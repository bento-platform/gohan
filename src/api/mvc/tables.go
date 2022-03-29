package mvc

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

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
	// can be any string -- expects "" by default

	// at least one of these parameters must be present
	if tableId == "" && dataType == "" {
		return c.JSON(http.StatusOK, map[string]interface{}{
			"code": 400,
			"errors": []map[string]interface{}{
				{
					"message": "Invalid or missing data type (specified ID: [])",
				},
			},
			"message":   "Bad Request",
			"timestamp": time.Now(),
		})
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

		// docId := r["_id"].(string)

		// cast map[string]interface{} to struct
		var resultingTable indexes.Table
		// mapstructure.Decode(source, &resultingTable)
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
