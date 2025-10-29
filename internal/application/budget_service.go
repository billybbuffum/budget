package application

import (
	"context"
	"fmt"
	"time"

	"github.com/billybbuffum/budget/internal/domain"
	"github.com/google/uuid"
)

// BudgetService handles budget-related business logic
type BudgetService struct {
	budgetRepo      domain.BudgetRepository
	categoryRepo    domain.CategoryRepository
	transactionRepo domain.TransactionRepository
}

// NewBudgetService creates a new budget service
func NewBudgetService(
	budgetRepo domain.BudgetRepository,
	categoryRepo domain.CategoryRepository,
	transactionRepo domain.TransactionRepository,
) *BudgetService {
	return &BudgetService{
		budgetRepo:      budgetRepo,
		categoryRepo:    categoryRepo,
		transactionRepo: transactionRepo,
	}
}

// CreateBudget creates a new budget
func (s *BudgetService) CreateBudget(ctx context.Context, categoryID string, amount float64, period, notes string) (*domain.Budget, error) {
	// Validate category exists
	if _, err := s.categoryRepo.GetByID(ctx, categoryID); err != nil {
		return nil, fmt.Errorf("category not found: %w", err)
	}

	if amount <= 0 {
		return nil, fmt.Errorf("budget amount must be positive")
	}

	if period == "" {
		return nil, fmt.Errorf("period is required (e.g., '2024-01')")
	}

	budget := &domain.Budget{
		ID:         uuid.New().String(),
		CategoryID: categoryID,
		Amount:     amount,
		Period:     period,
		Notes:      notes,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}

	if err := s.budgetRepo.Create(ctx, budget); err != nil {
		return nil, err
	}

	return budget, nil
}

// GetBudget retrieves a budget by ID
func (s *BudgetService) GetBudget(ctx context.Context, id string) (*domain.Budget, error) {
	return s.budgetRepo.GetByID(ctx, id)
}

// GetBudgetByCategoryAndPeriod retrieves a budget for a specific category and period
func (s *BudgetService) GetBudgetByCategoryAndPeriod(ctx context.Context, categoryID, period string) (*domain.Budget, error) {
	return s.budgetRepo.GetByCategoryAndPeriod(ctx, categoryID, period)
}

// ListBudgets retrieves all budgets
func (s *BudgetService) ListBudgets(ctx context.Context) ([]*domain.Budget, error) {
	return s.budgetRepo.List(ctx)
}

// ListBudgetsByPeriod retrieves budgets for a specific period
func (s *BudgetService) ListBudgetsByPeriod(ctx context.Context, period string) ([]*domain.Budget, error) {
	return s.budgetRepo.ListByPeriod(ctx, period)
}

// GetBudgetSummary calculates budget vs actual spending summary
func (s *BudgetService) GetBudgetSummary(ctx context.Context, period string) ([]*domain.BudgetSummary, error) {
	budgets, err := s.budgetRepo.ListByPeriod(ctx, period)
	if err != nil {
		return nil, err
	}

	// Parse period to get date range (assuming YYYY-MM format)
	startDate, endDate, err := parsePeriod(period)
	if err != nil {
		return nil, fmt.Errorf("invalid period format: %w", err)
	}

	var summaries []*domain.BudgetSummary
	for _, budget := range budgets {
		category, err := s.categoryRepo.GetByID(ctx, budget.CategoryID)
		if err != nil {
			continue
		}

		// Get transactions for this category in the period
		transactions, err := s.transactionRepo.ListByPeriod(ctx, startDate, endDate)
		if err != nil {
			continue
		}

		// Calculate actual spent for this category
		var actualSpent float64
		for _, txn := range transactions {
			if txn.CategoryID == budget.CategoryID {
				actualSpent += txn.Amount
			}
		}

		remaining := budget.Amount - actualSpent
		percentUsed := 0.0
		if budget.Amount > 0 {
			percentUsed = (actualSpent / budget.Amount) * 100
		}

		summary := &domain.BudgetSummary{
			Budget:      budget,
			Category:    category,
			ActualSpent: actualSpent,
			Remaining:   remaining,
			PercentUsed: percentUsed,
		}
		summaries = append(summaries, summary)
	}

	return summaries, nil
}

// UpdateBudget updates an existing budget
func (s *BudgetService) UpdateBudget(ctx context.Context, id, categoryID string, amount float64, period, notes string) (*domain.Budget, error) {
	budget, err := s.budgetRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if categoryID != "" {
		if _, err := s.categoryRepo.GetByID(ctx, categoryID); err != nil {
			return nil, fmt.Errorf("category not found: %w", err)
		}
		budget.CategoryID = categoryID
	}

	if amount > 0 {
		budget.Amount = amount
	}

	if period != "" {
		budget.Period = period
	}

	if notes != "" {
		budget.Notes = notes
	}

	budget.UpdatedAt = time.Now()

	if err := s.budgetRepo.Update(ctx, budget); err != nil {
		return nil, err
	}

	return budget, nil
}

// DeleteBudget deletes a budget
func (s *BudgetService) DeleteBudget(ctx context.Context, id string) error {
	return s.budgetRepo.Delete(ctx, id)
}

// parsePeriod converts a period string (YYYY-MM) to start and end dates in UTC
func parsePeriod(period string) (string, string, error) {
	// Parse the period as YYYY-MM in UTC
	t, err := time.Parse("2006-01", period)
	if err != nil {
		return "", "", err
	}

	// Ensure we're working in UTC
	t = t.UTC()

	// Start of month - go back 1 second to ensure we include 00:00:00
	startDate := t.Add(-time.Second).Format(time.RFC3339)

	// End of month at 23:59:59
	endDate := t.AddDate(0, 1, 0).Add(-time.Second).Format(time.RFC3339)

	return startDate, endDate, nil
}
