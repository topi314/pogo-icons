package pogoicons

import (
	"context"
	"io/fs"

	"github.com/disgoorg/disgo/bot"
	"github.com/disgoorg/disgo/handler"

	"github.com/topi314/pogo-icons/internal/pokeapi"
)

func New(client bot.Client, cfg Config, version string, goVersion string, assets fs.FS) *Bot {
	s := &Bot{
		cfg:        cfg,
		version:    version,
		goVersion:  goVersion,
		assets:     assets,
		client:     client,
		pokeClient: pokeapi.New("https://pokeapi.co/api/v2"),
	}

	client.AddEventListeners(s.Routes())

	return s
}

type Bot struct {
	cfg        Config
	version    string
	goVersion  string
	assets     fs.FS
	client     bot.Client
	pokeClient *pokeapi.Client
}

func (s *Bot) Start() {
	if s.cfg.Bot.SyncCommands {
		go func() {
			if err := handler.SyncCommands(s.client, commands, s.cfg.Bot.GuildIDs); err != nil {
				s.client.Logger().Error("failed to sync commands", err)
			}
		}()
	}

	if err := s.client.OpenGateway(context.Background()); err != nil {
		s.client.Logger().Error("failed to open gateway", err)
		return
	}
}
