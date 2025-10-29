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
	transactionRepo domain.TransactionRepository
	userRepo        domain.UserRepository
	categoryRepo    domain.CategoryRepository
}

// NewTransactionService creates a new transaction service
func NewTransactionService(
	transactionRepo domain.TransactionRepository,
	userRepo domain.UserRepository,
	categoryRepo domain.CategoryRepository,
) *TransactionService {
	return &TransactionService{
		transactionRepo: transactionRepo,
		userRepo:        userRepo,
		categoryRepo:    categoryRepo,
	}
}

// CreateTransaction creates a new transaction
func (s *TransactionService) CreateTransaction(ctx context.Context, userID, categoryID string, amount float64, description string, date time.Time) (*domain.Transaction, error) {
	// Validate user exists
	if _, err := s.userRepo.GetByID(ctx, userID); err != nil {
		return nil, fmt.Errorf("user not found: %w", err)
	}

	// Validate category exists
	if _, err := s.categoryRepo.GetByID(ctx, categoryID); err != nil {
		return nil, fmt.Errorf("category not found: %w", err)
	}

	if amount == 0 {
		return nil, fmt.Errorf("amount must be non-zero")
	}

	transaction := &domain.Transaction{
		ID:          uuid.New().String(),
		UserID:      userID,
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

// ListTransactionsByUser retrieves transactions for a specific user
func (s *TransactionService) ListTransactionsByUser(ctx context.Context, userID string) ([]*domain.Transaction, error) {
	return s.transactionRepo.ListByUser(ctx, userID)
}

// ListTransactionsByCategory retrieves transactions for a specific category
func (s *TransactionService) ListTransactionsByCategory(ctx context.Context, categoryID string) ([]*domain.Transaction, error) {
	return s.transactionRepo.ListByCategory(ctx, categoryID)
}

// ListTransactionsByPeriod retrieves transactions within a date range
func (s *TransactionService) ListTransactionsByPeriod(ctx context.Context, startDate, endDate time.Time) ([]*domain.Transaction, error) {
	return s.transactionRepo.ListByPeriod(ctx, startDate.Format(time.RFC3339), endDate.Format(time.RFC3339))
}

// UpdateTransaction updates an existing transaction
func (s *TransactionService) UpdateTransaction(ctx context.Context, id, userID, categoryID string, amount float64, description string, date time.Time) (*domain.Transaction, error) {
	transaction, err := s.transactionRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if userID != "" {
		if _, err := s.userRepo.GetByID(ctx, userID); err != nil {
			return nil, fmt.Errorf("user not found: %w", err)
		}
		transaction.UserID = userID
	}

	if categoryID != "" {
		if _, err := s.categoryRepo.GetByID(ctx, categoryID); err != nil {
			return nil, fmt.Errorf("category not found: %w", err)
		}
		transaction.CategoryID = categoryID
	}

	if amount != 0 {
		transaction.Amount = amount
	}

	if description != "" {
		transaction.Description = description
	}

	if !date.IsZero() {
		transaction.Date = date
	}

	transaction.UpdatedAt = time.Now()

	if err := s.transactionRepo.Update(ctx, transaction); err != nil {
		return nil, err
	}

	return transaction, nil
}

// DeleteTransaction deletes a transaction
func (s *TransactionService) DeleteTransaction(ctx context.Context, id string) error {
	return s.transactionRepo.Delete(ctx, id)
}
