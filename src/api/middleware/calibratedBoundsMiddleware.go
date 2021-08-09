package middleware

import (
	"net/http"
	"strconv"

	"github.com/labstack/echo"
)

func MandateCalibratedBounds(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		var (
			lowerBound int
			upperBound int

			lowerBoundPointer *int // simulate "nullable" int
			upperBoundPointer *int
		)

		lowerBoundQP := c.QueryParam("lowerBound")
		if len(lowerBoundQP) > 0 {
			lb, conversionErr := strconv.Atoi(lowerBoundQP)
			if conversionErr == nil {
				lowerBound = lb
				lowerBoundPointer = &lowerBound
			}
		}

		upperBoundQP := c.QueryParam("upperBound")
		if len(upperBoundQP) > 0 {
			ub, conversionErr := strconv.Atoi(upperBoundQP)
			if conversionErr == nil {
				upperBound = ub
				upperBoundPointer = &upperBound
			}
		}

		// allow call to pass if and only if:
		// - neither upper and lower bound parameters a provided
		// - both are provided
		// -- and if both are provided, that they are balanced
		if (upperBoundPointer == nil && lowerBoundPointer != nil) ||
			(upperBoundPointer != nil && lowerBoundPointer == nil) ||
			(upperBoundPointer != nil && lowerBoundPointer != nil && upperBound < lowerBound) {
			// if upper bound is less than the lower bound
			return echo.NewHTTPError(http.StatusBadRequest, "Invalid lower and upper bounds!")
		}

		return next(c)
	}
}
