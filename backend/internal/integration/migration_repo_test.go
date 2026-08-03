package integration

import (
	"context"
	"testing"

	migdomain "github.com/schemahub/backend/internal/migration/domain"
	migpg "github.com/schemahub/backend/internal/migration/repository/postgres"
)

func TestMigrationRepository_ListByProjectID_Pagination(t *testing.T) {
	pool := setup(t)
	ctx := context.Background()
	owner := createUser(t, pool)
	proj := createProject(t, pool, owner)
	repo := migpg.NewMigrationRepository(pool)

	for i := 0; i < 4; i++ {
		m := createMigrationRow(t, pool, proj, owner, versionOf(i))
		setCreatedAt(t, pool, "migrations", "id", m.ID, 100-i*10)
	}

	page1, cursor, _, err := repo.ListByProjectID(ctx, proj.ID, "", 2)
	requireNoErr(t, err, "page 1")
	if len(page1) != 2 || cursor == "" {
		t.Fatalf("page 1 = %d migrations, cursor %q; want 2 + cursor", len(page1), cursor)
	}
	page2, cursor2, _, err := repo.ListByProjectID(ctx, proj.ID, cursor, 2)
	requireNoErr(t, err, "page 2")
	if len(page2) != 2 || cursor2 != "" {
		t.Fatalf("page 2 = %d migrations, cursor %q; want final page (2, no cursor)", len(page2), cursor2)
	}
}

func TestMigrationRepository_ListByProjectID_TimestampTies(t *testing.T) {
	pool := setup(t)
	ctx := context.Background()
	owner := createUser(t, pool)
	proj := createProject(t, pool, owner)
	repo := migpg.NewMigrationRepository(pool)

	for i := 0; i < 5; i++ {
		createMigrationRow(t, pool, proj, owner, versionOf(i))
	}

	all := map[string]bool{}
	cursor := ""
	for page := 0; page < 6; page++ {
		rows, next, _, err := repo.ListByProjectID(ctx, proj.ID, cursor, 2)
		requireNoErr(t, err, "page")
		if len(rows) == 0 {
			break
		}
		for _, m := range rows {
			if all[m.ID] {
				t.Fatalf("migration %s duplicated across pages", m.ID)
			}
			all[m.ID] = true
		}
		if next == "" {
			break
		}
		cursor = next
	}
	if len(all) != 5 {
		t.Fatalf("covered %d distinct migrations, want 5", len(all))
	}
}

func TestMigrationRepository_GetByProjectAndVersion(t *testing.T) {
	pool := setup(t)
	ctx := context.Background()
	owner := createUser(t, pool)
	proj := createProject(t, pool, owner)
	repo := migpg.NewMigrationRepository(pool)

	createMigrationRow(t, pool, proj, owner, "1.0")

	got, err := repo.GetByProjectAndVersion(ctx, proj.ID, "1.0")
	requireNoErr(t, err, "GetByProjectAndVersion")
	if got.Version != "1.0" || got.ProjectID != proj.ID {
		t.Fatalf("unexpected migration: %+v", got)
	}
	if _, err := repo.GetByProjectAndVersion(ctx, proj.ID, "9.9"); err == nil {
		t.Fatal("missing version = nil error, want not-found")
	}
}

func TestMigrationRepository_RunsAndLogs(t *testing.T) {
	pool := setup(t)
	ctx := context.Background()
	owner := createUser(t, pool)
	proj := createProject(t, pool, owner)
	conn := createConnection(t, pool, proj, owner)
	m := createMigrationRow(t, pool, proj, owner, "2.0")
	repo := migpg.NewMigrationRepository(pool)

	run := &migdomain.MigrationRun{
		MigrationID:  m.ID,
		ConnectionID: conn.ID,
		Direction:    migdomain.MigrationDirectionUp,
		Status:       migdomain.RunStatusPending,
		ExecutedBy:   owner.ID,
	}
	requireNoErr(t, repo.CreateRun(ctx, run), "CreateRun")
	if run.ID == "" {
		t.Fatal("CreateRun did not populate run ID")
	}

	got, err := repo.GetRunByID(ctx, run.ID)
	requireNoErr(t, err, "GetRunByID")
	if got.Status != migdomain.RunStatusPending || got.ConnectionID != conn.ID {
		t.Fatalf("unexpected run: %+v", got)
	}

	got.Status = migdomain.RunStatusCompleted
	requireNoErr(t, repo.UpdateRun(ctx, got), "UpdateRun")
	updated, _ := repo.GetRunByID(ctx, run.ID)
	if updated.Status != migdomain.RunStatusCompleted {
		t.Fatalf("UpdateRun did not persist: %+v", updated)
	}

	active, err := repo.GetActiveRunForConnection(ctx, conn.ID)
	requireNoErr(t, err, "GetActiveRunForConnection")
	if active != nil {
		t.Fatalf("completed run reported active: %+v", active)
	}

	run2 := &migdomain.MigrationRun{
		MigrationID:  m.ID,
		ConnectionID: conn.ID,
		Direction:    migdomain.MigrationDirectionDown,
		Status:       migdomain.RunStatusRunning,
		ExecutedBy:   owner.ID,
	}
	requireNoErr(t, repo.CreateRun(ctx, run2), "CreateRun 2")
	active, _ = repo.GetActiveRunForConnection(ctx, conn.ID)
	if active == nil || active.ID != run2.ID {
		t.Fatalf("active run = %+v, want %s", active, run2.ID)
	}

	dur := int32(5)
	rowsAffected := int32(3)
	requireNoErr(t, repo.CreateLogEntry(ctx, &migdomain.MigrationLogEntry{
		MigrationRunID: run.ID,
		Sequence:       1,
		SQL:            "SELECT 1",
		DurationMs:     &dur,
		RowsAffected:   &rowsAffected,
	}), "CreateLogEntry")

	logs, err := repo.ListLogsByRunID(ctx, run.ID)
	requireNoErr(t, err, "ListLogsByRunID")
	if len(logs) != 1 || logs[0].SQL != "SELECT 1" {
		t.Fatalf("logs = %+v", logs)
	}

	runs, cursor, _, err := repo.ListRunsByMigrationID(ctx, m.ID, "", 1)
	requireNoErr(t, err, "ListRunsByMigrationID")
	if len(runs) != 1 || cursor == "" {
		t.Fatalf("runs page 1 = %d, cursor %q; want 1 + cursor", len(runs), cursor)
	}
	rest, cursor2, _, err := repo.ListRunsByMigrationID(ctx, m.ID, cursor, 1)
	requireNoErr(t, err, "ListRunsByMigrationID page 2")
	if len(rest) != 1 || cursor2 != "" {
		t.Fatalf("runs page 2 = %d, cursor %q; want final page", len(rest), cursor2)
	}
}

func TestMigrationRepository_SoftDeleteHides(t *testing.T) {
	pool := setup(t)
	ctx := context.Background()
	owner := createUser(t, pool)
	proj := createProject(t, pool, owner)
	repo := migpg.NewMigrationRepository(pool)

	m := createMigrationRow(t, pool, proj, owner, "3.0")
	requireNoErr(t, repo.SoftDelete(ctx, m.ID), "SoftDelete")
	if _, err := repo.GetByID(ctx, m.ID); err == nil {
		t.Fatal("GetByID after SoftDelete = nil error, want not-found")
	}
	all, _, _, err := repo.ListByProjectID(ctx, proj.ID, "", 10)
	requireNoErr(t, err, "ListByProjectID")
	if len(all) != 0 {
		t.Fatalf("ListByProjectID after SoftDelete = %d, want 0", len(all))
	}
}
