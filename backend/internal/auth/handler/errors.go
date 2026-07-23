package handler

import (
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func grpcUnauthenticated(msg string) error {
	return status.Error(codes.Unauthenticated, msg)
}

func grpcAlreadyExists(msg string) error {
	return status.Error(codes.AlreadyExists, msg)
}

func grpcNotFound(msg string) error {
	return status.Error(codes.NotFound, msg)
}

func grpcInvalidArgument(msg string) error {
	return status.Error(codes.InvalidArgument, msg)
}

func grpcInternal(msg string) error {
	return status.Error(codes.Internal, msg)
}
