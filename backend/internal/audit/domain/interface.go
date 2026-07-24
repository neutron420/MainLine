package domain

import (
	"context"
	"time"
)

type AuditRepository interface {
	Insert(ctx context.Context, entry *AuditEntry) error
	GetByID(ctx context.Context, id string) (*AuditEntry, error)
	List(ctx context.Context, filter *AuditFilter, cursor string, limit int32) ([]*AuditEntry, string, int32, error)
	ListAfterID(ctx context.Context, afterID string, eventType string, limit int) ([]*AuditEntry, error)
	GetStats(ctx context.Context, dateFrom, dateTo time.Time) (*AuditStats, error)
}

type AuditFilter struct {
	EventType    string
	ActorID      string
	ResourceType string
	ResourceID   string
	DateFrom     string
	DateTo       string
}
