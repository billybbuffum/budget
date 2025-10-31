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

// TransferMatcherService handles finding and scoring potential transfer matches
type TransferMatcherService struct {
	transactionRepo domain.TransactionRepository
	accountRepo     domain.AccountRepository
	suggestionRepo  domain.TransferSuggestionRepository
	config          TransferMatchConfig
}

// TransferMatchConfig holds configuration for the matching algorithm
type TransferMatchConfig struct {
	MaxDateDiffDays    int // Maximum days between matching transactions (default: 3)
	MinConfidenceScore int // Minimum score to create a suggestion (default: 10)
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
		config: TransferMatchConfig{
			MaxDateDiffDays:    3,
			MinConfidenceScore: 10,
		},
	}
}

// FindMatchesForTransactions finds potential transfer matches for the given transactions
func (s *TransferMatcherService) FindMatchesForTransactions(ctx context.Context, transactionIDs []string) error {
	for _, txnID := range transactionIDs {
		txn, err := s.transactionRepo.GetByID(ctx, txnID)
		if err != nil {
			return fmt.Errorf("failed to get transaction %s: %w", txnID, err)
		}

		// Skip if already a transfer or already linked
		if txn.Type != domain.TransactionTypeNormal || txn.TransferToAccountID != nil {
			continue
		}

		// Find candidates
		candidates, err := s.findCandidates(ctx, txn)
		if err != nil {
			return fmt.Errorf("failed to find candidates for transaction %s: %w", txnID, err)
		}

		// Score each candidate and create suggestions
		for _, candidate := range candidates {
			score := s.calculateMatchScore(txn, candidate)
			if score < s.config.MinConfidenceScore {
				continue
			}

			confidence := s.classifyConfidence(score)
			isCreditPayment := s.isCreditCardPayment(txn, candidate)

			suggestion := &domain.TransferSuggestion{
				ID:              uuid.New().String(),
				TransactionAID:  txn.ID,
				TransactionBID:  candidate.ID,
				Confidence:      confidence,
				Score:           score,
				IsCreditPayment: isCreditPayment,
				Status:          domain.StatusPending,
				CreatedAt:       time.Now(),
			}

			// Create suggestion (ignore duplicates - unique constraint will prevent them)
			if err := s.suggestionRepo.Create(ctx, suggestion); err != nil {
				// Log but don't fail - duplicate suggestions are OK (unique constraint)
				if !strings.Contains(err.Error(), "UNIQUE constraint failed") {
					return fmt.Errorf("failed to create suggestion: %w", err)
				}
			}
		}
	}

	return nil
}

// findCandidates finds potential matching transactions for the given transaction
func (s *TransferMatcherService) findCandidates(ctx context.Context, txn *domain.Transaction) ([]*domain.Transaction, error) {
	// Get all transactions
	allTransactions, err := s.transactionRepo.List(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list transactions: %w", err)
	}

	// Get the account for this transaction to check user ownership
	txnAccount, err := s.accountRepo.GetByID(ctx, txn.AccountID)
	if err != nil {
		return nil, fmt.Errorf("failed to get account: %w", err)
	}

	// Calculate date window
	dateMin := txn.Date.AddDate(0, 0, -s.config.MaxDateDiffDays)
	dateMax := txn.Date.AddDate(0, 0, s.config.MaxDateDiffDays)

	var candidates []*domain.Transaction
	for _, candidate := range allTransactions {
		// Skip self
		if candidate.ID == txn.ID {
			continue
		}

		// Must be in different account
		if candidate.AccountID == txn.AccountID {
			continue
		}

		// Must be normal type (not already a transfer)
		if candidate.Type != domain.TransactionTypeNormal {
			continue
		}

		// Must not be already linked
		if candidate.TransferToAccountID != nil {
			continue
		}

		// Must have opposite amount (one negative, one positive, same magnitude)
		if candidate.Amount != -txn.Amount {
			continue
		}

		// Must be within date window
		if candidate.Date.Before(dateMin) || candidate.Date.After(dateMax) {
			continue
		}

		// Must belong to same user (check account ownership)
		candidateAccount, err := s.accountRepo.GetByID(ctx, candidate.AccountID)
		if err != nil {
			continue // Skip if can't verify ownership
		}

		// For now, all accounts belong to same user in this system
		// If multi-user support is added, add user ID check here
		_ = txnAccount
		_ = candidateAccount

		candidates = append(candidates, candidate)
	}

	return candidates, nil
}

// calculateMatchScore scores a potential match between two transactions
func (s *TransferMatcherService) calculateMatchScore(txnA, txnB *domain.Transaction) int {
	score := 0

	// Date proximity scoring
	daysDiff := math.Abs(txnB.Date.Sub(txnA.Date).Hours() / 24)
	if daysDiff == 0 {
		score += 10 // Same day
	} else if daysDiff <= 1 {
		score += 5 // 1 day apart
	} else if daysDiff <= 3 {
		score += 2 // 2-3 days apart
	}

	// Description similarity
	if s.containsTransferKeywords(txnA.Description) && s.containsTransferKeywords(txnB.Description) {
		score += 5
	}

	// Round amount boost
	if s.isRoundAmount(txnA.Amount) {
		score += 3
	}

	// Credit card payment boost
	if s.isCreditCardPayment(txnA, txnB) {
		score += 5
		if s.containsPaymentKeywords(txnA.Description) || s.containsPaymentKeywords(txnB.Description) {
			score += 3
		}
	}

	return score
}

// classifyConfidence classifies the confidence level based on score
func (s *TransferMatcherService) classifyConfidence(score int) string {
	if score >= 15 {
		return domain.ConfidenceHigh
	} else if score >= 10 {
		return domain.ConfidenceMedium
	} else {
		return domain.ConfidenceLow
	}
}

// isCreditCardPayment checks if this is a credit card payment transfer
func (s *TransferMatcherService) isCreditCardPayment(txnA, txnB *domain.Transaction) bool {
	ctx := context.Background()

	accountA, err := s.accountRepo.GetByID(ctx, txnA.AccountID)
	if err != nil {
		return false
	}

	accountB, err := s.accountRepo.GetByID(ctx, txnB.AccountID)
	if err != nil {
		return false
	}

	return accountA.Type == domain.AccountTypeCredit || accountB.Type == domain.AccountTypeCredit
}

// containsTransferKeywords checks if description contains transfer-related keywords
func (s *TransferMatcherService) containsTransferKeywords(description string) bool {
	keywords := []string{"transfer", "xfer", "from", "to"}
	descLower := strings.ToLower(description)
	for _, kw := range keywords {
		if strings.Contains(descLower, kw) {
			return true
		}
	}
	return false
}

// containsPaymentKeywords checks if description contains payment-related keywords
func (s *TransferMatcherService) containsPaymentKeywords(description string) bool {
	keywords := []string{"payment", "autopay", "pay"}
	descLower := strings.ToLower(description)
	for _, kw := range keywords {
		if strings.Contains(descLower, kw) {
			return true
		}
	}
	return false
}

// isRoundAmount checks if the amount is a round number (ends in .00)
func (s *TransferMatcherService) isRoundAmount(amount int64) bool {
	return amount%100 == 0
}
