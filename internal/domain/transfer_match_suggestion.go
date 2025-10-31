package domain

import "time"

// ConfidenceLevel represents the confidence level of a transfer match
type ConfidenceLevel string

const (
	ConfidenceLevelHigh   ConfidenceLevel = "high"   // Score >= 15
	ConfidenceLevelMedium ConfidenceLevel = "medium" // Score 10-14
	ConfidenceLevelLow    ConfidenceLevel = "low"    // Score < 10
)

// SuggestionStatus represents the status of a transfer match suggestion
type SuggestionStatus string

const (
	SuggestionStatusPending  SuggestionStatus = "pending"  // Awaiting user review
	SuggestionStatusAccepted SuggestionStatus = "accepted" // User accepted the match
	SuggestionStatusRejected SuggestionStatus = "rejected" // User rejected the match
)

// TransferMatchSuggestion represents a potential transfer match between two transactions
type TransferMatchSuggestion struct {
	ID              string           `json:"id"`
	TransactionAID  string           `json:"transaction_a_id"`  // First transaction (typically outflow)
	TransactionBID  string           `json:"transaction_b_id"`  // Second transaction (typically inflow)
	MatchScore      int              `json:"match_score"`       // Calculated match score
	Confidence      ConfidenceLevel  `json:"confidence"`        // high, medium, or low
	Status          SuggestionStatus `json:"status"`            // pending, accepted, or rejected
	IsCreditPayment bool             `json:"is_credit_payment"` // True if one account is credit card
	CreatedAt       time.Time        `json:"created_at"`
	UpdatedAt       time.Time        `json:"updated_at"`
}
