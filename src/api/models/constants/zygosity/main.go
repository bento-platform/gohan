package zygosity

import (
	"gohan/api/models/constants"
)

const (
	Unknown constants.Zygosity = iota
	// Diploid or higher
	Heterozygous
	HomozygousReference
	HomozygousAlternate

	// Haploid (deliberately below diploid for sequential id'ing purposes)
	Reference
	Alternate
)

func IsKnown(value int) bool {
	return value > int(Unknown) && value <= int(Alternate)
}

func ZygosityToString(zyg constants.Zygosity) string {
	switch zyg {
	// Haploid
	case Reference:
		return "REFERENCE"
	case Alternate:
		return "ALTERNATE"

	// Diploid or higher
	case Heterozygous:
		return "HETEROZYGOUS"
	case HomozygousReference:
		return "HOMOZYGOUS_REFERENCE"
	case HomozygousAlternate:
		return "HOMOZYGOUS_ALTERNATE"
	default:
		return "UNKNOWN"
	}
}

// TODO: StringOrIntToZygosity
