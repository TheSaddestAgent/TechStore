package config

import (
	"fmt"
	postgres "techstore/internal/db"
	"techstore/internal/redis"

	"github.com/ilyakaznacheev/cleanenv"
)

type Config struct {
	HTTPPort int    `env:"HTTP_PORT" env-default:"8080"`
	Env      string `env:"ENV" env-default:"prod"`

	AuthServiceHost    string `env:"AUTH_SERVICE_HOST" env-default:"auth-service"`
	AuthServicePort    int    `env:"AUTH_SERVICE_PORT" env-default:"8081"`
	ProductServiceHost string `env:"PRODUCT_SERVICE_HOST" env-default:"product-service"`
	ProductServicePort int    `env:"PRODUCT_SERVICE_PORT" env-default:"8082"`

	postgres.Config

	RedisConfig redis.Config
}

func (c *Config) GetAuthServiceURL() string {
	return fmt.Sprintf("http://%s:%d", c.AuthServiceHost, c.AuthServicePort)
}

func (c *Config) GetProductServiceURL() string {
	return fmt.Sprintf("http://%s:%d", c.ProductServiceHost, c.ProductServicePort)
}

func ParseConfig() (*Config, error) {
	cfg := &Config{}

	if err := cleanenv.ReadEnv(cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config from env: %w", err)
	}

	return cfg, nil
}
