package application

import (
	"context"
	"fmt"
	"io"
	"time"

	"github.com/billybbuffum/budget/internal/domain"
	"github.com/billybbuffum/budget/internal/infrastructure/ofx"
	"github.com/google/uuid"
)

// ImportService handles transaction import logic
type ImportService struct {
	transactionRepo domain.TransactionRepository
	accountRepo     domain.AccountRepository
	budgetStateRepo domain.BudgetStateRepository
	ofxParser       *ofx.Parser
}

// NewImportService creates a new import service
func NewImportService(
	transactionRepo domain.TransactionRepository,
	accountRepo domain.AccountRepository,
	budgetStateRepo domain.BudgetStateRepository,
	ofxParser *ofx.Parser,
) *ImportService {
	return &ImportService{
		transactionRepo: transactionRepo,
		accountRepo:     accountRepo,
		budgetStateRepo: budgetStateRepo,
		ofxParser:       ofxParser,
	}
}

// ImportResult contains the result of an import operation
type ImportResult struct {
	TotalTransactions     int      `json:"total_transactions"`
	ImportedTransactions  int      `json:"imported_transactions"`
	SkippedDuplicates     int      `json:"skipped_duplicates"`
	Errors                []string `json:"errors,omitempty"`
	NewAccountBalance     int64    `json:"new_account_balance"`
	ImportedTransactionIDs []string `json:"imported_transaction_ids"`
}

// ImportFromOFX imports transactions from an OFX file
func (s *ImportService) ImportFromOFX(ctx context.Context, accountID string, reader io.Reader) (*ImportResult, error) {
	// Validate account exists
	account, err := s.accountRepo.GetByID(ctx, accountID)
	if err != nil {
		return nil, fmt.Errorf("account not found: %w", err)
	}

	// Parse OFX file (extracts ledger balance + last 90 days of transactions)
	parseResult, err := s.ofxParser.Parse(reader)
	if err != nil {
		return nil, fmt.Errorf("failed to parse OFX file: %w", err)
	}

	result := &ImportResult{
		TotalTransactions:      len(parseResult.Transactions),
		ImportedTransactions:   0,
		SkippedDuplicates:      0,
		Errors:                 []string{},
		ImportedTransactionIDs: []string{},
	}

	// Calculate balance delta using ledger balance from OFX file
	// This is the authoritative balance from the bank
	balanceDelta := int64(0)
	if parseResult.LedgerBalance != 0 {
		balanceDelta = parseResult.LedgerBalance - account.Balance
	}

	// Process each transaction (for categorization purposes only)
	// These transactions do NOT affect account balance since we're using ledger balance
	for _, ofxTxn := range parseResult.Transactions {
		// Normalize date to midnight UTC to ensure consistent comparison
		normalizedDate := time.Date(
			ofxTxn.Date.Year(),
			ofxTxn.Date.Month(),
			ofxTxn.Date.Day(),
			0, 0, 0, 0,
			time.UTC,
		)

		// Check for duplicate using FitID (Financial Institution Transaction ID)
		// FitID is a unique identifier from the bank, more reliable than date+amount+description
		existing, err := s.transactionRepo.FindByFitID(ctx, accountID, ofxTxn.FitID)
		if err != nil {
			result.Errors = append(result.Errors, fmt.Sprintf("error checking duplicate for transaction: %v", err))
			continue
		}

		if existing != nil {
			result.SkippedDuplicates++
			continue
		}

		// Create new transaction without category (uncategorized)
		transaction := &domain.Transaction{
			ID:          uuid.New().String(),
			Type:        domain.TransactionTypeNormal, // All imported transactions are normal type
			AccountID:   accountID,
			CategoryID:  nil, // Imported transactions start uncategorized
			Amount:      ofxTxn.Amount,
			Description: ofxTxn.Description,
			Date:        normalizedDate,
			FitID:       &ofxTxn.FitID, // Store FitID for duplicate detection
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		}

		if err := s.transactionRepo.Create(ctx, transaction); err != nil {
			result.Errors = append(result.Errors, fmt.Sprintf("failed to create transaction: %v", err))
			continue
		}

		result.ImportedTransactions++
		result.ImportedTransactionIDs = append(result.ImportedTransactionIDs, transaction.ID)
	}

	// Update account balance to match OFX ledger balance (if available)
	// and adjust Ready to Assign by the balance delta
	if parseResult.LedgerBalance != 0 {
		oldBalance := account.Balance
		account.Balance = parseResult.LedgerBalance
		account.UpdatedAt = time.Now()

		if err := s.accountRepo.Update(ctx, account); err != nil {
			// Rollback: delete imported transactions
			for _, txnID := range result.ImportedTransactionIDs {
				s.transactionRepo.Delete(ctx, txnID)
			}
			return nil, fmt.Errorf("failed to update account balance: %w", err)
		}

		// Adjust Ready to Assign by the balance delta only
		// This prevents double-counting when users have manually entered balances
		// Delta = New Balance - Old Balance
		// Example: OFX says $7,895.39, account had $0 -> add $7,895.39 to Ready to Assign
		// Example: OFX says $7,895.39, account had $10,000 -> subtract $2,104.61 from Ready to Assign
		if err := s.budgetStateRepo.AdjustReadyToAssign(ctx, balanceDelta); err != nil {
			// Rollback: delete imported transactions and reverse account balance
			for _, txnID := range result.ImportedTransactionIDs {
				s.transactionRepo.Delete(ctx, txnID)
			}
			account.Balance = oldBalance
			s.accountRepo.Update(ctx, account)
			return nil, fmt.Errorf("failed to adjust ready to assign: %w", err)
		}
	}

	result.NewAccountBalance = account.Balance

	return result, nil
}

// ValidateOFXFile validates that a file is a valid OFX file
func (s *ImportService) ValidateOFXFile(reader io.Reader) error {
	return s.ofxParser.ValidateOFXFile(reader)
}
