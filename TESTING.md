# Test Documentation

## Overview

This document describes the comprehensive test suite for the credit card payment sync removal and manual allocation helper feature.

## Test Structure

### Unit Tests

**Location:** `/home/user/budget/internal/application/allocation_service_test.go`

Tests the business logic in isolation using mock repositories.

#### Test Cases

1. **TestAllocationService_GetAllocationSummary_UnderfundedCalculation**
   - Real-time budgeted spending (allocate first, then spend) - underfunded should be nil
   - Retroactive full allocation (spend first, then allocate fully) - underfunded should be nil
   - Retroactive partial allocation (spend $500, allocate $300) - underfunded should be $200
   - No allocation (spend $500, no allocation) - underfunded should be $500
   - Multiple categories, mixed scenarios

2. **TestAllocationService_CoverUnderfundedPayment**
   - Success: cover underfunded with sufficient RTA
   - Error: category not found
   - Error: not a payment category
   - Error: payment category not underfunded
   - Error: insufficient RTA
   - Success: updates existing allocation for period

### HTTP Handler Tests

**Location:** `/home/user/budget/internal/infrastructure/http/handlers/allocation_handler_test.go`

Tests the HTTP request/response handling using mock services.

#### Test Cases

**TestAllocationHandler_CoverUnderfunded**
- Success: valid JSON request
- Error: invalid JSON body (400)
- Error: missing payment_category_id (400)
- Error: missing period (400)
- Error: category not found (400)
- Error: not a payment category (400)
- Error: not underfunded (400)
- Error: insufficient funds (400)

### Integration Tests

**Location:** `/home/user/budget/internal/integration/cover_underfunded_test.go`

Tests complete scenarios end-to-end with a real SQLite database.

#### Test Scenarios

1. **TestCoverUnderfunded_RealtimeBudgeting**
   - Create accounts and categories
   - Allocate budget first
   - Spend on credit card
   - Verify no underfunded amount (fully budgeted via expense allocation)
   - Attempt to cover underfunded fails (not underfunded)

2. **TestCoverUnderfunded_RetroactiveFullyBudgeted**
   - Spend $500 on CC without allocation
   - Verify underfunded = $500
   - Allocate $500 to expense category
   - Verify underfunded = nil (Option A calculation fix)
   - Attempt to cover underfunded fails (not underfunded)

3. **TestCoverUnderfunded_RetroactivePartiallyBudgeted**
   - Spend $500 on CC
   - Allocate only $300 to expense category
   - Verify underfunded = $200
   - Cover underfunded successfully
   - Verify payment allocation = $200
   - Verify underfunded = nil
   - Verify total budget = $300 (expense) + $200 (payment) = $500

4. **TestCoverUnderfunded_InsufficientFunds**
   - Spend $500 on CC
   - Add only $200 income
   - Verify underfunded = $500
   - Attempt to cover fails with "insufficient funds: need 500 but only 200 available"

5. **TestCoverUnderfunded_MultipleCategories**
   - Spend $300 on Groceries (fully allocated)
   - Spend $200 on Gas ($100 allocated, $100 unbudgeted)
   - Verify underfunded = $100
   - Verify underfunded_categories = ["Gas"]
   - Cover underfunded successfully
   - Verify underfunded = nil

6. **TestCoverUnderfunded_UpdatesExistingAllocation**
   - Spend $500 on CC
   - Manually allocate $200 to payment category
   - Verify underfunded = $300
   - Cover underfunded
   - Verify payment allocation updated to $500 (not $200 + $300)

## Running the Tests

### Run All Tests

```bash
go test ./...
```

### Run Specific Test Files

```bash
# Unit tests
go test ./internal/application -v

# Handler tests
go test ./internal/infrastructure/http/handlers -v

# Integration tests
go test ./internal/integration -v
```

### Run Specific Test Cases

```bash
# Run a specific test
go test ./internal/application -run TestAllocationService_GetAllocationSummary_UnderfundedCalculation -v

# Run all CoverUnderfunded tests
go test ./... -run CoverUnderfunded -v
```

### Run with Coverage

```bash
go test ./... -cover

# Generate detailed coverage report
go test ./... -coverprofile=coverage.out
go tool cover -html=coverage.out -o coverage.html
```

### Run with Race Detection

```bash
go test ./... -race
```

## Test Requirements

### Dependencies

The tests require the following Go packages:
- `github.com/mattn/go-sqlite3` - SQLite driver for integration tests

Install dependencies:
```bash
go mod download
```

### Test Database

Integration tests use an in-memory SQLite database (`:memory:`), so no database setup is required. Each test creates a fresh database instance.

## Key Testing Principles

1. **Isolation:** Each test is independent and can run in any order
2. **Repeatability:** Tests produce the same results every time
3. **Fast:** Unit tests use mocks; integration tests use in-memory database
4. **Clear:** Test names describe what's being tested
5. **Comprehensive:** Cover happy path, error cases, and edge cases

## What's Tested

### Critical Business Logic

- Zero-based budgeting calculations
- Underfunded amount calculation (Option A: unbudgeted spending only)
- Ready to Assign calculation
- Allocation upsert logic
- Payment category validation

### Data Integrity

- CRUD operations work correctly
- Foreign key constraints enforced
- Unique constraints enforced
- Balance updates are consistent

### API Layer

- Valid requests succeed
- Invalid requests return appropriate status codes
- Error messages are clear and helpful
- Response format is correct

## Test Coverage Goals

- **Unit Tests:** 80%+ coverage of business logic
- **Integration Tests:** Cover all critical user workflows
- **Handler Tests:** 100% coverage of HTTP endpoints

## Troubleshooting

### SQLite Driver Issues

If you get errors about CGO or missing SQLite:

```bash
# Install build dependencies (Ubuntu/Debian)
sudo apt-get install build-essential

# Install build dependencies (macOS)
xcode-select --install
```

### Test Timeout

For long-running tests:

```bash
go test ./... -timeout 5m
```

## Future Test Additions

Consider adding:
- Performance/benchmark tests
- Concurrent access tests
- Load testing for HTTP endpoints
- Property-based testing for complex calculations
