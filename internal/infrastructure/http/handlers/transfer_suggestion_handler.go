package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/billybbuffum/budget/internal/application"
	"github.com/billybbuffum/budget/internal/domain"
)

type TransferSuggestionHandler struct {
	suggestionRepo domain.TransferSuggestionRepository
	transactionRepo domain.TransactionRepository
	accountRepo    domain.AccountRepository
	linkService    *application.TransferLinkService
}

func NewTransferSuggestionHandler(
	suggestionRepo domain.TransferSuggestionRepository,
	transactionRepo domain.TransactionRepository,
	accountRepo domain.AccountRepository,
	linkService *application.TransferLinkService,
) *TransferSuggestionHandler {
	return &TransferSuggestionHandler{
		suggestionRepo:  suggestionRepo,
		transactionRepo: transactionRepo,
		accountRepo:     accountRepo,
		linkService:     linkService,
	}
}

// SuggestionWithDetails includes the full transaction and account details
type SuggestionWithDetails struct {
	ID              string                  `json:"id"`
	TransactionA    *domain.Transaction     `json:"transaction_a"`
	TransactionB    *domain.Transaction     `json:"transaction_b"`
	AccountA        *domain.Account         `json:"account_a"`
	AccountB        *domain.Account         `json:"account_b"`
	Confidence      string                  `json:"confidence"`
	Score           int                     `json:"score"`
	IsCreditPayment bool                    `json:"is_credit_payment"`
	CreatedAt       string                  `json:"created_at"`
}

// ListSuggestions returns all pending transfer suggestions with details
func (h *TransferSuggestionHandler) ListSuggestions(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Get query parameters
	status := r.URL.Query().Get("status")
	confidence := r.URL.Query().Get("confidence")

	var suggestions []*domain.TransferSuggestion
	var err error

	// Filter by status or confidence if provided
	if status == "pending" || status == "" {
		suggestions, err = h.suggestionRepo.ListPending(ctx)
	} else if confidence != "" {
		suggestions, err = h.suggestionRepo.ListByConfidence(ctx, confidence)
	} else {
		suggestions, err = h.suggestionRepo.List(ctx)
	}

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Enrich with transaction and account details
	var enrichedSuggestions []SuggestionWithDetails
	for _, suggestion := range suggestions {
		txnA, err := h.transactionRepo.GetByID(ctx, suggestion.TransactionAID)
		if err != nil {
			continue
		}

		txnB, err := h.transactionRepo.GetByID(ctx, suggestion.TransactionBID)
		if err != nil {
			continue
		}

		accountA, err := h.accountRepo.GetByID(ctx, txnA.AccountID)
		if err != nil {
			continue
		}

		accountB, err := h.accountRepo.GetByID(ctx, txnB.AccountID)
		if err != nil {
			continue
		}

		enrichedSuggestions = append(enrichedSuggestions, SuggestionWithDetails{
			ID:              suggestion.ID,
			TransactionA:    txnA,
			TransactionB:    txnB,
			AccountA:        accountA,
			AccountB:        accountB,
			Confidence:      suggestion.Confidence,
			Score:           suggestion.Score,
			IsCreditPayment: suggestion.IsCreditPayment,
			CreatedAt:       suggestion.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		})
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"suggestions": enrichedSuggestions,
	})
}

// AcceptSuggestion accepts a suggestion and links the transactions
func (h *TransferSuggestionHandler) AcceptSuggestion(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	suggestionID := r.PathValue("id")

	if suggestionID == "" {
		http.Error(w, "suggestion ID is required", http.StatusBadRequest)
		return
	}

	// Accept the suggestion (links the transactions)
	if err := h.linkService.AcceptSuggestion(ctx, suggestionID); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Get the suggestion to return linked transactions
	suggestion, err := h.suggestionRepo.GetByID(ctx, suggestionID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	txnA, _ := h.transactionRepo.GetByID(ctx, suggestion.TransactionAID)
	txnB, _ := h.transactionRepo.GetByID(ctx, suggestion.TransactionBID)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"linked_transactions": []*domain.Transaction{txnA, txnB},
	})
}

// RejectSuggestion rejects a suggestion
func (h *TransferSuggestionHandler) RejectSuggestion(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	suggestionID := r.PathValue("id")

	if suggestionID == "" {
		http.Error(w, "suggestion ID is required", http.StatusBadRequest)
		return
	}

	// Reject the suggestion
	if err := h.linkService.RejectSuggestion(ctx, suggestionID); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
	})
}

// ManualLinkRequest is the request body for manual linking
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
		http.Error(w, "both transaction IDs are required", http.StatusBadRequest)
		return
	}

	// Link the transactions
	if err := h.linkService.ManualLink(ctx, req.TransactionAID, req.TransactionBID); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	txnA, _ := h.transactionRepo.GetByID(ctx, req.TransactionAID)
	txnB, _ := h.transactionRepo.GetByID(ctx, req.TransactionBID)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"linked_transactions": []*domain.Transaction{txnA, txnB},
	})
}
