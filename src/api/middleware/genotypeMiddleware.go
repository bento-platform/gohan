package middleware

import (
	"fmt"
	"net/http"

	gq "api/models/constants/genotype-query"

	"github.com/labstack/echo"
)

func ValidatePotentialGenotypeQueryParameter(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		genotypeQP := c.QueryParam("genotype")

		if len(genotypeQP) > 0 {
			_, genotypeErr := gq.CastToGenoType(genotypeQP)
			if genotypeErr != nil {
				return echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("Invalid genotype query %s, %s", genotypeQP, genotypeErr))
			}
		}

		return next(c)
	}
}