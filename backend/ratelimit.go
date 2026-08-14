package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/ulule/limiter/v3"
	"github.com/ulule/limiter/v3/drivers/middleware/stdlib"
	sstore "github.com/ulule/limiter/v3/drivers/store/memory"
	sredis "github.com/ulule/limiter/v3/drivers/store/redis"
)

const (
	defaultRateLimitRequests int64         = 10
	defaultRateLimitWindow   time.Duration = 1 * time.Minute
)

var (
	limiterMiddleware *stdlib.Middleware
	limiterOnce       sync.Once
)

// parseRateLimitConfig builds the limiter rate from the RATE_LIMIT_REQUESTS and
// RATE_LIMIT_WINDOW environment variables. Unset, malformed, or non-positive
// values fall back to the defaults (10 requests / 1m) with a logged warning; it
// never panics.
func parseRateLimitConfig() limiter.Rate {
	requests := defaultRateLimitRequests
	if raw := os.Getenv("RATE_LIMIT_REQUESTS"); raw != "" {
		if n, err := strconv.ParseInt(raw, 10, 64); err != nil || n <= 0 {
			slog.Warn("Invalid RATE_LIMIT_REQUESTS, falling back to default",
				"value", raw, "default", defaultRateLimitRequests)
		} else {
			requests = n
		}
	}

	window := defaultRateLimitWindow
	if raw := os.Getenv("RATE_LIMIT_WINDOW"); raw != "" {
		if d, err := time.ParseDuration(raw); err != nil || d <= 0 {
			slog.Warn("Invalid RATE_LIMIT_WINDOW, falling back to default",
				"value", raw, "default", defaultRateLimitWindow)
		} else {
			window = d
		}
	}

	return limiter.Rate{Period: window, Limit: requests}
}

func getRateLimiter() *stdlib.Middleware {
	limiterOnce.Do(func() {
		rate := parseRateLimitConfig()
		slog.Info("Rate limiter configured", "requests", rate.Limit, "window", rate.Period)

		var store limiter.Store
		redisURL := os.Getenv("REDIS_URL")

		if redisURL == "" {
			slog.Info("REDIS_URL not set: Using memory storage for rate limiting (Not suitable for multi-instance)")
			store = sstore.NewStore()
		} else {
			opt, err := redis.ParseURL(redisURL)
			if err != nil {
				slog.Error("Failed to parse REDIS_URL, falling back to memory store", "error", err)
				store = sstore.NewStore()
			} else {
				client := redis.NewClient(opt)
				if err := client.Ping(context.Background()).Err(); err != nil {
					slog.Error("Failed to connect to Redis, falling back to memory store", "error", err)
					store = sstore.NewStore()
				} else {
					slog.Info("REDIS_URL set: Initializing distributed Redis rate limiter")
					var storeErr error
					store, storeErr = sredis.NewStoreWithOptions(client, limiter.StoreOptions{
						Prefix: "take3bounce_limit",
					})
					if storeErr != nil {
						slog.Error("Failed to create Redis store, falling back to memory store", "error", storeErr)
						store = sstore.NewStore()
					}
				}
			}
		}

		instance := limiter.New(store, rate, limiter.WithTrustForwardHeader(true))
		limiterMiddleware = stdlib.NewMiddleware(instance)
	})
	return limiterMiddleware
}

func rateLimitMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasPrefix(r.URL.Path, "/api") && r.URL.Path != "/api/status" {
			handler := getRateLimiter().Handler(next)
			handler.ServeHTTP(w, r)
			return
		}
		next.ServeHTTP(w, r)
	})
}
