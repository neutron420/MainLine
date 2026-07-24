package postgres

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/schemahub/backend/internal/schema/domain"
)

type SchemaRepository struct {
	db *pgxpool.Pool
}

func NewSchemaRepository(db *pgxpool.Pool) *SchemaRepository {
	return &SchemaRepository{db: db}
}

func (r *SchemaRepository) Create(ctx context.Context, s *domain.Schema) error {
	s.ID = uuid.NewString()
	s.CreatedAt = time.Now()
	s.UpdatedAt = time.Now()

	_, err := r.db.Exec(ctx,
		`INSERT INTO schemas (id, project_id, connection_id, schema_name, created_at, updated_at)
		 VALUES ($1, $2, $3, $4, $5, $6)`,
		s.ID, s.ProjectID, s.ConnectionID, s.SchemaName, s.CreatedAt, s.UpdatedAt,
	)
	return err
}

func (r *SchemaRepository) GetByID(ctx context.Context, id string) (*domain.Schema, error) {
	row := r.db.QueryRow(ctx,
		`SELECT id, project_id, connection_id, schema_name, current_version_id, last_introspected_at, created_at, updated_at, deleted_at
		 FROM schemas WHERE id = $1 AND deleted_at IS NULL`, id)

	s := &domain.Schema{}
	if err := row.Scan(&s.ID, &s.ProjectID, &s.ConnectionID, &s.SchemaName, &s.CurrentVersionID, &s.LastIntrospectedAt, &s.CreatedAt, &s.UpdatedAt, &s.DeletedAt); err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("schema not found")
		}
		return nil, err
	}
	return s, nil
}

func (r *SchemaRepository) ListByProjectID(ctx context.Context, projectID, cursor string, limit int32) ([]*domain.Schema, string, int32, error) {
	query := `SELECT id, project_id, connection_id, schema_name, current_version_id, last_introspected_at, created_at, updated_at, deleted_at
		FROM schemas WHERE project_id = $1 AND deleted_at IS NULL ORDER BY created_at DESC`

	var args []interface{}
	args = append(args, projectID)

	if cursor != "" {
		query = `SELECT id, project_id, connection_id, schema_name, current_version_id, last_introspected_at, created_at, updated_at, deleted_at
			FROM schemas WHERE project_id = $1 AND deleted_at IS NULL AND created_at < $2 ORDER BY created_at DESC`
		args = append(args, cursor)
	}

	if limit <= 0 { limit = 20 }
	query += fmt.Sprintf(" LIMIT %d", limit+1)

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, "", 0, err
	}
	defer rows.Close()

	var schemas []*domain.Schema
	count := int32(0)
	for rows.Next() {
		s := &domain.Schema{}
		if err := rows.Scan(&s.ID, &s.ProjectID, &s.ConnectionID, &s.SchemaName, &s.CurrentVersionID, &s.LastIntrospectedAt, &s.CreatedAt, &s.UpdatedAt, &s.DeletedAt); err != nil {
			return nil, "", 0, err
		}
		schemas = append(schemas, s)
		count++
	}

	var nextCursor string
	if int32(len(schemas)) > limit {
		nextCursor = schemas[len(schemas)-1].CreatedAt.Format(time.RFC3339Nano)
		schemas = schemas[:len(schemas)-1]
	}
	return schemas, nextCursor, count, nil
}

func (r *SchemaRepository) GetByConnectionAndSchema(ctx context.Context, connID, schemaName string) (*domain.Schema, error) {
	row := r.db.QueryRow(ctx,
		`SELECT id, project_id, connection_id, schema_name, current_version_id, last_introspected_at, created_at, updated_at, deleted_at
		 FROM schemas WHERE connection_id = $1 AND schema_name = $2 AND deleted_at IS NULL`, connID, schemaName)

	s := &domain.Schema{}
	if err := row.Scan(&s.ID, &s.ProjectID, &s.ConnectionID, &s.SchemaName, &s.CurrentVersionID, &s.LastIntrospectedAt, &s.CreatedAt, &s.UpdatedAt, &s.DeletedAt); err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("schema not found")
		}
		return nil, err
	}
	return s, nil
}

func (r *SchemaRepository) UpdateCurrentVersion(ctx context.Context, schemaID, versionID string) error {
	now := time.Now()
	_, err := r.db.Exec(ctx,
		`UPDATE schemas SET current_version_id = $1, last_introspected_at = $2, updated_at = $2 WHERE id = $3`,
		versionID, now, schemaID)
	return err
}

func (r *SchemaRepository) CreateVersion(ctx context.Context, v *domain.SchemaVersion) error {
	v.ID = uuid.NewString()
	v.CreatedAt = time.Now()

	_, err := r.db.Exec(ctx,
		`INSERT INTO schema_versions (id, schema_id, version, checksum, metadata, object_count, parent_version_id, created_by, created_at)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)`,
		v.ID, v.SchemaID, v.Version, v.Checksum, v.Metadata, v.ObjectCount, v.ParentVersionID, v.CreatedBy, v.CreatedAt)
	return err
}

func (r *SchemaRepository) GetVersionByID(ctx context.Context, id string) (*domain.SchemaVersion, error) {
	row := r.db.QueryRow(ctx,
		`SELECT id, schema_id, version, checksum, metadata, object_count, parent_version_id, created_by, created_at
		 FROM schema_versions WHERE id = $1`, id)

	v := &domain.SchemaVersion{}
	if err := row.Scan(&v.ID, &v.SchemaID, &v.Version, &v.Checksum, &v.Metadata, &v.ObjectCount, &v.ParentVersionID, &v.CreatedBy, &v.CreatedAt); err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("version not found")
		}
		return nil, err
	}
	return v, nil
}

func (r *SchemaRepository) ListVersionsBySchemaID(ctx context.Context, schemaID, cursor string, limit int32) ([]*domain.SchemaVersion, string, int32, error) {
	query := `SELECT id, schema_id, version, checksum, metadata, object_count, parent_version_id, created_by, created_at
		FROM schema_versions WHERE schema_id = $1 ORDER BY version DESC`

	var args []interface{}
	args = append(args, schemaID)

	if cursor != "" {
		query = `SELECT id, schema_id, version, checksum, metadata, object_count, parent_version_id, created_by, created_at
			FROM schema_versions WHERE schema_id = $1 AND version < $2 ORDER BY version DESC`
		args = append(args, cursor)
	}

	if limit <= 0 { limit = 20 }
	query += fmt.Sprintf(" LIMIT %d", limit+1)

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, "", 0, err
	}
	defer rows.Close()

	var versions []*domain.SchemaVersion
	count := int32(0)
	for rows.Next() {
		v := &domain.SchemaVersion{}
		if err := rows.Scan(&v.ID, &v.SchemaID, &v.Version, &v.Checksum, &v.Metadata, &v.ObjectCount, &v.ParentVersionID, &v.CreatedBy, &v.CreatedAt); err != nil {
			return nil, "", 0, err
		}
		versions = append(versions, v)
		count++
	}

	var nextCursor string
	if int32(len(versions)) > limit {
		nextCursor = fmt.Sprintf("%d", versions[len(versions)-1].Version)
		versions = versions[:len(versions)-1]
	}
	return versions, nextCursor, count, nil
}

func (r *SchemaRepository) GetLatestVersion(ctx context.Context, schemaID string) (*domain.SchemaVersion, error) {
	row := r.db.QueryRow(ctx,
		`SELECT id, schema_id, version, checksum, metadata, object_count, parent_version_id, created_by, created_at
		 FROM schema_versions WHERE schema_id = $1 ORDER BY version DESC LIMIT 1`, schemaID)

	v := &domain.SchemaVersion{}
	if err := row.Scan(&v.ID, &v.SchemaID, &v.Version, &v.Checksum, &v.Metadata, &v.ObjectCount, &v.ParentVersionID, &v.CreatedBy, &v.CreatedAt); err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("no versions found")
		}
		return nil, err
	}
	return v, nil
}

func (r *SchemaRepository) CreateObjects(ctx context.Context, objects []*domain.SchemaObject) error {
	for _, o := range objects {
		o.ID = uuid.NewString()
		_, err := r.db.Exec(ctx,
			`INSERT INTO schema_objects (id, schema_version_id, object_type, object_name, object_schema, definition)
			 VALUES ($1, $2, $3, $4, $5, $6)`,
			o.ID, o.SchemaVersionID, o.ObjectType, o.ObjectName, o.ObjectSchema, o.Definition)
		if err != nil {
			return err
		}
	}
	return nil
}

func (r *SchemaRepository) ListObjectsByVersionID(ctx context.Context, versionID, objectType, cursor string, limit int32) ([]*domain.SchemaObject, string, int32, error) {
	query := `SELECT id, schema_version_id, object_type, object_name, object_schema, definition, parent_object_id
		FROM schema_objects WHERE schema_version_id = $1`

	var args []interface{}
	args = append(args, versionID)

	if objectType != "" {
		query += " AND object_type = $2"
		args = append(args, objectType)
	}

	if cursor != "" {
		offset := len(args) + 1
		if objectType == "" {
			query += fmt.Sprintf(" AND id > $2")
		} else {
			query += fmt.Sprintf(" AND id > $%d", offset)
		}
		args = append(args, cursor)
	}

	query += " ORDER BY object_type, object_name"

	if limit <= 0 { limit = 20 }
	query += fmt.Sprintf(" LIMIT %d", limit+1)

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, "", 0, err
	}
	defer rows.Close()

	var objects []*domain.SchemaObject
	count := int32(0)
	for rows.Next() {
		o := &domain.SchemaObject{}
		if err := rows.Scan(&o.ID, &o.SchemaVersionID, &o.ObjectType, &o.ObjectName, &o.ObjectSchema, &o.Definition, &o.ParentObjectID); err != nil {
			return nil, "", 0, err
		}
		objects = append(objects, o)
		count++
	}

	var nextCursor string
	if int32(len(objects)) > limit {
		nextCursor = objects[len(objects)-1].ID
		objects = objects[:len(objects)-1]
	}
	return objects, nextCursor, count, nil
}

var _ = json.RawMessage{}
