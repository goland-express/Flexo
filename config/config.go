package config

import (
	"fmt"
	"log/slog"

	"github.com/caarlos0/env/v11"
	"github.com/joho/godotenv"
)

type Config struct {
	Token            string `env:"DISCORD_TOKEN,required"`
	LavalinkHost     string `env:"LAVALINK_HOST,required"`
	LavalinkPassword string `env:"LAVALINK_PASSWORD,required"`
	Prefix           string `env:"BOT_PREFIX" envDefault:"!"`
}

func Load() (*Config, error) {
	if err := godotenv.Load(); err != nil {
		slog.Warn("Could not load .env file, relying on environment variables", slog.Any("error", err))
	}

	cfg := &Config{}
	if err := env.Parse(cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config: %w", err)
	}

	return cfg, nil
}
