// Package testdb provides a shared harness for integration tests that run
// against a real PostgreSQL instance (see docker/docker-compose.test.yml).
//
// Tests using this package skip automatically when TEST_DATABASE_URL is not
// set, so the regular unit-test run (and CI without a database service)
// stays green.
package testdb

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/schemahub/backend/internal/pkg/database"
)

// URL returns the configured test database URL, or "" when integration tests
// should be skipped.
func URL() string { return os.Getenv("TEST_DATABASE_URL") }

var (
	migrateOnce sync.Once
	migrateErr  error
)

// Setup connects to the test database, applies migrations once per process,
// and truncates all tables so each test starts from a clean slate. It skips
// the test when TEST_DATABASE_URL is unset.
func Setup(t *testing.T) *pgxpool.Pool {
	t.Helper()

	url := URL()
	if url == "" {
		t.Skip("TEST_DATABASE_URL not set; skipping integration test")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	pool, err := database.Connect(ctx, url, 2, 8)
	if err != nil {
		t.Fatalf("connecting to test database: %v", err)
	}
	t.Cleanup(pool.Close)

	migrateOnce.Do(func() {
		migrateErr = database.RunMigrations(ctx, pool, migrationsDir())
	})
	if migrateErr != nil {
		t.Fatalf("applying migrations: %v", migrateErr)
	}

	if err := Truncate(ctx, pool); err != nil {
		t.Fatalf("truncating tables: %v", err)
	}

	return pool
}

// Truncate empties every table except the migration bookkeeping table.
func Truncate(ctx context.Context, pool *pgxpool.Pool) error {
	var names []string
	rows, err := pool.Query(ctx, `
		SELECT tablename FROM pg_tables
		WHERE schemaname = 'public' AND tablename <> 'schema_migrations'`)
	if err != nil {
		return fmt.Errorf("listing tables: %w", err)
	}
	for rows.Next() {
		var n string
		if err := rows.Scan(&n); err != nil {
			rows.Close()
			return err
		}
		names = append(names, n)
	}
	rows.Close()
	if err := rows.Err(); err != nil {
		return err
	}
	if len(names) == 0 {
		return nil
	}
	_, err = pool.Exec(ctx, "TRUNCATE "+strings.Join(names, ", ")+" RESTART IDENTITY CASCADE")
	if err != nil {
		return fmt.Errorf("truncating tables: %w", err)
	}
	return nil
}

// migrationsDir resolves backend/internal/pkg/database/migrations from the
// location of this source file, so tests work regardless of working directory.
func migrationsDir() string {
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		return "internal/pkg/database/migrations"
	}
	dir := filepath.Dir(file)
	for i := 0; i < 6; i++ {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return filepath.Join(dir, "internal", "pkg", "database", "migrations")
		}
		dir = filepath.Dir(dir)
	}
	return "internal/pkg/database/migrations"
}
