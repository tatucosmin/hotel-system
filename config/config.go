package config

import (
	"fmt"

	"github.com/caarlos0/env/v11"
)

type EnvType string

const (
	Env_Test EnvType = "test"
	Env_Dev  EnvType = "dev"
)

type Config struct {
	ServerPort       string  `env:"SERVER_PORT"`
	ServerHost       string  `env:"SERVER_HOST"`
	DatabaseHost     string  `env:"DB_HOST"`
	DatabasePort     string  `env:"DB_PORT"`
	DatabaseUser     string  `env:"DB_USER"`
	DatabaseName     string  `env:"DB_NAME"`
	DatabasePassword string  `env:"DB_PASS"`
	DatabaseTestPort string  `env:"DB_TEST_PORT"`
	Env              EnvType `env:"ENV" defaultEnv:"dev"`
	JwtSecret        string  `env:"JWT_SECRET"`
	ProjectRoot      string  `env:"PROJECT_ROOT"`
}

func New() (*Config, error) {
	cfg, err := env.ParseAs[Config]()
	if err != nil {
		return nil, fmt.Errorf("failed to load config from .envrc: %w", err)
	}
	return &cfg, nil
}

func (c *Config) DatabaseUrl() string {
	port := c.DatabasePort
	if c.Env == Env_Test {
		port = c.DatabaseTestPort
	}
	return fmt.Sprintf("postgresql://%s:%s@%s:%s/%s?sslmode=disable",
		c.DatabaseUser,
		c.DatabasePassword,
		c.DatabaseHost,
		port,
		c.DatabaseName,
	)
}
