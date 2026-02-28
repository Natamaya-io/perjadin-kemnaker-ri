# Stage 1: Build the Go binary
FROM golang:1.24-alpine AS builder

# Install SSL ca certificates and UPX
RUN apk --no-cache add ca-certificates tzdata upx

WORKDIR /app

# Copy Go modules first for better caching
COPY go.mod go.sum ./
RUN go mod download

# Copy the rest of the source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-s -w -extldflags '-static'" -o main ./cmd/api/main.go

# Compress binary
RUN upx --best --lzma main

# Stage 2: Ultra-lightweight runtime
FROM scratch

WORKDIR /app

# Copy the compiled, stripped binary
COPY --from=builder /app/main .

# Expose the application port
EXPOSE 8081

# Command to run the application
ENTRYPOINT ["./main"]
