package middleware

import (
	"fmt"
	"net/http"
	"strconv"

	z "api/models/constants/zygosity"

	"github.com/labstack/echo"
)

func ValidatePotentialGenotypeQueryParameter(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		genotypeQP := c.QueryParam("genotype")

		// TODO: improve checking
		if len(genotypeQP) > 0 {
			genInt, genotypeErr := strconv.Atoi(genotypeQP)

			if genotypeErr != nil || !z.IsValidQuery(genInt) {
				return echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("Invalid genotype query %s", genotypeQP))
			}
		}

		return next(c)
	}
}
