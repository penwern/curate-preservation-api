# Build stage
FROM golang:1.24.1-alpine AS builder

# Install build dependencies
RUN apk add --no-cache git ca-certificates gcc musl-dev make

# Set working directory
WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build arguments for version information
ARG VERSION=docker
ARG GIT_COMMIT=unknown
ARG BUILD_DATE=unknown

# Build the application with version information
RUN CGO_ENABLED=1 GOOS=linux go build -a \
    -ldflags "-extldflags '-static' \
    -X github.com/penwern/curate-preservation-api/cmd.Version=${VERSION} \
    -X github.com/penwern/curate-preservation-api/cmd.GitCommit=${GIT_COMMIT} \
    -X github.com/penwern/curate-preservation-api/cmd.BuildDate=${BUILD_DATE}" \
    -o preservation-api .

# Final stage
FROM alpine:latest

# Install runtime dependencies
RUN apk --no-cache add ca-certificates sqlite wget

# Create non-root user
RUN adduser -D -s /bin/sh apiuser

# Set working directory
WORKDIR /app

# Copy the binary from builder stage
COPY --from=builder /app/preservation-api .

# Create data directory for SQLite database
RUN mkdir -p /app/data && chown -R apiuser:apiuser /app

# Switch to non-root user
USER apiuser

# Expose port
EXPOSE 6910

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
  CMD wget --no-verbose --tries=1 --spider http://localhost:6910/api/v1/health || exit 1

# Run the application using Cobra serve command
CMD ["./preservation-api", "serve", "--port", "6910", "--db-connection", "/app/data/preservation_configs.db"]
