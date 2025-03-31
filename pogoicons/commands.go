package pogoicons

import (
	"errors"
	"fmt"
	"io/fs"
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

var backgroundChoices = []discord.ApplicationCommandOptionChoiceString{
	{
		Name:  "Generic",
		Value: "generic.png",
	},
	{
		Name:  "Generic Day",
		Value: "generic_day.png",
	},
	{
		Name:  "Generic Night",
		Value: "generic_night.png",
	},
	{
		Name:  "Spotlight Hour",
		Value: "spotlight_hour.png",
	},
	{
		Name:  "Community Day",
		Value: "community_day.png",
	},
	{
		Name:  "Raid",
		Value: "raid.png",
	},
	{
		Name:  "Max Raid",
		Value: "max_raid.png",
	},
	{
		Name:  "Mega Raid",
		Value: "mega_raid.png",
	},
	{
		Name:  "Shadow Raid",
		Value: "shadow_raid.png",
	},
	{
		Name:  "Ex Raid",
		Value: "ex_raid.png",
	},
}

var cosmeticChoices = []discord.ApplicationCommandOptionChoiceString{
	{
		Name:  "CA Star",
		Value: "ca_star.png",
	},
}

var commands = []discord.ApplicationCommandCreate{
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
		Name:        "icon",
		Description: "Generate an icon for a Pokémon",
		Options: []discord.ApplicationCommandOption{
			discord.ApplicationCommandOptionSubCommand{
				Name:        "generate",
				Description: "Generate an icon for a Pokémon with a background",
				Options: []discord.ApplicationCommandOption{
					discord.ApplicationCommandOptionString{
						Name:        "background",
						Description: "The background to use for the icon",
						Required:    true,
						Choices:     backgroundChoices,
					},
					discord.ApplicationCommandOptionInt{
						Name:         "pokemon",
						Description:  "The Pokémon to generate an icon for",
						Required:     true,
						Autocomplete: true,
					},
					discord.ApplicationCommandOptionString{
						Name:        "cosmetic",
						Description: "The cosmetic to use for the icon",
						Required:    false,
						Choices:     cosmeticChoices,
					},
					discord.ApplicationCommandOptionFloat{
						Name:        "pokemon-scale",
						Description: "The scale of the Pokémon image",
						Required:    false,
						MinValue:    json.Ptr(0.0),
						MaxValue:    json.Ptr(2.0),
					},
					discord.ApplicationCommandOptionFloat{
						Name:        "cosmetic-scale",
						Description: "The scale of the cosmetic image",
						Required:    false,
						MinValue:    json.Ptr(0.0),
						MaxValue:    json.Ptr(2.0),
					},
				},
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
}

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

	background := data.String("background")
	pokemon := data.Int("pokemon")
	pokemonScale, ok := data.OptFloat("pokemon-scale")
	if !ok {
		pokemonScale = 1.0
	}
	cosmetic := data.String("cosmetic")
	cosmeticScale, ok := data.OptFloat("cosmetic-scale")
	if !ok {
		cosmeticScale = 0.15
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

	var cosmeticAsset fs.File
	if cosmetic != "" {
		cosmeticAsset, err = s.assets.Open(path.Join("assets/cosmetics", cosmetic))
		if err != nil {
			if errors.Is(err, os.ErrNotExist) {
				return e.CreateMessage(discord.MessageCreate{
					Content: fmt.Sprintf("Cosmetic asset not found: %s", err),
				})
			}
			slog.ErrorContext(e.Ctx, "error opening cosmetic asset", slog.Any("err", err))
			return e.CreateMessage(discord.MessageCreate{
				Content: fmt.Sprintf("Error opening cosmetic asset: %s", err),
			})
		}
		defer cosmeticAsset.Close()
	}

	icon, err := pogoicon.Generate(pokemonSprite.Body, pokemonScale, backgroundAsset, cosmeticAsset, cosmeticScale)
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
