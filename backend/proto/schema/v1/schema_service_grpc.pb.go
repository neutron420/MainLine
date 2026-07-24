package schemav1

import (
	"context"

	"google.golang.org/grpc"
)

type SchemaServiceServer interface {
	IntrospectSchema(ctx context.Context, req *IntrospectSchemaRequest) (*IntrospectSchemaResponse, error)
	GetSchema(ctx context.Context, req *GetSchemaRequest) (*GetSchemaResponse, error)
	ListSchemas(ctx context.Context, req *ListSchemasRequest) (*ListSchemasResponse, error)

	ListSchemaVersions(ctx context.Context, req *ListSchemaVersionsRequest) (*ListSchemaVersionsResponse, error)
	GetSchemaVersion(ctx context.Context, req *GetSchemaVersionRequest) (*GetSchemaVersionResponse, error)
	CompareSchemaVersions(ctx context.Context, req *CompareSchemaVersionsRequest) (*CompareSchemaVersionsResponse, error)

	ListSchemaObjects(ctx context.Context, req *ListSchemaObjectsRequest) (*ListSchemaObjectsResponse, error)

	GetSchemaDiagram(ctx context.Context, req *GetSchemaDiagramRequest) (*GetSchemaDiagramResponse, error)
}

func RegisterSchemaServiceServer(s *grpc.Server, srv SchemaServiceServer) {
	s.RegisterService(&SchemaService_ServiceDesc, srv)
}

var SchemaService_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "schemahub.schema.v1.SchemaService",
	Methods: []grpc.MethodDesc{
		{MethodName: "IntrospectSchema", Handler: _SchemaService_IntrospectSchema_Handler},
		{MethodName: "GetSchema", Handler: _SchemaService_GetSchema_Handler},
		{MethodName: "ListSchemas", Handler: _SchemaService_ListSchemas_Handler},
		{MethodName: "ListSchemaVersions", Handler: _SchemaService_ListSchemaVersions_Handler},
		{MethodName: "GetSchemaVersion", Handler: _SchemaService_GetSchemaVersion_Handler},
		{MethodName: "CompareSchemaVersions", Handler: _SchemaService_CompareSchemaVersions_Handler},
		{MethodName: "ListSchemaObjects", Handler: _SchemaService_ListSchemaObjects_Handler},
		{MethodName: "GetSchemaDiagram", Handler: _SchemaService_GetSchemaDiagram_Handler},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "schema/v1/schema_service.proto",
}

func _SchemaService_IntrospectSchema_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, _ grpc.UnaryServerInterceptor) (interface{}, error) {
	req := &IntrospectSchemaRequest{}
	if err := dec(req); err != nil {
		return nil, err
	}
	return srv.(SchemaServiceServer).IntrospectSchema(ctx, req)
}

func _SchemaService_GetSchema_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, _ grpc.UnaryServerInterceptor) (interface{}, error) {
	req := &GetSchemaRequest{}
	if err := dec(req); err != nil {
		return nil, err
	}
	return srv.(SchemaServiceServer).GetSchema(ctx, req)
}

func _SchemaService_ListSchemas_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, _ grpc.UnaryServerInterceptor) (interface{}, error) {
	req := &ListSchemasRequest{}
	if err := dec(req); err != nil {
		return nil, err
	}
	return srv.(SchemaServiceServer).ListSchemas(ctx, req)
}

func _SchemaService_ListSchemaVersions_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, _ grpc.UnaryServerInterceptor) (interface{}, error) {
	req := &ListSchemaVersionsRequest{}
	if err := dec(req); err != nil {
		return nil, err
	}
	return srv.(SchemaServiceServer).ListSchemaVersions(ctx, req)
}

func _SchemaService_GetSchemaVersion_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, _ grpc.UnaryServerInterceptor) (interface{}, error) {
	req := &GetSchemaVersionRequest{}
	if err := dec(req); err != nil {
		return nil, err
	}
	return srv.(SchemaServiceServer).GetSchemaVersion(ctx, req)
}

func _SchemaService_CompareSchemaVersions_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, _ grpc.UnaryServerInterceptor) (interface{}, error) {
	req := &CompareSchemaVersionsRequest{}
	if err := dec(req); err != nil {
		return nil, err
	}
	return srv.(SchemaServiceServer).CompareSchemaVersions(ctx, req)
}

func _SchemaService_ListSchemaObjects_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, _ grpc.UnaryServerInterceptor) (interface{}, error) {
	req := &ListSchemaObjectsRequest{}
	if err := dec(req); err != nil {
		return nil, err
	}
	return srv.(SchemaServiceServer).ListSchemaObjects(ctx, req)
}

func _SchemaService_GetSchemaDiagram_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, _ grpc.UnaryServerInterceptor) (interface{}, error) {
	req := &GetSchemaDiagramRequest{}
	if err := dec(req); err != nil {
		return nil, err
	}
	return srv.(SchemaServiceServer).GetSchemaDiagram(ctx, req)
}

type UnimplementedSchemaServiceServer struct{}

func (UnimplementedSchemaServiceServer) IntrospectSchema(_ context.Context, _ *IntrospectSchemaRequest) (*IntrospectSchemaResponse, error) { return nil, nil }
func (UnimplementedSchemaServiceServer) GetSchema(_ context.Context, _ *GetSchemaRequest) (*GetSchemaResponse, error) { return nil, nil }
func (UnimplementedSchemaServiceServer) ListSchemas(_ context.Context, _ *ListSchemasRequest) (*ListSchemasResponse, error) { return nil, nil }
func (UnimplementedSchemaServiceServer) ListSchemaVersions(_ context.Context, _ *ListSchemaVersionsRequest) (*ListSchemaVersionsResponse, error) { return nil, nil }
func (UnimplementedSchemaServiceServer) GetSchemaVersion(_ context.Context, _ *GetSchemaVersionRequest) (*GetSchemaVersionResponse, error) { return nil, nil }
func (UnimplementedSchemaServiceServer) CompareSchemaVersions(_ context.Context, _ *CompareSchemaVersionsRequest) (*CompareSchemaVersionsResponse, error) { return nil, nil }
func (UnimplementedSchemaServiceServer) ListSchemaObjects(_ context.Context, _ *ListSchemaObjectsRequest) (*ListSchemaObjectsResponse, error) { return nil, nil }
func (UnimplementedSchemaServiceServer) GetSchemaDiagram(_ context.Context, _ *GetSchemaDiagramRequest) (*GetSchemaDiagramResponse, error) { return nil, nil }
