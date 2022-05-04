package tables

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"api/contexts"
	"api/models/constants"
	"api/models/dtos"
	"api/models/dtos/errors"
	"api/models/indexes"
	"api/mvc"
	esRepo "api/repositories/elasticsearch"
	"api/utils"

	"github.com/labstack/echo"
	"github.com/mitchellh/mapstructure"
)

func CreateTable(c echo.Context) error {
	fmt.Printf("[%s] - CreateTable hit!\n", time.Now())

	decoder := json.NewDecoder(c.Request().Body)
	var t dtos.CreateTableRequestDto
	err := decoder.Decode(&t)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]interface{}{
			"error": err,
		})
	}

	// TODO: improve verification
	if t.Name == "" {
		return c.JSON(http.StatusBadRequest, dtos.CreateTableResponseDto{
			Error: "'name' cannot be empty",
		})
	} else if t.Dataset == "" {
		return c.JSON(http.StatusBadRequest, dtos.CreateTableResponseDto{
			Error: "'dataset' cannot be empty",
		})
	} else if t.DataType == "" {
		return c.JSON(http.StatusBadRequest, dtos.CreateTableResponseDto{
			Error: "'data_type' cannot be empty",
		})
	}

	// ensure data_type is valid ('variant', etc..)
	if !utils.StringInSlice(t.DataType, constants.ValidTableDataTypes) {
		return c.JSON(http.StatusBadRequest, dtos.CreateTableResponseDto{
			Error: fmt.Sprintf("Invalid data_type: %s -- Must be one of the following: %s", t.DataType, constants.ValidTableDataTypes),
		})
	}

	// TODO: ensure dataset is a valid identifier (uuid ?)

	// avoid creating duplicate tables with the same name
	existingTables, error := esRepo.GetTablesByName(c, t.Name)
	if error != nil {
		return c.JSON(http.StatusInternalServerError, dtos.CreateTableResponseDto{
			Error: error.Error(),
		})
	}
	if len(existingTables) > 0 {
		return c.JSON(http.StatusBadRequest, dtos.CreateTableResponseDto{
			Error: fmt.Sprintf("A table with the name '%s' already exists", t.Name),
		})
	}

	// call repository
	table, error := esRepo.CreateTable(c, t)
	if error != nil {
		return c.JSON(http.StatusInternalServerError, dtos.CreateTableResponseDto{
			Error: error.Error(),
		})
	}

	return c.JSON(http.StatusOK, dtos.CreateTableResponseDto{
		Message: "Success",
		Table:   table,
	})
}

func GetTables(c echo.Context) error {
	fmt.Printf("[%s] - GetTables hit!\n", time.Now())

	// obtain tableId from the path
	tableId := c.Param("id")

	// obtain dataTypes from query parameter
	dataType := c.QueryParam("data-type")

	// at least one of these parameters must be present
	if tableId == "" && dataType == "" {
		return c.JSON(http.StatusBadRequest, errors.CreateSimpleBadRequest("Missing both id and data type - please provide at least one of them"))
	} else if dataType != "" {
		// ensure data_type is valid ('variant', etc..)
		if !utils.StringInSlice(dataType, constants.ValidTableDataTypes) {
			return c.JSON(http.StatusBadRequest, errors.CreateSimpleBadRequest(fmt.Sprintf("Invalid data_type: %s -- Must be one of the following: %s", dataType, constants.ValidTableDataTypes)))
		}
	}

	// call repository
	results, _ := esRepo.GetTables(c, tableId, dataType)
	if results == nil {
		// return empty result (assume there are no tables because the index doesn't exist)
		return c.JSON(http.StatusOK, []map[string]interface{}{})
	}
	// TODO: handle _ error better

	// gather data from "hits"
	docsHits := results["hits"].(map[string]interface{})["hits"]
	allDocHits := []map[string]interface{}{}
	mapstructure.Decode(docsHits, &allDocHits)

	// grab _source for each hit
	allSources := make([]indexes.Table, 0)

	for _, r := range allDocHits {
		source := r["_source"]
		byteSlice, _ := json.Marshal(source)

		// cast map[string]interface{} to table
		var resultingTable indexes.Table
		if err := json.Unmarshal(byteSlice, &resultingTable); err != nil {
			fmt.Println("failed to unmarshal:", err)
		}

		// accumulate structs
		allSources = append(allSources, resultingTable)
	}

	if tableId != "" && len(allSources) > 0 {
		// assume there is only 1 document in the database with this `id`
		// return a single object rather than the whole list
		return c.JSON(http.StatusOK, allSources[0])
	}

	return c.JSON(http.StatusOK, allSources)
}

func GetTableSummary(c echo.Context) error {
	fmt.Printf("[%s] - GetTableSummary hit!\n", time.Now())

	cfg := c.(*contexts.GohanContext).Config

	// obtain tableId from the path
	tableId := c.Param("id")
	// obtain other potentially relevant parameters from available query parameters
	// (these should be empty, but utilizing this common function is convenient to set up
	// the call to the variants index through the repository functions)
	var es, chromosome, lowerBound, upperBound, reference, alternative, genotype, assemblyId, _ = mvc.RetrieveCommonElements(c)
	// unused tableId from query parameter set to '_'

	// table id must be provided
	if tableId == "" {
		fmt.Println("Missing table id")
		return c.JSON(http.StatusBadRequest, errors.CreateSimpleBadRequest("Missing table id - please try again"))
	}

	// call repository
	// - get the table by id
	results, getTablesError := esRepo.GetTables(c, tableId, "")
	if getTablesError != nil {
		fmt.Printf("Failed to get tables with ID %s\n", tableId)
		return c.JSON(http.StatusInternalServerError, errors.CreateSimpleInternalServerError("Something went wrong.. Please try again later!"))
	}

	// gather data from "hits"
	docsHits := results["hits"].(map[string]interface{})["hits"]
	if docsHits == nil {
		fmt.Printf("No Tables with ID '%s' were found\n", tableId)
		return c.JSON(http.StatusBadRequest, errors.CreateSimpleBadRequest(fmt.Sprintf("Table with ID %s not found", tableId)))
	}

	// obtain hits (expecting 1)
	allDocHits := []map[string]interface{}{}
	mapstructure.Decode(docsHits, &allDocHits)

	// grab _source for each hit
	allSources := make([]interface{}, 0)
	// var allSources []indexes.Variant

	for _, r := range allDocHits {
		source := r["_source"]
		byteSlice, _ := json.Marshal(source)

		// cast map[string]interface{} to table
		var resultingTable indexes.Table
		if err := json.Unmarshal(byteSlice, &resultingTable); err != nil {
			fmt.Println("failed to unmarshal:", err)
		}

		// accumulate structs
		allSources = append(allSources, resultingTable)
	}

	if len(allSources) == 0 {
		fmt.Printf("Failed to get table summary with ID '%s'\n", tableId)
		return c.JSON(http.StatusBadRequest, errors.CreateSimpleBadRequest(fmt.Sprintf("Failed to get table summary with ID %s", tableId)))
	}

	// obtain table id from the one expected hit
	// and search for variants associated with it

	totalVariantsCount := 0.0

	docs, countError := esRepo.CountDocumentsContainerVariantOrSampleIdInPositionRange(cfg, es,
		chromosome, lowerBound, upperBound,
		"", "", // note : both variantId and sampleId are deliberately set to ""
		reference, alternative, genotype, assemblyId, tableId)
	if countError != nil {
		fmt.Printf("Failed to count variants with table ID %s\n", tableId)
		return c.JSON(http.StatusInternalServerError, errors.CreateSimpleInternalServerError("Something went wrong.. Please try again later!"))
	}

	totalVariantsCount = docs["count"].(float64)

	// obtain number of samples associated with this tableId
	resultingBuckets, bucketsError := esRepo.GetVariantsBucketsByKeywordAndTableId(cfg, es, "sample.id.keyword", tableId)
	if bucketsError != nil {
		fmt.Println(resultingBuckets)
	}

	// retrieve aggregations.items.buckets
	// and count number of samples
	bucketsMapped := []interface{}{}
	if aggs, aggsOk := resultingBuckets["aggregations"]; aggsOk {
		aggsMapped := aggs.(map[string]interface{})

		if items, itemsOk := aggsMapped["items"]; itemsOk {
			itemsMapped := items.(map[string]interface{})

			if buckets, bucketsOk := itemsMapped["buckets"]; bucketsOk {
				bucketsMapped = buckets.([]interface{})
			}
		}
	}

	fmt.Printf("Successfully Obtained Table ID '%s' Summary \n", tableId)

	return c.JSON(http.StatusOK, &dtos.TableSummaryResponseDto{
		Count: int(totalVariantsCount),
		DataTypeSpecific: map[string]interface{}{
			"samples": len(bucketsMapped),
		},
	})
}

func DeleteTable(c echo.Context) error {
	fmt.Printf("[%s] - DeleteTable hit!\n", time.Now())

	// obtain tableId from the path
	tableId := c.Param("id")

	// at least one of these parameters must be present
	if tableId == "" {
		fmt.Println("Missing table id")
		return c.JSON(http.StatusBadRequest, errors.CreateSimpleBadRequest("Missing table id - please try again"))
	}

	// call repository
	results, deleteError := esRepo.DeleteTableById(c, tableId)
	if deleteError != nil {
		fmt.Printf("Failed to delete tables with ID %s\n", tableId)
		return c.JSON(http.StatusInternalServerError, errors.CreateSimpleInternalServerError("Something went wrong.. Please try again later!"))
	}

	// gather data from "deleted"
	numDeleted := 0.0
	docsHits := results["deleted"]
	if docsHits != nil {
		numDeleted = docsHits.(float64)
	} else {
		fmt.Printf("No Tables with ID '%s' were deleted\n", tableId)
		return c.JSON(http.StatusBadRequest, errors.CreateSimpleBadRequest(fmt.Sprintf("Failed to delete tables with ID %s", tableId)))
	}
	if numDeleted == 0 {
		fmt.Printf("No Tables with ID '%s' were deleted\n", tableId)
		return c.JSON(http.StatusNotFound, errors.CreateSimpleNotFound(fmt.Sprintf("No table with ID %s", tableId)))
	}

	// TODO: spin the deletion of variants associated with
	// the tableId requested off in a go routine if the table
	// was successfully deleted and assume the deletion completes
	// successfully in the background

	// TODO: ensure that no variants exist without a valid tableId

	// delete variants associated with this table id
	deletedVariants, deleteVariantsError := esRepo.DeleteVariantsByTableId(c, tableId)
	if deleteVariantsError != nil {
		fmt.Printf("Failed to delete variants associated with table ID %s\n", tableId)
		return c.JSON(http.StatusInternalServerError, errors.CreateSimpleInternalServerError("Something went wrong.. Please try again later!"))
	}
	deletedVariantsResults := deletedVariants["deleted"]
	numDeletedVariants := 0.0
	if deletedVariantsResults != nil {
		numDeletedVariants = deletedVariantsResults.(float64)
	} else {
		msg := fmt.Sprintf("Failed to delete tables with ID %s", tableId)
		fmt.Println(msg)
		return c.JSON(http.StatusBadRequest, errors.CreateSimpleBadRequest(msg))
	}

	fmt.Printf("Successfully Deleted Table(s) with ID '%s' with %f variants!\n", tableId, numDeletedVariants)
	return c.NoContent(204)
}