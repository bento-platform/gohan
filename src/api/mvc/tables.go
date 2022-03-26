package mvc

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"api/models/indexes"
	esRepo "api/repositories/elasticsearch"

	"github.com/labstack/echo"
	"github.com/mitchellh/mapstructure"
)

func CreateTable(c echo.Context) error {
	fmt.Printf("[%s] - CreateTable hit!\n", time.Now())

	// TODO: elaborate

	// call repository
	esRepo.CreateTable(c)

	return c.JSON(http.StatusOK, make(map[string]interface{}))
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
