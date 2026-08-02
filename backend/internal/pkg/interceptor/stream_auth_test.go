package interceptor

import (
	"context"
	"reflect"
	"testing"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type fakeServerStream struct {
	grpc.ServerStream
	ctx     context.Context
	first   any
	recvCnt int
}

func (f *fakeServerStream) Context() context.Context { return f.ctx }

func (f *fakeServerStream) RecvMsg(m any) error {
	f.recvCnt++
	if f.first == nil {
		return nil
	}
	rv := reflect.ValueOf(m)
	sv := reflect.ValueOf(f.first)
	if rv.Kind() == reflect.Ptr && sv.Kind() == reflect.Ptr {
		rv.Elem().Set(sv.Elem())
	}
	return nil
}

type streamOKHandler struct{}

func (s *streamOKHandler) handle(srv interface{}, ss grpc.ServerStream) error { return nil }

func okStreamHandler(srv interface{}, ss grpc.ServerStream) error { return nil }

func TestStreamAuthInterceptor(t *testing.T) {
	t.Parallel()

	m := newTestManager(t)
	inter := StreamAuthInterceptor(m)
	info := &grpc.StreamServerInfo{FullMethod: "/x/Y"}

	t.Run("public methods pass through", func(t *testing.T) {
		pubInfo := &grpc.StreamServerInfo{FullMethod: "/schemahub.auth.v1.AuthService/Login"}
		ss := &fakeServerStream{ctx: context.Background()}
		if err := inter(nil, ss, pubInfo, okStreamHandler); err != nil {
			t.Errorf("public: err = %v", err)
		}
	})

	t.Run("missing token rejected", func(t *testing.T) {
		ss := &fakeServerStream{ctx: context.Background()}
		err := inter(nil, ss, info, okStreamHandler)
		if status.Code(err) != codes.Unauthenticated {
			t.Errorf("err = %v, want Unauthenticated", err)
		}
	})

	t.Run("valid token injects claims into stream context", func(t *testing.T) {
		token, err := m.GenerateAccessToken("user_9", "s@example.com", "admin")
		if err != nil {
			t.Fatal(err)
		}
		ss := &fakeServerStream{ctx: bearerCtx(token)}

		handler := func(srv interface{}, ss grpc.ServerStream) error {
			id, err1 := UserIDFromContext(ss.Context())
			role, err2 := UserRoleFromContext(ss.Context())
			if err1 != nil || err2 != nil {
				t.Fatalf("claims missing: %v %v", err1, err2)
			}
			if id != "user_9" || role != "admin" {
				t.Errorf("claims = %q/%q", id, role)
			}
			return nil
		}

		if err := inter(nil, ss, info, handler); err != nil {
			t.Errorf("valid token: %v", err)
		}
	})
}

func TestStreamRBACInterceptor(t *testing.T) {
	t.Parallel()

	type req struct{ ProjectId string }

	deny := func(ctx context.Context, userID, role, fullMethod string, req any) error {
		return status.Error(codes.PermissionDenied, "denied")
	}
	allow := func(ctx context.Context, userID, role, fullMethod string, req any) error {
		return nil
	}

	t.Run("rejects when not authenticated", func(t *testing.T) {
		inter := StreamRBACInterceptor(allow)
		ss := &fakeServerStream{ctx: context.Background()}
		err := inter(nil, ss, &grpc.StreamServerInfo{FullMethod: "/x/Y"}, okStreamHandler)
		if status.Code(err) != codes.Unauthenticated {
			t.Errorf("err = %v, want Unauthenticated", err)
		}
	})

	t.Run("deny fails the stream via first RecvMsg", func(t *testing.T) {
		inter := StreamRBACInterceptor(deny)
		ctx := context.WithValue(context.Background(), UserIDKey, "user_1")
		ss := &fakeServerStream{ctx: ctx, first: &req{ProjectId: "p1"}}
		gotReq := &req{}
		handler := func(srv interface{}, ss grpc.ServerStream) error {
			if err := ss.RecvMsg(gotReq); err != nil {
				return err
			}
			return nil
		}
		err := inter(nil, ss, &grpc.StreamServerInfo{FullMethod: "/x/Y"}, handler)
		if status.Code(err) != codes.PermissionDenied {
			t.Errorf("err = %v, want PermissionDenied", err)
		}
		if gotReq.ProjectId != "p1" {
			t.Errorf("request was decoded before check: %+v", gotReq)
		}
	})

	t.Run("allow decodes the request and passes through", func(t *testing.T) {
		inter := StreamRBACInterceptor(allow)
		ctx := context.WithValue(context.Background(), UserIDKey, "user_1")
		ss := &fakeServerStream{ctx: ctx, first: &req{ProjectId: "p1"}}
		gotReq := &req{}
		handler := func(srv interface{}, ss grpc.ServerStream) error {
			if err := ss.RecvMsg(gotReq); err != nil {
				return err
			}
			if gotReq.ProjectId != "p1" {
				t.Errorf("ProjectId = %q, want p1", gotReq.ProjectId)
			}
			return nil
		}
		if err := inter(nil, ss, &grpc.StreamServerInfo{FullMethod: "/x/Y"}, handler); err != nil {
			t.Errorf("allow: %v", err)
		}
	})

	t.Run("check receives the request and method", func(t *testing.T) {
		var gotReq any
		var gotMethod string
		inter := StreamRBACInterceptor(func(ctx context.Context, userID, role, fullMethod string, req any) error {
			gotReq = req
			gotMethod = fullMethod
			return nil
		})
		ctx := context.WithValue(context.Background(), UserIDKey, "user_1")
		ss := &fakeServerStream{ctx: ctx, first: &req{ProjectId: "p9"}}
		handler := func(srv interface{}, ss grpc.ServerStream) error {
			ss.RecvMsg(&req{})
			return nil
		}
		if err := inter(nil, ss, &grpc.StreamServerInfo{FullMethod: "/svc/M"}, handler); err != nil {
			t.Fatal(err)
		}
		if gotMethod != "/svc/M" {
			t.Errorf("method = %q, want /svc/M", gotMethod)
		}
		if r, ok := gotReq.(*req); !ok || r.ProjectId != "p9" {
			t.Errorf("req = %#v, want project p9", gotReq)
		}
	})
}
