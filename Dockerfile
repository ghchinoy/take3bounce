# Stage 1: Build Frontend
FROM node:20-alpine AS frontend-builder
WORKDIR /app/frontend
COPY frontend/package*.json ./
RUN npm install
COPY frontend/ .
RUN npm run build

# Stage 2: Build Backend
FROM golang:1.25-alpine AS backend-builder
WORKDIR /app/backend
COPY backend/go.mod backend/go.sum ./
RUN go mod download
COPY backend/ .
RUN go build -o server main.go

# Stage 3: Final Image
FROM alpine:latest
RUN apk add --no-cache ca-certificates
WORKDIR /app

# Copy binary
COPY --from=backend-builder /app/backend/server .

# Copy frontend assets
COPY --from=frontend-builder /app/frontend/dist ./dist

# Environment variables
ENV PORT=8080
# GOOGLE_CLOUD_PROJECT and GOOGLE_CLOUD_LOCATION should be provided at runtime or deployment

EXPOSE 8080
CMD ["./server"]
