package integration

import (
	"context"
	"fmt"
	"net/url"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	authdomain "github.com/schemahub/backend/internal/auth/domain"
	authpg "github.com/schemahub/backend/internal/auth/repository/postgres"
	migdomain "github.com/schemahub/backend/internal/migration/domain"
	migpg "github.com/schemahub/backend/internal/migration/repository/postgres"
	"github.com/schemahub/backend/internal/pkg/testdb"
	projectdomain "github.com/schemahub/backend/internal/project/domain"
	projectpg "github.com/schemahub/backend/internal/project/repository/postgres"
	schemadomain "github.com/schemahub/backend/internal/schema/domain"
	schemapg "github.com/schemahub/backend/internal/schema/repository/postgres"
)

const (
	targetDBName = "schemahub_target"
	testUserRole = "user"
)

// setup returns a pooled connection to the real test database. It skips the
// test when TEST_DATABASE_URL is unset and truncates all tables.
func setup(t *testing.T) *pgxpool.Pool {
	t.Helper()
	return testdb.Setup(t)
}

func newUUID(t *testing.T) string {
	t.Helper()
	return uuid.NewString()
}

func requireNoErr(t *testing.T, err error, msg string) {
	t.Helper()
	if err != nil {
		t.Fatalf("%s: %v", msg, err)
	}
}

func createUser(t *testing.T, pool *pgxpool.Pool) *authdomain.User {
	t.Helper()
	repo := authpg.NewUserRepository(pool)
	u := &authdomain.User{
		Email:        fmt.Sprintf("it-%s@example.com", newUUID(t)),
		PasswordHash: "testhash",
		DisplayName:  "Integration Tester",
		Role:         testUserRole,
		IsActive:     true,
	}
	requireNoErr(t, repo.Create(context.Background(), u), "creating user")
	got, err := repo.GetByEmail(context.Background(), u.Email)
	requireNoErr(t, err, "loading user")
	return got
}

// createProject inserts a project owned by user and makes the user an owner
// member (required for ListByUserID which joins project_members).
func createProject(t *testing.T, pool *pgxpool.Pool, owner *authdomain.User) *projectdomain.Project {
	t.Helper()
	repo := projectpg.NewProjectRepository(pool)
	p := &projectdomain.Project{
		Name:        "Integration Project",
		Slug:        fmt.Sprintf("it-%s", newUUID(t)[:8]),
		Description: "created by integration tests",
		Visibility:  projectdomain.VisibilityPrivate,
		Template:    "blank",
		CreatedBy:   owner.ID,
	}
	requireNoErr(t, repo.Create(context.Background(), p), "creating project")
	requireNoErr(t, repo.AddMember(context.Background(), &projectdomain.ProjectMember{
		ProjectID: p.ID,
		UserID:    owner.ID,
		Role:      projectdomain.RoleOwner,
	}), "adding owner member")
	return p
}

func createConnection(t *testing.T, pool *pgxpool.Pool, proj *projectdomain.Project, by *authdomain.User) *projectdomain.Connection {
	t.Helper()
	repo := projectpg.NewConnectionRepository(pool)
	c := &projectdomain.Connection{
		ProjectID:         proj.ID,
		Name:              "Integration Connection",
		Host:              "127.0.0.1",
		Port:              5432,
		DatabaseName:      "schemahub_test",
		Username:          "postgres",
		PasswordEncrypted: "opaque-plain-password",
		SSLMode:           projectdomain.SSLDisable,
		ConnectionStatus:  projectdomain.ConnStatusUnknown,
		CreatedBy:         by.ID,
	}
	requireNoErr(t, repo.Create(context.Background(), c), "creating connection")
	return c
}

func createSchemaRow(t *testing.T, pool *pgxpool.Pool, proj *projectdomain.Project, conn *projectdomain.Connection) *schemadomain.Schema {
	t.Helper()
	repo := schemapg.NewSchemaRepository(pool)
	s := &schemadomain.Schema{
		ProjectID:    proj.ID,
		ConnectionID: conn.ID,
		SchemaName:   "public",
	}
	requireNoErr(t, repo.Create(context.Background(), s), "creating schema row")
	return s
}

func createMigrationRow(t *testing.T, pool *pgxpool.Pool, proj *projectdomain.Project, by *authdomain.User, version string) *migdomain.Migration {
	t.Helper()
	repo := migpg.NewMigrationRepository(pool)
	m := &migdomain.Migration{
		ProjectID: proj.ID,
		Title:     fmt.Sprintf("Migration %s", version),
		Version:   version,
		UpSQL:     "CREATE TABLE it_migration (id int);",
		Checksum:  "checksum",
		Status:    migdomain.MigrationStatusDraft,
		CreatedBy: by.ID,
	}
	requireNoErr(t, repo.Create(context.Background(), m), "creating migration row")
	return m
}

// setCreatedAt forces a distinct created/updated timestamp so cursor
// pagination tests do not race microsecond-resolution clock values.
func setCreatedAt(t *testing.T, pool *pgxpool.Pool, table, idColumn, id string, secondsAgo int) {
	t.Helper()
	q := fmt.Sprintf("UPDATE %s SET created_at = now() - make_interval(secs => $1) WHERE %s = $2", table, idColumn)
	_, err := pool.Exec(context.Background(), q, secondsAgo, id)
	requireNoErr(t, err, fmt.Sprintf("setting %s.%s timestamp", table, idColumn))
}

// targetDBURL returns the executor target database URL built from the test
// database URL.
func targetDBURL(t *testing.T) string {
	t.Helper()
	u, err := url.Parse(testdb.URL())
	requireNoErr(t, err, "parsing TEST_DATABASE_URL")
	u.Path = "/" + targetDBName
	return u.String()
}

func adminDBURL(t *testing.T) string {
	t.Helper()
	u, err := url.Parse(testdb.URL())
	requireNoErr(t, err, "parsing TEST_DATABASE_URL")
	u.Path = "/postgres"
	return u.String()
}

// ensureTargetDB creates the executor target database idempotently and
// registers a cleanup that drops it again.
func ensureTargetDB(t *testing.T) string {
	t.Helper()
	if testdb.URL() == "" {
		t.Skip("TEST_DATABASE_URL not set; skipping integration test")
	}
	ctx := context.Background()

	admin, err := pgxpool.New(ctx, adminDBURL(t))
	requireNoErr(t, err, "connecting to admin database")
	defer admin.Close()

	var exists bool
	requireNoErr(t, admin.QueryRow(ctx,
		"SELECT EXISTS(SELECT 1 FROM pg_database WHERE datname = $1)", targetDBName).Scan(&exists),
		"checking target database")
	if !exists {
		_, err := admin.Exec(ctx, "CREATE DATABASE "+targetDBName)
		requireNoErr(t, err, "creating target database")
	}

	t.Cleanup(func() {
		ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
		defer cancel()
		admin, err := pgxpool.New(ctx, adminDBURL(t))
		if err != nil {
			return
		}
		defer admin.Close()
		_, _ = admin.Exec(ctx,
			"SELECT pg_terminate_backend(pid) FROM pg_stat_activity WHERE datname = $1 AND pid <> pg_backend_pid()",
			targetDBName)
		_, _ = admin.Exec(ctx, "DROP DATABASE IF EXISTS "+targetDBName)
	})

	return targetDBURL(t)
}

// tableExists reports whether a table exists in the given database.
func tableExists(t *testing.T, connStr, tableName string) bool {
	t.Helper()
	ctx := context.Background()
	pool, err := pgxpool.New(ctx, connStr)
	requireNoErr(t, err, "connecting for table check")
	defer pool.Close()
	var exists bool
	requireNoErr(t, pool.QueryRow(ctx,
		"SELECT to_regclass('public."+tableName+"') IS NOT NULL").Scan(&exists),
		"checking table existence")
	return exists
}

// waitForRun polls a migration run until it reaches the wanted terminal state.
func waitForRun(t *testing.T, repo migdomain.MigrationRepository, runID string, want migdomain.RunStatus) *migdomain.MigrationRun {
	t.Helper()
	deadline := time.Now().Add(15 * time.Second)
	for time.Now().Before(deadline) {
		run, err := repo.GetRunByID(context.Background(), runID)
		if err == nil && run.Status == want {
			return run
		}
		if err == nil && (run.Status == migdomain.RunStatusFailed || run.Status == migdomain.RunStatusCompleted) {
			t.Fatalf("run %s reached %q, want %q: %s", runID, run.Status, want, run.ErrorMessage)
		}
		time.Sleep(10 * time.Millisecond)
	}
	t.Fatalf("run %s never reached status %q", runID, want)
	return nil
}
