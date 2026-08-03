package handler

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"errors"
	"fmt"
	"testing"
	"time"

	jwtv5 "github.com/golang-jwt/jwt/v5"
	"github.com/schemahub/backend/internal/auth/domain"
	"github.com/schemahub/backend/internal/pkg/interceptor"
	"github.com/schemahub/backend/internal/pkg/jwt"
	authv1 "github.com/schemahub/backend/proto/auth/v1"
	"golang.org/x/crypto/bcrypt"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type fakeAuthRepo struct {
	users      map[string]*domain.User
	byEmail    map[string]*domain.User
	createErr  error
	getErr     error
	deletedIDs []string
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
	if u, ok := f.users[id]; ok {
		u.PasswordHash = passwordHash
	}
	return nil
}

func (f *fakeAuthRepo) SoftDelete(ctx context.Context, id string) error {
	f.deletedIDs = append(f.deletedIDs, id)
	delete(f.users, id)
	return nil
}

type fakeTokenRepo struct {
	tokens        map[string]*domain.RefreshToken
	created       []*domain.RefreshToken
	revoked       []string
	revokedFamily []string
	getErr        error
}

func newFakeTokenRepo() *fakeTokenRepo {
	return &fakeTokenRepo{tokens: map[string]*domain.RefreshToken{}}
}

func (f *fakeTokenRepo) Create(ctx context.Context, token *domain.RefreshToken) error {
	f.created = append(f.created, token)
	return nil
}

func (f *fakeTokenRepo) GetByHash(ctx context.Context, hash string) (*domain.RefreshToken, error) {
	if f.getErr != nil {
		return nil, f.getErr
	}
	t, ok := f.tokens[hash]
	if !ok {
		return nil, domain.ErrInvalidRefreshToken
	}
	return t, nil
}

func (f *fakeTokenRepo) Revoke(ctx context.Context, id string) error {
	f.revoked = append(f.revoked, id)
	return nil
}

func (f *fakeTokenRepo) RevokeFamily(ctx context.Context, family string) error {
	f.revokedFamily = append(f.revokedFamily, family)
	return nil
}

func (f *fakeTokenRepo) GetActiveByUserID(ctx context.Context, userID string) ([]*domain.RefreshToken, error) {
	var out []*domain.RefreshToken
	for _, t := range f.tokens {
		if t.UserID == userID && t.RevokedAt == nil {
			out = append(out, t)
		}
	}
	return out, nil
}

type fakeVerifyRepo struct {
	tokens map[string]*domain.VerificationToken
}

func newFakeVerifyRepo() *fakeVerifyRepo {
	return &fakeVerifyRepo{tokens: map[string]*domain.VerificationToken{}}
}

func (f *fakeVerifyRepo) Create(ctx context.Context, token *domain.VerificationToken) error {
	f.tokens[token.TokenHash] = token
	return nil
}

func (f *fakeVerifyRepo) GetByHash(ctx context.Context, hash string) (*domain.VerificationToken, error) {
	t, ok := f.tokens[hash]
	if !ok {
		return nil, errors.New("not found")
	}
	return t, nil
}

func (f *fakeVerifyRepo) Consume(ctx context.Context, id string) error {
	for _, t := range f.tokens {
		if t.ID == id {
			now := time.Now()
			t.ConsumedAt = &now
		}
	}
	return nil
}

func (f *fakeVerifyRepo) GetActiveByUserID(ctx context.Context, userID string) ([]*domain.VerificationToken, error) {
	var out []*domain.VerificationToken
	for _, t := range f.tokens {
		if t.UserID == userID && t.ConsumedAt == nil {
			out = append(out, t)
		}
	}
	return out, nil
}

type fakeOAuthRepo struct {
	identities map[string]*domain.OAuthIdentity
	byProvider map[string]*domain.OAuthIdentity
	deleted    []string
	getErr     error
}

func newFakeOAuthRepo() *fakeOAuthRepo {
	return &fakeOAuthRepo{
		identities: map[string]*domain.OAuthIdentity{},
		byProvider: map[string]*domain.OAuthIdentity{},
	}
}

func (f *fakeOAuthRepo) Create(ctx context.Context, identity *domain.OAuthIdentity) error {
	f.identities[identity.ID] = identity
	f.byProvider[identity.Provider+":"+identity.ProviderUserID] = identity
	return nil
}

func (f *fakeOAuthRepo) GetByProvider(ctx context.Context, provider, providerUserID string) (*domain.OAuthIdentity, error) {
	if f.getErr != nil {
		return nil, f.getErr
	}
	id, ok := f.byProvider[provider+":"+providerUserID]
	if !ok {
		return nil, errors.New("not found")
	}
	return id, nil
}

func (f *fakeOAuthRepo) GetByUserID(ctx context.Context, userID string) ([]*domain.OAuthIdentity, error) {
	if f.getErr != nil {
		return nil, f.getErr
	}
	var out []*domain.OAuthIdentity
	for _, id := range f.identities {
		if id.UserID == userID {
			out = append(out, id)
		}
	}
	return out, nil
}

func (f *fakeOAuthRepo) Delete(ctx context.Context, id string) error {
	f.deleted = append(f.deleted, id)
	delete(f.identities, id)
	return nil
}

func (f *fakeOAuthRepo) UpdateLastUsed(ctx context.Context, id string) error { return nil }
func (f *fakeOAuthRepo) GetExpiringSoon(ctx context.Context, within time.Duration) ([]*domain.OAuthIdentity, error) {
	return nil, nil
}
func (f *fakeOAuthRepo) UpdateTokens(ctx context.Context, id, accessTokenEncrypted, refreshTokenEncrypted string, expiresAt *time.Time) error {
	return nil
}

// testOAuthStateClaims mirrors the JSON shape of the domain oauthStateClaims so
// tests can sign valid state tokens with the same keypair.
type testOAuthStateClaims struct {
	jwtv5.RegisteredClaims
	Provider      string `json:"provider"`
	RedirectTo    string `json:"redirect_to"`
	Linking       bool   `json:"linking"`
	CodeChallenge string `json:"code_challenge"`
}

func codeChallenge(verifier string) string {
	h := sha256.Sum256([]byte(verifier))
	return base64.RawURLEncoding.EncodeToString(h[:])
}

func newTestAuthHandler(t *testing.T, repo *fakeAuthRepo, tokens *fakeTokenRepo, oauth *fakeOAuthRepo, verify *fakeVerifyRepo, m *jwt.Manager, cfg *domain.OAuthProviderConfig) *AuthHandler {
	t.Helper()
	svc := domain.NewAuthService(repo, tokens, oauth, verify, m, cfg)
	return NewAuthHandler(svc)
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
	return newTestAuthHandler(t, repo, newFakeTokenRepo(), newFakeOAuthRepo(), newFakeVerifyRepo(), m, &domain.OAuthProviderConfig{})
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

func TestAuthHandler_UpdateUserNotFound(t *testing.T) {
	t.Parallel()

	h := testAuthHandler(t, newFakeAuthRepo())

	ctx := context.WithValue(context.Background(), interceptor.UserIDKey, "ghost")
	_, err := h.UpdateUser(ctx, &authv1.UpdateUserRequest{DisplayName: "Ghost"})
	if status.Code(err) != codes.NotFound {
		t.Fatalf("UpdateUser() error code = %v, want NotFound (%v)", status.Code(err), err)
	}
}

func TestAuthHandler_RegisterWeakPassword(t *testing.T) {
	t.Parallel()

	repo := newFakeAuthRepo()
	h := testAuthHandler(t, repo)

	_, err := h.Register(context.Background(), &authv1.RegisterRequest{
		Email: "dev@schemahub.dev", Password: "short", DisplayName: "Dev",
	})
	if status.Code(err) != codes.InvalidArgument {
		t.Fatalf("Register() error code = %v, want InvalidArgument (%v)", status.Code(err), err)
	}
}

func TestAuthHandler_RegisterInvalidEmail(t *testing.T) {
	t.Parallel()

	repo := newFakeAuthRepo()
	h := testAuthHandler(t, repo)

	_, err := h.Register(context.Background(), &authv1.RegisterRequest{
		Email: "not-an-email", Password: "Password1", DisplayName: "Dev",
	})
	if status.Code(err) != codes.Internal {
		t.Fatalf("Register() error code = %v, want Internal (%v)", status.Code(err), err)
	}
}

func TestAuthHandler_Login(t *testing.T) {
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

	resp, err := h.Login(context.Background(), &authv1.LoginRequest{
		Email: "dev@schemahub.dev", Password: "Password1",
	})
	if err != nil {
		t.Fatalf("Login() error = %v", err)
	}
	if resp.User.Email != "dev@schemahub.dev" {
		t.Errorf("User.Email = %q, want dev@schemahub.dev", resp.User.Email)
	}
	if resp.AccessToken == "" || resp.RefreshToken == "" {
		t.Error("expected access and refresh tokens, got empty")
	}
}

func TestAuthHandler_LoginUnknownEmail(t *testing.T) {
	t.Parallel()

	h := testAuthHandler(t, newFakeAuthRepo())

	_, err := h.Login(context.Background(), &authv1.LoginRequest{
		Email: "ghost@schemahub.dev", Password: "Password1",
	})
	if status.Code(err) != codes.Unauthenticated {
		t.Fatalf("Login() error code = %v, want Unauthenticated (%v)", status.Code(err), err)
	}
}

func TestAuthHandler_RefreshToken(t *testing.T) {
	t.Parallel()

	repo := newFakeAuthRepo()
	repo.users["user_1"] = &domain.User{ID: "user_1", Email: "dev@schemahub.dev", IsActive: true}
	repo.byEmail["dev@schemahub.dev"] = repo.users["user_1"]
	tokens := newFakeTokenRepo()
	tokens.tokens["rt_abc"] = &domain.RefreshToken{
		ID: "rt_1", UserID: "user_1", Family: "fam_1", ExpiresAt: time.Now().Add(time.Hour),
	}
	h := newTestAuthHandler(t, repo, tokens, newFakeOAuthRepo(), newFakeVerifyRepo(), testManager(t), &domain.OAuthProviderConfig{})

	resp, err := h.RefreshToken(context.Background(), &authv1.RefreshTokenRequest{RefreshToken: "rt_abc"})
	if err != nil {
		t.Fatalf("RefreshToken() error = %v", err)
	}
	if resp.AccessToken == "" || resp.RefreshToken == "" {
		t.Error("expected access and refresh tokens, got empty")
	}
	if len(tokens.revoked) != 1 || tokens.revoked[0] != "rt_1" {
		t.Errorf("revoked = %v, want [rt_1]", tokens.revoked)
	}
	if len(tokens.created) != 1 {
		t.Errorf("created tokens = %d, want 1", len(tokens.created))
	}
}

func TestAuthHandler_RefreshTokenInvalid(t *testing.T) {
	t.Parallel()

	repo := newFakeAuthRepo()
	h := newTestAuthHandler(t, repo, newFakeTokenRepo(), newFakeOAuthRepo(), newFakeVerifyRepo(), testManager(t), &domain.OAuthProviderConfig{})

	_, err := h.RefreshToken(context.Background(), &authv1.RefreshTokenRequest{RefreshToken: "nope"})
	if status.Code(err) != codes.Unauthenticated {
		t.Fatalf("RefreshToken() error code = %v, want Unauthenticated (%v)", status.Code(err), err)
	}
}

func TestAuthHandler_Logout(t *testing.T) {
	t.Parallel()

	tokens := newFakeTokenRepo()
	tokens.tokens["rt_abc"] = &domain.RefreshToken{
		ID: "rt_1", UserID: "user_1", ExpiresAt: time.Now().Add(time.Hour),
	}
	h := newTestAuthHandler(t, newFakeAuthRepo(), tokens, newFakeOAuthRepo(), newFakeVerifyRepo(), testManager(t), &domain.OAuthProviderConfig{})

	_, err := h.Logout(context.Background(), &authv1.LogoutRequest{RefreshToken: "rt_abc"})
	if err != nil {
		t.Fatalf("Logout() error = %v", err)
	}
	if len(tokens.revoked) != 1 || tokens.revoked[0] != "rt_1" {
		t.Errorf("revoked = %v, want [rt_1]", tokens.revoked)
	}
}

func TestAuthHandler_LogoutUnknownToken(t *testing.T) {
	t.Parallel()

	h := newTestAuthHandler(t, newFakeAuthRepo(), newFakeTokenRepo(), newFakeOAuthRepo(), newFakeVerifyRepo(), testManager(t), &domain.OAuthProviderConfig{})

	if _, err := h.Logout(context.Background(), &authv1.LogoutRequest{RefreshToken: "nope"}); err != nil {
		t.Fatalf("Logout() error = %v, want nil for unknown token", err)
	}
}

func TestAuthHandler_VerifyEmail(t *testing.T) {
	t.Parallel()

	repo := newFakeAuthRepo()
	repo.users["user_1"] = &domain.User{ID: "user_1", Email: "dev@schemahub.dev", IsActive: true}
	repo.byEmail["dev@schemahub.dev"] = repo.users["user_1"]
	rawToken := "verify_token"
	hash := fmt.Sprintf("%x", sha256.Sum256([]byte(rawToken)))
	verify := newFakeVerifyRepo()
	verify.tokens[hash] = &domain.VerificationToken{
		ID: "vt_1", UserID: "user_1", Email: "dev@schemahub.dev", ExpiresAt: time.Now().Add(time.Hour),
	}
	h := newTestAuthHandler(t, repo, newFakeTokenRepo(), newFakeOAuthRepo(), verify, testManager(t), &domain.OAuthProviderConfig{})

	if _, err := h.VerifyEmail(context.Background(), &authv1.VerifyEmailRequest{Token: rawToken}); err != nil {
		t.Fatalf("VerifyEmail() error = %v", err)
	}
	if repo.users["user_1"].EmailVerifiedAt == nil {
		t.Error("expected EmailVerifiedAt to be set")
	}
}

func TestAuthHandler_VerifyEmailInvalidToken(t *testing.T) {
	t.Parallel()

	h := newTestAuthHandler(t, newFakeAuthRepo(), newFakeTokenRepo(), newFakeOAuthRepo(), newFakeVerifyRepo(), testManager(t), &domain.OAuthProviderConfig{})

	_, err := h.VerifyEmail(context.Background(), &authv1.VerifyEmailRequest{Token: "bogus"})
	if status.Code(err) != codes.Unauthenticated {
		t.Fatalf("VerifyEmail() error code = %v, want Unauthenticated (%v)", status.Code(err), err)
	}
}

func TestAuthHandler_SendVerificationEmail(t *testing.T) {
	t.Parallel()

	repo := newFakeAuthRepo()
	repo.users["user_1"] = &domain.User{ID: "user_1", Email: "dev@schemahub.dev", IsActive: true}
	repo.byEmail["dev@schemahub.dev"] = repo.users["user_1"]
	verify := newFakeVerifyRepo()
	h := newTestAuthHandler(t, repo, newFakeTokenRepo(), newFakeOAuthRepo(), verify, testManager(t), &domain.OAuthProviderConfig{})

	if _, err := h.SendVerificationEmail(context.Background(), &authv1.SendVerificationEmailRequest{Email: "dev@schemahub.dev"}); err != nil {
		t.Fatalf("SendVerificationEmail() error = %v", err)
	}
	if len(verify.tokens) != 1 {
		t.Errorf("verification tokens = %d, want 1", len(verify.tokens))
	}
}

func TestAuthHandler_SendVerificationEmailUnknownUser(t *testing.T) {
	t.Parallel()

	h := newTestAuthHandler(t, newFakeAuthRepo(), newFakeTokenRepo(), newFakeOAuthRepo(), newFakeVerifyRepo(), testManager(t), &domain.OAuthProviderConfig{})

	_, err := h.SendVerificationEmail(context.Background(), &authv1.SendVerificationEmailRequest{Email: "ghost@schemahub.dev"})
	if status.Code(err) != codes.NotFound {
		t.Fatalf("SendVerificationEmail() error code = %v, want NotFound (%v)", status.Code(err), err)
	}
}

func TestAuthHandler_ForgotPassword(t *testing.T) {
	t.Parallel()

	repo := newFakeAuthRepo()
	repo.users["user_1"] = &domain.User{ID: "user_1", Email: "dev@schemahub.dev", IsActive: true}
	repo.byEmail["dev@schemahub.dev"] = repo.users["user_1"]
	verify := newFakeVerifyRepo()
	h := newTestAuthHandler(t, repo, newFakeTokenRepo(), newFakeOAuthRepo(), verify, testManager(t), &domain.OAuthProviderConfig{})

	if _, err := h.ForgotPassword(context.Background(), &authv1.ForgotPasswordRequest{Email: "dev@schemahub.dev"}); err != nil {
		t.Fatalf("ForgotPassword() error = %v", err)
	}
	if len(verify.tokens) != 1 {
		t.Errorf("reset tokens = %d, want 1", len(verify.tokens))
	}
}

func TestAuthHandler_ForgotPasswordUnknownEmail(t *testing.T) {
	t.Parallel()

	h := newTestAuthHandler(t, newFakeAuthRepo(), newFakeTokenRepo(), newFakeOAuthRepo(), newFakeVerifyRepo(), testManager(t), &domain.OAuthProviderConfig{})

	if _, err := h.ForgotPassword(context.Background(), &authv1.ForgotPasswordRequest{Email: "ghost@schemahub.dev"}); err != nil {
		t.Fatalf("ForgotPassword() error = %v, want nil for unknown user", err)
	}
}

func TestAuthHandler_ResetPassword(t *testing.T) {
	t.Parallel()

	repo := newFakeAuthRepo()
	repo.users["user_1"] = &domain.User{ID: "user_1", Email: "dev@schemahub.dev", IsActive: true}
	repo.byEmail["dev@schemahub.dev"] = repo.users["user_1"]
	rawToken := "reset_token"
	hash := fmt.Sprintf("%x", sha256.Sum256([]byte(rawToken)))
	verify := newFakeVerifyRepo()
	verify.tokens[hash] = &domain.VerificationToken{
		ID: "vt_1", UserID: "user_1", Email: "dev@schemahub.dev", ExpiresAt: time.Now().Add(time.Hour),
	}
	h := newTestAuthHandler(t, repo, newFakeTokenRepo(), newFakeOAuthRepo(), verify, testManager(t), &domain.OAuthProviderConfig{})

	if _, err := h.ResetPassword(context.Background(), &authv1.ResetPasswordRequest{Token: rawToken, Password: "NewPassword1"}); err != nil {
		t.Fatalf("ResetPassword() error = %v", err)
	}
	if err := bcrypt.CompareHashAndPassword([]byte(repo.users["user_1"].PasswordHash), []byte("NewPassword1")); err != nil {
		t.Error("expected password hash to be updated to NewPassword1")
	}
}

func TestAuthHandler_ResetPasswordInvalidToken(t *testing.T) {
	t.Parallel()

	h := newTestAuthHandler(t, newFakeAuthRepo(), newFakeTokenRepo(), newFakeOAuthRepo(), newFakeVerifyRepo(), testManager(t), &domain.OAuthProviderConfig{})

	_, err := h.ResetPassword(context.Background(), &authv1.ResetPasswordRequest{Token: "bogus", Password: "NewPassword1"})
	if status.Code(err) != codes.Unauthenticated {
		t.Fatalf("ResetPassword() error code = %v, want Unauthenticated (%v)", status.Code(err), err)
	}
}

func TestAuthHandler_ResetPasswordWeakPassword(t *testing.T) {
	t.Parallel()

	repo := newFakeAuthRepo()
	repo.users["user_1"] = &domain.User{ID: "user_1", Email: "dev@schemahub.dev", IsActive: true}
	repo.byEmail["dev@schemahub.dev"] = repo.users["user_1"]
	rawToken := "reset_token"
	hash := fmt.Sprintf("%x", sha256.Sum256([]byte(rawToken)))
	verify := newFakeVerifyRepo()
	verify.tokens[hash] = &domain.VerificationToken{
		ID: "vt_1", UserID: "user_1", Email: "dev@schemahub.dev", ExpiresAt: time.Now().Add(time.Hour),
	}
	h := newTestAuthHandler(t, repo, newFakeTokenRepo(), newFakeOAuthRepo(), verify, testManager(t), &domain.OAuthProviderConfig{})

	_, err := h.ResetPassword(context.Background(), &authv1.ResetPasswordRequest{Token: rawToken, Password: "weak"})
	if status.Code(err) != codes.InvalidArgument {
		t.Fatalf("ResetPassword() error code = %v, want InvalidArgument (%v)", status.Code(err), err)
	}
}

func TestAuthHandler_DeleteAccount(t *testing.T) {
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
	tokens := newFakeTokenRepo()
	tokens.tokens["rt_abc"] = &domain.RefreshToken{ID: "rt_1", UserID: "user_1", ExpiresAt: time.Now().Add(time.Hour)}
	h := newTestAuthHandler(t, repo, tokens, newFakeOAuthRepo(), newFakeVerifyRepo(), testManager(t), &domain.OAuthProviderConfig{})

	ctx := context.WithValue(context.Background(), interceptor.UserIDKey, "user_1")
	if _, err := h.DeleteAccount(ctx, &authv1.DeleteAccountRequest{Password: "Password1"}); err != nil {
		t.Fatalf("DeleteAccount() error = %v", err)
	}
	if len(repo.deletedIDs) != 1 || repo.deletedIDs[0] != "user_1" {
		t.Errorf("deleted = %v, want [user_1]", repo.deletedIDs)
	}
	if len(tokens.revoked) != 1 || tokens.revoked[0] != "rt_1" {
		t.Errorf("revoked = %v, want [rt_1]", tokens.revoked)
	}
}

func TestAuthHandler_DeleteAccountWrongPassword(t *testing.T) {
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
	h := newTestAuthHandler(t, repo, newFakeTokenRepo(), newFakeOAuthRepo(), newFakeVerifyRepo(), testManager(t), &domain.OAuthProviderConfig{})

	ctx := context.WithValue(context.Background(), interceptor.UserIDKey, "user_1")
	_, err = h.DeleteAccount(ctx, &authv1.DeleteAccountRequest{Password: "WrongPass1"})
	if status.Code(err) != codes.InvalidArgument {
		t.Fatalf("DeleteAccount() error code = %v, want InvalidArgument (%v)", status.Code(err), err)
	}
}

func TestAuthHandler_GetOAuthURL(t *testing.T) {
	t.Parallel()

	h := newTestAuthHandler(t, newFakeAuthRepo(), newFakeTokenRepo(), newFakeOAuthRepo(), newFakeVerifyRepo(), testManager(t), &domain.OAuthProviderConfig{})

	for _, provider := range []string{"google", "github", "slack"} {
		resp, err := h.GetOAuthURL(context.Background(), &authv1.GetOAuthURLRequest{Provider: provider, RedirectTo: "https://app.schemahub.dev/callback"})
		if err != nil {
			t.Fatalf("GetOAuthURL(%s) error = %v", provider, err)
		}
		if resp.AuthUrl == "" {
			t.Errorf("GetOAuthURL(%s) AuthUrl = empty", provider)
		}
		if resp.StateToken == "" {
			t.Errorf("GetOAuthURL(%s) StateToken = empty", provider)
		}
	}
}

func TestAuthHandler_GetOAuthURLUnsupportedProvider(t *testing.T) {
	t.Parallel()

	h := newTestAuthHandler(t, newFakeAuthRepo(), newFakeTokenRepo(), newFakeOAuthRepo(), newFakeVerifyRepo(), testManager(t), &domain.OAuthProviderConfig{})

	_, err := h.GetOAuthURL(context.Background(), &authv1.GetOAuthURLRequest{Provider: "myspace"})
	if status.Code(err) != codes.Internal {
		t.Fatalf("GetOAuthURL() error code = %v, want Internal (%v)", status.Code(err), err)
	}
}

func TestAuthHandler_HandleOAuthCallbackUnsupportedProvider(t *testing.T) {
	t.Parallel()

	h := newTestAuthHandler(t, newFakeAuthRepo(), newFakeTokenRepo(), newFakeOAuthRepo(), newFakeVerifyRepo(), testManager(t), &domain.OAuthProviderConfig{})

	_, err := h.HandleOAuthCallback(context.Background(), &authv1.HandleOAuthCallbackRequest{
		Provider: "myspace", Code: "code", State: "state",
	})
	if status.Code(err) != codes.Internal {
		t.Fatalf("HandleOAuthCallback() error code = %v, want Internal (%v)", status.Code(err), err)
	}
}

func TestAuthHandler_HandleOAuthCallbackStateMismatch(t *testing.T) {
	t.Parallel()

	m := testManager(t)
	state, err := m.SignClaims(&testOAuthStateClaims{
		RegisteredClaims: jwtv5.RegisteredClaims{
			Subject: "state_1", ExpiresAt: jwtv5.NewNumericDate(time.Now().Add(time.Hour)),
		},
		Provider:      "google",
		CodeChallenge: codeChallenge("verifier_a"),
	})
	if err != nil {
		t.Fatal(err)
	}
	h := newTestAuthHandler(t, newFakeAuthRepo(), newFakeTokenRepo(), newFakeOAuthRepo(), newFakeVerifyRepo(), m, &domain.OAuthProviderConfig{})

	_, err = h.HandleOAuthCallback(context.Background(), &authv1.HandleOAuthCallbackRequest{
		Provider: "google", Code: "code", State: state, CodeVerifier: "verifier_b",
	})
	if status.Code(err) != codes.Internal {
		t.Fatalf("HandleOAuthCallback() error code = %v, want Internal (%v)", status.Code(err), err)
	}
}

func TestAuthHandler_LinkOAuthIdentityStateNotLinking(t *testing.T) {
	t.Parallel()

	m := testManager(t)
	state, err := m.SignClaims(&testOAuthStateClaims{
		RegisteredClaims: jwtv5.RegisteredClaims{
			Subject: "state_1", ExpiresAt: jwtv5.NewNumericDate(time.Now().Add(time.Hour)),
		},
		Provider:      "github",
		Linking:       false,
		CodeChallenge: codeChallenge("verifier_a"),
	})
	if err != nil {
		t.Fatal(err)
	}
	h := newTestAuthHandler(t, newFakeAuthRepo(), newFakeTokenRepo(), newFakeOAuthRepo(), newFakeVerifyRepo(), m, &domain.OAuthProviderConfig{})

	ctx := context.WithValue(context.Background(), interceptor.UserIDKey, "user_1")
	_, err = h.LinkOAuthIdentity(ctx, &authv1.LinkOAuthIdentityRequest{
		Provider: "github", Code: "code", State: state,
	})
	if status.Code(err) != codes.Internal {
		t.Fatalf("LinkOAuthIdentity() error code = %v, want Internal (%v)", status.Code(err), err)
	}
}

func TestAuthHandler_UnlinkOAuthIdentity(t *testing.T) {
	t.Parallel()

	repo := newFakeAuthRepo()
	repo.users["user_1"] = &domain.User{ID: "user_1", Email: "dev@schemahub.dev", PasswordHash: "hash", IsActive: true}
	repo.byEmail["dev@schemahub.dev"] = repo.users["user_1"]
	oauth := newFakeOAuthRepo()
	now := time.Now()
	oauth.identities["id1"] = &domain.OAuthIdentity{ID: "id1", UserID: "user_1", Provider: "google", CreatedAt: now}
	oauth.identities["id2"] = &domain.OAuthIdentity{ID: "id2", UserID: "user_1", Provider: "github", CreatedAt: now}
	h := newTestAuthHandler(t, repo, newFakeTokenRepo(), oauth, newFakeVerifyRepo(), testManager(t), &domain.OAuthProviderConfig{})

	ctx := context.WithValue(context.Background(), interceptor.UserIDKey, "user_1")
	if _, err := h.UnlinkOAuthIdentity(ctx, &authv1.UnlinkOAuthIdentityRequest{Provider: "google"}); err != nil {
		t.Fatalf("UnlinkOAuthIdentity() error = %v", err)
	}
	if len(oauth.deleted) != 1 || oauth.deleted[0] != "id1" {
		t.Errorf("deleted = %v, want [id1]", oauth.deleted)
	}
}

func TestAuthHandler_UnlinkOAuthIdentityNotLinked(t *testing.T) {
	t.Parallel()

	repo := newFakeAuthRepo()
	repo.users["user_1"] = &domain.User{ID: "user_1", Email: "dev@schemahub.dev", PasswordHash: "hash", IsActive: true}
	repo.byEmail["dev@schemahub.dev"] = repo.users["user_1"]
	oauth := newFakeOAuthRepo()
	oauth.identities["id2"] = &domain.OAuthIdentity{ID: "id2", UserID: "user_1", Provider: "github", CreatedAt: time.Now()}
	h := newTestAuthHandler(t, repo, newFakeTokenRepo(), oauth, newFakeVerifyRepo(), testManager(t), &domain.OAuthProviderConfig{})

	ctx := context.WithValue(context.Background(), interceptor.UserIDKey, "user_1")
	_, err := h.UnlinkOAuthIdentity(ctx, &authv1.UnlinkOAuthIdentityRequest{Provider: "google"})
	if status.Code(err) != codes.Internal {
		t.Fatalf("UnlinkOAuthIdentity() error code = %v, want Internal (%v)", status.Code(err), err)
	}
}

func TestAuthHandler_UnlinkOAuthIdentityLastAuthMethod(t *testing.T) {
	t.Parallel()

	repo := newFakeAuthRepo()
	repo.users["user_1"] = &domain.User{ID: "user_1", Email: "dev@schemahub.dev", IsActive: true}
	repo.byEmail["dev@schemahub.dev"] = repo.users["user_1"]
	oauth := newFakeOAuthRepo()
	oauth.identities["id1"] = &domain.OAuthIdentity{ID: "id1", UserID: "user_1", Provider: "google", CreatedAt: time.Now()}
	h := newTestAuthHandler(t, repo, newFakeTokenRepo(), oauth, newFakeVerifyRepo(), testManager(t), &domain.OAuthProviderConfig{})

	ctx := context.WithValue(context.Background(), interceptor.UserIDKey, "user_1")
	_, err := h.UnlinkOAuthIdentity(ctx, &authv1.UnlinkOAuthIdentityRequest{Provider: "google"})
	if status.Code(err) != codes.Internal {
		t.Fatalf("UnlinkOAuthIdentity() error code = %v, want Internal (%v)", status.Code(err), err)
	}
}

func TestAuthHandler_ListLinkedIdentities(t *testing.T) {
	t.Parallel()

	repo := newFakeAuthRepo()
	repo.users["user_1"] = &domain.User{ID: "user_1", Email: "dev@schemahub.dev", IsActive: true}
	repo.byEmail["dev@schemahub.dev"] = repo.users["user_1"]
	oauth := newFakeOAuthRepo()
	now := time.Now()
	oauth.identities["id1"] = &domain.OAuthIdentity{ID: "id1", UserID: "user_1", Provider: "google", ProviderEmail: "dev@gmail.com", CreatedAt: now}
	oauth.identities["id2"] = &domain.OAuthIdentity{ID: "id2", UserID: "user_1", Provider: "github", ProviderEmail: "dev@github.com", CreatedAt: now}
	h := newTestAuthHandler(t, repo, newFakeTokenRepo(), oauth, newFakeVerifyRepo(), testManager(t), &domain.OAuthProviderConfig{})

	ctx := context.WithValue(context.Background(), interceptor.UserIDKey, "user_1")
	resp, err := h.ListLinkedIdentities(ctx, &authv1.ListLinkedIdentitiesRequest{})
	if err != nil {
		t.Fatalf("ListLinkedIdentities() error = %v", err)
	}
	if len(resp.Identities) != 2 {
		t.Errorf("Identities len = %d, want 2", len(resp.Identities))
	}
	if resp.Identities[0].Provider != "google" && resp.Identities[1].Provider != "google" {
		t.Error("expected a google identity in the response")
	}
}

func testManager(t *testing.T) *jwt.Manager {
	t.Helper()
	priv, pub := testKeyPair(t)
	m, err := jwt.NewManager(priv, pub)
	if err != nil {
		t.Fatalf("NewManager() error = %v", err)
	}
	return m
}
