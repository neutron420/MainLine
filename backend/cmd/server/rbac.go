package main

import (
	"context"
	"fmt"

	auditRepo "github.com/schemahub/backend/internal/audit/repository/postgres"
	driftRepo "github.com/schemahub/backend/internal/drift/repository/postgres"
	migrationRepo "github.com/schemahub/backend/internal/migration/repository/postgres"
	projectRepo "github.com/schemahub/backend/internal/project/repository/postgres"
	schemaRepo "github.com/schemahub/backend/internal/schema/repository/postgres"
	auditv1 "github.com/schemahub/backend/proto/audit/v1"
	driftv1 "github.com/schemahub/backend/proto/drift/v1"
	eventv1 "github.com/schemahub/backend/proto/event/v1"
	migrationv1 "github.com/schemahub/backend/proto/migration/v1"
	projectv1 "github.com/schemahub/backend/proto/project/v1"
	schemav1 "github.com/schemahub/backend/proto/schema/v1"
)

// rbacEnforcer resolves the project scope of every request and verifies the
// caller is a member. Fine-grained role rules (owner-only deletes, etc.) remain
// enforced in the domain services.
type rbacEnforcer struct {
	projects    *projectRepo.ProjectRepository
	connections *projectRepo.ConnectionRepository
	schemas     *schemaRepo.SchemaRepository
	migrations  *migrationRepo.MigrationRepository
	drifts      *driftRepo.DriftRepository
	audits      *auditRepo.AuditRepository
}

// scoper returns the project ID a request is scoped to, or "" when the request
// has no project scope (allowed for any authenticated user).
type scoper func(ctx context.Context, e *rbacEnforcer, userID string, req any) (string, error)

var rbacScopes = map[string]scoper{
	// ── ProjectService ──
	"/schemahub.project.v1.ProjectService/GetProject":       byID,
	"/schemahub.project.v1.ProjectService/UpdateProject":    byID,
	"/schemahub.project.v1.ProjectService/DeleteProject":    byID,
	"/schemahub.project.v1.ProjectService/AddMember":        byProjectIDField,
	"/schemahub.project.v1.ProjectService/RemoveMember":     byProjectIDField,
	"/schemahub.project.v1.ProjectService/UpdateMemberRole": byProjectIDField,
	"/schemahub.project.v1.ProjectService/ListMembers":      byProjectIDField,
	"/schemahub.project.v1.ProjectService/CreateConnection": byProjectIDField,
	"/schemahub.project.v1.ProjectService/GetConnection":    byConnectionID,
	"/schemahub.project.v1.ProjectService/ListConnections":  byProjectIDField,
	"/schemahub.project.v1.ProjectService/UpdateConnection": byConnectionID,
	"/schemahub.project.v1.ProjectService/DeleteConnection": byConnectionID,
	"/schemahub.project.v1.ProjectService/TestConnection":   byConnectionIDFromReq,
	// ── SchemaService ──
	"/schemahub.schema.v1.SchemaService/IntrospectSchema":      byConnectionIDFromReq,
	"/schemahub.schema.v1.SchemaService/GetSchema":             bySchemaID,
	"/schemahub.schema.v1.SchemaService/ListSchemas":           byProjectIDField,
	"/schemahub.schema.v1.SchemaService/ListSchemaVersions":    bySchemaIDFromReq,
	"/schemahub.schema.v1.SchemaService/GetSchemaVersion":      byVersionID,
	"/schemahub.schema.v1.SchemaService/CompareSchemaVersions": byVersionAID,
	"/schemahub.schema.v1.SchemaService/ListSchemaObjects":     bySchemaVersionID,
	"/schemahub.schema.v1.SchemaService/GetSchemaDiagram":      bySchemaVersionID,
	// ── MigrationService ──
	"/schemahub.migration.v1.MigrationService/CreateMigration":   byProjectIDField,
	"/schemahub.migration.v1.MigrationService/GetMigration":      byMigrationID,
	"/schemahub.migration.v1.MigrationService/ListMigrations":    byProjectIDField,
	"/schemahub.migration.v1.MigrationService/UpdateMigration":   byMigrationID,
	"/schemahub.migration.v1.MigrationService/DeleteMigration":   byMigrationID,
	"/schemahub.migration.v1.MigrationService/ExecuteMigration":  byMigrationIDFromReq,
	"/schemahub.migration.v1.MigrationService/WatchMigration":    byRunIDFromReq,
	"/schemahub.migration.v1.MigrationService/RollbackMigration": byRunIDFromReq,
	"/schemahub.migration.v1.MigrationService/WatchRollback":     byRunIDFromReq,
	"/schemahub.migration.v1.MigrationService/DryRunMigration":   byMigrationIDFromReq,
	"/schemahub.migration.v1.MigrationService/GetMigrationRun":   byRunID,
	"/schemahub.migration.v1.MigrationService/ListMigrationRuns": byMigrationIDFromReq,
	"/schemahub.migration.v1.MigrationService/GetMigrationLogs":  byRunIDFromReq,
	// ── EventService ──
	"/schemahub.event.v1.EventService/Subscribe":        allProjectIDs,
	"/schemahub.event.v1.EventService/AcknowledgeEvent": byEventID,
	"/schemahub.event.v1.EventService/Heartbeat":        allProjectIDs,
	// ── AuditService ──
	"/schemahub.audit.v1.AuditService/ListAuditEntries": byAuditResource,
	"/schemahub.audit.v1.AuditService/GetAuditEntry":    byAuditEntryID,
	// ── DriftService ──
	"/schemahub.drift.v1.DriftService/CheckDrift":        byConnectionIDFromReq,
	"/schemahub.drift.v1.DriftService/ListDriftEvents":   byConnectionIDFromReq,
	"/schemahub.drift.v1.DriftService/GetDriftEvent":     byDriftEventID,
	"/schemahub.drift.v1.DriftService/ResolveDriftEvent": byDriftEventID,
	"/schemahub.drift.v1.DriftService/GetDriftStats":     byConnectionIDFromReq,
}

func (e *rbacEnforcer) enforce(ctx context.Context, userID, role, fullMethod string, req any) error {
	if role == "admin" {
		return nil
	}

	scope, ok := rbacScopes[fullMethod]
	if !ok {
		return nil
	}

	projectID, err := scope(ctx, e, userID, req)
	if err != nil {
		return fmt.Errorf("permission denied: %w", err)
	}
	if projectID == "" {
		return nil
	}

	_, err = e.projects.GetMember(ctx, projectID, userID)
	if err != nil {
		return fmt.Errorf("permission denied: not a member of this project")
	}
	return nil
}

func (e *rbacEnforcer) enforceAll(ctx context.Context, userID, role string, projectIDs []string) error {
	if role == "admin" {
		return nil
	}
	for _, pid := range projectIDs {
		if pid == "" {
			continue
		}
		if _, err := e.projects.GetMember(ctx, pid, userID); err != nil {
			return fmt.Errorf("permission denied: not a member of project %s", pid)
		}
	}
	return nil
}

// ── direct project ID extractors ──

func byID(ctx context.Context, e *rbacEnforcer, userID string, req any) (string, error) {
	switch r := req.(type) {
	case *projectv1.GetProjectRequest:
		return r.Id, nil
	case *projectv1.UpdateProjectRequest:
		return r.Id, nil
	case *projectv1.DeleteProjectRequest:
		return r.Id, nil
	default:
		return "", fmt.Errorf("unsupported request type %T", req)
	}
}

func byProjectIDField(ctx context.Context, e *rbacEnforcer, userID string, req any) (string, error) {
	switch r := req.(type) {
	case *projectv1.AddMemberRequest:
		return r.ProjectId, nil
	case *projectv1.RemoveMemberRequest:
		return r.ProjectId, nil
	case *projectv1.UpdateMemberRoleRequest:
		return r.ProjectId, nil
	case *projectv1.ListMembersRequest:
		return r.ProjectId, nil
	case *projectv1.CreateConnectionRequest:
		return r.ProjectId, nil
	case *projectv1.ListConnectionsRequest:
		return r.ProjectId, nil
	case *schemav1.ListSchemasRequest:
		return r.ProjectId, nil
	case *migrationv1.CreateMigrationRequest:
		return r.ProjectId, nil
	case *migrationv1.ListMigrationsRequest:
		return r.ProjectId, nil
	default:
		return "", fmt.Errorf("unsupported request type %T", req)
	}
}

// ── ID-based resolvers (ID → project) ──

func byConnectionID(ctx context.Context, e *rbacEnforcer, userID string, req any) (string, error) {
	var id string
	switch r := req.(type) {
	case *projectv1.GetConnectionRequest:
		id = r.Id
	case *projectv1.UpdateConnectionRequest:
		id = r.Id
	case *projectv1.DeleteConnectionRequest:
		id = r.Id
	default:
		return "", fmt.Errorf("unsupported request type %T", req)
	}
	c, err := e.connections.GetByID(ctx, id)
	if err != nil {
		return "", fmt.Errorf("resolving connection: %w", err)
	}
	return c.ProjectID, nil
}

func byConnectionIDFromReq(ctx context.Context, e *rbacEnforcer, userID string, req any) (string, error) {
	var id string
	switch r := req.(type) {
	case *projectv1.TestConnectionRequest:
		id = r.ConnectionId
	case *schemav1.IntrospectSchemaRequest:
		id = r.ConnectionId
	case *driftv1.CheckDriftRequest:
		id = r.ConnectionId
	case *driftv1.ListDriftEventsRequest:
		id = r.ConnectionId
	case *driftv1.GetDriftStatsRequest:
		id = r.ConnectionId
	default:
		return "", fmt.Errorf("unsupported request type %T", req)
	}
	return e.resolveConnectionProject(ctx, id)
}

func (e *rbacEnforcer) resolveConnectionProject(ctx context.Context, id string) (string, error) {
	if id == "" {
		return "", nil
	}
	c, err := e.connections.GetByID(ctx, id)
	if err != nil {
		return "", fmt.Errorf("resolving connection: %w", err)
	}
	return c.ProjectID, nil
}

func bySchemaID(ctx context.Context, e *rbacEnforcer, userID string, req any) (string, error) {
	var id string
	switch r := req.(type) {
	case *schemav1.GetSchemaRequest:
		id = r.Id
	default:
		return "", fmt.Errorf("unsupported request type %T", req)
	}
	s, err := e.schemas.GetByID(ctx, id)
	if err != nil {
		return "", fmt.Errorf("resolving schema: %w", err)
	}
	return s.ProjectID, nil
}

func bySchemaIDFromReq(ctx context.Context, e *rbacEnforcer, userID string, req any) (string, error) {
	var id string
	switch r := req.(type) {
	case *schemav1.ListSchemaVersionsRequest:
		id = r.SchemaId
	default:
		return "", fmt.Errorf("unsupported request type %T", req)
	}
	return e.resolveSchemaProject(ctx, id)
}

func (e *rbacEnforcer) resolveSchemaProject(ctx context.Context, id string) (string, error) {
	if id == "" {
		return "", nil
	}
	s, err := e.schemas.GetByID(ctx, id)
	if err != nil {
		return "", fmt.Errorf("resolving schema: %w", err)
	}
	return s.ProjectID, nil
}

func byVersionID(ctx context.Context, e *rbacEnforcer, userID string, req any) (string, error) {
	var id string
	switch r := req.(type) {
	case *schemav1.GetSchemaVersionRequest:
		id = r.Id
	default:
		return "", fmt.Errorf("unsupported request type %T", req)
	}
	return e.resolveVersionProject(ctx, id)
}

func byVersionAID(ctx context.Context, e *rbacEnforcer, userID string, req any) (string, error) {
	var id string
	switch r := req.(type) {
	case *schemav1.CompareSchemaVersionsRequest:
		id = r.VersionAId
	default:
		return "", fmt.Errorf("unsupported request type %T", req)
	}
	return e.resolveVersionProject(ctx, id)
}

func bySchemaVersionID(ctx context.Context, e *rbacEnforcer, userID string, req any) (string, error) {
	var id string
	switch r := req.(type) {
	case *schemav1.ListSchemaObjectsRequest:
		id = r.SchemaVersionId
	case *schemav1.GetSchemaDiagramRequest:
		id = r.SchemaVersionId
	default:
		return "", fmt.Errorf("unsupported request type %T", req)
	}
	return e.resolveVersionProject(ctx, id)
}

func (e *rbacEnforcer) resolveVersionProject(ctx context.Context, id string) (string, error) {
	if id == "" {
		return "", nil
	}
	v, err := e.schemas.GetVersionByID(ctx, id)
	if err != nil {
		return "", fmt.Errorf("resolving version: %w", err)
	}
	return e.resolveSchemaProject(ctx, v.SchemaID)
}

func byMigrationID(ctx context.Context, e *rbacEnforcer, userID string, req any) (string, error) {
	var id string
	switch r := req.(type) {
	case *migrationv1.GetMigrationRequest:
		id = r.Id
	case *migrationv1.UpdateMigrationRequest:
		id = r.Id
	case *migrationv1.DeleteMigrationRequest:
		id = r.Id
	default:
		return "", fmt.Errorf("unsupported request type %T", req)
	}
	return e.resolveMigrationProject(ctx, id)
}

func byMigrationIDFromReq(ctx context.Context, e *rbacEnforcer, userID string, req any) (string, error) {
	var id string
	switch r := req.(type) {
	case *migrationv1.ExecuteMigrationRequest:
		id = r.MigrationId
	case *migrationv1.DryRunMigrationRequest:
		id = r.MigrationId
	case *migrationv1.ListMigrationRunsRequest:
		id = r.MigrationId
	default:
		return "", fmt.Errorf("unsupported request type %T", req)
	}
	return e.resolveMigrationProject(ctx, id)
}

func (e *rbacEnforcer) resolveMigrationProject(ctx context.Context, id string) (string, error) {
	if id == "" {
		return "", nil
	}
	m, err := e.migrations.GetByID(ctx, id)
	if err != nil {
		return "", fmt.Errorf("resolving migration: %w", err)
	}
	return m.ProjectID, nil
}

func byRunID(ctx context.Context, e *rbacEnforcer, userID string, req any) (string, error) {
	var id string
	switch r := req.(type) {
	case *migrationv1.GetMigrationRunRequest:
		id = r.Id
	default:
		return "", fmt.Errorf("unsupported request type %T", req)
	}
	return e.resolveRunProject(ctx, id)
}

func byRunIDFromReq(ctx context.Context, e *rbacEnforcer, userID string, req any) (string, error) {
	var id string
	switch r := req.(type) {
	case *migrationv1.WatchMigrationRequest:
		id = r.RunId
	case *migrationv1.RollbackMigrationRequest:
		id = r.RunId
	case *migrationv1.WatchRollbackRequest:
		id = r.RunId
	case *migrationv1.GetMigrationLogsRequest:
		id = r.RunId
	default:
		return "", fmt.Errorf("unsupported request type %T", req)
	}
	return e.resolveRunProject(ctx, id)
}

func (e *rbacEnforcer) resolveRunProject(ctx context.Context, id string) (string, error) {
	if id == "" {
		return "", nil
	}
	run, err := e.migrations.GetRunByID(ctx, id)
	if err != nil {
		return "", fmt.Errorf("resolving run: %w", err)
	}
	return e.resolveMigrationProject(ctx, run.MigrationID)
}

func byEventID(ctx context.Context, e *rbacEnforcer, userID string, req any) (string, error) {
	var id string
	switch r := req.(type) {
	case *eventv1.AcknowledgeEventRequest:
		id = r.EventId
	default:
		return "", fmt.Errorf("unsupported request type %T", req)
	}
	return e.resolveAuditEntryProject(ctx, id)
}

func byDriftEventID(ctx context.Context, e *rbacEnforcer, userID string, req any) (string, error) {
	var id string
	switch r := req.(type) {
	case *driftv1.GetDriftEventRequest:
		id = r.Id
	case *driftv1.ResolveDriftEventRequest:
		id = r.Id
	default:
		return "", fmt.Errorf("unsupported request type %T", req)
	}
	return e.resolveDriftEventProject(ctx, id)
}

func (e *rbacEnforcer) resolveDriftEventProject(ctx context.Context, id string) (string, error) {
	if id == "" {
		return "", nil
	}
	evt, err := e.drifts.GetByID(ctx, id)
	if err != nil {
		return "", fmt.Errorf("resolving drift event: %w", err)
	}
	return e.resolveSchemaProject(ctx, evt.SchemaID)
}

func byAuditEntryID(ctx context.Context, e *rbacEnforcer, userID string, req any) (string, error) {
	var id string
	switch r := req.(type) {
	case *auditv1.GetAuditEntryRequest:
		id = r.Id
	default:
		return "", fmt.Errorf("unsupported request type %T", req)
	}
	return e.resolveAuditEntryProject(ctx, id)
}

func (e *rbacEnforcer) resolveAuditEntryProject(ctx context.Context, id string) (string, error) {
	if id == "" {
		return "", nil
	}
	entry, err := e.audits.GetByID(ctx, id)
	if err != nil {
		return "", fmt.Errorf("resolving audit entry: %w", err)
	}
	return e.resolveAuditResourceProject(ctx, entry.ResourceType, entry.ResourceID)
}

func byAuditResource(ctx context.Context, e *rbacEnforcer, userID string, req any) (string, error) {
	var resourceType, resourceID string
	switch r := req.(type) {
	case *auditv1.ListAuditEntriesRequest:
		resourceType, resourceID = r.ResourceType, r.ResourceId
	default:
		return "", fmt.Errorf("unsupported request type %T", req)
	}
	return e.resolveAuditResourceProject(ctx, resourceType, resourceID)
}

// resolveAuditResourceProject maps audit resource scopes to projects. Unknown
// or global resource types (e.g. "user", "auth", "") carry no project scope and
// are allowed for any authenticated caller.
func (e *rbacEnforcer) resolveAuditResourceProject(ctx context.Context, resourceType, resourceID string) (string, error) {
	switch resourceType {
	case "project", "member":
		return resourceID, nil
	case "connection":
		return e.resolveConnectionProject(ctx, resourceID)
	case "schema":
		return e.resolveSchemaProject(ctx, resourceID)
	case "migration":
		return e.resolveMigrationProject(ctx, resourceID)
	default:
		return "", nil
	}
}

// allProjectIDs enforces membership for every project in a list (Subscribe,
// Heartbeat). Empty lists are allowed.
func allProjectIDs(ctx context.Context, e *rbacEnforcer, userID string, req any) (string, error) {
	var ids []string
	switch r := req.(type) {
	case *eventv1.SubscribeRequest:
		ids = r.ProjectIds
	case *eventv1.HeartbeatRequest:
		ids = r.ProjectIds
	default:
		return "", fmt.Errorf("unsupported request type %T", req)
	}
	for _, pid := range ids {
		if pid == "" {
			continue
		}
		if _, err := e.projects.GetMember(ctx, pid, userID); err != nil {
			return "", fmt.Errorf("permission denied: not a member of project %s", pid)
		}
	}
	return "", nil
}
