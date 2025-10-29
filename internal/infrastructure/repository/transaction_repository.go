package repository

import (
	"context"
	"database/sql"
	"fmt"
	"time"

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
		INSERT INTO transactions (id, type, account_id, transfer_to_account_id, category_id, amount, description, date, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`
	_, err := r.db.ExecContext(ctx, query,
		transaction.ID, transaction.Type, transaction.AccountID, transaction.TransferToAccountID, transaction.CategoryID,
		transaction.Amount, transaction.Description, transaction.Date,
		transaction.CreatedAt, transaction.UpdatedAt)
	if err != nil {
		return fmt.Errorf("failed to create transaction: %w", err)
	}
	return nil
}

func (r *transactionRepository) GetByID(ctx context.Context, id string) (*domain.Transaction, error) {
	query := `
		SELECT id, type, account_id, transfer_to_account_id, category_id, amount, description, date, created_at, updated_at
		FROM transactions
		WHERE id = ?
	`
	transaction := &domain.Transaction{}
	var categoryID, transferToAccountID sql.NullString
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&transaction.ID, &transaction.Type, &transaction.AccountID, &transferToAccountID, &categoryID,
		&transaction.Amount, &transaction.Description, &transaction.Date,
		&transaction.CreatedAt, &transaction.UpdatedAt)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("transaction not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get transaction: %w", err)
	}
	if categoryID.Valid {
		transaction.CategoryID = &categoryID.String
	}
	if transferToAccountID.Valid {
		transaction.TransferToAccountID = &transferToAccountID.String
	}
	return transaction, nil
}

func (r *transactionRepository) List(ctx context.Context) ([]*domain.Transaction, error) {
	query := `
		SELECT id, type, account_id, transfer_to_account_id, category_id, amount, description, date, created_at, updated_at
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

func (r *transactionRepository) ListByAccount(ctx context.Context, accountID string) ([]*domain.Transaction, error) {
	query := `
		SELECT id, type, account_id, transfer_to_account_id, category_id, amount, description, date, created_at, updated_at
		FROM transactions
		WHERE account_id = ?
		ORDER BY date DESC
	`
	rows, err := r.db.QueryContext(ctx, query, accountID)
	if err != nil {
		return nil, fmt.Errorf("failed to list transactions by account: %w", err)
	}
	defer rows.Close()

	return r.scanTransactions(rows)
}

func (r *transactionRepository) ListByCategory(ctx context.Context, categoryID string) ([]*domain.Transaction, error) {
	query := `
		SELECT id, type, account_id, transfer_to_account_id, category_id, amount, description, date, created_at, updated_at
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
		SELECT id, type, account_id, transfer_to_account_id, category_id, amount, description, date, created_at, updated_at
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

func (r *transactionRepository) GetCategoryActivity(ctx context.Context, categoryID, period string) (int64, error) {
	// Parse period to get date range (YYYY-MM format)
	t, err := time.Parse("2006-01", period)
	if err != nil {
		return 0, fmt.Errorf("invalid period format: %w", err)
	}

	t = t.UTC()
	startDate := t.Format(time.RFC3339)
	endDate := t.AddDate(0, 1, 0).Add(-time.Second).Format(time.RFC3339)

	query := `
		SELECT COALESCE(SUM(amount), 0)
		FROM transactions
		WHERE category_id = ? AND date >= ? AND date <= ?
	`
	var activity int64
	err = r.db.QueryRowContext(ctx, query, categoryID, startDate, endDate).Scan(&activity)
	if err != nil {
		return 0, fmt.Errorf("failed to get category activity: %w", err)
	}
	return activity, nil
}

func (r *transactionRepository) Update(ctx context.Context, transaction *domain.Transaction) error {
	query := `
		UPDATE transactions
		SET type = ?, account_id = ?, transfer_to_account_id = ?, category_id = ?, amount = ?, description = ?, date = ?, updated_at = ?
		WHERE id = ?
	`
	result, err := r.db.ExecContext(ctx, query,
		transaction.Type, transaction.AccountID, transaction.TransferToAccountID, transaction.CategoryID, transaction.Amount,
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
		var categoryID, transferToAccountID sql.NullString
		if err := rows.Scan(&transaction.ID, &transaction.Type, &transaction.AccountID, &transferToAccountID, &categoryID,
			&transaction.Amount, &transaction.Description, &transaction.Date,
			&transaction.CreatedAt, &transaction.UpdatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan transaction: %w", err)
		}
		if categoryID.Valid {
			transaction.CategoryID = &categoryID.String
		}
		if transferToAccountID.Valid {
			transaction.TransferToAccountID = &transferToAccountID.String
		}
		transactions = append(transactions, transaction)
	}
	return transactions, nil
}
