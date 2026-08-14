# Three-Up TTS Generator


[![Google Gemini](https://img.shields.io/badge/Google%20Gemini-8E75B2?style=flat&logo=google&logoColor=white)](https://deepmind.google/technologies/gemini/)

## ⚡️ Quick Deployment

Launch your own private instance of the Three-Up Orchestrator to Google Cloud in just one click. The deployment script will automatically provision your Storage Bucket, configure CORS, and set up your IAM roles.

[![Run on Google Cloud](https://deploy.cloud.run/button.svg)](https://deploy.cloud.run/?git_repo=https://github.com/ghchinoy/take3bounce&utm_source=github&utm_medium=unpaidsoc&utm_campaign=FY-Q1-global-cloud-ai-starter-apps&utm_content=take3bounce&utm_term=-)

---

A web application that generates three different emotional variations ("Takes") of a given text script and synthesizes them into speech using Google's Gemini TTS.

It demonstrates how to orchestrate:
1. **Gemini Generative AI (`gemini-3.5-flash-lite`)** to rewrite the prompt strictly inserting emotion/technical voice tags.
2. **Gemini TTS (`gemini-3.1-flash-tts-preview`)** to read the tagged variations with different vocal energies.
3. A Lit Web Component frontend with custom text-tag visualization.

## Features

* **AI Orchestration:** Automates an entire "VO Booth" session, generating an enhanced script and 3 unique emotional variations from a single prompt.
* **Gemini TTS:** Leverages the latest `gemini-3.1-flash-tts-preview` model for high-fidelity voice synthesis.
* **Dual Design Systems:** A frontend built with Lit Web Components, featuring two toggleable aesthetic themes (Synthetix Studio Dark & Sunrise Studio Light).
* **Deep Observability:** Fully instrumented with OpenTelemetry to track exact latency costs of Gemini prompts and audio generation.

## Prerequisites

- **Go 1.26+**
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
GEMINI_MODEL=gemini-3.5-flash-lite
GEMINI_TTS_MODEL=gemini-3.1-flash-tts-preview
# Optional: alternate TTS model used only on the final retry. Leave unset to
# retry the primary TTS model instead of swapping to a different one.
#TTS_FALLBACK_MODEL=
# Optional: comma-separated exact origins (scheme+host+port) allowed to call the
# API cross-origin. Leave unset for same-origin only (the deployed single-service
# app works with no value). No wildcard is ever used. For local cross-origin dev
# set it to your app URL plus the Vite dev server.
#ALLOWED_ORIGINS=https://your-app.run.app,http://localhost:5173
```

### API CORS (`ALLOWED_ORIGINS`)

The generation API (`/api/*`) uses an env-driven origin **allowlist** — never a wildcard. `Access-Control-Allow-Origin` is only ever set to the request's own `Origin`, and only when that origin appears in `ALLOWED_ORIGINS`. Same-origin requests (the production single-service topology, where the UI and API share one Cloud Run URL) carry no `Origin` header and are never blocked, so the deployed app works with `ALLOWED_ORIGINS` unset. Set `ALLOWED_ORIGINS=http://localhost:5173` for the Vite dev server.

### Storage Bucket CORS

The application stores generated TTS audio in a GCS bucket (`GENMEDIA_BUCKET`) and streams it directly to the browser. Audio **playback** (`<audio>` element) is not CORS-gated, but the **Download** button performs a cross-origin `fetch()` of the file, which the bucket's CORS policy *does* gate. The bucket allowlist is derived from the same `ALLOWED_ORIGINS` value used by the API, so the Download button works for exactly the origins allowed to call the API.

> **Note:** For the Download button to work in production, `ALLOWED_ORIGINS` must include your app's public URL (you set this for the API anyway). Playback works regardless.

The deploy flow generates the bucket CORS config from `ALLOWED_ORIGINS` via `scripts/gen-cors.sh` and applies it automatically. To apply it manually (falls back to the committed `cors.json` dev default — `http://localhost:5173` — when `ALLOWED_ORIGINS` is empty):

```bash
ALLOWED_ORIGINS="https://your-app.run.app" scripts/gen-cors.sh > /tmp/cors.json
gcloud storage buckets update gs://<your-bucket-name> --cors-file=/tmp/cors.json
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

> **NOTE:** The figures below are an illustrative estimate for the **Gemini Flash-Lite text tier** and **Gemini Flash TTS Preview**. Verify current rates against the linked sources before relying on them.
> *(Sources: [Vertex AI Pricing](https://cloud.google.com/vertex-ai/generative-ai/pricing), [Cloud Text-to-Speech Pricing](https://cloud.google.com/text-to-speech/pricing?hl=en))**

### Pricing Rates (illustrative)
- **Gemini Flash-Lite (Text)**: $0.25/1M (Input), $1.50/1M (Output)
- **Gemini Flash TTS (Audio)**: $1.00/1M (Text Input), $20.00/1M (Audio Output tokens)
- *Note: Audio is billed at 200 tokens per second of generated speech.*


### Example Session: Generating 3 Takes

**Context:** A single user request providing a short 50-word script. The system generates an enhanced script, 3 tagged variations, and then synthesizes 3 separate audio files in parallel.

| Operation | Model | Tokens | Rate | Cost Estimate |
| :--- | :--- | :--- | :--- | :--- |
| **Text Gen (Input)** | `gemini-3.5-flash-lite` | ~450 | $0.25 / 1M | ~$0.000113 |
| **Text Gen (Output)** | `gemini-3.5-flash-lite` | ~400 | $1.50 / 1M | ~$0.000600 |
| **TTS Gen (Input)** | `gemini-3.1-flash-tts-preview` | ~250 | $1.00 / 1M | ~$0.000250 |
| **TTS Gen (Output)** | `gemini-3.1-flash-tts-preview` | ~12,000 (Audio) | $20.00 / 1M | ~$0.240000 |
| **Total Cost** | | **~13,100** | | **~$0.240963 (~24¢)** |

Generating an entire 3-take orchestrated VO session costs **roughly 24 cents**, making this architecture highly scalable for production use cases.

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
    # Required — the deploy script fails fast if any of these are unset:
    export PROJECT_ID="<NEW_PROJECT_ID>"       # or GOOGLE_CLOUD_PROJECT
    export GOOGLE_CLOUD_LOCATION="us-central1"
    export GENMEDIA_BUCKET="<NEW_BUCKET>"      # created + CORS-enabled automatically by the deploy flow

    # Optional:
    export USE_IAP="false"                       # Skips IAP setup and uses --allow-unauthenticated (default)
    export SERVICE_NAME="three-up-generator"     # Override the default service name
    export VPC_CONNECTOR="my-vpc-connector"      # Only set if egress must route through a VPC connector
    export AR_REPO="cloud-run-source-deploy"     # Artifact Registry repo for the built image (default shown)
    ```
3.  **Run the script**:
    ```bash
    ./scripts/deploy.sh
    ```

The script will automatically create the Artifact Registry repository and the service account in the new project, grant it the required Vertex AI and Storage permissions, build the image, and deploy it to Cloud Run publicly.

> **Migration note (public deploy):** Earlier revisions shipped with the author's private-infrastructure defaults baked into `scripts/deploy.sh` (a hard-coded bucket, service name, IAP group, and VPC connector). Those have been removed so the one-click / script deploy works cleanly in any project. **You must now set `GENMEDIA_BUCKET`** (plus `PROJECT_ID`/`GOOGLE_CLOUD_PROJECT` and `GOOGLE_CLOUD_LOCATION`); the script exits with a clear error before making any `gcloud` call if they are missing. The bucket is no longer defaulted at runtime — the backend also requires `GENMEDIA_BUCKET` at startup. Set `VPC_CONNECTOR` only if you actually need VPC egress (the flag is omitted otherwise), and override `AR_REPO` only to publish into a non-default Artifact Registry repository.

### Advanced Configuration (Production)

To enable production features like bot protection, distributed rate limiting, and scalable analytics, you can set the following environment variables (either in your `.env.deploy` file or in the Cloud Run configuration):

#### 1. BigQuery Analytics
To track generation events (voice actor popularity, text lengths, performance) in BigQuery:
- `BQ_DATASET`: Your BigQuery Dataset ID.
- `BQ_TABLE`: Your BigQuery Table ID.
- `DEMO_NAME`: (Optional) Name to identify this app in the metrics (default: `take3bounce`).

*Note: The service account deployed with Cloud Run automatically includes the required BigQuery permissions, provided the dataset already exists.*

#### 2. reCAPTCHA v3
To protect the text generation endpoint against automated abuse:
- `RECAPTCHA_SITE_KEY`: Your Google reCAPTCHA Enterprise site key.
  - *Note: Ensure your domain or Cloud Run URL is added to the allowed domains list in the Google Cloud Console.*

#### 3. Distributed Rate Limiting (Redis)
To strictly enforce rate limits across multiple horizontally scaled Cloud Run instances:
- `REDIS_URL`: A standard Redis connection string (e.g., `redis://10.0.0.3:6379/0`). This could be Google Cloud Memorystore or a Serverless Redis instance.

#### 4. Tuning the Rate Limit
The `/api` endpoints are rate-limited. Both the request count and the window are configurable via environment variables (invalid values fall back to the defaults with a logged warning — the app never crashes on a bad value):
- `RATE_LIMIT_REQUESTS`: Max requests allowed per window (default `10`).
- `RATE_LIMIT_WINDOW`: Window length as a [Go duration string](https://pkg.go.dev/time#ParseDuration) such as `1m`, `30s`, or `2h` (default `1m`).

## Security & Abuse Posture

This application ships as a **public demo** and is deliberately configured for easy, open deployment. Understand the following before exposing an instance to untrusted traffic:

- **The instance is PUBLIC by default.** When deployed without IAP (`USE_IAP=false`), the Cloud Run service allows unauthenticated access so anyone with the URL can call the API. This is a conscious choice for the demo experience.
- **reCAPTCHA fails open.** When `RECAPTCHA_SITE_KEY` is unset, request validation is skipped ("Simple Mode") and all requests are allowed through. This is intentional so the app runs out of the box without extra setup. To turn on bot protection, enable **Enterprise reCAPTCHA** (see below).
- **The in-memory rate limit is per-instance and non-durable.** Without `REDIS_URL`, each Cloud Run instance keeps its own counters in memory. Those counters **reset on restart** and are **not shared across instances**, so the effective global limit scales with the number of running instances. Set `REDIS_URL` for a single, distributed limit enforced across all instances.
- **CORS lockdown is the primary defense against quota burn.** Restricting allowed origins (configured in Phase 4) is the main control that prevents third-party sites from driving traffic through your deployment and burning your Gemini/TTS quota. The rate limit and reCAPTCHA are complementary layers, not substitutes for it.

### Enabling Enterprise reCAPTCHA

1. Create a **reCAPTCHA Enterprise** key in the Google Cloud Console and add your Cloud Run URL / domain to its allowed domains list.
2. Ensure the `recaptchaenterprise.googleapis.com` API is enabled (the deploy tooling enables this for you).
3. Set `RECAPTCHA_SITE_KEY` to the key value (in `.env.deploy` or the Cloud Run configuration) and redeploy. Once set, requests missing a valid token, with a mismatched action, or scoring below `0.5` are rejected.

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
