package handler

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/schemahub/backend/internal/pkg/interceptor"
	"github.com/schemahub/backend/internal/project/domain"
	projectv1 "github.com/schemahub/backend/proto/project/v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type fakeProjRepo struct {
	projects  map[string]*domain.Project
	members   map[string]*domain.ProjectMember
	memberErr error
}

func newFakeProjRepo() *fakeProjRepo {
	return &fakeProjRepo{
		projects: map[string]*domain.Project{},
		members:  map[string]*domain.ProjectMember{},
	}
}

func (f *fakeProjRepo) Create(ctx context.Context, p *domain.Project) error {
	p.ID = "proj_1"
	f.projects[p.ID] = p
	return nil
}

func (f *fakeProjRepo) GetByID(ctx context.Context, id string) (*domain.Project, error) {
	p, ok := f.projects[id]
	if !ok {
		return nil, domain.ErrProjectNotFound{ID: id}
	}
	return p, nil
}

func (f *fakeProjRepo) GetBySlug(ctx context.Context, slug string) (*domain.Project, error) {
	return nil, domain.ErrProjectNotFound{ID: slug}
}

func (f *fakeProjRepo) ListByUserID(ctx context.Context, userID, cursor string, limit int32) ([]*domain.Project, string, int32, error) {
	return nil, "", 0, nil
}

func (f *fakeProjRepo) Update(ctx context.Context, p *domain.Project) error {
	f.projects[p.ID] = p
	return nil
}

func (f *fakeProjRepo) SoftDelete(ctx context.Context, id string) error { return nil }

func (f *fakeProjRepo) AddMember(ctx context.Context, m *domain.ProjectMember) error {
	f.members[m.UserID] = m
	return nil
}

func (f *fakeProjRepo) RemoveMember(ctx context.Context, projectID, userID string) error {
	return nil
}

func (f *fakeProjRepo) UpdateMemberRole(ctx context.Context, projectID, userID string, role domain.ProjectRole) error {
	return nil
}

func (f *fakeProjRepo) GetMember(ctx context.Context, projectID, userID string) (*domain.ProjectMember, error) {
	if f.memberErr != nil {
		return nil, f.memberErr
	}
	m, ok := f.members[userID]
	if !ok {
		return nil, errors.New("member not found")
	}
	return m, nil
}

func (f *fakeProjRepo) ListMembers(ctx context.Context, projectID, cursor string, limit int32) ([]*domain.ProjectMember, string, int32, error) {
	var out []*domain.ProjectMember
	for _, m := range f.members {
		out = append(out, m)
	}
	return out, "", int32(len(out)), nil
}

func (f *fakeProjRepo) ListMemberUsers(ctx context.Context, projectID string) ([]*domain.ProjectMember, error) {
	var out []*domain.ProjectMember
	for _, m := range f.members {
		out = append(out, m)
	}
	return out, nil
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

type fakeConnRepo struct{}

func (f *fakeConnRepo) Create(ctx context.Context, c *domain.Connection) error { return nil }
func (f *fakeConnRepo) GetByID(ctx context.Context, id string) (*domain.Connection, error) {
	return nil, errors.New("not found")
}
func (f *fakeConnRepo) ListByProjectID(ctx context.Context, projectID, cursor string, limit int32) ([]*domain.Connection, string, int32, error) {
	return nil, "", 0, nil
}
func (f *fakeConnRepo) ListAll(ctx context.Context) ([]*domain.Connection, error) {
	return nil, nil
}
func (f *fakeConnRepo) Update(ctx context.Context, c *domain.Connection) error { return nil }
func (f *fakeConnRepo) SoftDelete(ctx context.Context, id string) error        { return nil }
func (f *fakeConnRepo) UpdateStatus(ctx context.Context, id string, s domain.ConnectionStatus, lastConnectedAt *time.Time) error {
	return nil
}

func testProjectHandler(t *testing.T, repo *fakeProjRepo, lookup *fakeUserLookup) *ProjectHandler {
	t.Helper()
	svc := domain.NewProjectService(repo, lookup)
	connSvc := domain.NewConnectionService(&fakeConnRepo{}, []byte("0123456789abcdef"))
	return NewProjectHandler(svc, connSvc)
}

func TestProjectHandler_CreateProject(t *testing.T) {
	t.Parallel()

	repo := newFakeProjRepo()
	repo.members["user_1"] = &domain.ProjectMember{ProjectID: "proj_1", UserID: "user_1", Role: domain.RoleOwner}
	h := testProjectHandler(t, repo, &fakeUserLookup{valid: true})

	ctx := context.WithValue(context.Background(), interceptor.UserIDKey, "user_1")
	resp, err := h.CreateProject(ctx, &projectv1.CreateProjectRequest{
		Name: "My App", Description: "app db", Visibility: "private",
	})
	if err != nil {
		t.Fatalf("CreateProject() error = %v", err)
	}
	if resp.Project.Name != "My App" {
		t.Errorf("Project.Name = %q, want My App", resp.Project.Name)
	}
	if resp.Project.Slug != "my-app" {
		t.Errorf("Project.Slug = %q, want my-app", resp.Project.Slug)
	}
	if resp.Project.CreatedBy != "user_1" {
		t.Errorf("Project.CreatedBy = %q, want user_1", resp.Project.CreatedBy)
	}
}

func TestProjectHandler_AddMemberByUserID(t *testing.T) {
	t.Parallel()

	repo := newFakeProjRepo()
	repo.members["user_1"] = &domain.ProjectMember{ProjectID: "proj_1", UserID: "user_1", Role: domain.RoleOwner}
	h := testProjectHandler(t, repo, &fakeUserLookup{ids: map[string]string{"teammate@schemahub.dev": "user_2"}, valid: true})

	ctx := context.WithValue(context.Background(), interceptor.UserIDKey, "user_1")
	_, err := h.AddMember(ctx, &projectv1.AddMemberRequest{
		ProjectId: "proj_1", UserId: "user_2", Role: "member",
	})
	if err != nil {
		t.Fatalf("AddMember() error = %v", err)
	}
	if _, ok := repo.members["user_2"]; !ok {
		t.Error("expected member user_2 to be added")
	}
}

func TestProjectHandler_AddMemberPermissionDenied(t *testing.T) {
	t.Parallel()

	repo := newFakeProjRepo()
	repo.members["user_1"] = &domain.ProjectMember{ProjectID: "proj_1", UserID: "user_1", Role: domain.RoleViewer}
	h := testProjectHandler(t, repo, &fakeUserLookup{valid: true})

	ctx := context.WithValue(context.Background(), interceptor.UserIDKey, "user_1")
	_, err := h.AddMember(ctx, &projectv1.AddMemberRequest{
		ProjectId: "proj_1", UserId: "user_2", Role: "member",
	})
	if status.Code(err) != codes.PermissionDenied {
		t.Fatalf("AddMember() error code = %v, want PermissionDenied (%v)", status.Code(err), err)
	}
}

func TestProjectHandler_ListMembers(t *testing.T) {
	t.Parallel()

	repo := newFakeProjRepo()
	repo.members["user_1"] = &domain.ProjectMember{ProjectID: "proj_1", UserID: "user_1", Role: domain.RoleOwner}
	repo.members["user_2"] = &domain.ProjectMember{ProjectID: "proj_1", UserID: "user_2", Role: domain.RoleMember}
	h := testProjectHandler(t, repo, &fakeUserLookup{valid: true})

	ctx := context.WithValue(context.Background(), interceptor.UserIDKey, "user_1")
	resp, err := h.ListMembers(ctx, &projectv1.ListMembersRequest{ProjectId: "proj_1", PageSize: 10})
	if err != nil {
		t.Fatalf("ListMembers() error = %v", err)
	}
	if len(resp.Members) != 2 {
		t.Errorf("Members len = %d, want 2", len(resp.Members))
	}
	if resp.TotalCount != 2 {
		t.Errorf("TotalCount = %d, want 2", resp.TotalCount)
	}
}

func TestProjectHandler_UpdateProject(t *testing.T) {
	t.Parallel()

	repo := newFakeProjRepo()
	repo.projects["proj_1"] = &domain.Project{
		ID: "proj_1", Name: "Old Name", Slug: "old-name", Visibility: domain.VisibilityPrivate,
	}
	repo.members["user_1"] = &domain.ProjectMember{ProjectID: "proj_1", UserID: "user_1", Role: domain.RoleOwner}
	h := testProjectHandler(t, repo, &fakeUserLookup{valid: true})

	ctx := context.WithValue(context.Background(), interceptor.UserIDKey, "user_1")
	resp, err := h.UpdateProject(ctx, &projectv1.UpdateProjectRequest{
		Id: "proj_1", Name: "New Name", Visibility: "public",
	})
	if err != nil {
		t.Fatalf("UpdateProject() error = %v", err)
	}
	if resp.Project.Name != "New Name" {
		t.Errorf("Project.Name = %q, want New Name", resp.Project.Name)
	}
	if resp.Project.Visibility != "public" {
		t.Errorf("Project.Visibility = %q, want public", resp.Project.Visibility)
	}
}
