# Issue #3: Underfunded Category Detection Nesting

**Priority:** 🟡 Medium
**Status:** 📋 Not Started
**Location:** `internal/application/allocation_service.go:299-358`

---

## Problem Statement

The underfunded category detection logic is deeply nested (6+ levels) within the `GetAllocationSummary` method, making it:
- Hard to read and understand
- Difficult to test independently
- Mixing concerns (summary calculation + underfunding detection)

---

## Current Implementation

```go
// Lines 299-358 in allocation_service.go
// Inside GetAllocationSummary, within the category loop:

// For payment categories, check if underfunded
var underfunded *int64
var underfundedCategories []string
if category.PaymentForAccountID != nil && *category.PaymentForAccountID != "" {
    // Get the credit card account balance
    account, err := s.accountRepo.GetByID(ctx, *category.PaymentForAccountID)
    if err == nil && account != nil {
        // Credit card balance is negative (you owe money)
        amountOwed := -account.Balance

        if amountOwed > 0 && available < amountOwed {
            // Underfunded: need more money
            shortfall := amountOwed - available
            underfunded = &shortfall

            // Find which expense categories are underfunded
            ccTransactions, err := s.transactionRepo.ListByAccount(ctx, *category.PaymentForAccountID)
            if err == nil {
                // Group by category and calculate spending per category
                categorySpending := make(map[string]int64)
                categoryNames := make(map[string]string)

                for _, txn := range ccTransactions {
                    if txn.CategoryID != nil && *txn.CategoryID != "" && txn.Amount < 0 {
                        categorySpending[*txn.CategoryID] += -txn.Amount

                        if _, exists := categoryNames[*txn.CategoryID]; !exists {
                            cat, err := s.categoryRepo.GetByID(ctx, *txn.CategoryID)
                            if err == nil {
                                categoryNames[*txn.CategoryID] = cat.Name
                            }
                        }
                    }
                }

                // Check each category to see if it has enough allocated
                for catID, spending := range categorySpending {
                    allAllocForCat, err := s.allocationRepo.List(ctx)
                    if err == nil {
                        var catTotalAllocated int64
                        for _, alloc := range allAllocForCat {
                            if alloc.CategoryID == catID {
                                catTotalAllocated += alloc.Amount
                            }
                        }

                        if spending > catTotalAllocated {
                            if name, exists := categoryNames[catID]; exists {
                                underfundedCategories = append(underfundedCategories, name)
                            }
                        }
                    }
                }
            }
        }
    }
}
```

---

## Issues Identified

### 1. Excessive Nesting
- **6+ levels deep** in some places
- Makes code hard to follow
- Hard to see the big picture logic

### 2. Mixed Concerns
`GetAllocationSummary` is doing TWO things:
- Calculating allocation summaries
- Detecting underfunded categories

Should be ONE responsibility per function.

### 3. Repeated Queries
- Calls `s.allocationRepo.List(ctx)` inside a loop
- Inefficient - same data fetched multiple times

### 4. Hard to Test
- Can't test underfunding logic independently
- Must test through `GetAllocationSummary`
- Makes unit testing difficult

### 5. Poor Readability
- Business logic is obscured by nesting
- Hard for new developers to understand

---

## Proposed Solution

Extract underfunding detection to a separate method:

```go
// CalculateUnderfundedInfo calculates if a payment category is underfunded
// and which expense categories contribute to the underfunding
func (s *AllocationService) CalculateUnderfundedInfo(
    ctx context.Context,
    paymentCategory *domain.Category,
    available int64,
) (*int64, []string, error) {
    // Clean, testable implementation
    // Returns: (shortfall amount, underfunded category names, error)
}
```

### Benefits

1. **Reduced Nesting**: Main function becomes much flatter
2. **Single Responsibility**: Each function has one clear purpose
3. **Testable**: Can unit test underfunding logic independently
4. **Reusable**: Can use this logic elsewhere if needed
5. **Readable**: Intent is clear from function name

---

## Refactored Implementation

### New Method

```go
// CalculateUnderfundedInfo checks if a payment category is underfunded
func (s *AllocationService) CalculateUnderfundedInfo(
    ctx context.Context,
    paymentCategory *domain.Category,
    available int64,
) (*int64, []string, error) {
    if paymentCategory.PaymentForAccountID == nil {
        return nil, nil, nil
    }

    // Get credit card account
    account, err := s.accountRepo.GetByID(ctx, *paymentCategory.PaymentForAccountID)
    if err != nil {
        return nil, nil, err
    }

    // Calculate amount owed
    amountOwed := -account.Balance
    if amountOwed <= 0 || available >= amountOwed {
        return nil, nil, nil // Not underfunded
    }

    // Calculate shortfall
    shortfall := amountOwed - available

    // Find underfunded expense categories
    underfundedCategories, err := s.findUnderfundedExpenseCategories(ctx, account.ID)
    if err != nil {
        return &shortfall, nil, err
    }

    return &shortfall, underfundedCategories, nil
}

// findUnderfundedExpenseCategories identifies which expense categories
// on a credit card have spending exceeding their allocation
func (s *AllocationService) findUnderfundedExpenseCategories(
    ctx context.Context,
    accountID string,
) ([]string, error) {
    // Get all transactions on this CC
    ccTransactions, err := s.transactionRepo.ListByAccount(ctx, accountID)
    if err != nil {
        return nil, err
    }

    // Group spending by category
    categorySpending := make(map[string]int64)
    categoryNames := make(map[string]string)

    for _, txn := range ccTransactions {
        if txn.CategoryID != nil && *txn.CategoryID != "" && txn.Amount < 0 {
            categorySpending[*txn.CategoryID] += -txn.Amount

            if _, exists := categoryNames[*txn.CategoryID]; !exists {
                cat, err := s.categoryRepo.GetByID(ctx, *txn.CategoryID)
                if err == nil {
                    categoryNames[*txn.CategoryID] = cat.Name
                }
            }
        }
    }

    // Check each category's allocation vs spending
    // TODO: This should use Issue #2's database aggregations for efficiency
    var underfunded []string
    allAllocations, err := s.allocationRepo.List(ctx)
    if err != nil {
        return nil, err
    }

    for catID, spending := range categorySpending {
        var catTotalAllocated int64
        for _, alloc := range allAllocations {
            if alloc.CategoryID == catID {
                catTotalAllocated += alloc.Amount
            }
        }

        if spending > catTotalAllocated {
            if name, exists := categoryNames[catID]; exists {
                underfunded = append(underfunded, name)
            }
        }
    }

    return underfunded, nil
}
```

### Updated GetAllocationSummary

```go
// In GetAllocationSummary, replace lines 299-358 with:

// For payment categories, check if underfunded
var underfunded *int64
var underfundedCategories []string

if category.PaymentForAccountID != nil && *category.PaymentForAccountID != "" {
    underfunded, underfundedCategories, err = s.CalculateUnderfundedInfo(ctx, category, available)
    if err != nil {
        // Log error but continue
        fmt.Printf("Warning: failed to calculate underfunded info: %v\n", err)
    }
}
```

---

## Discussion Notes

### Session 1: [Date]
*Discussion notes will be added here as we discuss this issue*

---

## Questions to Address

1. **Q:** Should underfunded calculation be async/cached?
   - **A:** [To be discussed]

2. **Q:** Do we need to sort underfunded categories?
   - **A:** [To be discussed]

3. **Q:** Should this integrate with Issue #2's database aggregations?
   - **A:** [To be discussed]

4. **Q:** Is there a better data structure for the result?
   - **A:** [To be discussed]

---

## Implementation Checklist

- [ ] Create `CalculateUnderfundedInfo` method
- [ ] Create `findUnderfundedExpenseCategories` helper method
- [ ] Update `GetAllocationSummary` to use new methods
- [ ] Add unit tests for `CalculateUnderfundedInfo`
- [ ] Add unit tests for `findUnderfundedExpenseCategories`
- [ ] Update integration tests
- [ ] Consider integration with Issue #2's optimizations

---

## Testing Strategy

### New Unit Tests Needed

```go
func TestCalculateUnderfundedInfo_NotUnderfunded(t *testing.T)
func TestCalculateUnderfundedInfo_Underfunded(t *testing.T)
func TestCalculateUnderfundedInfo_NoPaymentCategory(t *testing.T)
func TestFindUnderfundedExpenseCategories_AllCategoriesFullyFunded(t *testing.T)
func TestFindUnderfundedExpenseCategories_SomeUnderfunded(t *testing.T)
```

---

## Code Metrics

### Before

- **Cyclomatic Complexity**: ~15 (in GetAllocationSummary)
- **Max Nesting Depth**: 7
- **Lines of Code**: ~60 (for underfunding logic)

### After

- **Cyclomatic Complexity**: ~3 (in GetAllocationSummary), ~5 (in new methods)
- **Max Nesting Depth**: 3
- **Lines of Code**: ~30 (main method), ~30 (helper)

---

## Related Issues

- Issue #2: Database Aggregations (can optimize the allocation queries)

---

## Decision Log

| Date | Decision | Rationale |
|------|----------|-----------|
| TBD  | TBD      | TBD       |

---

**Last Updated:** 2025-10-31
