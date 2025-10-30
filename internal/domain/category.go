package domain

import "time"

// Category represents a budget category for expense tracking and budgeting
// All categories can receive budget allocations
// Income transactions don't require a category - they just increase Ready to Assign
// Payment categories are automatically created for credit card accounts
type Category struct {
	ID                  string    `json:"id"`
	Name                string    `json:"name"`
	Description         string    `json:"description"`
	Color               string    `json:"color"`               // Hex color for UI
	PaymentForAccountID *string   `json:"payment_for_account_id,omitempty"` // If set, this is a payment category for a credit card
	CreatedAt           time.Time `json:"created_at"`
	UpdatedAt           time.Time `json:"updated_at"`
}
