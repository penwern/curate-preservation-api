# Curate Preservation API

[![Go Report Card](https://goreportcard.com/badge/github.com/penwern/curate-preservation-api)](https://goreportcard.com/report/github.com/penwern/curate-preservation-api)
[![Go Version](https://img.shields.io/badge/go-1.24.1+-blue.svg)](https://golang.org/dl/)

A high-performance, standalone Go application that provides RESTful API endpoints for managing digital preservation configurations. Designed for seamless integration with the Penwern Curate Digital Preservation ecosystem, this API supports full CRUD operations and is compatible with both MySQL and SQLite3 databases. AKA CA4M API.

## 🌟 Key Features

- **RESTful API** - Clean, standards-compliant REST endpoints for preservation configurations
- **Dual Database Support** - Compatible with both MySQL and SQLite3 databases
- **Flexible Configuration** - Support for command-line flags, environment variables, and configuration files (YAML/JSON)
- **Production Ready** - Built-in middleware stack with logging, recovery, request ID tracking, and timeout handling
- **A3M Integration** - Full compatibility with A3M processing configuration format
- **Docker Support** - Ready-to-use Docker containers and Docker Compose setup
- **Graceful Shutdown** - Proper handling of shutdown signals for zero-downtime deployments
- **Health Checks** - Built-in health check endpoints for monitoring and orchestration
- **Security Features** - IP-based authentication bypass for trusted networks
- **Auto-migrations** - Automatic database schema migrations and default configuration creation

## 🏗️ Architecture

The API serves as the configuration management layer for the Penwern Curate Digital Preservation ecosystem:

- **Configuration Storage** - Manages preservation workflow configurations
- **A3M Integration** - Provides A3M-compatible processing parameters
- **Authentication Layer** - JWT-based authentication with IP whitelisting
- **Health Monitoring** - Built-in health checks for orchestration

```
[Pydio Cells] ↔ [Preservation API]
      ↓
  [Preservation Core]
```

## 🔧 Dependencies

### Required
- **Go**: Version 1.24.1 or higher
- **Database**: MySQL 5.7+ or SQLite3
- **Git**: For version control and dependency management

### Optional
- **Docker**: For containerized deployment
- **Pydio Cells**: For OIDC authentication (production environments)

## 🚀 Quick Start

### Installation

```bash
# Clone the repository
git clone https://github.com/penwern/curate-preservation-api.git
cd curate-preservation-api

# Install dependencies
make deps

# Build the application
make build

# Verify installation
./curate-preservation-api version
```

### Start with SQLite (Default)

```bash
# Start the server with default settings
./curate-preservation-api serve

# The API will be available at http://localhost:6910
```

### Start with MySQL

```bash
# Configure MySQL connection
./curate-preservation-api serve \
  --db-type mysql \
  --db-connection "username:password@tcp(localhost:3306)/preservation_db"
```

### Test the API

```bash
# Health check
curl http://localhost:6910/api/v1/health

# List configurations
curl http://localhost:6910/api/v1/preservation-configs

# Create a new configuration
curl -X POST http://localhost:6910/api/v1/preservation-configs \
  -H "Content-Type: application/json" \
  -d '{"name": "Test Config", "description": "My first preservation configuration"}'
```

## 📚 API Reference

### Base URL

```
http://localhost:6910/api/v1
```

### Endpoints

| Method | Endpoint | Description | Authentication |
|--------|----------|-------------|----------------|
| `GET` | `/health` | Health check endpoint | None |
| `HEAD` | `/health` | Health check endpoint (headers only) | None |
| `GET` | `/preservation-configs` | List all configurations | Required* |
| `POST` | `/preservation-configs` | Create new configuration | Required* |
| `GET` | `/preservation-configs/{id}` | Get configuration by ID | Required* |
| `PUT` | `/preservation-configs/{id}` | Update configuration | Required* |
| `DELETE` | `/preservation-configs/{id}` | Delete configuration | Required* |

**Authentication Notes:**
- \* Authentication is required for all `/preservation-configs` endpoints
- Authentication can be bypassed for requests from trusted IP addresses (configured via `--trusted-ips`)
- Authentication uses Bearer tokens validated against Pydio Cells OIDC
- Trusted IPs are typically used for internal services and administrative access

### Response Format

#### Success Responses
Successful responses return the requested data directly as JSON:

```json
// For single resources (GET /preservation-configs/1)
{
  "id": 1,
  "name": "Standard Configuration",
  "description": "Standard preservation workflow",
  "compress_aip": true,
  "a3m_config": { /* A3M configuration */ },
  "created_at": "2024-01-15T10:30:00Z",
  "updated_at": "2024-01-15T10:30:00Z"
}

// For collections (GET /preservation-configs)
[
  {
    "id": 1,
    "name": "Config 1",
    /* ... */
  },
  {
    "id": 2,
    "name": "Config 2",
    /* ... */
  }
]
```

### Example API Calls

#### Create Configuration

```bash
curl -X POST http://localhost:6910/api/v1/preservation-configs \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Standard Preservation",
    "description": "Standard preservation workflow configuration",
    "compress_aip": true,
    "a3m_config": {
      "examine_contents": true,
      "normalize": true,
      "aip_compression_level": 6,
      "aip_compression_algorithm": "bzip2"
    }
  }'
```

#### Update Configuration

```bash
curl -X PUT http://localhost:6910/api/v1/preservation-configs/1 \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Updated Configuration",
    "description": "Updated preservation settings",
    "compress_aip": false,
    "a3m_config": {
      "examine_contents": false,
      "normalize": true
    }
  }'
```

## ⚙️ Configuration

The application supports multiple configuration methods with the following precedence order:
1. Command-line flags (highest priority)
2. Environment variables
3. Configuration file
4. Default values (lowest priority)

### Environment Variables

All environment variables use the `CA4M_API_` prefix:

| Variable | Description | Default |
|----------|-------------|---------|
| `CA4M_API_DB_TYPE` | Database type (sqlite3/mysql) | `sqlite3` |
| `CA4M_API_DB_CONNECTION` | Database connection string | `preservation_configs.db` |
| `CA4M_API_SERVER_PORT` | Server port | `6910` |
| `CA4M_API_SERVER_SITE_DOMAIN` | Site domain for OIDC | `https://localhost:8080` |
| `CA4M_API_SERVER_ALLOW_INSECURE_TLS` | Allow insecure TLS connections | `false` |
| `CA4M_API_SERVER_TRUSTED_IPS` | Trusted IP addresses/ranges | *(empty)* |
| `CA4M_API_LOG_LEVEL` | Log level (debug, info, warn, error, fatal, panic) | `info` |
| `CA4M_API_LOG_FILE` | Log file path | *(empty)* |

### Configuration File (YAML)

```yaml
db:
    connection: preservation_configs.db
    type: sqlite3
log:
    file: "/var/log/curate/preservation-api.log"
    level: info
server:
    allow_insecure_tls: false
    port: 6910
    site_domain: localhost:8080
    trusted_ips:
        - 127.0.0.1
        - ::1
        - 10.0.0.0/8
        - 172.16.0.0/12
        - 192.168.0.0/16
```

### Configuration Management Commands

```bash
# Generate configuration file
./curate-preservation-api config generate

# Validate configuration file
./curate-preservation-api config validate
```

## 🐳 Docker Deployment

### Using Docker Compose (Recommended)

```bash
# Start with SQLite (default)
docker-compose up -d

# Start with MySQL
docker-compose -f docker-compose.mysql.yml up -d
```

### Manual Docker Setup

```bash
# Create a network
docker network create preservation-network

# Run with SQLite
docker run -d \
  --name preservation-api \
  --network preservation-network \
  -p 6910:6910 \
  -v preservation_data:/app/data \
  ghcr.io/penwern/curate-preservation-api:latest
```

### Docker Environment Variables

```bash
docker run -d \
  --name preservation-api \
  -p 6910:6910 \
  -e CA4M_API_PORT=6910 \
  -e CA4M_API_DB_TYPE=sqlite3 \
  -e CA4M_API_LOG_LEVEL=info \
  ghcr.io/penwern/curate-preservation-api:latest
```

## 🛠️ Development

### Environment Setup

```bash
# Clone and setup
git clone https://github.com/penwern/curate-preservation-api.git
cd curate-preservation-api

# Install development tools
make install-tools

# Install dependencies
make deps

# Run in development mode
make run
```

### Available Make Targets

| Target | Description |
|--------|-------------|
| `make build` | Build the binary |
| `make build-all` | Build for multiple platforms |
| `make test` | Run all tests |
| `make format` | Format all Go files |
| `make lint` | Run linting |
| `make check` | Run format-check, lint, and test |
| `make pre-commit` | Run all pre-commit checks |
| `make clean` | Clean build artifacts |
| `make run` | Run in development mode |

### Testing

```bash
# Run all tests
make test

# Run specific package tests
go test -v ./server/...

# Run tests with race detection
go test -race ./...
```

### Code Quality

The project uses several tools to maintain code quality:

- **golangci-lint**: Comprehensive linting
- **goimports**: Import formatting
- **gofumpt**: Stricter formatting than gofmt
- **gci**: Import organization

```bash
# Run all quality checks
make check

# Auto-fix formatting and linting issues
make fix

# Run individual tools
make format      # Format code
make lint        # Run linter
make format-check # Check formatting
```

## 📊 Data Model

### Preservation Configuration

The core data model represents a digital preservation configuration:

```go
type PreservationConfig struct {
    ID          int64               `json:"id"`
    Name        string              `json:"name"`
    Description string              `json:"description"`
    CompressAIP bool                `json:"compress_aip"`
    A3MConfig   A3MProcessingConfig `json:"a3m_config"`
    CreatedAt   time.Time           `json:"created_at"`
    UpdatedAt   time.Time           `json:"updated_at"`
}
```

**Core Fields:**
- **ID**: Unique identifier (auto-generated)
- **Name**: Human-readable name (required)
- **Description**: Optional description
- **CompressAIP**: Whether to compress the final AIP package (boolean)
- **A3MConfig**: Detailed A3M processing configuration
- **CreatedAt/UpdatedAt**: Timestamps (auto-managed)

### A3M Configuration Options

The A3M configuration includes all processing options:

| Field | Type | Description | Default |
|-------|------|-------------|---------|
| `assign_uuids_to_directories` | `bool` | Assign UUIDs to directories | `true` |
| `examine_contents` | `bool` | Examine file contents for identification | `false` |
| `generate_transfer_structure_report` | `bool` | Generate transfer structure reports | `true` |
| `document_empty_directories` | `bool` | Document empty directories | `true` |
| `extract_packages` | `bool` | Extract packages during processing | `true` |
| `delete_packages_after_extraction` | `bool` | Remove packages after extraction | `false` |
| `identify_transfer` | `bool` | Identify files in transfer | `true` |
| `identify_submission_and_metadata` | `bool` | Identify submission and metadata | `true` |
| `identify_before_normalization` | `bool` | Identify files before normalization | `true` |
| `normalize` | `bool` | Perform normalization | `true` |
| `transcribe_files` | `bool` | Transcribe text files | `true` |
| `perform_policy_checks_on_originals` | `bool` | Policy checks on original files | `true` |
| `perform_policy_checks_on_preservation_derivatives` | `bool` | Policy checks on preservation copies | `true` |
| `perform_policy_checks_on_access_derivatives` | `bool` | Policy checks on access copies | `true` |
| `thumbnail_mode` | `string` | Thumbnail generation mode | `"generate"` |
| `aip_compression_level` | `int` | AIP compression level (1-9) | `1` |
| `aip_compression_algorithm` | `string` | Compression algorithm | `"bzip2"` |

## 🚢 Releases

### Creating a Release

```bash
# List existing tags
git tag --list

# Create new release tag
git tag -a v0.1.5 -m "Release version 0.1.5"

# Push tag to trigger CI/CD
git push origin v0.1.5

# Verify release
git describe --tags
```

## 🤝 Contributing

### Component Overview

- **API Server**: HTTP server with Chi router and middleware stack
- **Authentication**: JWT-based authentication with IP whitelisting
- **Database Layer**: Repository pattern with migration support
- **A3M Integration**: gRPC client for A3M processing service
- **Configuration**: Viper-based configuration management
- **Logging**: Structured logging with Zap

### Code Standards

- Follow Go best practices and idioms
- Write clear, concise comments
- Include unit tests for new functionality
- Ensure all tests pass: `make test`
- Run linting: `make lint`
- Format code: `make format`

### Development Workflow

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

### Reporting Issues

When reporting issues, please include:

- Go version: `go version`
- Operating system and version
- Database type and version
- Steps to reproduce the issue
- Expected vs actual behavior
- Relevant log output

## 🙏 Acknowledgments

- [A3M](https://github.com/artefactual-labs/a3m) - Digital preservation processing
- [Chi](https://github.com/go-chi/chi) - Lightweight HTTP router
- [Cobra](https://github.com/spf13/cobra) - CLI framework
- [Viper](https://github.com/spf13/viper) - Configuration management
- [Zap](https://github.com/uber-go/zap) - High-performance logging

---

**Made with ❤️ by the Penwern team**