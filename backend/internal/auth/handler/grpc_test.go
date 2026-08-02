package handler

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"testing"
	"time"

	"github.com/schemahub/backend/internal/auth/domain"
	"github.com/schemahub/backend/internal/pkg/interceptor"
	"github.com/schemahub/backend/internal/pkg/jwt"
	authv1 "github.com/schemahub/backend/proto/auth/v1"
	"golang.org/x/crypto/bcrypt"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type fakeAuthRepo struct {
	users     map[string]*domain.User
	byEmail   map[string]*domain.User
	createErr error
	getErr    error
}

func newFakeAuthRepo() *fakeAuthRepo {
	return &fakeAuthRepo{users: map[string]*domain.User{}, byEmail: map[string]*domain.User{}}
}

func (f *fakeAuthRepo) Create(ctx context.Context, user *domain.User) error {
	if f.createErr != nil {
		return f.createErr
	}
	user.ID = "user_1"
	f.users[user.ID] = user
	f.byEmail[user.Email] = user
	return nil
}

func (f *fakeAuthRepo) GetByID(ctx context.Context, id string) (*domain.User, error) {
	if f.getErr != nil {
		return nil, f.getErr
	}
	u, ok := f.users[id]
	if !ok {
		return nil, domain.ErrUserNotFound
	}
	return u, nil
}

func (f *fakeAuthRepo) GetByEmail(ctx context.Context, email string) (*domain.User, error) {
	u, ok := f.byEmail[email]
	if !ok {
		return nil, errors.New("not found")
	}
	return u, nil
}

func (f *fakeAuthRepo) Update(ctx context.Context, user *domain.User) error {
	f.users[user.ID] = user
	f.byEmail[user.Email] = user
	return nil
}

func (f *fakeAuthRepo) UpdatePassword(ctx context.Context, id, passwordHash string) error {
	f.users[id].PasswordHash = passwordHash
	return nil
}

func (f *fakeAuthRepo) SoftDelete(ctx context.Context, id string) error { return nil }

type fakeTokenRepo struct{}

func (f *fakeTokenRepo) Create(ctx context.Context, token *domain.RefreshToken) error { return nil }
func (f *fakeTokenRepo) GetByHash(ctx context.Context, hash string) (*domain.RefreshToken, error) {
	return nil, errors.New("not found")
}
func (f *fakeTokenRepo) Revoke(ctx context.Context, id string) error { return nil }
func (f *fakeTokenRepo) RevokeFamily(ctx context.Context, family string) error {
	return nil
}
func (f *fakeTokenRepo) GetActiveByUserID(ctx context.Context, userID string) ([]*domain.RefreshToken, error) {
	return nil, nil
}

type fakeVerifyRepo struct{}

func (f *fakeVerifyRepo) Create(ctx context.Context, token *domain.VerificationToken) error {
	return nil
}
func (f *fakeVerifyRepo) GetByHash(ctx context.Context, hash string) (*domain.VerificationToken, error) {
	return nil, errors.New("not found")
}
func (f *fakeVerifyRepo) Consume(ctx context.Context, id string) error { return nil }
func (f *fakeVerifyRepo) GetActiveByUserID(ctx context.Context, userID string) ([]*domain.VerificationToken, error) {
	return nil, nil
}

type fakeOAuthRepo struct{}

func (f *fakeOAuthRepo) Create(ctx context.Context, identity *domain.OAuthIdentity) error { return nil }
func (f *fakeOAuthRepo) GetByProvider(ctx context.Context, provider, providerUserID string) (*domain.OAuthIdentity, error) {
	return nil, errors.New("not found")
}
func (f *fakeOAuthRepo) GetByUserID(ctx context.Context, userID string) ([]*domain.OAuthIdentity, error) {
	return nil, nil
}
func (f *fakeOAuthRepo) Delete(ctx context.Context, id string) error { return nil }
func (f *fakeOAuthRepo) UpdateLastUsed(ctx context.Context, id string) error {
	return nil
}
func (f *fakeOAuthRepo) GetExpiringSoon(ctx context.Context, within time.Duration) ([]*domain.OAuthIdentity, error) {
	return nil, nil
}
func (f *fakeOAuthRepo) UpdateTokens(ctx context.Context, id, accessTokenEncrypted, refreshTokenEncrypted string, expiresAt *time.Time) error {
	return nil
}

func testKeyPair(t *testing.T) (privatePEM, publicPEM string) {
	t.Helper()
	priv, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("generating RSA key: %v", err)
	}
	privatePEM = string(pem.EncodeToMemory(&pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(priv),
	}))
	pubBytes, err := x509.MarshalPKIXPublicKey(&priv.PublicKey)
	if err != nil {
		t.Fatalf("marshaling public key: %v", err)
	}
	publicPEM = string(pem.EncodeToMemory(&pem.Block{Type: "PUBLIC KEY", Bytes: pubBytes}))
	return privatePEM, publicPEM
}

func testAuthHandler(t *testing.T, repo *fakeAuthRepo) *AuthHandler {
	t.Helper()
	priv, pub := testKeyPair(t)
	m, err := jwt.NewManager(priv, pub)
	if err != nil {
		t.Fatalf("NewManager() error = %v", err)
	}
	svc := domain.NewAuthService(repo, &fakeTokenRepo{}, &fakeOAuthRepo{}, &fakeVerifyRepo{}, m, &domain.OAuthProviderConfig{})
	return NewAuthHandler(svc)
}

func TestAuthHandler_Register(t *testing.T) {
	t.Parallel()

	repo := newFakeAuthRepo()
	h := testAuthHandler(t, repo)

	resp, err := h.Register(context.Background(), &authv1.RegisterRequest{
		Email: "dev@schemahub.dev", Password: "Password1", DisplayName: "Dev",
	})
	if err != nil {
		t.Fatalf("Register() error = %v", err)
	}
	if resp.User.Email != "dev@schemahub.dev" {
		t.Errorf("User.Email = %q, want dev@schemahub.dev", resp.User.Email)
	}
	if resp.AccessToken == "" || resp.RefreshToken == "" {
		t.Error("expected access and refresh tokens, got empty")
	}
	if resp.ExpiresIn != 900 {
		t.Errorf("ExpiresIn = %d, want 900", resp.ExpiresIn)
	}
}

func TestAuthHandler_RegisterDuplicateEmail(t *testing.T) {
	t.Parallel()

	repo := newFakeAuthRepo()
	repo.createErr = domain.ErrEmailAlreadyExists
	h := testAuthHandler(t, repo)

	_, err := h.Register(context.Background(), &authv1.RegisterRequest{
		Email: "dup@schemahub.dev", Password: "Password1", DisplayName: "Dup",
	})
	if status.Code(err) != codes.AlreadyExists {
		t.Fatalf("Register() error code = %v, want AlreadyExists (%v)", status.Code(err), err)
	}
}

func TestAuthHandler_LoginWrongPassword(t *testing.T) {
	t.Parallel()

	hash, err := bcrypt.GenerateFromPassword([]byte("Password1"), bcrypt.DefaultCost)
	if err != nil {
		t.Fatal(err)
	}
	repo := newFakeAuthRepo()
	repo.users["user_1"] = &domain.User{
		ID: "user_1", Email: "dev@schemahub.dev", PasswordHash: string(hash), IsActive: true,
	}
	repo.byEmail["dev@schemahub.dev"] = repo.users["user_1"]
	h := testAuthHandler(t, repo)

	_, err = h.Login(context.Background(), &authv1.LoginRequest{
		Email: "dev@schemahub.dev", Password: "WrongPass1",
	})
	if status.Code(err) != codes.Unauthenticated {
		t.Fatalf("Login() error code = %v, want Unauthenticated (%v)", status.Code(err), err)
	}
}

func TestAuthHandler_GetCurrentUser(t *testing.T) {
	t.Parallel()

	repo := newFakeAuthRepo()
	repo.users["user_1"] = &domain.User{
		ID: "user_1", Email: "dev@schemahub.dev", DisplayName: "Dev", IsActive: true,
	}
	h := testAuthHandler(t, repo)

	ctx := context.WithValue(context.Background(), interceptor.UserIDKey, "user_1")
	resp, err := h.GetCurrentUser(ctx, &authv1.GetCurrentUserRequest{})
	if err != nil {
		t.Fatalf("GetCurrentUser() error = %v", err)
	}
	if resp.User.Id != "user_1" {
		t.Errorf("User.Id = %q, want user_1", resp.User.Id)
	}
}

func TestAuthHandler_GetCurrentUserNotFound(t *testing.T) {
	t.Parallel()

	repo := newFakeAuthRepo()
	h := testAuthHandler(t, repo)

	ctx := context.WithValue(context.Background(), interceptor.UserIDKey, "ghost")
	_, err := h.GetCurrentUser(ctx, &authv1.GetCurrentUserRequest{})
	if status.Code(err) != codes.NotFound {
		t.Fatalf("GetCurrentUser() error code = %v, want NotFound (%v)", status.Code(err), err)
	}
}

func TestAuthHandler_ChangePassword(t *testing.T) {
	t.Parallel()

	hash, err := bcrypt.GenerateFromPassword([]byte("Password1"), bcrypt.DefaultCost)
	if err != nil {
		t.Fatal(err)
	}
	repo := newFakeAuthRepo()
	repo.users["user_1"] = &domain.User{
		ID: "user_1", Email: "dev@schemahub.dev", PasswordHash: string(hash), IsActive: true,
	}
	repo.byEmail["dev@schemahub.dev"] = repo.users["user_1"]
	h := testAuthHandler(t, repo)

	ctx := context.WithValue(context.Background(), interceptor.UserIDKey, "user_1")
	_, err = h.ChangePassword(ctx, &authv1.ChangePasswordRequest{
		CurrentPassword: "Password1", NewPassword: "NewPassword1",
	})
	if err != nil {
		t.Fatalf("ChangePassword() error = %v", err)
	}
}

func TestAuthHandler_ChangePasswordMismatch(t *testing.T) {
	t.Parallel()

	hash, err := bcrypt.GenerateFromPassword([]byte("Password1"), bcrypt.DefaultCost)
	if err != nil {
		t.Fatal(err)
	}
	repo := newFakeAuthRepo()
	repo.users["user_1"] = &domain.User{
		ID: "user_1", Email: "dev@schemahub.dev", PasswordHash: string(hash), IsActive: true,
	}
	repo.byEmail["dev@schemahub.dev"] = repo.users["user_1"]
	h := testAuthHandler(t, repo)

	ctx := context.WithValue(context.Background(), interceptor.UserIDKey, "user_1")
	_, err = h.ChangePassword(ctx, &authv1.ChangePasswordRequest{
		CurrentPassword: "WrongPass1", NewPassword: "NewPassword1",
	})
	if status.Code(err) != codes.InvalidArgument {
		t.Fatalf("ChangePassword() error code = %v, want InvalidArgument (%v)", status.Code(err), err)
	}
}

func TestAuthHandler_UpdateUser(t *testing.T) {
	t.Parallel()

	repo := newFakeAuthRepo()
	repo.users["user_1"] = &domain.User{
		ID: "user_1", Email: "dev@schemahub.dev", DisplayName: "Old", IsActive: true,
	}
	repo.byEmail["dev@schemahub.dev"] = repo.users["user_1"]
	h := testAuthHandler(t, repo)

	ctx := context.WithValue(context.Background(), interceptor.UserIDKey, "user_1")
	resp, err := h.UpdateUser(ctx, &authv1.UpdateUserRequest{
		DisplayName: "New Name", AvatarUrl: "https://example.com/a.png",
	})
	if err != nil {
		t.Fatalf("UpdateUser() error = %v", err)
	}
	if resp.User.DisplayName != "New Name" {
		t.Errorf("DisplayName = %q, want New Name", resp.User.DisplayName)
	}
	if resp.User.AvatarUrl != "https://example.com/a.png" {
		t.Errorf("AvatarUrl = %q, want https://example.com/a.png", resp.User.AvatarUrl)
	}
}
