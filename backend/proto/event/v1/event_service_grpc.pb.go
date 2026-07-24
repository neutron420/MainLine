package eventv1

import (
	"context"

	"google.golang.org/grpc"
)

type EventServiceServer interface {
	Subscribe(req *SubscribeRequest, stream EventService_SubscribeServer) error
	AcknowledgeEvent(ctx context.Context, req *AcknowledgeEventRequest) (*AcknowledgeEventResponse, error)
	Heartbeat(ctx context.Context, req *HeartbeatRequest) (*HeartbeatResponse, error)
}

type UnimplementedEventServiceServer struct{}

func (UnimplementedEventServiceServer) Subscribe(req *SubscribeRequest, stream EventService_SubscribeServer) error {
	return nil
}

func (UnimplementedEventServiceServer) AcknowledgeEvent(ctx context.Context, req *AcknowledgeEventRequest) (*AcknowledgeEventResponse, error) {
	return &AcknowledgeEventResponse{}, nil
}

func (UnimplementedEventServiceServer) Heartbeat(ctx context.Context, req *HeartbeatRequest) (*HeartbeatResponse, error) {
	return &HeartbeatResponse{}, nil
}

func RegisterEventServiceServer(s *grpc.Server, srv EventServiceServer) {
	s.RegisterService(&EventService_ServiceDesc, srv)
}

type EventService_SubscribeServer interface {
	Send(*SchemaEvent) error
	grpc.ServerStream
}

var EventService_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "schemahub.event.v1.EventService",
	Methods: []grpc.MethodDesc{
		{
			MethodName: "AcknowledgeEvent",
			Handler:    _EventService_AcknowledgeEvent_Handler,
		},
		{
			MethodName: "Heartbeat",
			Handler:    _EventService_Heartbeat_Handler,
		},
	},
	Streams: []grpc.StreamDesc{
		{
			StreamName:    "Subscribe",
			Handler:       _EventService_Subscribe_Handler,
			ServerStreams: true,
		},
	},
	Metadata: "event/v1/event_service.proto",
}

func _EventService_AcknowledgeEvent_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, _ grpc.UnaryServerInterceptor) (interface{}, error) {
	req := &AcknowledgeEventRequest{}
	if err := dec(req); err != nil {
		return nil, err
	}
	return srv.(EventServiceServer).AcknowledgeEvent(ctx, req)
}

func _EventService_Heartbeat_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, _ grpc.UnaryServerInterceptor) (interface{}, error) {
	req := &HeartbeatRequest{}
	if err := dec(req); err != nil {
		return nil, err
	}
	return srv.(EventServiceServer).Heartbeat(ctx, req)
}

func _EventService_Subscribe_Handler(srv interface{}, stream grpc.ServerStream) error {
	req := &SubscribeRequest{}
	if err := stream.RecvMsg(req); err != nil {
		return err
	}
	return srv.(EventServiceServer).Subscribe(req, &eventSubscribeServer{stream})
}

type eventSubscribeServer struct {
	grpc.ServerStream
}

func (s *eventSubscribeServer) Send(m *SchemaEvent) error {
	return s.ServerStream.SendMsg(m)
}
