package handler

import (
	"context"
	"fmt"
	"time"

	"github.com/schemahub/backend/internal/migration/domain"
	"github.com/schemahub/backend/internal/pkg/errors"
	"github.com/schemahub/backend/internal/pkg/interceptor"
	migrationv1 "github.com/schemahub/backend/proto/migration/v1"
)

type MigrationHandler struct {
	migrationv1.UnimplementedMigrationServiceServer
	svc *domain.MigrationService
}

func NewMigrationHandler(svc *domain.MigrationService) *MigrationHandler {
	return &MigrationHandler{svc: svc}
}

func (h *MigrationHandler) CreateMigration(ctx context.Context, req *migrationv1.CreateMigrationRequest) (*migrationv1.CreateMigrationResponse, error) {
	userID, _ := interceptor.UserIDFromContext(ctx)

	m := &domain.Migration{
		ProjectID:   req.ProjectId,
		Title:       req.Title,
		Version:     req.Version,
		UpSQL:       req.UpSql,
		DownSQL:     req.DownSql,
		Description: req.Description,
		CreatedBy:   userID,
	}

	created, err := h.svc.Create(ctx, m)
	if err != nil {
		return nil, errors.ToGRPC(err)
	}

	return &migrationv1.CreateMigrationResponse{Migration: toProtoMigration(created)}, nil
}

func (h *MigrationHandler) GetMigration(ctx context.Context, req *migrationv1.GetMigrationRequest) (*migrationv1.GetMigrationResponse, error) {
	m, err := h.svc.GetByID(ctx, req.Id)
	if err != nil {
		return nil, errors.ToGRPC(err)
	}
	return &migrationv1.GetMigrationResponse{Migration: toProtoMigration(m)}, nil
}

func (h *MigrationHandler) ListMigrations(ctx context.Context, req *migrationv1.ListMigrationsRequest) (*migrationv1.ListMigrationsResponse, error) {
	migrations, next, total, err := h.svc.ListByProject(ctx, req.ProjectId, req.Cursor, req.PageSize)
	if err != nil {
		return nil, errors.ToGRPC(err)
	}

	var pm []*migrationv1.Migration
	for _, m := range migrations {
		pm = append(pm, toProtoMigration(m))
	}
	return &migrationv1.ListMigrationsResponse{Migrations: pm, NextCursor: next, TotalCount: total}, nil
}

func (h *MigrationHandler) UpdateMigration(ctx context.Context, req *migrationv1.UpdateMigrationRequest) (*migrationv1.UpdateMigrationResponse, error) {
	dm := &domain.Migration{ID: req.Id}
	if req.Title != nil {
		dm.Title = *req.Title
	}
	if req.Description != nil {
		dm.Description = *req.Description
	}
	if req.UpSql != nil {
		dm.UpSQL = *req.UpSql
	}
	if req.DownSql != nil {
		dm.DownSQL = *req.DownSql
	}
	if req.Status != nil {
		dm.Status = domain.MigrationStatus(*req.Status)
	}

	updated, err := h.svc.Update(ctx, dm)
	if err != nil {
		return nil, errors.ToGRPC(err)
	}
	return &migrationv1.UpdateMigrationResponse{Migration: toProtoMigration(updated)}, nil
}

func (h *MigrationHandler) DeleteMigration(ctx context.Context, req *migrationv1.DeleteMigrationRequest) (*migrationv1.DeleteMigrationResponse, error) {
	if err := h.svc.Delete(ctx, req.Id); err != nil {
		return nil, errors.ToGRPC(err)
	}
	return &migrationv1.DeleteMigrationResponse{}, nil
}

func (h *MigrationHandler) ExecuteMigration(ctx context.Context, req *migrationv1.ExecuteMigrationRequest) (*migrationv1.ExecuteMigrationResponse, error) {
	userID, _ := interceptor.UserIDFromContext(ctx)

	run, err := h.svc.Execute(ctx, req.MigrationId, req.ConnectionId, userID)
	if err != nil {
		return nil, errors.ToGRPC(err)
	}
	return &migrationv1.ExecuteMigrationResponse{Run: toProtoRun(run)}, nil
}

func (h *MigrationHandler) WatchMigration(req *migrationv1.WatchMigrationRequest, stream migrationv1.MigrationService_WatchMigrationServer) error {
	ch := h.svc.Subscribe(req.RunId)
	defer h.svc.Unsubscribe(req.RunId, ch)

	for msg := range ch {
		if err := stream.Send(toProtoStatus(msg)); err != nil {
			return err
		}
		if msg.State == domain.RunStatusCompleted || msg.State == domain.RunStatusFailed {
			break
		}
	}
	return nil
}

func (h *MigrationHandler) RollbackMigration(ctx context.Context, req *migrationv1.RollbackMigrationRequest) (*migrationv1.RollbackMigrationResponse, error) {
	userID, _ := interceptor.UserIDFromContext(ctx)

	run, err := h.svc.Rollback(ctx, req.RunId, userID)
	if err != nil {
		return nil, errors.ToGRPC(err)
	}
	return &migrationv1.RollbackMigrationResponse{Run: toProtoRun(run)}, nil
}

func (h *MigrationHandler) WatchRollback(req *migrationv1.WatchRollbackRequest, stream migrationv1.MigrationService_WatchRollbackServer) error {
	ch := h.svc.Subscribe(req.RunId)
	defer h.svc.Unsubscribe(req.RunId, ch)

	for msg := range ch {
		if err := stream.Send(toProtoStatus(msg)); err != nil {
			return err
		}
		if msg.State == domain.RunStatusCompleted || msg.State == domain.RunStatusFailed {
			break
		}
	}
	return nil
}

func (h *MigrationHandler) ValidateMigration(ctx context.Context, req *migrationv1.ValidateMigrationRequest) (*migrationv1.ValidateMigrationResponse, error) {
	valid, errs := h.svc.Validate(ctx, req.UpSql, req.DownSql)
	return &migrationv1.ValidateMigrationResponse{Valid: valid, Errors: errs}, nil
}

func (h *MigrationHandler) DryRunMigration(ctx context.Context, req *migrationv1.DryRunMigrationRequest) (*migrationv1.DryRunMigrationResponse, error) {
	valid, errs, warnings := h.svc.DryRun(ctx, req.MigrationId, req.ConnectionId)
	return &migrationv1.DryRunMigrationResponse{Valid: valid, Errors: errs, Warnings: warnings}, nil
}

func (h *MigrationHandler) GetMigrationRun(ctx context.Context, req *migrationv1.GetMigrationRunRequest) (*migrationv1.GetMigrationRunResponse, error) {
	run, err := h.svc.GetRunByID(ctx, req.Id)
	if err != nil {
		return nil, errors.ToGRPC(err)
	}
	return &migrationv1.GetMigrationRunResponse{Run: toProtoRun(run)}, nil
}

func (h *MigrationHandler) ListMigrationRuns(ctx context.Context, req *migrationv1.ListMigrationRunsRequest) (*migrationv1.ListMigrationRunsResponse, error) {
	runs, next, total, err := h.svc.ListRuns(ctx, req.MigrationId, req.Cursor, req.PageSize)
	if err != nil {
		return nil, errors.ToGRPC(err)
	}

	var pr []*migrationv1.MigrationRun
	for _, r := range runs {
		pr = append(pr, toProtoRun(r))
	}
	return &migrationv1.ListMigrationRunsResponse{Runs: pr, NextCursor: next, TotalCount: total}, nil
}

func (h *MigrationHandler) GetMigrationLogs(req *migrationv1.GetMigrationLogsRequest, stream migrationv1.MigrationService_GetMigrationLogsServer) error {
	entries, err := h.svc.GetLogs(stream.Context(), req.RunId)
	if err != nil {
		return errors.ToGRPC(err)
	}

	for _, e := range entries {
		if err := stream.Send(toProtoLogEntry(e)); err != nil {
			return err
		}
	}
	return nil
}

// â”€â”€ Converters â”€â”€

func toProtoMigration(m *domain.Migration) *migrationv1.Migration {
	if m == nil {
		return nil
	}
	pm := &migrationv1.Migration{
		Id:          m.ID,
		ProjectId:   m.ProjectID,
		Title:       m.Title,
		Description: m.Description,
		Version:     m.Version,
		UpSql:       m.UpSQL,
		DownSql:     m.DownSQL,
		Checksum:    m.Checksum,
		Status:      string(m.Status),
		CreatedBy:   m.CreatedBy,
		CreatedAt:   m.CreatedAt.Format(time.RFC3339),
		UpdatedAt:   m.UpdatedAt.Format(time.RFC3339),
	}
	return pm
}

func toProtoRun(r *domain.MigrationRun) *migrationv1.MigrationRun {
	if r == nil {
		return nil
	}
	pr := &migrationv1.MigrationRun{
		Id:           r.ID,
		MigrationId:  r.MigrationID,
		ConnectionId: r.ConnectionID,
		Direction:    string(r.Direction),
		Status:       string(r.Status),
		DurationMs:   r.DurationMs,
		ErrorMessage: r.ErrorMessage,
		ExecutedBy:   r.ExecutedBy,
		CreatedAt:    r.CreatedAt.Format(time.RFC3339),
	}
	if r.StartedAt != nil {
		pr.StartedAt = r.StartedAt.Format(time.RFC3339)
	}
	if r.CompletedAt != nil {
		pr.CompletedAt = r.CompletedAt.Format(time.RFC3339)
	}
	return pr
}

func toProtoStatus(msg *domain.MigrationStatusMessage) *migrationv1.MigrationStatusMessage {
	if msg == nil {
		return nil
	}
	ps := &migrationv1.MigrationStatusMessage{
		RunId:               msg.RunID,
		State:               string(msg.State),
		TotalStatements:     int32(msg.TotalStatements),
		CompletedStatements: int32(msg.CompletedStatements),
		CurrentStatement:    msg.CurrentStatement,
		StartedAt:           msg.StartedAt.Format(time.RFC3339),
		ElapsedMs:           msg.ElapsedMs,
		ErrorMessage:        msg.ErrorMessage,
	}
	if msg.LastLog != nil {
		ps.LastLog = toProtoLogEntry(msg.LastLog)
	}
	return ps
}

func toProtoLogEntry(e *domain.MigrationLogEntry) *migrationv1.MigrationLogEntry {
	if e == nil {
		return nil
	}
	pe := &migrationv1.MigrationLogEntry{
		Sequence:     int32(e.Sequence),
		Sql:          e.SQL,
		ErrorMessage: e.ErrorMessage,
		CreatedAt:    e.CreatedAt.Format(time.RFC3339),
	}
	if e.DurationMs != nil {
		pe.DurationMs = *e.DurationMs
	}
	if e.RowsAffected != nil {
		pe.RowsAffected = *e.RowsAffected
	}
	return pe
}

var _ = fmt.Sprintf
