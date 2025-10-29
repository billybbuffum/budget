package repository

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/billybbuffum/budget/internal/domain"
)

type transactionRepository struct {
	db *sql.DB
}

// NewTransactionRepository creates a new transaction repository
func NewTransactionRepository(db *sql.DB) domain.TransactionRepository {
	return &transactionRepository{db: db}
}

func (r *transactionRepository) Create(ctx context.Context, transaction *domain.Transaction) error {
	query := `
		INSERT INTO transactions (id, user_id, category_id, amount, description, date, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)
	`
	_, err := r.db.ExecContext(ctx, query,
		transaction.ID, transaction.UserID, transaction.CategoryID,
		transaction.Amount, transaction.Description, transaction.Date,
		transaction.CreatedAt, transaction.UpdatedAt)
	if err != nil {
		return fmt.Errorf("failed to create transaction: %w", err)
	}
	return nil
}

func (r *transactionRepository) GetByID(ctx context.Context, id string) (*domain.Transaction, error) {
	query := `
		SELECT id, user_id, category_id, amount, description, date, created_at, updated_at
		FROM transactions
		WHERE id = ?
	`
	transaction := &domain.Transaction{}
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&transaction.ID, &transaction.UserID, &transaction.CategoryID,
		&transaction.Amount, &transaction.Description, &transaction.Date,
		&transaction.CreatedAt, &transaction.UpdatedAt)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("transaction not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get transaction: %w", err)
	}
	return transaction, nil
}

func (r *transactionRepository) List(ctx context.Context) ([]*domain.Transaction, error) {
	query := `
		SELECT id, user_id, category_id, amount, description, date, created_at, updated_at
		FROM transactions
		ORDER BY date DESC
	`
	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to list transactions: %w", err)
	}
	defer rows.Close()

	return r.scanTransactions(rows)
}

func (r *transactionRepository) ListByUser(ctx context.Context, userID string) ([]*domain.Transaction, error) {
	query := `
		SELECT id, user_id, category_id, amount, description, date, created_at, updated_at
		FROM transactions
		WHERE user_id = ?
		ORDER BY date DESC
	`
	rows, err := r.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to list transactions by user: %w", err)
	}
	defer rows.Close()

	return r.scanTransactions(rows)
}

func (r *transactionRepository) ListByCategory(ctx context.Context, categoryID string) ([]*domain.Transaction, error) {
	query := `
		SELECT id, user_id, category_id, amount, description, date, created_at, updated_at
		FROM transactions
		WHERE category_id = ?
		ORDER BY date DESC
	`
	rows, err := r.db.QueryContext(ctx, query, categoryID)
	if err != nil {
		return nil, fmt.Errorf("failed to list transactions by category: %w", err)
	}
	defer rows.Close()

	return r.scanTransactions(rows)
}

func (r *transactionRepository) ListByPeriod(ctx context.Context, startDate, endDate string) ([]*domain.Transaction, error) {
	query := `
		SELECT id, user_id, category_id, amount, description, date, created_at, updated_at
		FROM transactions
		WHERE date >= ? AND date <= ?
		ORDER BY date DESC
	`
	rows, err := r.db.QueryContext(ctx, query, startDate, endDate)
	if err != nil {
		return nil, fmt.Errorf("failed to list transactions by period: %w", err)
	}
	defer rows.Close()

	return r.scanTransactions(rows)
}

func (r *transactionRepository) Update(ctx context.Context, transaction *domain.Transaction) error {
	query := `
		UPDATE transactions
		SET user_id = ?, category_id = ?, amount = ?, description = ?, date = ?, updated_at = ?
		WHERE id = ?
	`
	result, err := r.db.ExecContext(ctx, query,
		transaction.UserID, transaction.CategoryID, transaction.Amount,
		transaction.Description, transaction.Date, transaction.UpdatedAt, transaction.ID)
	if err != nil {
		return fmt.Errorf("failed to update transaction: %w", err)
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	if rows == 0 {
		return fmt.Errorf("transaction not found")
	}
	return nil
}

func (r *transactionRepository) Delete(ctx context.Context, id string) error {
	query := `DELETE FROM transactions WHERE id = ?`
	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete transaction: %w", err)
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	if rows == 0 {
		return fmt.Errorf("transaction not found")
	}
	return nil
}

func (r *transactionRepository) scanTransactions(rows *sql.Rows) ([]*domain.Transaction, error) {
	var transactions []*domain.Transaction
	for rows.Next() {
		transaction := &domain.Transaction{}
		if err := rows.Scan(&transaction.ID, &transaction.UserID, &transaction.CategoryID,
			&transaction.Amount, &transaction.Description, &transaction.Date,
			&transaction.CreatedAt, &transaction.UpdatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan transaction: %w", err)
		}
		transactions = append(transactions, transaction)
	}
	return transactions, nil
}
