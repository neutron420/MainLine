package middleware

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

type traceKeyType struct{}

var traceKey = traceKeyType{}

func TracingInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		md, ok := metadata.FromIncomingContext(ctx)
		traceID := ""

		if ok {
			if ids := md.Get("x-trace-id"); len(ids) > 0 {
				traceID = ids[0]
			}
		}

		if traceID == "" {
			traceID = fmt.Sprintf("trace_%s", uuid.NewString())
		}

		ctx = context.WithValue(ctx, traceKey, traceID)
		header := metadata.Pairs("x-trace-id", traceID)
		_ = grpc.SetHeader(ctx, header)

		return handler(ctx, req)
	}
}

func TraceIDFromContext(ctx context.Context) string {
	v := ctx.Value(traceKey)
	if v == nil {
		return ""
	}
	return v.(string)
}
