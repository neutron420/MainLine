package interceptor

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"testing"

	"github.com/schemahub/backend/internal/pkg/jwt"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
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

func newTestManager(t *testing.T) *jwt.Manager {
	t.Helper()
	priv, pub := generateTestKeyPair(t)
	m, err := jwt.NewManager(priv, pub)
	if err != nil {
		t.Fatalf("NewManager: %v", err)
	}
	return m
}

func bearerCtx(token string) context.Context {
	return metadata.NewIncomingContext(context.Background(), metadata.Pairs("authorization", "Bearer "+token))
}

func TestAuthInterceptor(t *testing.T) {
	t.Parallel()

	m := newTestManager(t)
	inter := AuthInterceptor(m)
	info := &grpc.UnaryServerInfo{FullMethod: "/x/Y"}

	t.Run("public methods pass through", func(t *testing.T) {
		pubInfo := &grpc.UnaryServerInfo{FullMethod: "/schemahub.auth.v1.AuthService/Login"}
		resp, err := inter(context.Background(), nil, pubInfo, okHandler)
		if err != nil || resp != "ok" {
			t.Errorf("public: resp = %v, err = %v", resp, err)
		}
	})

	t.Run("missing metadata", func(t *testing.T) {
		_, err := inter(context.Background(), nil, info, okHandler)
		if status.Code(err) != codes.Unauthenticated {
			t.Errorf("missing metadata: err = %v, want Unauthenticated", err)
		}
	})

	t.Run("missing authorization header", func(t *testing.T) {
		ctx := metadata.NewIncomingContext(context.Background(), metadata.Pairs("x", "1"))
		_, err := inter(ctx, nil, info, okHandler)
		if status.Code(err) != codes.Unauthenticated {
			t.Errorf("missing header: err = %v, want Unauthenticated", err)
		}
	})

	t.Run("invalid authorization format", func(t *testing.T) {
		ctx := metadata.NewIncomingContext(context.Background(), metadata.Pairs("authorization", "Token abc"))
		_, err := inter(ctx, nil, info, okHandler)
		if status.Code(err) != codes.Unauthenticated {
			t.Errorf("bad format: err = %v, want Unauthenticated", err)
		}
	})

	t.Run("invalid token", func(t *testing.T) {
		_, err := inter(bearerCtx("garbage.token.value"), nil, info, okHandler)
		if status.Code(err) != codes.Unauthenticated {
			t.Errorf("bad token: err = %v, want Unauthenticated", err)
		}
	})

	t.Run("valid token injects claims", func(t *testing.T) {
		token, err := m.GenerateAccessToken("user_7", "dev@example.com", "member")
		if err != nil {
			t.Fatal(err)
		}

		handler := func(ctx context.Context, req interface{}) (interface{}, error) {
			id, err1 := UserIDFromContext(ctx)
			email, err2 := UserEmailFromContext(ctx)
			role, err3 := UserRoleFromContext(ctx)
			if err1 != nil || err2 != nil || err3 != nil {
				t.Fatalf("claims missing: %v %v %v", err1, err2, err3)
			}
			if id != "user_7" || email != "dev@example.com" || role != "member" {
				t.Errorf("claims = %q/%q/%q", id, email, role)
			}
			return "ok", nil
		}

		if _, err := inter(bearerCtx(token), nil, info, handler); err != nil {
			t.Errorf("valid token: %v", err)
		}
	})
}
