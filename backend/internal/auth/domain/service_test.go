package domain

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/schemahub/backend/internal/pkg/jwt"
)

type fakeUserRepo struct {
	users map[string]*User
}

func (f *fakeUserRepo) Create(ctx context.Context, u *User) error {
	f.users[u.Email] = u
	return nil
}
func (f *fakeUserRepo) GetByID(ctx context.Context, id string) (*User, error) {
	for _, u := range f.users {
		if u.ID == id {
			return u, nil
		}
	}
	return nil, ErrUserNotFound
}
func (f *fakeUserRepo) GetByEmail(ctx context.Context, email string) (*User, error) {
	u, ok := f.users[email]
	if !ok {
		return nil, ErrUserNotFound
	}
	return u, nil
}
func (f *fakeUserRepo) Update(ctx context.Context, u *User) error { f.users[u.Email] = u; return nil }
func (f *fakeUserRepo) UpdatePassword(ctx context.Context, id, hash string) error {
	for _, u := range f.users {
		if u.ID == id {
			u.PasswordHash = hash
		}
	}
	return nil
}
func (f *fakeUserRepo) SoftDelete(ctx context.Context, id string) error { return nil }

type fakeTokenRepo struct{}

func (f *fakeTokenRepo) Create(ctx context.Context, t *RefreshToken) error { return nil }
func (f *fakeTokenRepo) GetByHash(ctx context.Context, h string) (*RefreshToken, error) {
	return nil, errors.New("not found")
}
func (f *fakeTokenRepo) Revoke(ctx context.Context, id string) error           { return nil }
func (f *fakeTokenRepo) RevokeFamily(ctx context.Context, family string) error { return nil }
func (f *fakeTokenRepo) GetActiveByUserID(ctx context.Context, id string) ([]*RefreshToken, error) {
	return nil, nil
}

type fakeVerifyRepo struct {
	tokens []*VerificationToken
}

func (f *fakeVerifyRepo) Create(ctx context.Context, t *VerificationToken) error {
	f.tokens = append(f.tokens, t)
	return nil
}
func (f *fakeVerifyRepo) GetByHash(ctx context.Context, h string) (*VerificationToken, error) {
	for _, t := range f.tokens {
		if t.TokenHash == h {
			return t, nil
		}
	}
	return nil, errors.New("not found")
}
func (f *fakeVerifyRepo) Consume(ctx context.Context, id string) error { return nil }
func (f *fakeVerifyRepo) GetActiveByUserID(ctx context.Context, id string) ([]*VerificationToken, error) {
	return nil, nil
}

type fakeOAuthRepo struct{}

func (f *fakeOAuthRepo) Create(ctx context.Context, i *OAuthIdentity) error { return nil }
func (f *fakeOAuthRepo) GetByProvider(ctx context.Context, p, uid string) (*OAuthIdentity, error) {
	return nil, errors.New("not found")
}
func (f *fakeOAuthRepo) GetByUserID(ctx context.Context, id string) ([]*OAuthIdentity, error) {
	return nil, nil
}
func (f *fakeOAuthRepo) Delete(ctx context.Context, id string) error         { return nil }
func (f *fakeOAuthRepo) UpdateLastUsed(ctx context.Context, id string) error { return nil }
func (f *fakeOAuthRepo) GetExpiringSoon(ctx context.Context, d time.Duration) ([]*OAuthIdentity, error) {
	return nil, nil
}
func (f *fakeOAuthRepo) UpdateTokens(ctx context.Context, id, a, r string, exp *time.Time) error {
	return nil
}

type fakeMailer struct {
	verifySent []string
	resetSent  []string
	err        error
}

func (f *fakeMailer) SendVerificationEmail(ctx context.Context, to, token string) error {
	if f.err != nil {
		return f.err
	}
	f.verifySent = append(f.verifySent, to+"|"+token)
	return nil
}
func (f *fakeMailer) SendPasswordResetEmail(ctx context.Context, to, token string) error {
	if f.err != nil {
		return f.err
	}
	f.resetSent = append(f.resetSent, to+"|"+token)
	return nil
}

func newTestAuthService(t *testing.T, verify *fakeVerifyRepo) (*AuthService, *fakeUserRepo, *fakeVerifyRepo, *fakeMailer) {
	t.Helper()
	users := &fakeUserRepo{users: map[string]*User{}}
	users.users["dev@example.com"] = &User{
		ID: "user_1", Email: "dev@example.com", DisplayName: "Dev", Role: "user", IsActive: true,
	}
	m := &fakeMailer{}
	s := NewAuthService(users, &fakeTokenRepo{}, &fakeOAuthRepo{}, verify, &jwt.Manager{}, nil)
	s.SetMailer(m)
	return s, users, verify, m
}

func TestSendVerificationEmailSendsToken(t *testing.T) {
	t.Parallel()
	s, _, verify, m := newTestAuthService(t, &fakeVerifyRepo{})

	err := s.SendVerificationEmail(context.Background(), "dev@example.com")
	if err != nil {
		t.Fatalf("SendVerificationEmail: %v", err)
	}
	if len(verify.tokens) != 1 {
		t.Fatalf("tokens stored = %d, want 1", len(verify.tokens))
	}
	if len(m.verifySent) != 1 {
		t.Fatalf("emails sent = %d, want 1", len(m.verifySent))
	}
	parts := strings.Split(m.verifySent[0], "|")
	if parts[0] != "dev@example.com" {
		t.Errorf("to = %q", parts[0])
	}
	if !strings.HasPrefix(parts[1], "verify_") {
		t.Errorf("token = %q, want verify_ prefix", parts[1])
	}
	if verify.tokens[0].TokenHash == parts[1] {
		t.Error("raw token must not be stored as-is")
	}
}

func TestSendVerificationEmailSkipsWhenVerified(t *testing.T) {
	t.Parallel()
	s, _, verify, m := newTestAuthService(t, &fakeVerifyRepo{})
	now := time.Now()
	s.userRepo.(*fakeUserRepo).users["dev@example.com"].EmailVerifiedAt = &now

	if err := s.SendVerificationEmail(context.Background(), "dev@example.com"); err != nil {
		t.Fatalf("SendVerificationEmail: %v", err)
	}
	if len(verify.tokens) != 0 {
		t.Error("token stored for already-verified user")
	}
	if len(m.verifySent) != 0 {
		t.Error("email sent to already-verified user")
	}
}

func TestSendVerificationEmailUnknownUser(t *testing.T) {
	t.Parallel()
	s, _, _, m := newTestAuthService(t, &fakeVerifyRepo{})

	err := s.SendVerificationEmail(context.Background(), "ghost@example.com")
	if !errors.Is(err, ErrUserNotFound) {
		t.Errorf("err = %v, want ErrUserNotFound", err)
	}
	if len(m.verifySent) != 0 {
		t.Error("email sent for unknown user")
	}
}

func TestForgotPasswordSendsResetToken(t *testing.T) {
	t.Parallel()
	s, _, verify, m := newTestAuthService(t, &fakeVerifyRepo{})

	if err := s.ForgotPassword(context.Background(), "dev@example.com"); err != nil {
		t.Fatalf("ForgotPassword: %v", err)
	}
	if len(verify.tokens) != 1 {
		t.Fatalf("tokens stored = %d, want 1", len(verify.tokens))
	}
	if len(m.resetSent) != 1 {
		t.Fatalf("reset emails sent = %d, want 1", len(m.resetSent))
	}
	if !strings.HasPrefix(strings.Split(m.resetSent[0], "|")[1], "reset_") {
		t.Errorf("token = %q, want reset_ prefix", strings.Split(m.resetSent[0], "|")[1])
	}
}

func TestForgotPasswordHidesUnknownUser(t *testing.T) {
	t.Parallel()
	s, _, _, m := newTestAuthService(t, &fakeVerifyRepo{})

	if err := s.ForgotPassword(context.Background(), "ghost@example.com"); err != nil {
		t.Fatalf("ForgotPassword must return nil for unknown user, got %v", err)
	}
	if len(m.resetSent) != 0 {
		t.Error("reset email sent for unknown user")
	}
}

func TestForgotPasswordMailFailureReturnsError(t *testing.T) {
	t.Parallel()
	s, _, _, m := newTestAuthService(t, &fakeVerifyRepo{})
	m.err = errors.New("smtp down")

	err := s.ForgotPassword(context.Background(), "dev@example.com")
	if err == nil || !strings.Contains(err.Error(), "sending reset email") {
		t.Errorf("err = %v, want wrapped send failure", err)
	}
}

func TestEmailFlowsWithoutMailer(t *testing.T) {
	t.Parallel()
	verify := &fakeVerifyRepo{}
	s, _, _, _ := newTestAuthService(t, verify)
	s.mailer = nil

	if err := s.SendVerificationEmail(context.Background(), "dev@example.com"); err != nil {
		t.Fatalf("SendVerificationEmail without mailer: %v", err)
	}
	if len(verify.tokens) != 1 {
		t.Error("token not stored when mailer is nil")
	}
}
