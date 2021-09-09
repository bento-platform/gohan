package zygosity

import (
	"api/models/constants"
)

const (
	Empty constants.Zygosity = iota - 1
	Unknown
	Homozygous
	Heterozygous
)

func IsValid(value int) bool {
	return value <= int(Heterozygous)
}

func IsValidQuery(value int) bool {
	return value > int(Empty) && IsValid(value)
}
