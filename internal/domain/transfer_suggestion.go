package domain

import "time"

// TransferSuggestion represents a suggested link between two transactions that may be a transfer
type TransferSuggestion struct {
	ID               string     `json:"id"`
	TransactionAID   string     `json:"transaction_a_id"`
	TransactionBID   string     `json:"transaction_b_id"`
	Confidence       string     `json:"confidence"` // "high", "medium", "low"
	Score            int        `json:"score"`
	IsCreditPayment  bool       `json:"is_credit_payment"`
	Status           string     `json:"status"` // "pending", "accepted", "rejected"
	CreatedAt        time.Time  `json:"created_at"`
	ReviewedAt       *time.Time `json:"reviewed_at,omitempty"`
}

// ConfidenceLevel constants
const (
	ConfidenceHigh   = "high"
	ConfidenceMedium = "medium"
	ConfidenceLow    = "low"
)

// SuggestionStatus constants
const (
	StatusPending  = "pending"
	StatusAccepted = "accepted"
	StatusRejected = "rejected"
)
