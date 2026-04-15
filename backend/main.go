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

	texporter "github.com/GoogleCloudPlatform/opentelemetry-operations-go/exporter/trace"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.21.0"

	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
)

// enableCORS is a middleware that injects Access-Control headers to allow
// the frontend (running on Vite dev server or from the same origin) to
// communicate with the Go backend API. It also handles preflight OPTIONS requests.
func enableCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
		w.Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}
		next.ServeHTTP(w, r)
	})
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
	if project == "" || location == "" {
		slog.Error("FATAL: GOOGLE_CLOUD_PROJECT and GOOGLE_CLOUD_LOCATION must be set in the environment")
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

	r.Use(enableCORS)
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
