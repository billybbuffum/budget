---
description: Verify clean architecture compliance across the codebase
allowed-tools: [Read, Grep, Glob]
---

# Check Clean Architecture Compliance

Verify that the Budget application maintains clean architecture principles.

## Clean Architecture Layers

```
┌─────────────────────────────────────┐
│     Infrastructure Layer            │
│  (HTTP, Database, External APIs)    │
│                                     │
│  ┌───────────────────────────────┐ │
│  │   Application Layer           │ │
│  │   (Business Logic Services)   │ │
│  │                               │ │
│  │  ┌─────────────────────────┐ │ │
│  │  │   Domain Layer          │ │ │
│  │  │   (Entities, Rules)     │ │ │
│  │  └─────────────────────────┘ │ │
│  └───────────────────────────────┘ │
└─────────────────────────────────────┘

Dependencies point INWARD →
```

## Verification Steps

### 1. Domain Layer Check

**Location:** `internal/domain/`

**Rules:**
✅ **MUST:**
- Contain only entities and interfaces
- Have NO external dependencies
- Define repository interfaces
- Include entity validation methods

❌ **MUST NOT:**
- Import database packages
- Import HTTP packages
- Import external libraries (except uuid)
- Contain business logic (that's for services)

**Check commands:**
```bash
# Check imports in domain files
grep -r "import" internal/domain/

# Should NOT see:
# - database/sql
# - net/http
# - internal/application
# - internal/infrastructure
```

### 2. Application Layer Check

**Location:** `internal/application/`

**Rules:**
✅ **MUST:**
- Contain business logic services
- Depend ONLY on domain interfaces
- Orchestrate domain entities
- Handle use cases

❌ **MUST NOT:**
- Import HTTP packages (no http.Request, http.ResponseWriter)
- Import database packages (no *sql.DB)
- Contain HTTP handling logic
- Contain SQL queries

**Check commands:**
```bash
# Check for HTTP imports (BAD)
grep -r "net/http" internal/application/

# Check for database imports (BAD)
grep -r "database/sql" internal/application/

# Verify services use interfaces, not concrete types
grep -r "repository\." internal/application/
```

### 3. Infrastructure Layer Check

**Location:** `internal/infrastructure/`

**Rules:**
✅ **MUST:**
- Implement domain repository interfaces
- Handle HTTP requests/responses
- Contain database queries
- Keep handlers thin (parse → call service → respond)

❌ **MUST NOT:**
- Contain business logic (that's for services)
- Have handlers with complex logic
- Have repositories with business rules

**Check commands:**
```bash
# Handlers should NOT have complex business logic
# Look for handlers with >50 lines (potential code smell)
find internal/infrastructure/http/handlers/ -name "*.go" -exec wc -l {} \;
```

### 4. Dependency Direction Check

**Rule:** Dependencies must point INWARD

```
Infrastructure → Application → Domain
     ✓              ✓           ✗ (no dependencies)
```

**Check:**
```bash
# Domain should NOT import application or infrastructure
grep -r "internal/application" internal/domain/
grep -r "internal/infrastructure" internal/domain/

# Application should NOT import infrastructure
grep -r "internal/infrastructure" internal/application/

# Both checks should return NOTHING
```

### 5. Service Pattern Check

**Application Services should:**
```go
type Service struct {
    repo domain.Repository  // ✓ Interface, not concrete type
}

func (s *Service) DoSomething(entity *domain.Entity) error {
    // ✓ Business logic here
    // ✓ Call repository methods
    // ✓ Return errors with context
}
```

**Handlers should:**
```go
func (h *Handler) HandleRequest(w http.ResponseWriter, r *http.Request) {
    // ✓ Parse request
    // ✓ Call service
    // ✓ Return response
    // ✗ NO business logic
}
```

### 6. Repository Pattern Check

**Repositories should:**
```go
type Repository struct {
    db *sql.DB  // ✓ Can have database dependency
}

func (r *Repository) Create(entity *domain.Entity) error {
    // ✓ SQL queries here
    // ✓ Data persistence
    // ✗ NO business logic
}
```

## Common Violations

### ❌ Handler with Business Logic
```go
func (h *Handler) CreateAccount(w http.ResponseWriter, r *http.Request) {
    // Parse request
    // ❌ BAD: Complex calculations in handler
    balance := calculateBalance(transactions)
    // ❌ BAD: Business rules in handler
    if balance < 0 && account.Type != "credit_card" {
        return errors.New("invalid")
    }
    // Should call service instead
}
```

### ❌ Service with Database/HTTP
```go
func (s *Service) CreateAccount(db *sql.DB, r *http.Request) error {
    // ❌ BAD: Service accepts *sql.DB
    // ❌ BAD: Service accepts *http.Request
    // Should use repository interface instead
}
```

### ❌ Repository with Business Logic
```go
func (r *Repository) Create(account *Account) error {
    // ❌ BAD: Business rule in repository
    if account.Balance < 0 && account.Type != "credit_card" {
        return errors.New("invalid balance")
    }
    // Should only handle persistence
}
```

## Architecture Report

After checking, provide a report:

```markdown
# Clean Architecture Compliance Report

## Domain Layer: [✅ Compliant / ⚠️ Issues Found]
[Findings]

## Application Layer: [✅ Compliant / ⚠️ Issues Found]
[Findings]

## Infrastructure Layer: [✅ Compliant / ⚠️ Issues Found]
[Findings]

## Dependency Direction: [✅ Correct / ⚠️ Violations]
[Findings]

## Violations Found
### Critical
- [Violation 1]

### Minor
- [Violation 2]

## Recommendations
1. [Action item 1]
2. [Action item 2]
```

## Quick Check Command

Run these to quickly check for violations:

```bash
# Domain importing application/infrastructure (SHOULD BE EMPTY)
echo "=== Domain Layer Violations ==="
grep -r "internal/application\|internal/infrastructure" internal/domain/ || echo "✅ No violations"

# Application importing infrastructure (SHOULD BE EMPTY)
echo "=== Application Layer Violations ==="
grep -r "internal/infrastructure" internal/application/ || echo "✅ No violations"

# Services with HTTP dependencies (SHOULD BE EMPTY)
echo "=== HTTP in Application Layer ==="
grep -r "net/http" internal/application/ || echo "✅ No violations"

# Services with database dependencies (SHOULD BE EMPTY)
echo "=== Database in Application Layer ==="
grep -r "database/sql" internal/application/ || echo "✅ No violations"
```
