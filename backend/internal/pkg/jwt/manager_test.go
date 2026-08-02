package jwt

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"strings"
	"testing"

	"github.com/golang-jwt/jwt/v5"
)

func generateTestKeyPair(t *testing.T) (privatePEM, publicPEM string) {
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
	publicPEM = string(pem.EncodeToMemory(&pem.Block{
		Type:  "PUBLIC KEY",
		Bytes: pubBytes,
	}))

	return privatePEM, publicPEM
}

func TestNewManagerInvalidKeys(t *testing.T) {
	t.Parallel()

	if _, err := NewManager("not a pem", ""); err == nil {
		t.Error("NewManager with garbage private key = nil error, want error")
	}

	priv, _ := generateTestKeyPair(t)
	if _, err := NewManager(priv, "not a pem"); err == nil {
		t.Error("NewManager with garbage public key = nil error, want error")
	}
}

func TestAccessTokenRoundTrip(t *testing.T) {
	t.Parallel()

	priv, pub := generateTestKeyPair(t)
	m, err := NewManager(priv, pub)
	if err != nil {
		t.Fatal(err)
	}

	token, err := m.GenerateAccessToken("user_1", "dev@schemahub.dev", "admin")
	if err != nil {
		t.Fatalf("GenerateAccessToken() error = %v", err)
	}
	if token == "" {
		t.Fatal("GenerateAccessToken() returned empty token")
	}

	claims, err := m.VerifyAccessToken(token)
	if err != nil {
		t.Fatalf("VerifyAccessToken() error = %v", err)
	}

	if claims["sub"] != "user_1" {
		t.Errorf("claims[sub] = %v, want user_1", claims["sub"])
	}
	if claims["email"] != "dev@schemahub.dev" {
		t.Errorf("claims[email] = %v, want dev@schemahub.dev", claims["email"])
	}
	if claims["role"] != "admin" {
		t.Errorf("claims[role] = %v, want admin", claims["role"])
	}
	if _, ok := claims["exp"]; !ok {
		t.Error("claims[exp] missing, want expiry set")
	}
	if _, ok := claims["jti"]; !ok {
		t.Error("claims[jti] missing, want token ID set")
	}
}

func TestVerifyAccessTokenRejectsTamperedToken(t *testing.T) {
	t.Parallel()

	priv, pub := generateTestKeyPair(t)
	m, err := NewManager(priv, pub)
	if err != nil {
		t.Fatal(err)
	}

	token, err := m.GenerateAccessToken("user_1", "a@b.dev", "member")
	if err != nil {
		t.Fatal(err)
	}

	tampered := token[:len(token)-4] + "AAAA"
	if _, err := m.VerifyAccessToken(tampered); err == nil {
		t.Error("VerifyAccessToken(tampered) = nil error, want error")
	}
}

func TestVerifyAccessTokenRejectsWrongSigningMethod(t *testing.T) {
	t.Parallel()

	priv, pub := generateTestKeyPair(t)
	m, err := NewManager(priv, pub)
	if err != nil {
		t.Fatal(err)
	}

	hsToken := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{"sub": "user_1"})
	hsString, err := hsToken.SignedString([]byte("shared-secret"))
	if err != nil {
		t.Fatal(err)
	}

	if _, err := m.VerifyAccessToken(hsString); err == nil {
		t.Error("VerifyAccessToken(HS256 token) = nil error, want error")
	}
}

func TestSignAndValidateClaims(t *testing.T) {
	t.Parallel()

	priv, pub := generateTestKeyPair(t)
	m, err := NewManager(priv, pub)
	if err != nil {
		t.Fatal(err)
	}

	claims := jwt.MapClaims{"sub": "user_9", "scope": "schema:read"}
	signed, err := m.SignClaims(claims)
	if err != nil {
		t.Fatalf("SignClaims() error = %v", err)
	}

	out := jwt.MapClaims{}
	if err := m.ValidateClaims(signed, out); err != nil {
		t.Fatalf("ValidateClaims() error = %v", err)
	}
	if out["sub"] != "user_9" {
		t.Errorf("validated sub = %v, want user_9", out["sub"])
	}
}

func TestValidateClaimsRejectsInvalid(t *testing.T) {
	t.Parallel()

	priv, pub := generateTestKeyPair(t)
	m, err := NewManager(priv, pub)
	if err != nil {
		t.Fatal(err)
	}

	err = m.ValidateClaims("garbage.token.here", jwt.MapClaims{})
	if err == nil {
		t.Error("ValidateClaims(garbage) = nil error, want error")
	}
}

func TestGenerateRefreshToken(t *testing.T) {
	t.Parallel()

	priv, pub := generateTestKeyPair(t)
	m, err := NewManager(priv, pub)
	if err != nil {
		t.Fatal(err)
	}

	token, err := m.GenerateRefreshToken()
	if err != nil {
		t.Fatalf("GenerateRefreshToken() error = %v", err)
	}
	if !strings.HasPrefix(token, "rt_") {
		t.Errorf("refresh token = %q, want prefix rt_", token)
	}
	if len(token) < 32 {
		t.Errorf("refresh token too short: %d chars", len(token))
	}
}
