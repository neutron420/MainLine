package integration

import (
	"context"
	"fmt"
	"strings"
	"testing"
	"time"

	projectdomain "github.com/schemahub/backend/internal/project/domain"
	projectpg "github.com/schemahub/backend/internal/project/repository/postgres"
)

func TestProjectRepository_RoundTrip(t *testing.T) {
	pool := setup(t)
	ctx := context.Background()
	owner := createUser(t, pool)
	repo := projectpg.NewProjectRepository(pool)

	p := &projectdomain.Project{
		Name:        "Round Trip Project",
		Slug:        fmt.Sprintf("rt-%s", newUUID(t)[:8]),
		Description: "original description",
		Visibility:  projectdomain.VisibilityPrivate,
		Template:    "starter",
		CreatedBy:   owner.ID,
	}
	requireNoErr(t, repo.Create(ctx, p), "Create")
	if p.ID == "" {
		t.Fatal("Create did not populate project ID")
	}

	got, err := repo.GetByID(ctx, p.ID)
	requireNoErr(t, err, "GetByID")
	if got.Name != p.Name || got.Slug != p.Slug || got.CreatedBy != owner.ID {
		t.Fatalf("GetByID returned unexpected project: %+v", got)
	}

	bySlug, err := repo.GetBySlug(ctx, p.Slug)
	requireNoErr(t, err, "GetBySlug")
	if bySlug.ID != p.ID {
		t.Fatalf("GetBySlug id = %s, want %s", bySlug.ID, p.ID)
	}

	got.Name = "Renamed Project"
	got.Description = "updated description"
	got.Visibility = projectdomain.VisibilityPublic
	requireNoErr(t, repo.Update(ctx, got), "Update")
	updated, _ := repo.GetByID(ctx, p.ID)
	if updated.Name != "Renamed Project" || updated.Description != "updated description" || updated.Visibility != projectdomain.VisibilityPublic {
		t.Fatalf("Update did not persist: %+v", updated)
	}
	if !updated.UpdatedAt.After(updated.CreatedAt) {
		t.Fatalf("Update did not bump updated_at: %+v", updated)
	}

	requireNoErr(t, repo.SoftDelete(ctx, p.ID), "SoftDelete")
	if _, err := repo.GetByID(ctx, p.ID); err == nil {
		t.Fatal("GetByID after SoftDelete = nil error, want not-found")
	}
	if _, err := repo.GetBySlug(ctx, p.Slug); err == nil {
		t.Fatal("GetBySlug after SoftDelete = nil error, want not-found")
	}
}

func TestProjectRepository_ListByUserID_Pagination(t *testing.T) {
	pool := setup(t)
	ctx := context.Background()
	owner := createUser(t, pool)
	repo := projectpg.NewProjectRepository(pool)

	// ListByUserID orders by updated_at DESC, so give each project a distinct
	// updated_at to make page boundaries deterministic.
	const n = 3
	var ids []string
	for i := 0; i < n; i++ {
		p := createProject(t, pool, owner)
		ids = append(ids, p.ID)
		setCreatedAt(t, pool, "projects", "id", p.ID, 100-i*10)
	}

	page1, cursor, _, err := repo.ListByUserID(ctx, owner.ID, "", 1)
	requireNoErr(t, err, "ListByUserID page 1")
	if len(page1) != 1 || cursor == "" {
		t.Fatalf("page 1 = %d projects, cursor %q; want 1 project + cursor", len(page1), cursor)
	}

	page2, cursor2, _, err := repo.ListByUserID(ctx, owner.ID, cursor, 1)
	requireNoErr(t, err, "ListByUserID page 2")
	if len(page2) != 1 || cursor2 == "" {
		t.Fatalf("page 2 = %d projects, cursor %q; want 1 project + cursor", len(page2), cursor2)
	}
	if page1[0].ID == page2[0].ID {
		t.Fatal("pages returned the same project")
	}

	page3, cursor3, _, err := repo.ListByUserID(ctx, owner.ID, cursor2, 1)
	requireNoErr(t, err, "ListByUserID page 3")
	if len(page3) != 1 || cursor3 != "" {
		t.Fatalf("page 3 = %d projects, cursor %q; want final page", len(page3), cursor3)
	}

	// Walk the full set: every project appears exactly once, no skips/dupes.
	seen := map[string]bool{page1[0].ID: true, page2[0].ID: true, page3[0].ID: true}
	if len(seen) != 3 {
		t.Fatalf("pages covered %d distinct projects, want 3", len(seen))
	}
	for _, id := range ids {
		if !seen[id] {
			t.Fatalf("project %s never appeared in pagination", id)
		}
	}
}

func TestProjectRepository_ListByUserID_PaginationTimestampTies(t *testing.T) {
	pool := setup(t)
	ctx := context.Background()
	owner := createUser(t, pool)
	repo := projectpg.NewProjectRepository(pool)

	// All projects share the same updated_at (created in quick succession):
	// the composite (updated_at, id) cursor must still page without skips.
	for i := 0; i < 5; i++ {
		createProject(t, pool, owner)
	}

	all := map[string]bool{}
	cursor := ""
	for page := 0; page < 6; page++ {
		projects, next, _, err := repo.ListByUserID(ctx, owner.ID, cursor, 2)
		requireNoErr(t, err, "ListByUserID page")
		if len(projects) == 0 {
			break
		}
		for _, p := range projects {
			if all[p.ID] {
				t.Fatalf("project %s duplicated across pages", p.ID)
			}
			all[p.ID] = true
		}
		if next == "" {
			break
		}
		cursor = next
	}
	if len(all) != 5 {
		t.Fatalf("covered %d distinct projects, want 5", len(all))
	}
}

func TestProjectRepository_Members(t *testing.T) {
	pool := setup(t)
	ctx := context.Background()
	owner := createUser(t, pool)
	proj := createProject(t, pool, owner)
	repo := projectpg.NewProjectRepository(pool)

	member1 := createUser(t, pool)
	member2 := createUser(t, pool)

	m := &projectdomain.ProjectMember{ProjectID: proj.ID, UserID: member1.ID, Role: projectdomain.RoleViewer}
	requireNoErr(t, repo.AddMember(ctx, m), "AddMember 1")
	if m.ID == "" {
		t.Fatal("AddMember did not populate member ID")
	}
	requireNoErr(t, repo.AddMember(ctx, &projectdomain.ProjectMember{ProjectID: proj.ID, UserID: member2.ID, Role: projectdomain.RoleMember}), "AddMember 2")

	got, err := repo.GetMember(ctx, proj.ID, member1.ID)
	requireNoErr(t, err, "GetMember")
	if got.Role != projectdomain.RoleViewer || got.UserID != member1.ID {
		t.Fatalf("GetMember returned unexpected member: %+v", got)
	}

	requireNoErr(t, repo.UpdateMemberRole(ctx, proj.ID, member1.ID, projectdomain.RoleAdmin), "UpdateMemberRole")
	updated, _ := repo.GetMember(ctx, proj.ID, member1.ID)
	if updated.Role != projectdomain.RoleAdmin {
		t.Fatalf("UpdateMemberRole did not persist: %+v", updated)
	}

	all, err := repo.ListMemberUsers(ctx, proj.ID)
	requireNoErr(t, err, "ListMemberUsers")
	if len(all) != 3 {
		t.Fatalf("ListMemberUsers = %d members, want 3", len(all))
	}

	requireNoErr(t, repo.RemoveMember(ctx, proj.ID, member2.ID), "RemoveMember")
	if _, err := repo.GetMember(ctx, proj.ID, member2.ID); err == nil {
		t.Fatal("GetMember after RemoveMember = nil error, want not-found")
	}
}

func TestProjectRepository_ListMembers_Pagination(t *testing.T) {
	pool := setup(t)
	ctx := context.Background()
	owner := createUser(t, pool)
	proj := createProject(t, pool, owner)
	repo := projectpg.NewProjectRepository(pool)

	for i := 0; i < 3; i++ {
		member := createUser(t, pool)
		requireNoErr(t, repo.AddMember(ctx, &projectdomain.ProjectMember{
			ProjectID: proj.ID,
			UserID:    member.ID,
			Role:      projectdomain.RoleMember,
		}), "AddMember")
		m, err := repo.GetMember(ctx, proj.ID, member.ID)
		requireNoErr(t, err, "GetMember for timestamp")
		setCreatedAt(t, pool, "project_members", "id", m.ID, 100-i*10)
	}

	page1, cursor, _, err := repo.ListMembers(ctx, proj.ID, "", 2)
	requireNoErr(t, err, "ListMembers page 1")
	if len(page1) != 2 || cursor == "" {
		t.Fatalf("page 1 = %d members, cursor %q; want 2 + cursor", len(page1), cursor)
	}

	page2, cursor2, _, err := repo.ListMembers(ctx, proj.ID, cursor, 2)
	requireNoErr(t, err, "ListMembers page 2")
	// 4 members total (owner + 3 added): page 1 = 2 + cursor, page 2 = 2 final.
	if len(page2) != 2 || cursor2 != "" {
		t.Fatalf("page 2 = %d members, cursor %q; want 2 final page", len(page2), cursor2)
	}
}

func TestConnectionRepository_RoundTrip(t *testing.T) {
	pool := setup(t)
	ctx := context.Background()
	owner := createUser(t, pool)
	proj := createProject(t, pool, owner)
	repo := projectpg.NewConnectionRepository(pool)

	c := &projectdomain.Connection{
		ProjectID:         proj.ID,
		Name:              "Primary DB",
		Host:              "db.internal",
		Port:              5432,
		DatabaseName:      "appdb",
		Username:          "app",
		PasswordEncrypted: "ciphertext",
		SSLMode:           projectdomain.SSLRequire,
		ConnectionStatus:  projectdomain.ConnStatusUnknown,
		CreatedBy:         owner.ID,
	}
	requireNoErr(t, repo.Create(ctx, c), "Create")
	if c.ID == "" {
		t.Fatal("Create did not populate connection ID")
	}

	got, err := repo.GetByID(ctx, c.ID)
	requireNoErr(t, err, "GetByID")
	if got.Host != "db.internal" || got.PasswordEncrypted != "ciphertext" || got.SSLMode != projectdomain.SSLRequire {
		t.Fatalf("GetByID returned unexpected connection: %+v", got)
	}

	got.Name = "Renamed DB"
	got.Port = 6432
	got.PasswordEncrypted = "ciphertext-2"
	requireNoErr(t, repo.Update(ctx, got), "Update")
	updated, _ := repo.GetByID(ctx, c.ID)
	if updated.Name != "Renamed DB" || updated.Port != 6432 || updated.PasswordEncrypted != "ciphertext-2" {
		t.Fatalf("Update did not persist: %+v", updated)
	}

	now := time.Now()
	requireNoErr(t, repo.UpdateStatus(ctx, c.ID, projectdomain.ConnStatusConnected, &now), "UpdateStatus")
	withStatus, _ := repo.GetByID(ctx, c.ID)
	if withStatus.ConnectionStatus != projectdomain.ConnStatusConnected || withStatus.LastConnectedAt == nil {
		t.Fatalf("UpdateStatus did not persist: %+v", withStatus)
	}

	all, err := repo.ListAll(ctx)
	requireNoErr(t, err, "ListAll")
	if len(all) != 1 {
		t.Fatalf("ListAll = %d connections, want 1", len(all))
	}

	requireNoErr(t, repo.SoftDelete(ctx, c.ID), "SoftDelete")
	if _, err := repo.GetByID(ctx, c.ID); err == nil {
		t.Fatal("GetByID after SoftDelete = nil error, want not-found")
	}
	all, _ = repo.ListAll(ctx)
	if len(all) != 0 {
		t.Fatalf("ListAll after SoftDelete = %d connections, want 0", len(all))
	}
}

func TestConnectionRepository_ListByProjectID_Pagination(t *testing.T) {
	pool := setup(t)
	ctx := context.Background()
	owner := createUser(t, pool)
	proj := createProject(t, pool, owner)
	repo := projectpg.NewConnectionRepository(pool)

	for i := 0; i < 5; i++ {
		c := createConnection(t, pool, proj, owner)
		setCreatedAt(t, pool, "connections", "id", c.ID, 100-i*10)
	}

	page1, cursor, _, err := repo.ListByProjectID(ctx, proj.ID, "", 2)
	requireNoErr(t, err, "ListByProjectID page 1")
	if len(page1) != 2 || cursor == "" {
		t.Fatalf("page 1 = %d connections, cursor %q; want 2 + cursor", len(page1), cursor)
	}

	page2, cursor2, _, err := repo.ListByProjectID(ctx, proj.ID, cursor, 2)
	requireNoErr(t, err, "ListByProjectID page 2")
	if len(page2) != 2 || cursor2 == "" {
		t.Fatalf("page 2 = %d connections, cursor %q; want 2 + cursor", len(page2), cursor2)
	}
	if page1[0].ID == page2[0].ID || page1[1].ID == page2[0].ID {
		t.Fatal("pages returned the same connection")
	}

	page3, cursor3, _, err := repo.ListByProjectID(ctx, proj.ID, cursor2, 2)
	requireNoErr(t, err, "ListByProjectID page 3")
	if len(page3) != 1 || cursor3 != "" {
		t.Fatalf("page 3 = %d connections, cursor %q; want 1 final page", len(page3), cursor3)
	}
}

func TestConnectionRepository_ListByProjectID_TimestampTies(t *testing.T) {
	pool := setup(t)
	ctx := context.Background()
	owner := createUser(t, pool)
	proj := createProject(t, pool, owner)
	repo := projectpg.NewConnectionRepository(pool)

	for i := 0; i < 5; i++ {
		createConnection(t, pool, proj, owner)
	}

	all := map[string]bool{}
	cursor := ""
	for page := 0; page < 6; page++ {
		conns, next, _, err := repo.ListByProjectID(ctx, proj.ID, cursor, 2)
		requireNoErr(t, err, "ListByProjectID page")
		if len(conns) == 0 {
			break
		}
		for _, c := range conns {
			if all[c.ID] {
				t.Fatalf("connection %s duplicated across pages", c.ID)
			}
			all[c.ID] = true
		}
		if next == "" {
			break
		}
		cursor = next
	}
	if len(all) != 5 {
		t.Fatalf("covered %d distinct connections, want 5", len(all))
	}
}

func TestProjectRepository_SlugValidation(t *testing.T) {
	pool := setup(t)
	ctx := context.Background()
	owner := createUser(t, pool)
	repo := projectpg.NewProjectRepository(pool)

	slug := strings.ToLower(fmt.Sprintf("slug-%s", newUUID(t)[:8]))
	p1 := &projectdomain.Project{Name: "A", Slug: slug, Visibility: projectdomain.VisibilityPrivate, CreatedBy: owner.ID}
	p2 := &projectdomain.Project{Name: "B", Slug: slug, Visibility: projectdomain.VisibilityPrivate, CreatedBy: owner.ID}

	requireNoErr(t, repo.Create(ctx, p1), "Create first")
	if err := repo.Create(ctx, p2); err == nil {
		t.Fatal("Create duplicate slug = nil error, want uniqueness violation")
	}
}
