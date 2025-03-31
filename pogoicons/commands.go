package pogoicons

import (
	"errors"
	"fmt"
	"io"
	"log/slog"
	"os"
	"path"
	"slices"
	"strconv"
	"strings"

	"github.com/disgoorg/disgo/bot"
	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/handler"
	"github.com/disgoorg/disgo/handler/middleware"
	"github.com/disgoorg/json"
	"go.gopad.dev/fuzzysearch/fuzzy"

	"github.com/topi314/pogo-icons/internal/pogoicon"
	"github.com/topi314/pogo-icons/internal/pokeapi"
)

func (s *Bot) commands() ([]discord.ApplicationCommandCreate, error) {
	var eventChoices []discord.ApplicationCommandOptionChoiceString
	for _, event := range s.assetCfg.Events {
		eventChoices = append(eventChoices, discord.ApplicationCommandOptionChoiceString{
			Name:  event.Name,
			Value: event.Name,
		})
	}

	var cosmeticChoices []discord.ApplicationCommandOptionChoiceString
	for _, cosmetic := range s.assetCfg.Cosmetics {
		cosmeticChoices = append(cosmeticChoices, discord.ApplicationCommandOptionChoiceString{
			Name:  cosmetic.Name,
			Value: cosmetic.Name,
		})
	}

	return []discord.ApplicationCommandCreate{
		discord.SlashCommandCreate{
			Name:        "info",
			Description: "Get some info about the bot",
			IntegrationTypes: []discord.ApplicationIntegrationType{
				discord.ApplicationIntegrationTypeUserInstall,
			},
			Contexts: []discord.InteractionContextType{
				discord.InteractionContextTypeGuild,
				discord.InteractionContextTypeBotDM,
				discord.InteractionContextTypePrivateChannel,
			},
		},
		discord.SlashCommandCreate{
			Name:        "generate",
			Description: "Generate an icon for a Pokémon",
			Options: []discord.ApplicationCommandOption{
				discord.ApplicationCommandOptionInt{
					Name:         "pokemon",
					Description:  "The Pokémon to generate an icon for",
					Required:     true,
					Autocomplete: true,
				},
				discord.ApplicationCommandOptionString{
					Name:        "event",
					Description: "The event this image is for",
					Required:    true,
					Choices:     eventChoices,
				},
				discord.ApplicationCommandOptionString{
					Name:        "cosmetic",
					Description: "The cosmetic to use for the icon",
					Required:    false,
					Choices:     cosmeticChoices,
				},
			},
			IntegrationTypes: []discord.ApplicationIntegrationType{
				discord.ApplicationIntegrationTypeUserInstall,
			},
			Contexts: []discord.InteractionContextType{
				discord.InteractionContextTypeGuild,
				discord.InteractionContextTypeBotDM,
				discord.InteractionContextTypePrivateChannel,
			},
		},
	}, nil
}

func (s *Bot) routes() bot.EventListener {
	r := handler.New()
	r.Use(middleware.Go)
	r.SlashCommand("/info", s.onInfo)
	r.Route("/generate", func(r handler.Router) {
		r.Autocomplete("/", s.onGenerateIconAutocomplete)
		r.SlashCommand("/", s.onGenerateIcon)
	})

	return r
}

func (s *Bot) onInfo(_ discord.SlashCommandInteractionData, e *handler.CommandEvent) error {
	return e.CreateMessage(discord.MessageCreate{
		Content: fmt.Sprintf("PogoIcons is a bot that generates evemt icons for Pokémon GO.\n\n**Version:** `%s`\n**Go Version:** `%s`\n",
			s.version,
			s.goVersion,
		),
		Flags: discord.MessageFlagEphemeral,
	})
}

func (s *Bot) onGenerateIconAutocomplete(e *handler.AutocompleteEvent) error {
	value := e.Data.String("pokemon")

	pokemon, err := s.pokeClient.GetPokemonSpecies(e.Ctx)
	if err != nil {
		slog.ErrorContext(e.Ctx, "error getting pokemon species", slog.Any("err", err))
		return e.AutocompleteResult([]discord.AutocompleteChoice{})
	}

	ranks := fuzzy.RankFindNormalizedFold(value, pokemon)
	if len(ranks) == 0 {
		return e.AutocompleteResult([]discord.AutocompleteChoice{})
	}
	choices := make([]discord.AutocompleteChoice, 0, max(25, len(ranks)))
	for i, rank := range ranks {
		if i >= 25 {
			break
		}
		choices = append(choices, discord.AutocompleteChoiceInt{
			Name:  strings.Title(rank.Target.Name),
			Value: rank.Target.ID(),
		})
	}

	return e.AutocompleteResult(choices)
}

func (s *Bot) onGenerateIcon(data discord.SlashCommandInteractionData, e *handler.CommandEvent) error {
	pokemon := data.Int("pokemon")
	eventName := data.String("event")
	cosmeticName := data.String("cosmetic")

	eventIndex := slices.IndexFunc(s.assetCfg.Events, func(e EventConfig) bool {
		return e.Name == eventName
	})
	if eventIndex == -1 {
		return e.CreateMessage(discord.MessageCreate{
			Content: fmt.Sprintf("Event not found: `%s`", eventName),
			Flags:   discord.MessageFlagEphemeral,
		})
	}
	event := s.assetCfg.Events[eventIndex]

	if err := e.DeferCreateMessage(false); err != nil {
		return err
	}

	var cosmetic CosmeticConfig
	if cosmeticIndex := slices.IndexFunc(s.assetCfg.Cosmetics, func(c CosmeticConfig) bool {
		return c.Name == cosmeticName
	}); cosmeticIndex > -1 {
		cosmetic = s.assetCfg.Cosmetics[cosmeticIndex]
	}

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
		_, err = e.UpdateInteractionResponse(discord.MessageUpdate{
			Content: json.Ptr(fmt.Sprintf("Error getting pokemon sprite: %s", err)),
		})
	}
	defer pokemonSprite.Body.Close()

	backgroundImage, err := s.assets.Open(path.Join("assets/backgrounds", event.Background))
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			_, err = e.UpdateInteractionResponse(discord.MessageUpdate{
				Content: json.Ptr(fmt.Sprintf("Background asset not found: %s", err)),
			})
		}
		slog.ErrorContext(e.Ctx, "error opening background asset", slog.Any("err", err))
		_, err = e.UpdateInteractionResponse(discord.MessageUpdate{
			Content: json.Ptr(fmt.Sprintf("Error opening background asset: %s", err)),
		})
	}
	defer backgroundImage.Close()

	var backgroundIconImage io.Reader
	if event.BackgroundIcon != "" {
		img, err := s.assets.Open(path.Join("assets/background_icons", event.BackgroundIcon))
		if err != nil {
			if errors.Is(err, os.ErrNotExist) {
				_, err = e.UpdateInteractionResponse(discord.MessageUpdate{
					Content: json.Ptr(fmt.Sprintf("Background icon asset not found: %s", err)),
				})
			}
			slog.ErrorContext(e.Ctx, "error opening background icon asset", slog.Any("err", err))
			_, err = e.UpdateInteractionResponse(discord.MessageUpdate{
				Content: json.Ptr(fmt.Sprintf("Error opening background icon asset: %s", err)),
			})
		}
		defer img.Close()
		backgroundIconImage = img
	}

	var cosmeticImage io.Reader
	if cosmeticName != "" {
		img, err := s.assets.Open(path.Join("assets/cosmetics", cosmetic.Image))
		if err != nil {
			if errors.Is(err, os.ErrNotExist) {
				_, err = e.UpdateInteractionResponse(discord.MessageUpdate{
					Content: json.Ptr(fmt.Sprintf("Cosmetic asset not found: %s", err)),
				})
				return err
			}
			slog.ErrorContext(e.Ctx, "error opening cosmetic asset", slog.Any("err", err))
			_, err = e.UpdateInteractionResponse(discord.MessageUpdate{
				Content: json.Ptr(fmt.Sprintf("Error opening cosmetic asset: %s", err)),
			})
		}
		defer img.Close()
		cosmeticImage = img
	}

	icon, err := pogoicon.Generate(backgroundImage, backgroundIconImage, event.BackgroundIconScale, pokemonSprite.Body, event.PokemonScale, cosmeticImage, cosmetic.Scale)
	if err != nil {
		slog.ErrorContext(e.Ctx, "error generating icon", slog.Any("err", err))
		_, err = e.UpdateInteractionResponse(discord.MessageUpdate{
			Content: json.Ptr(fmt.Sprintf("Error generating icon: %s", err)),
		})
	}

	_, err = e.UpdateInteractionResponse(discord.MessageUpdate{
		Content: json.Ptr(fmt.Sprintf("Generated icon for `%s` with background `%s`", p.Name, eventName)),
		Files: []*discord.File{
			discord.NewFile(fmt.Sprintf("%s_%s.png", p.Name, strings.ReplaceAll(strings.ToLower(eventName), " ", "_")), "", icon),
		},
	})

	return err
}
