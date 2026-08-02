package domain

import (
	"context"
	"errors"
	"testing"
)

type fakeProjRepo struct {
	project Project
	member  *ProjectMember
	added   *ProjectMember
}

func (f *fakeProjRepo) Create(ctx context.Context, p *Project) error {
	f.project = *p
	return nil
}

func (f *fakeProjRepo) GetByID(ctx context.Context, id string) (*Project, error) {
	if f.project.ID != id {
		return nil, ErrProjectNotFound{ID: id}
	}
	return &f.project, nil
}

func (f *fakeProjRepo) GetBySlug(ctx context.Context, slug string) (*Project, error) {
	if f.project.Slug == slug {
		return &f.project, nil
	}
	return nil, ErrProjectNotFound{ID: slug}
}

func (f *fakeProjRepo) ListByUserID(ctx context.Context, userID, cursor string, limit int32) ([]*Project, string, int32, error) {
	return []*Project{}, "", 0, nil
}

func (f *fakeProjRepo) Update(ctx context.Context, p *Project) error { return nil }
func (f *fakeProjRepo) SoftDelete(ctx context.Context, id string) error {
	return nil
}

func (f *fakeProjRepo) AddMember(ctx context.Context, m *ProjectMember) error {
	f.added = m
	return nil
}

func (f *fakeProjRepo) RemoveMember(ctx context.Context, projectID, userID string) error {
	return nil
}

func (f *fakeProjRepo) UpdateMemberRole(ctx context.Context, projectID, userID string, role ProjectRole) error {
	return nil
}

func (f *fakeProjRepo) GetMember(ctx context.Context, projectID, userID string) (*ProjectMember, error) {
	if f.member != nil {
		return f.member, nil
	}
	return nil, errors.New("member not found")
}

func (f *fakeProjRepo) ListMembers(ctx context.Context, projectID, cursor string, limit int32) ([]*ProjectMember, string, int32, error) {
	return []*ProjectMember{}, "", 0, nil
}

func (f *fakeProjRepo) ListMemberUsers(ctx context.Context, projectID string) ([]*ProjectMember, error) {
	return []*ProjectMember{}, nil
}

type fakeUserLookup struct {
	ids   map[string]string
	valid bool
}

func (f *fakeUserLookup) GetByEmail(ctx context.Context, email string) (string, error) {
	if !f.valid {
		return "", errors.New("lookup disabled")
	}
	id, ok := f.ids[email]
	if !ok {
		return "", errors.New("user not found")
	}
	return id, nil
}

func TestProjectService_AddMemberByEmail(t *testing.T) {
	t.Parallel()

	repo := &fakeProjRepo{
		member: &ProjectMember{ProjectID: "p1", UserID: "owner-1", Role: RoleOwner},
	}
	lookup := &fakeUserLookup{
		ids:   map[string]string{"teammate@example.com": "user-2"},
		valid: true,
	}
	svc := NewProjectService(repo, lookup)

	err := svc.AddMember(context.Background(), "p1", "", "teammate@example.com", "member", "owner-1")
	if err != nil {
		t.Fatalf("AddMember by email: %v", err)
	}
	if repo.added == nil || repo.added.UserID != "user-2" {
		t.Fatalf("expected member user-2 to be added, got %+v", repo.added)
	}
	if repo.added.Role != RoleMember {
		t.Fatalf("expected role member, got %s", repo.added.Role)
	}
}

func TestProjectService_AddMemberByEmailNotFound(t *testing.T) {
	t.Parallel()

	repo := &fakeProjRepo{
		member: &ProjectMember{ProjectID: "p1", UserID: "owner-1", Role: RoleOwner},
	}
	svc := NewProjectService(repo, &fakeUserLookup{ids: map[string]string{}, valid: true})

	err := svc.AddMember(context.Background(), "p1", "", "ghost@example.com", "member", "owner-1")
	if err == nil {
		t.Fatal("expected error for unknown email")
	}
	if _, ok := err.(ErrUserNotFoundByEmail); !ok {
		t.Fatalf("expected ErrUserNotFoundByEmail, got %T: %v", err, err)
	}
}

func TestProjectService_AddMemberRequiresUserOrEmail(t *testing.T) {
	t.Parallel()

	repo := &fakeProjRepo{
		member: &ProjectMember{ProjectID: "p1", UserID: "owner-1", Role: RoleOwner},
	}
	svc := NewProjectService(repo, &fakeUserLookup{valid: true})

	err := svc.AddMember(context.Background(), "p1", "", "", "member", "owner-1")
	if err == nil {
		t.Fatal("expected error when neither user_id nor email given")
	}
	if _, ok := err.(ErrNoUserSpecified); !ok {
		t.Fatalf("expected ErrNoUserSpecified, got %T: %v", err, err)
	}
}

func TestProjectService_AddMemberByUserIDStillWorks(t *testing.T) {
	t.Parallel()

	repo := &fakeProjRepo{
		member: &ProjectMember{ProjectID: "p1", UserID: "owner-1", Role: RoleOwner},
	}
	svc := NewProjectService(repo, &fakeUserLookup{valid: true})

	err := svc.AddMember(context.Background(), "p1", "user-3", "", "viewer", "owner-1")
	if err != nil {
		t.Fatalf("AddMember by user id: %v", err)
	}
	if repo.added == nil || repo.added.UserID != "user-3" {
		t.Fatalf("expected member user-3, got %+v", repo.added)
	}
}

func TestProjectService_AddMemberPermissionDenied(t *testing.T) {
	t.Parallel()

	repo := &fakeProjRepo{
		member: &ProjectMember{ProjectID: "p1", UserID: "viewer-1", Role: RoleViewer},
	}
	svc := NewProjectService(repo, &fakeUserLookup{valid: true})

	err := svc.AddMember(context.Background(), "p1", "", "teammate@example.com", "member", "viewer-1")
	if err == nil || err.Error() != "permission denied" {
		t.Fatalf("expected permission denied, got %v", err)
	}
}
