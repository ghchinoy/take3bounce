.PHONY: help dev dev-server dev-frontend build-frontend build-run deploy

help:
	@echo "Three-Up TTS Generator - Makefile Commands"
	@echo "----------------------------------------"
	@echo "Local Development:"
	@echo "  make dev            - Start both dev server and frontend concurrently"
	@echo "  make dev-server     - Start only the Go backend server"
	@echo "  make dev-frontend   - Start only the Vite frontend dev server"
	@echo "  make build-frontend - Build the frontend into backend/dist"
	@echo "  make build-run      - Build frontend and run the Go server locally"
	@echo ""
	@echo "Deployment:"
	@echo "  make deploy         - Build and deploy to Google Cloud Run"

dev:
	@$(MAKE) -j2 dev-server dev-frontend

dev-server:
	@echo "Starting Go backend server..."
	@cd backend && go run .

dev-frontend:
	@echo "Starting Vite frontend server..."
	@cd frontend && npm run dev

build-frontend:
	@echo "Building frontend..."
	@rm -rf backend/dist/*
	@cd frontend && npm run build
	@cp -R frontend/dist backend/

build-run: build-frontend dev-server

deploy:
	@./scripts/deploy.sh
