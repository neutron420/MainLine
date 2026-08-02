package handler

import (
	"context"
	"testing"

	"github.com/redis/go-redis/v9"
	"github.com/schemahub/backend/internal/event/domain"
	"github.com/schemahub/backend/internal/pkg/interceptor"
	eventv1 "github.com/schemahub/backend/proto/event/v1"
)

func TestEventHandler_AcknowledgeEvent(t *testing.T) {
	t.Parallel()

	h := NewEventHandler(domain.NewEventService(nil, nil))

	resp, err := h.AcknowledgeEvent(context.Background(), &eventv1.AcknowledgeEventRequest{
		EventId: "evt_1",
	})
	if err != nil {
		t.Fatalf("AcknowledgeEvent() error = %v", err)
	}
	if resp == nil {
		t.Fatal("AcknowledgeEvent() returned nil response")
	}
}

func TestEventHandler_HeartbeatNoUser(t *testing.T) {
	t.Parallel()

	h := NewEventHandler(domain.NewEventService(nil, nil))

	_, err := h.Heartbeat(context.Background(), &eventv1.HeartbeatRequest{
		ProjectIds: []string{"proj_1"},
	})
	if err == nil {
		t.Fatal("Heartbeat() without user in context = nil error, want error")
	}
}

func TestEventHandler_HeartbeatRedisUnavailable(t *testing.T) {
	t.Parallel()

	rdb := redis.NewClient(&redis.Options{Addr: "127.0.0.1:1"})
	defer rdb.Close()
	h := NewEventHandler(domain.NewEventService(rdb, nil))

	ctx := context.WithValue(context.Background(), interceptor.UserIDKey, "user_1")
	_, err := h.Heartbeat(ctx, &eventv1.HeartbeatRequest{
		ProjectIds: []string{"proj_1"},
	})
	if err == nil {
		t.Fatal("Heartbeat() with unreachable redis = nil error, want error")
	}
}
