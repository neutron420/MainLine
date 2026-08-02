package worker

import (
	"context"
	"testing"
	"time"

	"github.com/schemahub/backend/internal/drift/domain"
	projDomain "github.com/schemahub/backend/internal/project/domain"
	"github.com/schemahub/backend/pkg/encryption"
)

type fakeConnRepoForDrift struct {
	conns []projDomain.Connection
}

func (f *fakeConnRepoForDrift) Create(ctx context.Context, c *projDomain.Connection) error {
	return nil
}
func (f *fakeConnRepoForDrift) GetByID(ctx context.Context, id string) (*projDomain.Connection, error) {
	for i := range f.conns {
		if f.conns[i].ID == id {
			return &f.conns[i], nil
		}
	}
	return nil, nil
}
func (f *fakeConnRepoForDrift) ListByProjectID(ctx context.Context, projectID, cursor string, limit int32) ([]*projDomain.Connection, string, int32, error) {
	var out []*projDomain.Connection
	for i := range f.conns {
		out = append(out, &f.conns[i])
	}
	return out, "", int32(len(out)), nil
}
func (f *fakeConnRepoForDrift) ListAll(ctx context.Context) ([]*projDomain.Connection, error) {
	var out []*projDomain.Connection
	for i := range f.conns {
		out = append(out, &f.conns[i])
	}
	return out, nil
}
func (f *fakeConnRepoForDrift) Update(ctx context.Context, c *projDomain.Connection) error {
	return nil
}
func (f *fakeConnRepoForDrift) SoftDelete(ctx context.Context, id string) error { return nil }
func (f *fakeConnRepoForDrift) UpdateStatus(ctx context.Context, id string, status projDomain.ConnectionStatus, lastConnectedAt *time.Time) error {
	return nil
}

type recordingComparator struct {
	checked []string
	err     error
}

func (r *recordingComparator) CompareLiveWithVersion(ctx context.Context, connStr, connectionID string, schemaNames []string) ([]*domain.DriftEvent, error) {
	r.checked = append(r.checked, connectionID)
	return nil, r.err
}

func TestDriftCheckWorker_RunsForActiveConnections(t *testing.T) {
	key := []byte("0123456789abcdef0123456789abcdef")
	enc, err := encryption.Encrypt([]byte("secret"), key)
	if err != nil {
		t.Fatalf("encrypting password: %v", err)
	}

	repo := &fakeConnRepoForDrift{conns: []projDomain.Connection{
		{ID: "conn_1", Username: "u1", Host: "h1", Port: 5432, DatabaseName: "db1", SSLMode: projDomain.SSLRequire, PasswordEncrypted: enc},
		{ID: "conn_2", Username: "u2", Host: "h2", Port: 5432, DatabaseName: "db2", SSLMode: projDomain.SSLRequire, PasswordEncrypted: enc},
	}}
	comparator := &recordingComparator{}
	driftSvc := domain.NewDriftService(&fakeDriftRepoWorker{}, comparator)
	w := NewDriftCheckWorker(repo, driftSvc, key)

	if err := w.Run(context.Background()); err != nil {
		t.Fatalf("Run() error = %v", err)
	}

	if len(comparator.checked) != 2 {
		t.Fatalf("checked %d connections, want 2", len(comparator.checked))
	}
}

func TestDriftCheckWorker_SkipsDeletedConnections(t *testing.T) {
	key := []byte("0123456789abcdef0123456789abcdef")
	enc, err := encryption.Encrypt([]byte("secret"), key)
	if err != nil {
		t.Fatalf("encrypting password: %v", err)
	}

	now := time.Now()
	repo := &fakeConnRepoForDrift{conns: []projDomain.Connection{
		{ID: "conn_1", Username: "u1", Host: "h1", Port: 5432, DatabaseName: "db1", SSLMode: projDomain.SSLRequire, PasswordEncrypted: enc, DeletedAt: &now},
		{ID: "conn_2", Username: "u2", Host: "h2", Port: 5432, DatabaseName: "db2", SSLMode: projDomain.SSLRequire, PasswordEncrypted: enc},
	}}
	comparator := &recordingComparator{}
	driftSvc := domain.NewDriftService(&fakeDriftRepoWorker{}, comparator)
	w := NewDriftCheckWorker(repo, driftSvc, key)

	if err := w.Run(context.Background()); err != nil {
		t.Fatalf("Run() error = %v", err)
	}

	if len(comparator.checked) != 1 {
		t.Fatalf("checked %d connections, want 1 (deleted skipped)", len(comparator.checked))
	}
}

type fakeDriftRepoWorker struct{}

func (f *fakeDriftRepoWorker) Insert(ctx context.Context, e *domain.DriftEvent) error { return nil }
func (f *fakeDriftRepoWorker) GetByID(ctx context.Context, id string) (*domain.DriftEvent, error) {
	return nil, nil
}
func (f *fakeDriftRepoWorker) List(ctx context.Context, filter *domain.DriftFilter, cursor string, limit int32) ([]*domain.DriftEvent, string, int32, error) {
	return nil, "", 0, nil
}
func (f *fakeDriftRepoWorker) UpdateStatus(ctx context.Context, id string, status domain.DriftStatus, resolvedBy string) error {
	return nil
}
func (f *fakeDriftRepoWorker) GetStats(ctx context.Context, connectionID string) (*domain.DriftStats, error) {
	return &domain.DriftStats{}, nil
}
