package main

import (
	"fmt"
	"strings"
	"os"
	"bufio"
	"time"
	"encoding/json"
	"math/rand"
	"github.com/andrei-himself/pokedexcli/internal/pokeapi"
	"github.com/andrei-himself/pokedexcli/internal/pokecache"
)

type cliCommand struct {
	name string
	description string
	callback func(*config, *pokecache.Cache, []string) error
}

type config struct {
	Next *string
	Previous *string
	Pokedex map[string]pokeapi.Pokemon
}

var cliCommands map[string]cliCommand
var pokedex map[string]pokeapi.Pokemon

func cleanInput(text string) []string {
	splitted := strings.Fields(text)	
	cleaned := []string{}
	for _, w := range splitted {
		cleaned = append(cleaned, strings.ToLower(w))
	}
	return cleaned
}

func commandExit(c *config, cache *pokecache.Cache, input []string) error {
	fmt.Println("Closing the Pokedex... Goodbye!")
	os.Exit(0)
	return nil
}

func commandHelp(c *config, cache *pokecache.Cache, input []string) error {
	fmt.Println("Welcome to the Pokedex!")
	usage := "Usage:\n\n"
	for k, v := range cliCommands {
		usage += k + ": " + v.description + "\n"
	}
	fmt.Printf("%s", usage)
	return nil
}

func commandMap(c *config, cache *pokecache.Cache, input []string) error {
	body := []byte{}
	if v, ok := cache.Get(*c.Next); ok == true {
		fmt.Println("Using cache")
		body = v
	} else {
		v, err := pokeapi.Fetch(*c.Next)
		if err != nil {
			return err
		}
		body = v
		cache.Add(*c.Next, body)
	}

	locationArea := pokeapi.LocationArea{}
	err := json.Unmarshal(body, &locationArea)
	if err != nil {
		return err
	}
	c.Previous = locationArea.Previous
	c.Next = locationArea.Next
	len := len(locationArea.Results)
	for i := 0; i < len; i++ {
		fmt.Println(locationArea.Results[i].Name)
	}
	
	return nil
}

func commandMapb(c *config, cache *pokecache.Cache, input []string) error {
	if c.Previous == nil {
		fmt.Println("you're on the first page")
		return nil
	}

	body := []byte{}
	if v, ok := cache.Get(*c.Previous); ok {
		fmt.Println("Using cache")
		body = v
	} else {
		v, err := pokeapi.Fetch(*c.Previous)
		if err != nil {
			return err
		}
		body = v
		cache.Add(*c.Previous, body)
	}

	locationArea := pokeapi.LocationArea{}
	err := json.Unmarshal(body, &locationArea)
	if err != nil {
		return err
	}
	c.Previous = locationArea.Previous
	c.Next = locationArea.Next
	len := len(locationArea.Results)
	for i := 0; i < len; i++ {
		fmt.Println(locationArea.Results[i].Name)
	}

	return nil
}

func commandExplore(c *config, cache *pokecache.Cache, input []string) error {
	url := "https://pokeapi.co/api/v2/location-area/" + input[1]

	body := []byte{}
	if v, ok := cache.Get(url); ok {
		body = v
	} else {
		v, err := pokeapi.Fetch(url)
		if err != nil {
			return err
		}
		body = v
		cache.Add(url, body)
	}

	exploredLocation := pokeapi.ExploredLocation{}
	err := json.Unmarshal(body, &exploredLocation)
	if err != nil {
		return err
	}
	fmt.Printf("Exploring %s...\n", exploredLocation.Location.Name)
	len := len(exploredLocation.PokemonEncounters)
	if len == 0 {
		fmt.Println("No Pokémon found here")
	}
	for i := 0; i < len; i++ {
		fmt.Println("-", exploredLocation.PokemonEncounters[i].Pokemon.Name)
	}
	
	return nil
}

func commandCatch(c *config, cache *pokecache.Cache, input []string) error {
	url := "https://pokeapi.co/api/v2/pokemon/" + input[1]

	body := []byte{}
	if v, ok := cache.Get(url); ok {
		body = v
	} else {
		v, err := pokeapi.Fetch(url)
		if err != nil {
			return err
		}
		body = v
		cache.Add(url, body)
	}

	pokemon := pokeapi.Pokemon{}
	err := json.Unmarshal(body, &pokemon)
	if err != nil {
		return err
	}
	fmt.Printf("Throwing a Pokeball at %s...\n", pokemon.Name)
	chance := rand.Intn(pokemon.BaseExperience)
	if chance < pokemon.BaseExperience / 3 {
		fmt.Printf("%s escaped!\n", pokemon.Name)
		return nil
	}
	fmt.Printf("%s was caugth!\n", pokemon.Name)
	fmt.Println("You may now inspect it with the inspect command.")
	c.Pokedex[pokemon.Name] = pokemon

	return nil
}

func commandInspect(c *config, cache *pokecache.Cache, input []string) error {
	pokemon := pokeapi.Pokemon{}
	if v, ok := c.Pokedex[input[1]]; ok == false {
		fmt.Println("you have not caught that pokemon")
		return nil
	} else {
		pokemon = v
	}

	fmt.Println("Name:", pokemon.Name)
	fmt.Println("Height:", pokemon.Height)
	fmt.Println("Weight:", pokemon.Weight)
	fmt.Println("Stats:")
	for _, v := range pokemon.Stats {
		fmt.Printf("  -%s: %v\n", v.Stat.Name, v.BaseStat)
	}
	fmt.Println("Types:")
	for _, v := range pokemon.Types {
		fmt.Printf("  - %s\n", v.Type.Name)
	}

	return nil
}

func commandPokedex(c *config, cache *pokecache.Cache, input []string) error {
	fmt.Println("Your Pokedex:")
	for k, _ := range c.Pokedex {
		fmt.Println(" -", k)
	}

	return nil
}

func main() {
	cliCommands = map[string]cliCommand{
		"exit": {
			name: "exit",
			description: "Exit the Pokedex",
			callback: commandExit,
		},
		"help": {
			name: "help",
			description: "Displays a help message",
			callback: commandHelp,
		},
		"map": {
			name: "map",
			description: "Displays the next 20 location areas",
			callback: commandMap,
		},
		"mapb": {
			name: "mapb",
			description: "Displays the previous 20 location areas",
			callback: commandMapb,
		},
		"explore": {
			name: "explore",
			description: "Display a list of all the Pokémon in the given location",
			callback: commandExplore,
		},
		"catch": {
			name: "catch",
			description: "Try to catch the given Pokémon",
			callback: commandCatch,
		},
		"inspect": {
			name: "inspect",
			description: "Display information on the given Pokémon",
			callback: commandInspect,
		},
		"pokedex": {
			name: "pokedex",
			description: "Display the Pokémon names in your Pokedex",
			callback: commandPokedex,
		},
	}

	url := "https://pokeapi.co/api/v2/location-area/"
	config := config {
		Next: &url,
		Previous: nil,
		Pokedex: map[string]pokeapi.Pokemon{},
	}
	cache := pokecache.NewCache(10 * time.Second)

	scanner := bufio.NewScanner(os.Stdin)
	for ; ; {
		fmt.Printf("Pokedex > ")
		scanner.Scan()
		input := scanner.Text()
		cleaned := cleanInput(input)
		command := ""
		if len(cleaned) > 0 {
			command = cleaned[0]
		}
		if v, ok := cliCommands[command]; ok == false {
			fmt.Println("Unknown command")
		} else {
			if err := v.callback(&config, cache, cleaned); err != nil {
				fmt.Println(err)
			}
		}
	}
}