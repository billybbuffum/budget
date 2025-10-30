package domain

import "time"

// CategoryType represents whether a category is for income or expenses
type CategoryType string

const (
	CategoryTypeIncome  CategoryType = "income"
	CategoryTypeExpense CategoryType = "expense"
)

// Category represents a budget category (e.g., groceries, rent, salary)
// Income categories: for tracking where money comes from (no allocations)
// Expense categories: can have allocations assigned to them
type Category struct {
	ID          string       `json:"id"`
	Name        string       `json:"name"`
	Type        CategoryType `json:"type"`        // income or expense
	Description string       `json:"description"`
	Color       string       `json:"color"`       // Hex color for UI
	GroupID     *string      `json:"group_id"`    // Optional reference to category group
	CreatedAt   time.Time    `json:"created_at"`
	UpdatedAt   time.Time    `json:"updated_at"`
}
