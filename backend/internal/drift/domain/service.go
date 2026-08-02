package domain

import (
	"context"
	"fmt"

	"github.com/schemahub/backend/internal/pkg/errors"
)

type SchemaComparator interface {
	CompareLiveWithVersion(ctx context.Context, connStr, connectionID string, schemaNames []string) ([]*DriftEvent, error)
}

type DriftService struct {
	repo       DriftRepository
	comparator SchemaComparator
}

func NewDriftService(repo DriftRepository, comparator SchemaComparator) *DriftService {
	return &DriftService{repo: repo, comparator: comparator}
}

func (s *DriftService) CheckDrift(ctx context.Context, connStr, connectionID string, schemaNames []string) ([]*DriftEvent, error) {
	events, err := s.comparator.CompareLiveWithVersion(ctx, connStr, connectionID, schemaNames)
	if err != nil {
		return nil, fmt.Errorf("comparing live schema: %w", err)
	}

	for _, event := range events {
		event.ConnectionID = connectionID
		if err := s.repo.Insert(ctx, event); err != nil {
			return nil, fmt.Errorf("persisting drift event: %w", err)
		}
	}

	return events, nil
}

func (s *DriftService) List(ctx context.Context, filter *DriftFilter, cursor string, pageSize int32) ([]*DriftEvent, string, int32, error) {
	if pageSize <= 0 || pageSize > 100 {
		pageSize = 20
	}
	return s.repo.List(ctx, filter, cursor, pageSize)
}

func (s *DriftService) GetByID(ctx context.Context, id string) (*DriftEvent, error) {
	event, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	return event, nil
}

func (s *DriftService) Resolve(ctx context.Context, id, status, userID string) (*DriftEvent, error) {
	ds := DriftStatus(status)
	switch ds {
	case DriftStatusResolved, DriftStatusAcknowledged, DriftStatusFalsePositive:
	default:
		return nil, errors.New("INVALID_ARGUMENT", "invalid drift status: must be resolved, acknowledged, or false_positive")
	}

	event, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if event.Status == DriftStatusResolved {
		return nil, errors.New("FAILED_PRECONDITION", "drift event is already resolved")
	}

	if err := s.repo.UpdateStatus(ctx, id, ds, userID); err != nil {
		return nil, fmt.Errorf("updating drift status: %w", err)
	}

	event.Status = ds
	event.ResolvedBy = userID
	return event, nil
}

func (s *DriftService) GetStats(ctx context.Context, connectionID string) (*DriftStats, error) {
	return s.repo.GetStats(ctx, connectionID)
}
