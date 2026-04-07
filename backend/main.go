package main

import (
	"log/slog"
	"net/http"
	"os"

	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
)

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

func main() {
	_ = godotenv.Load()

	project := os.Getenv("GOOGLE_CLOUD_PROJECT")
	location := os.Getenv("GOOGLE_CLOUD_LOCATION")
	if project == "" || location == "" {
		slog.Error("FATAL: GOOGLE_CLOUD_PROJECT and GOOGLE_CLOUD_LOCATION must be set in the environment")
		os.Exit(1)
	}

	r := mux.NewRouter()
	r.HandleFunc("/api/variations", handleVariations).Methods("POST", "OPTIONS")

	// Serve static files from the frontend build
	r.PathPrefix("/").Handler(http.FileServer(http.Dir("./dist")))

	r.Use(enableCORS)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	slog.Info("Backend server listening", "port", port)
}
