# Issue #4: Transaction Creation Side Effects

**Priority:** 🟡 Medium
**Status:** 📋 Not Started
**Location:** `internal/application/transaction_service.go:38-198`

---

## Problem Statement

The `CreateTransaction` method has multiple side effects with manual rollback logic:

1. Creates transaction record
2. Updates account balance
3. For CC transactions: calculates expense category available budget
4. Creates/updates payment category allocation

This violates the Single Responsibility Principle and makes the code:
- Hard to reason about
- Error-prone (manual rollbacks)
- Not atomic (partial failures possible)
- Difficult to test

---

## Current Implementation

```go
func (s *TransactionService) CreateTransaction(...) (*domain.Transaction, error) {
    // 1. Create transaction
    if err := s.transactionRepo.Create(ctx, transaction); err != nil {
        return nil, err
    }

    // 2. Update account balance
    account.Balance += amount
    if err := s.accountRepo.Update(ctx, account); err != nil {
        // MANUAL ROLLBACK #1
        s.transactionRepo.Delete(ctx, transaction.ID)
        return nil, fmt.Errorf("failed to update account balance: %w", err)
    }

    // 3. For CC transactions, handle payment category allocation
    if account.Type == domain.AccountTypeCredit && categoryID != nil && amount < 0 {
        paymentCategory, err := s.categoryRepo.GetPaymentCategoryByAccountID(ctx, account.ID)
        if err != nil {
            // MANUAL ROLLBACK #2
            s.accountRepo.Update(ctx, &domain.Account{
                ID:        account.ID,
                Balance:   account.Balance - amount,
                UpdatedAt: time.Now(),
            })
            s.transactionRepo.Delete(ctx, transaction.ID)
            return nil, fmt.Errorf("failed to get payment category: %w", err)
        }

        // ... calculate amount to move (lines 110-152) ...

        if amountToMove > 0 {
            if err := s.allocationRepo.Create(ctx, paymentAlloc); err != nil {
                // MANUAL ROLLBACK #3
                s.accountRepo.Update(ctx, &domain.Account{
                    ID:        account.ID,
                    Balance:   account.Balance - amount,
                    UpdatedAt: time.Now(),
                })
                s.transactionRepo.Delete(ctx, transaction.ID)
                return nil, fmt.Errorf("failed to create payment allocation: %w", err)
            }
        }
    }

    return transaction, nil
}
```

---

## Issues Identified

### 1. Not Atomic
- If the process crashes between steps, data is inconsistent
- No database transaction wrapping the operations
- Partial failures leave the system in invalid state

### 2. Manual Rollback is Error-Prone
- Must manually reverse each operation
- Easy to forget a rollback step
- Rollback itself can fail, leaving inconsistent state
- Example: Lines 87, 98-103, 170-177, 185-192

### 3. Too Many Responsibilities
One method does:
- Transaction validation
- Transaction creation
- Account balance updates
- Payment category logic
- Allocation creation/updates

### 4. Hard to Test
- Must mock multiple repositories
- Hard to test failure scenarios
- Can't test steps independently

### 5. Code Duplication
Similar pattern in `CreateTransfer` (lines 200-327) with same issues

---

## Proposed Solutions

### Solution 1: Database Transactions (Recommended)

Wrap all operations in a database transaction for automatic rollback:

```go
func (s *TransactionService) CreateTransaction(...) (*domain.Transaction, error) {
    // Begin database transaction
    tx, err := s.db.BeginTx(ctx, nil)
    if err != nil {
        return nil, err
    }
    defer tx.Rollback() // Auto-rollback if not committed

    // All operations use tx instead of ctx
    if err := s.transactionRepo.CreateTx(tx, transaction); err != nil {
        return nil, err // Auto-rollback
    }

    if err := s.accountRepo.UpdateTx(tx, account); err != nil {
        return nil, err // Auto-rollback
    }

    if /* CC logic */ {
        if err := s.allocationRepo.CreateTx(tx, allocation); err != nil {
            return nil, err // Auto-rollback
        }
    }

    // Commit only if all succeeded
    if err := tx.Commit(); err != nil {
        return nil, err
    }

    return transaction, nil
}
```

**Pros:**
- Truly atomic operations
- Automatic rollback on any error
- No manual rollback code needed
- Standard database pattern

**Cons:**
- Requires repository refactoring to support transactions
- More complex setup initially

---

### Solution 2: Composable Functions

Break into smaller functions that can be composed:

```go
// Core transaction creation
func (s *TransactionService) createTransactionRecord(ctx, tx) error

// Account balance update
func (s *TransactionService) updateAccountBalance(ctx, accountID, delta) error

// Payment category allocation
func (s *TransactionService) handleCCPaymentAllocation(ctx, transaction) error

// Orchestrator
func (s *TransactionService) CreateTransaction(...) (*domain.Transaction, error) {
    tx, _ := s.db.BeginTx(ctx, nil)
    defer tx.Rollback()

    if err := s.createTransactionRecord(tx, ...); err != nil {
        return nil, err
    }

    if err := s.updateAccountBalance(tx, ...); err != nil {
        return nil, err
    }

    if needsCCHandling {
        if err := s.handleCCPaymentAllocation(tx, ...); err != nil {
            return nil, err
        }
    }

    return transaction, tx.Commit()
}
```

**Pros:**
- Better separation of concerns
- Each function is testable independently
- Still gets transaction safety

**Cons:**
- More functions to maintain
- Still requires repository refactoring

---

### Solution 3: Event Sourcing (Future)

Record events and apply them atomically:

**Pros:**
- Full audit trail
- Easier to test
- Can replay events

**Cons:**
- Major architectural change
- Overkill for current needs

**Decision:** Not recommended for now

---

## Recommended Approach

**Use Solution 1 (Database Transactions)** with elements of Solution 2 (composable functions).

### Implementation Steps

1. Add transaction support to repositories
2. Extract CC logic to helper method
3. Wrap in database transaction
4. Remove all manual rollback code
5. Add comprehensive tests

---

## Discussion Notes

### Session 1: [Date]
*Discussion notes will be added here as we discuss this issue*

---

## Questions to Address

1. **Q:** Does SQLite support nested transactions?
   - **A:** [To be discussed]

2. **Q:** Should we use a transaction manager or pass tx explicitly?
   - **A:** [To be discussed]

3. **Q:** How do we handle long-running transactions?
   - **A:** [To be discussed]

4. **Q:** Should CreateTransfer also be refactored?
   - **A:** [To be discussed]

---

## Implementation Checklist

### Phase 1: Repository Support
- [ ] Add `BeginTx()` method to database layer
- [ ] Add transaction-aware methods to each repository
  - [ ] `CreateTx(tx, entity)`
  - [ ] `UpdateTx(tx, entity)`
  - [ ] `DeleteTx(tx, id)`
- [ ] Test transaction rollback behavior

### Phase 2: Service Refactoring
- [ ] Extract CC payment logic to helper method
- [ ] Refactor `CreateTransaction` to use database transactions
- [ ] Remove all manual rollback code
- [ ] Refactor `CreateTransfer` similarly

### Phase 3: Testing
- [ ] Add transaction rollback tests
- [ ] Test partial failure scenarios
- [ ] Performance testing with transactions
- [ ] Integration tests

---

## Code Changes Required

### New Repository Interface Methods

```go
// In domain/repository.go
type Repository interface {
    BeginTx(ctx context.Context) (*sql.Tx, error)
    CreateTx(tx *sql.Tx, entity interface{}) error
    UpdateTx(tx *sql.Tx, entity interface{}) error
    DeleteTx(tx *sql.Tx, id string) error
}
```

### Database Transaction Helper

```go
// In infrastructure/database/sqlite.go
type SQLiteDB struct {
    db *sql.DB
}

func (s *SQLiteDB) BeginTx(ctx context.Context) (*sql.Tx, error) {
    return s.db.BeginTx(ctx, nil)
}

func (s *SQLiteDB) WithTransaction(ctx context.Context, fn func(*sql.Tx) error) error {
    tx, err := s.BeginTx(ctx)
    if err != nil {
        return err
    }

    defer func() {
        if err != nil {
            tx.Rollback()
        }
    }()

    err = fn(tx)
    if err != nil {
        return err
    }

    return tx.Commit()
}
```

---

## Testing Strategy

### Unit Tests

```go
func TestCreateTransaction_RollbackOnAccountUpdateFailure(t *testing.T)
func TestCreateTransaction_RollbackOnAllocationFailure(t *testing.T)
func TestCreateTransaction_CommitsOnSuccess(t *testing.T)
```

### Integration Tests

```go
func TestCreateTransaction_AtomicityUnderConcurrency(t *testing.T)
func TestCreateTransaction_RollbackDoesNotLeaveOrphans(t *testing.T)
```

---

## Performance Considerations

- Database transactions add minimal overhead
- SQLite's transaction model is efficient
- May need to tune transaction isolation level
- Should benchmark before/after

---

## Related Issues

- Issue #1: Payment Category Syncing (interacts with payment allocation logic)

---

## Decision Log

| Date | Decision | Rationale |
|------|----------|-----------|
| TBD  | TBD      | TBD       |

---

**Last Updated:** 2025-10-31
