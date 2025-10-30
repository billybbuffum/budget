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
	allocationRepo    domain.AllocationRepository
	budgetStateRepo   domain.BudgetStateRepository
}

// NewTransactionService creates a new transaction service
func NewTransactionService(
	transactionRepo domain.TransactionRepository,
	accountRepo domain.AccountRepository,
	categoryRepo domain.CategoryRepository,
	allocationRepo domain.AllocationRepository,
	budgetStateRepo domain.BudgetStateRepository,
) *TransactionService {
	return &TransactionService{
		transactionRepo: transactionRepo,
		accountRepo:     accountRepo,
		categoryRepo:    categoryRepo,
		allocationRepo:  allocationRepo,
		budgetStateRepo: budgetStateRepo,
	}
}

// CreateTransaction creates a new transaction and updates account balance
// Handles three types of transactions:
// 1. Normal income (positive amount): Increases account and Ready to Assign
// 2. Normal expense (negative amount): Decreases account, requires category
// 3. Credit card expense: Decreases card balance, moves budget from expense category to payment category
func (s *TransactionService) CreateTransaction(ctx context.Context, accountID string, categoryID *string, amount int64, description string, date time.Time) (*domain.Transaction, error) {
	// Validate account exists
	account, err := s.accountRepo.GetByID(ctx, accountID)
	if err != nil {
		return nil, fmt.Errorf("account not found: %w", err)
	}

	if amount == 0 {
		return nil, fmt.Errorf("amount must be non-zero")
	}

	// For expenses (negative amounts), category is required
	if amount < 0 && (categoryID == nil || *categoryID == "") {
		return nil, fmt.Errorf("category is required for expense transactions")
	}

	// Validate category if provided
	if categoryID != nil && *categoryID != "" {
		if _, err := s.categoryRepo.GetByID(ctx, *categoryID); err != nil {
			return nil, fmt.Errorf("category not found: %w", err)
		}
	}

	transaction := &domain.Transaction{
		ID:          uuid.New().String(),
		Type:        domain.TransactionTypeNormal,
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

	// Handle different transaction types
	if amount > 0 {
		// INCOME: Increase Ready to Assign
		if err := s.budgetStateRepo.AdjustReadyToAssign(ctx, amount); err != nil {
			// Rollback if Ready to Assign update fails
			s.accountRepo.Update(ctx, &domain.Account{
				ID:        account.ID,
				Balance:   account.Balance - amount,
				UpdatedAt: time.Now(),
			})
			s.transactionRepo.Delete(ctx, transaction.ID)
			return nil, fmt.Errorf("failed to adjust ready to assign: %w", err)
		}
	} else if account.Type == domain.AccountTypeCredit && categoryID != nil {
		// CREDIT CARD SPENDING: Move budgeted money from expense category to payment category
		// Only move money that's actually budgeted in the expense category
		// If overspending (spending more than allocated), only move what's available

		// Get the payment category for this credit card
		paymentCategory, err := s.categoryRepo.GetPaymentCategoryByAccountID(ctx, account.ID)
		if err != nil {
			// Rollback
			s.accountRepo.Update(ctx, &domain.Account{
				ID:        account.ID,
				Balance:   account.Balance - amount,
				UpdatedAt: time.Now(),
			})
			s.transactionRepo.Delete(ctx, transaction.ID)
			return nil, fmt.Errorf("failed to get payment category: %w", err)
		}

		// Get current period (YYYY-MM format)
		period := date.Format("2006-01")

		// Get the expense category's allocation to see how much budget is available
		expenseAlloc, err := s.allocationRepo.GetByCategoryAndPeriod(ctx, *categoryID, period)

		// Calculate how much money to move to payment category
		// We only move money that's actually budgeted
		amountToMove := int64(0)

		if err == nil && expenseAlloc != nil && expenseAlloc.Amount > 0 {
			// Get activity (spending) in the expense category for this period
			// BEFORE the current transaction (we need to check what's available NOW)
			startDate := date.AddDate(0, 0, -date.Day()+1) // First day of month
			endDate := startDate.AddDate(0, 1, -1)          // Last day of month

			transactions, err := s.transactionRepo.ListByCategory(ctx, *categoryID)
			if err == nil {
				var totalActivity int64 = 0
				for _, txn := range transactions {
					txnDate := txn.Date
					// Only include transactions BEFORE the one we just created
					// (exclude the current transaction ID)
					if txn.ID != transaction.ID &&
					   (txnDate.After(startDate) || txnDate.Equal(startDate)) &&
					   (txnDate.Before(endDate) || txnDate.Equal(endDate)) {
						totalActivity += txn.Amount
					}
				}

				// Calculate available budget in expense category BEFORE this transaction
				// Available = Allocated + Activity (activity is negative for expenses)
				available := expenseAlloc.Amount + totalActivity

				// Move the minimum of: spending amount or available budget
				spendingAmount := -amount // Convert negative amount to positive
				if available >= spendingAmount {
					// Enough budget available, move full amount
					amountToMove = spendingAmount
				} else if available > 0 {
					// Some budget available, move only what's available
					amountToMove = available
				}
				// If available <= 0, amountToMove stays 0 (no budget to move)
			}
		}
		// If expense category has no allocation, amountToMove stays 0

		// Only update payment category allocation if there's money to move
		if amountToMove > 0 {
			// Get or create allocation for payment category
			paymentAlloc, err := s.allocationRepo.GetByCategoryAndPeriod(ctx, paymentCategory.ID, period)
			if err != nil {
				// Create new allocation
				paymentAlloc = &domain.Allocation{
					ID:         uuid.New().String(),
					CategoryID: paymentCategory.ID,
					Amount:     amountToMove,
					Period:     period,
					CreatedAt:  time.Now(),
					UpdatedAt:  time.Now(),
				}
				if err := s.allocationRepo.Create(ctx, paymentAlloc); err != nil {
					// Rollback
					s.accountRepo.Update(ctx, &domain.Account{
						ID:        account.ID,
						Balance:   account.Balance - amount,
						UpdatedAt: time.Now(),
					})
					s.transactionRepo.Delete(ctx, transaction.ID)
					return nil, fmt.Errorf("failed to create payment allocation: %w", err)
				}
			} else {
				// Update existing allocation
				paymentAlloc.Amount += amountToMove
				paymentAlloc.UpdatedAt = time.Now()
				if err := s.allocationRepo.Update(ctx, paymentAlloc); err != nil {
					// Rollback
					s.accountRepo.Update(ctx, &domain.Account{
						ID:        account.ID,
						Balance:   account.Balance - amount,
						UpdatedAt: time.Now(),
					})
					s.transactionRepo.Delete(ctx, transaction.ID)
					return nil, fmt.Errorf("failed to update payment allocation: %w", err)
				}
			}
		}
		// Note: We don't decrease Ready to Assign because the money was already allocated
		// to the expense category. The allocation just moved from expense → payment category.
	}
	// For regular expense on non-credit accounts, no special handling needed
	// The money was allocated to the expense category and is now spent

	return transaction, nil
}

// CreateTransfer creates a transfer between two accounts
// Transfers move money between accounts without affecting Ready to Assign
// Amount should be positive (the amount to transfer)
func (s *TransactionService) CreateTransfer(ctx context.Context, fromAccountID, toAccountID string, amount int64, description string, date time.Time) (*domain.Transaction, error) {
	if amount <= 0 {
		return nil, fmt.Errorf("transfer amount must be positive")
	}

	if fromAccountID == toAccountID {
		return nil, fmt.Errorf("cannot transfer to the same account")
	}

	// Validate both accounts exist
	fromAccount, err := s.accountRepo.GetByID(ctx, fromAccountID)
	if err != nil {
		return nil, fmt.Errorf("source account not found: %w", err)
	}

	toAccount, err := s.accountRepo.GetByID(ctx, toAccountID)
	if err != nil {
		return nil, fmt.Errorf("destination account not found: %w", err)
	}

	// If transferring TO a credit card, check if we should categorize with payment category
	// Only categorize if there's money allocated (don't categorize overpayments)
	var outboundCategoryID *string
	if toAccount.Type == domain.AccountTypeCredit {
		paymentCategory, err := s.categoryRepo.GetPaymentCategoryByAccountID(ctx, toAccountID)
		if err == nil && paymentCategory != nil {
			// Check if payment category has any allocation
			// Get all allocations for this payment category
			allAllocations, err := s.allocationRepo.List(ctx)
			if err == nil {
				var totalAllocated int64
				for _, alloc := range allAllocations {
					if alloc.CategoryID == paymentCategory.ID {
						totalAllocated += alloc.Amount
					}
				}

				// Get all transactions already categorized with this payment category
				allTransactions, err := s.transactionRepo.ListByCategory(ctx, paymentCategory.ID)
				if err == nil {
					var totalSpent int64
					for _, txn := range allTransactions {
						if txn.Amount < 0 {
							totalSpent += -txn.Amount // Convert to positive
						}
					}

					// Available = Allocated - Already Spent
					available := totalAllocated - totalSpent

					// Only categorize if payment <= available
					// This prevents showing negative available when overpaying
					if available >= amount {
						outboundCategoryID = &paymentCategory.ID
					}
				}
			}
		}
	}

	// Create outbound transaction (negative) from source account
	outboundTxn := &domain.Transaction{
		ID:                  uuid.New().String(),
		Type:                domain.TransactionTypeTransfer,
		AccountID:           fromAccountID,
		TransferToAccountID: &toAccountID,
		CategoryID:          outboundCategoryID, // Categorize with payment category if allocated funds available
		Amount:              -amount, // Negative for outbound
		Description:         description,
		Date:                date,
		CreatedAt:           time.Now(),
		UpdatedAt:           time.Now(),
	}

	if err := s.transactionRepo.Create(ctx, outboundTxn); err != nil {
		return nil, err
	}

	// Create inbound transaction (positive) on destination account
	inboundTxn := &domain.Transaction{
		ID:                  uuid.New().String(),
		Type:                domain.TransactionTypeTransfer,
		AccountID:           toAccountID,
		TransferToAccountID: &fromAccountID, // Link back to source
		Amount:              amount, // Positive for inbound
		Description:         description,
		Date:                date,
		CreatedAt:           time.Now(),
		UpdatedAt:           time.Now(),
	}

	if err := s.transactionRepo.Create(ctx, inboundTxn); err != nil {
		// Rollback outbound transaction
		s.transactionRepo.Delete(ctx, outboundTxn.ID)
		return nil, err
	}

	// Update source account balance (decrease)
	fromAccount.Balance -= amount
	fromAccount.UpdatedAt = time.Now()
	if err := s.accountRepo.Update(ctx, fromAccount); err != nil {
		// Rollback both transactions
		s.transactionRepo.Delete(ctx, inboundTxn.ID)
		s.transactionRepo.Delete(ctx, outboundTxn.ID)
		return nil, fmt.Errorf("failed to update source account balance: %w", err)
	}

	// Update destination account balance (increase)
	toAccount.Balance += amount
	toAccount.UpdatedAt = time.Now()
	if err := s.accountRepo.Update(ctx, toAccount); err != nil {
		// Rollback everything
		fromAccount.Balance += amount // Restore source balance
		s.accountRepo.Update(ctx, fromAccount)
		s.transactionRepo.Delete(ctx, inboundTxn.ID)
		s.transactionRepo.Delete(ctx, outboundTxn.ID)
		return nil, fmt.Errorf("failed to update destination account balance: %w", err)
	}

	// Note: We DON'T adjust Ready to Assign because the money just moved between accounts
	// Total money in the system is the same

	// Return the outbound transaction (the one initiated by the user)
	return outboundTxn, nil
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

	// Update category if provided
	if categoryID != nil {
		if *categoryID != "" {
			if _, err := s.categoryRepo.GetByID(ctx, *categoryID); err != nil {
				return nil, fmt.Errorf("category not found: %w", err)
			}
		}
		oldTransaction.CategoryID = categoryID
	}

	if amount != 0 {
		// Validate category requirement for expenses
		if amount < 0 && (oldTransaction.CategoryID == nil || *oldTransaction.CategoryID == "") {
			return nil, fmt.Errorf("category is required for expense transactions")
		}
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

	// If this was an income transaction, decrease Ready to Assign
	if transaction.Amount > 0 {
		if err := s.budgetStateRepo.AdjustReadyToAssign(ctx, -transaction.Amount); err != nil {
			return fmt.Errorf("failed to adjust ready to assign: %w", err)
		}
	}

	return nil
}

// ListUncategorizedTransactions returns all transactions that don't have a category assigned
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
