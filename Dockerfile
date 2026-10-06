# ==========================================
# Stage 1: Build SvelteKit SPA Frontend
# ==========================================
FROM node:20-alpine AS frontend-builder
WORKDIR /app/frontend

COPY frontend/package.json frontend/package-lock.json* ./
RUN npm ci || npm install

COPY frontend/ .

ARG VITE_API_URL
ARG VITE_USE_REAL_API
ARG VITE_SHOW_DEMO_BANNER
ENV VITE_API_URL=$VITE_API_URL
ENV VITE_USE_REAL_API=$VITE_USE_REAL_API
ENV VITE_SHOW_DEMO_BANNER=$VITE_SHOW_DEMO_BANNER

RUN npm run build

# ==========================================
# Stage 2: Build Go Backend
# ==========================================
FROM golang:1.25-alpine AS backend-builder
WORKDIR /app/backend

RUN apk add --no-cache gcc musl-dev

COPY backend/go.mod backend/go.sum ./
RUN go mod download

COPY backend/ .

RUN mkdir -p ui/dist
COPY --from=frontend-builder /app/frontend/build/ ui/dist/

RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o perjadin-api ./cmd/api

# ==========================================
# Stage 3: Gotenberg Engine (Chromium + PDF)
# ==========================================
FROM gotenberg/gotenberg:8 AS gotenberg-source

# ==========================================
# Stage 4: THE SOTA FAT CONTAINER
# ==========================================
FROM alpine:latest
WORKDIR /app

# Install all stateful and infrastructure dependencies
RUN apk --no-cache add \
    postgresql \
    redis \
    supervisor \
    su-exec \
    tzdata \
    ca-certificates \
    chromium \
    ttf-freefont \
    font-noto-emoji \
    curl \
    fontconfig

# Copy Gotenberg Binary and its system requirements
COPY --from=gotenberg-source /usr/bin/gotenberg /usr/bin/gotenberg

# Copy Go Backend
COPY --from=backend-builder /app/backend/perjadin-api .
COPY --from=backend-builder /app/backend/db/migrations ./db/migrations
COPY --from=backend-builder /app/backend/templates ./templates

# Setup configuration and data directories
RUN mkdir -p uploads /var/lib/postgresql/data /run/postgresql /etc/supervisor.d

# Inject Supervisor and Entrypoint
COPY deploy/fat-container/supervisord.conf /etc/supervisord.conf
COPY deploy/fat-container/entrypoint.sh /entrypoint.sh
RUN chmod +x /entrypoint.sh

# Force internal loopback connections
ENV DB_HOST=127.0.0.1
ENV DB_PORT=5432
ENV DB_USER=postgres
ENV DB_PASSWORD=
ENV DB_NAME=postgres
ENV DB_SSLMODE=disable
ENV REDIS_HOST=127.0.0.1
ENV REDIS_PORT=6379
ENV GOTENBERG_URL=http://127.0.0.1:3000

EXPOSE 8081

HEALTHCHECK --interval=30s --timeout=5s --start-period=5s --retries=3 \
  CMD curl -f http://127.0.0.1:8081/api/v1/auth/demo-users || exit 1

ENTRYPOINT ["/entrypoint.sh"]
CMD ["/usr/bin/supervisord", "-c", "/etc/supervisord.conf"]
