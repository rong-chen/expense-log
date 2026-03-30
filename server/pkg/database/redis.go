package database

import (
	"context"
	"expense-log/internal/model"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

func InitRedis(cfg model.RedisConfig) *redis.Client {
	client := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%d", cfg.Host, cfg.Port),
		Password: cfg.Password,
		DB:       cfg.DB,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		panic(fmt.Errorf("Redis 连接失败: %w", err))
	}
	return client
}
