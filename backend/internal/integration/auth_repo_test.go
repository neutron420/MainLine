package integration

import (
	"context"
	"crypto/sha256"
	"errors"
	"fmt"
	"testing"
	"time"

	authdomain "github.com/schemahub/backend/internal/auth/domain"
	authpg "github.com/schemahub/backend/internal/auth/repository/postgres"
)

func hashHex(s string) string {
	h := sha256.Sum256([]byte(s))
	return fmt.Sprintf("%x", h)
}

func TestUserRepository_RoundTrip(t *testing.T) {
	pool := setup(t)
	ctx := context.Background()
	repo := authpg.NewUserRepository(pool)

	u := &authdomain.User{
		Email:        fmt.Sprintf("user-%s@example.com", newUUID(t)),
		PasswordHash: "hash-a",
		DisplayName:  "Alice",
		AvatarURL:    "https://example.com/a.png",
		Role:         testUserRole,
		IsActive:     true,
	}

	requireNoErr(t, repo.Create(ctx, u), "Create")

	got, err := repo.GetByEmail(ctx, u.Email)
	requireNoErr(t, err, "GetByEmail")
	if got.Email != u.Email || got.PasswordHash != "hash-a" || got.DisplayName != "Alice" || !got.IsActive {
		t.Fatalf("GetByEmail returned unexpected user: %+v", got)
	}

	byID, err := repo.GetByID(ctx, got.ID)
	requireNoErr(t, err, "GetByID")
	if byID.ID != got.ID {
		t.Fatalf("GetByID id = %s, want %s", byID.ID, got.ID)
	}

	got.DisplayName = "Alice Smith"
	got.AvatarURL = "https://example.com/b.png"
	requireNoErr(t, repo.Update(ctx, got), "Update")
	updated, _ := repo.GetByID(ctx, got.ID)
	if updated.DisplayName != "Alice Smith" || updated.AvatarURL != "https://example.com/b.png" {
		t.Fatalf("Update did not persist: %+v", updated)
	}

	requireNoErr(t, repo.UpdatePassword(ctx, got.ID, "hash-b"), "UpdatePassword")
	withNewPw, _ := repo.GetByID(ctx, got.ID)
	if withNewPw.PasswordHash != "hash-b" {
		t.Fatalf("UpdatePassword did not persist: %+v", withNewPw)
	}

	requireNoErr(t, repo.SoftDelete(ctx, got.ID), "SoftDelete")
	if _, err := repo.GetByID(ctx, got.ID); !errors.Is(err, authdomain.ErrUserNotFound) {
		t.Fatalf("GetByID after SoftDelete err = %v, want ErrUserNotFound", err)
	}
	if _, err := repo.GetByEmail(ctx, got.Email); !errors.Is(err, authdomain.ErrUserNotFound) {
		t.Fatalf("GetByEmail after SoftDelete err = %v, want ErrUserNotFound", err)
	}
}

func TestUserRepository_DuplicateEmail(t *testing.T) {
	pool := setup(t)
	ctx := context.Background()
	repo := authpg.NewUserRepository(pool)

	email := fmt.Sprintf("dup-%s@example.com", newUUID(t))
	first := &authdomain.User{Email: email, PasswordHash: "a", DisplayName: "A", Role: testUserRole, IsActive: true}
	second := &authdomain.User{Email: email, PasswordHash: "b", DisplayName: "B", Role: testUserRole, IsActive: true}

	requireNoErr(t, repo.Create(ctx, first), "Create first")
	if err := repo.Create(ctx, second); !errors.Is(err, authdomain.ErrEmailAlreadyExists) {
		t.Fatalf("Create duplicate err = %v, want ErrEmailAlreadyExists", err)
	}
}

func TestRefreshTokenRepository_RoundTrip(t *testing.T) {
	pool := setup(t)
	ctx := context.Background()
	user := createUser(t, pool)
	repo := authpg.NewRefreshTokenRepository(pool)

	rawA := "raw-token-a"
	rawB := "raw-token-b"
	family := fmt.Sprintf("fam-%s", newUUID(t)[:8])

	mk := func(raw string) *authdomain.RefreshToken {
		return &authdomain.RefreshToken{
			UserID:      user.ID,
			TokenHash:   hashHex(raw),
			ExpiresAt:   time.Now().Add(24 * time.Hour),
			CreatedByIP: "127.0.0.1",
			Family:      family,
		}
	}

	requireNoErr(t, repo.Create(ctx, mk(rawA)), "Create A")
	requireNoErr(t, repo.Create(ctx, mk(rawB)), "Create B")

	gotA, err := repo.GetByHash(ctx, rawA)
	requireNoErr(t, err, "GetByHash A")
	if gotA.UserID != user.ID || gotA.TokenHash != hashHex(rawA) || gotA.Family != family {
		t.Fatalf("GetByHash returned unexpected token: %+v", gotA)
	}
	if gotA.RevokedAt != nil {
		t.Fatalf("fresh token already revoked: %+v", gotA)
	}

	active, err := repo.GetActiveByUserID(ctx, user.ID)
	requireNoErr(t, err, "GetActiveByUserID")
	if len(active) != 2 {
		t.Fatalf("GetActiveByUserID = %d tokens, want 2", len(active))
	}

	requireNoErr(t, repo.Revoke(ctx, gotA.ID), "Revoke A")
	revoked, _ := repo.GetByHash(ctx, rawA)
	if revoked.RevokedAt == nil {
		t.Fatalf("Revoke did not set revoked_at: %+v", revoked)
	}
	active, _ = repo.GetActiveByUserID(ctx, user.ID)
	if len(active) != 1 || active[0].TokenHash != hashHex(rawB) {
		t.Fatalf("active after Revoke = %d tokens, want only B", len(active))
	}

	requireNoErr(t, repo.RevokeFamily(ctx, family), "RevokeFamily")
	active, _ = repo.GetActiveByUserID(ctx, user.ID)
	if len(active) != 0 {
		t.Fatalf("active after RevokeFamily = %d tokens, want 0", len(active))
	}

	if _, err := repo.GetByHash(ctx, "unknown-token"); !errors.Is(err, authdomain.ErrInvalidRefreshToken) {
		t.Fatalf("GetByHash(unknown) err = %v, want ErrInvalidRefreshToken", err)
	}
}

func TestRefreshTokenRepository_ExpiredTokenNotActive(t *testing.T) {
	pool := setup(t)
	ctx := context.Background()
	user := createUser(t, pool)
	repo := authpg.NewRefreshTokenRepository(pool)

	raw := "expired-token"
	requireNoErr(t, repo.Create(ctx, &authdomain.RefreshToken{
		UserID:      user.ID,
		TokenHash:   hashHex(raw),
		ExpiresAt:   time.Now().Add(-1 * time.Hour),
		CreatedByIP: "127.0.0.1",
		Family:      "fam-expired",
	}), "Create expired")

	active, err := repo.GetActiveByUserID(ctx, user.ID)
	requireNoErr(t, err, "GetActiveByUserID")
	if len(active) != 0 {
		t.Fatalf("expired token reported active: %+v", active)
	}
}

func TestVerificationTokenRepository_RoundTrip(t *testing.T) {
	pool := setup(t)
	ctx := context.Background()
	user := createUser(t, pool)
	repo := authpg.NewVerificationTokenRepo(pool)

	tok := &authdomain.VerificationToken{
		UserID:    user.ID,
		Email:     user.Email,
		TokenHash: hashHex("verify-token"),
		ExpiresAt: time.Now().Add(1 * time.Hour),
	}
	requireNoErr(t, repo.Create(ctx, tok), "Create")
	if tok.ID == "" {
		t.Fatal("Create did not populate token ID")
	}

	got, err := repo.GetByHash(ctx, tok.TokenHash)
	requireNoErr(t, err, "GetByHash")
	if got.ID != tok.ID || got.UserID != user.ID || got.Email != user.Email {
		t.Fatalf("GetByHash returned unexpected token: %+v", got)
	}
	if got.ConsumedAt != nil {
		t.Fatalf("fresh token already consumed: %+v", got)
	}

	active, err := repo.GetActiveByUserID(ctx, user.ID)
	requireNoErr(t, err, "GetActiveByUserID")
	if len(active) != 1 {
		t.Fatalf("GetActiveByUserID = %d tokens, want 1", len(active))
	}

	requireNoErr(t, repo.Consume(ctx, tok.ID), "Consume")
	if err := repo.Consume(ctx, tok.ID); err == nil {
		t.Fatal("second Consume = nil error, want already-consumed error")
	}
	consumed, _ := repo.GetByHash(ctx, tok.TokenHash)
	if consumed.ConsumedAt == nil {
		t.Fatalf("Consume did not set consumed_at: %+v", consumed)
	}
	active, _ = repo.GetActiveByUserID(ctx, user.ID)
	if len(active) != 0 {
		t.Fatalf("consumed token still active: %+v", active)
	}
}

func TestOAuthIdentityRepository_RoundTrip(t *testing.T) {
	pool := setup(t)
	ctx := context.Background()
	user := createUser(t, pool)
	repo := authpg.NewOAuthIdentityRepository(pool)

	expires := time.Now().Add(30 * time.Minute)
	ident := &authdomain.OAuthIdentity{
		UserID:                user.ID,
		Provider:              "google",
		ProviderUserID:        "google-123",
		ProviderEmail:         user.Email,
		AccessTokenEncrypted:  "access-enc",
		RefreshTokenEncrypted: "refresh-enc",
		ExpiresAt:             &expires,
	}
	requireNoErr(t, repo.Create(ctx, ident), "Create")

	dup := *ident
	if err := repo.Create(ctx, &dup); !errors.Is(err, authdomain.ErrProviderAlreadyLinked) {
		t.Fatalf("Create duplicate err = %v, want ErrProviderAlreadyLinked", err)
	}

	byProvider, err := repo.GetByProvider(ctx, "google", "google-123")
	requireNoErr(t, err, "GetByProvider")
	if byProvider == nil || byProvider.ProviderUserID != "google-123" || byProvider.UserID != user.ID {
		t.Fatalf("GetByProvider returned unexpected identity: %+v", byProvider)
	}
	missing, err := repo.GetByProvider(ctx, "google", "nope")
	if missing != nil || err != nil {
		t.Fatalf("GetByProvider(missing) = %+v, %v; want nil, nil", missing, err)
	}

	requireNoErr(t, repo.UpdateLastUsed(ctx, byProvider.ID), "UpdateLastUsed")
	afterUsed, _ := repo.GetByProvider(ctx, "google", "google-123")
	if afterUsed.LastUsedAt == nil {
		t.Fatalf("UpdateLastUsed did not set last_used_at: %+v", afterUsed)
	}

	newExp := time.Now().Add(2 * time.Hour)
	requireNoErr(t, repo.UpdateTokens(ctx, byProvider.ID, "access-enc-2", "refresh-enc-2", &newExp), "UpdateTokens")
	afterTokens, _ := repo.GetByProvider(ctx, "google", "google-123")
	if afterTokens.AccessTokenEncrypted != "access-enc-2" || afterTokens.RefreshTokenEncrypted != "refresh-enc-2" {
		t.Fatalf("UpdateTokens did not persist: %+v", afterTokens)
	}
	if afterTokens.ExpiresAt == nil || afterTokens.ExpiresAt.Sub(newExp) > time.Millisecond {
		t.Fatalf("UpdateTokens did not persist expiry: %+v", afterTokens)
	}

	requireNoErr(t, repo.Delete(ctx, byProvider.ID), "Delete")
	deleted, _ := repo.GetByProvider(ctx, "google", "google-123")
	if deleted != nil {
		t.Fatalf("identity still present after Delete: %+v", deleted)
	}
}

func TestOAuthIdentityRepository_GetByUserIDAndExpiringSoon(t *testing.T) {
	pool := setup(t)
	ctx := context.Background()
	user := createUser(t, pool)
	repo := authpg.NewOAuthIdentityRepository(pool)

	mk := func(provider, providerUserID, refresh string, expiresIn time.Duration) *authdomain.OAuthIdentity {
		exp := time.Now().Add(expiresIn)
		return &authdomain.OAuthIdentity{
			UserID:                user.ID,
			Provider:              provider,
			ProviderUserID:        providerUserID,
			ProviderEmail:         fmt.Sprintf("%s@example.com", provider),
			AccessTokenEncrypted:  "acc-" + provider,
			RefreshTokenEncrypted: refresh,
			ExpiresAt:             &exp,
		}
	}

	requireNoErr(t, repo.Create(ctx, mk("github", "gh-1", "refresh-1", 30*time.Minute)), "Create github")
	requireNoErr(t, repo.Create(ctx, mk("slack", "sl-1", "", 30*time.Minute)), "Create slack no refresh")
	requireNoErr(t, repo.Create(ctx, mk("google", "g-2", "refresh-2", 5*time.Hour)), "Create google far out")

	byUser, err := repo.GetByUserID(ctx, user.ID)
	requireNoErr(t, err, "GetByUserID")
	if len(byUser) != 3 {
		t.Fatalf("GetByUserID = %d identities, want 3", len(byUser))
	}

	expiring, err := repo.GetExpiringSoon(ctx, 2*time.Hour)
	requireNoErr(t, err, "GetExpiringSoon")
	if len(expiring) != 1 || expiring[0].Provider != "github" {
		t.Fatalf("GetExpiringSoon = %+v, want only github identity", expiring)
	}
}
