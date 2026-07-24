package worker

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/schemahub/backend/internal/drift/domain"
	eventDomain "github.com/schemahub/backend/internal/event/domain"
)

type EventPublisher interface {
	Publish(ctx context.Context, evt *eventDomain.SchemaEvent) error
}

type DriftAlertWorker struct {
	driftRepo domain.DriftRepository
	publisher EventPublisher
	rdb       *redis.Client
}

func NewDriftAlertWorker(driftRepo domain.DriftRepository, publisher EventPublisher, rdb *redis.Client) *DriftAlertWorker {
	return &DriftAlertWorker{driftRepo: driftRepo, publisher: publisher, rdb: rdb}
}

func (w *DriftAlertWorker) Name() string {
	return "drift-alerting"
}

func (w *DriftAlertWorker) Interval() time.Duration {
	return 1 * time.Minute
}

func (w *DriftAlertWorker) Run(ctx context.Context) error {
	events, _, _, err := w.driftRepo.List(ctx, &domain.DriftFilter{Status: string(domain.DriftStatusOpen)}, "", 100)
	if err != nil {
		return fmt.Errorf("fetching open drift events: %w", err)
	}

	for _, evt := range events {
		notifiedKey := fmt.Sprintf("drift:notified:%s", evt.ID)
		alreadyNotified, err := w.rdb.Exists(ctx, notifiedKey).Result()
		if err == nil && alreadyNotified > 0 {
			continue
		}

		alertEvent := &eventDomain.SchemaEvent{
			ID:        fmt.Sprintf("drift-%s", evt.ID),
			Type:      eventDomain.EventType("drift.detected"),
			ProjectID: evt.SchemaID,
			Timestamp: time.Now(),
			Metadata: map[string]string{
				"drift_id":     evt.ID,
				"object_type":  evt.ObjectType,
				"object_name":  evt.ObjectName,
				"drift_type":   string(evt.DriftType),
				"severity":     string(evt.Severity),
				"diff_summary": evt.DiffSummary,
			},
			Resource: &eventDomain.EventResource{
				Type: "drift_event",
				ID:   evt.ID,
			},
		}

		if err := w.publisher.Publish(ctx, alertEvent); err != nil {
			continue
		}

		w.rdb.Set(ctx, notifiedKey, "1", 1*time.Hour)
	}

	return nil
}
