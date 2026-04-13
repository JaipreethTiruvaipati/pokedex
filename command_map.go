package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/JaipreethTiruvaipati/pokedexcli/internal/pokecache"
)

// locationAreaResponse models one paginated response from:
// https://pokeapi.co/api/v2/location-area?offset=0&limit=20
type locationAreaResponse struct {
	// Count is the total number of location-area resources available in the API,
	// across ALL pages (not just this page).
	Count int `json:"count"`

	// Next is the URL for the next page of results.
	// It is a *string because the API can return null when there is no next page.
	// next != nil  => there is another page
	// next == nil  => current page is the last page
	Next *string `json:"next"`

	// Previous is the URL for the previous page of results.
	// It is a *string because the API can return null when there is no previous page.
	// previous != nil => user can go back
	// previous == nil => current page is the first page
	Previous *string `json:"previous"`

	// Results is the list of location areas returned for THIS page.
	// By default the API returns up to 20 items per page unless limit is changed.
	Results []struct {
		// Name is the human-readable resource key, e.g. "canalave-city-area".
		Name string `json:"name"`

		// URL is the endpoint for this specific location-area resource,
		// e.g. "https://pokeapi.co/api/v2/location-area/1/".
		URL string `json:"url"`
	} `json:"results"`
}

// fetchLocationAreas performs a GET request and unmarshals response JSON.
func fetchLocationAreas(url string, cache *pokecache.Cache) (locationAreaResponse, error) {

	// ── Step 1: Check cache first ──────────────────────────────────────────────
	if cached, ok := cache.Get(url); ok {
		// Cache HIT — we already have the bytes, no network needed
		fmt.Println("[cache] hit:", url)
		var parsed locationAreaResponse
		if err := json.Unmarshal(cached, &parsed); err != nil {
			return locationAreaResponse{}, fmt.Errorf("json parse failed (cached): %w", err)
		}
		return parsed, nil // return early, skip the HTTP call entirely
	}

	// ── Step 2: Cache MISS — do the real HTTP request ─────────────────────────
	fmt.Println("[cache] miss, fetching:", url)
	res, err := http.Get(url)
	if err != nil {
		return locationAreaResponse{}, fmt.Errorf("request failed: %w", err)
	}
	defer res.Body.Close()
	if res.StatusCode >= 400 {
		return locationAreaResponse{}, fmt.Errorf("bad status code: %d", res.StatusCode)
	}
	body, err := io.ReadAll(res.Body)
	if err != nil {
		return locationAreaResponse{}, fmt.Errorf("read body failed: %w", err)
	}

	// ── Step 3: Store raw bytes in cache BEFORE parsing ───────────────────────
	// We cache the raw bytes (not the parsed struct) so we can unmarshal them
	// the same way on a cache hit — no special-case logic needed.
	cache.Add(url, body)

	var parsed locationAreaResponse
	if err := json.Unmarshal(body, &parsed); err != nil {
		return locationAreaResponse{}, fmt.Errorf("json parse failed: %w", err)
	}
	return parsed, nil
}

// commandMap loads and prints the "next" page (20 location areas).
func commandMap(cfg *config, args []string) error {
	// If next is nil, there is no next page.
	if cfg.next == nil {
		fmt.Println("no more locations")
		return nil
	}
	resp, err := fetchLocationAreas(*cfg.next, cfg.cache)
	if err != nil {
		return err
	}
	// Update pagination state for future map/mapb calls.
	cfg.next = resp.Next
	cfg.previous = resp.Previous
	// Print each location area name on its own line.
	for _, area := range resp.Results {
		fmt.Println(area.Name)
	}
	return nil
}

// commandMapBack loads and prints the "previous" page (20 location areas).
func commandMapBack(cfg *config, args []string) error {
	// On very first page, previous is nil.
	if cfg.previous == nil {
		fmt.Println("you're on the first page")
		return nil
	}
	resp, err := fetchLocationAreas(*cfg.previous, cfg.cache)
	if err != nil {
		return err
	}
	// Update pagination state after moving backward.
	cfg.next = resp.Next
	cfg.previous = resp.Previous
	for _, area := range resp.Results {
		fmt.Println(area.Name)
	}
	return nil
}
