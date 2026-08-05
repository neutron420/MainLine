package worker

import (
	"context"
	"fmt"
	"regexp"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

var partitionBoundRe = regexp.MustCompile(`FROM \('([^']+)'\) TO \('([^']+)'\)`)

// parsePartitionRange extracts the [start, end) range from a partition bound
// expression such as "FOR VALUES FROM ('2026-10-01 00:00:00+00') TO ('2026-11-01 00:00:00+00')".
func parsePartitionRange(bound string) (start, end time.Time, ok bool) {
	m := partitionBoundRe.FindStringSubmatch(bound)
	if len(m) != 3 {
		return time.Time{}, time.Time{}, false
	}
	const layout = "2006-01-02 15:04:05-07"
	s, err1 := time.Parse(layout, m[1])
	e, err2 := time.Parse(layout, m[2])
	if err1 != nil || err2 != nil {
		return time.Time{}, time.Time{}, false
	}
	return s, e, true
}

type AuditPartitionWorker struct {
	pool *pgxpool.Pool
}

func NewAuditPartitionWorker(pool *pgxpool.Pool) *AuditPartitionWorker {
	return &AuditPartitionWorker{pool: pool}
}

func (w *AuditPartitionWorker) Name() string {
	return "audit-partition-manager"
}

func (w *AuditPartitionWorker) Interval() time.Duration {
	return 1 * time.Hour
}

// Run creates the audit_logs partition for the month after next, skipping
// creation when an existing partition already covers that range (e.g. a
// forward-looking buffer partition created by migrations).
func (w *AuditPartitionWorker) Run(ctx context.Context) error {
	now := time.Now().UTC()

	nextMonth := time.Date(now.Year(), now.Month()+2, 1, 0, 0, 0, 0, time.UTC)

	partitionName := fmt.Sprintf("audit_logs_%04d_%02d", nextMonth.Year(), nextMonth.Month())
	fromDate := nextMonth.Format("2006-01-02")
	toDate := nextMonth.AddDate(0, 1, 0).Format("2006-01-02")

	covered, err := w.rangeCovered(ctx, nextMonth, nextMonth.AddDate(0, 1, 0))
	if err != nil {
		return fmt.Errorf("checking partition coverage for %s: %w", partitionName, err)
	}
	if covered {
		return nil
	}

	query := fmt.Sprintf(`
		CREATE TABLE IF NOT EXISTS %s PARTITION OF audit_logs
		FOR VALUES FROM ('%s') TO ('%s')`,
		partitionName, fromDate, toDate)

	_, err = w.pool.Exec(ctx, query)
	if err != nil {
		return fmt.Errorf("creating partition %s: %w", partitionName, err)
	}

	return nil
}

// rangeCovered reports whether any existing audit_logs partition overlaps the
// [from, to) range.
func (w *AuditPartitionWorker) rangeCovered(ctx context.Context, from, to time.Time) (bool, error) {
	rows, err := w.pool.Query(ctx, `
		SELECT pg_get_expr(c.relpartbound, c.oid)
		FROM pg_class c
		JOIN pg_inherits i ON i.inhrelid = c.oid
		JOIN pg_class p ON p.oid = i.inhparent AND p.relname = 'audit_logs'
		JOIN pg_namespace n ON n.oid = c.relnamespace AND n.nspname = 'public'
		WHERE c.relkind = 'r' AND c.relispartition`)
	if err != nil {
		return false, err
	}
	defer rows.Close()

	for rows.Next() {
		var bound string
		if err := rows.Scan(&bound); err != nil {
			return false, err
		}
		start, end, ok := parsePartitionRange(bound)
		if !ok {
			continue
		}
		if start.Before(to) && end.After(from) {
			return true, nil
		}
	}
	return false, rows.Err()
}
