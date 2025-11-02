# Specification: Remove Credit Card Payment Syncing and Add Manual Allocation Helper

**Status:** Draft
**Created:** 2025-10-31
**Author:** AI-assisted specification
**Validated:** Yes (Domain Expert + Security Auditor)

---

## Executive Summary

This specification outlines the removal of the automatic `syncPaymentCategoryAllocations` function that retroactively adjusts credit card payment allocations, replacing it with a user-initiated manual helper. The current implementation violates zero-based budgeting principles by making hidden, retroactive adjustments with O(n²) performance complexity. The proposed solution preserves the correct real-time allocation logic (when transactions occur) while giving users explicit control over covering underfunded payment categories.

**Key Changes:**
1. **Remove:** `syncPaymentCategoryAllocations` function and all call sites
2. **Preserve:** Real-time allocation logic in transaction creation
3. **Add:** Manual helper API endpoint `POST /api/allocations/cover-underfunded`
4. **Add:** UI "Allocate to Cover" button for underfunded payment categories

---

## Business Requirements

### User Stories

**US-1: Transparent Budget Allocation**
As a budget user, I want to explicitly allocate money to cover credit card spending, so that I understand exactly where my money is going and maintain control over my budget.

**US-2: Real-Time Budget Movement**
As a budget user, when I spend money on a credit card, I want the budgeted amount to automatically move from the expense category to the payment category, so that my budget reflects my obligation to pay the credit card.

**US-3: Underfunded Awareness**
As a budget user, when my credit card payment category doesn't have enough money to cover the debt, I want to see a clear warning showing the shortfall amount and which categories contributed to it.

**US-4: Quick Coverage**
As a budget user, I want a single-click way to allocate money from "Ready to Assign" to cover underfunded payment categories, so that I can quickly resolve credit card budget shortfalls.

**US-5: Insufficient Funds Feedback**
As a budget user, when I don't have enough money in "Ready to Assign" to cover an underfunded payment category, I want to see a clear error message showing the shortfall, so I understand my financial situation.

### Success Criteria

- [ ] Automatic retroactive syncing is completely removed
- [ ] Real-time allocation continues to work when credit card transactions occur
- [ ] Underfunded payment categories display clear warnings with amounts and contributing categories
- [ ] "Allocate to Cover" button successfully creates allocations for underfunded amounts
- [ ] Error handling provides clear feedback for insufficient funds
- [ ] Performance improves (no more O(n²) database scans)
- [ ] Zero-based budgeting principles are maintained
- [ ] All tests pass (unit, integration, API, domain)

---

## Domain Validation

**Budget Domain Expert Review:**

### Status: ✅ APPROVED

**Findings:**

The proposed changes strongly align with zero-based budgeting principles and resolve fundamental domain issues:

1. **Retroactive Syncing Violates Zero-Based Budgeting**
   - Current `syncPaymentCategoryAllocations` function retroactively allocates in wrong periods
   - Uses `time.Now()` for historical spending, violating period-based budgeting
   - Hidden from users, breaking intentionality principle
   - O(n²) complexity causes performance degradation
   - **Verdict:** Remove entirely

2. **Real-Time Allocation Logic is Correct** ✅
   - Transaction service (lines 91-190) properly moves budget at transaction time
   - Only moves budgeted money (respects allocations)
   - Allocates in correct period (transaction period, not current)
   - Visible to users in budget view
   - **Verdict:** Preserve unchanged

3. **Underfunded Detection Already Implemented** ✅
   - Formula: `Underfunded = CC Balance (debt) - Payment Category Available`
   - Correctly identifies shortfall between debt and funds available
   - **Verdict:** Continue using

4. **Ready to Assign Already Accounts for Underfunded** ✅
   - Formula: `RTA = Total Inflows - Total Allocations - Total Underfunded`
   - Brilliant design: underfunded amounts already reduce RTA
   - Manual allocation makes implicit allocation explicit (RTA stays constant!)
   - **Verdict:** No changes needed

**Zero-Based Budgeting Impact:**

This change **improves** zero-based budgeting compliance:
- ✅ Every dollar has a job (underfunded reduces RTA)
- ✅ Forward-looking budgeting (removes retroactive adjustments)
- ✅ Intentionality (user explicitly chooses to allocate)
- ✅ Rollover by default (Available includes all history)

**Formulas Verified:**

All core budget formulas are correct and will be preserved:
- Category Available = Sum(All Allocations) - Sum(All Spending) ✅
- Underfunded = CC Balance - Payment Category Available ✅
- Ready to Assign = Inflows - Allocations - Underfunded ✅
- Real-Time Movement = min(expense available, spending amount) ✅

**Domain Constraints:**

1. Manual helper must verify category is a payment category (`PaymentForAccountID != nil`)
2. Must check Ready to Assign >= underfunded amount before allocating
3. Use upsert behavior (update existing or create new allocation)
4. Period must be "YYYY-MM" format
5. Preserve real-time allocation logic in transaction service unchanged
6. Keep underfunded calculation in GetAllocationSummary unchanged
7. Keep Ready to Assign formula with underfunded subtraction unchanged

---

## Security Review

**Security Auditor Review:**

### Status: ⚠️ APPROVED WITH REQUIRED CHANGES

**Security Considerations:**

The feature is fundamentally sound but requires critical security enhancements:

1. **Input Validation (HIGH PRIORITY)**
   - Missing UUID format validation for `payment_category_id`
   - Missing period format validation (YYYY-MM)
   - No verification that category is actually a payment category
   - Must validate before processing

2. **Error Handling (HIGH PRIORITY)**
   - Current pattern exposes internal error details
   - Must use generic user-facing messages with internal logging
   - Security risk: database structure and implementation details leaked

3. **Race Condition Risk (MEDIUM-HIGH PRIORITY)**
   - Concurrent allocations could cause negative Ready to Assign
   - Must use database transactions with Serializable isolation
   - Financial data integrity at risk

4. **Amount Validation (MEDIUM PRIORITY)**
   - No validation that underfunded amount is positive
   - No bounds checking for reasonable amounts
   - Must validate before creating allocation

**Security Requirements:**

### Input Validation Rules
- [ ] payment_category_id: UUID format validation with `uuid.Parse()`
- [ ] period: Regex validation `^\d{4}-(0[1-9]|1[0-2])$`
- [ ] period: Range validation (2 years past, 5 years future)
- [ ] category: Must be payment category (has `payment_for_account_id`)
- [ ] amount: Must be positive and > 0
- [ ] amount: Must not exceed Ready to Assign

### Authentication/Authorization
- Single-user application: no multi-user auth required currently
- Future-proof: validate data ownership patterns in place

### Data Protection
- Use parameterized SQL queries (already implemented ✅)
- Generic error messages to client (400/404/500)
- Detailed error logging internally only
- Don't log sensitive user data

### SQL Injection Prevention
- Current codebase uses parameterized queries correctly ✅
- Continue using `db.ExecContext(ctx, query, params...)`
- Never concatenate user input into SQL strings

### Transaction Isolation
```go
tx, err := s.db.BeginTx(ctx, &sql.TxOptions{
    Isolation: sql.LevelSerializable, // Prevents race conditions
})
```

**Potential Vulnerabilities Identified:**

1. Invalid input could cause unexpected errors (400/500)
2. Race conditions could create negative RTA state
3. Error messages could leak internal details
4. Missing payment category type validation could corrupt data

**Positive Security Practices Found:**

- ✅ SQL injection prevention (parameterized queries)
- ✅ No hardcoded credentials (uses environment variables)
- ✅ Foreign key constraints (referential integrity)
- ✅ Clean architecture separation (handlers/services/repositories)

---

## Technical Design

### Architecture Compliance

**Domain Layer:**
- No changes required (entities and interfaces remain unchanged)
- `domain.Allocation` struct already supports all needed fields
- `domain.Category` already has `PaymentForAccountID` field

**Application Layer:**
- **Remove:** `AllocationService.syncPaymentCategoryAllocations()` (lines 100-190)
- **Remove:** Call sites in `CreateOrUpdateAllocation` (lines 67, 91)
- **Add:** `AllocationService.AllocateToCoverUnderfunded()` method
- Business logic: validation, underfunded calculation, RTA check, allocation creation

**Infrastructure Layer:**
- **Add:** HTTP handler `POST /api/allocations/cover-underfunded`
- **Add:** Input validation utilities
- **Add:** Transaction support in repositories
- **Update:** Router to include new endpoint

**Dependencies:** ✅ Point inward (infrastructure → application → domain)

### Database Schema Changes

**No database migrations required.**

All necessary tables and columns already exist:
- `allocations` table has all required fields
- `categories` table has `payment_for_account_id` field
- Foreign key constraints already in place

**Existing Schema (No Changes):**
```sql
-- Allocations table (already exists)
CREATE TABLE allocations (
    id TEXT PRIMARY KEY,
    category_id TEXT NOT NULL,
    amount INTEGER NOT NULL,
    period TEXT NOT NULL,
    notes TEXT,
    created_at DATETIME NOT NULL,
    updated_at DATETIME NOT NULL,
    FOREIGN KEY (category_id) REFERENCES categories(id),
    UNIQUE(category_id, period)
);

-- Categories table (already exists)
CREATE TABLE categories (
    id TEXT PRIMARY KEY,
    name TEXT NOT NULL,
    payment_for_account_id TEXT,
    FOREIGN KEY (payment_for_account_id) REFERENCES accounts(id)
);
```

**Migration Notes:**
- ✅ Backward compatible (no schema changes)
- ✅ No data migration required
- ✅ Existing allocations remain unchanged

### API Design

#### Endpoint: POST /api/allocations/cover-underfunded

**Purpose:** Manually allocate funds from Ready to Assign to cover an underfunded payment category

**Request:**
- Method: POST
- Path: `/api/allocations/cover-underfunded`
- Headers: `Content-Type: application/json`
- Request Body:
```json
{
  "payment_category_id": "550e8400-e29b-41d4-a716-446655440000",
  "period": "2025-10"
}
```

**Response:**

Success (201 Created):
```json
{
  "allocation": {
    "id": "660e8400-e29b-41d4-a716-446655440000",
    "category_id": "550e8400-e29b-41d4-a716-446655440000",
    "amount": 20000,
    "period": "2025-10",
    "notes": "Cover underfunded credit card spending",
    "created_at": "2025-10-31T10:30:00Z",
    "updated_at": "2025-10-31T10:30:00Z"
  },
  "underfunded_amount": 20000,
  "ready_to_assign_after": 330000
}
```

Validation Error (400 Bad Request):
```json
{
  "error": "Invalid payment_category_id format"
}
```

Not Found (404 Not Found):
```json
{
  "error": "Payment category not found"
}
```

Insufficient Funds (400 Bad Request):
```json
{
  "error": "Insufficient funds: Ready to Assign: $33.00, Underfunded: $200.00"
}
```

Internal Error (500 Internal Server Error):
```json
{
  "error": "Failed to process allocation request"
}
```

**Validation:**
- `payment_category_id`: Required, UUID format, must exist, must be payment category
- `period`: Required, YYYY-MM format, must be within reasonable range (2 years past, 5 years future)
- Underfunded amount: Must be > 0 (calculated)
- Ready to Assign: Must be >= underfunded amount

**Side Effects:**
- Creates or updates allocation for payment category in specified period
- Updates Ready to Assign (decreases by allocation amount)
- Reduces or eliminates underfunded amount for payment category
- Updates allocation `updated_at` timestamp

**Security:**
- Input validation required before processing
- Database transaction with Serializable isolation
- Generic error messages (don't expose internal details)

### Service Layer Design

**Modified Services:**

**AllocationService** (`internal/application/allocation_service.go`)

**Remove Methods:**
```go
// DELETE this entire function (lines 100-190)
func (s *AllocationService) syncPaymentCategoryAllocations(ctx context.Context, categoryID string) error
```

**Remove Call Sites:**
```go
// Line 67: Remove this block
if category.PaymentForAccountID != nil {
    if err := s.syncPaymentCategoryAllocations(ctx, categoryID); err != nil {
        fmt.Printf("Warning: failed to sync payment category allocations: %v\n", err)
    }
}

// Line 91: Remove this block
if err := s.syncPaymentCategoryAllocations(ctx, categoryID); err != nil {
    fmt.Printf("Warning: failed to sync payment category allocations: %v\n", err)
}
```

**Add Methods:**
```go
// AllocateToCoverUnderfunded creates an allocation to cover underfunded payment category
func (s *AllocationService) AllocateToCoverUnderfunded(
    ctx context.Context,
    paymentCategoryID string,
    period string,
) (*domain.Allocation, int64, error) {

    // Returns: allocation created, underfunded amount covered, error

    // Implementation steps:
    // 1. Validate payment category exists and is payment category
    // 2. Calculate underfunded amount from GetAllocationSummary
    // 3. Check underfunded amount > 0
    // 4. Calculate Ready to Assign for period
    // 5. Verify RTA >= underfunded amount
    // 6. Create/update allocation with upsert behavior
    // 7. Return allocation, underfunded amount, nil
}
```

**Business Rules:**

1. **Payment Category Validation**
   - Must exist in database
   - Must have `payment_for_account_id` set (is payment category)
   - Return error if not payment category: "Category is not a payment category"

2. **Underfunded Calculation**
   - Use existing `GetAllocationSummary` to get underfunded amount
   - Formula: `CC Balance (debt) - Payment Category Available`
   - Must be > 0 to proceed
   - Return error if not underfunded: "Payment category is not underfunded"

3. **Ready to Assign Check**
   - Calculate using `CalculateReadyToAssignForPeriod`
   - Must be >= underfunded amount
   - Return error with amounts if insufficient: "Insufficient funds: Ready to Assign: $X, Underfunded: $Y"

4. **Allocation Creation (Upsert)**
   - Use `CreateOrUpdateAllocation` (existing method)
   - If allocation exists for (category_id, period): update amount
   - If not exists: create new allocation
   - Notes: "Cover underfunded credit card spending"

5. **Transaction Boundaries**
   - Entire operation within database transaction
   - Serializable isolation level to prevent race conditions
   - Atomic: all succeed or all rollback

### Repository Changes

**No new repositories required.**

**Modified Repositories:**

**AllocationRepository** (`internal/infrastructure/repository/allocation_repository.go`)

**Add Transaction Support:**
```go
// Add method to begin transaction
func (r *AllocationRepository) BeginTx(ctx context.Context, opts *sql.TxOptions) (*sql.Tx, error)

// Add method to execute operations within transaction
func (r *AllocationRepository) WithTx(tx *sql.Tx) *AllocationRepository
```

**Existing Methods Used:**
- `GetByID(ctx, id)` - Fetch category
- `Create(ctx, allocation)` - Create allocation
- `Update(ctx, allocation)` - Update allocation
- `GetAllocationSummary(ctx, period)` - Get underfunded amounts
- `CalculateReadyToAssignForPeriod(ctx, period)` - Check funds

---

## Test Plan

### Unit Tests

**Service Tests:**

**Test: AllocateToCoverUnderfunded - Success**
```go
func TestAllocationService_AllocateToCoverUnderfunded_Success(t *testing.T)
```
- Given: Payment category with $200 underfunded, RTA = $500
- When: Call AllocateToCoverUnderfunded(paymentCategoryID, "2025-10")
- Then:
  - Allocation created with amount = $200
  - Underfunded = $0
  - RTA = $300 (decreased by $200)
  - No error

**Test: AllocateToCoverUnderfunded - Invalid UUID**
```go
func TestAllocationService_AllocateToCoverUnderfunded_InvalidUUID(t *testing.T)
```
- Given: Invalid UUID format "not-a-uuid"
- When: Call AllocateToCoverUnderfunded("not-a-uuid", "2025-10")
- Then: Error = "Invalid payment_category_id format"

**Test: AllocateToCoverUnderfunded - Not Payment Category**
```go
func TestAllocationService_AllocateToCoverUnderfunded_NotPaymentCategory(t *testing.T)
```
- Given: Regular expense category (no payment_for_account_id)
- When: Call AllocateToCoverUnderfunded(expenseCategoryID, "2025-10")
- Then: Error = "Category is not a payment category"

**Test: AllocateToCoverUnderfunded - Not Underfunded**
```go
func TestAllocationService_AllocateToCoverUnderfunded_NotUnderfunded(t *testing.T)
```
- Given: Payment category with $0 underfunded
- When: Call AllocateToCoverUnderfunded(paymentCategoryID, "2025-10")
- Then: Error = "Payment category is not underfunded"

**Test: AllocateToCoverUnderfunded - Insufficient Funds**
```go
func TestAllocationService_AllocateToCoverUnderfunded_InsufficientFunds(t *testing.T)
```
- Given: Underfunded = $500, RTA = $100
- When: Call AllocateToCoverUnderfunded(paymentCategoryID, "2025-10")
- Then: Error = "Insufficient funds: Ready to Assign: $1.00, Underfunded: $5.00"

**Test: AllocateToCoverUnderfunded - Upsert Existing**
```go
func TestAllocationService_AllocateToCoverUnderfunded_UpsertExisting(t *testing.T)
```
- Given: Existing allocation = $100, underfunded = $50
- When: Call AllocateToCoverUnderfunded(paymentCategoryID, "2025-10")
- Then:
  - Allocation amount = $150 (updated, not duplicated)
  - Only one allocation record exists

**Test: AllocateToCoverUnderfunded - Invalid Period Format**
```go
func TestAllocationService_AllocateToCoverUnderfunded_InvalidPeriod(t *testing.T)
```
- Given: Invalid period "2025-13" (month 13)
- When: Call AllocateToCoverUnderfunded(paymentCategoryID, "2025-13")
- Then: Error = "Invalid period format, expected YYYY-MM"

**Test: SyncPaymentCategoryAllocations - Removed**
```go
func TestAllocationService_SyncNotCalled(t *testing.T)
```
- Given: Create any allocation
- When: CreateOrUpdateAllocation called
- Then: Verify syncPaymentCategoryAllocations is NOT called (removed)

### Integration Tests

**Repository Tests:**

**Test: Repository - Create Allocation with Transaction**
```go
func TestAllocationRepository_CreateWithTransaction(t *testing.T)
```
- Given: Database transaction started
- When: Create allocation within transaction, commit
- Then: Allocation persisted in database

**Test: Repository - Transaction Rollback**
```go
func TestAllocationRepository_TransactionRollback(t *testing.T)
```
- Given: Database transaction started, allocation created
- When: Error occurs, transaction rolled back
- Then: Allocation NOT persisted in database

**Test: Repository - Race Condition Prevention**
```go
func TestAllocationRepository_RaceConditionPrevention(t *testing.T)
```
- Given: Two concurrent allocation requests
- When: Both try to allocate from same RTA
- Then: Only sufficient allocations succeed, no negative RTA

### API Tests

**Handler Tests:**

**Test: POST /api/allocations/cover-underfunded - Success**
```go
func TestHandler_CoverUnderfunded_Success(t *testing.T)
```
- Request: POST with valid payment_category_id and period
- Response: 201 Created with allocation details
- Database: Allocation created/updated

**Test: POST /api/allocations/cover-underfunded - Invalid JSON**
```go
func TestHandler_CoverUnderfunded_InvalidJSON(t *testing.T)
```
- Request: POST with malformed JSON
- Response: 400 Bad Request

**Test: POST /api/allocations/cover-underfunded - Invalid UUID**
```go
func TestHandler_CoverUnderfunded_InvalidUUID(t *testing.T)
```
- Request: POST with payment_category_id = "not-a-uuid"
- Response: 400 Bad Request, error message about format

**Test: POST /api/allocations/cover-underfunded - Category Not Found**
```go
func TestHandler_CoverUnderfunded_CategoryNotFound(t *testing.T)
```
- Request: POST with non-existent payment_category_id
- Response: 404 Not Found

**Test: POST /api/allocations/cover-underfunded - Not Payment Category**
```go
func TestHandler_CoverUnderfunded_NotPaymentCategory(t *testing.T)
```
- Request: POST with expense category ID
- Response: 400 Bad Request

**Test: POST /api/allocations/cover-underfunded - Insufficient Funds**
```go
func TestHandler_CoverUnderfunded_InsufficientFunds(t *testing.T)
```
- Request: POST when RTA < underfunded
- Response: 400 Bad Request with specific error message

### Budget Domain Tests

**Test: Real-Time Allocation Still Works**
```go
func TestBudgetDomain_RealTimeAllocationPreserved(t *testing.T)
```
- Given: Allocate $500 to Groceries
- When: Create CC transaction for $100 in Groceries
- Then:
  - Groceries available = $400
  - Payment category allocation = $100
  - RTA unchanged
  - Underfunded = $0

**Test: Underfunded Calculation Correct**
```go
func TestBudgetDomain_UnderfundedCalculation(t *testing.T)
```
- Given: CC balance = -$500, payment category available = $300
- When: Calculate underfunded
- Then: Underfunded = $200 ($500 - $300)

**Test: Ready to Assign Accounts for Underfunded**
```go
func TestBudgetDomain_RTAAccountsForUnderfunded(t *testing.T)
```
- Given: Inflows = $5000, allocated = $1500, underfunded = $200
- When: Calculate RTA
- Then: RTA = $3300 ($5000 - $1500 - $200)

**Test: Manual Allocation Keeps RTA Constant**
```go
func TestBudgetDomain_ManualAllocationKeepsRTAConstant(t *testing.T)
```
- Given: RTA = $3300 before allocation
- When: Allocate $200 to cover underfunded
- Then:
  - RTA = $3300 after allocation (unchanged!)
  - Allocated increased by $200
  - Underfunded decreased by $200

**Test: Zero-Based Budgeting Formula Integrity**
```go
func TestBudgetDomain_FormulaIntegrity(t *testing.T)
```
- Verify: Category Available = Sum(Allocations) - Sum(Spending)
- Verify: Underfunded = CC Balance - Payment Category Available
- Verify: RTA = Inflows - Allocations - Underfunded
- All formulas must remain correct after changes

### Performance Tests

**Test: No O(n²) Operations**
```go
func TestPerformance_NoQuadraticComplexity(t *testing.T)
```
- Given: 1000 allocations, 1000 transactions, 10 credit cards
- When: Create new allocation
- Then:
  - Operation completes in < 100ms
  - No full table scans
  - No nested loops over all transactions

**Benchmark: Allocation Creation Before/After**
```go
func BenchmarkAllocationCreation_BeforeRemoval(b *testing.B)
func BenchmarkAllocationCreation_AfterRemoval(b *testing.B)
```
- Measure allocation creation time before and after removing sync
- Expected: Significant improvement (10x or more)

---

## Implementation Checklist

### Phase 1: Remove Syncing Logic

#### Domain Layer
- [ ] No changes required (domain entities unchanged)

#### Application Layer
- [ ] Delete `AllocationService.syncPaymentCategoryAllocations()` method (lines 100-190)
- [ ] Remove call site in `CreateOrUpdateAllocation` (line 67)
- [ ] Remove call site in `CreateOrUpdateAllocation` (line 91)
- [ ] Remove any imports used only by sync function

#### Infrastructure Layer
- [ ] No changes required for removal

#### Testing
- [ ] Run existing tests to verify no breakage
- [ ] Remove unit tests for syncPaymentCategoryAllocations
- [ ] Add test to verify sync is not called

### Phase 2: Add Validation Utilities

#### Infrastructure Layer
- [ ] Create `internal/infrastructure/http/validators/validators.go`
- [ ] Implement UUID validation function
- [ ] Implement period format validation (YYYY-MM regex)
- [ ] Implement period range validation (2 years past, 5 years future)
- [ ] Implement amount validation (positive, bounds checking)

#### Testing
- [ ] Unit tests for UUID validation
- [ ] Unit tests for period validation
- [ ] Unit tests for amount validation

### Phase 3: Add Manual Helper Service Method

#### Application Layer
- [ ] Add `AllocateToCoverUnderfunded()` method to AllocationService
- [ ] Implement payment category validation
- [ ] Implement underfunded calculation (use existing GetAllocationSummary)
- [ ] Implement RTA sufficiency check
- [ ] Implement allocation creation with upsert behavior
- [ ] Add transaction support with Serializable isolation
- [ ] Implement error handling with clear messages

#### Testing
- [ ] Unit test: Success case
- [ ] Unit test: Invalid UUID
- [ ] Unit test: Not payment category
- [ ] Unit test: Not underfunded
- [ ] Unit test: Insufficient funds
- [ ] Unit test: Upsert existing allocation
- [ ] Unit test: Invalid period format
- [ ] Integration test: Transaction commit
- [ ] Integration test: Transaction rollback
- [ ] Integration test: Race condition prevention

### Phase 4: Add API Endpoint

#### Infrastructure Layer
- [ ] Create handler `CoverUnderfundedHandler` in `allocation_handler.go`
- [ ] Implement request parsing and validation
- [ ] Call service method `AllocateToCoverUnderfunded`
- [ ] Implement response formatting (201/400/404/500)
- [ ] Implement secure error handling (generic messages)
- [ ] Add route `POST /api/allocations/cover-underfunded` to router

#### Configuration
- [ ] No changes to `cmd/server/main.go` (handler uses existing service)

#### Testing
- [ ] API test: Success (201)
- [ ] API test: Invalid JSON (400)
- [ ] API test: Invalid UUID (400)
- [ ] API test: Invalid period (400)
- [ ] API test: Category not found (404)
- [ ] API test: Not payment category (400)
- [ ] API test: Insufficient funds (400)
- [ ] API test: Internal error (500)

### Phase 5: Frontend UI (Separate Implementation)

#### Frontend Components
- [ ] Display underfunded warning in payment category row
- [ ] Show underfunded amount formatted as currency
- [ ] Show list of contributing expense categories
- [ ] Add "Allocate to Cover" button with loading state
- [ ] Implement button click handler (POST to API)
- [ ] Handle success response (refresh allocation summary)
- [ ] Handle error response (display error message)
- [ ] Disable button if RTA insufficient
- [ ] Show tooltip/message for insufficient funds

#### Frontend Testing
- [ ] UI test: Button appears for underfunded categories
- [ ] UI test: Button click creates allocation
- [ ] UI test: Success updates UI
- [ ] UI test: Error displays message
- [ ] UI test: Insufficient funds disables button

### Phase 6: Documentation

#### Code Documentation
- [ ] Add godoc comments to `AllocateToCoverUnderfunded` method
- [ ] Add godoc comments to validation functions
- [ ] Add godoc comments to handler
- [ ] Update inline comments for clarity

#### API Documentation
- [ ] Document POST /api/allocations/cover-underfunded in API docs
- [ ] Include request/response examples
- [ ] Document error codes and messages
- [ ] Update postman collection or equivalent

#### User Documentation
- [ ] Update README if needed
- [ ] Document new "Allocate to Cover" feature
- [ ] Explain removal of automatic syncing

### Phase 7: Verification

#### Code Quality
- [ ] Run `go fmt` on all changed files
- [ ] Run `go vet` and fix any issues
- [ ] Run `golangci-lint` if available

#### Architecture Compliance
- [ ] Verify dependencies point inward (infra → app → domain)
- [ ] Verify domain layer has no external dependencies
- [ ] Verify clean separation of concerns

#### Performance
- [ ] Run benchmarks before/after
- [ ] Verify no O(n²) operations
- [ ] Verify allocation creation is faster

#### Security
- [ ] Verify all input validation implemented
- [ ] Verify secure error handling (generic messages)
- [ ] Verify transaction isolation in place
- [ ] Verify no SQL injection vulnerabilities

#### Testing
- [ ] All unit tests pass
- [ ] All integration tests pass
- [ ] All API tests pass
- [ ] All budget domain tests pass
- [ ] Test coverage >= 80%

---

## Acceptance Criteria

### Functional

- [ ] ✅ Automatic retroactive syncing is completely removed
- [ ] ✅ Real-time allocation in transaction creation still works
- [ ] ✅ Underfunded detection works correctly
- [ ] ✅ Manual helper API endpoint creates allocations successfully
- [ ] ✅ UI button "Allocate to Cover" works with single click
- [ ] ✅ Error messages are clear and helpful
- [ ] ✅ Upsert behavior works (update existing or create new)
- [ ] ✅ Payment category validation prevents misuse
- [ ] ✅ Insufficient funds error provides clear guidance

### Non-Functional

- [ ] ✅ Performance improved (no O(n²) database scans)
- [ ] ✅ All tests passing (unit, integration, API, domain)
- [ ] ✅ Clean architecture maintained (dependencies point inward)
- [ ] ✅ Security requirements met (validation, error handling, transactions)
- [ ] ✅ Zero-based budgeting principles preserved
- [ ] ✅ Documentation complete and accurate
- [ ] ✅ No breaking API changes for existing endpoints
- [ ] ✅ No database migrations required

### Code Quality

- [ ] ✅ No architecture violations
- [ ] ✅ Go best practices followed
- [ ] ✅ Error handling comprehensive
- [ ] ✅ Input validation complete
- [ ] ✅ Test coverage >= 80%
- [ ] ✅ Code comments clear and helpful
- [ ] ✅ No lint warnings or errors

---

## Risks and Mitigations

**Risk:** Real-time allocation logic might have edge cases not covered by current tests
**Mitigation:** Comprehensive budget domain tests to verify formulas remain correct; manual testing with various scenarios before deployment

**Risk:** Race conditions in concurrent allocation requests
**Mitigation:** Database transactions with Serializable isolation level; integration tests specifically for race conditions

**Risk:** Users might not understand why automatic syncing was removed
**Mitigation:** Clear documentation explaining benefits (performance, transparency, intentionality); helpful error messages guide users

**Risk:** Performance improvement might not be as significant as expected
**Mitigation:** Benchmarks before/after to measure actual improvement; worst case is still better than O(n²) complexity

**Risk:** Frontend UI changes might have poor UX
**Mitigation:** Follow established patterns for buttons and error messages; user testing if available; progressive enhancement (API works even if UI not perfect)

**Risk:** Underfunded calculation might have bugs not caught by current tests
**Mitigation:** Budget domain expert validated the formula; comprehensive test scenarios including edge cases; existing code has been in production

---

## Dependencies

**No external dependencies.**

All required libraries and frameworks are already in use:
- Standard library (`database/sql`, `encoding/json`, etc.)
- UUID generation (already implemented)
- SQLite database (already configured)
- HTTP router (already configured)

**Internal Dependencies:**

This feature depends on existing functionality:
- `GetAllocationSummary` for underfunded calculation
- `CalculateReadyToAssignForPeriod` for funds checking
- `CreateOrUpdateAllocation` for allocation upsert
- Transaction service real-time allocation logic (must preserve)

---

## Future Enhancements

**Deferred for later implementation:**

1. **Partial Coverage UI**
   - Allow users to allocate less than full underfunded amount
   - Slider or input field for custom allocation amount
   - Show remaining underfunded after partial allocation

2. **Multiple Category Coverage**
   - Batch "Cover All Underfunded" button
   - Allocate to all underfunded payment categories at once
   - Show summary of total coverage needed

3. **Smart Suggestions**
   - Suggest which categories to reduce to free up RTA
   - Show available budget in other categories
   - Recommend reallocation strategies

4. **Payment Due Dates**
   - Track credit card payment due dates
   - Prioritize underfunded warnings by due date
   - Alert users before due dates approach

5. **Historical Trends**
   - Show average monthly spending per payment category
   - Suggest proactive allocations based on trends
   - Identify categories frequently underfunded

6. **Mobile-Friendly UI**
   - Optimize button layout for mobile screens
   - Swipe actions for quick allocation
   - Push notifications for underfunded categories

---

## Approval

- [x] Domain Expert: ✅ Validated (strongly aligned with zero-based budgeting)
- [x] Security Auditor: ⚠️ Reviewed (approved with required security enhancements)
- [ ] Stakeholder: Approved (pending review)
- [ ] Developer: Ready to implement

---

## Implementation

**Ready for implementation once security requirements are addressed:**

```
/implement-spec docs/spec-remove-cc-sync.md
```

**Estimated Implementation Time:**
- Phase 1 (Remove syncing): 1-2 hours
- Phase 2 (Validation utilities): 1-2 hours
- Phase 3 (Service method): 2-3 hours
- Phase 4 (API endpoint): 1-2 hours
- Phase 5 (Frontend UI): 2-3 hours
- Phase 6 (Documentation): 1 hour
- Phase 7 (Verification): 1 hour

**Total:** ~10-15 hours

**Complexity:** Medium

**Files to Create/Modify:** ~8 files
- Create: `validators.go` (validation utilities)
- Modify: `allocation_service.go` (remove sync, add helper)
- Modify: `allocation_handler.go` (add endpoint)
- Modify: `router.go` (add route)
- Create/Modify: Test files (~4 files)

---

## Notes

### Key Decisions Made

1. **Preserve Real-Time Allocation:** Confirmed correct and aligned with zero-based budgeting
2. **Remove Retroactive Sync:** Violates zero-based principles, causes performance issues
3. **Manual Helper Approach:** Gives users control, maintains intentionality
4. **Upsert Behavior:** Consistent with existing allocation patterns
5. **Transaction Isolation:** Serializable level to prevent race conditions
6. **Validation First:** Input validation before processing to prevent errors
7. **Generic Error Messages:** Don't expose internal implementation details

### Alternatives Considered

**Alternative 1: Keep Sync but Optimize**
- Rejected: Still violates zero-based budgeting principles (retroactive)
- Even with optimization, conceptually flawed

**Alternative 2: Automatic Coverage When RTA Available**
- Rejected: Removes user control, violates intentionality principle
- Users should explicitly choose allocations

**Alternative 3: Remove Underfunded Detection**
- Rejected: Users need visibility into CC debt coverage
- Detection is correct and helpful

**Alternative 4: Complex Batch Operations**
- Deferred: Start simple with single-category coverage
- Can add batch operations as future enhancement

### References

- Domain Expert Validation Report (above)
- Security Auditor Review Report (above)
- Existing codebase: `internal/application/allocation_service.go`
- Existing codebase: `internal/application/transaction_service.go`
- Zero-Based Budgeting Principles (documented in project)

---

**END OF SPECIFICATION**
