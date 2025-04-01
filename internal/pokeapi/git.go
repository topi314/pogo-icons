package pokeapi

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"path"
	"strings"
	"time"

	"github.com/go-git/go-billy/v5"
	"github.com/go-git/go-billy/v5/memfs"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/storage/memory"
)

func NewGit(repository string) (Client, error) {
	slog.Info("Creating new git client")
	fs := memfs.New()
	r, err := git.Clone(memory.NewStorage(), fs, &git.CloneOptions{
		URL:   repository,
		Auth:  nil,
		Depth: 1,
	})
	if err != nil {
		return nil, fmt.Errorf("error cloning repository: %w", err)
	}
	slog.Info("Cloned repository")

	c := &clientGit{
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
		fs:   fs,
		repo: r,
	}

	slog.Info("Loading data")
	if err = c.load(); err != nil {
		return nil, fmt.Errorf("error loading data: %w", err)
	}
	slog.Info("Data loaded")

	return c, nil
}

type clientGit struct {
	client  *http.Client
	fs      billy.Filesystem
	repo    *git.Repository
	pokemon []PokemonForm
}

func (c *clientGit) load() error {
	species, err := c.fs.ReadDir("data/api/v2/pokemon-species")
	if err != nil {
		return fmt.Errorf("error reading directory: %w", err)
	}

	var pokemon []PokemonForm
	for _, specie := range species {
		if !specie.IsDir() {
			continue
		}

		pokemonSpecie, err := c.parseSpecie(specie)
		if err != nil {
			return fmt.Errorf("error parsing specie: %w", err)
		}
		pokemon = append(pokemon, pokemonSpecie...)
	}
	c.pokemon = pokemon
	return nil
}

func (c *clientGit) parseSpecie(specie os.FileInfo) ([]PokemonForm, error) {
	filename := path.Join("data/api/v2/pokemon-species", specie.Name(), "index.json")
	file, err := c.fs.Open(filename)
	if err != nil {
		return nil, fmt.Errorf("error opening file %q: %w", filename, err)
	}
	defer file.Close()

	var p PokemonSpecie
	if err = json.NewDecoder(file).Decode(&p); err != nil {
		return nil, fmt.Errorf("error decoding specie: %w", err)
	}

	var forms []PokemonForm
	for _, variety := range p.Varieties {
		form, err := c.parsePokemon(variety.Pokemon.URL)
		if err != nil {
			return nil, fmt.Errorf("error parsing pokemon: %w", err)
		}
		forms = append(forms, form)
	}

	return forms, nil
}

func (c *clientGit) parsePokemon(url string) (PokemonForm, error) {
	file, err := c.fs.Open(path.Join("data", url, "index.json"))
	if err != nil {
		return PokemonForm{}, fmt.Errorf("error opening file: %w", err)
	}
	defer file.Close()

	var p Pokemon
	if err = json.NewDecoder(file).Decode(&p); err != nil {
		return PokemonForm{}, fmt.Errorf("error decoding pokemon: %w", err)
	}

	return newPokemonForm(p), nil
}

func (c *clientGit) GetPokemon(ctx context.Context) ([]PokemonForm, error) {
	return c.pokemon, nil
}

func (c *clientGit) GetPokemonForm(ctx context.Context, name string) (PokemonForm, error) {
	name = strings.ToLower(name)
	for _, p := range c.pokemon {
		if strings.ToLower(p.Value) == name || strings.ToLower(p.Name) == name {
			return p, nil
		}
	}
	return PokemonForm{}, fmt.Errorf("pokemon not found: %w", ErrNotFound)
}

func (c *clientGit) GetSprite(ctx context.Context, url string) (*http.Response, error) {
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
