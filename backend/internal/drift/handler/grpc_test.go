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
	events    map[string]*domain.DriftEvent
	resolved  *domain.DriftEvent
	notFound  bool
	updateErr error
}

func newFakeDriftRepo() *fakeDriftRepo {
	return &fakeDriftRepo{events: map[string]*domain.DriftEvent{}}
}

func (f *fakeDriftRepo) Insert(ctx context.Context, event *domain.DriftEvent) error { return nil }

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
	return &domain.DriftStats{}, nil
}

type fakeComparator struct{}

func (f *fakeComparator) CompareLiveWithVersion(ctx context.Context, connStr, schemaVersionID string, schemaNames []string) ([]*domain.DriftEvent, error) {
	return nil, nil
}

func testDriftHandler(t *testing.T, repo *fakeDriftRepo) *DriftHandler {
	t.Helper()
	svc := domain.NewDriftService(repo, &fakeComparator{})
	return NewDriftHandler(svc, nil)
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
