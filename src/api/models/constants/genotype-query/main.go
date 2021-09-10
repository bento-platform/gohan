package genotypeQuery

import (
	"api/models/constants"
	"errors"
	"strings"
)

const (
	UNCALLED constants.GenotypeQuery = ""

	// # Haploid
	REFERENCE constants.GenotypeQuery = "REFERENCE"
	ALTERNATE constants.GenotypeQuery = "ALTERNATE"

	// # Diploid or higher
	HOMOZYGOUS_REFERENCE constants.GenotypeQuery = "HOMOZYGOUS_REFERENCE"
	HETEROZYGOUS         constants.GenotypeQuery = "HETEROZYGOUS"
	HOMOZYGOUS_ALTERNATE constants.GenotypeQuery = "HOMOZYGOUS_ALTERNATE"
)

func CastToGenoType(text string) (constants.GenotypeQuery, error) {
	switch strings.ToLower(text) {
	case "":
		return UNCALLED, nil
	case "reference":
		return REFERENCE, nil
	case "alternate":
		return ALTERNATE, nil
	case "homozygous_reference":
		return HOMOZYGOUS_REFERENCE, nil
	case "heterozygous":
		return HETEROZYGOUS, nil
	case "homozygous_alternate":
		return HOMOZYGOUS_ALTERNATE, nil
	default:
		return UNCALLED, errors.New("unable to parse genotype query")
	}
}
