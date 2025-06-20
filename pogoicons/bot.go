package pogoicons

import (
	"context"
	"io"
	"io/fs"
	"log/slog"

	"github.com/disgoorg/disgo/bot"
	"github.com/disgoorg/disgo/handler"

	"github.com/topi314/pogo-icons/internal/icongen"
	"github.com/topi314/pogo-icons/internal/pokeapi"
)

func New(client bot.Client, pokeClient pokeapi.Client, cfg Config, version string, goVersion string, assets fs.FS, iconCfg icongen.Config) *Bot {
	s := &Bot{
		cfg:        cfg,
		version:    version,
		goVersion:  goVersion,
		assets:     assets,
		iconCfg:    iconCfg,
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
	iconCfg    icongen.Config
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

func (b *Bot) getPokemonImage(ctx context.Context, p string) (io.ReadCloser, error) {
	pf, err := b.pokeClient.GetPokemonForm(ctx, p)
	if err != nil {
		return nil, err
	}

	rs, err := b.pokeClient.GetSprite(ctx, pf.Sprite)
	if err != nil {
		return nil, err
	}

	return rs.Body, nil
}
