package cache

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/redis/go-redis/v9"
)

type RedisCache struct {
	client     *redis.Client
	defaultTTL time.Duration
	prefix     string
}

func NewRedisCache(client *redis.Client, defaultTTL time.Duration, prefix string) *RedisCache {
	return &RedisCache{
		client:     client,
		defaultTTL: defaultTTL,
		prefix:     prefix,
	}
}

func (rc *RedisCache) Get(ctx context.Context, key string, dest interface{}) error {
	val, err := rc.client.Get(ctx, rc.prefix+":"+key).Bytes()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return ErrMiss
		}
		return err
	}
	return json.Unmarshal(val, dest)
}

func (rc *RedisCache) Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
	newVal, err := json.Marshal(value)
	if err != nil {
		return err
	}
	if ttl > 0 {
		err = rc.client.Set(ctx, rc.prefix+":"+key, newVal, ttl).Err()
	} else {
		err = rc.client.Set(ctx, rc.prefix+":"+key, newVal, rc.defaultTTL).Err()
	}
	if err != nil {
		return err
	}
	return nil
}

func (rc *RedisCache) Delete(ctx context.Context, keys ...string) error {
	prefixedKeys := make([]string, len(keys))
	for i, key := range keys {
		prefixedKeys[i] = rc.prefix + ":" + key
	}
	err := rc.client.Del(ctx, prefixedKeys...).Err()
	if err != nil {
		return err
	}
	return nil
}

func (rc *RedisCache) Invalidate(ctx context.Context, pattern string) error {
	pattern = rc.prefix + ":" + pattern
	iter := rc.client.Scan(ctx, 0, pattern, 0).Iterator()

	var keys []string

	for iter.Next(ctx) {
		keys = append(keys, iter.Val())
	}

	if err := iter.Err(); err != nil {
		return err
	}

	if len(keys) > 0 {
		return rc.client.Del(ctx, keys...).Err()
	}
	return nil
}

func (rc *RedisCache) Close() error {
	return rc.client.Close()
}

func (rc *RedisCache) Ping(ctx context.Context) error {
	return rc.client.Ping(ctx).Err()
}
