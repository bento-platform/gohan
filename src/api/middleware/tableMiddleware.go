package middleware

import (
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
			return echo.NewHTTPError(http.StatusBadRequest, map[string]interface{}{
				"code": 400,
				"errors": []map[string]interface{}{
					{
						"message": "Missing table id",
					},
				},
				"message":   "Bad Request",
				"timestamp": time.Now(),
			})
		}

		// verify tableId is a valid UUID
		// - assume it's a valid table id if it's a uuid,
		//   further verification is done later
		if !utils.IsValidUUID(tableId) {
			fmt.Printf("Invalid table id %s\n", tableId)

			// TODO: formalize response dto model
			return c.JSON(http.StatusBadRequest, map[string]interface{}{
				"code": 400,
				"errors": []map[string]interface{}{
					{
						"message": fmt.Sprintf("Invalid table id %s - please provide a valid UUID", tableId),
					},
				},
				"message":   "Bad Request",
				"timestamp": time.Now(),
			})
		}

		return next(c)
	}
}
