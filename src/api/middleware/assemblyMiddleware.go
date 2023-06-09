package middleware

import (
	"gohan/api/contexts"
	"gohan/api/models/constants"
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

		// forward a type-safe value down the pipeline
		gc := c.(*contexts.GohanContext)
		gc.AssemblyId = constants.AssemblyId(assemblyId)

		return next(gc)
	}
}

/*
Echo middleware to ensure the optional `assemblyId`
HTTP query parameter is valid if present
*/
func OptionalAssemblyIdAttribute(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		gc := c.(*contexts.GohanContext)

		// check for assemblyId query parameter
		assemblyId := c.QueryParam("assemblyId")
		if len(assemblyId) > 0 {
			if !assid.IsKnownAssemblyId(assemblyId) {
				// if an id was provided and was invalid, return an error
				return echo.NewHTTPError(http.StatusBadRequest, "Invalid assemblyId!")
			}
			// forward a type-safe value down the pipeline
			gc.AssemblyId = constants.AssemblyId(assemblyId)
		} else {
			gc.AssemblyId = assid.Unknown
		}

		return next(gc)
	}
}
