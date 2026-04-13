package main

import "fmt"

// commandPokedex prints the names of all Pokemon the user has caught.
// It reads directly from cfg.pokedex — no API call needed.
// The pokedex map was populated by commandCatch each time a Pokemon was caught.
func commandPokedex(cfg *config, args []string) error {
	// If the user hasn't caught anything yet, the map is empty.
	if len(cfg.pokedex) == 0 {
		fmt.Println("You haven't caught any Pokemon yet!")
		return nil
	}

	fmt.Println("Your Pokedex:")
	// Range over the map and print each caught Pokemon name.
	// Note: map iteration order in Go is random, so the list order may vary.
	for name := range cfg.pokedex {
		fmt.Println(" -", name)
	}
	return nil
}
