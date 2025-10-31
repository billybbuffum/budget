package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/billybbuffum/budget/internal/application"
	"github.com/billybbuffum/budget/internal/domain"
)

// Mock repositories for handler testing

type mockAllocationRepositoryForHandler struct {
	allocations []*domain.Allocation
}

func (m *mockAllocationRepositoryForHandler) Create(ctx context.Context, allocation *domain.Allocation) error {
	m.allocations = append(m.allocations, allocation)
	return nil
}

func (m *mockAllocationRepositoryForHandler) GetByID(ctx context.Context, id string) (*domain.Allocation, error) {
	for _, a := range m.allocations {
		if a.ID == id {
			return a, nil
		}
	}
	return nil, errors.New("allocation not found")
}

func (m *mockAllocationRepositoryForHandler) GetByCategoryAndPeriod(ctx context.Context, categoryID, period string) (*domain.Allocation, error) {
	for _, a := range m.allocations {
		if a.CategoryID == categoryID && a.Period == period {
			return a, nil
		}
	}
	return nil, errors.New("allocation not found")
}

func (m *mockAllocationRepositoryForHandler) ListByPeriod(ctx context.Context, period string) ([]*domain.Allocation, error) {
	var result []*domain.Allocation
	for _, a := range m.allocations {
		if a.Period == period {
			result = append(result, a)
		}
	}
	return result, nil
}

func (m *mockAllocationRepositoryForHandler) List(ctx context.Context) ([]*domain.Allocation, error) {
	return m.allocations, nil
}

func (m *mockAllocationRepositoryForHandler) Update(ctx context.Context, allocation *domain.Allocation) error {
	for i, a := range m.allocations {
		if a.ID == allocation.ID {
			m.allocations[i] = allocation
			return nil
		}
	}
	return errors.New("allocation not found")
}

func (m *mockAllocationRepositoryForHandler) Delete(ctx context.Context, id string) error {
	for i, a := range m.allocations {
		if a.ID == id {
			m.allocations = append(m.allocations[:i], m.allocations[i+1:]...)
			return nil
		}
	}
	return errors.New("allocation not found")
}

type mockCategoryRepositoryForHandler struct {
	categories []*domain.Category
	getByIDErr error
	listErr    error
}

func (m *mockCategoryRepositoryForHandler) Create(ctx context.Context, category *domain.Category) error {
	m.categories = append(m.categories, category)
	return nil
}

func (m *mockCategoryRepositoryForHandler) GetByID(ctx context.Context, id string) (*domain.Category, error) {
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

func (m *mockCategoryRepositoryForHandler) GetPaymentCategoryByAccountID(ctx context.Context, accountID string) (*domain.Category, error) {
	for _, c := range m.categories {
		if c.PaymentForAccountID != nil && *c.PaymentForAccountID == accountID {
			return c, nil
		}
	}
	return nil, errors.New("payment category not found")
}

func (m *mockCategoryRepositoryForHandler) List(ctx context.Context) ([]*domain.Category, error) {
	if m.listErr != nil {
		return nil, m.listErr
	}
	return m.categories, nil
}

func (m *mockCategoryRepositoryForHandler) ListByGroup(ctx context.Context, groupID string) ([]*domain.Category, error) {
	var result []*domain.Category
	for _, c := range m.categories {
		if c.GroupID != nil && *c.GroupID == groupID {
			result = append(result, c)
		}
	}
	return result, nil
}

func (m *mockCategoryRepositoryForHandler) Update(ctx context.Context, category *domain.Category) error {
	for i, c := range m.categories {
		if c.ID == category.ID {
			m.categories[i] = category
			return nil
		}
	}
	return errors.New("category not found")
}

func (m *mockCategoryRepositoryForHandler) Delete(ctx context.Context, id string) error {
	for i, c := range m.categories {
		if c.ID == id {
			m.categories = append(m.categories[:i], m.categories[i+1:]...)
			return nil
		}
	}
	return errors.New("category not found")
}

type mockTransactionRepositoryForHandler struct {
	transactions []*domain.Transaction
}

func (m *mockTransactionRepositoryForHandler) Create(ctx context.Context, transaction *domain.Transaction) error {
	return nil
}

func (m *mockTransactionRepositoryForHandler) GetByID(ctx context.Context, id string) (*domain.Transaction, error) {
	return nil, nil
}

func (m *mockTransactionRepositoryForHandler) List(ctx context.Context) ([]*domain.Transaction, error) {
	return m.transactions, nil
}

func (m *mockTransactionRepositoryForHandler) ListByAccount(ctx context.Context, accountID string) ([]*domain.Transaction, error) {
	return nil, nil
}

func (m *mockTransactionRepositoryForHandler) ListByCategory(ctx context.Context, categoryID string) ([]*domain.Transaction, error) {
	return nil, nil
}

func (m *mockTransactionRepositoryForHandler) ListByPeriod(ctx context.Context, startDate, endDate string) ([]*domain.Transaction, error) {
	return nil, nil
}

func (m *mockTransactionRepositoryForHandler) ListUncategorized(ctx context.Context) ([]*domain.Transaction, error) {
	return nil, nil
}

func (m *mockTransactionRepositoryForHandler) GetCategoryActivity(ctx context.Context, categoryID, period string) (int64, error) {
	return 0, nil
}

func (m *mockTransactionRepositoryForHandler) FindDuplicate(ctx context.Context, accountID string, date time.Time, amount int64, description string) (*domain.Transaction, error) {
	return nil, nil
}

func (m *mockTransactionRepositoryForHandler) FindByFitID(ctx context.Context, accountID string, fitID string) (*domain.Transaction, error) {
	return nil, nil
}

func (m *mockTransactionRepositoryForHandler) Update(ctx context.Context, transaction *domain.Transaction) error {
	return nil
}

func (m *mockTransactionRepositoryForHandler) BulkUpdateCategory(ctx context.Context, transactionIDs []string, categoryID *string) error {
	return nil
}

func (m *mockTransactionRepositoryForHandler) Delete(ctx context.Context, id string) error {
	return nil
}

type mockBudgetStateRepositoryForHandler struct {
	state *domain.BudgetState
}

func (m *mockBudgetStateRepositoryForHandler) Get(ctx context.Context) (*domain.BudgetState, error) {
	if m.state == nil {
		m.state = &domain.BudgetState{ReadyToAssign: 0}
	}
	return m.state, nil
}

func (m *mockBudgetStateRepositoryForHandler) Update(ctx context.Context, state *domain.BudgetState) error {
	m.state = state
	return nil
}

func (m *mockBudgetStateRepositoryForHandler) AdjustReadyToAssign(ctx context.Context, delta int64) error {
	if m.state == nil {
		m.state = &domain.BudgetState{ReadyToAssign: 0}
	}
	m.state.ReadyToAssign += delta
	return nil
}

type mockAccountRepositoryForHandler struct {
	accounts []*domain.Account
}

func (m *mockAccountRepositoryForHandler) Create(ctx context.Context, account *domain.Account) error {
	m.accounts = append(m.accounts, account)
	return nil
}

func (m *mockAccountRepositoryForHandler) GetByID(ctx context.Context, id string) (*domain.Account, error) {
	for _, a := range m.accounts {
		if a.ID == id {
			return a, nil
		}
	}
	return nil, errors.New("account not found")
}

func (m *mockAccountRepositoryForHandler) List(ctx context.Context) ([]*domain.Account, error) {
	return m.accounts, nil
}

func (m *mockAccountRepositoryForHandler) Update(ctx context.Context, account *domain.Account) error {
	return nil
}

func (m *mockAccountRepositoryForHandler) Delete(ctx context.Context, id string) error {
	return nil
}

func (m *mockAccountRepositoryForHandler) GetTotalBalance(ctx context.Context) (int64, error) {
	var total int64
	for _, a := range m.accounts {
		total += a.Balance
	}
	return total, nil
}

func stringPtr(s string) *string {
	return &s
}

// Helper function to create allocation service with specific test data
func createTestAllocationService(
	categories []*domain.Category,
	transactions []*domain.Transaction,
	accounts []*domain.Account,
	allocations []*domain.Allocation,
) *application.AllocationService {
	allocRepo := &mockAllocationRepositoryForHandler{allocations: allocations}
	catRepo := &mockCategoryRepositoryForHandler{categories: categories}
	txnRepo := &mockTransactionRepositoryForHandler{transactions: transactions}
	budgetRepo := &mockBudgetStateRepositoryForHandler{}
	accountRepo := &mockAccountRepositoryForHandler{accounts: accounts}

	return application.NewAllocationService(allocRepo, catRepo, txnRepo, budgetRepo, accountRepo)
}

// Tests for CoverUnderfunded handler

func TestAllocationHandler_CoverUnderfunded_Success(t *testing.T) {
	// Setup - Create test data for a successful scenario
	accountID := "cc-account-1"
	paymentCategoryID := "payment-cat-1"

	categories := []*domain.Category{
		{
			ID:                  paymentCategoryID,
			Name:                "Credit Card Payment",
			PaymentForAccountID: stringPtr(accountID),
		},
	}

	accounts := []*domain.Account{
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
	}

	transactions := []*domain.Transaction{
		{
			ID:        "txn-income-1",
			AccountID: "checking",
			Amount:    200000, // $2000 income
			Date:      time.Date(2025, 10, 1, 0, 0, 0, 0, time.UTC),
			Type:      "income",
		},
	}

	allocations := []*domain.Allocation{}

	service := createTestAllocationService(categories, transactions, accounts, allocations)
	handler := NewAllocationHandler(service)

	// Create request
	requestBody := CoverUnderfundedRequest{
		CategoryID: paymentCategoryID,
		Period:     "2025-10",
	}
	body, _ := json.Marshal(requestBody)
	req := httptest.NewRequest("POST", "/api/allocations/cover-underfunded", bytes.NewReader(body))
	w := httptest.NewRecorder()

	// Execute
	handler.CoverUnderfunded(w, req)

	// Verify
	if w.Code != http.StatusCreated {
		t.Errorf("Expected status %d, got %d. Body: %s", http.StatusCreated, w.Code, w.Body.String())
	}

	var response domain.Allocation
	if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if response.CategoryID != paymentCategoryID {
		t.Errorf("Expected category ID %s, got %s", paymentCategoryID, response.CategoryID)
	}

	contentType := w.Header().Get("Content-Type")
	if contentType != "application/json" {
		t.Errorf("Expected Content-Type 'application/json', got '%s'", contentType)
	}
}

func TestAllocationHandler_CoverUnderfunded_MissingCategoryID(t *testing.T) {
	// Setup
	service := createTestAllocationService([]*domain.Category{}, []*domain.Transaction{}, []*domain.Account{}, []*domain.Allocation{})
	handler := NewAllocationHandler(service)

	// Create request without category_id
	requestBody := CoverUnderfundedRequest{
		Period: "2025-10",
		// CategoryID is missing
	}
	body, _ := json.Marshal(requestBody)
	req := httptest.NewRequest("POST", "/api/allocations/cover-underfunded", bytes.NewReader(body))
	w := httptest.NewRecorder()

	// Execute
	handler.CoverUnderfunded(w, req)

	// Verify
	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status %d, got %d", http.StatusBadRequest, w.Code)
	}

	responseBody := w.Body.String()
	if responseBody != "category_id is required\n" {
		t.Errorf("Expected error message about category_id, got: %s", responseBody)
	}
}

func TestAllocationHandler_CoverUnderfunded_MissingPeriod(t *testing.T) {
	// Setup
	service := createTestAllocationService([]*domain.Category{}, []*domain.Transaction{}, []*domain.Account{}, []*domain.Allocation{})
	handler := NewAllocationHandler(service)

	// Create request without period
	requestBody := CoverUnderfundedRequest{
		CategoryID: "payment-cat-1",
		// Period is missing
	}
	body, _ := json.Marshal(requestBody)
	req := httptest.NewRequest("POST", "/api/allocations/cover-underfunded", bytes.NewReader(body))
	w := httptest.NewRecorder()

	// Execute
	handler.CoverUnderfunded(w, req)

	// Verify
	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status %d, got %d", http.StatusBadRequest, w.Code)
	}

	responseBody := w.Body.String()
	if responseBody != "period is required\n" {
		t.Errorf("Expected error message about period, got: %s", responseBody)
	}
}

func TestAllocationHandler_CoverUnderfunded_InvalidJSON(t *testing.T) {
	// Setup
	service := createTestAllocationService([]*domain.Category{}, []*domain.Transaction{}, []*domain.Account{}, []*domain.Allocation{})
	handler := NewAllocationHandler(service)

	// Create request with invalid JSON
	req := httptest.NewRequest("POST", "/api/allocations/cover-underfunded", bytes.NewReader([]byte("invalid json")))
	w := httptest.NewRecorder()

	// Execute
	handler.CoverUnderfunded(w, req)

	// Verify
	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status %d, got %d", http.StatusBadRequest, w.Code)
	}

	responseBody := w.Body.String()
	if responseBody != "invalid request body\n" {
		t.Errorf("Expected error message about invalid request body, got: %s", responseBody)
	}
}

func TestAllocationHandler_CoverUnderfunded_NotPaymentCategory(t *testing.T) {
	// Setup - Category exists but is not a payment category
	categoryID := "normal-cat-1"
	categories := []*domain.Category{
		{
			ID:                  categoryID,
			Name:                "Groceries",
			PaymentForAccountID: nil, // Not a payment category
		},
	}

	service := createTestAllocationService(categories, []*domain.Transaction{}, []*domain.Account{}, []*domain.Allocation{})
	handler := NewAllocationHandler(service)

	// Create request
	requestBody := CoverUnderfundedRequest{
		CategoryID: categoryID,
		Period:     "2025-10",
	}
	body, _ := json.Marshal(requestBody)
	req := httptest.NewRequest("POST", "/api/allocations/cover-underfunded", bytes.NewReader(body))
	w := httptest.NewRecorder()

	// Execute
	handler.CoverUnderfunded(w, req)

	// Verify
	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status %d, got %d", http.StatusBadRequest, w.Code)
	}

	responseBody := w.Body.String()
	if responseBody != "category is not a payment category\n" {
		t.Errorf("Expected error message about payment category, got: %s", responseBody)
	}
}

func TestAllocationHandler_CoverUnderfunded_CategoryNotFound(t *testing.T) {
	// Setup - No categories exist
	service := createTestAllocationService([]*domain.Category{}, []*domain.Transaction{}, []*domain.Account{}, []*domain.Allocation{})
	handler := NewAllocationHandler(service)

	// Create request with nonexistent category
	requestBody := CoverUnderfundedRequest{
		CategoryID: "nonexistent",
		Period:     "2025-10",
	}
	body, _ := json.Marshal(requestBody)
	req := httptest.NewRequest("POST", "/api/allocations/cover-underfunded", bytes.NewReader(body))
	w := httptest.NewRecorder()

	// Execute
	handler.CoverUnderfunded(w, req)

	// Verify
	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status %d, got %d", http.StatusBadRequest, w.Code)
	}

	responseBody := w.Body.String()
	if len(responseBody) == 0 {
		t.Error("Expected error message in response body")
	}
}

func TestAllocationHandler_CoverUnderfunded_NotUnderfunded(t *testing.T) {
	// Setup - Payment category with positive balance (not underfunded)
	accountID := "cc-account-1"
	paymentCategoryID := "payment-cat-1"

	categories := []*domain.Category{
		{
			ID:                  paymentCategoryID,
			Name:                "Credit Card Payment",
			PaymentForAccountID: stringPtr(accountID),
		},
	}

	accounts := []*domain.Account{
		{
			ID:      accountID,
			Name:    "Credit Card",
			Type:    "credit_card",
			Balance: 10000, // Positive balance - overpaid
		},
	}

	// Enough allocation to cover
	allocations := []*domain.Allocation{
		{
			ID:         "alloc-1",
			CategoryID: paymentCategoryID,
			Period:     "2025-10",
			Amount:     50000,
		},
	}

	service := createTestAllocationService(categories, []*domain.Transaction{}, accounts, allocations)
	handler := NewAllocationHandler(service)

	// Create request
	requestBody := CoverUnderfundedRequest{
		CategoryID: paymentCategoryID,
		Period:     "2025-10",
	}
	body, _ := json.Marshal(requestBody)
	req := httptest.NewRequest("POST", "/api/allocations/cover-underfunded", bytes.NewReader(body))
	w := httptest.NewRecorder()

	// Execute
	handler.CoverUnderfunded(w, req)

	// Verify
	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status %d, got %d", http.StatusBadRequest, w.Code)
	}

	responseBody := w.Body.String()
	if responseBody != "payment category is not underfunded\n" {
		t.Errorf("Expected error message about underfunded, got: %s", responseBody)
	}
}

func TestAllocationHandler_CoverUnderfunded_ResponseFormat(t *testing.T) {
	// Setup - Valid scenario to test response format
	accountID := "cc-account-1"
	paymentCategoryID := "payment-cat-1"

	categories := []*domain.Category{
		{
			ID:                  paymentCategoryID,
			Name:                "Credit Card Payment",
			PaymentForAccountID: stringPtr(accountID),
		},
	}

	accounts := []*domain.Account{
		{
			ID:      accountID,
			Name:    "Credit Card",
			Type:    "credit_card",
			Balance: -30000, // Owe $300
		},
	}

	transactions := []*domain.Transaction{
		{
			ID:        "txn-income-1",
			AccountID: "checking",
			Amount:    200000, // $2000 income
			Date:      time.Date(2025, 10, 1, 0, 0, 0, 0, time.UTC),
			Type:      "income",
		},
	}

	service := createTestAllocationService(categories, transactions, accounts, []*domain.Allocation{})
	handler := NewAllocationHandler(service)

	// Create request
	requestBody := CoverUnderfundedRequest{
		CategoryID: paymentCategoryID,
		Period:     "2025-10",
	}
	body, _ := json.Marshal(requestBody)
	req := httptest.NewRequest("POST", "/api/allocations/cover-underfunded", bytes.NewReader(body))
	w := httptest.NewRecorder()

	// Execute
	handler.CoverUnderfunded(w, req)

	// Verify response format
	if w.Code != http.StatusCreated {
		t.Errorf("Expected status %d, got %d. Body: %s", http.StatusCreated, w.Code, w.Body.String())
	}

	// Verify Content-Type
	contentType := w.Header().Get("Content-Type")
	if contentType != "application/json" {
		t.Errorf("Expected Content-Type 'application/json', got '%s'", contentType)
	}

	// Verify response can be decoded as Allocation
	var response domain.Allocation
	if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
		t.Fatalf("Failed to decode response as Allocation: %v", err)
	}

	// Verify response has required fields
	if response.ID == "" {
		t.Error("Expected allocation ID to be set")
	}
	if response.CategoryID != paymentCategoryID {
		t.Errorf("Expected category ID %s, got %s", paymentCategoryID, response.CategoryID)
	}
	if response.Period != "2025-10" {
		t.Errorf("Expected period '2025-10', got %s", response.Period)
	}
	if response.Amount <= 0 {
		t.Errorf("Expected positive amount, got %d", response.Amount)
	}
}
