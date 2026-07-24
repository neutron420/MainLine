package projectv1

import (
	"context"

	"google.golang.org/grpc"
)

type ProjectServiceServer interface {
	CreateProject(ctx context.Context, req *CreateProjectRequest) (*CreateProjectResponse, error)
	GetProject(ctx context.Context, req *GetProjectRequest) (*GetProjectResponse, error)
	ListProjects(ctx context.Context, req *ListProjectsRequest) (*ListProjectsResponse, error)
	UpdateProject(ctx context.Context, req *UpdateProjectRequest) (*UpdateProjectResponse, error)
	DeleteProject(ctx context.Context, req *DeleteProjectRequest) (*DeleteProjectResponse, error)

	AddMember(ctx context.Context, req *AddMemberRequest) (*AddMemberResponse, error)
	RemoveMember(ctx context.Context, req *RemoveMemberRequest) (*RemoveMemberResponse, error)
	UpdateMemberRole(ctx context.Context, req *UpdateMemberRoleRequest) (*UpdateMemberRoleResponse, error)
	ListMembers(ctx context.Context, req *ListMembersRequest) (*ListMembersResponse, error)

	CreateConnection(ctx context.Context, req *CreateConnectionRequest) (*CreateConnectionResponse, error)
	GetConnection(ctx context.Context, req *GetConnectionRequest) (*GetConnectionResponse, error)
	ListConnections(ctx context.Context, req *ListConnectionsRequest) (*ListConnectionsResponse, error)
	UpdateConnection(ctx context.Context, req *UpdateConnectionRequest) (*UpdateConnectionResponse, error)
	DeleteConnection(ctx context.Context, req *DeleteConnectionRequest) (*DeleteConnectionResponse, error)
	TestConnection(ctx context.Context, req *TestConnectionRequest) (*TestConnectionResponse, error)
}

func RegisterProjectServiceServer(s *grpc.Server, srv ProjectServiceServer) {
	s.RegisterService(&ProjectService_ServiceDesc, srv)
}

var ProjectService_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "schemahub.project.v1.ProjectService",
	Methods: []grpc.MethodDesc{
		{MethodName: "CreateProject", Handler: _ProjectService_CreateProject_Handler},
		{MethodName: "GetProject", Handler: _ProjectService_GetProject_Handler},
		{MethodName: "ListProjects", Handler: _ProjectService_ListProjects_Handler},
		{MethodName: "UpdateProject", Handler: _ProjectService_UpdateProject_Handler},
		{MethodName: "DeleteProject", Handler: _ProjectService_DeleteProject_Handler},
		{MethodName: "AddMember", Handler: _ProjectService_AddMember_Handler},
		{MethodName: "RemoveMember", Handler: _ProjectService_RemoveMember_Handler},
		{MethodName: "UpdateMemberRole", Handler: _ProjectService_UpdateMemberRole_Handler},
		{MethodName: "ListMembers", Handler: _ProjectService_ListMembers_Handler},
		{MethodName: "CreateConnection", Handler: _ProjectService_CreateConnection_Handler},
		{MethodName: "GetConnection", Handler: _ProjectService_GetConnection_Handler},
		{MethodName: "ListConnections", Handler: _ProjectService_ListConnections_Handler},
		{MethodName: "UpdateConnection", Handler: _ProjectService_UpdateConnection_Handler},
		{MethodName: "DeleteConnection", Handler: _ProjectService_DeleteConnection_Handler},
		{MethodName: "TestConnection", Handler: _ProjectService_TestConnection_Handler},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "project/v1/project_service.proto",
}

func _ProjectService_CreateProject_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, _ grpc.UnaryServerInterceptor) (interface{}, error) {
	req := &CreateProjectRequest{}
	if err := dec(req); err != nil {
		return nil, err
	}
	return srv.(ProjectServiceServer).CreateProject(ctx, req)
}

func _ProjectService_GetProject_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, _ grpc.UnaryServerInterceptor) (interface{}, error) {
	req := &GetProjectRequest{}
	if err := dec(req); err != nil {
		return nil, err
	}
	return srv.(ProjectServiceServer).GetProject(ctx, req)
}

func _ProjectService_ListProjects_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, _ grpc.UnaryServerInterceptor) (interface{}, error) {
	req := &ListProjectsRequest{}
	if err := dec(req); err != nil {
		return nil, err
	}
	return srv.(ProjectServiceServer).ListProjects(ctx, req)
}

func _ProjectService_UpdateProject_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, _ grpc.UnaryServerInterceptor) (interface{}, error) {
	req := &UpdateProjectRequest{}
	if err := dec(req); err != nil {
		return nil, err
	}
	return srv.(ProjectServiceServer).UpdateProject(ctx, req)
}

func _ProjectService_DeleteProject_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, _ grpc.UnaryServerInterceptor) (interface{}, error) {
	req := &DeleteProjectRequest{}
	if err := dec(req); err != nil {
		return nil, err
	}
	return srv.(ProjectServiceServer).DeleteProject(ctx, req)
}

func _ProjectService_AddMember_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, _ grpc.UnaryServerInterceptor) (interface{}, error) {
	req := &AddMemberRequest{}
	if err := dec(req); err != nil {
		return nil, err
	}
	return srv.(ProjectServiceServer).AddMember(ctx, req)
}

func _ProjectService_RemoveMember_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, _ grpc.UnaryServerInterceptor) (interface{}, error) {
	req := &RemoveMemberRequest{}
	if err := dec(req); err != nil {
		return nil, err
	}
	return srv.(ProjectServiceServer).RemoveMember(ctx, req)
}

func _ProjectService_UpdateMemberRole_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, _ grpc.UnaryServerInterceptor) (interface{}, error) {
	req := &UpdateMemberRoleRequest{}
	if err := dec(req); err != nil {
		return nil, err
	}
	return srv.(ProjectServiceServer).UpdateMemberRole(ctx, req)
}

func _ProjectService_ListMembers_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, _ grpc.UnaryServerInterceptor) (interface{}, error) {
	req := &ListMembersRequest{}
	if err := dec(req); err != nil {
		return nil, err
	}
	return srv.(ProjectServiceServer).ListMembers(ctx, req)
}

func _ProjectService_CreateConnection_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, _ grpc.UnaryServerInterceptor) (interface{}, error) {
	req := &CreateConnectionRequest{}
	if err := dec(req); err != nil {
		return nil, err
	}
	return srv.(ProjectServiceServer).CreateConnection(ctx, req)
}

func _ProjectService_GetConnection_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, _ grpc.UnaryServerInterceptor) (interface{}, error) {
	req := &GetConnectionRequest{}
	if err := dec(req); err != nil {
		return nil, err
	}
	return srv.(ProjectServiceServer).GetConnection(ctx, req)
}

func _ProjectService_ListConnections_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, _ grpc.UnaryServerInterceptor) (interface{}, error) {
	req := &ListConnectionsRequest{}
	if err := dec(req); err != nil {
		return nil, err
	}
	return srv.(ProjectServiceServer).ListConnections(ctx, req)
}

func _ProjectService_UpdateConnection_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, _ grpc.UnaryServerInterceptor) (interface{}, error) {
	req := &UpdateConnectionRequest{}
	if err := dec(req); err != nil {
		return nil, err
	}
	return srv.(ProjectServiceServer).UpdateConnection(ctx, req)
}

func _ProjectService_DeleteConnection_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, _ grpc.UnaryServerInterceptor) (interface{}, error) {
	req := &DeleteConnectionRequest{}
	if err := dec(req); err != nil {
		return nil, err
	}
	return srv.(ProjectServiceServer).DeleteConnection(ctx, req)
}

func _ProjectService_TestConnection_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, _ grpc.UnaryServerInterceptor) (interface{}, error) {
	req := &TestConnectionRequest{}
	if err := dec(req); err != nil {
		return nil, err
	}
	return srv.(ProjectServiceServer).TestConnection(ctx, req)
}

type UnimplementedProjectServiceServer struct{}

func (UnimplementedProjectServiceServer) CreateProject(_ context.Context, _ *CreateProjectRequest) (*CreateProjectResponse, error) { return nil, nil }
func (UnimplementedProjectServiceServer) GetProject(_ context.Context, _ *GetProjectRequest) (*GetProjectResponse, error) { return nil, nil }
func (UnimplementedProjectServiceServer) ListProjects(_ context.Context, _ *ListProjectsRequest) (*ListProjectsResponse, error) { return nil, nil }
func (UnimplementedProjectServiceServer) UpdateProject(_ context.Context, _ *UpdateProjectRequest) (*UpdateProjectResponse, error) { return nil, nil }
func (UnimplementedProjectServiceServer) DeleteProject(_ context.Context, _ *DeleteProjectRequest) (*DeleteProjectResponse, error) { return nil, nil }
func (UnimplementedProjectServiceServer) AddMember(_ context.Context, _ *AddMemberRequest) (*AddMemberResponse, error) { return nil, nil }
func (UnimplementedProjectServiceServer) RemoveMember(_ context.Context, _ *RemoveMemberRequest) (*RemoveMemberResponse, error) { return nil, nil }
func (UnimplementedProjectServiceServer) UpdateMemberRole(_ context.Context, _ *UpdateMemberRoleRequest) (*UpdateMemberRoleResponse, error) { return nil, nil }
func (UnimplementedProjectServiceServer) ListMembers(_ context.Context, _ *ListMembersRequest) (*ListMembersResponse, error) { return nil, nil }
func (UnimplementedProjectServiceServer) CreateConnection(_ context.Context, _ *CreateConnectionRequest) (*CreateConnectionResponse, error) { return nil, nil }
func (UnimplementedProjectServiceServer) GetConnection(_ context.Context, _ *GetConnectionRequest) (*GetConnectionResponse, error) { return nil, nil }
func (UnimplementedProjectServiceServer) ListConnections(_ context.Context, _ *ListConnectionsRequest) (*ListConnectionsResponse, error) { return nil, nil }
func (UnimplementedProjectServiceServer) UpdateConnection(_ context.Context, _ *UpdateConnectionRequest) (*UpdateConnectionResponse, error) { return nil, nil }
func (UnimplementedProjectServiceServer) DeleteConnection(_ context.Context, _ *DeleteConnectionRequest) (*DeleteConnectionResponse, error) { return nil, nil }
func (UnimplementedProjectServiceServer) TestConnection(_ context.Context, _ *TestConnectionRequest) (*TestConnectionResponse, error) { return nil, nil }
