package mvc

import (
	"gohan/api/contexts"
	"gohan/api/models/constants"
	"strings"

	"github.com/elastic/go-elasticsearch/v7"
	"github.com/labstack/echo"
)

func RetrieveCommonElements(c echo.Context) (*elasticsearch.Client, string, int, int, string, string, []string, constants.GenotypeQuery, constants.AssemblyId) {
	gc := c.(*contexts.GohanContext)
	es := gc.Es7Client

	chromosome := gc.Chromosome

	lowerBound := gc.LowerBound
	upperBound := gc.UpperBound

	reference := c.QueryParam("reference")
	alternative := c.QueryParam("alternative")

	alleles := gc.Alleles

	// reference, alternative and alleles can have the
	// single-wildcard character 'N', which adheres to
	// the spec found at : https://droog.gs.washington.edu/parc/images/iupac.html

	// swap all 'N's into '?'s for elasticsearch
	reference = strings.Replace(reference, "N", "?", -1)
	alternative = strings.Replace(alternative, "N", "?", -1)

	genotype := gq.UNCALLED
	genotypeQP := c.QueryParam("genotype")
	if len(genotypeQP) > 0 {
		if parsedGenotype, gErr := gq.CastToGenoType(genotypeQP); gErr == nil {
			genotype = parsedGenotype
		}
	}

	assemblyId := a.Unknown
	assemblyIdQP := c.QueryParam("assemblyId")
	if len(assemblyIdQP) > 0 && a.IsKnownAssemblyId(assemblyIdQP) {
		assemblyId = a.CastToAssemblyId(assemblyIdQP)
	}

	return es, chromosome, lowerBound, upperBound, reference, alternative, alleles, genotype, assemblyId
}
