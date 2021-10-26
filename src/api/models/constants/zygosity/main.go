package zygosity

import (
	"api/models/constants"
)

const (
	Unknown constants.Zygosity = iota
	Homozygous
	Heterozygous
	Empty
)

func IsValid(value int) bool {
	return value <= int(Heterozygous)
}

func IsValidQuery(value int) bool {
	return value > int(Empty) && IsValid(value)
}
