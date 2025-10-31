# Quick Start: Testing AllocateToCoverUnderfunded

## Run All Tests

```bash
# Quick run - all tests
go test ./internal/application ./internal/infrastructure/http/handlers

# Verbose output
go test ./internal/application ./internal/infrastructure/http/handlers -v

# Use the convenience script
chmod +x run_tests.sh
./run_tests.sh
```

## Run Specific Tests

```bash
# Service layer only
go test ./internal/application -v -run "AllocateToCoverUnderfunded"

# Handler layer only
go test ./internal/infrastructure/http/handlers -v -run "CoverUnderfunded"

# Helper function tests
go test ./internal/application -v -run "CalculateReadyToAssignWithoutUnderfunded"
```

## Check Coverage

```bash
# Generate coverage report
go test ./internal/application -coverprofile=coverage.out
go tool cover -html=coverage.out

# Quick coverage summary
go test ./internal/application -cover
```

## Test What Was Added

The new `AllocateToCoverUnderfunded` functionality includes:

1. **Service Method** (`internal/application/allocation_service.go:333-389`)
   - Allocates funds to cover underfunded credit card payment categories
   - Validates category is a payment category
   - Checks if category is actually underfunded
   - Ensures sufficient funds are available

2. **Helper Method** (`internal/application/allocation_service.go:393-439`)
   - Calculates ready-to-assign without subtracting underfunded amounts
   - Used to determine available funds for allocation

3. **API Endpoint** (`POST /api/allocations/cover-underfunded`)
   - Accepts `category_id` and `period` in JSON body
   - Returns HTTP 201 with allocation on success
   - Returns HTTP 400 with error message on failure

## Files Created

- `/home/user/budget/internal/application/allocation_service_test.go` - Service tests (9 tests)
- `/home/user/budget/internal/infrastructure/http/handlers/allocation_handler_test.go` - Handler tests (8 tests)
- `/home/user/budget/TEST_SUMMARY.md` - Detailed documentation
- `/home/user/budget/run_tests.sh` - Convenience script
- `/home/user/budget/TESTING_QUICK_START.md` - This file

## Expected Test Results

All 17 tests should pass:
- 9 service layer unit tests
- 8 handler layer integration tests

```
PASS
ok      github.com/billybbuffum/budget/internal/application     0.1s
PASS
ok      github.com/billybbuffum/budget/internal/infrastructure/http/handlers     0.1s
```

## Common Issues

**Issue**: Tests fail with "category not found"
- **Fix**: Check that mock data includes the category being tested

**Issue**: Underfunded amount calculation is wrong
- **Fix**: Verify account balance and transaction amounts are set correctly

**Issue**: Insufficient funds error
- **Fix**: Ensure inflows are greater than allocations + underfunded amount

## Next Steps

1. Run the tests to verify they pass
2. Review coverage reports
3. Add integration tests with real database (future)
4. Set up CI/CD to run tests automatically

## Need Help?

- See `TEST_SUMMARY.md` for detailed documentation
- Check test files for examples
- Review the implementation in `allocation_service.go`
