package main

import "fmt"

// commandInspect prints detailed information about a caught Pokemon.
//
// Key design: we do NOT make an API call here.
// The Pokemon was already fetched and stored in cfg.pokedex when the user caught it.
// This is why we saved the full Pokemon struct (not just the name) in the map.
func commandInspect(cfg *config, args []string) error {
	// Guard: user must provide a Pokemon name.
	if len(args) == 0 {
		fmt.Println("Usage: inspect <pokemon-name>")
		return nil
	}

	pokemonName := args[0]

	// Look up the Pokemon in the pokedex map.
	// If it's not there, the user hasn't caught it yet.
	pokemon, caught := cfg.pokedex[pokemonName]
	if !caught {
		fmt.Println("you have not caught that pokemon")
		return nil
	}

	// Print all details — data comes from the stored struct, no API call needed.
	fmt.Println("Name:", pokemon.Name)
	fmt.Println("Height:", pokemon.Height)
	fmt.Println("Weight:", pokemon.Weight)

	fmt.Println("Stats:")
	for _, s := range pokemon.Stats {
		// Each stat prints as:  -hp: 40
		fmt.Printf("  -%s: %d\n", s.Stat.Name, s.BaseStat)
	}

	fmt.Println("Types:")
	for _, t := range pokemon.Types {
		// Each type prints as:  - normal
		fmt.Printf("  - %s\n", t.Type.Name)
	}

	return nil
}
