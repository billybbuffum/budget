# Development Workflows for Budget Application

> Common workflows for implementing features, fixing bugs, and maintaining the Budget application using Claude Code tools.

---

## Table of Contents

1. [Implementing a Specification](#implementing-a-specification)
2. [Creating a New Feature](#creating-a-new-feature)
3. [Adding an API Endpoint](#adding-an-api-endpoint)
4. [Reviewing Code](#reviewing-code)
5. [Fixing a Bug](#fixing-a-bug)
6. [Refactoring Code](#refactoring-code)
7. [Testing](#testing)
8. [Deployment](#deployment)

---

## Implementing a Specification

### Quick Method (Automated)

**Best for**: Most specifications, automated workflow

```
/implement-spec https://github.com/billybbuffum/budget/blob/branch/docs/spec.md
```

**What happens**:
1. Fetches and analyzes spec
2. Invokes budget-domain-expert for validation
3. Creates implementation plan
4. Invokes test-generator for test strategy
5. Invokes code-reviewer for verification
6. Guides you through implementation step-by-step

**Time**: 15-60 minutes depending on complexity

---

### Manual Method (Full Control)

**Best for**: Complex specs requiring careful consideration

**Step 1: Read and understand**
```
"Please read the specification at [URL] and summarize the requirements"
```

**Step 2: Validate business logic**
```
"Invoke the budget-domain-expert agent to validate the business logic requirements in this spec"
```

**Step 3: Create implementation**
- For new features:
  ```
  /new-feature feature-name
  ```
- For modifications:
  ```
  "Implement the changes following clean architecture patterns"
  ```

**Step 4: Generate tests**
```
"Invoke the test-generator agent to create comprehensive tests for the changes"
```

**Step 5: Review code**
```
"Invoke the code-reviewer agent to review the implementation"
```

**Step 6: Verify**
```
/run-tests
/check-architecture
```

**Time**: 30-90 minutes

---

### Hybrid Method (Best of Both)

**Best for**: Most situations

```
"Please implement this specification: [URL]

After reading it:
1. Invoke budget-domain-expert to validate
2. Create implementation
3. Invoke test-generator for tests
4. Invoke code-reviewer for final check"
```

Claude will follow your instructions explicitly.

**Time**: 20-60 minutes

---

## Creating a New Feature

### Method 1: Scaffold with Command

```
/new-feature recurring-transactions
```

**What it creates**:
- Domain entity and repository interface
- Application service with business logic
- Infrastructure repository implementation
- HTTP handler
- Route registration
- Database schema

**Then**:
```
"Invoke test-generator to create tests for RecurringTransactionService"
"Invoke code-reviewer to verify the implementation"
```

---

### Method 2: Manual Creation

**Step 1: Domain layer**
```
"Create a RecurringTransaction entity in internal/domain/ following the pattern of existing entities"
```

**Step 2: Application layer**
```
"Create RecurringTransactionService in internal/application/ with CRUD operations"
```

**Step 3: Infrastructure layer**
```
"Create repository implementation in internal/infrastructure/repository/"
"Create HTTP handler in internal/infrastructure/http/handlers/"
```

**Step 4: Wire it up**
```
"Update router.go and main.go to wire the new feature"
```

---

## Adding an API Endpoint

### Quick Method

```
/new-endpoint "Add GET /api/allocations/history endpoint that returns allocation history for a category"
```

**Provides**:
- Handler method implementation
- Route registration
- Service method (if needed)
- Request/response format
- Error handling

---

### Manual Method

```
"Add a new endpoint GET /api/allocations/history

Requirements:
- Query parameter: category_id (required)
- Query parameter: limit (optional, default 10)
- Returns array of allocations ordered by period DESC
- Include proper error handling"
```

Then test it:
```
/test-endpoint /api/allocations/history?category_id=xxx
```

---

## Reviewing Code

### Full Code Review

```
/review-pr
```

**What it checks**:
- Clean architecture compliance
- Go best practices
- Budget domain logic
- Security issues
- Test coverage
- Documentation

---

### Targeted Reviews

**Review specific component**:
```
"Invoke the code-reviewer agent to review AllocationService focusing on the rollover calculation logic"
```

**Security audit**:
```
"Invoke the security-auditor agent to check for vulnerabilities in the transaction endpoints"
```

**Domain logic validation**:
```
"Invoke the budget-domain-expert agent to verify the Ready to Assign calculation in AllocationService"
```

**Architecture check**:
```
/check-architecture
```

---

## Fixing a Bug

### Workflow

**Step 1: Understand the bug**
```
"The allocation rollover is not calculating correctly. Here's the issue: [description]"
```

**Step 2: Locate the code**
```
"Find all code related to allocation rollover calculation"
```

**Step 3: Validate current logic**
```
"Invoke the budget-domain-expert agent to review the current rollover implementation and identify the issue"
```

**Step 4: Fix the bug**
```
"Fix the rollover calculation based on the domain expert's findings"
```

**Step 5: Add tests**
```
"Invoke test-generator to create tests that verify the rollover bug is fixed and prevent regression"
```

**Step 6: Verify**
```
/run-tests
"Invoke code-reviewer to verify the fix"
```

---

## Refactoring Code

### Safe Refactoring Workflow

**Step 1: Identify code smell**
```
"I want to refactor AllocationService - it's getting too large and has duplicate code"
```

**Step 2: Get refactoring plan**
```
"Invoke the refactoring-assistant agent to analyze AllocationService and propose refactoring improvements"
```

**Step 3: Ensure tests exist**
```
"Are there tests for AllocationService? If not, invoke test-generator to create them first"
```

**Step 4: Execute refactoring**
```
"Implement the refactoring plan proposed by the refactoring-assistant"
```

**Step 5: Verify**
```
/run-tests
/check-architecture
"Invoke code-reviewer to verify the refactoring maintains clean architecture"
```

---

## Testing

### Generate Tests

**For a service**:
```
"Invoke the test-generator agent to create comprehensive tests for AccountService"
```

**For a repository**:
```
"Invoke the test-generator agent to create integration tests for AccountRepository"
```

**For a handler**:
```
"Invoke the test-generator agent to create end-to-end tests for AccountHandler"
```

---

### Run Tests

**All tests**:
```
/run-tests
```

**Specific package**:
```bash
go test ./internal/application/... -v
```

**With coverage**:
```bash
go test ./... -cover -coverprofile=coverage.out
go tool cover -html=coverage.out
```

---

### Test an API Endpoint

```
/test-endpoint /api/accounts
```

**Or test manually**:
```bash
# Create account
curl -X POST http://localhost:8080/api/accounts \
  -H "Content-Type: application/json" \
  -d '{"name":"Test","type":"checking","balance":100000}'

# Get accounts
curl http://localhost:8080/api/accounts
```

---

## Deployment

### Pre-Deployment Checklist

```
"Run pre-deployment checks:
1. /run-tests to ensure all tests pass
2. /check-architecture to verify clean architecture
3. Invoke security-auditor to check for vulnerabilities
4. Invoke code-reviewer for final code review"
```

---

### Build and Deploy

**Build Docker image**:
```bash
docker-compose build
```

**Start locally**:
```bash
docker-compose up -d
```

**Verify health**:
```bash
curl http://localhost:8080/health
```

**Test endpoints**:
```
/test-endpoint /api/accounts/summary
/test-endpoint /api/allocations/ready-to-assign
```

---

## Advanced Workflows

### Full Feature Implementation (Example)

**Scenario**: Implement category groups feature

**Week 1: Planning and Domain**
```
Day 1:
- "Invoke budget-domain-expert to discuss category groups feature"
- "Design the CategoryGroup entity and its relationship to Category"

Day 2:
- /new-feature category-groups
- "Create domain entity and repository interface"
```

**Week 2: Implementation**
```
Day 3-4:
- "Implement CategoryGroupService with CRUD operations"
- "Create repository implementation"
- "Add HTTP handlers and routes"

Day 5:
- "Invoke test-generator to create comprehensive test suite"
- /run-tests
```

**Week 3: Review and Deploy**
```
Day 6:
- "Invoke code-reviewer for full review"
- "Invoke security-auditor for security check"
- Fix any issues found

Day 7:
- /check-architecture
- /run-tests
- "Update documentation"
- Create PR
```

---

### Bug Fix with Root Cause Analysis

```
"There's a bug: credit card spending doesn't update the payment category correctly

Steps:
1. Invoke budget-domain-expert to explain correct credit card behavior
2. Find the relevant code in TransactionService
3. Invoke code-reviewer to identify the bug
4. Fix the issue
5. Invoke test-generator to create regression tests
6. /run-tests to verify fix"
```

---

### Performance Optimization

```
"The /api/allocations/summary endpoint is slow

Steps:
1. Analyze the current implementation
2. Invoke refactoring-assistant to suggest optimizations
3. Check database queries for missing indexes
4. Implement optimizations
5. Invoke code-reviewer to verify changes
6. Test performance improvement"
```

---

## Quick Reference

### Most Common Workflows

| Task | Command |
|------|---------|
| Implement spec | `/implement-spec <url>` |
| New feature | `/new-feature <name>` |
| New endpoint | `/new-endpoint <description>` |
| Review code | `/review-pr` |
| Check architecture | `/check-architecture` |
| Run tests | `/run-tests` |
| Test endpoint | `/test-endpoint <path>` |
| Generate docs | `/generate-docs` |

### Most Common Agent Invocations

| Task | Agent |
|------|-------|
| Validate budget logic | `budget-domain-expert` |
| Generate tests | `test-generator` |
| Review code | `code-reviewer` |
| Security audit | `security-auditor` |
| Refactor code | `refactoring-assistant` |
| Document API | `api-documenter` |

---

## Tips and Best Practices

### 1. Start with Domain Expert

For any budget-related change:
```
"Invoke budget-domain-expert to validate this approach before I implement it"
```

### 2. Generate Tests Early

Don't wait until the end:
```
"Invoke test-generator now so I can use TDD"
```

### 3. Use Architecture Checks

After any significant change:
```
/check-architecture
```

### 4. Combine Commands and Agents

```
/new-feature user-auth
"Now invoke test-generator and code-reviewer"
```

### 5. Explicit is Better

Instead of:
```
"Review my code"
```

Say:
```
"Invoke code-reviewer agent to review AllocationService"
```

---

## Troubleshooting Workflows

### "I'm not sure which approach to use"

**Simple change** (< 50 lines):
- Direct implementation, no commands/agents

**Medium change** (50-200 lines):
- Use `/new-endpoint` or direct implementation
- Invoke code-reviewer after

**Complex change** (> 200 lines):
- Use `/new-feature` or `/implement-spec`
- Invoke all relevant agents

### "An agent didn't find issues"

That's good! Means your code is solid.

### "I want to skip some steps"

The workflows are guidelines, not strict rules. Adapt as needed.

---

**Last Updated**: October 31, 2025

*These workflows are living documents. Update them as you discover better approaches.*
