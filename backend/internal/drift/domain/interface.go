package domain

import "context"

type DriftRepository interface {
	Insert(ctx context.Context, event *DriftEvent) error
	GetByID(ctx context.Context, id string) (*DriftEvent, error)
	List(ctx context.Context, filter *DriftFilter, cursor string, limit int32) ([]*DriftEvent, string, int32, error)
	UpdateStatus(ctx context.Context, id string, status DriftStatus, resolvedBy string) error
	GetStats(ctx context.Context, connectionID string) (*DriftStats, error)
}

type DriftFilter struct {
	ConnectionID string
	Status       string
	Severity     string
	DriftType    string
}
