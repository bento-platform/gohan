package middleware

import (
	"gohan/api/contexts"

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

		if len(chromQP) == 0 {
			// if no chromosome is provided, assume "wildcard" search
			gc.Chromosome = "*"
		} else {
			gc.Chromosome = chromQP
		}

		return next(gc)
	}
}
