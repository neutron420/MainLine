package worker

import (
	"context"
	"fmt"
	"time"

	"github.com/schemahub/backend/internal/drift/domain"
	projDomain "github.com/schemahub/backend/internal/project/domain"
	"github.com/schemahub/backend/pkg/encryption"
)

// DriftCheckWorker periodically runs drift detection for every active connection.
type DriftCheckWorker struct {
	connRepo      projDomain.ConnectionRepository
	driftSvc      *domain.DriftService
	encryptionKey []byte
}

func NewDriftCheckWorker(connRepo projDomain.ConnectionRepository, driftSvc *domain.DriftService, encryptionKey []byte) *DriftCheckWorker {
	return &DriftCheckWorker{connRepo: connRepo, driftSvc: driftSvc, encryptionKey: encryptionKey}
}

func (w *DriftCheckWorker) Name() string {
	return "drift-check"
}

func (w *DriftCheckWorker) Interval() time.Duration {
	return 10 * time.Minute
}

func (w *DriftCheckWorker) Run(ctx context.Context) error {
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
			continue
		}

		connStr := fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=%s",
			c.Username, string(decrypted), c.Host, c.Port, c.DatabaseName, string(c.SSLMode))

		if _, err := w.driftSvc.CheckDrift(ctx, connStr, c.ID, []string{"public"}); err != nil {
			continue
		}
	}

	return nil
}
