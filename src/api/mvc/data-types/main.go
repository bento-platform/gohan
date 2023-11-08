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
	variantDataTypeJson["last_ingested"] = resultsMap["last_ingested"]

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
