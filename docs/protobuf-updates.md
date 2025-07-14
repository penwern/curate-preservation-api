# Protobuf Updates

This document explains how protobuf definitions are automatically updated in this project.

## Overview

This project uses [Buf](https://buf.build) to manage protobuf definitions from the A3M project. The protobuf definitions are automatically updated weekly via GitHub Actions and can also be updated manually.

## Automatic Updates

### Weekly Schedule

A GitHub Actions workflow runs every Monday at 8:00 AM UTC to:

1. Check for updates to the A3M protobuf definitions (`buf.build/penwern/a3m`)
2. Update dependencies if changes are found
3. Regenerate Go code from the latest protobuf definitions
4. Create a pull request with the changes
5. Run tests to ensure compatibility

### Workflow Details

- **Workflow File**: `.github/workflows/update-protobuf.yml`
- **Trigger**: Weekly schedule (Mondays) + manual dispatch
- **Dependencies**: `buf.build/penwern/a3m`
- **Generated Code**: `common/proto/a3m/gen/go/`

### Pull Request Process

When updates are found, the workflow will:

1. Create a new branch: `chore/update-protobuf-definitions-<timestamp>`
2. Commit changes with a descriptive message
3. Create a pull request with:
   - Detailed description of changes
   - Testing checklist
   - Review guidelines
   - Auto-merge policy information

## Manual Updates

### Using the Script

For local development, you can update protobuf definitions manually:

```bash
# Run the update script
./scripts/update-protobuf.sh
```

This script will:
- Check for buf CLI installation
- Update protobuf dependencies
- Regenerate Go code
- Run tests to verify changes
- Show a summary of any changes

### Manual Process

If you prefer to run the commands manually:

```bash
# Navigate to protobuf directory
cd common/proto/a3m

# Update dependencies
buf dep update

# Generate Go code
buf generate

# Navigate back to project root
cd ../../..

# Test the changes
go mod tidy
go test ./models/... -short
go vet ./common/proto/...
```

## Configuration Files

### `buf.yaml`

Defines the buf module configuration and dependencies:

```yaml
version: v2
deps:
  - buf.build/penwern/a3m
```

### `buf.gen.yaml`

Configures code generation for Go:

```yaml
version: v2
clean: true
plugins:
  - remote: buf.build/protocolbuffers/go
    opt:
      - paths=source_relative
    out: gen/go
  - remote: buf.build/grpc/go
    opt:
      - paths=source_relative
    out: gen/go
inputs:
  - module: buf.build/penwern/a3m:main
    paths:
      - a3m/api/transferservice/v1beta1
```

### `buf.lock`

Lock file that pins specific versions of dependencies (automatically managed).

## Integration with Go Code

The generated protobuf code is used in the `models` package:

- **`models/a3m_config.go`**: Wraps the generated `ProcessingConfig` with custom JSON marshaling
- **`models/preservation_config.go`**: Uses the A3M config in the main configuration struct

## Testing

### Automated Tests

The CI workflow includes several validation steps:

1. **Protobuf Validation**: Validates buf configuration and dependencies
2. **Code Generation**: Ensures generated code matches committed code
3. **Build Verification**: Confirms generated code compiles successfully
4. **Model Tests**: Runs tests for protobuf-related models
5. **Integration Tests**: Tests the full workflow with protobuf models

### Manual Testing

After updating protobuf definitions:

```bash
# Run all tests
go test ./...

# Test protobuf-specific functionality
go test ./models/ -v

# Test JSON serialization
go test ./models/ -run "JSON" -v

# Run benchmarks
go test ./models/ -bench=. -benchmem -run=^$
```

## Troubleshooting

### Common Issues

1. **Buf CLI not installed**:
   ```bash
   # Install buf CLI
   BUF_VERSION="1.28.1"
   curl -sSL "https://github.com/bufbuild/buf/releases/download/v${BUF_VERSION}/buf-$(uname -s)-$(uname -m)" \
     -o /usr/local/bin/buf
   chmod +x /usr/local/bin/buf
   ```

2. **Generated code out of sync**:
   ```bash
   cd common/proto/a3m
   buf generate
   cd ../../..
   git add .
   git commit -m "chore: regenerate protobuf code"
   ```

3. **Dependency conflicts**:
   ```bash
   cd common/proto/a3m
   buf dep update
   buf generate
   ```

4. **Build failures after update**:
   - Check for breaking changes in the protobuf definitions
   - Update model code to handle new fields or removed fields
   - Run tests to identify specific issues

### Getting Help

- Check the [Buf documentation](https://docs.buf.build)
- Review the A3M protobuf definitions at `buf.build/penwern/a3m`
- Look at recent GitHub Actions runs for detailed logs
- Check the `#development` channel for assistance

## Best Practices

### Reviewing Protobuf Updates

When reviewing automatically generated PRs:

1. **Check for Breaking Changes**: Look for removed or renamed fields
2. **Validate Backward Compatibility**: Ensure existing functionality works
3. **Test New Features**: Verify any new fields or methods work correctly
4. **Update Documentation**: Add documentation for significant changes
5. **Run Full Test Suite**: Ensure all tests pass

### Development Workflow

1. **Before Development**: Ensure protobuf definitions are up to date
2. **After Updates**: Run the full test suite
3. **Before Commits**: Verify generated code is committed
4. **In PRs**: Include protobuf changes in the same PR as related code changes

### Version Management

- Lock file (`buf.lock`) should always be committed
- Generated code should be committed with the source changes
- Use semantic versioning for releases that include protobuf changes
- Document breaking changes in release notes

## Security Considerations

- Protobuf updates are fetched from trusted sources (`buf.build/penwern/a3m`)
- Generated code is automatically tested before PR creation
- All changes go through the normal PR review process
- Dependencies are pinned in the lock file for reproducibility