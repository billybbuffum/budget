.PHONY: test test-unit test-handler test-integration test-coverage test-verbose clean

# Run all tests
test:
	go test ./...

# Run unit tests only
test-unit:
	go test ./internal/application -v

# Run handler tests only
test-handler:
	go test ./internal/infrastructure/http/handlers -v

# Run integration tests only
test-integration:
	go test ./internal/integration -v

# Run tests with coverage
test-coverage:
	go test ./... -cover
	go test ./... -coverprofile=coverage.out
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

# Run tests with verbose output
test-verbose:
	go test ./... -v

# Run tests with race detection
test-race:
	go test ./... -race

# Run specific test
test-cover-underfunded:
	go test ./... -run CoverUnderfunded -v

# Clean test artifacts
clean:
	rm -f coverage.out coverage.html

# Run all tests and show coverage
ci: test-coverage
