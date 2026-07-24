package driftv1

import (
	"context"

	"google.golang.org/grpc"
)

type DriftServiceServer interface {
	CheckDrift(ctx context.Context, req *CheckDriftRequest) (*CheckDriftResponse, error)
	ListDriftEvents(ctx context.Context, req *ListDriftEventsRequest) (*ListDriftEventsResponse, error)
	GetDriftEvent(ctx context.Context, req *GetDriftEventRequest) (*GetDriftEventResponse, error)
	ResolveDriftEvent(ctx context.Context, req *ResolveDriftEventRequest) (*ResolveDriftEventResponse, error)
	GetDriftStats(ctx context.Context, req *GetDriftStatsRequest) (*GetDriftStatsResponse, error)
}

type UnimplementedDriftServiceServer struct{}

func (UnimplementedDriftServiceServer) CheckDrift(ctx context.Context, req *CheckDriftRequest) (*CheckDriftResponse, error) {
	return &CheckDriftResponse{}, nil
}

func (UnimplementedDriftServiceServer) ListDriftEvents(ctx context.Context, req *ListDriftEventsRequest) (*ListDriftEventsResponse, error) {
	return &ListDriftEventsResponse{}, nil
}

func (UnimplementedDriftServiceServer) GetDriftEvent(ctx context.Context, req *GetDriftEventRequest) (*GetDriftEventResponse, error) {
	return &GetDriftEventResponse{}, nil
}

func (UnimplementedDriftServiceServer) ResolveDriftEvent(ctx context.Context, req *ResolveDriftEventRequest) (*ResolveDriftEventResponse, error) {
	return &ResolveDriftEventResponse{}, nil
}

func (UnimplementedDriftServiceServer) GetDriftStats(ctx context.Context, req *GetDriftStatsRequest) (*GetDriftStatsResponse, error) {
	return &GetDriftStatsResponse{}, nil
}

func RegisterDriftServiceServer(s *grpc.Server, srv DriftServiceServer) {
	s.RegisterService(&DriftService_ServiceDesc, srv)
}

var DriftService_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "schemahub.drift.v1.DriftService",
	Methods: []grpc.MethodDesc{
		{MethodName: "CheckDrift", Handler: _DriftService_CheckDrift_Handler},
		{MethodName: "ListDriftEvents", Handler: _DriftService_ListDriftEvents_Handler},
		{MethodName: "GetDriftEvent", Handler: _DriftService_GetDriftEvent_Handler},
		{MethodName: "ResolveDriftEvent", Handler: _DriftService_ResolveDriftEvent_Handler},
		{MethodName: "GetDriftStats", Handler: _DriftService_GetDriftStats_Handler},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "drift/v1/drift_service.proto",
}

func _DriftService_CheckDrift_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, _ grpc.UnaryServerInterceptor) (interface{}, error) {
	req := &CheckDriftRequest{}
	if err := dec(req); err != nil {
		return nil, err
	}
	return srv.(DriftServiceServer).CheckDrift(ctx, req)
}

func _DriftService_ListDriftEvents_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, _ grpc.UnaryServerInterceptor) (interface{}, error) {
	req := &ListDriftEventsRequest{}
	if err := dec(req); err != nil {
		return nil, err
	}
	return srv.(DriftServiceServer).ListDriftEvents(ctx, req)
}

func _DriftService_GetDriftEvent_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, _ grpc.UnaryServerInterceptor) (interface{}, error) {
	req := &GetDriftEventRequest{}
	if err := dec(req); err != nil {
		return nil, err
	}
	return srv.(DriftServiceServer).GetDriftEvent(ctx, req)
}

func _DriftService_ResolveDriftEvent_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, _ grpc.UnaryServerInterceptor) (interface{}, error) {
	req := &ResolveDriftEventRequest{}
	if err := dec(req); err != nil {
		return nil, err
	}
	return srv.(DriftServiceServer).ResolveDriftEvent(ctx, req)
}

func _DriftService_GetDriftStats_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, _ grpc.UnaryServerInterceptor) (interface{}, error) {
	req := &GetDriftStatsRequest{}
	if err := dec(req); err != nil {
		return nil, err
	}
	return srv.(DriftServiceServer).GetDriftStats(ctx, req)
}
