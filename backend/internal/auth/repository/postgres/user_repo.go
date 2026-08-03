package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/schemahub/backend/internal/auth/domain"
	"github.com/schemahub/backend/internal/auth/repository"
)

type userRepo struct {
	pool *pgxpool.Pool
}

func NewUserRepository(pool *pgxpool.Pool) repository.UserRepository {
	return &userRepo{pool: pool}
}

func (r *userRepo) Create(ctx context.Context, u *domain.User) error {
	_, err := r.pool.Exec(ctx,
		`INSERT INTO users (id, email, password_hash, display_name, avatar_url, role, is_active)
		 VALUES (gen_random_uuid(), $1, $2, $3, $4, $5, $6)
		 RETURNING id`,
		u.Email, u.PasswordHash, u.DisplayName, u.AvatarURL, u.Role, u.IsActive)
	if err != nil {
		if pgxErrCode(err) == "23505" {
			return domain.ErrEmailAlreadyExists
		}
		return fmt.Errorf("inserting user: %w", err)
	}
	return nil
}

func (r *userRepo) GetByID(ctx context.Context, id string) (*domain.User, error) {
	row := r.pool.QueryRow(ctx,
		`SELECT id, email, password_hash, display_name, avatar_url, role, is_active,
		        email_verified_at, last_login_at, created_at, updated_at
		 FROM users WHERE id = $1 AND deleted_at IS NULL`, id)
	return scanUser(row)
}

func (r *userRepo) GetByEmail(ctx context.Context, email string) (*domain.User, error) {
	row := r.pool.QueryRow(ctx,
		`SELECT id, email, password_hash, display_name, avatar_url, role, is_active,
		        email_verified_at, last_login_at, created_at, updated_at
		 FROM users WHERE email = $1 AND deleted_at IS NULL`, email)
	return scanUser(row)
}

func (r *userRepo) Update(ctx context.Context, u *domain.User) error {
	_, err := r.pool.Exec(ctx,
		`UPDATE users SET display_name=$1, avatar_url=$2, updated_at=now()
		 WHERE id=$3 AND deleted_at IS NULL`,
		u.DisplayName, u.AvatarURL, u.ID)
	return err
}

func (r *userRepo) SoftDelete(ctx context.Context, id string) error {
	_, err := r.pool.Exec(ctx,
		`UPDATE users SET deleted_at=now() WHERE id=$1`, id)
	return err
}

func (r *userRepo) UpdatePassword(ctx context.Context, id, passwordHash string) error {
	_, err := r.pool.Exec(ctx,
		`UPDATE users SET password_hash=$1, updated_at=now() WHERE id=$2 AND deleted_at IS NULL`,
		passwordHash, id)
	return err
}

func scanUser(row pgx.Row) (*domain.User, error) {
	u := &domain.User{}
	err := row.Scan(&u.ID, &u.Email, &u.PasswordHash, &u.DisplayName,
		&u.AvatarURL, &u.Role, &u.IsActive,
		&u.EmailVerifiedAt, &u.LastLoginAt, &u.CreatedAt, &u.UpdatedAt)
	if err == pgx.ErrNoRows {
		return nil, domain.ErrUserNotFound
	}
	return u, err
}

func pgxErrCode(err error) string {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		return pgErr.Code
	}
	return ""
}
