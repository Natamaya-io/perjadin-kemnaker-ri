# =============================================================================
# SOTA Cache-Optimized Multi-Stage Dockerfile
# Layer strategy: Heavy layers (node_modules) cached first, source last.
# Result: After first push, only the ~2MB "build" layer changes each deploy.
# =============================================================================

# ---- Stage 1: Install ALL deps (this layer cached unless package.json changes) ----
FROM public.ecr.aws/docker/library/node:20-alpine AS deps-all
WORKDIR /app
COPY package.json package-lock.json .npmrc ./
# npm ci is faster & more reproducible than npm install
RUN npm ci

# ---- Stage 2: Build (only re-runs when src changes, deps layer is reused) ----
FROM public.ecr.aws/docker/library/node:20-alpine AS builder
WORKDIR /app

# Build-time env vars for Vite
ARG VITE_API_URL=/api/v1
ARG VITE_USE_REAL_API=true
ARG VITE_SHOW_DEMO_BANNER=false
ENV VITE_API_URL=$VITE_API_URL
ENV VITE_USE_REAL_API=$VITE_USE_REAL_API
ENV VITE_SHOW_DEMO_BANNER=$VITE_SHOW_DEMO_BANNER

# Reuse cached node_modules from stage 1
COPY --from=deps-all /app/node_modules ./node_modules

# Copy source code LAST (most frequently changing layer)
COPY . .
RUN npm run build

# ---- Stage 3: Production deps only (this layer cached unless package.json changes) ----
FROM public.ecr.aws/docker/library/node:20-alpine AS deps-prod
WORKDIR /app
COPY package.json package-lock.json ./
RUN npm ci --omit=dev && \
    find node_modules -type f -name "*.md" -delete && \
    find node_modules -type f -name "*.ts" -delete && \
    find node_modules -type f -name "*.map" -delete && \
    find node_modules -type d -name "test" -exec rm -rf {} + 2>/dev/null || true && \
    find node_modules -type d -name "tests" -exec rm -rf {} + 2>/dev/null || true && \
    find node_modules -type d -name "example" -exec rm -rf {} + 2>/dev/null || true && \
    find node_modules -type d -name "examples" -exec rm -rf {} + 2>/dev/null || true && \
    find node_modules -type d -name "docs" -exec rm -rf {} + 2>/dev/null || true

# ---- Stage 4: Minimal production runtime ----
FROM public.ecr.aws/docker/library/alpine:3.20
RUN apk add --no-cache nodejs

WORKDIR /app
ENV NODE_ENV=production

# Layer order matters: heaviest+most-stable first → lightest+most-changed last
# Layer A: prod node_modules (~100MB, changes only on package.json update)
COPY --from=deps-prod /app/node_modules ./node_modules
# Layer B: package.json (rarely changes)
COPY --from=builder /app/package.json ./
# Layer C: built app (~3-5MB, changes on EVERY deploy - this is the ONLY layer pushed each time)
COPY --from=builder /app/build ./build

EXPOSE 3000
CMD ["node", "build/index.js"]
