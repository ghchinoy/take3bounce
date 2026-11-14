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

// TelemetryRow represents a single analytics event row written to BigQuery.
type TelemetryRow struct {
	Timestamp time.Time         `bigquery:"timestamp"`
	DemoName  string            `bigquery:"demo_name"`
	EventType string            `bigquery:"event_type"`
	Metadata  bigquery.NullJSON `bigquery:"metadata"` // JSON encoded string of metadata
}

// initBQClient lazily initializes the shared BigQuery client when BQ_DATASET
// and BQ_TABLE are configured.
func initBQClient(ctx context.Context) *bigquery.Client {
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

// LogTelemetryEvent asynchronously logs an event to BigQuery or slogs it if BigQuery is disabled.
func LogTelemetryEvent(ctx context.Context, eventType string, metadata map[string]any) {
	baseCtx := context.WithoutCancel(ctx)
	go func() {
		insertCtx, cancel := context.WithTimeout(baseCtx, 10*time.Second)
		defer cancel()

		client := initBQClient(insertCtx)

		metadataJSON, err := json.Marshal(metadata)
		if err != nil {
			metadataJSON = []byte("{}")
		}

		if client == nil {
			slog.Info("Telemetry Event (Local)", "event_type", eventType, "metadata", string(metadataJSON))
			return
		}

		row := &TelemetryRow{
			Timestamp: time.Now().UTC(),
			DemoName:  demoName,
			EventType: eventType,
			Metadata:  bigquery.NullJSON{JSONVal: string(metadataJSON), Valid: true},
		}

		inserter := client.Dataset(datasetID).Table(tableID).Inserter()
		if err := inserter.Put(insertCtx, []*TelemetryRow{row}); err != nil {
			slog.Error("BigQuery Insert Error", "error", err)
			return
		}
		slog.Info("Logged event to BigQuery", "event_type", eventType)
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
		LogTelemetryEvent(r.Context(), req.EventType, req.Metadata)
	}
	w.WriteHeader(http.StatusOK)
}
