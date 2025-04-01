package pokeapi

import (
	"context"
	"errors"
	"net/http"
)

var ErrNotFound = errors.New("not found")

type Client interface {
	GetPokemon(ctx context.Context) ([]PokemonForm, error)
	GetPokemonForm(ctx context.Context, name string) (PokemonForm, error)
	GetSprite(ctx context.Context, url string) (*http.Response, error)
}
