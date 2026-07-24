package domain

import "context"

type UserRepository interface {
	Create(ctx context.Context, user *User) error
	GetByID(ctx context.Context, id string) (*User, error)
	GetByEmail(ctx context.Context, email string) (*User, error)
	Update(ctx context.Context, user *User) error
	UpdatePassword(ctx context.Context, id, passwordHash string) error
	SoftDelete(ctx context.Context, id string) error
}

type RefreshTokenRepository interface {
	Create(ctx context.Context, token *RefreshToken) error
	GetByHash(ctx context.Context, hash string) (*RefreshToken, error)
	Revoke(ctx context.Context, id string) error
	RevokeFamily(ctx context.Context, family string) error
	GetActiveByUserID(ctx context.Context, userID string) ([]*RefreshToken, error)
}

type VerificationTokenRepository interface {
	Create(ctx context.Context, token *VerificationToken) error
	GetByHash(ctx context.Context, hash string) (*VerificationToken, error)
	Consume(ctx context.Context, id string) error
	GetActiveByUserID(ctx context.Context, userID string) ([]*VerificationToken, error)
}

type OAuthIdentityRepository interface {
	Create(ctx context.Context, identity *OAuthIdentity) error
	GetByProvider(ctx context.Context, provider, providerUserID string) (*OAuthIdentity, error)
	GetByUserID(ctx context.Context, userID string) ([]*OAuthIdentity, error)
	Delete(ctx context.Context, id string) error
	UpdateLastUsed(ctx context.Context, id string) error
}
