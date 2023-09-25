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

func GetDataTypes(c echo.Context) error {
	es := c.(*contexts.GohanContext).Es7Client
	cfg := c.(*contexts.GohanContext).Config

	// accumulate number of variants associated with each
	// sampleId fetched from the variants overview
	resultsMap, err := variantService.GetVariantsOverview(es, cfg)

	if err != nil {
		// Could not talk to Elasticsearch, return an error
		return c.JSON(http.StatusInternalServerError, map[string]interface{}{
			"error": err.Error(),
		})
	}

	variantDataTypeJson["count"] = sumAllValues(resultsMap["sampleIDs"])

	// Data types are basically stand-ins for schema blocks
	return c.JSON(http.StatusOK, []map[string]interface{}{
		variantDataTypeJson,
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
