package domain

import "time"

// CategoryType represents whether a category group is for income or expenses
// Note: Individual categories no longer have types, only groups do
type CategoryType string

const (
	CategoryTypeIncome  CategoryType = "income"
	CategoryTypeExpense CategoryType = "expense"
)

// Category represents a budget category for expense tracking and budgeting
// All categories can receive budget allocations
// Income transactions don't require a category - they just increase Ready to Assign
// Payment categories are automatically created for credit card accounts
type Category struct {
	ID                  string    `json:"id"`
	Name                string    `json:"name"`
	Description         string    `json:"description"`
	Color               string    `json:"color"`                            // Hex color for UI
	GroupID             *string   `json:"group_id,omitempty"`               // Optional reference to category group
	PaymentForAccountID *string   `json:"payment_for_account_id,omitempty"` // If set, this is a payment category for a credit card
	CreatedAt           time.Time `json:"created_at"`
	UpdatedAt           time.Time `json:"updated_at"`
}
