package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"log/slog"
	"os"
	"os/signal"

	"github.com/topi314/pogo-icons/internal/poke-icon"
	"github.com/topi314/pogo-icons/internal/pokeapi"
)

const (
	ScaleWidth  = -1
	ScaleHeight = 450
)

func main() {
	pokemon := flag.String("pokemon", "", "Pokemon name or ID")
	background := flag.String("background", "", "Background image name")
	endpoint := flag.String("endpoint", "https://pokeapi.co/api/v2", "PokeAPI endpoint")
	output := flag.String("output", "output.png", "Output file name")
	flag.Parse()

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, os.Kill)
	defer cancel()

	if *pokemon == "" {
		slog.ErrorContext(ctx, "Pokemon name or ID is required")
		return
	}

	if *background == "" {
		slog.ErrorContext(ctx, "Background image name is required")
		return
	}

	pokeClient := pokeapi.New(*endpoint)
	p, err := pokeClient.GetPokemon(ctx, *pokemon)
	if err != nil {
		if errors.Is(err, pokeapi.ErrNotFound) {
			slog.ErrorContext(ctx, "Pokemon not found", slog.String("pokemon", *pokemon))
			return
		}
		slog.ErrorContext(ctx, "Error while getting Pokemon", slog.Any("err", err))
		return
	}

	pokemonImage, err := pokeClient.GetSprite(ctx, p.Sprites.Other.OfficialArtwork.FrontDefault)
	if err != nil {
		slog.Error("Error while getting Pokemon image", slog.Any("err", err))
		return
	}
	defer pokemonImage.Body.Close()

	backgroundImage, err := os.Open(fmt.Sprintf("assets/backgrounds/%s.png", *background))
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			slog.ErrorContext(ctx, "Background image not found", slog.String("background", *background))
			return
		}
		slog.ErrorContext(ctx, "Error while opening background", slog.Any("err", err))
		return
	}
	defer backgroundImage.Close()

	r, err := pokeicon.Generate(ctx, pokemonImage.Body, backgroundImage, pokeicon.Options{
		ScaleWidth:  ScaleWidth,
		ScaleHeight: ScaleHeight,
	})
	if err != nil {
		slog.ErrorContext(ctx, "Error while generating image", slog.Any("err", err))
		return
	}

	outputFile, err := os.OpenFile("output.png", os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0666)
	if err != nil {
		slog.ErrorContext(ctx, "error opening output file", slog.String("err", err.Error()))
	}
	defer outputFile.Close()

	if _, err = io.Copy(outputFile, r); err != nil {
		slog.ErrorContext(ctx, "error copying to output file", slog.Any("err", err))
		return
	}

	slog.InfoContext(ctx, "Generated image", slog.String("output", *output))
}
