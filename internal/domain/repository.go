package domain

import "context"

// AccountRepository defines the interface for account data operations
type AccountRepository interface {
	Create(ctx context.Context, account *Account) error
	GetByID(ctx context.Context, id string) (*Account, error)
	List(ctx context.Context) ([]*Account, error)
	Update(ctx context.Context, account *Account) error
	Delete(ctx context.Context, id string) error
	GetTotalBalance(ctx context.Context) (int64, error)
}

// CategoryRepository defines the interface for category data operations
type CategoryRepository interface {
	Create(ctx context.Context, category *Category) error
	GetByID(ctx context.Context, id string) (*Category, error)
	List(ctx context.Context) ([]*Category, error)
	ListByType(ctx context.Context, categoryType CategoryType) ([]*Category, error)
	Update(ctx context.Context, category *Category) error
	Delete(ctx context.Context, id string) error
}

// TransactionRepository defines the interface for transaction data operations
type TransactionRepository interface {
	Create(ctx context.Context, transaction *Transaction) error
	GetByID(ctx context.Context, id string) (*Transaction, error)
	List(ctx context.Context) ([]*Transaction, error)
	ListByAccount(ctx context.Context, accountID string) ([]*Transaction, error)
	ListByCategory(ctx context.Context, categoryID string) ([]*Transaction, error)
	ListByPeriod(ctx context.Context, startDate, endDate string) ([]*Transaction, error)
	GetCategoryActivity(ctx context.Context, categoryID, period string) (int64, error)
	Update(ctx context.Context, transaction *Transaction) error
	Delete(ctx context.Context, id string) error
}

// AllocationRepository defines the interface for allocation data operations
type AllocationRepository interface {
	Create(ctx context.Context, allocation *Allocation) error
	GetByID(ctx context.Context, id string) (*Allocation, error)
	GetByCategoryAndPeriod(ctx context.Context, categoryID, period string) (*Allocation, error)
	ListByPeriod(ctx context.Context, period string) ([]*Allocation, error)
	List(ctx context.Context) ([]*Allocation, error)
	GetTotalAllocated(ctx context.Context) (int64, error)
	Update(ctx context.Context, allocation *Allocation) error
	Delete(ctx context.Context, id string) error
}

// BudgetStateRepository defines the interface for budget state operations
type BudgetStateRepository interface {
	Get(ctx context.Context) (*BudgetState, error)
	Update(ctx context.Context, state *BudgetState) error
	AdjustReadyToAssign(ctx context.Context, delta int64) error
}
