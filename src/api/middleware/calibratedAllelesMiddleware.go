package middleware

import (
	"gohan/api/models/dtos/errors"
	"net/http"
	"strings"

	"github.com/labstack/echo"
)

func MandateCalibratedAlleles(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		if len(c.QueryParam("alleles")) > 0 {
			allelesQP := strings.Split(c.QueryParam("alleles"), ",")

			// ensure the allele query is properly formatted
			if allelesQP[len(allelesQP)-1] == "" {
				return echo.NewHTTPError(
					http.StatusBadRequest,
					errors.CreateSimpleBadRequest("Found an empty allele! Please double check your request!"))
			}

			// ensure no more than 2 alleles are provided at once
			if len(allelesQP) > 2 {
				return echo.NewHTTPError(
					http.StatusBadRequest,
					errors.CreateSimpleBadRequest("Too many alleles! Please only provide 1 or 2"))
			}
		}

		return next(c)
	}
}
