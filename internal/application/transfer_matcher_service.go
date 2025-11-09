package application

import (
	"context"
	"fmt"
	"math"
	"strings"
	"time"

	"github.com/billybbuffum/budget/internal/domain"
	"github.com/google/uuid"
)

// TransferMatcherService handles automatic detection of transfer matches
type TransferMatcherService struct {
	transactionRepo domain.TransactionRepository
	accountRepo     domain.AccountRepository
	suggestionRepo  domain.TransferSuggestionRepository
}

// NewTransferMatcherService creates a new transfer matcher service
func NewTransferMatcherService(
	transactionRepo domain.TransactionRepository,
	accountRepo domain.AccountRepository,
	suggestionRepo domain.TransferSuggestionRepository,
) *TransferMatcherService {
	return &TransferMatcherService{
		transactionRepo: transactionRepo,
		accountRepo:     accountRepo,
		suggestionRepo:  suggestionRepo,
	}
}

// FindMatchesForTransaction finds potential transfer matches for a given transaction
func (s *TransferMatcherService) FindMatchesForTransaction(ctx context.Context, txn *domain.Transaction) error {
	// Skip if already a transfer
	if txn.Type == domain.TransactionTypeTransfer {
		return nil
	}

	// Find candidate transactions
	candidates, err := s.findCandidates(ctx, txn)
	if err != nil {
		return fmt.Errorf("failed to find candidates: %w", err)
	}

	// Score each candidate and create suggestions
	for _, candidate := range candidates {
		score := s.calculateMatchScore(txn, candidate)

		// Only create suggestions with score >= 10 (minimum medium confidence)
		if score < 10 {
			continue
		}

		// Check if suggestion already exists
		existing, err := s.suggestionRepo.FindByTransactions(ctx, txn.ID, candidate.ID)
		if err != nil {
			return fmt.Errorf("failed to check existing suggestion: %w", err)
		}
		if existing != nil {
			continue // Skip if suggestion already exists
		}

		// Determine if this is a credit card payment
		isCreditPayment, err := s.isCreditCardPayment(ctx, txn, candidate)
		if err != nil {
			return fmt.Errorf("failed to check credit payment: %w", err)
		}

		// Create suggestion
		suggestion := &domain.TransferMatchSuggestion{
			ID:              uuid.New().String(),
			TransactionAID:  txn.ID,
			TransactionBID:  candidate.ID,
			MatchScore:      score,
			Confidence:      s.classifyConfidence(score),
			Status:          domain.SuggestionStatusPending,
			IsCreditPayment: isCreditPayment,
			CreatedAt:       time.Now().UTC(),
			UpdatedAt:       time.Now().UTC(),
		}

		if err := s.suggestionRepo.Create(ctx, suggestion); err != nil {
			return fmt.Errorf("failed to create suggestion: %w", err)
		}
	}

	return nil
}

// findCandidates finds potential matching transactions
func (s *TransferMatcherService) findCandidates(ctx context.Context, txn *domain.Transaction) ([]*domain.Transaction, error) {
	// Get all transactions in the date window (±3 days)
	startDate := txn.Date.AddDate(0, 0, -3).Format(time.RFC3339)
	endDate := txn.Date.AddDate(0, 0, 3).Format(time.RFC3339)

	allTxns, err := s.transactionRepo.ListByPeriod(ctx, startDate, endDate)
	if err != nil {
		return nil, fmt.Errorf("failed to list transactions by period: %w", err)
	}

	// Get source account to check user ownership and type
	sourceAccount, err := s.accountRepo.GetByID(ctx, txn.AccountID)
	if err != nil {
		return nil, fmt.Errorf("failed to get source account: %w", err)
	}

	var candidates []*domain.Transaction
	for _, candidate := range allTxns {
		// Skip self
		if candidate.ID == txn.ID {
			continue
		}

		// Must be different account
		if candidate.AccountID == txn.AccountID {
			continue
		}

		// Must be opposite amount
		if txn.Amount+candidate.Amount != 0 {
			continue
		}

		// Must be type='normal' (not already a transfer)
		if candidate.Type != domain.TransactionTypeNormal {
			continue
		}

		// Get candidate account to verify same user (for now we assume same database = same user)
		// In a multi-user system, we'd check user_id here
		candidateAccount, err := s.accountRepo.GetByID(ctx, candidate.AccountID)
		if err != nil {
			continue // Skip if account not found
		}

		// For credit card spending (negative on credit card), only match if it's a payment
		// (positive on credit card, negative on checking/savings)
		if candidateAccount.Type == domain.AccountTypeCredit && candidate.Amount < 0 {
			continue // Skip credit card spending
		}
		if sourceAccount.Type == domain.AccountTypeCredit && txn.Amount < 0 {
			continue // Skip credit card spending
		}

		candidates = append(candidates, candidate)
	}

	return candidates, nil
}

// calculateMatchScore calculates the match score for two transactions
func (s *TransferMatcherService) calculateMatchScore(txnA, txnB *domain.Transaction) int {
	score := 0

	// Base score for same day (10 points)
	daysDiff := int(math.Abs(float64(txnB.Date.Sub(txnA.Date).Hours() / 24)))
	if daysDiff == 0 {
		score += 10
	} else {
		// Date proximity penalty: 10 points minus 2 per day
		score += 10 - (daysDiff * 2)
		if score < 0 {
			score = 0
		}
	}

	// Round amount boost (+3 points)
	if s.isRoundAmount(txnA.Amount) {
		score += 3
	}

	// Description similarity boost (+5 points)
	if s.hasTransferKeywords(txnA.Description) || s.hasTransferKeywords(txnB.Description) {
		score += 5
	}

	return score
}

// isRoundAmount checks if an amount is a round number (e.g., $1,000.00)
func (s *TransferMatcherService) isRoundAmount(amountCents int64) bool {
	absAmount := int64(math.Abs(float64(amountCents)))
	// Round if divisible by $100 (10000 cents) and amount >= $100
	return absAmount >= 10000 && absAmount%10000 == 0
}

// hasTransferKeywords checks if description contains transfer-related keywords
func (s *TransferMatcherService) hasTransferKeywords(description string) bool {
	lowerDesc := strings.ToLower(description)
	keywords := []string{"transfer", "xfer", "from", "to", "payment"}
	for _, keyword := range keywords {
		if strings.Contains(lowerDesc, keyword) {
			return true
		}
	}
	return false
}

// isCreditCardPayment checks if the transaction pair represents a credit card payment
func (s *TransferMatcherService) isCreditCardPayment(ctx context.Context, txnA, txnB *domain.Transaction) (bool, error) {
	accountA, err := s.accountRepo.GetByID(ctx, txnA.AccountID)
	if err != nil {
		return false, err
	}
	accountB, err := s.accountRepo.GetByID(ctx, txnB.AccountID)
	if err != nil {
		return false, err
	}

	// One account must be credit card type
	return accountA.Type == domain.AccountTypeCredit || accountB.Type == domain.AccountTypeCredit, nil
}

// classifyConfidence classifies the confidence level based on score
func (s *TransferMatcherService) classifyConfidence(score int) domain.ConfidenceLevel {
	if score >= 15 {
		return domain.ConfidenceLevelHigh
	}
	if score >= 10 {
		return domain.ConfidenceLevelMedium
	}
	return domain.ConfidenceLevelLow
}

// GetMatchScoreForPair calculates and returns the match score for a specific transaction pair
// Used for manual linking validation
func (s *TransferMatcherService) GetMatchScoreForPair(ctx context.Context, txnAID, txnBID string) (int, error) {
	txnA, err := s.transactionRepo.GetByID(ctx, txnAID)
	if err != nil {
		return 0, err
	}
	txnB, err := s.transactionRepo.GetByID(ctx, txnBID)
	if err != nil {
		return 0, err
	}

	return s.calculateMatchScore(txnA, txnB), nil
}
