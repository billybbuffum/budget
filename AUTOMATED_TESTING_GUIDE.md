# Automated Testing Guide - Credit Card Sync Removal Feature

## Quick Start

Run all tests for the credit card sync removal feature:

```bash
# Run all tests in the project
go test ./...

# Run with verbose output to see all test names
go test -v ./...

# Run with coverage report
go test -cover ./...
```

## Running Specific Test Suites

### Validator Tests
```bash
# Run validator tests only
go test ./internal/infrastructure/http/validators/

# Run with verbose output
go test -v ./internal/infrastructure/http/validators/

# Run specific validator test
go test -v ./internal/infrastructure/http/validators/ -run TestValidateUUID
go test -v ./internal/infrastructure/http/validators/ -run TestValidatePeriodFormat
go test -v ./internal/infrastructure/http/validators/ -run TestValidatePeriodRange
go test -v ./internal/infrastructure/http/validators/ -run TestValidateAmountPositive
go test -v ./internal/infrastructure/http/validators/ -run TestValidateAmountBounds
```

### Service Tests
```bash
# Run allocation service tests only
go test ./internal/application/ -run TestAllocationService

# Run with verbose output
go test -v ./internal/application/ -run TestAllocationService

# Run specific service test scenarios
go test -v ./internal/application/ -run TestAllocationService_AllocateToCoverUnderfunded_Success
go test -v ./internal/application/ -run TestAllocationService_AllocateToCoverUnderfunded_CategoryNotFound
go test -v ./internal/application/ -run TestAllocationService_AllocateToCoverUnderfunded_InsufficientFunds
```

### Handler Tests
```bash
# Run allocation handler tests only
go test ./internal/infrastructure/http/handlers/ -run TestAllocationHandler

# Run with verbose output
go test -v ./internal/infrastructure/http/handlers/ -run TestAllocationHandler

# Run specific handler test scenarios
go test -v ./internal/infrastructure/http/handlers/ -run TestAllocationHandler_CoverUnderfunded_Success
go test -v ./internal/infrastructure/http/handlers/ -run TestAllocationHandler_CoverUnderfunded_InvalidJSON
go test -v ./internal/infrastructure/http/handlers/ -run TestAllocationHandler_CoverUnderfunded_InsufficientFunds
```

## Coverage Analysis

### Generate Coverage Report
```bash
# Generate coverage profile
go test -coverprofile=coverage.out ./...

# View coverage summary
go tool cover -func=coverage.out

# Generate HTML coverage report
go tool cover -html=coverage.out -o coverage.html

# Open coverage report in browser
open coverage.html  # macOS
xdg-open coverage.html  # Linux
```

### Package-Specific Coverage
```bash
# Coverage for validators only
go test -coverprofile=validators_coverage.out ./internal/infrastructure/http/validators/
go tool cover -func=validators_coverage.out

# Coverage for service only
go test -coverprofile=service_coverage.out ./internal/application/
go tool cover -func=service_coverage.out

# Coverage for handlers only
go test -coverprofile=handlers_coverage.out ./internal/infrastructure/http/handlers/
go tool cover -func=handlers_coverage.out
```

## Test Patterns

### Run Tests Matching a Pattern
```bash
# Run all tests with "Invalid" in the name
go test -v ./... -run Invalid

# Run all tests with "Success" in the name
go test -v ./... -run Success

# Run all tests with "Error" in the name
go test -v ./... -run Error
```

### Run Tests in Parallel
```bash
# Run tests with parallel execution
go test -v -parallel 4 ./...
```

### Run Tests with Race Detection
```bash
# Run with race detector (important for concurrent code)
go test -race ./...
```

## Test Output Formats

### Standard Output
```bash
# Standard test output
go test ./...
# Output: ok/FAIL for each package

# Example:
# ok      github.com/billybbuffum/budget/internal/infrastructure/http/validators    0.123s
# ok      github.com/billybbuffum/budget/internal/application    0.456s
# ok      github.com/billybbuffum/budget/internal/infrastructure/http/handlers    0.789s
```

### Verbose Output
```bash
# Verbose output shows each test
go test -v ./...

# Example:
# === RUN   TestValidateUUID
# === RUN   TestValidateUUID/valid_UUID_v4
# --- PASS: TestValidateUUID (0.00s)
#     --- PASS: TestValidateUUID/valid_UUID_v4 (0.00s)
```

### JSON Output
```bash
# JSON output for CI/CD integration
go test -json ./... > test_results.json
```

## Continuous Integration

### CI/CD Script
```bash
#!/bin/bash
# ci_test.sh

set -e

echo "Running all tests..."
go test ./...

echo "Running tests with coverage..."
go test -coverprofile=coverage.out ./...

echo "Coverage summary:"
go tool cover -func=coverage.out

echo "Checking coverage threshold (80% minimum)..."
COVERAGE=$(go tool cover -func=coverage.out | grep total | awk '{print substr($3, 1, length($3)-1)}')
THRESHOLD=80

if (( $(echo "$COVERAGE < $THRESHOLD" | bc -l) )); then
    echo "Coverage $COVERAGE% is below threshold $THRESHOLD%"
    exit 1
fi

echo "All tests passed with $COVERAGE% coverage!"
```

Make executable and run:
```bash
chmod +x ci_test.sh
./ci_test.sh
```

## Debugging Failed Tests

### Run Single Failing Test
```bash
# Run only the failing test with verbose output
go test -v ./internal/application/ -run TestAllocationService_AllocateToCoverUnderfunded_InsufficientFunds
```

### Add Debug Output
Temporarily add print statements in test:
```go
t.Logf("Debug: allocation = %+v", allocation)
t.Logf("Debug: error = %v", err)
```

### Use Test Timeout
```bash
# Set custom timeout (default is 10 minutes)
go test -timeout 30s ./...
```

## Common Issues

### Test Cache
Go caches test results. To force re-run:
```bash
# Clear test cache
go clean -testcache

# Run tests without cache
go test -count=1 ./...
```

### Import Errors
If you see import errors:
```bash
# Download dependencies
go mod download

# Tidy dependencies
go mod tidy

# Verify dependencies
go mod verify
```

### Build Errors
```bash
# Check for build issues
go build ./...

# Check for syntax errors
go vet ./...

# Run linter (if installed)
golangci-lint run
```

## Best Practices

### Before Committing
```bash
# Run this before every commit:
go test ./...                    # All tests pass
go test -race ./...              # No race conditions
go vet ./...                     # No suspicious code
go fmt ./...                     # Code is formatted
```

### Watch Mode (with external tool)
```bash
# Install gotestsum
go install gotest.tools/gotestsum@latest

# Run in watch mode
gotestsum --watch ./...
```

### Test Coverage Goals
- Validators: 100% coverage (achieved)
- Service methods: 100% coverage (achieved)
- HTTP handlers: 100% coverage (achieved)
- Overall project: Aim for 80%+

## Test File Structure

```
budget/
├── internal/
│   ├── application/
│   │   ├── allocation_service.go
│   │   └── allocation_service_test.go        ← Service tests
│   └── infrastructure/
│       └── http/
│           ├── handlers/
│           │   ├── allocation_handler.go
│           │   └── allocation_handler_test.go ← Handler tests
│           └── validators/
│               ├── validators.go
│               └── validators_test.go         ← Validator tests
├── TEST_SUMMARY.md                            ← Test documentation
└── AUTOMATED_TESTING_GUIDE.md                 ← This file
```

## Expected Test Output

### Successful Run
```
$ go test ./...
ok      github.com/billybbuffum/budget/internal/application    0.123s
ok      github.com/billybbuffum/budget/internal/infrastructure/http/handlers    0.456s
ok      github.com/billybbuffum/budget/internal/infrastructure/http/validators    0.089s
```

### With Coverage
```
$ go test -cover ./...
ok      github.com/billybbuffum/budget/internal/application    0.123s  coverage: 95.2% of statements
ok      github.com/billybbuffum/budget/internal/infrastructure/http/handlers    0.456s  coverage: 98.5% of statements
ok      github.com/billybbuffum/budget/internal/infrastructure/http/validators    0.089s  coverage: 100.0% of statements
```

## Next Steps

After all tests pass:

1. **Review Coverage**: Check coverage report for any missed lines
2. **Integration Testing**: Consider adding integration tests with real database
3. **Performance Testing**: Test with large datasets
4. **Security Testing**: Verify no sensitive data in error messages
5. **Documentation**: Update API documentation with new endpoint

## Resources

- Go Testing Package: https://pkg.go.dev/testing
- Go Coverage Tool: https://go.dev/blog/cover
- Table-Driven Tests: https://go.dev/wiki/TableDrivenTests
- Go Testing Best Practices: https://go.dev/doc/effective_go#testing
