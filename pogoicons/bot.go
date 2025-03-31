package pogoicons

import (
	"context"
	"io/fs"
	"log/slog"
	"time"

	"github.com/disgoorg/disgo/bot"
	"github.com/disgoorg/disgo/handler"

	"github.com/topi314/pogo-icons/internal/pokeapi"
)

func New(client bot.Client, cfg Config, version string, goVersion string, assets fs.FS, assetCfg AssetConfig) *Bot {
	s := &Bot{
		cfg:        cfg,
		version:    version,
		goVersion:  goVersion,
		assets:     assets,
		assetCfg:   assetCfg,
		client:     client,
		pokeClient: pokeapi.New(cfg.PokeAPIURL),
	}

	client.AddEventListeners(s.routes())

	return s
}

type Bot struct {
	cfg        Config
	version    string
	goVersion  string
	assets     fs.FS
	assetCfg   AssetConfig
	client     bot.Client
	pokeClient *pokeapi.Client
}

func (s *Bot) Start() {
	if s.cfg.Bot.SyncCommands {
		go func() {
			slog.Info("Syncing commands")
			commands, err := s.commands()
			if err != nil {
				s.client.Logger().Error("failed to sync commands", err)
				return
			}
			if err = handler.SyncCommands(s.client, commands, s.cfg.Bot.GuildIDs); err != nil {
				s.client.Logger().Error("failed to sync commands", err)
			}
		}()
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	slog.Info("Fetching Pokémon species")
	if _, err := s.pokeClient.GetPokemonSpecies(ctx); err != nil {
		s.client.Logger().Error("failed to fetch pokemon species", err)
		return
	}

	if err := s.client.OpenGateway(context.Background()); err != nil {
		s.client.Logger().Error("failed to open gateway", err)
		return
	}
}
