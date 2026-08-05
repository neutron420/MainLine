package domain

import (
	"context"
	"fmt"
	"regexp"
	"strings"
	"time"
)

type ProjectVisibility string

const (
	VisibilityPrivate ProjectVisibility = "private"
	VisibilityTeam    ProjectVisibility = "team"
	VisibilityPublic  ProjectVisibility = "public"
)

type ProjectRole string

const (
	RoleOwner  ProjectRole = "owner"
	RoleAdmin  ProjectRole = "admin"
	RoleMember ProjectRole = "member"
	RoleViewer ProjectRole = "viewer"
)

type Project struct {
	ID          string
	Name        string
	Slug        string
	Description string
	Visibility  ProjectVisibility
	Template    string
	CreatedBy   string
	CreatedAt   time.Time
	UpdatedAt   time.Time
	DeletedAt   *time.Time
}

type ProjectMember struct {
	ID        string
	ProjectID string
	UserID    string
	Role      ProjectRole
	InvitedBy *string
	JoinedAt  *time.Time
	CreatedAt time.Time
}

type InvitationStatus string

const (
	InvitationPending  InvitationStatus = "pending"
	InvitationAccepted InvitationStatus = "accepted"
	InvitationRevoked  InvitationStatus = "revoked"
)

type ProjectInvitation struct {
	ID        string
	ProjectID string
	Email     string
	Role      ProjectRole
	Token     string
	Status    InvitationStatus
	InvitedBy string
	ExpiresAt time.Time
	CreatedAt time.Time
	AcceptedAt *time.Time
}

type ProjectRepository interface {
	Create(ctx context.Context, p *Project) error
	GetByID(ctx context.Context, id string) (*Project, error)
	GetBySlug(ctx context.Context, slug string) (*Project, error)
	ListByUserID(ctx context.Context, userID, cursor string, limit int32) ([]*Project, string, int32, error)
	Update(ctx context.Context, p *Project) error
	SoftDelete(ctx context.Context, id string) error

	AddMember(ctx context.Context, m *ProjectMember) error
	RemoveMember(ctx context.Context, projectID, userID string) error
	UpdateMemberRole(ctx context.Context, projectID, userID string, role ProjectRole) error
	GetMember(ctx context.Context, projectID, userID string) (*ProjectMember, error)
	ListMembers(ctx context.Context, projectID, cursor string, limit int32) ([]*ProjectMember, string, int32, error)
	ListMemberUsers(ctx context.Context, projectID string) ([]*ProjectMember, error)

	CreateInvitation(ctx context.Context, inv *ProjectInvitation) error
	GetInvitationByToken(ctx context.Context, token string) (*ProjectInvitation, error)
	MarkInvitationAccepted(ctx context.Context, id, projectID, userID string) error
}

var slugRegex = regexp.MustCompile(`^[a-z0-9-]+$`)

var validTemplates = map[string]bool{
	"blank":     true,
	"starter":   true,
	"ecommerce": true,
}

func ValidTemplate(t string) bool {
	return validTemplates[t]
}

func GenerateSlug(name string) string {
	s := strings.ToLower(name)
	s = regexp.MustCompile(`[^a-z0-9]+`).ReplaceAllString(s, "-")
	s = strings.Trim(s, "-")
	if len(s) > 200 {
		s = s[:200]
	}
	if s == "" {
		s = fmt.Sprintf("project-%d", time.Now().Unix())
	}
	return s
}

func (r ProjectRole) Permissions() int {
	switch r {
	case RoleOwner:
		return 100
	case RoleAdmin:
		return 80
	case RoleMember:
		return 50
	case RoleViewer:
		return 10
	default:
		return 0
	}
}

func (r ProjectRole) CanManageMembers() bool {
	return r == RoleOwner || r == RoleAdmin
}

func (r ProjectRole) CanWrite() bool {
	return r.Permissions() >= RoleMember.Permissions()
}

func (r ProjectRole) CanManageConnections() bool {
	return r == RoleOwner || r == RoleAdmin
}

func ValidateVisibility(v string) (ProjectVisibility, error) {
	switch v {
	case "private":
		return VisibilityPrivate, nil
	case "team":
		return VisibilityTeam, nil
	case "public":
		return VisibilityPublic, nil
	default:
		return "", fmt.Errorf("invalid visibility: %s (must be private, team, or public)", v)
	}
}

func ValidateRole(v string) (ProjectRole, error) {
	switch v {
	case "owner":
		return RoleOwner, nil
	case "admin":
		return RoleAdmin, nil
	case "member":
		return RoleMember, nil
	case "viewer":
		return RoleViewer, nil
	default:
		return "", fmt.Errorf("invalid role: %s (must be owner, admin, member, or viewer)", v)
	}
}
