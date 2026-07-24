package interceptor

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

func IdempotencyInterceptor(rdb *redis.Client) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		md, ok := metadata.FromIncomingContext(ctx)
		if !ok {
			return handler(ctx, req)
		}

		keys := md.Get("x-idempotency-key")
		if len(keys) == 0 {
			return handler(ctx, req)
		}
		key := keys[0]

		cacheKey := fmt.Sprintf("idempotent:%s", key)
		exists, err := rdb.Exists(ctx, cacheKey).Result()
		if err != nil {
			return handler(ctx, req)
		}

		if exists > 0 {
			return nil, status.Errorf(codes.AlreadyExists, "request with idempotency key %s has already been processed", key)
		}

		if err := rdb.Set(ctx, cacheKey, "1", 5*time.Minute).Err(); err != nil {
			return handler(ctx, req)
		}

		return handler(ctx, req)
	}
}
