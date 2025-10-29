package domain

import "context"

// UserRepository defines the interface for user data operations
type UserRepository interface {
	Create(ctx context.Context, user *User) error
	GetByID(ctx context.Context, id string) (*User, error)
	GetByEmail(ctx context.Context, email string) (*User, error)
	List(ctx context.Context) ([]*User, error)
	Update(ctx context.Context, user *User) error
	Delete(ctx context.Context, id string) error
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
	ListByUser(ctx context.Context, userID string) ([]*Transaction, error)
	ListByCategory(ctx context.Context, categoryID string) ([]*Transaction, error)
	ListByPeriod(ctx context.Context, startDate, endDate string) ([]*Transaction, error)
	Update(ctx context.Context, transaction *Transaction) error
	Delete(ctx context.Context, id string) error
}

// BudgetRepository defines the interface for budget data operations
type BudgetRepository interface {
	Create(ctx context.Context, budget *Budget) error
	GetByID(ctx context.Context, id string) (*Budget, error)
	GetByCategoryAndPeriod(ctx context.Context, categoryID, period string) (*Budget, error)
	ListByPeriod(ctx context.Context, period string) ([]*Budget, error)
	List(ctx context.Context) ([]*Budget, error)
	Update(ctx context.Context, budget *Budget) error
	Delete(ctx context.Context, id string) error
}
