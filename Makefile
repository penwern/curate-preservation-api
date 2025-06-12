# Build variables
BINARY_NAME=preservation-api
VERSION?=dev
GIT_COMMIT?=$(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
BUILD_DATE?=$(shell date -u +"%Y-%m-%dT%H:%M:%SZ")
LDFLAGS=-ldflags "-X github.com/penwern/curate-preservation-api/cmd.Version=$(VERSION) -X github.com/penwern/curate-preservation-api/cmd.GitCommit=$(GIT_COMMIT) -X github.com/penwern/curate-preservation-api/cmd.BuildDate=$(BUILD_DATE)"

.PHONY: build clean test run help install deps

# Default target
all: build

# Build the binary
build:
	@echo "Building $(BINARY_NAME)..."
	go build $(LDFLAGS) -o $(BINARY_NAME) .

# Build for multiple platforms
build-all:
	@echo "Building for multiple platforms..."
	GOOS=linux GOARCH=amd64 go build $(LDFLAGS) -o $(BINARY_NAME)-linux-amd64 .
	GOOS=darwin GOARCH=amd64 go build $(LDFLAGS) -o $(BINARY_NAME)-darwin-amd64 .
	GOOS=darwin GOARCH=arm64 go build $(LDFLAGS) -o $(BINARY_NAME)-darwin-arm64 .
	GOOS=windows GOARCH=amd64 go build $(LDFLAGS) -o $(BINARY_NAME)-windows-amd64.exe .

# Install dependencies
deps:
	@echo "Installing dependencies..."
	go mod download
	go mod tidy

# Run tests
test:
	@echo "Running tests..."
	go test -v ./...

# Run the application
run:
	@echo "Running $(BINARY_NAME)..."
	go run $(LDFLAGS) . serve

# Clean build artifacts
clean:
	@echo "Cleaning..."
	rm -f $(BINARY_NAME)
	rm -f $(BINARY_NAME)-*
	go clean

# Install the binary to GOPATH/bin
install:
	@echo "Installing $(BINARY_NAME)..."
	go install $(LDFLAGS) .

# Generate configuration file
config:
	@echo "Generating sample configuration..."
	./$(BINARY_NAME) config generate

# Show help
help:
	@echo "Available targets:"
	@echo "  build      - Build the binary"
	@echo "  build-all  - Build for multiple platforms"
	@echo "  deps       - Install dependencies"
	@echo "  test       - Run tests"
	@echo "  run        - Run the application"
	@echo "  clean      - Clean build artifacts"
	@echo "  install    - Install binary to GOPATH/bin"
	@echo "  config     - Generate sample configuration"
	@echo "  help       - Show this help message" 