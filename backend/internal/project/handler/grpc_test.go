package handler

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	pkgErrors "github.com/schemahub/backend/internal/pkg/errors"
	"github.com/schemahub/backend/internal/pkg/interceptor"
	"github.com/schemahub/backend/internal/project/domain"
	"github.com/schemahub/backend/pkg/encryption"
	projectv1 "github.com/schemahub/backend/proto/project/v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type fakeProjRepo struct {
	projects    map[string]*domain.Project
	slugs       map[string]*domain.Project
	members     map[string]*domain.ProjectMember
	memberErr   error
	listErr     error
	deleted     []string
	removed     []string
	roleUpdates []string
}

func newFakeProjRepo() *fakeProjRepo {
	return &fakeProjRepo{
		projects: map[string]*domain.Project{},
		slugs:    map[string]*domain.Project{},
		members:  map[string]*domain.ProjectMember{},
	}
}

func (f *fakeProjRepo) Create(ctx context.Context, p *domain.Project) error {
	p.ID = "proj_1"
	f.projects[p.ID] = p
	f.slugs[p.Slug] = p
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
	p, ok := f.slugs[slug]
	if !ok {
		return nil, domain.ErrProjectNotFound{ID: slug}
	}
	return p, nil
}

func (f *fakeProjRepo) ListByUserID(ctx context.Context, userID, cursor string, limit int32) ([]*domain.Project, string, int32, error) {
	if f.listErr != nil {
		return nil, "", 0, f.listErr
	}
	var out []*domain.Project
	for _, p := range f.projects {
		if p.CreatedBy == userID {
			out = append(out, p)
		}
	}
	total := int32(len(out))
	next := ""
	if int32(len(out)) > limit {
		out = out[:limit]
		next = "cursor_next"
	}
	return out, next, total, nil
}

func (f *fakeProjRepo) Update(ctx context.Context, p *domain.Project) error {
	f.projects[p.ID] = p
	if p.Slug != "" {
		f.slugs[p.Slug] = p
	}
	return nil
}

func (f *fakeProjRepo) SoftDelete(ctx context.Context, id string) error {
	f.deleted = append(f.deleted, id)
	delete(f.projects, id)
	return nil
}

func (f *fakeProjRepo) AddMember(ctx context.Context, m *domain.ProjectMember) error {
	f.members[m.UserID] = m
	return nil
}

func (f *fakeProjRepo) RemoveMember(ctx context.Context, projectID, userID string) error {
	f.removed = append(f.removed, userID)
	delete(f.members, userID)
	return nil
}

func (f *fakeProjRepo) UpdateMemberRole(ctx context.Context, projectID, userID string, role domain.ProjectRole) error {
	f.roleUpdates = append(f.roleUpdates, userID+"|"+string(role))
	if m, ok := f.members[userID]; ok {
		m.Role = role
	}
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

func (f *fakeProjRepo) CreateInvitation(ctx context.Context, inv *domain.ProjectInvitation) error {
	return nil
}

func (f *fakeProjRepo) GetInvitationByToken(ctx context.Context, token string) (*domain.ProjectInvitation, error) {
	return nil, errors.New("invitation not found")
}

func (f *fakeProjRepo) MarkInvitationAccepted(ctx context.Context, id, projectID, userID string) error {
	return nil
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

type fakeConnRepo struct {
	conns       map[string]*domain.Connection
	getErr      error
	deleted     []string
	statusCalls []string
}

func newFakeConnRepo() *fakeConnRepo {
	return &fakeConnRepo{conns: map[string]*domain.Connection{}}
}

func (f *fakeConnRepo) Create(ctx context.Context, c *domain.Connection) error {
	c.ID = "conn_1"
	now := time.Now()
	c.CreatedAt = now
	c.UpdatedAt = now
	f.conns[c.ID] = c
	return nil
}

func (f *fakeConnRepo) GetByID(ctx context.Context, id string) (*domain.Connection, error) {
	if f.getErr != nil {
		return nil, f.getErr
	}
	c, ok := f.conns[id]
	if !ok {
		return nil, pkgErrors.New("NOT_FOUND", "connection not found")
	}
	return c, nil
}

func (f *fakeConnRepo) ListByProjectID(ctx context.Context, projectID, cursor string, limit int32) ([]*domain.Connection, string, int32, error) {
	var out []*domain.Connection
	for _, c := range f.conns {
		if c.ProjectID == projectID {
			out = append(out, c)
		}
	}
	return out, "", int32(len(out)), nil
}

func (f *fakeConnRepo) ListAll(ctx context.Context) ([]*domain.Connection, error) {
	return nil, nil
}

func (f *fakeConnRepo) Update(ctx context.Context, c *domain.Connection) error {
	f.conns[c.ID] = c
	return nil
}

func (f *fakeConnRepo) SoftDelete(ctx context.Context, id string) error {
	f.deleted = append(f.deleted, id)
	delete(f.conns, id)
	return nil
}

func (f *fakeConnRepo) UpdateStatus(ctx context.Context, id string, s domain.ConnectionStatus, lastConnectedAt *time.Time) error {
	f.statusCalls = append(f.statusCalls, id+"|"+string(s))
	if c, ok := f.conns[id]; ok {
		c.ConnectionStatus = s
		c.LastConnectedAt = lastConnectedAt
	}
	return nil
}

func testProjectHandler(t *testing.T, repo *fakeProjRepo, lookup *fakeUserLookup) *ProjectHandler {
	t.Helper()
	return testProjectHandlerWithConns(t, repo, lookup, newFakeConnRepo())
}

func testProjectHandlerWithConns(t *testing.T, repo *fakeProjRepo, lookup *fakeUserLookup, conns *fakeConnRepo) *ProjectHandler {
	t.Helper()
	svc := domain.NewProjectService(repo, lookup)
	connSvc := domain.NewConnectionService(conns, []byte("0123456789abcdef"))
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

func TestProjectHandler_GetProject(t *testing.T) {
	t.Parallel()

	repo := newFakeProjRepo()
	repo.projects["proj_1"] = &domain.Project{
		ID: "proj_1", Name: "My App", Slug: "my-app", Visibility: domain.VisibilityPrivate,
		CreatedBy: "user_1", CreatedAt: time.Now(), UpdatedAt: time.Now(),
	}
	h := testProjectHandler(t, repo, &fakeUserLookup{valid: true})

	resp, err := h.GetProject(context.Background(), &projectv1.GetProjectRequest{Id: "proj_1"})
	if err != nil {
		t.Fatalf("GetProject() error = %v", err)
	}
	if resp.Project.Name != "My App" {
		t.Errorf("Project.Name = %q, want My App", resp.Project.Name)
	}
	if resp.Project.Slug != "my-app" {
		t.Errorf("Project.Slug = %q, want my-app", resp.Project.Slug)
	}
}

func TestProjectHandler_GetProjectNotFound(t *testing.T) {
	t.Parallel()

	h := testProjectHandler(t, newFakeProjRepo(), &fakeUserLookup{valid: true})

	_, err := h.GetProject(context.Background(), &projectv1.GetProjectRequest{Id: "ghost"})
	if status.Code(err) != codes.NotFound {
		t.Fatalf("GetProject() error code = %v, want NotFound (%v)", status.Code(err), err)
	}
}

func TestProjectHandler_ListProjects(t *testing.T) {
	t.Parallel()

	repo := newFakeProjRepo()
	repo.projects["proj_1"] = &domain.Project{ID: "proj_1", Name: "One", CreatedBy: "user_1", CreatedAt: time.Now(), UpdatedAt: time.Now()}
	repo.projects["proj_2"] = &domain.Project{ID: "proj_2", Name: "Two", CreatedBy: "user_1", CreatedAt: time.Now(), UpdatedAt: time.Now()}
	repo.projects["proj_3"] = &domain.Project{ID: "proj_3", Name: "Other", CreatedBy: "user_9", CreatedAt: time.Now(), UpdatedAt: time.Now()}
	h := testProjectHandler(t, repo, &fakeUserLookup{valid: true})

	ctx := context.WithValue(context.Background(), interceptor.UserIDKey, "user_1")
	resp, err := h.ListProjects(ctx, &projectv1.ListProjectsRequest{PageSize: 10})
	if err != nil {
		t.Fatalf("ListProjects() error = %v", err)
	}
	if len(resp.Projects) != 2 {
		t.Errorf("Projects len = %d, want 2", len(resp.Projects))
	}
	if resp.TotalCount != 2 {
		t.Errorf("TotalCount = %d, want 2", resp.TotalCount)
	}
	if resp.NextCursor != "" {
		t.Errorf("NextCursor = %q, want empty", resp.NextCursor)
	}
}

func TestProjectHandler_ListProjectsPagination(t *testing.T) {
	t.Parallel()

	repo := newFakeProjRepo()
	for i, name := range []string{"One", "Two", "Three"} {
		repo.projects[fmt.Sprintf("proj_%d", i+1)] = &domain.Project{
			ID: fmt.Sprintf("proj_%d", i+1), Name: name, CreatedBy: "user_1",
			CreatedAt: time.Now(), UpdatedAt: time.Now(),
		}
	}
	h := testProjectHandler(t, repo, &fakeUserLookup{valid: true})

	ctx := context.WithValue(context.Background(), interceptor.UserIDKey, "user_1")
	resp, err := h.ListProjects(ctx, &projectv1.ListProjectsRequest{PageSize: 1})
	if err != nil {
		t.Fatalf("ListProjects() error = %v", err)
	}
	if len(resp.Projects) != 1 {
		t.Errorf("Projects len = %d, want 1", len(resp.Projects))
	}
	if resp.NextCursor != "cursor_next" {
		t.Errorf("NextCursor = %q, want cursor_next", resp.NextCursor)
	}
	if resp.TotalCount != 3 {
		t.Errorf("TotalCount = %d, want 3", resp.TotalCount)
	}
}

func TestProjectHandler_CreateProjectInvalidVisibility(t *testing.T) {
	t.Parallel()

	h := testProjectHandler(t, newFakeProjRepo(), &fakeUserLookup{valid: true})

	ctx := context.WithValue(context.Background(), interceptor.UserIDKey, "user_1")
	_, err := h.CreateProject(ctx, &projectv1.CreateProjectRequest{
		Name: "My App", Visibility: "bogus",
	})
	if status.Code(err) != codes.Internal {
		t.Fatalf("CreateProject() error code = %v, want Internal (%v)", status.Code(err), err)
	}
}

func TestProjectHandler_CreateProjectSlugConflict(t *testing.T) {
	t.Parallel()

	repo := newFakeProjRepo()
	repo.slugs["my-app"] = &domain.Project{ID: "proj_1", Slug: "my-app"}
	h := testProjectHandler(t, repo, &fakeUserLookup{valid: true})

	ctx := context.WithValue(context.Background(), interceptor.UserIDKey, "user_1")
	_, err := h.CreateProject(ctx, &projectv1.CreateProjectRequest{
		Name: "My App", Visibility: "private",
	})
	if status.Code(err) != codes.AlreadyExists {
		t.Fatalf("CreateProject() error code = %v, want AlreadyExists (%v)", status.Code(err), err)
	}
}

func TestProjectHandler_UpdateProjectNotFound(t *testing.T) {
	t.Parallel()

	h := testProjectHandler(t, newFakeProjRepo(), &fakeUserLookup{valid: true})

	ctx := context.WithValue(context.Background(), interceptor.UserIDKey, "user_1")
	_, err := h.UpdateProject(ctx, &projectv1.UpdateProjectRequest{Id: "ghost", Name: "X"})
	if status.Code(err) != codes.NotFound {
		t.Fatalf("UpdateProject() error code = %v, want NotFound (%v)", status.Code(err), err)
	}
}

func TestProjectHandler_UpdateProjectPermissionDenied(t *testing.T) {
	t.Parallel()

	repo := newFakeProjRepo()
	repo.projects["proj_1"] = &domain.Project{ID: "proj_1", Name: "My App"}
	repo.members["user_1"] = &domain.ProjectMember{ProjectID: "proj_1", UserID: "user_1", Role: domain.RoleViewer}
	h := testProjectHandler(t, repo, &fakeUserLookup{valid: true})

	ctx := context.WithValue(context.Background(), interceptor.UserIDKey, "user_1")
	_, err := h.UpdateProject(ctx, &projectv1.UpdateProjectRequest{Id: "proj_1", Name: "X"})
	if status.Code(err) != codes.PermissionDenied {
		t.Fatalf("UpdateProject() error code = %v, want PermissionDenied (%v)", status.Code(err), err)
	}
}

func TestProjectHandler_DeleteProject(t *testing.T) {
	t.Parallel()

	repo := newFakeProjRepo()
	repo.projects["proj_1"] = &domain.Project{ID: "proj_1", Name: "My App"}
	repo.members["user_1"] = &domain.ProjectMember{ProjectID: "proj_1", UserID: "user_1", Role: domain.RoleOwner}
	h := testProjectHandler(t, repo, &fakeUserLookup{valid: true})

	ctx := context.WithValue(context.Background(), interceptor.UserIDKey, "user_1")
	_, err := h.DeleteProject(ctx, &projectv1.DeleteProjectRequest{Id: "proj_1"})
	if err != nil {
		t.Fatalf("DeleteProject() error = %v", err)
	}
	if len(repo.deleted) != 1 || repo.deleted[0] != "proj_1" {
		t.Errorf("deleted = %v, want [proj_1]", repo.deleted)
	}
}

func TestProjectHandler_DeleteProjectNotOwner(t *testing.T) {
	t.Parallel()

	repo := newFakeProjRepo()
	repo.projects["proj_1"] = &domain.Project{ID: "proj_1", Name: "My App"}
	repo.members["user_1"] = &domain.ProjectMember{ProjectID: "proj_1", UserID: "user_1", Role: domain.RoleMember}
	h := testProjectHandler(t, repo, &fakeUserLookup{valid: true})

	ctx := context.WithValue(context.Background(), interceptor.UserIDKey, "user_1")
	_, err := h.DeleteProject(ctx, &projectv1.DeleteProjectRequest{Id: "proj_1"})
	if status.Code(err) != codes.Internal {
		t.Fatalf("DeleteProject() error code = %v, want Internal (%v)", status.Code(err), err)
	}
}

func TestProjectHandler_AddMemberByEmail(t *testing.T) {
	t.Parallel()

	repo := newFakeProjRepo()
	repo.members["user_1"] = &domain.ProjectMember{ProjectID: "proj_1", UserID: "user_1", Role: domain.RoleOwner}
	h := testProjectHandler(t, repo, &fakeUserLookup{ids: map[string]string{"teammate@schemahub.dev": "user_2"}, valid: true})

	ctx := context.WithValue(context.Background(), interceptor.UserIDKey, "user_1")
	_, err := h.AddMember(ctx, &projectv1.AddMemberRequest{
		ProjectId: "proj_1", Email: "teammate@schemahub.dev", Role: "member",
	})
	if err != nil {
		t.Fatalf("AddMember() error = %v", err)
	}
	if _, ok := repo.members["user_2"]; !ok {
		t.Error("expected member user_2 to be added via email invite")
	}
}

func TestProjectHandler_AddMemberEmailNotFound(t *testing.T) {
	t.Parallel()

	repo := newFakeProjRepo()
	repo.members["user_1"] = &domain.ProjectMember{ProjectID: "proj_1", UserID: "user_1", Role: domain.RoleOwner}
	h := testProjectHandler(t, repo, &fakeUserLookup{valid: true})

	ctx := context.WithValue(context.Background(), interceptor.UserIDKey, "user_1")
	_, err := h.AddMember(ctx, &projectv1.AddMemberRequest{
		ProjectId: "proj_1", Email: "ghost@schemahub.dev", Role: "member",
	})
	if status.Code(err) != codes.NotFound {
		t.Fatalf("AddMember() error code = %v, want NotFound (%v)", status.Code(err), err)
	}
}

func TestProjectHandler_AddMemberNoUserSpecified(t *testing.T) {
	t.Parallel()

	repo := newFakeProjRepo()
	repo.members["user_1"] = &domain.ProjectMember{ProjectID: "proj_1", UserID: "user_1", Role: domain.RoleOwner}
	h := testProjectHandler(t, repo, &fakeUserLookup{valid: true})

	ctx := context.WithValue(context.Background(), interceptor.UserIDKey, "user_1")
	_, err := h.AddMember(ctx, &projectv1.AddMemberRequest{
		ProjectId: "proj_1", Role: "member",
	})
	if status.Code(err) != codes.InvalidArgument {
		t.Fatalf("AddMember() error code = %v, want InvalidArgument (%v)", status.Code(err), err)
	}
}

func TestProjectHandler_AddMemberInvalidRole(t *testing.T) {
	t.Parallel()

	repo := newFakeProjRepo()
	repo.members["user_1"] = &domain.ProjectMember{ProjectID: "proj_1", UserID: "user_1", Role: domain.RoleOwner}
	h := testProjectHandler(t, repo, &fakeUserLookup{valid: true})

	ctx := context.WithValue(context.Background(), interceptor.UserIDKey, "user_1")
	_, err := h.AddMember(ctx, &projectv1.AddMemberRequest{
		ProjectId: "proj_1", UserId: "user_2", Role: "bogus",
	})
	if status.Code(err) != codes.Internal {
		t.Fatalf("AddMember() error code = %v, want Internal (%v)", status.Code(err), err)
	}
}

func TestProjectHandler_RemoveMember(t *testing.T) {
	t.Parallel()

	repo := newFakeProjRepo()
	repo.members["user_1"] = &domain.ProjectMember{ProjectID: "proj_1", UserID: "user_1", Role: domain.RoleOwner}
	repo.members["user_2"] = &domain.ProjectMember{ProjectID: "proj_1", UserID: "user_2", Role: domain.RoleMember}
	h := testProjectHandler(t, repo, &fakeUserLookup{valid: true})

	ctx := context.WithValue(context.Background(), interceptor.UserIDKey, "user_1")
	_, err := h.RemoveMember(ctx, &projectv1.RemoveMemberRequest{ProjectId: "proj_1", UserId: "user_2"})
	if err != nil {
		t.Fatalf("RemoveMember() error = %v", err)
	}
	if len(repo.removed) != 1 || repo.removed[0] != "user_2" {
		t.Errorf("removed = %v, want [user_2]", repo.removed)
	}
}

func TestProjectHandler_RemoveMemberLastOwner(t *testing.T) {
	t.Parallel()

	repo := newFakeProjRepo()
	repo.members["user_1"] = &domain.ProjectMember{ProjectID: "proj_1", UserID: "user_1", Role: domain.RoleOwner}
	h := testProjectHandler(t, repo, &fakeUserLookup{valid: true})

	ctx := context.WithValue(context.Background(), interceptor.UserIDKey, "user_1")
	_, err := h.RemoveMember(ctx, &projectv1.RemoveMemberRequest{ProjectId: "proj_1", UserId: "user_1"})
	if status.Code(err) != codes.FailedPrecondition {
		t.Fatalf("RemoveMember() error code = %v, want FailedPrecondition (%v)", status.Code(err), err)
	}
}

func TestProjectHandler_RemoveMemberNotFound(t *testing.T) {
	t.Parallel()

	repo := newFakeProjRepo()
	repo.members["user_1"] = &domain.ProjectMember{ProjectID: "proj_1", UserID: "user_1", Role: domain.RoleOwner}
	h := testProjectHandler(t, repo, &fakeUserLookup{valid: true})

	ctx := context.WithValue(context.Background(), interceptor.UserIDKey, "user_1")
	_, err := h.RemoveMember(ctx, &projectv1.RemoveMemberRequest{ProjectId: "proj_1", UserId: "ghost"})
	if status.Code(err) != codes.NotFound {
		t.Fatalf("RemoveMember() error code = %v, want NotFound (%v)", status.Code(err), err)
	}
}

func TestProjectHandler_UpdateMemberRole(t *testing.T) {
	t.Parallel()

	repo := newFakeProjRepo()
	repo.members["user_1"] = &domain.ProjectMember{ProjectID: "proj_1", UserID: "user_1", Role: domain.RoleOwner}
	repo.members["user_2"] = &domain.ProjectMember{ProjectID: "proj_1", UserID: "user_2", Role: domain.RoleMember}
	h := testProjectHandler(t, repo, &fakeUserLookup{valid: true})

	ctx := context.WithValue(context.Background(), interceptor.UserIDKey, "user_1")
	_, err := h.UpdateMemberRole(ctx, &projectv1.UpdateMemberRoleRequest{
		ProjectId: "proj_1", UserId: "user_2", Role: "admin",
	})
	if err != nil {
		t.Fatalf("UpdateMemberRole() error = %v", err)
	}
	if len(repo.roleUpdates) != 1 || repo.roleUpdates[0] != "user_2|admin" {
		t.Errorf("roleUpdates = %v, want [user_2|admin]", repo.roleUpdates)
	}
}

func TestProjectHandler_UpdateMemberRoleInvalidRole(t *testing.T) {
	t.Parallel()

	repo := newFakeProjRepo()
	repo.members["user_1"] = &domain.ProjectMember{ProjectID: "proj_1", UserID: "user_1", Role: domain.RoleOwner}
	h := testProjectHandler(t, repo, &fakeUserLookup{valid: true})

	ctx := context.WithValue(context.Background(), interceptor.UserIDKey, "user_1")
	_, err := h.UpdateMemberRole(ctx, &projectv1.UpdateMemberRoleRequest{
		ProjectId: "proj_1", UserId: "user_2", Role: "bogus",
	})
	if status.Code(err) != codes.Internal {
		t.Fatalf("UpdateMemberRole() error code = %v, want Internal (%v)", status.Code(err), err)
	}
}

func TestProjectHandler_CreateConnection(t *testing.T) {
	t.Parallel()

	repo := newFakeProjRepo()
	h := testProjectHandlerWithConns(t, repo, &fakeUserLookup{valid: true}, newFakeConnRepo())

	ctx := context.WithValue(context.Background(), interceptor.UserIDKey, "user_1")
	resp, err := h.CreateConnection(ctx, &projectv1.CreateConnectionRequest{
		ProjectId: "proj_1", Name: "prod", Host: "db.example.com", Port: 5432,
		DatabaseName: "app", Username: "app_user", Password: "s3cret", SslMode: "require",
	})
	if err != nil {
		t.Fatalf("CreateConnection() error = %v", err)
	}
	if resp.Connection.Name != "prod" {
		t.Errorf("Connection.Name = %q, want prod", resp.Connection.Name)
	}
	if resp.Connection.CreatedBy != "user_1" {
		t.Errorf("Connection.CreatedBy = %q, want user_1", resp.Connection.CreatedBy)
	}
}

func TestProjectHandler_GetConnection(t *testing.T) {
	t.Parallel()

	conns := newFakeConnRepo()
	conns.conns["conn_1"] = &domain.Connection{
		ID: "conn_1", ProjectID: "proj_1", Name: "prod", Host: "db.example.com",
		Port: 5432, DatabaseName: "app", Username: "app_user", SSLMode: domain.SSLRequire,
		CreatedAt: time.Now(), UpdatedAt: time.Now(),
	}
	h := testProjectHandlerWithConns(t, newFakeProjRepo(), &fakeUserLookup{valid: true}, conns)

	resp, err := h.GetConnection(context.Background(), &projectv1.GetConnectionRequest{Id: "conn_1"})
	if err != nil {
		t.Fatalf("GetConnection() error = %v", err)
	}
	if resp.Connection.Name != "prod" {
		t.Errorf("Connection.Name = %q, want prod", resp.Connection.Name)
	}
	if resp.Connection.DatabaseName != "app" {
		t.Errorf("Connection.DatabaseName = %q, want app", resp.Connection.DatabaseName)
	}
}

func TestProjectHandler_GetConnectionNotFound(t *testing.T) {
	t.Parallel()

	h := testProjectHandlerWithConns(t, newFakeProjRepo(), &fakeUserLookup{valid: true}, newFakeConnRepo())

	_, err := h.GetConnection(context.Background(), &projectv1.GetConnectionRequest{Id: "ghost"})
	if status.Code(err) != codes.NotFound {
		t.Fatalf("GetConnection() error code = %v, want NotFound (%v)", status.Code(err), err)
	}
}

func TestProjectHandler_ListConnections(t *testing.T) {
	t.Parallel()

	conns := newFakeConnRepo()
	now := time.Now()
	conns.conns["conn_1"] = &domain.Connection{ID: "conn_1", ProjectID: "proj_1", Name: "prod", CreatedAt: now, UpdatedAt: now}
	conns.conns["conn_2"] = &domain.Connection{ID: "conn_2", ProjectID: "proj_1", Name: "staging", CreatedAt: now, UpdatedAt: now}
	conns.conns["conn_3"] = &domain.Connection{ID: "conn_3", ProjectID: "proj_2", Name: "other", CreatedAt: now, UpdatedAt: now}
	h := testProjectHandlerWithConns(t, newFakeProjRepo(), &fakeUserLookup{valid: true}, conns)

	resp, err := h.ListConnections(context.Background(), &projectv1.ListConnectionsRequest{ProjectId: "proj_1", PageSize: 10})
	if err != nil {
		t.Fatalf("ListConnections() error = %v", err)
	}
	if len(resp.Connections) != 2 {
		t.Errorf("Connections len = %d, want 2", len(resp.Connections))
	}
	if resp.TotalCount != 2 {
		t.Errorf("TotalCount = %d, want 2", resp.TotalCount)
	}
}

func TestProjectHandler_UpdateConnection(t *testing.T) {
	t.Parallel()

	conns := newFakeConnRepo()
	conns.conns["conn_1"] = &domain.Connection{
		ID: "conn_1", ProjectID: "proj_1", Name: "old", Host: "old.example.com", Port: 5432,
		DatabaseName: "app", Username: "app_user", SSLMode: domain.SSLRequire,
		CreatedAt: time.Now(), UpdatedAt: time.Now(),
	}
	h := testProjectHandlerWithConns(t, newFakeProjRepo(), &fakeUserLookup{valid: true}, conns)

	resp, err := h.UpdateConnection(context.Background(), &projectv1.UpdateConnectionRequest{
		Id: "conn_1", Name: "new", Host: "new.example.com", Port: 5433,
	})
	if err != nil {
		t.Fatalf("UpdateConnection() error = %v", err)
	}
	if resp.Connection.Name != "new" {
		t.Errorf("Connection.Name = %q, want new", resp.Connection.Name)
	}
	if resp.Connection.Host != "new.example.com" {
		t.Errorf("Connection.Host = %q, want new.example.com", resp.Connection.Host)
	}
}

func TestProjectHandler_UpdateConnectionNotFound(t *testing.T) {
	t.Parallel()

	h := testProjectHandlerWithConns(t, newFakeProjRepo(), &fakeUserLookup{valid: true}, newFakeConnRepo())

	_, err := h.UpdateConnection(context.Background(), &projectv1.UpdateConnectionRequest{Id: "ghost", Name: "x"})
	if status.Code(err) != codes.Internal {
		t.Fatalf("UpdateConnection() error code = %v, want Internal (%v)", status.Code(err), err)
	}
}

func TestProjectHandler_DeleteConnection(t *testing.T) {
	t.Parallel()

	conns := newFakeConnRepo()
	conns.conns["conn_1"] = &domain.Connection{ID: "conn_1", ProjectID: "proj_1", Name: "prod"}
	h := testProjectHandlerWithConns(t, newFakeProjRepo(), &fakeUserLookup{valid: true}, conns)

	_, err := h.DeleteConnection(context.Background(), &projectv1.DeleteConnectionRequest{Id: "conn_1"})
	if err != nil {
		t.Fatalf("DeleteConnection() error = %v", err)
	}
	if len(conns.deleted) != 1 || conns.deleted[0] != "conn_1" {
		t.Errorf("deleted = %v, want [conn_1]", conns.deleted)
	}
}

func TestProjectHandler_TestConnection(t *testing.T) {
	t.Parallel()

	key := []byte("0123456789abcdef")
	enc, err := encryption.Encrypt([]byte("s3cret"), key)
	if err != nil {
		t.Fatal(err)
	}
	conns := newFakeConnRepo()
	conns.conns["conn_1"] = &domain.Connection{
		ID: "conn_1", ProjectID: "proj_1", Name: "prod", Host: "localhost", Port: 99999,
		DatabaseName: "app", Username: "app_user", SSLMode: domain.SSLDisable,
		PasswordEncrypted: enc, CreatedAt: time.Now(), UpdatedAt: time.Now(),
	}
	h := testProjectHandlerWithConns(t, newFakeProjRepo(), &fakeUserLookup{valid: true}, conns)

	resp, err := h.TestConnection(context.Background(), &projectv1.TestConnectionRequest{ConnectionId: "conn_1"})
	if err != nil {
		t.Fatalf("TestConnection() error = %v", err)
	}
	if resp.Success {
		t.Error("Success = true, want false for unreachable database")
	}
	if resp.Error == "" {
		t.Error("Error = empty, want failure description")
	}
	if len(conns.statusCalls) != 1 || conns.statusCalls[0] != "conn_1|failed" {
		t.Errorf("statusCalls = %v, want [conn_1|failed]", conns.statusCalls)
	}
}
