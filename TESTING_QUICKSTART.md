# Testing Quick Start Guide

## Quick Reference: How to Test the "Allocate to Cover" Feature

### Option 1: Automated Tests (Fastest)

```bash
# Install Playwright (first time only)
npm install -D @playwright/test

# Install browsers (first time only)
npx playwright install

# Start the application
docker-compose up -d

# Run all "Allocate to Cover" tests
npx playwright test tests/e2e/allocate-to-cover.spec.ts

# Run in headed mode (see browser)
npx playwright test tests/e2e/allocate-to-cover.spec.ts --headed

# Run specific test
npx playwright test -g "successfully allocate"

# Generate HTML report
npx playwright test --reporter=html
npx playwright show-report
```

### Option 2: Manual Testing (Most Visual)

```bash
# 1. Start the application
docker-compose up -d

# 2. Open browser
open http://localhost:8080

# 3. Create test data:
#    - Add checking account with $5,000 income
#    - Add credit card account
#    - Add credit card transaction: -$200 (creates debt)
#    - Payment category auto-created
#    - Allocate $100 to payment category (creates $100 underfunded)

# 4. Navigate to Budget tab

# 5. Verify underfunded warning:
#    - Warning: "⚠️ Underfunded - Need $100.00 more"
#    - Button: "Allocate to Cover" (blue)

# 6. Click "Allocate to Cover" button

# 7. Verify success:
#    - Toast: "Successfully allocated $100.00..."
#    - Underfunded warning removed
#    - Ready to Assign decreased by $100
```

### Option 3: Interactive Testing with Playwright MCP

If you have Playwright MCP access, use the detailed guide:
- **File:** `/Users/billybuffum/development/budget/docs/ui-test-execution-guide.md`

Example command:
```
Use Playwright MCP to test the "Allocate to Cover" button:

1. Navigate to http://localhost:8080
2. Create checking account with $5,000 income
3. Create credit card with -$200 debt
4. Allocate $100 to payment category
5. Navigate to Budget tab
6. Verify underfunded warning appears
7. Click "Allocate to Cover" button
8. Verify success toast
9. Verify underfunded warning removed
10. Take screenshots
```

## Test Files

- **Test Plan:** `docs/ui-test-plan-allocate-to-cover.md` (14 scenarios)
- **Execution Guide:** `docs/ui-test-execution-guide.md` (detailed steps)
- **Automated Tests:** `tests/e2e/allocate-to-cover.spec.ts` (10 test cases)
- **Summary:** `docs/ui-test-summary-allocate-to-cover.md`

## Common Issues

### Application Won't Start
```bash
# Check if port 8080 is in use
lsof -i :8080

# Restart Docker
docker-compose down
docker-compose up -d

# Check logs
docker-compose logs -f
```

### Tests Fail with "Cannot connect"
```bash
# Ensure application is running
curl http://localhost:8080/health

# Should return: {"status": "healthy"}
```

### Playwright Not Installed
```bash
npm install -D @playwright/test
npx playwright install chromium
```

## Quick Test Checklist

Minimum tests to verify feature works:

- [ ] Underfunded warning displays with correct amount
- [ ] "Allocate to Cover" button appears for underfunded payment categories
- [ ] Clicking button creates allocation successfully
- [ ] Success toast appears
- [ ] Underfunded warning removed after allocation
- [ ] Ready to Assign decreases by allocated amount
- [ ] Error message appears when insufficient funds
- [ ] No button for regular (non-payment) categories

## Success Criteria

Feature is ready for production when:
- ✅ All automated tests pass
- ✅ Manual testing shows intuitive UX
- ✅ Error handling is clear and helpful
- ✅ No JavaScript console errors
- ✅ Works on desktop and mobile
- ✅ Performance is acceptable (< 500ms API response)

## Need More Details?

See the comprehensive guides:
- **Test Plan:** `docs/ui-test-plan-allocate-to-cover.md`
- **Execution Guide:** `docs/ui-test-execution-guide.md`
- **Summary:** `docs/ui-test-summary-allocate-to-cover.md`
