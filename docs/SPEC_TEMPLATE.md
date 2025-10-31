# Specification: [Feature Name]

**Status:** Draft | Review | Approved | Implemented
**Created:** YYYY-MM-DD
**Author:** [Name or AI-assisted]
**Validated:** Yes | No | Pending

---

## Executive Summary

[1-2 paragraph overview of the feature and its value]

---

## Business Requirements

### User Stories

**As a** [user type]
**I want to** [action]
**So that** [benefit]

**As a** [user type]
**I want to** [action]
**So that** [benefit]

### Success Criteria

What does "done" look like from a business perspective?

- [ ] Users can [accomplish goal 1]
- [ ] Users can [accomplish goal 2]
- [ ] System maintains [constraint]

---

## Domain Validation

### Budget Domain Expert Review

**Status:** ✅ Approved | ⚠️ Concerns | ❌ Rejected | ⏳ Pending

**Findings:**
[Summary of budget-domain-expert agent review]

**Zero-Based Budgeting Impact:**
- Impact on Ready to Assign: [None | Affects calculation because...]
- Impact on Category Available: [None | Affects rollover because...]
- Impact on Allocations: [None | Changes how allocations work because...]
- Impact on Credit Cards: [None | Affects credit card logic because...]

**Formulas Verified:**
- [Formula 1]: ✅ Correct
- [Formula 2]: ✅ Correct

**Domain Constraints:**
- [Constraint 1]: [Explanation]
- [Constraint 2]: [Explanation]

**Recommendations:**
- [Recommendation from domain expert]

---

## Security Review

### Security Auditor Review

**Status:** ✅ Approved | ⚠️ Concerns | ❌ Rejected | ⏳ Pending

**Security Considerations:**
[Summary of security-auditor agent review]

**Identified Risks:**
- **Risk:** [SQL Injection in...]
  **Mitigation:** [Use parameterized queries]

- **Risk:** [Insufficient validation of...]
  **Mitigation:** [Add validation for...]

**Security Requirements:**
- [ ] Input validation for [specific fields]
- [ ] Authentication required for [endpoints]
- [ ] Authorization checks for [operations]
- [ ] Data protection for [sensitive fields]
- [ ] SQL injection prevention via parameterized queries
- [ ] Error messages sanitized (no info leakage)

---

## Technical Design

### Architecture Compliance

**Clean Architecture Layers:**

**Domain Layer** (`internal/domain/`):
- New entities: [Entity1, Entity2]
- New repository interfaces: [Repository1Interface]
- Validation methods: [Entity.Validate()]

**Application Layer** (`internal/application/`):
- New services: [Service1, Service2]
- Modified services: [ExistingService]
- Business logic: [Description]

**Infrastructure Layer** (`internal/infrastructure/`):
- Repository implementations: [Repository1]
- HTTP handlers: [Handler1]
- Router changes: [New routes]

**Dependencies:** ✅ Point inward (Infrastructure → Application → Domain)

---

### Database Schema Changes

```sql
-- New tables
CREATE TABLE table_name (
    id TEXT PRIMARY KEY,
    field1 TEXT NOT NULL,
    amount INTEGER NOT NULL,  -- Always INTEGER for money (cents)
    created_at DATETIME NOT NULL,
    updated_at DATETIME NOT NULL,
    FOREIGN KEY (field1) REFERENCES other_table(id) ON DELETE CASCADE
);

-- Indexes
CREATE INDEX idx_table_field ON table_name(field1);
CREATE INDEX idx_table_date ON table_name(created_at);

-- Modified tables (if any)
ALTER TABLE existing_table ADD COLUMN new_field TEXT;
```

**Migration Strategy:**
- Backward compatible: Yes | No
- Data migration required: Yes | No
- Migration script: [Description or location]

**Schema Decisions:**
- Amount storage: INTEGER (cents) ✅
- ID type: TEXT (UUID) ✅
- Timestamps: DATETIME (UTC) ✅
- Foreign keys: CASCADE delete ✅

---

### API Design

#### Endpoint 1: [METHOD] /api/resource

**Purpose:** [What this endpoint accomplishes]

**Request:**
```
Method: POST
Path: /api/resource
Headers:
  Content-Type: application/json

Body:
{
  "field1": "string",
  "field2": 12345,
  "field3": "2024-01-01T00:00:00Z"
}
```

**Validation:**
- `field1`: Required, non-empty string, max length 255
- `field2`: Required, positive integer (cents)
- `field3`: Required, valid RFC3339 timestamp

**Response:**

*Success (201 Created):*
```json
{
  "id": "uuid",
  "field1": "string",
  "field2": 12345,
  "field3": "2024-01-01T00:00:00Z",
  "created_at": "2024-01-01T00:00:00Z",
  "updated_at": "2024-01-01T00:00:00Z"
}
```

*Error (400 Bad Request):*
```json
{
  "error": "field1 is required"
}
```

*Error (404 Not Found):*
```json
{
  "error": "resource not found"
}
```

**Side Effects:**
- Updates account balance by [amount]
- Creates allocation for [category]
- Triggers [calculation]

**curl Example:**
```bash
curl -X POST http://localhost:8080/api/resource \
  -H "Content-Type: application/json" \
  -d '{
    "field1": "value",
    "field2": 12345,
    "field3": "2024-01-01T00:00:00Z"
  }'
```

---

#### Endpoint 2: [METHOD] /api/resource/{id}

[Same detailed structure as above]

---

### Service Layer Design

#### New Services

**ServiceName** (`internal/application/service_name.go`)

```go
type ServiceName struct {
    repo domain.RepositoryInterface
}

// CreateResource creates a new resource
func (s *ServiceName) CreateResource(resource *domain.Entity) error {
    // Business logic
    // Validation
    // Call repository
}

// GetResource retrieves a resource by ID
func (s *ServiceName) GetResource(id string) (*domain.Entity, error) {
    // Business logic
}
```

**Business Rules:**
1. [Rule 1]: When [condition], then [action]
2. [Rule 2]: Must [constraint] before [operation]
3. [Rule 3]: Cannot [action] if [condition]

**Error Handling:**
- Returns `ErrNotFound` if resource doesn't exist
- Returns `ErrInvalidInput` if validation fails
- Returns `ErrConflict` if unique constraint violated

#### Modified Services

**ExistingService** (`internal/application/existing_service.go`)

New methods:
```go
func (s *ExistingService) NewMethod(param Type) (Result, error) {
    // Implementation
}
```

---

### Repository Changes

#### New Repositories

**RepositoryName** (`internal/infrastructure/repository/repository_name.go`)

Implements: `domain.RepositoryInterface`

```go
type RepositoryName struct {
    db *sql.DB
}

func (r *RepositoryName) Create(entity *domain.Entity) error {
    query := `INSERT INTO table_name (...) VALUES (?)`
    _, err := r.db.Exec(query, entity.Field1, ...)
    return err
}
```

**SQL Queries:**
- All queries use parameterized statements (SQL injection prevention)
- Proper error handling and rollback on failure
- Transaction boundaries for multi-step operations

#### Modified Repositories

**ExistingRepository**

New methods: [Description]

---

## Test Plan

### Unit Tests

#### Service Tests

**Test:** `TestServiceName_CreateResource`
- **Given:** Valid resource data
- **When:** CreateResource is called
- **Then:** Resource is created successfully
- **Mock:** Repository returns nil error

**Test:** `TestServiceName_CreateResource_ValidationError`
- **Given:** Invalid resource data (empty required field)
- **When:** CreateResource is called
- **Then:** Returns ErrInvalidInput
- **Mock:** N/A (validation happens before repository call)

**Test:** `TestServiceName_CreateResource_RepositoryError`
- **Given:** Valid resource data
- **When:** CreateResource is called and repository fails
- **Then:** Returns wrapped error
- **Mock:** Repository returns database error

---

### Integration Tests

#### Repository Tests

**Test:** `TestRepository_Create`
- **Given:** In-memory test database with schema
- **When:** Create is called with valid entity
- **Then:** Entity is inserted and can be retrieved
- **Verify:** Row exists in database

**Test:** `TestRepository_Create_UniqueConstraint`
- **Given:** Existing entity with same unique field
- **When:** Create is called with duplicate
- **Then:** Returns constraint violation error
- **Verify:** Database unchanged

---

### API Tests

#### Handler Tests

**Test:** `TestHandler_CreateResource`
- **Given:** Valid JSON request body
- **When:** POST /api/resource is called
- **Then:** Returns 201 Created with resource
- **Mock:** Service returns created entity

**Test:** `TestHandler_CreateResource_InvalidJSON`
- **Given:** Malformed JSON
- **When:** POST /api/resource is called
- **Then:** Returns 400 Bad Request
- **Verify:** Error message is descriptive

**Test:** `TestHandler_CreateResource_ValidationError`
- **Given:** Valid JSON with invalid data
- **When:** POST /api/resource is called
- **Then:** Returns 400 Bad Request
- **Mock:** Service returns ErrInvalidInput

---

### Budget Domain Tests

**Test:** `TestFeature_ReadyToAssignCalculation`
- **Given:** [Initial state]
- **When:** [Feature action]
- **Then:** Ready to Assign = [expected value]
- **Verify:** Formula is correct

**Test:** `TestFeature_RolloverBehavior`
- **Given:** Unspent allocation from previous period
- **When:** New period starts
- **Then:** Available includes rollover
- **Verify:** Rollover calculation correct

---

### Manual Testing Scenarios

1. **Happy Path:**
   - Step 1: [User action]
   - Step 2: [Expected result]
   - Step 3: [Verification]

2. **Error Scenario:**
   - Step 1: [Invalid action]
   - Step 2: [Expected error]
   - Step 3: [Verification]

---

## Implementation Checklist

### Phase 1: Domain Layer
- [ ] Create `Entity` in `internal/domain/entity.go`
- [ ] Define `RepositoryInterface` in `internal/domain/repository.go`
- [ ] Add `Entity.Validate()` method
- [ ] Add domain error types if needed

### Phase 2: Application Layer
- [ ] Create `Service` in `internal/application/service.go`
- [ ] Implement `CreateResource(entity) error`
- [ ] Implement `GetResource(id) (*Entity, error)`
- [ ] Implement `UpdateResource(entity) error`
- [ ] Implement `DeleteResource(id) error`
- [ ] Add comprehensive error handling

### Phase 3: Infrastructure - Repository
- [ ] Create `Repository` in `internal/infrastructure/repository/repository.go`
- [ ] Implement `Create(entity) error`
- [ ] Implement `GetByID(id) (*Entity, error)`
- [ ] Implement `Update(entity) error`
- [ ] Implement `Delete(id) error`
- [ ] Use parameterized queries (SQL injection prevention)

### Phase 4: Infrastructure - HTTP
- [ ] Create `Handler` in `internal/infrastructure/http/handlers/handler.go`
- [ ] Implement `CreateResource(w, r)`
- [ ] Implement `GetResource(w, r)`
- [ ] Implement `UpdateResource(w, r)`
- [ ] Implement `DeleteResource(w, r)`
- [ ] Add proper status codes (200, 201, 400, 404, 500)

### Phase 5: Infrastructure - Router
- [ ] Add routes in `internal/infrastructure/http/router.go`
- [ ] POST /api/resource
- [ ] GET /api/resource/{id}
- [ ] PUT /api/resource/{id}
- [ ] DELETE /api/resource/{id}

### Phase 6: Database
- [ ] Update schema in `internal/infrastructure/database/sqlite.go`
- [ ] Add CREATE TABLE statement
- [ ] Add indexes
- [ ] Test schema creation

### Phase 7: Dependency Injection
- [ ] Update `cmd/server/main.go`
- [ ] Initialize repository with DB
- [ ] Initialize service with repository
- [ ] Initialize handler with service
- [ ] Wire to router

### Phase 8: Testing
- [ ] Generate unit tests for service
- [ ] Generate integration tests for repository
- [ ] Generate API tests for handlers
- [ ] Run all tests: `go test ./... -v`
- [ ] Verify coverage: `go test ./... -cover`

### Phase 9: Documentation
- [ ] Update `Claude.md` with new entity/endpoints
- [ ] Add godoc comments to exported functions
- [ ] Update README if needed
- [ ] Create API documentation

### Phase 10: Verification
- [ ] Run `/check-architecture` - verify clean architecture
- [ ] Invoke `code-reviewer` agent - code review
- [ ] Invoke `security-auditor` agent - security check
- [ ] Manual testing with curl
- [ ] Performance testing (if applicable)

---

## Acceptance Criteria

### Functional Requirements
- [ ] Requirement 1: [Specific functionality works]
- [ ] Requirement 2: [Specific functionality works]
- [ ] Edge case 1: [Handled correctly]
- [ ] Edge case 2: [Handled correctly]

### Non-Functional Requirements
- [ ] All tests passing
- [ ] Test coverage > 80%
- [ ] Clean architecture maintained
- [ ] Security requirements met
- [ ] Performance acceptable (< Xms response time)
- [ ] Documentation complete

### Code Quality
- [ ] No architecture violations detected
- [ ] Go best practices followed
- [ ] Error handling comprehensive
- [ ] Input validation complete
- [ ] SQL injection prevented
- [ ] Code reviewed and approved

### Budget Domain Specific
- [ ] Ready to Assign calculation correct
- [ ] Category Available calculation correct
- [ ] Rollover behavior correct (if applicable)
- [ ] Credit card logic correct (if applicable)
- [ ] Money stored as INTEGER (cents)

---

## Risks and Mitigations

### Risk: [Describe potential risk]
**Impact:** High | Medium | Low
**Likelihood:** High | Medium | Low
**Mitigation:** [How to address or prevent]

### Risk: Breaking change to existing API
**Impact:** High
**Likelihood:** Medium
**Mitigation:** Version the API, provide migration path, deprecation notice

### Risk: Performance degradation with large datasets
**Impact:** Medium
**Likelihood:** Low
**Mitigation:** Add indexes, implement pagination, load testing

---

## Dependencies

### External Dependencies
- None | Requires [library] version X.Y

### Internal Dependencies
- Depends on [existing feature]
- Must be implemented before [future feature]

### Database
- Requires SQLite 3.x
- Requires schema version X

---

## Future Enhancements

Features intentionally deferred for later implementation:

1. [Enhancement 1]: [Why deferred]
2. [Enhancement 2]: [Why deferred]

---

## Approval Checklist

- [x] Domain Expert: Validated
- [x] Security Auditor: Reviewed
- [ ] Technical Lead: Approved
- [ ] Product Owner: Approved
- [ ] Stakeholder: Approved
- [ ] Developer: Ready to implement

---

## Implementation

**This specification is ready for implementation.**

To implement this specification, run:

```bash
/implement-spec docs/spec-[feature-name].md
```

The `/implement-spec` command will:
1. Read this specification
2. Invoke budget-domain-expert to validate (already done)
3. Create implementation following the technical design
4. Invoke test-generator to create tests from test plan
5. Invoke code-reviewer to verify implementation
6. Guide through verification and deployment

**Estimated Effort:**
- Complexity: Low | Medium | High
- Estimated time: X hours/days
- Files to create: ~N new files
- Files to modify: ~M existing files

---

## Notes and Decisions

### Decision 1: [Topic]
**Decision:** [What was decided]
**Rationale:** [Why this decision was made]
**Alternatives considered:** [Other options]

### Decision 2: [Topic]
**Decision:** [What was decided]
**Rationale:** [Why this decision was made]

---

## Appendix

### Glossary
- **Term 1**: Definition
- **Term 2**: Definition

### References
- [Related specification]
- [External documentation]
- [Design document]

---

**Version History**

- v1.0 (YYYY-MM-DD): Initial specification
- v1.1 (YYYY-MM-DD): Updated based on review feedback
