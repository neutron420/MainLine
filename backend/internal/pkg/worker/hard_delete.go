package worker

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type HardDeleteWorker struct {
	pool *pgxpool.Pool
}

func NewHardDeleteWorker(pool *pgxpool.Pool) *HardDeleteWorker {
	return &HardDeleteWorker{pool: pool}
}

func (w *HardDeleteWorker) Name() string {
	return "hard-delete-cleanup"
}

func (w *HardDeleteWorker) Interval() time.Duration {
	return 24 * time.Hour
}

func (w *HardDeleteWorker) Run(ctx context.Context) error {
	tables := []string{"users", "projects", "connections", "schemas", "migrations"}
	cutoff := time.Now().Add(-90 * 24 * time.Hour)

	for _, table := range tables {
		query := fmt.Sprintf(`DELETE FROM %s WHERE deleted_at IS NOT NULL AND deleted_at < $1`, table)
		tag, err := w.pool.Exec(ctx, query, cutoff)
		if err != nil {
			return fmt.Errorf("hard deleting from %s: %w", table, err)
		}
		if tag.RowsAffected() > 0 {
		}
	}

	return nil
}
