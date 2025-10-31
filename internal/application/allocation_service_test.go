package application

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/billybbuffum/budget/internal/domain"
)

// Mock repositories for testing

type mockAllocationRepository struct {
	allocations        []*domain.Allocation
	createErr          error
	getByIDErr         error
	getByCategoryErr   error
	listErr            error
	updateErr          error
	deleteErr          error
	nextID             int
}

func (m *mockAllocationRepository) Create(ctx context.Context, allocation *domain.Allocation) error {
	if m.createErr != nil {
		return m.createErr
	}
	if allocation.ID == "" {
		m.nextID++
		allocation.ID = string(rune('a' + m.nextID))
	}
	m.allocations = append(m.allocations, allocation)
	return nil
}

func (m *mockAllocationRepository) GetByID(ctx context.Context, id string) (*domain.Allocation, error) {
	if m.getByIDErr != nil {
		return nil, m.getByIDErr
	}
	for _, a := range m.allocations {
		if a.ID == id {
			return a, nil
		}
	}
	return nil, errors.New("allocation not found")
}

func (m *mockAllocationRepository) GetByCategoryAndPeriod(ctx context.Context, categoryID, period string) (*domain.Allocation, error) {
	if m.getByCategoryErr != nil {
		return nil, m.getByCategoryErr
	}
	for _, a := range m.allocations {
		if a.CategoryID == categoryID && a.Period == period {
			return a, nil
		}
	}
	return nil, errors.New("allocation not found")
}

func (m *mockAllocationRepository) ListByPeriod(ctx context.Context, period string) ([]*domain.Allocation, error) {
	if m.listErr != nil {
		return nil, m.listErr
	}
	var result []*domain.Allocation
	for _, a := range m.allocations {
		if a.Period == period {
			result = append(result, a)
		}
	}
	return result, nil
}

func (m *mockAllocationRepository) List(ctx context.Context) ([]*domain.Allocation, error) {
	if m.listErr != nil {
		return nil, m.listErr
	}
	return m.allocations, nil
}

func (m *mockAllocationRepository) Update(ctx context.Context, allocation *domain.Allocation) error {
	if m.updateErr != nil {
		return m.updateErr
	}
	for i, a := range m.allocations {
		if a.ID == allocation.ID {
			m.allocations[i] = allocation
			return nil
		}
	}
	return errors.New("allocation not found")
}

func (m *mockAllocationRepository) Delete(ctx context.Context, id string) error {
	if m.deleteErr != nil {
		return m.deleteErr
	}
	for i, a := range m.allocations {
		if a.ID == id {
			m.allocations = append(m.allocations[:i], m.allocations[i+1:]...)
			return nil
		}
	}
	return errors.New("allocation not found")
}

type mockCategoryRepository struct {
	categories []*domain.Category
	getByIDErr error
	listErr    error
}

func (m *mockCategoryRepository) Create(ctx context.Context, category *domain.Category) error {
	m.categories = append(m.categories, category)
	return nil
}

func (m *mockCategoryRepository) GetByID(ctx context.Context, id string) (*domain.Category, error) {
	if m.getByIDErr != nil {
		return nil, m.getByIDErr
	}
	for _, c := range m.categories {
		if c.ID == id {
			return c, nil
		}
	}
	return nil, errors.New("category not found")
}

func (m *mockCategoryRepository) GetPaymentCategoryByAccountID(ctx context.Context, accountID string) (*domain.Category, error) {
	for _, c := range m.categories {
		if c.PaymentForAccountID != nil && *c.PaymentForAccountID == accountID {
			return c, nil
		}
	}
	return nil, errors.New("payment category not found")
}

func (m *mockCategoryRepository) List(ctx context.Context) ([]*domain.Category, error) {
	if m.listErr != nil {
		return nil, m.listErr
	}
	return m.categories, nil
}

func (m *mockCategoryRepository) ListByGroup(ctx context.Context, groupID string) ([]*domain.Category, error) {
	var result []*domain.Category
	for _, c := range m.categories {
		if c.GroupID != nil && *c.GroupID == groupID {
			result = append(result, c)
		}
	}
	return result, nil
}

func (m *mockCategoryRepository) Update(ctx context.Context, category *domain.Category) error {
	for i, c := range m.categories {
		if c.ID == category.ID {
			m.categories[i] = category
			return nil
		}
	}
	return errors.New("category not found")
}

func (m *mockCategoryRepository) Delete(ctx context.Context, id string) error {
	for i, c := range m.categories {
		if c.ID == id {
			m.categories = append(m.categories[:i], m.categories[i+1:]...)
			return nil
		}
	}
	return errors.New("category not found")
}

type mockTransactionRepository struct {
	transactions []*domain.Transaction
	listErr      error
}

func (m *mockTransactionRepository) Create(ctx context.Context, transaction *domain.Transaction) error {
	m.transactions = append(m.transactions, transaction)
	return nil
}

func (m *mockTransactionRepository) GetByID(ctx context.Context, id string) (*domain.Transaction, error) {
	for _, t := range m.transactions {
		if t.ID == id {
			return t, nil
		}
	}
	return nil, errors.New("transaction not found")
}

func (m *mockTransactionRepository) List(ctx context.Context) ([]*domain.Transaction, error) {
	if m.listErr != nil {
		return nil, m.listErr
	}
	return m.transactions, nil
}

func (m *mockTransactionRepository) ListByAccount(ctx context.Context, accountID string) ([]*domain.Transaction, error) {
	var result []*domain.Transaction
	for _, t := range m.transactions {
		if t.AccountID == accountID {
			result = append(result, t)
		}
	}
	return result, nil
}

func (m *mockTransactionRepository) ListByCategory(ctx context.Context, categoryID string) ([]*domain.Transaction, error) {
	var result []*domain.Transaction
	for _, t := range m.transactions {
		if t.CategoryID != nil && *t.CategoryID == categoryID {
			result = append(result, t)
		}
	}
	return result, nil
}

func (m *mockTransactionRepository) ListByPeriod(ctx context.Context, startDate, endDate string) ([]*domain.Transaction, error) {
	return nil, nil
}

func (m *mockTransactionRepository) ListUncategorized(ctx context.Context) ([]*domain.Transaction, error) {
	var result []*domain.Transaction
	for _, t := range m.transactions {
		if t.CategoryID == nil {
			result = append(result, t)
		}
	}
	return result, nil
}

func (m *mockTransactionRepository) GetCategoryActivity(ctx context.Context, categoryID, period string) (int64, error) {
	var total int64
	for _, t := range m.transactions {
		if t.CategoryID != nil && *t.CategoryID == categoryID {
			txnPeriod := t.Date.Format("2006-01")
			if txnPeriod == period {
				total += t.Amount
			}
		}
	}
	return total, nil
}

func (m *mockTransactionRepository) FindDuplicate(ctx context.Context, accountID string, date time.Time, amount int64, description string) (*domain.Transaction, error) {
	return nil, nil
}

func (m *mockTransactionRepository) FindByFitID(ctx context.Context, accountID string, fitID string) (*domain.Transaction, error) {
	return nil, nil
}

func (m *mockTransactionRepository) Update(ctx context.Context, transaction *domain.Transaction) error {
	return nil
}

func (m *mockTransactionRepository) BulkUpdateCategory(ctx context.Context, transactionIDs []string, categoryID *string) error {
	return nil
}

func (m *mockTransactionRepository) Delete(ctx context.Context, id string) error {
	return nil
}

type mockBudgetStateRepository struct {
	state *domain.BudgetState
}

func (m *mockBudgetStateRepository) Get(ctx context.Context) (*domain.BudgetState, error) {
	if m.state == nil {
		m.state = &domain.BudgetState{ReadyToAssign: 0}
	}
	return m.state, nil
}

func (m *mockBudgetStateRepository) Update(ctx context.Context, state *domain.BudgetState) error {
	m.state = state
	return nil
}

func (m *mockBudgetStateRepository) AdjustReadyToAssign(ctx context.Context, delta int64) error {
	if m.state == nil {
		m.state = &domain.BudgetState{ReadyToAssign: 0}
	}
	m.state.ReadyToAssign += delta
	return nil
}

type mockAccountRepository struct {
	accounts []*domain.Account
}

func (m *mockAccountRepository) Create(ctx context.Context, account *domain.Account) error {
	m.accounts = append(m.accounts, account)
	return nil
}

func (m *mockAccountRepository) GetByID(ctx context.Context, id string) (*domain.Account, error) {
	for _, a := range m.accounts {
		if a.ID == id {
			return a, nil
		}
	}
	return nil, errors.New("account not found")
}

func (m *mockAccountRepository) List(ctx context.Context) ([]*domain.Account, error) {
	return m.accounts, nil
}

func (m *mockAccountRepository) Update(ctx context.Context, account *domain.Account) error {
	return nil
}

func (m *mockAccountRepository) Delete(ctx context.Context, id string) error {
	return nil
}

func (m *mockAccountRepository) GetTotalBalance(ctx context.Context) (int64, error) {
	var total int64
	for _, a := range m.accounts {
		total += a.Balance
	}
	return total, nil
}

// Helper function to create a pointer to a string
func stringPtr(s string) *string {
	return &s
}

func int64Ptr(i int64) *int64 {
	return &i
}

// Tests for AllocateToCoverUnderfunded

func TestAllocationService_AllocateToCoverUnderfunded_Success(t *testing.T) {
	// Setup
	ctx := context.Background()
	period := "2025-10"

	accountID := "cc-account-1"
	paymentCategoryID := "payment-cat-1"
	normalCategoryID := "normal-cat-1"

	// Create mock repositories
	allocRepo := &mockAllocationRepository{
		allocations: []*domain.Allocation{
			{
				ID:         "alloc-1",
				CategoryID: normalCategoryID,
				Period:     period,
				Amount:     50000, // $500 allocated to normal category
			},
		},
	}

	catRepo := &mockCategoryRepository{
		categories: []*domain.Category{
			{
				ID:                  paymentCategoryID,
				Name:                "Credit Card Payment",
				PaymentForAccountID: stringPtr(accountID),
			},
			{
				ID:   normalCategoryID,
				Name: "Groceries",
			},
		},
	}

	// Credit card has -$300 balance (we owe $300)
	// Payment category has no allocations yet
	// So underfunded = $300 (amount owed) - $0 (available) = $300
	txnRepo := &mockTransactionRepository{
		transactions: []*domain.Transaction{
			// Income
			{
				ID:        "txn-income-1",
				AccountID: "checking",
				Amount:    200000, // $2000 income
				Date:      time.Date(2025, 10, 1, 0, 0, 0, 0, time.UTC),
				Type:      "income",
			},
			// Credit card spending
			{
				ID:         "txn-cc-1",
				AccountID:  accountID,
				Amount:     -50000, // $500 spent on card
				Date:       time.Date(2025, 10, 15, 0, 0, 0, 0, time.UTC),
				CategoryID: stringPtr("cat-food"),
			},
		},
	}

	accountRepo := &mockAccountRepository{
		accounts: []*domain.Account{
			{
				ID:      accountID,
				Name:    "Credit Card",
				Type:    "credit_card",
				Balance: -30000, // Owe $300
			},
			{
				ID:      "checking",
				Name:    "Checking",
				Type:    "checking",
				Balance: 200000, // $2000
			},
		},
	}

	budgetRepo := &mockBudgetStateRepository{}

	service := NewAllocationService(allocRepo, catRepo, txnRepo, budgetRepo, accountRepo)

	// Execute
	allocation, err := service.AllocateToCoverUnderfunded(ctx, paymentCategoryID, period)

	// Verify
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if allocation == nil {
		t.Fatal("Expected allocation to be created")
	}

	if allocation.CategoryID != paymentCategoryID {
		t.Errorf("Expected category ID %s, got %s", paymentCategoryID, allocation.CategoryID)
	}

	if allocation.Period != period {
		t.Errorf("Expected period %s, got %s", period, allocation.Period)
	}

	// The underfunded amount is: amountOwed - available
	// amountOwed = -(-30000) = 30000 (we owe $300)
	// available = 0 (no allocations to payment category yet)
	// underfunded = 30000 - 0 = 30000
	expectedAmount := int64(30000)
	if allocation.Amount != expectedAmount {
		t.Errorf("Expected amount %d, got %d", expectedAmount, allocation.Amount)
	}

	if allocation.Notes != "Allocated to cover underfunded credit card balance" {
		t.Errorf("Expected specific notes, got: %s", allocation.Notes)
	}
}

func TestAllocationService_AllocateToCoverUnderfunded_NotPaymentCategory(t *testing.T) {
	// Setup
	ctx := context.Background()
	period := "2025-10"
	categoryID := "normal-cat-1"

	catRepo := &mockCategoryRepository{
		categories: []*domain.Category{
			{
				ID:                  categoryID,
				Name:                "Groceries",
				PaymentForAccountID: nil, // Not a payment category
			},
		},
	}

	allocRepo := &mockAllocationRepository{}
	txnRepo := &mockTransactionRepository{}
	budgetRepo := &mockBudgetStateRepository{}
	accountRepo := &mockAccountRepository{}

	service := NewAllocationService(allocRepo, catRepo, txnRepo, budgetRepo, accountRepo)

	// Execute
	_, err := service.AllocateToCoverUnderfunded(ctx, categoryID, period)

	// Verify
	if err == nil {
		t.Fatal("Expected error for non-payment category")
	}

	if err.Error() != "category is not a payment category" {
		t.Errorf("Expected specific error message, got: %v", err)
	}
}

func TestAllocationService_AllocateToCoverUnderfunded_CategoryNotFound(t *testing.T) {
	// Setup
	ctx := context.Background()
	period := "2025-10"
	categoryID := "nonexistent"

	catRepo := &mockCategoryRepository{
		categories: []*domain.Category{},
	}

	allocRepo := &mockAllocationRepository{}
	txnRepo := &mockTransactionRepository{}
	budgetRepo := &mockBudgetStateRepository{}
	accountRepo := &mockAccountRepository{}

	service := NewAllocationService(allocRepo, catRepo, txnRepo, budgetRepo, accountRepo)

	// Execute
	_, err := service.AllocateToCoverUnderfunded(ctx, categoryID, period)

	// Verify
	if err == nil {
		t.Fatal("Expected error for nonexistent category")
	}
}

func TestAllocationService_AllocateToCoverUnderfunded_NotUnderfunded(t *testing.T) {
	// Setup
	ctx := context.Background()
	period := "2025-10"

	accountID := "cc-account-1"
	paymentCategoryID := "payment-cat-1"

	allocRepo := &mockAllocationRepository{
		allocations: []*domain.Allocation{
			{
				ID:         "alloc-payment",
				CategoryID: paymentCategoryID,
				Period:     period,
				Amount:     100000, // $1000 allocated - more than needed
			},
		},
	}

	catRepo := &mockCategoryRepository{
		categories: []*domain.Category{
			{
				ID:                  paymentCategoryID,
				Name:                "Credit Card Payment",
				PaymentForAccountID: stringPtr(accountID),
			},
		},
	}

	// Credit card has positive balance (they owe us!)
	txnRepo := &mockTransactionRepository{
		transactions: []*domain.Transaction{
			{
				ID:        "txn-income",
				AccountID: "checking",
				Amount:    100000,
				Date:      time.Date(2025, 10, 1, 0, 0, 0, 0, time.UTC),
				Type:      "income",
			},
		},
	}

	accountRepo := &mockAccountRepository{
		accounts: []*domain.Account{
			{
				ID:      accountID,
				Name:    "Credit Card",
				Type:    "credit_card",
				Balance: 10000, // Positive balance - we overpaid
			},
		},
	}

	budgetRepo := &mockBudgetStateRepository{}

	service := NewAllocationService(allocRepo, catRepo, txnRepo, budgetRepo, accountRepo)

	// Execute
	_, err := service.AllocateToCoverUnderfunded(ctx, paymentCategoryID, period)

	// Verify
	if err == nil {
		t.Fatal("Expected error when payment category is not underfunded")
	}

	if err.Error() != "payment category is not underfunded" {
		t.Errorf("Expected specific error message, got: %v", err)
	}
}

func TestAllocationService_AllocateToCoverUnderfunded_InsufficientFunds(t *testing.T) {
	// Setup
	ctx := context.Background()
	period := "2025-10"

	accountID := "cc-account-1"
	paymentCategoryID := "payment-cat-1"
	normalCategoryID := "normal-cat-1"

	allocRepo := &mockAllocationRepository{
		allocations: []*domain.Allocation{
			{
				ID:         "alloc-normal",
				CategoryID: normalCategoryID,
				Period:     period,
				Amount:     95000, // $950 already allocated to normal category
			},
		},
	}

	catRepo := &mockCategoryRepository{
		categories: []*domain.Category{
			{
				ID:                  paymentCategoryID,
				Name:                "Credit Card Payment",
				PaymentForAccountID: stringPtr(accountID),
			},
			{
				ID:   normalCategoryID,
				Name: "Groceries",
			},
		},
	}

	// Only $1000 income, already allocated $950, so only $50 available
	// But we need $800 for underfunded CC
	txnRepo := &mockTransactionRepository{
		transactions: []*domain.Transaction{
			{
				ID:        "txn-income",
				AccountID: "checking",
				Amount:    100000, // $1000 income
				Date:      time.Date(2025, 10, 1, 0, 0, 0, 0, time.UTC),
				Type:      "income",
			},
			{
				ID:         "txn-cc",
				AccountID:  accountID,
				Amount:     -50000, // $500 spent on card
				Date:       time.Date(2025, 10, 15, 0, 0, 0, 0, time.UTC),
				CategoryID: stringPtr("cat-food"),
			},
		},
	}

	accountRepo := &mockAccountRepository{
		accounts: []*domain.Account{
			{
				ID:      accountID,
				Name:    "Credit Card",
				Type:    "credit_card",
				Balance: -30000, // Owe $300
			},
		},
	}

	budgetRepo := &mockBudgetStateRepository{}

	service := NewAllocationService(allocRepo, catRepo, txnRepo, budgetRepo, accountRepo)

	// Execute
	_, err := service.AllocateToCoverUnderfunded(ctx, paymentCategoryID, period)

	// Verify
	if err == nil {
		t.Fatal("Expected error for insufficient funds")
	}

	// Should mention insufficient funds
	if err.Error()[:len("insufficient funds")] != "insufficient funds" {
		t.Errorf("Expected insufficient funds error, got: %v", err)
	}
}

// Tests for calculateReadyToAssignWithoutUnderfunded

func TestAllocationService_CalculateReadyToAssignWithoutUnderfunded_Basic(t *testing.T) {
	// Setup
	ctx := context.Background()
	period := "2025-10"

	normalCategoryID := "normal-cat-1"
	paymentCategoryID := "payment-cat-1"

	allocRepo := &mockAllocationRepository{
		allocations: []*domain.Allocation{
			{
				ID:         "alloc-normal",
				CategoryID: normalCategoryID,
				Period:     period,
				Amount:     50000, // $500 allocated
			},
			{
				ID:         "alloc-payment",
				CategoryID: paymentCategoryID,
				Period:     period,
				Amount:     30000, // $300 allocated to payment (should be excluded)
			},
		},
	}

	catRepo := &mockCategoryRepository{
		categories: []*domain.Category{
			{
				ID:   normalCategoryID,
				Name: "Groceries",
			},
			{
				ID:                  paymentCategoryID,
				Name:                "CC Payment",
				PaymentForAccountID: stringPtr("cc-1"),
			},
		},
	}

	txnRepo := &mockTransactionRepository{
		transactions: []*domain.Transaction{
			{
				ID:        "txn-income",
				AccountID: "checking",
				Amount:    100000, // $1000 income
				Date:      time.Date(2025, 10, 1, 0, 0, 0, 0, time.UTC),
				Type:      "income",
			},
		},
	}

	budgetRepo := &mockBudgetStateRepository{}
	accountRepo := &mockAccountRepository{}

	service := NewAllocationService(allocRepo, catRepo, txnRepo, budgetRepo, accountRepo)

	// Execute
	rta, err := service.calculateReadyToAssignWithoutUnderfunded(ctx, period)

	// Verify
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	// RTA = $1000 (income) - $500 (normal allocation) - $0 (payment excluded) = $500
	expectedRTA := int64(50000)
	if rta != expectedRTA {
		t.Errorf("Expected RTA %d, got %d", expectedRTA, rta)
	}
}

func TestAllocationService_CalculateReadyToAssignWithoutUnderfunded_MultiplePaymentCategories(t *testing.T) {
	// Setup
	ctx := context.Background()
	period := "2025-10"

	normalCategoryID := "normal-cat-1"
	paymentCategory1ID := "payment-cat-1"
	paymentCategory2ID := "payment-cat-2"

	allocRepo := &mockAllocationRepository{
		allocations: []*domain.Allocation{
			{
				ID:         "alloc-normal",
				CategoryID: normalCategoryID,
				Period:     period,
				Amount:     20000, // $200 allocated
			},
			{
				ID:         "alloc-payment-1",
				CategoryID: paymentCategory1ID,
				Period:     period,
				Amount:     30000, // $300 allocated to payment 1 (excluded)
			},
			{
				ID:         "alloc-payment-2",
				CategoryID: paymentCategory2ID,
				Period:     period,
				Amount:     40000, // $400 allocated to payment 2 (excluded)
			},
		},
	}

	catRepo := &mockCategoryRepository{
		categories: []*domain.Category{
			{
				ID:   normalCategoryID,
				Name: "Groceries",
			},
			{
				ID:                  paymentCategory1ID,
				Name:                "CC Payment 1",
				PaymentForAccountID: stringPtr("cc-1"),
			},
			{
				ID:                  paymentCategory2ID,
				Name:                "CC Payment 2",
				PaymentForAccountID: stringPtr("cc-2"),
			},
		},
	}

	txnRepo := &mockTransactionRepository{
		transactions: []*domain.Transaction{
			{
				ID:        "txn-income",
				AccountID: "checking",
				Amount:    100000, // $1000 income
				Date:      time.Date(2025, 10, 1, 0, 0, 0, 0, time.UTC),
				Type:      "income",
			},
		},
	}

	budgetRepo := &mockBudgetStateRepository{}
	accountRepo := &mockAccountRepository{}

	service := NewAllocationService(allocRepo, catRepo, txnRepo, budgetRepo, accountRepo)

	// Execute
	rta, err := service.calculateReadyToAssignWithoutUnderfunded(ctx, period)

	// Verify
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	// RTA = $1000 (income) - $200 (normal allocation) = $800
	// Payment allocations should be excluded
	expectedRTA := int64(80000)
	if rta != expectedRTA {
		t.Errorf("Expected RTA %d, got %d", expectedRTA, rta)
	}
}

func TestAllocationService_CalculateReadyToAssignWithoutUnderfunded_OnlyIncludesUpToPeriod(t *testing.T) {
	// Setup
	ctx := context.Background()
	period := "2025-10"

	normalCategoryID := "normal-cat-1"

	allocRepo := &mockAllocationRepository{
		allocations: []*domain.Allocation{
			{
				ID:         "alloc-sep",
				CategoryID: normalCategoryID,
				Period:     "2025-09",
				Amount:     10000, // $100 in September
			},
			{
				ID:         "alloc-oct",
				CategoryID: normalCategoryID,
				Period:     "2025-10",
				Amount:     20000, // $200 in October
			},
			{
				ID:         "alloc-nov",
				CategoryID: normalCategoryID,
				Period:     "2025-11",
				Amount:     30000, // $300 in November (should not be included)
			},
		},
	}

	catRepo := &mockCategoryRepository{
		categories: []*domain.Category{
			{
				ID:   normalCategoryID,
				Name: "Groceries",
			},
		},
	}

	txnRepo := &mockTransactionRepository{
		transactions: []*domain.Transaction{
			{
				ID:        "txn-sep",
				AccountID: "checking",
				Amount:    50000, // $500 in September
				Date:      time.Date(2025, 9, 1, 0, 0, 0, 0, time.UTC),
				Type:      "income",
			},
			{
				ID:        "txn-oct",
				AccountID: "checking",
				Amount:    60000, // $600 in October
				Date:      time.Date(2025, 10, 1, 0, 0, 0, 0, time.UTC),
				Type:      "income",
			},
			{
				ID:        "txn-nov",
				AccountID: "checking",
				Amount:    70000, // $700 in November (should not be included)
				Date:      time.Date(2025, 11, 1, 0, 0, 0, 0, time.UTC),
				Type:      "income",
			},
		},
	}

	budgetRepo := &mockBudgetStateRepository{}
	accountRepo := &mockAccountRepository{}

	service := NewAllocationService(allocRepo, catRepo, txnRepo, budgetRepo, accountRepo)

	// Execute
	rta, err := service.calculateReadyToAssignWithoutUnderfunded(ctx, period)

	// Verify
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	// RTA = ($500 + $600) income - ($100 + $200) allocations = $800
	expectedRTA := int64(80000)
	if rta != expectedRTA {
		t.Errorf("Expected RTA %d, got %d", expectedRTA, rta)
	}
}

func TestAllocationService_CalculateReadyToAssignWithoutUnderfunded_ExcludesTransfers(t *testing.T) {
	// Setup
	ctx := context.Background()
	period := "2025-10"

	allocRepo := &mockAllocationRepository{
		allocations: []*domain.Allocation{},
	}

	catRepo := &mockCategoryRepository{
		categories: []*domain.Category{},
	}

	txnRepo := &mockTransactionRepository{
		transactions: []*domain.Transaction{
			{
				ID:        "txn-income",
				AccountID: "checking",
				Amount:    100000, // $1000 income
				Date:      time.Date(2025, 10, 1, 0, 0, 0, 0, time.UTC),
				Type:      "income",
			},
			{
				ID:        "txn-transfer",
				AccountID: "checking",
				Amount:    50000, // $500 transfer (should be excluded)
				Date:      time.Date(2025, 10, 5, 0, 0, 0, 0, time.UTC),
				Type:      "transfer",
			},
		},
	}

	budgetRepo := &mockBudgetStateRepository{}
	accountRepo := &mockAccountRepository{}

	service := NewAllocationService(allocRepo, catRepo, txnRepo, budgetRepo, accountRepo)

	// Execute
	rta, err := service.calculateReadyToAssignWithoutUnderfunded(ctx, period)

	// Verify
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	// RTA = $1000 (income only, transfer excluded)
	expectedRTA := int64(100000)
	if rta != expectedRTA {
		t.Errorf("Expected RTA %d, got %d", expectedRTA, rta)
	}
}
