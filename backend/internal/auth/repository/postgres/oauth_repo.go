package postgres

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/schemahub/backend/internal/auth/domain"
	"github.com/schemahub/backend/internal/auth/repository"
)

type oauthIdentityRepo struct {
	pool *pgxpool.Pool
}

func NewOAuthIdentityRepository(pool *pgxpool.Pool) repository.OAuthIdentityRepository {
	return &oauthIdentityRepo{pool: pool}
}

func (r *oauthIdentityRepo) Create(ctx context.Context, o *domain.OAuthIdentity) error {
	_, err := r.pool.Exec(ctx,
		`INSERT INTO oauth_identities (id, user_id, provider, provider_user_id, provider_email,
		 access_token_encrypted, refresh_token_encrypted, expires_at)
		 VALUES (gen_random_uuid(), $1, $2, $3, $4, $5, $6, $7)`,
		o.UserID, o.Provider, o.ProviderUserID, o.ProviderEmail,
		o.AccessTokenEncrypted, o.RefreshTokenEncrypted, o.ExpiresAt)
	if pgxErrCode(err) == "23505" {
		return domain.ErrProviderAlreadyLinked
	}
	return err
}

func (r *oauthIdentityRepo) GetByProvider(ctx context.Context, provider, providerUserID string) (*domain.OAuthIdentity, error) {
	row := r.pool.QueryRow(ctx,
		`SELECT id, user_id, provider, provider_user_id, provider_email,
		        access_token_encrypted, refresh_token_encrypted, expires_at, created_at, last_used_at
		 FROM oauth_identities WHERE provider=$1 AND provider_user_id=$2`, provider, providerUserID)
	o := &domain.OAuthIdentity{}
	err := row.Scan(&o.ID, &o.UserID, &o.Provider, &o.ProviderUserID, &o.ProviderEmail,
		&o.AccessTokenEncrypted, &o.RefreshTokenEncrypted, &o.ExpiresAt, &o.CreatedAt, &o.LastUsedAt)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	return o, err
}

func (r *oauthIdentityRepo) GetByUserID(ctx context.Context, userID string) ([]*domain.OAuthIdentity, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT id, user_id, provider, provider_user_id, provider_email,
		        access_token_encrypted, refresh_token_encrypted, expires_at, created_at, last_used_at
		 FROM oauth_identities WHERE user_id=$1`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var identities []*domain.OAuthIdentity
	for rows.Next() {
		o := &domain.OAuthIdentity{}
		if err := rows.Scan(&o.ID, &o.UserID, &o.Provider, &o.ProviderUserID, &o.ProviderEmail,
			&o.AccessTokenEncrypted, &o.RefreshTokenEncrypted, &o.ExpiresAt, &o.CreatedAt, &o.LastUsedAt); err != nil {
			return nil, err
		}
		identities = append(identities, o)
	}
	return identities, nil
}

func (r *oauthIdentityRepo) Delete(ctx context.Context, id string) error {
	_, err := r.pool.Exec(ctx, `DELETE FROM oauth_identities WHERE id=$1`, id)
	return err
}

func (r *oauthIdentityRepo) UpdateLastUsed(ctx context.Context, id string) error {
	_, err := r.pool.Exec(ctx,
		`UPDATE oauth_identities SET last_used_at=now() WHERE id=$1`, id)
	return err
}

func (r *oauthIdentityRepo) GetExpiringSoon(ctx context.Context, within time.Duration) ([]*domain.OAuthIdentity, error) {
	cutoff := time.Now().Add(within)
	rows, err := r.pool.Query(ctx,
		`SELECT id, user_id, provider, provider_user_id, provider_email,
		        access_token_encrypted, refresh_token_encrypted, expires_at, created_at, last_used_at
		 FROM oauth_identities WHERE expires_at IS NOT NULL AND expires_at <= $1 AND refresh_token_encrypted IS NOT NULL AND refresh_token_encrypted != ''`,
		cutoff)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var identities []*domain.OAuthIdentity
	for rows.Next() {
		o := &domain.OAuthIdentity{}
		if err := rows.Scan(&o.ID, &o.UserID, &o.Provider, &o.ProviderUserID, &o.ProviderEmail,
			&o.AccessTokenEncrypted, &o.RefreshTokenEncrypted, &o.ExpiresAt, &o.CreatedAt, &o.LastUsedAt); err != nil {
			return nil, err
		}
		identities = append(identities, o)
	}
	return identities, nil
}

func (r *oauthIdentityRepo) UpdateTokens(ctx context.Context, id, accessTokenEncrypted, refreshTokenEncrypted string, expiresAt *time.Time) error {
	_, err := r.pool.Exec(ctx,
		`UPDATE oauth_identities SET access_token_encrypted=$1, refresh_token_encrypted=$2, expires_at=$3 WHERE id=$4`,
		accessTokenEncrypted, refreshTokenEncrypted, expiresAt, id)
	return err
}


