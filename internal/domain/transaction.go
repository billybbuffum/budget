package domain

import "time"

// TransactionType represents the type of transaction
type TransactionType string

const (
	TransactionTypeNormal   TransactionType = "normal"   // Regular inflow or outflow
	TransactionTypeTransfer TransactionType = "transfer" // Transfer between accounts
)

// Transaction represents a single financial transaction
// Normal transactions:
//   - Positive amounts = Inflows - CategoryID optional
//   - Negative amounts = Outflows - CategoryID required
// Transfer transactions:
//   - Move money between accounts
//   - No category needed
//   - Amount is negative on source account
type Transaction struct {
	ID                  string           `json:"id"`
	Type                TransactionType  `json:"type"`                             // normal or transfer
	AccountID           string           `json:"account_id"`                       // Source account
	TransferToAccountID *string          `json:"transfer_to_account_id,omitempty"` // Destination account (transfers only)
	CategoryID          *string          `json:"category_id,omitempty"`            // Category (normal transactions only, nullable for imports)
	Amount              int64            `json:"amount"`                           // Amount in cents (positive=inflow, negative=outflow)
	Description         string           `json:"description"`
	Date                time.Time        `json:"date"`                             // When the transaction occurred
	FitID               *string          `json:"fitid,omitempty"`                  // Financial Institution Transaction ID (for OFX imports, duplicate detection)
	CreatedAt           time.Time        `json:"created_at"`
	UpdatedAt           time.Time        `json:"updated_at"`
}
