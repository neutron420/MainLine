package handler

import (
	"context"
	"fmt"
	"time"

	"github.com/schemahub/backend/internal/drift/domain"
	"github.com/schemahub/backend/internal/pkg/errors"
	"github.com/schemahub/backend/internal/pkg/interceptor"
	driftv1 "github.com/schemahub/backend/proto/drift/v1"
)

type DriftHandler struct {
	driftv1.UnimplementedDriftServiceServer
	svc        *domain.DriftService
	connString func(ctx context.Context, connID string) (string, error)
}

func NewDriftHandler(svc *domain.DriftService, connString func(ctx context.Context, connID string) (string, error)) *DriftHandler {
	return &DriftHandler{svc: svc, connString: connString}
}

func (h *DriftHandler) CheckDrift(ctx context.Context, req *driftv1.CheckDriftRequest) (*driftv1.CheckDriftResponse, error) {
	connStr, err := h.connString(ctx, req.ConnectionId)
	if err != nil {
		return nil, errors.ToGRPC(err)
	}

	events, err := h.svc.CheckDrift(ctx, connStr, req.ConnectionId, req.SchemaNames)
	if err != nil {
		return nil, errors.ToGRPC(err)
	}

	var pe []*driftv1.DriftEvent
	for _, e := range events {
		pe = append(pe, toProtoEvent(e))
	}
	return &driftv1.CheckDriftResponse{
		Events:      pe,
		HasDrift:    len(pe) > 0,
		TotalDrifts: int32(len(pe)),
	}, nil
}

func (h *DriftHandler) ListDriftEvents(ctx context.Context, req *driftv1.ListDriftEventsRequest) (*driftv1.ListDriftEventsResponse, error) {
	filter := &domain.DriftFilter{
		ConnectionID: req.ConnectionId,
		Status:       req.Status,
		Severity:     req.Severity,
		DriftType:    req.DriftType,
	}

	events, next, total, err := h.svc.List(ctx, filter, req.Cursor, req.PageSize)
	if err != nil {
		return nil, errors.ToGRPC(err)
	}

	var pe []*driftv1.DriftEvent
	for _, e := range events {
		pe = append(pe, toProtoEvent(e))
	}
	return &driftv1.ListDriftEventsResponse{Events: pe, NextCursor: next, TotalCount: total}, nil
}

func (h *DriftHandler) GetDriftEvent(ctx context.Context, req *driftv1.GetDriftEventRequest) (*driftv1.GetDriftEventResponse, error) {
	event, err := h.svc.GetByID(ctx, req.Id)
	if err != nil {
		return nil, errors.ToGRPC(err)
	}
	return &driftv1.GetDriftEventResponse{Event: toProtoEvent(event)}, nil
}

func (h *DriftHandler) ResolveDriftEvent(ctx context.Context, req *driftv1.ResolveDriftEventRequest) (*driftv1.ResolveDriftEventResponse, error) {
	userID, _ := interceptor.UserIDFromContext(ctx)

	event, err := h.svc.Resolve(ctx, req.Id, req.Status, userID)
	if err != nil {
		return nil, errors.ToGRPC(err)
	}
	return &driftv1.ResolveDriftEventResponse{Event: toProtoEvent(event)}, nil
}

func (h *DriftHandler) GetDriftStats(ctx context.Context, req *driftv1.GetDriftStatsRequest) (*driftv1.GetDriftStatsResponse, error) {
	stats, err := h.svc.GetStats(ctx, req.ConnectionId)
	if err != nil {
		return nil, errors.ToGRPC(err)
	}
	return &driftv1.GetDriftStatsResponse{
		TotalOpen:          stats.TotalOpen,
		TotalResolved:      stats.TotalResolved,
		TotalAcknowledged:  stats.TotalAcknowledged,
		TotalFalsePositive: stats.TotalFalsePositive,
		BySeverity:         stats.BySeverity,
		ByDriftType:        stats.ByDriftType,
	}, nil
}

func toProtoEvent(e *domain.DriftEvent) *driftv1.DriftEvent {
	if e == nil {
		return nil
	}
	pe := &driftv1.DriftEvent{
		Id:                 e.ID,
		ConnectionId:       e.ConnectionID,
		SchemaId:           e.SchemaID,
		ExpectedVersionId:  e.ExpectedVersionID,
		DriftType:          string(e.DriftType),
		ObjectType:         e.ObjectType,
		ObjectName:         e.ObjectName,
		ExpectedDefinition: e.ExpectedDefinition,
		ActualDefinition:   e.ActualDefinition,
		DiffSummary:        e.DiffSummary,
		Severity:           string(e.Severity),
		Status:             string(e.Status),
		DetectedAt:         e.DetectedAt.Format(time.RFC3339),
	}
	if e.ResolvedAt != nil {
		pe.ResolvedAt = e.ResolvedAt.Format(time.RFC3339)
	}
	pe.ResolvedBy = e.ResolvedBy
	return pe
}

var _ = fmt.Sprintf
