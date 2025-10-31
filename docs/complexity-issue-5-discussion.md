# Issue #5: Vestigial BudgetState Model

**Priority:** 🟢 Low
**Status:** 📋 Not Started
**Location:** `internal/domain/budget_state.go`

---

## Problem Statement

The `BudgetState` model exists with a `ReadyToAssign` field, but this field is deprecated and no longer used. The comment at `allocation_service.go:452-459` explicitly states:

```go
// GetReadyToAssign reads the Ready to Assign amount from the database
// DEPRECATED: This now returns 0 as Ready to Assign is calculated per-period
// Use CalculateReadyToAssignForPeriod instead
```

This creates:
- Confusion for developers (why does this model exist?)
- Unused database schema
- Potential source of bugs if accidentally used
- Maintenance burden

---

## Current Implementation

### Domain Model

```go
// internal/domain/budget_state.go
package domain

import "time"

// BudgetState represents the current state of the budget
// This is a singleton record that tracks values that need to be coordinated
type BudgetState struct {
    ID            string    `json:"id"`
    ReadyToAssign int64     `json:"ready_to_assign"` // Amount available to allocate (in cents)
    UpdatedAt     time.Time `json:"updated_at"`
}
```

### Database Schema

```sql
-- In migrations
CREATE TABLE budget_state (
    id TEXT PRIMARY KEY,
    ready_to_assign INTEGER NOT NULL DEFAULT 0,
    updated_at TIMESTAMP NOT NULL
);
```

### Deprecated Usage

```go
// internal/application/allocation_service.go:452-459
func (s *AllocationService) GetReadyToAssign(ctx context.Context) (int64, error) {
    state, err := s.budgetStateRepo.Get(ctx)
    if err != nil {
        return 0, fmt.Errorf("failed to get budget state: %w", err)
    }
    return state.ReadyToAssign, nil
}
```

---

## Issues Identified

### 1. Dead Code
- Model defined but not actively used
- Repository exists but deprecated
- Database table exists but not updated

### 2. Developer Confusion
- New developers might use the deprecated model
- Unclear why it exists
- Adds cognitive load

### 3. Maintenance Burden
- Must maintain unused code
- Database migrations include it
- Tests might cover deprecated functionality

### 4. Potential Bugs
- If someone accidentally uses `GetReadyToAssign()` instead of `CalculateReadyToAssignForPeriod()`
- Data could become stale/inconsistent

---

## Proposed Solution

**REMOVE** the entire `BudgetState` model and related code.

### Why This Works

1. **RTA is now calculated:** `CalculateReadyToAssignForPeriod` computes RTA dynamically
2. **No data loss:** RTA was derived data, not source data
3. **Simplification:** One less model to maintain
4. **Clarity:** Makes it obvious there's only one way to get RTA

---

## Code to Remove

### 1. Domain Model
- File: `internal/domain/budget_state.go`

### 2. Repository Interface
```go
// From internal/domain/repository.go
type BudgetStateRepository interface {
    Get(ctx context.Context) (*BudgetState, error)
    Update(ctx context.Context, state *BudgetState) error
}
```

### 3. Repository Implementation
- File: `internal/infrastructure/repository/budget_state_repository.go`

### 4. Service References
- `internal/application/allocation_service.go:17` (field)
- `internal/application/allocation_service.go:27` (constructor param)
- `internal/application/allocation_service.go:452-459` (method)
- `internal/application/transaction_service.go:18` (field)
- `internal/application/transaction_service.go:26` (constructor param)

### 5. Handler References
- Any HTTP handlers using the deprecated method

---

## Migration Strategy

### Database Migration

Create a migration to drop the table:

```sql
-- migrations/NNNN_remove_budget_state.sql
DROP TABLE IF EXISTS budget_state;
```

### Code Migration

1. Remove model file
2. Remove repository implementation
3. Remove from repository interfaces
4. Remove from service constructors
5. Remove deprecated `GetReadyToAssign` method
6. Update any code calling the deprecated method
7. Update tests

---

## Discussion Notes

### Session 1: [Date]
*Discussion notes will be added here as we discuss this issue*

---

## Questions to Address

1. **Q:** Is there ANY current usage of `GetReadyToAssign()`?
   - **A:** [To be discussed - need to grep codebase]

2. **Q:** Should we keep the model but remove the deprecated field?
   - **A:** [To be discussed - does BudgetState serve any other purpose?]

3. **Q:** Do we need a multi-step deprecation (warning → removal)?
   - **A:** [To be discussed - who are the users?]

4. **Q:** Are there any other fields in BudgetState that are used?
   - **A:** [To be discussed]

---

## Implementation Checklist

### Phase 1: Analysis
- [ ] Search codebase for ALL usages of `BudgetState`
- [ ] Search codebase for ALL usages of `GetReadyToAssign`
- [ ] Verify no external API dependencies
- [ ] Check if any UI code uses this

### Phase 2: Code Removal
- [ ] Remove domain model file
- [ ] Remove repository interface
- [ ] Remove repository implementation
- [ ] Update service constructors
- [ ] Remove deprecated method
- [ ] Update any callers

### Phase 3: Database
- [ ] Create migration to drop table
- [ ] Test migration (up and down)
- [ ] Verify no foreign keys reference this table

### Phase 4: Testing
- [ ] Remove tests for BudgetState
- [ ] Verify all remaining tests pass
- [ ] Integration test for RTA calculation

---

## Search for Usages

```bash
# Find all references to BudgetState
grep -r "BudgetState" --include="*.go" .

# Find all references to GetReadyToAssign
grep -r "GetReadyToAssign" --include="*.go" .

# Find all references to budgetStateRepo
grep -r "budgetStateRepo" --include="*.go" .
```

---

## Risk Assessment

### Low Risk ✅
- Model is already deprecated
- Removal aligns with current design
- RTA calculation exists and works

### Potential Issues
- **External API:** If REST API exposes this, clients might break
  - **Mitigation:** Check API routes, deprecate endpoint first if needed
- **UI Dependencies:** Frontend might call the old endpoint
  - **Mitigation:** Search for API calls in frontend code

---

## Alternative Solutions Considered

### Alternative 1: Keep Model, Remove Field
- Remove `ready_to_assign` field but keep model for future use
- **Rejected:** Model serves no purpose without the field

### Alternative 2: Deprecation Period
- Mark as deprecated, remove in next version
- **Rejected:** Already marked deprecated, no active users

### Alternative 3: Convert to View/Cache
- Keep as a materialized view for performance
- **Rejected:** Premature optimization, calculate-on-demand works fine

---

## Files to Update

```
internal/domain/budget_state.go                          [DELETE]
internal/domain/repository.go                            [EDIT - remove interface]
internal/infrastructure/repository/budget_state_repository.go [DELETE]
internal/application/allocation_service.go               [EDIT - remove field, param, method]
internal/application/transaction_service.go              [EDIT - remove field, param]
cmd/server/main.go                                       [EDIT - remove from DI]
migrations/NNNN_remove_budget_state.sql                  [CREATE]
```

---

## Testing Verification

After removal, verify:

```bash
# All tests pass
go test ./...

# No references remain
grep -r "BudgetState" --include="*.go" . | grep -v "test"

# API still works
curl http://localhost:8080/api/allocations/summary?period=2024-11
```

---

## Related Issues

None - this is standalone cleanup

---

## Decision Log

| Date | Decision | Rationale |
|------|----------|-----------|
| TBD  | TBD      | TBD       |

---

**Last Updated:** 2025-10-31
