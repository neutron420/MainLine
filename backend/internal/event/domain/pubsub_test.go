package domain

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
)

type fakeAuditLogger struct {
	events  []*SchemaEvent
	inserts []*SchemaEvent
	err     error
}

func (f *fakeAuditLogger) InsertEvent(ctx context.Context, evt *SchemaEvent) error {
	if f.err != nil {
		return f.err
	}
	f.inserts = append(f.inserts, evt)
	return nil
}

func (f *fakeAuditLogger) ListEventsAfter(ctx context.Context, afterID string, projectIDs []string, eventTypes []EventType, limit int) ([]*SchemaEvent, error) {
	if f.err != nil {
		return nil, f.err
	}
	return f.events, nil
}

func newTestEventServiceWithAudit(t *testing.T, audit AuditLogger) (*EventService, *miniredis.Miniredis, *redis.Client) {
	t.Helper()
	srv := miniredis.RunT(t)
	rdb := redis.NewClient(&redis.Options{Addr: srv.Addr()})
	t.Cleanup(func() { rdb.Close() })
	return NewEventService(rdb, audit), srv, rdb
}

func newSubscriber(id string, projectIDs []string, eventTypes []EventType) *Subscriber {
	return &Subscriber{
		ID:         id,
		ProjectIDs: projectIDs,
		EventTypes: eventTypes,
		Buffer:     make(chan *SchemaEvent, 16),
		Done:       make(chan struct{}),
	}
}

func testEvent(id, projectID string, evtType EventType) *SchemaEvent {
	return &SchemaEvent{
		ID:        id,
		Type:      evtType,
		Version:   1,
		ProjectID: projectID,
		Payload:   `{"key":"value"}`,
	}
}

func recvEvent(t *testing.T, ch <-chan *SchemaEvent, timeout time.Duration) *SchemaEvent {
	t.Helper()
	select {
	case evt, ok := <-ch:
		if !ok {
			t.Fatal("subscriber buffer closed while waiting for event")
		}
		return evt
	case <-time.After(timeout):
		t.Fatal("timed out waiting for event")
		return nil
	}
}

func assertNoEvent(t *testing.T, ch <-chan *SchemaEvent, wait time.Duration) {
	t.Helper()
	select {
	case evt := <-ch:
		t.Fatalf("received unexpected event %v", evt)
	case <-time.After(wait):
	}
}

func waitForPubSub(t *testing.T, rdb *redis.Client, channel string) {
	t.Helper()
	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		counts, err := rdb.PubSubNumSub(context.Background(), channel).Result()
		if err == nil && counts[channel] >= 1 {
			return
		}
		time.Sleep(10 * time.Millisecond)
	}
	t.Fatalf("pub/sub subscription to %s not established", channel)
}

func TestSubscribeReplay(t *testing.T) {
	t.Parallel()
	audit := &fakeAuditLogger{events: []*SchemaEvent{
		testEvent("evt_1", "proj_1", EventTypeSchemaRefreshed),
		testEvent("evt_2", "proj_1", EventTypeDriftDetected),
	}}
	svc, _, _ := newTestEventServiceWithAudit(t, audit)
	ctx := context.Background()
	sub := newSubscriber("sub_1", []string{"proj_1"}, nil)

	ch, err := svc.Subscribe(ctx, sub, "evt_0")
	if err != nil {
		t.Fatalf("Subscribe: %v", err)
	}
	evt := recvEvent(t, ch, 2*time.Second)
	if evt.ID != "evt_1" || evt.Type != EventTypeSchemaRefreshed {
		t.Errorf("first replayed event = %+v, want evt_1", evt)
	}
	evt = recvEvent(t, ch, 2*time.Second)
	if evt.ID != "evt_2" {
		t.Errorf("second replayed event = %+v, want evt_2", evt)
	}
	assertNoEvent(t, ch, 200*time.Millisecond)
}

func TestSubscribeReplayAuditError(t *testing.T) {
	t.Parallel()
	audit := &fakeAuditLogger{err: errors.New("audit log unavailable")}
	svc, _, _ := newTestEventServiceWithAudit(t, audit)
	ctx := context.Background()
	sub := newSubscriber("sub_1", []string{"proj_1"}, nil)

	ch, err := svc.Subscribe(ctx, sub, "evt_0")
	if err != nil {
		t.Fatalf("Subscribe: %v", err)
	}
	assertNoEvent(t, ch, 300*time.Millisecond)
}

func TestSubscribeLiveFanOut(t *testing.T) {
	t.Parallel()
	svc, _, rdb := newTestEventServiceWithAudit(t, nil)
	ctx := context.Background()
	sub := newSubscriber("sub_1", []string{"proj_1"}, nil)
	t.Cleanup(svc.Stop)

	ch, err := svc.Subscribe(ctx, sub, "")
	if err != nil {
		t.Fatalf("Subscribe: %v", err)
	}
	waitForPubSub(t, rdb, "schema:events:global")

	data, err := json.Marshal(testEvent("live_1", "proj_1", EventTypeMigrationStarted))
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}
	if err := rdb.Publish(ctx, "schema:events:global", string(data)).Err(); err != nil {
		t.Fatalf("raw redis Publish: %v", err)
	}
	evt := recvEvent(t, ch, 2*time.Second)
	if evt.ID != "live_1" {
		t.Errorf("received event = %+v, want live_1", evt)
	}

	if err := svc.Publish(ctx, testEvent("live_2", "proj_1", EventTypeMigrationCompleted)); err != nil {
		t.Fatalf("svc.Publish: %v", err)
	}
	evt = recvEvent(t, ch, 2*time.Second)
	if evt.ID != "live_2" {
		t.Errorf("received event = %+v, want live_2", evt)
	}
}

func TestSubscribeMultipleSubscribers(t *testing.T) {
	t.Parallel()
	svc, _, rdb := newTestEventServiceWithAudit(t, nil)
	ctx := context.Background()
	sub1 := newSubscriber("sub_1", []string{"proj_1"}, nil)
	sub2 := newSubscriber("sub_2", []string{"proj_1"}, nil)
	t.Cleanup(svc.Stop)

	ch1, err := svc.Subscribe(ctx, sub1, "")
	if err != nil {
		t.Fatalf("Subscribe sub_1: %v", err)
	}
	ch2, err := svc.Subscribe(ctx, sub2, "")
	if err != nil {
		t.Fatalf("Subscribe sub_2: %v", err)
	}
	waitForPubSub(t, rdb, "schema:events:global")

	if err := svc.Publish(ctx, testEvent("fan_1", "proj_1", EventTypeConnectionCreated)); err != nil {
		t.Fatalf("Publish: %v", err)
	}
	if evt := recvEvent(t, ch1, 2*time.Second); evt.ID != "fan_1" {
		t.Errorf("sub_1 received %+v, want fan_1", evt)
	}
	if evt := recvEvent(t, ch2, 2*time.Second); evt.ID != "fan_1" {
		t.Errorf("sub_2 received %+v, want fan_1", evt)
	}
}

func TestSubscribeMultipleProjects(t *testing.T) {
	t.Parallel()
	svc, _, rdb := newTestEventServiceWithAudit(t, nil)
	ctx := context.Background()
	sub := newSubscriber("sub_1", []string{"proj_1", "proj_2"}, nil)
	t.Cleanup(svc.Stop)

	ch, err := svc.Subscribe(ctx, sub, "")
	if err != nil {
		t.Fatalf("Subscribe: %v", err)
	}
	waitForPubSub(t, rdb, "schema:events:global")

	if err := svc.Publish(ctx, testEvent("m_1", "proj_1", EventTypeMemberAdded)); err != nil {
		t.Fatalf("Publish proj_1: %v", err)
	}
	if evt := recvEvent(t, ch, 2*time.Second); evt.ID != "m_1" {
		t.Errorf("received %+v, want m_1", evt)
	}
	if err := svc.Publish(ctx, testEvent("m_2", "proj_2", EventTypeMemberRemoved)); err != nil {
		t.Fatalf("Publish proj_2: %v", err)
	}
	if evt := recvEvent(t, ch, 2*time.Second); evt.ID != "m_2" {
		t.Errorf("received %+v, want m_2", evt)
	}
}

func TestSubscribeEventTypeFiltering(t *testing.T) {
	t.Parallel()
	svc, _, rdb := newTestEventServiceWithAudit(t, nil)
	ctx := context.Background()
	subAll := newSubscriber("sub_all", []string{"proj_1"}, nil)
	subMig := newSubscriber("sub_mig", []string{"proj_1"}, []EventType{EventTypeMigrationStarted})
	t.Cleanup(svc.Stop)

	chAll, err := svc.Subscribe(ctx, subAll, "")
	if err != nil {
		t.Fatalf("Subscribe sub_all: %v", err)
	}
	chMig, err := svc.Subscribe(ctx, subMig, "")
	if err != nil {
		t.Fatalf("Subscribe sub_mig: %v", err)
	}
	waitForPubSub(t, rdb, "schema:events:global")

	if err := svc.Publish(ctx, testEvent("e_1", "proj_1", EventTypeSchemaVersionCreated)); err != nil {
		t.Fatalf("Publish: %v", err)
	}
	if evt := recvEvent(t, chAll, 2*time.Second); evt.ID != "e_1" {
		t.Errorf("sub_all received %+v, want e_1", evt)
	}
	assertNoEvent(t, chMig, 300*time.Millisecond)

	if err := svc.Publish(ctx, testEvent("e_2", "proj_1", EventTypeMigrationStarted)); err != nil {
		t.Fatalf("Publish: %v", err)
	}
	if evt := recvEvent(t, chMig, 2*time.Second); evt.ID != "e_2" {
		t.Errorf("sub_mig received %+v, want e_2", evt)
	}
	if evt := recvEvent(t, chAll, 2*time.Second); evt.ID != "e_2" {
		t.Errorf("sub_all received %+v, want e_2", evt)
	}
}

func TestUnsubscribeRemovesSubscriber(t *testing.T) {
	t.Parallel()
	svc, _, rdb := newTestEventServiceWithAudit(t, nil)
	ctx := context.Background()
	sub := newSubscriber("sub_1", []string{"proj_1"}, nil)
	t.Cleanup(svc.Stop)

	ch, err := svc.Subscribe(ctx, sub, "")
	if err != nil {
		t.Fatalf("Subscribe: %v", err)
	}
	waitForPubSub(t, rdb, "schema:events:global")

	svc.Unsubscribe("sub_1")
	if _, ok := <-ch; ok {
		t.Error("subscriber buffer still open after Unsubscribe")
	}

	svc.subMu.Lock()
	if len(svc.subscribers) != 0 {
		t.Errorf("subscribers = %d, want 0", len(svc.subscribers))
	}
	if len(svc.channelSubs) != 0 {
		t.Errorf("channelSubs = %v, want empty", svc.channelSubs)
	}
	svc.subMu.Unlock()

	if err := svc.Publish(ctx, testEvent("u_1", "proj_1", EventTypeRoleChanged)); err != nil {
		t.Fatalf("Publish after unsubscribe: %v", err)
	}
	svc.Unsubscribe("sub_1")
}

func TestPublishChannels(t *testing.T) {
	t.Parallel()
	svc, _, rdb := newTestEventServiceWithAudit(t, nil)
	ctx := context.Background()
	ps := rdb.Subscribe(ctx, "schema:events:project:proj_1", "schema:events:global")
	defer ps.Close()
	waitForPubSub(t, rdb, "schema:events:project:proj_1")
	waitForPubSub(t, rdb, "schema:events:global")
	ch := ps.Channel()

	if err := svc.Publish(ctx, testEvent("pub_1", "proj_1", EventTypeConnectionStatusChanged)); err != nil {
		t.Fatalf("Publish: %v", err)
	}

	got := map[string]bool{}
	for len(got) < 2 {
		select {
		case msg := <-ch:
			var parsed map[string]interface{}
			if err := json.Unmarshal([]byte(msg.Payload), &parsed); err != nil {
				t.Fatalf("unmarshal payload: %v", err)
			}
			if parsed["ID"] != "pub_1" {
				t.Errorf("payload ID = %v, want pub_1", parsed["ID"])
			}
			got[msg.Channel] = true
		case <-time.After(2 * time.Second):
			t.Fatalf("expected messages on both channels, got %v", got)
		}
	}
	if !got["schema:events:project:proj_1"] {
		t.Error("project channel did not receive event")
	}
	if !got["schema:events:global"] {
		t.Error("global channel did not receive event")
	}
}

func TestPublishEdgeCases(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	svc, _, _ := newTestEventServiceWithAudit(t, nil)
	emptyProject := testEvent("ep_1", "", EventTypeSchemaRefreshed)
	emptyProject.Payload = ""
	if err := svc.Publish(ctx, emptyProject); err != nil {
		t.Errorf("Publish with empty project ID/payload: %v", err)
	}

	audit := &fakeAuditLogger{}
	svcAudit, _, _ := newTestEventServiceWithAudit(t, audit)
	if err := svcAudit.Publish(ctx, testEvent("a_1", "proj_1", EventTypeDriftDetected)); err != nil {
		t.Fatalf("Publish with audit: %v", err)
	}
	if len(audit.inserts) != 1 || audit.inserts[0].ID != "a_1" {
		t.Errorf("audit inserts = %+v, want [a_1]", audit.inserts)
	}

	auditErr := &fakeAuditLogger{err: errors.New("audit log unavailable")}
	svcErr, _, _ := newTestEventServiceWithAudit(t, auditErr)
	if err := svcErr.Publish(ctx, testEvent("a_2", "proj_1", EventTypeDriftResolved)); err != nil {
		t.Errorf("Publish must ignore audit failure, got %v", err)
	}
}

func TestSendHeartbeatMultiProjectTTL(t *testing.T) {
	t.Parallel()
	svc, srv, rdb := newTestEventServiceWithAudit(t, nil)
	ctx := context.Background()

	if err := svc.SendHeartbeat(ctx, "user_1", []string{"proj_1", "proj_2"}); err != nil {
		t.Fatalf("SendHeartbeat: %v", err)
	}

	for _, project := range []string{"proj_1", "proj_2"} {
		users, err := svc.GetPresence(ctx, project)
		if err != nil {
			t.Fatalf("GetPresence(%s): %v", project, err)
		}
		if len(users) != 1 || users[0] != "user_1" {
			t.Errorf("presence for %s = %v, want [user_1]", project, users)
		}
	}

	ttl, err := rdb.TTL(ctx, "presence:project:proj_1:user_1").Result()
	if err != nil {
		t.Fatalf("TTL: %v", err)
	}
	if ttl != 60*time.Second {
		t.Errorf("presence TTL = %v, want 60s", ttl)
	}

	srv.FastForward(61 * 1e9)
	users, err := svc.GetPresence(ctx, "proj_2")
	if err != nil {
		t.Fatalf("GetPresence after expiry: %v", err)
	}
	if len(users) != 0 {
		t.Errorf("presence after expiry = %v, want empty", users)
	}
}

func TestAcknowledgeTTL(t *testing.T) {
	t.Parallel()
	svc, _, rdb := newTestEventServiceWithAudit(t, nil)
	ctx := context.Background()

	if err := svc.Acknowledge(ctx, "user_1", "evt_1"); err != nil {
		t.Fatalf("Acknowledge: %v", err)
	}
	acked, err := svc.IsAcknowledged(ctx, "user_1", "evt_1")
	if err != nil {
		t.Fatalf("IsAcknowledged: %v", err)
	}
	if !acked {
		t.Error("evt_1 should be acknowledged")
	}

	ttl, err := rdb.TTL(ctx, "schema:events:acked:user_1").Result()
	if err != nil {
		t.Fatalf("TTL: %v", err)
	}
	if ttl < 23*time.Hour || ttl > 25*time.Hour {
		t.Errorf("ack TTL = %v, want ~24h", ttl)
	}

	other, err := svc.IsAcknowledged(ctx, "user_2", "evt_1")
	if err != nil {
		t.Fatalf("IsAcknowledged other: %v", err)
	}
	if other {
		t.Error("evt_1 acknowledged for wrong user")
	}
}

func TestStopCancelsPubSub(t *testing.T) {
	t.Parallel()
	svc, _, rdb := newTestEventServiceWithAudit(t, nil)
	ctx := context.Background()
	sub := newSubscriber("sub_1", []string{"proj_1"}, nil)

	ch, err := svc.Subscribe(ctx, sub, "")
	if err != nil {
		t.Fatalf("Subscribe: %v", err)
	}
	waitForPubSub(t, rdb, "schema:events:global")

	if err := svc.Publish(ctx, testEvent("s_1", "proj_1", EventTypeMigrationStarted)); err != nil {
		t.Fatalf("Publish: %v", err)
	}
	if evt := recvEvent(t, ch, 2*time.Second); evt.ID != "s_1" {
		t.Errorf("received %+v, want s_1", evt)
	}

	svc.Stop()
	svc.psMu.Lock()
	started := svc.psStarted
	svc.psMu.Unlock()
	if started {
		t.Error("psStarted still true after Stop")
	}

	if err := svc.Publish(ctx, testEvent("s_2", "proj_1", EventTypeMigrationCompleted)); err != nil {
		t.Fatalf("Publish after Stop: %v", err)
	}
	assertNoEvent(t, ch, 300*time.Millisecond)

	svc.Stop()
}
