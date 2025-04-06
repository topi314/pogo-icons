package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"io/fs"
	"log/slog"
	"os"
	"os/signal"
	"path"
	"strings"

	"github.com/BurntSushi/toml"

	"github.com/topi314/pogo-icons/internal/pogoicon"
	"github.com/topi314/pogo-icons/internal/pokeapi"
	"github.com/topi314/pogo-icons/pogoicons"
)

func main() {
	pokemon := flag.String("pokemon", "", "A list of Pokemon names or IDs (comma separated)")
	event := flag.String("event", "", "Event name")
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

	assetConfig, err := fs.ReadFile(assetsDir, "assets/config.toml")
	if err != nil {
		slog.ErrorContext(ctx, "Error while reading asset config", slog.Any("err", err))
		return
	}
	var assetCfg pogoicons.AssetConfig
	if err = toml.Unmarshal(assetConfig, &assetCfg); err != nil {
		slog.ErrorContext(ctx, "Error while unmarshalling events", slog.Any("err", err))
		return
	}

	var eventCfg *pogoicons.EventConfig
	for _, e := range assetCfg.Events {
		if e.Name == *event {
			eventCfg = &e
			break
		}
	}
	if eventCfg == nil {
		slog.ErrorContext(ctx, "Event not found", slog.String("event", *event))
		return
	}

	var pokemonImages []io.Reader
	for _, p := range strings.Split(*pokemon, ",") {
		pf, err := pokeClient.GetPokemonForm(ctx, p)
		if err != nil {
			if errors.Is(err, pokeapi.ErrNotFound) {
				slog.ErrorContext(ctx, "Pokemon not found", slog.String("pokemon", p))
				return
			}
			slog.ErrorContext(ctx, "Error while getting Pokemon", slog.Any("err", err))
			return
		}

		pokemonImage, err := pokeClient.GetSprite(ctx, pf.Sprite)
		if err != nil {
			slog.Error("Error while getting Pokemon image", slog.Any("err", err))
			return
		}
		defer pokemonImage.Body.Close()
		pokemonImages = append(pokemonImages, pokemonImage.Body)
	}

	r, err := pogoicon.Generate(backgroundImage)
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
