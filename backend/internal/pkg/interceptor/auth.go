package interceptor

import (
	"context"
	"strings"

	"github.com/schemahub/backend/internal/pkg/jwt"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

type contextKey string

const (
	UserIDKey    contextKey = "user_id"
	UserEmailKey contextKey = "user_email"
	UserRoleKey  contextKey = "user_role"
)

var publicMethods = map[string]bool{
	"/schemahub.auth.v1.AuthService/Register":            true,
	"/schemahub.auth.v1.AuthService/Login":               true,
	"/schemahub.auth.v1.AuthService/RefreshToken":        true,
	"/schemahub.auth.v1.AuthService/GetOAuthURL":         true,
	"/schemahub.auth.v1.AuthService/HandleOAuthCallback": true,
	"/schemahub.auth.v1.AuthService/ForgotPassword":      true,
	"/schemahub.auth.v1.AuthService/ResetPassword":       true,
}

func AuthInterceptor(jwtManager *jwt.Manager) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		if publicMethods[info.FullMethod] {
			return handler(ctx, req)
		}

		md, ok := metadata.FromIncomingContext(ctx)
		if !ok {
			return nil, status.Error(codes.Unauthenticated, "missing metadata")
		}

		authHeaders := md.Get("authorization")
		if len(authHeaders) == 0 {
			return nil, status.Error(codes.Unauthenticated, "missing authorization header")
		}

		token := strings.TrimPrefix(authHeaders[0], "Bearer ")
		if token == authHeaders[0] {
			return nil, status.Error(codes.Unauthenticated, "invalid authorization format")
		}

		claims, err := jwtManager.VerifyAccessToken(token)
		if err != nil {
			return nil, status.Error(codes.Unauthenticated, "invalid or expired token")
		}

		ctx = context.WithValue(ctx, UserIDKey, claims["sub"])
		ctx = context.WithValue(ctx, UserEmailKey, claims["email"])
		ctx = context.WithValue(ctx, UserRoleKey, claims["role"])

		return handler(ctx, req)
	}
}
