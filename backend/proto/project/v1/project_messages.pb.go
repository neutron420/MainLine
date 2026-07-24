package projectv1

import (
	"fmt"
)

type Project struct {
	ID           string `json:"id"`
	Name         string `json:"name"`
	Slug         string `json:"slug"`
	Description  string `json:"description"`
	Visibility   string `json:"visibility"`
	MemberCount  int32  `json:"member_count"`
	CreatedBy    string `json:"created_by"`
	CreatedAt    string `json:"created_at"`
	UpdatedAt    string `json:"updated_at"`
}

type ProjectMember struct {
	UserID      string `json:"user_id"`
	Email       string `json:"email"`
	DisplayName string `json:"display_name"`
	Role        string `json:"role"`
	JoinedAt    string `json:"joined_at"`
}

type CreateProjectRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Visibility  string `json:"visibility"`
}

type CreateProjectResponse struct {
	Project *Project `json:"project"`
}

type GetProjectRequest struct {
	ID string `json:"id"`
}

type GetProjectResponse struct {
	Project *Project `json:"project"`
}

type ListProjectsRequest struct {
	Cursor   string `json:"cursor"`
	PageSize int32  `json:"page_size"`
}

type ListProjectsResponse struct {
	Projects    []*Project `json:"projects"`
	NextCursor  string     `json:"next_cursor"`
	TotalCount  int32      `json:"total_count"`
}

type UpdateProjectRequest struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Visibility  string `json:"visibility"`
}

type UpdateProjectResponse struct {
	Project *Project `json:"project"`
}

type DeleteProjectRequest struct {
	ID string `json:"id"`
}

type DeleteProjectResponse struct{}

type AddMemberRequest struct {
	ProjectID string `json:"project_id"`
	UserID    string `json:"user_id"`
	Role      string `json:"role"`
}

type AddMemberResponse struct{}

type RemoveMemberRequest struct {
	ProjectID string `json:"project_id"`
	UserID    string `json:"user_id"`
}

type RemoveMemberResponse struct{}

type UpdateMemberRoleRequest struct {
	ProjectID string `json:"project_id"`
	UserID    string `json:"user_id"`
	Role      string `json:"role"`
}

type UpdateMemberRoleResponse struct{}

type ListMembersRequest struct {
	ProjectID string `json:"project_id"`
	Cursor    string `json:"cursor"`
	PageSize  int32  `json:"page_size"`
}

type ListMembersResponse struct {
	Members    []*ProjectMember `json:"members"`
	NextCursor string          `json:"next_cursor"`
	TotalCount int32           `json:"total_count"`
}

func (r *CreateProjectRequest) Validate() error {
	if r.Name == "" {
		return fmt.Errorf("project name is required")
	}
	if len(r.Name) > 200 {
		return fmt.Errorf("project name must be 200 characters or less")
	}
	if r.Visibility == "" {
		r.Visibility = "private"
	}
	return nil
}

func (r *UpdateProjectRequest) Validate() error {
	if r.ID == "" {
		return fmt.Errorf("project id is required")
	}
	if r.Name != "" && len(r.Name) > 200 {
		return fmt.Errorf("project name must be 200 characters or less")
	}
	return nil
}

func (r *AddMemberRequest) Validate() error {
	if r.ProjectID == "" {
		return fmt.Errorf("project id is required")
	}
	if r.UserID == "" {
		return fmt.Errorf("user id is required")
	}
	if r.Role == "" {
		r.Role = "member"
	}
	return nil
}

func (r *UpdateMemberRoleRequest) Validate() error {
	if r.ProjectID == "" {
		return fmt.Errorf("project id is required")
	}
	if r.UserID == "" {
		return fmt.Errorf("user id is required")
	}
	if r.Role == "" {
		return fmt.Errorf("role is required")
	}
	return nil
}

// ── Connection Messages ──

type Connection struct {
	ID               string `json:"id"`
	ProjectID        string `json:"project_id"`
	Name             string `json:"name"`
	Host             string `json:"host"`
	Port             int32  `json:"port"`
	DatabaseName     string `json:"database_name"`
	Username         string `json:"username"`
	SSLMode          string `json:"ssl_mode"`
	ConnectionStatus string `json:"connection_status"`
	LastConnectedAt  string `json:"last_connected_at"`
	CreatedBy        string `json:"created_by"`
	CreatedAt        string `json:"created_at"`
	UpdatedAt        string `json:"updated_at"`
}

type CreateConnectionRequest struct {
	ProjectID    string `json:"project_id"`
	Name         string `json:"name"`
	Host         string `json:"host"`
	Port         int32  `json:"port"`
	DatabaseName string `json:"database_name"`
	Username     string `json:"username"`
	Password     string `json:"password"`
	SSLMode      string `json:"ssl_mode"`
}

type CreateConnectionResponse struct {
	Connection *Connection `json:"connection"`
}

type GetConnectionRequest struct {
	ID string `json:"id"`
}

type GetConnectionResponse struct {
	Connection *Connection `json:"connection"`
}

type ListConnectionsRequest struct {
	ProjectID string `json:"project_id"`
	Cursor    string `json:"cursor"`
	PageSize  int32  `json:"page_size"`
}

type ListConnectionsResponse struct {
	Connections []*Connection `json:"connections"`
	NextCursor  string        `json:"next_cursor"`
	TotalCount  int32         `json:"total_count"`
}

type UpdateConnectionRequest struct {
	ID           string `json:"id"`
	Name         string `json:"name"`
	Host         string `json:"host"`
	Port         int32  `json:"port"`
	DatabaseName string `json:"database_name"`
	Username     string `json:"username"`
	Password     string `json:"password"`
	SSLMode      string `json:"ssl_mode"`
}

type UpdateConnectionResponse struct {
	Connection *Connection `json:"connection"`
}

type DeleteConnectionRequest struct {
	ID string `json:"id"`
}

type DeleteConnectionResponse struct{}

type TestConnectionRequest struct {
	ConnectionID string `json:"connection_id"`
}

type TestConnectionResponse struct {
	Success       bool   `json:"success"`
	LatencyMs     int32  `json:"latency_ms"`
	ServerVersion string `json:"server_version"`
	DatabaseName  string `json:"database_name"`
	Error         string `json:"error"`
}

func (r *CreateConnectionRequest) Validate() error {
	if r.Name == "" {
		return fmt.Errorf("connection name is required")
	}
	if r.Host == "" {
		return fmt.Errorf("host is required")
	}
	if r.Port <= 0 || r.Port > 65535 {
		r.Port = 5432
	}
	if r.DatabaseName == "" {
		return fmt.Errorf("database name is required")
	}
	if r.Username == "" {
		return fmt.Errorf("username is required")
	}
	if r.Password == "" {
		return fmt.Errorf("password is required")
	}
	if r.SSLMode == "" {
		r.SSLMode = "require"
	}
	return nil
}

func (r *UpdateConnectionRequest) Validate() error {
	if r.ID == "" {
		return fmt.Errorf("connection id is required")
	}
	return nil
}
