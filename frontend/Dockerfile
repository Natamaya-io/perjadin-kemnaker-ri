# Stage 1: Build the SvelteKit application
FROM node:20-alpine AS builder

WORKDIR /app

# Copy package files strictly
COPY package.json package-lock.json* ./

# Clean install to ensure precise dependency resolution
RUN npm install

# Copy the rest of the application
COPY . .

# Build the SvelteKit application
RUN npm run build

# Stage 2: Create a lightweight runtime image
FROM node:20-alpine

# Set Node to production environment for performance
ENV NODE_ENV=production
ENV PORT=3000

WORKDIR /app

# Copy ONLY what is absolutely necessary from builder
COPY --from=builder /app/build ./build
COPY --from=builder /app/package.json ./package.json
COPY --from=builder /app/package-lock.json ./package-lock.json

# Clean install only production dependencies and wipe npm cache immediately
# to prevent temporary files from inflating the final image layer
RUN npm install --omit=dev && npm cache clean --force

# Use the non-root node user created by the alpine image
USER node

# Expose the application port
EXPOSE 3000

# Run the compiled Node.js server
CMD ["node", "build/index.js"]
