# Multi-stage build for minimum image size
# Stage 1: Build stage
FROM golang:1.23-alpine AS builder

# Install build dependencies
RUN apk add --no-cache git ca-certificates tzdata

# Set working directory
WORKDIR /app

# Copy go mod files first for better caching
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the application with optimizations
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags='-w -s -extldflags "-static"' \
    -a -installsuffix cgo \
    -o main ./cmd/app/main.go

# Stage 2: Final minimal image
FROM alpine:3.20

RUN apk add --no-cache ca-certificates tzdata wget

# Copy the binary from builder stage
COPY --from=builder /app/main /main

# Copy public assets
COPY --from=builder /app/public /public

# Expose port
EXPOSE 3000

# Health check via HTTP
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:3000/health || exit 1

# Run the binary
ENTRYPOINT ["/main"]