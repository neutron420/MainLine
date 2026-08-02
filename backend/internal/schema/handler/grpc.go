package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/schemahub/backend/internal/pkg/errors"
	"github.com/schemahub/backend/internal/pkg/interceptor"
	"github.com/schemahub/backend/internal/schema/domain"
	schemav1 "github.com/schemahub/backend/proto/schema/v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type SchemaHandler struct {
	schemav1.UnimplementedSchemaServiceServer
	svc        *domain.SchemaService
	connString func(ctx context.Context, connID string) (string, error)
}

func NewSchemaHandler(svc *domain.SchemaService, connString func(ctx context.Context, connID string) (string, error)) *SchemaHandler {
	return &SchemaHandler{svc: svc, connString: connString}
}

func (h *SchemaHandler) IntrospectSchema(ctx context.Context, req *schemav1.IntrospectSchemaRequest) (*schemav1.IntrospectSchemaResponse, error) {
	userID, _ := interceptor.UserIDFromContext(ctx)

	connStr, err := h.connString(ctx, req.ConnectionId)
	if err != nil {
		return nil, status.Error(codes.FailedPrecondition, fmt.Sprintf("cannot connect: %v", err))
	}

	schema, version, err := h.svc.Introspect(ctx, connStr, req.ConnectionId, req.SchemaNames, userID)
	if err != nil {
		return nil, errors.ToGRPC(err)
	}
	return &schemav1.IntrospectSchemaResponse{
		Schema:        toProtoSchema(schema),
		SchemaVersion: toProtoVersion(version),
	}, nil
}

func (h *SchemaHandler) GetSchema(ctx context.Context, req *schemav1.GetSchemaRequest) (*schemav1.GetSchemaResponse, error) {
	s, err := h.svc.GetSchemaByID(ctx, req.Id)
	if err != nil {
		return nil, errors.ToGRPC(err)
	}
	return &schemav1.GetSchemaResponse{Schema: toProtoSchema(s)}, nil
}

func (h *SchemaHandler) ListSchemas(ctx context.Context, req *schemav1.ListSchemasRequest) (*schemav1.ListSchemasResponse, error) {
	schemas, next, total, err := h.svc.ListSchemas(ctx, req.ProjectId, req.Cursor, req.PageSize)
	if err != nil {
		return nil, errors.ToGRPC(err)
	}
	var ps []*schemav1.Schema
	for _, s := range schemas {
		ps = append(ps, toProtoSchema(s))
	}
	return &schemav1.ListSchemasResponse{Schemas: ps, NextCursor: next, TotalCount: total}, nil
}

func (h *SchemaHandler) ListSchemaVersions(ctx context.Context, req *schemav1.ListSchemaVersionsRequest) (*schemav1.ListSchemaVersionsResponse, error) {
	versions, next, total, err := h.svc.ListVersions(ctx, req.SchemaId, req.Cursor, req.PageSize)
	if err != nil {
		return nil, errors.ToGRPC(err)
	}
	var pv []*schemav1.SchemaVersion
	for _, v := range versions {
		pv = append(pv, toProtoVersion(v))
	}
	return &schemav1.ListSchemaVersionsResponse{Versions: pv, NextCursor: next, TotalCount: total}, nil
}

func (h *SchemaHandler) GetSchemaVersion(ctx context.Context, req *schemav1.GetSchemaVersionRequest) (*schemav1.GetSchemaVersionResponse, error) {
	v, err := h.svc.GetVersionByID(ctx, req.Id)
	if err != nil {
		return nil, errors.ToGRPC(err)
	}
	return &schemav1.GetSchemaVersionResponse{Version: toProtoVersion(v)}, nil
}

func (h *SchemaHandler) CompareSchemaVersions(ctx context.Context, req *schemav1.CompareSchemaVersionsRequest) (*schemav1.CompareSchemaVersionsResponse, error) {
	diff, err := h.svc.CompareVersions(ctx, req.VersionAId, req.VersionBId)
	if err != nil {
		return nil, errors.ToGRPC(err)
	}
	return &schemav1.CompareSchemaVersionsResponse{Diff: toProtoDiff(diff)}, nil
}

func (h *SchemaHandler) ListSchemaObjects(ctx context.Context, req *schemav1.ListSchemaObjectsRequest) (*schemav1.ListSchemaObjectsResponse, error) {
	objects, next, total, err := h.svc.ListObjects(ctx, req.SchemaVersionId, req.ObjectType, req.Cursor, req.PageSize)
	if err != nil {
		return nil, errors.ToGRPC(err)
	}
	var po []*schemav1.SchemaObject
	for _, o := range objects {
		po = append(po, toProtoObject(o))
	}
	return &schemav1.ListSchemaObjectsResponse{Objects: po, NextCursor: next, TotalCount: total}, nil
}

func (h *SchemaHandler) GetSchemaDiagram(ctx context.Context, req *schemav1.GetSchemaDiagramRequest) (*schemav1.GetSchemaDiagramResponse, error) {
	data, err := h.svc.GetDiagram(ctx, req.SchemaVersionId, req.IncludeDetails)
	if err != nil {
		return nil, errors.ToGRPC(err)
	}

	var nodes []*schemav1.DiagramNode
	for _, n := range data.Nodes {
		var cols []*schemav1.ColumnInfo
		for _, c := range n.Data.Columns {
			cols = append(cols, &schemav1.ColumnInfo{
				Name: c.Name, Type: c.DataType, Nullable: c.IsNullable,
			})
		}
		nodes = append(nodes, &schemav1.DiagramNode{
			Id:       n.ID,
			Type:     n.Type,
			Position: &schemav1.DiagramPosition{X: n.Position.X, Y: n.Position.Y},
			Data:     &schemav1.DiagramNodeData{Label: n.Data.Label, Columns: cols},
		})
	}

	var edges []*schemav1.DiagramEdge
	for _, e := range data.Edges {
		edges = append(edges, &schemav1.DiagramEdge{
			Id: e.ID, Source: e.Source, Target: e.Target, Label: e.Label,
		})
	}

	return &schemav1.GetSchemaDiagramResponse{Nodes: nodes, Edges: edges}, nil
}

func toProtoSchema(s *domain.Schema) *schemav1.Schema {
	ps := &schemav1.Schema{
		Id: s.ID, ProjectId: s.ProjectID, ConnectionId: s.ConnectionID,
		SchemaName: s.SchemaName, CreatedAt: s.CreatedAt.Format(time.RFC3339),
		UpdatedAt: s.UpdatedAt.Format(time.RFC3339),
	}
	if s.CurrentVersionID != nil {
		ps.CurrentVersionId = *s.CurrentVersionID
	}
	if s.LastIntrospectedAt != nil {
		ps.LastIntrospectedAt = s.LastIntrospectedAt.Format(time.RFC3339)
	}
	return ps
}

func toProtoVersion(v *domain.SchemaVersion) *schemav1.SchemaVersion {
	pv := &schemav1.SchemaVersion{
		Id: v.ID, SchemaId: v.SchemaID, Version: int32(v.Version),
		Checksum: v.Checksum, ObjectCount: int32(v.ObjectCount),
		CreatedBy: v.CreatedBy, CreatedAt: v.CreatedAt.Format(time.RFC3339),
	}
	if v.ParentVersionID != nil {
		pv.ParentVersionId = *v.ParentVersionID
	}
	return pv
}

func toProtoObject(o *domain.SchemaObject) *schemav1.SchemaObject {
	po := &schemav1.SchemaObject{
		Id: o.ID, SchemaVersionId: o.SchemaVersionID,
		ObjectType: o.ObjectType, ObjectName: o.ObjectName, ObjectSchema: o.ObjectSchema,
		Definition: string(o.Definition),
	}
	if o.ParentObjectID != nil {
		po.ParentObjectId = *o.ParentObjectID
	}
	return po
}

func toProtoDiff(d *domain.DiffResult) *schemav1.SchemaDiff {
	return &schemav1.SchemaDiff{
		AddedObjects:    toDiffObjects(d.AddedObjects),
		RemovedObjects:  toDiffObjects(d.RemovedObjects),
		ModifiedObjects: toDiffObjects(d.ModifiedObjects),
	}
}

func toDiffObjects(in []domain.DiffObject) []*schemav1.DiffObject {
	var out []*schemav1.DiffObject
	for _, o := range in {
		d := &schemav1.DiffObject{Type: o.Type, Name: o.Name, Definition: string(o.Definition)}
		for _, c := range o.Changes {
			d.Changes = append(d.Changes, &schemav1.FieldChange{
				Field:  c.Field,
				Before: fmt.Sprintf("%v", c.Before),
				After:  fmt.Sprintf("%v", c.After),
			})
		}
		out = append(out, d)
	}
	return out
}

var _ = json.Marshal
