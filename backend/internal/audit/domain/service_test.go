package domain

import (
	"context"
	"testing"
	"time"
)

type fakeAuditRepo struct {
	entries map[string]*AuditEntry
}

func (f *fakeAuditRepo) Insert(ctx context.Context, entry *AuditEntry) error {
	if f.entries == nil {
		f.entries = map[string]*AuditEntry{}
	}
	f.entries[entry.ID] = entry
	return nil
}

func (f *fakeAuditRepo) GetByID(ctx context.Context, id string) (*AuditEntry, error) {
	e, ok := f.entries[id]
	if !ok {
		return nil, context.DeadlineExceeded
	}
	return e, nil
}

func (f *fakeAuditRepo) List(ctx context.Context, filter *AuditFilter, cursor string, limit int32) ([]*AuditEntry, string, int32, error) {
	var out []*AuditEntry
	for _, e := range f.entries {
		out = append(out, e)
	}
	return out, "", int32(len(out)), nil
}

func (f *fakeAuditRepo) ListAfterID(ctx context.Context, afterID string, eventType string, limit int) ([]*AuditEntry, error) {
	return nil, nil
}

func (f *fakeAuditRepo) GetStats(ctx context.Context, dateFrom, dateTo time.Time) (*AuditStats, error) {
	return &AuditStats{TotalEntries: 5}, nil
}

func TestAuditService_InsertAndGet(t *testing.T) {
	t.Parallel()

	repo := &fakeAuditRepo{}
	svc := NewAuditService(repo)

	entry := &AuditEntry{ID: "a1", ActorID: "user_1"}
	if err := svc.Insert(context.Background(), entry); err != nil {
		t.Fatalf("Insert() error = %v", err)
	}

	got, err := svc.GetByID(context.Background(), "a1")
	if err != nil {
		t.Fatalf("GetByID() error = %v", err)
	}
	if got.ID != "a1" {
		t.Errorf("GetByID() = %q, want a1", got.ID)
	}
}

func TestAuditService_ListDefaultsPageSize(t *testing.T) {
	t.Parallel()

	repo := &fakeAuditRepo{entries: map[string]*AuditEntry{
		"a1": {ID: "a1"},
		"a2": {ID: "a2"},
	}}
	svc := NewAuditService(repo)

	entries, _, total, err := svc.List(context.Background(), nil, "", 0)
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 2 || total != 2 {
		t.Errorf("List() = %d entries, total %d; want 2, 2", len(entries), total)
	}
}

func TestAuditService_ListAfterIDClampsLimit(t *testing.T) {
	t.Parallel()

	svc := NewAuditService(&fakeAuditRepo{})
	if _, err := svc.ListAfterID(context.Background(), "a1", "migration", 0); err != nil {
		t.Fatalf("ListAfterID() error = %v", err)
	}
	if _, err := svc.ListAfterID(context.Background(), "a1", "migration", 5000); err != nil {
		t.Fatalf("ListAfterID(5000) error = %v", err)
	}
}

func TestAuditService_GetStats(t *testing.T) {
	t.Parallel()

	svc := NewAuditService(&fakeAuditRepo{})

	t.Run("valid dates", func(t *testing.T) {
		stats, err := svc.GetStats(context.Background(), "2026-01-01T00:00:00Z", "2026-01-31T00:00:00Z")
		if err != nil {
			t.Fatalf("GetStats() error = %v", err)
		}
		if stats.TotalEntries != 5 {
			t.Errorf("TotalEntries = %d, want 5", stats.TotalEntries)
		}
	})

	t.Run("empty dates", func(t *testing.T) {
		if _, err := svc.GetStats(context.Background(), "", ""); err != nil {
			t.Errorf("GetStats(empty) error = %v, want nil", err)
		}
	})

	t.Run("invalid date_from", func(t *testing.T) {
		if _, err := svc.GetStats(context.Background(), "not-a-date", ""); err == nil {
			t.Error("GetStats(bad from) = nil error, want INVALID_ARGUMENT")
		}
	})

	t.Run("invalid date_to", func(t *testing.T) {
		if _, err := svc.GetStats(context.Background(), "", "not-a-date"); err == nil {
			t.Error("GetStats(bad to) = nil error, want INVALID_ARGUMENT")
		}
	})
}
