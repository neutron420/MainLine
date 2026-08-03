package postgres

import (
	"context"
	"crypto/sha256"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/schemahub/backend/internal/auth/domain"
	"github.com/schemahub/backend/internal/auth/repository"
)

type refreshTokenRepo struct {
	pool *pgxpool.Pool
}

func NewRefreshTokenRepository(pool *pgxpool.Pool) repository.RefreshTokenRepository {
	return &refreshTokenRepo{pool: pool}
}

func hashToken(token string) string {
	h := sha256.Sum256([]byte(token))
	return fmt.Sprintf("%x", h)
}

func (r *refreshTokenRepo) Create(ctx context.Context, t *domain.RefreshToken) error {
	_, err := r.pool.Exec(ctx,
		`INSERT INTO refresh_tokens (id, user_id, token_hash, expires_at, created_by_ip, family)
		 VALUES (gen_random_uuid(), $1, $2, $3, $4, $5)`,
		t.UserID, t.TokenHash, t.ExpiresAt, t.CreatedByIP, t.Family)
	return err
}

func (r *refreshTokenRepo) GetByHash(ctx context.Context, rawToken string) (*domain.RefreshToken, error) {
	row := r.pool.QueryRow(ctx,
		`SELECT id, user_id, token_hash, expires_at, revoked_at, created_at, created_by_ip::text, family
		 FROM refresh_tokens WHERE token_hash = $1`, hashToken(rawToken))
	t := &domain.RefreshToken{}
	err := row.Scan(&t.ID, &t.UserID, &t.TokenHash, &t.ExpiresAt,
		&t.RevokedAt, &t.CreatedAt, &t.CreatedByIP, &t.Family)
	if err == pgx.ErrNoRows {
		return nil, domain.ErrInvalidRefreshToken
	}
	return t, err
}

func (r *refreshTokenRepo) Revoke(ctx context.Context, id string) error {
	_, err := r.pool.Exec(ctx,
		`UPDATE refresh_tokens SET revoked_at=now() WHERE id=$1`, id)
	return err
}

func (r *refreshTokenRepo) RevokeFamily(ctx context.Context, family string) error {
	_, err := r.pool.Exec(ctx,
		`UPDATE refresh_tokens SET revoked_at=now() WHERE family=$1 AND revoked_at IS NULL`, family)
	return err
}

func (r *refreshTokenRepo) GetActiveByUserID(ctx context.Context, userID string) ([]*domain.RefreshToken, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT id, user_id, token_hash, expires_at, revoked_at, created_at, created_by_ip::text, family
		 FROM refresh_tokens
		 WHERE user_id=$1 AND revoked_at IS NULL AND expires_at > now()
		 ORDER BY created_at DESC`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tokens []*domain.RefreshToken
	for rows.Next() {
		t := &domain.RefreshToken{}
		if err := rows.Scan(&t.ID, &t.UserID, &t.TokenHash, &t.ExpiresAt,
			&t.RevokedAt, &t.CreatedAt, &t.CreatedByIP, &t.Family); err != nil {
			return nil, err
		}
		tokens = append(tokens, t)
	}
	return tokens, nil
}
