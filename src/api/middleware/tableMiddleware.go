package middleware

import (
	"api/models/dtos"
	"api/utils"
	"fmt"
	"net/http"
	"time"

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
			return c.JSON(http.StatusBadRequest, &dtos.GeneralErrorResponseDto{
				Code:      400,
				Message:   "Bad Request",
				Timestamp: time.Now(),
				Errors: []dtos.GeneralError{
					{
						Message: "Missing table id",
					},
				},
			})
		}

		// verify tableId is a valid UUID
		// - assume it's a valid table id if it's a uuid,
		//   further verification is done later
		if !utils.IsValidUUID(tableId) {
			fmt.Printf("Invalid table id %s\n", tableId)

			return c.JSON(http.StatusBadRequest, &dtos.GeneralErrorResponseDto{
				Code:      400,
				Message:   "Bad Request",
				Timestamp: time.Now(),
				Errors: []dtos.GeneralError{
					{
						Message: fmt.Sprintf("Invalid table id %s - please provide a valid UUID", tableId),
					},
				},
			})
		}

		return next(c)
	}
}
