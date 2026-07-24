package postgres

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/schemahub/backend/internal/auth/domain"
	"github.com/schemahub/backend/internal/auth/repository"
)

type verificationTokenRepo struct {
	pool *pgxpool.Pool
}

func NewVerificationTokenRepo(pool *pgxpool.Pool) repository.VerificationTokenRepository {
	return &verificationTokenRepo{pool: pool}
}

func (r *verificationTokenRepo) Create(ctx context.Context, t *domain.VerificationToken) error {
	query := `INSERT INTO verification_tokens (user_id, email, token_hash, expires_at, created_at) VALUES ($1, $2, $3, $4, $5) RETURNING id`
	return r.pool.QueryRow(ctx, query, t.UserID, t.Email, t.TokenHash, t.ExpiresAt, time.Now()).Scan(&t.ID)
}

func (r *verificationTokenRepo) GetByHash(ctx context.Context, hash string) (*domain.VerificationToken, error) {
	query := `SELECT id, user_id, email, token_hash, expires_at, consumed_at, created_at FROM verification_tokens WHERE token_hash = $1`
	t := &domain.VerificationToken{}
	var consumedAt *time.Time
	err := r.pool.QueryRow(ctx, query, hash).Scan(&t.ID, &t.UserID, &t.Email, &t.TokenHash, &t.ExpiresAt, &consumedAt, &t.CreatedAt)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("verification token not found")
		}
		return nil, fmt.Errorf("getting verification token: %w", err)
	}
	t.ConsumedAt = consumedAt
	return t, nil
}

func (r *verificationTokenRepo) Consume(ctx context.Context, id string) error {
	query := `UPDATE verification_tokens SET consumed_at = $1 WHERE id = $2 AND consumed_at IS NULL`
	tag, err := r.pool.Exec(ctx, query, time.Now(), id)
	if err != nil {
		return fmt.Errorf("consuming verification token: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return fmt.Errorf("verification token not found or already consumed")
	}
	return nil
}

func (r *verificationTokenRepo) GetActiveByUserID(ctx context.Context, userID string) ([]*domain.VerificationToken, error) {
	query := `SELECT id, user_id, email, token_hash, expires_at, consumed_at, created_at FROM verification_tokens WHERE user_id = $1 AND consumed_at IS NULL AND expires_at > NOW()`
	rows, err := r.pool.Query(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("getting active verification tokens: %w", err)
	}
	defer rows.Close()

	var tokens []*domain.VerificationToken
	for rows.Next() {
		t := &domain.VerificationToken{}
		var consumedAt *time.Time
		if err := rows.Scan(&t.ID, &t.UserID, &t.Email, &t.TokenHash, &t.ExpiresAt, &consumedAt, &t.CreatedAt); err != nil {
			return nil, fmt.Errorf("scanning verification token: %w", err)
		}
		t.ConsumedAt = consumedAt
		tokens = append(tokens, t)
	}
	return tokens, rows.Err()
}
