package domain

import "time"

// Allocation represents money assigned to a category for a specific period
type Allocation struct {
	ID         string    `json:"id"`
	CategoryID string    `json:"category_id"`
	Period     string    `json:"period"`      // Format: YYYY-MM
	Amount     int64     `json:"amount"`      // Allocated amount in cents
	Notes      string    `json:"notes"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

// AllocationSummary provides a summary view of allocation vs actual spending
type AllocationSummary struct {
	Allocation *Allocation `json:"allocation"`
	Category   *Category   `json:"category"`
	Activity   int64       `json:"activity"`   // Sum of transactions (negative for spending)
	Available  int64       `json:"available"`  // Allocated + Activity (Activity is negative)
}
