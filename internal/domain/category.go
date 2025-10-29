package domain

import "time"

// CategoryType represents whether a category is for income or expenses
type CategoryType string

const (
	CategoryTypeIncome  CategoryType = "income"
	CategoryTypeExpense CategoryType = "expense"
)

// Category represents a budget category (e.g., groceries, rent, salary)
type Category struct {
	ID          string       `json:"id"`
	Name        string       `json:"name"`
	Type        CategoryType `json:"type"`
	Description string       `json:"description"`
	Color       string       `json:"color"` // Hex color for UI
	CreatedAt   time.Time    `json:"created_at"`
	UpdatedAt   time.Time    `json:"updated_at"`
}
