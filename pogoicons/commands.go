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

	"github.com/topi314/pogo-icons/internal/icongen"
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
			Description: "Generate a Pokémon GO event icon",
			Options: []discord.ApplicationCommandOption{
				discord.ApplicationCommandOptionString{
					Name:        "event",
					Description: "The event this image is for",
					Required:    true,
					Choices:     eventChoices,
				},
				discord.ApplicationCommandOptionString{
					Name:         "pokemon1",
					Description:  "The Pokémon to include",
					Required:     true,
					Autocomplete: true,
				},
				discord.ApplicationCommandOptionString{
					Name:         "pokemon2",
					Description:  "The Pokémon to include",
					Autocomplete: true,
				},
				discord.ApplicationCommandOptionString{
					Name:         "pokemon3",
					Description:  "The Pokémon to include",
					Autocomplete: true,
				},
				discord.ApplicationCommandOptionString{
					Name:         "pokemon4",
					Description:  "The Pokémon to include",
					Autocomplete: true,
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
		r.Use(middleware.Defer(discord.InteractionTypeApplicationCommand, false, false))
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
	opt := e.Data.Focused()
	value := e.Data.String(opt.Name)

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
	event := data.String("event")
	pokemonList := []string{data.String("pokemon1")}
	if pokemon, ok := data.OptString("pokemon2"); ok {
		pokemonList = append(pokemonList, pokemon)
	}
	if pokemon, ok := data.OptString("pokemon3"); ok {
		pokemonList = append(pokemonList, pokemon)
	}
	if pokemon, ok := data.OptString("pokemon4"); ok {
		pokemonList = append(pokemonList, pokemon)
	}
	var cosmetics []string
	if cosmetic, ok := data.OptString("cosmetic"); ok {
		cosmetics = append(cosmetics, cosmetic)
	}

	icon, err := icongen.Generate(b.assets, b.iconCfg, b.getPokemonImage, event, pokemonList, cosmetics)
	if err != nil {
		slog.ErrorContext(e.Ctx, "error generating icon", slog.Any("err", err))
		_, err = e.UpdateInteractionResponse(discord.MessageUpdate{
			Content: json.Ptr(fmt.Sprintf("Error generating icon: %s", err)),
		})
		return err
	}

	_, err = e.UpdateInteractionResponse(discord.MessageUpdate{
		Content: json.Ptr(fmt.Sprintf("Generated icon for `%s` with `%s`", event, strings.Join(pokemonList, ","))),
		Files: []*discord.File{
			discord.NewFile(fmt.Sprintf("%s_%s.png", strings.ReplaceAll(strings.ToLower(event), " ", "_"), strings.Join(pokemonList, "_")), "", icon),
		},
	})

	return err
}
