package utils

import (
	c "api/models/constants"
	z "api/models/constants/zygosity"
)

func StringInSlice(a string, list []string) bool {
	for _, b := range list {
		if b == a {
			return true
		}
	}
	return false
}

// TODO: StringOrIntToZygosity
func ZygosityToString(zyg c.Zygosity) string {
	switch zyg {
	case z.Heterozygous:
		return "HETEROZYGOUS"
	case z.HomozygousAlternate:
		return "HOMOZYGOUS_ALTERNATE"
	case z.HomozygousReference:
		return "HOMOZYGOUS_REFERENCE"
	default:
		return "UNKNOWN"
	}
}
