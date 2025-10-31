package application

import (
	"context"
	"fmt"
	"time"

	"github.com/billybbuffum/budget/internal/domain"
)

// TransferLinkService handles linking transactions as transfers
type TransferLinkService struct {
	transactionRepo domain.TransactionRepository
	accountRepo     domain.AccountRepository
	categoryRepo    domain.CategoryRepository
	allocationRepo  domain.AllocationRepository
	suggestionRepo  domain.TransferSuggestionRepository
}

// NewTransferLinkService creates a new transfer link service
func NewTransferLinkService(
	transactionRepo domain.TransactionRepository,
	accountRepo domain.AccountRepository,
	categoryRepo domain.CategoryRepository,
	allocationRepo domain.AllocationRepository,
	suggestionRepo domain.TransferSuggestionRepository,
) *TransferLinkService {
	return &TransferLinkService{
		transactionRepo: transactionRepo,
		accountRepo:     accountRepo,
		categoryRepo:    categoryRepo,
		allocationRepo:  allocationRepo,
		suggestionRepo:  suggestionRepo,
	}
}

// AcceptSuggestion accepts a transfer match suggestion and links the transactions
func (s *TransferLinkService) AcceptSuggestion(ctx context.Context, suggestionID string) error {
	// Get suggestion
	suggestion, err := s.suggestionRepo.GetByID(ctx, suggestionID)
	if err != nil {
		return fmt.Errorf("suggestion not found: %w", err)
	}

	if suggestion.Status != domain.SuggestionStatusPending {
		return fmt.Errorf("suggestion already processed with status: %s", suggestion.Status)
	}

	// Link the transactions
	if err := s.linkTransactions(ctx, suggestion.TransactionAID, suggestion.TransactionBID, suggestion.IsCreditPayment); err != nil {
		return fmt.Errorf("failed to link transactions: %w", err)
	}

	// Update suggestion status
	suggestion.Status = domain.SuggestionStatusAccepted
	suggestion.UpdatedAt = time.Now().UTC()
	if err := s.suggestionRepo.Update(ctx, suggestion); err != nil {
		return fmt.Errorf("failed to update suggestion status: %w", err)
	}

	return nil
}

// RejectSuggestion rejects a transfer match suggestion
func (s *TransferLinkService) RejectSuggestion(ctx context.Context, suggestionID string) error {
	// Get suggestion
	suggestion, err := s.suggestionRepo.GetByID(ctx, suggestionID)
	if err != nil {
		return fmt.Errorf("suggestion not found: %w", err)
	}

	if suggestion.Status != domain.SuggestionStatusPending {
		return fmt.Errorf("suggestion already processed with status: %s", suggestion.Status)
	}

	// Update suggestion status
	suggestion.Status = domain.SuggestionStatusRejected
	suggestion.UpdatedAt = time.Now().UTC()
	if err := s.suggestionRepo.Update(ctx, suggestion); err != nil {
		return fmt.Errorf("failed to update suggestion status: %w", err)
	}

	return nil
}

// ManualLink manually links two transactions as a transfer
func (s *TransferLinkService) ManualLink(ctx context.Context, txnAID, txnBID string) error {
	// Validate both transactions exist
	txnA, err := s.transactionRepo.GetByID(ctx, txnAID)
	if err != nil {
		return fmt.Errorf("transaction A not found: %w", err)
	}
	txnB, err := s.transactionRepo.GetByID(ctx, txnBID)
	if err != nil {
		return fmt.Errorf("transaction B not found: %w", err)
	}

	// Validate linking rules
	if err := s.validateLinkRules(ctx, txnA, txnB); err != nil {
		return err
	}

	// Check if this is a credit card payment
	accountA, err := s.accountRepo.GetByID(ctx, txnA.AccountID)
	if err != nil {
		return fmt.Errorf("account A not found: %w", err)
	}
	accountB, err := s.accountRepo.GetByID(ctx, txnB.AccountID)
	if err != nil {
		return fmt.Errorf("account B not found: %w", err)
	}
	isCreditPayment := accountA.Type == domain.AccountTypeCredit || accountB.Type == domain.AccountTypeCredit

	// Link the transactions
	return s.linkTransactions(ctx, txnAID, txnBID, isCreditPayment)
}

// validateLinkRules validates that two transactions can be linked
func (s *TransferLinkService) validateLinkRules(ctx context.Context, txnA, txnB *domain.Transaction) error {
	// Both must be normal transactions
	if txnA.Type != domain.TransactionTypeNormal {
		return fmt.Errorf("transaction A is already a transfer")
	}
	if txnB.Type != domain.TransactionTypeNormal {
		return fmt.Errorf("transaction B is already a transfer")
	}

	// Must be different accounts
	if txnA.AccountID == txnB.AccountID {
		return fmt.Errorf("cannot link transactions from the same account")
	}

	// Must have opposite amounts
	if txnA.Amount+txnB.Amount != 0 {
		return fmt.Errorf("transaction amounts must be opposite (±X)")
	}

	return nil
}

// linkTransactions converts two normal transactions to transfer transactions
func (s *TransferLinkService) linkTransactions(ctx context.Context, txnAID, txnBID string, isCreditPayment bool) error {
	// Get both transactions
	txnA, err := s.transactionRepo.GetByID(ctx, txnAID)
	if err != nil {
		return fmt.Errorf("transaction A not found: %w", err)
	}
	txnB, err := s.transactionRepo.GetByID(ctx, txnBID)
	if err != nil {
		return fmt.Errorf("transaction B not found: %w", err)
	}

	// Determine which is outflow and which is inflow
	var outflowTxn, inflowTxn *domain.Transaction
	if txnA.Amount < 0 {
		outflowTxn = txnA
		inflowTxn = txnB
	} else {
		outflowTxn = txnB
		inflowTxn = txnA
	}

	// Handle credit card payment categorization
	var paymentCategoryID *string
	if isCreditPayment {
		// Find which account is the credit card
		accountA, err := s.accountRepo.GetByID(ctx, txnA.AccountID)
		if err != nil {
			return fmt.Errorf("account A not found: %w", err)
		}
		accountB, err := s.accountRepo.GetByID(ctx, txnB.AccountID)
		if err != nil {
			return fmt.Errorf("account B not found: %w", err)
		}

		var creditAccountID string
		var checkingTxn *domain.Transaction
		if accountA.Type == domain.AccountTypeCredit {
			creditAccountID = accountA.ID
			checkingTxn = txnB // The other transaction is from checking
		} else {
			creditAccountID = accountB.ID
			checkingTxn = txnA
		}

		// Get payment category for the credit card
		paymentCategory, err := s.categoryRepo.GetPaymentCategoryByAccountID(ctx, creditAccountID)
		if err == nil && paymentCategory != nil {
			// Check if we have available budget for this payment
			amount := -checkingTxn.Amount // Convert to positive
			if s.hasAvailableBudget(ctx, paymentCategory.ID, amount) {
				paymentCategoryID = &paymentCategory.ID
			}
			// If no available budget, paymentCategoryID stays nil (overpayment case)
		}
	}

	// Update outflow transaction
	outflowTxn.Type = domain.TransactionTypeTransfer
	outflowTxn.TransferToAccountID = &inflowTxn.AccountID
	if isCreditPayment && paymentCategoryID != nil && outflowTxn.Amount < 0 {
		outflowTxn.CategoryID = paymentCategoryID // Apply payment category to outflow (checking side)
	} else {
		outflowTxn.CategoryID = nil // Remove category for normal transfers
	}
	outflowTxn.UpdatedAt = time.Now().UTC()

	if err := s.transactionRepo.Update(ctx, outflowTxn); err != nil {
		return fmt.Errorf("failed to update outflow transaction: %w", err)
	}

	// Update inflow transaction
	inflowTxn.Type = domain.TransactionTypeTransfer
	inflowTxn.TransferToAccountID = &outflowTxn.AccountID
	inflowTxn.CategoryID = nil // Inflow side never has category
	inflowTxn.UpdatedAt = time.Now().UTC()

	if err := s.transactionRepo.Update(ctx, inflowTxn); err != nil {
		// Rollback outflow transaction update
		outflowTxn.Type = domain.TransactionTypeNormal
		outflowTxn.TransferToAccountID = nil
		s.transactionRepo.Update(ctx, outflowTxn)
		return fmt.Errorf("failed to update inflow transaction: %w", err)
	}

	return nil
}

// hasAvailableBudget checks if a payment category has available budget
func (s *TransferLinkService) hasAvailableBudget(ctx context.Context, categoryID string, amount int64) bool {
	// Get all allocations for this category
	allAllocations, err := s.allocationRepo.List(ctx)
	if err != nil {
		return false
	}

	var totalAllocated int64
	for _, alloc := range allAllocations {
		if alloc.CategoryID == categoryID {
			totalAllocated += alloc.Amount
		}
	}

	// Get all transactions already categorized
	allTransactions, err := s.transactionRepo.ListByCategory(ctx, categoryID)
	if err != nil {
		return false
	}

	var totalSpent int64
	for _, txn := range allTransactions {
		if txn.Amount < 0 {
			totalSpent += -txn.Amount // Convert to positive
		}
	}

	// Available = Allocated - Already Spent
	available := totalAllocated - totalSpent

	// Check if payment amount <= available
	return available >= amount
}
