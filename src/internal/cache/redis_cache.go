package cache

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
)

type RedisCache struct {
	client *redis.Client
}

func NewRedisCache(client *redis.Client) *RedisCache {
	return &RedisCache{client: client}
}

func (rc *RedisCache) Get(ctx context.Context, key string, dest interface{}) error {
	rc.client.Get(ctx, key)
}

func (rc *RedisCache) Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
	//TODO implement me
	panic("implement me")
}

func (rc *RedisCache) Delete(ctx context.Context, keys ...string) error {
	//TODO implement me
	panic("implement me")
}

func (rc *RedisCache) Invalidate(ctx context.Context, pattern string) error {
	//TODO implement me
	panic("implement me")
}

func (rc *RedisCache) Close() error {
	return rc.client.Close()
}

func (rc *RedisCache) Ping(ctx context.Context) error {
	return rc.client.Ping(ctx).Err()
}
