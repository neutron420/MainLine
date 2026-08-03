package integration

import (
	"context"
	"testing"
	"time"

	driftdomain "github.com/schemahub/backend/internal/drift/domain"
	driftpg "github.com/schemahub/backend/internal/drift/repository/postgres"
)

func TestDriftRepository_RoundTripAndStatus(t *testing.T) {
	pool := setup(t)
	ctx := context.Background()
	owner := createUser(t, pool)
	proj := createProject(t, pool, owner)
	conn := createConnection(t, pool, proj, owner)
	repo := driftpg.NewDriftRepository(pool)

	evt := &driftdomain.DriftEvent{
		ConnectionID:       conn.ID,
		DriftType:          driftdomain.DriftTypeModifiedObject,
		ObjectType:         "table",
		ObjectName:         "users",
		ExpectedDefinition: `{"sql":"CREATE TABLE users (id int)"}`,
		ActualDefinition:   `{"sql":"CREATE TABLE users (id bigint)"}`,
		DiffSummary:        `{"changed":["id"]}`,
		Severity:           driftdomain.SeverityCritical,
		Status:             driftdomain.DriftStatusOpen,
	}
	requireNoErr(t, repo.Insert(ctx, evt), "Insert")
	if evt.ID == "" {
		t.Fatal("Insert did not populate event ID")
	}

	got, err := repo.GetByID(ctx, evt.ID)
	requireNoErr(t, err, "GetByID")
	if got.ObjectName != "users" || got.DriftType != driftdomain.DriftTypeModifiedObject {
		t.Fatalf("unexpected event: %+v", got)
	}

	requireNoErr(t, repo.UpdateStatus(ctx, evt.ID, driftdomain.DriftStatusResolved, owner.ID), "UpdateStatus")
	resolved, _ := repo.GetByID(ctx, evt.ID)
	if resolved.Status != driftdomain.DriftStatusResolved || resolved.ResolvedBy != owner.ID || resolved.ResolvedAt == nil {
		t.Fatalf("UpdateStatus not persisted: %+v", resolved)
	}
}

func TestDriftRepository_ListFiltersAndPaginationTies(t *testing.T) {
	pool := setup(t)
	ctx := context.Background()
	owner := createUser(t, pool)
	proj := createProject(t, pool, owner)
	conn := createConnection(t, pool, proj, owner)
	repo := driftpg.NewDriftRepository(pool)

	// 4 events on conn, 1 event on a second connection.
	conn2 := createConnection(t, pool, proj, owner)
	for i := 0; i < 4; i++ {
		requireNoErr(t, repo.Insert(ctx, &driftdomain.DriftEvent{
			ConnectionID: conn.ID,
			DriftType:    driftdomain.DriftTypeModifiedObject,
			ObjectType:   "table",
			ObjectName:   "t" + time.Now().String(),
			Severity:     driftdomain.SeverityWarning,
			Status:       driftdomain.DriftStatusOpen,
		}), "Insert")
	}
	requireNoErr(t, repo.Insert(ctx, &driftdomain.DriftEvent{
		ConnectionID: conn2.ID,
		DriftType:    driftdomain.DriftTypeMissingObject,
		ObjectType:   "table",
		ObjectName:   "other",
		Severity:     driftdomain.SeverityInfo,
		Status:       driftdomain.DriftStatusOpen,
	}), "Insert 2")

	filter := &driftdomain.DriftFilter{ConnectionID: conn.ID}
	all := map[string]bool{}
	cursor := ""
	for page := 0; page < 6; page++ {
		events, next, _, err := repo.List(ctx, filter, cursor, 2)
		requireNoErr(t, err, "page")
		if len(events) == 0 {
			break
		}
		for _, e := range events {
			if all[e.ID] {
				t.Fatalf("event %s duplicated across pages", e.ID)
			}
			all[e.ID] = true
		}
		if next == "" {
			break
		}
		cursor = next
	}
	if len(all) != 4 {
		t.Fatalf("covered %d distinct events on conn, want 4", len(all))
	}

	filter = &driftdomain.DriftFilter{ConnectionID: conn2.ID, DriftType: string(driftdomain.DriftTypeMissingObject)}
	events, _, _, err := repo.List(ctx, filter, "", 10)
	requireNoErr(t, err, "filtered list")
	if len(events) != 1 || events[0].ObjectName != "other" {
		t.Fatalf("filtered = %+v", events)
	}

	filter = &driftdomain.DriftFilter{Status: string(driftdomain.DriftStatusResolved)}
	events, _, _, _ = repo.List(ctx, filter, "", 10)
	if len(events) != 0 {
		t.Fatalf("status filter = %d events, want 0", len(events))
	}
}

func TestDriftRepository_GetStats(t *testing.T) {
	pool := setup(t)
	ctx := context.Background()
	owner := createUser(t, pool)
	proj := createProject(t, pool, owner)
	conn := createConnection(t, pool, proj, owner)
	repo := driftpg.NewDriftRepository(pool)

	for i := 0; i < 3; i++ {
		requireNoErr(t, repo.Insert(ctx, &driftdomain.DriftEvent{
			ConnectionID: conn.ID,
			DriftType:    driftdomain.DriftTypeModifiedObject,
			ObjectType:   "table",
			ObjectName:   "obj",
			Severity:     driftdomain.SeverityCritical,
			Status:       driftdomain.DriftStatusOpen,
		}), "Insert open")
	}
	requireNoErr(t, repo.Insert(ctx, &driftdomain.DriftEvent{
		ConnectionID: conn.ID,
		DriftType:    driftdomain.DriftTypeExtraObject,
		ObjectType:   "table",
		ObjectName:   "extra",
		Severity:     driftdomain.SeverityInfo,
		Status:       driftdomain.DriftStatusAcknowledged,
	}), "Insert acknowledged")

	stats, err := repo.GetStats(ctx, conn.ID)
	requireNoErr(t, err, "GetStats")
	if stats.TotalOpen != 3 || stats.TotalAcknowledged != 1 || stats.TotalResolved != 0 {
		t.Fatalf("stats = %+v", stats)
	}
	if stats.BySeverity["critical"] != 3 || stats.BySeverity["info"] != 1 {
		t.Fatalf("BySeverity = %+v", stats.BySeverity)
	}
	if stats.ByDriftType["modified_object"] != 3 || stats.ByDriftType["extra_object"] != 1 {
		t.Fatalf("ByDriftType = %+v", stats.ByDriftType)
	}
}
