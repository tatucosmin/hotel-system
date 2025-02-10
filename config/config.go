package config

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/caarlos0/env/v11"
)

type EnvType string

const (
	Env_Test EnvType = "test"
	Env_Dev  EnvType = "dev"
	Env_Prod EnvType = "prod"
)

type Config struct {
	ServerPort           string  `env:"SERVER_PORT"`
	ServerHost           string  `env:"SERVER_HOST"`
	DatabaseHost         string  `env:"DB_HOST"`
	DatabasePort         string  `env:"DB_PORT"`
	DatabaseUser         string  `env:"DB_USER"`
	DatabaseName         string  `env:"DB_NAME"`
	DatabasePassword     string  `env:"DB_PASS"`
	DatabaseTestPort     string  `env:"DB_TEST_PORT"`
	Env                  EnvType `env:"ENV" defaultEnv:"dev"`
	JwtSecret            string  `env:"JWT_SECRET"`
	ProjectRoot          string  `env:"PROJECT_ROOT"`
	S3LocalStackEndpoint string  `env:"LOCALSTACK_S3_ENDPOINT"`
	S3Bucket             string  `env:"S3_BUCKET"`
	S3Client             *s3.Client
}

func New() (*Config, error) {
	sdkConfig, err := awsconfig.LoadDefaultConfig(context.Background())
	if err != nil {
		return nil, fmt.Errorf("failed to load aws sdk config: %w", err)
	}

	cfg, err := env.ParseAs[Config]()
	if err != nil {
		return nil, fmt.Errorf("failed to load config from .envrc: %w", err)
	}

	s3Client := s3.NewFromConfig(sdkConfig, func(opts *s3.Options) {
		if cfg.Env != Env_Prod {
			opts.BaseEndpoint = aws.String(cfg.S3LocalStackEndpoint)
			opts.UsePathStyle = true
		}
	})

	cfg.S3Client = s3Client
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
