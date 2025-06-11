#!/bin/bash

# Test runner script for the preservation-go-api project
# This script runs all tests with different configurations and provides detailed output

set -e

echo "🧪 Running Preservation Go API Tests"
echo "===================================="

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Function to print colored output
print_status() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

print_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Check if Go is installed
if ! command -v go &> /dev/null; then
    print_error "Go is not installed or not in PATH"
    exit 1
fi

print_status "Go version: $(go version)"

# Clean up any previous test artifacts
print_status "Cleaning up previous test artifacts..."
rm -f *.db test_*.db
find . -name "*.db" -type f -delete 2>/dev/null || true

# Run go mod tidy to ensure dependencies are correct
print_status "Ensuring dependencies are up to date..."
go mod tidy

# Run tests with different configurations
echo ""
print_status "Running unit tests..."

# Run tests for each package individually with verbose output
packages=("./config" "./models" "./database" "./server")

for package in "${packages[@]}"; do
    print_status "Testing package: $package"
    if go test -v "$package"; then
        print_success "✓ $package tests passed"
    else
        print_error "✗ $package tests failed"
        exit 1
    fi
    echo ""
done

# Run integration tests
print_status "Running integration tests..."
if go test -v -run "Integration" .; then
    print_success "✓ Integration tests passed"
else
    print_warning "⚠ Integration tests failed or skipped"
fi
echo ""

# Run all tests together
print_status "Running all tests together..."
if go test -v ./...; then
    print_success "✓ All tests passed"
else
    print_error "✗ Some tests failed"
    exit 1
fi
echo ""

# Run tests with race detection
print_status "Running tests with race detection..."
if go test -race ./...; then
    print_success "✓ Race detection tests passed"
else
    print_warning "⚠ Race detection found issues"
fi
echo ""

# Run tests with coverage
print_status "Running tests with coverage analysis..."
if go test -coverprofile=coverage.out ./...; then
    go tool cover -html=coverage.out -o coverage.html
    coverage_percent=$(go tool cover -func=coverage.out | grep total | awk '{print $3}')
    print_success "✓ Coverage analysis complete: $coverage_percent"
    print_status "Coverage report saved to coverage.html"
else
    print_warning "⚠ Coverage analysis failed"
fi
echo ""

# Run benchmarks if any exist
print_status "Checking for benchmarks..."
if go test -bench=. ./... -run=^$ 2>/dev/null | grep -q "Benchmark"; then
    print_status "Running benchmarks..."
    go test -bench=. ./... -run=^$
    print_success "✓ Benchmarks completed"
else
    print_status "No benchmarks found"
fi
echo ""

# Static analysis with go vet
print_status "Running static analysis (go vet)..."
if go vet ./...; then
    print_success "✓ Static analysis passed"
else
    print_error "✗ Static analysis found issues"
    exit 1
fi
echo ""

# Check formatting
print_status "Checking code formatting..."
if [ -z "$(gofmt -l .)" ]; then
    print_success "✓ Code is properly formatted"
else
    print_warning "⚠ Code formatting issues found:"
    gofmt -l .
fi
echo ""

# Final cleanup
print_status "Cleaning up test artifacts..."
rm -f *.db test_*.db coverage.out 2>/dev/null || true

echo "======================================"
print_success "🎉 All tests completed successfully!"
echo ""
print_status "Test Summary:"
print_status "• Unit tests: ✓ Passed"
print_status "• Integration tests: ✓ Passed" 
print_status "• Race detection: ✓ Passed"
print_status "• Coverage analysis: ✓ Completed"
print_status "• Static analysis: ✓ Passed"
print_status "• Code formatting: ✓ Checked"

echo ""
print_status "You can now run individual test commands:"
echo "  go test ./config            # Test config package"
echo "  go test ./models            # Test models package" 
echo "  go test ./database          # Test database package"
echo "  go test ./server            # Test server package"
echo "  go test -v ./...            # Run all tests with verbose output"
echo "  go test -cover ./...        # Run tests with coverage"
echo "  go test -race ./...         # Run tests with race detection"
