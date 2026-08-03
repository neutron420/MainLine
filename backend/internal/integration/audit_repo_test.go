package integration

import (
	"context"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	auditdomain "github.com/schemahub/backend/internal/audit/domain"
	auditpg "github.com/schemahub/backend/internal/audit/repository/postgres"
)

func insertAudit(t *testing.T, pool *pgxpool.Pool, repo *auditpg.AuditRepository, evtType, actorID, actorEmail, resourceType, resourceID string, daysAgo int) string {
	t.Helper()
	ctx := context.Background()
	traceID := newUUID(t)
	entry := &auditdomain.AuditEntry{
		EventType:    evtType,
		ActorID:      actorID,
		ActorEmail:   actorEmail,
		Action:       "do_something",
		ResourceType: resourceType,
		ResourceID:   resourceID,
		Metadata:     map[string]string{"k": "v"},
		IPAddress:    "10.0.0.1",
		UserAgent:    "test-agent",
		TraceID:      traceID,
	}
	requireNoErr(t, repo.Insert(ctx, entry), "Insert")

	var id string
	requireNoErr(t, pool.QueryRow(ctx, "SELECT id FROM audit_logs WHERE trace_id = $1", traceID).Scan(&id), "fetching audit id")
	if daysAgo > 0 {
		_, err := pool.Exec(ctx, "UPDATE audit_logs SET created_at = now() - make_interval(days => $1) WHERE id = $2", daysAgo, id)
		requireNoErr(t, err, "aging audit row")
	}
	return id
}

func TestAuditRepository_RoundTripAndFilters(t *testing.T) {
	pool := setup(t)
	ctx := context.Background()
	repo := auditpg.NewAuditRepository(pool)
	userA := createUser(t, pool)
	userB := createUser(t, pool)

	id := insertAudit(t, pool, repo, "migration.run", userA.ID, userA.Email, "migration", newUUID(t), 0)
	id2 := insertAudit(t, pool, repo, "project.create", userB.ID, userB.Email, "project", newUUID(t), 2)
	insertAudit(t, pool, repo, "schema.introspect", userA.ID, userA.Email, "schema", newUUID(t), 8)

	got, err := repo.GetByID(ctx, id)
	requireNoErr(t, err, "GetByID")
	if got.EventType != "migration.run" || got.ActorID != userA.ID || got.Metadata["k"] != "v" {
		t.Fatalf("unexpected entry: %+v", got)
	}

	filter := &auditdomain.AuditFilter{EventType: "migration.run"}
	entries, _, _, err := repo.List(ctx, filter, "", 10)
	requireNoErr(t, err, "List by event type")
	if len(entries) != 1 || entries[0].ID != id {
		t.Fatalf("event-type filter = %+v", entries)
	}

	filter = &auditdomain.AuditFilter{ActorID: userA.ID}
	entries, _, _, _ = repo.List(ctx, filter, "", 10)
	if len(entries) != 2 {
		t.Fatalf("actor filter = %d entries, want 2", len(entries))
	}

	filter = &auditdomain.AuditFilter{ResourceType: "project"}
	entries, _, _, _ = repo.List(ctx, filter, "", 10)
	if len(entries) != 1 || entries[0].ID != id2 {
		t.Fatalf("resource filter = %+v", entries)
	}

	filter = &auditdomain.AuditFilter{DateFrom: time.Now().Add(-24 * time.Hour).UTC().Format(time.RFC3339)}
	entries, _, _, _ = repo.List(ctx, filter, "", 10)
	if len(entries) != 1 {
		t.Fatalf("date-from filter = %d entries, want 1", len(entries))
	}

	filter = &auditdomain.AuditFilter{DateTo: time.Now().Add(-24 * time.Hour).UTC().Format(time.RFC3339)}
	entries, _, _, _ = repo.List(ctx, filter, "", 10)
	if len(entries) != 2 {
		t.Fatalf("date-to filter = %d entries, want 2", len(entries))
	}
}

func TestAuditRepository_ListPaginationTies(t *testing.T) {
	pool := setup(t)
	ctx := context.Background()
	repo := auditpg.NewAuditRepository(pool)
	user := createUser(t, pool)
	resID := newUUID(t)

	for i := 0; i < 5; i++ {
		insertAudit(t, pool, repo, "audit.event", user.ID, user.Email, "audit", resID, 0)
	}

	all := map[string]bool{}
	cursor := ""
	for page := 0; page < 6; page++ {
		entries, next, _, err := repo.List(ctx, &auditdomain.AuditFilter{}, cursor, 2)
		requireNoErr(t, err, "page")
		if len(entries) == 0 {
			break
		}
		for _, e := range entries {
			if all[e.ID] {
				t.Fatalf("entry %s duplicated across pages", e.ID)
			}
			all[e.ID] = true
		}
		if next == "" {
			break
		}
		cursor = next
	}
	if len(all) != 5 {
		t.Fatalf("covered %d distinct entries, want 5", len(all))
	}
}

func TestAuditRepository_ListAfterID(t *testing.T) {
	pool := setup(t)
	ctx := context.Background()
	repo := auditpg.NewAuditRepository(pool)
	user := createUser(t, pool)
	resID := newUUID(t)

	var ids []string
	for i := 0; i < 3; i++ {
		ids = append(ids, insertAudit(t, pool, repo, "audit.event", user.ID, user.Email, "audit", resID, 0))
	}

	after, err := repo.ListAfterID(ctx, ids[0], "", 10)
	requireNoErr(t, err, "ListAfterID")
	if len(after) != 2 {
		t.Fatalf("ListAfterID = %d entries, want 2", len(after))
	}
	if after[0].ID != ids[1] || after[1].ID != ids[2] {
		t.Fatalf("ListAfterID order = %+v", after)
	}
}

func TestAuditRepository_GetStats(t *testing.T) {
	pool := setup(t)
	ctx := context.Background()
	repo := auditpg.NewAuditRepository(pool)
	userA := createUser(t, pool)
	userB := createUser(t, pool)
	resID := newUUID(t)

	insertAudit(t, pool, repo, "migration.run", userA.ID, userA.Email, "migration", resID, 0)
	insertAudit(t, pool, repo, "migration.run", userA.ID, userA.Email, "migration", resID, 0)
	insertAudit(t, pool, repo, "project.create", userB.ID, userB.Email, "project", resID, 1)

	from := time.Now().Add(-2 * 24 * time.Hour)
	to := time.Now().Add(24 * time.Hour)
	stats, err := repo.GetStats(ctx, from, to)
	requireNoErr(t, err, "GetStats")
	if stats.TotalEntries != 3 {
		t.Fatalf("TotalEntries = %d, want 3", stats.TotalEntries)
	}
	if stats.ByEventType["migration.run"] != 2 || stats.ByEventType["project.create"] != 1 {
		t.Fatalf("ByEventType = %+v", stats.ByEventType)
	}
	if stats.UniqueActors != 2 {
		t.Fatalf("UniqueActors = %d, want 2", stats.UniqueActors)
	}
}
