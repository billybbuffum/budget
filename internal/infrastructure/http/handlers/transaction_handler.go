package handlers

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/billybbuffum/budget/internal/application"
)

type TransactionHandler struct {
	transactionService *application.TransactionService
}

func NewTransactionHandler(transactionService *application.TransactionService) *TransactionHandler {
	return &TransactionHandler{transactionService: transactionService}
}

type CreateTransactionRequest struct {
	AccountID   string    `json:"account_id"`
	CategoryID  *string   `json:"category_id,omitempty"` // Optional for income, required for expenses
	Amount      int64     `json:"amount"`                // in cents (positive=inflow, negative=outflow)
	Description string    `json:"description"`
	Date        time.Time `json:"date"`
}

type CreateTransferRequest struct {
	FromAccountID string    `json:"from_account_id"`
	ToAccountID   string    `json:"to_account_id"`
	Amount        int64     `json:"amount"`      // in cents (must be positive)
	Description   string    `json:"description"`
	Date          time.Time `json:"date"`
}

type UpdateTransactionRequest struct {
	AccountID   string    `json:"account_id"`
	CategoryID  *string   `json:"category_id,omitempty"`
	Amount      int64     `json:"amount"`
	Description string    `json:"description"`
	Date        time.Time `json:"date"`
}

func (h *TransactionHandler) CreateTransaction(w http.ResponseWriter, r *http.Request) {
	var req CreateTransactionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	transaction, err := h.transactionService.CreateTransaction(
		r.Context(), req.AccountID, req.CategoryID, req.Amount, req.Description, req.Date)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(transaction)
}

func (h *TransactionHandler) GetTransaction(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		http.Error(w, "transaction id is required", http.StatusBadRequest)
		return
	}

	transaction, err := h.transactionService.GetTransaction(r.Context(), id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(transaction)
}

func (h *TransactionHandler) ListTransactions(w http.ResponseWriter, r *http.Request) {
	accountID := r.URL.Query().Get("account_id")
	categoryID := r.URL.Query().Get("category_id")
	startDate := r.URL.Query().Get("start_date")
	endDate := r.URL.Query().Get("end_date")
	uncategorized := r.URL.Query().Get("uncategorized")

	var transactions interface{}
	var err error

	if uncategorized == "true" {
		transactions, err = h.transactionService.ListUncategorizedTransactions(r.Context())
	} else if accountID != "" {
		transactions, err = h.transactionService.ListTransactionsByAccount(r.Context(), accountID)
	} else if categoryID != "" {
		transactions, err = h.transactionService.ListTransactionsByCategory(r.Context(), categoryID)
	} else if startDate != "" && endDate != "" {
		start, err1 := time.Parse(time.RFC3339, startDate)
		end, err2 := time.Parse(time.RFC3339, endDate)
		if err1 != nil || err2 != nil {
			http.Error(w, "invalid date format, use RFC3339", http.StatusBadRequest)
			return
		}
		transactions, err = h.transactionService.ListTransactionsByPeriod(r.Context(), start, end)
	} else {
		transactions, err = h.transactionService.ListTransactions(r.Context())
	}

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(transactions)
}

func (h *TransactionHandler) UpdateTransaction(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		http.Error(w, "transaction id is required", http.StatusBadRequest)
		return
	}

	var req UpdateTransactionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	transaction, err := h.transactionService.UpdateTransaction(
		r.Context(), id, req.AccountID, req.CategoryID, req.Amount, req.Description, req.Date)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(transaction)
}

func (h *TransactionHandler) DeleteTransaction(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		http.Error(w, "transaction id is required", http.StatusBadRequest)
		return
	}

	if err := h.transactionService.DeleteTransaction(r.Context(), id); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *TransactionHandler) CreateTransfer(w http.ResponseWriter, r *http.Request) {
	var req CreateTransferRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	transaction, err := h.transactionService.CreateTransfer(
		r.Context(), req.FromAccountID, req.ToAccountID, req.Amount, req.Description, req.Date)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(transaction)
}

type BulkCategorizeRequest struct {
	TransactionIDs []string `json:"transaction_ids"`
	CategoryID     *string  `json:"category_id,omitempty"`
}

func (h *TransactionHandler) BulkCategorizeTransactions(w http.ResponseWriter, r *http.Request) {
	var req BulkCategorizeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	if len(req.TransactionIDs) == 0 {
		http.Error(w, "transaction_ids is required", http.StatusBadRequest)
		return
	}

	if err := h.transactionService.BulkCategorizeTransactions(r.Context(), req.TransactionIDs, req.CategoryID); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
