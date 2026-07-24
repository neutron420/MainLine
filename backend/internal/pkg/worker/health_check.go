package worker

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/schemahub/backend/internal/project/domain"
	"github.com/schemahub/backend/pkg/encryption"
)

type ConnectionHealthWorker struct {
	connRepo      domain.ConnectionRepository
	encryptionKey  []byte
}

func NewConnectionHealthWorker(connRepo domain.ConnectionRepository, encryptionKey []byte) *ConnectionHealthWorker {
	return &ConnectionHealthWorker{connRepo: connRepo, encryptionKey: encryptionKey}
}

func (w *ConnectionHealthWorker) Name() string {
	return "connection-health-check"
}

func (w *ConnectionHealthWorker) Interval() time.Duration {
	return 5 * time.Minute
}

func (w *ConnectionHealthWorker) Run(ctx context.Context) error {
	connections, err := w.connRepo.ListAll(ctx)
	if err != nil {
		return fmt.Errorf("listing connections: %w", err)
	}

	for _, c := range connections {
		if c.DeletedAt != nil {
			continue
		}

		decrypted, err := encryption.Decrypt(c.PasswordEncrypted, w.encryptionKey)
		if err != nil {
			w.connRepo.UpdateStatus(ctx, c.ID, domain.ConnStatusFailed, nil)
			continue
		}

		connStr := fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=%s",
			c.Username, string(decrypted), c.Host, c.Port, c.DatabaseName, string(c.SSLMode))

		pool, err := pgxpool.New(ctx, connStr)
		if err != nil {
			w.connRepo.UpdateStatus(ctx, c.ID, domain.ConnStatusFailed, nil)
			continue
		}

		var version string
		if err := pool.QueryRow(ctx, "SELECT version()").Scan(&version); err != nil {
			pool.Close()
			w.connRepo.UpdateStatus(ctx, c.ID, domain.ConnStatusFailed, nil)
			continue
		}

		pool.Close()
		now := time.Now()
		w.connRepo.UpdateStatus(ctx, c.ID, domain.ConnStatusConnected, &now)
	}

	return nil
}
