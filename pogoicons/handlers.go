package pogoicons

import (
	"errors"
	"fmt"
	"log/slog"
	"os"
	"path"
	"strconv"

	"github.com/disgoorg/disgo/bot"
	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/handler"
	"github.com/disgoorg/disgo/handler/middleware"
	"github.com/disgoorg/json"

	"github.com/topi314/pogo-icons/internal/pogoicon"
	"github.com/topi314/pogo-icons/internal/pokeapi"
)

func (s *Bot) Routes() bot.EventListener {
	r := handler.New()
	r.Use(middleware.Go)
	r.SlashCommand("/info", s.onInfo)
	r.Route("/icon/generate", func(r handler.Router) {
		r.Autocomplete("/", s.onGenerateIconAutocomplete)
		r.SlashCommand("/", s.onGenerateIcon)
	})

	return r
}

func (s *Bot) onInfo(_ discord.SlashCommandInteractionData, e *handler.CommandEvent) error {
	return e.CreateMessage(discord.MessageCreate{
		Content: fmt.Sprintf("PogoIcons is a bot that generates icons for Pokémon GO.\n\nVersion: %s\nGo Version: %s\n",
			s.version,
			s.goVersion,
		),
		Flags: discord.MessageFlagEphemeral,
	})
}

func (s *Bot) onGenerateIconAutocomplete(e *handler.AutocompleteEvent) error {
	option := e.Data.Focused()
	switch option.Name {
	case "pokemon":
		return e.AutocompleteResult([]discord.AutocompleteChoice{
			discord.AutocompleteChoiceInt{
				Name:  "Bulbasaur",
				Value: 1,
			},
		})

	default:
		return fmt.Errorf("autocomplete option does not have a pokemon option")
	}
}

func (s *Bot) onGenerateIcon(data discord.SlashCommandInteractionData, e *handler.CommandEvent) error {
	if err := e.DeferCreateMessage(false); err != nil {
		return err
	}

	pokemon := data.Int("pokemon")
	background := data.String("background")

	p, err := s.pokeClient.GetPokemon(e.Ctx, strconv.Itoa(pokemon))
	if err != nil {
		if errors.Is(err, pokeapi.ErrNotFound) {
			_, err = e.UpdateInteractionResponse(discord.MessageUpdate{
				Content: json.Ptr(fmt.Sprintf("Pokemon with ID %d not found", pokemon)),
			})
			return err
		}
		slog.ErrorContext(e.Ctx, "error getting pokemon", slog.Any("err", err))
		_, err = e.UpdateInteractionResponse(discord.MessageUpdate{
			Content: json.Ptr(fmt.Sprintf("Error getting pokemon: %s", err)),
		})
		return err
	}

	pokemonSprite, err := s.pokeClient.GetSprite(e.Ctx, p.Sprites.Other.OfficialArtwork.FrontDefault)
	if err != nil {
		slog.ErrorContext(e.Ctx, "error getting pokemon sprite", slog.Any("err", err))
		return e.CreateMessage(discord.MessageCreate{
			Content: fmt.Sprintf("Error getting pokemon sprite: %s", err),
		})
	}
	defer pokemonSprite.Body.Close()

	backgroundAsset, err := s.assets.Open(path.Join("assets/backgrounds", background))
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return e.CreateMessage(discord.MessageCreate{
				Content: fmt.Sprintf("Background asset not found: %s", err),
			})
		}
		slog.ErrorContext(e.Ctx, "error opening background asset", slog.Any("err", err))
		return e.CreateMessage(discord.MessageCreate{
			Content: fmt.Sprintf("Error opening background asset: %s", err),
		})
	}
	defer backgroundAsset.Close()

	icon, err := pogoicon.Generate(e.Ctx, pokemonSprite.Body, backgroundAsset, pogoicon.Options{
		FFMPEG:      s.cfg.FFMPEG,
		ScaleWidth:  -1,
		ScaleHeight: 450,
	})
	if err != nil {
		slog.ErrorContext(e.Ctx, "error generating icon", slog.Any("err", err))
		return e.CreateMessage(discord.MessageCreate{
			Content: fmt.Sprintf("Error generating icon: %s", err),
		})
	}

	_, err = e.UpdateInteractionResponse(discord.MessageUpdate{
		Content: json.Ptr(fmt.Sprintf("Generated icon for `%s` with background `%s`", p.Name, background)),
		Files: []*discord.File{
			discord.NewFile(fmt.Sprintf("%s-%s", p.Name, background), "", icon),
		},
	})

	return err
}
