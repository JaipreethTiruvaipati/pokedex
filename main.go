package main

import (
	"bufio"
	"fmt"
	"os"
	"sort"
	"time" // needed for 5 * time.Second

	"github.com/JaipreethTiruvaipati/pokedexcli/internal/pokecache"
)

// config holds shared REPL state.
// next and previous are pagination URLs for the location-area endpoint.
type config struct {
	next     *string
	previous *string
	cache    *pokecache.Cache
	pokedex  map[string]Pokemon // stores caught Pokemon; key = pokemon name
}

// cliCommand describes one command that the REPL understands.
//
// name:        command text typed by the user (example: "help")
// description: short explanation shown in the help output
// callback takes a pointer to config so commands can read/update shared state.
type cliCommand struct {
	name        string
	description string
	callback    func(*config, []string) error
}

// strPtr is a tiny helper to get a *string from a string literal/value.

func strPtr(s string) *string {
	return &s
}

// getCommands is our command registry.
//
// Why this exists:
// - Central place to add/remove commands
// - REPL can look up a command name without hardcoding many if/else blocks
// Add all supported commands here.
func getCommands() map[string]cliCommand {
	return map[string]cliCommand{
		"help": {
			name:        "help",
			description: "Displays a help message",
			callback:    commandHelp,
		},
		"exit": {
			name:        "exit",
			description: "Exit the Pokedex",
			callback:    commandExit,
		},
		"map": {
			name:        "map",
			description: "Displays the next 20 location areas",
			callback:    commandMap,
		},
		"mapb": {
			name:        "mapb",
			description: "Displays the previous 20 location areas",
			callback:    commandMapBack,
		},
		"explore": {
			name:        "explore",
			description: "Explore a location area to see its Pokemon",
			callback:    commandExplore,
		},
		"catch": {
			name:        "catch",
			description: "Attempt to catch a Pokemon",
			callback:    commandCatch,
		},
		"inspect": {
			name:        "inspect",
			description: "Inspect a caught Pokemon's details",
			callback:    commandInspect,
		},
		"pokedex": {
			name:        "pokedex",
			description: "List all caught Pokemon",
			callback:    commandPokedex,
		},
	}
}

// commandHelp prints the greeting and usage text.
//
// It dynamically generates the usage lines from the command registry.
// That means whenever you register a new command, help updates automatically.
func commandHelp(cfg *config, args []string) error {
	_ = cfg // not used now, but kept for consistent callback signature
	fmt.Println("Welcome to the Pokedex!")
	fmt.Println("Usage:")
	fmt.Println()

	commands := getCommands()

	// Map iteration order in Go is random.
	// Sort keys so help output is stable and predictable.
	keys := make([]string, 0, len(commands))
	for key := range commands {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	for _, key := range keys {
		cmd := commands[key]
		fmt.Printf("%s: %s\n", cmd.name, cmd.description)
	}

	return nil
}

// commandExit prints a goodbye message, then terminates the process immediately.
//
// os.Exit(0) means "clean, successful exit" (exit status code 0).
func commandExit(cfg *config, args []string) error {
	_ = cfg // not used now, but kept for consistent callback signature
	fmt.Println("Closing the Pokedex... Goodbye!")
	os.Exit(0)

	// This line is unreachable because os.Exit stops the program immediately.
	// It is included only to satisfy the func() error signature.
	return nil
}

func main() {

	// Start config on the first page of location areas.

	cfg := &config{
		next:     strPtr("https://pokeapi.co/api/v2/location-area?offset=0&limit=20"),
		previous: nil,
		cache:    pokecache.NewCache(5 * time.Second), // 5s TTL — tune this as you like
		pokedex:  make(map[string]Pokemon),            // make() initializes the map so writes don't panic
	}
	// Scanner reads user input line-by-line from stdin.
	scanner := bufio.NewScanner(os.Stdin)

	// Infinite REPL loop: prompt -> read -> parse -> dispatch command.
	for {
		// Show prompt without newline, so cursor stays on same line.
		fmt.Print("Pokedex > ")

		// Read one line from user.
		// (In a production app, you'd also check scanner.Err().)
		scanner.Scan()
		input := scanner.Text()

		// Normalize and split input:
		// - lowercases everything
		// - splits by whitespace
		// Example: "  HeLP  now " -> ["help", "now"]
		words := cleanInput(input)

		// If user hit Enter on an empty line, reprompt.
		if len(words) == 0 {
			continue
		}

		// The first word is treated as the command name.
		commandName := words[0]

		// Fetch command registry and try to find this command.
		commands := getCommands()
		cmd, exists := commands[commandName]

		// If command is unknown, notify user and continue REPL loop.
		if !exists {
			fmt.Println("Unknown command")
			continue
		}

		// Run command callback.
		// If callback returns an error, print it.
		if err := cmd.callback(cfg, words[1:]); err != nil {
			fmt.Println(err)
		}
	}
}
