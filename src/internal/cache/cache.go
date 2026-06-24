package cache

import (
	"context"
	"errors"
	"time"
)

var (
	ErrMiss = errors.New("cache miss")
)

type Cache interface {
	Get(ctx context.Context, key string, dest interface{}) error
	Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error
	Delete(ctx context.Context, keys ...string) error
	Invalidate(ctx context.Context, pattern string) error
	Ping(ctx context.Context) error
	Close() error
}
