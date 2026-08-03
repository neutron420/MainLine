package handler

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
	"github.com/schemahub/backend/internal/event/domain"
	"github.com/schemahub/backend/internal/pkg/interceptor"
	eventv1 "github.com/schemahub/backend/proto/event/v1"
	"google.golang.org/grpc"
)

type fakeEventStream struct {
	grpc.ServerStream
	ctx  context.Context
	mu   sync.Mutex
	evts []*eventv1.SchemaEvent
}

func (f *fakeEventStream) Context() context.Context { return f.ctx }

func (f *fakeEventStream) Send(e *eventv1.SchemaEvent) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.evts = append(f.evts, e)
	return nil
}

func (f *fakeEventStream) count() int {
	f.mu.Lock()
	defer f.mu.Unlock()
	return len(f.evts)
}

type fakeAuditLogger struct {
	events []*domain.SchemaEvent
}

func (f *fakeAuditLogger) InsertEvent(ctx context.Context, evt *domain.SchemaEvent) error {
	return nil
}

func (f *fakeAuditLogger) ListEventsAfter(ctx context.Context, afterID string, projectIDs []string, eventTypes []domain.EventType, limit int) ([]*domain.SchemaEvent, error) {
	return f.events, nil
}

func TestEventHandler_AcknowledgeEvent(t *testing.T) {
	t.Parallel()

	srv := miniredis.RunT(t)
	rdb := redis.NewClient(&redis.Options{Addr: srv.Addr()})
	defer rdb.Close()

	h := NewEventHandler(domain.NewEventService(rdb, nil))

	ctx := context.WithValue(context.Background(), interceptor.UserIDKey, "user_1")
	resp, err := h.AcknowledgeEvent(ctx, &eventv1.AcknowledgeEventRequest{
		EventId: "evt_1",
	})
	if err != nil {
		t.Fatalf("AcknowledgeEvent() error = %v", err)
	}
	if resp == nil {
		t.Fatal("AcknowledgeEvent() returned nil response")
	}

	acked, err := h.svc.IsAcknowledged(ctx, "user_1", "evt_1")
	if err != nil {
		t.Fatalf("IsAcknowledged() error = %v", err)
	}
	if !acked {
		t.Error("event evt_1 not marked as acknowledged for user_1")
	}
}

func TestEventHandler_AcknowledgeEventNoUser(t *testing.T) {
	t.Parallel()

	h := NewEventHandler(domain.NewEventService(nil, nil))

	_, err := h.AcknowledgeEvent(context.Background(), &eventv1.AcknowledgeEventRequest{
		EventId: "evt_1",
	})
	if err == nil {
		t.Fatal("AcknowledgeEvent() without user in context = nil error, want error")
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

func TestEventHandler_Subscribe(t *testing.T) {
	t.Parallel()

	srv := miniredis.RunT(t)
	rdb := redis.NewClient(&redis.Options{Addr: srv.Addr()})
	defer rdb.Close()

	h := NewEventHandler(domain.NewEventService(rdb, nil))

	ctx, cancel := context.WithCancel(context.WithValue(context.Background(), interceptor.UserIDKey, "user_1"))
	defer cancel()

	stream := &fakeEventStream{ctx: ctx}
	done := make(chan error, 1)
	go func() {
		done <- h.Subscribe(&eventv1.SubscribeRequest{
			ProjectIds: []string{"proj_1"},
			EventTypes: []string{string(domain.EventTypeSchemaVersionCreated)},
		}, stream)
	}()

	deadline := time.Now().Add(3 * time.Second)
	for stream.count() == 0 && time.Now().Before(deadline) {
		if err := h.svc.Publish(ctx, &domain.SchemaEvent{
			ID: "evt_1", Type: domain.EventTypeSchemaVersionCreated,
			ProjectID: "proj_1", Timestamp: time.Now(),
		}); err != nil {
			t.Fatalf("Publish() error = %v", err)
		}
		time.Sleep(10 * time.Millisecond)
	}
	if stream.count() == 0 {
		t.Fatal("Subscribe() did not deliver the published event")
	}

	cancel()
	select {
	case err := <-done:
		if err != context.Canceled {
			t.Errorf("Subscribe() error = %v, want context.Canceled", err)
		}
	case <-time.After(3 * time.Second):
		t.Fatal("Subscribe() did not return after context cancel")
	}
}

func TestEventHandler_SubscribeNoUser(t *testing.T) {
	t.Parallel()

	h := NewEventHandler(domain.NewEventService(nil, nil))
	stream := &fakeEventStream{ctx: context.Background()}

	err := h.Subscribe(&eventv1.SubscribeRequest{ProjectIds: []string{"proj_1"}}, stream)
	if err == nil {
		t.Fatal("Subscribe() without user in context = nil error, want error")
	}
}

func TestEventHandler_SubscribeReplaysEvents(t *testing.T) {
	t.Parallel()

	srv := miniredis.RunT(t)
	rdb := redis.NewClient(&redis.Options{Addr: srv.Addr()})
	defer rdb.Close()

	audit := &fakeAuditLogger{events: []*domain.SchemaEvent{
		{ID: "evt_2", Type: domain.EventTypeSchemaVersionCreated, ProjectID: "proj_1", Timestamp: time.Now()},
		{ID: "evt_3", Type: domain.EventTypeMigrationStarted, ProjectID: "proj_1", Timestamp: time.Now()},
	}}
	h := NewEventHandler(domain.NewEventService(rdb, audit))

	ctx, cancel := context.WithCancel(context.WithValue(context.Background(), interceptor.UserIDKey, "user_1"))
	defer cancel()

	stream := &fakeEventStream{ctx: ctx}
	done := make(chan error, 1)
	go func() {
		done <- h.Subscribe(&eventv1.SubscribeRequest{
			ProjectIds:  []string{"proj_1"},
			LastEventId: "evt_1",
		}, stream)
	}()

	deadline := time.Now().Add(3 * time.Second)
	for stream.count() < 2 && time.Now().Before(deadline) {
		time.Sleep(5 * time.Millisecond)
	}
	if stream.count() != 2 {
		t.Errorf("Subscribe() replayed %d events, want 2", stream.count())
	}

	cancel()
	select {
	case <-done:
	case <-time.After(3 * time.Second):
		t.Fatal("Subscribe() did not return after context cancel")
	}
}

func TestEventHandler_HeartbeatSetsPresence(t *testing.T) {
	t.Parallel()

	srv := miniredis.RunT(t)
	rdb := redis.NewClient(&redis.Options{Addr: srv.Addr()})
	defer rdb.Close()

	h := NewEventHandler(domain.NewEventService(rdb, nil))
	ctx := context.WithValue(context.Background(), interceptor.UserIDKey, "user_1")

	resp, err := h.Heartbeat(ctx, &eventv1.HeartbeatRequest{ProjectIds: []string{"proj_1"}})
	if err != nil {
		t.Fatalf("Heartbeat() error = %v", err)
	}
	if resp == nil {
		t.Fatal("Heartbeat() returned nil response")
	}

	users, err := h.svc.GetPresence(context.Background(), "proj_1")
	if err != nil {
		t.Fatalf("GetPresence() error = %v", err)
	}
	if len(users) != 1 || users[0] != "user_1" {
		t.Errorf("GetPresence() = %v, want [user_1]", users)
	}

	users, err = h.svc.GetPresence(context.Background(), "proj_2")
	if err != nil {
		t.Fatalf("GetPresence() error = %v", err)
	}
	if len(users) != 0 {
		t.Errorf("GetPresence(empty project) = %v, want []", users)
	}
}

func TestEventHandler_AcknowledgeEventMissingEventID(t *testing.T) {
	t.Parallel()

	srv := miniredis.RunT(t)
	rdb := redis.NewClient(&redis.Options{Addr: srv.Addr()})
	defer rdb.Close()

	h := NewEventHandler(domain.NewEventService(rdb, nil))
	ctx := context.WithValue(context.Background(), interceptor.UserIDKey, "user_1")

	_, err := h.AcknowledgeEvent(ctx, &eventv1.AcknowledgeEventRequest{})
	if err == nil {
		t.Fatal("AcknowledgeEvent() with empty event id = nil error, want error")
	}
}
