package zygosity

import (
	"api/models/constants"
)

const (
	Unknown constants.Zygosity = iota
	Heterozygous
	HomozygousReference
	HomozygousAlternate
)

func IsKnown(value int) bool {
	return value > int(Unknown) && value <= int(HomozygousAlternate)
}

func ZygosityToString(zyg constants.Zygosity) string {
	switch zyg {
	case Heterozygous:
		return "HETEROZYGOUS"
	case HomozygousAlternate:
		return "HOMOZYGOUS_ALTERNATE"
	case HomozygousReference:
		return "HOMOZYGOUS_REFERENCE"
	default:
		return "UNKNOWN"
	}
}

// TODO: StringOrIntToZygosity
