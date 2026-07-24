package handler

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/schemahub/backend/internal/event/domain"
	"github.com/schemahub/backend/internal/pkg/interceptor"
	eventv1 "github.com/schemahub/backend/proto/event/v1"
)

type EventHandler struct {
	eventv1.UnimplementedEventServiceServer
	svc *domain.EventService
}

func NewEventHandler(svc *domain.EventService) *EventHandler {
	return &EventHandler{svc: svc}
}

func (h *EventHandler) Subscribe(req *eventv1.SubscribeRequest, stream eventv1.EventService_SubscribeServer) error {
	userID, err := interceptor.UserIDFromContext(stream.Context())
	if err != nil {
		return err
	}

	var eventTypes []domain.EventType
	for _, et := range req.EventTypes {
		eventTypes = append(eventTypes, domain.EventType(et))
	}

	sub := &domain.Subscriber{
		ID:         uuid.NewString(),
		UserID:     userID,
		ProjectIDs: req.ProjectIDs,
		EventTypes: eventTypes,
		Buffer:     make(chan *domain.SchemaEvent, 100),
		Done:       make(chan struct{}),
	}

	eventCh, err := h.svc.Subscribe(stream.Context(), sub, req.LastEventID)
	if err != nil {
		return err
	}
	defer h.svc.Unsubscribe(sub.ID)

	for {
		select {
		case evt, ok := <-eventCh:
			if !ok {
				return nil
			}
			pbEvt := toProtoEvent(evt)
			if err := stream.Send(pbEvt); err != nil {
				return err
			}
		case <-stream.Context().Done():
			return stream.Context().Err()
		}
	}
}

func (h *EventHandler) AcknowledgeEvent(ctx context.Context, req *eventv1.AcknowledgeEventRequest) (*eventv1.AcknowledgeEventResponse, error) {
	return &eventv1.AcknowledgeEventResponse{}, nil
}

func (h *EventHandler) Heartbeat(ctx context.Context, req *eventv1.HeartbeatRequest) (*eventv1.HeartbeatResponse, error) {
	userID, err := interceptor.UserIDFromContext(ctx)
	if err != nil {
		return nil, err
	}

	if err := h.svc.SendHeartbeat(ctx, userID, req.ProjectIDs); err != nil {
		return nil, err
	}

	return &eventv1.HeartbeatResponse{}, nil
}

func toProtoEvent(e *domain.SchemaEvent) *eventv1.SchemaEvent {
	if e == nil {
		return nil
	}
	pe := &eventv1.SchemaEvent{
		ID:        e.ID,
		Type:      string(e.Type),
		Version:   e.Version,
		Timestamp: e.Timestamp.Format(time.RFC3339Nano),
		ProjectID: e.ProjectID,
		Payload:   e.Payload,
		Metadata:  e.Metadata,
	}
	if e.Actor != nil {
		pe.Actor = &eventv1.EventActor{ID: e.Actor.ID, Email: e.Actor.Email}
	}
	if e.Resource != nil {
		pe.Resource = &eventv1.EventResource{Type: e.Resource.Type, ID: e.Resource.ID}
	}
	return pe
}

var _ = fmt.Sprintf
