package main

import (
	"context"
	"flag"
	"io"
	"io/fs"
	"log/slog"
	"os"
	"os/signal"
	"strings"

	"github.com/BurntSushi/toml"

	"github.com/topi314/pogo-icons/internal/pogoicon"
	"github.com/topi314/pogo-icons/internal/pokeapi"
)

func main() {
	pokemon := flag.String("pokemon", "", "A list of Pokemon names or IDs (comma separated)")
	event := flag.String("event", "", "Event name")
	cosmetics := flag.String("cosmetics", "", "A list od cosmetics names (comma separated)")
	endpoint := flag.String("endpoint", "https://pokeapi.co/api/v2", "PokeAPI endpoint URL (default: https://pokeapi.co/api/v2)")
	assets := flag.String("assets", "assets", "Assets directory (default: assets)")
	output := flag.String("output", "output.png", "Output file name (default: output.png)")
	flag.Parse()

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, os.Kill)
	defer cancel()

	if *pokemon == "" {
		slog.ErrorContext(ctx, "Pokemon name or ID is required")
		return
	}

	if *event == "" {
		slog.ErrorContext(ctx, "Event name is required")
		return
	}

	pokeClient := pokeapi.NewAPI(*endpoint)
	assetsDir := os.DirFS(*assets)

	assetConfig, err := fs.ReadFile(assetsDir, "config.toml")
	if err != nil {
		slog.ErrorContext(ctx, "Error while reading asset config", slog.Any("err", err))
		return
	}
	var cfg pogoicon.Config
	if err = toml.Unmarshal(assetConfig, &cfg); err != nil {
		slog.ErrorContext(ctx, "Error while unmarshalling events", slog.Any("err", err))
		return
	}

	var getPokemonImage = func(p string) (io.ReadCloser, error) {
		pf, err := pokeClient.GetPokemonForm(ctx, p)
		if err != nil {
			return nil, err
		}

		pokemonImage, err := pokeClient.GetSprite(ctx, pf.Sprite)
		if err != nil {
			return nil, err
		}

		return pokemonImage.Body, nil
	}

	pokemonList := strings.Split(*pokemon, ",")
	cosmeticList := strings.Split(*cosmetics, ",")

	r, err := pogoicon.Generate(assetsDir, cfg, getPokemonImage, *event, pokemonList, cosmeticList)
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
