package assemblyId

import (
	"gohan/api/models/constants"
	"strings"
)

const (
	Unknown constants.AssemblyId = "Unknown"

	GRCh38 constants.AssemblyId = "GRCh38"
	GRCh37 constants.AssemblyId = "GRCh37"
	NCBI36 constants.AssemblyId = "NCBI36"
	NCBI35 constants.AssemblyId = "NCBI35"
	NCBI34 constants.AssemblyId = "NCBI34"
	Other  constants.AssemblyId = "Other"
)

func CastToAssemblyId(text string) constants.AssemblyId {
	switch strings.ToLower(text) {
	case "grch38":
		return GRCh38
	case "grch37":
		return GRCh37
	case "ncbi36":
		return NCBI36
	case "ncbi35":
		return NCBI35
	case "ncbi34":
		return NCBI34
	case "other":
		return Other
	default:
		return Unknown
	}
}

func IsKnownAssemblyId(text string) bool {
	// attempt to cast to assemblyId and
	// return if unknown assemblyId
	return CastToAssemblyId(text) != Unknown
}
