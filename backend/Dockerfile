# Stage 1: Build the Go binary
FROM golang:1.24-alpine AS builder

# Install SSL ca certificates (required if the API makes external HTTPS calls)
RUN apk --no-cache add ca-certificates tzdata

WORKDIR /app

# Copy Go modules first for better caching
COPY go.mod go.sum ./
RUN go mod download

# Copy the rest of the source code
COPY . .

# Build the application with strict optimizations
# -s: disable symbol table
# -w: disable DWARF generation
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o main ./cmd/api/main.go

# Stage 2: Ultra-lightweight runtime (Scratch is an empty 0MB image)
FROM scratch

WORKDIR /app

# Copy the timezone data and CA certificates from the builder
COPY --from=builder /usr/share/zoneinfo /usr/share/zoneinfo
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

# Copy the compiled, stripped binary
COPY --from=builder /app/main .
COPY --from=builder /app/.env.example .env

# Expose the application port
EXPOSE 8081

# Command to run the application
ENTRYPOINT ["./main"]
