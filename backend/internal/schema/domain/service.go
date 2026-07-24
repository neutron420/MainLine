package domain

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/schemahub/backend/pkg/encryption"
)

type SchemaService struct {
	repo       SchemaRepository
	introspect *IntrospectionService
	differ     *Differ
}

func NewSchemaService(repo SchemaRepository) *SchemaService {
	return &SchemaService{
		repo:       repo,
		introspect: NewIntrospectionService(),
		differ:     NewDiffer(),
	}
}

func (s *SchemaService) Introspect(ctx context.Context, connStr, connID string, schemaNames []string, userID string) (*Schema, *SchemaVersion, error) {
	meta, err := s.introspect.Introspect(ctx, connStr, schemaNames)
	if err != nil {
		return nil, nil, fmt.Errorf("introspecting: %w", err)
	}

	schemaName := "public"
	if len(schemaNames) > 0 {
		schemaName = schemaNames[0]
	}

	schema, err := s.repo.GetByConnectionAndSchema(ctx, connID, schemaName)
	if err != nil {
		schema = &Schema{
			ProjectID:    "",
			ConnectionID: connID,
			SchemaName:   schemaName,
		}
		if err := s.repo.Create(ctx, schema); err != nil {
			return nil, nil, fmt.Errorf("creating schema: %w", err)
		}
	}

	checksum := ComputeChecksum(meta)

	latest, _ := s.repo.GetLatestVersion(ctx, schema.ID)
	if latest != nil && latest.Checksum == checksum {
		return schema, latest, nil
	}

	var parentID *string
	if latest != nil {
		parentID = &latest.ID
	}

	var objCount int
	var metadata SchemaMetadata
	if err := json.Unmarshal(meta, &metadata); err == nil {
		objCount = len(metadata.Tables) + len(metadata.Enums) + len(metadata.Extensions)
	}

	version := &SchemaVersion{
		SchemaID:    schema.ID,
		Version:     1,
		Checksum:    checksum,
		Metadata:    meta,
		ObjectCount: objCount,
		CreatedBy:   userID,
	}

	if latest != nil {
		version.Version = latest.Version + 1
		version.ParentVersionID = parentID
	}

	if err := s.repo.CreateVersion(ctx, version); err != nil {
		return nil, nil, fmt.Errorf("creating version: %w", err)
	}

	if err := s.repo.UpdateCurrentVersion(ctx, schema.ID, version.ID); err != nil {
		return nil, nil, fmt.Errorf("updating current version: %w", err)
	}

	var objects []*SchemaObject
	for _, tbl := range metadata.Tables {
		def, _ := json.Marshal(tbl)
		objects = append(objects, &SchemaObject{
			SchemaVersionID: version.ID,
			ObjectType:      "table",
			ObjectName:      tbl.Name,
			ObjectSchema:    tbl.Schema,
			Definition:      def,
		})
	}
	for _, en := range metadata.Enums {
		def, _ := json.Marshal(en)
		objects = append(objects, &SchemaObject{
			SchemaVersionID: version.ID,
			ObjectType:      "enum",
			ObjectName:      en.Name,
			ObjectSchema:    "public",
			Definition:      def,
		})
	}
	for _, ext := range metadata.Extensions {
		def, _ := json.Marshal(ext)
		objects = append(objects, &SchemaObject{
			SchemaVersionID: version.ID,
			ObjectType:      "extension",
			ObjectName:      ext.Name,
			ObjectSchema:    "public",
			Definition:      def,
		})
	}

	if len(objects) > 0 {
		if err := s.repo.CreateObjects(ctx, objects); err != nil {
			return nil, nil, fmt.Errorf("creating objects: %w", err)
		}
	}

	return schema, version, nil
}

func (s *SchemaService) GetSchemaByID(ctx context.Context, id string) (*Schema, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *SchemaService) ListSchemas(ctx context.Context, projectID, cursor string, pageSize int32) ([]*Schema, string, int32, error) {
	if pageSize <= 0 || pageSize > 100 {
		pageSize = 20
	}
	return s.repo.ListByProjectID(ctx, projectID, cursor, pageSize)
}

func (s *SchemaService) ListVersions(ctx context.Context, schemaID, cursor string, pageSize int32) ([]*SchemaVersion, string, int32, error) {
	if pageSize <= 0 || pageSize > 100 {
		pageSize = 20
	}
	return s.repo.ListVersionsBySchemaID(ctx, schemaID, cursor, pageSize)
}

func (s *SchemaService) GetVersionByID(ctx context.Context, id string) (*SchemaVersion, error) {
	return s.repo.GetVersionByID(ctx, id)
}

func (s *SchemaService) CompareVersions(ctx context.Context, versionAID, versionBID string) (*DiffResult, error) {
	vA, err := s.repo.GetVersionByID(ctx, versionAID)
	if err != nil {
		return nil, fmt.Errorf("version A not found: %w", err)
	}
	vB, err := s.repo.GetVersionByID(ctx, versionBID)
	if err != nil {
		return nil, fmt.Errorf("version B not found: %w", err)
	}

	return s.differ.Diff(vA.Metadata, vB.Metadata)
}

func (s *SchemaService) ListObjects(ctx context.Context, versionID, objectType, cursor string, pageSize int32) ([]*SchemaObject, string, int32, error) {
	if pageSize <= 0 || pageSize > 100 {
		pageSize = 20
	}
	return s.repo.ListObjectsByVersionID(ctx, versionID, objectType, cursor, pageSize)
}

type DiagramData struct {
	Nodes []DiagramNode `json:"nodes"`
	Edges []DiagramEdge `json:"edges"`
}

type DiagramNode struct {
	ID       string          `json:"id"`
	Type     string          `json:"type"`
	Position DiagramPosition `json:"position"`
	Data     DiagramNodeData `json:"data"`
}

type DiagramPosition struct {
	X float64 `json:"x"`
	Y float64 `json:"y"`
}

type DiagramNodeData struct {
	Label    string       `json:"label"`
	Columns  []ColumnInfo `json:"columns,omitempty"`
	TableDef TableInfo    `json:"-"`
}

type DiagramEdge struct {
	ID           string `json:"id"`
	Source       string `json:"source"`
	Target       string `json:"target"`
	SourceHandle string `json:"source_handle,omitempty"`
	TargetHandle string `json:"target_handle,omitempty"`
	Label        string `json:"label,omitempty"`
}

func (s *SchemaService) GetDiagram(ctx context.Context, versionID string, includeDetails bool) (*DiagramData, error) {
	v, err := s.repo.GetVersionByID(ctx, versionID)
	if err != nil {
		return nil, fmt.Errorf("version not found: %w", err)
	}

	var meta SchemaMetadata
	if err := json.Unmarshal(v.Metadata, &meta); err != nil {
		return nil, fmt.Errorf("unmarshaling metadata: %w", err)
	}

	objects, _, _, err := s.repo.ListObjectsByVersionID(ctx, versionID, "", "", 1000)
	if err != nil {
		return nil, fmt.Errorf("listing objects: %w", err)
	}
	_ = objects

	diagram := &DiagramData{
		Nodes: []DiagramNode{},
		Edges: []DiagramEdge{},
	}

	x := 0.0
	y := 0.0
	edgeID := 0

	sort.Slice(meta.Tables, func(i, j int) bool {
		return meta.Tables[i].Name < meta.Tables[j].Name
	})

	for _, tbl := range meta.Tables {
		nodeID := tbl.Schema + "." + tbl.Name
		columns := tbl.Columns
		if !includeDetails {
			columns = nil
		}

		diagram.Nodes = append(diagram.Nodes, DiagramNode{
			ID:   nodeID,
			Type: "table",
			Position: DiagramPosition{
				X: x,
				Y: y,
			},
			Data: DiagramNodeData{
				Label:   tbl.Name,
				Columns: columns,
			},
		})

		y += 200
		if y > 800 {
			y = 0
			x += 300
		}

		for _, fk := range tbl.Constr.ForeignKeys {
			targetID := tbl.Schema + "." + fk.RefTable
			edgeID++
			diagram.Edges = append(diagram.Edges, DiagramEdge{
				ID:     fmt.Sprintf("e-%d", edgeID),
				Source: nodeID,
				Target: targetID,
				Label:  fk.Name,
			})
		}
	}

	return diagram, nil
}

var _ = uuid.NewString
var _ = pgxpool.Pool{}
var _ = encryption.Encrypt
var _ = time.Now
