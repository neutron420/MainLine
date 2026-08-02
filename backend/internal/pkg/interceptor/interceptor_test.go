package interceptor

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"net"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/redis/go-redis/v9"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/peer"
	"google.golang.org/grpc/status"
)

// ── helpers ──

func okHandler(ctx context.Context, req interface{}) (interface{}, error) {
	return "ok", nil
}

func errHandler(code codes.Code, msg string) grpc.UnaryHandler {
	return func(ctx context.Context, req interface{}) (interface{}, error) {
		return nil, status.Error(code, msg)
	}
}

// ── context helpers ──

func TestUserFromContext(t *testing.T) {
	t.Parallel()

	if _, err := UserIDFromContext(context.Background()); err == nil {
		t.Error("UserIDFromContext(empty) = nil error, want error")
	}
	if _, err := UserEmailFromContext(context.Background()); err == nil {
		t.Error("UserEmailFromContext(empty) = nil error, want error")
	}
	if _, err := UserRoleFromContext(context.Background()); err == nil {
		t.Error("UserRoleFromContext(empty) = nil error, want error")
	}

	ctx := context.WithValue(context.Background(), UserIDKey, "user_1")
	ctx = context.WithValue(ctx, UserEmailKey, "a@b.com")
	ctx = context.WithValue(ctx, UserRoleKey, "admin")

	id, err := UserIDFromContext(ctx)
	if err != nil || id != "user_1" {
		t.Errorf("UserIDFromContext() = %q, %v; want user_1", id, err)
	}
	email, _ := UserEmailFromContext(ctx)
	if email != "a@b.com" {
		t.Errorf("UserEmailFromContext() = %q", email)
	}
	role, _ := UserRoleFromContext(ctx)
	if role != "admin" {
		t.Errorf("UserRoleFromContext() = %q", role)
	}
}

// ── token bucket / rate limiter ──

func TestTokenBucket(t *testing.T) {
	t.Parallel()

	tb := newTokenBucket(10, 5)
	for i := 0; i < 5; i++ {
		if !tb.allow() {
			t.Fatalf("allow() %d = false, want true (burst 5)", i)
		}
	}
	if tb.allow() {
		t.Error("allow() beyond burst = true, want false")
	}

	// refill: 10 tokens/sec → ~1 token per 100ms
	time.Sleep(120 * time.Millisecond)
	if !tb.allow() {
		t.Error("allow() after refill = false, want true")
	}
}

func TestRateLimiterByUser(t *testing.T) {
	t.Parallel()

	rl := NewRateLimiter(1, 2)
	defer rl.Stop()

	ctx := context.WithValue(context.Background(), UserIDKey, "user_1")
	inter := rl.UnaryServerInterceptor()

	var calls atomic.Int32
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		calls.Add(1)
		return "ok", nil
	}

	info := &grpc.UnaryServerInfo{FullMethod: "/x/Y"}
	if _, err := inter(ctx, nil, info, handler); err != nil {
		t.Fatalf("first call: %v", err)
	}
	if _, err := inter(ctx, nil, info, handler); err != nil {
		t.Fatalf("second call: %v", err)
	}
	if _, err := inter(ctx, nil, info, handler); err == nil {
		t.Error("third call = nil error, want ResourceExhausted")
	}
	if calls.Load() != 2 {
		t.Errorf("handler calls = %d, want 2", calls.Load())
	}
}

func TestRateLimiterFallsBackToIP(t *testing.T) {
	t.Parallel()

	rl := NewRateLimiter(1, 1)
	defer rl.Stop()

	ctx := peer.NewContext(context.Background(), &peer.Peer{
		Addr: &net.TCPAddr{IP: net.ParseIP("10.0.0.1"), Port: 5555},
	})
	inter := rl.UnaryServerInterceptor()
	info := &grpc.UnaryServerInfo{FullMethod: "/x/Y"}

	if _, err := inter(ctx, nil, info, okHandler); err != nil {
		t.Fatalf("call with IP: %v", err)
	}
	if _, err := inter(ctx, nil, info, okHandler); err == nil {
		t.Error("second call with same IP = nil error, want ResourceExhausted")
	}

	// different IP gets its own bucket
	ctx2 := peer.NewContext(context.Background(), &peer.Peer{
		Addr: &net.TCPAddr{IP: net.ParseIP("10.0.0.2"), Port: 5555},
	})
	if _, err := inter(ctx2, nil, info, okHandler); err != nil {
		t.Errorf("call from other IP: %v", err)
	}
}

func TestExtractIP(t *testing.T) {
	t.Parallel()

	if got := extractIP(context.Background()); got != "" {
		t.Errorf("extractIP(no peer) = %q, want empty", got)
	}

	ctx := peer.NewContext(context.Background(), &peer.Peer{
		Addr: &net.TCPAddr{IP: net.ParseIP("192.168.1.10"), Port: 1111},
	})
	if got := extractIP(ctx); got != "192.168.1.10" {
		t.Errorf("extractIP() = %q, want 192.168.1.10", got)
	}
}

// ── validation ──

type validReq struct{ ok bool }

func (v validReq) Validate() error {
	if !v.ok {
		return errors.New("request invalid")
	}
	return nil
}

func TestValidationInterceptor(t *testing.T) {
	t.Parallel()

	inter := ValidationInterceptor()
	info := &grpc.UnaryServerInfo{FullMethod: "/x/Y"}

	if _, err := inter(context.Background(), validReq{ok: false}, info, okHandler); status.Code(err) != codes.InvalidArgument {
		t.Errorf("invalid req: err = %v, want InvalidArgument", err)
	}

	if _, err := inter(context.Background(), validReq{ok: true}, info, okHandler); err != nil {
		t.Errorf("valid req: %v", err)
	}

	// requests without Validate() pass through
	if _, err := inter(context.Background(), "plain", info, okHandler); err != nil {
		t.Errorf("plain req: %v", err)
	}
}

// ── retry ──

func TestIsRetriable(t *testing.T) {
	t.Parallel()

	serializationErr := &pgconn.PgError{Code: "40001"}
	deadlockErr := &pgconn.PgError{Code: "40P01"}
	otherErr := &pgconn.PgError{Code: "23505"}

	if !isRetriable(serializationErr) {
		t.Error("40001 should be retriable")
	}
	if !isRetriable(deadlockErr) {
		t.Error("40P01 should be retriable")
	}
	if isRetriable(otherErr) {
		t.Error("23505 should not be retriable")
	}
	if isRetriable(errors.New("plain error")) {
		t.Error("plain error should not be retriable")
	}
}

func TestDBRetryInterceptor(t *testing.T) {
	t.Parallel()

	info := &grpc.UnaryServerInfo{FullMethod: "/x/Y"}
	retriableErr := &pgconn.PgError{Code: "40001", Message: "serialization failure"}

	t.Run("succeeds immediately", func(t *testing.T) {
		inter := DBRetryInterceptor(2)
		resp, err := inter(context.Background(), nil, info, okHandler)
		if err != nil || resp != "ok" {
			t.Errorf("resp = %v, err = %v", resp, err)
		}
	})

	t.Run("succeeds after retries", func(t *testing.T) {
		inter := DBRetryInterceptor(3)
		var attempts atomic.Int32
		handler := func(ctx context.Context, req interface{}) (interface{}, error) {
			if attempts.Add(1) < 3 {
				return nil, retriableErr
			}
			return "ok", nil
		}
		resp, err := inter(context.Background(), nil, info, handler)
		if err != nil || resp != "ok" {
			t.Errorf("resp = %v, err = %v; want ok", resp, err)
		}
		if attempts.Load() != 3 {
			t.Errorf("attempts = %d, want 3", attempts.Load())
		}
	})

	t.Run("gives up after max retries", func(t *testing.T) {
		inter := DBRetryInterceptor(1)
		var attempts atomic.Int32
		handler := func(ctx context.Context, req interface{}) (interface{}, error) {
			attempts.Add(1)
			return nil, retriableErr
		}
		if _, err := inter(context.Background(), nil, info, handler); err == nil {
			t.Error("persistent retriable error = nil, want error")
		}
		if attempts.Load() != 2 {
			t.Errorf("attempts = %d, want 2 (initial + 1 retry)", attempts.Load())
		}
	})

	t.Run("non-retriable passes through", func(t *testing.T) {
		inter := DBRetryInterceptor(3)
		handler := errHandler(codes.NotFound, "nope")
		_, err := inter(context.Background(), nil, info, handler)
		if status.Code(err) != codes.NotFound {
			t.Errorf("err = %v, want NotFound", err)
		}
	})
}

// ── recovery ──

func TestRecoveryInterceptor(t *testing.T) {
	t.Parallel()

	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	inter := RecoveryInterceptor(logger)
	info := &grpc.UnaryServerInfo{FullMethod: "/x/Y"}

	panicHandler := func(ctx context.Context, req interface{}) (interface{}, error) {
		panic("boom")
	}

	_, err := inter(context.Background(), nil, info, panicHandler)
	if status.Code(err) != codes.Internal {
		t.Errorf("err = %v, want Internal", err)
	}
	if err == nil || !strings.Contains(err.Error(), "internal server error") {
		t.Errorf("err = %v, want sanitized message", err)
	}
}

func TestLoggingInterceptor(t *testing.T) {
	t.Parallel()

	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	inter := LoggingInterceptor(logger)
	info := &grpc.UnaryServerInfo{FullMethod: "/x/Y"}

	if _, err := inter(context.Background(), nil, info, okHandler); err != nil {
		t.Errorf("success path: %v", err)
	}
	if _, err := inter(context.Background(), nil, info, errHandler(codes.Internal, "x")); err == nil {
		t.Error("error path should propagate error")
	}
}

// ── idempotency (redis unreachable → graceful fallback) ──

func TestIdempotencyInterceptorFallback(t *testing.T) {
	t.Parallel()

	// client pointing nowhere → all redis ops error → interceptor must fall through
	rdb := redis.NewClient(&redis.Options{Addr: "127.0.0.1:1"})
	defer rdb.Close()

	inter := IdempotencyInterceptor(rdb)
	info := &grpc.UnaryServerInfo{FullMethod: "/x/Y"}

	t.Run("no metadata", func(t *testing.T) {
		resp, err := inter(context.Background(), nil, info, okHandler)
		if err != nil || resp != "ok" {
			t.Errorf("resp = %v, err = %v", resp, err)
		}
	})

	t.Run("no idempotency key", func(t *testing.T) {
		ctx := metadata.NewIncomingContext(context.Background(), metadata.Pairs("other", "1"))
		if _, err := inter(ctx, nil, info, okHandler); err != nil {
			t.Errorf("no key: %v", err)
		}
	})

	t.Run("redis down still calls handler", func(t *testing.T) {
		ctx := metadata.NewIncomingContext(context.Background(), metadata.Pairs("x-idempotency-key", "k1"))
		if _, err := inter(ctx, nil, info, okHandler); err != nil {
			t.Errorf("redis down: %v", err)
		}
	})
}
