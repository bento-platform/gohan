package mvc

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"api/contexts"
	"api/models/constants"
	"api/models/dtos"
	"api/models/indexes"
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

	log.Println("Incoming:")
	log.Println(t)

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
		// TODO: homogenize response model
		return c.JSON(http.StatusOK, map[string]interface{}{
			"code": 400,
			"errors": []map[string]interface{}{
				{
					"message": "Missing both id and data type - please provide at least one of them",
				},
			},
			"message":   "Bad Request",
			"timestamp": time.Now(),
		})
	} else if dataType != "" {
		// ensure data_type is valid ('variant', etc..)
		if !utils.StringInSlice(dataType, constants.ValidTableDataTypes) {
			return c.JSON(http.StatusBadRequest, dtos.CreateTableResponseDto{
				Error: fmt.Sprintf("Invalid data_type: %s -- Must be one of the following: %s", dataType, constants.ValidTableDataTypes),
			})
		}
	}

	// call repository
	results, _ := esRepo.GetTables(c, tableId, dataType)

	// gather data from "hits"
	docsHits := results["hits"].(map[string]interface{})["hits"]
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
	var es, chromosome, lowerBound, upperBound, reference, alternative, genotype, assemblyId, _ = retrieveCommonElements(c)
	// unused tableId from query parameter set to '_'

	// table id must be provided
	if tableId == "" {
		fmt.Println("Missing table id")

		// TODO: formalize response dto model
		return c.JSON(http.StatusBadRequest, map[string]interface{}{
			"code": 400,
			"errors": []map[string]interface{}{
				{
					"message": "Missing table id - please try again",
				},
			},
			"message":   "Bad Request",
			"timestamp": time.Now(),
		})
	}

	// call repository
	// - get the table by id
	results, getTablesError := esRepo.GetTables(c, tableId, "")
	if getTablesError != nil {
		fmt.Printf("Failed to get tables with ID %s\n", tableId)
		return c.JSON(http.StatusInternalServerError, map[string]interface{}{
			"code": 500,
			"errors": []map[string]interface{}{
				{
					"message": "Something went wrong.. Please try again later!",
				},
			},
			"message": "Internal Server Error", "timestamp": time.Now(),
		})
	}

	// gather data from "hits"
	docsHits := results["hits"].(map[string]interface{})["hits"]
	if docsHits == nil {
		fmt.Printf("No Tables with ID '%s' were deleted\n", tableId)
		return c.JSON(http.StatusBadRequest, map[string]interface{}{
			"code": 400,
			"errors": []map[string]interface{}{
				{
					"message": fmt.Sprintf("Table with ID %s not found", tableId),
				},
			},
			"message": "Bad Request", "timestamp": time.Now(),
		})
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
		fmt.Printf("No Variants associated with table ID '%s'\n", tableId)
		return c.JSON(http.StatusBadRequest, map[string]interface{}{
			"code": 400,
			"errors": []map[string]interface{}{
				{
					"message": fmt.Sprintf("Failed to get table summary with ID %s", tableId),
				},
			},
			"message": "Bad Request", "timestamp": time.Now(),
		})
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
		return c.JSON(http.StatusInternalServerError, map[string]interface{}{
			"code": 500,
			"errors": []map[string]interface{}{
				{
					"message": "Something went wrong.. Please try again later!",
				},
			},
			"message": "Internal Server Error", "timestamp": time.Now(),
		})
	}

	totalVariantsCount = docs["count"].(float64)

	fmt.Printf("Successfully Obtained Table ID '%s' Summary \n", tableId)

	return c.JSON(http.StatusOK, &dtos.TableSummaryResponseDto{
		Count:            int(totalVariantsCount),
		DataTypeSpecific: map[string]interface{}{},
	})
}

func DeleteTable(c echo.Context) error {
	fmt.Printf("[%s] - DeleteTable hit!\n", time.Now())

	// obtain tableId from the path
	tableId := c.Param("id")

	// at least one of these parameters must be present
	if tableId == "" {
		fmt.Println("Missing table id")

		// TODO: formalize response dto model
		return c.JSON(http.StatusOK, map[string]interface{}{
			"code": 400,
			"errors": []map[string]interface{}{
				{
					"message": "Missing table id - please try again",
				},
			},
			"message":   "Bad Request",
			"timestamp": time.Now(),
		})
	}

	// call repository
	results, deleteError := esRepo.DeleteTableById(c, tableId)
	if deleteError != nil {
		fmt.Printf("Failed to delete tables with ID %s\n", tableId)
		return c.JSON(http.StatusInternalServerError, map[string]interface{}{
			"code": 500,
			"errors": []map[string]interface{}{
				{
					"message": "Something went wrong.. Please try again later!",
				},
			},
			"message": "Internal Server Error", "timestamp": time.Now(),
		})
	}

	// gather data from "deleted"
	numDeleted := 0.0
	docsHits := results["deleted"]
	if docsHits != nil {
		numDeleted = docsHits.(float64)
	} else {
		fmt.Printf("No Tables with ID '%s' were deleted\n", tableId)
		return c.JSON(http.StatusBadRequest, map[string]interface{}{
			"code": 400,
			"errors": []map[string]interface{}{
				{
					"message": fmt.Sprintf("Failed to delete tables with ID %s", tableId),
				},
			},
			"message": "Bad Request", "timestamp": time.Now(),
		})
	}
	if numDeleted == 0 {
		fmt.Printf("No Tables with ID '%s' were deleted\n", tableId)
		return c.JSON(http.StatusNotFound, map[string]interface{}{
			"code": 404,
			"errors": []map[string]interface{}{
				{
					"message": fmt.Sprintf("No table with ID %s", tableId),
				},
			},
			"message": "Not Found", "timestamp": time.Now(),
		})
	}

	// delete variants associated with this table id
	deletedVariants, deleteVariantsError := esRepo.DeleteVariantsByTableId(c, tableId)
	if deleteVariantsError != nil {
		fmt.Printf("Failed to delete variants associated with table ID %s\n", tableId)
		return c.JSON(http.StatusInternalServerError, map[string]interface{}{
			"code": 500,
			"errors": []map[string]interface{}{
				{
					"message": "Something went wrong.. Please try again later!",
				},
			},
			"message": "Internal Server Error", "timestamp": time.Now(),
		})
	}
	deletedVariantsResults := deletedVariants["deleted"]
	numDeletedVariants := 0.0
	if deletedVariantsResults != nil {
		numDeletedVariants = deletedVariantsResults.(float64)
	} else {
		fmt.Printf("No Tables with ID '%s' were deleted\n", tableId)
		return c.JSON(http.StatusBadRequest, map[string]interface{}{
			"code": 400,
			"errors": []map[string]interface{}{
				{
					"message": fmt.Sprintf("Failed to delete tables with ID %s", tableId),
				},
			},
			"message": "Bad Request", "timestamp": time.Now(),
		})
	}

	fmt.Printf("Successfully Deleted Table(s) with ID '%s' with %f variants!\n", tableId, numDeletedVariants)
	return c.NoContent(204)
}
