package middleware

import (
	"fmt"
	"net/http"

	"gohan/api/contexts"
	"gohan/api/models/constants"
	gq "gohan/api/models/constants/genotype-query"

	"github.com/labstack/echo"
)

/*
Echo middleware to ensure the validity of the optionally provided `genotype` HTTP query parameter
*/
func ValidatePotentialGenotypeQueryParameter(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		gc := c.(*contexts.GohanContext)

		// check for a genotype query parameter
		var (
			genotype    constants.GenotypeQuery
			genotypeErr error
		)
		genotypeQP := c.QueryParam("genotype")
		if len(genotypeQP) > 0 {
			// validate that the genotype string
			// converts to a valid GenotypeQuery
			// (skips "UNCALLED" values)
			genotype, genotypeErr = gq.CastToGenoType(genotypeQP)
			if genotypeErr != nil {
				return echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("Invalid genotype query %s, %s", genotypeQP, genotypeErr))
			}
		}

		gc.Genotype = genotype
		return next(gc)
	}
}
