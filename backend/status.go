package main

import (
	"encoding/json"
	"net/http"
	"os"
)

type StatusResponse struct {
	Mode             string   `json:"mode"`
	Missing          []string `json:"missing"`
	RecaptchaSiteKey string   `json:"recaptcha_site_key,omitempty"`
	RedisConfigured  bool     `json:"redis_configured"`
}

func handleStatus(w http.ResponseWriter, r *http.Request) {
	var missing []string
	siteKey := os.Getenv("RECAPTCHA_SITE_KEY")

	if siteKey == "" {
		missing = append(missing, "recaptcha")
	}
	
	redisUrl := os.Getenv("REDIS_URL")
	if redisUrl == "" {
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
		RedisConfigured:  redisUrl != "",
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}
