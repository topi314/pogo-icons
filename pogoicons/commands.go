package pogoicons

import (
	"github.com/disgoorg/disgo/discord"
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
					discord.ApplicationCommandOptionInt{
						Name:         "pokemon",
						Description:  "The Pokémon to generate an icon for",
						Required:     true,
						Autocomplete: true,
					},
					discord.ApplicationCommandOptionString{
						Name:        "background",
						Description: "The background to use for the icon",
						Required:    true,
						Choices:     backgroundChoices,
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
