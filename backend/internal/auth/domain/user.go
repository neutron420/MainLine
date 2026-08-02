package domain

import (
	"errors"
	"time"
)

var (
	ErrEmailAlreadyExists        = errors.New("email already registered")
	ErrInvalidCredentials        = errors.New("invalid email or password")
	ErrUserNotFound              = errors.New("user not found")
	ErrInvalidRefreshToken       = errors.New("invalid or expired refresh token")
	ErrTokenRevoked              = errors.New("token has been revoked")
	ErrPasswordMismatch          = errors.New("current password is incorrect")
	ErrWeakPassword              = errors.New("password does not meet strength requirements")
	ErrEmailNotVerified          = errors.New("email not verified")
	ErrProviderAlreadyLinked     = errors.New("provider already linked to another account")
	ErrLastAuthMethod            = errors.New("cannot remove last authentication method")
	ErrOAuthStateMismatch        = errors.New("OAuth state mismatch — possible CSRF")
	ErrEmailVerificationRequired = errors.New("email verification required")
	ErrPermissionDenied          = errors.New("permission denied")
)

type User struct {
	ID              string
	Email           string
	PasswordHash    string
	DisplayName     string
	AvatarURL       string
	Role            string
	IsActive        bool
	EmailVerifiedAt *time.Time
	LastLoginAt     *time.Time
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

type RefreshToken struct {
	ID          string
	UserID      string
	TokenHash   string
	ExpiresAt   time.Time
	RevokedAt   *time.Time
	CreatedAt   time.Time
	CreatedByIP string
	Family      string
}

type VerificationToken struct {
	ID         string
	UserID     string
	Email      string
	TokenHash  string
	ExpiresAt  time.Time
	ConsumedAt *time.Time
	CreatedAt  time.Time
}

type OAuthIdentity struct {
	ID                    string
	UserID                string
	Provider              string
	ProviderUserID        string
	ProviderEmail         string
	AccessTokenEncrypted  string
	RefreshTokenEncrypted string
	ExpiresAt             *time.Time
	CreatedAt             time.Time
	LastUsedAt            *time.Time
}

func ValidatePassword(password string) error {
	if len(password) < 8 {
		return ErrWeakPassword
	}
	if len(password) > 128 {
		return ErrWeakPassword
	}
	hasUpper, hasLower, hasDigit := false, false, false
	for _, c := range password {
		switch {
		case c >= 'A' && c <= 'Z':
			hasUpper = true
		case c >= 'a' && c <= 'z':
			hasLower = true
		case c >= '0' && c <= '9':
			hasDigit = true
		}
	}
	if !hasUpper || !hasLower || !hasDigit {
		return ErrWeakPassword
	}
	return nil
}

func ValidateEmail(email string) bool {
	if len(email) < 3 || len(email) > 320 {
		return false
	}
	hasAt := false
	for _, c := range email {
		if c == '@' {
			if hasAt {
				return false
			}
			hasAt = true
		}
	}
	return hasAt
}
