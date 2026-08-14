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
	"log/slog"
	"net/http"
	"context"
	"os"
	"strings"

	texporter "github.com/GoogleCloudPlatform/opentelemetry-operations-go/exporter/trace"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.21.0"

	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
)

// parseOriginSet parses a comma-separated ALLOWED_ORIGINS value into a set of
// exact origins (scheme+host+port). Whitespace around entries is trimmed and
// empty entries are dropped, so an unset or blank value yields an empty set
// (same-origin-only: no cross-origin caller is ever allowed).
func parseOriginSet(raw string) map[string]struct{} {
	set := make(map[string]struct{})
	for _, o := range strings.Split(raw, ",") {
		o = strings.TrimSpace(o)
		if o != "" {
			set[o] = struct{}{}
		}
	}
	return set
}

// newCORSMiddleware builds the CORS middleware from a fixed set of allowed
// origins (parsed once at startup). The generation API is never exposed with a
// wildcard: Access-Control-Allow-Origin is only ever set to the request's own
// Origin, and only when that origin is in the allowlist.
//
// Same-origin requests carry no Origin header, so they are never blocked and
// need no configuration — the deployed single-service app works with
// ALLOWED_ORIGINS unset. Cross-origin dev (Vite on http://localhost:5173) works
// by listing that origin in ALLOWED_ORIGINS.
func newCORSMiddleware(allowed map[string]struct{}) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// The response varies by Origin regardless of whether the origin
			// is allowlisted, so emit Vary: Origin on every response the
			// middleware handles for correct HTTP caching semantics.
			w.Header().Add("Vary", "Origin")
			origin := r.Header.Get("Origin")
			if origin != "" {
				if _, ok := allowed[origin]; ok {
					w.Header().Set("Access-Control-Allow-Origin", origin)
					w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS")
					w.Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type, Content-Length, Accept-Encoding, Authorization")
				}
			}
			if r.Method == http.MethodOptions {
				// Preflight: 204 No Content. CORS headers above are present only
				// when the origin was allowlisted.
				w.WriteHeader(http.StatusNoContent)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

// main initializes the Go web server. It loads the environment configuration,
// validates required Google Cloud credentials, registers the API endpoints,
// and starts serving both the static frontend assets and dynamic API routes.

func initTracer(ctx context.Context, projectID string) (*sdktrace.TracerProvider, error) {
	exporter, err := texporter.New(texporter.WithProjectID(projectID))
	if err != nil {
		return nil, err
	}
	res, err := resource.New(ctx,
		resource.WithAttributes(
			semconv.ServiceName("threeup-api"),
		),
	)
	if err != nil {
		return nil, err
	}
	tp := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exporter),
		sdktrace.WithResource(res),
	)
	otel.SetTracerProvider(tp)
	return tp, nil
}

func main() {
	_ = godotenv.Load()

	project := os.Getenv("GOOGLE_CLOUD_PROJECT")
	location := os.Getenv("GOOGLE_CLOUD_LOCATION")
	bucket := os.Getenv("GENMEDIA_BUCKET")
	if project == "" || location == "" || bucket == "" {
		slog.Error("FATAL: GOOGLE_CLOUD_PROJECT, GOOGLE_CLOUD_LOCATION, and GENMEDIA_BUCKET must be set in the environment")
		os.Exit(1)
	}

	ctx := context.Background()
	tp, err := initTracer(ctx, project)
	if err != nil {
		slog.Error("Failed to initialize tracer", "error", err)
		os.Exit(1)
	}
	defer tp.Shutdown(ctx)

	r := mux.NewRouter()
	r.HandleFunc("/api/variations", handleVariations).Methods("POST", "OPTIONS")
	r.HandleFunc("/api/retry-audio", handleRetryAudio).Methods("POST", "OPTIONS")
	r.HandleFunc("/api/variation-single", handleGenerateOne).Methods("POST", "OPTIONS")
	r.HandleFunc("/api/status", handleStatus).Methods("GET", "OPTIONS")
	r.HandleFunc("/api/track", handleTrack).Methods("POST", "OPTIONS")

	// Serve static files from the frontend build
	r.PathPrefix("/").Handler(http.FileServer(http.Dir("./dist")))

	allowedOrigins := parseOriginSet(os.Getenv("ALLOWED_ORIGINS"))
	if len(allowedOrigins) == 0 {
		slog.Info("CORS: no ALLOWED_ORIGINS set; same-origin only (no cross-origin callers allowed)")
	} else {
		slog.Info("CORS: cross-origin allowlist configured", "count", len(allowedOrigins))
	}
	r.Use(newCORSMiddleware(allowedOrigins))
	r.Use(rateLimitMiddleware)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	slog.Info("Backend server listening", "port", port)
	tracedHandler := otelhttp.NewHandler(r, "threeup-api")
	if err := http.ListenAndServe(":"+port, tracedHandler); err != nil {
		slog.Error("Server failed", "error", err)
		os.Exit(1)
	}
}
