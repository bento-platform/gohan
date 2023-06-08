package middleware

import (
	"gohan/api/contexts"
	"gohan/api/models/dtos/errors"
	"net/http"
	"strings"

	"github.com/labstack/echo"
)

func MandateCalibratedAlleles(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		gc := c.(*contexts.GohanContext)

		var alleles []string
		var allelesQP string = c.QueryParam("alleles")
		if len(allelesQP) > 0 {
			alleles = strings.Split(allelesQP, ",")

			// ensure the allele query is properly formatted
			if alleles[len(alleles)-1] == "" {
				return echo.NewHTTPError(
					http.StatusBadRequest,
					errors.CreateSimpleBadRequest("Found an empty allele! Please double check your request!"))
			}

			// ensure no more than 2 alleles are provided at once
			if len(alleles) > 2 {
				return echo.NewHTTPError(
					http.StatusBadRequest,
					errors.CreateSimpleBadRequest("Too many alleles! Please only provide 1 or 2"))
			}

			for i, a := range alleles {
				alleles[i] = strings.Replace(a, "N", "?", -1)
				// alleles = append(alleles, strings.Replace(a, "N", "?", -1))
			}
		}

		gc.Alleles = alleles
		return next(c)
	}
}
