---
description: Create a validated specification for a new feature with domain expert review
argument-hint: <feature-description>
allowed-tools: [Read, Write, Grep, Glob, Task]
---

# Create Feature Specification

Create a comprehensive, validated specification for a new feature before implementation.

## Feature Description
{{arg}}

## Specification Creation Workflow

Follow this process to create a well-validated specification that's ready for implementation.

---

### Phase 1: Requirements Gathering

**1. Understand the feature request**

Based on the feature description, clarify:
- **What** is the feature trying to accomplish?
- **Who** will use this feature?
- **Why** is this feature needed?
- **When** will users interact with it?
- **How** should it work from a user perspective?

**2. Ask clarifying questions** (if needed)

If the feature description is vague, ask:
- What are the key user stories?
- What are the edge cases?
- What are the success criteria?
- Are there performance requirements?
- Are there security considerations?

**3. Document initial requirements**

Create a summary of:
- Core functionality
- User workflows
- Expected outcomes
- Constraints and limitations

---

### Phase 2: Domain Validation

**4. Invoke budget-domain-expert for validation**

```
Invoke the budget-domain-expert agent to:
- Validate the feature aligns with zero-based budgeting principles
- Check for conflicts with existing budget logic
- Verify formulas and calculations are correct
- Identify potential issues with:
  * Ready to Assign calculations
  * Category Available with rollover
  * Credit card logic
  * Allocation rules
  * Transaction handling
- Suggest domain-specific improvements
- Return validation report with approval or concerns
```

**5. Address domain expert findings**

- If approved: proceed to next phase
- If concerns raised: adjust requirements to address issues
- Document any domain-specific constraints or rules

---

### Phase 3: Security Review

**6. Invoke security-auditor for early security review**

```
Invoke the security-auditor agent to:
- Identify potential security vulnerabilities in the design
- Review data handling requirements
- Check for input validation needs
- Verify authentication/authorization considerations
- Identify sensitive data protection requirements
- Suggest security best practices
- Return security considerations and requirements
```

**7. Document security requirements**

From security-auditor findings, document:
- Input validation rules
- Data protection requirements
- Access control needs
- Potential attack vectors to guard against

---

### Phase 4: Technical Design

**8. Design data model**

Define database changes needed:
- New tables/entities
- Modifications to existing tables
- New columns and their types
- Indexes needed
- Foreign key relationships
- Unique constraints

**Remember:**
- Store amounts as INTEGER (cents)
- Use TEXT for UUIDs
- Use DATETIME for timestamps (UTC)
- Add appropriate indexes
- Define CASCADE behavior for foreign keys

**9. Design API contracts**

For each endpoint needed, define:

```
## Endpoint: [METHOD] /api/resource

**Purpose:** [What this endpoint does]

**Request:**
- Method: GET/POST/PUT/DELETE
- Headers: Content-Type: application/json
- Path Parameters: {id}, etc.
- Query Parameters: ?param=value
- Request Body:
  {
    "field": "type",
    "field2": "type"
  }

**Response:**
- Success (200/201):
  {
    "field": "type"
  }
- Error (400/404/500):
  {
    "error": "message"
  }

**Validation:**
- Field1: required, type, constraints
- Field2: optional, type, default

**Side Effects:**
- Updates account balance
- Creates allocation
- etc.
```

**10. Design service layer**

Define business logic needed:
- New services or modifications to existing services
- Service methods and their signatures
- Repository interfaces needed
- Business rules and validation logic
- Transaction boundaries

**11. Plan clean architecture compliance**

Verify design follows clean architecture:
- **Domain layer**: New entities, repository interfaces
- **Application layer**: Services with business logic
- **Infrastructure layer**: Repository implementations, HTTP handlers
- **Dependencies**: Point inward (infrastructure → application → domain)

---

### Phase 5: Test Planning

**12. Define test scenarios**

Create acceptance criteria as test scenarios:

**Unit Tests:**
- Service methods to test
- Expected inputs and outputs
- Error cases to handle
- Edge cases to verify

**Integration Tests:**
- Repository operations to test
- Database constraints to verify
- Transaction rollback scenarios

**API Tests:**
- Endpoint requests to test
- Status codes to verify
- Response formats to validate
- Error handling to check

**Budget Domain Tests:**
- Zero-based budgeting calculations
- Rollover behavior
- Credit card logic (if applicable)
- Allocation rules

**13. Define acceptance criteria**

Clear criteria for "done":
- [ ] Functional requirement 1 met
- [ ] Functional requirement 2 met
- [ ] All tests passing
- [ ] Architecture compliance verified
- [ ] Security requirements met
- [ ] Documentation updated

---

### Phase 6: Specification Document Creation

**14. Generate specification document**

Create `docs/spec-<feature-name>.md` with this structure:

```markdown
# Specification: <Feature Name>

**Status:** Draft
**Created:** YYYY-MM-DD
**Author:** AI-assisted specification
**Validated:** Yes (Domain Expert + Security Auditor)

---

## Executive Summary

[1-2 paragraph overview of the feature]

---

## Business Requirements

### User Stories

As a [user type], I want to [action], so that [benefit].

### Success Criteria

- [ ] Criterion 1
- [ ] Criterion 2

---

## Domain Validation

**Budget Domain Expert Review:**
- Status: ✅ Approved / ⚠️ Concerns
- Findings: [Summary of domain expert feedback]
- Zero-Based Budgeting Impact: [How this affects budgeting logic]
- Formulas Verified: [Any calculations validated]

**Domain Constraints:**
- [Constraint 1]
- [Constraint 2]

---

## Security Review

**Security Auditor Review:**
- Status: ✅ Approved / ⚠️ Concerns
- Security Considerations: [Summary]

**Security Requirements:**
- [ ] Input validation for [fields]
- [ ] Authentication/authorization for [endpoints]
- [ ] Data protection for [sensitive data]
- [ ] SQL injection prevention via parameterized queries

---

## Technical Design

### Architecture Compliance

- **Domain Layer:** [New entities, interfaces]
- **Application Layer:** [Services, business logic]
- **Infrastructure Layer:** [Repositories, handlers]
- **Dependencies:** ✅ Point inward

### Database Schema Changes

```sql
-- New tables
CREATE TABLE ...

-- Modified tables
ALTER TABLE ...

-- New indexes
CREATE INDEX ...
```

**Migration Notes:**
- [Backward compatibility considerations]
- [Data migration required?]

### API Design

#### Endpoint 1: [METHOD] /api/resource

[Full endpoint specification from Phase 4]

#### Endpoint 2: [METHOD] /api/resource/{id}

[Full endpoint specification from Phase 4]

### Service Layer Design

**New Services:**
- ServiceName
  - Method1(params) return
  - Method2(params) return

**Modified Services:**
- ExistingService
  - NewMethod(params) return

**Business Rules:**
1. [Rule 1]
2. [Rule 2]

### Repository Changes

**New Repositories:**
- RepositoryName implementing domain.RepositoryInterface

**Modified Repositories:**
- ExistingRepository
  - NewMethod(params) return

---

## Test Plan

### Unit Tests

**Service Tests:**
- Test: TestServiceName_Method
  - Given: [conditions]
  - When: [action]
  - Then: [expected result]

### Integration Tests

**Repository Tests:**
- Test: TestRepository_Create
  - Verify: [database operation]

### API Tests

**Handler Tests:**
- Test: TestHandler_CreateResource
  - Request: POST /api/resource
  - Expect: 201 Created

### Budget Domain Tests

- Test: [Specific budget calculation]
  - Verify: [correct behavior]

---

## Implementation Checklist

### Domain Layer
- [ ] Create [Entity] in internal/domain/
- [ ] Define [Repository] interface in internal/domain/repository.go
- [ ] Add domain validation methods

### Application Layer
- [ ] Create/modify [Service] in internal/application/
- [ ] Implement business logic
- [ ] Add error handling

### Infrastructure Layer
- [ ] Implement [Repository] in internal/infrastructure/repository/
- [ ] Create [Handler] in internal/infrastructure/http/handlers/
- [ ] Update router in internal/infrastructure/http/router.go
- [ ] Update database schema in internal/infrastructure/database/sqlite.go

### Configuration
- [ ] Update cmd/server/main.go for dependency injection

### Testing
- [ ] Generate unit tests
- [ ] Generate integration tests
- [ ] Generate API tests
- [ ] Verify all tests pass

### Documentation
- [ ] Update Claude.md
- [ ] Add API documentation
- [ ] Update README if needed

---

## Acceptance Criteria

### Functional
- [ ] [Requirement 1] works as specified
- [ ] [Requirement 2] works as specified

### Non-Functional
- [ ] All tests passing
- [ ] Clean architecture maintained
- [ ] Security requirements met
- [ ] Performance acceptable
- [ ] Documentation complete

### Code Quality
- [ ] No architecture violations
- [ ] Go best practices followed
- [ ] Error handling comprehensive
- [ ] Input validation complete

---

## Risks and Mitigations

**Risk:** [Potential issue]
**Mitigation:** [How to address]

**Risk:** [Potential issue]
**Mitigation:** [How to address]

---

## Dependencies

- No external dependencies / Depends on [feature X]
- Requires [library] version Y

---

## Future Enhancements

[Features intentionally deferred for later]

---

## Approval

- [x] Domain Expert: Validated
- [x] Security Auditor: Reviewed
- [ ] Stakeholder: Approved (pending review)
- [ ] Developer: Ready to implement

---

## Implementation

**Ready for implementation with:**

```
/implement-spec docs/spec-<feature-name>.md
```

---

## Notes

[Any additional context, decisions made, alternatives considered]
```

**15. Save the specification**

Write the generated spec to:
- `docs/spec-<feature-name>.md`

---

### Phase 7: Review and Finalization

**16. Verify completeness**

Check that the spec includes:
- [ ] Clear business requirements
- [ ] Domain expert validation ✅
- [ ] Security review ✅
- [ ] Complete technical design
- [ ] API contracts defined
- [ ] Database changes specified
- [ ] Test plan created
- [ ] Acceptance criteria clear
- [ ] Implementation checklist

**17. Generate summary**

Provide a summary:

```markdown
# Specification Created: <Feature Name>

**Location:** docs/spec-<feature-name>.md

**Status:**
- Domain validation: ✅ Approved / ⚠️ Concerns addressed
- Security review: ✅ Approved / ⚠️ Requirements identified
- Technical design: ✅ Complete
- Ready for implementation: ✅ Yes

**Key Highlights:**
- [Highlight 1]
- [Highlight 2]

**Next Steps:**
1. Review the specification with stakeholders
2. Make any final adjustments
3. When approved, implement with:
   /implement-spec docs/spec-<feature-name>.md

**Estimated Implementation:**
- Complexity: Low / Medium / High
- Estimated effort: [X] hours/days
- Files to create/modify: ~[N] files
```

---

## Best Practices for Spec Creation

### Be Specific
- Define exact field names, types, constraints
- Specify exact endpoint paths and methods
- Document exact error messages and status codes

### Think Through Edge Cases
- What if the user provides invalid input?
- What if a related entity doesn't exist?
- What if there's a concurrent modification?
- What if the operation partially fails?

### Consider the Budget Domain
- How does this affect Ready to Assign?
- How does this interact with allocations?
- How does rollover work?
- What about credit cards?

### Design for Testability
- Clear inputs and outputs
- Deterministic behavior
- Mockable dependencies
- Observable side effects

### Follow Clean Architecture
- Domain entities are pure
- Services use interfaces
- Infrastructure implements interfaces
- Dependencies point inward

---

## Budget Application Specifics

### Always Consider

**Money Handling:**
- Amounts in INTEGER (cents)
- Never use REAL/float
- Conversions: dollars * 100 = cents

**Zero-Based Budgeting:**
- Ready to Assign = Balance - Allocated
- Available includes all history (rollover)
- One allocation per category per period

**Clean Architecture:**
- Domain: no external dependencies
- Application: interfaces only
- Infrastructure: implementations

**Security:**
- Parameterized SQL queries
- Input validation
- Error messages don't leak info

---

## Success Criteria for This Command

At the end, you should have:
- ✅ A complete specification document
- ✅ Domain expert validated business logic
- ✅ Security considerations identified
- ✅ Technical design ready to implement
- ✅ Test plan defined
- ✅ Acceptance criteria clear
- ✅ Ready for /implement-spec

---

**Remember:** The goal is to catch issues early, when they're cheap to fix. A good spec saves hours of refactoring later!
