package domain

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"
)

type fakeConnRepo struct {
	conns map[string]*Connection
}

func (f *fakeConnRepo) Create(ctx context.Context, c *Connection) error {
	if f.conns == nil {
		f.conns = map[string]*Connection{}
	}
	f.conns[c.ID] = c
	return nil
}

func (f *fakeConnRepo) GetByID(ctx context.Context, id string) (*Connection, error) {
	c, ok := f.conns[id]
	if !ok {
		return nil, errors.New("connection not found")
	}
	return c, nil
}

func (f *fakeConnRepo) ListByProjectID(ctx context.Context, projectID, cursor string, limit int32) ([]*Connection, string, int32, error) {
	var out []*Connection
	for _, c := range f.conns {
		if c.ProjectID == projectID {
			out = append(out, c)
		}
	}
	return out, "", int32(len(out)), nil
}

func (f *fakeConnRepo) ListAll(ctx context.Context) ([]*Connection, error) {
	var out []*Connection
	for _, c := range f.conns {
		out = append(out, c)
	}
	return out, nil
}

func (f *fakeConnRepo) Update(ctx context.Context, c *Connection) error {
	f.conns[c.ID] = c
	return nil
}

func (f *fakeConnRepo) SoftDelete(ctx context.Context, id string) error {
	if c, ok := f.conns[id]; ok {
		c.DeletedAt = &c.CreatedAt
	}
	return nil
}

func (f *fakeConnRepo) UpdateStatus(ctx context.Context, id string, status ConnectionStatus, lastConnectedAt *time.Time) error {
	if c, ok := f.conns[id]; ok {
		c.ConnectionStatus = status
		c.LastConnectedAt = lastConnectedAt
	}
	return nil
}

func testKey() []byte {
	key := make([]byte, 32)
	for i := range key {
		key[i] = byte(i + 1)
	}
	return key
}

func TestConnectionService_CreateEncryptsPassword(t *testing.T) {
	t.Parallel()

	repo := &fakeConnRepo{}
	svc := NewConnectionService(repo, testKey())

	conn := &Connection{ID: "c1", Name: "Prod", Host: "db.example.com", Port: 5432, DatabaseName: "app", Username: "admin", SSLMode: SSLRequire}
	created, err := svc.Create(context.Background(), conn, "s3cr3t!")
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	if created.PasswordEncrypted == "" {
		t.Fatal("PasswordEncrypted is empty, want encrypted value")
	}
	if strings.Contains(created.PasswordEncrypted, "s3cr3t!") {
		t.Error("PasswordEncrypted contains plaintext password")
	}
	if created.PasswordEncrypted == "s3cr3t!" {
		t.Error("password stored in plaintext")
	}
	if created.ConnectionStatus != ConnStatusUnknown {
		t.Errorf("ConnectionStatus = %q, want unknown", created.ConnectionStatus)
	}
}

func TestConnectionService_GetConnectionString(t *testing.T) {
	t.Parallel()

	repo := &fakeConnRepo{}
	svc := NewConnectionService(repo, testKey())

	conn := &Connection{ID: "c1", Name: "Prod", Host: "db.example.com", Port: 5432, DatabaseName: "app", Username: "admin", SSLMode: SSLRequire}
	if _, err := svc.Create(context.Background(), conn, "pw"); err != nil {
		t.Fatal(err)
	}

	got, err := svc.GetConnectionString(context.Background(), "c1")
	if err != nil {
		t.Fatalf("GetConnectionString() error = %v", err)
	}

	want := "postgres://admin:pw@db.example.com:5432/app?sslmode=require"
	if got != want {
		t.Errorf("GetConnectionString() = %q, want %q", got, want)
	}
}

func TestConnectionService_GetConnectionStringUnknown(t *testing.T) {
	t.Parallel()

	svc := NewConnectionService(&fakeConnRepo{}, testKey())
	if _, err := svc.GetConnectionString(context.Background(), "missing"); err == nil {
		t.Error("GetConnectionString(missing) = nil error, want error")
	}
}

func TestConnectionService_Update(t *testing.T) {
	t.Parallel()

	repo := &fakeConnRepo{}
	svc := NewConnectionService(repo, testKey())

	conn := &Connection{ID: "c1", Name: "Old", Host: "old.com", Port: 5432, DatabaseName: "app", Username: "admin", SSLMode: SSLDisable, PasswordEncrypted: "old"}
	repo.conns = map[string]*Connection{"c1": conn}

	updated, err := svc.Update(context.Background(), "c1", "New Name", "new.com", 6543, "", "", "newpass", "require")
	if err != nil {
		t.Fatalf("Update() error = %v", err)
	}

	if updated.Name != "New Name" || updated.Host != "new.com" || updated.Port != 6543 {
		t.Errorf("Update() did not apply fields: %+v", updated)
	}
	if updated.SSLMode != SSLRequire {
		t.Errorf("SSLMode = %q, want require", updated.SSLMode)
	}
	if updated.PasswordEncrypted == "old" || updated.PasswordEncrypted == "newpass" {
		t.Error("password was not re-encrypted")
	}

	got, err := svc.GetConnectionString(context.Background(), "c1")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(got, "newpass@new.com:6543") {
		t.Errorf("connection string not updated: %q", got)
	}
}

func TestConnectionService_UpdateMissing(t *testing.T) {
	t.Parallel()

	svc := NewConnectionService(&fakeConnRepo{}, testKey())
	if _, err := svc.Update(context.Background(), "nope", "x", "y", 1, "z", "u", "", ""); err == nil {
		t.Error("Update(missing) = nil error, want error")
	}
}

func TestConnectionService_ListDefaultsPageSize(t *testing.T) {
	t.Parallel()

	repo := &fakeConnRepo{
		conns: map[string]*Connection{
			"c1": {ID: "c1", ProjectID: "p1"},
			"c2": {ID: "c2", ProjectID: "p1"},
			"c3": {ID: "c3", ProjectID: "p2"},
		},
	}
	svc := NewConnectionService(repo, testKey())

	conns, _, total, err := svc.List(context.Background(), "p1", "", 0)
	if err != nil {
		t.Fatal(err)
	}
	if total != 2 || len(conns) != 2 {
		t.Errorf("List() = %d conns, total %d; want 2, 2", len(conns), total)
	}
}

func TestConnectionService_Delete(t *testing.T) {
	t.Parallel()

	repo := &fakeConnRepo{conns: map[string]*Connection{"c1": {ID: "c1"}}}
	svc := NewConnectionService(repo, testKey())

	if err := svc.Delete(context.Background(), "c1"); err != nil {
		t.Fatalf("Delete() error = %v", err)
	}
}

func TestConnection_Validate(t *testing.T) {
	t.Parallel()

	valid := &Connection{Name: "x", Host: "h", Port: 5432, DatabaseName: "d", Username: "u"}
	if err := valid.Validate(); err != nil {
		t.Errorf("Validate(valid) error = %v, want nil", err)
	}
	if valid.SSLMode != SSLRequire {
		t.Errorf("default SSLMode = %q, want require", valid.SSLMode)
	}

	tests := []struct {
		name string
		c    *Connection
	}{
		{name: "missing name", c: &Connection{Host: "h", Port: 5432, DatabaseName: "d", Username: "u"}},
		{name: "missing host", c: &Connection{Name: "x", Port: 5432, DatabaseName: "d", Username: "u"}},
		{name: "bad port", c: &Connection{Name: "x", Host: "h", Port: 0, DatabaseName: "d", Username: "u"}},
		{name: "port too high", c: &Connection{Name: "x", Host: "h", Port: 70000, DatabaseName: "d", Username: "u"}},
		{name: "missing database", c: &Connection{Name: "x", Host: "h", Port: 5432, Username: "u"}},
		{name: "missing username", c: &Connection{Name: "x", Host: "h", Port: 5432, DatabaseName: "d"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.c.Validate(); err == nil {
				t.Errorf("Validate(%s) = nil error, want error", tt.name)
			}
		})
	}
}
