package redis

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
)

func NewRedisClient(config *redis.Options) (*redis.Client, error) {
	client := redis.NewClient(config)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := client.Ping(ctx).Err(); err != nil {
		return nil, err
	}
	return client, nil
}
