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
		md, ok := metadata.FromIncomingContext(ctx)
		if !ok {
			return handler(ctx, req)
		}

		origin := ""
		if origins := md.Get("origin"); len(origins) > 0 {
			origin = origins[0]
		}

		if origin != "" && !isAllowedOrigin(origin, allowedOrigins) {
			return nil, status.Error(codes.PermissionDenied, fmt.Sprintf("origin %s not allowed", origin))
		}

		resp, err := handler(ctx, req)

		if err == nil {
			header := metadata.Pairs(
				"access-control-allow-origin", origin,
				"access-control-allow-methods", "POST, GET, OPTIONS",
				"access-control-allow-headers", "Content-Type, Authorization, X-Idempotency-Key",
				"access-control-expose-headers", "X-RateLimit-Limit, X-RateLimit-Remaining, X-RateLimit-Reset",
			)
			if err := grpc.SetHeader(ctx, header); err != nil {
			}
		}

		return resp, err
	}
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
