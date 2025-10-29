package domain

import "time"

// Transaction represents a single income or expense transaction
type Transaction struct {
	ID          string    `json:"id"`
	UserID      string    `json:"user_id"`
	CategoryID  string    `json:"category_id"`
	Amount      float64   `json:"amount"`       // Positive for income, can be positive for expenses too
	Description string    `json:"description"`
	Date        time.Time `json:"date"`         // When the transaction occurred
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}
