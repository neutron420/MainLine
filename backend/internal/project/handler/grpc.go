package handler

import (
	"context"
	"time"

	"github.com/schemahub/backend/internal/pkg/errors"
	"github.com/schemahub/backend/internal/pkg/interceptor"
	"github.com/schemahub/backend/internal/project/domain"
	projectv1 "github.com/schemahub/backend/proto/project/v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type ProjectHandler struct {
	projectv1.UnimplementedProjectServiceServer
	svc     *domain.ProjectService
	connSvc *domain.ConnectionService
}

func NewProjectHandler(svc *domain.ProjectService, connSvc *domain.ConnectionService) *ProjectHandler {
	return &ProjectHandler{svc: svc, connSvc: connSvc}
}

func userIDFromContext(ctx context.Context) string {
	id, _ := interceptor.UserIDFromContext(ctx)
	return id
}

func (h *ProjectHandler) CreateProject(ctx context.Context, req *projectv1.CreateProjectRequest) (*projectv1.CreateProjectResponse, error) {
	userID := userIDFromContext(ctx)

	p, err := h.svc.Create(ctx, req.Name, req.Description, req.Visibility, userID)
	if err != nil {
		return nil, mapProjectError(err)
	}

	return &projectv1.CreateProjectResponse{Project: toProtoProject(p)}, nil
}

func (h *ProjectHandler) GetProject(ctx context.Context, req *projectv1.GetProjectRequest) (*projectv1.GetProjectResponse, error) {
	p, err := h.svc.GetByID(ctx, req.Id)
	if err != nil {
		return nil, mapProjectError(err)
	}
	return &projectv1.GetProjectResponse{Project: toProtoProject(p)}, nil
}

func (h *ProjectHandler) ListProjects(ctx context.Context, req *projectv1.ListProjectsRequest) (*projectv1.ListProjectsResponse, error) {
	userID := userIDFromContext(ctx)

	projects, nextCursor, total, err := h.svc.List(ctx, userID, req.Cursor, req.PageSize)
	if err != nil {
		return nil, mapProjectError(err)
	}

	var protoProjects []*projectv1.Project
	for _, p := range projects {
		protoProjects = append(protoProjects, toProtoProject(p))
	}

	return &projectv1.ListProjectsResponse{
		Projects:   protoProjects,
		NextCursor: nextCursor,
		TotalCount: total,
	}, nil
}

func (h *ProjectHandler) UpdateProject(ctx context.Context, req *projectv1.UpdateProjectRequest) (*projectv1.UpdateProjectResponse, error) {
	userID := userIDFromContext(ctx)

	p, err := h.svc.Update(ctx, req.Id, req.Name, req.Description, req.Visibility, userID)
	if err != nil {
		return nil, mapProjectError(err)
	}

	return &projectv1.UpdateProjectResponse{Project: toProtoProject(p)}, nil
}

func (h *ProjectHandler) DeleteProject(ctx context.Context, req *projectv1.DeleteProjectRequest) (*projectv1.DeleteProjectResponse, error) {
	userID := userIDFromContext(ctx)

	if err := h.svc.Delete(ctx, req.Id, userID); err != nil {
		return nil, mapProjectError(err)
	}

	return &projectv1.DeleteProjectResponse{}, nil
}

func (h *ProjectHandler) AddMember(ctx context.Context, req *projectv1.AddMemberRequest) (*projectv1.AddMemberResponse, error) {
	userID := userIDFromContext(ctx)

	if err := h.svc.AddMember(ctx, req.ProjectId, req.UserId, req.Role, userID); err != nil {
		return nil, mapProjectError(err)
	}

	return &projectv1.AddMemberResponse{}, nil
}

func (h *ProjectHandler) RemoveMember(ctx context.Context, req *projectv1.RemoveMemberRequest) (*projectv1.RemoveMemberResponse, error) {
	userID := userIDFromContext(ctx)

	if err := h.svc.RemoveMember(ctx, req.ProjectId, req.UserId, userID); err != nil {
		return nil, mapProjectError(err)
	}

	return &projectv1.RemoveMemberResponse{}, nil
}

func (h *ProjectHandler) UpdateMemberRole(ctx context.Context, req *projectv1.UpdateMemberRoleRequest) (*projectv1.UpdateMemberRoleResponse, error) {
	userID := userIDFromContext(ctx)

	if err := h.svc.UpdateMemberRole(ctx, req.ProjectId, req.UserId, req.Role, userID); err != nil {
		return nil, mapProjectError(err)
	}

	return &projectv1.UpdateMemberRoleResponse{}, nil
}

func (h *ProjectHandler) ListMembers(ctx context.Context, req *projectv1.ListMembersRequest) (*projectv1.ListMembersResponse, error) {
	userID := userIDFromContext(ctx)

	members, nextCursor, total, err := h.svc.ListMembers(ctx, req.ProjectId, req.Cursor, req.PageSize, userID)
	if err != nil {
		return nil, mapProjectError(err)
	}

	var protoMembers []*projectv1.ProjectMember
	for _, m := range members {
		protoMembers = append(protoMembers, toProtoMember(m))
	}

	return &projectv1.ListMembersResponse{
		Members:    protoMembers,
		NextCursor: nextCursor,
		TotalCount: total,
	}, nil
}

func toProtoProject(p *domain.Project) *projectv1.Project {
	return &projectv1.Project{
		Id:          p.ID,
		Name:        p.Name,
		Slug:        p.Slug,
		Description: p.Description,
		Visibility:  string(p.Visibility),
		CreatedBy:   p.CreatedBy,
		CreatedAt:   p.CreatedAt.Format(time.RFC3339),
		UpdatedAt:   p.UpdatedAt.Format(time.RFC3339),
	}
}

func toProtoMember(m *domain.ProjectMember) *projectv1.ProjectMember {
	return &projectv1.ProjectMember{
		UserId:   m.UserID,
		Role:     string(m.Role),
		JoinedAt: optionalTime(m.JoinedAt),
	}
}

func optionalTime(t *time.Time) string {
	if t == nil {
		return ""
	}
	return t.Format(time.RFC3339)
}

func mapProjectError(err error) error {
	switch e := err.(type) {
	case domain.ErrProjectNotFound:
		return status.Error(codes.NotFound, e.Error())
	case domain.ErrProjectSlugConflict:
		return status.Error(codes.AlreadyExists, e.Error())
	case domain.ErrMemberNotFound:
		return status.Error(codes.NotFound, e.Error())
	case domain.ErrLastOwner:
		return status.Error(codes.FailedPrecondition, e.Error())
	}

	if err.Error() == "permission denied" {
		return status.Error(codes.PermissionDenied, err.Error())
	}

	return errors.ToGRPC(err)
}

// â”€â”€ Connection Handlers â”€â”€

func toProtoConn(c *domain.Connection) *projectv1.Connection {
	return &projectv1.Connection{
		Id:               c.ID,
		ProjectId:        c.ProjectID,
		Name:             c.Name,
		Host:             c.Host,
		Port:             int32(c.Port),
		DatabaseName:     c.DatabaseName,
		Username:         c.Username,
		SslMode:          string(c.SSLMode),
		ConnectionStatus: string(c.ConnectionStatus),
		LastConnectedAt:  optionalTime(c.LastConnectedAt),
		CreatedBy:        c.CreatedBy,
		CreatedAt:        c.CreatedAt.Format(time.RFC3339),
		UpdatedAt:        c.UpdatedAt.Format(time.RFC3339),
	}
}

func (h *ProjectHandler) CreateConnection(ctx context.Context, req *projectv1.CreateConnectionRequest) (*projectv1.CreateConnectionResponse, error) {
	userID := userIDFromContext(ctx)

	c := &domain.Connection{
		ProjectID:    req.ProjectId,
		Name:         req.Name,
		Host:         req.Host,
		Port:         int(req.Port),
		DatabaseName: req.DatabaseName,
		Username:     req.Username,
		SSLMode:      domain.SSLMode(req.SslMode),
		CreatedBy:    userID,
	}

	created, err := h.connSvc.Create(ctx, c, req.Password)
	if err != nil {
		return nil, mapConnectionError(err)
	}
	return &projectv1.CreateConnectionResponse{Connection: toProtoConn(created)}, nil
}

func (h *ProjectHandler) GetConnection(ctx context.Context, req *projectv1.GetConnectionRequest) (*projectv1.GetConnectionResponse, error) {
	c, err := h.connSvc.GetByID(ctx, req.Id)
	if err != nil {
		return nil, mapConnectionError(err)
	}
	return &projectv1.GetConnectionResponse{Connection: toProtoConn(c)}, nil
}

func (h *ProjectHandler) ListConnections(ctx context.Context, req *projectv1.ListConnectionsRequest) (*projectv1.ListConnectionsResponse, error) {
	conns, nextCursor, total, err := h.connSvc.List(ctx, req.ProjectId, req.Cursor, req.PageSize)
	if err != nil {
		return nil, mapConnectionError(err)
	}

	var proto []*projectv1.Connection
	for _, c := range conns {
		proto = append(proto, toProtoConn(c))
	}
	return &projectv1.ListConnectionsResponse{Connections: proto, NextCursor: nextCursor, TotalCount: total}, nil
}

func (h *ProjectHandler) UpdateConnection(ctx context.Context, req *projectv1.UpdateConnectionRequest) (*projectv1.UpdateConnectionResponse, error) {
	c, err := h.connSvc.Update(ctx, req.Id, req.Name, req.Host, req.Port, req.DatabaseName, req.Username, req.Password, req.SslMode)
	if err != nil {
		return nil, mapConnectionError(err)
	}
	return &projectv1.UpdateConnectionResponse{Connection: toProtoConn(c)}, nil
}

func (h *ProjectHandler) DeleteConnection(ctx context.Context, req *projectv1.DeleteConnectionRequest) (*projectv1.DeleteConnectionResponse, error) {
	if err := h.connSvc.Delete(ctx, req.Id); err != nil {
		return nil, mapConnectionError(err)
	}
	return &projectv1.DeleteConnectionResponse{}, nil
}

func (h *ProjectHandler) TestConnection(ctx context.Context, req *projectv1.TestConnectionRequest) (*projectv1.TestConnectionResponse, error) {
	success, latency, version, dbName, err := h.connSvc.Test(ctx, req.ConnectionId)
	if err != nil {
		return &projectv1.TestConnectionResponse{
			Success:   success,
			LatencyMs: latency,
			Error:     err.Error(),
		}, nil
	}
	return &projectv1.TestConnectionResponse{
		Success:       success,
		LatencyMs:     latency,
		ServerVersion: version,
		DatabaseName:  dbName,
	}, nil
}

func mapConnectionError(err error) error {
	if err == nil {
		return nil
	}
	return errors.ToGRPC(err)
}
