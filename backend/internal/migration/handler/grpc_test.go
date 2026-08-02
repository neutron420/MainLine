package handler

import (
	"context"
	"testing"
	"time"

	"github.com/schemahub/backend/internal/migration/domain"
	"github.com/schemahub/backend/internal/pkg/errors"
	migrationv1 "github.com/schemahub/backend/proto/migration/v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type fakeMigRepo struct {
	migrations map[string]*domain.Migration
	notFound   bool
}

func newFakeMigRepo() *fakeMigRepo {
	return &fakeMigRepo{migrations: map[string]*domain.Migration{}}
}

func (f *fakeMigRepo) Create(ctx context.Context, m *domain.Migration) error {
	m.ID = "mig_1"
	f.migrations[m.ID] = m
	return nil
}

func (f *fakeMigRepo) GetByID(ctx context.Context, id string) (*domain.Migration, error) {
	if f.notFound {
		return nil, errors.New("NOT_FOUND", "migration not found")
	}
	m, ok := f.migrations[id]
	if !ok {
		return nil, errors.New("NOT_FOUND", "migration not found")
	}
	return m, nil
}

func (f *fakeMigRepo) ListByProjectID(ctx context.Context, projectID, cursor string, limit int32) ([]*domain.Migration, string, int32, error) {
	var out []*domain.Migration
	for _, m := range f.migrations {
		if m.ProjectID == projectID {
			out = append(out, m)
		}
	}
	return out, "", int32(len(out)), nil
}

func (f *fakeMigRepo) Update(ctx context.Context, m *domain.Migration) error {
	f.migrations[m.ID] = m
	return nil
}

func (f *fakeMigRepo) SoftDelete(ctx context.Context, id string) error { return nil }

func (f *fakeMigRepo) GetByProjectAndVersion(ctx context.Context, projectID, version string) (*domain.Migration, error) {
	return nil, errors.New("NOT_FOUND", "not found")
}

func (f *fakeMigRepo) CreateRun(ctx context.Context, r *domain.MigrationRun) error { return nil }
func (f *fakeMigRepo) GetRunByID(ctx context.Context, id string) (*domain.MigrationRun, error) {
	return nil, errors.New("NOT_FOUND", "run not found")
}
func (f *fakeMigRepo) UpdateRun(ctx context.Context, r *domain.MigrationRun) error { return nil }
func (f *fakeMigRepo) ListRunsByMigrationID(ctx context.Context, migrationID, cursor string, limit int32) ([]*domain.MigrationRun, string, int32, error) {
	return nil, "", 0, nil
}
func (f *fakeMigRepo) GetActiveRunForConnection(ctx context.Context, connectionID string) (*domain.MigrationRun, error) {
	return nil, errors.New("NOT_FOUND", "no active run")
}
func (f *fakeMigRepo) CreateLogEntry(ctx context.Context, entry *domain.MigrationLogEntry) error {
	return nil
}
func (f *fakeMigRepo) ListLogsByRunID(ctx context.Context, runID string) ([]*domain.MigrationLogEntry, error) {
	return nil, nil
}

func testMigrationHandler(t *testing.T, repo *fakeMigRepo) *MigrationHandler {
	t.Helper()
	svc := domain.NewMigrationService(repo, nil)
	return NewMigrationHandler(svc)
}

func TestMigrationHandler_ListMigrations(t *testing.T) {
	t.Parallel()

	repo := newFakeMigRepo()
	repo.migrations["mig_1"] = &domain.Migration{
		ID: "mig_1", ProjectID: "proj_1", Title: "create users", Version: "2024-01-01",
		Status: domain.MigrationStatusDraft, CreatedAt: time.Now(), UpdatedAt: time.Now(),
	}
	repo.migrations["mig_2"] = &domain.Migration{
		ID: "mig_2", ProjectID: "proj_1", Title: "add indexes", Version: "2024-01-02",
		Status: domain.MigrationStatusDraft, CreatedAt: time.Now(), UpdatedAt: time.Now(),
	}
	h := testMigrationHandler(t, repo)

	resp, err := h.ListMigrations(context.Background(), &migrationv1.ListMigrationsRequest{
		ProjectId: "proj_1", PageSize: 10,
	})
	if err != nil {
		t.Fatalf("ListMigrations() error = %v", err)
	}
	if len(resp.Migrations) != 2 {
		t.Errorf("Migrations len = %d, want 2", len(resp.Migrations))
	}
	if resp.TotalCount != 2 {
		t.Errorf("TotalCount = %d, want 2", resp.TotalCount)
	}
}

func TestMigrationHandler_GetMigration(t *testing.T) {
	t.Parallel()

	repo := newFakeMigRepo()
	repo.migrations["mig_1"] = &domain.Migration{
		ID: "mig_1", ProjectID: "proj_1", Title: "create users", Version: "2024-01-01",
		UpSQL: "CREATE TABLE users (id INT);", Status: domain.MigrationStatusDraft,
		CreatedAt: time.Now(), UpdatedAt: time.Now(),
	}
	h := testMigrationHandler(t, repo)

	resp, err := h.GetMigration(context.Background(), &migrationv1.GetMigrationRequest{Id: "mig_1"})
	if err != nil {
		t.Fatalf("GetMigration() error = %v", err)
	}
	if resp.Migration.Title != "create users" {
		t.Errorf("Migration.Title = %q, want create users", resp.Migration.Title)
	}
}

func TestMigrationHandler_GetMigrationNotFound(t *testing.T) {
	t.Parallel()

	repo := newFakeMigRepo()
	repo.notFound = true
	h := testMigrationHandler(t, repo)

	_, err := h.GetMigration(context.Background(), &migrationv1.GetMigrationRequest{Id: "ghost"})
	if status.Code(err) != codes.NotFound {
		t.Fatalf("GetMigration() error code = %v, want NotFound (%v)", status.Code(err), err)
	}
}

func TestMigrationHandler_CreateMigration(t *testing.T) {
	t.Parallel()

	repo := newFakeMigRepo()
	h := testMigrationHandler(t, repo)

	resp, err := h.CreateMigration(context.Background(), &migrationv1.CreateMigrationRequest{
		ProjectId: "proj_1", Title: "create users", Version: "2024-01-01",
		UpSql: "CREATE TABLE users (id INT);",
	})
	if err != nil {
		t.Fatalf("CreateMigration() error = %v", err)
	}
	if resp.Migration.Id != "mig_1" {
		t.Errorf("Migration.Id = %q, want mig_1", resp.Migration.Id)
	}
	if resp.Migration.Status != "draft" {
		t.Errorf("Migration.Status = %q, want draft", resp.Migration.Status)
	}
	if resp.Migration.Checksum == "" {
		t.Error("expected non-empty checksum")
	}
}
