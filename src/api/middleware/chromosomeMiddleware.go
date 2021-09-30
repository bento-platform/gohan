package middleware

import (
	"api/models/constants/chromosome"
	"net/http"

	"github.com/labstack/echo"
)

/*
	Echo middleware to ensure a valid `chromosome` HTTP query parameter was provided
*/
func ValidateOptionalChromosomeAttribute(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		// check for chromosome query parameter
		chromQP := c.QueryParam("chromosome")

		// verify:
		if len(chromQP) > 0 && !chromosome.IsValidHumanChromosome(chromQP) {
			// if chromosome less than 1 or greater than 23
			// and not 'x', 'y' or 'm'
			return echo.NewHTTPError(http.StatusBadRequest, "Please provide a valid 'chromosome' (either 1-23, X, Y, or M)")
		}

		return next(c)
	}
}
