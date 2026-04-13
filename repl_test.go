package main

import "testing"

func TestCleanInput(t *testing.T) {
	cases := []struct {
		input    string
		expected []string
	}{
		{
			input:    "  hello  world  ",
			expected: []string{"hello", "world"},
		},
		{
			input:    "Charmander Bulbasaur PIKACHU",
			expected: []string{"charmander", "bulbasaur", "pikachu"},
		},
	}

	for _, c := range cases {
		actual := cleanInput(c.input)

		// Check the length of the actual slice
		if len(actual) != len(c.expected) {
			t.Errorf("Lengths don't match: len(%v) vs len(%v)", len(actual), len(c.expected))
			// Skip to next loop iteration so we don't crash checking out-of-bounds indexes
			continue
		}

		// Check each word in the slice matches
		for i := range actual {
			word := actual[i]
			expectedWord := c.expected[i]

			if word != expectedWord {
				t.Errorf("Values don't match: expected %v, but got %v", expectedWord, word)
			}
		}
	}
}
