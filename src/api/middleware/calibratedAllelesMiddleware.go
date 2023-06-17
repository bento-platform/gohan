package middleware

import (
	"fmt"
	"gohan/api/contexts"
	"gohan/api/models/dtos/errors"
	"gohan/api/utils"
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

			// check validity of each provided character
			for i, allele := range alleles {
				// - accept lower cases, but transform to upper for elasticsearch (case sensitive)
				upperAllele := strings.ToUpper(allele)
				// - check validity of the rest
				for _, nuc := range upperAllele {
					if !utils.StringInSlice(string(nuc), utils.AcceptedNucleotideCharacters) {
						// return status 400 if any allele is incorrect
						return echo.NewHTTPError(
							http.StatusBadRequest,
							errors.CreateSimpleBadRequest(fmt.Sprintf("Nucleotide %s unacceptable! Please double check and try again.", string(nuc))))
					}
				}
				// - replace 'N' with uppercase '?' for elasticsearch
				upperAllele = strings.Replace(upperAllele, "N", "?", -1)

				alleles[i] = upperAllele
			}
		}

		gc.Alleles = alleles
		return next(gc)
	}
}
