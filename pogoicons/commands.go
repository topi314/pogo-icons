package pogoicons

import (
	"fmt"
	"log/slog"
	"strings"

	"github.com/disgoorg/disgo/bot"
	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/handler"
	"github.com/disgoorg/disgo/handler/middleware"
	"github.com/disgoorg/json"
	"go.gopad.dev/fuzzysearch/fuzzy"

	"github.com/topi314/pogo-icons/internal/pogoicon"
)

func (b *Bot) commands() ([]discord.ApplicationCommandCreate, error) {
	var eventChoices []discord.ApplicationCommandOptionChoiceString
	for _, event := range b.iconCfg.Events {
		eventChoices = append(eventChoices, discord.ApplicationCommandOptionChoiceString{
			Name:  event.Name,
			Value: event.Name,
		})
	}

	var cosmeticChoices []discord.ApplicationCommandOptionChoiceString
	for _, cosmetic := range b.iconCfg.Cosmetics {
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
	values := strings.Split(value, ",")

	pokemon, err := b.pokeClient.GetPokemon(e.Ctx)
	if err != nil {
		slog.ErrorContext(e.Ctx, "error getting pokemon", slog.Any("err", err))
		return e.AutocompleteResult([]discord.AutocompleteChoice{})
	}

	for _, v := range values {
		ranks := fuzzy.RankFindNormalizedFold(v, pokemon)
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
	}

	return e.AutocompleteResult(choices)
}

func (b *Bot) onGenerateIcon(data discord.SlashCommandInteractionData, e *handler.CommandEvent) error {
	event := data.String("event")
	pokemon := data.String("pokemon")
	cosmetic := data.String("cosmetic")

	pokemonList := strings.Split(pokemon, ",")
	cosmetics := strings.Split(cosmetic, ",")

	icon, err := pogoicon.Generate(b.assets, b.iconCfg, b.getPokemonImage, event, pokemonList, cosmetics)
	if err != nil {
		slog.ErrorContext(e.Ctx, "error generating icon", slog.Any("err", err))
		_, err = e.UpdateInteractionResponse(discord.MessageUpdate{
			Content: json.Ptr(fmt.Sprintf("Error generating icon: %s", err)),
		})
	}

	_, err = e.UpdateInteractionResponse(discord.MessageUpdate{
		Content: json.Ptr(fmt.Sprintf("Generated icon for `%s` with `%s`", event, pokemon)),
		Files: []*discord.File{
			discord.NewFile(fmt.Sprintf("%s_%s.png", strings.ReplaceAll(strings.ToLower(event), " ", "_"), pokemon), "", icon),
		},
	})

	return err
}
