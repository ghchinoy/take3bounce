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
		setRequests  bool
		setWindow    bool
		wantRequests int64
		wantWindow   time.Duration
	}{
		{
			name:         "defaults when unset",
			wantRequests: defaultRateLimitRequests,
			wantWindow:   defaultRateLimitWindow,
		},
		{
			name:         "valid values applied",
			requests:     "50",
			window:       "30s",
			setRequests:  true,
			setWindow:    true,
			wantRequests: 50,
			wantWindow:   30 * time.Second,
		},
		{
			name:         "non-numeric requests falls back",
			requests:     "abc",
			setRequests:  true,
			wantRequests: defaultRateLimitRequests,
			wantWindow:   defaultRateLimitWindow,
		},
		{
			name:         "zero requests falls back",
			requests:     "0",
			setRequests:  true,
			wantRequests: defaultRateLimitRequests,
			wantWindow:   defaultRateLimitWindow,
		},
		{
			name:         "negative requests falls back",
			requests:     "-5",
			setRequests:  true,
			wantRequests: defaultRateLimitRequests,
			wantWindow:   defaultRateLimitWindow,
		},
		{
			name:         "malformed window falls back",
			window:       "not-a-duration",
			setWindow:    true,
			wantRequests: defaultRateLimitRequests,
			wantWindow:   defaultRateLimitWindow,
		},
		{
			name:         "non-positive window falls back",
			window:       "0s",
			setWindow:    true,
			wantRequests: defaultRateLimitRequests,
			wantWindow:   defaultRateLimitWindow,
		},
		{
			name:         "valid requests with invalid window",
			requests:     "25",
			window:       "bogus",
			setRequests:  true,
			setWindow:    true,
			wantRequests: 25,
			wantWindow:   defaultRateLimitWindow,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.setRequests {
				t.Setenv("RATE_LIMIT_REQUESTS", tt.requests)
			} else {
				t.Setenv("RATE_LIMIT_REQUESTS", "")
			}
			if tt.setWindow {
				t.Setenv("RATE_LIMIT_WINDOW", tt.window)
			} else {
				t.Setenv("RATE_LIMIT_WINDOW", "")
			}

			rate := parseRateLimitConfig()

			if rate.Limit != tt.wantRequests {
				t.Errorf("Limit = %d, want %d", rate.Limit, tt.wantRequests)
			}
			if rate.Period != tt.wantWindow {
				t.Errorf("Period = %v, want %v", rate.Period, tt.wantWindow)
			}
		})
	}
}
