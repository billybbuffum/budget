package domain

import "time"

const (
	// CreditCardPaymentsGroupName is the name of the special group that contains
	// all credit card payment categories. This group is automatically managed and
	// cannot be renamed or deleted by users.
	CreditCardPaymentsGroupName = "Credit Card Payments"
)

// CategoryGroup represents a grouping of categories for budget organization
// Groups are used purely for visual organization on the budget page
// Examples: Housing, Transportation, Entertainment, etc.
type CategoryGroup struct {
	ID           string    `json:"id"`
	Name         string    `json:"name"`
	Description  string    `json:"description"`
	DisplayOrder int       `json:"display_order"` // For controlling display order in UI
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}
