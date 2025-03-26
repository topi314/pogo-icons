package server

import (
	"fmt"
	"log/slog"
	"os"

	"github.com/BurntSushi/toml"
)

func LoadConfig(cfgPath string) (Config, error) {
	file, err := os.Open(cfgPath)
	if err != nil {
		return Config{}, fmt.Errorf("failed to open config file: %w", err)
	}
	defer func() {
		_ = file.Close()
	}()

	cfg := defaultConfig()
	if _, err = toml.NewDecoder(file).Decode(&cfg); err != nil {
		return Config{}, fmt.Errorf("failed to decode config file: %w", err)
	}

	return cfg, nil
}

func defaultConfig() Config {
	return Config{
		Log: LogConfig{
			Level:     slog.LevelInfo,
			Format:    LogFormatText,
			AddSource: false,
			NoColor:   false,
		},
		ListenAddr: "0.0.0.0",
		AssetsDir:  "assets",
	}
}

type Config struct {
	Dev        bool      `toml:"dev"`
	ListenAddr string    `toml:"listen_addr"`
	AssetsDir  string    `toml:"assets_dir"`
	Log        LogConfig `toml:"log"`
}

func (c Config) String() string {
	return fmt.Sprintf("Dev: %t\nListenAddr: %s\nAssetsDir: %s\nLog: %s",
		c.Dev,
		c.ListenAddr,
		c.AssetsDir,
		c.Log,
	)
}

type LogFormat string

const (
	LogFormatJSON   LogFormat = "json"
	LogFormatText   LogFormat = "text"
	LogFormatLogFMT LogFormat = "log-fmt"
)

type LogConfig struct {
	Level     slog.Level `toml:"level"`
	Format    LogFormat  `toml:"format"`
	AddSource bool       `toml:"add_source"`
	NoColor   bool       `toml:"no_color"`
}

func (c LogConfig) String() string {
	return fmt.Sprintf("\n Level: %s\n Format: %s\n AddSource: %t\n NoColor: %t",
		c.Level,
		c.Format,
		c.AddSource,
		c.NoColor,
	)
}
