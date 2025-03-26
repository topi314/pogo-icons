package main

import (
	"flag"
	"log/slog"
	"os"
	"os/signal"
	"runtime/debug"
	"syscall"

	"github.com/charmbracelet/log"
	"github.com/muesli/termenv"

	"github.com/topi314/pogo-icons/server"
)

func main() {
	cfgPath := flag.String("config", "config.toml", "path to config file")
	flag.Parse()

	cfg, err := server.LoadConfig(*cfgPath)
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

	slog.Info("Starting server...", slog.String("version", version), slog.String("go_version", goVersion))
	slog.Info("Config loaded", slog.Any("config", cfg))

	s := server.New(cfg, version, goVersion)
	go s.Start()

	slog.Info("Server started", slog.Any("addr", cfg.ListenAddr))
	si := make(chan os.Signal, 1)
	signal.Notify(si, syscall.SIGINT, syscall.SIGTERM)
	<-si
}

func setupLogger(cfg server.LogConfig) {
	var formatter log.Formatter
	switch cfg.Format {
	case server.LogFormatJSON:
		formatter = log.JSONFormatter
	case server.LogFormatText:
		formatter = log.TextFormatter
	case server.LogFormatLogFMT:
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
	if cfg.Format == server.LogFormatText && !cfg.NoColor {
		handler.SetColorProfile(termenv.TrueColor)
	}

	slog.SetDefault(slog.New(handler))
}
