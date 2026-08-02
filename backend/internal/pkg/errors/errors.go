package errors

import (
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Error struct {
	Code    string
	Message string
	Err     error
}

func (e *Error) Error() string {
	if e.Err != nil {
		return e.Message + ": " + e.Err.Error()
	}
	return e.Message
}

func (e *Error) Unwrap() error {
	return e.Err
}

func New(code, msg string) *Error {
	return &Error{Code: code, Message: msg}
}

func Wrap(code, msg string, err error) *Error {
	return &Error{Code: code, Message: msg, Err: err}
}

func ToGRPC(err error) error {
	if err == nil {
		return nil
	}
	var e *Error
	if as, ok := err.(*Error); ok {
		e = as
	} else {
		return status.Error(codes.Internal, "internal server error")
	}

	switch e.Code {
	case "NOT_FOUND":
		return status.Error(codes.NotFound, e.Message)
	case "ALREADY_EXISTS":
		return status.Error(codes.AlreadyExists, e.Message)
	case "INVALID_ARGUMENT":
		return status.Error(codes.InvalidArgument, e.Message)
	case "UNAUTHENTICATED":
		return status.Error(codes.Unauthenticated, e.Message)
	case "PERMISSION_DENIED":
		return status.Error(codes.PermissionDenied, e.Message)
	case "FAILED_PRECONDITION":
		return status.Error(codes.FailedPrecondition, e.Message)
	case "RESOURCE_EXHAUSTED":
		return status.Error(codes.ResourceExhausted, e.Message)
	default:
		return status.Error(codes.Internal, e.Message)
	}
}
