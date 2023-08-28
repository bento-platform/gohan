package middleware

import (
	"fmt"
	"gohan/api/contexts"
	"gohan/api/models/dtos/errors"
	"gohan/api/utils"
	"net/http"

	"github.com/google/uuid"
	"github.com/labstack/echo"
)

/*
Echo middleware to ensure a valid `dataset` HTTP query parameter was provided
*/
func MandateDatasetAttribute(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		// check for dataset query parameter
		dataset := c.QueryParam("dataset")
		if len(dataset) == 0 {
			// if no id was provided, or is invalid, return an error
			return c.JSON(http.StatusBadRequest, errors.CreateSimpleBadRequest("missing dataset"))
		}

		// verify dataset is a valid UUID
		// - assume it's a valid dataset if it's a uuid,
		//   further verification is done later
		if !utils.IsValidUUID(dataset) {
			fmt.Printf("Invalid dataset %s\n", dataset)

			return c.JSON(http.StatusBadRequest, errors.CreateSimpleBadRequest(fmt.Sprintf("invalid dataset %s - please provide a valid uuid", dataset)))
		}

		// forward a type-safe value down the pipeline
		gc := c.(*contexts.GohanContext)
		gc.Dataset = uuid.MustParse(dataset)

		return next(gc)
	}
}

func MandateDatasetPathParam(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		dataset := c.Param("dataset")
		if !utils.IsValidUUID(dataset) {
			fmt.Printf("Invalid dataset %s\n", dataset)

			return c.JSON(http.StatusBadRequest, errors.CreateSimpleBadRequest(fmt.Sprintf("invalid dataset %s - please provide a valid uuid", dataset)))
		}

		gc := c.(*contexts.GohanContext)
		gc.Dataset = uuid.MustParse(dataset)

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

		// check for dataset query parameter
		dataset := c.QueryParam("dataset")
		if len(dataset) > 0 {
			// verify dataset is a valid UUID
			// - assume it's a valid dataset if it's a uuid,
			//   further verification is done later
			if !utils.IsValidUUID(dataset) {
				fmt.Printf("Invalid dataset %s\n", dataset)

				return c.JSON(http.StatusBadRequest, errors.CreateSimpleBadRequest(fmt.Sprintf("invalid dataset %s - please provide a valid uuid", dataset)))
			}

			// forward a type-safe value down the pipeline
			gc.Dataset = uuid.MustParse(dataset)
		}

		return next(gc)
	}
}
