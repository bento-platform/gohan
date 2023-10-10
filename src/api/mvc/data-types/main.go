package dataTypes

import (
	"net/http"

	"gohan/api/contexts"
	"gohan/api/models/schemas"

	variantService "gohan/api/services/variants"

	"github.com/labstack/echo"
)

var variantDataTypeJson = map[string]interface{}{
	"id":              "variant",
	"label":           "Variants",
	"queryable":       true,
	"schema":          schemas.VARIANT_SCHEMA,
	"metadata_schema": schemas.OBJECT_SCHEMA,
}

func fetchVariantData(c echo.Context) (map[string]interface{}, error) {
	gc := c.(*contexts.GohanContext)
	cfg := gc.Config
	es := gc.Es7Client

	resultsMap, err := variantService.GetVariantsOverview(es, cfg)
	if err != nil {
		return nil, err
	}

	variantDataTypeJson["count"] = sumAllValues(resultsMap["sampleIDs"])
	if latestCreated, ok := resultsMap["last_created_time"].(string); ok {
		variantDataTypeJson["last_ingested"] = latestCreated
	}

	return variantDataTypeJson, nil
}

func GetDataTypes(c echo.Context) error {
	variantData, err := fetchVariantData(c)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]interface{}{
			"error": err.Error(),
		})
	}

	return c.JSON(http.StatusOK, []map[string]interface{}{
		variantData,
	})
}

func GetReducedDataTypes(c echo.Context) error {
	variantData, err := fetchVariantData(c)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]interface{}{
			"error": err.Error(),
		})
	}

	if variantData == nil {
		return c.JSON(http.StatusInternalServerError, map[string]interface{}{
			"error": "Failed to retrieve variant data.",
		})
	}

	count, _ := variantData["count"]
	id, _ := variantData["id"].(string)
	label, _ := variantData["label"].(string)
	last_ingested, _ := variantData["last_ingested"].(string)
	queryable, _ := variantData["queryable"].(bool)

	// Create a reduced response
	reducedResponse := map[string]interface{}{
		"count":         count,
		"id":            id,
		"label":         label,
		"last_ingested": last_ingested,
		"queryable":     queryable,
	}

	return c.JSON(http.StatusOK, []map[string]interface{}{
		reducedResponse,
	})
}

func GetVariantDataType(c echo.Context) error {
	return c.JSON(http.StatusOK, variantDataTypeJson)
}

func GetVariantDataTypeSchema(c echo.Context) error {
	return c.JSON(http.StatusOK, schemas.VARIANT_SCHEMA)
}

func GetVariantDataTypeMetadataSchema(c echo.Context) error {
	return c.JSON(http.StatusOK, schemas.VARIANT_METADATA_SCHEMA)
}

// - helpers
func sumAllValues(keyedValues interface{}) float64 {
	tmpValueStrings := keyedValues.(map[string]interface{})
	sum := 0.0
	for _, k := range tmpValueStrings {
		sum += k.(float64)
	}
	return sum
}
