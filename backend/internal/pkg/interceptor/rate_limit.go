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

func RateLimitInterceptor(rdb *redis.Client, limit int) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		if publicMethods[info.FullMethod] {
			return handler(ctx, req)
		}

		userID, _ := UserIDFromContext(ctx)
		if userID == "" {
			return handler(ctx, req)
		}

		key := fmt.Sprintf("ratelimit:%s:%s", userID, info.FullMethod)
		pipe := rdb.Pipeline()
		incr := pipe.Incr(ctx, key)
		pipe.Expire(ctx, key, time.Minute)
		if _, err := pipe.Exec(ctx); err != nil {
			return handler(ctx, req)
		}

		if incr.Val() > int64(limit) {
			return nil, status.Error(codes.ResourceExhausted, "rate limit exceeded")
		}

		return handler(ctx, req)
	}
}