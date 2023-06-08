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
