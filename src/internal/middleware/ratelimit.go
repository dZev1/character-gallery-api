package middleware

import (
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/redis/go-redis/v9"
)

type RateLimiter struct {
	mu     sync.RWMutex
	client *redis.Client
	limits map[string]int
	window time.Duration
}

func NewRateLimiter(client *redis.Client, window time.Duration) *RateLimiter {
	return &RateLimiter{
		client: client,
		limits: make(map[string]int),
		window: window,
	}
}

func (rl *RateLimiter) SetLimit(method, path string, max int) {
	rl.mu.Lock()
	defer rl.mu.Unlock()
	rl.limits[method+" "+path] = max
}

func (rl *RateLimiter) Limit(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		route := r.Method + " " + strings.TrimSuffix(r.URL.Path, "/")

		rl.mu.RLock()
		max, ok := rl.limits[route]
		rl.mu.RUnlock()

		if !ok {
			next.ServeHTTP(w, r)
			return
		}

		ip := r.RemoteAddr
		if idx := strings.LastIndex(ip, ":"); idx != -1 {
			ip = ip[:idx]
		}

		now := time.Now().UnixMilli() / int64(rl.window.Milliseconds())
		key := fmt.Sprintf("ratelimit:%s:%s:%d", ip, route, now)

		pipe := rl.client.Pipeline()
		incr := pipe.Incr(r.Context(), key)
		pipe.Expire(r.Context(), key, rl.window*2)
		_, err := pipe.Exec(r.Context())
		if err != nil {
			next.ServeHTTP(w, r)
			return
		}

		count := incr.Val()

		remaining := int64(max) - count
		if remaining < 0 {
			remaining = 0
		}

		w.Header().Set("X-RateLimit-Limit", fmt.Sprintf("%d", max))
		w.Header().Set("X-RateLimit-Remaining", fmt.Sprintf("%d", remaining))

		if count > int64(max) {
			w.Header().Set("Retry-After", fmt.Sprintf("%.0f", rl.window.Seconds()))
			http.Error(w, "429 too many requests", http.StatusTooManyRequests)
			return
		}

		next.ServeHTTP(w, r)
	})
}
