package middleware

import (
	"gohan/api/contexts"
	"strings"

	"github.com/labstack/echo"
)

/*
Echo middleware to prepare the context for an optionall provided singular `id` HTTP query parameter
*/
func CalibrateOptionalSampleIdsSingularAttribute(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		gc := c.(*contexts.GohanContext)

		// check for id query parameter
		sampleId := c.QueryParam("id")
		if len(sampleId) == 0 {
			sampleId = "*" // wildcard
		}

		gc.SampleIds = append(gc.SampleIds, sampleId)
		return next(gc)
	}
}

/*
Echo middleware to prepare the context for an optionally provided pluralized `id` (spelled `ids`) HTTP query parameter
*/
func CalibrateOptionalSampleIdsPluralAttribute(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		gc := c.(*contexts.GohanContext)

		// check for id's query parameter
		var (
			sampleIdQP = c.QueryParam("ids")
			sampleIds  []string
		)
		if len(sampleIdQP) > 0 {
			sampleIds = strings.Split(sampleIdQP, ",")
		} else {
			sampleIds = append(sampleIds, "*") // wildcard
		}

		gc.SampleIds = append(gc.SampleIds, sampleIds...)
		return next(gc)
	}
}
