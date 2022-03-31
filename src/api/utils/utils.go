package utils

import "strings"

func StringInSlice(a string, list []string) bool {
	for _, b := range list {
		if b == a {
			return true
		}
	}
	return false
}

func GetLeadingStringInBetweenSquareBrackets(str string) (bracketString string, theRestString string) {
	var (
		start = "["
		end   = "]"
	)
	s := strings.Index(str, start)
	if s == -1 {
		return
	}

	// Assume that if the open bracket is not at index 0,
	// it's an open bracket for an array of some sort within the string rather
	// than a marker for a prepended status code (i.e. elasticsearch)
	if s != 0 {
		return
	}

	e := strings.Index(str[s:], end)
	if e == -1 {
		return
	}

	return strings.Trim(str[s:e+1], " "), strings.Trim(str[e+1:], " ")
}
