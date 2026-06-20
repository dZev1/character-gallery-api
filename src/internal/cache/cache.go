package cache

import (
	"context"
	"time"
)

type Cache interface {
	Get(ctx context.Context, key string, dest interface{}) error
	Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error
	Delete(ctx context.Context, keys ...string) error
	Invalidate(ctx context.Context, pattern string) error
	Ping(ctx context.Context) error
	Close() error
}
