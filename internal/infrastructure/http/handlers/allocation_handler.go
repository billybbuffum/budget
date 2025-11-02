package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"net/http"

	"github.com/billybbuffum/budget/internal/domain"
	"github.com/billybbuffum/budget/internal/infrastructure/http/validators"
)

// AllocationServiceInterface defines the interface for allocation operations
type AllocationServiceInterface interface {
	CreateAllocation(ctx context.Context, categoryID string, amount int64, period, notes string) (*domain.Allocation, error)
	GetAllocation(ctx context.Context, id string) (*domain.Allocation, error)
	ListAllocations(ctx context.Context) ([]*domain.Allocation, error)
	ListAllocationsByPeriod(ctx context.Context, period string) ([]*domain.Allocation, error)
	DeleteAllocation(ctx context.Context, id string) error
	GetAllocationSummary(ctx context.Context, period string) ([]*domain.AllocationSummary, error)
	AllocateToCoverUnderfunded(ctx context.Context, paymentCategoryID string, period string) (*domain.Allocation, int64, error)
	CalculateReadyToAssignForPeriod(ctx context.Context, period string) (int64, error)
}

type AllocationHandler struct {
	allocationService AllocationServiceInterface
}

func NewAllocationHandler(allocationService AllocationServiceInterface) *AllocationHandler {
	return &AllocationHandler{
		allocationService: allocationService,
	}
}

type CreateAllocationRequest struct {
	CategoryID string `json:"category_id"`
	Amount     int64  `json:"amount"` // in cents
	Period     string `json:"period"` // YYYY-MM
	Notes      string `json:"notes"`
}

func (h *AllocationHandler) CreateAllocation(w http.ResponseWriter, r *http.Request) {
	var req CreateAllocationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	allocation, err := h.allocationService.CreateAllocation(r.Context(), req.CategoryID, req.Amount, req.Period, req.Notes)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(allocation)
}

func (h *AllocationHandler) GetAllocation(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		http.Error(w, "allocation id is required", http.StatusBadRequest)
		return
	}

	allocation, err := h.allocationService.GetAllocation(r.Context(), id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(allocation)
}

func (h *AllocationHandler) ListAllocations(w http.ResponseWriter, r *http.Request) {
	period := r.URL.Query().Get("period")

	var allocations interface{}
	var err error

	if period != "" {
		allocations, err = h.allocationService.ListAllocationsByPeriod(r.Context(), period)
	} else {
		allocations, err = h.allocationService.ListAllocations(r.Context())
	}

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(allocations)
}

func (h *AllocationHandler) GetAllocationSummary(w http.ResponseWriter, r *http.Request) {
	period := r.URL.Query().Get("period")
	if period == "" {
		http.Error(w, "period query parameter is required", http.StatusBadRequest)
		return
	}

	summary, err := h.allocationService.GetAllocationSummary(r.Context(), period)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Calculate Ready to Assign for this period
	readyToAssign, err := h.allocationService.CalculateReadyToAssignForPeriod(r.Context(), period)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Include Ready to Assign in response
	response := map[string]interface{}{
		"categories":      summary,
		"ready_to_assign": readyToAssign,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (h *AllocationHandler) GetReadyToAssign(w http.ResponseWriter, r *http.Request) {
	period := r.URL.Query().Get("period")
	if period == "" {
		http.Error(w, "period query parameter is required", http.StatusBadRequest)
		return
	}

	readyToAssign, err := h.allocationService.CalculateReadyToAssignForPeriod(r.Context(), period)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	response := map[string]int64{
		"ready_to_assign": readyToAssign,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (h *AllocationHandler) DeleteAllocation(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		http.Error(w, "allocation id is required", http.StatusBadRequest)
		return
	}

	if err := h.allocationService.DeleteAllocation(r.Context(), id); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// CoverUnderfundedRequest represents the request body for covering underfunded payment categories
type CoverUnderfundedRequest struct {
	PaymentCategoryID string `json:"payment_category_id"`
	Period            string `json:"period"` // YYYY-MM
}

// CoverUnderfunded handles POST /api/allocations/cover-underfunded
// Creates an allocation to cover an underfunded payment category
func (h *AllocationHandler) CoverUnderfunded(w http.ResponseWriter, r *http.Request) {
	var req CoverUnderfundedRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Printf("ERROR: Failed to decode request body: %v", err)
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	// Validate UUID format
	if err := validators.ValidateUUID(req.PaymentCategoryID); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Validate period format and range
	if err := validators.ValidatePeriodFormat(req.Period); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err := validators.ValidatePeriodRange(req.Period); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Call service method to allocate to cover underfunded
	allocation, underfundedAmount, err := h.allocationService.AllocateToCoverUnderfunded(
		r.Context(),
		req.PaymentCategoryID,
		req.Period,
	)

	if err != nil {
		// Log detailed error internally
		log.Printf("ERROR: Failed to cover underfunded for category %s: %v", req.PaymentCategoryID, err)

		// Use typed error checking for appropriate status codes
		if errors.Is(err, domain.ErrCategoryNotFound) {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}

		if errors.Is(err, domain.ErrNotPaymentCategory) ||
			errors.Is(err, domain.ErrNotUnderfunded) ||
			errors.Is(err, domain.ErrInsufficientFunds) {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		// For all other errors, return generic internal server error
		http.Error(w, "Failed to process allocation request", http.StatusInternalServerError)
		return
	}

	// Calculate Ready to Assign after the allocation
	readyToAssignAfter, err := h.allocationService.CalculateReadyToAssignForPeriod(r.Context(), req.Period)
	if err != nil {
		log.Printf("WARNING: Failed to calculate Ready to Assign after allocation: %v", err)
		// Continue with response even if RTA calculation fails
		readyToAssignAfter = 0
	}

	// Prepare successful response
	response := map[string]interface{}{
		"allocation":            allocation,
		"underfunded_amount":    underfundedAmount,
		"ready_to_assign_after": readyToAssignAfter,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(response)
}
