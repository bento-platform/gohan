package middleware

import (
	"gohan/api/contexts"
	"gohan/api/models/constants/chromosome"
	"net/http"

	"github.com/labstack/echo"
)

/*
Echo middleware to ensure a valid `chromosome` HTTP query parameter was provided
*/
func ValidateOptionalChromosomeAttribute(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		gc := c.(*contexts.GohanContext)

		// check for chromosome query parameter
		chromQP := c.QueryParam("chromosome")

		// verify:
		if len(chromQP) > 0 && !chromosome.IsValidHumanChromosome(chromQP) {
			// if chromosome less than 1 or greater than 23
			// and not 'x', 'y' or 'm'
			return echo.NewHTTPError(http.StatusBadRequest, "Please provide a valid 'chromosome' (either 1-23, X, Y, or M)")
		}

		if len(chromQP) == 0 {
			// if no chromosome is provided, assume "wildcard" search
			gc.Chromosome = "*"
		} else {
			gc.Chromosome = chromQP
		}

		return next(gc)
	}
}
