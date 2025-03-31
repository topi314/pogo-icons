package pogoicons

import (
	"fmt"
	"log/slog"
	"os"

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
		PokeAPIURL: "https://pokeapi.co/api/v2",
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
	PokeAPIURL string    `toml:"pokeapi_url"`
	Bot        BotConfig `toml:"bot"`
	Log        LogConfig `toml:"log"`
}

func (c Config) String() string {
	return fmt.Sprintf("PokeAPIURL: %s\nBot: %s\nLog: %s",
		c.PokeAPIURL,
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
		c.Token,
		c.GuildIDs,
		c.SyncCommands,
	)
}

type AssetConfig struct {
	Events    []EventConfig    `toml:"events"`
	Cosmetics []CosmeticConfig `toml:"cosmetics"`
}

type EventConfig struct {
	Name                string  `toml:"name"`
	Background          string  `toml:"background"`
	BackgroundIcon      string  `toml:"background_icon"`
	BackgroundIconScale float64 `toml:"background_icon_scale"`
	PokemonScale        float64 `toml:"pokemon_scale"`
}

type CosmeticConfig struct {
	Name  string  `toml:"name"`
	Image string  `toml:"image"`
	Scale float64 `toml:"scale"`
}
