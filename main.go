package main

import (
	"embed"
	"flag"
	"io/fs"
	"log/slog"
	"os"
	"os/signal"
	"runtime/debug"
	"syscall"

	"github.com/BurntSushi/toml"
	"github.com/charmbracelet/log"
	"github.com/disgoorg/disgo"
	"github.com/disgoorg/disgo/bot"
	"github.com/muesli/termenv"

	"github.com/topi314/pogo-icons/internal/icongen"
	"github.com/topi314/pogo-icons/internal/pokeapi"
	"github.com/topi314/pogo-icons/pogoicons"
)

var (
	//go:embed assets
	assets embed.FS

	//go:embed assets/generate.toml
	generateConfig []byte
)

func main() {
	cfgPath := flag.String("config", "config.toml", "path to config file")
	flag.Parse()

	cfg, err := pogoicons.LoadConfig(*cfgPath)
	if err != nil {
		slog.Error("Error while loading config", slog.Any("err", err))
		return
	}

	setupLogger(cfg.Log)

	version := "unknown"
	goVersion := "unknown"
	if info, ok := debug.ReadBuildInfo(); ok {
		version = info.Main.Version
		goVersion = info.GoVersion
	}

	slog.Info("Starting bpt...", slog.String("version", version), slog.String("go_version", goVersion))
	slog.Info("Config loaded", slog.Any("config", cfg))

	client, err := disgo.New(cfg.Bot.Token, bot.WithDefaultGateway())
	if err != nil {
		slog.Error("Error while creating bot client", slog.Any("err", err))
		return
	}

	var assetCfg icongen.Config
	if err = toml.Unmarshal(generateConfig, &assetCfg); err != nil {
		slog.Error("Error while unmarshalling events", slog.Any("err", err))
		return
	}

	subAssets, err := fs.Sub(assets, "assets")
	if err != nil {
		slog.Error("Error while creating assets sub fs", slog.Any("err", err))
		return
	}

	pokeClient, err := pokeapi.NewGit(cfg.Repository, cfg.ClonePath)
	if err != nil {
		slog.Error("Error while creating pokeapi client", slog.Any("err", err))
		return
	}

	b := pogoicons.New(client, pokeClient, cfg, version, goVersion, subAssets, assetCfg)
	go b.Start()

	slog.Info("Bot started")
	si := make(chan os.Signal, 1)
	signal.Notify(si, syscall.SIGINT, syscall.SIGTERM)
	<-si
}

func setupLogger(cfg pogoicons.LogConfig) {
	var formatter log.Formatter
	switch cfg.Format {
	case pogoicons.LogFormatJSON:
		formatter = log.JSONFormatter
	case pogoicons.LogFormatText:
		formatter = log.TextFormatter
	case pogoicons.LogFormatLogFMT:
		formatter = log.LogfmtFormatter
	default:
		slog.Error("Unknown log format", slog.String("format", string(cfg.Format)))
		os.Exit(-1)
	}

	handler := log.NewWithOptions(os.Stdout, log.Options{
		Level:           log.Level(cfg.Level),
		ReportTimestamp: true,
		ReportCaller:    cfg.AddSource,
		Formatter:       formatter,
	})
	if cfg.Format == pogoicons.LogFormatText && !cfg.NoColor {
		handler.SetColorProfile(termenv.TrueColor)
	}

	slog.SetDefault(slog.New(handler))
}
