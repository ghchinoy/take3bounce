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
	"testing"
	"time"
)

// TestParseRateLimitConfig verifies that valid env values are applied and that
// unset/invalid values fall back to the 10 requests / 1m defaults without
// panicking.
func TestParseRateLimitConfig(t *testing.T) {
	tests := []struct {
		name         string
		requests     string // "" means unset
		window       string // "" means unset
		wantRequests int64
		wantWindow   time.Duration
	}{
		{
			name:         "defaults_when_unset",
			wantRequests: defaultRateLimitRequests,
			wantWindow:   defaultRateLimitWindow,
		},
		{
			name:         "valid_values_applied",
			requests:     "50",
			window:       "30s",
			wantRequests: 50,
			wantWindow:   30 * time.Second,
		},
		{
			name:         "non_numeric_requests_falls_back",
			requests:     "abc",
			wantRequests: defaultRateLimitRequests,
			wantWindow:   defaultRateLimitWindow,
		},
		{
			name:         "zero_requests_falls_back",
			requests:     "0",
			wantRequests: defaultRateLimitRequests,
			wantWindow:   defaultRateLimitWindow,
		},
		{
			name:         "negative_requests_falls_back",
			requests:     "-5",
			wantRequests: defaultRateLimitRequests,
			wantWindow:   defaultRateLimitWindow,
		},
		{
			name:         "malformed_window_falls_back",
			window:       "not-a-duration",
			wantRequests: defaultRateLimitRequests,
			wantWindow:   defaultRateLimitWindow,
		},
		{
			name:         "non_positive_window_falls_back",
			window:       "0s",
			wantRequests: defaultRateLimitRequests,
			wantWindow:   defaultRateLimitWindow,
		},
		{
			name:         "valid_requests_with_invalid_window",
			requests:     "25",
			window:       "bogus",
			wantRequests: 25,
			wantWindow:   defaultRateLimitWindow,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv("RATE_LIMIT_REQUESTS", tt.requests)
			t.Setenv("RATE_LIMIT_WINDOW", tt.window)

			rate := parseRateLimitConfig()

			if rate.Limit != tt.wantRequests {
				t.Errorf("parseRateLimitConfig(requests=%q, window=%q).Limit = %d, want %d", tt.requests, tt.window, rate.Limit, tt.wantRequests)
			}
			if rate.Period != tt.wantWindow {
				t.Errorf("parseRateLimitConfig(requests=%q, window=%q).Period = %v, want %v", tt.requests, tt.window, rate.Period, tt.wantWindow)
			}
		})
	}
}

// TestNewLimiterStoreFallback verifies that empty, malformed, and unreachable
// Redis URLs fall back cleanly and quickly to the in-memory store without
// leaking background connection pool retries.
func TestNewLimiterStoreFallback(t *testing.T) {
	tests := []struct {
		name     string
		redisURL string
	}{
		{
			name:     "empty_url_uses_memory_store",
			redisURL: "",
		},
		{
			name:     "malformed_url_falls_back_to_memory_store",
			redisURL: "://not-a-valid-redis-url",
		},
		{
			name:     "unreachable_redis_falls_back_fast_and_closes_client",
			redisURL: "redis://127.0.0.1:1/0?dial_timeout=100ms",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			start := time.Now()
			store := newLimiterStore(context.Background(), tt.redisURL)
			elapsed := time.Since(start)

			if store == nil {
				t.Fatalf("newLimiterStore(%q) returned nil store", tt.redisURL)
			}
			if elapsed > 3*time.Second {
				t.Errorf("newLimiterStore(%q) took %v, want <= 3s", tt.redisURL, elapsed)
			}
		})
	}
}
