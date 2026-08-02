package domain

import (
	"context"
	"testing"

	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
)

func newTestEventService(t *testing.T) (*EventService, *miniredis.Miniredis) {
	t.Helper()
	srv := miniredis.RunT(t)
	rdb := redis.NewClient(&redis.Options{Addr: srv.Addr()})
	t.Cleanup(func() { rdb.Close() })
	return NewEventService(rdb, nil), srv
}

func TestEventServiceAcknowledge(t *testing.T) {
	t.Parallel()

	svc, _ := newTestEventService(t)
	ctx := context.Background()

	if err := svc.Acknowledge(ctx, "user_1", "evt_1"); err != nil {
		t.Fatalf("Acknowledge: %v", err)
	}

	acked, err := svc.IsAcknowledged(ctx, "user_1", "evt_1")
	if err != nil {
		t.Fatalf("IsAcknowledged: %v", err)
	}
	if !acked {
		t.Error("evt_1 should be acknowledged for user_1")
	}

	other, err := svc.IsAcknowledged(ctx, "user_2", "evt_1")
	if err != nil {
		t.Fatalf("IsAcknowledged other user: %v", err)
	}
	if other {
		t.Error("evt_1 must not be acknowledged for user_2")
	}
}

func TestEventServiceAcknowledgeRequiresIDs(t *testing.T) {
	t.Parallel()

	svc, _ := newTestEventService(t)
	ctx := context.Background()

	if err := svc.Acknowledge(ctx, "", "evt_1"); err == nil {
		t.Error("Acknowledge with empty userID = nil error")
	}
	if err := svc.Acknowledge(ctx, "user_1", ""); err == nil {
		t.Error("Acknowledge with empty eventID = nil error")
	}
}

func TestEventServiceSendHeartbeatAndPresence(t *testing.T) {
	t.Parallel()

	svc, srv := newTestEventService(t)
	ctx := context.Background()

	if err := svc.SendHeartbeat(ctx, "user_1", []string{"proj_1"}); err != nil {
		t.Fatalf("SendHeartbeat: %v", err)
	}

	users, err := svc.GetPresence(ctx, "proj_1")
	if err != nil {
		t.Fatalf("GetPresence: %v", err)
	}
	if len(users) != 1 || users[0] != "user_1" {
		t.Errorf("presence = %v, want [user_1]", users)
	}

	srv.FastForward(61 * 1e9) // presence keys expire after 60s
	users, err = svc.GetPresence(ctx, "proj_1")
	if err != nil {
		t.Fatalf("GetPresence after expiry: %v", err)
	}
	if len(users) != 0 {
		t.Errorf("presence after expiry = %v, want empty", users)
	}
}
