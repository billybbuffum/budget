package domain

import "time"

// TransactionType represents the type of transaction
type TransactionType string

const (
	TransactionTypeNormal   TransactionType = "normal"   // Regular income or expense
	TransactionTypeTransfer TransactionType = "transfer" // Transfer between accounts
)

// Transaction represents a single financial transaction
// Normal transactions:
//   - Positive amounts = Inflows (income) - CategoryID optional
//   - Negative amounts = Outflows (expenses) - CategoryID required
// Transfer transactions:
//   - Move money between accounts
//   - No category needed
//   - Amount is negative on source account
type Transaction struct {
	ID                  string           `json:"id"`
	Type                TransactionType  `json:"type"`                          // normal or transfer
	AccountID           string           `json:"account_id"`                    // Source account
	TransferToAccountID *string          `json:"transfer_to_account_id,omitempty"` // Destination account (transfers only)
	CategoryID          *string          `json:"category_id,omitempty"`         // Category (normal transactions only)
	Amount              int64            `json:"amount"`                        // Amount in cents
	Description         string           `json:"description"`
	Date                time.Time        `json:"date"`
	CreatedAt           time.Time        `json:"created_at"`
	UpdatedAt           time.Time        `json:"updated_at"`
}
