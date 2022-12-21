package middleware

import (
	"fmt"
	"gohan/api/models/dtos/errors"
	"gohan/api/utils"
	"net/http"

	"github.com/labstack/echo"
)

/*
Echo middleware to ensure a valid `tableId` HTTP query parameter was provided
*/
func MandateTableIdAttribute(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		// check for tableId query parameter
		tableId := c.QueryParam("tableId")
		if len(tableId) == 0 {
			// if no id was provided, or is invalid, return an error
			return c.JSON(http.StatusBadRequest, errors.CreateSimpleBadRequest("missing table id"))
		}

		// verify tableId is a valid UUID
		// - assume it's a valid table id if it's a uuid,
		//   further verification is done later
		if !utils.IsValidUUID(tableId) {
			fmt.Printf("Invalid table id %s\n", tableId)

			return c.JSON(http.StatusBadRequest, errors.CreateSimpleBadRequest(fmt.Sprintf("invalid table id %s - please provide a valid uuid", tableId)))
		}

		return next(c)
	}
}
