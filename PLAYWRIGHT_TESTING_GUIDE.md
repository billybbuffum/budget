# Playwright Testing Guide - Credit Card & Transfer Features

## Overview
This guide provides comprehensive Playwright test scenarios for the credit card, transfer, and category refactoring features implemented in this branch.

## Prerequisites

### Setup Playwright
```bash
npm init playwright@latest
```

### Test Configuration
```javascript
// playwright.config.js
import { defineConfig } from '@playwright/test';

export default defineConfig({
  testDir: './tests',
  fullyParallel: true,
  forbidOnly: !!process.env.CI,
  retries: process.env.CI ? 2 : 0,
  workers: process.env.CI ? 1 : undefined,
  reporter: 'html',
  use: {
    baseURL: 'http://localhost:8080',
    trace: 'on-first-retry',
  },
});
```

### Start Server
```bash
# In one terminal
go build -o budget ./cmd/server
./budget

# In another terminal
npx playwright test
```

---

## Test Suite 1: Category Management (No Type Field)

### Test 1.1: Create Category with Predefined Color Palette

**Feature**: Categories no longer have an income/expense type. Color selection uses predefined palette.

```javascript
import { test, expect } from '@playwright/test';

test('should create category using predefined color palette', async ({ page }) => {
  await page.goto('/');

  // Navigate to categories
  await page.click('text=Categories');
  await page.click('text=+ Add Category');

  // Verify NO category type dropdown exists
  await expect(page.locator('#category-type')).not.toBeVisible();

  // Fill in category name
  await page.fill('#category-name', 'Test Groceries');

  // Verify predefined color palette is visible
  await expect(page.locator('.color-swatch')).toHaveCount(10);

  // Select green color (for groceries)
  await page.click('.color-swatch[data-color="#10b981"]');

  // Verify selection visual feedback
  await expect(page.locator('.color-swatch[data-color="#10b981"]')).toHaveClass(/selected/);
  await expect(page.locator('.color-swatch[data-color="#10b981"] .color-check')).not.toHaveClass(/hidden/);

  // Verify hidden input has correct value
  const colorValue = await page.inputValue('#category-color');
  expect(colorValue).toBe('#10b981');

  // Add description and submit
  await page.fill('#category-description', 'Weekly grocery shopping');
  await page.click('button[type="submit"]:has-text("Add Category")');

  // Verify success
  await expect(page.locator('#toast')).toContainText('Category created successfully');

  // Verify category appears in list
  await expect(page.locator('text=Test Groceries')).toBeVisible();
});
```

### Test 1.2: All 10 Predefined Colors Available

```javascript
test('should have all 10 predefined colors with correct semantic meanings', async ({ page }) => {
  await page.goto('/');
  await page.click('text=Categories');
  await page.click('text=+ Add Category');

  // Define expected colors with their semantic meanings
  const expectedColors = [
    { color: '#f97316', title: 'Orange - Household' },
    { color: '#3b82f6', title: 'Blue - Transportation' },
    { color: '#10b981', title: 'Green - Groceries' },
    { color: '#a855f7', title: 'Purple - Entertainment' },
    { color: '#ef4444', title: 'Red - Utilities' },
    { color: '#ec4899', title: 'Pink - Health' },
    { color: '#eab308', title: 'Yellow - Shopping' },
    { color: '#6366f1', title: 'Indigo - Subscriptions' },
    { color: '#14b8a6', title: 'Teal - Savings' },
    { color: '#6b7280', title: 'Gray - Other' }
  ];

  for (const { color, title } of expectedColors) {
    const swatch = page.locator(`.color-swatch[data-color="${color}"]`);
    await expect(swatch).toBeVisible();
    await expect(swatch).toHaveAttribute('title', title);
  }

  // Blue should be selected by default
  await expect(page.locator('.color-swatch[data-color="#3b82f6"]')).toHaveClass(/selected/);
});
```

### Test 1.3: Categories List Shows No Type

```javascript
test('should display categories without type indicator', async ({ page }) => {
  await page.goto('/');
  await page.click('text=Categories');

  // Categories should not show "(expense)" or "(income)" labels
  const categoryCards = page.locator('#categories-list > div');
  const count = await categoryCards.count();

  for (let i = 0; i < count; i++) {
    const card = categoryCards.nth(i);
    const text = await card.textContent();
    expect(text).not.toContain('(expense)');
    expect(text).not.toContain('(income)');
  }
});
```

---

## Test Suite 2: Optional Income Categorization

### Test 2.1: Category Required for Expenses

```javascript
test('should require category for expense transactions', async ({ page }) => {
  await page.goto('/');
  await page.click('text=+ Add Transaction');

  // Set to outflow (expense)
  await page.selectOption('#transaction-type', 'outflow');

  // Verify category is required
  await expect(page.locator('#category-required-indicator')).toHaveText('*');
  await expect(page.locator('#transaction-category')).toHaveAttribute('required', '');
  await expect(page.locator('#category-hint')).toContainText('Required for expenses');
});
```

### Test 2.2: Category Optional for Income

```javascript
test('should make category optional for income transactions', async ({ page }) => {
  await page.goto('/');
  await page.click('text=+ Add Transaction');

  // Set to inflow (income)
  await page.selectOption('#transaction-type', 'inflow');

  // Verify category is optional
  await expect(page.locator('#category-required-indicator')).toHaveText('');
  await expect(page.locator('#transaction-category')).not.toHaveAttribute('required');
  await expect(page.locator('#category-hint')).toContainText('optional for income');
});
```

### Test 2.3: Create Income Transaction Without Category

```javascript
test('should successfully create income transaction without category', async ({ page }) => {
  await page.goto('/');

  // Create transaction
  await page.click('text=+ Add Transaction');
  await page.selectOption('#transaction-account', { index: 1 }); // First account
  await page.selectOption('#transaction-type', 'inflow');

  // Leave category empty
  await page.selectOption('#transaction-category', '');

  await page.fill('#transaction-amount', '1000');
  await page.fill('#transaction-date', '2025-10-30');
  await page.fill('#transaction-description', 'Paycheck');

  await page.click('button[type="submit"]:has-text("Add Transaction")');

  // Verify success
  await expect(page.locator('#toast')).toContainText('Transaction added successfully');

  // Navigate to transactions view and verify
  await page.click('text=All Transactions');
  await expect(page.locator('text=Paycheck')).toBeVisible();

  // Verify no category color dot appears
  const transactionCard = page.locator('text=Paycheck').locator('..');
  await expect(transactionCard.locator('.w-2.h-2.rounded-full')).not.toBeVisible();
});
```

### Test 2.4: Dynamic Category Requirement Toggle

```javascript
test('should toggle category requirement when switching transaction type', async ({ page }) => {
  await page.goto('/');
  await page.click('text=+ Add Transaction');

  const categorySelect = page.locator('#transaction-category');
  const indicator = page.locator('#category-required-indicator');

  // Start with outflow (required)
  await page.selectOption('#transaction-type', 'outflow');
  await expect(categorySelect).toHaveAttribute('required', '');
  await expect(indicator).toHaveText('*');

  // Switch to inflow (optional)
  await page.selectOption('#transaction-type', 'inflow');
  await expect(categorySelect).not.toHaveAttribute('required');
  await expect(indicator).toHaveText('');

  // Switch back to outflow (required again)
  await page.selectOption('#transaction-type', 'outflow');
  await expect(categorySelect).toHaveAttribute('required', '');
  await expect(indicator).toHaveText('*');
});
```

---

## Test Suite 3: Credit Card Support

### Test 3.1: Create Credit Card Account

```javascript
test('should create credit card account type', async ({ page }) => {
  await page.goto('/');
  await page.click('text=Accounts');
  await page.click('text=+ Add Account');

  // Verify credit card type is available
  await expect(page.locator('#account-type option[value="credit"]')).toBeVisible();

  await page.fill('#account-name', 'Chase Sapphire');
  await page.selectOption('#account-type', 'credit');
  await page.fill('#account-balance', '0');

  await page.click('button[type="submit"]:has-text("Add Account")');

  await expect(page.locator('#toast')).toContainText('Account created successfully');
  await expect(page.locator('text=Chase Sapphire')).toBeVisible();
});
```

### Test 3.2: Payment Category Auto-Created

```javascript
test('should auto-create payment category when credit card is created', async ({ page }) => {
  await page.goto('/');

  // Create credit card
  await page.click('text=Accounts');
  await page.click('text=+ Add Account');
  await page.fill('#account-name', 'Test CC');
  await page.selectOption('#account-type', 'credit');
  await page.fill('#account-balance', '0');
  await page.click('button[type="submit"]:has-text("Add Account")');

  // Go to budget view
  await page.click('text=Budget');

  // Verify payment category exists
  await expect(page.locator('text=Payment: Test CC')).toBeVisible();
  await expect(page.locator('text=(Auto-managed)')).toBeVisible();
});
```

### Test 3.3: Payment Category Visual Distinction

```javascript
test('should display payment category with orange background and auto-managed label', async ({ page }) => {
  await page.goto('/');

  // Assuming a credit card already exists
  await page.click('text=Budget');

  // Find payment category card
  const paymentCard = page.locator('.bg-orange-50').filter({ hasText: 'Payment:' });
  await expect(paymentCard).toBeVisible();

  // Verify auto-managed label
  await expect(paymentCard.locator('text=(Auto-managed)')).toBeVisible();
  await expect(paymentCard.locator('.text-orange-600')).toBeVisible();

  // Verify allocated amount is NOT editable (no cursor-pointer class)
  const allocatedDiv = paymentCard.locator('text=Allocated').locator('..').locator('div').nth(1);
  await expect(allocatedDiv).not.toHaveClass(/cursor-pointer/);
});
```

### Test 3.4: Payment Category Hidden from Dropdowns

```javascript
test('should hide payment categories from transaction category dropdown', async ({ page }) => {
  await page.goto('/');

  // Open transaction modal
  await page.click('text=+ Add Transaction');

  // Get all category options
  const categoryOptions = page.locator('#transaction-category option');
  const count = await categoryOptions.count();

  // Verify no payment categories in dropdown
  for (let i = 0; i < count; i++) {
    const optionText = await categoryOptions.nth(i).textContent();
    expect(optionText).not.toContain('Payment:');
  }
});
```

### Test 3.5: Credit Card Spending Allocates to Payment Category

```javascript
test('should allocate credit card spending to payment category', async ({ page }) => {
  await page.goto('/');

  // Get initial payment category state
  await page.click('text=Budget');
  const paymentCard = page.locator('.bg-orange-50').filter({ hasText: 'Payment:' }).first();
  const initialAllocated = await paymentCard.locator('text=Allocated').locator('..').locator('.font-semibold').textContent();

  // Create expense on credit card
  await page.click('text=+ Add Transaction');

  // Select credit card account
  const ccOption = await page.locator('#transaction-account option').filter({ hasText: /credit/i }).first();
  const ccValue = await ccOption.getAttribute('value');
  await page.selectOption('#transaction-account', ccValue);

  // Select a category (e.g., Groceries)
  await page.selectOption('#transaction-category', { index: 1 });
  await page.selectOption('#transaction-type', 'outflow');
  await page.fill('#transaction-amount', '50.00');
  await page.fill('#transaction-date', '2025-10-30');
  await page.fill('#transaction-description', 'Grocery shopping');

  await page.click('button[type="submit"]:has-text("Add Transaction")');
  await expect(page.locator('#toast')).toContainText('successfully');

  // Verify payment category allocated increased by $50
  await page.click('text=Budget');
  const newAllocated = await paymentCard.locator('text=Allocated').locator('..').locator('.font-semibold').textContent();

  // Parse currency and verify
  const initialAmount = parseFloat(initialAllocated.replace(/[$,]/g, ''));
  const newAmount = parseFloat(newAllocated.replace(/[$,]/g, ''));
  expect(newAmount).toBeCloseTo(initialAmount + 50, 2);
});
```

---

## Test Suite 4: Transfer Functionality

### Test 4.1: Transfer Button Visible

```javascript
test('should show transfer button in header', async ({ page }) => {
  await page.goto('/');

  await expect(page.locator('button:has-text("Transfer")')).toBeVisible();
  await expect(page.locator('button:has-text("Transfer")')).toHaveClass(/btn-secondary/);
});
```

### Test 4.2: Transfer Modal Opens

```javascript
test('should open transfer modal with correct fields', async ({ page }) => {
  await page.goto('/');
  await page.click('button:has-text("Transfer")');

  // Verify modal is visible
  await expect(page.locator('#transfer-modal')).toHaveClass(/active/);
  await expect(page.locator('h3:has-text("Transfer Between Accounts")')).toBeVisible();

  // Verify all required fields
  await expect(page.locator('#transfer-from-account')).toBeVisible();
  await expect(page.locator('#transfer-to-account')).toBeVisible();
  await expect(page.locator('#transfer-amount')).toBeVisible();
  await expect(page.locator('#transfer-date')).toBeVisible();
  await expect(page.locator('#transfer-description')).toBeVisible();
});
```

### Test 4.3: Require Minimum 2 Accounts for Transfer

```javascript
test('should show error if less than 2 accounts exist', async ({ page }) => {
  // This test requires starting with a fresh database or single account
  await page.goto('/');

  // If only 0-1 accounts exist
  const accountCount = await page.locator('#accounts-list > div').count();

  if (accountCount < 2) {
    await page.click('button:has-text("Transfer")');
    await expect(page.locator('#toast')).toContainText('You need at least 2 accounts');
  }
});
```

### Test 4.4: Create Transfer Between Accounts

```javascript
test('should successfully create transfer between accounts', async ({ page }) => {
  await page.goto('/');

  // Open transfer modal
  await page.click('button:has-text("Transfer")');

  // Select from account (first account)
  await page.selectOption('#transfer-from-account', { index: 1 });

  // Select to account (second account)
  await page.selectOption('#transfer-to-account', { index: 2 });

  // Fill amount
  await page.fill('#transfer-amount', '250.00');
  await page.fill('#transfer-date', '2025-10-30');
  await page.fill('#transfer-description', 'Moving savings');

  await page.click('button[type="submit"]:has-text("Create Transfer")');

  // Verify success
  await expect(page.locator('#toast')).toContainText('Transfer created successfully');

  // Verify transfer appears in transactions
  await page.click('text=All Transactions');
  await expect(page.locator('text=Moving savings')).toBeVisible();
  await expect(page.locator('text=Transfer:')).toBeVisible();
});
```

### Test 4.5: Transfer to Same Account Validation

```javascript
test('should prevent transfer to same account', async ({ page }) => {
  await page.goto('/');
  await page.click('button:has-text("Transfer")');

  // Select same account for both
  const firstAccount = await page.locator('#transfer-from-account option').nth(1).getAttribute('value');
  await page.selectOption('#transfer-from-account', firstAccount);
  await page.selectOption('#transfer-to-account', firstAccount);

  await page.fill('#transfer-amount', '100');
  await page.fill('#transfer-date', '2025-10-30');

  await page.click('button[type="submit"]:has-text("Create Transfer")');

  await expect(page.locator('#toast')).toContainText('Cannot transfer to the same account');
});
```

### Test 4.6: Transfer Updates Both Account Balances

```javascript
test('should update both account balances correctly', async ({ page }) => {
  await page.goto('/');
  await page.click('text=Accounts');

  // Get initial balances
  const accounts = await page.locator('#accounts-list > div').all();
  const fromAccountText = await accounts[0].textContent();
  const toAccountText = await accounts[1].textContent();

  const parseBalance = (text) => parseFloat(text.match(/\$([\d,]+\.\d{2})/)[1].replace(',', ''));
  const fromInitial = parseBalance(fromAccountText);
  const toInitial = parseBalance(toAccountText);

  // Create transfer
  await page.click('button:has-text("Transfer")');
  await page.selectOption('#transfer-from-account', { index: 1 });
  await page.selectOption('#transfer-to-account', { index: 2 });
  await page.fill('#transfer-amount', '100.00');
  await page.fill('#transfer-date', '2025-10-30');
  await page.click('button[type="submit"]:has-text("Create Transfer")');

  await expect(page.locator('#toast')).toContainText('successfully');

  // Reload accounts view
  await page.click('text=Budget');
  await page.click('text=Accounts');

  // Get new balances
  const accountsAfter = await page.locator('#accounts-list > div').all();
  const fromAfterText = await accountsAfter[0].textContent();
  const toAfterText = await accountsAfter[1].textContent();

  const fromFinal = parseBalance(fromAfterText);
  const toFinal = parseBalance(toAfterText);

  // Verify changes
  expect(fromFinal).toBeCloseTo(fromInitial - 100, 2);
  expect(toFinal).toBeCloseTo(toInitial + 100, 2);
});
```

---

## Test Suite 5: Credit Card Payment Tracking

### Test 5.1: Transfer to Credit Card Categorizes with Payment Category

```javascript
test('should categorize transfer to credit card with payment category', async ({ page }) => {
  await page.goto('/');

  // Get credit card payment category initial state
  await page.click('text=Budget');
  const paymentCard = page.locator('.bg-orange-50').filter({ hasText: 'Payment:' }).first();
  const initialSpent = await paymentCard.locator('text=Spent').locator('..').locator('.font-semibold').textContent();

  // Create transfer TO credit card
  await page.click('button:has-text("Transfer")');

  // Select checking as from, credit card as to
  const checkingOption = await page.locator('#transfer-from-account option').filter({ hasText: /checking/i }).first();
  const ccOption = await page.locator('#transfer-to-account option').filter({ hasText: /credit/i }).first();

  await page.selectOption('#transfer-from-account', await checkingOption.getAttribute('value'));
  await page.selectOption('#transfer-to-account', await ccOption.getAttribute('value'));

  await page.fill('#transfer-amount', '75.00');
  await page.fill('#transfer-date', '2025-10-30');
  await page.fill('#transfer-description', 'CC Payment');

  await page.click('button[type="submit"]:has-text("Create Transfer")');
  await expect(page.locator('#toast')).toContainText('successfully');

  // Verify payment category spent increased
  await page.click('text=Budget');
  const newSpent = await paymentCard.locator('text=Spent').locator('..').locator('.font-semibold').textContent();

  const initialAmount = parseFloat(initialSpent.replace(/[$,]/g, ''));
  const newAmount = parseFloat(newSpent.replace(/[$,]/g, ''));
  expect(newAmount).toBeCloseTo(initialAmount + 75, 2);
});
```

### Test 5.2: Payment Category Shows Available = $0 After Full Payment

```javascript
test('should show available $0 in payment category after paying full balance', async ({ page }) => {
  await page.goto('/');

  // First, make a CC purchase to have a balance
  await page.click('text=+ Add Transaction');
  const ccOption = await page.locator('#transaction-account option').filter({ hasText: /credit/i }).first();
  await page.selectOption('#transaction-account', await ccOption.getAttribute('value'));
  await page.selectOption('#transaction-category', { index: 1 });
  await page.selectOption('#transaction-type', 'outflow');
  await page.fill('#transaction-amount', '100.00');
  await page.fill('#transaction-date', '2025-10-30');
  await page.click('button[type="submit"]:has-text("Add Transaction")');

  // Get allocated amount
  await page.click('text=Budget');
  const paymentCard = page.locator('.bg-orange-50').filter({ hasText: 'Payment:' }).first();
  const allocated = await paymentCard.locator('text=Allocated').locator('..').locator('.font-semibold').textContent();
  const allocatedAmount = parseFloat(allocated.replace(/[$,]/g, ''));

  // Pay the full amount
  await page.click('button:has-text("Transfer")');
  const checkingOption = await page.locator('#transfer-from-account option').filter({ hasText: /checking/i }).first();
  const ccPaymentOption = await page.locator('#transfer-to-account option').filter({ hasText: /credit/i }).first();

  await page.selectOption('#transfer-from-account', await checkingOption.getAttribute('value'));
  await page.selectOption('#transfer-to-account', await ccPaymentOption.getAttribute('value'));
  await page.fill('#transfer-amount', allocatedAmount.toFixed(2));
  await page.fill('#transfer-date', '2025-10-30');
  await page.click('button[type="submit"]:has-text("Create Transfer")');

  // Verify available is $0
  await page.click('text=Budget');
  const available = await paymentCard.locator('text=Available').locator('..').locator('.font-bold').textContent();
  expect(available).toBe('$0.00');
});
```

---

## Test Suite 6: Integration Tests

### Test 6.1: Complete Credit Card Flow

```javascript
test('should complete full credit card lifecycle', async ({ page }) => {
  await page.goto('/');

  // Step 1: Create credit card
  await page.click('text=Accounts');
  await page.click('text=+ Add Account');
  await page.fill('#account-name', 'Integration Test CC');
  await page.selectOption('#account-type', 'credit');
  await page.fill('#account-balance', '0');
  await page.click('button[type="submit"]:has-text("Add Account")');

  // Step 2: Verify payment category created
  await page.click('text=Budget');
  await expect(page.locator('text=Payment: Integration Test CC')).toBeVisible();

  // Step 3: Make purchase on credit card
  await page.click('text=+ Add Transaction');
  const ccOption = await page.locator('#transaction-account option').filter({ hasText: 'Integration Test CC' }).first();
  await page.selectOption('#transaction-account', await ccOption.getAttribute('value'));
  await page.selectOption('#transaction-category', { index: 1 });
  await page.selectOption('#transaction-type', 'outflow');
  await page.fill('#transaction-amount', '150.00');
  await page.fill('#transaction-date', '2025-10-30');
  await page.fill('#transaction-description', 'Restaurant');
  await page.click('button[type="submit"]:has-text("Add Transaction")');

  // Step 4: Verify payment category allocated $150
  await page.click('text=Budget');
  const paymentCard = page.locator('.bg-orange-50').filter({ hasText: 'Payment: Integration Test CC' });
  await expect(paymentCard.locator('text=Allocated').locator('..').locator('.font-semibold')).toContainText('$150.00');
  await expect(paymentCard.locator('text=Spent').locator('..').locator('.font-semibold')).toContainText('$0.00');
  await expect(paymentCard.locator('text=Available')).toBeVisible();

  // Step 5: Pay credit card from checking
  await page.click('button:has-text("Transfer")');
  const checkingOption = await page.locator('#transfer-from-account option').filter({ hasText: /checking/i }).first();
  const ccPayOption = await page.locator('#transfer-to-account option').filter({ hasText: 'Integration Test CC' }).first();
  await page.selectOption('#transfer-from-account', await checkingOption.getAttribute('value'));
  await page.selectOption('#transfer-to-account', await ccPayOption.getAttribute('value'));
  await page.fill('#transfer-amount', '150.00');
  await page.fill('#transfer-date', '2025-10-30');
  await page.fill('#transfer-description', 'CC Payment');
  await page.click('button[type="submit"]:has-text("Create Transfer")');

  // Step 6: Verify payment category shows spent $150, available $0
  await page.click('text=Budget');
  await expect(paymentCard.locator('text=Spent').locator('..').locator('.font-semibold')).toContainText('$150.00');
  await expect(paymentCard.locator('text=Available').locator('..').locator('.font-bold')).toContainText('$0.00');

  // Step 7: Verify CC balance is $0
  await page.click('text=Accounts');
  const ccCard = page.locator('text=Integration Test CC').locator('..');
  await expect(ccCard.locator('.text-xl.font-bold')).toContainText('$0.00');
});
```

### Test 6.2: Multiple Categories with Mixed Income/Expense

```javascript
test('should handle multiple categories without type distinction', async ({ page }) => {
  await page.goto('/');

  // Create several categories (no type needed)
  const categories = [
    { name: 'Salary Income', color: '#10b981', description: 'Monthly salary' },
    { name: 'Rent', color: '#ef4444', description: 'Monthly rent' },
    { name: 'Groceries', color: '#10b981', description: 'Food shopping' }
  ];

  for (const cat of categories) {
    await page.click('text=Categories');
    await page.click('text=+ Add Category');
    await page.fill('#category-name', cat.name);
    await page.click(`.color-swatch[data-color="${cat.color}"]`);
    await page.fill('#category-description', cat.description);
    await page.click('button[type="submit"]:has-text("Add Category")');
    await expect(page.locator('#toast')).toContainText('successfully');
  }

  // Create income transaction with category
  await page.click('text=+ Add Transaction');
  await page.selectOption('#transaction-account', { index: 1 });
  await page.selectOption('#transaction-category', { label: 'Salary Income' });
  await page.selectOption('#transaction-type', 'inflow');
  await page.fill('#transaction-amount', '5000.00');
  await page.fill('#transaction-date', '2025-10-30');
  await page.click('button[type="submit"]:has-text("Add Transaction")');

  // Create expense transactions
  await page.click('text=+ Add Transaction');
  await page.selectOption('#transaction-account', { index: 1 });
  await page.selectOption('#transaction-category', { label: 'Rent' });
  await page.selectOption('#transaction-type', 'outflow');
  await page.fill('#transaction-amount', '1500.00');
  await page.fill('#transaction-date', '2025-10-30');
  await page.click('button[type="submit"]:has-text("Add Transaction")');

  // Verify all transactions appear correctly
  await page.click('text=All Transactions');
  await expect(page.locator('text=Salary Income').locator('..').locator('text=+$5,000.00')).toBeVisible();
  await expect(page.locator('text=Rent').locator('..').locator('text=-$1,500.00')).toBeVisible();
});
```

### Test 6.3: Payment Category Not in Management View

```javascript
test('should not show payment categories in categories management view', async ({ page }) => {
  await page.goto('/');
  await page.click('text=Categories');

  // Get all category cards
  const categoryCards = page.locator('#categories-list > div');
  const count = await categoryCards.count();

  // Verify none contain "Payment:" prefix
  for (let i = 0; i < count; i++) {
    const cardText = await categoryCards.nth(i).textContent();
    expect(cardText).not.toContain('Payment:');
    expect(cardText).not.toContain('(Auto-managed)');
  }
});
```

---

## Test Suite 7: Edge Cases

### Test 7.1: Transfer with Empty Description

```javascript
test('should allow transfer without description', async ({ page }) => {
  await page.goto('/');
  await page.click('button:has-text("Transfer")');

  await page.selectOption('#transfer-from-account', { index: 1 });
  await page.selectOption('#transfer-to-account', { index: 2 });
  await page.fill('#transfer-amount', '25.00');
  await page.fill('#transfer-date', '2025-10-30');
  // Leave description empty

  await page.click('button[type="submit"]:has-text("Create Transfer")');
  await expect(page.locator('#toast')).toContainText('successfully');
});
```

### Test 7.2: Large Transfer Amount

```javascript
test('should handle large transfer amounts', async ({ page }) => {
  await page.goto('/');
  await page.click('button:has-text("Transfer")');

  await page.selectOption('#transfer-from-account', { index: 1 });
  await page.selectOption('#transfer-to-account', { index: 2 });
  await page.fill('#transfer-amount', '999999.99');
  await page.fill('#transfer-date', '2025-10-30');

  await page.click('button[type="submit"]:has-text("Create Transfer")');
  await expect(page.locator('#toast')).toContainText('successfully');

  // Verify amount displays correctly
  await page.click('text=All Transactions');
  await expect(page.locator('text=$999,999.99')).toBeVisible();
});
```

### Test 7.3: Decimal Precision in Transfers

```javascript
test('should handle decimal precision correctly', async ({ page }) => {
  await page.goto('/');
  await page.click('button:has-text("Transfer")');

  await page.selectOption('#transfer-from-account', { index: 1 });
  await page.selectOption('#transfer-to-account', { index: 2 });
  await page.fill('#transfer-amount', '123.45');
  await page.fill('#transfer-date', '2025-10-30');

  await page.click('button[type="submit"]:has-text("Create Transfer")');

  await page.click('text=All Transactions');
  // Should show exactly $123.45, not $123.44 or $123.46
  await expect(page.locator('text=$123.45')).toBeVisible();
});
```

### Test 7.4: Category Color Persists After Creation

```javascript
test('should persist category color selection', async ({ page }) => {
  await page.goto('/');
  await page.click('text=Categories');

  // Create category with purple color
  await page.click('text=+ Add Category');
  await page.fill('#category-name', 'Test Purple Category');
  await page.click('.color-swatch[data-color="#a855f7"]');
  await page.click('button[type="submit"]:has-text("Add Category")');

  // Verify color appears in category list
  const categoryCard = page.locator('text=Test Purple Category').locator('..');
  const colorDot = categoryCard.locator('.rounded-full').first();

  const bgColor = await colorDot.evaluate((el) =>
    window.getComputedStyle(el).backgroundColor
  );

  // Purple color should be applied
  expect(bgColor).toBe('rgb(168, 85, 247)'); // #a855f7 in RGB
});
```

### Test 7.5: Cancel Modal Actions

```javascript
test('should cancel modal without saving data', async ({ page }) => {
  await page.goto('/');

  // Open transaction modal
  await page.click('text=+ Add Transaction');
  await page.selectOption('#transaction-account', { index: 1 });
  await page.fill('#transaction-amount', '100.00');

  // Click cancel
  await page.click('button:has-text("Cancel")');

  // Modal should close
  await expect(page.locator('#transaction-modal')).not.toHaveClass(/active/);

  // Reopen and verify data was not saved
  await page.click('text=+ Add Transaction');
  const amount = await page.inputValue('#transaction-amount');
  expect(amount).toBe('');
});
```

---

## Test Suite 8: Accessibility & UX

### Test 8.1: Keyboard Navigation in Modals

```javascript
test('should support keyboard navigation', async ({ page }) => {
  await page.goto('/');
  await page.click('text=+ Add Transaction');

  // Tab through form fields
  await page.keyboard.press('Tab'); // Focus account
  await page.keyboard.press('Tab'); // Focus category
  await page.keyboard.press('Tab'); // Focus type
  await page.keyboard.press('Tab'); // Focus amount

  // Verify amount field has focus
  const focusedElement = await page.evaluate(() => document.activeElement.id);
  expect(focusedElement).toBe('transaction-amount');

  // Escape should close modal
  await page.keyboard.press('Escape');
  await expect(page.locator('#transaction-modal')).not.toHaveClass(/active/);
});
```

### Test 8.2: Modal Click Outside to Close

```javascript
test('should close modal when clicking outside', async ({ page }) => {
  await page.goto('/');
  await page.click('text=+ Add Transaction');

  await expect(page.locator('#transaction-modal')).toHaveClass(/active/);

  // Click on modal backdrop
  await page.locator('#transaction-modal').click({ position: { x: 5, y: 5 } });

  await expect(page.locator('#transaction-modal')).not.toHaveClass(/active/);
});
```

### Test 8.3: Toast Notifications Auto-Dismiss

```javascript
test('should auto-dismiss toast after 3 seconds', async ({ page }) => {
  await page.goto('/');

  // Trigger a toast
  await page.click('text=Accounts');
  await page.click('text=+ Add Account');
  await page.fill('#account-name', 'Toast Test');
  await page.selectOption('#account-type', 'checking');
  await page.fill('#account-balance', '100');
  await page.click('button[type="submit"]:has-text("Add Account")');

  // Toast should be visible immediately
  await expect(page.locator('#toast.active')).toBeVisible();

  // Wait 3.5 seconds
  await page.waitForTimeout(3500);

  // Toast should be dismissed
  await expect(page.locator('#toast.active')).not.toBeVisible();
});
```

### Test 8.4: Form Validation Messages

```javascript
test('should show validation for required fields', async ({ page }) => {
  await page.goto('/');
  await page.click('text=+ Add Transaction');

  // Try to submit without filling required fields
  await page.click('button[type="submit"]:has-text("Add Transaction")');

  // Check HTML5 validation
  const accountField = page.locator('#transaction-account');
  const validationMessage = await accountField.evaluate((el) => el.validationMessage);
  expect(validationMessage).toBeTruthy();
});
```

---

## Running All Tests

### Run All Tests
```bash
npx playwright test
```

### Run Specific Test Suite
```bash
npx playwright test --grep "Category Management"
```

### Run Tests in UI Mode (Interactive)
```bash
npx playwright test --ui
```

### Run Tests with Trace
```bash
npx playwright test --trace on
```

### Generate HTML Report
```bash
npx playwright show-report
```

---

## Test Data Setup

### Seed Script (Optional)
Create `tests/seed.js` to set up test data:

```javascript
export async function seedTestData() {
  // Create checking account
  await fetch('http://localhost:8080/api/accounts', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({
      name: 'Test Checking',
      type: 'checking',
      balance: 500000 // $5000
    })
  });

  // Create credit card
  await fetch('http://localhost:8080/api/accounts', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({
      name: 'Test Credit Card',
      type: 'credit',
      balance: 0
    })
  });

  // Create category
  await fetch('http://localhost:8080/api/categories', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({
      name: 'Groceries',
      color: '#10b981',
      description: 'Food shopping'
    })
  });
}
```

### Use in Tests
```javascript
import { test, expect } from '@playwright/test';
import { seedTestData } from './seed.js';

test.beforeEach(async () => {
  await seedTestData();
});
```

---

## Expected Test Coverage

- ✅ Category creation without type field
- ✅ Predefined color palette (all 10 colors)
- ✅ Optional income categorization
- ✅ Dynamic category requirement toggle
- ✅ Credit card account creation
- ✅ Payment category auto-creation
- ✅ Payment category visual distinction
- ✅ Payment category filtering
- ✅ Transfer functionality
- ✅ Transfer validation
- ✅ Credit card payment tracking
- ✅ Account balance updates
- ✅ Complete integration flows
- ✅ Edge cases and error handling
- ✅ Accessibility and UX

---

## Troubleshooting

### Server Not Starting
```bash
# Check if port is in use
lsof -i :8080

# Kill process
kill -9 <PID>
```

### Database Issues
```bash
# Reset database
rm budget.db
./budget  # Will recreate with migrations
```

### Flaky Tests
- Add explicit waits: `await page.waitForSelector('text=...')`
- Use `waitForLoadState`: `await page.waitForLoadState('networkidle')`
- Increase timeout: `test.setTimeout(60000)`

---

## Summary

This comprehensive test suite covers all features implemented in this branch:

1. **Category Refactoring** - Removal of type field, unified category list
2. **Predefined Color Palette** - 10 semantic colors for categories
3. **Optional Income Categorization** - Dynamic requirement based on transaction type
4. **Credit Card Support** - Auto-payment categories, visual distinction
5. **Transfer Functionality** - Account-to-account transfers
6. **Payment Tracking** - Credit card payment properly reflected in budget
7. **Integration Flows** - Complete user workflows
8. **Edge Cases** - Validation, error handling, UX

Run these tests after any changes to ensure all features continue working correctly!
