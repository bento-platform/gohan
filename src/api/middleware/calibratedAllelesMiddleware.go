package middleware

import (
	"net/http"
	"strings"

	"github.com/labstack/echo"
)

func MandateCalibratedAlleles(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		allelesQP := strings.Split(c.QueryParam("alleles"), ",")

		// ensure no more than 2 alleles are provided at once
		if len(allelesQP) > 2 {
			return echo.NewHTTPError(http.StatusBadRequest, "Too many alleles! Please only provide 1 or 2")
		}

		return next(c)
	}
}
