package repository

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/billybbuffum/budget/internal/domain"
)

type categoryRepository struct {
	db *sql.DB
}

// NewCategoryRepository creates a new category repository
func NewCategoryRepository(db *sql.DB) domain.CategoryRepository {
	return &categoryRepository{db: db}
}

func (r *categoryRepository) Create(ctx context.Context, category *domain.Category) error {
	query := `
		INSERT INTO categories (id, name, description, color, group_id, payment_for_account_id, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)
	`
	_, err := r.db.ExecContext(ctx, query,
		category.ID, category.Name, category.Description,
		category.Color, category.GroupID, category.PaymentForAccountID, category.CreatedAt, category.UpdatedAt)
	if err != nil {
		return fmt.Errorf("failed to create category: %w", err)
	}
	return nil
}

func (r *categoryRepository) GetByID(ctx context.Context, id string) (*domain.Category, error) {
	query := `
		SELECT id, name, description, color, group_id, payment_for_account_id, created_at, updated_at
		FROM categories
		WHERE id = ?
	`
	category := &domain.Category{}
	var groupID, paymentForAccountID sql.NullString
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&category.ID, &category.Name, &category.Description,
		&category.Color, &groupID, &paymentForAccountID, &category.CreatedAt, &category.UpdatedAt)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("category not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get category: %w", err)
	}
	if groupID.Valid {
		category.GroupID = &groupID.String
	}
	if paymentForAccountID.Valid {
		category.PaymentForAccountID = &paymentForAccountID.String
	}
	return category, nil
}

func (r *categoryRepository) GetPaymentCategoryByAccountID(ctx context.Context, accountID string) (*domain.Category, error) {
	query := `
		SELECT id, name, description, color, group_id, payment_for_account_id, created_at, updated_at
		FROM categories
		WHERE payment_for_account_id = ?
	`
	category := &domain.Category{}
	var groupID, paymentForAccountID sql.NullString
	err := r.db.QueryRowContext(ctx, query, accountID).Scan(
		&category.ID, &category.Name, &category.Description,
		&category.Color, &groupID, &paymentForAccountID, &category.CreatedAt, &category.UpdatedAt)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("payment category not found for account")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get payment category: %w", err)
	}
	if groupID.Valid {
		category.GroupID = &groupID.String
	}
	if paymentForAccountID.Valid {
		category.PaymentForAccountID = &paymentForAccountID.String
	}
	return category, nil
}

func (r *categoryRepository) List(ctx context.Context) ([]*domain.Category, error) {
	query := `
		SELECT id, name, description, color, group_id, payment_for_account_id, created_at, updated_at
		FROM categories
		ORDER BY name
	`
	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to list categories: %w", err)
	}
	defer rows.Close()

	var categories []*domain.Category
	for rows.Next() {
		category := &domain.Category{}
		var groupID, paymentForAccountID sql.NullString
		if err := rows.Scan(&category.ID, &category.Name,
			&category.Description, &category.Color, &groupID, &paymentForAccountID, &category.CreatedAt, &category.UpdatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan category: %w", err)
		}
		if groupID.Valid {
			category.GroupID = &groupID.String
		}
		if paymentForAccountID.Valid {
			category.PaymentForAccountID = &paymentForAccountID.String
		}
		categories = append(categories, category)
	}
	return categories, nil
}

func (r *categoryRepository) ListByGroup(ctx context.Context, groupID string) ([]*domain.Category, error) {
	query := `
		SELECT id, name, description, color, group_id, payment_for_account_id, created_at, updated_at
		FROM categories
		WHERE group_id = ?
		ORDER BY name
	`
	rows, err := r.db.QueryContext(ctx, query, groupID)
	if err != nil {
		return nil, fmt.Errorf("failed to list categories by group: %w", err)
	}
	defer rows.Close()

	var categories []*domain.Category
	for rows.Next() {
		category := &domain.Category{}
		var groupID, paymentForAccountID sql.NullString
		if err := rows.Scan(&category.ID, &category.Name,
			&category.Description, &category.Color, &groupID, &paymentForAccountID, &category.CreatedAt, &category.UpdatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan category: %w", err)
		}
		if groupID.Valid {
			category.GroupID = &groupID.String
		}
		if paymentForAccountID.Valid {
			category.PaymentForAccountID = &paymentForAccountID.String
		}
		categories = append(categories, category)
	}
	return categories, nil
}

func (r *categoryRepository) Update(ctx context.Context, category *domain.Category) error {
	query := `
		UPDATE categories
		SET name = ?, description = ?, color = ?, group_id = ?, updated_at = ?
		WHERE id = ?
	`
	result, err := r.db.ExecContext(ctx, query,
		category.Name, category.Description,
		category.Color, category.GroupID, category.UpdatedAt, category.ID)
	if err != nil {
		return fmt.Errorf("failed to update category: %w", err)
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	if rows == 0 {
		return fmt.Errorf("category not found")
	}
	return nil
}

func (r *categoryRepository) Delete(ctx context.Context, id string) error {
	query := `DELETE FROM categories WHERE id = ?`
	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete category: %w", err)
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	if rows == 0 {
		return fmt.Errorf("category not found")
	}
	return nil
}
