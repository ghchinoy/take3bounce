# Take3Bounce Dashboards & Analytics

This document outlines the setup, expected outcomes, and analyses for the two primary observability and analytics platforms used by Take3Bounce: **Looker Studio (Business Intelligence)** and **Google Cloud Observability (Operational Health)**.

---

## 1. Looker Studio (Business Intelligence & Marketing)

Looker Studio connects directly to BigQuery to visualize user behavior, adoption rates, and generation trends. This dashboard is intended for product managers and marketing teams.

### Setup Instructions

1. **Ensure BigQuery is Populated:** The app must be deployed with `BQ_DATASET` and `BQ_TABLE` environment variables configured.
2. **Create the Unpacked View:** In the BigQuery Console, create a view to parse the JSON metadata and handle timezone conversions automatically:
   ```sql
   CREATE OR REPLACE VIEW `[PROJECT_ID].[DATASET_ID].generations_unpacked` AS
   SELECT 
     DATETIME(timestamp, "America/Los_Angeles") as local_time,
     timestamp as utc_time,
     demo_name,
     event_type,
     JSON_VALUE(metadata.model) as ai_model,
     JSON_VALUE(metadata.voiceActor) as voice_actor,
     CAST(JSON_VALUE(metadata.textLength) AS INT64) as text_length,
     CAST(JSON_VALUE(metadata.takeCount) AS INT64) as takes_generated,
     JSON_VALUE(metadata.location) as click_location
   FROM `[PROJECT_ID].[DATASET_ID].generations`
   ```
3. **Connect Looker Studio:** Add the `generations_unpacked` View as a Data Source in Looker Studio.
4. **Create Legend Translations (Optional):** Create a Calculated Field in Looker Studio to format the raw `event_type` strings:
   ```sql
   CASE 
     WHEN event_type = "threeup_generated" THEN "Three-Up Orchestrator"
     WHEN event_type = "oneup_generated" THEN "One-Up Generator"
     WHEN event_type = "retry_audio_generated" THEN "Audio Sandbox & Showcase"
     WHEN event_type = "deploy_button_clicked" THEN "Deployed to Cloud Run"
     ELSE event_type 
   END
   ```

### Analyses & Expected Outcomes

*   **Feature Adoption (Pie Chart):** Breakdown of event types (3-Up vs 1-Up). Tells the story of which tools users are finding most valuable.
*   **Most Popular Voice Actor (Bar Chart):** Count of `takes_generated` grouped by `voice_actor`. Highlights user preferences to guide future preset additions.
*   **Average Script Length (Scorecard):** Average of `text_length`. Indicates whether users are testing short snippets or processing full production scripts (useful for cost projections).
*   **Conversion Funnel (Scorecard):** Count of `deploy_button_clicked` events. Shows the conversion rate of open-source adoption from the public demo.

---

## 2. Google Cloud Observability (Operational Health)

Cloud Observability (Cloud Monitoring & Cloud Trace) provides deep technical insights into application performance, scaling, and API latency. This dashboard is intended for engineering and DevOps.

### Setup Instructions

1. **Enable OpenTelemetry:** The application natively exports OTel spans to Cloud Trace.
2. **Create the TTS Latency Metric:**
   - Go to **Logging > Log-based Metrics** in the Google Cloud Console.
   - Click **Create Metric** and choose **Distribution**.
   - Name it `tts_generation_latency`.
   - **Filter:**
     ```text
     logName="projects/[PROJECT_ID]/logs/cloudaudit.googleapis.com%2Factivity" OR resource.type="cloud_run_revision"
     jsonPayload.spanName="TTS_Generation"
     ```
   - **Extractor:** `EXTRACT(jsonPayload.latency)`
3. **Build the Custom Dashboard:**
   - Go to **Monitoring > Dashboards** and create a new dashboard for the `take3` Cloud Run service.

### Analyses & Expected Outcomes

*   **Total Traffic & Error Rates (Line Chart):** Groups HTTP response codes (2xx, 4xx, 5xx). Essential for spotting 429 Rate Limit spikes or 403 ReCaptcha failures.
*   **TTS Generation Latency (Line Chart):** Uses the `tts_generation_latency` log-based metric (p50, p90, p99 percentiles). Identifies bottlenecks or slowdowns originating from the Gemini API.
*   **Instance Count (Stacked Area Chart):** Tracks Cloud Run auto-scaling. Shows the relationship between traffic spikes and container spin-ups.
*   **CPU / Memory Utilization (Line Chart):** Ensures the Go backend is efficiently handling concurrent TTS generation requests without resource starvation.

