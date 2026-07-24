package domain

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"time"
)

type Schema struct {
	ID                 string
	ProjectID          string
	ConnectionID       string
	SchemaName         string
	CurrentVersionID   *string
	LastIntrospectedAt *time.Time
	CreatedAt          time.Time
	UpdatedAt          time.Time
	DeletedAt          *time.Time
}

type SchemaVersion struct {
	ID              string
	SchemaID        string
	Version         int
	Checksum        string
	Metadata        json.RawMessage
	ObjectCount     int
	ParentVersionID *string
	CreatedBy       string
	CreatedAt       time.Time
}

type SchemaObject struct {
	ID              string
	SchemaVersionID string
	ObjectType      string
	ObjectName      string
	ObjectSchema    string
	Definition      json.RawMessage
	ParentObjectID  *string
}

type SchemaRepository interface {
	Create(ctx context.Context, s *Schema) error
	GetByID(ctx context.Context, id string) (*Schema, error)
	ListByProjectID(ctx context.Context, projectID, cursor string, limit int32) ([]*Schema, string, int32, error)
	GetByConnectionAndSchema(ctx context.Context, connID, schemaName string) (*Schema, error)
	UpdateCurrentVersion(ctx context.Context, schemaID, versionID string) error

	CreateVersion(ctx context.Context, v *SchemaVersion) error
	GetVersionByID(ctx context.Context, id string) (*SchemaVersion, error)
	ListVersionsBySchemaID(ctx context.Context, schemaID, cursor string, limit int32) ([]*SchemaVersion, string, int32, error)
	GetLatestVersion(ctx context.Context, schemaID string) (*SchemaVersion, error)

	CreateObjects(ctx context.Context, objects []*SchemaObject) error
	ListObjectsByVersionID(ctx context.Context, versionID, objectType, cursor string, limit int32) ([]*SchemaObject, string, int32, error)
}

func ComputeChecksum(data json.RawMessage) string {
	h := sha256.Sum256(data)
	return fmt.Sprintf("%x", h)
}
