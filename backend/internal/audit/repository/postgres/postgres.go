package postgres

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/schemahub/backend/internal/audit/domain"
	eventDomain "github.com/schemahub/backend/internal/event/domain"
)

type AuditRepository struct {
	pool *pgxpool.Pool
}

func NewAuditRepository(pool *pgxpool.Pool) *AuditRepository {
	return &AuditRepository{pool: pool}
}

func (r *AuditRepository) Insert(ctx context.Context, entry *domain.AuditEntry) error {
	metaJSON, _ := json.Marshal(entry.Metadata)
	changesJSON := []byte("null")
	if entry.ResourceChanges != "" {
		changesJSON = []byte(entry.ResourceChanges)
	}

	_, err := r.pool.Exec(ctx, `
		INSERT INTO audit_logs (event_type, actor_id, actor_email, action, resource_type, resource_id, resource_changes, metadata, ip_address, user_agent, trace_id)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)`,
		entry.EventType, nullIfEmpty(entry.ActorID), nullIfEmpty(entry.ActorEmail),
		entry.Action, entry.ResourceType, entry.ResourceID,
		changesJSON, metaJSON,
		nullIfEmpty(entry.IPAddress), nullIfEmpty(entry.UserAgent),
		entry.TraceID,
	)
	if err != nil {
		return fmt.Errorf("inserting audit entry: %w", err)
	}
	return nil
}

func (r *AuditRepository) GetByID(ctx context.Context, id string) (*domain.AuditEntry, error) {
	row := r.pool.QueryRow(ctx, `
		SELECT id, event_type, actor_id, actor_email, action, resource_type, resource_id, resource_changes, metadata, ip_address, user_agent, trace_id, created_at
		FROM audit_logs WHERE id = $1`, id)

	return scanEntry(row)
}

func (r *AuditRepository) List(ctx context.Context, filter *domain.AuditFilter, cursor string, limit int32) ([]*domain.AuditEntry, string, int32, error) {
	where := "WHERE 1=1"
	args := []interface{}{}
	argIdx := 1

	if filter.EventType != "" {
		where += fmt.Sprintf(" AND event_type = $%d", argIdx)
		args = append(args, filter.EventType)
		argIdx++
	}
	if filter.ActorID != "" {
		where += fmt.Sprintf(" AND actor_id = $%d", argIdx)
		args = append(args, filter.ActorID)
		argIdx++
	}
	if filter.ResourceType != "" {
		where += fmt.Sprintf(" AND resource_type = $%d", argIdx)
		args = append(args, filter.ResourceType)
		argIdx++
	}
	if filter.ResourceID != "" {
		where += fmt.Sprintf(" AND resource_id = $%d", argIdx)
		args = append(args, filter.ResourceID)
		argIdx++
	}
	if filter.DateFrom != "" {
		where += fmt.Sprintf(" AND created_at >= $%d", argIdx)
		args = append(args, filter.DateFrom)
		argIdx++
	}
	if filter.DateTo != "" {
		where += fmt.Sprintf(" AND created_at <= $%d", argIdx)
		args = append(args, filter.DateTo)
		argIdx++
	}
	if cursor != "" {
		where += fmt.Sprintf(" AND created_at < (SELECT created_at FROM audit_logs WHERE id = $%d)", argIdx)
		args = append(args, cursor)
		argIdx++
	}

	args = append(args, limit+1)
	query := fmt.Sprintf(`
		SELECT id, event_type, actor_id, actor_email, action, resource_type, resource_id, resource_changes, metadata, ip_address, user_agent, trace_id, created_at
		FROM audit_logs %s ORDER BY created_at DESC LIMIT $%d`, where, argIdx)

	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, "", 0, fmt.Errorf("listing audit entries: %w", err)
	}
	defer rows.Close()

	var entries []*domain.AuditEntry
	for rows.Next() {
		entry, err := scanEntry(rows)
		if err != nil {
			return nil, "", 0, err
		}
		entries = append(entries, entry)
	}

	var nextCursor string
	if len(entries) > int(limit) {
		entries = entries[:limit]
		nextCursor = entries[len(entries)-1].ID
	}

	var total int32
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM audit_logs %s", where)
	countArgs := args[:len(args)-1]
	if err := r.pool.QueryRow(ctx, countQuery, countArgs...).Scan(&total); err != nil {
		return nil, "", 0, fmt.Errorf("counting audit entries: %w", err)
	}

	return entries, nextCursor, total, nil
}

func (r *AuditRepository) ListAfterID(ctx context.Context, afterID string, eventType string, limit int) ([]*domain.AuditEntry, error) {
	where := "WHERE created_at > (SELECT created_at FROM audit_logs WHERE id = $1)"
	args := []interface{}{afterID}
	argIdx := 2

	if eventType != "" {
		where += fmt.Sprintf(" AND event_type = $%d", argIdx)
		args = append(args, eventType)
		argIdx++
	}

	args = append(args, limit)
	query := fmt.Sprintf(`
		SELECT id, event_type, actor_id, actor_email, action, resource_type, resource_id, resource_changes, metadata, ip_address, user_agent, trace_id, created_at
		FROM audit_logs %s ORDER BY created_at ASC LIMIT $%d`, where, argIdx)

	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("listing audit entries after id: %w", err)
	}
	defer rows.Close()

	var entries []*domain.AuditEntry
	for rows.Next() {
		entry, err := scanEntry(rows)
		if err != nil {
			return nil, err
		}
		entries = append(entries, entry)
	}
	return entries, nil
}

func (r *AuditRepository) InsertEvent(ctx context.Context, evt *eventDomain.SchemaEvent) error {
	actorID := ""
	actorEmail := ""
	if evt.Actor != nil {
		actorID = evt.Actor.ID
		actorEmail = evt.Actor.Email
	}
	resourceType := ""
	resourceID := ""
	if evt.Resource != nil {
		resourceType = evt.Resource.Type
		resourceID = evt.Resource.ID
	}
	metaJSON, _ := json.Marshal(evt.Metadata)

	_, err := r.pool.Exec(ctx, `
		INSERT INTO audit_logs (event_type, actor_id, actor_email, action, resource_type, resource_id, metadata, trace_id)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`,
		evt.Type, nullIfEmpty(actorID), nullIfEmpty(actorEmail),
		string(evt.Type), resourceType, resourceID,
		metaJSON, evt.ID,
	)
	if err != nil {
		return fmt.Errorf("inserting event audit entry: %w", err)
	}
	return nil
}

func (r *AuditRepository) ListEventsAfter(ctx context.Context, afterID string, projectIDs []string, eventTypes []eventDomain.EventType, limit int) ([]*eventDomain.SchemaEvent, error) {
	where := "WHERE created_at > (SELECT created_at FROM audit_logs WHERE id = $1)"
	args := []interface{}{afterID}
	argIdx := 2

	if len(eventTypes) > 0 {
		where += fmt.Sprintf(" AND event_type = ANY($%d)", argIdx)
		typeStrs := make([]string, len(eventTypes))
		for i, et := range eventTypes {
			typeStrs[i] = string(et)
		}
		args = append(args, typeStrs)
		argIdx++
	}

	if len(projectIDs) > 0 {
		where += fmt.Sprintf(" AND resource_id = ANY($%d)", argIdx)
		args = append(args, projectIDs)
		argIdx++
	}

	args = append(args, limit)
	query := fmt.Sprintf(`
		SELECT id, event_type, actor_id, actor_email, action, resource_type, resource_id, metadata, created_at
		FROM audit_logs %s ORDER BY created_at ASC LIMIT $%d`, where, argIdx)

	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("listing events after id: %w", err)
	}
	defer rows.Close()

	var events []*eventDomain.SchemaEvent
	for rows.Next() {
		var (
			id, eventType, action, resourceType, resourceID string
			createdAt                                        time.Time
			actorID, actorEmail                              *string
			metadata                                         []byte
		)
		if err := rows.Scan(&id, &eventType, &actorID, &actorEmail, &action, &resourceType, &resourceID, &metadata, &createdAt); err != nil {
			return nil, fmt.Errorf("scanning event: %w", err)
		}
		evt := &eventDomain.SchemaEvent{
			ID:        id,
			Type:      eventDomain.EventType(eventType),
			Timestamp: createdAt,
			Resource:  &eventDomain.EventResource{Type: resourceType, ID: resourceID},
		}
		if actorID != nil {
			evt.Actor = &eventDomain.EventActor{ID: *actorID}
			if actorEmail != nil {
				evt.Actor.Email = *actorEmail
			}
		}
		if metadata != nil {
			json.Unmarshal(metadata, &evt.Metadata)
		}
		events = append(events, evt)
	}
	return events, nil
}

func (r *AuditRepository) GetStats(ctx context.Context, dateFrom, dateTo time.Time) (*domain.AuditStats, error) {
	where := "WHERE 1=1"
	args := []interface{}{}
	argIdx := 1

	if !dateFrom.IsZero() {
		where += fmt.Sprintf(" AND created_at >= $%d", argIdx)
		args = append(args, dateFrom)
		argIdx++
	}
	if !dateTo.IsZero() {
		where += fmt.Sprintf(" AND created_at <= $%d", argIdx)
		args = append(args, dateTo)
		argIdx++
	}

	stats := &domain.AuditStats{
		ByEventType: make(map[string]int32),
		ByAction:    make(map[string]int32),
		DateFrom:    dateFrom,
		DateTo:      dateTo,
	}

	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM audit_logs %s", where)
	if err := r.pool.QueryRow(ctx, countQuery, args...).Scan(&stats.TotalEntries); err != nil {
		return nil, fmt.Errorf("counting total entries: %w", err)
	}

	byTypeQuery := fmt.Sprintf("SELECT event_type, COUNT(*) as cnt FROM audit_logs %s GROUP BY event_type", where)
	rows, err := r.pool.Query(ctx, byTypeQuery, args...)
	if err != nil {
		return nil, fmt.Errorf("querying by event type: %w", err)
	}
	defer rows.Close()
	for rows.Next() {
		var et string
		var cnt int32
		if err := rows.Scan(&et, &cnt); err != nil {
			return nil, err
		}
		stats.ByEventType[et] = cnt
	}

	byActionQuery := fmt.Sprintf("SELECT action, COUNT(*) as cnt FROM audit_logs %s GROUP BY action", where)
	rows2, err := r.pool.Query(ctx, byActionQuery, args...)
	if err != nil {
		return nil, fmt.Errorf("querying by action: %w", err)
	}
	defer rows2.Close()
	for rows2.Next() {
		var action string
		var cnt int32
		if err := rows2.Scan(&action, &cnt); err != nil {
			return nil, err
		}
		stats.ByAction[action] = cnt
	}

	uniqueQuery := fmt.Sprintf("SELECT COUNT(DISTINCT actor_id) FROM audit_logs %s", where)
	if err := r.pool.QueryRow(ctx, uniqueQuery, args...).Scan(&stats.UniqueActors); err != nil {
		return nil, fmt.Errorf("counting unique actors: %w", err)
	}

	return stats, nil
}

func scanEntry(row interface{ Scan(dest ...interface{}) error }) (*domain.AuditEntry, error) {
	var (
		id, eventType, action, resourceType, resourceID, traceID string
		createdAt                                                time.Time
		actorID, actorEmail, ipAddr, userAgent                   *string
		resourceChanges, metadata                                []byte
	)
	if err := row.Scan(&id, &eventType, &actorID, &actorEmail, &action, &resourceType, &resourceID, &resourceChanges, &metadata, &ipAddr, &userAgent, &traceID, &createdAt); err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("audit entry not found")
		}
		return nil, fmt.Errorf("scanning audit entry: %w", err)
	}

	entry := &domain.AuditEntry{
		ID:           id,
		EventType:    eventType,
		Action:       action,
		ResourceType: resourceType,
		ResourceID:   resourceID,
		TraceID:      traceID,
		CreatedAt:    createdAt,
	}

	if actorID != nil {
		entry.ActorID = *actorID
	}
	if actorEmail != nil {
		entry.ActorEmail = *actorEmail
	}
	if ipAddr != nil {
		entry.IPAddress = *ipAddr
	}
	if userAgent != nil {
		entry.UserAgent = *userAgent
	}
	if resourceChanges != nil {
		entry.ResourceChanges = string(resourceChanges)
	}
	if metadata != nil {
		json.Unmarshal(metadata, &entry.Metadata)
	}

	return entry, nil
}

func nullIfEmpty(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}
