# Specification: Remove Credit Card Payment Syncing and Add Manual Allocation Helper

**Issue:** #1 - Credit Card Payment Category Syncing
**Type:** Refactoring + Feature Enhancement
**Priority:** High
**Estimated Effort:** 4-6 hours
**Status:** Approved for Implementation

---

## Executive Summary

Remove the complex retroactive credit card payment syncing logic and replace it with a user-friendly UI that helps users consciously allocate money to cover underfunded credit card spending. This change:
- Removes ~125 lines of complex code
- Improves performance (eliminates O(n²) operations on every allocation)
- Aligns with zero-based budgeting principles
- Improves user experience through transparency

---

## Problem Statement

### Current Behavior (Problems)

1. **Hidden Complexity:** When users allocate money to any expense category, a hidden `syncPaymentCategoryAllocations` function runs in the background, scanning all credit cards and retroactively adjusting payment category allocations.

2. **Performance Issues:** The sync function has O(n²) or worse complexity and runs on EVERY allocation create/update, regardless of whether it's needed.

3. **Violates ZBB Principles:** Zero-based budgeting is forward-looking ("what job does this dollar have?"), but this function is retroactive ("let me fix past allocations").

4. **User Confusion:** Users don't know when/why payment category allocations change automatically.

### Desired Behavior

1. **Real-time handling:** Credit card spending moves budget from expense categories to payment categories at transaction time (already implemented in `CreateTransaction`)

2. **Transparent underfunding:** When credit cards are underfunded, clearly show this to the user

3. **User control:** Provide a simple UI button to help users allocate money to cover underfunded credit cards

---

## Requirements

### Functional Requirements

#### FR1: Remove Retroactive Syncing
- **MUST** remove `syncPaymentCategoryAllocations` function from `allocation_service.go`
- **MUST** remove all calls to this function
- **MUST** maintain existing real-time allocation logic in `CreateTransaction`

#### FR2: Underfunded CC Detection
- **MUST** continue to calculate underfunded payment categories in `GetAllocationSummary`
- **MUST** return `underfunded` amount (shortfall)
- **MUST** return `underfunded_categories` (list of expense category names)

#### FR3: Manual Allocation Helper API
- **MUST** create new endpoint: `POST /api/allocations/cover-underfunded`
- **MUST** accept payment category ID and period
- **MUST** automatically allocate the underfunded amount to the payment category
- **MUST** return updated allocation

#### FR4: UI Helper Button
- **MUST** show underfunded warning in budget UI when payment category is underfunded
- **MUST** display clear message: "Your [Card Name] is underfunded by $X.XX"
- **MUST** show which expense categories contributed to underfunding
- **MUST** provide button: "Allocate to Cover"
- **MUST** update UI immediately after allocation

### Non-Functional Requirements

#### NFR1: Performance
- Allocation creation **MUST** be faster (no full database scans)
- **MUST** benchmark before/after to verify improvement

#### NFR2: Backward Compatibility
- **MUST NOT** break existing API contracts
- **MUST NOT** require database migration (data structure unchanged)

#### NFR3: Testing
- **MUST** have unit tests for all new functions
- **MUST** have integration tests for API endpoint
- **MUST** have UI tests for the helper button

---

## User Stories

### Story 1: Normal ZBB Flow (No Change)
**As a** budget-conscious user
**I want** my credit card payments to be automatically tracked when I make purchases
**So that** I don't have to manually move budget allocations

**Acceptance Criteria:**
- Given I have $500 allocated to "Groceries"
- When I spend $50 at the store with my credit card
- Then $50 is automatically moved from "Groceries" to "Credit Card Payment" allocation
- And my available in "Groceries" shows $450
- And my "Credit Card Payment" shows $50 allocated

### Story 2: Underfunded Credit Card (New Feature)
**As a** user who spent on credit before budgeting
**I want** to see clearly that my credit card is underfunded
**So that** I can allocate money to cover it

**Acceptance Criteria:**
- Given I spent $50 on "Groceries" with my credit card before allocating budget
- When I view the budget page
- Then I see an underfunded warning: "Your Visa is underfunded by $50.00"
- And it shows "Underfunded categories: Groceries"
- And I see a button "Allocate to Cover"

### Story 3: Quick Fix with Helper Button (New Feature)
**As a** user with an underfunded credit card
**I want** a quick way to allocate money to cover it
**So that** I don't have to manually calculate and allocate

**Acceptance Criteria:**
- Given my Visa is underfunded by $50
- And I have $50+ in "Ready to Assign"
- When I click "Allocate to Cover"
- Then $50 is allocated to "Visa Payment" category
- And my "Ready to Assign" decreases by $50
- And the underfunded warning disappears
- And I see a success message

### Story 4: Insufficient Funds (New Feature)
**As a** user with an underfunded credit card but insufficient "Ready to Assign"
**I want** to be informed that I need more budget
**So that** I can make an informed decision

**Acceptance Criteria:**
- Given my Visa is underfunded by $100
- But I only have $50 in "Ready to Assign"
- When I click "Allocate to Cover"
- Then I see an error: "Insufficient funds. You have $50.00 available but need $100.00. Please move money from another category or add income."
- And no allocation is created

---

## Technical Specification

### Backend Changes

#### 1. Remove Syncing Logic

**File:** `internal/application/allocation_service.go`

**Actions:**
- Delete function `syncPaymentCategoryAllocations` (lines 100-225)
- Remove call in `CreateAllocation` (lines 67-69)
- Remove call in `CreateAllocation` (lines 91-95)

**Before:**
```go
// Line 67-69
if err := s.syncPaymentCategoryAllocations(ctx, categoryID); err != nil {
    fmt.Printf("Warning: failed to sync payment category allocations: %v\n", err)
}
```

**After:**
```go
// Remove entirely
```

#### 2. Add Helper Method

**File:** `internal/application/allocation_service.go`

**New Method:**
```go
// AllocateToCoverUnderfunded allocates money to a payment category to cover its underfunded amount
// This is a user-initiated action to resolve underfunded credit cards
func (s *AllocationService) AllocateToCoverUnderfunded(
    ctx context.Context,
    paymentCategoryID string,
    period string,
) (*domain.Allocation, error) {
    // 1. Verify this is a payment category
    category, err := s.categoryRepo.GetByID(ctx, paymentCategoryID)
    if err != nil {
        return nil, fmt.Errorf("category not found: %w", err)
    }

    if category.PaymentForAccountID == nil || *category.PaymentForAccountID == "" {
        return nil, fmt.Errorf("category is not a payment category")
    }

    // 2. Calculate current underfunded amount
    summaries, err := s.GetAllocationSummary(ctx, period)
    if err != nil {
        return nil, fmt.Errorf("failed to get allocation summary: %w", err)
    }

    var underfundedAmount int64
    for _, summary := range summaries {
        if summary.Category.ID == paymentCategoryID {
            if summary.Underfunded == nil || *summary.Underfunded <= 0 {
                return nil, fmt.Errorf("payment category is not underfunded")
            }
            underfundedAmount = *summary.Underfunded
            break
        }
    }

    if underfundedAmount == 0 {
        return nil, fmt.Errorf("payment category not found or not underfunded")
    }

    // 3. Check if sufficient Ready to Assign
    readyToAssign, err := s.CalculateReadyToAssignForPeriod(ctx, period)
    if err != nil {
        return nil, fmt.Errorf("failed to calculate ready to assign: %w", err)
    }

    if readyToAssign < underfundedAmount {
        return nil, fmt.Errorf("insufficient funds: have %d, need %d", readyToAssign, underfundedAmount)
    }

    // 4. Create or update allocation
    allocation, err := s.CreateAllocation(ctx, paymentCategoryID, underfundedAmount, period, "Allocated to cover underfunded credit card")
    if err != nil {
        return nil, fmt.Errorf("failed to create allocation: %w", err)
    }

    return allocation, nil
}
```

#### 3. Add API Endpoint

**File:** `internal/infrastructure/http/handlers/allocation_handler.go`

**New Handler:**
```go
type CoverUnderfundedRequest struct {
    PaymentCategoryID string `json:"payment_category_id"`
    Period            string `json:"period"` // YYYY-MM
}

func (h *AllocationHandler) CoverUnderfunded(w http.ResponseWriter, r *http.Request) {
    var req CoverUnderfundedRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        http.Error(w, "invalid request body", http.StatusBadRequest)
        return
    }

    if req.PaymentCategoryID == "" {
        http.Error(w, "payment_category_id is required", http.StatusBadRequest)
        return
    }

    if req.Period == "" {
        http.Error(w, "period is required", http.StatusBadRequest)
        return
    }

    allocation, err := h.allocationService.AllocateToCoverUnderfunded(
        r.Context(),
        req.PaymentCategoryID,
        req.Period,
    )
    if err != nil {
        // Check for specific error types
        if strings.Contains(err.Error(), "insufficient funds") {
            w.WriteHeader(http.StatusBadRequest)
            json.NewEncoder(w).Encode(map[string]string{
                "error": err.Error(),
                "code":  "INSUFFICIENT_FUNDS",
            })
            return
        }

        http.Error(w, err.Error(), http.StatusBadRequest)
        return
    }

    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(http.StatusCreated)
    json.NewEncoder(w).Encode(map[string]interface{}{
        "allocation": allocation,
        "message":    "Successfully allocated to cover underfunded credit card",
    })
}
```

**File:** `internal/infrastructure/http/router.go`

**Add Route:**
```go
// In setupRoutes() function
router.HandleFunc("POST /api/allocations/cover-underfunded", allocationHandler.CoverUnderfunded)
```

---

### Frontend Changes

#### UI Component Specification

**Location:** Budget page, within each payment category row

**Visual Design:**

```
┌─────────────────────────────────────────────────────────────┐
│ Credit Card Payments                                          │
├─────────────────────────────────────────────────────────────┤
│                                                               │
│ Visa Payment                                                  │
│ Allocated: $100.00  |  Activity: -$50.00  |  Available: $50.00│
│                                                               │
│ ⚠️  Underfunded by $50.00                                     │
│ Underfunded categories: Groceries, Gas                        │
│ [ Allocate to Cover ]                                         │
│                                                               │
└─────────────────────────────────────────────────────────────┘
```

**Component Structure:**

```javascript
// Pseudocode
function PaymentCategoryRow({ category, summary, period }) {
    const [loading, setLoading] = useState(false);
    const [error, setError] = useState(null);

    const isUnderfunded = summary.underfunded && summary.underfunded > 0;
    const underfundedAmount = formatCurrency(summary.underfunded);
    const underfundedCategories = summary.underfunded_categories.join(', ');

    const handleAllocateToCover = async () => {
        setLoading(true);
        setError(null);

        try {
            const response = await fetch('/api/allocations/cover-underfunded', {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify({
                    payment_category_id: category.id,
                    period: period
                })
            });

            const data = await response.json();

            if (!response.ok) {
                if (data.code === 'INSUFFICIENT_FUNDS') {
                    setError(data.error);
                } else {
                    setError('Failed to allocate. Please try again.');
                }
                return;
            }

            // Show success message
            showToast('Successfully allocated to cover credit card');

            // Refresh budget data
            refreshBudgetSummary();
        } catch (err) {
            setError('Network error. Please try again.');
        } finally {
            setLoading(false);
        }
    };

    return (
        <div className="category-row">
            {/* Regular category display */}
            <CategoryHeader category={category} />
            <CategoryAmounts summary={summary} />

            {/* Underfunded warning (only show if underfunded) */}
            {isUnderfunded && (
                <div className="underfunded-warning">
                    <div className="warning-icon">⚠️</div>
                    <div className="warning-content">
                        <p className="warning-message">
                            Underfunded by {underfundedAmount}
                        </p>
                        <p className="warning-details">
                            Underfunded categories: {underfundedCategories}
                        </p>
                        <button
                            onClick={handleAllocateToCover}
                            disabled={loading}
                            className="btn-allocate-cover"
                        >
                            {loading ? 'Allocating...' : 'Allocate to Cover'}
                        </button>
                        {error && (
                            <p className="error-message">{error}</p>
                        )}
                    </div>
                </div>
            )}
        </div>
    );
}
```

**CSS Styling:**

```css
.underfunded-warning {
    display: flex;
    gap: 12px;
    padding: 12px;
    margin-top: 8px;
    background-color: #FEF3C7; /* Light yellow */
    border: 1px solid #F59E0B; /* Orange */
    border-radius: 6px;
}

.warning-icon {
    font-size: 20px;
}

.warning-content {
    flex: 1;
}

.warning-message {
    font-weight: 600;
    color: #92400E; /* Dark orange */
    margin: 0 0 4px 0;
}

.warning-details {
    font-size: 14px;
    color: #78350F;
    margin: 0 0 8px 0;
}

.btn-allocate-cover {
    padding: 8px 16px;
    background-color: #F59E0B;
    color: white;
    border: none;
    border-radius: 4px;
    cursor: pointer;
    font-weight: 500;
}

.btn-allocate-cover:hover {
    background-color: #D97706;
}

.btn-allocate-cover:disabled {
    background-color: #D1D5DB;
    cursor: not-allowed;
}

.error-message {
    margin-top: 8px;
    color: #DC2626;
    font-size: 14px;
}
```

---

## Test Specifications

### Unit Tests

**File:** `internal/application/allocation_service_test.go`

```go
func TestAllocateToCoverUnderfunded_Success(t *testing.T) {
    // Setup: Create payment category, CC account, underfunded spending
    // Action: Call AllocateToCoverUnderfunded
    // Assert: Allocation created with correct amount
}

func TestAllocateToCoverUnderfunded_NotPaymentCategory(t *testing.T) {
    // Setup: Regular expense category
    // Action: Call AllocateToCoverUnderfunded
    // Assert: Error returned
}

func TestAllocateToCoverUnderfunded_NotUnderfunded(t *testing.T) {
    // Setup: Payment category with sufficient allocation
    // Action: Call AllocateToCoverUnderfunded
    // Assert: Error returned
}

func TestAllocateToCoverUnderfunded_InsufficientReadyToAssign(t *testing.T) {
    // Setup: Underfunded by $100, only $50 ready to assign
    // Action: Call AllocateToCoverUnderfunded
    // Assert: Error with "insufficient funds" message
}
```

### Integration Tests

**File:** `internal/infrastructure/http/handlers/allocation_handler_test.go`

```go
func TestCoverUnderfundedEndpoint_Success(t *testing.T) {
    // Setup: Create underfunded CC via test database
    // Action: POST /api/allocations/cover-underfunded
    // Assert: 201 status, allocation created, RTA decreased
}

func TestCoverUnderfundedEndpoint_InsufficientFunds(t *testing.T) {
    // Setup: Underfunded CC, insufficient RTA
    // Action: POST /api/allocations/cover-underfunded
    // Assert: 400 status, error code INSUFFICIENT_FUNDS
}

func TestCoverUnderfundedEndpoint_InvalidCategory(t *testing.T) {
    // Setup: Non-existent category ID
    // Action: POST /api/allocations/cover-underfunded
    // Assert: 400 status, error message
}
```

### Manual QA Test Cases

#### Test Case 1: Normal Flow
1. Create credit card account "Visa"
2. Add income transaction $1000
3. Spend $50 on Groceries with Visa BEFORE allocating
4. Navigate to budget page
5. **Verify:** Visa Payment shows underfunded warning
6. **Verify:** Warning shows "$50.00" and "Groceries"
7. Click "Allocate to Cover"
8. **Verify:** Success message appears
9. **Verify:** Warning disappears
10. **Verify:** Visa Payment allocation shows $50
11. **Verify:** Ready to Assign decreased by $50

#### Test Case 2: Insufficient Funds
1. Create credit card account "Visa"
2. Add income transaction $30
3. Spend $50 on Groceries with Visa
4. Navigate to budget page
5. Click "Allocate to Cover"
6. **Verify:** Error message: "Insufficient funds: have $30.00, need $50.00..."
7. **Verify:** No allocation created
8. **Verify:** Warning still visible

#### Test Case 3: Multiple Underfunded Categories
1. Spend $30 on Groceries with Visa
2. Spend $20 on Gas with Visa
3. Navigate to budget page
4. **Verify:** Warning shows "Underfunded by $50.00"
5. **Verify:** Details show "Groceries, Gas"
6. Click "Allocate to Cover"
7. **Verify:** Single allocation of $50 covers both

---

## Performance Benchmarks

### Before (With Sync)

**Benchmark:** Create 100 allocations sequentially

```go
// Expected results (to be measured)
func BenchmarkCreateAllocation_WithSync(b *testing.B) {
    // Setup: 5 credit cards, 1000 transactions, 500 existing allocations
    // Measure: Time to create 100 allocations
    // Expected: ~5-10 seconds (due to syncing overhead)
}
```

### After (Without Sync)

```go
func BenchmarkCreateAllocation_WithoutSync(b *testing.B) {
    // Setup: Same as above
    // Measure: Time to create 100 allocations
    // Expected: <1 second (no syncing)
}
```

**Target:** 5-10x performance improvement

---

## Implementation Steps

### Phase 1: Remove Syncing (1-2 hours)
1. [ ] Create feature branch: `remove-cc-sync`
2. [ ] Remove `syncPaymentCategoryAllocations` function
3. [ ] Remove calls to the function
4. [ ] Run existing tests to verify nothing breaks
5. [ ] Commit: "Remove retroactive CC payment syncing"

### Phase 2: Add Backend Helper (1-2 hours)
6. [ ] Implement `AllocateToCoverUnderfunded` method
7. [ ] Write unit tests for the method
8. [ ] Add API handler `CoverUnderfunded`
9. [ ] Add route in router
10. [ ] Write integration tests for endpoint
11. [ ] Commit: "Add helper endpoint to cover underfunded CC"

### Phase 3: Add Frontend UI (2-3 hours)
12. [ ] Add underfunded warning component
13. [ ] Wire up "Allocate to Cover" button
14. [ ] Handle error states (insufficient funds)
15. [ ] Add CSS styling
16. [ ] Manual testing in browser
17. [ ] Commit: "Add UI helper for underfunded credit cards"

### Phase 4: Testing & Documentation (1 hour)
18. [ ] Run full test suite
19. [ ] Performance benchmark comparison
20. [ ] Update user documentation
21. [ ] Create PR with detailed description
22. [ ] Code review

---

## Acceptance Criteria

### Code Quality
- [ ] All existing tests pass
- [ ] New unit tests added with >80% coverage
- [ ] Integration tests added for new endpoint
- [ ] No linting errors
- [ ] Code reviewed and approved

### Functionality
- [ ] Sync function removed
- [ ] Real-time allocation still works correctly
- [ ] New API endpoint works as specified
- [ ] UI displays underfunded warnings
- [ ] Button successfully allocates money
- [ ] Error handling works (insufficient funds, etc.)

### Performance
- [ ] Allocation creation is measurably faster
- [ ] No performance regressions in other operations

### Documentation
- [ ] Code comments added for new functions
- [ ] User documentation updated
- [ ] API documentation updated

---

## Rollback Plan

If issues are discovered after deployment:

1. **Immediate:** Revert the PR
2. **Temporary Fix:** Re-enable sync function if critical
3. **Investigation:** Identify root cause
4. **Resolution:** Fix and redeploy

**Risk Level:** Low
- Real-time logic already exists and is tested
- Removing code is lower risk than adding
- Frontend feature is additive, doesn't break existing flow

---

## Open Questions

1. **Q:** Should we add analytics to track how often users use "Allocate to Cover"?
   - **A:** TBD - Discuss with product team

2. **Q:** Should we show underfunded warning in multiple places (dashboard, etc.)?
   - **A:** TBD - Start with budget page, expand later if needed

3. **Q:** Should we support partial allocation (allocate $30 when underfunded by $50)?
   - **A:** TBD - Current spec allocates full amount or nothing

---

## Success Metrics

After 2 weeks in production:

- [ ] No increase in support tickets about CC payments
- [ ] "Allocate to Cover" button used by >10% of users with underfunded CCs
- [ ] Average allocation creation time decreased by >5x
- [ ] No bugs reported related to CC payment tracking

---

**Document Version:** 1.0
**Last Updated:** 2025-10-31
**Author:** Claude
**Approved By:** [Pending]
