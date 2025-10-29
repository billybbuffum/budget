package repository

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/billybbuffum/budget/internal/domain"
)

type accountRepository struct {
	db *sql.DB
}

// NewAccountRepository creates a new account repository
func NewAccountRepository(db *sql.DB) domain.AccountRepository {
	return &accountRepository{db: db}
}

func (r *accountRepository) Create(ctx context.Context, account *domain.Account) error {
	query := `
		INSERT INTO accounts (id, name, balance, type, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?)
	`
	_, err := r.db.ExecContext(ctx, query,
		account.ID, account.Name, account.Balance, account.Type,
		account.CreatedAt, account.UpdatedAt)
	if err != nil {
		return fmt.Errorf("failed to create account: %w", err)
	}
	return nil
}

func (r *accountRepository) GetByID(ctx context.Context, id string) (*domain.Account, error) {
	query := `
		SELECT id, name, balance, type, created_at, updated_at
		FROM accounts
		WHERE id = ?
	`
	account := &domain.Account{}
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&account.ID, &account.Name, &account.Balance, &account.Type,
		&account.CreatedAt, &account.UpdatedAt)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("account not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get account: %w", err)
	}
	return account, nil
}

func (r *accountRepository) List(ctx context.Context) ([]*domain.Account, error) {
	query := `
		SELECT id, name, balance, type, created_at, updated_at
		FROM accounts
		ORDER BY created_at DESC
	`
	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to list accounts: %w", err)
	}
	defer rows.Close()

	var accounts []*domain.Account
	for rows.Next() {
		account := &domain.Account{}
		if err := rows.Scan(&account.ID, &account.Name, &account.Balance, &account.Type,
			&account.CreatedAt, &account.UpdatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan account: %w", err)
		}
		accounts = append(accounts, account)
	}
	return accounts, nil
}

func (r *accountRepository) Update(ctx context.Context, account *domain.Account) error {
	query := `
		UPDATE accounts
		SET name = ?, balance = ?, type = ?, updated_at = ?
		WHERE id = ?
	`
	result, err := r.db.ExecContext(ctx, query,
		account.Name, account.Balance, account.Type, account.UpdatedAt, account.ID)
	if err != nil {
		return fmt.Errorf("failed to update account: %w", err)
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	if rows == 0 {
		return fmt.Errorf("account not found")
	}
	return nil
}

func (r *accountRepository) Delete(ctx context.Context, id string) error {
	query := `DELETE FROM accounts WHERE id = ?`
	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete account: %w", err)
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	if rows == 0 {
		return fmt.Errorf("account not found")
	}
	return nil
}

func (r *accountRepository) GetTotalBalance(ctx context.Context) (int64, error) {
	query := `SELECT COALESCE(SUM(balance), 0) FROM accounts`
	var total int64
	err := r.db.QueryRowContext(ctx, query).Scan(&total)
	if err != nil {
		return 0, fmt.Errorf("failed to get total balance: %w", err)
	}
	return total, nil
}
