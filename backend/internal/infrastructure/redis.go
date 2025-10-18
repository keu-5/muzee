package infrastructure

import (
	"context"

	"github.com/keu-5/muzee/backend/config"
	"github.com/redis/go-redis/v9"
	"go.uber.org/fx"
)

func NewRedisClient(lc fx.Lifecycle, cfg *config.Config, logger *Logger) *redis.Client {
	client := redis.NewClient(&redis.Options{
		Addr:     cfg.RedisAddr,
		Password: cfg.RedisPassword,
		DB:       cfg.RedisDB,
	})

	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			if err := client.Ping(ctx).Err(); err != nil {
				logger.Fatalf("failed to connect to redis: %v", err)
				return err
			}
			logger.Info("✓ Connected to Redis")
			return nil
		},
		OnStop: func(ctx context.Context) error {
			logger.Info("Closing Redis connection")
			return client.Close()
		},
	})

	return client
}
