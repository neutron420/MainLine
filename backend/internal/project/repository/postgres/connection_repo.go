package postgres

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/schemahub/backend/internal/project/domain"
)

type ConnectionRepository struct {
	db *pgxpool.Pool
}

func NewConnectionRepository(db *pgxpool.Pool) *ConnectionRepository {
	return &ConnectionRepository{db: db}
}

func (r *ConnectionRepository) Create(ctx context.Context, c *domain.Connection) error {
	c.ID = uuid.NewString()
	c.CreatedAt = time.Now()
	c.UpdatedAt = time.Now()

	_, err := r.db.Exec(ctx,
		`INSERT INTO connections (id, project_id, name, host, port, database_name, username, password_encrypted, ssl_mode, connection_status, created_by, created_at, updated_at)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)`,
		c.ID, c.ProjectID, c.Name, c.Host, c.Port, c.DatabaseName, c.Username, c.PasswordEncrypted, string(c.SSLMode), string(c.ConnectionStatus), c.CreatedBy, c.CreatedAt, c.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("inserting connection: %w", err)
	}
	return nil
}

func (r *ConnectionRepository) GetByID(ctx context.Context, id string) (*domain.Connection, error) {
	row := r.db.QueryRow(ctx,
		`SELECT id, project_id, name, host, port, database_name, username, password_encrypted, ssl_mode, connection_status, last_connected_at, created_by, created_at, updated_at, deleted_at
		 FROM connections WHERE id = $1 AND deleted_at IS NULL`, id,
	)

	c := &domain.Connection{}
	var deletedAt *time.Time
	if err := row.Scan(&c.ID, &c.ProjectID, &c.Name, &c.Host, &c.Port, &c.DatabaseName, &c.Username, &c.PasswordEncrypted, &c.SSLMode, &c.ConnectionStatus, &c.LastConnectedAt, &c.CreatedBy, &c.CreatedAt, &c.UpdatedAt, &deletedAt); err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("connection not found")
		}
		return nil, fmt.Errorf("scanning connection: %w", err)
	}
	return c, nil
}

func (r *ConnectionRepository) ListByProjectID(ctx context.Context, projectID, cursor string, limit int32) ([]*domain.Connection, string, int32, error) {
	query := `SELECT id, project_id, name, host, port, database_name, username, password_encrypted, ssl_mode, connection_status, last_connected_at, created_by, created_at, updated_at, deleted_at
		FROM connections WHERE project_id = $1 AND deleted_at IS NULL ORDER BY created_at DESC`

	var args []interface{}
	args = append(args, projectID)

	if cursor != "" {
		query = `SELECT id, project_id, name, host, port, database_name, username, password_encrypted, ssl_mode, connection_status, last_connected_at, created_by, created_at, updated_at, deleted_at
			FROM connections WHERE project_id = $1 AND deleted_at IS NULL AND created_at < $2 ORDER BY created_at DESC`
		args = append(args, cursor)
	}

	if limit <= 0 {
		limit = 20
	}
	query += fmt.Sprintf(" LIMIT %d", limit+1)

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, "", 0, fmt.Errorf("querying connections: %w", err)
	}
	defer rows.Close()

	var conns []*domain.Connection
	count := int32(0)
	for rows.Next() {
		c := &domain.Connection{}
		var deletedAt *time.Time
		if err := rows.Scan(&c.ID, &c.ProjectID, &c.Name, &c.Host, &c.Port, &c.DatabaseName, &c.Username, &c.PasswordEncrypted, &c.SSLMode, &c.ConnectionStatus, &c.LastConnectedAt, &c.CreatedBy, &c.CreatedAt, &c.UpdatedAt, &deletedAt); err != nil {
			return nil, "", 0, fmt.Errorf("scanning connection: %w", err)
		}
		conns = append(conns, c)
		count++
	}

	var nextCursor string
	if int32(len(conns)) > limit {
		nextCursor = conns[len(conns)-1].CreatedAt.Format(time.RFC3339Nano)
		conns = conns[:len(conns)-1]
	}

	return conns, nextCursor, count, nil
}

func (r *ConnectionRepository) Update(ctx context.Context, c *domain.Connection) error {
	c.UpdatedAt = time.Now()
	_, err := r.db.Exec(ctx,
		`UPDATE connections SET name = $1, host = $2, port = $3, database_name = $4, username = $5, password_encrypted = $6, ssl_mode = $7, updated_at = $8 WHERE id = $9 AND deleted_at IS NULL`,
		c.Name, c.Host, c.Port, c.DatabaseName, c.Username, c.PasswordEncrypted, string(c.SSLMode), c.UpdatedAt, c.ID,
	)
	if err != nil {
		return fmt.Errorf("updating connection: %w", err)
	}
	return nil
}

func (r *ConnectionRepository) SoftDelete(ctx context.Context, id string) error {
	_, err := r.db.Exec(ctx, `UPDATE connections SET deleted_at = NOW() WHERE id = $1 AND deleted_at IS NULL`, id)
	if err != nil {
		return fmt.Errorf("soft deleting connection: %w", err)
	}
	return nil
}

func (r *ConnectionRepository) UpdateStatus(ctx context.Context, id string, status domain.ConnectionStatus, lastConnectedAt *time.Time) error {
	_, err := r.db.Exec(ctx,
		`UPDATE connections SET connection_status = $1, last_connected_at = $2, updated_at = NOW() WHERE id = $3`,
		string(status), lastConnectedAt, id,
	)
	if err != nil {
		return fmt.Errorf("updating connection status: %w", err)
	}
	return nil
}
