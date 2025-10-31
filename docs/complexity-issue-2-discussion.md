# Issue #2: Multiple Calculation Methods for Same Data

**Priority:** 🔴 High
**Status:** 📋 Not Started
**Location:** `internal/application/allocation_service.go`

---

## Problem Statement

Aggregation calculations (sums, totals) are performed in-memory by iterating through all records, resulting in:
- O(n) complexity for every request
- Duplicated aggregation logic across methods
- Inefficient use of database capabilities

---

## Current Implementation

### Example 1: GetAllocationSummary (Line 264-275)
```go
// Get all allocations for this category across all periods
allAllocations, err := s.allocationRepo.List(ctx)
if err != nil {
    continue
}

var totalAllocated int64
for _, alloc := range allAllocations {
    if alloc.CategoryID == category.ID {
        totalAllocated += alloc.Amount
    }
}
```

### Example 2: CalculateReadyToAssignForPeriod (Line 384-397)
```go
// Get all transactions to calculate inflows
allTransactions, err := s.transactionRepo.List(ctx)
if err != nil {
    return 0, fmt.Errorf("failed to list transactions: %w", err)
}

// Calculate total inflows through this period
var totalInflows int64
for _, txn := range allTransactions {
    txnPeriod := txn.Date.Format("2006-01")
    if txn.Amount > 0 && txnPeriod <= period && txn.Type != "transfer" {
        totalInflows += txn.Amount
    }
}
```

---

## Issues Identified

### 1. Performance Problems
- **Loads ALL records** into memory on every request
- **O(n) iteration** through all records to calculate sums
- **No caching** - recalculates from scratch every time
- **Scalability issue** - Performance degrades as data grows

### 2. Code Duplication
Same aggregation patterns repeated in:
- `GetAllocationSummary` (lines 264-290)
- `CalculateReadyToAssignForPeriod` (lines 384-427)
- `syncPaymentCategoryAllocations` (lines 138-148, 186-221)

### 3. Database Inefficiency
- Database can calculate SUM() in O(1) with indexes
- Network overhead transferring all records
- Memory overhead storing all records

---

## Proposed Solution

Move aggregations to the database layer using SQL.

### New Repository Methods

```go
// In domain/repository.go - Add to interfaces

type AllocationRepository interface {
    // ... existing methods ...

    // New methods
    GetTotalAllocatedByCategory(ctx context.Context, categoryID string) (int64, error)
    GetTotalAllocatedByCategoryAndPeriod(ctx context.Context, categoryID, period string) (int64, error)
    GetTotalAllocationsThroughPeriod(ctx context.Context, period string) (int64, error)
    GetTotalAllocationsExcludingCategories(ctx context.Context, period string, excludeCategoryIDs []string) (int64, error)
}

type TransactionRepository interface {
    // ... existing methods ...

    // New methods
    GetTotalSpentByCategory(ctx context.Context, categoryID string) (int64, error)
    GetTotalInflowsThroughPeriod(ctx context.Context, period string) (int64, error)
    GetCategoryActivityForPeriod(ctx context.Context, categoryID, period string) (int64, error)
}
```

### Example SQL Implementation

```sql
-- GetTotalAllocatedByCategory
SELECT COALESCE(SUM(amount), 0)
FROM allocations
WHERE category_id = ?;

-- GetTotalSpentByCategory
SELECT COALESCE(SUM(ABS(amount)), 0)
FROM transactions
WHERE category_id = ? AND amount < 0;

-- GetTotalInflowsThroughPeriod
SELECT COALESCE(SUM(amount), 0)
FROM transactions
WHERE amount > 0
  AND type != 'transfer'
  AND strftime('%Y-%m', date) <= ?;
```

---

## Benefits

### Performance
- **O(n) → O(1)**: Database index scan instead of full table scan
- **Reduced memory**: No need to load all records
- **Reduced network**: Only transfer the sum, not all records

### Code Quality
- **DRY principle**: No repeated aggregation logic
- **Clarity**: Intent is clear in repository method names
- **Testability**: Can test aggregations independently

### Scalability
- Works efficiently with 100 records or 1,000,000 records
- Database optimizations (indexes) apply automatically

---

## Discussion Notes

### Session 1: [Date]
*Discussion notes will be added here as we discuss this issue*

---

## Questions to Address

1. **Q:** Should we add database indexes to support these aggregations?
   - **A:** [To be discussed]

2. **Q:** Do we need caching on top of database aggregations?
   - **A:** [To be discussed]

3. **Q:** How do we handle timezone considerations in period calculations?
   - **A:** [To be discussed]

4. **Q:** Should we batch multiple aggregations into a single query?
   - **A:** [To be discussed]

---

## Implementation Checklist

### Phase 1: Repository Layer
- [ ] Add new methods to repository interfaces
- [ ] Implement SQL queries in SQLite repository
- [ ] Add unit tests for new repository methods
- [ ] Add database indexes if needed

### Phase 2: Service Layer Refactoring
- [ ] Update `GetAllocationSummary` to use new methods
- [ ] Update `CalculateReadyToAssignForPeriod` to use new methods
- [ ] Update any other methods using in-memory aggregation
- [ ] Add/update integration tests

### Phase 3: Optimization
- [ ] Benchmark before/after performance
- [ ] Document performance improvements
- [ ] Consider caching strategy if needed

---

## Database Indexes Needed

```sql
-- For allocation aggregations
CREATE INDEX IF NOT EXISTS idx_allocations_category
ON allocations(category_id);

CREATE INDEX IF NOT EXISTS idx_allocations_period
ON allocations(period);

CREATE INDEX IF NOT EXISTS idx_allocations_category_period
ON allocations(category_id, period);

-- For transaction aggregations
CREATE INDEX IF NOT EXISTS idx_transactions_category
ON transactions(category_id);

CREATE INDEX IF NOT EXISTS idx_transactions_date
ON transactions(date);

CREATE INDEX IF NOT EXISTS idx_transactions_type
ON transactions(type);
```

---

## Migration Strategy

### Option 1: Gradual Migration
1. Add new repository methods
2. Update one service method at a time
3. Test each change independently

### Option 2: Big Bang
1. Add all repository methods
2. Update all service methods at once
3. Comprehensive testing before deployment

**Recommended:** Option 1 (Gradual)

---

## Performance Benchmarks

### Before (Current Implementation)

| Operation | Records | Time | Memory |
|-----------|---------|------|--------|
| GetAllocationSummary | 1,000 | TBD | TBD |
| GetAllocationSummary | 10,000 | TBD | TBD |
| CalculateRTA | 1,000 | TBD | TBD |
| CalculateRTA | 10,000 | TBD | TBD |

### After (Database Aggregation)

| Operation | Records | Time | Memory |
|-----------|---------|------|--------|
| GetAllocationSummary | 1,000 | TBD | TBD |
| GetAllocationSummary | 10,000 | TBD | TBD |
| CalculateRTA | 1,000 | TBD | TBD |
| CalculateRTA | 10,000 | TBD | TBD |

---

## Related Issues

- Issue #3: Underfunded Category Detection (also uses in-memory aggregation)
- Issue #1: Payment Category Syncing (uses in-memory aggregation)

---

## Decision Log

| Date | Decision | Rationale |
|------|----------|-----------|
| TBD  | TBD      | TBD       |

---

**Last Updated:** 2025-10-31
