package middleware

import (
	"gohan/api/contexts"
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
		if len(assemblyId) == 0 {
			// if no id was provided, return an error
			return echo.NewHTTPError(http.StatusBadRequest, "Missing assemblyId!")
		}

		// forward a type-safe value down the pipeline
		gc := c.(*contexts.GohanContext)
		gc.AssemblyId = assemblyId

		return next(gc)
	}
}
