package domain

import (
	"context"
	"crypto/sha256"
	"fmt"
	"time"
)

type MigrationStatus string

const (
	MigrationStatusDraft      MigrationStatus = "draft"
	MigrationStatusPending    MigrationStatus = "pending"
	MigrationStatusRunning    MigrationStatus = "running"
	MigrationStatusCompleted  MigrationStatus = "completed"
	MigrationStatusFailed     MigrationStatus = "failed"
	MigrationStatusRolledBack MigrationStatus = "rolled_back"
)

type MigrationDirection string

const (
	MigrationDirectionUp   MigrationDirection = "up"
	MigrationDirectionDown MigrationDirection = "down"
)

type RunStatus string

const (
	RunStatusPending    RunStatus = "pending"
	RunStatusRunning    RunStatus = "running"
	RunStatusCompleted  RunStatus = "completed"
	RunStatusFailed     RunStatus = "failed"
	RunStatusRolledBack RunStatus = "rolled_back"
)

type Migration struct {
	ID          string
	ProjectID   string
	Title       string
	Description string
	Version     string
	UpSQL       string
	DownSQL     string
	Checksum    string
	Status      MigrationStatus
	CreatedBy   string
	CreatedAt   time.Time
	UpdatedAt   time.Time
	DeletedAt   *time.Time
}

type MigrationRun struct {
	ID           string
	MigrationID  string
	ConnectionID string
	Direction    MigrationDirection
	Status       RunStatus
	StartedAt    *time.Time
	CompletedAt  *time.Time
	DurationMs   int32
	ErrorMessage string
	ExecutedBy   string
	CreatedAt    time.Time
}

type MigrationLogEntry struct {
	ID             string
	MigrationRunID string
	Sequence       int
	SQL            string
	DurationMs     *int32
	RowsAffected   *int32
	ErrorMessage   string
	CreatedAt      time.Time
}

type MigrationStatusMessage struct {
	RunID               string
	State               RunStatus
	TotalStatements     int32
	CompletedStatements int32
	CurrentStatement    string
	StartedAt           time.Time
	ElapsedMs           int64
	ErrorMessage        string
	LastLog             *MigrationLogEntry
}

type ValidationError struct {
	Line    int
	Column  int
	Message string
}

func ComputeChecksum(sql string) string {
	h := sha256.Sum256([]byte(sql))
	return fmt.Sprintf("%x", h)
}

func (m *Migration) Validate() error {
	if m.Title == "" {
		return fmt.Errorf("title is required")
	}
	if m.Version == "" {
		return fmt.Errorf("version is required")
	}
	if m.UpSQL == "" {
		return fmt.Errorf("up_sql is required")
	}
	return nil
}

// ── Repository Interface ──

type MigrationRepository interface {
	Create(ctx context.Context, m *Migration) error
	GetByID(ctx context.Context, id string) (*Migration, error)
	ListByProjectID(ctx context.Context, projectID, cursor string, limit int32) ([]*Migration, string, int32, error)
	Update(ctx context.Context, m *Migration) error
	SoftDelete(ctx context.Context, id string) error
	GetByProjectAndVersion(ctx context.Context, projectID, version string) (*Migration, error)

	CreateRun(ctx context.Context, r *MigrationRun) error
	GetRunByID(ctx context.Context, id string) (*MigrationRun, error)
	UpdateRun(ctx context.Context, r *MigrationRun) error
	ListRunsByMigrationID(ctx context.Context, migrationID, cursor string, limit int32) ([]*MigrationRun, string, int32, error)
	GetActiveRunForConnection(ctx context.Context, connectionID string) (*MigrationRun, error)

	CreateLogEntry(ctx context.Context, entry *MigrationLogEntry) error
	ListLogsByRunID(ctx context.Context, runID string) ([]*MigrationLogEntry, error)
}
