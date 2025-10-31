# Test Quick Reference

## Run Commands

| Command | Description |
|---------|-------------|
| `make test` | Run all tests |
| `make test-unit` | Run unit tests only |
| `make test-handler` | Run handler tests only |
| `make test-integration` | Run integration tests only |
| `make test-coverage` | Generate HTML coverage report |
| `make test-verbose` | Run with verbose output |
| `make test-race` | Run with race detection |
| `make test-cover-underfunded` | Run CoverUnderfunded tests only |

## Test Files

| File | Type | Test Count | Description |
|------|------|------------|-------------|
| `internal/application/allocation_service_test.go` | Unit | 11 cases | Business logic with mocks |
| `internal/infrastructure/http/handlers/allocation_handler_test.go` | Handler | 8 cases | HTTP request/response |
| `internal/integration/cover_underfunded_test.go` | Integration | 6 scenarios | End-to-end workflows |

## Test Scenarios

### Unit Tests

#### GetAllocationSummary - Underfunded Calculation
- ✓ Real-time budgeting (allocate → spend) → underfunded = nil
- ✓ Retroactive full budget (spend → allocate all) → underfunded = nil
- ✓ Retroactive partial budget (spend $500 → allocate $300) → underfunded = $200
- ✓ No allocation (spend $500 → no budget) → underfunded = $500
- ✓ Multiple categories with mixed scenarios

#### CoverUnderfundedPayment
- ✓ Success with sufficient RTA
- ✓ Error: category not found
- ✓ Error: not a payment category
- ✓ Error: not underfunded
- ✓ Error: insufficient RTA
- ✓ Updates existing allocation (upsert)

### Handler Tests

#### POST /api/allocations/cover-underfunded
- ✓ Success: valid request → 200 OK
- ✓ Error: invalid JSON → 400 Bad Request
- ✓ Error: missing payment_category_id → 400 Bad Request
- ✓ Error: missing period → 400 Bad Request
- ✓ Error: category not found → 400 Bad Request
- ✓ Error: not payment category → 400 Bad Request
- ✓ Error: not underfunded → 400 Bad Request
- ✓ Error: insufficient funds → 400 Bad Request

### Integration Tests

#### Real-Time Budgeting
- Allocate $500 to Groceries
- Spend $500 on CC in Groceries
- ✓ Underfunded = nil (fully covered)
- ✓ CoverUnderfunded fails (not underfunded)

#### Retroactive Full Budget
- Spend $500 on CC (no allocation)
- ✓ Underfunded = $500
- Allocate $500 to Groceries
- ✓ Underfunded = nil (Option A fix!)

#### Retroactive Partial Budget
- Spend $500 on CC
- Allocate $300 to Groceries
- ✓ Underfunded = $200
- Cover underfunded
- ✓ Payment allocation = $200
- ✓ Underfunded = nil
- ✓ Total budget = $500

#### Insufficient Funds
- Spend $500 on CC
- Only $200 income
- ✓ Underfunded = $500
- ✓ CoverUnderfunded fails (insufficient funds)

#### Multiple Categories
- Groceries: $300 spent, $300 allocated ✓
- Gas: $200 spent, $100 allocated ✗
- ✓ Underfunded = $100
- ✓ Underfunded categories = ["Gas"]
- Cover underfunded
- ✓ Underfunded = nil

#### Update Existing Allocation
- Spend $500 on CC
- Manually allocate $200 to payment
- ✓ Underfunded = $300
- Cover underfunded
- ✓ Payment allocation = $500 (updated, not added)

## Key Test Patterns

### Table-Driven Tests
```go
tests := []struct {
    name    string
    input   Input
    want    Output
    wantErr bool
}{
    {"case 1", input1, output1, false},
    {"case 2", input2, output2, true},
}

for _, tt := range tests {
    t.Run(tt.name, func(t *testing.T) {
        // test logic
    })
}
```

### Mock Repository Pattern
```go
type mockRepository struct {
    data []*domain.Entity
    err  error
}

func (m *mockRepository) Create(ctx context.Context, entity *domain.Entity) error {
    if m.err != nil {
        return m.err
    }
    m.data = append(m.data, entity)
    return nil
}
```

### Integration Test Setup
```go
db := setupTestDB(t)
defer db.Close()

repo := repository.NewRepository(db)
service := application.NewService(repo)

// Test scenario
```

## Common Assertions

| Assertion | Code |
|-----------|------|
| Check error | `if err != nil { t.Errorf("...") }` |
| Check no error | `if err == nil { t.Error("...") }` |
| Check value | `if got != want { t.Errorf("got %v, want %v", got, want) }` |
| Check nil | `if x != nil { t.Error("...") }` |
| Check not nil | `if x == nil { t.Error("...") }` |
| Check contains | `if !contains(s, substr) { t.Error("...") }` |

## Critical Test Values

| Scenario | Amounts (cents) |
|----------|-----------------|
| Full budget | Spend: 50000, Allocate: 50000 → Underfunded: nil |
| Partial budget | Spend: 50000, Allocate: 30000 → Underfunded: 20000 |
| No budget | Spend: 50000, Allocate: 0 → Underfunded: 50000 |
| Insufficient RTA | Spend: 50000, Income: 20000 → Error |

## Test Database Schema

```
accounts
├── id (TEXT PRIMARY KEY)
├── name (TEXT)
├── type (checking/savings/cash/credit)
└── balance (INTEGER)

categories
├── id (TEXT PRIMARY KEY)
├── name (TEXT)
├── group_id (TEXT FK)
└── payment_for_account_id (TEXT FK, nullable)

transactions
├── id (TEXT PRIMARY KEY)
├── account_id (TEXT FK)
├── category_id (TEXT FK, nullable)
├── amount (INTEGER)
└── date (DATETIME)

allocations
├── id (TEXT PRIMARY KEY)
├── category_id (TEXT FK)
├── amount (INTEGER)
├── period (TEXT)
└── UNIQUE(category_id, period)
```

## Helper Functions

```go
// String pointer
func strPtr(s string) *string { return &s }

// Int64 pointer
func int64Ptr(i int64) *int64 { return &i }

// Contains check
func contains(s, substr string) bool { ... }
```

## Coverage Goals

- **Unit Tests:** 80%+ coverage
- **Handler Tests:** 100% coverage of endpoint
- **Integration Tests:** All critical workflows

## Debugging Tips

1. **Test fails:** Run with `-v` flag for verbose output
2. **Unclear error:** Check assertion messages
3. **Mock not working:** Verify mock setup matches expectations
4. **Integration test fails:** Check database schema initialization
5. **Race condition:** Run with `-race` flag

## Quick Test Run

```bash
# Fast check (all tests)
go test ./...

# Detailed output
go test ./... -v

# Coverage report
go test ./... -cover

# Specific test
go test ./internal/application -run TestAllocationService_CoverUnderfundedPayment -v
```

## What's Tested

| Feature | Unit | Handler | Integration |
|---------|------|---------|-------------|
| Underfunded calculation | ✓ | - | ✓ |
| CoverUnderfunded validation | ✓ | ✓ | ✓ |
| HTTP request handling | - | ✓ | - |
| Real-time budgeting | ✓ | - | ✓ |
| Retroactive budgeting | ✓ | - | ✓ |
| Multiple categories | ✓ | - | ✓ |
| Error handling | ✓ | ✓ | ✓ |
| Budget integrity | - | - | ✓ |

## Status Codes

| Scenario | Status Code |
|----------|-------------|
| Success | 200 OK |
| Invalid JSON | 400 Bad Request |
| Missing field | 400 Bad Request |
| Not payment category | 400 Bad Request |
| Not underfunded | 400 Bad Request |
| Insufficient funds | 400 Bad Request |

---

**Last Updated:** Generated with comprehensive test suite
**Total Tests:** 25 test cases across 9 test functions
**Total Code:** ~1,800 lines of test code
