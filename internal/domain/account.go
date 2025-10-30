package domain

import "time"

// AccountType represents the type of account
type AccountType string

const (
	AccountTypeChecking AccountType = "checking"
	AccountTypeSavings  AccountType = "savings"
	AccountTypeCash     AccountType = "cash"
	AccountTypeCredit   AccountType = "credit" // Credit cards - negative balance = debt
)

// Account represents a financial account that holds money
type Account struct {
	ID        string      `json:"id"`
	Name      string      `json:"name"`
	Balance   int64       `json:"balance"`    // Balance in cents
	Type      AccountType `json:"type"`
	CreatedAt time.Time   `json:"created_at"`
	UpdatedAt time.Time   `json:"updated_at"`
}
