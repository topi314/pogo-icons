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

	"github.com/topi314/pogo-icons/internal/pogoicon"
	"github.com/topi314/pogo-icons/internal/pokeapi"
)

func main() {
	background := flag.String("background", "generic", "Background image name (default: generic)")
	backgroundIcon := flag.String("background-icon", "", "Background icon name")
	backgroundIconScale := flag.Float64("background-icon-scale", 1.0, "Background icon scale (default: 1.0)")
	pokemon := flag.String("pokemon", "", "Pokemon name or ID")
	pokemonScale := flag.Float64("scale", 1, "Pokemon scale (default: 1.0)")
	cosmetic := flag.String("cosmetic", "", "Cosmetic image name")
	cosmeticScale := flag.Float64("cosmetic-scale", 0.15, "Cosmetic scale (default: 0.2)")
	endpoint := flag.String("endpoint", "https://pokeapi.co/api/v2", "PokeAPI endpoint URL (default: https://pokeapi.co/api/v2)")
	output := flag.String("output", "output.png", "Output file name (default: output.png)")
	flag.Parse()

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, os.Kill)
	defer cancel()

	if *pokemon == "" {
		slog.ErrorContext(ctx, "Pokemon name or ID is required")
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

	var backgroundIconImage io.Reader
	if *backgroundIcon != "" {
		img, err := os.Open(fmt.Sprintf("assets/background_icons/%s.png", *backgroundIcon))
		if err != nil {
			if errors.Is(err, os.ErrNotExist) {
				slog.ErrorContext(ctx, "Background icon image not found", slog.String("background-icon", *backgroundIcon))
				return
			}
			slog.ErrorContext(ctx, "Error while opening background icon", slog.Any("err", err))
			return
		}
		defer img.Close()
		backgroundIconImage = img
	}

	var cosmeticImage io.Reader
	if *cosmetic != "" {
		img, err := os.Open(fmt.Sprintf("assets/cosmetics/%s.png", *cosmetic))
		if err != nil {
			if errors.Is(err, os.ErrNotExist) {
				slog.ErrorContext(ctx, "Cosmetic image not found", slog.String("cosmetic", *cosmetic))
				return
			}
			slog.ErrorContext(ctx, "Error while opening cosmetic", slog.Any("err", err))
			return
		}
		defer img.Close()
		cosmeticImage = img
	}

	r, err := pogoicon.Generate(backgroundImage, backgroundIconImage, *backgroundIconScale, pokemonImage.Body, *pokemonScale, cosmeticImage, *cosmeticScale)
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
