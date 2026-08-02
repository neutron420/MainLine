package handler

import (
	"context"
	"sort"
	"testing"
	"time"

	"github.com/schemahub/backend/internal/pkg/errors"
	"github.com/schemahub/backend/internal/schema/domain"
	schemav1 "github.com/schemahub/backend/proto/schema/v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type fakeSchemaRepo struct {
	schemas  map[string]*domain.Schema
	versions map[string]*domain.SchemaVersion
	notFound bool
}

func newFakeSchemaRepo() *fakeSchemaRepo {
	return &fakeSchemaRepo{
		schemas:  map[string]*domain.Schema{},
		versions: map[string]*domain.SchemaVersion{},
	}
}

func (f *fakeSchemaRepo) Create(ctx context.Context, s *domain.Schema) error { return nil }
func (f *fakeSchemaRepo) GetByID(ctx context.Context, id string) (*domain.Schema, error) {
	if f.notFound {
		return nil, errors.New("NOT_FOUND", "schema not found")
	}
	s, ok := f.schemas[id]
	if !ok {
		return nil, errors.New("NOT_FOUND", "schema not found")
	}
	return s, nil
}
func (f *fakeSchemaRepo) ListByProjectID(ctx context.Context, projectID, cursor string, limit int32) ([]*domain.Schema, string, int32, error) {
	return nil, "", 0, nil
}
func (f *fakeSchemaRepo) GetByConnectionAndSchema(ctx context.Context, connID, schemaName string) (*domain.Schema, error) {
	return nil, errors.New("NOT_FOUND", "not found")
}
func (f *fakeSchemaRepo) UpdateCurrentVersion(ctx context.Context, schemaID, versionID string) error {
	return nil
}
func (f *fakeSchemaRepo) CreateVersion(ctx context.Context, v *domain.SchemaVersion) error {
	return nil
}
func (f *fakeSchemaRepo) GetVersionByID(ctx context.Context, id string) (*domain.SchemaVersion, error) {
	if f.notFound {
		return nil, errors.New("NOT_FOUND", "version not found")
	}
	v, ok := f.versions[id]
	if !ok {
		return nil, errors.New("NOT_FOUND", "version not found")
	}
	return v, nil
}
func (f *fakeSchemaRepo) ListVersionsBySchemaID(ctx context.Context, schemaID, cursor string, limit int32) ([]*domain.SchemaVersion, string, int32, error) {
	var out []*domain.SchemaVersion
	for _, v := range f.versions {
		if v.SchemaID == schemaID {
			out = append(out, v)
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Version > out[j].Version })
	return out, "", int32(len(out)), nil
}
func (f *fakeSchemaRepo) GetLatestVersion(ctx context.Context, schemaID string) (*domain.SchemaVersion, error) {
	return nil, errors.New("NOT_FOUND", "not found")
}
func (f *fakeSchemaRepo) CreateObjects(ctx context.Context, objects []*domain.SchemaObject) error {
	return nil
}
func (f *fakeSchemaRepo) ListObjectsByVersionID(ctx context.Context, versionID, objectType, cursor string, limit int32) ([]*domain.SchemaObject, string, int32, error) {
	return nil, "", 0, nil
}

func testSchemaHandler(t *testing.T, repo *fakeSchemaRepo) *SchemaHandler {
	t.Helper()
	svc := domain.NewSchemaService(repo)
	return NewSchemaHandler(svc, nil)
}

func TestSchemaHandler_GetSchema(t *testing.T) {
	t.Parallel()

	repo := newFakeSchemaRepo()
	repo.schemas["schema_1"] = &domain.Schema{
		ID: "schema_1", ProjectID: "proj_1", ConnectionID: "conn_1", SchemaName: "public",
		CreatedAt: time.Now(), UpdatedAt: time.Now(),
	}
	h := testSchemaHandler(t, repo)

	resp, err := h.GetSchema(context.Background(), &schemav1.GetSchemaRequest{Id: "schema_1"})
	if err != nil {
		t.Fatalf("GetSchema() error = %v", err)
	}
	if resp.Schema.Id != "schema_1" {
		t.Errorf("Schema.Id = %q, want schema_1", resp.Schema.Id)
	}
	if resp.Schema.SchemaName != "public" {
		t.Errorf("Schema.SchemaName = %q, want public", resp.Schema.SchemaName)
	}
}

func TestSchemaHandler_GetSchemaNotFound(t *testing.T) {
	t.Parallel()

	repo := newFakeSchemaRepo()
	repo.notFound = true
	h := testSchemaHandler(t, repo)

	_, err := h.GetSchema(context.Background(), &schemav1.GetSchemaRequest{Id: "ghost"})
	if status.Code(err) != codes.NotFound {
		t.Fatalf("GetSchema() error code = %v, want NotFound (%v)", status.Code(err), err)
	}
}

func TestSchemaHandler_ListSchemaVersions(t *testing.T) {
	t.Parallel()

	repo := newFakeSchemaRepo()
	repo.versions["v1"] = &domain.SchemaVersion{
		ID: "v1", SchemaID: "schema_1", Version: 1, Checksum: "abc", CreatedAt: time.Now(),
	}
	repo.versions["v2"] = &domain.SchemaVersion{
		ID: "v2", SchemaID: "schema_1", Version: 2, Checksum: "def", CreatedAt: time.Now(),
	}
	h := testSchemaHandler(t, repo)

	resp, err := h.ListSchemaVersions(context.Background(), &schemav1.ListSchemaVersionsRequest{
		SchemaId: "schema_1", PageSize: 10,
	})
	if err != nil {
		t.Fatalf("ListSchemaVersions() error = %v", err)
	}
	if len(resp.Versions) != 2 {
		t.Errorf("Versions len = %d, want 2", len(resp.Versions))
	}
	if resp.Versions[0].Checksum != "def" {
		t.Errorf("first version Checksum = %q, want def (newest first)", resp.Versions[0].Checksum)
	}
}

func TestSchemaHandler_GetSchemaVersion(t *testing.T) {
	t.Parallel()

	repo := newFakeSchemaRepo()
	repo.versions["v1"] = &domain.SchemaVersion{
		ID: "v1", SchemaID: "schema_1", Version: 1, Checksum: "abc", CreatedAt: time.Now(),
	}
	h := testSchemaHandler(t, repo)

	resp, err := h.GetSchemaVersion(context.Background(), &schemav1.GetSchemaVersionRequest{Id: "v1"})
	if err != nil {
		t.Fatalf("GetSchemaVersion() error = %v", err)
	}
	if resp.Version.Id != "v1" {
		t.Errorf("Version.Id = %q, want v1", resp.Version.Id)
	}
}
