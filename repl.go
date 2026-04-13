package main

import (
	"strings"
)

func cleanInput(text string) []string {
	// First, lowercase the entire string
	lowered := strings.ToLower(text)

	// Second, split the string into a slice of words by whitespace
	// strings.Fields handles multiple spaces and trims for us!
	words := strings.Fields(lowered)

	return words
}
