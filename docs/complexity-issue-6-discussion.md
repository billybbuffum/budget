# Issue #6: Unclear Transfer to CC Logic

**Priority:** 🟢 Low
**Status:** 📋 Not Started
**Location:** `internal/application/transaction_service.go:223-261`

---

## Problem Statement

When transferring money TO a credit card, there's complex conditional logic about when to categorize the transfer with the payment category:

- Only categorizes if payment category has allocated funds
- Only categorizes if available >= transfer amount
- Results in asymmetric behavior (transfers TO cc are special, FROM cc aren't)

This creates:
- Unclear business intent
- Additional complexity (~40 lines of code)
- Inconsistent user experience
- Hard to explain to users

---

## Current Implementation

```go
// Lines 223-261 in transaction_service.go
func (s *TransactionService) CreateTransfer(...) (*domain.Transaction, error) {
    // ... validation ...

    // If transferring TO a credit card, check if we should categorize with payment category
    // Only categorize if there's money allocated (don't categorize overpayments)
    var outboundCategoryID *string
    if toAccount.Type == domain.AccountTypeCredit {
        paymentCategory, err := s.categoryRepo.GetPaymentCategoryByAccountID(ctx, toAccountID)
        if err == nil && paymentCategory != nil {
            // Check if payment category has any allocation
            // Get all allocations for this payment category
            allAllocations, err := s.allocationRepo.List(ctx)
            if err == nil {
                var totalAllocated int64
                for _, alloc := range allAllocations {
                    if alloc.CategoryID == paymentCategory.ID {
                        totalAllocated += alloc.Amount
                    }
                }

                // Get all transactions already categorized with this payment category
                allTransactions, err := s.transactionRepo.ListByCategory(ctx, paymentCategory.ID)
                if err == nil {
                    var totalSpent int64
                    for _, txn := range allTransactions {
                        if txn.Amount < 0 {
                            totalSpent += -txn.Amount // Convert to positive
                        }
                    }

                    // Available = Allocated - Already Spent
                    available := totalAllocated - totalSpent

                    // Only categorize if payment <= available
                    // This prevents showing negative available when overpaying
                    if available >= amount {
                        outboundCategoryID = &paymentCategory.ID
                    }
                }
            }
        }
    }

    // Create outbound transaction with conditional category
    outboundTxn := &domain.Transaction{
        // ...
        CategoryID: outboundCategoryID, // Might be nil!
        // ...
    }
    // ...
}
```

---

## Issues Identified

### 1. Unclear Business Intent

**Current behavior:**
- Transfer $100 to CC with $100 allocated → Categorized ✅
- Transfer $150 to CC with $100 allocated → NOT categorized ❌

**Questions:**
- Why is overpayment treated differently?
- What should the user expect?
- How does this align with zero-based budgeting?

### 2. Complexity

- ~40 lines of code
- Nested conditionals
- Multiple repository queries
- In-memory aggregation (also Issue #2 problem)

### 3. Asymmetric Behavior

- Transfers TO credit cards: Special logic
- Transfers FROM credit cards: No special logic
- Transfers between checking accounts: No special logic

This asymmetry suggests the logic might be unnecessary.

### 4. User Experience Issues

**Scenario:** User pays off credit card
- Has $500 debt
- Allocated $400 to payment category
- Transfers $500 (full payment)
- Result: Transfer is NOT categorized

**Problem:** User's "available" in payment category is still $400, but they paid $500. Where did the extra $100 come from in the budget?

---

## Proposed Solution

**Simplify:** ALL transfers to credit cards should be categorized with the payment category.

### Simplified Implementation

```go
// Simplified version
var outboundCategoryID *string
if toAccount.Type == domain.AccountTypeCredit {
    paymentCategory, err := s.categoryRepo.GetPaymentCategoryByAccountID(ctx, toAccountID)
    if err == nil && paymentCategory != nil {
        outboundCategoryID = &paymentCategory.ID
    }
}
```

**Reduces ~40 lines to ~5 lines.**

---

## Why This Works (Zero-Based Budgeting Perspective)

### In Zero-Based Budgeting:

1. **Every dollar has a job**
2. **Credit card payments are expenses** (money leaving your budget)
3. **Payment category represents the job:** "Pay off this card"

### Current Logic Problem:

- If you overpay, the money is "uncategorized"
- This violates ZBB principle (money without a job)
- You can't track where that money came from

### Proposed Logic Benefit:

- ALL CC payments are categorized
- Payment category might go negative (overpaid)
- **Negative available = You paid more than budgeted**
- This is actually GOOD feedback for the user!

---

## User Experience Comparison

### Current Behavior

**Scenario:** $500 debt, $400 allocated, pay $500

| Category | Allocated | Activity | Available |
|----------|-----------|----------|-----------|
| CC Payment | $400 | $0 | $400 |

**Problem:** Where did the $500 payment come from? It's invisible!

### Proposed Behavior

| Category | Allocated | Activity | Available |
|----------|-----------|----------|-----------|
| CC Payment | $400 | -$500 | -$100 |

**Benefit:** Clear that you paid $100 more than budgeted. User can see this and decide:
- "I'll allocate $100 to cover the overpayment"
- "I'll take $100 from another category"

This is MORE aligned with ZBB!

---

## Discussion Notes

### Session 1: [Date]
*Discussion notes will be added here as we discuss this issue*

---

## Questions to Address

1. **Q:** Why was the original logic implemented this way?
   - **A:** [To be discussed - check git history/comments]

2. **Q:** Are there edge cases where NOT categorizing is better?
   - **A:** [To be discussed]

3. **Q:** Should overpayments prompt the user to allocate more?
   - **A:** [To be discussed - UX consideration]

4. **Q:** How do other budgeting apps (YNAB, etc.) handle this?
   - **A:** [To be discussed]

---

## Implementation Checklist

### Phase 1: Analysis
- [ ] Research why original logic was implemented
- [ ] Check git history for context
- [ ] Survey other budgeting apps' behavior
- [ ] Confirm no hidden edge cases

### Phase 2: Code Changes
- [ ] Simplify CreateTransfer logic
- [ ] Remove allocation/spending calculation
- [ ] Update tests
- [ ] Add test for overpayment scenario

### Phase 3: Documentation
- [ ] Update user documentation
- [ ] Explain overpayment behavior
- [ ] Add to FAQ if needed

---

## Code Changes

### Before (Lines 223-261)
40+ lines of complex conditional logic

### After (Proposed)
```go
// If transferring TO a credit card, categorize with payment category
var outboundCategoryID *string
if toAccount.Type == domain.AccountTypeCredit {
    paymentCategory, err := s.categoryRepo.GetPaymentCategoryByAccountID(ctx, toAccountID)
    if err == nil && paymentCategory != nil {
        outboundCategoryID = &paymentCategory.ID
    }
}

// Create outbound transaction
outboundTxn := &domain.Transaction{
    ID:                  uuid.New().String(),
    Type:                domain.TransactionTypeTransfer,
    AccountID:           fromAccountID,
    TransferToAccountID: &toAccountID,
    CategoryID:          outboundCategoryID,
    Amount:              -amount,
    Description:         description,
    Date:                date,
    CreatedAt:           time.Now(),
    UpdatedAt:           time.Now(),
}
```

**Line count:** 5 lines instead of 40+

---

## Testing Strategy

### Test Cases to Add

```go
func TestCreateTransfer_ToCreditCard_FullyAllocated(t *testing.T) {
    // Transfer = Allocated
    // Should categorize, available = 0
}

func TestCreateTransfer_ToCreditCard_Overpayment(t *testing.T) {
    // Transfer > Allocated
    // Should categorize, available = negative
}

func TestCreateTransfer_ToCreditCard_Underpayment(t *testing.T) {
    // Transfer < Allocated
    // Should categorize, available = positive
}

func TestCreateTransfer_ToCreditCard_NoAllocation(t *testing.T) {
    // No allocation yet
    // Should categorize, available = negative
}
```

---

## UI Considerations

If implementing this change, the UI should:

1. **Show negative available clearly**
   - Example: "CC Payment: Available -$100" in red

2. **Provide user guidance**
   - Tooltip: "You paid $100 more than budgeted. Consider allocating funds to cover this."

3. **Allow easy reallocation**
   - Button: "Allocate to cover overpayment"

---

## Alternative Solutions Considered

### Alternative 1: Keep Current Logic
- **Pros:** No changes needed
- **Cons:** Complexity remains, violates ZBB principles
- **Decision:** Not recommended

### Alternative 2: Prompt User During Transfer
- Ask user: "Categorize this payment?"
- **Pros:** User control
- **Cons:** Extra friction, most users will always say yes
- **Decision:** Not recommended (over-engineering)

### Alternative 3: Two-Step Process
- Transfer money first (uncategorized)
- Then categorize separately
- **Pros:** Maximum flexibility
- **Cons:** Extra steps, easy to forget
- **Decision:** Not recommended

---

## Impact Assessment

### Code Impact
- **Lines Removed:** ~35
- **Complexity Reduced:** Yes
- **Performance Improved:** Yes (fewer queries)

### User Impact
- **Breaking Change:** No (behavior improves)
- **Documentation Needed:** Yes
- **Migration Needed:** No

### Business Logic Impact
- **Aligns Better with ZBB:** Yes
- **More Predictable:** Yes
- **Easier to Explain:** Yes

---

## Related Issues

- Issue #2: Database Aggregations (current code does in-memory aggregation)
- Issue #4: Transaction Side Effects (this is part of CreateTransfer)

---

## Decision Log

| Date | Decision | Rationale |
|------|----------|-----------|
| TBD  | TBD      | TBD       |

---

**Last Updated:** 2025-10-31
