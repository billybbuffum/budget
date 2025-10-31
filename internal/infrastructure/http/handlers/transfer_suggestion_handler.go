package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/billybbuffum/budget/internal/application"
	"github.com/billybbuffum/budget/internal/domain"
)

// TransferSuggestionHandler handles transfer suggestion HTTP requests
type TransferSuggestionHandler struct {
	linkService       *application.TransferLinkService
	suggestionRepo    domain.TransferSuggestionRepository
	transactionRepo   domain.TransactionRepository
}

// NewTransferSuggestionHandler creates a new transfer suggestion handler
func NewTransferSuggestionHandler(
	linkService *application.TransferLinkService,
	suggestionRepo domain.TransferSuggestionRepository,
	transactionRepo domain.TransactionRepository,
) *TransferSuggestionHandler {
	return &TransferSuggestionHandler{
		linkService:     linkService,
		suggestionRepo:  suggestionRepo,
		transactionRepo: transactionRepo,
	}
}

// SuggestionWithTransactions represents a suggestion with full transaction details
type SuggestionWithTransactions struct {
	domain.TransferMatchSuggestion
	TransactionA *domain.Transaction `json:"transaction_a"`
	TransactionB *domain.Transaction `json:"transaction_b"`
}

// ListPendingSuggestions returns all pending transfer match suggestions
func (h *TransferSuggestionHandler) ListPendingSuggestions(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	suggestions, err := h.suggestionRepo.ListPending(ctx)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Enrich suggestions with full transaction details
	enrichedSuggestions := make([]SuggestionWithTransactions, 0, len(suggestions))
	for _, suggestion := range suggestions {
		txnA, err := h.transactionRepo.GetByID(ctx, suggestion.TransactionAID)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		txnB, err := h.transactionRepo.GetByID(ctx, suggestion.TransactionBID)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		enrichedSuggestions = append(enrichedSuggestions, SuggestionWithTransactions{
			TransferMatchSuggestion: *suggestion,
			TransactionA:            txnA,
			TransactionB:            txnB,
		})
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(enrichedSuggestions)
}

// ListAllSuggestions returns all transfer match suggestions (including accepted/rejected)
func (h *TransferSuggestionHandler) ListAllSuggestions(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	suggestions, err := h.suggestionRepo.List(ctx)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Enrich suggestions with full transaction details
	enrichedSuggestions := make([]SuggestionWithTransactions, 0, len(suggestions))
	for _, suggestion := range suggestions {
		txnA, err := h.transactionRepo.GetByID(ctx, suggestion.TransactionAID)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		txnB, err := h.transactionRepo.GetByID(ctx, suggestion.TransactionBID)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		enrichedSuggestions = append(enrichedSuggestions, SuggestionWithTransactions{
			TransferMatchSuggestion: *suggestion,
			TransactionA:            txnA,
			TransactionB:            txnB,
		})
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(enrichedSuggestions)
}

// AcceptSuggestion accepts a transfer match suggestion
func (h *TransferSuggestionHandler) AcceptSuggestion(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	suggestionID := r.PathValue("id")

	if err := h.linkService.AcceptSuggestion(ctx, suggestionID); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "suggestion accepted and transactions linked"})
}

// RejectSuggestion rejects a transfer match suggestion
func (h *TransferSuggestionHandler) RejectSuggestion(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	suggestionID := r.PathValue("id")

	if err := h.linkService.RejectSuggestion(ctx, suggestionID); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "suggestion rejected"})
}

// ManualLinkRequest represents a request to manually link transactions
type ManualLinkRequest struct {
	TransactionAID string `json:"transaction_a_id"`
	TransactionBID string `json:"transaction_b_id"`
}

// ManualLink manually links two transactions as a transfer
func (h *TransferSuggestionHandler) ManualLink(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var req ManualLinkRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	if req.TransactionAID == "" || req.TransactionBID == "" {
		http.Error(w, "both transaction_a_id and transaction_b_id are required", http.StatusBadRequest)
		return
	}

	if err := h.linkService.ManualLink(ctx, req.TransactionAID, req.TransactionBID); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "transactions manually linked"})
}
