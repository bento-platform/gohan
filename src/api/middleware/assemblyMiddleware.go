package middleware

import (
	assid "gohan/api/models/constants/assembly-id"
	"net/http"

	"github.com/labstack/echo"
)

/*
Echo middleware to ensure a valid `assemblyId` HTTP query parameter was provided
*/
func MandateAssemblyIdAttribute(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		// check for assemblyId query parameter
		assemblyId := c.QueryParam("assemblyId")
		if len(assemblyId) == 0 || !assid.IsKnownAssemblyId(assemblyId) {
			// if no id was provided, or it was invalid, return an error
			return echo.NewHTTPError(http.StatusBadRequest, "Missing or unknown assemblyId!")
		}

		return next(c)
	}
}
