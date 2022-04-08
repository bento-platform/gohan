package utils

import (
	"math/rand"
	"strings"
	"time"

	"github.com/google/uuid"
)

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

func RandomString(n int) string {
	rand.Seed(time.Now().UnixNano())

	var letters = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789")

	s := make([]rune, n)
	for i := range s {
		s[i] = letters[rand.Intn(len(letters))]
	}
	return string(s)
}

func IsValidUUID(u string) bool {
	_, err := uuid.Parse(u)
	return err == nil
}
func KeyExists(decoded map[string]interface{}, key string) bool {
	val, ok := decoded[key]
	return ok && val != nil
}
