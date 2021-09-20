package middleware

import (
	"net/http"
	"strconv"

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
		i, conversionErr := strconv.Atoi(chromQP)
		if conversionErr != nil {
			// if invalid chromosome
			return echo.NewHTTPError(http.StatusBadRequest, "Error converting 'chromosome' query parameter! Check your input")
		}

		if i <= 0 {
			// if chromosome less than 0
			return echo.NewHTTPError(http.StatusBadRequest, "Please provide a 'chromosome' greater than 0!")
		}

		return next(c)
	}
}
