package interceptor

import (
	"context"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type PermissionCheckFunc func(ctx context.Context, userID, role, fullMethod string, req any) error

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

		if err := checkFn(ctx, userID, role, info.FullMethod, req); err != nil {
			return nil, status.Error(codes.PermissionDenied, err.Error())
		}

		return handler(ctx, req)
	}
}

func StreamRBACInterceptor(checkFn PermissionCheckFunc) grpc.StreamServerInterceptor {
	return func(srv interface{}, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		if publicMethods[info.FullMethod] {
			return handler(srv, ss)
		}

		userID, _ := UserIDFromContext(ss.Context())
		role, _ := UserRoleFromContext(ss.Context())
		if userID == "" {
			return status.Error(codes.Unauthenticated, "not authenticated")
		}

		return handler(srv, &checkStream{
			ServerStream: ss,
			checkFn:      checkFn,
			userID:       userID,
			role:         role,
			fullMethod:   info.FullMethod,
		})
	}
}

// checkStream authorizes the first message received on a server-streaming RPC.
// The generated handler performs the initial RecvMsg, so the check runs on the
// concrete request type before the handler sees it. Subsequent messages (client
// streaming) pass through unchecked, matching the unary interceptor semantics.
type checkStream struct {
	grpc.ServerStream
	checkFn    PermissionCheckFunc
	userID     string
	role       string
	fullMethod string
	checked    bool
}

func (s *checkStream) RecvMsg(m any) error {
	if err := s.ServerStream.RecvMsg(m); err != nil {
		return err
	}
	if !s.checked {
		s.checked = true
		if err := s.checkFn(s.Context(), s.userID, s.role, s.fullMethod, m); err != nil {
			return status.Error(codes.PermissionDenied, err.Error())
		}
	}
	return nil
}
