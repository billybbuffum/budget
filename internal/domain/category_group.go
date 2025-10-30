package domain

import "time"

// CategoryGroup represents a grouping of categories (e.g., Housing, Transportation)
// Groups help organize categories for better budget visualization and management
type CategoryGroup struct {
	ID           string       `json:"id"`
	Name         string       `json:"name"`
	Type         CategoryType `json:"type"`         // income or expense (must match contained categories)
	Description  string       `json:"description"`
	DisplayOrder int          `json:"display_order"` // For controlling display order in UI
	CreatedAt    time.Time    `json:"created_at"`
	UpdatedAt    time.Time    `json:"updated_at"`
}
