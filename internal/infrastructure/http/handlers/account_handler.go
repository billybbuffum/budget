package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/billybbuffum/budget/internal/application"
	"github.com/billybbuffum/budget/internal/domain"
)

type AccountHandler struct {
	accountService *application.AccountService
}

func NewAccountHandler(accountService *application.AccountService) *AccountHandler {
	return &AccountHandler{accountService: accountService}
}

type CreateAccountRequest struct {
	Name    string `json:"name"`
	Balance int64  `json:"balance"` // in cents
	Type    string `json:"type"`    // checking, savings, cash
}

type UpdateAccountRequest struct {
	Name    string `json:"name"`
	Balance int64  `json:"balance"`
	Type    string `json:"type"`
}

func (h *AccountHandler) CreateAccount(w http.ResponseWriter, r *http.Request) {
	var req CreateAccountRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	account, err := h.accountService.CreateAccount(r.Context(), req.Name, req.Balance, domain.AccountType(req.Type))
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(account)
}

func (h *AccountHandler) GetAccount(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		http.Error(w, "account id is required", http.StatusBadRequest)
		return
	}

	account, err := h.accountService.GetAccount(r.Context(), id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(account)
}

func (h *AccountHandler) ListAccounts(w http.ResponseWriter, r *http.Request) {
	accounts, err := h.accountService.ListAccounts(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(accounts)
}

func (h *AccountHandler) UpdateAccount(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		http.Error(w, "account id is required", http.StatusBadRequest)
		return
	}

	var req UpdateAccountRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	account, err := h.accountService.UpdateAccount(r.Context(), id, req.Name, req.Balance, domain.AccountType(req.Type))
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(account)
}

func (h *AccountHandler) DeleteAccount(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		http.Error(w, "account id is required", http.StatusBadRequest)
		return
	}

	if err := h.accountService.DeleteAccount(r.Context(), id); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *AccountHandler) GetSummary(w http.ResponseWriter, r *http.Request) {
	totalBalance, err := h.accountService.GetTotalBalance(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	summary := map[string]int64{
		"total_balance": totalBalance,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(summary)
}
