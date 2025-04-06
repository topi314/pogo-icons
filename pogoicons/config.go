package pogoicons

import (
	"fmt"
	"log/slog"
	"os"
	"strings"

	"github.com/BurntSushi/toml"
	"github.com/disgoorg/snowflake/v2"
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
		Repository: "https://github.com/PokeAPI/api-data",
		Bot: BotConfig{
			Token:        "",
			GuildIDs:     nil,
			SyncCommands: true,
		},
		Log: LogConfig{
			Level:     slog.LevelInfo,
			Format:    LogFormatText,
			AddSource: false,
			NoColor:   false,
		},
	}
}

type Config struct {
	Repository string    `toml:"repository"`
	Bot        BotConfig `toml:"bot"`
	Log        LogConfig `toml:"log"`
}

func (c Config) String() string {
	return fmt.Sprintf("Repository: %s\nBot: %s\nLog: %s",
		c.Repository,
		c.Bot,
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

type BotConfig struct {
	Token        string         `toml:"token"`
	GuildIDs     []snowflake.ID `toml:"guild_ids"`
	SyncCommands bool           `toml:"sync_commands"`
}

func (c BotConfig) String() string {
	return fmt.Sprintf("\n Token: %s\n GuildIDs: %s\n SyncCommands: %t",
		strings.Repeat("*", len(c.Token)),
		c.GuildIDs,
		c.SyncCommands,
	)
}
