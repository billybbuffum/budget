# Manual Test: Allocate to Cover Feature

**Date:** 2025-11-01
**Feature:** Allocate to Cover Button for Underfunded Payment Categories
**App URL:** http://localhost:8080
**Tester:** _________________
**Result:** ☐ PASS  ☐ FAIL

---

## Prerequisites

✅ Application is running at http://localhost:8080
✅ Browser is open (Chrome, Firefox, or Safari)
✅ You have 10-15 minutes for testing

---

## Test Scenario: Create Underfunded Credit Card Payment Category

### Step 1: Setup - Create Checking Account with Income

**Goal:** Create a source of funds

**Actions:**
1. Open http://localhost:8080
2. Click **"+ Add Transaction"** button (top right)
3. Fill in the form:
   - **Account:** Click "Add Account" → Name: "Test Checking", Type: "Checking" → Save
   - **Category:** Click "Add Category" → Name: "Salary", Group: Create new group "Test Income" → Save
   - **Amount:** `5000` (positive for income)
   - **Description:** "Test Income"
   - **Date:** Today's date
4. Click **Save**

**Expected Result:**
- ✅ Success toast appears
- ✅ Checking account shows $5,000.00 balance
- ✅ Transaction appears in sidebar

**Actual Result:**
☐ PASS  ☐ FAIL
**Notes:** ___________________________________________

---

### Step 2: Setup - Create Credit Card Account

**Goal:** Create a credit card account (which auto-creates a payment category)

**Actions:**
1. In the left sidebar, under "Accounts", click **"+ Add"**
2. Fill in the form:
   - **Name:** "Test Credit Card"
   - **Type:** Select "Checking" (if no credit card option exists)
   - **Initial Balance:** `0`
3. Click **Save**

**Expected Result:**
- ✅ Credit card account appears in sidebar
- ✅ Balance shows $0.00

**Actual Result:**
☐ PASS  ☐ FAIL
**Notes:** ___________________________________________

---

### Step 3: Create Credit Card Spending (Groceries)

**Goal:** Make a credit card purchase to create debt

**Actions:**
1. Click **"+ Add Transaction"** button
2. Fill in the form:
   - **Account:** Select "Test Credit Card"
   - **Category:** Click "Add Category" → Name: "Groceries", Group: "Test Expenses" (create if needed) → Save
   - **Amount:** `-200` (negative for spending)
   - **Description:** "Groceries purchase"
   - **Date:** Today's date
3. Click **Save**

**Expected Result:**
- ✅ Success toast appears
- ✅ Credit card balance shows -$200.00 (debt)
- ✅ Transaction appears in list

**Actual Result:**
☐ PASS  ☐ FAIL
**Notes:** ___________________________________________

---

### Step 4: Navigate to Budget Tab

**Goal:** View the budget to see Ready to Assign

**Actions:**
1. Look for navigation tabs/links
2. Click **"Budget"** (if it exists) or check if you're already on the budget view

**Expected Result:**
- ✅ Budget view loads
- ✅ "Ready to Assign" box shows at top
- ✅ Shows amount like $5,000.00

**Actual Result:**
☐ PASS  ☐ FAIL
**Ready to Assign Amount:** $ ___________
**Notes:** ___________________________________________

---

### Step 5: Check for Payment Category

**Goal:** Verify a payment category was auto-created for the credit card

**Actions:**
1. Scroll through category groups in the budget view
2. Look for a category with a name like:
   - "Test Credit Card - Payment" or
   - "Test Credit Card Payment" or
   - Similar naming pattern

**Expected Result:**
- ✅ Payment category exists
- ✅ Category name references the credit card account
- ✅ May show in a special group or within regular categories

**Actual Result:**
☐ PASS  ☐ FAIL
**Payment Category Name:** ___________________________________________
**Notes:** ___________________________________________

---

### Step 6: Expand Category Groups (if collapsed)

**Goal:** Make sure all categories are visible

**Actions:**
1. Look for an **"Expand All"** button (usually near "Categories" heading)
2. If found, click it
3. Alternatively, click individual group headers to expand them

**Expected Result:**
- ✅ All category groups expand
- ✅ All categories become visible
- ✅ Payment category is visible

**Actual Result:**
☐ PASS  ☐ FAIL
**Notes:** ___________________________________________

---

### Step 7: Allocate Partial Amount to Payment Category

**Goal:** Create an underfunded state by allocating less than the debt

**Actions:**
1. Find the payment category for "Test Credit Card"
2. Look for an "Allocated" or "Amount" input field in that row
3. Click the input field
4. Enter: `100` (only $100, but debt is $200)
5. Press Enter or click outside to save
6. Wait for save confirmation

**Expected Result:**
- ✅ Allocation saved successfully
- ✅ Category shows $100.00 allocated
- ✅ Ready to Assign decreases by $100 (should show $4,900.00)

**Actual Result:**
☐ PASS  ☐ FAIL
**Ready to Assign After:** $ ___________
**Allocated Amount:** $ ___________
**Notes:** ___________________________________________

---

### Step 8: ⭐ VERIFY UNDERFUNDED WARNING

**Goal:** Verify the underfunded warning appears

**Actions:**
1. Look at the payment category row
2. Look for warning indicators:
   - ⚠️ emoji
   - Red text
   - Text saying "Underfunded"
   - Amount showing "$100.00 more" or similar

**Expected Result:**
- ✅ ⚠️ Warning icon appears
- ✅ Red text saying "Underfunded"
- ✅ Shows "Need $100.00 more" or similar
- ✅ May show contributing categories (e.g., "Groceries")

**Actual Result:**
☐ PASS  ☐ FAIL
**Warning Text Seen:** ___________________________________________
**Underfunded Amount Shown:** $ ___________
**Screenshot Taken?** ☐ Yes  ☐ No
**Notes:** ___________________________________________

---

### Step 9: ⭐ VERIFY "ALLOCATE TO COVER" BUTTON

**Goal:** Verify the "Allocate to Cover" button appears

**Actions:**
1. Still looking at the payment category row
2. Look for a button labeled "Allocate to Cover"
3. Note its styling:
   - Color (should be blue)
   - Position (likely next to the underfunded warning)

**Expected Result:**
- ✅ "Allocate to Cover" button is visible
- ✅ Button is blue (bg-blue-600 class)
- ✅ Button has white text
- ✅ Button appears clickable (not disabled)

**Actual Result:**
☐ PASS  ☐ FAIL
**Button Text:** ___________________________________________
**Button Color:** ___________________________________________
**Screenshot Taken?** ☐ Yes  ☐ No
**Notes:** ___________________________________________

---

### Step 10: ⭐ CLICK "ALLOCATE TO COVER" BUTTON

**Goal:** Test the allocation functionality

**Actions:**
1. Hover over the "Allocate to Cover" button
2. Click the button
3. Watch for:
   - Loading state ("Allocating..." text)
   - Success toast notification
   - UI updates

**Expected Result:**
- ✅ Button changes to "Allocating..." briefly
- ✅ Green success toast appears
- ✅ Toast says something like "Successfully allocated $100.00 to cover..."
- ✅ UI refreshes automatically

**Actual Result:**
☐ PASS  ☐ FAIL
**Loading State Seen?** ☐ Yes  ☐ No
**Toast Message:** ___________________________________________
**Screenshot Taken?** ☐ Yes  ☐ No
**Notes:** ___________________________________________

---

### Step 11: ⭐ VERIFY UNDERFUNDED RESOLVED

**Goal:** Confirm the underfunded state is resolved

**Actions:**
1. Look at the payment category row again
2. Check for:
   - Underfunded warning gone
   - "Allocate to Cover" button gone
   - Allocated amount increased
   - Ready to Assign decreased

**Expected Result:**
- ✅ ⚠️ Warning icon is GONE
- ✅ "Underfunded" text is GONE
- ✅ "Allocate to Cover" button is GONE
- ✅ Allocated amount now shows $200.00 (was $100, added $100)
- ✅ Ready to Assign now shows $4,800.00 (was $4,900, decreased by $100)

**Actual Result:**
☐ PASS  ☐ FAIL
**Warning Still Visible?** ☐ Yes  ☐ No
**Button Still Visible?** ☐ Yes  ☐ No
**New Allocated Amount:** $ ___________
**New Ready to Assign:** $ ___________
**Screenshot Taken?** ☐ Yes  ☐ No
**Notes:** ___________________________________________

---

### Step 12: ⭐ VERIFY AVAILABLE BALANCE

**Goal:** Confirm the payment category "Available" amount is correct

**Actions:**
1. Look at the payment category row
2. Find the "Available" column
3. Note the amount

**Expected Result:**
- ✅ Available should show $200.00
- ✅ This matches the credit card debt ($200)
- ✅ This means the payment category is fully funded

**Calculation:**
- Allocated: $200
- Spent/Debt: $200
- Available: $200 - $0 (no actual payment made yet) = $200

**Actual Result:**
☐ PASS  ☐ FAIL
**Available Amount Shown:** $ ___________
**Notes:** ___________________________________________

---

## Bonus Test: Insufficient Funds Scenario

**Goal:** Test error handling when RTA is less than underfunded amount

### Steps:

1. Create another credit card account ("Test CC 2")
2. Make a $500 purchase on it
3. Allocate $100 to its payment category (creates $400 underfunded)
4. Allocate most of your Ready to Assign to other categories (leave only $100 RTA)
5. Try to click "Allocate to Cover" on the $400 underfunded category

**Expected Result:**
- ✅ Red error toast appears
- ✅ Error message says something like: "Insufficient funds: Ready to Assign: $100.00, Underfunded: $400.00"
- ✅ Allocation is NOT created
- ✅ Underfunded state remains

**Actual Result:**
☐ PASS  ☐ FAIL
**Error Message:** ___________________________________________
**Notes:** ___________________________________________

---

## Bonus Test: Regular Category (No Button)

**Goal:** Verify button does NOT appear for non-payment categories

### Steps:

1. Look at the "Groceries" category row (or any regular expense category)
2. Check if "Allocate to Cover" button appears

**Expected Result:**
- ✅ No "Allocate to Cover" button on regular categories
- ✅ Only payment categories should have this button

**Actual Result:**
☐ PASS  ☐ FAIL
**Notes:** ___________________________________________

---

## Overall Test Result

**Feature Works As Expected?** ☐ YES  ☐ NO

**Critical Issues Found:**
___________________________________________
___________________________________________

**Minor Issues Found:**
___________________________________________
___________________________________________

**Recommendations:**
___________________________________________
___________________________________________

**Tester Signature:** _________________
**Date Completed:** _________________

---

## Troubleshooting

### Issue: Payment category not appearing

**Possible Causes:**
- Payment categories might not auto-create for "checking" type accounts
- May need to manually create a category with specific settings
- Check if there's a setting to enable/disable payment category creation

**Solution:**
- Manually create a category
- Link it to the credit card account
- Check account settings

### Issue: Category groups are collapsed

**Solution:**
- Click "Expand All" button
- Click individual group headers
- Check if there's a setting to remember expand/collapse state

### Issue: Changes not saving

**Solution:**
- Check browser console for errors (F12 → Console tab)
- Verify API calls are succeeding
- Try refreshing the page

### Issue: Numbers don't match expected

**Possible Cause:**
- Previous test data still in database
- Multiple test runs accumulated data

**Solution:**
- Note actual numbers and adjust test expectations
- Or reset database and rerun test

---

## Success Criteria Checklist

✅ All expected UI elements appear
✅ Button click creates allocation
✅ Success toast appears
✅ Underfunded state resolves
✅ Ready to Assign updates correctly
✅ No console errors
✅ Feature is intuitive and easy to use

**Overall Feature Assessment:** ☐ Production Ready  ☐ Needs Work
