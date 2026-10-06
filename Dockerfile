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
# Stage 3: THE SOTA FAT CONTAINER
# Built ON TOP of gotenberg/gotenberg:8
# to guarantee identical LibreOffice rendering
# ==========================================
FROM gotenberg/gotenberg:8
USER root
WORKDIR /app

# Install PostgreSQL, Redis, Supervisor, and Locales on Debian
RUN apt-get update && apt-get install -y --no-install-recommends \
    curl ca-certificates gnupg locales \
    && sed -i -e "s/# en_US.UTF-8 UTF-8/en_US.UTF-8 UTF-8/" /etc/locale.gen \
    && dpkg-reconfigure --frontend=noninteractive locales \
    && curl -fsSL https://www.postgresql.org/media/keys/ACCC4CF8.asc | gpg --dearmor -o /etc/apt/trusted.gpg.d/postgresql.gpg \
    && echo "deb http://apt.postgresql.org/pub/repos/apt/ trixie-pgdg main" > /etc/apt/sources.list.d/pgdg.list \
    && apt-get update && apt-get install -y --no-install-recommends \
    postgresql-16 \
    redis-server \
    supervisor \
    gosu \
    && rm -rf /var/lib/apt/lists/*

ENV LANG=en_US.UTF-8
ENV LANGUAGE=en_US:en
ENV LC_ALL=en_US.UTF-8

# Copy Go Backend
COPY --from=backend-builder /app/backend/perjadin-api .
COPY --from=backend-builder /app/backend/db/migrations ./db/migrations
COPY --from=backend-builder /app/backend/templates ./templates

# Setup configuration and data directories
RUN mkdir -p uploads /var/lib/postgresql/data /run/postgresql /etc/supervisor/conf.d

# Inject Supervisor, Entrypoint, and PG Wrapper
COPY deploy/fat-container/supervisord.conf /etc/supervisor/conf.d/supervisord.conf
COPY deploy/fat-container/entrypoint.sh /entrypoint.sh
COPY deploy/fat-container/entrypoint-pg.sh /entrypoint-pg.sh
RUN chmod +x /entrypoint.sh /entrypoint-pg.sh

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
CMD ["/usr/bin/supervisord", "-n", "-c", "/etc/supervisor/conf.d/supervisord.conf"]
