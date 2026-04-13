package main

import (
	"encoding/json"
	"fmt"
	"io"
	"math/rand" // provides rand.Float64() for the catch roll
	"net/http"

	"github.com/JaipreethTiruvaipati/pokedexcli/internal/pokecache"
)

// Pokemon holds the data we care about from:
// https://pokeapi.co/api/v2/pokemon/{name}
//
// Pokemon holds the data we care about from:
// https://pokeapi.co/api/v2/pokemon/{name}
//
// All these fields are populated automatically by json.Unmarshal
// when we catch a Pokemon — no extra API call needed for inspect.
type Pokemon struct {
	Name           string `json:"name"`
	BaseExperience int    `json:"base_experience"`
	Height         int    `json:"height"` // in decimetres (1 unit = 10 cm)
	Weight         int    `json:"weight"` // in hectograms (1 unit = 100 g)

	// Stats is a list of base stats like hp, attack, defense, etc.
	// Each stat has a base_stat value and a nested stat.name string.
	Stats []struct {
		BaseStat int `json:"base_stat"`
		Stat     struct {
			Name string `json:"name"`
		} `json:"stat"`
	} `json:"stats"`

	// Types is a list of types the Pokemon belongs to (e.g. normal, flying).
	// Each entry has a nested type.name string.
	Types []struct {
		Type struct {
			Name string `json:"name"`
		} `json:"type"`
	} `json:"types"`
}

// fetchPokemon retrieves a Pokemon's data from the API (or cache).
// Same cache-first pattern used in fetchLocationAreas and fetchLocationAreaDetail.
func fetchPokemon(url string, cache *pokecache.Cache) (Pokemon, error) {
	// ── Cache hit ─────────────────────────────────────────────────────────────
	if cached, ok := cache.Get(url); ok {
		fmt.Println("[cache] hit:", url)
		var p Pokemon
		if err := json.Unmarshal(cached, &p); err != nil {
			return Pokemon{}, fmt.Errorf("json parse failed (cached): %w", err)
		}
		return p, nil
	}
	// ── Cache miss — real HTTP request ────────────────────────────────────────
	fmt.Println("[cache] miss, fetching:", url)
	res, err := http.Get(url)
	if err != nil {
		return Pokemon{}, fmt.Errorf("request failed: %w", err)
	}
	defer res.Body.Close()
	if res.StatusCode >= 400 {
		return Pokemon{}, fmt.Errorf("bad status: %d", res.StatusCode)
	}
	body, err := io.ReadAll(res.Body)
	if err != nil {
		return Pokemon{}, fmt.Errorf("read body failed: %w", err)
	}
	// Store in cache before parsing so future catches of the same Pokemon are instant.
	cache.Add(url, body)
	var p Pokemon
	if err := json.Unmarshal(body, &p); err != nil {
		return Pokemon{}, fmt.Errorf("json parse failed: %w", err)
	}
	return p, nil
}

// commandCatch attempts to catch a named Pokemon.
//
// How the catch chance works:
//   catchChance = 50.0 / base_experience
//
// A random float between 0.0 and 1.0 is rolled.
// If the roll is LESS than catchChance → caught!
// If the roll is GREATER OR EQUAL → escaped.
//
// Examples:
//   Caterpie  (base_exp=39):  catchChance = 50/39  = 1.28 → always caught (>1.0 = guaranteed)
//   Pikachu   (base_exp=112): catchChance = 50/112 = 0.45 → ~45% chance
//   Dragonite (base_exp=270): catchChance = 50/270 = 0.19 → ~19% chance
//   Mewtwo    (base_exp=340): catchChance = 50/340 = 0.15 → ~15% chance

func commandCatch(cfg *config, args []string) error {
	// Guard: user must provide a Pokemon name.
	if len(args) == 0 {
		fmt.Println("Usage: catch <pokemon-name>")
		return nil
	}
	pokemonName := args[0]
	url := "https://pokeapi.co/api/v2/pokemon/" + pokemonName
	// Print the "throwing" message BEFORE we know the outcome.
	// This mimics the suspense of throwing a Pokeball.
	fmt.Println("Throwing a Pokeball at " + pokemonName + "...")

	pokemon, err := fetchPokemon(url, cfg.cache)
	if err != nil {
		return err
	}
	// Calculate catch chance from base experience.
	// We use 50.0 as the numerator — feel free to tune this up or down.
	catchChance := 50.0 / float64(pokemon.BaseExperience)
	// rand.Float64() returns a random float in [0.0, 1.0).
	// If our roll beats catchChance, the Pokemon escapes.
	roll := rand.Float64()
	if roll < catchChance {
		// Caught! Add to the user's pokedex.
		cfg.pokedex[pokemon.Name] = pokemon
		fmt.Println(pokemonName + " was caught!")
		fmt.Println("You may now inspect it with the inspect command.")
	} else {
		fmt.Println(pokemonName + " escaped!")
	}
	return nil
}
