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

		ctx, err := authenticate(ctx, jwtManager, info.FullMethod)
		if err != nil {
			return nil, err
		}

		return handler(ctx, req)
	}
}

func StreamAuthInterceptor(jwtManager *jwt.Manager) grpc.StreamServerInterceptor {
	return func(srv interface{}, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		if publicMethods[info.FullMethod] {
			return handler(srv, ss)
		}

		ctx, err := authenticate(ss.Context(), jwtManager, info.FullMethod)
		if err != nil {
			return err
		}

		return handler(srv, &contextStream{ServerStream: ss, ctx: ctx})
	}
}

// contextStream swaps the context of a server stream so auth claims set by the
// interceptor are visible to the handler.
type contextStream struct {
	grpc.ServerStream
	ctx context.Context
}

func (s *contextStream) Context() context.Context { return s.ctx }

func authenticate(ctx context.Context, jwtManager *jwt.Manager, fullMethod string) (context.Context, error) {
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

	return ctx, nil
}
