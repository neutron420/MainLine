package migrationv1

import (
	"context"

	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

type MigrationServiceServer interface {
	CreateMigration(ctx context.Context, req *CreateMigrationRequest) (*CreateMigrationResponse, error)
	GetMigration(ctx context.Context, req *GetMigrationRequest) (*GetMigrationResponse, error)
	ListMigrations(ctx context.Context, req *ListMigrationsRequest) (*ListMigrationsResponse, error)
	UpdateMigration(ctx context.Context, req *UpdateMigrationRequest) (*UpdateMigrationResponse, error)
	DeleteMigration(ctx context.Context, req *DeleteMigrationRequest) (*DeleteMigrationResponse, error)

	ExecuteMigration(ctx context.Context, req *ExecuteMigrationRequest) (*ExecuteMigrationResponse, error)
	WatchMigration(req *WatchMigrationRequest, stream MigrationService_WatchMigrationServer) error
	RollbackMigration(ctx context.Context, req *RollbackMigrationRequest) (*RollbackMigrationResponse, error)
	WatchRollback(req *WatchRollbackRequest, stream MigrationService_WatchRollbackServer) error

	ValidateMigration(ctx context.Context, req *ValidateMigrationRequest) (*ValidateMigrationResponse, error)
	DryRunMigration(ctx context.Context, req *DryRunMigrationRequest) (*DryRunMigrationResponse, error)

	GetMigrationRun(ctx context.Context, req *GetMigrationRunRequest) (*GetMigrationRunResponse, error)
	ListMigrationRuns(ctx context.Context, req *ListMigrationRunsRequest) (*ListMigrationRunsResponse, error)
	GetMigrationLogs(req *GetMigrationLogsRequest, stream MigrationService_GetMigrationLogsServer) error
}

type UnimplementedMigrationServiceServer struct{}

func (UnimplementedMigrationServiceServer) CreateMigration(ctx context.Context, req *CreateMigrationRequest) (*CreateMigrationResponse, error) {
	return nil, nil
}
func (UnimplementedMigrationServiceServer) GetMigration(ctx context.Context, req *GetMigrationRequest) (*GetMigrationResponse, error) {
	return nil, nil
}
func (UnimplementedMigrationServiceServer) ListMigrations(ctx context.Context, req *ListMigrationsRequest) (*ListMigrationsResponse, error) {
	return nil, nil
}
func (UnimplementedMigrationServiceServer) UpdateMigration(ctx context.Context, req *UpdateMigrationRequest) (*UpdateMigrationResponse, error) {
	return nil, nil
}
func (UnimplementedMigrationServiceServer) DeleteMigration(ctx context.Context, req *DeleteMigrationRequest) (*DeleteMigrationResponse, error) {
	return nil, nil
}
func (UnimplementedMigrationServiceServer) ExecuteMigration(ctx context.Context, req *ExecuteMigrationRequest) (*ExecuteMigrationResponse, error) {
	return nil, nil
}
func (UnimplementedMigrationServiceServer) WatchMigration(req *WatchMigrationRequest, stream MigrationService_WatchMigrationServer) error {
	return nil
}
func (UnimplementedMigrationServiceServer) RollbackMigration(ctx context.Context, req *RollbackMigrationRequest) (*RollbackMigrationResponse, error) {
	return nil, nil
}
func (UnimplementedMigrationServiceServer) WatchRollback(req *WatchRollbackRequest, stream MigrationService_WatchRollbackServer) error {
	return nil
}
func (UnimplementedMigrationServiceServer) ValidateMigration(ctx context.Context, req *ValidateMigrationRequest) (*ValidateMigrationResponse, error) {
	return nil, nil
}
func (UnimplementedMigrationServiceServer) DryRunMigration(ctx context.Context, req *DryRunMigrationRequest) (*DryRunMigrationResponse, error) {
	return nil, nil
}
func (UnimplementedMigrationServiceServer) GetMigrationRun(ctx context.Context, req *GetMigrationRunRequest) (*GetMigrationRunResponse, error) {
	return nil, nil
}
func (UnimplementedMigrationServiceServer) ListMigrationRuns(ctx context.Context, req *ListMigrationRunsRequest) (*ListMigrationRunsResponse, error) {
	return nil, nil
}
func (UnimplementedMigrationServiceServer) GetMigrationLogs(req *GetMigrationLogsRequest, stream MigrationService_GetMigrationLogsServer) error {
	return nil
}

func RegisterMigrationServiceServer(s *grpc.Server, srv MigrationServiceServer) {
	s.RegisterService(&MigrationService_ServiceDesc, srv)
}

type MigrationService_WatchMigrationServer interface {
	Send(*MigrationStatusMessage) error
	grpc.ServerStream
}

type MigrationService_WatchRollbackServer interface {
	Send(*MigrationStatusMessage) error
	grpc.ServerStream
}

type MigrationService_GetMigrationLogsServer interface {
	Send(*MigrationLogEntry) error
	grpc.ServerStream
}

var MigrationService_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "schemahub.migration.v1.MigrationService",
	Methods: []grpc.MethodDesc{
		{MethodName: "CreateMigration", Handler: _MigrationService_CreateMigration_Handler},
		{MethodName: "GetMigration", Handler: _MigrationService_GetMigration_Handler},
		{MethodName: "ListMigrations", Handler: _MigrationService_ListMigrations_Handler},
		{MethodName: "UpdateMigration", Handler: _MigrationService_UpdateMigration_Handler},
		{MethodName: "DeleteMigration", Handler: _MigrationService_DeleteMigration_Handler},
		{MethodName: "ExecuteMigration", Handler: _MigrationService_ExecuteMigration_Handler},
		{MethodName: "RollbackMigration", Handler: _MigrationService_RollbackMigration_Handler},
		{MethodName: "ValidateMigration", Handler: _MigrationService_ValidateMigration_Handler},
		{MethodName: "DryRunMigration", Handler: _MigrationService_DryRunMigration_Handler},
		{MethodName: "GetMigrationRun", Handler: _MigrationService_GetMigrationRun_Handler},
		{MethodName: "ListMigrationRuns", Handler: _MigrationService_ListMigrationRuns_Handler},
	},
	Streams: []grpc.StreamDesc{
		{
			StreamName:    "WatchMigration",
			Handler:       _MigrationService_WatchMigration_Handler,
			ServerStreams: true,
		},
		{
			StreamName:    "WatchRollback",
			Handler:       _MigrationService_WatchRollback_Handler,
			ServerStreams: true,
		},
		{
			StreamName:    "GetMigrationLogs",
			Handler:       _MigrationService_GetMigrationLogs_Handler,
			ServerStreams: true,
		},
	},
	Metadata: "migration/v1/migration_service.proto",
}

func _MigrationService_CreateMigration_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, _ grpc.UnaryServerInterceptor) (interface{}, error) {
	req := &CreateMigrationRequest{}
	if err := dec(req); err != nil {
		return nil, err
	}
	return srv.(MigrationServiceServer).CreateMigration(ctx, req)
}

func _MigrationService_GetMigration_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, _ grpc.UnaryServerInterceptor) (interface{}, error) {
	req := &GetMigrationRequest{}
	if err := dec(req); err != nil {
		return nil, err
	}
	return srv.(MigrationServiceServer).GetMigration(ctx, req)
}

func _MigrationService_ListMigrations_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, _ grpc.UnaryServerInterceptor) (interface{}, error) {
	req := &ListMigrationsRequest{}
	if err := dec(req); err != nil {
		return nil, err
	}
	return srv.(MigrationServiceServer).ListMigrations(ctx, req)
}

func _MigrationService_UpdateMigration_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, _ grpc.UnaryServerInterceptor) (interface{}, error) {
	req := &UpdateMigrationRequest{}
	if err := dec(req); err != nil {
		return nil, err
	}
	return srv.(MigrationServiceServer).UpdateMigration(ctx, req)
}

func _MigrationService_DeleteMigration_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, _ grpc.UnaryServerInterceptor) (interface{}, error) {
	req := &DeleteMigrationRequest{}
	if err := dec(req); err != nil {
		return nil, err
	}
	return srv.(MigrationServiceServer).DeleteMigration(ctx, req)
}

func _MigrationService_ExecuteMigration_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, _ grpc.UnaryServerInterceptor) (interface{}, error) {
	req := &ExecuteMigrationRequest{}
	if err := dec(req); err != nil {
		return nil, err
	}
	return srv.(MigrationServiceServer).ExecuteMigration(ctx, req)
}

func _MigrationService_RollbackMigration_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, _ grpc.UnaryServerInterceptor) (interface{}, error) {
	req := &RollbackMigrationRequest{}
	if err := dec(req); err != nil {
		return nil, err
	}
	return srv.(MigrationServiceServer).RollbackMigration(ctx, req)
}

func _MigrationService_ValidateMigration_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, _ grpc.UnaryServerInterceptor) (interface{}, error) {
	req := &ValidateMigrationRequest{}
	if err := dec(req); err != nil {
		return nil, err
	}
	return srv.(MigrationServiceServer).ValidateMigration(ctx, req)
}

func _MigrationService_DryRunMigration_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, _ grpc.UnaryServerInterceptor) (interface{}, error) {
	req := &DryRunMigrationRequest{}
	if err := dec(req); err != nil {
		return nil, err
	}
	return srv.(MigrationServiceServer).DryRunMigration(ctx, req)
}

func _MigrationService_GetMigrationRun_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, _ grpc.UnaryServerInterceptor) (interface{}, error) {
	req := &GetMigrationRunRequest{}
	if err := dec(req); err != nil {
		return nil, err
	}
	return srv.(MigrationServiceServer).GetMigrationRun(ctx, req)
}

func _MigrationService_ListMigrationRuns_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, _ grpc.UnaryServerInterceptor) (interface{}, error) {
	req := &ListMigrationRunsRequest{}
	if err := dec(req); err != nil {
		return nil, err
	}
	return srv.(MigrationServiceServer).ListMigrationRuns(ctx, req)
}

func _MigrationService_WatchMigration_Handler(srv interface{}, stream grpc.ServerStream) error {
	req := &WatchMigrationRequest{}
	if err := stream.RecvMsg(req); err != nil {
		return err
	}
	// Set content-type to allow gRPC-Web streaming
	_ = metadata.New(map[string]string{})
	return srv.(MigrationServiceServer).WatchMigration(req, &migrationWatchServer{stream})
}

func _MigrationService_WatchRollback_Handler(srv interface{}, stream grpc.ServerStream) error {
	req := &WatchRollbackRequest{}
	if err := stream.RecvMsg(req); err != nil {
		return err
	}
	return srv.(MigrationServiceServer).WatchRollback(req, &migrationWatchRollbackServer{stream})
}

func _MigrationService_GetMigrationLogs_Handler(srv interface{}, stream grpc.ServerStream) error {
	req := &GetMigrationLogsRequest{}
	if err := stream.RecvMsg(req); err != nil {
		return err
	}
	return srv.(MigrationServiceServer).GetMigrationLogs(req, &migrationLogsServer{stream})
}

type migrationWatchServer struct {
	grpc.ServerStream
}

func (s *migrationWatchServer) Send(m *MigrationStatusMessage) error {
	return s.ServerStream.SendMsg(m)
}

type migrationWatchRollbackServer struct {
	grpc.ServerStream
}

func (s *migrationWatchRollbackServer) Send(m *MigrationStatusMessage) error {
	return s.ServerStream.SendMsg(m)
}

type migrationLogsServer struct {
	grpc.ServerStream
}

func (s *migrationLogsServer) Send(m *MigrationLogEntry) error {
	return s.ServerStream.SendMsg(m)
}
