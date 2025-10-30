package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/billybbuffum/budget/internal/application"
)

type AllocationHandler struct {
	allocationService *application.AllocationService
}

func NewAllocationHandler(allocationService *application.AllocationService) *AllocationHandler {
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
