package domain

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/schemahub/backend/internal/pkg/jwt"
	"golang.org/x/crypto/bcrypt"
)

type AuthService struct {
	userRepo    UserRepository
	tokenRepo   RefreshTokenRepository
	oauthRepo   OAuthIdentityRepository
	jwtManager  *jwt.Manager
}

func NewAuthService(
	userRepo UserRepository,
	tokenRepo RefreshTokenRepository,
	oauthRepo OAuthIdentityRepository,
	jwtManager *jwt.Manager,
) *AuthService {
	return &AuthService{
		userRepo:   userRepo,
		tokenRepo:  tokenRepo,
		oauthRepo:  oauthRepo,
		jwtManager: jwtManager,
	}
}

func (s *AuthService) Register(ctx context.Context, email, password, displayName string) (*User, string, string, error) {
	if !ValidateEmail(email) {
		return nil, "", "", fmt.Errorf("invalid email format")
	}
	if err := ValidatePassword(password); err != nil {
		return nil, "", "", err
	}
	if displayName == "" {
		return nil, "", "", fmt.Errorf("display name is required")
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, "", "", fmt.Errorf("hashing password: %w", err)
	}

	user := &User{
		Email:        email,
		PasswordHash: string(hash),
		DisplayName:  displayName,
		Role:         "user",
		IsActive:     true,
	}

	if err := s.userRepo.Create(ctx, user); err != nil {
		return nil, "", "", err
	}

	created, err := s.userRepo.GetByEmail(ctx, email)
	if err != nil {
		return nil, "", "", err
	}

	accessToken, err := s.jwtManager.GenerateAccessToken(created.ID, created.Email, created.Role)
	if err != nil {
		return nil, "", "", fmt.Errorf("generating access token: %w", err)
	}

	refreshToken, err := s.generateRefreshToken(ctx, created.ID, "")
	if err != nil {
		return nil, "", "", err
	}

	return created, accessToken, refreshToken, nil
}

func (s *AuthService) Login(ctx context.Context, email, password string) (*User, string, string, error) {
	user, err := s.userRepo.GetByEmail(ctx, email)
	if err != nil {
		return nil, "", "", ErrInvalidCredentials
	}

	if !user.IsActive {
		return nil, "", "", ErrInvalidCredentials
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		return nil, "", "", ErrInvalidCredentials
	}

	accessToken, err := s.jwtManager.GenerateAccessToken(user.ID, user.Email, user.Role)
	if err != nil {
		return nil, "", "", fmt.Errorf("generating access token: %w", err)
	}

	refreshToken, err := s.generateRefreshToken(ctx, user.ID, "")
	if err != nil {
		return nil, "", "", err
	}

	return user, accessToken, refreshToken, nil
}

func (s *AuthService) RefreshToken(ctx context.Context, rawToken string) (string, string, error) {
	stored, err := s.tokenRepo.GetByHash(ctx, rawToken)
	if err != nil {
		return "", "", err
	}

	if stored.RevokedAt != nil {
		if err := s.tokenRepo.RevokeFamily(ctx, stored.Family); err != nil {
			return "", "", fmt.Errorf("revoking token family: %w", err)
		}
		return "", "", ErrTokenRevoked
	}

	if time.Now().After(stored.ExpiresAt) {
		return "", "", ErrInvalidRefreshToken
	}

	if err := s.tokenRepo.Revoke(ctx, stored.ID); err != nil {
		return "", "", fmt.Errorf("revoking old token: %w", err)
	}

	user, err := s.userRepo.GetByID(ctx, stored.UserID)
	if err != nil {
		return "", "", err
	}

	accessToken, err := s.jwtManager.GenerateAccessToken(user.ID, user.Email, user.Role)
	if err != nil {
		return "", "", err
	}

	newRefresh, err := s.generateRefreshToken(ctx, user.ID, stored.Family)
	if err != nil {
		return "", "", err
	}

	return accessToken, newRefresh, nil
}

func (s *AuthService) Logout(ctx context.Context, rawToken string) error {
	stored, err := s.tokenRepo.GetByHash(ctx, rawToken)
	if err != nil {
		return nil
	}
	return s.tokenRepo.Revoke(ctx, stored.ID)
}

func (s *AuthService) GetUserByID(ctx context.Context, id string) (*User, error) {
	return s.userRepo.GetByID(ctx, id)
}

func (s *AuthService) UpdateUser(ctx context.Context, id, displayName, avatarURL string) (*User, error) {
	user, err := s.userRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if displayName != "" {
		user.DisplayName = displayName
	}
	if avatarURL != "" {
		user.AvatarURL = avatarURL
	}
	if err := s.userRepo.Update(ctx, user); err != nil {
		return nil, err
	}
	return user, nil
}

func (s *AuthService) ChangePassword(ctx context.Context, userID, currentPassword, newPassword string) error {
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return err
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(currentPassword)); err != nil {
		return ErrPasswordMismatch
	}

	if err := ValidatePassword(newPassword); err != nil {
		return err
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("hashing password: %w", err)
	}

	return s.userRepo.UpdatePassword(ctx, userID, string(hash))
}

func (s *AuthService) generateRefreshToken(ctx context.Context, userID, family string) (string, error) {
	if family == "" {
		b := make([]byte, 16)
		if _, err := rand.Read(b); err != nil {
			return "", fmt.Errorf("generating family: %w", err)
		}
		family = hex.EncodeToString(b)
	}

	b := make([]byte, 48)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("generating token: %w", err)
	}
	rawToken := fmt.Sprintf("rt_%x", b)
	tokenHash := fmt.Sprintf("%x", sha256.Sum256([]byte(rawToken)))

	t := &RefreshToken{
		UserID:    userID,
		TokenHash: tokenHash,
		ExpiresAt: time.Now().Add(7 * 24 * time.Hour),
		Family:    family,
	}

	if err := s.tokenRepo.Create(ctx, t); err != nil {
		return "", fmt.Errorf("storing refresh token: %w", err)
	}

	return rawToken, nil
}

// ── OAuth ──

func (s *AuthService) GetOAuthURL(provider, redirectTo string, linking bool) (string, string, error) {
	return "", "", fmt.Errorf("not implemented")
}

func (s *AuthService) HandleOAuthCallback(ctx context.Context, provider, code, state, codeVerifier string) (*User, string, string, bool, bool, error) {
	return nil, "", "", false, false, fmt.Errorf("not implemented")
}

func (s *AuthService) LinkOAuthIdentity(ctx context.Context, userID, provider, code, state string) error {
	return fmt.Errorf("not implemented")
}

func (s *AuthService) UnlinkOAuthIdentity(ctx context.Context, userID, provider string) error {
	return fmt.Errorf("not implemented")
}

func (s *AuthService) ListLinkedIdentities(ctx context.Context, userID string) ([]*OAuthIdentity, error) {
	return nil, fmt.Errorf("not implemented")
}

