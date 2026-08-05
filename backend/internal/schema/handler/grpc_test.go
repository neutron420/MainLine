package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"reflect"
	"sort"
	"strings"
	"testing"
	"time"

	"github.com/schemahub/backend/internal/pkg/errors"
	"github.com/schemahub/backend/internal/pkg/interceptor"
	"github.com/schemahub/backend/internal/schema/domain"
	schemav1 "github.com/schemahub/backend/proto/schema/v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type fakeSchemaRepo struct {
	schemas  map[string]*domain.Schema
	versions map[string]*domain.SchemaVersion
	objects  map[string][]*domain.SchemaObject
	notFound bool
}

func newFakeSchemaRepo() *fakeSchemaRepo {
	return &fakeSchemaRepo{
		schemas:  map[string]*domain.Schema{},
		versions: map[string]*domain.SchemaVersion{},
		objects:  map[string][]*domain.SchemaObject{},
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
	var out []*domain.Schema
	for _, s := range f.schemas {
		if s.ProjectID == projectID {
			out = append(out, s)
		}
	}
	return out, "", int32(len(out)), nil
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
	return nil, nil
}
func (f *fakeSchemaRepo) CreateObjects(ctx context.Context, objects []*domain.SchemaObject) error {
	return nil
}
func (f *fakeSchemaRepo) ListObjectsByVersionID(ctx context.Context, versionID, objectType, cursor string, limit int32) ([]*domain.SchemaObject, string, int32, error) {
	var out []*domain.SchemaObject
	for _, o := range f.objects[versionID] {
		if objectType != "" && o.ObjectType != objectType {
			continue
		}
		out = append(out, o)
	}
	return out, "", int32(len(out)), nil
}

func testSchemaHandler(t *testing.T, repo *fakeSchemaRepo) *SchemaHandler {
	t.Helper()
	svc := domain.NewSchemaService(repo)
	return NewSchemaHandler(svc, nil)
}

func testSchemaHandlerWithConn(t *testing.T, repo *fakeSchemaRepo, connInfo func(context.Context, string) (string, string, error)) *SchemaHandler {
	t.Helper()
	svc := domain.NewSchemaService(repo)
	return NewSchemaHandler(svc, connInfo)
}

func failConnector(ctx context.Context, connStr string) (domain.DBPool, error) {
	return nil, fmt.Errorf("connector disabled")
}

type introspectRows struct {
	rows [][]any
	idx  int
}

func (r *introspectRows) Next() bool {
	r.idx++
	return r.idx <= len(r.rows)
}

func (r *introspectRows) Scan(dest ...any) error {
	row := r.rows[r.idx-1]
	for i, d := range dest {
		val := row[i]
		if val == nil {
			continue
		}
		dv := reflect.ValueOf(d)
		if dv.Kind() != reflect.Ptr {
			return fmt.Errorf("scan destination %d is not a pointer", i)
		}
		ev := dv.Elem()
		if ev.Kind() == reflect.Ptr {
			ev.Set(reflect.New(ev.Type().Elem()))
			ev = ev.Elem()
		}
		sv := reflect.ValueOf(val)
		switch {
		case sv.Type().AssignableTo(ev.Type()):
			ev.Set(sv)
		case ev.Kind() == reflect.String:
			ev.SetString(fmt.Sprintf("%v", val))
		case ev.Kind() == reflect.Int || ev.Kind() == reflect.Int32 || ev.Kind() == reflect.Int64:
			ev.SetInt(sv.Int())
		}
	}
	return nil
}

func (r *introspectRows) Close() {}

type introspectPool struct {
	tables     [][]any
	columns    [][]any
	indexes    [][]any
	pks        [][]any
	fks        [][]any
	uniques    [][]any
	enums      [][]any
	extensions [][]any
}

func (p *introspectPool) Query(ctx context.Context, sql string, args ...any) (domain.Rows, error) {
	rows := &introspectRows{}
	switch {
	case strings.Contains(sql, "information_schema.tables"):
		rows.rows = p.tables
	case strings.Contains(sql, "information_schema.columns"):
		rows.rows = p.columns
	case strings.Contains(sql, "pg_indexes"):
		rows.rows = p.indexes
	case strings.Contains(sql, "PRIMARY KEY"):
		rows.rows = p.pks
	case strings.Contains(sql, "FOREIGN KEY"):
		rows.rows = p.fks
	case strings.Contains(sql, "UNIQUE"):
		rows.rows = p.uniques
	case strings.Contains(sql, "pg_enum"):
		rows.rows = p.enums
	case strings.Contains(sql, "pg_extension"):
		rows.rows = p.extensions
	default:
		return nil, fmt.Errorf("unexpected introspection query: %s", sql)
	}
	return rows, nil
}

func (p *introspectPool) Close() {}

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

func TestSchemaHandler_IntrospectSchema(t *testing.T) {
	pool := &introspectPool{
		tables:     [][]any{{"users", "public"}},
		columns:    [][]any{{"id", "integer", "NO", nil, nil, 1}, {"email", "text", "YES", nil, nil, 2}},
		indexes:    [][]any{{"users_pkey", "CREATE UNIQUE INDEX users_pkey ON public.users USING btree (id)"}},
		pks:        [][]any{{"id"}},
		enums:      [][]any{{"mood", []string{"happy", "sad"}}},
		extensions: [][]any{{"postgis"}},
	}
	domain.SetConnector(func(ctx context.Context, connStr string) (domain.DBPool, error) {
		return pool, nil
	})
	defer domain.SetConnector(failConnector)

	repo := newFakeSchemaRepo()
	h := testSchemaHandlerWithConn(t, repo, func(ctx context.Context, connID string) (string, string, error) {
		return "postgres://fake", "proj_1", nil
	})
	ctx := context.WithValue(context.Background(), interceptor.UserIDKey, "user_1")

	resp, err := h.IntrospectSchema(ctx, &schemav1.IntrospectSchemaRequest{
		ConnectionId: "conn_1", SchemaNames: []string{"public"},
	})
	if err != nil {
		t.Fatalf("IntrospectSchema() error = %v", err)
	}
	if resp.Schema == nil || resp.Schema.ConnectionId != "conn_1" {
		t.Errorf("Schema = %+v, want connection conn_1", resp.Schema)
	}
	if resp.SchemaVersion == nil {
		t.Fatal("IntrospectSchema() returned nil schema version")
	}
	if resp.SchemaVersion.Version != 1 {
		t.Errorf("SchemaVersion.Version = %d, want 1", resp.SchemaVersion.Version)
	}
	if resp.SchemaVersion.ObjectCount != 3 {
		t.Errorf("SchemaVersion.ObjectCount = %d, want 3", resp.SchemaVersion.ObjectCount)
	}
	if resp.SchemaVersion.CreatedBy != "user_1" {
		t.Errorf("SchemaVersion.CreatedBy = %q, want user_1", resp.SchemaVersion.CreatedBy)
	}
}

func TestSchemaHandler_IntrospectSchemaConnStringError(t *testing.T) {
	t.Parallel()

	repo := newFakeSchemaRepo()
	h := testSchemaHandlerWithConn(t, repo, func(ctx context.Context, connID string) (string, string, error) {
		return "", "", fmt.Errorf("connection not found")
	})

	_, err := h.IntrospectSchema(context.Background(), &schemav1.IntrospectSchemaRequest{ConnectionId: "conn_1"})
	if status.Code(err) != codes.FailedPrecondition {
		t.Fatalf("IntrospectSchema() error code = %v, want FailedPrecondition (%v)", status.Code(err), err)
	}
}

func TestSchemaHandler_IntrospectSchemaIntrospectionError(t *testing.T) {
	domain.SetConnector(func(ctx context.Context, connStr string) (domain.DBPool, error) {
		return nil, fmt.Errorf("connection refused")
	})
	defer domain.SetConnector(failConnector)

	repo := newFakeSchemaRepo()
	h := testSchemaHandlerWithConn(t, repo, func(ctx context.Context, connID string) (string, string, error) {
		return "postgres://fake", "proj_1", nil
	})

	_, err := h.IntrospectSchema(context.Background(), &schemav1.IntrospectSchemaRequest{ConnectionId: "conn_1"})
	if status.Code(err) != codes.Internal {
		t.Fatalf("IntrospectSchema() error code = %v, want Internal (%v)", status.Code(err), err)
	}
}

func TestSchemaHandler_ListSchemas(t *testing.T) {
	t.Parallel()

	repo := newFakeSchemaRepo()
	repo.schemas["schema_1"] = &domain.Schema{
		ID: "schema_1", ProjectID: "proj_1", ConnectionID: "conn_1", SchemaName: "public",
		CreatedAt: time.Now(), UpdatedAt: time.Now(),
	}
	repo.schemas["schema_2"] = &domain.Schema{
		ID: "schema_2", ProjectID: "proj_1", ConnectionID: "conn_2", SchemaName: "analytics",
		CreatedAt: time.Now(), UpdatedAt: time.Now(),
	}
	h := testSchemaHandler(t, repo)

	resp, err := h.ListSchemas(context.Background(), &schemav1.ListSchemasRequest{
		ProjectId: "proj_1", PageSize: 10,
	})
	if err != nil {
		t.Fatalf("ListSchemas() error = %v", err)
	}
	if len(resp.Schemas) != 2 {
		t.Errorf("Schemas len = %d, want 2", len(resp.Schemas))
	}
	if resp.TotalCount != 2 {
		t.Errorf("TotalCount = %d, want 2", resp.TotalCount)
	}
}

func TestSchemaHandler_GetSchemaVersionNotFound(t *testing.T) {
	t.Parallel()

	repo := newFakeSchemaRepo()
	repo.notFound = true
	h := testSchemaHandler(t, repo)

	_, err := h.GetSchemaVersion(context.Background(), &schemav1.GetSchemaVersionRequest{Id: "ghost"})
	if status.Code(err) != codes.NotFound {
		t.Fatalf("GetSchemaVersion() error code = %v, want NotFound (%v)", status.Code(err), err)
	}
}

func TestSchemaHandler_CompareSchemaVersions(t *testing.T) {
	t.Parallel()

	repo := newFakeSchemaRepo()
	repo.versions["v1"] = &domain.SchemaVersion{
		ID: "v1", SchemaID: "schema_1", Version: 1, CreatedAt: time.Now(),
		Metadata: json.RawMessage(`{"tables":[{"name":"users","schema":"public"}],"enums":[],"extensions":[]}`),
	}
	repo.versions["v2"] = &domain.SchemaVersion{
		ID: "v2", SchemaID: "schema_1", Version: 2, CreatedAt: time.Now(),
		Metadata: json.RawMessage(`{"tables":[{"name":"users","schema":"public"},{"name":"posts","schema":"public"}],"enums":[],"extensions":[]}`),
	}
	h := testSchemaHandler(t, repo)

	resp, err := h.CompareSchemaVersions(context.Background(), &schemav1.CompareSchemaVersionsRequest{
		VersionAId: "v1", VersionBId: "v2",
	})
	if err != nil {
		t.Fatalf("CompareSchemaVersions() error = %v", err)
	}
	if len(resp.Diff.AddedObjects) != 1 {
		t.Fatalf("AddedObjects len = %d, want 1", len(resp.Diff.AddedObjects))
	}
	if resp.Diff.AddedObjects[0].Name != "public.posts" {
		t.Errorf("added object Name = %q, want public.posts", resp.Diff.AddedObjects[0].Name)
	}
}

func TestSchemaHandler_CompareSchemaVersionsMissing(t *testing.T) {
	t.Parallel()

	repo := newFakeSchemaRepo()
	repo.versions["v1"] = &domain.SchemaVersion{
		ID: "v1", SchemaID: "schema_1", Version: 1, CreatedAt: time.Now(),
		Metadata: json.RawMessage(`{"tables":[],"enums":[],"extensions":[]}`),
	}
	h := testSchemaHandler(t, repo)

	_, err := h.CompareSchemaVersions(context.Background(), &schemav1.CompareSchemaVersionsRequest{
		VersionAId: "v1", VersionBId: "ghost",
	})
	if status.Code(err) != codes.Internal {
		t.Fatalf("CompareSchemaVersions() error code = %v, want Internal (%v)", status.Code(err), err)
	}
}

func TestSchemaHandler_ListSchemaObjects(t *testing.T) {
	t.Parallel()

	repo := newFakeSchemaRepo()
	repo.objects["v1"] = []*domain.SchemaObject{
		{ID: "obj_1", SchemaVersionID: "v1", ObjectType: "table", ObjectName: "users", ObjectSchema: "public", Definition: json.RawMessage(`{}`)},
		{ID: "obj_2", SchemaVersionID: "v1", ObjectType: "table", ObjectName: "posts", ObjectSchema: "public", Definition: json.RawMessage(`{}`)},
	}
	h := testSchemaHandler(t, repo)

	resp, err := h.ListSchemaObjects(context.Background(), &schemav1.ListSchemaObjectsRequest{
		SchemaVersionId: "v1", PageSize: 10,
	})
	if err != nil {
		t.Fatalf("ListSchemaObjects() error = %v", err)
	}
	if len(resp.Objects) != 2 {
		t.Errorf("Objects len = %d, want 2", len(resp.Objects))
	}
	if resp.TotalCount != 2 {
		t.Errorf("TotalCount = %d, want 2", resp.TotalCount)
	}
	if resp.Objects[0].ObjectName != "users" {
		t.Errorf("first object Name = %q, want users", resp.Objects[0].ObjectName)
	}
}

func TestSchemaHandler_GetSchemaDiagram(t *testing.T) {
	t.Parallel()

	repo := newFakeSchemaRepo()
	repo.versions["v1"] = &domain.SchemaVersion{
		ID: "v1", SchemaID: "schema_1", Version: 1, CreatedAt: time.Now(),
		Metadata: json.RawMessage(`{"tables":[{"name":"users","schema":"public","columns":[{"name":"id","data_type":"integer","is_nullable":false,"ordinal_position":1}],"indexes":[],"constraints":{"primary_key":["id"],"foreign_keys":[{"column":"org_id","ref_table":"orgs","ref_column":"id","name":"fk_orgs"}],"uniques":[]}}],"enums":[],"extensions":[]}`),
	}
	h := testSchemaHandler(t, repo)

	resp, err := h.GetSchemaDiagram(context.Background(), &schemav1.GetSchemaDiagramRequest{
		SchemaVersionId: "v1", IncludeDetails: true,
	})
	if err != nil {
		t.Fatalf("GetSchemaDiagram() error = %v", err)
	}
	if len(resp.Nodes) != 1 {
		t.Fatalf("Nodes len = %d, want 1", len(resp.Nodes))
	}
	if resp.Nodes[0].Id != "public.users" {
		t.Errorf("Node Id = %q, want public.users", resp.Nodes[0].Id)
	}
	if len(resp.Nodes[0].Data.Columns) != 1 {
		t.Errorf("Node columns len = %d, want 1", len(resp.Nodes[0].Data.Columns))
	}
	if len(resp.Edges) != 1 {
		t.Fatalf("Edges len = %d, want 1", len(resp.Edges))
	}
	if resp.Edges[0].Target != "public.orgs" {
		t.Errorf("Edge Target = %q, want public.orgs", resp.Edges[0].Target)
	}
}

func TestSchemaHandler_GetSchemaDiagramVersionNotFound(t *testing.T) {
	t.Parallel()

	repo := newFakeSchemaRepo()
	repo.notFound = true
	h := testSchemaHandler(t, repo)

	_, err := h.GetSchemaDiagram(context.Background(), &schemav1.GetSchemaDiagramRequest{
		SchemaVersionId: "ghost",
	})
	if status.Code(err) != codes.Internal {
		t.Fatalf("GetSchemaDiagram() error code = %v, want Internal (%v)", status.Code(err), err)
	}
}
