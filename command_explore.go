package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/JaipreethTiruvaipati/pokedexcli/internal/pokecache"
)

// locationAreaDetail models the response from:
// https://pokeapi.co/api/v2/location-area/{name}
// We only need pokemon_encounters — everything else in the response is ignored.
type locationAreaDetail struct {
	PokemonEncounters []struct {
		Pokemon struct {
			Name string `json:"name"`
		} `json:"pokemon"`
	} `json:"pokemon_encounters"`
}

// fetchLocationAreaDetail fetches the detail page for one location area.
// It uses the cache exactly the same way as fetchLocationAreas:
// check cache first → return cached bytes if hit → fetch from network on miss → store in cache.
func fetchLocationAreaDetail(url string, cache *pokecache.Cache) (locationAreaDetail, error) {
	// ── Cache hit ─────────────────────────────────────────────────────────────
	if cached, ok := cache.Get(url); ok {
		fmt.Println("[cache] hit:", url)
		var parsed locationAreaDetail
		if err := json.Unmarshal(cached, &parsed); err != nil {
			return locationAreaDetail{}, fmt.Errorf("json parse failed (cached): %w", err)
		}
		return parsed, nil
	}

	// ── Cache miss — real HTTP request ────────────────────────────────────────
	fmt.Println("[cache] miss, fetching:", url)
	res, err := http.Get(url)
	if err != nil {
		return locationAreaDetail{}, fmt.Errorf("request failed: %w", err)
	}
	defer res.Body.Close()

	if res.StatusCode >= 400 {
		return locationAreaDetail{}, fmt.Errorf("bad status: %d", res.StatusCode)
	}

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return locationAreaDetail{}, fmt.Errorf("read body failed: %w", err)
	}

	// Store raw bytes before parsing so future calls hit the cache.
	cache.Add(url, body)

	var parsed locationAreaDetail
	if err := json.Unmarshal(body, &parsed); err != nil {
		return locationAreaDetail{}, fmt.Errorf("json parse failed: %w", err)
	}
	return parsed, nil
}

// commandExplore lists all Pokémon found in the named location area.
// Usage: explore <location-area-name>
// Example: explore pastoria-city-area
func commandExplore(cfg *config, args []string) error {
	// Guard: user must provide an area name.
	if len(args) == 0 {
		fmt.Println("Usage: explore <location-area-name>")
		return nil
	}

	areaName := args[0]
	url := "https://pokeapi.co/api/v2/location-area/" + areaName

	fmt.Println("Exploring " + areaName + "...")

	detail, err := fetchLocationAreaDetail(url, cfg.cache)
	if err != nil {
		return err
	}

	fmt.Println("Found Pokemon:")
	for _, encounter := range detail.PokemonEncounters {
		fmt.Println(" -", encounter.Pokemon.Name)
	}

	return nil
}
