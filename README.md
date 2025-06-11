# Preservation Configuration API Server

A standalone Go application that provides RESTful API endpoints for managing digital preservation configurations. The API supports CRUD operations (Create, Read, Update, Delete) and is compatible with both MySQL and SQLite3 databases.

## Features

- RESTful API for preservation configurations
- Database support for both MySQL and SQLite3
- Configurable via command-line flags or configuration file
- Graceful shutdown handling
- Integration with A3M processing configuration format
- Automatic middleware stack (logging, recovery, request ID, real IP, timeout)
- JSON-based API responses
- Default configuration creation on database initialization

## API Endpoints

| Method | Path | Description |
|--------|------|-------------|
| GET | `/api/v1/health` | Health check endpoint |
| GET | `/api/v1/preservation-configs` | List all preservation configurations |
| POST | `/api/v1/preservation-configs` | Create a new preservation configuration |
| GET | `/api/v1/preservation-configs/{id}` | Get a specific preservation configuration by ID |
| PUT | `/api/v1/preservation-configs/{id}` | Update an existing preservation configuration |
| DELETE | `/api/v1/preservation-configs/{id}` | Delete a preservation configuration |

## Usage

### Running the API Server

```bash
# Run with default settings (SQLite3 database in current directory)
go run main.go

# Run with custom port (default is 6910)
go run main.go --port 8080

# Run with MySQL
go run main.go --db mysql --conn "user:password@tcp(localhost:3306)/dbname"

# Run with SQLite3 file in specific location
go run main.go --db sqlite3 --conn "/path/to/database.db"

# Run with configuration file
go run main.go --config config.json
```

### Configuration File Format

```json
{
  "db_type": "mysql",
  "db_connection": "user:password@tcp(localhost:3306)/dbname",
  "port": 8080
}
```

## API Usage Examples

### Health Check

```bash
curl http://localhost:6910/api/v1/health
```

Response:

```json
{
  "status": "ok"
}
```

### List All Configurations

```bash
curl http://localhost:6910/api/v1/preservation-configs
```

### Create a New Configuration

```bash
curl -X POST http://localhost:6910/api/v1/preservation-configs \
  -H "Content-Type: application/json" \
  -d '{
    "name": "My Preservation Config",
    "description": "Custom configuration for digital preservation"
  }'
```

### Get a Specific Configuration

```bash
curl http://localhost:6910/api/v1/preservation-configs/1
```

### Update a Configuration

```bash
curl -X PUT http://localhost:6910/api/v1/preservation-configs/1 \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Updated Config Name",
    "description": "Updated description",
    "a3m_config": {
      "examine_contents": true,
      "normalize": false
    }
  }'
```

### Delete a Configuration

```bash
curl -X DELETE http://localhost:6910/api/v1/preservation-configs/1
```

## Preservation Configuration Model

The API manages digital preservation configurations with the following properties:

- **ID**: Unique identifier for the configuration (auto-generated)
- **Name**: Human-readable name for the configuration (required)
- **Description**: Optional description
- **A3M Configuration Options**:
  - `assign_uuids_to_directories`: Assign UUIDs to directories
  - `examine_contents`: Content examination
  - `generate_transfer_structure_report`: Transfer structure report generation  
  - `document_empty_directories`: Empty directory documentation
  - `extract_packages`: Package extraction settings
  - `delete_packages_after_extraction`: Delete packages after extraction
  - `identify_transfer`: File identification for transfer
  - `identify_submission_and_metadata`: Submission and metadata identification
  - `identify_before_normalization`: Identification before normalization
  - `normalize`: Normalization options
  - `transcribe_files`: File transcription
  - `perform_policy_checks_on_originals`: Policy checks on original files
  - `perform_policy_checks_on_preservation_derivatives`: Policy checks on preservation derivatives
  - `perform_policy_checks_on_access_derivatives`: Policy checks on access derivatives
  - `thumbnail_mode`: Thumbnail generation mode
  - `aip_compression_level`: AIP compression level (1-9)
  - `aip_compression_algorithm`: AIP compression algorithm
- **Timestamps**: `created_at` and `updated_at` (auto-managed)

## Integration with A3M

Preservation configurations are stored with full A3M `ProcessingConfig` compatibility. The API creates a default configuration on database initialization and supports all A3M processing configuration options. Configurations can be converted to the A3M format for use with the preservation system using the `ToA3MConfig()` method on the `PreservationConfig` model.

## Project Structure

```txt
preservation-go-api/
├── main.go                    # Application entry point
├── go.mod                     # Go module definition
├── go.sum                     # Go module checksums
├── README.md                  # Project documentation
├── run_tests.sh               # Test runner script
├── coverage.html              # Test coverage report
├── config/
│   ├── config.go              # Configuration handling
│   └── config_test.go         # Configuration tests
├── database/
│   ├── db.go                  # Database connection management
│   ├── repository.go          # Database operations (CRUD)
│   └── db_test.go             # Database tests
├── models/
│   └── processing_config.go   # Data models and A3M integration
└── server/
    ├── server.go              # HTTP server implementation
    ├── routes.go              # API route definitions and handlers
    └── server_test.go         # Server and API tests
```

## Development

### Prerequisites

- Go 1.24.1+
- MySQL or SQLite3

### Dependencies

- github.com/go-chi/chi/v5 - HTTP router and middleware
- github.com/go-chi/render - Response rendering helpers
- github.com/go-sql-driver/mysql - MySQL driver
- github.com/mattn/go-sqlite3 - SQLite3 driver
- github.com/penwern/curate-preservation-core - A3M processing configuration types

### Running Tests

```bash
# Run all tests
go test ./...

# Run tests with coverage
go test -cover ./...

# Generate coverage report
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out -o coverage.html

# Or use the provided script
./run_tests.sh
```

### Building

```bash
# Build the application
go build -o preservation-api main.go

# Run the built binary
./preservation-api --port 8080
```
