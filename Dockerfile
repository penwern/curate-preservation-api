# Build stage
FROM golang:1.24.1-alpine AS builder

# Install build dependencies
RUN apk add --no-cache git ca-certificates gcc musl-dev make

# Set working directory
WORKDIR /app

# Copy go mod files first for better caching
COPY go.mod go.sum ./

# Download dependencies with cache mount
RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    go mod download

# Copy source code
COPY . .

# Build the application with cache mounts and optimized flags
RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    CGO_ENABLED=1 GOOS=linux go build \
    -ldflags "-s -w -extldflags '-static'" \
    -trimpath \
    -o preservation-api .

# Final stage
FROM alpine:latest

# Install runtime dependencies
RUN apk --no-cache add ca-certificates sqlite wget curl

# Create non-root user
RUN adduser -D -s /bin/sh apiuser

# Set working directory
WORKDIR /app

# Copy the binary from builder stage
COPY --from=builder /app/preservation-api .

# Copy the database migrations directory
COPY --from=builder /app/database/migrations ./database/migrations

# Create data directory for SQLite database
RUN mkdir -p /app/data && chown -R apiuser:apiuser /app

# Create logs directory following Linux standards
RUN mkdir -p /var/log/curate && chown -R apiuser:apiuser /var/log/curate

# Switch to non-root user
USER apiuser

# Expose port
EXPOSE 6910

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
  CMD wget --no-verbose --tries=1 --spider http://localhost:6910/api/v1/health || exit 1

# Run the application using Cobra serve command
CMD ["./preservation-api", "serve", "--port", "6910", "--db-connection", "/app/data/preservation_configs.db"]
