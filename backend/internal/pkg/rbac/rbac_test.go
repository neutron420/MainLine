package rbac

import (
	"context"
	"errors"
	"testing"
)

type fakeMemberGetter struct {
	roles map[string]string // key: projectID:userID
	err   error
}

func (f *fakeMemberGetter) GetMember(ctx context.Context, projectID, userID string) (string, error) {
	if f.err != nil {
		return "", f.err
	}
	role, ok := f.roles[projectID+":"+userID]
	if !ok {
		return "", errors.New("not a member")
	}
	return role, nil
}

type fakeProjectResolver struct {
	mapping map[string]string // resourceID -> projectID
	err     error
}

func (f *fakeProjectResolver) ResolveProjectID(ctx context.Context, resource ResourceType, resourceID string) (string, error) {
	if f.err != nil {
		return "", f.err
	}
	pid, ok := f.mapping[resourceID]
	if !ok {
		return "", errors.New("resource not found")
	}
	return pid, nil
}

func TestCheck_GlobalAdminBypasses(t *testing.T) {
	t.Parallel()

	c := NewChecker(&fakeMemberGetter{}, &fakeProjectResolver{})
	err := c.Check(context.Background(), "user_1", GlobalRoleAdmin, ResourceProject, ActionDelete, "proj_1")
	if err != nil {
		t.Errorf("global admin should bypass all checks, got error: %v", err)
	}

	err = c.Check(context.Background(), "user_1", GlobalRoleAdmin, ResourceMigration, ActionExecute, "mig_1")
	if err != nil {
		t.Errorf("global admin should bypass child checks, got error: %v", err)
	}
}

func TestCheck_ProjectAccess(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		role    string
		action  Action
		wantErr bool
	}{
		{name: "owner read", role: "owner", action: ActionRead, wantErr: false},
		{name: "admin update", role: "admin", action: ActionUpdate, wantErr: false},
		{name: "member create", role: "member", action: ActionCreate, wantErr: false},
		{name: "viewer read", role: "viewer", action: ActionRead, wantErr: false},
		{name: "non-member denied", role: "", action: ActionRead, wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mg := &fakeMemberGetter{roles: map[string]string{"proj_1:user_1": tt.role}}
			c := NewChecker(mg, &fakeProjectResolver{})

			err := c.Check(context.Background(), "user_1", GlobalRoleUser, ResourceProject, tt.action, "proj_1")
			if (err != nil) != tt.wantErr {
				t.Errorf("Check() error = %v, wantErr = %v", err, tt.wantErr)
			}
		})
	}
}

func TestCheck_ChildResourceResolvesProject(t *testing.T) {
	t.Parallel()

	resolver := &fakeProjectResolver{mapping: map[string]string{
		"mig_1":    "proj_1",
		"schema_1": "proj_1",
	}}
	mg := &fakeMemberGetter{roles: map[string]string{"proj_1:user_1": "member"}}
	c := NewChecker(mg, resolver)

	resources := []ResourceType{ResourceConnection, ResourceSchema, ResourceMigration, ResourceMember}
	for _, res := range resources {
		err := c.Check(context.Background(), "user_1", GlobalRoleUser, res, ActionRead, "mig_1")
		if err != nil {
			t.Errorf("Check(%s) error = %v, want nil", res, err)
		}
	}
}

func TestCheck_ChildResourceDeniedWhenNotMember(t *testing.T) {
	t.Parallel()

	resolver := &fakeProjectResolver{mapping: map[string]string{"mig_1": "proj_1"}}
	c := NewChecker(&fakeMemberGetter{roles: map[string]string{}}, resolver)

	err := c.Check(context.Background(), "user_1", GlobalRoleUser, ResourceMigration, ActionExecute, "mig_1")
	if err == nil {
		t.Error("Check() = nil error, want permission denied")
	}
}

func TestCheck_ResolverErrorDenies(t *testing.T) {
	t.Parallel()

	c := NewChecker(&fakeMemberGetter{}, &fakeProjectResolver{err: errors.New("db down")})

	err := c.Check(context.Background(), "user_1", GlobalRoleUser, ResourceSchema, ActionRead, "schema_x")
	if err == nil {
		t.Error("Check() with resolver error = nil error, want permission denied")
	}
}

func TestCheck_UnknownResource(t *testing.T) {
	t.Parallel()

	c := NewChecker(&fakeMemberGetter{}, &fakeProjectResolver{})
	err := c.Check(context.Background(), "user_1", GlobalRoleUser, ResourceType("banana"), ActionRead, "x")
	if err == nil {
		t.Error("Check(unknown resource) = nil error, want error")
	}
}

func TestCheck_ViewerCanReadButMemberGetterErrorStillDenies(t *testing.T) {
	t.Parallel()

	mg := &fakeMemberGetter{err: errors.New("query failed")}
	c := NewChecker(mg, &fakeProjectResolver{})

	err := c.Check(context.Background(), "user_1", GlobalRoleUser, ResourceProject, ActionRead, "proj_1")
	if err == nil {
		t.Error("Check() with member getter error = nil error, want permission denied")
	}
}
