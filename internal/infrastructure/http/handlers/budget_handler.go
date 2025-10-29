package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/billybbuffum/budget/internal/application"
)

type BudgetHandler struct {
	budgetService *application.BudgetService
}

func NewBudgetHandler(budgetService *application.BudgetService) *BudgetHandler {
	return &BudgetHandler{budgetService: budgetService}
}

type CreateBudgetRequest struct {
	CategoryID string  `json:"category_id"`
	Amount     float64 `json:"amount"`
	Period     string  `json:"period"`
	Notes      string  `json:"notes"`
}

type UpdateBudgetRequest struct {
	CategoryID string  `json:"category_id"`
	Amount     float64 `json:"amount"`
	Period     string  `json:"period"`
	Notes      string  `json:"notes"`
}

func (h *BudgetHandler) CreateBudget(w http.ResponseWriter, r *http.Request) {
	var req CreateBudgetRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	budget, err := h.budgetService.CreateBudget(r.Context(), req.CategoryID, req.Amount, req.Period, req.Notes)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(budget)
}

func (h *BudgetHandler) GetBudget(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		http.Error(w, "budget id is required", http.StatusBadRequest)
		return
	}

	budget, err := h.budgetService.GetBudget(r.Context(), id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(budget)
}

func (h *BudgetHandler) ListBudgets(w http.ResponseWriter, r *http.Request) {
	period := r.URL.Query().Get("period")

	var budgets interface{}
	var err error

	if period != "" {
		budgets, err = h.budgetService.ListBudgetsByPeriod(r.Context(), period)
	} else {
		budgets, err = h.budgetService.ListBudgets(r.Context())
	}

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(budgets)
}

func (h *BudgetHandler) GetBudgetSummary(w http.ResponseWriter, r *http.Request) {
	period := r.URL.Query().Get("period")
	if period == "" {
		http.Error(w, "period query parameter is required", http.StatusBadRequest)
		return
	}

	summary, err := h.budgetService.GetBudgetSummary(r.Context(), period)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(summary)
}

func (h *BudgetHandler) UpdateBudget(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		http.Error(w, "budget id is required", http.StatusBadRequest)
		return
	}

	var req UpdateBudgetRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	budget, err := h.budgetService.UpdateBudget(r.Context(), id, req.CategoryID, req.Amount, req.Period, req.Notes)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(budget)
}

func (h *BudgetHandler) DeleteBudget(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		http.Error(w, "budget id is required", http.StatusBadRequest)
		return
	}

	if err := h.budgetService.DeleteBudget(r.Context(), id); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
