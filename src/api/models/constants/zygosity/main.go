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
