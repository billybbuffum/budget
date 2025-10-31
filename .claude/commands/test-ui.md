---
description: Test UI interactively with Playwright MCP and optionally generate automated tests
argument-hint: <feature-or-workflow-description>
allowed-tools: [Task, Bash, Write]
---

# Test UI with Playwright

Test user interface interactively using Playwright MCP, find bugs, and optionally generate automated test code.

## Feature/Workflow to Test
{{arg}}

## UI Testing Workflow

### Phase 1: Understand What to Test

Based on the feature description, determine:
- Which pages/components to test
- What user workflow to verify
- Expected behavior and outcomes
- Success criteria

### Phase 2: Invoke UI Tester Agent

**Invoke the ui-tester agent to perform interactive testing:**

```
Invoke the ui-tester agent with task:
"Test the {{arg}} workflow interactively using Playwright MCP.

Create a detailed test plan with specific steps, then execute using Playwright MCP.

Report:
1. Test plan created
2. Interactive test results (passed/failed for each step)
3. Bugs found (if any)
4. Screenshots/evidence (if available)
5. Recommendations for fixes
6. Whether ready for automated test generation

If tests pass completely, generate Playwright test files for CI/CD."
```

The ui-tester agent will:
1. Create test plan
2. Use Playwright MCP to actually test in browser
3. Report detailed results
4. Identify bugs if found
5. Generate automated tests if all passes

### Phase 3: Review Results

**If tests passed:**
```markdown
✅ All UI tests passed!

Tests Created:
- tests/e2e/[feature].spec.ts

Next Steps:
- Run tests: npx playwright test
- Add to CI/CD pipeline
- Feature is ready for deployment
```

**If bugs found:**
```markdown
❌ Bugs found during testing

Issues:
1. [Bug description]
   - Steps to reproduce
   - Expected vs actual behavior
   - Suggested fix

Next Steps:
1. Fix the identified bugs
2. Rerun: /test-ui "{{arg}}"
3. Repeat until all tests pass
```

### Phase 4: Fix Bugs (if needed)

If bugs were found:
1. Review bug reports from ui-tester agent
2. Fix the issues in the code
3. Re-run this command:
   ```
   /test-ui "{{arg}}"
   ```
4. Repeat until all tests pass

### Phase 5: Generate Automated Tests

Once interactive tests pass, the agent will automatically generate:
- Playwright test files (`.spec.ts`)
- Test configuration
- README with test running instructions

## Common UI Test Scenarios

### Test Allocation Creation
```
/test-ui "allocation creation workflow"
```

Tests:
- Navigate to Budget page
- Click Add Allocation button
- Fill in category and amount
- Save allocation
- Verify allocation appears
- Verify Ready to Assign updates

### Test Transaction Entry
```
/test-ui "transaction creation and balance update"
```

Tests:
- Navigate to Transactions page
- Add new transaction
- Verify account balance updates
- Verify transaction appears in list

### Test Budget Calculations
```
/test-ui "Ready to Assign calculation with income and allocations"
```

Tests:
- Record initial Ready to Assign
- Add income transaction
- Verify Ready to Assign increases
- Create allocation
- Verify Ready to Assign decreases
- Verify math is correct

### Test Form Validation
```
/test-ui "allocation form validation and error handling"
```

Tests:
- Try to submit empty form
- Try to submit with invalid data
- Verify error messages appear
- Verify form prevents invalid submission

### Test Responsive Design
```
/test-ui "allocation workflow on mobile, tablet, and desktop"
```

Tests:
- Test at 375px width (mobile)
- Test at 768px width (tablet)
- Test at 1920px width (desktop)
- Verify layout adapts
- Verify all functions work

### Test Credit Card Workflow
```
/test-ui "credit card creation and payment category auto-creation"
```

Tests:
- Create credit card account
- Verify payment category appears
- Spend on credit card
- Verify payment category available increases

## Requirements

### Playwright MCP Must Be Installed

Check if installed:
```bash
claude mcp list | grep playwright
```

If not installed:
```bash
npm install -g @playwright/test @modelcontextprotocol/server-playwright
claude mcp add playwright --scope user
```

See `MCP_RECOMMENDATIONS.md` for full installation instructions.

### Application Must Be Running

Ensure the Budget app is running:
```bash
# Check if running
curl http://localhost:8080/health

# If not running, start it
docker-compose up -d

# Or run directly
go run cmd/server/main.go
```

## Understanding the Output

### Successful Test Output

```markdown
# UI Test Results: Allocation Creation

## Test Plan
1. Navigate to http://localhost:8080
2. Click Budget tab
3. Click Add Allocation
4. Select "Groceries" category
5. Enter $500 amount
6. Click Save
7. Verify success message
8. Verify allocation appears

## Interactive Test Results (Playwright MCP)
✅ Step 1: Navigated successfully
✅ Step 2: Budget tab clicked
✅ Step 3: Add Allocation modal opened
✅ Step 4: Category selected: Groceries
✅ Step 5: Amount entered: $500
✅ Step 6: Save button clicked
✅ Step 7: Success message appeared
✅ Step 8: Allocation row created with correct values

## Summary
✅ All tests passed
✅ No bugs found
✅ UI behaves as expected

## Automated Tests Generated
Created: tests/e2e/allocations.spec.ts

Run with: npx playwright test allocations.spec.ts
```

### Failed Test Output

```markdown
# UI Test Results: Allocation Creation

## Test Plan
[Same as above]

## Interactive Test Results (Playwright MCP)
✅ Step 1-5: Passed
❌ Step 6: Save button clicked - NO RESPONSE
⏭️  Step 7-8: Skipped (previous step failed)

## Bug Found
**Issue**: Save button does not respond to clicks

**Evidence**:
- Button clicked via Playwright MCP
- No network request triggered
- No success message appeared
- Console error: "Cannot read property 'save' of undefined"

**Location**: Budget page, Add Allocation modal

**Suggested Fix**:
Check the save button event handler in app.js.
The save function may not be properly bound.

## Next Steps
1. Fix the save button event handler
2. Rerun: /test-ui "allocation creation workflow"
3. Verify fix resolves the issue
```

## Tips for Effective UI Testing

### 1. Start the App First
Always ensure the app is running before testing:
```bash
docker-compose up -d && sleep 2  # Wait for app to start
```

### 2. Test Specific Workflows
Be specific about what to test:
```
Good: /test-ui "allocation creation with valid data"
Better: /test-ui "create allocation, verify Ready to Assign decreases"
```

### 3. Test Error Cases Too
Don't just test happy path:
```
/test-ui "allocation creation with missing required fields"
/test-ui "allocation creation with invalid amount"
```

### 4. Test Calculations
Budget app is all about calculations:
```
/test-ui "verify rollover calculation from previous month"
/test-ui "verify Ready to Assign after income and allocation"
```

### 5. Retest After Fixes
After fixing bugs:
```
/test-ui "[same workflow that failed]"
```

## Integration with Development Workflow

### During Feature Development

```
# 1. Implement feature
/implement-spec docs/spec-feature.md

# 2. Test UI interactively
/test-ui "feature workflow"

# 3. Fix any bugs found
[Make fixes]

# 4. Retest
/test-ui "feature workflow"

# 5. Tests pass → automated tests generated
# 6. Commit feature + tests
```

### Before Creating PR

```
# Test all major workflows
/test-ui "allocation creation and management"
/test-ui "transaction entry and balance updates"
/test-ui "budget calculations and rollover"

# Ensure all tests pass before PR
```

### After UI Changes

```
# Any time UI changes, retest affected workflows
/test-ui "[affected workflow]"
```

## Generated Test Files

After successful testing, you'll have:

**tests/e2e/[feature].spec.ts**
- Automated Playwright tests
- Can run in CI/CD
- Regression testing

**How to run generated tests:**
```bash
# Run all tests
npx playwright test

# Run specific test
npx playwright test allocations.spec.ts

# Run with visible browser
npx playwright test --headed

# Debug mode
npx playwright test --debug
```

## Troubleshooting

### Playwright MCP Not Found
```bash
# Install Playwright MCP
npm install -g @playwright/test @modelcontextprotocol/server-playwright
claude mcp add playwright --scope user

# Verify installation
claude mcp list
```

### App Not Responding
```bash
# Check if app is running
curl http://localhost:8080/health

# Restart app
docker-compose restart
```

### Tests Timing Out
```
Increase timeout in test:
- Slow network
- Heavy page load
- Complex calculations

Agent will adjust timeouts automatically.
```

### Browser Not Opening
```
Playwright MCP should open browser automatically.
If not:
- Check Playwright is installed
- Check MCP configuration
- Try: npx playwright install
```

## Success Criteria

This command succeeds when:
- ✅ Test plan created
- ✅ Interactive tests run via Playwright MCP
- ✅ All tests pass OR bugs clearly documented
- ✅ Automated tests generated (if tests passed)
- ✅ Next steps clearly identified

## Remember

**Playwright MCP = Interactive Testing**
- Opens real browser
- Tests like a real user
- Finds bugs fast
- Use during development

**Generated Tests = Automation**
- Runs in CI/CD
- Regression testing
- No browser needed
- Use for deployment

**Workflow:**
1. Test interactively with MCP → Find bugs
2. Fix bugs → Retest
3. All pass → Generate automated tests
4. Commit tests → CI/CD runs them forever
