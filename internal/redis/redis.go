package redis

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

type Config struct {
	Host     string        `env:"REDIS_HOST"     env-default:"localhost"`
	Port     int           `env:"REDIS_PORT"     env-default:"6379"`
	Password string        `env:"REDIS_PASSWORD" env-default:""`
	DB       int           `env:"REDIS_DB"       env-default:"0"`
	Timeout  time.Duration `env:"REDIS_TIMEOUT"  env-default:"10s"`
}

type RedisClient struct {
	Client *redis.Client
}

func New(config Config) (*RedisClient, error) {
	client := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%d", config.Host, config.Port),
		Password: config.Password,
		DB:       config.DB,
	})

	ctx, cancel := context.WithTimeout(context.Background(), config.Timeout)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		client.Close()
		return nil, fmt.Errorf("failed to connect to Redis: %w", err)
	}

	return &RedisClient{
		Client: client,
	}, nil
}

func (r *RedisClient) Close() {
	if r.Client != nil {
		if errClose := r.Client.Close(); errClose != nil {
			fmt.Print("error occured while closing redis client: ", errClose)
			return
		}
	}
}
