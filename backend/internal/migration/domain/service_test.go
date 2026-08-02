package domain

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"
)

type fakeMigrationRepo struct {
	mu         sync.Mutex
	migrations map[string]*Migration
	runs       map[string]*MigrationRun
	activeRun  *MigrationRun
	logs       []*MigrationLogEntry
}

func (f *fakeMigrationRepo) Create(ctx context.Context, m *Migration) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.migrations == nil {
		f.migrations = map[string]*Migration{}
	}
	f.migrations[m.ID] = m
	return nil
}

func (f *fakeMigrationRepo) GetByID(ctx context.Context, id string) (*Migration, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	m, ok := f.migrations[id]
	if !ok {
		return nil, errors.New("migration not found")
	}
	return m, nil
}

func (f *fakeMigrationRepo) ListByProjectID(ctx context.Context, projectID, cursor string, limit int32) ([]*Migration, string, int32, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	var out []*Migration
	for _, m := range f.migrations {
		if m.ProjectID == projectID {
			out = append(out, m)
		}
	}
	return out, "", int32(len(out)), nil
}

func (f *fakeMigrationRepo) Update(ctx context.Context, m *Migration) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.migrations[m.ID] = m
	return nil
}

func (f *fakeMigrationRepo) SoftDelete(ctx context.Context, id string) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	if m, ok := f.migrations[id]; ok {
		m.Status = MigrationStatusRolledBack
	}
	return nil
}

func (f *fakeMigrationRepo) GetByProjectAndVersion(ctx context.Context, projectID, version string) (*Migration, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	for _, m := range f.migrations {
		if m.ProjectID == projectID && m.Version == version {
			return m, nil
		}
	}
	return nil, nil
}

func (f *fakeMigrationRepo) CreateRun(ctx context.Context, r *MigrationRun) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.runs == nil {
		f.runs = map[string]*MigrationRun{}
	}
	f.runs[r.ID] = r
	return nil
}

func (f *fakeMigrationRepo) GetRunByID(ctx context.Context, id string) (*MigrationRun, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	r, ok := f.runs[id]
	if !ok {
		return nil, errors.New("run not found")
	}
	return r, nil
}

func (f *fakeMigrationRepo) UpdateRun(ctx context.Context, r *MigrationRun) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.runs[r.ID] = r
	return nil
}

func (f *fakeMigrationRepo) ListRunsByMigrationID(ctx context.Context, migrationID, cursor string, limit int32) ([]*MigrationRun, string, int32, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	var out []*MigrationRun
	for _, r := range f.runs {
		if r.MigrationID == migrationID {
			out = append(out, r)
		}
	}
	return out, "", int32(len(out)), nil
}

func (f *fakeMigrationRepo) GetActiveRunForConnection(ctx context.Context, connectionID string) (*MigrationRun, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.activeRun, nil
}

func (f *fakeMigrationRepo) CreateLogEntry(ctx context.Context, entry *MigrationLogEntry) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.logs = append(f.logs, entry)
	return nil
}

func (f *fakeMigrationRepo) ListLogsByRunID(ctx context.Context, runID string) ([]*MigrationLogEntry, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	var out []*MigrationLogEntry
	for _, l := range f.logs {
		if l.MigrationRunID == runID {
			out = append(out, l)
		}
	}
	return out, nil
}

func newTestMigration(id, projectID, version, upSQL string) *Migration {
	return &Migration{ID: id, ProjectID: projectID, Title: "test", Version: version, UpSQL: upSQL, Status: MigrationStatusDraft}
}

// failingConnString always errors so the async executor finishes without a DB.
func failingConnString(ctx context.Context, connID string) (string, error) {
	return "", errors.New("no database available")
}

// waitForRun polls until the run reaches a terminal state, so tests don't race
// the spawned executor goroutine.
func waitForRun(t *testing.T, repo *fakeMigrationRepo, runID string, want RunStatus) *MigrationRun {
	t.Helper()
	deadline := time.Now().Add(3 * time.Second)
	for time.Now().Before(deadline) {
		run, err := repo.GetRunByID(context.Background(), runID)
		if err == nil && run.Status == want {
			return run
		}
		time.Sleep(5 * time.Millisecond)
	}
	t.Fatalf("run %s never reached status %q", runID, want)
	return nil
}

func TestMigrationService_Create(t *testing.T) {
	t.Parallel()

	t.Run("valid", func(t *testing.T) {
		repo := &fakeMigrationRepo{}
		svc := NewMigrationService(repo, failingConnString)

		m := newTestMigration("m1", "p1", "001", "CREATE TABLE a (id int);")
		created, err := svc.Create(context.Background(), m)
		if err != nil {
			t.Fatalf("Create() error = %v", err)
		}
		if created.Status != MigrationStatusDraft {
			t.Errorf("Status = %q, want draft", created.Status)
		}
		if created.Checksum == "" {
			t.Error("Checksum empty, want computed")
		}
	})

	t.Run("validation error", func(t *testing.T) {
		repo := &fakeMigrationRepo{}
		svc := NewMigrationService(repo, failingConnString)
		if _, err := svc.Create(context.Background(), &Migration{}); err == nil {
			t.Error("Create(invalid) = nil error, want INVALID_ARGUMENT")
		}
	})

	t.Run("duplicate version", func(t *testing.T) {
		repo := &fakeMigrationRepo{
			migrations: map[string]*Migration{"m1": newTestMigration("m1", "p1", "001", "x")},
		}
		svc := NewMigrationService(repo, failingConnString)

		dup := newTestMigration("m2", "p1", "001", "CREATE TABLE b (id int);")
		if _, err := svc.Create(context.Background(), dup); err == nil {
			t.Error("Create(duplicate version) = nil error, want ALREADY_EXISTS")
		}
	})
}

func TestMigrationService_Update(t *testing.T) {
	t.Parallel()

	t.Run("updates draft", func(t *testing.T) {
		repo := &fakeMigrationRepo{
			migrations: map[string]*Migration{"m1": newTestMigration("m1", "p1", "001", "OLD SQL;")},
		}
		svc := NewMigrationService(repo, failingConnString)

		updated, err := svc.Update(context.Background(), &Migration{ID: "m1", Title: "New title", UpSQL: "CREATE TABLE c (id int);"})
		if err != nil {
			t.Fatalf("Update() error = %v", err)
		}
		if updated.Title != "New title" {
			t.Errorf("Title = %q, want New title", updated.Title)
		}
		if updated.UpSQL != "CREATE TABLE c (id int);" {
			t.Errorf("UpSQL not updated")
		}
		if updated.Checksum == "" || updated.Checksum == ComputeChecksum("OLD SQL;") {
			t.Error("Checksum not recomputed after UpSQL change")
		}
	})

	t.Run("cannot update completed", func(t *testing.T) {
		m := newTestMigration("m1", "p1", "001", "x")
		m.Status = MigrationStatusCompleted
		repo := &fakeMigrationRepo{migrations: map[string]*Migration{"m1": m}}
		svc := NewMigrationService(repo, failingConnString)

		if _, err := svc.Update(context.Background(), &Migration{ID: "m1", Title: "t"}); err == nil {
			t.Error("Update(completed) = nil error, want FAILED_PRECONDITION")
		}
	})

	t.Run("missing migration", func(t *testing.T) {
		svc := NewMigrationService(&fakeMigrationRepo{}, failingConnString)
		if _, err := svc.Update(context.Background(), &Migration{ID: "nope"}); err == nil {
			t.Error("Update(missing) = nil error, want error")
		}
	})
}

func TestMigrationService_Delete(t *testing.T) {
	t.Parallel()

	t.Run("draft can be deleted", func(t *testing.T) {
		repo := &fakeMigrationRepo{
			migrations: map[string]*Migration{"m1": newTestMigration("m1", "p1", "001", "x")},
		}
		svc := NewMigrationService(repo, failingConnString)
		if err := svc.Delete(context.Background(), "m1"); err != nil {
			t.Fatalf("Delete() error = %v", err)
		}
	})

	t.Run("running cannot be deleted", func(t *testing.T) {
		m := newTestMigration("m1", "p1", "001", "x")
		m.Status = MigrationStatusRunning
		repo := &fakeMigrationRepo{migrations: map[string]*Migration{"m1": m}}
		svc := NewMigrationService(repo, failingConnString)
		if err := svc.Delete(context.Background(), "m1"); err == nil {
			t.Error("Delete(running) = nil error, want FAILED_PRECONDITION")
		}
	})
}

func TestMigrationService_Execute(t *testing.T) {
	t.Parallel()

	t.Run("starts run and executor fails connection", func(t *testing.T) {
		m := newTestMigration("m1", "p1", "001", "CREATE TABLE a (id int);")
		m.Status = MigrationStatusDraft
		repo := &fakeMigrationRepo{
			migrations: map[string]*Migration{"m1": m},
			runs:       map[string]*MigrationRun{},
		}
		svc := NewMigrationService(repo, failingConnString)

		run, err := svc.Execute(context.Background(), "m1", "conn_1", "user_1")
		if err != nil {
			t.Fatalf("Execute() error = %v", err)
		}
		if run.Status != RunStatusPending {
			t.Errorf("run Status = %q, want pending", run.Status)
		}
		if run.Direction != MigrationDirectionUp {
			t.Errorf("direction = %q, want up", run.Direction)
		}
		if run.ExecutedBy != "user_1" {
			t.Errorf("ExecutedBy = %q, want user_1", run.ExecutedBy)
		}

		// executor goroutine must fail gracefully (no DB) and finalize the migration
		waitForRun(t, repo, run.ID, RunStatusFailed)
		got, _ := repo.GetByID(context.Background(), "m1")
		if got.Status != MigrationStatusFailed {
			t.Errorf("migration Status = %q, want failed", got.Status)
		}
	})

	t.Run("completed cannot execute", func(t *testing.T) {
		m := newTestMigration("m1", "p1", "001", "x")
		m.Status = MigrationStatusCompleted
		repo := &fakeMigrationRepo{
			migrations: map[string]*Migration{"m1": m},
			runs:       map[string]*MigrationRun{},
		}
		svc := NewMigrationService(repo, failingConnString)
		if _, err := svc.Execute(context.Background(), "m1", "conn_1", "u"); err == nil {
			t.Error("Execute(completed) = nil error, want FAILED_PRECONDITION")
		}
	})

	t.Run("active run blocks", func(t *testing.T) {
		m := newTestMigration("m1", "p1", "001", "x")
		m.Status = MigrationStatusDraft
		repo := &fakeMigrationRepo{
			migrations: map[string]*Migration{"m1": m},
			runs:       map[string]*MigrationRun{},
			activeRun:  &MigrationRun{ID: "run_active"},
		}
		svc := NewMigrationService(repo, failingConnString)
		if _, err := svc.Execute(context.Background(), "m1", "conn_1", "u"); err == nil {
			t.Error("Execute(active run) = nil error, want FAILED_PRECONDITION")
		}
	})
}

func TestMigrationService_Rollback(t *testing.T) {
	t.Parallel()

	completedRun := &MigrationRun{ID: "run_1", MigrationID: "m1", ConnectionID: "conn_1", Status: RunStatusCompleted}

	t.Run("rolls back completed run", func(t *testing.T) {
		m := newTestMigration("m1", "p1", "001", "x")
		m.DownSQL = "DROP TABLE a;"
		repo := &fakeMigrationRepo{
			migrations: map[string]*Migration{"m1": m},
			runs:       map[string]*MigrationRun{"run_1": completedRun},
		}
		svc := NewMigrationService(repo, failingConnString)

		run, err := svc.Rollback(context.Background(), "run_1", "user_2")
		if err != nil {
			t.Fatalf("Rollback() error = %v", err)
		}
		if run.Direction != MigrationDirectionDown {
			t.Errorf("direction = %q, want down", run.Direction)
		}
		if run.ExecutedBy != "user_2" {
			t.Errorf("ExecutedBy = %q, want user_2", run.ExecutedBy)
		}
		waitForRun(t, repo, run.ID, RunStatusFailed)
	})

	t.Run("missing down sql", func(t *testing.T) {
		m := newTestMigration("m1", "p1", "001", "x")
		repo := &fakeMigrationRepo{
			migrations: map[string]*Migration{"m1": m},
			runs:       map[string]*MigrationRun{"run_1": completedRun},
		}
		svc := NewMigrationService(repo, failingConnString)
		if _, err := svc.Rollback(context.Background(), "run_1", "u"); err == nil {
			t.Error("Rollback(no down sql) = nil error, want FAILED_PRECONDITION")
		}
	})

	t.Run("only completed runs", func(t *testing.T) {
		m := newTestMigration("m1", "p1", "001", "x")
		m.DownSQL = "DROP TABLE a;"
		failedRun := &MigrationRun{ID: "run_2", MigrationID: "m1", Status: RunStatusFailed}
		repo := &fakeMigrationRepo{
			migrations: map[string]*Migration{"m1": m},
			runs:       map[string]*MigrationRun{"run_2": failedRun},
		}
		svc := NewMigrationService(repo, failingConnString)
		if _, err := svc.Rollback(context.Background(), "run_2", "u"); err == nil {
			t.Error("Rollback(failed run) = nil error, want FAILED_PRECONDITION")
		}
	})
}

func TestMigrationService_Watchers(t *testing.T) {
	t.Parallel()

	svc := NewMigrationService(&fakeMigrationRepo{}, failingConnString)

	ch1 := svc.Subscribe("run_1")
	ch2 := svc.Subscribe("run_1")

	svc.broadcast("run_1", &MigrationStatusMessage{RunID: "run_1", State: RunStatusRunning})
	svc.broadcast("run_1", &MigrationStatusMessage{RunID: "run_1", State: RunStatusRunning})

	if len(ch1) != 2 {
		t.Errorf("ch1 buffered %d messages, want 2", len(ch1))
	}
	if len(ch2) != 2 {
		t.Errorf("ch2 buffered %d messages, want 2", len(ch2))
	}

	// broadcast to a run with no watchers must not block
	svc.broadcast("no_watchers", &MigrationStatusMessage{RunID: "x"})

	svc.Unsubscribe("run_1", ch1)

	// drain the two buffered messages, then the closed channel signals ok=false
	for i := 0; i < 2; i++ {
		<-ch1
	}
	if _, ok := <-ch1; ok {
		t.Error("ch1 not closed after Unsubscribe")
	}
	if len(svc.watchers["run_1"]) != 1 {
		t.Errorf("watchers after unsubscribe = %d, want 1", len(svc.watchers["run_1"]))
	}
}

func TestMigrationService_ValidateAndDryRun(t *testing.T) {
	t.Parallel()

	t.Run("validate", func(t *testing.T) {
		svc := NewMigrationService(&fakeMigrationRepo{}, failingConnString)
		ok, _ := svc.Validate(context.Background(), "CREATE TABLE a (id int);", "")
		if !ok {
			t.Error("Validate(valid) = false, want true")
		}
		ok, errs := svc.Validate(context.Background(), "", "")
		if ok || len(errs) == 0 {
			t.Error("Validate(empty) should fail with errors")
		}
	})

	t.Run("dry run missing migration", func(t *testing.T) {
		svc := NewMigrationService(&fakeMigrationRepo{}, failingConnString)
		ok, errs, _ := svc.DryRun(context.Background(), "missing", "conn_1")
		if ok || len(errs) == 0 {
			t.Error("DryRun(missing) should fail")
		}
	})

	t.Run("dry run invalid sql", func(t *testing.T) {
		repo := &fakeMigrationRepo{
			migrations: map[string]*Migration{"m1": newTestMigration("m1", "p1", "001", "FROBNICATE x;")},
		}
		svc := NewMigrationService(repo, failingConnString)
		ok, errs, _ := svc.DryRun(context.Background(), "m1", "conn_1")
		if ok || len(errs) == 0 {
			t.Error("DryRun(invalid sql) should fail validation")
		}
	})
}

func TestMigrationService_ListDefaults(t *testing.T) {
	t.Parallel()

	t.Run("list migrations", func(t *testing.T) {
		repo := &fakeMigrationRepo{
			migrations: map[string]*Migration{
				"m1": newTestMigration("m1", "p1", "001", "x"),
				"m2": newTestMigration("m2", "p1", "002", "x"),
			},
		}
		svc := NewMigrationService(repo, failingConnString)
		ms, _, total, err := svc.ListByProject(context.Background(), "p1", "", 0)
		if err != nil {
			t.Fatal(err)
		}
		if len(ms) != 2 || total != 2 {
			t.Errorf("List() = %d, total %d; want 2, 2", len(ms), total)
		}
	})

	t.Run("list runs", func(t *testing.T) {
		repo := &fakeMigrationRepo{
			runs: map[string]*MigrationRun{
				"r1": {ID: "r1", MigrationID: "m1"},
				"r2": {ID: "r2", MigrationID: "m1"},
			},
		}
		svc := NewMigrationService(repo, failingConnString)
		runs, _, total, err := svc.ListRuns(context.Background(), "m1", "", 0)
		if err != nil {
			t.Fatal(err)
		}
		if len(runs) != 2 || total != 2 {
			t.Errorf("ListRuns() = %d, total %d; want 2, 2", len(runs), total)
		}
	})

	t.Run("get run", func(t *testing.T) {
		repo := &fakeMigrationRepo{runs: map[string]*MigrationRun{"r1": {ID: "r1"}}}
		svc := NewMigrationService(repo, failingConnString)
		run, err := svc.GetRunByID(context.Background(), "r1")
		if err != nil || run.ID != "r1" {
			t.Errorf("GetRunByID() = %+v, err %v", run, err)
		}
	})
}

func TestSplitSQLStatements(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		sql  string
		want int
	}{
		{name: "one", sql: "CREATE TABLE a (id int);", want: 1},
		{name: "two", sql: "CREATE TABLE a (id int); CREATE TABLE b (id int);", want: 2},
		{name: "quoted semicolon", sql: "INSERT INTO t VALUES ('a;b');", want: 1},
		{name: "comment", sql: "CREATE TABLE a (id int); -- done\n", want: 1},
		{name: "empty", sql: "", want: 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := len(splitSQLStatements(tt.sql)); got != tt.want {
				t.Errorf("splitSQLStatements() = %d, want %d", got, tt.want)
			}
		})
	}
}
