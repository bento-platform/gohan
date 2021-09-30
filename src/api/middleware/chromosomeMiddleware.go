package middleware

import (
	"api/models/constants/chromosome"
	"net/http"

	"github.com/labstack/echo"
)

/*
	Echo middleware to ensure a valid `chromosome` HTTP query parameter was provided
*/
func MandateChromosomeAttribute(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		// check for chromosome query parameter
		chromQP := c.QueryParam("chromosome")
		if len(chromQP) == 0 {
			// if no id was provided return an error
			return echo.NewHTTPError(http.StatusBadRequest, "Missing 'chromosome' query parameter for querying!")
		}

		// verify:
		if !chromosome.IsValidHumanChromosome(chromQP) {
			// if chromosome less than 1 or greater than 23
			// and not 'x', 'y' or 'm'
			return echo.NewHTTPError(http.StatusBadRequest, "Please provide a 'chromosome' greater than 0!")
		}

		return next(c)
	}
}
