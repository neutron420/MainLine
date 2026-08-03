package postgres

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/schemahub/backend/internal/pkg/pagination"
	"github.com/schemahub/backend/internal/project/domain"
)

type ProjectRepository struct {
	db *pgxpool.Pool
}

func NewProjectRepository(db *pgxpool.Pool) *ProjectRepository {
	return &ProjectRepository{db: db}
}

func (r *ProjectRepository) Create(ctx context.Context, p *domain.Project) error {
	p.ID = uuid.NewString()
	p.CreatedAt = time.Now()
	p.UpdatedAt = time.Now()

	_, err := r.db.Exec(ctx,
		`INSERT INTO projects (id, name, slug, description, visibility, template, created_by, created_at, updated_at)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)`,
		p.ID, p.Name, p.Slug, p.Description, string(p.Visibility), p.Template, p.CreatedBy, p.CreatedAt, p.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("inserting project: %w", err)
	}
	return nil
}

func (r *ProjectRepository) GetByID(ctx context.Context, id string) (*domain.Project, error) {
	row := r.db.QueryRow(ctx,
		`SELECT id, name, slug, description, visibility, template, created_by, created_at, updated_at, deleted_at
		 FROM projects WHERE id = $1 AND deleted_at IS NULL`, id,
	)

	p := &domain.Project{}
	var deletedAt *time.Time
	if err := row.Scan(&p.ID, &p.Name, &p.Slug, &p.Description, &p.Visibility, &p.Template, &p.CreatedBy, &p.CreatedAt, &p.UpdatedAt, &deletedAt); err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("project not found")
		}
		return nil, fmt.Errorf("scanning project: %w", err)
	}
	return p, nil
}

func (r *ProjectRepository) GetBySlug(ctx context.Context, slug string) (*domain.Project, error) {
	row := r.db.QueryRow(ctx,
		`SELECT id, name, slug, description, visibility, template, created_by, created_at, updated_at, deleted_at
		 FROM projects WHERE slug = $1 AND deleted_at IS NULL`, slug,
	)

	p := &domain.Project{}
	var deletedAt *time.Time
	if err := row.Scan(&p.ID, &p.Name, &p.Slug, &p.Description, &p.Visibility, &p.Template, &p.CreatedBy, &p.CreatedAt, &p.UpdatedAt, &deletedAt); err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("project not found")
		}
		return nil, fmt.Errorf("scanning project: %w", err)
	}
	return p, nil
}

func (r *ProjectRepository) ListByUserID(ctx context.Context, userID, cursor string, limit int32) ([]*domain.Project, string, int32, error) {
	query := `SELECT p.id, p.name, p.slug, p.description, p.visibility, p.template, p.created_by, p.created_at, p.updated_at, p.deleted_at,
		(SELECT COUNT(*) FROM project_members pm WHERE pm.project_id = p.id) as member_count
		FROM projects p
		JOIN project_members pm ON p.id = pm.project_id
		WHERE pm.user_id = $1 AND p.deleted_at IS NULL
		ORDER BY p.updated_at DESC`

	var args []interface{}
	args = append(args, userID)

	if cursor != "" {
		ts, id, ok := pagination.Decode(cursor)
		if !ok {
			return nil, "", 0, fmt.Errorf("invalid project cursor")
		}
		query = `SELECT p.id, p.name, p.slug, p.description, p.visibility, p.template, p.created_by, p.created_at, p.updated_at, p.deleted_at,
			(SELECT COUNT(*) FROM project_members pm WHERE pm.project_id = p.id) as member_count
			FROM projects p
			JOIN project_members pm ON p.id = pm.project_id
			WHERE pm.user_id = $1 AND p.deleted_at IS NULL AND (p.updated_at, p.id) < ($2::timestamptz, $3)
			ORDER BY p.updated_at DESC, p.id DESC`
		args = append(args, ts, id)
	}

	if limit <= 0 {
		limit = 20
	}
	query += fmt.Sprintf(" LIMIT %d", limit+1)

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, "", 0, fmt.Errorf("querying projects: %w", err)
	}
	defer rows.Close()

	var projects []*domain.Project
	count := int32(0)
	for rows.Next() {
		p := &domain.Project{}
		var deletedAt *time.Time
		var memberCount int32
		if err := rows.Scan(&p.ID, &p.Name, &p.Slug, &p.Description, &p.Visibility, &p.Template, &p.CreatedBy, &p.CreatedAt, &p.UpdatedAt, &deletedAt, &memberCount); err != nil {
			return nil, "", 0, fmt.Errorf("scanning project: %w", err)
		}
		projects = append(projects, p)
		count++
	}

	var nextCursor string
	if int32(len(projects)) > limit {
		projects = projects[:len(projects)-1]
		nextCursor = pagination.Encode(projects[len(projects)-1].UpdatedAt, projects[len(projects)-1].ID)
	}

	return projects, nextCursor, count, nil
}

func (r *ProjectRepository) Update(ctx context.Context, p *domain.Project) error {
	p.UpdatedAt = time.Now()
	_, err := r.db.Exec(ctx,
		`UPDATE projects SET name = $1, description = $2, visibility = $3, updated_at = $4 WHERE id = $5 AND deleted_at IS NULL`,
		p.Name, p.Description, string(p.Visibility), p.UpdatedAt, p.ID,
	)
	if err != nil {
		return fmt.Errorf("updating project: %w", err)
	}
	return nil
}

func (r *ProjectRepository) SoftDelete(ctx context.Context, id string) error {
	_, err := r.db.Exec(ctx, `UPDATE projects SET deleted_at = NOW() WHERE id = $1 AND deleted_at IS NULL`, id)
	if err != nil {
		return fmt.Errorf("soft deleting project: %w", err)
	}
	return nil
}

func (r *ProjectRepository) AddMember(ctx context.Context, m *domain.ProjectMember) error {
	m.ID = uuid.NewString()
	m.CreatedAt = time.Now()
	_, err := r.db.Exec(ctx,
		`INSERT INTO project_members (id, project_id, user_id, role, created_at)
		 VALUES ($1, $2, $3, $4, $5) ON CONFLICT (project_id, user_id) DO UPDATE SET role = $4`,
		m.ID, m.ProjectID, m.UserID, string(m.Role), m.CreatedAt,
	)
	if err != nil {
		return fmt.Errorf("inserting member: %w", err)
	}
	return nil
}

func (r *ProjectRepository) RemoveMember(ctx context.Context, projectID, userID string) error {
	_, err := r.db.Exec(ctx,
		`DELETE FROM project_members WHERE project_id = $1 AND user_id = $2`,
		projectID, userID,
	)
	if err != nil {
		return fmt.Errorf("removing member: %w", err)
	}
	return nil
}

func (r *ProjectRepository) UpdateMemberRole(ctx context.Context, projectID, userID string, role domain.ProjectRole) error {
	_, err := r.db.Exec(ctx,
		`UPDATE project_members SET role = $1 WHERE project_id = $2 AND user_id = $3`,
		string(role), projectID, userID,
	)
	if err != nil {
		return fmt.Errorf("updating member role: %w", err)
	}
	return nil
}

func (r *ProjectRepository) GetMember(ctx context.Context, projectID, userID string) (*domain.ProjectMember, error) {
	row := r.db.QueryRow(ctx,
		`SELECT id, project_id, user_id, role, invited_by, joined_at, created_at
		 FROM project_members WHERE project_id = $1 AND user_id = $2`,
		projectID, userID,
	)

	m := &domain.ProjectMember{}
	if err := row.Scan(&m.ID, &m.ProjectID, &m.UserID, &m.Role, &m.InvitedBy, &m.JoinedAt, &m.CreatedAt); err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("member not found")
		}
		return nil, fmt.Errorf("scanning member: %w", err)
	}
	return m, nil
}

func (r *ProjectRepository) ListMembers(ctx context.Context, projectID, cursor string, limit int32) ([]*domain.ProjectMember, string, int32, error) {
	query := `SELECT pm.id, pm.project_id, pm.user_id, pm.role, pm.invited_by, pm.joined_at, pm.created_at
		FROM project_members pm WHERE pm.project_id = $1 ORDER BY pm.created_at ASC`

	var args []interface{}
	args = append(args, projectID)

	if cursor != "" {
		ts, id, ok := pagination.Decode(cursor)
		if !ok {
			return nil, "", 0, fmt.Errorf("invalid member cursor")
		}
		query = `SELECT pm.id, pm.project_id, pm.user_id, pm.role, pm.invited_by, pm.joined_at, pm.created_at
			FROM project_members pm WHERE pm.project_id = $1 AND (pm.created_at, pm.id) > ($2::timestamptz, $3)
			ORDER BY pm.created_at ASC, pm.id ASC`
		args = append(args, ts, id)
	}

	if limit <= 0 {
		limit = 20
	}
	query += fmt.Sprintf(" LIMIT %d", limit+1)

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, "", 0, fmt.Errorf("querying members: %w", err)
	}
	defer rows.Close()

	var members []*domain.ProjectMember
	count := int32(0)
	for rows.Next() {
		m := &domain.ProjectMember{}
		if err := rows.Scan(&m.ID, &m.ProjectID, &m.UserID, &m.Role, &m.InvitedBy, &m.JoinedAt, &m.CreatedAt); err != nil {
			return nil, "", 0, fmt.Errorf("scanning member: %w", err)
		}
		members = append(members, m)
		count++
	}

	var nextCursor string
	if int32(len(members)) > limit {
		members = members[:len(members)-1]
		nextCursor = pagination.Encode(members[len(members)-1].CreatedAt, members[len(members)-1].ID)
	}

	return members, nextCursor, count, nil
}

func (r *ProjectRepository) ListMemberUsers(ctx context.Context, projectID string) ([]*domain.ProjectMember, error) {
	rows, err := r.db.Query(ctx,
		`SELECT id, project_id, user_id, role, invited_by, joined_at, created_at
		 FROM project_members WHERE project_id = $1`, projectID,
	)
	if err != nil {
		return nil, fmt.Errorf("querying all members: %w", err)
	}
	defer rows.Close()

	var members []*domain.ProjectMember
	for rows.Next() {
		m := &domain.ProjectMember{}
		if err := rows.Scan(&m.ID, &m.ProjectID, &m.UserID, &m.Role, &m.InvitedBy, &m.JoinedAt, &m.CreatedAt); err != nil {
			return nil, fmt.Errorf("scanning member: %w", err)
		}
		members = append(members, m)
	}
	return members, nil
}
