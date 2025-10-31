---
description: Implement a specification by orchestrating agents and workflows
argument-hint: <spec-url>
allowed-tools: [Read, Write, Edit, Bash, Grep, Glob, WebFetch, Task]
---

# Implement Specification

Orchestrate a complete implementation workflow for a specification.

## Specification URL
{{arg}}

## Implementation Workflow

Follow these steps to implement the specification thoroughly:

### Phase 1: Fetch and Analyze Specification

1. **Fetch the specification**:
   - If URL is a GitHub file, use WebFetch or read it if already cloned
   - Parse and understand all requirements
   - Identify key components and changes needed

2. **Summarize findings**:
   - List all requirements
   - Identify affected components (entities, services, handlers, etc.)
   - Note any breaking changes
   - Highlight edge cases or concerns

### Phase 2: Domain Validation

3. **Invoke budget-domain-expert agent**:
   ```
   Invoke the budget-domain-expert agent to:
   - Validate business logic requirements
   - Verify zero-based budgeting calculations
   - Check for budget domain violations
   - Identify potential issues with:
     * Ready to Assign calculations
     * Category Available with rollover
     * Credit card logic
     * Allocation rules
   - Return validation report
   ```

4. **Review validation results**:
   - Show me what the domain expert found
   - Address any concerns before proceeding
   - Adjust plan if needed based on domain expert feedback

### Phase 3: Architecture Planning

5. **Determine implementation approach**:
   Based on spec requirements, identify:
   - [ ] New entities needed (domain layer)
   - [ ] Service changes (application layer)
   - [ ] Repository changes (infrastructure layer)
   - [ ] HTTP handler changes (infrastructure layer)
   - [ ] Database schema changes
   - [ ] API endpoint changes

6. **Check architecture compliance**:
   - Verify approach follows clean architecture
   - Domain layer: no external dependencies
   - Application layer: uses interfaces only
   - Infrastructure layer: implements interfaces
   - Dependencies point inward

### Phase 4: Implementation

7. **Implement changes systematically**:

   **If creating new feature:**
   - Follow `/new-feature` pattern
   - Create domain entity and repository interface
   - Implement service with business logic
   - Create repository implementation
   - Add HTTP handler
   - Register routes
   - Update database schema

   **If modifying existing feature:**
   - Read existing code first
   - Make minimal, focused changes
   - Preserve existing patterns
   - Update tests as needed

8. **Key implementation guidelines**:
   - Store amounts as INTEGER (cents)
   - Use parameterized SQL queries
   - Wrap errors with context
   - Follow existing code patterns
   - Maintain clean architecture layers

### Phase 5: Testing Strategy

9. **Invoke test-generator agent**:
   ```
   Invoke the test-generator agent to:
   - Create test plan for new/changed code
   - Generate unit tests for services
   - Generate integration tests for repositories
   - Generate handler tests for API endpoints
   - Focus on:
     * Happy path scenarios
     * Error cases
     * Edge cases
     * Budget calculation correctness
   - Return test implementation
   ```

10. **Review test coverage**:
    - Ensure critical paths are tested
    - Verify budget calculations are tested
    - Check error handling is tested

### Phase 6: Code Review

11. **Invoke code-reviewer agent**:
    ```
    Invoke the code-reviewer agent to:
    - Review all implementation changes
    - Check clean architecture compliance
    - Verify Go best practices
    - Validate budget domain logic
    - Check for:
      * Security issues
      * Error handling
      * Code quality
      * Test coverage
    - Return detailed review
    ```

12. **Address review findings**:
    - Fix any critical issues
    - Address important suggestions
    - Document decisions for suggestions not implemented

### Phase 7: Verification

13. **Run tests**:
    ```bash
    # Compile check
    go build ./...

    # Run all tests
    go test ./... -v

    # Check for issues
    go vet ./...
    ```

14. **Test API endpoints** (if applicable):
    - Use curl to test new/modified endpoints
    - Verify request/response formats
    - Test error cases
    - Validate status codes

15. **Manual verification**:
    - Test the feature manually if needed
    - Verify business logic works correctly
    - Check edge cases

### Phase 8: Documentation

16. **Update documentation**:
    - [ ] Update Claude.md with new entities/endpoints
    - [ ] Add comments to complex logic
    - [ ] Update API documentation
    - [ ] Document any breaking changes

### Phase 9: Final Check

17. **Architecture compliance check**:
    - Run `/check-architecture` to verify clean architecture
    - Ensure no layer violations introduced

18. **Security check** (if needed):
    - Invoke security-auditor agent for sensitive changes
    - Especially for:
      * Financial calculations
      * Database operations
      * User input handling

### Phase 10: Completion

19. **Summary**:
    - List all changes made
    - List all files created/modified
    - Confirm all requirements satisfied
    - Note any deviations from spec (with reasons)

20. **Next steps**:
    - Create commit with descriptive message
    - Push to branch
    - Ready for PR review

## Important Notes

### Budget Application Specific

- **Money Handling**: Always use INTEGER (cents), never REAL
- **Zero-Based Budgeting**:
  - Ready to Assign = Total Balance - Total Allocated
  - Available includes all history (rollover)
  - One allocation per category per period

- **Clean Architecture**:
  - Domain: entities + interfaces only
  - Application: business logic using interfaces
  - Infrastructure: implementations (DB, HTTP)

### Error Handling

- Wrap errors with context: `fmt.Errorf("context: %w", err)`
- Return appropriate HTTP status codes
- Validate all inputs

### Testing

- Unit tests for services (mock repositories)
- Integration tests for repositories (test DB)
- Handler tests for HTTP endpoints
- Test budget calculations thoroughly

## Execution Strategy

**For Simple Specs** (small changes):
- May skip some agent invocations
- Focus on implementation and basic testing

**For Complex Specs** (major changes):
- Use all agents as specified
- Thorough testing and validation
- Multiple review passes

**For Breaking Changes**:
- Extra caution with domain expert
- Extensive testing
- Migration strategy planning

## Success Criteria

- [ ] All spec requirements implemented
- [ ] Domain expert validated business logic
- [ ] Tests passing
- [ ] Code reviewer approved
- [ ] Architecture compliant
- [ ] Documentation updated
- [ ] No regressions introduced

## Final Checklist

Before marking complete:
- [ ] Code compiles without errors
- [ ] All tests pass
- [ ] Architecture is clean
- [ ] Security is verified
- [ ] Documentation is updated
- [ ] Ready for PR

---

**Remember**: This is a thorough workflow. Adapt based on spec complexity. For simple changes, streamline. For complex changes, follow rigorously.
