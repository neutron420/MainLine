package integration

import (
	"context"
	"fmt"
	"testing"

	migdomain "github.com/schemahub/backend/internal/migration/domain"
	migpg "github.com/schemahub/backend/internal/migration/repository/postgres"
)

// resolverFor returns a connString resolver that maps the given connection ID
// to the executor target database.
func resolverFor(connID, target string) func(ctx context.Context, id string) (string, error) {
	return func(ctx context.Context, id string) (string, error) {
		if id != connID {
			return "", fmt.Errorf("unknown connection %s", id)
		}
		return target, nil
	}
}

func newExecutorService(t *testing.T, connID, target string) (*migdomain.MigrationService, *migpg.MigrationRepository) {
	t.Helper()
	pool := setup(t)
	repo := migpg.NewMigrationRepository(pool)
	svc := migdomain.NewMigrationService(repo, resolverFor(connID, target))
	return svc, repo
}

func TestMigrationExecutor_AppliesMigrationToTarget(t *testing.T) {
	target := ensureTargetDB(t)
	pool := setup(t)
	ctx := context.Background()
	owner := createUser(t, pool)
	proj := createProject(t, pool, owner)
	conn := createConnection(t, pool, proj, owner)
	repo := migpg.NewMigrationRepository(pool)
	svc := migdomain.NewMigrationService(repo, resolverFor(conn.ID, target))

	m := createMigrationRow(t, pool, proj, owner, "1.0")
	m.UpSQL = "CREATE TABLE it_applied (id serial PRIMARY KEY, name text NOT NULL); INSERT INTO it_applied (name) VALUES ('a'), ('b');"
	requireNoErr(t, repo.Update(ctx, m), "update migration SQL")

	run, err := svc.Execute(ctx, m.ID, conn.ID, owner.ID)
	requireNoErr(t, err, "Execute")
	finished := waitForRun(t, repo, run.ID, migdomain.RunStatusCompleted)
	if finished.ErrorMessage != "" {
		t.Fatalf("run error: %s", finished.ErrorMessage)
	}
	if !tableExists(t, target, "it_applied") {
		t.Fatal("it_applied table was not created in target database")
	}

	// Migration should have been finalized as completed.
	mig, err := repo.GetByID(ctx, m.ID)
	requireNoErr(t, err, "GetByID")
	if mig.Status != migdomain.MigrationStatusCompleted {
		t.Fatalf("migration status = %q, want completed", mig.Status)
	}

	// Logs should contain both statements in order.
	logs, err := repo.ListLogsByRunID(ctx, run.ID)
	requireNoErr(t, err, "ListLogsByRunID")
	if len(logs) != 2 {
		t.Fatalf("log entries = %d, want 2", len(logs))
	}
}

func TestMigrationExecutor_RollsBackOnFailure(t *testing.T) {
	target := ensureTargetDB(t)
	pool := setup(t)
	ctx := context.Background()
	owner := createUser(t, pool)
	proj := createProject(t, pool, owner)
	conn := createConnection(t, pool, proj, owner)
	repo := migpg.NewMigrationRepository(pool)
	svc := migdomain.NewMigrationService(repo, resolverFor(conn.ID, target))

	// First statement succeeds, second fails: the whole transaction must
	// roll back, leaving no trace of the first table.
	m := createMigrationRow(t, pool, proj, owner, "2.0")
	m.UpSQL = "CREATE TABLE it_partial (id int); INSERT INTO it_partial VALUES ('not-an-int');"
	requireNoErr(t, repo.Update(ctx, m), "update migration SQL")

	run, err := svc.Execute(ctx, m.ID, conn.ID, owner.ID)
	requireNoErr(t, err, "Execute")
	finished := waitForRun(t, repo, run.ID, migdomain.RunStatusFailed)
	if finished.ErrorMessage == "" {
		t.Fatal("failed run has no error message")
	}
	if tableExists(t, target, "it_partial") {
		t.Fatal("it_partial survived a failed run; transaction did not roll back")
	}

	mig, err := repo.GetByID(ctx, m.ID)
	requireNoErr(t, err, "GetByID")
	if mig.Status != migdomain.MigrationStatusFailed {
		t.Fatalf("migration status = %q, want failed", mig.Status)
	}

	logs, err := repo.ListLogsByRunID(ctx, run.ID)
	requireNoErr(t, err, "ListLogsByRunID")
	if len(logs) == 0 || logs[len(logs)-1].ErrorMessage == "" {
		t.Fatalf("failed statement was not logged: %+v", logs)
	}
}

func TestMigrationExecutor_ConnectionStringError(t *testing.T) {
	pool := setup(t)
	ctx := context.Background()
	owner := createUser(t, pool)
	proj := createProject(t, pool, owner)
	conn := createConnection(t, pool, proj, owner)
	repo := migpg.NewMigrationRepository(pool)
	svc := migdomain.NewMigrationService(repo, func(ctx context.Context, id string) (string, error) {
		return "", fmt.Errorf("no database for %s", id)
	})

	m := createMigrationRow(t, pool, proj, owner, "3.0")
	run, err := svc.Execute(ctx, m.ID, conn.ID, owner.ID)
	requireNoErr(t, err, "Execute")
	finished := waitForRun(t, repo, run.ID, migdomain.RunStatusFailed)
	if finished.ErrorMessage == "" || finished.ErrorMessage != "connection string: no database for "+conn.ID {
		t.Fatalf("unexpected error message: %q", finished.ErrorMessage)
	}
}

func TestMigrationExecutor_RejectsNonDraft(t *testing.T) {
	target := ensureTargetDB(t)
	pool := setup(t)
	ctx := context.Background()
	owner := createUser(t, pool)
	proj := createProject(t, pool, owner)
	conn := createConnection(t, pool, proj, owner)
	repo := migpg.NewMigrationRepository(pool)
	svc := migdomain.NewMigrationService(repo, resolverFor(conn.ID, target))

	m := createMigrationRow(t, pool, proj, owner, "4.0")
	m.Status = migdomain.MigrationStatusCompleted
	requireNoErr(t, repo.Update(ctx, m), "set status")

	if _, err := svc.Execute(ctx, m.ID, conn.ID, owner.ID); err == nil {
		t.Fatal("Execute on completed migration = nil error, want FAILED_PRECONDITION")
	}
}

func TestMigrationService_DryRunAndValidate(t *testing.T) {
	target := ensureTargetDB(t)
	pool := setup(t)
	ctx := context.Background()
	owner := createUser(t, pool)
	proj := createProject(t, pool, owner)
	conn := createConnection(t, pool, proj, owner)
	repo := migpg.NewMigrationRepository(pool)
	svc := migdomain.NewMigrationService(repo, resolverFor(conn.ID, target))

	m := createMigrationRow(t, pool, proj, owner, "5.0")
	m.UpSQL = "CREATE TABLE it_dry (id int);"
	requireNoErr(t, repo.Update(ctx, m), "update migration SQL")

	ok, logs, errs := svc.DryRun(ctx, m.ID, conn.ID)
	if !ok {
		t.Fatalf("DryRun returned ok=false: logs=%v errs=%v", logs, errs)
	}
	// Dry-run must not leave the table behind.
	if tableExists(t, target, "it_dry") {
		t.Fatal("DryRun committed DDL to the target database")
	}

	ok, errs = svc.Validate(ctx, "CREATE TABLE ok (id int);", "DROP TABLE IF EXISTS ok;")
	if !ok || len(errs) != 0 {
		t.Fatalf("Validate(valid) = ok=%v errs=%v", ok, errs)
	}
	_, errs = svc.Validate(ctx, "THIS IS NOT SQL", "")
	if len(errs) == 0 {
		t.Fatal("Validate(invalid) returned no errors")
	}
}
