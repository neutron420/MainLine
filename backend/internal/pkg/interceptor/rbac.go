package interceptor

import (
	"context"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type PermissionCheckFunc func(ctx context.Context, userID, role, fullMethod string) error

func RBACInterceptor(checkFn PermissionCheckFunc) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		if publicMethods[info.FullMethod] {
			return handler(ctx, req)
		}

		userID, _ := UserIDFromContext(ctx)
		role, _ := UserRoleFromContext(ctx)

		if userID == "" {
			return nil, status.Error(codes.Unauthenticated, "not authenticated")
		}

		if err := checkFn(ctx, userID, role, info.FullMethod); err != nil {
			return nil, status.Error(codes.PermissionDenied, err.Error())
		}

		return handler(ctx, req)
	}
}
