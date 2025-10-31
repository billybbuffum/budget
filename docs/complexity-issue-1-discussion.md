# Issue #1: Credit Card Payment Category Syncing

**Priority:** 🔴 High
**Status:** 📋 Not Started
**Location:** `internal/application/allocation_service.go:100-225`

---

## Problem Statement

The `syncPaymentCategoryAllocations` function performs retroactive allocation syncing with the following behavior:

1. Runs on EVERY allocation create/update
2. Iterates through ALL credit card accounts
3. For each CC, finds ALL transactions matching the expense category
4. Calculates total CC spending per category
5. Recalculates payment category allocations from scratch

---

## Current Implementation

```go
// Lines 100-225 in allocation_service.go
func (s *AllocationService) syncPaymentCategoryAllocations(ctx context.Context, categoryID string) error {
    // Nested loops through:
    // - All accounts
    // - All transactions per account
    // - All allocations
    // Complexity: O(n²) or worse
}
```

---

## Issues Identified

### 1. Performance Issues
- **Complexity:** O(n²) or O(n³) depending on dataset size
- **Triggered:** On every single allocation create/update
- **Scale:** Gets worse as transactions/allocations grow

### 2. Hidden Side Effects
- Creating an allocation for groceries triggers payment category updates
- User has no visibility into these side effects
- Makes debugging difficult

### 3. Violates Zero-Based Budgeting Principles
- ZBB is forward-looking: "What job does this dollar have?"
- This function is retroactive: "Let me fix past allocations"
- Goes against the core principle

### 4. Weak Error Handling
- Only prints warnings to console
- Doesn't fail the allocation operation
- Could lead to inconsistent state

### 5. Redundant with Existing Logic
- `transaction_service.go:91-195` already handles CC spending at transaction time
- This function re-does work that was already done correctly

---

## Proposed Solution

**REMOVE** the entire `syncPaymentCategoryAllocations` function and its calls.

### Why This Works

1. **Real-time handling exists:** `CreateTransaction` already moves budgeted money to payment category when CC spending occurs
2. **Simpler:** No retroactive logic needed
3. **Faster:** No expensive recalculations
4. **More predictable:** No hidden side effects

### Code to Remove

- Line 100-225: `syncPaymentCategoryAllocations` function definition
- Line 67-69: Call in `CreateAllocation` (update path)
- Line 91-95: Call in `CreateAllocation` (create path)

---

## Discussion Notes

### Session 1: [Date]
*Discussion notes will be added here as we discuss this issue*

---

### Session 2: [Date]
*Additional discussion notes*

---

## Questions to Address

1. **Q:** Are there edge cases where retroactive syncing is needed?
   - **A:** [To be discussed]

2. **Q:** What happens to existing allocations that were created by this sync?
   - **A:** [To be discussed]

3. **Q:** Should we add a migration to clean up orphaned payment allocations?
   - **A:** [To be discussed]

4. **Q:** How do we handle historical data that relied on this logic?
   - **A:** [To be discussed]

---

## Risks & Mitigations

### Risk 1: Breaking Existing Behavior
- **Mitigation:** Add comprehensive tests before removal
- **Mitigation:** Test with production-like data

### Risk 2: Users Expect Retroactive Behavior
- **Mitigation:** Document the change
- **Mitigation:** Provide migration guide if needed

---

## Testing Strategy

### Before Removal
1. Document current behavior with integration tests
2. Identify all scenarios that trigger sync
3. Verify transaction-time logic covers all cases

### After Removal
1. Verify no regression in CC payment tracking
2. Test allocation creation performance improvement
3. Test edge cases (overspending, underfunding, etc.)

---

## Implementation Checklist

- [ ] Review current behavior with stakeholder
- [ ] Create comprehensive test suite
- [ ] Remove function definition
- [ ] Remove all function calls
- [ ] Run full test suite
- [ ] Performance benchmark comparison
- [ ] Update documentation
- [ ] Create migration guide if needed

---

## Alternative Solutions Considered

### Alternative 1: Optimize Instead of Remove
- **Pros:** Less disruptive
- **Cons:** Still violates ZBB principles, still adds complexity
- **Decision:** Not recommended

### Alternative 2: Make Async/Background Job
- **Pros:** Removes performance impact
- **Cons:** Still retroactive, adds deployment complexity
- **Decision:** Not recommended

### Alternative 3: User-Triggered Manual Sync
- **Pros:** User control
- **Cons:** Shouldn't be needed if real-time logic works
- **Decision:** Not recommended

---

## Related Issues

- Issue #4: Transaction Creation Side Effects (contains the real-time logic we're relying on)

---

## Decision Log

| Date | Decision | Rationale |
|------|----------|-----------|
| TBD  | TBD      | TBD       |

---

**Last Updated:** 2025-10-31
