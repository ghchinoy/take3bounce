# Project Context: The "Three-Up" Audio Orchestrator

## Overview
The "Three-Up" Audio Orchestrator is a performance comparison engine designed to transform a single text input into three distinct audio "performances" (Takes A, B, and C) using automated AI direction. It simulates a professional Voice Over (VO) booth experience by applying different Personas, Subtexts, and Technical Energies to the same script, generating marked-up scripts and TTS audio via Gemini.

## Architecture
- **Backend:** Go (1.25), `gorilla/mux` for routing. It uses the `google.golang.org/genai` SDK configured for the **Vertex AI backend** to generate script markups (Gemini Pro) and synthesize audio (Gemini Flash TTS).
- **Frontend:** TypeScript, Lit, and Vite. It utilizes `@material/web` components alongside custom Web Components from the `scream-ui` project, specifically `@ghchinoy/lit-text-ui` and `@ghchinoy/lit-audio-ui`.
- **Deployment:** Dockerized multi-stage build deployed to Google Cloud Run via `scripts/deploy.sh`. It uses a dedicated service account (`threeup-sa`) with `roles/aiplatform.user` and optionally supports IAP for secure access.

## Design System: "Synthetix Studio"
The frontend adheres to a bespoke design system created via Stitch, built on a "Precision Brutalism" North Star.
- **Theme:** Dark mode (`#0e0e0f` background).
- **Typography:** `Space Grotesk` for headlines/monumental elements, `Inter` for dense technical script readouts.
- **Visuals:** High-frequency neon accents (Primary Cyan `#8ff5ff`, Secondary Purple `#c97cff`, Tertiary Green `#8eff71`).
- **Core Rule ("No-Line"):** Do NOT use 1px solid borders to section off the interface. Boundaries must be defined solely through background color shifts (tonal layering) to mimic high-end physical studio gear.

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
- **Backend:** `cd backend && go run main.go` (Requires `GOOGLE_CLOUD_PROJECT` and `GOOGLE_CLOUD_LOCATION` in the environment or `.env`).
- **Frontend:** `cd frontend && npm run dev` (Vite proxies `/api` to the backend on `localhost:8080`).
- **Build/Deploy:** Ensure dependencies use the public npm registry (e.g., `"@ghchinoy/lit-text-ui": "^0.2.0"`), then run `make deploy` from the root.