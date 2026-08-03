package handler

import (
	"context"
	"testing"
	"time"

	"github.com/schemahub/backend/internal/audit/domain"
	"github.com/schemahub/backend/internal/pkg/errors"
	auditv1 "github.com/schemahub/backend/proto/audit/v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

type fakeAuditRepo struct {
	entries      []*domain.AuditEntry
	stats        *domain.AuditStats
	lastFilter   *domain.AuditFilter
	lastAfterID  string
	lastEventType string
	listErr      error
	getErr       error
	afterEntries []*domain.AuditEntry
	afterErr     error
}

func (f *fakeAuditRepo) Insert(ctx context.Context, entry *domain.AuditEntry) error { return nil }

func (f *fakeAuditRepo) GetByID(ctx context.Context, id string) (*domain.AuditEntry, error) {
	if f.getErr != nil {
		return nil, f.getErr
	}
	for _, e := range f.entries {
		if e.ID == id {
			return e, nil
		}
	}
	return nil, errors.New("NOT_FOUND", "audit entry not found")
}

func (f *fakeAuditRepo) List(ctx context.Context, filter *domain.AuditFilter, cursor string, limit int32) ([]*domain.AuditEntry, string, int32, error) {
	f.lastFilter = filter
	if f.listErr != nil {
		return nil, "", 0, f.listErr
	}
	return f.entries, "", int32(len(f.entries)), nil
}

func (f *fakeAuditRepo) ListAfterID(ctx context.Context, afterID string, eventType string, limit int) ([]*domain.AuditEntry, error) {
	f.lastAfterID = afterID
	f.lastEventType = eventType
	if f.afterErr != nil {
		return nil, f.afterErr
	}
	return f.afterEntries, nil
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

// fakeAuditStream implements grpc.ServerStreamingServer[AuditEntry] so the
// TailAuditEntries handler can be exercised without a live gRPC connection.
type fakeAuditStream struct {
	ctx  context.Context
	sent []*auditv1.AuditEntry
}

func (f *fakeAuditStream) Send(e *auditv1.AuditEntry) error {
	f.sent = append(f.sent, e)
	return nil
}

func (f *fakeAuditStream) Context() context.Context          { return f.ctx }
func (f *fakeAuditStream) SetHeader(metadata.MD) error       { return nil }
func (f *fakeAuditStream) SendHeader(metadata.MD) error      { return nil }
func (f *fakeAuditStream) SetTrailer(metadata.MD)            {}
func (f *fakeAuditStream) SendMsg(m any) error               { return nil }
func (f *fakeAuditStream) RecvMsg(m any) error               { return nil }

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

func TestAuditHandler_GetAuditEntry(t *testing.T) {
	t.Parallel()

	repo := &fakeAuditRepo{entries: []*domain.AuditEntry{
		{ID: "e1", EventType: "schema_version_created", ActorID: "user_1", Action: "create", ResourceType: "schema", CreatedAt: time.Now()},
	}}
	h := testAuditHandler(t, repo)

	resp, err := h.GetAuditEntry(context.Background(), &auditv1.GetAuditEntryRequest{Id: "e1"})
	if err != nil {
		t.Fatalf("GetAuditEntry() error = %v", err)
	}
	if resp.Entry.Id != "e1" {
		t.Errorf("Entry.Id = %q, want e1", resp.Entry.Id)
	}
	if resp.Entry.EventType != "schema_version_created" {
		t.Errorf("Entry.EventType = %q, want schema_version_created", resp.Entry.EventType)
	}
}

func TestAuditHandler_GetAuditEntryNotFound(t *testing.T) {
	t.Parallel()

	h := testAuditHandler(t, &fakeAuditRepo{})

	_, err := h.GetAuditEntry(context.Background(), &auditv1.GetAuditEntryRequest{Id: "ghost"})
	if status.Code(err) != codes.Internal {
		t.Fatalf("GetAuditEntry() error code = %v, want Internal (%v)", status.Code(err), err)
	}
}

func TestAuditHandler_ListAuditEntriesFilters(t *testing.T) {
	t.Parallel()

	repo := &fakeAuditRepo{entries: []*domain.AuditEntry{
		{ID: "e1", EventType: "schema_version_created", ActorID: "user_1", ResourceType: "schema", ResourceID: "sch_1", CreatedAt: time.Now()},
	}}
	h := testAuditHandler(t, repo)

	_, err := h.ListAuditEntries(context.Background(), &auditv1.ListAuditEntriesRequest{
		EventType:    "schema_version_created",
		ActorId:      "user_1",
		ResourceType: "schema",
		ResourceId:   "sch_1",
		DateFrom:     "2026-01-01T00:00:00Z",
		DateTo:       "2026-01-02T00:00:00Z",
		Cursor:       "cursor_1",
		PageSize:     25,
	})
	if err != nil {
		t.Fatalf("ListAuditEntries() error = %v", err)
	}
	if repo.lastFilter == nil {
		t.Fatal("expected the repository to receive a filter")
	}
	if repo.lastFilter.EventType != "schema_version_created" {
		t.Errorf("filter.EventType = %q, want schema_version_created", repo.lastFilter.EventType)
	}
	if repo.lastFilter.ActorID != "user_1" {
		t.Errorf("filter.ActorID = %q, want user_1", repo.lastFilter.ActorID)
	}
	if repo.lastFilter.ResourceType != "schema" {
		t.Errorf("filter.ResourceType = %q, want schema", repo.lastFilter.ResourceType)
	}
	if repo.lastFilter.ResourceID != "sch_1" {
		t.Errorf("filter.ResourceID = %q, want sch_1", repo.lastFilter.ResourceID)
	}
	if repo.lastFilter.DateFrom != "2026-01-01T00:00:00Z" {
		t.Errorf("filter.DateFrom = %q, want 2026-01-01T00:00:00Z", repo.lastFilter.DateFrom)
	}
	if repo.lastFilter.DateTo != "2026-01-02T00:00:00Z" {
		t.Errorf("filter.DateTo = %q, want 2026-01-02T00:00:00Z", repo.lastFilter.DateTo)
	}
}

func TestAuditHandler_TailAuditEntries(t *testing.T) {
	t.Parallel()

	repo := &fakeAuditRepo{afterEntries: []*domain.AuditEntry{
		{ID: "e3", EventType: "migration_completed", ActorID: "user_1", Action: "execute", CreatedAt: time.Now()},
		{ID: "e4", EventType: "schema_version_created", ActorID: "user_1", Action: "create", CreatedAt: time.Now()},
	}}
	h := testAuditHandler(t, repo)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	stream := &fakeAuditStream{ctx: ctx}
	err := h.TailAuditEntries(&auditv1.TailAuditEntriesRequest{
		SinceEventId: "e2", EventType: "schema_version_created",
	}, stream)
	if err != context.Canceled {
		t.Fatalf("TailAuditEntries() error = %v, want context.Canceled", err)
	}
	if len(stream.sent) != 2 {
		t.Fatalf("sent entries = %d, want 2", len(stream.sent))
	}
	if stream.sent[0].Id != "e3" {
		t.Errorf("sent[0].Id = %q, want e3", stream.sent[0].Id)
	}
	if repo.lastAfterID != "e2" {
		t.Errorf("afterID = %q, want e2", repo.lastAfterID)
	}
	if repo.lastEventType != "schema_version_created" {
		t.Errorf("eventType = %q, want schema_version_created", repo.lastEventType)
	}
}

func TestAuditHandler_TailAuditEntriesNoSince(t *testing.T) {
	t.Parallel()

	h := testAuditHandler(t, &fakeAuditRepo{})

	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	stream := &fakeAuditStream{ctx: ctx}
	err := h.TailAuditEntries(&auditv1.TailAuditEntriesRequest{}, stream)
	if err != context.Canceled {
		t.Fatalf("TailAuditEntries() error = %v, want context.Canceled", err)
	}
	if len(stream.sent) != 0 {
		t.Errorf("sent entries = %d, want 0", len(stream.sent))
	}
}

func TestAuditHandler_TailAuditEntriesRepoError(t *testing.T) {
	t.Parallel()

	repo := &fakeAuditRepo{afterErr: errors.New("NOT_FOUND", "audit entry not found")}
	h := testAuditHandler(t, repo)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	err := h.TailAuditEntries(&auditv1.TailAuditEntriesRequest{SinceEventId: "e9"}, &fakeAuditStream{ctx: ctx})
	if status.Code(err) != codes.NotFound {
		t.Fatalf("TailAuditEntries() error code = %v, want NotFound (%v)", status.Code(err), err)
	}
}
