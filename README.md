# Three-Up TTS Generator

A web application that generates three different emotional variations ("Takes") of a given text script and synthesizes them into speech using Google's Gemini TTS.

It demonstrates how to orchestrate:
1. **Gemini Generative AI (`gemini-3-flash-preview`)** to rewrite the prompt strictly inserting emotion/technical voice tags.
2. **Gemini TTS (`gemini-3.1-flash-tts-preview`)** to read the tagged variations with different vocal energies.
3. A Lit Web Component frontend with custom text-tag visualization.

## Features

* **AI Orchestration:** Automates an entire "VO Booth" session, generating an enhanced script and 3 unique emotional variations from a single prompt.
* **Gemini TTS:** Leverages the latest `gemini-3.1-flash-tts-preview` model for high-fidelity voice synthesis.
* **Dual Design Systems:** A frontend built with Lit Web Components, featuring two toggleable aesthetic themes (Synthetix Studio Dark & Sunrise Studio Light).
* **Deep Observability:** Fully instrumented with OpenTelemetry to track exact latency costs of Gemini prompts and audio generation.

## Prerequisites

- **Go 1.25+**
- **Node.js 20+**
- A **Google Cloud Project** with Vertex AI enabled.
- Application Default Credentials (ADC) configured locally (`gcloud auth application-default login`).

## Configuration

Edit `backend/.env` to configure your Google Cloud project details and bucket for audio generation:

```env
GOOGLE_CLOUD_PROJECT=your-project-id
GOOGLE_CLOUD_LOCATION=us-central1
PORT=8080
GENMEDIA_BUCKET=your-bucket-name
GEMINI_MODEL=gemini-3-flash-preview
GEMINI_TTS_MODEL=gemini-3.1-flash-tts-preview
```

### Storage Bucket CORS

The application stores generated TTS audio in a GCS bucket (`GENMEDIA_BUCKET`) and streams it directly to the browser. To allow cross-origin audio playback, you must configure CORS on this bucket so the browser can stream the `206 Partial Content` audio:

```bash
gcloud storage buckets update gs://<your-bucket-name> --cors-file=cors.json
```

## Running Locally

To run the full stack locally during development, you can use the provided Makefile.

**To start everything concurrently (frontend & backend):**
```bash
make dev
```
- The backend API runs on `http://localhost:8080`
- The frontend dev server runs on `http://localhost:5173`
- *The frontend server automatically proxies `/api` requests to the backend.*

**Individual Service Commands:**
- `make dev-server`: Run just the Go backend.
- `make dev-frontend`: Run just the Vite frontend.
- `make build-frontend`: Build the production frontend bundle into `backend/dist`.
- `make build-run`: Build the frontend, place it in the backend's directory, and run the backend. Since the backend is configured to serve static files from `dist/`, the app will be fully available on `http://localhost:8080`.

## Testing

To test functionality, open the frontend (`http://localhost:5173` during `make dev`, or `http://localhost:8080` via `make build-run`), paste a script like:

> "Our kittens are raised in a cage-free environment with 24/7 medical supervision."

Click **Generate Three-Up Takes**. After processing, the UI will display the three variations with inline audio tags (e.g., `[happy]`, `[sarcasm]`) and audio players to listen to the generated Gemini TTS output.

## 💰 Cost Analysis (Estimate)

> **NOTE:** Pricing based on **Gemini 3.1 Flash Lite Preview** and estimated **3.1 Flash TTS Preview** (based on 2.5 Flash TTS) as of **April 2026**.
> *(Source: Google Cloud Vertex AI Pricing)*

### Example Session: Generating 3 Takes

**Context:** A single user request providing a short 50-word script. The system generates an enhanced script, 3 tagged variations, and then synthesizes 3 separate audio files in parallel.

| Operation | Model | Tokens | Rate | Cost Estimate |
| :--- | :--- | :--- | :--- | :--- |
| **Text Gen (Input)** | `gemini-3.1-flash-lite` | ~300 | $0.25 / 1M | ~$0.000075 |
| **Text Gen (Output)** | `gemini-3.1-flash-lite` | ~300 | $1.50 / 1M | ~$0.000450 |
| **TTS Gen (Input)** | `gemini-3.1-flash-tts` | ~300 | $0.50 / 1M | ~$0.000150 |
| **Total Cost** | | **~900** | | **~$0.000675 (< 0.1¢)** |

Generating an entire 3-take orchestrated VO session costs **less than one-tenth of a cent**, making this architecture highly scalable for production use cases.

## Architecture & Design

For a deeper dive into the system architecture, component design, and operational learnings from building with Gemini TTS, please refer to the documentation:

* [Architecture & Operational Learnings](docs/ARCHITECTURE.md)
* [Synthetix Studio Design System (Dark)](docs/DESIGN.md)
* [Sunrise Studio Design System (Light)](docs/DESIGN_SUNRISE.md)

## Deployment

The application is containerized using Docker and is configured for deployment to **Google Cloud Run**.

1. Ensure your `gcloud` CLI is configured and authenticated.
2. Run the deployment script:

```bash
make deploy
# OR
./scripts/deploy.sh
```

This script will:
1. Load environment variables from `backend/.env` (or use defaults).
2. Build the Docker container using `gcloud builds submit`.
3. Deploy the application to Cloud Run with unauthenticated access enabled.

Upon success, `gcloud` will output the public URL of your application.

### Deploying to a New Public Project (No IAP)

If you want to deploy this application publicly to a different Google Cloud project (without Identity-Aware Proxy restrictions), you can configure the deployment script via environment variables.

1.  **Enable Required APIs** in your new project:
    ```bash
    gcloud services enable run.googleapis.com cloudbuild.googleapis.com aiplatform.googleapis.com storage.googleapis.com --project=<NEW_PROJECT_ID>
    ```
2.  **Configure Environment Variables** (either export them in your terminal or create a `.env.deploy` file):
    ```bash
    export PROJECT_ID="<NEW_PROJECT_ID>"
    export USE_IAP="false"                 # Skips IAP setup and uses --allow-unauthenticated
    export SERVICE_NAME="threeup-audio"    # (Optional) Override the default service name
    export GENMEDIA_BUCKET="<NEW_BUCKET>"  # Ensure your target project has this bucket created and CORS-enabled
    ```
3.  **Run the script**:
    ```bash
    ./scripts/deploy.sh
    ```

The script will automatically create the necessary service account in the new project, grant it the required Vertex AI and Storage permissions, build the image, and deploy it to Cloud Run publicly.

## Observability & Tracing (OpenTelemetry)

The Three-Up backend is fully instrumented with **OpenTelemetry (OTel)**, providing deep visibility into the orchestration engine's performance. By default, it exports traces directly to **Google Cloud Trace** when deployed.

### What is Traced?
1. **HTTP Requests:** Every incoming API call (e.g., `/api/variations`, `/api/variation-single`) is tracked from start to finish via the `otelhttp` middleware.
2. **LLM Text Generation (`LLM_Generate_Text`):** Captures the exact latency of the Gemini prompt logic.
3. **TTS Audio Synthesis (`TTS_Generation`):** Each parallel TTS audio request has its own child span. It captures the specific `take`, the `voiceName`, and crucially, the retry `attempt` number if the Vertex API throws a safety block.
4. **Google Cloud Storage (`GCS_Audio_Upload`):** Tracks the final network latency of uploading the generated WAV files back to the bucket.
5. **Downstream Linkage:** Google's internal API network timings are automatically appended as leaf nodes to your traces!

### Testing Traces Locally
To view traces generated from your local machine:
1. Ensure your `GOOGLE_CLOUD_PROJECT` is set in your `.env` file.
2. Authenticate locally with Application Default Credentials:
   `gcloud auth application-default login`
3. Run the backend: `cd backend && go run .`
4. Generate a take in the UI, then navigate to the **Trace** page in your Google Cloud Console. You'll see a beautiful waterfall chart breaking down the exact millisecond cost of every Gemini and GCS interaction.
