---
description: Run all tests and report results
allowed-tools: [Bash]
---

# Run Tests

Run all tests for the Budget application.

## Test Execution

### 1. Run All Go Tests
```bash
go test ./... -v
```

### 2. Run Tests with Coverage
```bash
go test ./... -cover -coverprofile=coverage.out
```

### 3. View Coverage Report
```bash
go tool cover -html=coverage.out -o coverage.html
echo "Coverage report generated: coverage.html"
```

### 4. Run Specific Package Tests
```bash
# Test application layer only
go test ./internal/application/... -v

# Test infrastructure layer only
go test ./internal/infrastructure/... -v

# Test specific file
go test ./internal/application/account_service_test.go -v
```

### 5. Run Tests with Race Detection
```bash
go test ./... -race
```

## Test Analysis

After running tests, analyze:
- [ ] All tests passing?
- [ ] Any flaky tests (sometimes pass, sometimes fail)?
- [ ] What's the coverage percentage?
- [ ] Any race conditions detected?

## If Tests Don't Exist Yet

Currently, the Budget application has **no tests**. Consider:

1. **Invoke test-generator agent** to create tests:
   ```
   Use the test-generator sub agent to create comprehensive tests
   ```

2. **Priority for test creation:**
   - ✅ AllocationService (critical business logic)
   - ✅ TransactionService (balance updates)
   - ✅ Repository implementations
   - ✅ HTTP handlers

## Test Quality Checks

### Code Quality
```bash
# Check for code issues
go vet ./...

# Format code
go fmt ./...

# Check for common mistakes
go vet ./...
```

### Build Check
```bash
# Ensure code compiles
go build ./...
```

### Dependency Check
```bash
# Check for outdated dependencies
go list -u -m all
```

## Continuous Testing During Development

### Watch Mode (requires additional tools)
If you have `entr` or similar:
```bash
# Re-run tests on file changes
find . -name "*.go" | entr -c go test ./...
```

### Quick Test Loop
```bash
# Build, test, run
go build ./... && go test ./... && go run cmd/server/main.go
```

## Test Report Format

```markdown
# Test Results

## Summary
- Total Tests: X
- Passed: X
- Failed: X
- Coverage: X%

## Failed Tests
[List any failed tests with error messages]

## Coverage Analysis
[Areas with good/poor coverage]

## Recommendations
- [ ] Areas needing more tests
- [ ] Flaky tests to fix
- [ ] Test quality improvements
```

## Budget App Testing Guidelines

When tests are created, they should cover:

### Critical Business Logic
- **Allocation calculations**
  - Ready to Assign = Balance - Allocated
  - Available per category with rollover
  - One allocation per category per period

- **Transaction balance updates**
  - Creating transaction updates balance
  - Updating transaction adjusts balance
  - Deleting transaction reverts balance
  - Operations are atomic

### Repository Operations
- CRUD operations for all entities
- Query filters work correctly
- Foreign key constraints enforced
- Unique constraints enforced

### HTTP Handlers
- Valid requests return correct status
- Invalid requests return 400
- Missing resources return 404
- Error handling is consistent

## No Tests Yet?

If no tests exist, here's the plan:

1. **Start with critical paths:**
   ```
   /new-feature test-suite - Generate tests for the application
   ```

2. **Or invoke the test generator:**
   ```
   Use test-generator agent to create comprehensive test suite
   ```

3. **Focus areas:**
   - AllocationService: Zero-based budgeting logic
   - TransactionService: Balance updates
   - Repositories: Database operations
   - Handlers: API endpoints
