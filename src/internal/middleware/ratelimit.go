package middleware

import (
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/redis/go-redis/v9"
)

type rateLimitRule struct {
	method   string
	segments []string
	max      int
}

func (r *rateLimitRule) match(method, path string) (int, bool) {
	if r.method != method {
		return 0, false
	}
	segments := strings.Split(strings.Trim(path, "/"), "/")
	if len(segments) != len(r.segments) {
		return 0, false
	}
	for i, seg := range r.segments {
		if len(seg) > 0 && seg[0] == '{' && seg[len(seg)-1] == '}' {
			continue
		}
		if seg != segments[i] {
			return 0, false
		}
	}
	return r.max, true
}

type RateLimiter struct {
	mu     sync.RWMutex
	client *redis.Client
	rules  []rateLimitRule
	window time.Duration
}

func NewRateLimiter(client *redis.Client, window time.Duration) *RateLimiter {
	return &RateLimiter{
		client: client,
		rules:  make([]rateLimitRule, 0),
		window: window,
	}
}

func (rl *RateLimiter) SetLimit(method, path string, max int) {
	rl.mu.Lock()
	defer rl.mu.Unlock()
	rl.rules = append(rl.rules, rateLimitRule{
		method:   method,
		segments: strings.Split(strings.Trim(path, "/"), "/"),
		max:      max,
	})
}

func (rl *RateLimiter) Limit(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		path := strings.TrimSuffix(r.URL.Path, "/")

		rl.mu.RLock()
		var max int
		var ok bool
		for _, rule := range rl.rules {
			max, ok = rule.match(r.Method, path)
			if ok {
				break
			}
		}
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
		key := fmt.Sprintf("ratelimit:%s:%s:%d", ip, r.Method+" "+path, now)

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
