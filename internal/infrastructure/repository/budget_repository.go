package repository

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/billybbuffum/budget/internal/domain"
)

type budgetRepository struct {
	db *sql.DB
}

// NewBudgetRepository creates a new budget repository
func NewBudgetRepository(db *sql.DB) domain.BudgetRepository {
	return &budgetRepository{db: db}
}

func (r *budgetRepository) Create(ctx context.Context, budget *domain.Budget) error {
	query := `
		INSERT INTO budgets (id, category_id, amount, period, notes, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?)
	`
	_, err := r.db.ExecContext(ctx, query,
		budget.ID, budget.CategoryID, budget.Amount, budget.Period,
		budget.Notes, budget.CreatedAt, budget.UpdatedAt)
	if err != nil {
		return fmt.Errorf("failed to create budget: %w", err)
	}
	return nil
}

func (r *budgetRepository) GetByID(ctx context.Context, id string) (*domain.Budget, error) {
	query := `
		SELECT id, category_id, amount, period, notes, created_at, updated_at
		FROM budgets
		WHERE id = ?
	`
	budget := &domain.Budget{}
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&budget.ID, &budget.CategoryID, &budget.Amount, &budget.Period,
		&budget.Notes, &budget.CreatedAt, &budget.UpdatedAt)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("budget not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get budget: %w", err)
	}
	return budget, nil
}

func (r *budgetRepository) GetByCategoryAndPeriod(ctx context.Context, categoryID, period string) (*domain.Budget, error) {
	query := `
		SELECT id, category_id, amount, period, notes, created_at, updated_at
		FROM budgets
		WHERE category_id = ? AND period = ?
	`
	budget := &domain.Budget{}
	err := r.db.QueryRowContext(ctx, query, categoryID, period).Scan(
		&budget.ID, &budget.CategoryID, &budget.Amount, &budget.Period,
		&budget.Notes, &budget.CreatedAt, &budget.UpdatedAt)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("budget not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get budget: %w", err)
	}
	return budget, nil
}

func (r *budgetRepository) ListByPeriod(ctx context.Context, period string) ([]*domain.Budget, error) {
	query := `
		SELECT id, category_id, amount, period, notes, created_at, updated_at
		FROM budgets
		WHERE period = ?
		ORDER BY created_at DESC
	`
	rows, err := r.db.QueryContext(ctx, query, period)
	if err != nil {
		return nil, fmt.Errorf("failed to list budgets by period: %w", err)
	}
	defer rows.Close()

	return r.scanBudgets(rows)
}

func (r *budgetRepository) List(ctx context.Context) ([]*domain.Budget, error) {
	query := `
		SELECT id, category_id, amount, period, notes, created_at, updated_at
		FROM budgets
		ORDER BY period DESC, created_at DESC
	`
	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to list budgets: %w", err)
	}
	defer rows.Close()

	return r.scanBudgets(rows)
}

func (r *budgetRepository) Update(ctx context.Context, budget *domain.Budget) error {
	query := `
		UPDATE budgets
		SET category_id = ?, amount = ?, period = ?, notes = ?, updated_at = ?
		WHERE id = ?
	`
	result, err := r.db.ExecContext(ctx, query,
		budget.CategoryID, budget.Amount, budget.Period,
		budget.Notes, budget.UpdatedAt, budget.ID)
	if err != nil {
		return fmt.Errorf("failed to update budget: %w", err)
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	if rows == 0 {
		return fmt.Errorf("budget not found")
	}
	return nil
}

func (r *budgetRepository) Delete(ctx context.Context, id string) error {
	query := `DELETE FROM budgets WHERE id = ?`
	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete budget: %w", err)
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	if rows == 0 {
		return fmt.Errorf("budget not found")
	}
	return nil
}

func (r *budgetRepository) scanBudgets(rows *sql.Rows) ([]*domain.Budget, error) {
	var budgets []*domain.Budget
	for rows.Next() {
		budget := &domain.Budget{}
		if err := rows.Scan(&budget.ID, &budget.CategoryID, &budget.Amount,
			&budget.Period, &budget.Notes, &budget.CreatedAt, &budget.UpdatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan budget: %w", err)
		}
		budgets = append(budgets, budget)
	}
	return budgets, nil
}
