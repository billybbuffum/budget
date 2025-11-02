# UI Test Execution Guide: Allocate to Cover Feature

## Purpose

This guide provides step-by-step instructions for executing interactive UI tests for the "Allocate to Cover" button feature using Playwright MCP.

## Important Note About Playwright MCP

**Playwright MCP (Model Context Protocol)** is an interactive testing tool that:
- Actually opens a real browser
- Performs user actions (clicks, typing, navigation)
- Reports results in real-time
- Can be used during development to verify features work correctly

This is different from automated test files (`.spec.ts`) which run in CI/CD.

## Prerequisites

### 1. Start the Application

```bash
# Using Docker Compose (recommended)
docker-compose up -d

# OR using Go directly
go run cmd/server/main.go
```

Verify application is running: http://localhost:8080

### 2. Verify Application State

Open browser to http://localhost:8080 and verify:
- Application loads
- Budget tab is visible
- No JavaScript errors in console

### 3. Prepare Test Environment

Option A: Use existing database (may have existing data)
Option B: Reset database for clean test state

```bash
# To reset database (if needed)
docker-compose down -v
docker-compose up -d
```

## Test Execution with Playwright MCP

### How to Use Playwright MCP

Playwright MCP is invoked through your AI assistant. You provide natural language instructions for what to test, and Playwright MCP executes the actions in a real browser.

**Example command format:**
```
Use Playwright MCP to test the following:

1. Navigate to http://localhost:8080
2. Click the Budget tab
3. Verify "Ready to Assign" is displayed
4. Take screenshot
```

### Test Scenario 1: Basic Underfunded Display

**Objective:** Verify underfunded warning appears correctly

**Test Setup Commands for Playwright MCP:**

```
Use Playwright MCP to test underfunded warning display:

1. Navigate to http://localhost:8080

2. Create test data via API:
   - POST /api/accounts: Create checking account "Test Checking"
   - POST /api/categories: Create income category "Salary"
   - POST /api/transactions: Add $5,000 income to checking account
   - POST /api/accounts: Create credit card account "Test CC"
   - POST /api/categories: Create expense category "Groceries"
   - POST /api/transactions: Add -$200 expense to credit card (Groceries)
   - GET /api/categories: Find payment category (auto-created, name contains "Payment")
   - POST /api/allocations: Allocate $100 to payment category for current period

3. Navigate to Budget tab

4. Locate the payment category row (look for orange colored category with "Payment" in name)

5. Verify the following elements are visible:
   - Warning emoji: ⚠️
   - Text: "Underfunded - Need $100.00 more"
   - Warning text color is red
   - Contributing categories text: "Contributing categories: Groceries"

6. Take screenshot: 'underfunded-warning.png'

7. Report results:
   - Is underfunded warning visible? (Yes/No)
   - Is amount formatted correctly as $100.00? (Yes/No)
   - Is text color red? (Yes/No)
   - Are contributing categories listed? (Yes/No)
```

**Expected Result:**
- All checks pass
- Screenshot shows red warning with correct amount
- Contributing categories displayed

---

### Test Scenario 2: Allocate to Cover Button Appearance

**Test Commands for Playwright MCP:**

```
Use Playwright MCP to test "Allocate to Cover" button:

1. Continue from previous test scenario (underfunded state exists)

2. Locate "Allocate to Cover" button in payment category row

3. Verify button properties:
   - Text: "Allocate to Cover"
   - Background color: Blue (class includes 'bg-blue-600')
   - Text color: White
   - Button is enabled (not disabled)
   - Title attribute contains: "Allocate from Ready to Assign"

4. Hover over button

5. Verify hover state:
   - Background color darkens (bg-blue-700)

6. Take screenshot of button in normal state

7. Take screenshot of button in hover state

8. Report results
```

**Expected Result:**
- Button visible with correct styling
- Hover effect works
- Button is clickable

---

### Test Scenario 3: Successful Allocation (Happy Path)

**Test Commands for Playwright MCP:**

```
Use Playwright MCP to test successful allocation workflow:

1. Continue from previous scenario (underfunded payment category exists)

2. Before clicking button, record current values:
   - Ready to Assign amount (e.g., $4,900.00)
   - Payment category allocated amount (e.g., $100.00)
   - Underfunded amount (e.g., $100.00)

3. Click "Allocate to Cover" button

4. Immediately observe:
   - Button text changes to "Allocating..."
   - Button becomes disabled
   - Take screenshot: 'button-loading-state.png'

5. Wait for API response (POST /api/allocations/cover-underfunded)

6. After API response, verify:
   - Success toast appears with message: "Successfully allocated..."
   - Toast color is green
   - Take screenshot: 'success-toast.png'

7. Wait for UI to refresh (2-3 seconds)

8. Verify final state:
   - Underfunded warning is NO LONGER visible
   - Payment category allocated increased by $100 (now $200.00)
   - Ready to Assign decreased by $100 (now $4,800.00)
   - "Allocate to Cover" button is NO LONGER visible
   - Take screenshot: 'after-allocation-success.png'

9. Report results with before/after values
```

**Expected Result:**
- Loading state shown during API call
- Success toast appears
- Underfunded warning removed
- Allocations and RTA updated correctly
- No JavaScript errors in console

---

### Test Scenario 4: Insufficient Funds Error

**Test Commands for Playwright MCP:**

```
Use Playwright MCP to test insufficient funds error handling:

1. Reset to clean state (refresh page or recreate test data)

2. Create scenario with insufficient funds:
   - POST /api/accounts: Create checking account with $200 income
   - POST /api/accounts: Create credit card with -$500 debt
   - POST /api/allocations: Allocate $100 to another category (reduces RTA)
   - Result: RTA = $100, Underfunded = $500

3. Navigate to Budget tab

4. Verify Ready to Assign shows $100.00

5. Verify payment category shows underfunded $500.00

6. Click "Allocate to Cover" button

7. Wait for API response (expect 400 error)

8. Verify error handling:
   - Error toast appears (red color)
   - Error message contains: "Insufficient funds"
   - Error message shows RTA amount: "$100.00"
   - Error message shows underfunded amount: "$500.00"
   - Take screenshot: 'insufficient-funds-error.png'

9. Verify button recovery:
   - Button text reverts to "Allocate to Cover"
   - Button becomes enabled again

10. Verify no changes made:
    - Underfunded warning still visible
    - Ready to Assign unchanged ($100.00)
    - Payment category allocation unchanged

11. Report results
```

**Expected Result:**
- Clear error message explaining why operation failed
- Button recovers to normal state
- No data changed
- User understands the problem

---

### Test Scenario 5: Multiple Underfunded Categories

**Test Commands for Playwright MCP:**

```
Use Playwright MCP to test multiple underfunded payment categories:

1. Create test data with TWO credit cards:
   - Create checking account with $10,000 income
   - Create credit card 1 with -$300 debt
   - Create credit card 2 with -$400 debt
   - Allocate $100 to CC1 payment category (underfunded $200)
   - Allocate $150 to CC2 payment category (underfunded $250)

2. Navigate to Budget tab

3. Verify TWO underfunded warnings are visible:
   - First payment category: "Need $200.00 more"
   - Second payment category: "Need $250.00 more"

4. Verify TWO "Allocate to Cover" buttons visible

5. Note Ready to Assign: $9,750 ($10,000 - $100 - $150)

6. Click first "Allocate to Cover" button

7. Wait for success

8. Verify first payment category:
   - Underfunded warning removed
   - Allocated increased to $300

9. Verify second payment category:
   - Still shows underfunded warning
   - Unchanged

10. Verify Ready to Assign: $9,550 ($9,750 - $200)

11. Click second "Allocate to Cover" button

12. Wait for success

13. Verify second payment category:
    - Underfunded warning removed
    - Allocated increased to $400

14. Verify Ready to Assign: $9,300 ($9,550 - $250)

15. Verify NO underfunded warnings visible

16. Take screenshots at each stage

17. Report results
```

**Expected Result:**
- Both categories work independently
- Each button covers its own underfunded amount
- RTA decreases by total: $10,000 - $250 - $450 = $9,300
- No interference between categories

---

### Test Scenario 6: Contributing Categories Display

**Test Commands for Playwright MCP:**

```
Use Playwright MCP to test contributing categories display:

1. Create credit card with multiple spending categories:
   - Create checking with $5,000 income
   - Create credit card account
   - Create expense on credit card: -$150 Groceries
   - Create expense on credit card: -$100 Gas
   - Create expense on credit card: -$125 Dining
   - Total debt: -$375
   - Allocate $100 to payment category (underfunded $275)

2. Navigate to Budget tab

3. Locate payment category row

4. Verify underfunded warning shows: "$275.00"

5. Verify contributing categories text:
   - Text: "Contributing categories:"
   - Lists: "Groceries, Gas, Dining" (comma-separated)
   - Text color: Red
   - Font size: Small (text-xs class)

6. Take screenshot: 'contributing-categories.png'

7. Report results
```

**Expected Result:**
- Contributing categories listed correctly
- Comma-separated format
- Red text color
- Helpful context for user

---

### Test Scenario 7: No Button for Regular Categories

**Test Commands for Playwright MCP:**

```
Use Playwright MCP to verify button only appears for payment categories:

1. Create regular expense category:
   - POST /api/categories: Create "Groceries" expense category
   - POST /api/allocations: Allocate $500 to Groceries

2. Navigate to Budget tab

3. Locate Groceries category row

4. Verify "Allocate to Cover" button does NOT appear

5. Verify only payment categories have the button

6. Take screenshot: 'no-button-regular-category.png'

7. Report results
```

**Expected Result:**
- "Allocate to Cover" button ONLY on payment categories
- Regular categories show standard UI

---

### Test Scenario 8: Fully Funded Payment Category

**Test Commands for Playwright MCP:**

```
Use Playwright MCP to verify no warning when fully funded:

1. Create credit card with matching allocation:
   - Create credit card with -$300 debt
   - Allocate exactly $300 to payment category

2. Navigate to Budget tab

3. Verify payment category shows:
   - Allocated: $300.00
   - Available: $300.00
   - NO underfunded warning
   - NO "Allocate to Cover" button

4. Take screenshot: 'fully-funded-no-warning.png'

5. Report results
```

**Expected Result:**
- No warning when payment category available >= credit card debt
- No button appears
- Normal category display

---

### Test Scenario 9: Ready to Assign Accounting

**Test Commands for Playwright MCP:**

```
Use Playwright MCP to verify RTA formula accounts for underfunded:

Formula to verify: RTA = Total Inflows - Total Allocations - Total Underfunded

1. Create controlled scenario:
   - Income: $5,000
   - Regular category allocation: $1,500
   - Credit card debt: $500
   - Payment category allocation: $300
   - Underfunded: $200

2. Calculate expected RTA:
   RTA = $5,000 - $1,500 - $200 = $3,300

3. Navigate to Budget tab

4. Verify Ready to Assign displays: $3,300.00

5. Click "Allocate to Cover" to cover $200 underfunded

6. Wait for success

7. Calculate new expected RTA:
   - Total allocations now: $1,500 + $200 = $1,700
   - Underfunded now: $0
   - RTA = $5,000 - $1,700 - $0 = $3,300

8. Verify Ready to Assign UNCHANGED: $3,300.00
   (Because underfunded moved to allocation, net effect is zero)

9. Take screenshots showing RTA before and after

10. Report results with calculations
```

**Expected Result:**
- RTA correctly accounts for underfunded before allocation
- RTA unchanged after covering (underfunded → allocation)
- Zero-based budgeting principle maintained

---

## Interpreting Test Results

### Success Indicators

For each test scenario, check:
- ✅ All expected elements visible
- ✅ Correct styling and formatting
- ✅ API calls successful (200/201 status codes)
- ✅ No JavaScript errors in console
- ✅ UI updates automatically
- ✅ User feedback is clear

### Failure Indicators

- ❌ Elements not visible or missing
- ❌ Incorrect calculations
- ❌ API errors (400, 500 status codes)
- ❌ JavaScript console errors
- ❌ UI doesn't refresh
- ❌ Poor user experience

### Bug Report Template

When a test fails, document:

```
Bug ID: BUG-ALLOCATE-XXX
Test Scenario: [Test name]
Severity: Critical / High / Medium / Low

Steps to Reproduce:
1. [Step]
2. [Step]
3. [Step]

Expected Result:
[What should happen]

Actual Result:
[What actually happened]

Screenshots:
- [Screenshot filename]

Console Errors:
[JavaScript errors from browser console]

Network Errors:
[API errors from Network tab]

Additional Context:
[Any other relevant information]
```

## After Testing is Complete

### If All Tests Pass ✅

1. **Document Success:**
   - All test scenarios executed successfully
   - Screenshots captured
   - No bugs found
   - Feature ready for production

2. **Generate Automated Tests:**
   - Use the Playwright test file: `/Users/billybuffum/development/budget/tests/e2e/allocate-to-cover.spec.ts`
   - Run: `npx playwright test allocate-to-cover.spec.ts`
   - Integrate into CI/CD pipeline

3. **Update Documentation:**
   - Mark feature as tested and approved
   - Update user documentation
   - Add to release notes

### If Tests Fail ❌

1. **Document Bugs:**
   - Create detailed bug reports
   - Include screenshots and reproduction steps
   - Create GitHub issues

2. **Prioritize Fixes:**
   - Critical: Blocks feature from working
   - High: Major functionality broken
   - Medium: Edge cases or UX issues
   - Low: Minor cosmetic issues

3. **Fix and Retest:**
   - Fix bugs in priority order
   - Retest ONLY the failed scenarios
   - Once fixed, run full test suite again

4. **Iterate:**
   - Repeat until all tests pass
   - Get approval before marking complete

## Tips for Effective Testing

1. **Test in Order:**
   - Start with basic scenarios (underfunded display)
   - Progress to complex scenarios (multiple categories)
   - This helps isolate problems

2. **Take Screenshots:**
   - Visual evidence is valuable
   - Compare before/after states
   - Share with team

3. **Check Console:**
   - Always keep DevTools console open
   - JavaScript errors indicate bugs
   - Network tab shows API calls

4. **Fresh Database:**
   - Use clean state for each test run
   - Avoids data contamination
   - Makes tests reproducible

5. **Real User Perspective:**
   - Does the feature make sense?
   - Is it intuitive?
   - Are error messages helpful?

## Browser DevTools Checklist

Open DevTools (F12) and monitor:

- **Console Tab:**
  - No errors (red messages)
  - No warnings (yellow messages)
  - Log API calls (if debug mode on)

- **Network Tab:**
  - Filter: XHR to see API calls
  - Check status codes (200, 201, 400, etc.)
  - Verify request/response payloads
  - Check response times

- **Elements Tab:**
  - Inspect element classes
  - Verify colors (red, blue)
  - Check visibility (display: none?)

## Summary Report Template

After completing all tests:

```
# UI Test Results: Allocate to Cover Feature
Date: [Date]
Tester: [Name]
Browser: Chrome [version]
Application: http://localhost:8080

## Test Execution Summary
- Total Scenarios: 9
- Passed: X
- Failed: Y
- Blocked: Z

## Passed Tests
✅ TC-001: Underfunded warning display
✅ TC-002: Button appearance
[...]

## Failed Tests
❌ TC-XXX: [Test name]
- Reason: [Brief description]
- Bug ID: BUG-ALLOCATE-XXX
[...]

## Bugs Found
1. [Bug summary]
   - Severity: High
   - Status: Open
   - Issue: #XXX

## Overall Assessment
- Feature Status: Ready / Needs Work
- Recommendation: Approve / Fix Bugs First
- Notes: [Additional comments]

## Screenshots
[List of screenshot files]

## Next Steps
1. [Action item]
2. [Action item]
```

## Questions or Issues?

If you encounter problems during testing:
1. Check application logs: `docker-compose logs -f`
2. Verify database state: `sqlite3 budget.db`
3. Restart application: `docker-compose restart`
4. Reset database: `docker-compose down -v && docker-compose up -d`

## Conclusion

Interactive testing with Playwright MCP is a powerful way to verify UI functionality before creating automated tests. Take your time, be thorough, and document everything!
