package ploidy

import (
	"gohan/api/models/constants"
)

const (
	Unknown constants.Ploidy = iota

	Haploid
	Diploid
)

func IsKnown(value int) bool {
	return value > int(Unknown) && value <= int(Diploid)
}
