package middleware

import (
	"gohan/api/contexts"
	"net/http"
	"strings"

	"github.com/labstack/echo"
)

/*
Echo middleware to ensure a singular `id` HTTP query parameter was provided
*/
func MandateSampleIdsSingularAttribute(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		gc := c.(*contexts.GohanContext)

		// check for id query parameter
		sampleId := c.QueryParam("id")
		if len(sampleId) == 0 {
			// if no id was provided return an error
			return echo.NewHTTPError(http.StatusBadRequest, "Missing 'id' query parameter for sample id querying!")
		}

		gc.SampleIds = append(gc.SampleIds, sampleId)
		return next(gc)
	}
}

/*
Echo middleware to ensure a pluralized `id` (spelled `ids`) HTTP query parameter was provided
*/
func MandateSampleIdsPluralAttribute(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		gc := c.(*contexts.GohanContext)

		// check for id's query parameter
		sampleIdQP := c.QueryParam("ids")
		if len(sampleIdQP) == 0 {
			// if no ids were provided return an error
			return echo.NewHTTPError(http.StatusBadRequest, "Missing 'ids' query parameter for sample id querying!")
		}
		sampleIds := strings.Split(sampleIdQP, ",")

		gc.SampleIds = append(gc.SampleIds, sampleIds...)
		return next(gc)
	}
}
