package main

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"os"
	"sync"
	"time"

	"cloud.google.com/go/bigquery"
)

var (
	bqClient     *bigquery.Client
	bqClientOnce sync.Once
	datasetID    string
	tableID      string
	demoName     string
)

type TelemetryRow struct {
	Timestamp time.Time `bigquery:"timestamp"`
	DemoName  string    `bigquery:"demo_name"`
	EventType string    `bigquery:"event_type"`
	Metadata  bigquery.NullJSON `bigquery:"metadata"` // JSON encoded string of metadata
}

func getBQClient() *bigquery.Client {
	bqClientOnce.Do(func() {
		datasetID = os.Getenv("BQ_DATASET")
		tableID = os.Getenv("BQ_TABLE")
		demoName = os.Getenv("DEMO_NAME")
		if demoName == "" {
			demoName = "take3bounce"
		}

		if datasetID == "" || tableID == "" {
			slog.Info("BQ_DATASET or BQ_TABLE not set: Skipping BigQuery telemetry initialization")
			return
		}

		projectID := os.Getenv("GOOGLE_CLOUD_PROJECT")
		ctx := context.Background()
		client, err := bigquery.NewClient(ctx, projectID)
		if err != nil {
			slog.Error("Warning: Could not initialize BigQuery client", "error", err)
			return
		}
		bqClient = client
		slog.Info("BigQuery telemetry initialized", "dataset", datasetID, "table", tableID)
	})
	return bqClient
}

// LogTelemetryEvent asynchronously logs an event to BigQuery or slogs it if BQ is disabled.
func LogTelemetryEvent(eventType string, metadata map[string]interface{}) {
	go func() {
		client := getBQClient()

		metadataJSON, err := json.Marshal(metadata)
		if err != nil {
			metadataJSON = []byte("{}")
		}

		if client == nil {
			slog.Info("Telemetry Event (Local)", "event_type", eventType, "metadata", string(metadataJSON))
			return
		}

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		row := &TelemetryRow{
			Timestamp: time.Now().UTC(),
			DemoName:  demoName,
			EventType: eventType,
			Metadata:  bigquery.NullJSON{JSONVal: string(metadataJSON), Valid: true},
		}

		inserter := client.Dataset(datasetID).Table(tableID).Inserter()
		if err := inserter.Put(ctx, []*TelemetryRow{row}); err != nil {
			slog.Error("BigQuery Insert Error", "error", err)
		} else {
			slog.Info("Logged event to BigQuery", "event_type", eventType)
		}
	}()
}

// handleTrack receives generic frontend events and logs them to BigQuery.
func handleTrack(w http.ResponseWriter, r *http.Request) {
	var req TrackRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	if req.EventType != "" {
		LogTelemetryEvent(req.EventType, req.Metadata)
	}
	w.WriteHeader(http.StatusOK)
}
