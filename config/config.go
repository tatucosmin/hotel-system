package config

import (
	"fmt"

	"github.com/caarlos0/env/v11"
)

type Config struct {
	Port             string `env:"PORT"`
	DatabaseHost     string `env:"DB_HOST"`
	DatabasePort     string `env:"DB_PORT"`
	DatabaseUser     string `env:"DB_USER"`
	DatabaseName     string `env:"DB_NAME"`
	DatabasePassword string `env:"DB_PASS"`
}

func New() (*Config, error) {
	cfg, err := env.ParseAs[Config]()
	if err != nil {
		return nil, fmt.Errorf("failed to load config from .envrc: %w", err)
	}
	return &cfg, nil
}

func (c *Config) DatabaseUrl() string {
	return fmt.Sprintf("postgresql://%s:%s@%s:%s/%s",
		c.DatabaseUser,
		c.DatabasePassword,
		c.DatabaseHost,
		c.DatabasePort,
		c.DatabaseName,
	)
}
