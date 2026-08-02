package middleware

import (
	"context"
	"fmt"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

func CORSInterceptor(allowedOrigins []string) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		if err := checkOrigin(ctx, allowedOrigins); err != nil {
			return nil, err
		}

		resp, err := handler(ctx, req)

		if err == nil {
			setCORSHeaders(ctx)
		}

		return resp, err
	}
}

func CORSStreamInterceptor(allowedOrigins []string) grpc.StreamServerInterceptor {
	return func(srv interface{}, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		if err := checkOrigin(ss.Context(), allowedOrigins); err != nil {
			return err
		}

		if err := handler(srv, ss); err != nil {
			return err
		}

		setCORSHeaders(ss.Context())
		return nil
	}
}

func checkOrigin(ctx context.Context, allowedOrigins []string) error {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return nil
	}

	origin := ""
	if origins := md.Get("origin"); len(origins) > 0 {
		origin = origins[0]
	}

	if origin != "" && !isAllowedOrigin(origin, allowedOrigins) {
		return status.Error(codes.PermissionDenied, fmt.Sprintf("origin %s not allowed", origin))
	}

	return nil
}

func setCORSHeaders(ctx context.Context) {
	header := metadata.Pairs(
		"access-control-allow-origin", originOf(ctx),
		"access-control-allow-methods", "POST, GET, OPTIONS",
		"access-control-allow-headers", "Content-Type, Authorization, X-Idempotency-Key",
		"access-control-expose-headers", "X-RateLimit-Limit, X-RateLimit-Remaining, X-RateLimit-Reset",
	)
	if err := grpc.SetHeader(ctx, header); err != nil {
	}
}

func originOf(ctx context.Context) string {
	if md, ok := metadata.FromIncomingContext(ctx); ok {
		if origins := md.Get("origin"); len(origins) > 0 {
			return origins[0]
		}
	}
	return ""
}

func isAllowedOrigin(origin string, allowed []string) bool {
	if len(allowed) == 0 {
		return true
	}
	for _, a := range allowed {
		if a == "*" || a == origin {
			return true
		}
	}
	return false
}
