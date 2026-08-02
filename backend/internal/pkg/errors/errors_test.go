package errors

import (
	"errors"
	"strings"
	"testing"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestError_Error(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		e       *Error
		wantSub string
	}{
		{name: "with cause", e: Wrap("NOT_FOUND", "user not found", errors.New("row missing")), wantSub: "user not found: row missing"},
		{name: "without cause", e: New("NOT_FOUND", "user not found"), wantSub: "user not found"},
		{name: "empty message", e: New("NOT_FOUND", ""), wantSub: ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.e.Error()
			if !strings.Contains(got, tt.wantSub) {
				t.Errorf("Error() = %q, want it to contain %q", got, tt.wantSub)
			}
		})
	}
}

func TestError_Unwrap(t *testing.T) {
	t.Parallel()

	cause := errors.New("root cause")
	e := Wrap("NOT_FOUND", "context", cause)

	if !errors.Is(e, cause) {
		t.Errorf("errors.Is(e, cause) = false, want true")
	}
	if unwrapped := e.Unwrap(); unwrapped != cause {
		t.Errorf("Unwrap() = %v, want %v", unwrapped, cause)
	}
}

func TestToGRPC(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		err      error
		wantCode codes.Code
	}{
		{name: "nil", err: nil, wantCode: codes.OK},
		{name: "not found", err: New("NOT_FOUND", "x"), wantCode: codes.NotFound},
		{name: "already exists", err: New("ALREADY_EXISTS", "x"), wantCode: codes.AlreadyExists},
		{name: "invalid argument", err: New("INVALID_ARGUMENT", "x"), wantCode: codes.InvalidArgument},
		{name: "unauthenticated", err: New("UNAUTHENTICATED", "x"), wantCode: codes.Unauthenticated},
		{name: "permission denied", err: New("PERMISSION_DENIED", "x"), wantCode: codes.PermissionDenied},
		{name: "failed precondition", err: New("FAILED_PRECONDITION", "x"), wantCode: codes.FailedPrecondition},
		{name: "resource exhausted", err: New("RESOURCE_EXHAUSTED", "x"), wantCode: codes.ResourceExhausted},
		{name: "unknown code maps to internal", err: New("SOMETHING_ELSE", "x"), wantCode: codes.Internal},
		{name: "plain error maps to internal", err: errors.New("boom"), wantCode: codes.Internal},
		{name: "wrapped code", err: Wrap("NOT_FOUND", "x", errors.New("y")), wantCode: codes.NotFound},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ToGRPC(tt.err)
			if tt.wantCode == codes.OK {
				if got != nil {
					t.Errorf("ToGRPC(nil) = %v, want nil", got)
				}
				return
			}
			if got == nil {
				t.Fatal("ToGRPC() = nil, want status error")
			}
			if code := status.Code(got); code != tt.wantCode {
				t.Errorf("ToGRPC() code = %v, want %v", code, tt.wantCode)
			}
		})
	}
}
