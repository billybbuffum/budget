# Manual Testing Guide: Credit Card, Transfer, and Category Refactoring Features

## Overview
This branch implements several major features:
1. **Simplified Category Model** - Removed category types (income/expense distinction)
2. **Optional Income Categorization** - Income transactions no longer require categories
3. **Credit Card Support** - Credit cards with automatic payment category allocation
4. **Transfer Functionality** - Move money between accounts without affecting Ready to Assign
5. **Enhanced UI** - Updated interface to support all new features

## Prerequisites

### Setup
1. Clone the repository and checkout the branch:
   ```bash
   git checkout claude/refactor-transaction-categories-011CUcC5UGZs7GYszcw1GqKb
   ```

2. Build and start the server:
   ```bash
   go build -o budget ./cmd/server
   ./budget
   ```

3. Open your browser to `http://localhost:8080`

### Understanding the Mental Model
The budgeting system follows this logic:
- **Ready to Assign**: Pool of unallocated money (starts with account balances)
- **Allocations**: Money you've assigned to categories for spending
- **Available**: What's left in a category (Allocated - Spent)
- **Credit Cards**: Spending on credit automatically moves budget from expense category → payment category

---

## Test Plan

### Phase 1: Basic Account Setup

#### Test 1.1: Create Checking Account
**Steps:**
1. Navigate to "Accounts" view
2. Click "+ Add Account"
3. Fill in:
   - Name: "Main Checking"
   - Type: "Checking"
   - Starting Balance: 5000.00
4. Click "Add Account"

**Expected Results:**
- ✅ Account appears in accounts list
- ✅ Balance shows $5,000.00 in green
- ✅ Total Balance shows $5,000.00

#### Test 1.2: Create Savings Account
**Steps:**
1. Click "+ Add Account"
2. Fill in:
   - Name: "Emergency Fund"
   - Type: "Savings"
   - Starting Balance: 2000.00
3. Click "Add Account"

**Expected Results:**
- ✅ Both accounts visible
- ✅ Total Balance shows $7,000.00

#### Test 1.3: Create Credit Card Account (NEW FEATURE)
**Steps:**
1. Click "+ Add Account"
2. Fill in:
   - Name: "Visa Card"
   - Type: "Credit Card" ← **New option**
   - Starting Balance: -500.00 (negative for debt)
3. Click "Add Account"

**Expected Results:**
- ✅ Credit card appears with **negative balance in red**: -$500.00
- ✅ Total Balance shows $6,500.00 ($7,000 - $500 debt)
- ✅ Navigate to "Categories" → **"Visa Card Payment" category auto-created** ← **KEY TEST**
  - Should have red color (#FF6B6B)
  - Should NOT appear in user-editable categories list (filtered out)

---

### Phase 2: Category Management

#### Test 2.1: Create Categories (Simplified - No Type Required)
**Steps:**
1. Navigate to "Categories" view
2. Click "+ Add Category"
3. Fill in:
   - Name: "Groceries"
   - Color: Green (#22c55e)
   - Description: "Food and household items"
   - **NOTE: No "Type" dropdown should be visible** ← **KEY TEST**
4. Click "Add Category"

5. Repeat for:
   - "Gas" - Orange color, "Fuel for car"
   - "Rent" - Blue color, "Monthly rent"
   - "Entertainment" - Purple color, "Fun stuff"

**Expected Results:**
- ✅ **No category type selection** (old UI had Income/Expense dropdown)
- ✅ All categories appear in single list (not split into Income/Expense sections)
- ✅ "Visa Card Payment" should **NOT** appear in the list ← **KEY TEST**

**API Verification (Optional):**
```bash
curl http://localhost:8080/api/categories | jq '.[] | {name, payment_for_account_id}'
```
- User categories should have `payment_for_account_id: null`
- Payment category should have `payment_for_account_id: "<visa-account-id>"`

---

### Phase 3: Income & Budget Allocation

#### Test 3.1: Add Income WITHOUT Category (NEW FEATURE)
**Steps:**
1. Navigate to "Budget" or "All Transactions" view
2. Click "+ Add Transaction"
3. Fill in:
   - Account: "Main Checking"
   - **Category: Leave empty** ← **KEY TEST**
   - Type: "Inflow (Income)"
   - Amount: 3000.00
   - Date: Today
   - Description: "Paycheck"
4. Click "Add Transaction"

**Expected Results:**
- ✅ Transaction creates successfully **without category** ← **KEY TEST**
- ✅ Checking account balance: $8,000.00 ($5,000 + $3,000)
- ✅ **Budget view → Ready to Assign: $10,000.00** ($7,000 initial + $3,000 income)

**UI Behavior Check:**
- When you change transaction type from "Outflow" to "Inflow":
  - ✅ Category field label changes from "Category *" to "Category" (no asterisk)
  - ✅ Helper text shows: "Required for expenses, optional for income"

#### Test 3.2: Allocate Budget to Categories
**Steps:**
1. Navigate to "Budget" view
2. For "Groceries" category:
   - Click on the "Allocated" amount ($0.00)
   - Enter: 500.00
   - Press Enter or click away
3. Repeat for:
   - Gas: $200.00
   - Rent: $1500.00
   - Entertainment: $100.00

**Expected Results:**
- ✅ Ready to Assign decreases: $10,000 → $7,700
- ✅ Each category shows:
  - Allocated: (amount you entered)
  - Spent: $0.00
  - Available: (same as allocated)

---

### Phase 4: Normal Spending

#### Test 4.1: Regular Expense Transaction
**Steps:**
1. Click "+ Add Transaction"
2. Fill in:
   - Account: "Main Checking"
   - Category: "Groceries"
   - Type: "Outflow (Expense)"
   - Amount: 150.00
   - Date: Today
   - Description: "Weekly grocery shopping"
3. Click "Add Transaction"

**Expected Results:**
- ✅ Checking balance: $7,850.00
- ✅ Groceries category in Budget view:
  - Allocated: $500.00
  - Spent: $150.00
  - Available: $350.00 (in green)
- ✅ Ready to Assign: Still $7,700 (unchanged)

---

### Phase 5: Credit Card Spending (CORE NEW FEATURE)

#### Test 5.1: Credit Card Purchase with Category
**Steps:**
1. Click "+ Add Transaction"
2. Fill in:
   - Account: **"Visa Card"** ← Credit card
   - Category: "Gas"
   - Type: "Outflow (Expense)"
   - Amount: 50.00
   - Date: Today
   - Description: "Gas station"
3. Click "Add Transaction"

**Expected Results - This is the MAGIC:** ✨
- ✅ **Visa Card balance**: -$550.00 (debt increased)
- ✅ **Gas category**:
  - Allocated: $200.00
  - Spent: $50.00
  - Available: $150.00
- ✅ **"Visa Card Payment" category** (navigate to Budget view, scroll down):
  - Allocated: **$50.00** ← **AUTO-ALLOCATED** ← **KEY TEST**
  - Available: $50.00
  - **This money "moved" from Gas to Visa Payment automatically**

**What Just Happened:**
The system automatically moved $50 from your "Gas" budget to "Visa Card Payment" budget. This represents money you need to set aside to pay the credit card bill.

#### Test 5.2: Multiple Credit Card Purchases
**Steps:**
1. Add another transaction:
   - Account: "Visa Card"
   - Category: "Groceries"
   - Amount: 75.00
   - Description: "Groceries on credit"

2. Add third transaction:
   - Account: "Visa Card"
   - Category: "Entertainment"
   - Amount: 25.00
   - Description: "Movie tickets"

**Expected Results:**
- ✅ Visa Card balance: -$650.00
- ✅ Each category shows spending
- ✅ **"Visa Card Payment" allocation: $150.00** ($50 + $75 + $25)
- ✅ This $150 is ready to pay your credit card bill

---

### Phase 6: Transfers (NEW FEATURE)

#### Test 6.1: Transfer Between Checking and Savings
**Steps:**
1. Click the new **"Transfer" button** in header (next to "+ Add Transaction")
2. Fill in:
   - From Account: "Main Checking"
   - To Account: "Emergency Fund"
   - Amount: 1000.00
   - Date: Today
   - Description: "Monthly savings transfer"
3. Click "Create Transfer"

**Expected Results:**
- ✅ Checking balance: $6,850.00 ($7,850 - $1,000)
- ✅ Savings balance: $3,000.00 ($2,000 + $1,000)
- ✅ **Total Balance: Still $6,500.00** (unchanged - money just moved) ← **KEY TEST**
- ✅ **Ready to Assign: Still $7,700** (unchanged) ← **KEY TEST**
- ✅ Navigate to "All Transactions":
  - Should see 2 transfer transactions
  - One shows -$1,000 from Checking
  - One shows +$1,000 to Savings
  - Description format: "Transfer: Main Checking → Emergency Fund"
  - **No category** (transfers don't have categories)

#### Test 6.2: Pay Credit Card (Transfer from Checking to Credit Card)
**Steps:**
1. Click "Transfer" button
2. Fill in:
   - From Account: "Main Checking"
   - To Account: "Visa Card"
   - Amount: 150.00
   - Description: "Pay credit card bill"
3. Click "Create Transfer"

**Expected Results:**
- ✅ Checking balance: $6,700.00
- ✅ Visa Card balance: **-$500.00** (paid down from -$650) ← **KEY TEST**
- ✅ Total Balance: Still $6,500.00
- ✅ **Ready to Assign: Still $7,700** (transfers don't affect this)
- ✅ Visa Card Payment category still shows $150.00 available
  - **Note:** The allocation stays there - you've used "real" money to pay the bill

---

### Phase 7: Edge Cases & Validations

#### Test 7.1: Try to Create Transfer to Same Account
**Steps:**
1. Click "Transfer"
2. Select same account for both From and To
3. Try to submit

**Expected Results:**
- ✅ Error message: "Cannot transfer to the same account"

#### Test 7.2: Try to Create Transfer with Less Than 2 Accounts
**Steps:**
1. Create a fresh database (or test with only 1 account)
2. Click "Transfer"

**Expected Results:**
- ✅ Error message: "You need at least 2 accounts to make a transfer"

#### Test 7.3: Try to Create Expense WITHOUT Category
**Steps:**
1. Click "+ Add Transaction"
2. Select Type: "Outflow (Expense)"
3. Try to submit without selecting category

**Expected Results:**
- ✅ Error: "Please select a category for expenses"
- **Income allows no category, but expenses require one** ← **KEY TEST**

#### Test 7.4: Income with Optional Category
**Steps:**
1. Click "+ Add Transaction"
2. Fill in:
   - Account: "Main Checking"
   - Category: "Entertainment" (choose one this time)
   - Type: "Inflow (Income)"
   - Amount: 50.00
   - Description: "Side gig payment"

**Expected Results:**
- ✅ Transaction creates successfully
- ✅ Income CAN have a category, it's just optional
- ✅ Ready to Assign increases by $50

---

### Phase 8: Transaction Display & Filtering

#### Test 8.1: View All Transactions
**Steps:**
1. Navigate to "All Transactions"
2. Review the list

**Expected Results:**
- ✅ Normal transactions show: Date • Account • Category
- ✅ **Transfer transactions show**: "Transfer: From Account → To Account" ← **KEY TEST**
- ✅ Income without category shows: Date • Account (no category)
- ✅ Transactions sorted by date (newest first)
- ✅ Color coding:
  - Positive amounts in green
  - Negative amounts in red

#### Test 8.2: Verify Payment Category is Hidden
**Steps:**
1. Try to create new transaction
2. Look at Category dropdown

**Expected Results:**
- ✅ "Groceries", "Gas", "Rent", "Entertainment" visible
- ✅ **"Visa Card Payment" NOT in dropdown** ← **KEY TEST**
- Payment categories are system-managed, not user-selectable

---

### Phase 9: Budget Summary Verification

#### Test 9.1: Final Budget State
**Steps:**
1. Navigate to "Budget" view
2. Review all numbers

**Expected State After All Tests:**
```
Ready to Assign: $7,750.00

User Categories (visible):
├── Groceries
│   ├── Allocated: $500.00
│   ├── Spent: $225.00 ($150 checking + $75 credit)
│   └── Available: $275.00
├── Gas
│   ├── Allocated: $200.00
│   ├── Spent: $50.00
│   └── Available: $150.00
├── Rent
│   ├── Allocated: $1,500.00
│   ├── Spent: $0.00
│   └── Available: $1,500.00
└── Entertainment
    ├── Allocated: $100.00
    ├── Spent: $25.00
    └── Available: $75.00

System Category (not visible in UI):
└── Visa Card Payment
    ├── Allocated: $150.00 (auto-allocated from credit card spending)
    ├── Spent: $0.00
    └── Available: $150.00
```

**Account Balances:**
- Main Checking: $6,700.00
- Emergency Fund: $3,000.00
- Visa Card: -$500.00
- **Total: $6,500.00** (should match initial $7,000 - $500 debt)

---

## API Testing (Optional Advanced Verification)

### Check Transaction Structure
```bash
# List all transactions
curl http://localhost:8080/api/transactions | jq .

# Verify transfer transactions have:
# - type: "transfer"
# - transfer_to_account_id: <uuid>
# - category_id: null

# Verify normal transactions have:
# - type: "normal"
# - category_id: <uuid> or null (for income)
```

### Check Category Structure
```bash
# List categories
curl http://localhost:8080/api/categories | jq .

# User categories should have:
# - payment_for_account_id: null

# Payment category should have:
# - payment_for_account_id: <credit-card-account-id>
```

---

## Regression Testing Checklist

Test that old features still work:

- [ ] Can create/edit/delete accounts
- [ ] Can create/edit/delete categories
- [ ] Can view account summary with total balance
- [ ] Month navigation in Budget view works
- [ ] Allocating budget decreases Ready to Assign
- [ ] Spending decreases category Available amount
- [ ] All modals open and close correctly
- [ ] Form validation works (required fields, etc.)

---

## Known Behaviors (Not Bugs)

1. **Payment categories don't appear in UI lists** - This is intentional. They're system-managed.

2. **Transfers don't change Ready to Assign** - Correct! You're just moving money between accounts.

3. **Credit card spending increases payment category allocation** - This is the core feature! The system automatically sets aside money to pay the credit card.

4. **Paying credit card (via transfer) doesn't reduce payment category allocation** - The allocation represents "money you should use for payment". Actually paying is a separate action.

5. **Income without category doesn't show "Unknown" in transaction list** - Correct, it just doesn't display a category.

---

## Bug Reporting Template

If you find issues, please report with:

```
**Bug Title:** Brief description

**Steps to Reproduce:**
1. ...
2. ...
3. ...

**Expected Result:**
What should happen

**Actual Result:**
What actually happened

**Screenshots:**
(if applicable)

**Browser/Environment:**
- Browser: Chrome/Firefox/Safari
- OS: Windows/Mac/Linux
```

---

## Success Criteria

All tests pass if:
- ✅ Credit card accounts can be created and display correctly
- ✅ Payment categories are auto-created and hidden from UI
- ✅ Credit card spending auto-allocates to payment category
- ✅ Transfers work and don't affect Ready to Assign
- ✅ Income transactions work without categories
- ✅ Category creation no longer requires type selection
- ✅ All transaction types display correctly in lists
- ✅ No build errors or console errors
- ✅ All calculations are accurate

---

**Estimated Testing Time:** 30-45 minutes for complete walkthrough

**Questions?** Check the commit history for implementation details or reach out to the development team.
