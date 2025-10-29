package repository

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/billybbuffum/budget/internal/domain"
)

type budgetStateRepository struct {
	db *sql.DB
}

// NewBudgetStateRepository creates a new budget state repository
func NewBudgetStateRepository(db *sql.DB) domain.BudgetStateRepository {
	return &budgetStateRepository{db: db}
}

func (r *budgetStateRepository) Get(ctx context.Context) (*domain.BudgetState, error) {
	query := `
		SELECT id, ready_to_assign, updated_at
		FROM budget_state
		WHERE id = 'singleton'
	`
	state := &domain.BudgetState{}
	err := r.db.QueryRowContext(ctx, query).Scan(
		&state.ID, &state.ReadyToAssign, &state.UpdatedAt)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("budget state not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get budget state: %w", err)
	}
	return state, nil
}

func (r *budgetStateRepository) Update(ctx context.Context, state *domain.BudgetState) error {
	query := `
		UPDATE budget_state
		SET ready_to_assign = ?, updated_at = ?
		WHERE id = 'singleton'
	`
	state.UpdatedAt = time.Now()
	result, err := r.db.ExecContext(ctx, query, state.ReadyToAssign, state.UpdatedAt)
	if err != nil {
		return fmt.Errorf("failed to update budget state: %w", err)
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	if rows == 0 {
		return fmt.Errorf("budget state not found")
	}
	return nil
}

func (r *budgetStateRepository) AdjustReadyToAssign(ctx context.Context, delta int64) error {
	query := `
		UPDATE budget_state
		SET ready_to_assign = ready_to_assign + ?, updated_at = ?
		WHERE id = 'singleton'
	`
	result, err := r.db.ExecContext(ctx, query, delta, time.Now())
	if err != nil {
		return fmt.Errorf("failed to adjust ready to assign: %w", err)
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	if rows == 0 {
		return fmt.Errorf("budget state not found")
	}
	return nil
}
