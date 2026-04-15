package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/ulule/limiter/v3"
	"github.com/ulule/limiter/v3/drivers/middleware/stdlib"
	sstore "github.com/ulule/limiter/v3/drivers/store/memory"
	sredis "github.com/ulule/limiter/v3/drivers/store/redis"
)

var (
	limiterMiddleware *stdlib.Middleware
	limiterOnce       sync.Once
)

func getRateLimiter() *stdlib.Middleware {
	limiterOnce.Do(func() {
		rate := limiter.Rate{
			Period: 1 * time.Minute,
			Limit:  10,
		}

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
