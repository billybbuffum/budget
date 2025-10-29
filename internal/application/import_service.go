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
	ofxParser       *ofx.Parser
}

// NewImportService creates a new import service
func NewImportService(
	transactionRepo domain.TransactionRepository,
	accountRepo domain.AccountRepository,
	ofxParser *ofx.Parser,
) *ImportService {
	return &ImportService{
		transactionRepo: transactionRepo,
		accountRepo:     accountRepo,
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

	// Parse OFX file
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

	// Track balance changes
	balanceChange := int64(0)

	// Process each transaction
	for _, ofxTxn := range parseResult.Transactions {
		// Check for duplicate
		dateStr := ofxTxn.Date.Format(time.RFC3339)
		existing, err := s.transactionRepo.FindDuplicate(ctx, accountID, dateStr, ofxTxn.Amount, ofxTxn.Description)
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
			AccountID:   accountID,
			CategoryID:  nil, // Imported transactions start uncategorized
			Amount:      ofxTxn.Amount,
			Description: ofxTxn.Description,
			Date:        ofxTxn.Date,
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		}

		if err := s.transactionRepo.Create(ctx, transaction); err != nil {
			result.Errors = append(result.Errors, fmt.Sprintf("failed to create transaction: %v", err))
			continue
		}

		result.ImportedTransactions++
		result.ImportedTransactionIDs = append(result.ImportedTransactionIDs, transaction.ID)
		balanceChange += ofxTxn.Amount
	}

	// Update account balance if any transactions were imported
	if result.ImportedTransactions > 0 {
		account.Balance += balanceChange
		account.UpdatedAt = time.Now()

		if err := s.accountRepo.Update(ctx, account); err != nil {
			// Rollback: delete imported transactions
			for _, txnID := range result.ImportedTransactionIDs {
				s.transactionRepo.Delete(ctx, txnID)
			}
			return nil, fmt.Errorf("failed to update account balance: %w", err)
		}
	}

	result.NewAccountBalance = account.Balance

	return result, nil
}

// ValidateOFXFile validates that a file is a valid OFX file
func (s *ImportService) ValidateOFXFile(reader io.Reader) error {
	return s.ofxParser.ValidateOFXFile(reader)
}
