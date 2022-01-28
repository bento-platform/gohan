package mvc

import (
	"net/http"

	"api/models/schemas"

	"github.com/labstack/echo"
)

var variantDataTypeJson = map[string]interface{}{
	"id":              "variant",
	"schema":          schemas.VARIANT_SCHEMA,
	"metadata_schema": schemas.VARIANT_TABLE_METADATA_SCHEMA,
}

func GetDataTypes(c echo.Context) error {
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
	return c.JSON(http.StatusOK, schemas.VARIANT_TABLE_METADATA_SCHEMA)
}

func FakeBentoTables(c echo.Context) error {
	return c.JSON(http.StatusOK, []string{})
}

func FakeBentoTableSchema(c echo.Context) error {
	return c.JSON(http.StatusOK, map[string]interface{}{
		"assembly_ids": []string{
			"GRCh38",
			"GRCh37",
			"NCBI36",
			"Other",
		},
		"data_type": "variant",
		"id":        "fake",
		"metadata": map[string]string{
			"created": "2021-09-14T17:49:47.154843Z",
			"name":    "Fake Variants Table",
			"updated": "2021-09-14T17:49:47.154843Z",
		},
		"name":   "Fake Variants Table",
		"schema": schemas.VARIANT_SCHEMA,
	})
}

// # TODO: Consistent snake or kebab
