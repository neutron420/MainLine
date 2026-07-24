package auditv1

import (
	"context"

	"google.golang.org/grpc"
)

type AuditServiceServer interface {
	ListAuditEntries(ctx context.Context, req *ListAuditEntriesRequest) (*ListAuditEntriesResponse, error)
	GetAuditEntry(ctx context.Context, req *GetAuditEntryRequest) (*GetAuditEntryResponse, error)
	TailAuditEntries(req *TailAuditEntriesRequest, stream AuditService_TailAuditEntriesServer) error
	GetAuditStats(ctx context.Context, req *GetAuditStatsRequest) (*GetAuditStatsResponse, error)
}

type UnimplementedAuditServiceServer struct{}

func (UnimplementedAuditServiceServer) ListAuditEntries(ctx context.Context, req *ListAuditEntriesRequest) (*ListAuditEntriesResponse, error) {
	return &ListAuditEntriesResponse{}, nil
}

func (UnimplementedAuditServiceServer) GetAuditEntry(ctx context.Context, req *GetAuditEntryRequest) (*GetAuditEntryResponse, error) {
	return &GetAuditEntryResponse{}, nil
}

func (UnimplementedAuditServiceServer) TailAuditEntries(req *TailAuditEntriesRequest, stream AuditService_TailAuditEntriesServer) error {
	return nil
}

func (UnimplementedAuditServiceServer) GetAuditStats(ctx context.Context, req *GetAuditStatsRequest) (*GetAuditStatsResponse, error) {
	return &GetAuditStatsResponse{}, nil
}

func RegisterAuditServiceServer(s *grpc.Server, srv AuditServiceServer) {
	s.RegisterService(&AuditService_ServiceDesc, srv)
}

type AuditService_TailAuditEntriesServer interface {
	Send(*AuditEntry) error
	grpc.ServerStream
}

var AuditService_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "schemahub.audit.v1.AuditService",
	Methods: []grpc.MethodDesc{
		{
			MethodName: "ListAuditEntries",
			Handler:    _AuditService_ListAuditEntries_Handler,
		},
		{
			MethodName: "GetAuditEntry",
			Handler:    _AuditService_GetAuditEntry_Handler,
		},
		{
			MethodName: "GetAuditStats",
			Handler:    _AuditService_GetAuditStats_Handler,
		},
	},
	Streams: []grpc.StreamDesc{
		{
			StreamName:    "TailAuditEntries",
			Handler:       _AuditService_TailAuditEntries_Handler,
			ServerStreams: true,
		},
	},
	Metadata: "audit/v1/audit_service.proto",
}

func _AuditService_ListAuditEntries_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, _ grpc.UnaryServerInterceptor) (interface{}, error) {
	req := &ListAuditEntriesRequest{}
	if err := dec(req); err != nil {
		return nil, err
	}
	return srv.(AuditServiceServer).ListAuditEntries(ctx, req)
}

func _AuditService_GetAuditEntry_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, _ grpc.UnaryServerInterceptor) (interface{}, error) {
	req := &GetAuditEntryRequest{}
	if err := dec(req); err != nil {
		return nil, err
	}
	return srv.(AuditServiceServer).GetAuditEntry(ctx, req)
}

func _AuditService_GetAuditStats_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, _ grpc.UnaryServerInterceptor) (interface{}, error) {
	req := &GetAuditStatsRequest{}
	if err := dec(req); err != nil {
		return nil, err
	}
	return srv.(AuditServiceServer).GetAuditStats(ctx, req)
}

func _AuditService_TailAuditEntries_Handler(srv interface{}, stream grpc.ServerStream) error {
	req := &TailAuditEntriesRequest{}
	if err := stream.RecvMsg(req); err != nil {
		return err
	}
	return srv.(AuditServiceServer).TailAuditEntries(req, &auditTailServer{stream})
}

type auditTailServer struct {
	grpc.ServerStream
}

func (s *auditTailServer) Send(m *AuditEntry) error {
	return s.ServerStream.SendMsg(m)
}
