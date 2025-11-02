import { test, expect } from '@playwright/test';

/**
 * E2E Tests: Allocate to Cover Underfunded Payment Categories
 *
 * Tests the "Allocate to Cover" button functionality for credit card
 * payment categories that are underfunded.
 *
 * Based on specification: docs/spec-remove-cc-sync.md
 */

// Helper functions
async function createAccount(page: any, name: string, type: string) {
  // Navigate to accounts page or use API
  const response = await page.request.post('/api/accounts', {
    data: {
      name: name,
      type: type
    }
  });
  return await response.json();
}

async function createCategoryGroup(page: any, name: string) {
  const response = await page.request.post('/api/category-groups', {
    data: {
      name: name,
      order: 1
    }
  });
  return await response.json();
}

async function createCategory(page: any, name: string, groupId: string) {
  const response = await page.request.post('/api/categories', {
    data: {
      name: name,
      description: '',
      color: '#3B82F6',
      group_id: groupId
    }
  });
  return await response.json();
}

async function createTransaction(page: any, accountId: string, categoryId: string, amount: number, description: string, date: string) {
  const response = await page.request.post('/api/transactions', {
    data: {
      account_id: accountId,
      category_id: categoryId,
      amount: amount,
      description: description,
      date: date
    }
  });
  return await response.json();
}

async function createAllocation(page: any, categoryId: string, amount: number, period: string) {
  const response = await page.request.post('/api/allocations', {
    data: {
      category_id: categoryId,
      amount: amount,
      period: period
    }
  });
  return await response.json();
}

async function getCurrentPeriod(): Promise<string> {
  const now = new Date();
  const year = now.getFullYear();
  const month = String(now.getMonth() + 1).padStart(2, '0');
  return `${year}-${month}`;
}

async function expandAllGroups(page: any) {
  // Click "Expand All" button to make category groups visible
  const expandButton = page.locator('button:has-text("Expand All")');
  if (await expandButton.isVisible()) {
    await expandButton.click();
    await page.waitForTimeout(500); // Wait for expand animation
  }
}

test.describe('Allocate to Cover - Underfunded Payment Categories', () => {
  let testGroupId: string;

  test.beforeEach(async ({ page }) => {
    // Navigate to application
    await page.goto('http://localhost:8080');

    // Wait for application to load by checking for the header and ready to assign box
    await page.waitForSelector('header', { timeout: 10000 });
    await page.waitForSelector('#ready-to-assign', { timeout: 10000 });
    await page.waitForLoadState('networkidle');

    // Create a shared category group for all tests
    const group = await createCategoryGroup(page, 'Test Categories');
    testGroupId = group.id;
  });

  test('TC-001: Display underfunded warning for payment category', async ({ page }) => {
    // Setup test data
    const period = await getCurrentPeriod();

    // Create checking account with funds
    const checking = await createAccount(page, 'Test Checking', 'checking');
    const incomeCategory = await createCategory(page, 'Salary', testGroupId);
    await createTransaction(page, checking.id, incomeCategory.id, 500000, 'Income', new Date().toISOString());

    // Create credit card account
    const creditCard = await createAccount(page, 'Test Credit Card', 'checking');

    // Find payment category (auto-created)
    const categoriesResponse = await page.request.get('/api/categories');
    const categories = await categoriesResponse.json();
    const paymentCategory = categories.find((c: any) => c.name.includes('Payment'));

    // Create credit card spending
    const groceries = await createCategory(page, 'Groceries', testGroupId);
    await createTransaction(page, creditCard.id, groceries.id, -20000, 'Groceries', new Date().toISOString());

    // Create partial allocation (creates underfunded state)
    await createAllocation(page, paymentCategory.id, 10000, period);

    // Navigate to Budget page
    await page.click('text=Budget');
    await expandAllGroups(page);
    await page.waitForLoadState('networkidle');

    // Verify underfunded warning appears
    const underfundedWarning = page.locator('text=⚠️ Underfunded');
    await expect(underfundedWarning).toBeVisible();

    // Verify amount is displayed
    await expect(page.locator('text=/Need \\$100\\.00 more/')).toBeVisible();

    // Verify warning is red
    const warningElement = page.locator('.text-red-600, .text-red-400').filter({ hasText: 'Underfunded' });
    await expect(warningElement).toBeVisible();

    // Take screenshot
    await page.screenshot({ path: 'test-results/underfunded-warning-display.png' });
  });

  test('TC-002: Display "Allocate to Cover" button for underfunded category', async ({ page }) => {
    // Setup underfunded state (similar to TC-001)
    const period = await getCurrentPeriod();
    const checking = await createAccount(page, 'Test Checking', 'checking');
    const incomeCategory = await createCategory(page, 'Salary', testGroupId);
    await createTransaction(page, checking.id, incomeCategory.id, 500000, 'Income', new Date().toISOString());

    const creditCard = await createAccount(page, 'Test Credit Card', 'checking');
    const categoriesResponse = await page.request.get('/api/categories');
    const categories = await categoriesResponse.json();
    const paymentCategory = categories.find((c: any) => c.name.includes('Payment'));

    const groceries = await createCategory(page, 'Groceries', testGroupId);
    await createTransaction(page, creditCard.id, groceries.id, -20000, 'Groceries', new Date().toISOString());
    await createAllocation(page, paymentCategory.id, 10000, period);

    // Navigate to Budget page
    await page.click('text=Budget');
    await expandAllGroups(page);
    await page.waitForLoadState('networkidle');

    // Verify "Allocate to Cover" button appears
    const allocateButton = page.locator('button:has-text("Allocate to Cover")');
    await expect(allocateButton).toBeVisible();

    // Verify button styling
    await expect(allocateButton).toHaveClass(/bg-blue-600/);
    await expect(allocateButton).toHaveClass(/text-white/);

    // Verify button is enabled
    await expect(allocateButton).toBeEnabled();

    // Verify tooltip
    const title = await allocateButton.getAttribute('title');
    expect(title).toContain('Allocate from Ready to Assign');

    // Hover and verify hover state
    await allocateButton.hover();
    await page.screenshot({ path: 'test-results/allocate-button-hover.png' });
  });

  test('TC-003: Successfully allocate to cover underfunded amount', async ({ page }) => {
    // Setup underfunded state
    const period = await getCurrentPeriod();
    const checking = await createAccount(page, 'Test Checking', 'checking');
    const incomeCategory = await createCategory(page, 'Salary', testGroupId);
    await createTransaction(page, checking.id, incomeCategory.id, 500000, 'Income', new Date().toISOString());

    const creditCard = await createAccount(page, 'Test Credit Card', 'checking');
    const categoriesResponse = await page.request.get('/api/categories');
    const categories = await categoriesResponse.json();
    const paymentCategory = categories.find((c: any) => c.name.includes('Payment'));

    const groceries = await createCategory(page, 'Groceries', testGroupId);
    await createTransaction(page, creditCard.id, groceries.id, -20000, 'Groceries', new Date().toISOString());
    await createAllocation(page, paymentCategory.id, 10000, period);

    // Navigate to Budget page
    await page.click('text=Budget');
    await expandAllGroups(page);
    await page.waitForLoadState('networkidle');

    // Get initial Ready to Assign value
    const rtaElement = page.locator('#ready-to-assign');
    const initialRTA = await rtaElement.textContent();

    // Take screenshot before
    await page.screenshot({ path: 'test-results/before-allocation.png' });

    // Click "Allocate to Cover" button
    const allocateButton = page.locator('button:has-text("Allocate to Cover")');
    await allocateButton.click();

    // Verify loading state
    await expect(page.locator('button:has-text("Allocating...")')).toBeVisible();
    await expect(allocateButton).toBeDisabled();

    // Wait for success
    await page.waitForResponse(response =>
      response.url().includes('/api/allocations/cover-underfunded') && response.status() === 201
    );

    // Verify success toast
    await expect(page.locator('text=/Successfully allocated/')).toBeVisible();

    // Take screenshot during success
    await page.screenshot({ path: 'test-results/allocation-success-toast.png' });

    // Wait for UI to refresh
    await page.waitForLoadState('networkidle');

    // Verify underfunded warning is gone
    await expect(page.locator('text=⚠️ Underfunded')).not.toBeVisible();

    // Verify Ready to Assign decreased by $100
    const newRTA = await rtaElement.textContent();
    // Note: Would need to parse and compare actual values

    // Take screenshot after
    await page.screenshot({ path: 'test-results/after-allocation-success.png' });
  });

  test('TC-004: Error handling for insufficient funds', async ({ page }) => {
    // Setup: Create underfunded state where RTA < underfunded
    const period = await getCurrentPeriod();
    const checking = await createAccount(page, 'Test Checking', 'checking');
    const incomeCategory = await createCategory(page, 'Salary', testGroupId);

    // Only $200 income
    await createTransaction(page, checking.id, incomeCategory.id, 20000, 'Income', new Date().toISOString());

    const creditCard = await createAccount(page, 'Test Credit Card', 'checking');
    const categoriesResponse = await page.request.get('/api/categories');
    const categories = await categoriesResponse.json();
    const paymentCategory = categories.find((c: any) => c.name.includes('Payment'));

    // Create $500 debt
    const groceries = await createCategory(page, 'Groceries', testGroupId);
    await createTransaction(page, creditCard.id, groceries.id, -50000, 'Groceries', new Date().toISOString());

    // Allocate to other categories to reduce RTA to $100
    const gas = await createCategory(page, 'Gas', testGroupId);
    await createAllocation(page, gas.id, 10000, period);

    // Navigate to Budget page
    await page.click('text=Budget');
    await expandAllGroups(page);
    await page.waitForLoadState('networkidle');

    // Verify RTA is less than underfunded
    // RTA should be ~$100, underfunded should be ~$500

    // Click "Allocate to Cover" button
    const allocateButton = page.locator('button:has-text("Allocate to Cover")');
    await allocateButton.click();

    // Wait for error response
    await page.waitForResponse(response =>
      response.url().includes('/api/allocations/cover-underfunded') && response.status() === 400
    );

    // Verify error toast appears
    const errorToast = page.locator('text=/Insufficient funds/');
    await expect(errorToast).toBeVisible();

    // Verify error message shows amounts
    await expect(page.locator('text=/Ready to Assign.*Underfunded/')).toBeVisible();

    // Verify button returns to normal state
    await expect(page.locator('button:has-text("Allocate to Cover")')).toBeVisible();
    await expect(allocateButton).toBeEnabled();

    // Verify underfunded warning still visible
    await expect(page.locator('text=⚠️ Underfunded')).toBeVisible();

    // Take screenshot
    await page.screenshot({ path: 'test-results/insufficient-funds-error.png' });
  });

  test('TC-005: Handle multiple underfunded payment categories', async ({ page }) => {
    // Setup: Create two credit cards, both underfunded
    const period = await getCurrentPeriod();
    const checking = await createAccount(page, 'Test Checking', 'checking');
    const incomeCategory = await createCategory(page, 'Salary', testGroupId);
    await createTransaction(page, checking.id, incomeCategory.id, 1000000, 'Income', new Date().toISOString());

    // Credit Card 1
    const creditCard1 = await createAccount(page, 'Test Credit Card 1', 'checking');
    const groceries = await createCategory(page, 'Groceries', testGroupId);
    await createTransaction(page, creditCard1.id, groceries.id, -30000, 'Groceries', new Date().toISOString());

    // Credit Card 2
    const creditCard2 = await createAccount(page, 'Test Credit Card 2', 'checking');
    const gas = await createCategory(page, 'Gas', testGroupId);
    await createTransaction(page, creditCard2.id, gas.id, -40000, 'Gas', new Date().toISOString());

    // Get payment categories
    const categoriesResponse = await page.request.get('/api/categories');
    const categories = await categoriesResponse.json();
    const paymentCategories = categories.filter((c: any) => c.name.includes('Payment'));

    // Create partial allocations (underfunded state)
    await createAllocation(page, paymentCategories[0].id, 10000, period);
    await createAllocation(page, paymentCategories[1].id, 15000, period);

    // Navigate to Budget page
    await page.click('text=Budget');
    await expandAllGroups(page);
    await page.waitForLoadState('networkidle');

    // Verify both have underfunded warnings
    const underfundedWarnings = page.locator('text=⚠️ Underfunded');
    await expect(underfundedWarnings).toHaveCount(2);

    // Verify both have "Allocate to Cover" buttons
    const allocateButtons = page.locator('button:has-text("Allocate to Cover")');
    await expect(allocateButtons).toHaveCount(2);

    // Click first button
    await allocateButtons.first().click();
    await page.waitForResponse(response =>
      response.url().includes('/api/allocations/cover-underfunded') && response.status() === 201
    );

    // Wait for UI refresh
    await page.waitForLoadState('networkidle');

    // Verify first category no longer underfunded
    await expect(underfundedWarnings).toHaveCount(1);

    // Click second button
    await allocateButtons.first().click(); // Now first in the remaining list
    await page.waitForResponse(response =>
      response.url().includes('/api/allocations/cover-underfunded') && response.status() === 201
    );

    // Wait for UI refresh
    await page.waitForLoadState('networkidle');

    // Verify no underfunded warnings remain
    await expect(underfundedWarnings).toHaveCount(0);

    // Take screenshot
    await page.screenshot({ path: 'test-results/multiple-categories-covered.png' });
  });

  test('TC-006: Verify contributing categories are displayed', async ({ page }) => {
    // Setup: Create credit card with multiple spending categories
    const period = await getCurrentPeriod();
    const checking = await createAccount(page, 'Test Checking', 'checking');
    const incomeCategory = await createCategory(page, 'Salary', testGroupId);
    await createTransaction(page, checking.id, incomeCategory.id, 500000, 'Income', new Date().toISOString());

    const creditCard = await createAccount(page, 'Test Credit Card', 'checking');

    // Create multiple expense categories and transactions
    const groceries = await createCategory(page, 'Groceries', testGroupId);
    const gas = await createCategory(page, 'Gas', testGroupId);
    const dining = await createCategory(page, 'Dining', testGroupId);

    await createTransaction(page, creditCard.id, groceries.id, -15000, 'Groceries', new Date().toISOString());
    await createTransaction(page, creditCard.id, gas.id, -10000, 'Gas', new Date().toISOString());
    await createTransaction(page, creditCard.id, dining.id, -12500, 'Dining', new Date().toISOString());

    // Get payment category
    const categoriesResponse = await page.request.get('/api/categories');
    const categories = await categoriesResponse.json();
    const paymentCategory = categories.find((c: any) => c.name.includes('Payment'));

    // Create partial allocation
    await createAllocation(page, paymentCategory.id, 10000, period);

    // Navigate to Budget page
    await page.click('text=Budget');
    await expandAllGroups(page);
    await page.waitForLoadState('networkidle');

    // Verify contributing categories text appears
    const contributingText = page.locator('text=/Contributing categories:/');
    await expect(contributingText).toBeVisible();

    // Verify categories are listed
    await expect(page.locator('text=/Groceries.*Gas.*Dining/')).toBeVisible();

    // Verify text is red and small
    const contributingElement = page.locator('.text-red-500, .text-red-400').filter({ hasText: 'Contributing categories' });
    await expect(contributingElement).toBeVisible();

    // Take screenshot
    await page.screenshot({ path: 'test-results/contributing-categories-display.png' });
  });

  test('TC-007: Verify button does not appear for non-payment categories', async ({ page }) => {
    // Setup: Create regular expense category with allocation
    const period = await getCurrentPeriod();
    const groceries = await createCategory(page, 'Groceries', testGroupId);
    await createAllocation(page, groceries.id, 50000, period);

    // Navigate to Budget page
    await page.click('text=Budget');
    await expandAllGroups(page);
    await page.waitForLoadState('networkidle');

    // Verify "Allocate to Cover" button does NOT appear for regular categories
    const groceriesRow = page.locator('text=Groceries').locator('..');
    const allocateButton = groceriesRow.locator('button:has-text("Allocate to Cover")');
    await expect(allocateButton).not.toBeVisible();

    // Take screenshot
    await page.screenshot({ path: 'test-results/no-button-regular-category.png' });
  });

  test('TC-008: Verify no warning when payment category is fully funded', async ({ page }) => {
    // Setup: Create credit card with exactly matching allocation
    const period = await getCurrentPeriod();
    const checking = await createAccount(page, 'Test Checking', 'checking');
    const incomeCategory = await createCategory(page, 'Salary', testGroupId);
    await createTransaction(page, checking.id, incomeCategory.id, 500000, 'Income', new Date().toISOString());

    const creditCard = await createAccount(page, 'Test Credit Card', 'checking');
    const groceries = await createCategory(page, 'Groceries', testGroupId);
    await createTransaction(page, creditCard.id, groceries.id, -30000, 'Groceries', new Date().toISOString());

    // Get payment category
    const categoriesResponse = await page.request.get('/api/categories');
    const categories = await categoriesResponse.json();
    const paymentCategory = categories.find((c: any) => c.name.includes('Payment'));

    // Create FULL allocation (no underfunded)
    await createAllocation(page, paymentCategory.id, 30000, period);

    // Navigate to Budget page
    await page.click('text=Budget');
    await expandAllGroups(page);
    await page.waitForLoadState('networkidle');

    // Verify NO underfunded warning
    await expect(page.locator('text=⚠️ Underfunded')).not.toBeVisible();

    // Verify NO "Allocate to Cover" button
    const paymentRow = page.locator(`text=${paymentCategory.name}`).locator('..');
    await expect(paymentRow.locator('button:has-text("Allocate to Cover")')).not.toBeVisible();

    // Take screenshot
    await page.screenshot({ path: 'test-results/no-warning-fully-funded.png' });
  });

  test('TC-009: Verify double-click prevention', async ({ page }) => {
    // Setup underfunded state
    const period = await getCurrentPeriod();
    const checking = await createAccount(page, 'Test Checking', 'checking');
    const incomeCategory = await createCategory(page, 'Salary', testGroupId);
    await createTransaction(page, checking.id, incomeCategory.id, 500000, 'Income', new Date().toISOString());

    const creditCard = await createAccount(page, 'Test Credit Card', 'checking');
    const categoriesResponse = await page.request.get('/api/categories');
    const categories = await categoriesResponse.json();
    const paymentCategory = categories.find((c: any) => c.name.includes('Payment'));

    const groceries = await createCategory(page, 'Groceries', testGroupId);
    await createTransaction(page, creditCard.id, groceries.id, -20000, 'Groceries', new Date().toISOString());
    await createAllocation(page, paymentCategory.id, 10000, period);

    // Navigate to Budget page
    await page.click('text=Budget');
    await expandAllGroups(page);
    await page.waitForLoadState('networkidle');

    // Track API calls
    let apiCallCount = 0;
    page.on('response', response => {
      if (response.url().includes('/api/allocations/cover-underfunded')) {
        apiCallCount++;
      }
    });

    // Click button multiple times rapidly
    const allocateButton = page.locator('button:has-text("Allocate to Cover")');
    await allocateButton.click();
    await allocateButton.click(); // Second click should be prevented
    await allocateButton.click(); // Third click should be prevented

    // Wait for response
    await page.waitForTimeout(2000);

    // Verify only one API call was made
    expect(apiCallCount).toBe(1);
  });

  test('TC-010: Verify Ready to Assign accounting for underfunded', async ({ page }) => {
    // Setup: Controlled scenario to verify RTA formula
    const period = await getCurrentPeriod();

    // Create income: $5,000
    const checking = await createAccount(page, 'Test Checking', 'checking');
    const incomeCategory = await createCategory(page, 'Salary', testGroupId);
    await createTransaction(page, checking.id, incomeCategory.id, 500000, 'Income', new Date().toISOString());

    // Create regular allocation: $1,500
    const groceries = await createCategory(page, 'Groceries', testGroupId);
    await createAllocation(page, groceries.id, 150000, period);

    // Create credit card debt: $500
    const creditCard = await createAccount(page, 'Test Credit Card', 'checking');
    const gas = await createCategory(page, 'Gas', testGroupId);
    await createTransaction(page, creditCard.id, gas.id, -50000, 'Gas', new Date().toISOString());

    // Payment category allocation: $300 (underfunded by $200)
    const categoriesResponse = await page.request.get('/api/categories');
    const categories = await categoriesResponse.json();
    const paymentCategory = categories.find((c: any) => c.name.includes('Payment'));
    await createAllocation(page, paymentCategory.id, 30000, period);

    // Navigate to Budget page
    await page.click('text=Budget');
    await expandAllGroups(page);
    await page.waitForLoadState('networkidle');

    // Get RTA value
    // Expected: $5,000 - $1,500 - $200 (underfunded) = $3,300
    const rtaElement = page.locator('#ready-to-assign');
    const rtaText = await rtaElement.textContent();
    expect(rtaText).toContain('$3,300.00');

    // Click "Allocate to Cover" to cover $200 underfunded
    await page.locator('button:has-text("Allocate to Cover")').click();
    await page.waitForResponse(response =>
      response.url().includes('/api/allocations/cover-underfunded') && response.status() === 201
    );

    // Wait for refresh
    await page.waitForLoadState('networkidle');

    // Get new RTA value
    // Expected: $5,000 - $1,500 - $200 (now in allocations) - $0 (underfunded) = $3,300
    // RTA should be UNCHANGED (underfunded moved to allocation)
    const newRtaText = await rtaElement.textContent();
    expect(newRtaText).toContain('$3,300.00');
  });
});
