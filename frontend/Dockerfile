# Stage 1: Build the SvelteKit app
FROM node:20-alpine AS builder

WORKDIR /app

# Copy package files and install all dependencies (needed for build)
COPY package.json package-lock.json .npmrc ./
RUN npm install

# Copy source code and build the application
COPY . .
RUN npm run build

# Stage 2: Install production dependencies
FROM node:20-alpine AS deps
WORKDIR /app
COPY package.json package-lock.json ./
# Install, then aggressively prune node_modules
RUN npm install --omit=dev && \
    find node_modules -type f -name "*.md" -delete && \
    find node_modules -type f -name "*.ts" -delete && \
    find node_modules -type f -name "*.map" -delete && \
    find node_modules -type d -name "test" -exec rm -rf {} + && \
    find node_modules -type d -name "tests" -exec rm -rf {} + && \
    find node_modules -type d -name "example" -exec rm -rf {} + && \
    find node_modules -type d -name "examples" -exec rm -rf {} + && \
    find node_modules -type d -name "doc" -exec rm -rf {} + && \
    find node_modules -type d -name "docs" -exec rm -rf {} +

# Stage 3: Production runtime
FROM alpine:3.20

# Install nodejs
RUN apk add --no-cache nodejs

WORKDIR /app

# Set Node environment to production
ENV NODE_ENV=production

# Copy built app and production node_modules
COPY --from=deps /app/node_modules ./node_modules
COPY --from=builder /app/build ./build
COPY --from=builder /app/package.json ./

# Expose the production port
EXPOSE 3000

# Start the Node.js server
CMD ["node", "build/index.js"]
