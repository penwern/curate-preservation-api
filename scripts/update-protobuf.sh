#!/bin/bash
# Script to update protobuf definitions locally
# This script mirrors the GitHub Actions workflow for local development

set -euo pipefail

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Function to print colored output
print_info() {
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

# Check if buf is installed
if ! command -v buf &> /dev/null; then
    print_error "buf CLI is not installed. Please install it first:"
    echo "  # Install buf CLI"
    echo "  BUF_VERSION=\"1.28.1\""
    echo "  curl -sSL \"https://github.com/bufbuild/buf/releases/download/v\${BUF_VERSION}/buf-\$(uname -s)-\$(uname -m)\" \\"
    echo "    -o /usr/local/bin/buf"
    echo "  chmod +x /usr/local/bin/buf"
    exit 1
fi

print_info "buf version: $(buf --version)"

# Navigate to the protobuf directory
PROTO_DIR="common/proto/a3m"
if [ ! -d "$PROTO_DIR" ]; then
    print_error "Protobuf directory not found: $PROTO_DIR"
    print_error "Please run this script from the project root directory"
    exit 1
fi

cd "$PROTO_DIR"

print_info "Current working directory: $(pwd)"

# Check if this is a remote dependency setup
if [ -f buf.yaml ] && grep -q "buf.build" buf.yaml; then
    print_info "Detected remote dependency setup (buf.build/penwern/a3m)"
    print_info "Skipping dependency update - using remote module"
    REMOTE_DEPS=true
else
    REMOTE_DEPS=false
    # Backup current state for local dependencies
    print_info "Creating backup of current state..."
    cp buf.lock buf.lock.backup 2>/dev/null || print_warning "No existing buf.lock found"

    # Show current dependencies
    print_info "Current dependencies:"
    if [ -f buf.lock ]; then
        cat buf.lock
    else
        print_warning "No lock file found"
    fi

    # Update dependencies
    print_info "Updating buf dependencies..."
    if buf dep update; then
        print_success "Dependencies updated successfully"
    else
        print_error "Failed to update dependencies"
        exit 1
    fi

    # Check if dependencies changed
    DEPS_CHANGED=false
    if [ -f buf.lock.backup ] && ! diff -q buf.lock.backup buf.lock >/dev/null 2>&1; then
        DEPS_CHANGED=true
        print_success "Dependencies were updated!"
        print_info "Dependency changes:"
        diff buf.lock.backup buf.lock || true
    else
        print_info "Dependencies are already up to date"
    fi
fi

# Generate Go code
print_info "Regenerating Go code from protobuf definitions..."
if buf generate; then
    print_success "Go code generated successfully"
else
    print_error "Failed to generate Go code"
    exit 1
fi

# Navigate back to project root
cd ../../..

# Check for changes in generated files
print_info "Checking for changes in generated files..."
if git diff --quiet; then
    print_success "No changes detected in generated files"
    print_info "Your protobuf definitions are up to date!"
else
    print_warning "Changes detected in generated files:"
    git diff --name-only
    
    echo ""
    print_info "Summary of changes:"
    if [ "$REMOTE_DEPS" = true ]; then
        echo "  ✓ Regenerated Go code from remote protobuf module"
    else
        if [ "${DEPS_CHANGED:-false}" = true ]; then
            echo "  ✓ Updated protobuf dependencies in buf.lock"
        fi
        echo "  ✓ Regenerated Go code from latest protobuf definitions"
    fi
    echo "  ✓ Source: buf.build/penwern/a3m"
    
    echo ""
    print_warning "Next steps:"
    echo "  1. Review the changes: git diff"
    echo "  2. Test the changes: go test ./..."
    echo "  3. Commit the changes: git add . && git commit -m \"chore: update protobuf definitions\""
fi

# Run tests to verify everything works
print_info "Running tests to verify protobuf updates..."
if go mod tidy && go test ./models/... -short; then
    print_success "All tests passed!"
else
    print_error "Some tests failed. Please review the changes."
    exit 1
fi

# Validate the generated code
print_info "Validating generated code..."
if go vet ./common/proto/... && go build ./common/proto/...; then
    print_success "Generated code is valid!"
else
    print_error "Generated code has issues. Please review."
    exit 1
fi

# Clean up backup (only if we created one)
if [ "$REMOTE_DEPS" = false ]; then
    rm -f "$PROTO_DIR/buf.lock.backup"
fi

print_success "Protobuf update completed successfully!"
echo ""
print_info "If you see changes above, consider creating a PR with:"
echo "  git add ."
echo "  git commit -m \"chore: update protobuf definitions\""
echo "  git push origin HEAD"