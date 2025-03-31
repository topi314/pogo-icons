package pokeapi

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"
)

var ErrNotFound = errors.New("not found")

func New(endpoint string) *Client {
	return &Client{
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
		endpoint: endpoint,
	}
}

type Client struct {
	client         *http.Client
	endpoint       string
	pokemonSpecies []PokemonSpecie
}

func (c *Client) GetPokemonSpecies(ctx context.Context) ([]PokemonSpecie, error) {
	if c.pokemonSpecies != nil {
		return c.pokemonSpecies, nil
	}

	url := c.endpoint + "/pokemon-species"

	var allPokemonSpecies []PokemonSpecie
	for {
		p, err := c.getPokemonSpeciesPage(ctx, url)
		if err != nil {
			return nil, err
		}

		allPokemonSpecies = append(allPokemonSpecies, p.Results...)
		if p.Next == "" {
			break
		}

		url = p.Next
	}

	c.pokemonSpecies = allPokemonSpecies

	return allPokemonSpecies, nil
}

func (c *Client) getPokemonSpeciesPage(ctx context.Context, url string) (Page[PokemonSpecie], error) {
	rq, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return Page[PokemonSpecie]{}, fmt.Errorf("error creating request: %w", err)
	}
	rs, err := c.client.Do(rq)
	if err != nil {
		return Page[PokemonSpecie]{}, fmt.Errorf("error executing request: %w", err)
	}
	defer rs.Body.Close()

	var p Page[PokemonSpecie]
	if err = json.NewDecoder(rs.Body).Decode(&p); err != nil {
		return Page[PokemonSpecie]{}, fmt.Errorf("error decoding response: %w", err)
	}

	return p, nil
}

// GetPokemon fetches a single Pokémon by id or name.
func (c *Client) GetPokemon(ctx context.Context, name string) (Pokemon, error) {
	rq, err := http.NewRequestWithContext(ctx, http.MethodGet, c.endpoint+"/pokemon/"+name, nil)
	if err != nil {
		return Pokemon{}, fmt.Errorf("error creating request: %w", err)
	}

	rs, err := c.client.Do(rq)
	if err != nil {
		return Pokemon{}, fmt.Errorf("error executing request: %w", err)
	}
	defer rs.Body.Close()

	if rs.StatusCode == http.StatusNotFound {
		return Pokemon{}, fmt.Errorf("failed to find pokemon: %w", ErrNotFound)
	}

	if rs.StatusCode != http.StatusOK {
		return Pokemon{}, fmt.Errorf("error fetching pokemon: %w", fmt.Errorf("unexpected status code: %d", rs.StatusCode))
	}

	var p Pokemon
	if err = json.NewDecoder(rs.Body).Decode(&p); err != nil {
		return Pokemon{}, fmt.Errorf("error decoding response: %w", err)
	}

	return p, nil
}

func (c *Client) GetSprite(ctx context.Context, url string) (*http.Response, error) {
	rq, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("error creating sprite request: %w", err)
	}

	rs, err := c.client.Do(rq)
	if err != nil {
		return nil, fmt.Errorf("error executing sprite request: %w", err)
	}

	return rs, nil
}
