package repository

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/billybbuffum/budget/internal/domain"
)

type allocationRepository struct {
	db *sql.DB
}

// NewAllocationRepository creates a new allocation repository
func NewAllocationRepository(db *sql.DB) domain.AllocationRepository {
	return &allocationRepository{db: db}
}

func (r *allocationRepository) Create(ctx context.Context, allocation *domain.Allocation) error {
	query := `
		INSERT INTO allocations (id, category_id, amount, period, notes, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?)
	`
	_, err := r.db.ExecContext(ctx, query,
		allocation.ID, allocation.CategoryID, allocation.Amount, allocation.Period,
		allocation.Notes, allocation.CreatedAt, allocation.UpdatedAt)
	if err != nil {
		return fmt.Errorf("failed to create allocation: %w", err)
	}
	return nil
}

func (r *allocationRepository) GetByID(ctx context.Context, id string) (*domain.Allocation, error) {
	query := `
		SELECT id, category_id, amount, period, notes, created_at, updated_at
		FROM allocations
		WHERE id = ?
	`
	allocation := &domain.Allocation{}
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&allocation.ID, &allocation.CategoryID, &allocation.Amount, &allocation.Period,
		&allocation.Notes, &allocation.CreatedAt, &allocation.UpdatedAt)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("allocation not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get allocation: %w", err)
	}
	return allocation, nil
}

func (r *allocationRepository) GetByCategoryAndPeriod(ctx context.Context, categoryID, period string) (*domain.Allocation, error) {
	query := `
		SELECT id, category_id, amount, period, notes, created_at, updated_at
		FROM allocations
		WHERE category_id = ? AND period = ?
	`
	allocation := &domain.Allocation{}
	err := r.db.QueryRowContext(ctx, query, categoryID, period).Scan(
		&allocation.ID, &allocation.CategoryID, &allocation.Amount, &allocation.Period,
		&allocation.Notes, &allocation.CreatedAt, &allocation.UpdatedAt)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("allocation not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get allocation: %w", err)
	}
	return allocation, nil
}

func (r *allocationRepository) ListByPeriod(ctx context.Context, period string) ([]*domain.Allocation, error) {
	query := `
		SELECT id, category_id, amount, period, notes, created_at, updated_at
		FROM allocations
		WHERE period = ?
		ORDER BY created_at DESC
	`
	rows, err := r.db.QueryContext(ctx, query, period)
	if err != nil {
		return nil, fmt.Errorf("failed to list allocations by period: %w", err)
	}
	defer rows.Close()

	return r.scanAllocations(rows)
}

func (r *allocationRepository) List(ctx context.Context) ([]*domain.Allocation, error) {
	query := `
		SELECT id, category_id, amount, period, notes, created_at, updated_at
		FROM allocations
		ORDER BY period DESC, created_at DESC
	`
	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to list allocations: %w", err)
	}
	defer rows.Close()

	return r.scanAllocations(rows)
}

func (r *allocationRepository) Update(ctx context.Context, allocation *domain.Allocation) error {
	query := `
		UPDATE allocations
		SET category_id = ?, amount = ?, period = ?, notes = ?, updated_at = ?
		WHERE id = ?
	`
	result, err := r.db.ExecContext(ctx, query,
		allocation.CategoryID, allocation.Amount, allocation.Period,
		allocation.Notes, allocation.UpdatedAt, allocation.ID)
	if err != nil {
		return fmt.Errorf("failed to update allocation: %w", err)
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	if rows == 0 {
		return fmt.Errorf("allocation not found")
	}
	return nil
}

func (r *allocationRepository) Delete(ctx context.Context, id string) error {
	query := `DELETE FROM allocations WHERE id = ?`
	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete allocation: %w", err)
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	if rows == 0 {
		return fmt.Errorf("allocation not found")
	}
	return nil
}

func (r *allocationRepository) scanAllocations(rows *sql.Rows) ([]*domain.Allocation, error) {
	var allocations []*domain.Allocation
	for rows.Next() {
		allocation := &domain.Allocation{}
		if err := rows.Scan(&allocation.ID, &allocation.CategoryID, &allocation.Amount,
			&allocation.Period, &allocation.Notes, &allocation.CreatedAt, &allocation.UpdatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan allocation: %w", err)
		}
		allocations = append(allocations, allocation)
	}
	return allocations, nil
}
