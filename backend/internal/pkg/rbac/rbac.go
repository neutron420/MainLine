package rbac

import (
	"context"
	"fmt"
)

type Role string

const (
	RoleOwner  Role = "owner"
	RoleAdmin  Role = "admin"
	RoleMember Role = "member"
	RoleViewer Role = "viewer"
)

type GlobalRole string

const (
	GlobalRoleAdmin GlobalRole = "admin"
	GlobalRoleUser  GlobalRole = "user"
)

type ResourceType string

const (
	ResourceProject    ResourceType = "project"
	ResourceConnection ResourceType = "connection"
	ResourceSchema     ResourceType = "schema"
	ResourceMigration  ResourceType = "migration"
	ResourceMember     ResourceType = "member"
)

type Action string

const (
	ActionCreate  Action = "create"
	ActionRead    Action = "read"
	ActionUpdate  Action = "update"
	ActionDelete  Action = "delete"
	ActionExecute Action = "execute"
)

type ProjectMemberGetter interface {
	GetMember(ctx context.Context, projectID, userID string) (role string, err error)
}

type ResourceProjectResolver interface {
	ResolveProjectID(ctx context.Context, resource ResourceType, resourceID string) (string, error)
}

type Checker struct {
	memberGetter   ProjectMemberGetter
	projectResolver ResourceProjectResolver
}

func NewChecker(mg ProjectMemberGetter, pr ResourceProjectResolver) *Checker {
	return &Checker{memberGetter: mg, projectResolver: pr}
}

func (c *Checker) Check(ctx context.Context, userID string, globalRole GlobalRole, resource ResourceType, action Action, resourceID string) error {
	if globalRole == GlobalRoleAdmin {
		return nil
	}

	switch resource {
	case ResourceProject:
		return c.checkProject(ctx, userID, resourceID)
	case ResourceConnection, ResourceSchema, ResourceMigration, ResourceMember:
		return c.checkProjectChild(ctx, userID, resource, action, resourceID)
	default:
		return fmt.Errorf("unknown resource type: %s", resource)
	}
}

func (c *Checker) checkProject(ctx context.Context, userID, projectID string) error {
	member, err := c.memberGetter.GetMember(ctx, projectID, userID)
	if err != nil {
		return fmt.Errorf("permission denied")
	}
	role := Role(member)
	if role == RoleViewer || role == RoleMember || role == RoleAdmin || role == RoleOwner {
		return nil
	}
	return fmt.Errorf("permission denied")
}

func (c *Checker) checkProjectChild(ctx context.Context, userID string, resource ResourceType, action Action, resourceID string) error {
	projectID, err := c.projectResolver.ResolveProjectID(ctx, resource, resourceID)
	if err != nil {
		return fmt.Errorf("permission denied: %w", err)
	}
	return c.checkProject(ctx, userID, projectID)
}
