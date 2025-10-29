package domain

import "time"

// BudgetState represents the current state of the budget
// This is a singleton record that tracks values that need to be coordinated
type BudgetState struct {
	ID            string    `json:"id"`
	ReadyToAssign int64     `json:"ready_to_assign"` // Amount available to allocate (in cents)
	UpdatedAt     time.Time `json:"updated_at"`
}
