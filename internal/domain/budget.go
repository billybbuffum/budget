package domain

import "time"

// Budget represents a budget allocation for a category in a specific period
type Budget struct {
	ID           string    `json:"id"`
	CategoryID   string    `json:"category_id"`
	Amount       float64   `json:"amount"`        // Budgeted amount
	Period       string    `json:"period"`        // e.g., "2024-01" for monthly budget
	Notes        string    `json:"notes"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// BudgetSummary provides a summary view of budget vs actual spending
type BudgetSummary struct {
	Budget       *Budget  `json:"budget"`
	Category     *Category `json:"category"`
	ActualSpent  float64  `json:"actual_spent"`
	Remaining    float64  `json:"remaining"`
	PercentUsed  float64  `json:"percent_used"`
}
