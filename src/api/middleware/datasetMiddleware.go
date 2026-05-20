package middleware

import (
	"fmt"
	"gohan/api/contexts"
	"gohan/api/models/dtos/errors"
	"net/http"

	"github.com/labstack/echo"
)

/*
Echo middleware to ensure a valid `dataset` HTTP query parameter was provided
*/
func MandateDatasetAttribute(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		dataset := c.QueryParam("dataset")
		if len(dataset) == 0 {
			return c.JSON(http.StatusBadRequest, errors.CreateSimpleBadRequest("missing dataset"))
		}

		gc := c.(*contexts.GohanContext)
		gc.Dataset = dataset

		return next(gc)
	}
}

func MandateDatasetPathParam(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		dataset := c.Param("dataset")
		if len(dataset) == 0 {
			return c.JSON(http.StatusBadRequest, errors.CreateSimpleBadRequest("missing dataset"))
		}

		gc := c.(*contexts.GohanContext)
		gc.Dataset = dataset

		return next(gc)
	}
}

func MandateDataTypePathParam(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		dataType := c.Param("dataType")
		if dataType != "variant" {
			fmt.Printf("Invalid data-type provided: %s\n", dataType)
			return c.JSON(http.StatusBadRequest, errors.CreateSimpleBadRequest(
				fmt.Sprintf("invalid data-type %s - please provide a valid data-type (e.g. \"variant\")", dataType),
			))
		}
		gc := c.(*contexts.GohanContext)
		gc.DataType = dataType
		return next(gc)
	}
}

/*
Echo middleware to ensure a `dataset` HTTP query parameter is valid if provided
*/
func OptionalDatasetAttribute(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		gc := c.(*contexts.GohanContext)

		dataset := c.QueryParam("dataset")
		if len(dataset) > 0 {
			gc.Dataset = dataset
		}

		return next(gc)
	}
}
