package application

import (
	"context"
	"fmt"
	"time"

	"github.com/billybbuffum/budget/internal/domain"
	"github.com/google/uuid"
)

// TransactionService handles transaction-related business logic
type TransactionService struct {
	transactionRepo   domain.TransactionRepository
	accountRepo       domain.AccountRepository
	categoryRepo      domain.CategoryRepository
	budgetStateRepo   domain.BudgetStateRepository
}

// NewTransactionService creates a new transaction service
func NewTransactionService(
	transactionRepo domain.TransactionRepository,
	accountRepo domain.AccountRepository,
	categoryRepo domain.CategoryRepository,
	budgetStateRepo domain.BudgetStateRepository,
) *TransactionService {
	return &TransactionService{
		transactionRepo: transactionRepo,
		accountRepo:     accountRepo,
		categoryRepo:    categoryRepo,
		budgetStateRepo: budgetStateRepo,
	}
}

// CreateTransaction creates a new transaction and updates account balance
// Positive amount = inflow (adds to account), Negative amount = outflow (subtracts from account)
// categoryID can be nil for imported transactions that haven't been categorized yet
func (s *TransactionService) CreateTransaction(ctx context.Context, accountID string, categoryID *string, amount int64, description string, date time.Time) (*domain.Transaction, error) {
	// Validate account exists
	account, err := s.accountRepo.GetByID(ctx, accountID)
	if err != nil {
		return nil, fmt.Errorf("account not found: %w", err)
	}

	// Validate category exists if provided
	if categoryID != nil {
		if _, err := s.categoryRepo.GetByID(ctx, *categoryID); err != nil {
			return nil, fmt.Errorf("category not found: %w", err)
		}
	}

	if amount == 0 {
		return nil, fmt.Errorf("amount must be non-zero")
	}

	transaction := &domain.Transaction{
		ID:          uuid.New().String(),
		AccountID:   accountID,
		CategoryID:  categoryID,
		Amount:      amount,
		Description: description,
		Date:        date,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	if err := s.transactionRepo.Create(ctx, transaction); err != nil {
		return nil, err
	}

	// Update account balance
	account.Balance += amount
	account.UpdatedAt = time.Now()
	if err := s.accountRepo.Update(ctx, account); err != nil {
		// Rollback transaction creation if balance update fails
		s.transactionRepo.Delete(ctx, transaction.ID)
		return nil, fmt.Errorf("failed to update account balance: %w", err)
	}

	return transaction, nil
}

// GetTransaction retrieves a transaction by ID
func (s *TransactionService) GetTransaction(ctx context.Context, id string) (*domain.Transaction, error) {
	return s.transactionRepo.GetByID(ctx, id)
}

// ListTransactions retrieves all transactions
func (s *TransactionService) ListTransactions(ctx context.Context) ([]*domain.Transaction, error) {
	return s.transactionRepo.List(ctx)
}

// ListTransactionsByAccount retrieves transactions for a specific account
func (s *TransactionService) ListTransactionsByAccount(ctx context.Context, accountID string) ([]*domain.Transaction, error) {
	return s.transactionRepo.ListByAccount(ctx, accountID)
}

// ListTransactionsByCategory retrieves transactions for a specific category
func (s *TransactionService) ListTransactionsByCategory(ctx context.Context, categoryID string) ([]*domain.Transaction, error) {
	return s.transactionRepo.ListByCategory(ctx, categoryID)
}

// ListTransactionsByPeriod retrieves transactions within a date range
func (s *TransactionService) ListTransactionsByPeriod(ctx context.Context, startDate, endDate time.Time) ([]*domain.Transaction, error) {
	return s.transactionRepo.ListByPeriod(ctx, startDate.Format(time.RFC3339), endDate.Format(time.RFC3339))
}

// UpdateTransaction updates an existing transaction and adjusts account balance
func (s *TransactionService) UpdateTransaction(ctx context.Context, id, accountID string, categoryID *string, amount int64, description string, date time.Time) (*domain.Transaction, error) {
	// Get existing transaction
	oldTransaction, err := s.transactionRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// Get old account to reverse balance change
	oldAccount, err := s.accountRepo.GetByID(ctx, oldTransaction.AccountID)
	if err != nil {
		return nil, fmt.Errorf("old account not found: %w", err)
	}

	// Reverse old balance change
	oldAccount.Balance -= oldTransaction.Amount
	oldAccount.UpdatedAt = time.Now()

	// Update transaction fields
	if accountID != "" && accountID != oldTransaction.AccountID {
		// Validate new account exists
		newAccount, err := s.accountRepo.GetByID(ctx, accountID)
		if err != nil {
			return nil, fmt.Errorf("new account not found: %w", err)
		}
		// Update old account (remove old transaction amount)
		if err := s.accountRepo.Update(ctx, oldAccount); err != nil {
			return nil, err
		}
		// Update new account (add new transaction amount)
		newAccount.Balance += amount
		newAccount.UpdatedAt = time.Now()
		if err := s.accountRepo.Update(ctx, newAccount); err != nil {
			return nil, err
		}
		oldTransaction.AccountID = accountID
	} else {
		// Same account, just adjust balance difference
		oldAccount.Balance += amount
		if err := s.accountRepo.Update(ctx, oldAccount); err != nil {
			return nil, err
		}
	}

	// Update category if provided (can be nil to clear category)
	if categoryID != nil {
		if *categoryID != "" {
			// Validate category exists if not empty
			if _, err := s.categoryRepo.GetByID(ctx, *categoryID); err != nil {
				return nil, fmt.Errorf("category not found: %w", err)
			}
		}
		oldTransaction.CategoryID = categoryID
	}

	if amount != 0 {
		oldTransaction.Amount = amount
	}

	if description != "" {
		oldTransaction.Description = description
	}

	if !date.IsZero() {
		oldTransaction.Date = date
	}

	oldTransaction.UpdatedAt = time.Now()

	if err := s.transactionRepo.Update(ctx, oldTransaction); err != nil {
		return nil, err
	}

	return oldTransaction, nil
}

// DeleteTransaction deletes a transaction and reverses its effect on account balance
func (s *TransactionService) DeleteTransaction(ctx context.Context, id string) error {
	// Get transaction to know the account and amount
	transaction, err := s.transactionRepo.GetByID(ctx, id)
	if err != nil {
		return err
	}

	// Get account and reverse balance change
	account, err := s.accountRepo.GetByID(ctx, transaction.AccountID)
	if err != nil {
		return fmt.Errorf("account not found: %w", err)
	}

	account.Balance -= transaction.Amount
	account.UpdatedAt = time.Now()

	// Delete transaction first
	if err := s.transactionRepo.Delete(ctx, id); err != nil {
		return err
	}

	// Then update account balance
	if err := s.accountRepo.Update(ctx, account); err != nil {
		return fmt.Errorf("failed to update account balance: %w", err)
	}

	return nil
}

// ListUncategorizedTransactions retrieves all transactions without a category
func (s *TransactionService) ListUncategorizedTransactions(ctx context.Context) ([]*domain.Transaction, error) {
	return s.transactionRepo.ListUncategorized(ctx)
}

// BulkCategorizeTransactions assigns a category to multiple transactions at once
func (s *TransactionService) BulkCategorizeTransactions(ctx context.Context, transactionIDs []string, categoryID *string) error {
	if len(transactionIDs) == 0 {
		return fmt.Errorf("no transaction IDs provided")
	}

	// Validate category exists if provided
	if categoryID != nil && *categoryID != "" {
		if _, err := s.categoryRepo.GetByID(ctx, *categoryID); err != nil {
			return fmt.Errorf("category not found: %w", err)
		}
	}

	return s.transactionRepo.BulkUpdateCategory(ctx, transactionIDs, categoryID)
}
