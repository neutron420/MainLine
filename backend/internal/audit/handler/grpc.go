package handler

import (
	"context"
	"time"

	"github.com/schemahub/backend/internal/audit/domain"
	"github.com/schemahub/backend/internal/pkg/errors"
	auditv1 "github.com/schemahub/backend/proto/audit/v1"
)

type AuditHandler struct {
	auditv1.UnimplementedAuditServiceServer
	svc *domain.AuditService
}

func NewAuditHandler(svc *domain.AuditService) *AuditHandler {
	return &AuditHandler{svc: svc}
}

func (h *AuditHandler) ListAuditEntries(ctx context.Context, req *auditv1.ListAuditEntriesRequest) (*auditv1.ListAuditEntriesResponse, error) {
	filter := &domain.AuditFilter{
		EventType:    req.EventType,
		ActorID:      req.ActorId,
		ResourceType: req.ResourceType,
		ResourceID:   req.ResourceId,
		DateFrom:     req.DateFrom,
		DateTo:       req.DateTo,
	}

	entries, next, total, err := h.svc.List(ctx, filter, req.Cursor, req.PageSize)
	if err != nil {
		return nil, errors.ToGRPC(err)
	}

	var pe []*auditv1.AuditEntry
	for _, e := range entries {
		pe = append(pe, toProtoEntry(e))
	}
	return &auditv1.ListAuditEntriesResponse{Entries: pe, NextCursor: next, TotalCount: total}, nil
}

func (h *AuditHandler) GetAuditEntry(ctx context.Context, req *auditv1.GetAuditEntryRequest) (*auditv1.GetAuditEntryResponse, error) {
	entry, err := h.svc.GetByID(ctx, req.Id)
	if err != nil {
		return nil, errors.ToGRPC(err)
	}
	return &auditv1.GetAuditEntryResponse{Entry: toProtoEntry(entry)}, nil
}

func (h *AuditHandler) TailAuditEntries(req *auditv1.TailAuditEntriesRequest, stream auditv1.AuditService_TailAuditEntriesServer) error {
	ctx := stream.Context()

	for {
		var entries []*domain.AuditEntry
		var err error

		if req.SinceEventId != "" {
			entries, err = h.svc.ListAfterID(ctx, req.SinceEventId, req.EventType, 100)
			if err != nil {
				return errors.ToGRPC(err)
			}
		}

		if len(entries) > 0 {
			for _, e := range entries {
				if err := stream.Send(toProtoEntry(e)); err != nil {
					return err
				}
				req.SinceEventId = e.ID
			}
		}

		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(5 * time.Second):
		}
	}
}

func (h *AuditHandler) GetAuditStats(ctx context.Context, req *auditv1.GetAuditStatsRequest) (*auditv1.GetAuditStatsResponse, error) {
	stats, err := h.svc.GetStats(ctx, req.DateFrom, req.DateTo)
	if err != nil {
		return nil, errors.ToGRPC(err)
	}

	return &auditv1.GetAuditStatsResponse{
		TotalEntries: stats.TotalEntries,
		ByEventType:  stats.ByEventType,
		ByAction:     stats.ByAction,
		UniqueActors: stats.UniqueActors,
		DateFrom:     stats.DateFrom.Format(time.RFC3339),
		DateTo:       stats.DateTo.Format(time.RFC3339),
	}, nil
}

func toProtoEntry(e *domain.AuditEntry) *auditv1.AuditEntry {
	if e == nil {
		return nil
	}
	return &auditv1.AuditEntry{
		Id:              e.ID,
		EventType:       e.EventType,
		ActorId:         e.ActorID,
		ActorEmail:      e.ActorEmail,
		Action:          e.Action,
		ResourceType:    e.ResourceType,
		ResourceId:      e.ResourceID,
		ResourceChanges: e.ResourceChanges,
		Metadata:        e.Metadata,
		IpAddress:       e.IPAddress,
		UserAgent:       e.UserAgent,
		TraceId:         e.TraceID,
		CreatedAt:       e.CreatedAt.Format(time.RFC3339),
	}
}
