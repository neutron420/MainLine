package handler

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/schemahub/backend/internal/migration/domain"
	"github.com/schemahub/backend/internal/pkg/errors"
	"github.com/schemahub/backend/internal/pkg/interceptor"
	migrationv1 "github.com/schemahub/backend/proto/migration/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type fakeMigRepo struct {
	migrations map[string]*domain.Migration
	runs       map[string]*domain.MigrationRun
	logs       []*domain.MigrationLogEntry
	activeRun  *domain.MigrationRun
	notFound   bool
	runSeq     int
	getRunErr  error
	logsErr    error
}

func newFakeMigRepo() *fakeMigRepo {
	return &fakeMigRepo{
		migrations: map[string]*domain.Migration{},
		runs:       map[string]*domain.MigrationRun{},
	}
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
	for _, m := range f.migrations {
		if m.ProjectID == projectID && m.Version == version {
			return m, nil
		}
	}
	return nil, nil
}

func (f *fakeMigRepo) CreateRun(ctx context.Context, r *domain.MigrationRun) error {
	f.runSeq++
	r.ID = fmt.Sprintf("run_%d", f.runSeq)
	f.runs[r.ID] = r
	return nil
}

func (f *fakeMigRepo) GetRunByID(ctx context.Context, id string) (*domain.MigrationRun, error) {
	if f.getRunErr != nil {
		return nil, f.getRunErr
	}
	r, ok := f.runs[id]
	if !ok {
		return nil, errors.New("NOT_FOUND", "run not found")
	}
	return r, nil
}

func (f *fakeMigRepo) UpdateRun(ctx context.Context, r *domain.MigrationRun) error {
	f.runs[r.ID] = r
	return nil
}

func (f *fakeMigRepo) ListRunsByMigrationID(ctx context.Context, migrationID, cursor string, limit int32) ([]*domain.MigrationRun, string, int32, error) {
	var out []*domain.MigrationRun
	for _, r := range f.runs {
		if r.MigrationID == migrationID {
			out = append(out, r)
		}
	}
	return out, "", int32(len(out)), nil
}

func (f *fakeMigRepo) GetActiveRunForConnection(ctx context.Context, connectionID string) (*domain.MigrationRun, error) {
	return f.activeRun, nil
}

func (f *fakeMigRepo) CreateLogEntry(ctx context.Context, entry *domain.MigrationLogEntry) error {
	f.logs = append(f.logs, entry)
	return nil
}

func (f *fakeMigRepo) ListLogsByRunID(ctx context.Context, runID string) ([]*domain.MigrationLogEntry, error) {
	if f.logsErr != nil {
		return nil, f.logsErr
	}
	var out []*domain.MigrationLogEntry
	for _, l := range f.logs {
		if l.MigrationRunID == runID {
			out = append(out, l)
		}
	}
	return out, nil
}

func testMigrationHandler(t *testing.T, repo *fakeMigRepo) *MigrationHandler {
	t.Helper()
	svc := domain.NewMigrationService(repo, nil)
	return NewMigrationHandler(svc)
}

func testMigrationHandlerWithConn(t *testing.T, repo *fakeMigRepo, connString func(context.Context, string) (string, error)) *MigrationHandler {
	t.Helper()
	svc := domain.NewMigrationService(repo, connString)
	return NewMigrationHandler(svc)
}

func failingConnString(ctx context.Context, connID string) (string, error) {
	return "", fmt.Errorf("no database available")
}

func strPtr(s string) *string { return &s }

type fakeMigWatchStream struct {
	grpc.ServerStream
	ctx     context.Context
	mu      sync.Mutex
	status  []*migrationv1.MigrationStatusMessage
	sendErr error
}

func (f *fakeMigWatchStream) Context() context.Context { return f.ctx }

func (f *fakeMigWatchStream) Send(m *migrationv1.MigrationStatusMessage) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.sendErr != nil {
		return f.sendErr
	}
	f.status = append(f.status, m)
	return nil
}

type fakeMigLogsStream struct {
	grpc.ServerStream
	ctx  context.Context
	mu   sync.Mutex
	logs []*migrationv1.MigrationLogEntry
}

func (f *fakeMigLogsStream) Context() context.Context { return f.ctx }

func (f *fakeMigLogsStream) Send(e *migrationv1.MigrationLogEntry) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.logs = append(f.logs, e)
	return nil
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

func TestMigrationHandler_CreateMigrationInvalid(t *testing.T) {
	t.Parallel()

	repo := newFakeMigRepo()
	h := testMigrationHandler(t, repo)

	_, err := h.CreateMigration(context.Background(), &migrationv1.CreateMigrationRequest{
		ProjectId: "proj_1", Version: "2024-01-01",
	})
	if status.Code(err) != codes.InvalidArgument {
		t.Fatalf("CreateMigration() error code = %v, want InvalidArgument (%v)", status.Code(err), err)
	}
}

func TestMigrationHandler_CreateMigrationDuplicateVersion(t *testing.T) {
	t.Parallel()

	repo := newFakeMigRepo()
	repo.migrations["mig_0"] = &domain.Migration{
		ID: "mig_0", ProjectID: "proj_1", Title: "exists", Version: "2024-01-01",
		UpSQL: "CREATE TABLE a (id INT);", Status: domain.MigrationStatusDraft,
		CreatedAt: time.Now(), UpdatedAt: time.Now(),
	}
	h := testMigrationHandler(t, repo)

	_, err := h.CreateMigration(context.Background(), &migrationv1.CreateMigrationRequest{
		ProjectId: "proj_1", Title: "duplicate", Version: "2024-01-01",
		UpSql: "CREATE TABLE b (id INT);",
	})
	if status.Code(err) != codes.AlreadyExists {
		t.Fatalf("CreateMigration() error code = %v, want AlreadyExists (%v)", status.Code(err), err)
	}
}

func TestMigrationHandler_CreateMigrationSetsCreator(t *testing.T) {
	t.Parallel()

	repo := newFakeMigRepo()
	h := testMigrationHandler(t, repo)
	ctx := context.WithValue(context.Background(), interceptor.UserIDKey, "user_1")

	resp, err := h.CreateMigration(ctx, &migrationv1.CreateMigrationRequest{
		ProjectId: "proj_1", Title: "create users", Version: "2024-01-01",
		UpSql: "CREATE TABLE users (id INT);",
	})
	if err != nil {
		t.Fatalf("CreateMigration() error = %v", err)
	}
	if resp.Migration.CreatedBy != "user_1" {
		t.Errorf("Migration.CreatedBy = %q, want user_1", resp.Migration.CreatedBy)
	}
}

func TestMigrationHandler_UpdateMigration(t *testing.T) {
	t.Parallel()

	repo := newFakeMigRepo()
	repo.migrations["mig_1"] = &domain.Migration{
		ID: "mig_1", ProjectID: "proj_1", Title: "old title", Version: "2024-01-01",
		UpSQL: "CREATE TABLE a (id INT);", Status: domain.MigrationStatusDraft,
		CreatedAt: time.Now(), UpdatedAt: time.Now(),
	}
	h := testMigrationHandler(t, repo)

	newUpSQL := "CREATE TABLE users (id INT);"
	resp, err := h.UpdateMigration(context.Background(), &migrationv1.UpdateMigrationRequest{
		Id: "mig_1", Title: strPtr("new title"), UpSql: &newUpSQL,
	})
	if err != nil {
		t.Fatalf("UpdateMigration() error = %v", err)
	}
	if resp.Migration.Title != "new title" {
		t.Errorf("Migration.Title = %q, want new title", resp.Migration.Title)
	}
	if resp.Migration.UpSql != newUpSQL {
		t.Errorf("Migration.UpSql = %q, want %q", resp.Migration.UpSql, newUpSQL)
	}
	if resp.Migration.Checksum != domain.ComputeChecksum(newUpSQL) {
		t.Error("Migration.Checksum not recomputed after UpSql change")
	}
}

func TestMigrationHandler_UpdateMigrationNotFound(t *testing.T) {
	t.Parallel()

	repo := newFakeMigRepo()
	h := testMigrationHandler(t, repo)

	_, err := h.UpdateMigration(context.Background(), &migrationv1.UpdateMigrationRequest{Id: "ghost"})
	if status.Code(err) != codes.NotFound {
		t.Fatalf("UpdateMigration() error code = %v, want NotFound (%v)", status.Code(err), err)
	}
}

func TestMigrationHandler_UpdateMigrationNotDraft(t *testing.T) {
	t.Parallel()

	repo := newFakeMigRepo()
	repo.migrations["mig_1"] = &domain.Migration{
		ID: "mig_1", ProjectID: "proj_1", Title: "t", Version: "2024-01-01",
		UpSQL: "CREATE TABLE a (id INT);", Status: domain.MigrationStatusCompleted,
		CreatedAt: time.Now(), UpdatedAt: time.Now(),
	}
	h := testMigrationHandler(t, repo)

	_, err := h.UpdateMigration(context.Background(), &migrationv1.UpdateMigrationRequest{Id: "mig_1", Title: strPtr("x")})
	if status.Code(err) != codes.FailedPrecondition {
		t.Fatalf("UpdateMigration() error code = %v, want FailedPrecondition (%v)", status.Code(err), err)
	}
}

func TestMigrationHandler_DeleteMigration(t *testing.T) {
	t.Parallel()

	repo := newFakeMigRepo()
	repo.migrations["mig_1"] = &domain.Migration{
		ID: "mig_1", ProjectID: "proj_1", Title: "t", Version: "2024-01-01",
		UpSQL: "CREATE TABLE a (id INT);", Status: domain.MigrationStatusDraft,
		CreatedAt: time.Now(), UpdatedAt: time.Now(),
	}
	h := testMigrationHandler(t, repo)

	if _, err := h.DeleteMigration(context.Background(), &migrationv1.DeleteMigrationRequest{Id: "mig_1"}); err != nil {
		t.Fatalf("DeleteMigration() error = %v", err)
	}
}

func TestMigrationHandler_DeleteMigrationNotFound(t *testing.T) {
	t.Parallel()

	repo := newFakeMigRepo()
	h := testMigrationHandler(t, repo)

	_, err := h.DeleteMigration(context.Background(), &migrationv1.DeleteMigrationRequest{Id: "ghost"})
	if status.Code(err) != codes.NotFound {
		t.Fatalf("DeleteMigration() error code = %v, want NotFound (%v)", status.Code(err), err)
	}
}

func TestMigrationHandler_DeleteMigrationRunning(t *testing.T) {
	t.Parallel()

	repo := newFakeMigRepo()
	repo.migrations["mig_1"] = &domain.Migration{
		ID: "mig_1", ProjectID: "proj_1", Title: "t", Version: "2024-01-01",
		UpSQL: "CREATE TABLE a (id INT);", Status: domain.MigrationStatusRunning,
		CreatedAt: time.Now(), UpdatedAt: time.Now(),
	}
	h := testMigrationHandler(t, repo)

	_, err := h.DeleteMigration(context.Background(), &migrationv1.DeleteMigrationRequest{Id: "mig_1"})
	if status.Code(err) != codes.FailedPrecondition {
		t.Fatalf("DeleteMigration() error code = %v, want FailedPrecondition (%v)", status.Code(err), err)
	}
}

func TestMigrationHandler_ExecuteMigration(t *testing.T) {
	t.Parallel()

	repo := newFakeMigRepo()
	repo.migrations["mig_1"] = &domain.Migration{
		ID: "mig_1", ProjectID: "proj_1", Title: "create users", Version: "2024-01-01",
		UpSQL: "CREATE TABLE users (id INT);", Status: domain.MigrationStatusDraft,
		CreatedAt: time.Now(), UpdatedAt: time.Now(),
	}
	h := testMigrationHandlerWithConn(t, repo, failingConnString)
	ctx := context.WithValue(context.Background(), interceptor.UserIDKey, "user_1")

	resp, err := h.ExecuteMigration(ctx, &migrationv1.ExecuteMigrationRequest{MigrationId: "mig_1", ConnectionId: "conn_1"})
	if err != nil {
		t.Fatalf("ExecuteMigration() error = %v", err)
	}
	if resp.Run.Direction != "up" {
		t.Errorf("Run.Direction = %q, want up", resp.Run.Direction)
	}
	if resp.Run.ExecutedBy != "user_1" {
		t.Errorf("Run.ExecutedBy = %q, want user_1", resp.Run.ExecutedBy)
	}
	if resp.Run.ConnectionId != "conn_1" {
		t.Errorf("Run.ConnectionId = %q, want conn_1", resp.Run.ConnectionId)
	}
}

func TestMigrationHandler_ExecuteMigrationNotFound(t *testing.T) {
	t.Parallel()

	repo := newFakeMigRepo()
	repo.notFound = true
	h := testMigrationHandlerWithConn(t, repo, failingConnString)

	_, err := h.ExecuteMigration(context.Background(), &migrationv1.ExecuteMigrationRequest{MigrationId: "ghost", ConnectionId: "conn_1"})
	if status.Code(err) != codes.NotFound {
		t.Fatalf("ExecuteMigration() error code = %v, want NotFound (%v)", status.Code(err), err)
	}
}

func TestMigrationHandler_ExecuteMigrationNotDraft(t *testing.T) {
	t.Parallel()

	repo := newFakeMigRepo()
	repo.migrations["mig_1"] = &domain.Migration{
		ID: "mig_1", ProjectID: "proj_1", Title: "t", Version: "2024-01-01",
		UpSQL: "CREATE TABLE a (id INT);", Status: domain.MigrationStatusCompleted,
		CreatedAt: time.Now(), UpdatedAt: time.Now(),
	}
	h := testMigrationHandlerWithConn(t, repo, failingConnString)

	_, err := h.ExecuteMigration(context.Background(), &migrationv1.ExecuteMigrationRequest{MigrationId: "mig_1", ConnectionId: "conn_1"})
	if status.Code(err) != codes.FailedPrecondition {
		t.Fatalf("ExecuteMigration() error code = %v, want FailedPrecondition (%v)", status.Code(err), err)
	}
}

func TestMigrationHandler_ExecuteMigrationActiveRun(t *testing.T) {
	t.Parallel()

	repo := newFakeMigRepo()
	repo.migrations["mig_1"] = &domain.Migration{
		ID: "mig_1", ProjectID: "proj_1", Title: "t", Version: "2024-01-01",
		UpSQL: "CREATE TABLE a (id INT);", Status: domain.MigrationStatusDraft,
		CreatedAt: time.Now(), UpdatedAt: time.Now(),
	}
	repo.activeRun = &domain.MigrationRun{ID: "run_active"}
	h := testMigrationHandlerWithConn(t, repo, failingConnString)

	_, err := h.ExecuteMigration(context.Background(), &migrationv1.ExecuteMigrationRequest{MigrationId: "mig_1", ConnectionId: "conn_1"})
	if status.Code(err) != codes.FailedPrecondition {
		t.Fatalf("ExecuteMigration() error code = %v, want FailedPrecondition (%v)", status.Code(err), err)
	}
}

func TestMigrationHandler_WatchMigration(t *testing.T) {
	t.Parallel()

	repo := newFakeMigRepo()
	repo.migrations["mig_1"] = &domain.Migration{
		ID: "mig_1", ProjectID: "proj_1", Title: "create users", Version: "2024-01-01",
		UpSQL: "CREATE TABLE users (id INT);", Status: domain.MigrationStatusDraft,
		CreatedAt: time.Now(), UpdatedAt: time.Now(),
	}
	h := testMigrationHandlerWithConn(t, repo, failingConnString)

	stream := &fakeMigWatchStream{ctx: context.Background()}
	done := make(chan error, 1)
	go func() {
		done <- h.WatchMigration(&migrationv1.WatchMigrationRequest{RunId: "run_1"}, stream)
	}()

	time.Sleep(50 * time.Millisecond)

	if _, err := h.svc.Execute(context.Background(), "mig_1", "conn_1", "user_1"); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	select {
	case err := <-done:
		if err != nil {
			t.Fatalf("WatchMigration() error = %v", err)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("WatchMigration() did not return after a terminal status")
	}

	stream.mu.Lock()
	defer stream.mu.Unlock()
	if len(stream.status) == 0 {
		t.Fatal("WatchMigration() received no status messages")
	}
	last := stream.status[len(stream.status)-1]
	if last.State != "failed" {
		t.Errorf("last status state = %q, want failed", last.State)
	}
	if last.RunId != "run_1" {
		t.Errorf("last status RunId = %q, want run_1", last.RunId)
	}
}

func TestMigrationHandler_WatchMigrationSendError(t *testing.T) {
	t.Parallel()

	repo := newFakeMigRepo()
	repo.migrations["mig_1"] = &domain.Migration{
		ID: "mig_1", ProjectID: "proj_1", Title: "create users", Version: "2024-01-01",
		UpSQL: "CREATE TABLE users (id INT);", Status: domain.MigrationStatusDraft,
		CreatedAt: time.Now(), UpdatedAt: time.Now(),
	}
	h := testMigrationHandlerWithConn(t, repo, failingConnString)

	sendErr := fmt.Errorf("client disconnected")
	stream := &fakeMigWatchStream{ctx: context.Background(), sendErr: sendErr}
	done := make(chan error, 1)
	go func() {
		done <- h.WatchMigration(&migrationv1.WatchMigrationRequest{RunId: "run_1"}, stream)
	}()

	time.Sleep(50 * time.Millisecond)

	if _, err := h.svc.Execute(context.Background(), "mig_1", "conn_1", "user_1"); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	select {
	case err := <-done:
		if err != sendErr {
			t.Errorf("WatchMigration() error = %v, want send error", err)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("WatchMigration() did not return after stream send failure")
	}
}

func TestMigrationHandler_RollbackMigration(t *testing.T) {
	t.Parallel()

	repo := newFakeMigRepo()
	repo.migrations["mig_1"] = &domain.Migration{
		ID: "mig_1", ProjectID: "proj_1", Title: "create users", Version: "2024-01-01",
		UpSQL: "CREATE TABLE users (id INT);", DownSQL: "DROP TABLE users;",
		Status: domain.MigrationStatusDraft, CreatedAt: time.Now(), UpdatedAt: time.Now(),
	}
	repo.runs["run_1"] = &domain.MigrationRun{
		ID: "run_1", MigrationID: "mig_1", ConnectionID: "conn_1",
		Direction: domain.MigrationDirectionUp, Status: domain.RunStatusCompleted,
		ExecutedBy: "user_1", CreatedAt: time.Now(),
	}
	h := testMigrationHandlerWithConn(t, repo, failingConnString)
	ctx := context.WithValue(context.Background(), interceptor.UserIDKey, "user_2")

	resp, err := h.RollbackMigration(ctx, &migrationv1.RollbackMigrationRequest{RunId: "run_1"})
	if err != nil {
		t.Fatalf("RollbackMigration() error = %v", err)
	}
	if resp.Run.Direction != "down" {
		t.Errorf("Run.Direction = %q, want down", resp.Run.Direction)
	}
	if resp.Run.ExecutedBy != "user_2" {
		t.Errorf("Run.ExecutedBy = %q, want user_2", resp.Run.ExecutedBy)
	}
}

func TestMigrationHandler_RollbackMigrationRunNotFound(t *testing.T) {
	t.Parallel()

	repo := newFakeMigRepo()
	h := testMigrationHandlerWithConn(t, repo, failingConnString)

	_, err := h.RollbackMigration(context.Background(), &migrationv1.RollbackMigrationRequest{RunId: "ghost"})
	if status.Code(err) != codes.NotFound {
		t.Fatalf("RollbackMigration() error code = %v, want NotFound (%v)", status.Code(err), err)
	}
}

func TestMigrationHandler_RollbackMigrationNoDownSQL(t *testing.T) {
	t.Parallel()

	repo := newFakeMigRepo()
	repo.migrations["mig_1"] = &domain.Migration{
		ID: "mig_1", ProjectID: "proj_1", Title: "t", Version: "2024-01-01",
		UpSQL: "CREATE TABLE a (id INT);", Status: domain.MigrationStatusDraft,
		CreatedAt: time.Now(), UpdatedAt: time.Now(),
	}
	repo.runs["run_1"] = &domain.MigrationRun{
		ID: "run_1", MigrationID: "mig_1", ConnectionID: "conn_1",
		Direction: domain.MigrationDirectionUp, Status: domain.RunStatusCompleted,
		CreatedAt: time.Now(),
	}
	h := testMigrationHandlerWithConn(t, repo, failingConnString)

	_, err := h.RollbackMigration(context.Background(), &migrationv1.RollbackMigrationRequest{RunId: "run_1"})
	if status.Code(err) != codes.FailedPrecondition {
		t.Fatalf("RollbackMigration() error code = %v, want FailedPrecondition (%v)", status.Code(err), err)
	}
}

func TestMigrationHandler_RollbackMigrationNotCompleted(t *testing.T) {
	t.Parallel()

	repo := newFakeMigRepo()
	repo.migrations["mig_1"] = &domain.Migration{
		ID: "mig_1", ProjectID: "proj_1", Title: "t", Version: "2024-01-01",
		UpSQL: "CREATE TABLE a (id INT);", DownSQL: "DROP TABLE a;",
		Status: domain.MigrationStatusDraft, CreatedAt: time.Now(), UpdatedAt: time.Now(),
	}
	repo.runs["run_1"] = &domain.MigrationRun{
		ID: "run_1", MigrationID: "mig_1", ConnectionID: "conn_1",
		Direction: domain.MigrationDirectionUp, Status: domain.RunStatusFailed,
		CreatedAt: time.Now(),
	}
	h := testMigrationHandlerWithConn(t, repo, failingConnString)

	_, err := h.RollbackMigration(context.Background(), &migrationv1.RollbackMigrationRequest{RunId: "run_1"})
	if status.Code(err) != codes.FailedPrecondition {
		t.Fatalf("RollbackMigration() error code = %v, want FailedPrecondition (%v)", status.Code(err), err)
	}
}

func TestMigrationHandler_WatchRollback(t *testing.T) {
	t.Parallel()

	repo := newFakeMigRepo()
	repo.migrations["mig_1"] = &domain.Migration{
		ID: "mig_1", ProjectID: "proj_1", Title: "create users", Version: "2024-01-01",
		UpSQL: "CREATE TABLE users (id INT);", DownSQL: "DROP TABLE users;",
		Status: domain.MigrationStatusDraft, CreatedAt: time.Now(), UpdatedAt: time.Now(),
	}
	repo.runs["run_1"] = &domain.MigrationRun{
		ID: "run_1", MigrationID: "mig_1", ConnectionID: "conn_1",
		Direction: domain.MigrationDirectionUp, Status: domain.RunStatusCompleted,
		CreatedAt: time.Now(),
	}
	if err := repo.CreateRun(context.Background(), &domain.MigrationRun{
		MigrationID: "mig_1", ConnectionID: "conn_1",
		Direction: domain.MigrationDirectionUp, Status: domain.RunStatusCompleted,
		CreatedAt: time.Now(),
	}); err != nil {
		t.Fatalf("CreateRun() error = %v", err)
	}
	h := testMigrationHandlerWithConn(t, repo, failingConnString)

	stream := &fakeMigWatchStream{ctx: context.Background()}
	done := make(chan error, 1)
	go func() {
		done <- h.WatchRollback(&migrationv1.WatchRollbackRequest{RunId: "run_2"}, stream)
	}()

	time.Sleep(50 * time.Millisecond)

	if _, err := h.svc.Rollback(context.Background(), "run_1", "user_2"); err != nil {
		t.Fatalf("Rollback() error = %v", err)
	}

	select {
	case err := <-done:
		if err != nil {
			t.Fatalf("WatchRollback() error = %v", err)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("WatchRollback() did not return after a terminal status")
	}

	stream.mu.Lock()
	defer stream.mu.Unlock()
	if len(stream.status) == 0 {
		t.Fatal("WatchRollback() received no status messages")
	}
	if last := stream.status[len(stream.status)-1]; last.State != "failed" {
		t.Errorf("last status state = %q, want failed", last.State)
	}
}

func TestMigrationHandler_ValidateMigration(t *testing.T) {
	t.Parallel()

	repo := newFakeMigRepo()
	h := testMigrationHandler(t, repo)

	t.Run("valid sql", func(t *testing.T) {
		resp, err := h.ValidateMigration(context.Background(), &migrationv1.ValidateMigrationRequest{
			UpSql: "CREATE TABLE users (id INT);",
		})
		if err != nil {
			t.Fatalf("ValidateMigration() error = %v", err)
		}
		if !resp.Valid {
			t.Errorf("Valid = false, want true (errors: %v)", resp.Errors)
		}
	})

	t.Run("invalid sql", func(t *testing.T) {
		resp, err := h.ValidateMigration(context.Background(), &migrationv1.ValidateMigrationRequest{
			UpSql: "DROP DATABASE x;",
		})
		if err != nil {
			t.Fatalf("ValidateMigration() error = %v", err)
		}
		if resp.Valid {
			t.Error("Valid = true, want false")
		}
		if len(resp.Errors) == 0 {
			t.Error("Errors empty, want validation errors")
		}
	})
}

func TestMigrationHandler_DryRunMigration(t *testing.T) {
	t.Parallel()

	t.Run("missing migration", func(t *testing.T) {
		repo := newFakeMigRepo()
		repo.notFound = true
		h := testMigrationHandlerWithConn(t, repo, failingConnString)

		resp, err := h.DryRunMigration(context.Background(), &migrationv1.DryRunMigrationRequest{
			MigrationId: "ghost", ConnectionId: "conn_1",
		})
		if err != nil {
			t.Fatalf("DryRunMigration() error = %v", err)
		}
		if resp.Valid {
			t.Error("Valid = true, want false")
		}
		if len(resp.Errors) == 0 {
			t.Error("Errors empty, want error for missing migration")
		}
	})

	t.Run("invalid sql", func(t *testing.T) {
		repo := newFakeMigRepo()
		repo.migrations["mig_1"] = &domain.Migration{
			ID: "mig_1", ProjectID: "proj_1", Title: "t", Version: "2024-01-01",
			UpSQL: "FROBNICATE x;", Status: domain.MigrationStatusDraft,
			CreatedAt: time.Now(), UpdatedAt: time.Now(),
		}
		h := testMigrationHandlerWithConn(t, repo, failingConnString)

		resp, err := h.DryRunMigration(context.Background(), &migrationv1.DryRunMigrationRequest{
			MigrationId: "mig_1", ConnectionId: "conn_1",
		})
		if err != nil {
			t.Fatalf("DryRunMigration() error = %v", err)
		}
		if resp.Valid {
			t.Error("Valid = true, want false")
		}
		if len(resp.Errors) == 0 {
			t.Error("Errors empty, want validation errors")
		}
	})

	t.Run("connection failure", func(t *testing.T) {
		repo := newFakeMigRepo()
		repo.migrations["mig_1"] = &domain.Migration{
			ID: "mig_1", ProjectID: "proj_1", Title: "t", Version: "2024-01-01",
			UpSQL: "CREATE TABLE a (id INT);", Status: domain.MigrationStatusDraft,
			CreatedAt: time.Now(), UpdatedAt: time.Now(),
		}
		h := testMigrationHandlerWithConn(t, repo, failingConnString)

		resp, err := h.DryRunMigration(context.Background(), &migrationv1.DryRunMigrationRequest{
			MigrationId: "mig_1", ConnectionId: "conn_1",
		})
		if err != nil {
			t.Fatalf("DryRunMigration() error = %v", err)
		}
		if resp.Valid {
			t.Error("Valid = true, want false")
		}
		found := false
		for _, e := range resp.Errors {
			if strings.Contains(e, "connection") {
				found = true
			}
		}
		if !found {
			t.Errorf("Errors = %v, want a connection error", resp.Errors)
		}
	})
}

func TestMigrationHandler_GetMigrationRun(t *testing.T) {
	t.Parallel()

	repo := newFakeMigRepo()
	repo.runs["run_1"] = &domain.MigrationRun{
		ID: "run_1", MigrationID: "mig_1", ConnectionID: "conn_1",
		Direction: domain.MigrationDirectionUp, Status: domain.RunStatusCompleted,
		ExecutedBy: "user_1", CreatedAt: time.Now(),
	}
	h := testMigrationHandler(t, repo)

	resp, err := h.GetMigrationRun(context.Background(), &migrationv1.GetMigrationRunRequest{Id: "run_1"})
	if err != nil {
		t.Fatalf("GetMigrationRun() error = %v", err)
	}
	if resp.Run.Id != "run_1" {
		t.Errorf("Run.Id = %q, want run_1", resp.Run.Id)
	}
	if resp.Run.Status != "completed" {
		t.Errorf("Run.Status = %q, want completed", resp.Run.Status)
	}
}

func TestMigrationHandler_GetMigrationRunNotFound(t *testing.T) {
	t.Parallel()

	repo := newFakeMigRepo()
	h := testMigrationHandler(t, repo)

	_, err := h.GetMigrationRun(context.Background(), &migrationv1.GetMigrationRunRequest{Id: "ghost"})
	if status.Code(err) != codes.NotFound {
		t.Fatalf("GetMigrationRun() error code = %v, want NotFound (%v)", status.Code(err), err)
	}
}

func TestMigrationHandler_ListMigrationRuns(t *testing.T) {
	t.Parallel()

	repo := newFakeMigRepo()
	repo.runs["run_1"] = &domain.MigrationRun{ID: "run_1", MigrationID: "mig_1", Direction: domain.MigrationDirectionUp, Status: domain.RunStatusCompleted, CreatedAt: time.Now()}
	repo.runs["run_2"] = &domain.MigrationRun{ID: "run_2", MigrationID: "mig_1", Direction: domain.MigrationDirectionUp, Status: domain.RunStatusFailed, CreatedAt: time.Now()}
	h := testMigrationHandler(t, repo)

	resp, err := h.ListMigrationRuns(context.Background(), &migrationv1.ListMigrationRunsRequest{
		MigrationId: "mig_1", PageSize: 10,
	})
	if err != nil {
		t.Fatalf("ListMigrationRuns() error = %v", err)
	}
	if len(resp.Runs) != 2 {
		t.Errorf("Runs len = %d, want 2", len(resp.Runs))
	}
	if resp.TotalCount != 2 {
		t.Errorf("TotalCount = %d, want 2", resp.TotalCount)
	}
}

func TestMigrationHandler_GetMigrationLogs(t *testing.T) {
	t.Parallel()

	repo := newFakeMigRepo()
	dur := int32(5)
	repo.logs = []*domain.MigrationLogEntry{
		{MigrationRunID: "run_1", Sequence: 1, SQL: "CREATE TABLE a (id INT);", DurationMs: &dur, CreatedAt: time.Now()},
		{MigrationRunID: "run_1", Sequence: 2, SQL: "CREATE INDEX idx_a ON a (id);", CreatedAt: time.Now()},
	}
	h := testMigrationHandler(t, repo)

	stream := &fakeMigLogsStream{ctx: context.Background()}
	if err := h.GetMigrationLogs(&migrationv1.GetMigrationLogsRequest{RunId: "run_1"}, stream); err != nil {
		t.Fatalf("GetMigrationLogs() error = %v", err)
	}
	if len(stream.logs) != 2 {
		t.Fatalf("stream received %d entries, want 2", len(stream.logs))
	}
	if stream.logs[0].Sequence != 1 {
		t.Errorf("first entry Sequence = %d, want 1", stream.logs[0].Sequence)
	}
	if stream.logs[0].DurationMs != 5 {
		t.Errorf("first entry DurationMs = %d, want 5", stream.logs[0].DurationMs)
	}
}

func TestMigrationHandler_GetMigrationLogsError(t *testing.T) {
	t.Parallel()

	repo := newFakeMigRepo()
	repo.logsErr = errors.New("NOT_FOUND", "logs not found")
	h := testMigrationHandler(t, repo)

	stream := &fakeMigLogsStream{ctx: context.Background()}
	err := h.GetMigrationLogs(&migrationv1.GetMigrationLogsRequest{RunId: "run_1"}, stream)
	if status.Code(err) != codes.NotFound {
		t.Fatalf("GetMigrationLogs() error code = %v, want NotFound (%v)", status.Code(err), err)
	}
}
