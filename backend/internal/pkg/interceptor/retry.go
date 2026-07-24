package interceptor

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5/pgconn"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var retriableCodes = map[string]bool{
	"40001": true,
	"40P01": true,
}

func DBRetryInterceptor(maxRetries int) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		var lastErr error

		for attempt := 0; attempt <= maxRetries; attempt++ {
			resp, err := handler(ctx, req)
			if err == nil {
				return resp, nil
			}

			if !isRetriable(err) {
				return nil, err
			}

			lastErr = err

			if attempt < maxRetries {
				backoff := time.Duration(100*(1<<attempt)) * time.Millisecond
				select {
				case <-ctx.Done():
					return nil, status.Error(codes.DeadlineExceeded, "context cancelled during retry")
				case <-time.After(backoff):
				}
			}
		}

		return nil, lastErr
	}
}

func isRetriable(err error) bool {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		return retriableCodes[pgErr.Code]
	}
	return false
}
