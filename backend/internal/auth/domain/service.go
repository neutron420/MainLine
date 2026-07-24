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
	verifyRepo  VerificationTokenRepository
	jwtManager  *jwt.Manager
	oauthConfigs *OAuthProviderConfig
}

func NewAuthService(
	userRepo UserRepository,
	tokenRepo RefreshTokenRepository,
	oauthRepo OAuthIdentityRepository,
	verifyRepo VerificationTokenRepository,
	jwtManager *jwt.Manager,
	oauthConfigs *OAuthProviderConfig,
) *AuthService {
	return &AuthService{
		userRepo:     userRepo,
		tokenRepo:    tokenRepo,
		oauthRepo:    oauthRepo,
		verifyRepo:   verifyRepo,
		jwtManager:   jwtManager,
		oauthConfigs: oauthConfigs,
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

func (s *AuthService) SendVerificationEmail(ctx context.Context, email string) error {
	user, err := s.userRepo.GetByEmail(ctx, email)
	if err != nil {
		return ErrUserNotFound
	}
	if user.EmailVerifiedAt != nil {
		return nil
	}

	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return fmt.Errorf("generating verification token: %w", err)
	}
	rawToken := fmt.Sprintf("verify_%x", b)
	tokenHash := fmt.Sprintf("%x", sha256.Sum256([]byte(rawToken)))

	t := &VerificationToken{
		UserID:    user.ID,
		Email:     user.Email,
		TokenHash: tokenHash,
		ExpiresAt: time.Now().Add(24 * time.Hour),
	}

	if err := s.verifyRepo.Create(ctx, t); err != nil {
		return fmt.Errorf("storing verification token: %w", err)
	}

	return nil
}

func (s *AuthService) VerifyEmail(ctx context.Context, rawToken string) error {
	tokenHash := fmt.Sprintf("%x", sha256.Sum256([]byte(rawToken)))
	vt, err := s.verifyRepo.GetByHash(ctx, tokenHash)
	if err != nil {
		return ErrInvalidCredentials
	}

	if vt.ConsumedAt != nil {
		return ErrInvalidCredentials
	}

	if time.Now().After(vt.ExpiresAt) {
		return ErrInvalidCredentials
	}

	now := time.Now()
	vt.ConsumedAt = &now
	if err := s.verifyRepo.Consume(ctx, vt.ID); err != nil {
		return fmt.Errorf("consuming verification token: %w", err)
	}

	user, err := s.userRepo.GetByID(ctx, vt.UserID)
	if err != nil {
		return ErrUserNotFound
	}

	user.EmailVerifiedAt = &now
	if err := s.userRepo.Update(ctx, user); err != nil {
		return fmt.Errorf("updating user verification: %w", err)
	}

	return nil
}

func (s *AuthService) DeleteAccount(ctx context.Context, userID, password string) error {
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return ErrUserNotFound
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		return ErrPasswordMismatch
	}

	activeTokens, err := s.tokenRepo.GetActiveByUserID(ctx, userID)
	if err == nil {
		for _, t := range activeTokens {
			if err := s.tokenRepo.Revoke(ctx, t.ID); err != nil {
				return fmt.Errorf("revoking token %s: %w", t.ID, err)
			}
		}
	}

	return s.userRepo.SoftDelete(ctx, userID)
}

func (s *AuthService) ForgotPassword(ctx context.Context, email string) error {
	user, err := s.userRepo.GetByEmail(ctx, email)
	if err != nil {
		return nil
	}

	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return fmt.Errorf("generating reset token: %w", err)
	}
	rawToken := fmt.Sprintf("reset_%x", b)
	tokenHash := fmt.Sprintf("%x", sha256.Sum256([]byte(rawToken)))

	t := &VerificationToken{
		UserID:    user.ID,
		Email:     user.Email,
		TokenHash: tokenHash,
		ExpiresAt: time.Now().Add(1 * time.Hour),
	}

	if err := s.verifyRepo.Create(ctx, t); err != nil {
		return fmt.Errorf("storing reset token: %w", err)
	}

	return nil
}

func (s *AuthService) ResetPassword(ctx context.Context, token, newPassword string) error {
	tokenHash := fmt.Sprintf("%x", sha256.Sum256([]byte(token)))
	vt, err := s.verifyRepo.GetByHash(ctx, tokenHash)
	if err != nil {
		return ErrInvalidCredentials
	}

	if vt.ConsumedAt != nil {
		return ErrInvalidCredentials
	}

	if time.Now().After(vt.ExpiresAt) {
		return ErrInvalidCredentials
	}

	now := time.Now()
	vt.ConsumedAt = &now
	if err := s.verifyRepo.Consume(ctx, vt.ID); err != nil {
		return fmt.Errorf("consuming reset token: %w", err)
	}

	user, err := s.userRepo.GetByID(ctx, vt.UserID)
	if err != nil {
		return ErrUserNotFound
	}

	if err := ValidatePassword(newPassword); err != nil {
		return err
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("hashing password: %w", err)
	}

	if err := s.userRepo.UpdatePassword(ctx, user.ID, string(hash)); err != nil {
		return fmt.Errorf("updating password: %w", err)
	}

	activeTokens, err := s.tokenRepo.GetActiveByUserID(ctx, user.ID)
	if err == nil {
		for _, t := range activeTokens {
			if err := s.tokenRepo.Revoke(ctx, t.ID); err != nil {
				return fmt.Errorf("revoking token %s: %w", t.ID, err)
			}
		}
	}

	return nil
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



