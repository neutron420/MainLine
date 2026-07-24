package interceptor

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type methodLimit struct {
	limit int
	window time.Duration
}

var endpointLimits = map[string]methodLimit{
	"/schemahub.auth.v1.AuthService/Login":         {limit: 5, window: time.Minute},
	"/schemahub.auth.v1.AuthService/Register":      {limit: 3, window: time.Minute},
	"/schemahub.auth.v1.AuthService/ForgotPassword": {limit: 3, window: time.Minute},
}

func RateLimitInterceptor(rdb *redis.Client, defaultLimit int) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		if publicMethods[info.FullMethod] {
			return handler(ctx, req)
		}

		limit := defaultLimit
		window := time.Minute

		if ml, ok := endpointLimits[info.FullMethod]; ok {
			limit = ml.limit
			window = ml.window
		}

		userID, _ := UserIDFromContext(ctx)
		ip := extractIP(ctx)

		key := fmt.Sprintf("ratelimit:%s:%s", info.FullMethod, userID)
		if userID == "" {
			key = fmt.Sprintf("ratelimit:%s:%s", info.FullMethod, ip)
		}

		pipe := rdb.Pipeline()
		incr := pipe.Incr(ctx, key)
		pipe.Expire(ctx, key, window)
		if _, err := pipe.Exec(ctx); err != nil {
			return handler(ctx, req)
		}

		if incr.Val() > int64(limit) {
			return nil, status.Error(codes.ResourceExhausted, fmt.Sprintf("rate limit exceeded: %d per %v", limit, window))
		}

		return handler(ctx, req)
	}
}


