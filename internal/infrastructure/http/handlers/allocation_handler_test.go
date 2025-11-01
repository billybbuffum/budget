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

	"github.com/billybbuffum/budget/internal/domain"
)

// Mock AllocationService for handler tests

type mockAllocationService struct {
	allocateToCoverUnderfundedResult    *domain.Allocation
	allocateToCoverUnderfundedUnderfunded int64
	allocateToCoverUnderfundedError     error
	calculateReadyToAssignResult        int64
	calculateReadyToAssignError         error
}

func (m *mockAllocationService) AllocateToCoverUnderfunded(
	ctx context.Context,
	paymentCategoryID string,
	period string,
) (*domain.Allocation, int64, error) {
	if m.allocateToCoverUnderfundedError != nil {
		return nil, 0, m.allocateToCoverUnderfundedError
	}
	return m.allocateToCoverUnderfundedResult, m.allocateToCoverUnderfundedUnderfunded, nil
}

func (m *mockAllocationService) CalculateReadyToAssignForPeriod(ctx context.Context, period string) (int64, error) {
	if m.calculateReadyToAssignError != nil {
		return 0, m.calculateReadyToAssignError
	}
	return m.calculateReadyToAssignResult, nil
}

func (m *mockAllocationService) CreateAllocation(ctx context.Context, categoryID string, amount int64, period, notes string) (*domain.Allocation, error) {
	return nil, nil
}

func (m *mockAllocationService) GetAllocation(ctx context.Context, id string) (*domain.Allocation, error) {
	return nil, nil
}

func (m *mockAllocationService) ListAllocations(ctx context.Context) ([]*domain.Allocation, error) {
	return nil, nil
}

func (m *mockAllocationService) GetAllocationsByPeriod(ctx context.Context, period string) ([]*domain.Allocation, error) {
	return nil, nil
}

func (m *mockAllocationService) UpdateAllocation(ctx context.Context, id string, amount int64, notes string) (*domain.Allocation, error) {
	return nil, nil
}

func (m *mockAllocationService) DeleteAllocation(ctx context.Context, id string) error {
	return nil
}

func (m *mockAllocationService) GetAllocationSummary(ctx context.Context, period string) ([]*domain.AllocationSummary, error) {
	return nil, nil
}

// Tests for CoverUnderfunded handler

func TestAllocationHandler_CoverUnderfunded_Success(t *testing.T) {
	// Setup
	paymentCategoryID := "550e8400-e29b-41d4-a716-446655440000"
	period := "2025-10"
	underfundedAmount := int64(20000)    // $200
	readyToAssignAfter := int64(330000)  // $3300

	allocation := &domain.Allocation{
		ID:         "allocation-id",
		CategoryID: paymentCategoryID,
		Period:     period,
		Amount:     underfundedAmount,
		Notes:      "Cover underfunded credit card spending",
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}

	mockService := &mockAllocationService{
		allocateToCoverUnderfundedResult:      allocation,
		allocateToCoverUnderfundedUnderfunded: underfundedAmount,
		calculateReadyToAssignResult:          readyToAssignAfter,
	}

	handler := NewAllocationHandler(mockService)

	requestBody := CoverUnderfundedRequest{
		PaymentCategoryID: paymentCategoryID,
		Period:            period,
	}
	body, _ := json.Marshal(requestBody)

	req := httptest.NewRequest("POST", "/api/allocations/cover-underfunded", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	// Act
	handler.CoverUnderfunded(w, req)

	// Assert
	if w.Code != http.StatusCreated {
		t.Errorf("CoverUnderfunded() status = %d, want %d", w.Code, http.StatusCreated)
	}

	var response map[string]interface{}
	if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if response["underfunded_amount"] != float64(underfundedAmount) {
		t.Errorf("Response underfunded_amount = %v, want %v", response["underfunded_amount"], underfundedAmount)
	}

	if response["ready_to_assign_after"] != float64(readyToAssignAfter) {
		t.Errorf("Response ready_to_assign_after = %v, want %v", response["ready_to_assign_after"], readyToAssignAfter)
	}

	allocationData, ok := response["allocation"].(map[string]interface{})
	if !ok {
		t.Fatal("Response allocation is not a map")
	}

	if allocationData["id"] != allocation.ID {
		t.Errorf("Response allocation.id = %v, want %v", allocationData["id"], allocation.ID)
	}

	if allocationData["category_id"] != paymentCategoryID {
		t.Errorf("Response allocation.category_id = %v, want %v", allocationData["category_id"], paymentCategoryID)
	}
}

func TestAllocationHandler_CoverUnderfunded_InvalidJSON(t *testing.T) {
	// Setup
	mockService := &mockAllocationService{}
	handler := NewAllocationHandler(mockService)

	req := httptest.NewRequest("POST", "/api/allocations/cover-underfunded", bytes.NewReader([]byte("invalid json")))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	// Act
	handler.CoverUnderfunded(w, req)

	// Assert
	if w.Code != http.StatusBadRequest {
		t.Errorf("CoverUnderfunded() status = %d, want %d", w.Code, http.StatusBadRequest)
	}

	if !bytes.Contains(w.Body.Bytes(), []byte("invalid request body")) {
		t.Errorf("CoverUnderfunded() body = %s, want 'invalid request body'", w.Body.String())
	}
}

func TestAllocationHandler_CoverUnderfunded_InvalidUUID(t *testing.T) {
	// Setup
	mockService := &mockAllocationService{}
	handler := NewAllocationHandler(mockService)

	requestBody := CoverUnderfundedRequest{
		PaymentCategoryID: "not-a-uuid",
		Period:            "2025-10",
	}
	body, _ := json.Marshal(requestBody)

	req := httptest.NewRequest("POST", "/api/allocations/cover-underfunded", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	// Act
	handler.CoverUnderfunded(w, req)

	// Assert
	if w.Code != http.StatusBadRequest {
		t.Errorf("CoverUnderfunded() status = %d, want %d", w.Code, http.StatusBadRequest)
	}

	if !bytes.Contains(w.Body.Bytes(), []byte("invalid UUID format")) {
		t.Errorf("CoverUnderfunded() body = %s, want 'invalid UUID format'", w.Body.String())
	}
}

func TestAllocationHandler_CoverUnderfunded_InvalidPeriodFormat(t *testing.T) {
	tests := []struct {
		name   string
		period string
	}{
		{"invalid month - 13", "2024-13"},
		{"invalid month - 00", "2024-00"},
		{"two digit year", "24-01"},
		{"single digit month", "2024-1"},
		{"slash separator", "2024/01"},
		{"empty string", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			mockService := &mockAllocationService{}
			handler := NewAllocationHandler(mockService)

			requestBody := CoverUnderfundedRequest{
				PaymentCategoryID: "550e8400-e29b-41d4-a716-446655440000",
				Period:            tt.period,
			}
			body, _ := json.Marshal(requestBody)

			req := httptest.NewRequest("POST", "/api/allocations/cover-underfunded", bytes.NewReader(body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			// Act
			handler.CoverUnderfunded(w, req)

			// Assert
			if w.Code != http.StatusBadRequest {
				t.Errorf("CoverUnderfunded() status = %d, want %d", w.Code, http.StatusBadRequest)
			}

			if !bytes.Contains(w.Body.Bytes(), []byte("invalid period format")) {
				t.Errorf("CoverUnderfunded() body = %s, want 'invalid period format'", w.Body.String())
			}
		})
	}
}

func TestAllocationHandler_CoverUnderfunded_PeriodOutOfRange(t *testing.T) {
	// Setup - period too far in the past
	mockService := &mockAllocationService{}
	handler := NewAllocationHandler(mockService)

	// Calculate a period 3 years ago (should fail)
	threeYearsAgo := time.Now().AddDate(-3, 0, 0).Format("2006-01")

	requestBody := CoverUnderfundedRequest{
		PaymentCategoryID: "550e8400-e29b-41d4-a716-446655440000",
		Period:            threeYearsAgo,
	}
	body, _ := json.Marshal(requestBody)

	req := httptest.NewRequest("POST", "/api/allocations/cover-underfunded", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	// Act
	handler.CoverUnderfunded(w, req)

	// Assert
	if w.Code != http.StatusBadRequest {
		t.Errorf("CoverUnderfunded() status = %d, want %d", w.Code, http.StatusBadRequest)
	}

	if !bytes.Contains(w.Body.Bytes(), []byte("too far in the past")) {
		t.Errorf("CoverUnderfunded() body = %s, want 'too far in the past'", w.Body.String())
	}
}

func TestAllocationHandler_CoverUnderfunded_CategoryNotFound(t *testing.T) {
	// Setup
	mockService := &mockAllocationService{
		allocateToCoverUnderfundedError: errors.New("payment category not found"),
	}
	handler := NewAllocationHandler(mockService)

	requestBody := CoverUnderfundedRequest{
		PaymentCategoryID: "550e8400-e29b-41d4-a716-446655440000",
		Period:            "2025-10",
	}
	body, _ := json.Marshal(requestBody)

	req := httptest.NewRequest("POST", "/api/allocations/cover-underfunded", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	// Act
	handler.CoverUnderfunded(w, req)

	// Assert
	if w.Code != http.StatusNotFound {
		t.Errorf("CoverUnderfunded() status = %d, want %d", w.Code, http.StatusNotFound)
	}

	if !bytes.Contains(w.Body.Bytes(), []byte("payment category not found")) {
		t.Errorf("CoverUnderfunded() body = %s, want 'payment category not found'", w.Body.String())
	}
}

func TestAllocationHandler_CoverUnderfunded_NotPaymentCategory(t *testing.T) {
	// Setup
	mockService := &mockAllocationService{
		allocateToCoverUnderfundedError: errors.New("category is not a payment category"),
	}
	handler := NewAllocationHandler(mockService)

	requestBody := CoverUnderfundedRequest{
		PaymentCategoryID: "550e8400-e29b-41d4-a716-446655440000",
		Period:            "2025-10",
	}
	body, _ := json.Marshal(requestBody)

	req := httptest.NewRequest("POST", "/api/allocations/cover-underfunded", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	// Act
	handler.CoverUnderfunded(w, req)

	// Assert
	if w.Code != http.StatusBadRequest {
		t.Errorf("CoverUnderfunded() status = %d, want %d", w.Code, http.StatusBadRequest)
	}

	if !bytes.Contains(w.Body.Bytes(), []byte("category is not a payment category")) {
		t.Errorf("CoverUnderfunded() body = %s, want 'category is not a payment category'", w.Body.String())
	}
}

func TestAllocationHandler_CoverUnderfunded_NotUnderfunded(t *testing.T) {
	// Setup
	mockService := &mockAllocationService{
		allocateToCoverUnderfundedError: errors.New("payment category is not underfunded"),
	}
	handler := NewAllocationHandler(mockService)

	requestBody := CoverUnderfundedRequest{
		PaymentCategoryID: "550e8400-e29b-41d4-a716-446655440000",
		Period:            "2025-10",
	}
	body, _ := json.Marshal(requestBody)

	req := httptest.NewRequest("POST", "/api/allocations/cover-underfunded", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	// Act
	handler.CoverUnderfunded(w, req)

	// Assert
	if w.Code != http.StatusBadRequest {
		t.Errorf("CoverUnderfunded() status = %d, want %d", w.Code, http.StatusBadRequest)
	}

	if !bytes.Contains(w.Body.Bytes(), []byte("payment category is not underfunded")) {
		t.Errorf("CoverUnderfunded() body = %s, want 'payment category is not underfunded'", w.Body.String())
	}
}

func TestAllocationHandler_CoverUnderfunded_InsufficientFunds(t *testing.T) {
	// Setup
	mockService := &mockAllocationService{
		allocateToCoverUnderfundedError: errors.New("insufficient funds: Ready to Assign: $1.00, Underfunded: $5.00"),
	}
	handler := NewAllocationHandler(mockService)

	requestBody := CoverUnderfundedRequest{
		PaymentCategoryID: "550e8400-e29b-41d4-a716-446655440000",
		Period:            "2025-10",
	}
	body, _ := json.Marshal(requestBody)

	req := httptest.NewRequest("POST", "/api/allocations/cover-underfunded", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	// Act
	handler.CoverUnderfunded(w, req)

	// Assert
	if w.Code != http.StatusBadRequest {
		t.Errorf("CoverUnderfunded() status = %d, want %d", w.Code, http.StatusBadRequest)
	}

	if !bytes.Contains(w.Body.Bytes(), []byte("insufficient funds")) {
		t.Errorf("CoverUnderfunded() body = %s, want 'insufficient funds'", w.Body.String())
	}

	// Verify the error message includes the amounts
	if !bytes.Contains(w.Body.Bytes(), []byte("$1.00")) || !bytes.Contains(w.Body.Bytes(), []byte("$5.00")) {
		t.Errorf("CoverUnderfunded() body should include specific amounts: %s", w.Body.String())
	}
}

func TestAllocationHandler_CoverUnderfunded_InternalServerError(t *testing.T) {
	// Setup - simulate an unexpected error
	mockService := &mockAllocationService{
		allocateToCoverUnderfundedError: errors.New("database connection failed"),
	}
	handler := NewAllocationHandler(mockService)

	requestBody := CoverUnderfundedRequest{
		PaymentCategoryID: "550e8400-e29b-41d4-a716-446655440000",
		Period:            "2025-10",
	}
	body, _ := json.Marshal(requestBody)

	req := httptest.NewRequest("POST", "/api/allocations/cover-underfunded", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	// Act
	handler.CoverUnderfunded(w, req)

	// Assert
	if w.Code != http.StatusInternalServerError {
		t.Errorf("CoverUnderfunded() status = %d, want %d", w.Code, http.StatusInternalServerError)
	}

	// Should return generic error message, not the internal error
	if !bytes.Contains(w.Body.Bytes(), []byte("Failed to process allocation request")) {
		t.Errorf("CoverUnderfunded() body = %s, want 'Failed to process allocation request'", w.Body.String())
	}

	// Should NOT expose internal error details
	if bytes.Contains(w.Body.Bytes(), []byte("database connection failed")) {
		t.Errorf("CoverUnderfunded() should not expose internal error: %s", w.Body.String())
	}
}

func TestAllocationHandler_CoverUnderfunded_RTACalculationFailure(t *testing.T) {
	// Setup - success in allocation, but RTA calculation fails
	paymentCategoryID := "550e8400-e29b-41d4-a716-446655440000"
	period := "2025-10"
	underfundedAmount := int64(20000)

	allocation := &domain.Allocation{
		ID:         "allocation-id",
		CategoryID: paymentCategoryID,
		Period:     period,
		Amount:     underfundedAmount,
		Notes:      "Cover underfunded credit card spending",
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}

	mockService := &mockAllocationService{
		allocateToCoverUnderfundedResult:      allocation,
		allocateToCoverUnderfundedUnderfunded: underfundedAmount,
		calculateReadyToAssignError:           errors.New("failed to calculate RTA"),
	}

	handler := NewAllocationHandler(mockService)

	requestBody := CoverUnderfundedRequest{
		PaymentCategoryID: paymentCategoryID,
		Period:            period,
	}
	body, _ := json.Marshal(requestBody)

	req := httptest.NewRequest("POST", "/api/allocations/cover-underfunded", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	// Act
	handler.CoverUnderfunded(w, req)

	// Assert
	// Should still succeed (201) even if RTA calculation fails
	if w.Code != http.StatusCreated {
		t.Errorf("CoverUnderfunded() status = %d, want %d", w.Code, http.StatusCreated)
	}

	var response map[string]interface{}
	if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	// RTA should be 0 when calculation fails
	if response["ready_to_assign_after"] != float64(0) {
		t.Errorf("Response ready_to_assign_after = %v, want 0", response["ready_to_assign_after"])
	}

	// But allocation should still be present
	if response["allocation"] == nil {
		t.Error("Response allocation should not be nil")
	}
}

func TestAllocationHandler_CoverUnderfunded_EmptyRequestBody(t *testing.T) {
	// Setup
	mockService := &mockAllocationService{}
	handler := NewAllocationHandler(mockService)

	req := httptest.NewRequest("POST", "/api/allocations/cover-underfunded", bytes.NewReader([]byte("{}")))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	// Act
	handler.CoverUnderfunded(w, req)

	// Assert
	if w.Code != http.StatusBadRequest {
		t.Errorf("CoverUnderfunded() status = %d, want %d", w.Code, http.StatusBadRequest)
	}

	// Should fail validation (empty UUID)
	if !bytes.Contains(w.Body.Bytes(), []byte("invalid")) {
		t.Errorf("CoverUnderfunded() body = %s, should contain validation error", w.Body.String())
	}
}

func TestAllocationHandler_CoverUnderfunded_ContentTypeJSON(t *testing.T) {
	// Setup
	paymentCategoryID := "550e8400-e29b-41d4-a716-446655440000"
	period := "2025-10"
	underfundedAmount := int64(20000)

	allocation := &domain.Allocation{
		ID:         "allocation-id",
		CategoryID: paymentCategoryID,
		Period:     period,
		Amount:     underfundedAmount,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}

	mockService := &mockAllocationService{
		allocateToCoverUnderfundedResult:      allocation,
		allocateToCoverUnderfundedUnderfunded: underfundedAmount,
		calculateReadyToAssignResult:          330000,
	}

	handler := NewAllocationHandler(mockService)

	requestBody := CoverUnderfundedRequest{
		PaymentCategoryID: paymentCategoryID,
		Period:            period,
	}
	body, _ := json.Marshal(requestBody)

	req := httptest.NewRequest("POST", "/api/allocations/cover-underfunded", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	// Act
	handler.CoverUnderfunded(w, req)

	// Assert
	if w.Code != http.StatusCreated {
		t.Errorf("CoverUnderfunded() status = %d, want %d", w.Code, http.StatusCreated)
	}

	// Verify response has JSON content type
	contentType := w.Header().Get("Content-Type")
	if contentType != "application/json" {
		t.Errorf("CoverUnderfunded() Content-Type = %s, want application/json", contentType)
	}
}
