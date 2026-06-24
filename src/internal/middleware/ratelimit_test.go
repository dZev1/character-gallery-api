package middleware

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/redis/go-redis/v9"
)

var globalRLClient *redis.Client

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

	globalRLClient = client

	code := m.Run()

	client.Close()
	os.Exit(code)
}

func newTestLimiter(t *testing.T, window time.Duration) *RateLimiter {
	t.Helper()

	if globalRLClient == nil {
		t.Skip("REDIS_TEST_URL not set")
	}

	ctx := context.Background()
	_ = globalRLClient.FlushDB(ctx).Err()

	t.Cleanup(func() {
		_ = globalRLClient.FlushDB(ctx).Err()
	})

	return NewRateLimiter(globalRLClient, window)
}

func okHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte("ok"))
}

func TestRateLimiter_UnderLimit(t *testing.T) {
	rl := newTestLimiter(t, time.Minute)
	rl.SetLimit("GET", "/test", 5)

	req := httptest.NewRequest("GET", "/test", nil)

	for i := 0; i < 5; i++ {
		rr := httptest.NewRecorder()
		rl.Limit(http.HandlerFunc(okHandler)).ServeHTTP(rr, req)

		if rr.Code != http.StatusOK {
			t.Errorf("request %d: got %d, want %d", i+1, rr.Code, http.StatusOK)
		}
	}
}

func TestRateLimiter_OverLimit(t *testing.T) {
	rl := newTestLimiter(t, time.Minute)
	rl.SetLimit("GET", "/test", 3)

	req := httptest.NewRequest("GET", "/test", nil)

	for i := 0; i < 3; i++ {
		rr := httptest.NewRecorder()
		rl.Limit(http.HandlerFunc(okHandler)).ServeHTTP(rr, req)
		if rr.Code != http.StatusOK {
			t.Errorf("request %d: got %d, want %d", i+1, rr.Code, http.StatusOK)
		}
	}

	rr := httptest.NewRecorder()
	rl.Limit(http.HandlerFunc(okHandler)).ServeHTTP(rr, req)
	if rr.Code != http.StatusTooManyRequests {
		t.Errorf("expected 429, got %d", rr.Code)
	}
}

func TestRateLimiter_Headers(t *testing.T) {
	rl := newTestLimiter(t, 5*time.Minute)
	rl.SetLimit("GET", "/test", 10)

	req := httptest.NewRequest("GET", "/test", nil)
	rr := httptest.NewRecorder()
	rl.Limit(http.HandlerFunc(okHandler)).ServeHTTP(rr, req)

	if rr.Header().Get("X-RateLimit-Limit") != "10" {
		t.Errorf("got X-RateLimit-Limit=%q, want %q", rr.Header().Get("X-RateLimit-Limit"), "10")
	}
	if rr.Header().Get("X-RateLimit-Remaining") != "9" {
		t.Errorf("got X-RateLimit-Remaining=%q, want %q", rr.Header().Get("X-RateLimit-Remaining"), "9")
	}
}

func TestRateLimiter_NoLimitConfigured(t *testing.T) {
	rl := newTestLimiter(t, time.Minute)

	req := httptest.NewRequest("GET", "/unlimited", nil)
	rr := httptest.NewRecorder()
	rl.Limit(http.HandlerFunc(okHandler)).ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("got %d, want %d", rr.Code, http.StatusOK)
	}
}

func TestRateLimiter_DifferentRoutes(t *testing.T) {
	rl := newTestLimiter(t, time.Minute)
	rl.SetLimit("GET", "/heavy", 2)
	rl.SetLimit("GET", "/light", 10)

	heavy := httptest.NewRequest("GET", "/heavy", nil)
	light := httptest.NewRequest("GET", "/light", nil)

	for i := 0; i < 2; i++ {
		rr := httptest.NewRecorder()
		rl.Limit(http.HandlerFunc(okHandler)).ServeHTTP(rr, heavy)
		if rr.Code != http.StatusOK {
			t.Errorf("heavy request %d: got %d", i+1, rr.Code)
		}
	}

	rr := httptest.NewRecorder()
	rl.Limit(http.HandlerFunc(okHandler)).ServeHTTP(rr, heavy)
	if rr.Code != http.StatusTooManyRequests {
		t.Errorf("heavy over limit: expected 429, got %d", rr.Code)
	}

	for i := 0; i < 10; i++ {
		rr := httptest.NewRecorder()
		rl.Limit(http.HandlerFunc(okHandler)).ServeHTTP(rr, light)
		if rr.Code != http.StatusOK {
			t.Errorf("light request %d: got %d", i+1, rr.Code)
		}
	}
}

func TestRateLimiter_RetryAfterHeader(t *testing.T) {
	rl := newTestLimiter(t, time.Minute)
	rl.SetLimit("GET", "/test", 1)

	req := httptest.NewRequest("GET", "/test", nil)

	rr := httptest.NewRecorder()
	rl.Limit(http.HandlerFunc(okHandler)).ServeHTTP(rr, req)

	rr2 := httptest.NewRecorder()
	rl.Limit(http.HandlerFunc(okHandler)).ServeHTTP(rr2, req)

	if rr2.Code != http.StatusTooManyRequests {
		t.Fatalf("expected 429, got %d", rr2.Code)
	}

	if rr2.Header().Get("Retry-After") == "" {
		t.Error("expected Retry-After header on 429")
	}
}

func TestRateLimiter_PostAndGetSeparate(t *testing.T) {
	rl := newTestLimiter(t, time.Minute)
	rl.SetLimit("POST", "/resource", 3)
	rl.SetLimit("GET", "/resource", 5)

	post := httptest.NewRequest("POST", "/resource", nil)
	get := httptest.NewRequest("GET", "/resource", nil)

	for i := 0; i < 3; i++ {
		rr := httptest.NewRecorder()
		rl.Limit(http.HandlerFunc(okHandler)).ServeHTTP(rr, post)
		if rr.Code != http.StatusOK {
			t.Errorf("POST %d: got %d", i+1, rr.Code)
		}
	}

	rr := httptest.NewRecorder()
	rl.Limit(http.HandlerFunc(okHandler)).ServeHTTP(rr, post)
	if rr.Code != http.StatusTooManyRequests {
		t.Errorf("POST over limit: expected 429, got %d", rr.Code)
	}

	for i := 0; i < 5; i++ {
		rr := httptest.NewRecorder()
		rl.Limit(http.HandlerFunc(okHandler)).ServeHTTP(rr, get)
		if rr.Code != http.StatusOK {
			t.Errorf("GET %d: got %d", i+1, rr.Code)
		}
	}
}
