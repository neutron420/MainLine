package handler

import (
	"context"
	"testing"
	"time"

	"github.com/schemahub/backend/internal/drift/domain"
	"github.com/schemahub/backend/internal/pkg/errors"
	"github.com/schemahub/backend/internal/pkg/interceptor"
	driftv1 "github.com/schemahub/backend/proto/drift/v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type fakeDriftRepo struct {
	events     map[string]*domain.DriftEvent
	resolved   *domain.DriftEvent
	notFound   bool
	updateErr  error
	stats      *domain.DriftStats
	lastFilter *domain.DriftFilter
}

func newFakeDriftRepo() *fakeDriftRepo {
	return &fakeDriftRepo{events: map[string]*domain.DriftEvent{}}
}

func (f *fakeDriftRepo) Insert(ctx context.Context, event *domain.DriftEvent) error {
	f.events[event.ID] = event
	return nil
}

func (f *fakeDriftRepo) GetByID(ctx context.Context, id string) (*domain.DriftEvent, error) {
	if f.notFound {
		return nil, errors.New("NOT_FOUND", "drift event not found")
	}
	e, ok := f.events[id]
	if !ok {
		return nil, errors.New("NOT_FOUND", "drift event not found")
	}
	return e, nil
}

func (f *fakeDriftRepo) List(ctx context.Context, filter *domain.DriftFilter, cursor string, limit int32) ([]*domain.DriftEvent, string, int32, error) {
	f.lastFilter = filter
	var out []*domain.DriftEvent
	for _, e := range f.events {
		out = append(out, e)
	}
	return out, "", int32(len(out)), nil
}

func (f *fakeDriftRepo) UpdateStatus(ctx context.Context, id string, status domain.DriftStatus, resolvedBy string) error {
	if f.updateErr != nil {
		return f.updateErr
	}
	f.events[id].Status = status
	f.events[id].ResolvedBy = resolvedBy
	f.resolved = f.events[id]
	return nil
}

func (f *fakeDriftRepo) GetStats(ctx context.Context, connectionID string) (*domain.DriftStats, error) {
	if f.stats != nil {
		return f.stats, nil
	}
	return &domain.DriftStats{}, nil
}

type fakeComparator struct {
	events     []*domain.DriftEvent
	compareErr error
}

func (f *fakeComparator) CompareLiveWithVersion(ctx context.Context, connStr, schemaVersionID string, schemaNames []string) ([]*domain.DriftEvent, error) {
	if f.compareErr != nil {
		return nil, f.compareErr
	}
	return f.events, nil
}

func testDriftHandler(t *testing.T, repo *fakeDriftRepo) *DriftHandler {
	t.Helper()
	return testDriftHandlerWithConn(t, repo, &fakeComparator{}, nil)
}

func testDriftHandlerWithConn(t *testing.T, repo *fakeDriftRepo, comparator *fakeComparator, connString func(ctx context.Context, connID string) (string, error)) *DriftHandler {
	t.Helper()
	svc := domain.NewDriftService(repo, comparator)
	return NewDriftHandler(svc, connString)
}

func TestDriftHandler_ListDriftEvents(t *testing.T) {
	t.Parallel()

	repo := newFakeDriftRepo()
	repo.events["d1"] = &domain.DriftEvent{
		ID: "d1", ConnectionID: "conn_1", DriftType: domain.DriftTypeMissingObject,
		ObjectName: "users", Severity: domain.SeverityCritical, Status: domain.DriftStatusOpen,
		DetectedAt: time.Now(),
	}
	h := testDriftHandler(t, repo)

	resp, err := h.ListDriftEvents(context.Background(), &driftv1.ListDriftEventsRequest{
		ConnectionId: "conn_1", PageSize: 10,
	})
	if err != nil {
		t.Fatalf("ListDriftEvents() error = %v", err)
	}
	if len(resp.Events) != 1 {
		t.Errorf("Events len = %d, want 1", len(resp.Events))
	}
	if resp.Events[0].ObjectName != "users" {
		t.Errorf("Events[0].ObjectName = %q, want users", resp.Events[0].ObjectName)
	}
}

func TestDriftHandler_ResolveDriftEvent(t *testing.T) {
	t.Parallel()

	repo := newFakeDriftRepo()
	repo.events["d1"] = &domain.DriftEvent{
		ID: "d1", DriftType: domain.DriftTypeMissingObject, ObjectName: "users",
		Severity: domain.SeverityWarning, Status: domain.DriftStatusOpen, DetectedAt: time.Now(),
	}
	h := testDriftHandler(t, repo)

	ctx := context.WithValue(context.Background(), interceptor.UserIDKey, "user_1")
	resp, err := h.ResolveDriftEvent(ctx, &driftv1.ResolveDriftEventRequest{
		Id: "d1", Status: "resolved",
	})
	if err != nil {
		t.Fatalf("ResolveDriftEvent() error = %v", err)
	}
	if resp.Event.Status != "resolved" {
		t.Errorf("Event.Status = %q, want resolved", resp.Event.Status)
	}
	if repo.resolved == nil || repo.resolved.ResolvedBy != "user_1" {
		t.Error("expected resolved event recorded with ResolvedBy user_1")
	}
}

func TestDriftHandler_ResolveDriftEventInvalidStatus(t *testing.T) {
	t.Parallel()

	repo := newFakeDriftRepo()
	repo.events["d1"] = &domain.DriftEvent{
		ID: "d1", Status: domain.DriftStatusOpen, DetectedAt: time.Now(),
	}
	h := testDriftHandler(t, repo)

	ctx := context.WithValue(context.Background(), interceptor.UserIDKey, "user_1")
	_, err := h.ResolveDriftEvent(ctx, &driftv1.ResolveDriftEventRequest{
		Id: "d1", Status: "bogus",
	})
	if status.Code(err) != codes.InvalidArgument {
		t.Fatalf("ResolveDriftEvent() error code = %v, want InvalidArgument (%v)", status.Code(err), err)
	}
}

func TestDriftHandler_ResolveDriftEventAlreadyResolved(t *testing.T) {
	t.Parallel()

	repo := newFakeDriftRepo()
	repo.events["d1"] = &domain.DriftEvent{
		ID: "d1", Status: domain.DriftStatusResolved, DetectedAt: time.Now(),
	}
	h := testDriftHandler(t, repo)

	ctx := context.WithValue(context.Background(), interceptor.UserIDKey, "user_1")
	_, err := h.ResolveDriftEvent(ctx, &driftv1.ResolveDriftEventRequest{
		Id: "d1", Status: "resolved",
	})
	if status.Code(err) != codes.FailedPrecondition {
		t.Fatalf("ResolveDriftEvent() error code = %v, want FailedPrecondition (%v)", status.Code(err), err)
	}
}

func TestDriftHandler_GetDriftEvent(t *testing.T) {
	t.Parallel()

	repo := newFakeDriftRepo()
	repo.events["d1"] = &domain.DriftEvent{
		ID: "d1", ConnectionID: "conn_1", DriftType: domain.DriftTypeMissingObject,
		ObjectName: "users", Severity: domain.SeverityCritical, Status: domain.DriftStatusOpen,
		DetectedAt: time.Now(),
	}
	h := testDriftHandler(t, repo)

	resp, err := h.GetDriftEvent(context.Background(), &driftv1.GetDriftEventRequest{Id: "d1"})
	if err != nil {
		t.Fatalf("GetDriftEvent() error = %v", err)
	}
	if resp.Event.Id != "d1" {
		t.Errorf("Event.Id = %q, want d1", resp.Event.Id)
	}
	if resp.Event.ObjectName != "users" {
		t.Errorf("Event.ObjectName = %q, want users", resp.Event.ObjectName)
	}
}

func TestDriftHandler_GetDriftEventNotFound(t *testing.T) {
	t.Parallel()

	repo := newFakeDriftRepo()
	repo.notFound = true
	h := testDriftHandler(t, repo)

	_, err := h.GetDriftEvent(context.Background(), &driftv1.GetDriftEventRequest{Id: "ghost"})
	if status.Code(err) != codes.NotFound {
		t.Fatalf("GetDriftEvent() error code = %v, want NotFound (%v)", status.Code(err), err)
	}
}

func TestDriftHandler_ListDriftEventsFilters(t *testing.T) {
	t.Parallel()

	repo := newFakeDriftRepo()
	repo.events["d1"] = &domain.DriftEvent{
		ID: "d1", ConnectionID: "conn_1", DriftType: domain.DriftTypeMissingObject,
		ObjectName: "users", Severity: domain.SeverityCritical, Status: domain.DriftStatusOpen,
		DetectedAt: time.Now(),
	}
	h := testDriftHandler(t, repo)

	resp, err := h.ListDriftEvents(context.Background(), &driftv1.ListDriftEventsRequest{
		ConnectionId: "conn_1", Status: "open", Severity: "critical", DriftType: "missing_object",
		Cursor: "cursor_1", PageSize: 25,
	})
	if err != nil {
		t.Fatalf("ListDriftEvents() error = %v", err)
	}
	if len(resp.Events) != 1 {
		t.Errorf("Events len = %d, want 1", len(resp.Events))
	}
	if repo.lastFilter == nil {
		t.Fatal("expected the repository to receive a filter")
	}
	if repo.lastFilter.ConnectionID != "conn_1" {
		t.Errorf("filter.ConnectionID = %q, want conn_1", repo.lastFilter.ConnectionID)
	}
	if repo.lastFilter.Status != "open" {
		t.Errorf("filter.Status = %q, want open", repo.lastFilter.Status)
	}
	if repo.lastFilter.Severity != "critical" {
		t.Errorf("filter.Severity = %q, want critical", repo.lastFilter.Severity)
	}
	if repo.lastFilter.DriftType != "missing_object" {
		t.Errorf("filter.DriftType = %q, want missing_object", repo.lastFilter.DriftType)
	}
}

func TestDriftHandler_CheckDrift(t *testing.T) {
	t.Parallel()

	repo := newFakeDriftRepo()
	comparator := &fakeComparator{events: []*domain.DriftEvent{
		{ID: "d_new", DriftType: domain.DriftTypeExtraObject, ObjectName: "orphan", Severity: domain.SeverityWarning, Status: domain.DriftStatusOpen, DetectedAt: time.Now()},
	}}
	connString := func(ctx context.Context, connID string) (string, error) {
		return "postgres://app:secret@db.example.com:5432/app?sslmode=require", nil
	}
	h := testDriftHandlerWithConn(t, repo, comparator, connString)

	resp, err := h.CheckDrift(context.Background(), &driftv1.CheckDriftRequest{
		ConnectionId: "conn_1", SchemaNames: []string{"public"},
	})
	if err != nil {
		t.Fatalf("CheckDrift() error = %v", err)
	}
	if len(resp.Events) != 1 {
		t.Fatalf("Events len = %d, want 1", len(resp.Events))
	}
	if resp.Events[0].ObjectName != "orphan" {
		t.Errorf("Events[0].ObjectName = %q, want orphan", resp.Events[0].ObjectName)
	}
	if !resp.HasDrift {
		t.Error("HasDrift = false, want true")
	}
	if resp.TotalDrifts != 1 {
		t.Errorf("TotalDrifts = %d, want 1", resp.TotalDrifts)
	}
	if repo.events["d_new"] == nil {
		t.Error("expected detected drift event to be persisted")
	}
}

func TestDriftHandler_CheckDriftNoDrift(t *testing.T) {
	t.Parallel()

	h := testDriftHandlerWithConn(t, newFakeDriftRepo(), &fakeComparator{}, func(ctx context.Context, connID string) (string, error) {
		return "postgres://app:secret@db.example.com:5432/app?sslmode=require", nil
	})

	resp, err := h.CheckDrift(context.Background(), &driftv1.CheckDriftRequest{ConnectionId: "conn_1"})
	if err != nil {
		t.Fatalf("CheckDrift() error = %v", err)
	}
	if resp.HasDrift {
		t.Error("HasDrift = true, want false")
	}
	if resp.TotalDrifts != 0 {
		t.Errorf("TotalDrifts = %d, want 0", resp.TotalDrifts)
	}
}

func TestDriftHandler_CheckDriftConnStringError(t *testing.T) {
	t.Parallel()

	h := testDriftHandlerWithConn(t, newFakeDriftRepo(), &fakeComparator{}, func(ctx context.Context, connID string) (string, error) {
		return "", errors.New("NOT_FOUND", "connection not found")
	})

	_, err := h.CheckDrift(context.Background(), &driftv1.CheckDriftRequest{ConnectionId: "ghost"})
	if status.Code(err) != codes.NotFound {
		t.Fatalf("CheckDrift() error code = %v, want NotFound (%v)", status.Code(err), err)
	}
}

func TestDriftHandler_CheckDriftComparatorError(t *testing.T) {
	t.Parallel()

	comparator := &fakeComparator{compareErr: errors.New("INTERNAL", "live schema unavailable")}
	h := testDriftHandlerWithConn(t, newFakeDriftRepo(), comparator, func(ctx context.Context, connID string) (string, error) {
		return "postgres://app:secret@db.example.com:5432/app?sslmode=require", nil
	})

	_, err := h.CheckDrift(context.Background(), &driftv1.CheckDriftRequest{ConnectionId: "conn_1"})
	if status.Code(err) != codes.Internal {
		t.Fatalf("CheckDrift() error code = %v, want Internal (%v)", status.Code(err), err)
	}
}

func TestDriftHandler_GetDriftStats(t *testing.T) {
	t.Parallel()

	repo := newFakeDriftRepo()
	repo.stats = &domain.DriftStats{
		TotalOpen:          3,
		TotalResolved:      1,
		TotalAcknowledged:  2,
		TotalFalsePositive: 1,
		BySeverity:         map[string]int32{"critical": 2},
		ByDriftType:        map[string]int32{"missing_object": 3},
	}
	h := testDriftHandler(t, repo)

	resp, err := h.GetDriftStats(context.Background(), &driftv1.GetDriftStatsRequest{ConnectionId: "conn_1"})
	if err != nil {
		t.Fatalf("GetDriftStats() error = %v", err)
	}
	if resp.TotalOpen != 3 {
		t.Errorf("TotalOpen = %d, want 3", resp.TotalOpen)
	}
	if resp.TotalResolved != 1 {
		t.Errorf("TotalResolved = %d, want 1", resp.TotalResolved)
	}
	if resp.TotalAcknowledged != 2 {
		t.Errorf("TotalAcknowledged = %d, want 2", resp.TotalAcknowledged)
	}
	if resp.BySeverity["critical"] != 2 {
		t.Errorf("BySeverity = %v, want critical:2", resp.BySeverity)
	}
}
