package zygosity

import (
	"api/models/constants"
)

const (
	Unknown constants.Zygosity = iota
	Homozygous
	Heterozygous
)
