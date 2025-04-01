package pogoicons

import (
	"context"
	"io/fs"
	"log/slog"

	"github.com/disgoorg/disgo/bot"
	"github.com/disgoorg/disgo/handler"

	"github.com/topi314/pogo-icons/internal/pokeapi"
)

func New(client bot.Client, pokeClient pokeapi.Client, cfg Config, version string, goVersion string, assets fs.FS, assetCfg AssetConfig) *Bot {
	s := &Bot{
		cfg:        cfg,
		version:    version,
		goVersion:  goVersion,
		assets:     assets,
		assetCfg:   assetCfg,
		client:     client,
		pokeClient: pokeClient,
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
	pokeClient pokeapi.Client
}

func (b *Bot) Start() {
	if b.cfg.Bot.SyncCommands {
		go func() {
			slog.Info("Syncing commands")
			commands, err := b.commands()
			if err != nil {
				b.client.Logger().Error("failed to sync commands", err)
				return
			}
			if err = handler.SyncCommands(b.client, commands, b.cfg.Bot.GuildIDs); err != nil {
				b.client.Logger().Error("failed to sync commands", err)
			}
		}()
	}

	if err := b.client.OpenGateway(context.Background()); err != nil {
		b.client.Logger().Error("failed to open gateway", err)
		return
	}
}
