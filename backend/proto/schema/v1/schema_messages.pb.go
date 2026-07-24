package schemav1

import "fmt"

// ── Entities ──

type Schema struct {
	ID                 string `json:"id"`
	ProjectID          string `json:"project_id"`
	ConnectionID       string `json:"connection_id"`
	SchemaName         string `json:"schema_name"`
	CurrentVersionID   string `json:"current_version_id"`
	LastIntrospectedAt string `json:"last_introspected_at"`
	CreatedAt          string `json:"created_at"`
	UpdatedAt          string `json:"updated_at"`
}

type SchemaVersion struct {
	ID              string `json:"id"`
	SchemaID        string `json:"schema_id"`
	Version         int32  `json:"version"`
	Checksum        string `json:"checksum"`
	ObjectCount     int32  `json:"object_count"`
	ParentVersionID string `json:"parent_version_id"`
	CreatedBy       string `json:"created_by"`
	CreatedAt       string `json:"created_at"`
}

type SchemaObject struct {
	ID              string `json:"id"`
	SchemaVersionID string `json:"schema_version_id"`
	ObjectType      string `json:"object_type"`
	ObjectName      string `json:"object_name"`
	ObjectSchema    string `json:"object_schema"`
	Definition      string `json:"definition"`
	ParentObjectID  string `json:"parent_object_id"`
}

type DiagramNode struct {
	ID       string         `json:"id"`
	Type     string         `json:"type"`
	Position *DiagramPosition `json:"position"`
	Data     *DiagramNodeData `json:"data"`
}

type DiagramNodeData struct {
	Label     string          `json:"label"`
	Columns   []*ColumnInfo   `json:"columns"`
	ObjectType string         `json:"object_type"`
}

type ColumnInfo struct {
	Name     string `json:"name"`
	Type     string `json:"type"`
	Nullable bool   `json:"nullable"`
	Default  string `json:"default"`
	IsPK     bool   `json:"is_pk"`
	IsFK     bool   `json:"is_fk"`
}

type DiagramPosition struct {
	X float64 `json:"x"`
	Y float64 `json:"y"`
}

type DiagramEdge struct {
	ID           string `json:"id"`
	Source       string `json:"source"`
	Target       string `json:"target"`
	SourceHandle string `json:"source_handle"`
	TargetHandle string `json:"target_handle"`
	Label        string `json:"label"`
}

// ── Requests / Responses ──

type IntrospectSchemaRequest struct {
	ConnectionID string   `json:"connection_id"`
	SchemaNames  []string `json:"schema_names"`
}

type IntrospectSchemaResponse struct {
	Schema        *Schema        `json:"schema"`
	SchemaVersion *SchemaVersion `json:"schema_version"`
}

type GetSchemaRequest struct {
	ID string `json:"id"`
}

type GetSchemaResponse struct {
	Schema *Schema `json:"schema"`
}

type ListSchemasRequest struct {
	ProjectID string `json:"project_id"`
	Cursor    string `json:"cursor"`
	PageSize  int32  `json:"page_size"`
}

type ListSchemasResponse struct {
	Schemas    []*Schema `json:"schemas"`
	NextCursor string    `json:"next_cursor"`
	TotalCount int32     `json:"total_count"`
}

type ListSchemaVersionsRequest struct {
	SchemaID string `json:"schema_id"`
	Cursor   string `json:"cursor"`
	PageSize int32  `json:"page_size"`
}

type ListSchemaVersionsResponse struct {
	Versions   []*SchemaVersion `json:"versions"`
	NextCursor string           `json:"next_cursor"`
	TotalCount int32            `json:"total_count"`
}

type GetSchemaVersionRequest struct {
	ID string `json:"id"`
}

type GetSchemaVersionResponse struct {
	Version *SchemaVersion `json:"version"`
}

type CompareSchemaVersionsRequest struct {
	VersionAID string `json:"version_a_id"`
	VersionBID string `json:"version_b_id"`
}

type CompareSchemaVersionsResponse struct {
	Diff *SchemaDiff `json:"diff"`
}

type SchemaDiff struct {
	AddedObjects    []*DiffObject `json:"added_objects"`
	RemovedObjects  []*DiffObject `json:"removed_objects"`
	ModifiedObjects []*DiffObject `json:"modified_objects"`
}

type DiffObject struct {
	Type       string         `json:"type"`
	Name       string         `json:"name"`
	Definition string         `json:"definition"`
	Changes    []*FieldChange `json:"changes,omitempty"`
}

type FieldChange struct {
	Field  string `json:"field"`
	Before string `json:"before"`
	After  string `json:"after"`
}

type ListSchemaObjectsRequest struct {
	SchemaVersionID string `json:"schema_version_id"`
	ObjectType      string `json:"object_type"`
	Cursor          string `json:"cursor"`
	PageSize        int32  `json:"page_size"`
}

type ListSchemaObjectsResponse struct {
	Objects    []*SchemaObject `json:"objects"`
	NextCursor string          `json:"next_cursor"`
	TotalCount int32           `json:"total_count"`
}

type GetSchemaDiagramRequest struct {
	SchemaVersionID string `json:"schema_version_id"`
	IncludeDetails  bool   `json:"include_details"`
}

type GetSchemaDiagramResponse struct {
	Nodes []*DiagramNode `json:"nodes"`
	Edges []*DiagramEdge `json:"edges"`
}

func (r *IntrospectSchemaRequest) Validate() error {
	if r.ConnectionID == "" {
		return fmt.Errorf("connection id is required")
	}
	if len(r.SchemaNames) == 0 {
		r.SchemaNames = []string{"public"}
	}
	return nil
}
