package postgres

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/schemahub/backend/internal/migration/domain"
	"github.com/schemahub/backend/internal/pkg/pagination"
)

type MigrationRepository struct {
	db *pgxpool.Pool
}

func NewMigrationRepository(db *pgxpool.Pool) *MigrationRepository {
	return &MigrationRepository{db: db}
}

func (r *MigrationRepository) Create(ctx context.Context, m *domain.Migration) error {
	m.ID = uuid.NewString()
	m.CreatedAt = time.Now()
	m.UpdatedAt = time.Now()

	_, err := r.db.Exec(ctx,
		`INSERT INTO migrations (id, project_id, title, description, version, up_sql, down_sql, checksum, status, created_by, created_at, updated_at)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)`,
		m.ID, m.ProjectID, m.Title, m.Description, m.Version, m.UpSQL, m.DownSQL,
		m.Checksum, string(m.Status), m.CreatedBy, m.CreatedAt, m.UpdatedAt)
	return err
}

func (r *MigrationRepository) GetByID(ctx context.Context, id string) (*domain.Migration, error) {
	row := r.db.QueryRow(ctx,
		`SELECT id, project_id, title, description, version, up_sql, down_sql, checksum, status, created_by, created_at, updated_at, deleted_at
		 FROM migrations WHERE id = $1 AND deleted_at IS NULL`, id)

	m := &domain.Migration{}
	var status string
	if err := row.Scan(&m.ID, &m.ProjectID, &m.Title, &m.Description, &m.Version,
		&m.UpSQL, &m.DownSQL, &m.Checksum, &status, &m.CreatedBy,
		&m.CreatedAt, &m.UpdatedAt, &m.DeletedAt); err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("migration not found")
		}
		return nil, err
	}
	m.Status = domain.MigrationStatus(status)
	return m, nil
}

func (r *MigrationRepository) ListByProjectID(ctx context.Context, projectID, cursor string, limit int32) ([]*domain.Migration, string, int32, error) {
	query := `SELECT id, project_id, title, description, version, up_sql, down_sql, checksum, status, created_by, created_at, updated_at, deleted_at
		FROM migrations WHERE project_id = $1 AND deleted_at IS NULL ORDER BY created_at DESC`

	var args []interface{}
	args = append(args, projectID)

	if cursor != "" {
		ts, id, ok := pagination.Decode(cursor)
		if !ok {
			return nil, "", 0, fmt.Errorf("invalid migration cursor")
		}
		query = `SELECT id, project_id, title, description, version, up_sql, down_sql, checksum, status, created_by, created_at, updated_at, deleted_at
			FROM migrations WHERE project_id = $1 AND deleted_at IS NULL AND (created_at, id) < ($2::timestamptz, $3)
			ORDER BY created_at DESC, id DESC`
		args = append(args, ts, id)
	}

	if limit <= 0 {
		limit = 20
	}
	query += fmt.Sprintf(" LIMIT %d", limit+1)

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, "", 0, err
	}
	defer rows.Close()

	var migrations []*domain.Migration
	for rows.Next() {
		m := &domain.Migration{}
		var status string
		if err := rows.Scan(&m.ID, &m.ProjectID, &m.Title, &m.Description, &m.Version,
			&m.UpSQL, &m.DownSQL, &m.Checksum, &status, &m.CreatedBy,
			&m.CreatedAt, &m.UpdatedAt, &m.DeletedAt); err != nil {
			return nil, "", 0, err
		}
		m.Status = domain.MigrationStatus(status)
		migrations = append(migrations, m)
	}

	var nextCursor string
	if int32(len(migrations)) > limit {
		migrations = migrations[:len(migrations)-1]
		nextCursor = pagination.Encode(migrations[len(migrations)-1].CreatedAt, migrations[len(migrations)-1].ID)
	}
	return migrations, nextCursor, int32(len(migrations)), nil
}

func (r *MigrationRepository) Update(ctx context.Context, m *domain.Migration) error {
	m.UpdatedAt = time.Now()
	_, err := r.db.Exec(ctx,
		`UPDATE migrations SET title=$1, description=$2, version=$3, up_sql=$4, down_sql=$5,
		 checksum=$6, status=$7, updated_at=$8 WHERE id=$9 AND deleted_at IS NULL`,
		m.Title, m.Description, m.Version, m.UpSQL, m.DownSQL,
		m.Checksum, string(m.Status), m.UpdatedAt, m.ID)
	return err
}

func (r *MigrationRepository) SoftDelete(ctx context.Context, id string) error {
	_, err := r.db.Exec(ctx, `UPDATE migrations SET deleted_at = now() WHERE id = $1`, id)
	return err
}

func (r *MigrationRepository) GetByProjectAndVersion(ctx context.Context, projectID, version string) (*domain.Migration, error) {
	row := r.db.QueryRow(ctx,
		`SELECT id, project_id, title, description, version, up_sql, down_sql, checksum, status, created_by, created_at, updated_at, deleted_at
		 FROM migrations WHERE project_id = $1 AND version = $2 AND deleted_at IS NULL`, projectID, version)

	m := &domain.Migration{}
	var status string
	if err := row.Scan(&m.ID, &m.ProjectID, &m.Title, &m.Description, &m.Version,
		&m.UpSQL, &m.DownSQL, &m.Checksum, &status, &m.CreatedBy,
		&m.CreatedAt, &m.UpdatedAt, &m.DeletedAt); err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("migration not found")
		}
		return nil, err
	}
	m.Status = domain.MigrationStatus(status)
	return m, nil
}

func (r *MigrationRepository) CreateRun(ctx context.Context, run *domain.MigrationRun) error {
	run.ID = uuid.NewString()
	run.CreatedAt = time.Now()

	_, err := r.db.Exec(ctx,
		`INSERT INTO migration_runs (id, migration_id, connection_id, direction, status, executed_by, created_at)
		 VALUES ($1, $2, $3, $4, $5, $6, $7)`,
		run.ID, run.MigrationID, run.ConnectionID, string(run.Direction),
		string(run.Status), run.ExecutedBy, run.CreatedAt)
	return err
}

type rowScanner interface {
	Scan(dest ...any) error
}

func scanRun(r rowScanner) (*domain.MigrationRun, error) {
	run := &domain.MigrationRun{}
	var status, direction string
	var startedAt, completedAt *time.Time
	var durationMs *int32
	var errorMessage *string
	if err := r.Scan(&run.ID, &run.MigrationID, &run.ConnectionID, &direction, &status,
		&startedAt, &completedAt, &durationMs, &errorMessage,
		&run.ExecutedBy, &run.CreatedAt); err != nil {
		return nil, err
	}
	run.StartedAt = startedAt
	run.CompletedAt = completedAt
	if durationMs != nil {
		run.DurationMs = *durationMs
	}
	if errorMessage != nil {
		run.ErrorMessage = *errorMessage
	}
	run.Status = domain.RunStatus(status)
	run.Direction = domain.MigrationDirection(direction)
	return run, nil
}

func (r *MigrationRepository) GetRunByID(ctx context.Context, id string) (*domain.MigrationRun, error) {
	row := r.db.QueryRow(ctx,
		`SELECT id, migration_id, connection_id, direction, status, started_at, completed_at, duration_ms, error_message, executed_by, created_at
		 FROM migration_runs WHERE id = $1`, id)

	run, err := scanRun(row)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("run not found")
		}
		return nil, err
	}
	return run, nil
}

func (r *MigrationRepository) UpdateRun(ctx context.Context, run *domain.MigrationRun) error {
	_, err := r.db.Exec(ctx,
		`UPDATE migration_runs SET status=$1, started_at=$2, completed_at=$3, duration_ms=$4, error_message=$5
		 WHERE id=$6`,
		string(run.Status), run.StartedAt, run.CompletedAt, run.DurationMs, run.ErrorMessage, run.ID)
	return err
}

func (r *MigrationRepository) ListRunsByMigrationID(ctx context.Context, migrationID, cursor string, limit int32) ([]*domain.MigrationRun, string, int32, error) {
	query := `SELECT id, migration_id, connection_id, direction, status, started_at, completed_at, duration_ms, error_message, executed_by, created_at
		FROM migration_runs WHERE migration_id = $1 ORDER BY created_at DESC`

	var args []interface{}
	args = append(args, migrationID)

	if cursor != "" {
		ts, id, ok := pagination.Decode(cursor)
		if !ok {
			return nil, "", 0, fmt.Errorf("invalid migration run cursor")
		}
		query = `SELECT id, migration_id, connection_id, direction, status, started_at, completed_at, duration_ms, error_message, executed_by, created_at
			FROM migration_runs WHERE migration_id = $1 AND (created_at, id) < ($2::timestamptz, $3)
			ORDER BY created_at DESC, id DESC`
		args = append(args, ts, id)
	}

	if limit <= 0 {
		limit = 20
	}
	query += fmt.Sprintf(" LIMIT %d", limit+1)

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, "", 0, err
	}
	defer rows.Close()

	var runs []*domain.MigrationRun
	for rows.Next() {
		run, err := scanRun(rows)
		if err != nil {
			return nil, "", 0, err
		}
		runs = append(runs, run)
	}

	var nextCursor string
	if int32(len(runs)) > limit {
		runs = runs[:len(runs)-1]
		nextCursor = pagination.Encode(runs[len(runs)-1].CreatedAt, runs[len(runs)-1].ID)
	}
	return runs, nextCursor, int32(len(runs)), nil
}

func (r *MigrationRepository) GetActiveRunForConnection(ctx context.Context, connectionID string) (*domain.MigrationRun, error) {
	row := r.db.QueryRow(ctx,
		`SELECT id, migration_id, connection_id, direction, status, started_at, completed_at, duration_ms, error_message, executed_by, created_at
		 FROM migration_runs WHERE connection_id = $1 AND status IN ('pending', 'running') LIMIT 1`, connectionID)

	run, err := scanRun(row)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return run, nil
}

func (r *MigrationRepository) CreateLogEntry(ctx context.Context, entry *domain.MigrationLogEntry) error {
	entry.ID = uuid.NewString()
	entry.CreatedAt = time.Now()

	_, err := r.db.Exec(ctx,
		`INSERT INTO migration_logs (id, migration_run_id, sequence, sql, duration_ms, rows_affected, error_message, created_at)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`,
		entry.ID, entry.MigrationRunID, entry.Sequence, entry.SQL,
		entry.DurationMs, entry.RowsAffected, entry.ErrorMessage, entry.CreatedAt)
	return err
}

func (r *MigrationRepository) ListLogsByRunID(ctx context.Context, runID string) ([]*domain.MigrationLogEntry, error) {
	rows, err := r.db.Query(ctx,
		`SELECT id, migration_run_id, sequence, sql, duration_ms, rows_affected, error_message, created_at
		 FROM migration_logs WHERE migration_run_id = $1 ORDER BY sequence`, runID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var entries []*domain.MigrationLogEntry
	for rows.Next() {
		e := &domain.MigrationLogEntry{}
		if err := rows.Scan(&e.ID, &e.MigrationRunID, &e.Sequence, &e.SQL,
			&e.DurationMs, &e.RowsAffected, &e.ErrorMessage, &e.CreatedAt); err != nil {
			return nil, err
		}
		entries = append(entries, e)
	}
	return entries, nil
}
