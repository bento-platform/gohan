package sort

import (
	"api/models/constants"
	"strings"
)

const (
	Undefined  constants.SortDirection = ""
	Ascending  constants.SortDirection = "asc"
	Descending constants.SortDirection = "desc"
)

func CastToSortDirection(text string) constants.SortDirection {
	switch strings.ToLower(text) {
	case "asc":
		return Ascending
	case "desc":
		return Descending
	default:
		return Undefined
	}
}
