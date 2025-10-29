package domain

import "time"

// Transaction represents a single income or expense transaction
// Positive amounts = Inflows (money coming in)
// Negative amounts = Outflows (money going out/expenses)
type Transaction struct {
	ID          string    `json:"id"`
	AccountID   string    `json:"account_id"`              // Which account this transaction affects
	CategoryID  *string   `json:"category_id,omitempty"`   // Category for tracking/allocation (nullable for imported transactions)
	Amount      int64     `json:"amount"`                  // Amount in cents (positive=inflow, negative=outflow)
	Description string    `json:"description"`
	Date        time.Time `json:"date"`                    // When the transaction occurred
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}
