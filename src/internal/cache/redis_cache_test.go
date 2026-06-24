package cache

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/redis/go-redis/v9"
)

var globalClient *redis.Client

func TestMain(m *testing.M) {
	dsn := os.Getenv("REDIS_TEST_URL")
	if dsn == "" {
		fmt.Println("SKIP: REDIS_TEST_URL not set")
		os.Exit(0)
	}

	opts, err := redis.ParseURL(dsn)
	if err != nil {
		fmt.Fprintf(os.Stderr, "invalid REDIS_TEST_URL: %v\n", err)
		os.Exit(1)
	}
	opts.Protocol = 2

	client := redis.NewClient(opts)

	ctx := context.Background()
	if err := client.Ping(ctx).Err(); err != nil {
		fmt.Fprintf(os.Stderr, "failed to connect to Redis: %v\n", err)
		os.Exit(1)
	}

	globalClient = client

	code := m.Run()

	client.Close()
	os.Exit(code)
}

func newTestCache(t *testing.T, prefix string) *RedisCache {
	t.Helper()

	if globalClient == nil {
		t.Skip("REDIS_TEST_URL not set")
	}

	c := NewRedisCache(globalClient, time.Minute, prefix)

	ctx := context.Background()
	_ = globalClient.FlushDB(ctx).Err()

	t.Cleanup(func() {
		_ = globalClient.FlushDB(ctx).Err()
	})

	return c
}

type testStruct struct {
	Name  string `json:"name"`
	Value int    `json:"value"`
}

func TestRedisCache_GetSet(t *testing.T) {
	c := newTestCache(t, "test")
	ctx := context.Background()
	key := "getset"

	want := testStruct{Name: "foo", Value: 42}
	err := c.Set(ctx, key, want, 0)
	if err != nil {
		t.Fatal(err)
	}

	var got testStruct
	err = c.Get(ctx, key, &got)
	if err != nil {
		t.Fatal(err)
	}

	if got != want {
		t.Errorf("got %+v, want %+v", got, want)
	}
}

func TestRedisCache_ErrMiss(t *testing.T) {
	c := newTestCache(t, "test")
	ctx := context.Background()

	var got testStruct
	err := c.Get(ctx, "nonexistent", &got)
	if err != ErrMiss {
		t.Errorf("got %v, want %v", err, ErrMiss)
	}
}

func TestRedisCache_GetSetWithTTL(t *testing.T) {
	c := newTestCache(t, "test")
	ctx := context.Background()
	key := "ttl"

	err := c.Set(ctx, key, "data", 100*time.Millisecond)
	if err != nil {
		t.Fatal(err)
	}

	var got string
	err = c.Get(ctx, key, &got)
	if err != nil {
		t.Fatal(err)
	}
	if got != "data" {
		t.Errorf("got %q, want %q", got, "data")
	}

	time.Sleep(150 * time.Millisecond)

	err = c.Get(ctx, key, &got)
	if err != ErrMiss {
		t.Errorf("expected ErrMiss after TTL, got %v", err)
	}
}

func TestRedisCache_Delete(t *testing.T) {
	c := newTestCache(t, "test")
	ctx := context.Background()

	_ = c.Set(ctx, "del1", "a", 0)
	_ = c.Set(ctx, "del2", "b", 0)

	err := c.Delete(ctx, "del1")
	if err != nil {
		t.Fatal(err)
	}

	var got string
	if err := c.Get(ctx, "del1", &got); err != ErrMiss {
		t.Errorf("expected ErrMiss after delete, got %v", err)
	}

	if err := c.Get(ctx, "del2", &got); err != nil {
		t.Fatal(err)
	}
	if got != "b" {
		t.Errorf("got %q, want %q", got, "b")
	}
}

func TestRedisCache_DeleteMultiple(t *testing.T) {
	c := newTestCache(t, "test")
	ctx := context.Background()

	_ = c.Set(ctx, "a", 1, 0)
	_ = c.Set(ctx, "b", 2, 0)
	_ = c.Set(ctx, "c", 3, 0)

	err := c.Delete(ctx, "a", "b")
	if err != nil {
		t.Fatal(err)
	}

	var got int
	if err := c.Get(ctx, "a", &got); err != ErrMiss {
		t.Errorf("expected ErrMiss, got %v", err)
	}
	if err := c.Get(ctx, "b", &got); err != ErrMiss {
		t.Errorf("expected ErrMiss, got %v", err)
	}
	if err := c.Get(ctx, "c", &got); err != nil {
		t.Fatal(err)
	}
	if got != 3 {
		t.Errorf("got %d, want %d", got, 3)
	}
}

func TestRedisCache_Invalidate(t *testing.T) {
	c := newTestCache(t, "test")
	ctx := context.Background()

	_ = c.Set(ctx, "char:1", 1, 0)
	_ = c.Set(ctx, "char:2", 2, 0)
	_ = c.Set(ctx, "char:10", 10, 0)
	_ = c.Set(ctx, "item:1", 100, 0)

	err := c.Invalidate(ctx, "char:*")
	if err != nil {
		t.Fatal(err)
	}

	var got int
	if err := c.Get(ctx, "char:1", &got); err != ErrMiss {
		t.Errorf("expected ErrMiss, got %v", err)
	}
	if err := c.Get(ctx, "char:2", &got); err != ErrMiss {
		t.Errorf("expected ErrMiss, got %v", err)
	}
	if err := c.Get(ctx, "char:10", &got); err != ErrMiss {
		t.Errorf("expected ErrMiss, got %v", err)
	}
	if err := c.Get(ctx, "item:1", &got); err != nil {
		t.Fatal(err)
	}
	if got != 100 {
		t.Errorf("got %d, want %d", got, 100)
	}
}

func TestRedisCache_InvalidateNoMatch(t *testing.T) {
	c := newTestCache(t, "test")
	ctx := context.Background()

	_ = c.Set(ctx, "a", 1, 0)

	err := c.Invalidate(ctx, "nonexistent:*")
	if err != nil {
		t.Fatal(err)
	}

	var got int
	if err := c.Get(ctx, "a", &got); err != nil {
		t.Fatal(err)
	}
	if got != 1 {
		t.Errorf("got %d, want %d", got, 1)
	}
}

func TestRedisCache_PrefixIsolation(t *testing.T) {
	ca := newTestCache(t, "alpha")
	cb := newTestCache(t, "beta")
	ctx := context.Background()

	_ = ca.Set(ctx, "x", "alpha-val", 0)
	_ = cb.Set(ctx, "x", "beta-val", 0)

	var got string
	if err := ca.Get(ctx, "x", &got); err != nil {
		t.Fatal(err)
	}
	if got != "alpha-val" {
		t.Errorf("got %q, want %q", got, "alpha-val")
	}

	if err := cb.Get(ctx, "x", &got); err != nil {
		t.Fatal(err)
	}
	if got != "beta-val" {
		t.Errorf("got %q, want %q", got, "beta-val")
	}
}

func TestRedisCache_DefaultTTL(t *testing.T) {
	if globalClient == nil {
		t.Skip("REDIS_TEST_URL not set")
	}
	c := NewRedisCache(globalClient, 100*time.Millisecond, "test")
	ctx := context.Background()

	_ = c.Set(ctx, "default-ttl", "data", 0)

	time.Sleep(150 * time.Millisecond)

	var got string
	err := c.Get(ctx, "default-ttl", &got)
	if err != ErrMiss {
		t.Errorf("expected ErrMiss after default TTL, got %v", err)
	}
}

func TestRedisCache_Ping(t *testing.T) {
	c := newTestCache(t, "test")
	ctx := context.Background()

	if err := c.Ping(ctx); err != nil {
		t.Errorf("Ping failed: %v", err)
	}
}

func TestRedisCache_JSONTypes(t *testing.T) {
	c := newTestCache(t, "test")
	ctx := context.Background()

	type nested struct {
		A string `json:"a"`
		B int    `json:"b"`
	}

	want := nested{A: "hello", B: 99}
	err := c.Set(ctx, "nested", want, 0)
	if err != nil {
		t.Fatal(err)
	}

	var got nested
	err = c.Get(ctx, "nested", &got)
	if err != nil {
		t.Fatal(err)
	}

	if got != want {
		t.Errorf("got %+v, want %+v", got, want)
	}
}
