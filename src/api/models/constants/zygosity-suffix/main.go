// TODO: Rename?
package ZygositySuffix

import (
	"api/models/constants"
)

const (
	Unknown constants.ZygositySuffix = iota
	Reference
	Alternate
	Empty
)

func IsValidZygositySuffix(value int) bool {
	return value >= int(Unknown) && value <= int(Alternate)
}

func IsRelevantZygositySuffix(value int) bool {
	return IsValidZygositySuffix(value) && value >= int(Empty)
}
