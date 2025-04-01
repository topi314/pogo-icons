package pogoicons

import (
	"errors"
	"fmt"
	"io"
	"log/slog"
	"os"
	"path"
	"slices"
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

func (b *Bot) commands() ([]discord.ApplicationCommandCreate, error) {
	var eventChoices []discord.ApplicationCommandOptionChoiceString
	for _, event := range b.assetCfg.Events {
		eventChoices = append(eventChoices, discord.ApplicationCommandOptionChoiceString{
			Name:  event.Name,
			Value: event.Name,
		})
	}

	var cosmeticChoices []discord.ApplicationCommandOptionChoiceString
	for _, cosmetic := range b.assetCfg.Cosmetics {
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
				discord.ApplicationCommandOptionString{
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
				discord.ApplicationCommandOptionBool{
					Name:        "shiny",
					Description: "Whether to use the shiny variant of the Pokémon",
				},
				discord.ApplicationCommandOptionString{
					Name:        "cosmetic",
					Description: "The cosmetic to use for the icon",
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

func (b *Bot) routes() bot.EventListener {
	r := handler.New()
	r.Use(middleware.Go)
	r.SlashCommand("/info", b.onInfo)
	r.Route("/generate", func(r handler.Router) {
		r.Autocomplete("/", b.onGenerateIconAutocomplete)
		r.SlashCommand("/", b.onGenerateIcon)
	})

	return r
}

func (b *Bot) onInfo(_ discord.SlashCommandInteractionData, e *handler.CommandEvent) error {
	return e.CreateMessage(discord.MessageCreate{
		Content: fmt.Sprintf("PogoIcons is a bot that generates evemt icons for Pokémon GO.\n\n**Version:** `%s`\n**Go Version:** `%s`\n",
			b.version,
			b.goVersion,
		),
		Flags: discord.MessageFlagEphemeral,
	})
}

func (b *Bot) onGenerateIconAutocomplete(e *handler.AutocompleteEvent) error {
	value := e.Data.String("pokemon")

	pokemon, err := b.pokeClient.GetPokemon(e.Ctx)
	if err != nil {
		slog.ErrorContext(e.Ctx, "error getting pokemon", slog.Any("err", err))
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
		choices = append(choices, discord.AutocompleteChoiceString{
			Name:  rank.Target.Name,
			Value: rank.Target.Value,
		})
	}

	return e.AutocompleteResult(choices)
}

func (b *Bot) onGenerateIcon(data discord.SlashCommandInteractionData, e *handler.CommandEvent) error {
	pokemon := data.String("pokemon")
	eventName := data.String("event")
	shiny := data.Bool("shiny")
	cosmeticName := data.String("cosmetic")

	eventIndex := slices.IndexFunc(b.assetCfg.Events, func(e EventConfig) bool {
		return e.Name == eventName
	})
	if eventIndex == -1 {
		return e.CreateMessage(discord.MessageCreate{
			Content: fmt.Sprintf("Event not found: `%s`", eventName),
			Flags:   discord.MessageFlagEphemeral,
		})
	}
	event := b.assetCfg.Events[eventIndex]

	if err := e.DeferCreateMessage(false); err != nil {
		return err
	}

	var cosmetic CosmeticConfig
	if cosmeticIndex := slices.IndexFunc(b.assetCfg.Cosmetics, func(c CosmeticConfig) bool {
		return c.Name == cosmeticName
	}); cosmeticIndex > -1 {
		cosmetic = b.assetCfg.Cosmetics[cosmeticIndex]
	}

	p, err := b.pokeClient.GetPokemonForm(e.Ctx, pokemon)
	if err != nil {
		if errors.Is(err, pokeapi.ErrNotFound) {
			_, err = e.UpdateInteractionResponse(discord.MessageUpdate{
				Content: json.Ptr(fmt.Sprintf("Pokemon `%s` not found", pokemon)),
			})
			return err
		}
		slog.ErrorContext(e.Ctx, "error getting pokemon", slog.Any("err", err))
		_, err = e.UpdateInteractionResponse(discord.MessageUpdate{
			Content: json.Ptr(fmt.Sprintf("Error getting pokemon: %s", err)),
		})
		return err
	}

	sprite := p.Sprite
	if shiny && p.ShinySprite != "" {
		sprite = p.ShinySprite
	}

	pokemonSprite, err := b.pokeClient.GetSprite(e.Ctx, sprite)
	if err != nil {
		slog.ErrorContext(e.Ctx, "error getting pokemon sprite", slog.Any("err", err))
		_, err = e.UpdateInteractionResponse(discord.MessageUpdate{
			Content: json.Ptr(fmt.Sprintf("Error getting pokemon sprite: %s", err)),
		})
	}
	defer pokemonSprite.Body.Close()

	backgroundImage, err := b.assets.Open(path.Join("assets/backgrounds", event.Background))
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
		img, err := b.assets.Open(path.Join("assets/background_icons", event.BackgroundIcon))
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
		img, err := b.assets.Open(path.Join("assets/cosmetics", cosmetic.Image))
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
