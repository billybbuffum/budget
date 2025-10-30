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
		INSERT INTO categories (id, name, description, color, group_id, payment_for_account_id, archived_at, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
	`
	_, err := r.db.ExecContext(ctx, query,
		category.ID, category.Name, category.Description,
		category.Color, category.GroupID, category.PaymentForAccountID, category.ArchivedAt, category.CreatedAt, category.UpdatedAt)
	if err != nil {
		return fmt.Errorf("failed to create category: %w", err)
	}
	return nil
}

func (r *categoryRepository) GetByID(ctx context.Context, id string) (*domain.Category, error) {
	query := `
		SELECT id, name, description, color, group_id, payment_for_account_id, archived_at, created_at, updated_at
		FROM categories
		WHERE id = ? AND archived_at IS NULL
	`
	category := &domain.Category{}
	var groupID, paymentForAccountID sql.NullString
	var archivedAt sql.NullTime
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&category.ID, &category.Name, &category.Description,
		&category.Color, &groupID, &paymentForAccountID, &archivedAt, &category.CreatedAt, &category.UpdatedAt)
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
	if archivedAt.Valid {
		category.ArchivedAt = &archivedAt.Time
	}
	return category, nil
}

func (r *categoryRepository) List(ctx context.Context) ([]*domain.Category, error) {
	query := `
		SELECT id, name, description, color, group_id, payment_for_account_id, archived_at, created_at, updated_at
		FROM categories
		WHERE archived_at IS NULL
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
		var archivedAt sql.NullTime
		if err := rows.Scan(&category.ID, &category.Name,
			&category.Description, &category.Color, &groupID, &paymentForAccountID, &archivedAt, &category.CreatedAt, &category.UpdatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan category: %w", err)
		}
		if groupID.Valid {
			category.GroupID = &groupID.String
		}
		if paymentForAccountID.Valid {
			category.PaymentForAccountID = &paymentForAccountID.String
		}
		if archivedAt.Valid {
			category.ArchivedAt = &archivedAt.Time
		}
		categories = append(categories, category)
	}
	return categories, nil
}

func (r *categoryRepository) ListByGroup(ctx context.Context, groupID string) ([]*domain.Category, error) {
	query := `
		SELECT id, name, description, color, group_id, payment_for_account_id, archived_at, created_at, updated_at
		FROM categories
		WHERE group_id = ? AND archived_at IS NULL
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
		var grpID, paymentForAccountID sql.NullString
		var archivedAt sql.NullTime
		if err := rows.Scan(&category.ID, &category.Name,
			&category.Description, &category.Color, &grpID, &paymentForAccountID, &archivedAt, &category.CreatedAt, &category.UpdatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan category: %w", err)
		}
		if grpID.Valid {
			category.GroupID = &grpID.String
		}
		if paymentForAccountID.Valid {
			category.PaymentForAccountID = &paymentForAccountID.String
		}
		if archivedAt.Valid {
			category.ArchivedAt = &archivedAt.Time
		}
		categories = append(categories, category)
	}
	return categories, nil
}

func (r *categoryRepository) Update(ctx context.Context, category *domain.Category) error {
	query := `
		UPDATE categories
		SET name = ?, description = ?, color = ?, group_id = ?, payment_for_account_id = ?, archived_at = ?, updated_at = ?
		WHERE id = ?
	`
	result, err := r.db.ExecContext(ctx, query,
		category.Name, category.Description,
		category.Color, category.GroupID, category.PaymentForAccountID, category.ArchivedAt, category.UpdatedAt, category.ID)
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

func (r *categoryRepository) GetPaymentCategoryByAccountID(ctx context.Context, accountID string) (*domain.Category, error) {
	query := `
		SELECT id, name, description, color, group_id, payment_for_account_id, archived_at, created_at, updated_at
		FROM categories
		WHERE payment_for_account_id = ? AND archived_at IS NULL
	`
	category := &domain.Category{}
	var groupID, paymentForAccountID sql.NullString
	var archivedAt sql.NullTime
	err := r.db.QueryRowContext(ctx, query, accountID).Scan(
		&category.ID, &category.Name, &category.Description,
		&category.Color, &groupID, &paymentForAccountID, &archivedAt, &category.CreatedAt, &category.UpdatedAt)
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
	if archivedAt.Valid {
		category.ArchivedAt = &archivedAt.Time
	}
	return category, nil
}

func (r *categoryRepository) Delete(ctx context.Context, id string) error {
	// Soft delete: set archived_at to current time
	query := `
		UPDATE categories
		SET archived_at = datetime('now'), updated_at = datetime('now')
		WHERE id = ? AND archived_at IS NULL
	`
	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to archive category: %w", err)
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	if rows == 0 {
		return fmt.Errorf("category not found or already archived")
	}
	return nil
}

// FindByNameIncludingArchived finds a category by name, including archived categories
// Used to detect if a category with the same name existed previously
func (r *categoryRepository) FindByNameIncludingArchived(ctx context.Context, name string) (*domain.Category, error) {
	query := `
		SELECT id, name, description, color, group_id, payment_for_account_id, archived_at, created_at, updated_at
		FROM categories
		WHERE LOWER(name) = LOWER(?)
		ORDER BY archived_at IS NULL DESC, updated_at DESC
		LIMIT 1
	`
	category := &domain.Category{}
	var groupID, paymentForAccountID sql.NullString
	var archivedAt sql.NullTime
	err := r.db.QueryRowContext(ctx, query, name).Scan(
		&category.ID, &category.Name, &category.Description,
		&category.Color, &groupID, &paymentForAccountID, &archivedAt, &category.CreatedAt, &category.UpdatedAt)
	if err == sql.ErrNoRows {
		return nil, nil // Not found is not an error
	}
	if err != nil {
		return nil, fmt.Errorf("failed to find category: %w", err)
	}
	if groupID.Valid {
		category.GroupID = &groupID.String
	}
	if paymentForAccountID.Valid {
		category.PaymentForAccountID = &paymentForAccountID.String
	}
	if archivedAt.Valid {
		category.ArchivedAt = &archivedAt.Time
	}
	return category, nil
}

// RestoreCategory restores an archived category
func (r *categoryRepository) RestoreCategory(ctx context.Context, id string) error {
	query := `
		UPDATE categories
		SET archived_at = NULL, updated_at = datetime('now')
		WHERE id = ? AND archived_at IS NOT NULL
	`
	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to restore category: %w", err)
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	if rows == 0 {
		return fmt.Errorf("archived category not found")
	}
	return nil
}
