package domain

import (
	"context"
	"fmt"
	"time"

	"github.com/schemahub/backend/internal/pkg/errors"
)

type AuditService struct {
	repo AuditRepository
}

func NewAuditService(repo AuditRepository) *AuditService {
	return &AuditService{repo: repo}
}

func (s *AuditService) Insert(ctx context.Context, entry *AuditEntry) error {
	return s.repo.Insert(ctx, entry)
}

func (s *AuditService) GetByID(ctx context.Context, id string) (*AuditEntry, error) {
	entry, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("getting audit entry: %w", err)
	}
	return entry, nil
}

func (s *AuditService) List(ctx context.Context, filter *AuditFilter, cursor string, pageSize int32) ([]*AuditEntry, string, int32, error) {
	if pageSize <= 0 || pageSize > 100 {
		pageSize = 20
	}
	return s.repo.List(ctx, filter, cursor, pageSize)
}

func (s *AuditService) ListAfterID(ctx context.Context, afterID string, eventType string, limit int) ([]*AuditEntry, error) {
	if limit <= 0 || limit > 1000 {
		limit = 1000
	}
	return s.repo.ListAfterID(ctx, afterID, eventType, limit)
}

func (s *AuditService) GetStats(ctx context.Context, dateFrom, dateTo string) (*AuditStats, error) {
	var from, to time.Time
	var err error

	if dateFrom != "" {
		from, err = time.Parse(time.RFC3339, dateFrom)
		if err != nil {
			return nil, errors.New("INVALID_ARGUMENT", "invalid date_from format, use RFC3339")
		}
	}
	if dateTo != "" {
		to, err = time.Parse(time.RFC3339, dateTo)
		if err != nil {
			return nil, errors.New("INVALID_ARGUMENT", "invalid date_to format, use RFC3339")
		}
	}

	return s.repo.GetStats(ctx, from, to)
}
