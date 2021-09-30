package chromosome

import (
	"strconv"
	"strings"
)

func IsValidHumanChromosome(text string) bool {

	// Check if number can be represented as an int as is non-zero
	chromNumber, _ := strconv.Atoi(text)
	if chromNumber > 0 {
		// It can..
		// Check if it in range 1-23
		if chromNumber < 24 {
			return true
		}
	} else {
		// No it can't..
		// Check if it is an X, Y..
		loweredText := strings.ToLower(text)
		switch loweredText {
		case "x":
			return true
		case "y":
			return true
		}

		// ..or M (MT)
		switch strings.Contains(loweredText, "m") {
		case true:
			return true
		}
	}

	return false
}
