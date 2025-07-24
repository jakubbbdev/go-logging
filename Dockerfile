# Multi-stage Dockerfile for Go Logging Library
# Stage 1: Build
FROM golang:1.23-alpine AS builder

# Install dependencies
RUN apk add --no-cache git ca-certificates tzdata

# Set working directory
WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the application with optimizations
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags='-w -s -extldflags "-static"' \
    -a -installsuffix cgo \
    -o main cmd/examples/enterprise/main.go

# Stage 2: Runtime
FROM scratch

# Copy CA certificates for HTTPS
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

# Copy timezone data
COPY --from=builder /usr/share/zoneinfo /usr/share/zoneinfo

# Copy the binary
COPY --from=builder /app/main /main

# Environment variables for container logging
ENV LOG_LEVEL=info
ENV LOG_FORMAT=json
ENV LOG_OUTPUT=console
ENV JSON_LOGS=true
ENV LOG_USE_COLORS=false
ENV SERVICE_NAME=go-logging-demo
ENV VERSION=1.0.0
ENV ENVIRONMENT=production

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD /main --health-check || exit 1

# Run as non-root user
USER 1000:1000

# Expose ports (if needed for metrics)
EXPOSE 8080

# Entry point
ENTRYPOINT ["/main"] 