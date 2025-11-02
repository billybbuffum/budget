# UI Test Plan: Allocate to Cover Underfunded Payment Categories

**Feature:** Credit Card Payment Category "Allocate to Cover" Button
**Specification:** /Users/billybuffum/development/budget/docs/spec-remove-cc-sync.md
**Date Created:** 2025-11-01
**Status:** Ready for Interactive Testing

## Overview

This test plan covers interactive UI testing of the "Allocate to Cover" button feature for underfunded credit card payment categories. This feature allows users to manually allocate funds from Ready to Assign to cover underfunded payment categories with a single click.

## Test Environment

- **Application URL:** http://localhost:8080
- **Browser:** Chrome (Playwright default)
- **Test Data:** Fresh database with test data created during test execution

## Prerequisites

Before executing tests, ensure:
1. Application is running: `docker-compose up -d` or `go run cmd/server/main.go`
2. Database is accessible
3. Frontend is built: `npm run build:css`

## Test Data Setup

### Initial State Required:
1. **Checking Account**
   - Name: "Test Checking"
   - Type: checking
   - Balance: $5,000.00 (via income transaction)

2. **Credit Card Account**
   - Name: "Test Credit Card"
   - Type: checking (with negative balance)
   - Balance: -$500.00 (debt)

3. **Categories**
   - Income category: "Salary"
   - Expense categories: "Groceries", "Gas", "Dining"
   - Payment category: Auto-created for credit card

4. **Transactions**
   - Income: +$5,000 to Checking, category: Salary
   - CC Expense: -$200 on Credit Card, category: Groceries
   - CC Expense: -$150 on Credit Card, category: Gas
   - CC Expense: -$150 on Credit Card, category: Dining
   - Total CC debt: $500

5. **Initial Allocations**
   - Payment category allocation: $300 (creates underfunded state)
   - Underfunded amount: $500 - $300 = $200

## Test Scenarios

---

### Test 1: Underfunded Warning Display

**Objective:** Verify underfunded payment categories display correct warnings

**Test ID:** TC-ALLOCATE-001

**Steps:**
1. Navigate to http://localhost:8080
2. Click "Budget" tab
3. Select current period (2025-11)
4. Scroll to payment category row (orange colored, name ends with " - Payment")
5. Locate underfunded warning section
6. Verify warning text: "⚠️ Underfunded - Need $200.00 more"
7. Verify contributing categories text: "Contributing categories: Groceries, Gas, Dining"
8. Verify warning color is red (text-red-600)
9. Take screenshot: `underfunded-warning-display.png`

**Expected Results:**
- ✅ Warning visible with warning emoji (⚠️)
- ✅ Amount formatted as currency: "$200.00"
- ✅ Text color is red
- ✅ Contributing categories listed: "Groceries, Gas, Dining"
- ✅ Warning appears in payment category row only

**Pass Criteria:**
- All expected results verified
- Screenshot shows correct formatting

---

### Test 2: "Allocate to Cover" Button Appearance

**Objective:** Verify button appears and displays correctly for underfunded categories

**Test ID:** TC-ALLOCATE-002

**Steps:**
1. Navigate to Budget page with underfunded payment category
2. Locate "Allocate to Cover" button in payment category row
3. Verify button text: "Allocate to Cover"
4. Verify button styling:
   - Background: Blue (bg-blue-600)
   - Text: White
   - Rounded corners
   - Padding: px-3 py-1
5. Hover over button
6. Verify hover state: Darker blue (bg-blue-700)
7. Check button title attribute: "Allocate from Ready to Assign to cover this underfunded amount"
8. Verify button is enabled (not disabled)
9. Take screenshot: `allocate-button-appearance.png`

**Expected Results:**
- ✅ Button visible next to underfunded warning
- ✅ Button text: "Allocate to Cover"
- ✅ Blue background with white text
- ✅ Hover changes to darker blue
- ✅ Tooltip explains functionality
- ✅ Button is enabled and clickable

**Pass Criteria:**
- Button styling matches design
- Hover state works
- Tooltip is informative

---

### Test 3: Successful Allocation Workflow (Happy Path)

**Objective:** Test complete workflow from underfunded to covered state

**Test ID:** TC-ALLOCATE-003

**Pre-conditions:**
- Credit card payment category underfunded: $200
- Ready to Assign: $4,500 (sufficient funds)

**Steps:**
1. Navigate to Budget page
2. Note initial values:
   - Ready to Assign: $4,500.00
   - Payment category allocated: $300.00
   - Payment category underfunded: $200.00
3. Click "Allocate to Cover" button
4. Observe button state changes:
   - Text changes to "Allocating..."
   - Button becomes disabled
5. Wait for API response (POST /api/allocations/cover-underfunded)
6. Verify success toast appears:
   - Message: "Successfully allocated $200.00 to cover [Payment Category Name]"
   - Toast color: Green (success)
7. Verify button state after success:
   - Text reverts to "Allocate to Cover"
   - Button re-enabled
8. Verify underfunded warning disappears (no longer shown)
9. Verify payment category allocation updated:
   - New allocated: $500.00 ($300 + $200)
10. Verify Ready to Assign updated:
    - New RTA: $4,300.00 ($4,500 - $200)
11. Verify allocation summary refreshes automatically
12. Take screenshots:
    - `before-allocation.png`
    - `during-allocation-loading.png`
    - `after-allocation-success.png`

**Expected Results:**
- ✅ Button shows loading state during API call
- ✅ Success toast displays with correct amount
- ✅ Underfunded warning removed from UI
- ✅ Payment category allocated = $300 + $200 = $500
- ✅ Ready to Assign = $4,500 - $200 = $4,300
- ✅ UI refreshes without manual page reload

**Pass Criteria:**
- All calculations correct
- UI updates automatically
- No errors in console
- User receives clear success feedback

---

### Test 4: Insufficient Funds Error Handling

**Objective:** Verify correct error handling when RTA < underfunded amount

**Test ID:** TC-ALLOCATE-004

**Pre-conditions:**
- Create scenario with:
  - Credit card debt: $500
  - Payment category allocated: $0
  - Underfunded: $500
  - Modify Ready to Assign to $100 (create allocations to other categories)

**Steps:**
1. Navigate to Budget page
2. Verify Ready to Assign: $100.00
3. Verify payment category underfunded: $500.00
4. Note that underfunded > RTA
5. Click "Allocate to Cover" button
6. Observe button state:
   - Changes to "Allocating..."
   - Becomes disabled
7. Wait for API error response
8. Verify error toast appears:
   - Message: "Insufficient funds: Ready to Assign: $100.00, Underfunded: $500.00"
   - Toast color: Red (error)
   - Toast duration: Visible for 5+ seconds
9. Verify button state after error:
   - Text reverts to "Allocate to Cover"
   - Button re-enabled
10. Verify underfunded warning still visible
11. Verify no allocation created:
    - Payment category allocated unchanged
    - Ready to Assign unchanged
12. Check browser console for errors (should be logged but handled gracefully)
13. Take screenshot: `insufficient-funds-error.png`

**Expected Results:**
- ✅ Error toast displays clear message with amounts
- ✅ Button returns to normal state after error
- ✅ Underfunded warning remains visible
- ✅ No changes to allocations
- ✅ Ready to Assign unchanged
- ✅ User understands why operation failed

**Pass Criteria:**
- Error message is clear and actionable
- No data corruption
- Button recovers correctly
- Console error logged (for debugging)

---

### Test 5: Multiple Underfunded Categories

**Objective:** Verify handling of multiple underfunded payment categories

**Test ID:** TC-ALLOCATE-005

**Pre-conditions:**
- Create two credit card accounts:
  - Credit Card 1: -$300 debt, $100 allocated, $200 underfunded
  - Credit Card 2: -$400 debt, $150 allocated, $250 underfunded
- Ready to Assign: $3,000 (sufficient for both)

**Steps:**
1. Navigate to Budget page
2. Verify both payment categories show underfunded warnings:
   - Payment 1: "⚠️ Underfunded - Need $200.00 more"
   - Payment 2: "⚠️ Underfunded - Need $250.00 more"
3. Verify both have "Allocate to Cover" buttons
4. Note Ready to Assign: $3,000.00
5. Click "Allocate to Cover" for Payment Category 1
6. Wait for success
7. Verify Payment Category 1:
   - Underfunded warning removed
   - Allocated increased by $200
8. Verify Payment Category 2:
   - Still shows underfunded warning
   - Unchanged
9. Verify Ready to Assign: $2,800.00 ($3,000 - $200)
10. Click "Allocate to Cover" for Payment Category 2
11. Wait for success
12. Verify Payment Category 2:
    - Underfunded warning removed
    - Allocated increased by $250
13. Verify Ready to Assign: $2,550.00 ($2,800 - $250)
14. Take screenshots at each stage

**Expected Results:**
- ✅ Each payment category displays independently
- ✅ Buttons work independently
- ✅ Covering one doesn't affect the other
- ✅ RTA decreases by sum of covered amounts: $3,000 - $200 - $250 = $2,550
- ✅ Both categories can be covered sequentially

**Pass Criteria:**
- Independent operation of each button
- Correct RTA calculation after each operation
- No interference between categories

---

### Test 6: Underfunded Calculation Accuracy

**Objective:** Verify underfunded calculation matches formula

**Test ID:** TC-ALLOCATE-006

**Formula:** Underfunded = |Credit Card Balance| - Payment Category Available

**Steps:**
1. Create credit card account
2. Create transactions totaling -$500 (debt)
3. Verify credit card balance: -$500.00
4. Create allocation to payment category: $300.00
5. Verify payment category available: $300.00
6. Calculate expected underfunded:
   - Underfunded = abs(-500) - 300 = 500 - 300 = 200
7. Navigate to Budget page
8. Verify displayed underfunded: "$200.00"
9. Verify calculation matches
10. Test with different amounts:
    - CC Balance: -$1,000, Allocated: $750, Expected: $250
    - CC Balance: -$250, Allocated: $0, Expected: $250
    - CC Balance: -$100, Allocated: $150, Expected: $0 (not underfunded)
11. Take screenshots showing calculations

**Expected Results:**
- ✅ Underfunded = |CC Balance| - Payment Available
- ✅ Display matches calculated value exactly
- ✅ Formula holds for various scenarios
- ✅ When Available >= Balance, no underfunded warning shown

**Pass Criteria:**
- All calculations match formula
- Edge cases handled correctly
- No rounding errors

---

### Test 7: Contributing Categories Display

**Objective:** Verify contributing categories are listed correctly

**Test ID:** TC-ALLOCATE-007

**Steps:**
1. Create credit card with transactions in multiple categories:
   - Groceries: -$150
   - Gas: -$100
   - Dining: -$125
   - Total debt: -$375
2. Create payment category allocation: $100 (underfunded by $275)
3. Navigate to Budget page
4. Locate underfunded warning
5. Verify contributing categories text appears
6. Verify text format: "Contributing categories: Groceries, Gas, Dining"
7. Verify categories are comma-separated
8. Verify text color: Red (text-red-500 or text-red-400)
9. Verify text size: Small (text-xs)
10. Test with single category:
    - Only Groceries: -$200
    - Verify: "Contributing categories: Groceries"
11. Take screenshot: `contributing-categories-display.png`

**Expected Results:**
- ✅ Text: "Contributing categories: [list]"
- ✅ Categories comma-separated
- ✅ Displayed below underfunded amount
- ✅ Red text color matching warning
- ✅ Small font size for secondary info
- ✅ Works with single or multiple categories

**Pass Criteria:**
- Correct category names listed
- Proper formatting and styling
- Useful context for user

---

### Test 8: Ready to Assign Accounting

**Objective:** Verify RTA formula accounts for underfunded amounts

**Test ID:** TC-ALLOCATE-008

**Formula:** RTA = Total Inflows - Total Allocations - Total Underfunded

**Steps:**
1. Create clean test state:
   - Income: $5,000
   - Allocations to regular categories: $1,500
   - Credit card debt: $500
   - Payment category allocation: $300
   - Underfunded: $200
2. Calculate expected RTA:
   - RTA = $5,000 - $1,500 - $200 = $3,300
3. Navigate to Budget page
4. Verify Ready to Assign displays: $3,300.00
5. Click "Allocate to Cover" to cover $200 underfunded
6. After allocation success, verify new state:
   - Total allocations: $1,500 + $200 = $1,700
   - Underfunded: $0 (covered)
   - RTA = $5,000 - $1,700 - $0 = $3,300
7. Verify RTA unchanged (underfunded moved to allocation)
8. Take screenshots before and after

**Expected Results:**
- ✅ Before: RTA = Inflows - Allocations - Underfunded = $3,300
- ✅ After: RTA = Inflows - (Allocations + covered) - 0 = $3,300
- ✅ RTA unchanged after covering underfunded
- ✅ Formula demonstrates underfunded is accounted in RTA

**Pass Criteria:**
- RTA calculation matches formula
- Covering underfunded doesn't change RTA (moves from underfunded to allocation)
- Zero-based budgeting principle maintained

---

### Test 9: Loading State and User Feedback

**Objective:** Verify loading states and user feedback during allocation

**Test ID:** TC-ALLOCATE-009

**Steps:**
1. Navigate to Budget page with underfunded payment category
2. Open browser DevTools Network tab
3. Throttle network to "Slow 3G" to observe loading state
4. Click "Allocate to Cover" button
5. Immediately observe:
   - Button text changes to "Allocating..."
   - Button becomes disabled (disabled attribute)
   - Cursor changes to not-allowed
6. Monitor network request:
   - POST to /api/allocations/cover-underfunded
   - Request body contains payment_category_id and period
7. Wait for response
8. Observe success state:
   - Toast appears with success message
   - Button text reverts
   - Button re-enables
   - UI refreshes
9. Reset state and test error scenario with network failure
10. Disconnect network
11. Click "Allocate to Cover"
12. Verify error handling:
    - Error toast appears
    - Button recovers
13. Restore network and verify still works
14. Take screenshots of each state

**Expected Results:**
- ✅ Loading state is clear and immediate
- ✅ Button disabled during API call prevents double-clicks
- ✅ Success feedback is positive and informative
- ✅ Error feedback is clear and actionable
- ✅ Network errors handled gracefully
- ✅ UI always recovers to usable state

**Pass Criteria:**
- User always knows what's happening
- No way to trigger duplicate requests
- All states have appropriate feedback

---

### Test 10: Period Context

**Objective:** Verify allocation uses correct budget period

**Test ID:** TC-ALLOCATE-010

**Steps:**
1. Navigate to Budget page
2. Verify current period selector shows: "November 2025"
3. Create underfunded state in current period
4. Click "Allocate to Cover"
5. Verify allocation created for period: "2025-11"
6. Change period to "December 2025"
7. Create new underfunded state (new transactions)
8. Click "Allocate to Cover"
9. Verify allocation created for period: "2025-12"
10. Switch back to November
11. Verify November allocation still exists
12. Switch to December
13. Verify December allocation exists
14. Take screenshots showing period context

**Expected Results:**
- ✅ Allocation created for currently selected period
- ✅ Period from getCurrentPeriod() function
- ✅ Multiple periods can have separate allocations
- ✅ Period selector controls which allocation is created

**Pass Criteria:**
- Correct period used for allocation
- Period switching doesn't break functionality
- Each period maintains independent state

---

## Regression Tests

### Test 11: Non-Payment Categories

**Objective:** Verify "Allocate to Cover" button does NOT appear for non-payment categories

**Steps:**
1. Navigate to Budget page
2. Locate regular expense categories (Groceries, Gas, etc.)
3. Verify "Allocate to Cover" button does NOT appear
4. Create underspending in regular category (allocated > spent)
5. Verify no "Allocate to Cover" button
6. Take screenshot

**Expected Results:**
- ✅ Button only appears for payment categories
- ✅ Regular categories show standard allocation UI

---

### Test 12: Zero Underfunded State

**Objective:** Verify no warning or button when payment category is fully funded

**Steps:**
1. Create credit card with debt: -$300
2. Allocate exactly $300 to payment category
3. Navigate to Budget page
4. Verify payment category row shows:
   - Allocated: $300.00
   - Available: $300.00
   - NO underfunded warning
   - NO "Allocate to Cover" button
5. Take screenshot

**Expected Results:**
- ✅ No underfunded warning shown
- ✅ No "Allocate to Cover" button
- ✅ Payment category displays normally

---

## Edge Cases

### Test 13: Concurrent Clicks (Double-Click Prevention)

**Objective:** Verify button cannot be clicked twice

**Steps:**
1. Navigate to underfunded payment category
2. Click "Allocate to Cover" button rapidly twice
3. Verify only one API request sent
4. Verify button disabled after first click prevents second click
5. Check browser console for errors

**Expected Results:**
- ✅ Only one allocation created
- ✅ Button disabled state prevents duplicate requests
- ✅ No errors from double-click

---

### Test 14: Partial Coverage Not Allowed

**Objective:** Verify system covers full underfunded amount (no partial)

**Steps:**
1. Create underfunded: $500
2. Click "Allocate to Cover"
3. Verify allocation created for full $500 (not partial)
4. Verify underfunded becomes $0 (fully covered)

**Expected Results:**
- ✅ Full amount allocated (no partial coverage option)
- ✅ Underfunded fully resolved

---

## Browser Console Checks

For each test, verify:
- No JavaScript errors in console
- API calls logged (if debug mode enabled)
- Network requests successful (200 status codes)
- No CORS errors
- No 404 or 500 errors

## Performance Criteria

- Button click to loading state: < 100ms
- API response time: < 500ms (local)
- UI refresh after success: < 200ms
- Toast display: Immediate
- No UI freezing or lag

## Accessibility Checks

- Button has proper title/tooltip
- Colors meet WCAG contrast requirements
- Button keyboard accessible (Tab to focus, Enter to activate)
- Screen reader friendly (semantic HTML)

## Test Execution Checklist

Before starting:
- [ ] Application running at http://localhost:8080
- [ ] Database initialized
- [ ] Browser DevTools open
- [ ] Network tab monitoring
- [ ] Console tab monitoring
- [ ] Screenshots directory ready

During testing:
- [ ] Execute each test in order
- [ ] Take screenshots at key points
- [ ] Note any deviations from expected results
- [ ] Record console errors
- [ ] Record network errors
- [ ] Note performance issues

After testing:
- [ ] Compile test results
- [ ] Organize screenshots
- [ ] Document bugs found
- [ ] Create bug reports for failures
- [ ] Update test plan with findings

## Test Results Template

```
Test ID: TC-ALLOCATE-XXX
Test Name: [Name]
Status: ✅ PASS / ❌ FAIL / ⚠️ PARTIAL
Execution Date: [Date]
Browser: Chrome [version]

Results:
- Step 1: ✅ PASS
- Step 2: ✅ PASS
- Step 3: ❌ FAIL - [description]

Bugs Found:
1. [Bug description]
   - Severity: High/Medium/Low
   - Steps to reproduce: [steps]
   - Expected: [expected]
   - Actual: [actual]
   - Screenshot: [filename]

Notes:
[Any additional observations]
```

## Success Criteria

All tests must pass:
- ✅ All 14 test scenarios pass
- ✅ No critical or high severity bugs
- ✅ Browser console clean (no errors)
- ✅ Network requests successful
- ✅ UI matches specification
- ✅ User experience is smooth and intuitive

## Next Steps After Testing

If all tests pass:
1. Generate automated Playwright test files
2. Integrate tests into CI/CD pipeline
3. Update documentation
4. Mark feature as complete

If tests fail:
1. Document bugs with reproduction steps
2. Create GitHub issues for bugs
3. Prioritize bug fixes
4. Retest after fixes
5. Repeat until all tests pass
