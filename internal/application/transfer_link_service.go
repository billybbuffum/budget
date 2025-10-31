package application

import (
	"context"
	"fmt"
	"math"
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

// AcceptSuggestion accepts a transfer suggestion and links the transactions
func (s *TransferLinkService) AcceptSuggestion(ctx context.Context, suggestionID string) error {
	// Get the suggestion
	suggestion, err := s.suggestionRepo.GetByID(ctx, suggestionID)
	if err != nil {
		return fmt.Errorf("failed to get suggestion: %w", err)
	}

	// Check status
	if suggestion.Status != domain.StatusPending {
		return fmt.Errorf("suggestion is not pending (status: %s)", suggestion.Status)
	}

	// Get both transactions
	txnA, err := s.transactionRepo.GetByID(ctx, suggestion.TransactionAID)
	if err != nil {
		return fmt.Errorf("failed to get transaction A: %w", err)
	}

	txnB, err := s.transactionRepo.GetByID(ctx, suggestion.TransactionBID)
	if err != nil {
		return fmt.Errorf("failed to get transaction B: %w", err)
	}

	// Validate both are still normal and unlinked
	if txnA.Type != domain.TransactionTypeNormal || txnB.Type != domain.TransactionTypeNormal {
		return fmt.Errorf("one or both transactions are already transfers")
	}
	if txnA.TransferToAccountID != nil || txnB.TransferToAccountID != nil {
		return fmt.Errorf("one or both transactions are already linked")
	}

	// Link the transactions
	if err := s.linkTransactions(ctx, txnA, txnB); err != nil {
		return fmt.Errorf("failed to link transactions: %w", err)
	}

	// Mark suggestion as accepted
	if err := s.suggestionRepo.Accept(ctx, suggestionID); err != nil {
		return fmt.Errorf("failed to mark suggestion as accepted: %w", err)
	}

	// Delete any other suggestions involving these transactions
	if err := s.suggestionRepo.DeleteByTransactionID(ctx, txnA.ID); err != nil {
		// Log but don't fail
	}
	if err := s.suggestionRepo.DeleteByTransactionID(ctx, txnB.ID); err != nil {
		// Log but don't fail
	}

	return nil
}

// RejectSuggestion rejects a transfer suggestion
func (s *TransferLinkService) RejectSuggestion(ctx context.Context, suggestionID string) error {
	// Get the suggestion
	suggestion, err := s.suggestionRepo.GetByID(ctx, suggestionID)
	if err != nil {
		return fmt.Errorf("failed to get suggestion: %w", err)
	}

	// Check status
	if suggestion.Status != domain.StatusPending {
		return fmt.Errorf("suggestion is not pending (status: %s)", suggestion.Status)
	}

	// Mark as rejected
	if err := s.suggestionRepo.Reject(ctx, suggestionID); err != nil {
		return fmt.Errorf("failed to mark suggestion as rejected: %w", err)
	}

	return nil
}

// ManualLink manually links two transactions as a transfer
func (s *TransferLinkService) ManualLink(ctx context.Context, transactionAID, transactionBID string) error {
	// Get both transactions
	txnA, err := s.transactionRepo.GetByID(ctx, transactionAID)
	if err != nil {
		return fmt.Errorf("failed to get transaction A: %w", err)
	}

	txnB, err := s.transactionRepo.GetByID(ctx, transactionBID)
	if err != nil {
		return fmt.Errorf("failed to get transaction B: %w", err)
	}

	// Validate
	if txnA.AccountID == txnB.AccountID {
		return fmt.Errorf("transactions must be in different accounts")
	}

	if txnA.Type != domain.TransactionTypeNormal || txnB.Type != domain.TransactionTypeNormal {
		return fmt.Errorf("both transactions must be normal (not already transfers)")
	}

	if txnA.TransferToAccountID != nil || txnB.TransferToAccountID != nil {
		return fmt.Errorf("one or both transactions are already linked")
	}

	// Link the transactions
	if err := s.linkTransactions(ctx, txnA, txnB); err != nil {
		return fmt.Errorf("failed to link transactions: %w", err)
	}

	return nil
}

// linkTransactions converts two transactions to transfers and links them
func (s *TransferLinkService) linkTransactions(ctx context.Context, txnA, txnB *domain.Transaction) error {
	// Get accounts
	accountA, err := s.accountRepo.GetByID(ctx, txnA.AccountID)
	if err != nil {
		return fmt.Errorf("failed to get account A: %w", err)
	}

	accountB, err := s.accountRepo.GetByID(ctx, txnB.AccountID)
	if err != nil {
		return fmt.Errorf("failed to get account B: %w", err)
	}

	// Determine which is the source (negative) and destination (positive)
	var sourceAccount, destAccount *domain.Account
	var sourceTxn, destTxn *domain.Transaction

	if txnA.Amount < 0 {
		sourceAccount = accountA
		destAccount = accountB
		sourceTxn = txnA
		destTxn = txnB
	} else {
		sourceAccount = accountB
		destAccount = accountA
		sourceTxn = txnB
		destTxn = txnA
	}

	// Check if this is a credit card payment and apply payment category
	var sourceCategoryID *string
	if destAccount.Type == domain.AccountTypeCredit {
		sourceCategoryID = s.applyCreditCardPaymentCategory(ctx, destAccount.ID, math.Abs(float64(sourceTxn.Amount)))
	}

	// Update source transaction
	sourceTxn.Type = domain.TransactionTypeTransfer
	sourceTxn.TransferToAccountID = &destAccount.ID
	sourceTxn.CategoryID = sourceCategoryID
	sourceTxn.UpdatedAt = time.Now()

	if err := s.transactionRepo.Update(ctx, sourceTxn); err != nil {
		return fmt.Errorf("failed to update source transaction: %w", err)
	}

	// Update destination transaction
	destTxn.Type = domain.TransactionTypeTransfer
	destTxn.TransferToAccountID = &sourceAccount.ID
	destTxn.CategoryID = nil // Destination never has category
	destTxn.UpdatedAt = time.Now()

	if err := s.transactionRepo.Update(ctx, destTxn); err != nil {
		// Rollback source transaction
		sourceTxn.Type = domain.TransactionTypeNormal
		sourceTxn.TransferToAccountID = nil
		s.transactionRepo.Update(ctx, sourceTxn)
		return fmt.Errorf("failed to update destination transaction: %w", err)
	}

	return nil
}

// applyCreditCardPaymentCategory applies payment category if there's budget available
func (s *TransferLinkService) applyCreditCardPaymentCategory(ctx context.Context, creditAccountID string, amount float64) *string {
	paymentCategory, err := s.categoryRepo.GetPaymentCategoryByAccountID(ctx, creditAccountID)
	if err != nil || paymentCategory == nil {
		return nil
	}

	// Get all allocations for this payment category
	allAllocations, err := s.allocationRepo.List(ctx)
	if err != nil {
		return nil
	}

	var totalAllocated int64
	for _, alloc := range allAllocations {
		if alloc.CategoryID == paymentCategory.ID {
			totalAllocated += alloc.Amount
		}
	}

	// Get all transactions already categorized with this payment category
	allTransactions, err := s.transactionRepo.ListByCategory(ctx, paymentCategory.ID)
	if err != nil {
		return nil
	}

	var totalSpent int64
	for _, txn := range allTransactions {
		if txn.Amount < 0 {
			totalSpent += -txn.Amount // Convert to positive
		}
	}

	// Available = Allocated - Already Spent
	available := totalAllocated - totalSpent

	// Only categorize if payment <= available
	paymentAmount := int64(amount)
	if available >= paymentAmount {
		return &paymentCategory.ID
	}

	return nil
}
