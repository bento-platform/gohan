package middleware

import (
	"net/http"
	"strings"

	"github.com/labstack/echo"
)

/*
	Echo middleware to ensure a singular `id` HTTP query parameter was provided
*/
func MandateSampleIdsSingularAttribute(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		// check for id query parameter
		sampleId := c.QueryParam("id")
		if len(sampleId) == 0 {
			// if no id was provided return an error
			return echo.NewHTTPError(http.StatusBadRequest, "Missing 'id' query parameter for sample id querying!")
		}

		return next(c)
	}
}

/*
	Echo middleware to ensure a pluralized `id` (spelled `ids`) HTTP query parameter was provided
*/
func MandateSampleIdsPluralAttribute(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		// check for id's query parameter
		sampleIds := strings.Split(c.QueryParam("ids"), ",")
		if len(sampleIds) == 0 {
			// if no ids were provided return an error
			return echo.NewHTTPError(http.StatusBadRequest, "Missing 'ids' query parameter for sample id querying!")
		}

		return next(c)
	}
}
