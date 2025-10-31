package repository

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/billybbuffum/budget/internal/domain"
)

type transferSuggestionRepository struct {
	db *sql.DB
}

// NewTransferSuggestionRepository creates a new transfer suggestion repository
func NewTransferSuggestionRepository(db *sql.DB) domain.TransferSuggestionRepository {
	return &transferSuggestionRepository{db: db}
}

func (r *transferSuggestionRepository) Create(ctx context.Context, suggestion *domain.TransferMatchSuggestion) error {
	query := `
		INSERT INTO transfer_match_suggestions (id, transaction_a_id, transaction_b_id, match_score, confidence, status, is_credit_payment, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
	`
	_, err := r.db.ExecContext(ctx, query,
		suggestion.ID, suggestion.TransactionAID, suggestion.TransactionBID,
		suggestion.MatchScore, suggestion.Confidence, suggestion.Status,
		suggestion.IsCreditPayment, suggestion.CreatedAt, suggestion.UpdatedAt)
	if err != nil {
		return fmt.Errorf("failed to create transfer suggestion: %w", err)
	}
	return nil
}

func (r *transferSuggestionRepository) GetByID(ctx context.Context, id string) (*domain.TransferMatchSuggestion, error) {
	query := `
		SELECT id, transaction_a_id, transaction_b_id, match_score, confidence, status, is_credit_payment, created_at, updated_at
		FROM transfer_match_suggestions
		WHERE id = ?
	`
	suggestion := &domain.TransferMatchSuggestion{}
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&suggestion.ID, &suggestion.TransactionAID, &suggestion.TransactionBID,
		&suggestion.MatchScore, &suggestion.Confidence, &suggestion.Status,
		&suggestion.IsCreditPayment, &suggestion.CreatedAt, &suggestion.UpdatedAt)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("transfer suggestion not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get transfer suggestion: %w", err)
	}
	return suggestion, nil
}

func (r *transferSuggestionRepository) List(ctx context.Context) ([]*domain.TransferMatchSuggestion, error) {
	query := `
		SELECT id, transaction_a_id, transaction_b_id, match_score, confidence, status, is_credit_payment, created_at, updated_at
		FROM transfer_match_suggestions
		ORDER BY created_at DESC
	`
	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to list transfer suggestions: %w", err)
	}
	defer rows.Close()

	return r.scanSuggestions(rows)
}

func (r *transferSuggestionRepository) ListByStatus(ctx context.Context, status domain.SuggestionStatus) ([]*domain.TransferMatchSuggestion, error) {
	query := `
		SELECT id, transaction_a_id, transaction_b_id, match_score, confidence, status, is_credit_payment, created_at, updated_at
		FROM transfer_match_suggestions
		WHERE status = ?
		ORDER BY created_at DESC
	`
	rows, err := r.db.QueryContext(ctx, query, status)
	if err != nil {
		return nil, fmt.Errorf("failed to list transfer suggestions by status: %w", err)
	}
	defer rows.Close()

	return r.scanSuggestions(rows)
}

func (r *transferSuggestionRepository) ListPending(ctx context.Context) ([]*domain.TransferMatchSuggestion, error) {
	return r.ListByStatus(ctx, domain.SuggestionStatusPending)
}

func (r *transferSuggestionRepository) FindByTransactions(ctx context.Context, txnAID, txnBID string) (*domain.TransferMatchSuggestion, error) {
	// Check both directions (A->B and B->A)
	query := `
		SELECT id, transaction_a_id, transaction_b_id, match_score, confidence, status, is_credit_payment, created_at, updated_at
		FROM transfer_match_suggestions
		WHERE (transaction_a_id = ? AND transaction_b_id = ?)
		   OR (transaction_a_id = ? AND transaction_b_id = ?)
		LIMIT 1
	`
	suggestion := &domain.TransferMatchSuggestion{}
	err := r.db.QueryRowContext(ctx, query, txnAID, txnBID, txnBID, txnAID).Scan(
		&suggestion.ID, &suggestion.TransactionAID, &suggestion.TransactionBID,
		&suggestion.MatchScore, &suggestion.Confidence, &suggestion.Status,
		&suggestion.IsCreditPayment, &suggestion.CreatedAt, &suggestion.UpdatedAt)
	if err == sql.ErrNoRows {
		return nil, nil // Not found is not an error
	}
	if err != nil {
		return nil, fmt.Errorf("failed to find transfer suggestion: %w", err)
	}
	return suggestion, nil
}

func (r *transferSuggestionRepository) Update(ctx context.Context, suggestion *domain.TransferMatchSuggestion) error {
	query := `
		UPDATE transfer_match_suggestions
		SET match_score = ?, confidence = ?, status = ?, is_credit_payment = ?, updated_at = ?
		WHERE id = ?
	`
	result, err := r.db.ExecContext(ctx, query,
		suggestion.MatchScore, suggestion.Confidence, suggestion.Status,
		suggestion.IsCreditPayment, suggestion.UpdatedAt, suggestion.ID)
	if err != nil {
		return fmt.Errorf("failed to update transfer suggestion: %w", err)
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to check rows affected: %w", err)
	}
	if rows == 0 {
		return fmt.Errorf("transfer suggestion not found")
	}
	return nil
}

func (r *transferSuggestionRepository) Delete(ctx context.Context, id string) error {
	query := `DELETE FROM transfer_match_suggestions WHERE id = ?`
	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete transfer suggestion: %w", err)
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to check rows affected: %w", err)
	}
	if rows == 0 {
		return fmt.Errorf("transfer suggestion not found")
	}
	return nil
}

func (r *transferSuggestionRepository) scanSuggestions(rows *sql.Rows) ([]*domain.TransferMatchSuggestion, error) {
	var suggestions []*domain.TransferMatchSuggestion
	for rows.Next() {
		suggestion := &domain.TransferMatchSuggestion{}
		err := rows.Scan(
			&suggestion.ID, &suggestion.TransactionAID, &suggestion.TransactionBID,
			&suggestion.MatchScore, &suggestion.Confidence, &suggestion.Status,
			&suggestion.IsCreditPayment, &suggestion.CreatedAt, &suggestion.UpdatedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan transfer suggestion: %w", err)
		}
		suggestions = append(suggestions, suggestion)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating transfer suggestions: %w", err)
	}
	return suggestions, nil
}
