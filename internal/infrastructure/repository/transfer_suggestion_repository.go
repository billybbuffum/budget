package repository

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/billybbuffum/budget/internal/domain"
)

type transferSuggestionRepository struct {
	db *sql.DB
}

// NewTransferSuggestionRepository creates a new transfer suggestion repository
func NewTransferSuggestionRepository(db *sql.DB) domain.TransferSuggestionRepository {
	return &transferSuggestionRepository{db: db}
}

func (r *transferSuggestionRepository) Create(ctx context.Context, suggestion *domain.TransferSuggestion) error {
	query := `
		INSERT INTO transfer_match_suggestions (id, transaction_a_id, transaction_b_id, confidence, score, is_credit_payment, status, created_at, reviewed_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
	`
	_, err := r.db.ExecContext(ctx, query,
		suggestion.ID, suggestion.TransactionAID, suggestion.TransactionBID,
		suggestion.Confidence, suggestion.Score, suggestion.IsCreditPayment,
		suggestion.Status, suggestion.CreatedAt, suggestion.ReviewedAt)
	if err != nil {
		return fmt.Errorf("failed to create transfer suggestion: %w", err)
	}
	return nil
}

func (r *transferSuggestionRepository) GetByID(ctx context.Context, id string) (*domain.TransferSuggestion, error) {
	query := `
		SELECT id, transaction_a_id, transaction_b_id, confidence, score, is_credit_payment, status, created_at, reviewed_at
		FROM transfer_match_suggestions
		WHERE id = ?
	`
	suggestion := &domain.TransferSuggestion{}
	var reviewedAt sql.NullTime
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&suggestion.ID, &suggestion.TransactionAID, &suggestion.TransactionBID,
		&suggestion.Confidence, &suggestion.Score, &suggestion.IsCreditPayment,
		&suggestion.Status, &suggestion.CreatedAt, &reviewedAt)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("transfer suggestion not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get transfer suggestion: %w", err)
	}
	if reviewedAt.Valid {
		suggestion.ReviewedAt = &reviewedAt.Time
	}
	return suggestion, nil
}

func (r *transferSuggestionRepository) List(ctx context.Context) ([]*domain.TransferSuggestion, error) {
	query := `
		SELECT id, transaction_a_id, transaction_b_id, confidence, score, is_credit_payment, status, created_at, reviewed_at
		FROM transfer_match_suggestions
		ORDER BY score DESC, created_at DESC
	`
	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to list transfer suggestions: %w", err)
	}
	defer rows.Close()

	return r.scanSuggestions(rows)
}

func (r *transferSuggestionRepository) ListPending(ctx context.Context) ([]*domain.TransferSuggestion, error) {
	query := `
		SELECT id, transaction_a_id, transaction_b_id, confidence, score, is_credit_payment, status, created_at, reviewed_at
		FROM transfer_match_suggestions
		WHERE status = 'pending'
		ORDER BY score DESC, created_at DESC
	`
	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to list pending transfer suggestions: %w", err)
	}
	defer rows.Close()

	return r.scanSuggestions(rows)
}

func (r *transferSuggestionRepository) ListByConfidence(ctx context.Context, confidence string) ([]*domain.TransferSuggestion, error) {
	query := `
		SELECT id, transaction_a_id, transaction_b_id, confidence, score, is_credit_payment, status, created_at, reviewed_at
		FROM transfer_match_suggestions
		WHERE confidence = ?
		ORDER BY score DESC, created_at DESC
	`
	rows, err := r.db.QueryContext(ctx, query, confidence)
	if err != nil {
		return nil, fmt.Errorf("failed to list transfer suggestions by confidence: %w", err)
	}
	defer rows.Close()

	return r.scanSuggestions(rows)
}

func (r *transferSuggestionRepository) Accept(ctx context.Context, suggestionID string) error {
	query := `
		UPDATE transfer_match_suggestions
		SET status = 'accepted', reviewed_at = ?
		WHERE id = ?
	`
	now := time.Now()
	result, err := r.db.ExecContext(ctx, query, now, suggestionID)
	if err != nil {
		return fmt.Errorf("failed to accept transfer suggestion: %w", err)
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	if rows == 0 {
		return fmt.Errorf("transfer suggestion not found")
	}
	return nil
}

func (r *transferSuggestionRepository) Reject(ctx context.Context, suggestionID string) error {
	query := `
		UPDATE transfer_match_suggestions
		SET status = 'rejected', reviewed_at = ?
		WHERE id = ?
	`
	now := time.Now()
	result, err := r.db.ExecContext(ctx, query, now, suggestionID)
	if err != nil {
		return fmt.Errorf("failed to reject transfer suggestion: %w", err)
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	if rows == 0 {
		return fmt.Errorf("transfer suggestion not found")
	}
	return nil
}

func (r *transferSuggestionRepository) DeleteByTransactionID(ctx context.Context, transactionID string) error {
	query := `
		DELETE FROM transfer_match_suggestions
		WHERE transaction_a_id = ? OR transaction_b_id = ?
	`
	_, err := r.db.ExecContext(ctx, query, transactionID, transactionID)
	if err != nil {
		return fmt.Errorf("failed to delete transfer suggestions by transaction ID: %w", err)
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
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	if rows == 0 {
		return fmt.Errorf("transfer suggestion not found")
	}
	return nil
}

func (r *transferSuggestionRepository) scanSuggestions(rows *sql.Rows) ([]*domain.TransferSuggestion, error) {
	var suggestions []*domain.TransferSuggestion
	for rows.Next() {
		suggestion := &domain.TransferSuggestion{}
		var reviewedAt sql.NullTime
		if err := rows.Scan(&suggestion.ID, &suggestion.TransactionAID, &suggestion.TransactionBID,
			&suggestion.Confidence, &suggestion.Score, &suggestion.IsCreditPayment,
			&suggestion.Status, &suggestion.CreatedAt, &reviewedAt); err != nil {
			return nil, fmt.Errorf("failed to scan transfer suggestion: %w", err)
		}
		if reviewedAt.Valid {
			suggestion.ReviewedAt = &reviewedAt.Time
		}
		suggestions = append(suggestions, suggestion)
	}
	return suggestions, nil
}
