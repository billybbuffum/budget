package repository

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/billybbuffum/budget/internal/domain"
)

type categoryGroupRepository struct {
	db *sql.DB
}

// NewCategoryGroupRepository creates a new category group repository
func NewCategoryGroupRepository(db *sql.DB) domain.CategoryGroupRepository {
	return &categoryGroupRepository{db: db}
}

func (r *categoryGroupRepository) Create(ctx context.Context, group *domain.CategoryGroup) error {
	query := `
		INSERT INTO category_groups (id, name, description, display_order, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?)
	`
	_, err := r.db.ExecContext(ctx, query,
		group.ID, group.Name, group.Description,
		group.DisplayOrder, group.CreatedAt, group.UpdatedAt)
	if err != nil {
		return fmt.Errorf("failed to create category group: %w", err)
	}
	return nil
}

func (r *categoryGroupRepository) GetByID(ctx context.Context, id string) (*domain.CategoryGroup, error) {
	query := `
		SELECT id, name, description, display_order, created_at, updated_at
		FROM category_groups
		WHERE id = ?
	`
	group := &domain.CategoryGroup{}
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&group.ID, &group.Name, &group.Description,
		&group.DisplayOrder, &group.CreatedAt, &group.UpdatedAt)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("category group not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get category group: %w", err)
	}
	return group, nil
}

func (r *categoryGroupRepository) List(ctx context.Context) ([]*domain.CategoryGroup, error) {
	query := `
		SELECT id, name, description, display_order, created_at, updated_at
		FROM category_groups
		ORDER BY display_order, name
	`
	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to list category groups: %w", err)
	}
	defer rows.Close()

	var groups []*domain.CategoryGroup
	for rows.Next() {
		group := &domain.CategoryGroup{}
		if err := rows.Scan(&group.ID, &group.Name,
			&group.Description, &group.DisplayOrder, &group.CreatedAt, &group.UpdatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan category group: %w", err)
		}
		groups = append(groups, group)
	}
	return groups, nil
}

func (r *categoryGroupRepository) Update(ctx context.Context, group *domain.CategoryGroup) error {
	query := `
		UPDATE category_groups
		SET name = ?, description = ?, display_order = ?, updated_at = ?
		WHERE id = ?
	`
	result, err := r.db.ExecContext(ctx, query,
		group.Name, group.Description,
		group.DisplayOrder, group.UpdatedAt, group.ID)
	if err != nil {
		return fmt.Errorf("failed to update category group: %w", err)
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	if rows == 0 {
		return fmt.Errorf("category group not found")
	}
	return nil
}

func (r *categoryGroupRepository) Delete(ctx context.Context, id string) error {
	query := `DELETE FROM category_groups WHERE id = ?`
	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete category group: %w", err)
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	if rows == 0 {
		return fmt.Errorf("category group not found")
	}
	return nil
}
