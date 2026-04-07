# Three-Up TTS Generator

A web application that generates three different emotional variations ("Takes") of a given text script and synthesizes them into speech using Google's Gemini TTS.

It demonstrates how to orchestrate:
1. **Gemini Generative AI (`gemini-2.5-pro`)** to rewrite the prompt strictly inserting emotion/technical voice tags.
2. **Gemini TTS (`gemini-3.1-flash-tts-preview`)** to read the tagged variations with different vocal energies.
3. A Lit Web Component frontend with custom text-tag visualization.

## Prerequisites

- **Go 1.25+**
- **Node.js 20+**
- A **Google Cloud Project** with Vertex AI enabled.
- Application Default Credentials (ADC) configured locally (`gcloud auth application-default login`).

## Configuration

Edit `backend/.env` to configure your Google Cloud project details:

```env
GOOGLE_CLOUD_PROJECT=your_project_id
GOOGLE_CLOUD_LOCATION=us-central1
PORT=8080
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