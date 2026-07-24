package postgres

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/schemahub/backend/internal/drift/domain"
)

type DriftRepository struct {
	pool *pgxpool.Pool
}

func NewDriftRepository(pool *pgxpool.Pool) *DriftRepository {
	return &DriftRepository{pool: pool}
}

func (r *DriftRepository) Insert(ctx context.Context, event *domain.DriftEvent) error {
	_, err := r.pool.Exec(ctx, `
		INSERT INTO drift_events (connection_id, schema_id, expected_version_id, drift_type, object_type, object_name, expected_definition, actual_definition, diff_summary, severity, status)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)`,
		event.ConnectionID, nullIfEmpty(event.SchemaID), nullIfEmpty(event.ExpectedVersionID),
		string(event.DriftType), event.ObjectType, event.ObjectName,
		nullIfEmpty(event.ExpectedDefinition), nullIfEmpty(event.ActualDefinition),
		nullIfEmpty(event.DiffSummary), string(event.Severity), string(event.Status),
	)
	if err != nil {
		return fmt.Errorf("inserting drift event: %w", err)
	}
	return nil
}

func (r *DriftRepository) GetByID(ctx context.Context, id string) (*domain.DriftEvent, error) {
	row := r.pool.QueryRow(ctx, `
		SELECT id, connection_id, schema_id, expected_version_id, drift_type, object_type, object_name,
		       expected_definition, actual_definition, diff_summary, severity, status, detected_at, resolved_at, resolved_by
		FROM drift_events WHERE id = $1`, id)

	return scanEvent(row)
}

func (r *DriftRepository) List(ctx context.Context, filter *domain.DriftFilter, cursor string, limit int32) ([]*domain.DriftEvent, string, int32, error) {
	where := "WHERE 1=1"
	args := []interface{}{}
	argIdx := 1

	if filter.ConnectionID != "" {
		where += fmt.Sprintf(" AND connection_id = $%d", argIdx)
		args = append(args, filter.ConnectionID)
		argIdx++
	}
	if filter.Status != "" {
		where += fmt.Sprintf(" AND status = $%d", argIdx)
		args = append(args, filter.Status)
		argIdx++
	}
	if filter.Severity != "" {
		where += fmt.Sprintf(" AND severity = $%d", argIdx)
		args = append(args, filter.Severity)
		argIdx++
	}
	if filter.DriftType != "" {
		where += fmt.Sprintf(" AND drift_type = $%d", argIdx)
		args = append(args, filter.DriftType)
		argIdx++
	}
	if cursor != "" {
		where += fmt.Sprintf(" AND detected_at < (SELECT detected_at FROM drift_events WHERE id = $%d)", argIdx)
		args = append(args, cursor)
		argIdx++
	}

	args = append(args, limit+1)
	query := fmt.Sprintf(`
		SELECT id, connection_id, schema_id, expected_version_id, drift_type, object_type, object_name,
		       expected_definition, actual_definition, diff_summary, severity, status, detected_at, resolved_at, resolved_by
		FROM drift_events %s ORDER BY detected_at DESC LIMIT $%d`, where, argIdx)

	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, "", 0, fmt.Errorf("listing drift events: %w", err)
	}
	defer rows.Close()

	var events []*domain.DriftEvent
	for rows.Next() {
		event, err := scanEvent(rows)
		if err != nil {
			return nil, "", 0, err
		}
		events = append(events, event)
	}

	var nextCursor string
	if len(events) > int(limit) {
		events = events[:limit]
		nextCursor = events[len(events)-1].ID
	}

	var total int32
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM drift_events %s", where)
	countArgs := args[:len(args)-1]
	if err := r.pool.QueryRow(ctx, countQuery, countArgs...).Scan(&total); err != nil {
		return nil, "", 0, fmt.Errorf("counting drift events: %w", err)
	}

	return events, nextCursor, total, nil
}

func (r *DriftRepository) UpdateStatus(ctx context.Context, id string, status domain.DriftStatus, resolvedBy string) error {
	_, err := r.pool.Exec(ctx, `
		UPDATE drift_events SET status = $1, resolved_at = now(), resolved_by = $2 WHERE id = $3`,
		string(status), nullIfEmpty(resolvedBy), id)
	if err != nil {
		return fmt.Errorf("updating drift event status: %w", err)
	}
	return nil
}

func (r *DriftRepository) GetStats(ctx context.Context, connectionID string) (*domain.DriftStats, error) {
	where := "WHERE 1=1"
	args := []interface{}{}
	if connectionID != "" {
		where += " AND connection_id = $1"
		args = append(args, connectionID)
	}

	stats := &domain.DriftStats{
		BySeverity:  make(map[string]int32),
		ByDriftType: make(map[string]int32),
	}

	if err := r.pool.QueryRow(ctx, fmt.Sprintf("SELECT COUNT(*) FILTER (WHERE status='open') FROM drift_events %s", where), args...).Scan(&stats.TotalOpen); err != nil {
		return nil, fmt.Errorf("counting open: %w", err)
	}
	if err := r.pool.QueryRow(ctx, fmt.Sprintf("SELECT COUNT(*) FILTER (WHERE status='resolved') FROM drift_events %s", where), args...).Scan(&stats.TotalResolved); err != nil {
		return nil, fmt.Errorf("counting resolved: %w", err)
	}
	if err := r.pool.QueryRow(ctx, fmt.Sprintf("SELECT COUNT(*) FILTER (WHERE status='acknowledged') FROM drift_events %s", where), args...).Scan(&stats.TotalAcknowledged); err != nil {
		return nil, fmt.Errorf("counting acknowledged: %w", err)
	}
	if err := r.pool.QueryRow(ctx, fmt.Sprintf("SELECT COUNT(*) FILTER (WHERE status='false_positive') FROM drift_events %s", where), args...).Scan(&stats.TotalFalsePositive); err != nil {
		return nil, fmt.Errorf("counting false positive: %w", err)
	}

	rows, err := r.pool.Query(ctx, fmt.Sprintf("SELECT severity, COUNT(*) FROM drift_events %s GROUP BY severity", where), args...)
	if err != nil {
		return nil, fmt.Errorf("querying by severity: %w", err)
	}
	defer rows.Close()
	for rows.Next() {
		var sev string
		var cnt int32
		if err := rows.Scan(&sev, &cnt); err != nil {
			return nil, err
		}
		stats.BySeverity[sev] = cnt
	}

	rows2, err := r.pool.Query(ctx, fmt.Sprintf("SELECT drift_type, COUNT(*) FROM drift_events %s GROUP BY drift_type", where), args...)
	if err != nil {
		return nil, fmt.Errorf("querying by drift type: %w", err)
	}
	defer rows2.Close()
	for rows2.Next() {
		var dt string
		var cnt int32
		if err := rows2.Scan(&dt, &cnt); err != nil {
			return nil, err
		}
		stats.ByDriftType[dt] = cnt
	}

	return stats, nil
}

func scanEvent(row interface{ Scan(dest ...interface{}) error }) (*domain.DriftEvent, error) {
	var (
		id, driftType, objType, objName, severity, status string
		detectedAt                                        time.Time
		connID, schemaID, expVerID, expDef, actDef, diffSummary, resolvedBy *string
		resolvedAt                                        *time.Time
	)
	err := row.Scan(&id, &connID, &schemaID, &expVerID, &driftType, &objType, &objName,
		&expDef, &actDef, &diffSummary, &severity, &status, &detectedAt, &resolvedAt, &resolvedBy)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("drift event not found")
		}
		return nil, fmt.Errorf("scanning drift event: %w", err)
	}

	event := &domain.DriftEvent{
		ID:        id,
		DriftType: domain.DriftType(driftType),
		ObjectType: objType,
		ObjectName: objName,
		Severity:  domain.Severity(severity),
		Status:    domain.DriftStatus(status),
		DetectedAt: detectedAt,
	}
	if connID != nil {
		event.ConnectionID = *connID
	}
	if schemaID != nil {
		event.SchemaID = *schemaID
	}
	if expVerID != nil {
		event.ExpectedVersionID = *expVerID
	}
	if expDef != nil {
		event.ExpectedDefinition = *expDef
	}
	if actDef != nil {
		event.ActualDefinition = *actDef
	}
	if diffSummary != nil {
		event.DiffSummary = *diffSummary
	}
	if resolvedAt != nil {
		event.ResolvedAt = resolvedAt
	}
	if resolvedBy != nil {
		event.ResolvedBy = *resolvedBy
	}
	return event, nil
}

func nullIfEmpty(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}
