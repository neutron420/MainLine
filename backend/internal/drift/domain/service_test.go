package domain

import (
	"context"
	"errors"
	"testing"
)

type fakeDriftRepo struct {
	events    map[string]*DriftEvent
	insertErr error
	listErr   error
}

func (f *fakeDriftRepo) Insert(ctx context.Context, e *DriftEvent) error {
	if f.insertErr != nil {
		return f.insertErr
	}
	if f.events == nil {
		f.events = map[string]*DriftEvent{}
	}
	f.events[e.ID] = e
	return nil
}

func (f *fakeDriftRepo) GetByID(ctx context.Context, id string) (*DriftEvent, error) {
	e, ok := f.events[id]
	if !ok {
		return nil, errors.New("drift event not found")
	}
	return e, nil
}

func (f *fakeDriftRepo) List(ctx context.Context, filter *DriftFilter, cursor string, limit int32) ([]*DriftEvent, string, int32, error) {
	if f.listErr != nil {
		return nil, "", 0, f.listErr
	}
	var out []*DriftEvent
	for _, e := range f.events {
		out = append(out, e)
	}
	return out, "", int32(len(out)), nil
}

func (f *fakeDriftRepo) UpdateStatus(ctx context.Context, id string, status DriftStatus, resolvedBy string) error {
	e, ok := f.events[id]
	if !ok {
		return errors.New("drift event not found")
	}
	e.Status = status
	e.ResolvedBy = resolvedBy
	return nil
}

func (f *fakeDriftRepo) GetStats(ctx context.Context, connectionID string) (*DriftStats, error) {
	return &DriftStats{TotalOpen: 3}, nil
}

type fakeComparator struct {
	events []*DriftEvent
	err    error
}

func (f *fakeComparator) CompareLiveWithVersion(ctx context.Context, connStr, schemaVersionID string, schemaNames []string) ([]*DriftEvent, error) {
	if f.err != nil {
		return nil, f.err
	}
	return f.events, nil
}

func TestCheckDrift_PersistsEvents(t *testing.T) {
	t.Parallel()

	repo := &fakeDriftRepo{}
	comparator := &fakeComparator{
		events: []*DriftEvent{
			{ID: "ev1", ObjectName: "users.email", Severity: SeverityCritical},
			{ID: "ev2", ObjectName: "idx_users", Severity: SeverityWarning},
		},
	}
	svc := NewDriftService(repo, comparator)

	events, err := svc.CheckDrift(context.Background(), "postgres://x", "conn_1", []string{"public"})
	if err != nil {
		t.Fatalf("CheckDrift() error = %v", err)
	}
	if len(events) != 2 {
		t.Fatalf("events len = %d, want 2", len(events))
	}
	for _, e := range events {
		if e.ConnectionID != "conn_1" {
			t.Errorf("event ConnectionID = %q, want conn_1", e.ConnectionID)
		}
	}
}

func TestCheckDrift_NoEvents(t *testing.T) {
	t.Parallel()

	svc := NewDriftService(&fakeDriftRepo{}, &fakeComparator{})
	events, err := svc.CheckDrift(context.Background(), "postgres://x", "conn_1", nil)
	if err != nil {
		t.Fatalf("CheckDrift() error = %v", err)
	}
	if len(events) != 0 {
		t.Errorf("events len = %d, want 0", len(events))
	}
}

func TestCheckDrift_ComparatorError(t *testing.T) {
	t.Parallel()

	svc := NewDriftService(&fakeDriftRepo{}, &fakeComparator{err: errors.New("boom")})
	if _, err := svc.CheckDrift(context.Background(), "postgres://x", "conn_1", nil); err == nil {
		t.Error("CheckDrift() = nil error, want error from comparator")
	}
}

func TestCheckDrift_InsertError(t *testing.T) {
	t.Parallel()

	repo := &fakeDriftRepo{insertErr: errors.New("db full")}
	svc := NewDriftService(repo, &fakeComparator{events: []*DriftEvent{{ID: "ev1"}}})
	if _, err := svc.CheckDrift(context.Background(), "postgres://x", "conn_1", nil); err == nil {
		t.Error("CheckDrift() = nil error, want persist error")
	}
}

func TestResolve_ValidTransitions(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		status string
	}{
		{name: "acknowledged", status: "acknowledged"},
		{name: "resolved", status: "resolved"},
		{name: "false positive", status: "false_positive"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &fakeDriftRepo{
				events: map[string]*DriftEvent{"ev1": {ID: "ev1", Status: DriftStatusOpen}},
			}
			svc := NewDriftService(repo, &fakeComparator{})

			event, err := svc.Resolve(context.Background(), "ev1", tt.status, "user_9")
			if err != nil {
				t.Fatalf("Resolve(%s) error = %v", tt.status, err)
			}
			if string(event.Status) != tt.status {
				t.Errorf("status = %q, want %q", event.Status, tt.status)
			}
			if event.ResolvedBy != "user_9" {
				t.Errorf("ResolvedBy = %q, want user_9", event.ResolvedBy)
			}
		})
	}
}

func TestResolve_InvalidStatus(t *testing.T) {
	t.Parallel()

	svc := NewDriftService(&fakeDriftRepo{}, &fakeComparator{})
	if _, err := svc.Resolve(context.Background(), "ev1", "bogus", "user_9"); err == nil {
		t.Error("Resolve(bogus) = nil error, want INVALID_ARGUMENT")
	}
}

func TestResolve_AlreadyResolved(t *testing.T) {
	t.Parallel()

	repo := &fakeDriftRepo{
		events: map[string]*DriftEvent{"ev1": {ID: "ev1", Status: DriftStatusResolved}},
	}
	svc := NewDriftService(repo, &fakeComparator{})
	if _, err := svc.Resolve(context.Background(), "ev1", "resolved", "user_9"); err == nil {
		t.Error("Resolve(already resolved) = nil error, want FAILED_PRECONDITION")
	}
}

func TestResolve_NotFound(t *testing.T) {
	t.Parallel()

	svc := NewDriftService(&fakeDriftRepo{}, &fakeComparator{})
	if _, err := svc.Resolve(context.Background(), "missing", "resolved", "user_9"); err == nil {
		t.Error("Resolve(missing) = nil error, want error")
	}
}

func TestList_DefaultPageSize(t *testing.T) {
	t.Parallel()

	repo := &fakeDriftRepo{
		events:  map[string]*DriftEvent{"ev1": {ID: "ev1"}},
		listErr: nil,
	}
	svc := NewDriftService(repo, &fakeComparator{})

	events, cursor, total, err := svc.List(context.Background(), nil, "", 0)
	if err != nil {
		t.Fatalf("List() error = %v", err)
	}
	if len(events) != 1 || total != 1 {
		t.Errorf("List() = %d events, total %d; want 1, 1", len(events), total)
	}
	if cursor != "" {
		t.Errorf("cursor = %q, want empty", cursor)
	}
}

func TestList_ClampsPageSize(t *testing.T) {
	t.Parallel()

	svc := NewDriftService(&fakeDriftRepo{}, &fakeComparator{})
	_, _, _, err := svc.List(context.Background(), nil, "", 500)
	if err != nil {
		t.Fatalf("List(500) error = %v", err)
	}
}

func TestGetStats(t *testing.T) {
	t.Parallel()

	svc := NewDriftService(&fakeDriftRepo{}, &fakeComparator{})
	stats, err := svc.GetStats(context.Background(), "conn_1")
	if err != nil {
		t.Fatalf("GetStats() error = %v", err)
	}
	if stats.TotalOpen != 3 {
		t.Errorf("TotalOpen = %d, want 3", stats.TotalOpen)
	}
}
