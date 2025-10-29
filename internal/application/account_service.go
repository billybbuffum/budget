package application

import (
	"context"
	"fmt"
	"time"

	"github.com/billybbuffum/budget/internal/domain"
	"github.com/google/uuid"
)

// AccountService handles account-related business logic
type AccountService struct {
	accountRepo     domain.AccountRepository
	budgetStateRepo domain.BudgetStateRepository
}

// NewAccountService creates a new account service
func NewAccountService(accountRepo domain.AccountRepository, budgetStateRepo domain.BudgetStateRepository) *AccountService {
	return &AccountService{
		accountRepo:     accountRepo,
		budgetStateRepo: budgetStateRepo,
	}
}

// CreateAccount creates a new account
func (s *AccountService) CreateAccount(ctx context.Context, name string, balance int64, accountType domain.AccountType) (*domain.Account, error) {
	if name == "" {
		return nil, fmt.Errorf("account name is required")
	}

	if accountType != domain.AccountTypeChecking &&
	   accountType != domain.AccountTypeSavings &&
	   accountType != domain.AccountTypeCash {
		return nil, fmt.Errorf("invalid account type")
	}

	account := &domain.Account{
		ID:        uuid.New().String(),
		Name:      name,
		Balance:   balance,
		Type:      accountType,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	if err := s.accountRepo.Create(ctx, account); err != nil {
		return nil, err
	}

	// If account has a starting balance, add it to Ready to Assign
	// This represents money in accounts that hasn't been assigned to categories yet
	if balance != 0 {
		if err := s.budgetStateRepo.AdjustReadyToAssign(ctx, balance); err != nil {
			// Rollback account creation if Ready to Assign update fails
			s.accountRepo.Delete(ctx, account.ID)
			return nil, fmt.Errorf("failed to adjust ready to assign: %w", err)
		}
	}

	return account, nil
}

// GetAccount retrieves an account by ID
func (s *AccountService) GetAccount(ctx context.Context, id string) (*domain.Account, error) {
	return s.accountRepo.GetByID(ctx, id)
}

// ListAccounts retrieves all accounts
func (s *AccountService) ListAccounts(ctx context.Context) ([]*domain.Account, error) {
	return s.accountRepo.List(ctx)
}

// UpdateAccount updates an existing account
func (s *AccountService) UpdateAccount(ctx context.Context, id, name string, balance int64, accountType domain.AccountType) (*domain.Account, error) {
	account, err := s.accountRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if name != "" {
		account.Name = name
	}

	// Calculate the delta if balance is changing
	oldBalance := account.Balance
	balanceDelta := balance - oldBalance

	// Allow updating balance to any value (including negative for credit cards potentially)
	account.Balance = balance

	if accountType != "" {
		if accountType != domain.AccountTypeChecking &&
		   accountType != domain.AccountTypeSavings &&
		   accountType != domain.AccountTypeCash {
			return nil, fmt.Errorf("invalid account type")
		}
		account.Type = accountType
	}

	account.UpdatedAt = time.Now()

	if err := s.accountRepo.Update(ctx, account); err != nil {
		return nil, err
	}

	// Adjust Ready to Assign by the balance delta
	// If balance increased, increase Ready to Assign; if decreased, decrease it
	if balanceDelta != 0 {
		if err := s.budgetStateRepo.AdjustReadyToAssign(ctx, balanceDelta); err != nil {
			return nil, fmt.Errorf("failed to adjust ready to assign: %w", err)
		}
	}

	return account, nil
}

// DeleteAccount deletes an account and adjusts Ready to Assign
func (s *AccountService) DeleteAccount(ctx context.Context, id string) error {
	// Get the account first to know its balance
	account, err := s.accountRepo.GetByID(ctx, id)
	if err != nil {
		return err
	}

	// Delete the account
	if err := s.accountRepo.Delete(ctx, id); err != nil {
		return err
	}

	// Subtract the account balance from Ready to Assign
	// This represents money that no longer exists in any account
	if account.Balance != 0 {
		if err := s.budgetStateRepo.AdjustReadyToAssign(ctx, -account.Balance); err != nil {
			return fmt.Errorf("failed to adjust ready to assign: %w", err)
		}
	}

	return nil
}

// GetTotalBalance returns the sum of all account balances
func (s *AccountService) GetTotalBalance(ctx context.Context) (int64, error) {
	return s.accountRepo.GetTotalBalance(ctx)
}
