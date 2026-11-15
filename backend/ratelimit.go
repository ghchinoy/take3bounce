// Copyright 2026 Google LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

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
		n, err := strconv.ParseInt(raw, 10, 64)
		if err != nil || n <= 0 {
			slog.Warn("Invalid RATE_LIMIT_REQUESTS, falling back to default",
				"value", raw, "default", defaultRateLimitRequests)
		}
		if err == nil && n > 0 {
			requests = n
		}
	}

	window := defaultRateLimitWindow
	if raw := os.Getenv("RATE_LIMIT_WINDOW"); raw != "" {
		d, err := time.ParseDuration(raw)
		if err != nil || d <= 0 {
			slog.Warn("Invalid RATE_LIMIT_WINDOW, falling back to default",
				"value", raw, "default", defaultRateLimitWindow)
		}
		if err == nil && d > 0 {
			window = d
		}
	}

	return limiter.Rate{Period: window, Limit: requests}
}

// newLimiterStore initializes a Redis-backed rate-limiter store when redisURL
// is configured and reachable, falling back to an in-memory store otherwise.
func newLimiterStore(ctx context.Context, redisURL string) limiter.Store {
	if redisURL == "" {
		slog.Info("REDIS_URL not set: Using memory storage for rate limiting (Not suitable for multi-instance)")
		return sstore.NewStore()
	}

	opt, err := redis.ParseURL(redisURL)
	if err != nil {
		slog.Error("Failed to parse REDIS_URL, falling back to memory store", "error", err)
		return sstore.NewStore()
	}

	if opt.DialTimeout == 0 {
		opt.DialTimeout = 1500 * time.Millisecond
	}
	if opt.ReadTimeout == 0 {
		opt.ReadTimeout = 1500 * time.Millisecond
	}
	if opt.WriteTimeout == 0 {
		opt.WriteTimeout = 1500 * time.Millisecond
	}
	opt.MaxRetries = 1

	client := redis.NewClient(opt)
	pingCtx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()
	if err := client.Ping(pingCtx).Err(); err != nil {
		_ = client.Close()
		slog.Error("Failed to connect to Redis, falling back to memory store", "error", err)
		return sstore.NewStore()
	}

	slog.Info("REDIS_URL set: Initializing distributed Redis rate limiter")
	store, err := sredis.NewStoreWithOptions(client, limiter.StoreOptions{
		Prefix: "take3bounce_limit",
	})
	if err != nil {
		_ = client.Close()
		slog.Error("Failed to create Redis store, falling back to memory store", "error", err)
		return sstore.NewStore()
	}
	return store
}

// rateLimiter lazily initializes and returns the singleton HTTP rate-limiter middleware.
func rateLimiter(ctx context.Context) *stdlib.Middleware {
	limiterOnce.Do(func() {
		rate := parseRateLimitConfig()
		slog.Info("Rate limiter configured", "requests", rate.Limit, "window", rate.Period)

		store := newLimiterStore(ctx, os.Getenv("REDIS_URL"))
		instance := limiter.New(store, rate, limiter.WithTrustForwardHeader(true))
		limiterMiddleware = stdlib.NewMiddleware(instance)
	})
	return limiterMiddleware
}

// rateLimitMiddleware enforces rate limits on all /api routes except /api/status.
func rateLimitMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasPrefix(r.URL.Path, "/api") && r.URL.Path != "/api/status" {
			handler := rateLimiter(r.Context()).Handler(next)
			handler.ServeHTTP(w, r)
			return
		}
		next.ServeHTTP(w, r)
	})
}
