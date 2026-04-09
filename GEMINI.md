# Project Context: The "Three-Up" Audio Orchestrator

## Overview
The "Three-Up" Audio Orchestrator is a performance comparison engine designed to transform a single text input into three distinct audio "performances" (Takes A, B, and C) using automated AI direction. It simulates a professional Voice Over (VO) booth experience by applying different Personas, Subtexts, and Technical Energies to the same script, generating marked-up scripts and TTS audio via Gemini.

## Architecture
- **Backend:** Go (1.25), `gorilla/mux` for routing. It uses the `google.golang.org/genai` SDK configured for the **Vertex AI backend** to generate script markups (Gemini Pro) and synthesize audio (Gemini Flash TTS).
- **Frontend:** TypeScript, Lit, and Vite. It utilizes `@material/web` components alongside custom Web Components from the `scream-ui` project, specifically `@ghchinoy/lit-text-ui` and `@ghchinoy/lit-audio-ui`.
- **Deployment:** Dockerized multi-stage build deployed to Google Cloud Run via `scripts/deploy.sh`. It uses a dedicated service account (`threeup-sa`) with `roles/aiplatform.user` and optionally supports IAP for secure access.

## Design Systems: "Synthetix Studio" (Dark) & "Sunrise Studio" (Light)
The frontend must support a dual-theme architecture based on two Stitch design systems, with a seamless toggle between them. See `docs/DESIGN.md` (Dark) and `docs/DESIGN_SUNRISE.md` (Light) for full tokens, and `docs/design/` for visual references.
- **Synthetix Studio (Dark):** "Precision Brutalism" using a `#0e0e0f` background, `Space Grotesk` & `Inter` fonts, and strict tonal layering (no borders) with cyan/purple/green neon accents.
- **Sunrise Studio (Light):** "Approachable Clarity" using a `#FAFAFA` background, `Sora` & `Nunito Sans` fonts, soft drop shadows, full rounded corners, and warm orange/green/blue accents.

## Task Management (Beads / `bd`)
This project strictly uses **Beads (`bd`)** for issue tracking. Do NOT use markdown TODO lists.
- The project is configured to use a `dolt` SQL server in `--shared-server` mode.
- Implementor agents should run `bd ready` to find available work.
- Planners/Designers should break down features into fine-grained tasks and use `bd create` and `bd dep add` to establish blocking relationships (e.g., Design System -> Layout -> Components).
- See `AGENTS.md` for specific `bd` command workflows and the mandatory session close protocol.

## Agent Collaboration Guidelines
- **Planners & Designers:** Focus on the "why" and "what". Document strategies in `docs/`. When dealing with UI, focus on functional hierarchy, state management (e.g., "Processing" states), and comparison-first layouts suitable for A/B/C auditing. Engage with the `Stitch` MCP for UI conceptualization.
- **Implementors:** Focus on the "how". Pick up tasks from `bd`. Pay strict attention to **Code Health**:
  - *Concurrency:* Use Goroutines/WaitGroups for parallel tasks (like generating 3 TTS audio files).
  - *Resilience:* Strip Markdown artifacts when parsing JSON from LLMs. Handle missing environment variables by failing fast (`log.Fatalf`).
  - *Robustness:* Dynamically handle MIME types from APIs rather than hardcoding. Provide clear, user-facing error states in the UI.

## Local Development
- **Backend:** `cd backend && go run .` (Requires `GOOGLE_CLOUD_PROJECT` and `GOOGLE_CLOUD_LOCATION` in the environment or `.env`). DO NOT use `go run main.go` as it will fail to compile secondary package files.
- **Frontend:** `cd frontend && npm run dev` (Vite proxies `/api` to the backend on `localhost:8080`).
- **Build/Deploy:** Ensure dependencies use the public npm registry (e.g., `"@ghchinoy/lit-text-ui": "^0.2.0"`), then run `make deploy` from the root.

## Operational Learnings & Architecture "Gotchas"
- **Vite & Lit Components:** If locally symlinked Lit components cause a blank page with a `NotSupportedError: Failed to execute 'define' on 'CustomElementRegistry'` in the console, Vite is double-bundling the libraries. Fix this by adding `resolve.dedupe: ['lit', '@material/web']` to `vite.config.ts`.
- **Vite Multi-Page Apps:** Vite ignores all HTML files except `index.html` during `npm run build`. To compile a secondary page (like a sandbox), you MUST explicitly map it in `vite.config.ts` under `build.rollupOptions.input`.
- **IPv6 Proxy Issues:** When proxying Vite (`/api`) to a Go backend, always set the proxy target to `http://127.0.0.1:8080`. Using `http://localhost:8080` often results in a 502 `ECONNREFUSED` error because Node attempts to route to the IPv6 `::1` loopback, while the Go server binds to the IPv4 wildcard.
- **Gemini TTS Modality (Raw PCM):** When explicitly requesting the `AUDIO` modality from `gemini-3.1-flash-tts-preview`, the engine does NOT return a formatted MP3. It returns raw `audio/l16` (16-bit PCM, 24kHz) bytes. To make this playable in an HTML `<audio>` element, the backend *must* dynamically calculate and prepend a standard 44-byte `RIFF/WAVE` header to the byte array before saving/sending.
- **TTS Safety Filters:** If `GenerateContent` succeeds without a Go `error` but returns 0 Candidates, it is almost always a Safety Filter block (e.g. intense emotional tags triggering `PROHIBITED_CONTENT`). Always extract and log `ttsResp.PromptFeedback.BlockReason` to catch this.
- **Firebase Storage Tokens & CORS:** To serve GCS audio directly to a browser `<audio>` tag bypassing IAM, you must emulate Firebase Storage. During the Go `storage` upload, generate a UUID and inject it into the object's `Metadata["firebaseStorageDownloadTokens"]`. Append `?alt=media&token=<UUID>` to the returned `firebasestorage.googleapis.com` URL. Additionally, the bucket itself must have CORS explicitly configured (`gcloud storage buckets update gs://<bucket> --cors-file=cors.json`) or the browser will block the `206 Partial Content` audio streaming requests.
- **Lit DOM Reuse & Canvas Components:** When rendering dynamic lists of canvas-based Web Components (like `<ui-audio-tag-editor>`) that change based on state, you MUST use Lit's `repeat(items, keyFn, template)` directive instead of a standard array `.map()`. This prevents Lit from reusing stale DOM nodes and forces a proper canvas redraw.
- **Native Audio Downloads (Cross-Origin):** Standard `<a download>` attributes often fail to trigger a file-save dialog when the source URL is cross-origin (like a GCS bucket), instead opening the audio in a new tab. To force a native file download dialog, the frontend must `fetch` the URL, convert it to a `blob()`, generate a local `window.URL.createObjectURL(blob)`, and programmatically trigger a click on a temporary anchor element.
- **Vertex AI 3.1 TTS Safety Overrides:** The `gemini-3.1-flash-tts-preview` model on Vertex AI explicitly rejects `BLOCK_NONE` overrides for Hate Speech, Sexually Explicit, and Dangerous Content, causing Internal 500 errors. Only `HARM_CATEGORY_HARASSMENT` can be successfully overridden to `HarmBlockThresholdBlockNone` within the `GenerateContentConfig`.
- **Go Build & Standalone Scripts:** Standalone Go scripts declaring `package main` must be placed in a dedicated sub-directory (e.g., `backend/tools/`) to prevent `main redeclared in this block` compilation errors during `go build .` of the primary application.
