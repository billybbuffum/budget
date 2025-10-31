---
name: go-testing
description: Go testing best practices and patterns for unit, integration, and E2E tests
triggers: [test, testing, unit test, integration test, mock, table-driven]
---

# Go Testing Skill

## Testing Philosophy

- Tests document behavior
- Tests enable refactoring
- Tests catch regressions
- Tests should be fast and reliable

## Test Types

### Unit Tests
**What:** Test individual functions/methods in isolation
**How:** Mock dependencies
**Where:** `*_test.go` files next to code
**Speed:** Very fast (<1ms each)

### Integration Tests
**What:** Test components working together
**How:** Use real dependencies (test database)
**Where:** `*_test.go` files
**Speed:** Fast (1-100ms each)

### E2E Tests
**What:** Test full system end-to-end
**How:** Real HTTP requests, real database
**Where:** Separate test package or `*_test.go`
**Speed:** Slower (100ms-1s each)

## Test File Structure

```
internal/application/
  account_service.go
  account_service_test.go  ← Test file

internal/infrastructure/repository/
  account_repository.go
  account_repository_test.go
```

## Basic Test Structure

```go
package application

import "testing"

func TestAccountService_CreateAccount(t *testing.T) {
    // Arrange: Set up test data and dependencies
    mockRepo := &MockAccountRepository{}
    service := NewAccountService(mockRepo)
    account := &domain.Account{Name: "Test", Type: "checking"}

    // Act: Execute the function being tested
    err := service.CreateAccount(account)

    // Assert: Verify the results
    if err != nil {
        t.Errorf("Expected no error, got %v", err)
    }
}
```

## Table-Driven Tests

Best practice for testing multiple scenarios:

```go
func TestAccountService_CreateAccount(t *testing.T) {
    tests := []struct {
        name    string
        account *domain.Account
        wantErr bool
        errMsg  string
    }{
        {
            name:    "valid checking account",
            account: &domain.Account{Name: "Checking", Type: "checking", Balance: 100000},
            wantErr: false,
        },
        {
            name:    "empty name",
            account: &domain.Account{Name: "", Type: "checking"},
            wantErr: true,
            errMsg:  "name is required",
        },
        {
            name:    "invalid type",
            account: &domain.Account{Name: "Test", Type: "invalid"},
            wantErr: true,
            errMsg:  "invalid account type",
        },
        {
            name:    "negative balance for non-credit",
            account: &domain.Account{Name: "Test", Type: "checking", Balance: -100},
            wantErr: true,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            mockRepo := &MockAccountRepository{}
            service := NewAccountService(mockRepo)

            err := service.CreateAccount(tt.account)

            if (err != nil) != tt.wantErr {
                t.Errorf("CreateAccount() error = %v, wantErr %v", err, tt.wantErr)
                return
            }

            if err != nil && tt.errMsg != "" && !strings.Contains(err.Error(), tt.errMsg) {
                t.Errorf("Expected error containing %q, got %q", tt.errMsg, err.Error())
            }
        })
    }
}
```

## Mocking

### Interface-Based Mocking

```go
// Define interface in domain
type AccountRepository interface {
    Create(account *Account) error
    GetByID(id string) (*Account, error)
}

// Create mock in test
type MockAccountRepository struct {
    CreateFunc  func(account *Account) error
    GetByIDFunc func(id string) (*Account, error)
}

func (m *MockAccountRepository) Create(account *Account) error {
    if m.CreateFunc != nil {
        return m.CreateFunc(account)
    }
    return nil
}

func (m *MockAccountRepository) GetByID(id string) (*Account, error) {
    if m.GetByIDFunc != nil {
        return m.GetByIDFunc(id)
    }
    return nil, nil
}

// Use mock in test
func TestService(t *testing.T) {
    mock := &MockAccountRepository{
        CreateFunc: func(account *Account) error {
            return errors.New("database error")
        },
    }

    service := NewService(mock)
    err := service.CreateAccount(&Account{})

    if err == nil {
        t.Error("Expected error")
    }
}
```

### Simple Mock Structs

```go
type MockRepository struct {
    accounts []*Account
    err      error
}

func (m *MockRepository) Create(account *Account) error {
    if m.err != nil {
        return m.err
    }
    m.accounts = append(m.accounts, account)
    return nil
}

func (m *MockRepository) GetByID(id string) (*Account, error) {
    if m.err != nil {
        return nil, m.err
    }
    for _, acc := range m.accounts {
        if acc.ID == id {
            return acc, nil
        }
    }
    return nil, ErrNotFound
}
```

## Integration Tests

### Testing Repositories with Real Database

```go
func setupTestDB(t *testing.T) *sql.DB {
    db, err := sql.Open("sqlite3", ":memory:")
    if err != nil {
        t.Fatalf("Failed to open database: %v", err)
    }

    // Initialize schema
    if err := database.InitSchema(db); err != nil {
        t.Fatalf("Failed to initialize schema: %v", err)
    }

    return db
}

func TestAccountRepository_Create(t *testing.T) {
    db := setupTestDB(t)
    defer db.Close()

    repo := repository.NewAccountRepository(db)
    account := &domain.Account{
        ID:      uuid.New().String(),
        Name:    "Test Account",
        Type:    "checking",
        Balance: 100000,
    }

    // Test Create
    err := repo.Create(account)
    if err != nil {
        t.Fatalf("Create failed: %v", err)
    }

    // Verify by reading back
    retrieved, err := repo.GetByID(account.ID)
    if err != nil {
        t.Fatalf("GetByID failed: %v", err)
    }

    if retrieved.Name != account.Name {
        t.Errorf("Name = %v, want %v", retrieved.Name, account.Name)
    }
    if retrieved.Balance != account.Balance {
        t.Errorf("Balance = %v, want %v", retrieved.Balance, account.Balance)
    }
}
```

### Test Helpers

```go
// Helper to create test account
func createTestAccount(t *testing.T, repo domain.AccountRepository) *domain.Account {
    t.Helper()  // Mark as helper function

    account := &domain.Account{
        ID:      uuid.New().String(),
        Name:    "Test Account",
        Type:    "checking",
        Balance: 100000,
    }

    if err := repo.Create(account); err != nil {
        t.Fatalf("Failed to create test account: %v", err)
    }

    return account
}

// Use helper
func TestSomething(t *testing.T) {
    repo := setupTestRepo(t)
    account := createTestAccount(t, repo)  // Clean and reusable
    // ... rest of test
}
```

## HTTP Handler Tests

```go
func TestAccountHandler_CreateAccount(t *testing.T) {
    // Mock service
    mockService := &MockAccountService{
        CreateAccountFunc: func(account *domain.Account) error {
            return nil
        },
    }

    handler := handlers.NewAccountHandler(mockService)

    // Create request
    body := `{"name":"Test","type":"checking","balance":100000}`
    req := httptest.NewRequest("POST", "/api/accounts", strings.NewReader(body))
    req.Header.Set("Content-Type", "application/json")

    // Record response
    w := httptest.NewRecorder()

    // Execute handler
    handler.CreateAccount(w, req)

    // Assert response
    if w.Code != http.StatusCreated {
        t.Errorf("Status = %v, want %v", w.Code, http.StatusCreated)
    }

    // Parse response body
    var response domain.Account
    if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
        t.Fatalf("Failed to decode response: %v", err)
    }

    if response.Name != "Test" {
        t.Errorf("Name = %v, want Test", response.Name)
    }
}
```

## Test Organization

### Subtests

```go
func TestAccountService(t *testing.T) {
    t.Run("Create", func(t *testing.T) {
        // Create tests
    })

    t.Run("Update", func(t *testing.T) {
        // Update tests
    })

    t.Run("Delete", func(t *testing.T) {
        // Delete tests
    })
}
```

### Test Fixtures

```go
// fixtures_test.go
var (
    validAccount = &domain.Account{
        ID:      "test-id",
        Name:    "Test Account",
        Type:    "checking",
        Balance: 100000,
    }

    invalidAccount = &domain.Account{
        Name: "",  // Invalid: empty name
    }
)
```

## Assertions

### Basic Assertions

```go
// Equality
if got != want {
    t.Errorf("got %v, want %v", got, want)
}

// Error checking
if err != nil {
    t.Errorf("unexpected error: %v", err)
}
if err == nil {
    t.Error("expected error, got nil")
}

// Error type checking
if !errors.Is(err, ErrNotFound) {
    t.Errorf("expected ErrNotFound, got %v", err)
}

// Nil checking
if result != nil {
    t.Errorf("expected nil, got %v", result)
}
```

### Deep Comparison

```go
import "reflect"

if !reflect.DeepEqual(got, want) {
    t.Errorf("got %+v, want %+v", got, want)
}
```

### Custom Comparison

```go
func compareAccounts(t *testing.T, got, want *Account) {
    t.Helper()

    if got.Name != want.Name {
        t.Errorf("Name: got %v, want %v", got.Name, want.Name)
    }
    if got.Balance != want.Balance {
        t.Errorf("Balance: got %v, want %v", got.Balance, want.Balance)
    }
}
```

## Test Coverage

```bash
# Run tests with coverage
go test ./... -cover

# Generate coverage profile
go test ./... -coverprofile=coverage.out

# View coverage in browser
go tool cover -html=coverage.out

# Get coverage by function
go test ./... -coverprofile=coverage.out
go tool cover -func=coverage.out
```

## Test Best Practices

### DO ✅

- **Use table-driven tests** for multiple scenarios
- **Name tests descriptively**: `TestFunction_Scenario`
- **Test behavior, not implementation**
- **Keep tests simple and focused**
- **Use test helpers** for repeated setup
- **Mark helpers with `t.Helper()`**
- **Test error cases** as well as happy path
- **Use subtests** for organization
- **Mock external dependencies**
- **Make tests independent** (no shared state)

### DON'T ❌

- **Don't use global state** in tests
- **Don't test private functions** directly
- **Don't make tests dependent** on each other
- **Don't skip assertions**
- **Don't test implementation details**
- **Don't make tests too complex**
- **Don't ignore test failures**

## Common Test Patterns

### Testing Error Cases

```go
func TestService_ErrorHandling(t *testing.T) {
    tests := []struct {
        name    string
        setup   func() error
        wantErr bool
    }{
        {
            name: "database error",
            setup: func() error {
                return errors.New("database error")
            },
            wantErr: true,
        },
        {
            name: "success",
            setup: func() error {
                return nil
            },
            wantErr: false,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            err := tt.setup()
            if (err != nil) != tt.wantErr {
                t.Errorf("error = %v, wantErr %v", err, tt.wantErr)
            }
        })
    }
}
```

### Testing with Context

```go
func TestService_WithTimeout(t *testing.T) {
    ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
    defer cancel()

    err := service.DoSomething(ctx)

    if err != nil {
        t.Errorf("unexpected error: %v", err)
    }
}
```

### Parallel Tests

```go
func TestParallel(t *testing.T) {
    tests := []struct {
        name string
        // ...
    }{
        // test cases
    }

    for _, tt := range tests {
        tt := tt  // Capture range variable
        t.Run(tt.name, func(t *testing.T) {
            t.Parallel()  // Run in parallel
            // Test implementation
        })
    }
}
```

## Budget App Testing Examples

### Testing Allocation Service

```go
func TestAllocationService_CalculateReadyToAssign(t *testing.T) {
    mockAccountRepo := &MockAccountRepository{
        accounts: []*domain.Account{
            {Balance: 500000},  // $5,000
            {Balance: 200000},  // $2,000
        },
    }

    mockAllocationRepo := &MockAllocationRepository{
        allocations: []*domain.Allocation{
            {Amount: 120000},  // $1,200
            {Amount: 50000},   // $500
        },
    }

    service := NewAllocationService(mockAccountRepo, mockAllocationRepo)

    readyToAssign, err := service.GetReadyToAssign()

    if err != nil {
        t.Fatalf("unexpected error: %v", err)
    }

    expected := 530000  // $5,300
    if readyToAssign != expected {
        t.Errorf("ReadyToAssign = %d, want %d", readyToAssign, expected)
    }
}
```

### Testing Transaction Balance Updates

```go
func TestTransactionService_Create_UpdatesBalance(t *testing.T) {
    db := setupTestDB(t)
    defer db.Close()

    accountRepo := repository.NewAccountRepository(db)
    txnRepo := repository.NewTransactionRepository(db)
    service := NewTransactionService(txnRepo, accountRepo)

    // Create test account
    account := &domain.Account{
        ID:      uuid.New().String(),
        Balance: 100000,  // $1,000
    }
    accountRepo.Create(account)

    // Create transaction
    txn := &domain.Transaction{
        ID:        uuid.New().String(),
        AccountID: account.ID,
        Amount:    -50000,  // Spend $500
    }

    err := service.CreateTransaction(txn)
    if err != nil {
        t.Fatalf("CreateTransaction failed: %v", err)
    }

    // Verify balance updated
    updated, _ := accountRepo.GetByID(account.ID)
    expected := 50000  // $500 remaining

    if updated.Balance != expected {
        t.Errorf("Balance = %d, want %d", updated.Balance, expected)
    }
}
```

## Running Tests

```bash
# Run all tests
go test ./...

# Run with verbose output
go test -v ./...

# Run specific test
go test -run TestAccountService_Create

# Run tests in specific package
go test ./internal/application

# Run with race detection
go test -race ./...

# Run with coverage
go test -cover ./...
```
