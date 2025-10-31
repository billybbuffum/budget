---
description: Review a pull request with budget app specific checks
allowed-tools: [Read, Grep, Glob, Bash, Task]
---

# Review Pull Request

Perform a comprehensive code review for the Budget application.

## Review Process

### 1. Get PR Changes
```bash
# See what files changed
git diff --name-only main...HEAD

# See the actual changes
git diff main...HEAD
```

### 2. Invoke Code Reviewer Agent
Use the code-reviewer sub agent to perform detailed review:
- Clean architecture compliance
- Go best practices
- Budget domain logic correctness
- Code quality

### 3. Additional Checks

**Database Changes:**
- [ ] Schema changes are in `sqlite.go`
- [ ] Migrations are backward compatible
- [ ] Amounts stored as INTEGER (cents)
- [ ] Proper foreign key constraints
- [ ] Indexes added where needed

**API Changes:**
- [ ] Endpoints follow RESTful conventions
- [ ] Status codes are appropriate
- [ ] Request/response formats are documented
- [ ] Error handling is consistent

**Tests:**
- [ ] Tests are included for new features
- [ ] Existing tests still pass
- [ ] Critical business logic is tested
- [ ] Edge cases are covered

**Documentation:**
- [ ] Claude.md updated with new features
- [ ] API endpoints documented
- [ ] Complex logic has comments

**Security:**
- [ ] No SQL injection vulnerabilities
- [ ] Input validation is present
- [ ] No sensitive data in logs
- [ ] Error messages don't leak info

### 4. Run Tests
```bash
# Run Go tests
go test ./...

# Check for compilation errors
go build ./...
```

### 5. Check Architecture
Verify layer separation:
- Domain: No external dependencies
- Application: Only domain interfaces
- Infrastructure: Implements interfaces

### 6. Budget Domain Verification

**If allocation logic changed:**
- [ ] Ready to Assign = Balance - Allocated
- [ ] Available includes rollover
- [ ] One allocation per category per period

**If transaction logic changed:**
- [ ] Balance updates are atomic
- [ ] Creating/updating/deleting updates balance
- [ ] Amounts are in cents

**If credit card logic changed:**
- [ ] Balances are negative
- [ ] Payment category auto-creation
- [ ] Payment allocation logic

## Review Checklist Summary

### Critical Issues 🔴
Look for issues that MUST be fixed:
- Architecture violations
- Security vulnerabilities
- Data corruption risks
- Breaking changes without migration

### Important Issues 🟡
Look for issues that SHOULD be fixed:
- Go best practice violations
- Missing error handling
- Incomplete validation
- Poor code organization

### Suggestions 🔵
Nice to have improvements:
- Performance optimizations
- Code clarity improvements
- Additional tests
- Better naming

## Final Review Output

Provide a summary with:
1. **Overall assessment**
2. **Critical issues** (must fix)
3. **Important issues** (should fix)
4. **Suggestions** (nice to have)
5. **Good patterns** (positive feedback)
6. **Recommendation**: Approve / Request Changes / Needs Discussion

## Commands to Run

```bash
# Check for compilation errors
go build ./...

# Run tests (when they exist)
go test ./...

# Check for common issues
go vet ./...

# Format code
go fmt ./...
```
