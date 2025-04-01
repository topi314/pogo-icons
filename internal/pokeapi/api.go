package pokeapi

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

func NewAPI(endpoint string) Client {
	return &clientAPI{
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
		endpoint: endpoint,
	}
}

type clientAPI struct {
	client       *http.Client
	endpoint     string
	pokemonForms []PokemonForm
}

func (c *clientAPI) GetPokemon(ctx context.Context) ([]PokemonForm, error) {
	if c.pokemonForms != nil {
		return c.pokemonForms, nil
	}

	url := c.endpoint + "/pokemon"

	var allPokemonForms []PokemonForm
	for {
		p, err := c.getPokemonPage(ctx, url)
		if err != nil {
			return nil, err
		}

		for _, result := range p.Results {
			pokemonForm := newPokemonForm(result)
			allPokemonForms = append(allPokemonForms, pokemonForm)
		}

		if p.Next == "" {
			break
		}

		url = p.Next
	}

	c.pokemonForms = allPokemonForms
	return allPokemonForms, nil
}

func (c *clientAPI) getPokemonPage(ctx context.Context, url string) (Page[Pokemon], error) {
	rq, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return Page[Pokemon]{}, fmt.Errorf("error creating request: %w", err)
	}
	rs, err := c.client.Do(rq)
	if err != nil {
		return Page[Pokemon]{}, fmt.Errorf("error executing request: %w", err)
	}
	defer rs.Body.Close()

	var p Page[Pokemon]
	if err = json.NewDecoder(rs.Body).Decode(&p); err != nil {
		return Page[Pokemon]{}, fmt.Errorf("error decoding response: %w", err)
	}

	return p, nil
}

func (c *clientAPI) GetPokemonForm(ctx context.Context, name string) (PokemonForm, error) {
	rq, err := http.NewRequestWithContext(ctx, http.MethodGet, c.endpoint+"/pokemon/"+name, nil)
	if err != nil {
		return PokemonForm{}, fmt.Errorf("error creating request: %w", err)
	}

	rs, err := c.client.Do(rq)
	if err != nil {
		return PokemonForm{}, fmt.Errorf("error executing request: %w", err)
	}
	defer rs.Body.Close()

	if rs.StatusCode == http.StatusNotFound {
		return PokemonForm{}, fmt.Errorf("failed to find pokemon: %w", ErrNotFound)
	}

	if rs.StatusCode != http.StatusOK {
		return PokemonForm{}, fmt.Errorf("error fetching pokemon: %w", fmt.Errorf("unexpected status code: %d", rs.StatusCode))
	}

	var p Pokemon
	if err = json.NewDecoder(rs.Body).Decode(&p); err != nil {
		return PokemonForm{}, fmt.Errorf("error decoding response: %w", err)
	}

	return newPokemonForm(p), nil
}

func (c *clientAPI) GetSprite(ctx context.Context, url string) (*http.Response, error) {
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
