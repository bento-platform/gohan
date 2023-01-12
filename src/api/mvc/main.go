package mvc

import (
	"gohan/api/contexts"
	"gohan/api/models/constants"
	a "gohan/api/models/constants/assembly-id"
	gq "gohan/api/models/constants/genotype-query"
	"log"
	"strconv"
	"strings"

	"github.com/elastic/go-elasticsearch/v7"
	"github.com/labstack/echo"
)

func RetrieveCommonElements(c echo.Context) (*elasticsearch.Client, string, int, int, string, string, []string, constants.GenotypeQuery, constants.AssemblyId, string) {
	es := c.(*contexts.GohanContext).Es7Client

	chromosome := c.QueryParam("chromosome")
	if len(chromosome) == 0 {
		// if no chromosome is provided, assume "wildcard" search
		chromosome = "*"
	}

	lowerBoundQP := c.QueryParam("lowerBound")
	var (
		lowerBound int
		lbErr      error
	)
	if len(lowerBoundQP) > 0 {
		lowerBound, lbErr = strconv.Atoi(lowerBoundQP)
		if lbErr != nil {
			log.Fatal(lbErr)
		}
	}

	upperBoundQP := c.QueryParam("upperBound")
	var (
		upperBound int
		ubErr      error
	)
	if len(upperBoundQP) > 0 {
		upperBound, ubErr = strconv.Atoi(upperBoundQP)
		if ubErr != nil {
			log.Fatal(ubErr)
		}
	}

	reference := c.QueryParam("reference")
	alternative := c.QueryParam("alternative")

	var alleles []string
	allelesQP := strings.Split(c.QueryParam("alleles"), ",")

	// reference, alternative and alleles can have the
	// single-wildcard character 'N', which adheres to
	// the spec found at : https://droog.gs.washington.edu/parc/images/iupac.html

	// swap all 'N's into '?'s for elasticsearch
	reference = strings.Replace(reference, "N", "?", -1)
	alternative = strings.Replace(alternative, "N", "?", -1)
	if len(allelesQP) > 0 && allelesQP[0] != "" { // check it isn't empty
		for _, a := range allelesQP {
			alleles = append(alleles, strings.Replace(a, "N", "?", -1))
		}
	}

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

	tableId := c.QueryParam("tableId")
	if len(tableId) == 0 {
		// if no tableId is provided, assume "wildcard" search
		tableId = "*"
	}

	return es, chromosome, lowerBound, upperBound, reference, alternative, alleles, genotype, assemblyId, tableId
}
