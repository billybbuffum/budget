package domain

import (
	"context"
	"time"
)

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
	GetPaymentCategoryByAccountID(ctx context.Context, accountID string) (*Category, error)
	List(ctx context.Context) ([]*Category, error)
	ListByGroup(ctx context.Context, groupID string) ([]*Category, error)
	Update(ctx context.Context, category *Category) error
	Delete(ctx context.Context, id string) error
}

// CategoryGroupRepository defines the interface for category group data operations
type CategoryGroupRepository interface {
	Create(ctx context.Context, group *CategoryGroup) error
	GetByID(ctx context.Context, id string) (*CategoryGroup, error)
	List(ctx context.Context) ([]*CategoryGroup, error)
	Update(ctx context.Context, group *CategoryGroup) error
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
	ListUncategorized(ctx context.Context) ([]*Transaction, error)
	GetCategoryActivity(ctx context.Context, categoryID, period string) (int64, error)
	FindDuplicate(ctx context.Context, accountID string, date time.Time, amount int64, description string) (*Transaction, error)
	FindByFitID(ctx context.Context, accountID string, fitID string) (*Transaction, error)
	Update(ctx context.Context, transaction *Transaction) error
	BulkUpdateCategory(ctx context.Context, transactionIDs []string, categoryID *string) error
	Delete(ctx context.Context, id string) error
}

// AllocationRepository defines the interface for allocation data operations
type AllocationRepository interface {
	Create(ctx context.Context, allocation *Allocation) error
	GetByID(ctx context.Context, id string) (*Allocation, error)
	GetByCategoryAndPeriod(ctx context.Context, categoryID, period string) (*Allocation, error)
	ListByPeriod(ctx context.Context, period string) ([]*Allocation, error)
	List(ctx context.Context) ([]*Allocation, error)
	Update(ctx context.Context, allocation *Allocation) error
	Delete(ctx context.Context, id string) error
}

// BudgetStateRepository defines the interface for budget state operations
type BudgetStateRepository interface {
	Get(ctx context.Context) (*BudgetState, error)
	Update(ctx context.Context, state *BudgetState) error
	AdjustReadyToAssign(ctx context.Context, delta int64) error
}
