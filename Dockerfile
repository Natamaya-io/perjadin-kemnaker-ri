# ==========================================
# Stage 1: Build SvelteKit SPA Frontend
# ==========================================
FROM node:20-alpine AS frontend-builder
WORKDIR /app/frontend

COPY frontend/package.json frontend/package-lock.json* ./
RUN npm ci || npm install

COPY frontend/ .

# Inject Build Args for Vite/SvelteKit
ARG VITE_API_URL
ARG VITE_USE_REAL_API
ARG VITE_SHOW_DEMO_BANNER
ENV VITE_API_URL=$VITE_API_URL
ENV VITE_USE_REAL_API=$VITE_USE_REAL_API
ENV VITE_SHOW_DEMO_BANNER=$VITE_SHOW_DEMO_BANNER

# This will output to frontend/build/ (based on svelte.config.js adapter-static)
RUN npm run build

# ==========================================
# Stage 2: Build Go Backend (with embedded UI)
# ==========================================
FROM golang:1.25-alpine AS backend-builder
WORKDIR /app/backend

# Install build dependencies
RUN apk add --no-cache gcc musl-dev

COPY backend/go.mod backend/go.sum ./
RUN go mod download

COPY backend/ .

# Inject the built frontend files into the Go package for embedding
RUN mkdir -p ui/dist
COPY --from=frontend-builder /app/frontend/build/ ui/dist/

# Build statically with Mimalloc / Jemalloc flags if needed, here we use standard static build
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o perjadin-api ./cmd/api

# ==========================================
# Stage 3: Final Production Image
# ==========================================
FROM alpine:latest
WORKDIR /app

# Install CA certificates (webpki-roots analog) and tzdata
RUN apk --no-cache add ca-certificates tzdata

COPY --from=backend-builder /app/backend/perjadin-api .
COPY --from=backend-builder /app/backend/db/migrations ./db/migrations
COPY --from=backend-builder /app/backend/templates ./templates

RUN mkdir -p uploads
EXPOSE 8081

# Liveness probe (Deep Health Probe via Echo)
HEALTHCHECK --interval=30s --timeout=5s --start-period=5s --retries=3 \
  CMD wget --no-verbose --tries=1 --spider http://localhost:8081/api/v1/auth/demo-users || exit 1

CMD ["./perjadin-api"]
