package chromosome

import (
	"fmt"
	"strconv"
	"strings"
)

func ValidListOfHumanChromosomes() []string {
	var humChroms []string
	for i := 1; i < 24; i++ {
		humChroms = append(humChroms, fmt.Sprint(i))
	}
	humChroms = append(humChroms, "X")
	humChroms = append(humChroms, "Y")
	humChroms = append(humChroms, "M")
	return humChroms
}

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
