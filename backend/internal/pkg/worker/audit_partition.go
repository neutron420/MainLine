package worker

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

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

func (w *AuditPartitionWorker) Run(ctx context.Context) error {
	now := time.Now().UTC()

	nextMonth := time.Date(now.Year(), now.Month()+2, 1, 0, 0, 0, 0, time.UTC)
	if now.Month() == 12 {
		nextMonth = time.Date(now.Year()+1, 1, 1, 0, 0, 0, 0, time.UTC)
	}

	partitionName := fmt.Sprintf("audit_logs_%04d_%02d", nextMonth.Year(), nextMonth.Month())
	fromDate := nextMonth.Format("2006-01-02")
	toDate := time.Date(nextMonth.Year(), nextMonth.Month()+1, 1, 0, 0, 0, 0, time.UTC).Format("2006-01-02")

	query := fmt.Sprintf(`
		CREATE TABLE IF NOT EXISTS %s PARTITION OF audit_logs
		FOR VALUES FROM ('%s') TO ('%s')`,
		partitionName, fromDate, toDate)

	_, err := w.pool.Exec(ctx, query)
	if err != nil {
		return fmt.Errorf("creating partition %s: %w", partitionName, err)
	}

	return nil
}
