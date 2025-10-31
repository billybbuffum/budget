# Test Execution Guide

## Prerequisites

Ensure you have the required dependencies:

```bash
cd /home/user/budget
go mod download
```

Required packages:
- `github.com/mattn/go-sqlite3` (SQLite driver)
- Standard Go testing library

## Running Tests

### Option 1: Using Makefile (Recommended)

```bash
# Run all tests
make test

# Run unit tests only (business logic with mocks)
make test-unit

# Run HTTP handler tests only
make test-handler

# Run integration tests only (end-to-end scenarios)
make test-integration

# Run tests with coverage report
make test-coverage
# Opens coverage.html in browser

# Run tests with verbose output
make test-verbose

# Run tests with race detection
make test-race

# Run only CoverUnderfunded-related tests
make test-cover-underfunded

# Clean test artifacts
make clean
```

### Option 2: Using Go Commands Directly

```bash
# Run all tests
go test ./...

# Run with verbose output
go test ./... -v

# Run with coverage
go test ./... -cover

# Run specific package
go test ./internal/application -v
go test ./internal/infrastructure/http/handlers -v
go test ./internal/integration -v

# Run specific test function
go test ./internal/application -run TestAllocationService_GetAllocationSummary_UnderfundedCalculation -v

# Run all tests matching pattern
go test ./... -run CoverUnderfunded -v

# Run with race detection
go test ./... -race

# Generate coverage report
go test ./... -coverprofile=coverage.out
go tool cover -html=coverage.out -o coverage.html
```

## Expected Output

### Successful Test Run

```
$ make test
go test ./...
?       github.com/billybbuffum/budget/cmd/budget       [no test files]
ok      github.com/billybbuffum/budget/internal/application     0.234s
ok      github.com/billybbuffum/budget/internal/infrastructure/http/handlers    0.156s
ok      github.com/billybbuffum/budget/internal/integration     1.456s
```

### Test with Coverage

```
$ make test-coverage
go test ./... -cover
...
ok      github.com/billybbuffum/budget/internal/application     0.234s  coverage: 85.4% of statements
ok      github.com/billybbuffum/budget/internal/infrastructure/http/handlers    0.156s  coverage: 92.1% of statements
ok      github.com/billybbuffum/budget/internal/integration     1.456s  coverage: 78.9% of statements
...
Coverage report generated: coverage.html
```

### Verbose Test Output

```
$ make test-verbose
=== RUN   TestAllocationService_GetAllocationSummary_UnderfundedCalculation
=== RUN   TestAllocationService_GetAllocationSummary_UnderfundedCalculation/Real-time_budgeted_spending_-_no_underfunded
=== RUN   TestAllocationService_GetAllocationSummary_UnderfundedCalculation/Retroactive_full_allocation_-_no_underfunded
=== RUN   TestAllocationService_GetAllocationSummary_UnderfundedCalculation/Retroactive_partial_allocation_-_underfunded_$200
=== RUN   TestAllocationService_GetAllocationSummary_UnderfundedCalculation/No_allocation_-_fully_underfunded
=== RUN   TestAllocationService_GetAllocationSummary_UnderfundedCalculation/Multiple_categories_-_mixed_scenarios
--- PASS: TestAllocationService_GetAllocationSummary_UnderfundedCalculation (0.00s)
    --- PASS: TestAllocationService_GetAllocationSummary_UnderfundedCalculation/Real-time_budgeted_spending_-_no_underfunded (0.00s)
    --- PASS: TestAllocationService_GetAllocationSummary_UnderfundedCalculation/Retroactive_full_allocation_-_no_underfunded (0.00s)
    --- PASS: TestAllocationService_GetAllocationSummary_UnderfundedCalculation/Retroactive_partial_allocation_-_underfunded_$200 (0.00s)
    --- PASS: TestAllocationService_GetAllocationSummary_UnderfundedCalculation/No_allocation_-_fully_underfunded (0.00s)
    --- PASS: TestAllocationService_GetAllocationSummary_UnderfundedCalculation/Multiple_categories_-_mixed_scenarios (0.00s)
...
```

## Test Structure

```
/home/user/budget/
├── internal/
│   ├── application/
│   │   └── allocation_service_test.go
│   │       ├── TestAllocationService_GetAllocationSummary_UnderfundedCalculation
│   │       │   ├── Real-time budgeted spending
│   │       │   ├── Retroactive full allocation
│   │       │   ├── Retroactive partial allocation
│   │       │   ├── No allocation
│   │       │   └── Multiple categories
│   │       └── TestAllocationService_CoverUnderfundedPayment
│   │           ├── Success with sufficient RTA
│   │           ├── Error: category not found
│   │           ├── Error: not a payment category
│   │           ├── Error: not underfunded
│   │           ├── Error: insufficient RTA
│   │           └── Success: updates existing allocation
│   │
│   ├── infrastructure/http/handlers/
│   │   └── allocation_handler_test.go
│   │       └── TestAllocationHandler_CoverUnderfunded
│   │           ├── Success: valid request
│   │           ├── Error: invalid JSON
│   │           ├── Error: missing payment_category_id
│   │           ├── Error: missing period
│   │           ├── Error: category not found
│   │           ├── Error: not a payment category
│   │           ├── Error: not underfunded
│   │           └── Error: insufficient funds
│   │
│   └── integration/
│       └── cover_underfunded_test.go
│           ├── TestCoverUnderfunded_RealtimeBudgeting
│           ├── TestCoverUnderfunded_RetroactiveFullyBudgeted
│           ├── TestCoverUnderfunded_RetroactivePartiallyBudgeted
│           ├── TestCoverUnderfunded_InsufficientFunds
│           ├── TestCoverUnderfunded_MultipleCategories
│           └── TestCoverUnderfunded_UpdatesExistingAllocation
```

## Test Execution Flow

### Unit Tests (Fast)
```
Mock Repositories → Service Methods → Assertions
```
- Uses in-memory mock data
- No database required
- Very fast execution (<1ms per test)

### Handler Tests (Fast)
```
HTTP Request → Handler → Mock Service → HTTP Response
```
- Uses httptest package
- Mocks service layer
- Fast execution (<1ms per test)

### Integration Tests (Slower)
```
In-Memory DB → Repositories → Services → Assertions
```
- Creates real SQLite database in memory
- Tests full stack except HTTP layer
- Slower but still fast (<100ms per test)

## Debugging Failed Tests

### Step 1: Identify the Failing Test

Run with verbose output to see which test failed:
```bash
go test ./... -v
```

### Step 2: Run Only the Failing Test

```bash
# Example: if TestAllocationService_CoverUnderfundedPayment fails
go test ./internal/application -run TestAllocationService_CoverUnderfundedPayment -v
```

### Step 3: Check the Error Message

The test output will show:
- Expected value
- Actual value
- Line number
- Error description

Example:
```
allocation_service_test.go:456: Expected underfunded = 20000, got 30000
```

### Step 4: Check Mock Setup

For unit tests, verify:
- Mock data is set up correctly
- Mock methods return expected values
- Test case input matches what you're testing

### Step 5: Check Database State

For integration tests, add debug prints:
```go
// Add this to see what's in the database
allocs, _ := allocRepo.List(ctx)
for _, a := range allocs {
    t.Logf("Allocation: %s, Amount: %d, Period: %s", a.CategoryID, a.Amount, a.Period)
}
```

## Common Issues and Solutions

### Issue: "no test files"
**Cause:** Test file not in correct package or directory
**Solution:** Ensure test file is in same directory as code being tested

### Issue: "undefined: repository.NewAccountRepository"
**Cause:** Import path incorrect or package not exported
**Solution:** Check import paths and ensure types are exported

### Issue: "database locked"
**Cause:** SQLite database not properly closed
**Solution:** Ensure `defer db.Close()` is called in each test

### Issue: "foreign key constraint failed"
**Cause:** Trying to create entity with reference to non-existent entity
**Solution:** Create parent entities before child entities in test setup

### Issue: Test passes locally but fails in CI
**Cause:** Race condition or time-dependent logic
**Solution:** Run with `-race` flag: `go test ./... -race`

### Issue: "context deadline exceeded"
**Cause:** Test takes too long
**Solution:** Increase timeout: `go test ./... -timeout 5m`

## Performance Benchmarks

Expected test execution times:

| Test Type | Count | Expected Time |
|-----------|-------|---------------|
| Unit tests | 11 | < 0.5s |
| Handler tests | 8 | < 0.2s |
| Integration tests | 6 | < 2.0s |
| **Total** | **25** | **< 3s** |

If tests take significantly longer:
1. Check if external dependencies are being called
2. Verify database is in-memory (`:memory:`)
3. Look for network calls or file I/O
4. Consider parallelizing tests: `t.Parallel()`

## Coverage Analysis

### View Coverage Report

```bash
make test-coverage
# Opens coverage.html in browser
```

### Coverage by Package

Expected coverage:
- `internal/application`: 85%+ (high priority, core business logic)
- `internal/infrastructure/http/handlers`: 90%+ (critical API endpoints)
- `internal/integration`: 70%+ (focuses on workflows, not line coverage)

### Missing Coverage Areas

Use coverage report to identify:
- Uncovered error paths
- Edge cases not tested
- Dead code that can be removed

## Continuous Integration

### GitHub Actions Example

```yaml
name: Tests
on: [push, pull_request]
jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      - uses: actions/setup-go@v4
        with:
          go-version: '1.21'
      - run: go mod download
      - run: make test-coverage
      - uses: actions/upload-artifact@v3
        with:
          name: coverage-report
          path: coverage.html
```

### Pre-commit Hook

Create `.git/hooks/pre-commit`:
```bash
#!/bin/bash
echo "Running tests..."
make test
if [ $? -ne 0 ]; then
    echo "Tests failed. Commit aborted."
    exit 1
fi
```

## Test Maintenance

### Adding New Tests

1. Identify what needs testing
2. Choose appropriate test type (unit/handler/integration)
3. Add test case to existing test function (table-driven)
4. Or create new test function if testing new feature
5. Run tests to verify: `make test`
6. Update documentation if adding new test file

### Updating Existing Tests

1. When code changes, update corresponding tests
2. When bugs are found, add regression test
3. Keep test data realistic and representative
4. Maintain test readability and clarity

### Removing Tests

1. Remove obsolete tests when features are removed
2. Archive tests for deprecated features (don't delete history)
3. Update documentation to reflect removed tests

## Best Practices

1. **Run tests before committing:** `make test`
2. **Write tests first (TDD):** Test → Code → Refactor
3. **Keep tests simple:** One assertion per test case when possible
4. **Use descriptive names:** Test name should describe scenario
5. **Clean up:** Use `defer` for cleanup (db.Close(), etc.)
6. **Avoid flaky tests:** No randomness, no time dependencies
7. **Test public interfaces:** Don't test private methods
8. **Use table-driven tests:** For multiple similar scenarios
9. **Mock external dependencies:** Database, APIs, file system
10. **Document complex tests:** Add comments for non-obvious logic

## Getting Help

If you encounter issues:

1. Check error messages carefully
2. Run with `-v` flag for verbose output
3. Review test documentation (TESTING.md)
4. Check test code for examples
5. Verify database schema is correct
6. Ensure all dependencies are installed

## Quick Commands Reference

```bash
# Most common commands
make test                  # Run all tests
make test-coverage         # Run with coverage report
make test-verbose          # Run with verbose output
go test ./... -v          # Alternative verbose run
go test ./... -run Name   # Run specific test
```

---

**Ready to test!** Start with `make test` to run all tests.
