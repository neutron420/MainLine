package domain

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/schemahub/backend/internal/pkg/errors"
)

type AuditLogger interface {
	InsertEvent(ctx context.Context, evt *SchemaEvent) error
	ListEventsAfter(ctx context.Context, afterID string, projectIDs []string, eventTypes []EventType, limit int) ([]*SchemaEvent, error)
}

type EventService struct {
	rdb         *redis.Client
	audit       AuditLogger
	subMu       sync.RWMutex
	subscribers map[string]*Subscriber
	channelSubs map[string]map[string]bool

	pubsub    *redis.PubSub
	psMu      sync.Mutex
	psStarted bool
	psCtx     context.Context
	psCancel  context.CancelFunc
}

func NewEventService(rdb *redis.Client, audit AuditLogger) *EventService {
	return &EventService{
		rdb:         rdb,
		audit:       audit,
		subscribers: make(map[string]*Subscriber),
		channelSubs: make(map[string]map[string]bool),
	}
}

func (s *EventService) Subscribe(ctx context.Context, sub *Subscriber, lastEventID string) (<-chan *SchemaEvent, error) {
	s.subMu.Lock()
	s.subscribers[sub.ID] = sub
	for _, pid := range sub.ProjectIDs {
		channel := fmt.Sprintf("schema:events:project:%s", pid)
		if s.channelSubs[channel] == nil {
			s.channelSubs[channel] = make(map[string]bool)
		}
		s.channelSubs[channel][sub.ID] = true
	}
	s.subMu.Unlock()

	s.ensurePubSubRunning()

	if lastEventID != "" {
		go s.replayEvents(ctx, sub, lastEventID)
	}

	return sub.Buffer, nil
}

func (s *EventService) Unsubscribe(subID string) {
	s.subMu.Lock()
	sub, ok := s.subscribers[subID]
	if ok {
		for channel, subs := range s.channelSubs {
			delete(subs, subID)
			if len(subs) == 0 {
				delete(s.channelSubs, channel)
			}
		}
		delete(s.subscribers, subID)
		close(sub.Done)
		close(sub.Buffer)
	}
	s.subMu.Unlock()
}

func (s *EventService) Publish(ctx context.Context, evt *SchemaEvent) error {
	if s.audit != nil {
		if err := s.audit.InsertEvent(ctx, evt); err != nil {
			log.Printf("failed to persist event to audit log: %v", err)
		}
	}

	data, err := json.Marshal(evt)
	if err != nil {
		return fmt.Errorf("marshaling event: %w", err)
	}

	projectChannel := fmt.Sprintf("schema:events:project:%s", evt.ProjectID)
	if err := s.rdb.Publish(ctx, projectChannel, string(data)).Err(); err != nil {
		return fmt.Errorf("publishing to project channel: %w", err)
	}

	if err := s.rdb.Publish(ctx, "schema:events:global", string(data)).Err(); err != nil {
		return fmt.Errorf("publishing to global channel: %w", err)
	}

	return nil
}

func (s *EventService) SendHeartbeat(ctx context.Context, userID string, projectIDs []string) error {
	now := time.Now().Unix()
	for _, pid := range projectIDs {
		key := fmt.Sprintf("presence:project:%s:%s", pid, userID)
		if err := s.rdb.Set(ctx, key, now, 60*time.Second).Err(); err != nil {
			return fmt.Errorf("setting presence key: %w", err)
		}
	}
	return nil
}

func (s *EventService) GetPresence(ctx context.Context, projectID string) ([]string, error) {
	pattern := fmt.Sprintf("presence:project:%s:*", projectID)
	keys, err := s.rdb.Keys(ctx, pattern).Result()
	if err != nil {
		return nil, fmt.Errorf("querying presence keys: %w", err)
	}

	var userIDs []string
	prefix := fmt.Sprintf("presence:project:%s:", projectID)
	for _, key := range keys {
		uid := key[len(prefix):]
		userIDs = append(userIDs, uid)
	}
	return userIDs, nil
}

const ackKeyTTL = 24 * time.Hour

func ackKey(userID string) string {
	return fmt.Sprintf("schema:events:acked:%s", userID)
}

// Acknowledge records that a user has processed an event. Acks live in a Redis
// set keyed by user and expire after ackKeyTTL.
func (s *EventService) Acknowledge(ctx context.Context, userID, eventID string) error {
	if userID == "" || eventID == "" {
		return fmt.Errorf("user id and event id are required")
	}
	key := ackKey(userID)
	pipe := s.rdb.TxPipeline()
	pipe.SAdd(ctx, key, eventID)
	pipe.Expire(ctx, key, ackKeyTTL)
	if _, err := pipe.Exec(ctx); err != nil {
		return fmt.Errorf("storing acknowledgement: %w", err)
	}
	return nil
}

// IsAcknowledged reports whether the user already acknowledged an event.
func (s *EventService) IsAcknowledged(ctx context.Context, userID, eventID string) (bool, error) {
	member, err := s.rdb.SIsMember(ctx, ackKey(userID), eventID).Result()
	if err != nil {
		return false, fmt.Errorf("checking acknowledgement: %w", err)
	}
	return member, nil
}

func (s *EventService) ensurePubSubRunning() {
	s.psMu.Lock()
	defer s.psMu.Unlock()
	if s.psStarted {
		return
	}

	s.psCtx, s.psCancel = context.WithCancel(context.Background())
	s.pubsub = s.rdb.Subscribe(s.psCtx, "schema:events:global")

	ch := s.pubsub.Channel()
	go s.listenPubSub(ch)

	s.psStarted = true
}

func (s *EventService) listenPubSub(ch <-chan *redis.Message) {
	for msg := range ch {
		var evt SchemaEvent
		if err := json.Unmarshal([]byte(msg.Payload), &evt); err != nil {
			log.Printf("failed to unmarshal event from redis: %v", err)
			continue
		}

		s.subMu.RLock()
		projectChannel := fmt.Sprintf("schema:events:project:%s", evt.ProjectID)
		subIDs := make(map[string]bool)
		for sid := range s.channelSubs["schema:events:global"] {
			subIDs[sid] = true
		}
		for sid := range s.channelSubs[projectChannel] {
			subIDs[sid] = true
		}
		s.subMu.RUnlock()

		for sid := range subIDs {
			s.subMu.RLock()
			sub, ok := s.subscribers[sid]
			s.subMu.RUnlock()
			if !ok {
				continue
			}

			if !matchesEventType(sub.EventTypes, EventType(evt.Type)) {
				continue
			}

			select {
			case sub.Buffer <- &evt:
			default:
				log.Printf("subscriber %s buffer full, dropping event %s", sid, evt.ID)
			}
		}
	}
}

func (s *EventService) replayEvents(ctx context.Context, sub *Subscriber, afterID string) {
	events, err := s.audit.ListEventsAfter(ctx, afterID, sub.ProjectIDs, sub.EventTypes, 1000)
	if err != nil {
		log.Printf("failed to replay events for subscriber %s: %v", sub.ID, err)
		return
	}

	for _, evt := range events {
		select {
		case sub.Buffer <- evt:
		case <-sub.Done:
			return
		default:
			return
		}
	}
}

func (s *EventService) Stop() {
	s.psMu.Lock()
	if s.psCancel != nil {
		s.psCancel()
	}
	if s.pubsub != nil {
		s.pubsub.Close()
	}
	s.psStarted = false
	s.psMu.Unlock()
}

func matchesEventType(filter []EventType, evt EventType) bool {
	if len(filter) == 0 {
		return true
	}
	for _, ft := range filter {
		if ft == evt {
			return true
		}
	}
	return false
}

var _ = errors.New
