package handler

import (
	"context"
	"testing"
	"time"

	"github.com/schemahub/backend/internal/audit/domain"
	"github.com/schemahub/backend/internal/pkg/errors"
	auditv1 "github.com/schemahub/backend/proto/audit/v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type fakeAuditRepo struct {
	entries []*domain.AuditEntry
	stats   *domain.AuditStats
}

func (f *fakeAuditRepo) Insert(ctx context.Context, entry *domain.AuditEntry) error { return nil }

func (f *fakeAuditRepo) GetByID(ctx context.Context, id string) (*domain.AuditEntry, error) {
	for _, e := range f.entries {
		if e.ID == id {
			return e, nil
		}
	}
	return nil, errors.New("NOT_FOUND", "audit entry not found")
}

func (f *fakeAuditRepo) List(ctx context.Context, filter *domain.AuditFilter, cursor string, limit int32) ([]*domain.AuditEntry, string, int32, error) {
	return f.entries, "", int32(len(f.entries)), nil
}

func (f *fakeAuditRepo) ListAfterID(ctx context.Context, afterID string, eventType string, limit int) ([]*domain.AuditEntry, error) {
	return nil, nil
}

func (f *fakeAuditRepo) GetStats(ctx context.Context, dateFrom, dateTo time.Time) (*domain.AuditStats, error) {
	if f.stats != nil {
		return f.stats, nil
	}
	return &domain.AuditStats{DateFrom: dateFrom, DateTo: dateTo}, nil
}

func testAuditHandler(t *testing.T, repo *fakeAuditRepo) *AuditHandler {
	t.Helper()
	return NewAuditHandler(domain.NewAuditService(repo))
}

func TestAuditHandler_ListAuditEntries(t *testing.T) {
	t.Parallel()

	repo := &fakeAuditRepo{entries: []*domain.AuditEntry{
		{ID: "e1", EventType: "schema_version_created", ActorID: "user_1", Action: "create", ResourceType: "schema", CreatedAt: time.Now()},
		{ID: "e2", EventType: "migration_completed", ActorID: "user_1", Action: "execute", ResourceType: "migration", CreatedAt: time.Now()},
	}}
	h := testAuditHandler(t, repo)

	resp, err := h.ListAuditEntries(context.Background(), &auditv1.ListAuditEntriesRequest{
		ActorId: "user_1", PageSize: 10,
	})
	if err != nil {
		t.Fatalf("ListAuditEntries() error = %v", err)
	}
	if len(resp.Entries) != 2 {
		t.Errorf("Entries len = %d, want 2", len(resp.Entries))
	}
	if resp.TotalCount != 2 {
		t.Errorf("TotalCount = %d, want 2", resp.TotalCount)
	}
	if resp.Entries[0].EventType != "schema_version_created" {
		t.Errorf("Entries[0].EventType = %q, want schema_version_created", resp.Entries[0].EventType)
	}
}

func TestAuditHandler_GetAuditStats(t *testing.T) {
	t.Parallel()

	now := time.Now()
	repo := &fakeAuditRepo{stats: &domain.AuditStats{
		TotalEntries: 5,
		ByEventType:  map[string]int32{"schema_version_created": 5},
		ByAction:     map[string]int32{"create": 5},
		UniqueActors: 2,
		DateFrom:     now, DateTo: now,
	}}
	h := testAuditHandler(t, repo)

	resp, err := h.GetAuditStats(context.Background(), &auditv1.GetAuditStatsRequest{})
	if err != nil {
		t.Fatalf("GetAuditStats() error = %v", err)
	}
	if resp.TotalEntries != 5 {
		t.Errorf("TotalEntries = %d, want 5", resp.TotalEntries)
	}
	if resp.ByEventType["schema_version_created"] != 5 {
		t.Errorf("ByEventType = %v, want schema_version_created:5", resp.ByEventType)
	}
	if resp.UniqueActors != 2 {
		t.Errorf("UniqueActors = %d, want 2", resp.UniqueActors)
	}
}

func TestAuditHandler_GetAuditStatsInvalidDate(t *testing.T) {
	t.Parallel()

	h := testAuditHandler(t, &fakeAuditRepo{})

	_, err := h.GetAuditStats(context.Background(), &auditv1.GetAuditStatsRequest{
		DateFrom: "not-a-date",
	})
	if status.Code(err) != codes.InvalidArgument {
		t.Fatalf("GetAuditStats() error code = %v, want InvalidArgument (%v)", status.Code(err), err)
	}
}
