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
	"encoding/json"
	"log/slog"
	"net/http"
	"os"
)

// StatusResponse reports the runtime configuration mode and optional integrations
// (ReCaptcha Enterprise and Redis rate limiting) to the frontend.
type StatusResponse struct {
	Mode             string   `json:"mode"`
	Missing          []string `json:"missing"`
	RecaptchaSiteKey string   `json:"recaptcha_site_key,omitempty"`
	RedisConfigured  bool     `json:"redis_configured"`
}

// handleStatus returns the current server operating mode and configured security integrations.
func handleStatus(w http.ResponseWriter, _ *http.Request) {
	var missing []string
	siteKey := os.Getenv("RECAPTCHA_SITE_KEY")
	if siteKey == "" {
		missing = append(missing, "recaptcha")
	}

	redisURL := os.Getenv("REDIS_URL")
	if redisURL == "" {
		missing = append(missing, "redis")
	}

	mode := "production"
	if len(missing) > 0 {
		mode = "simple"
	}

	resp := StatusResponse{
		Mode:             mode,
		Missing:          missing,
		RecaptchaSiteKey: siteKey,
		RedisConfigured:  redisURL != "",
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		slog.Error("Failed to encode status response", "error", err)
	}
}
