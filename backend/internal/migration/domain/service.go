package domain

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/schemahub/backend/internal/pkg/errors"
)

type MigrationService struct {
	repo       MigrationRepository
	validator  *SQLValidator
	connString func(ctx context.Context, connID string) (string, error)
	watchMu    sync.RWMutex
	watchers   map[string][]chan *MigrationStatusMessage
}

func NewMigrationService(repo MigrationRepository, connString func(ctx context.Context, connID string) (string, error)) *MigrationService {
	return &MigrationService{
		repo:       repo,
		validator:  NewSQLValidator(),
		connString: connString,
		watchers:   make(map[string][]chan *MigrationStatusMessage),
	}
}

// ── CRUD ──

func (s *MigrationService) Create(ctx context.Context, m *Migration) (*Migration, error) {
	if err := m.Validate(); err != nil {
		return nil, errors.New("INVALID_ARGUMENT", err.Error())
	}

	existing, _ := s.repo.GetByProjectAndVersion(ctx, m.ProjectID, m.Version)
	if existing != nil {
		return nil, errors.New("ALREADY_EXISTS", fmt.Sprintf("migration version %s already exists in project", m.Version))
	}

	m.Checksum = ComputeChecksum(m.UpSQL)
	m.Status = MigrationStatusDraft

	if err := s.repo.Create(ctx, m); err != nil {
		return nil, fmt.Errorf("creating migration: %w", err)
	}

	return m, nil
}

func (s *MigrationService) GetByID(ctx context.Context, id string) (*Migration, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *MigrationService) ListByProject(ctx context.Context, projectID, cursor string, pageSize int32) ([]*Migration, string, int32, error) {
	if pageSize <= 0 || pageSize > 100 {
		pageSize = 20
	}
	return s.repo.ListByProjectID(ctx, projectID, cursor, pageSize)
}

func (s *MigrationService) Update(ctx context.Context, m *Migration) (*Migration, error) {
	existing, err := s.repo.GetByID(ctx, m.ID)
	if err != nil {
		return nil, err
	}

	if existing.Status != MigrationStatusDraft {
		return nil, errors.New("FAILED_PRECONDITION", "can only update migrations in draft status")
	}

	if m.Title != "" {
		existing.Title = m.Title
	}
	if m.Description != "" {
		existing.Description = m.Description
	}
	if m.UpSQL != "" {
		existing.UpSQL = m.UpSQL
		existing.Checksum = ComputeChecksum(m.UpSQL)
	}
	if m.DownSQL != "" {
		existing.DownSQL = m.DownSQL
	}
	if m.Status != "" {
		existing.Status = m.Status
	}

	if err := s.repo.Update(ctx, existing); err != nil {
		return nil, fmt.Errorf("updating migration: %w", err)
	}

	return existing, nil
}

func (s *MigrationService) Delete(ctx context.Context, id string) error {
	m, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if m.Status == MigrationStatusRunning {
		return errors.New("FAILED_PRECONDITION", "cannot delete a running migration")
	}
	return s.repo.SoftDelete(ctx, id)
}

// ── Execute ──

func (s *MigrationService) Execute(ctx context.Context, migrationID, connectionID, userID string) (*MigrationRun, error) {
	m, err := s.repo.GetByID(ctx, migrationID)
	if err != nil {
		return nil, err
	}

	if m.Status != MigrationStatusDraft && m.Status != MigrationStatusPending {
		return nil, errors.New("FAILED_PRECONDITION", "migration must be in draft or pending status")
	}

	activeRun, _ := s.repo.GetActiveRunForConnection(ctx, connectionID)
	if activeRun != nil {
		return nil, errors.New("FAILED_PRECONDITION", "a migration is already running on this connection")
	}

	run := &MigrationRun{
		MigrationID:  migrationID,
		ConnectionID: connectionID,
		Direction:    MigrationDirectionUp,
		Status:       RunStatusPending,
		ExecutedBy:   userID,
	}

	if err := s.repo.CreateRun(ctx, run); err != nil {
		return nil, fmt.Errorf("creating run: %w", err)
	}

	m.Status = MigrationStatusRunning
	if err := s.repo.Update(ctx, m); err != nil {
		return nil, fmt.Errorf("updating migration status: %w", err)
	}

	go s.executeAsync(run.ID, m.UpSQL, connectionID)

	return run, nil
}

func (s *MigrationService) executeAsync(runID, sql, connID string) {
	ctx := context.Background()

	run, err := s.repo.GetRunByID(ctx, runID)
	if err != nil {
		return
	}

	run.Status = RunStatusRunning
	now := time.Now()
	run.StartedAt = &now
	_ = s.repo.UpdateRun(ctx, run)

	s.broadcast(runID, &MigrationStatusMessage{
		RunID: runID, State: RunStatusRunning, StartedAt: now,
	})

	connStr, err := s.connString(ctx, connID)
	if err != nil {
		run.Status = RunStatusFailed
		run.ErrorMessage = fmt.Sprintf("connection string: %v", err)
		completed := time.Now()
		run.CompletedAt = &completed
		run.DurationMs = int32(time.Since(now).Milliseconds())
		_ = s.repo.UpdateRun(ctx, run)
		s.broadcast(runID, &MigrationStatusMessage{
			RunID: runID, State: RunStatusFailed, ErrorMessage: run.ErrorMessage,
		})
		s.finalizeMigration(ctx, run.MigrationID, MigrationStatusFailed)
		return
	}

	pool, err := pgxpool.New(ctx, connStr)
	if err != nil {
		run.Status = RunStatusFailed
		run.ErrorMessage = fmt.Sprintf("connecting: %v", err)
		completed := time.Now()
		run.CompletedAt = &completed
		run.DurationMs = int32(time.Since(now).Milliseconds())
		_ = s.repo.UpdateRun(ctx, run)
		s.broadcast(runID, &MigrationStatusMessage{
			RunID: runID, State: RunStatusFailed, ErrorMessage: run.ErrorMessage,
		})
		s.finalizeMigration(ctx, run.MigrationID, MigrationStatusFailed)
		return
	}
	defer pool.Close()

	stmts := splitSQLStatements(sql)
	total := len(stmts)

	s.broadcast(runID, &MigrationStatusMessage{
		RunID: runID, State: RunStatusRunning, TotalStatements: int32(total),
		CompletedStatements: 0, StartedAt: now,
	})

	tx, err := pool.Begin(ctx)
	if err != nil {
		run.Status = RunStatusFailed
		run.ErrorMessage = fmt.Sprintf("begin transaction: %v", err)
		completed := time.Now()
		run.CompletedAt = &completed
		run.DurationMs = int32(time.Since(now).Milliseconds())
		_ = s.repo.UpdateRun(ctx, run)
		s.broadcast(runID, &MigrationStatusMessage{
			RunID: runID, State: RunStatusFailed, ErrorMessage: run.ErrorMessage,
		})
		s.finalizeMigration(ctx, run.MigrationID, MigrationStatusFailed)
		return
	}

	success := true
	for i, stmt := range stmts {
		if strings.TrimSpace(stmt) == "" {
			continue
		}

		stmtStart := time.Now()
		_, err := tx.Exec(ctx, stmt)
		stmtDuration := int32(time.Since(stmtStart).Milliseconds())

		logEntry := &MigrationLogEntry{
			MigrationRunID: runID,
			Sequence:       i + 1,
			SQL:            stmt,
			DurationMs:     &stmtDuration,
		}

		if err != nil {
			logEntry.ErrorMessage = err.Error()
			_ = s.repo.CreateLogEntry(ctx, logEntry)
			_ = tx.Rollback(ctx)

			run.Status = RunStatusFailed
			run.ErrorMessage = fmt.Sprintf("statement %d: %v", i+1, err)
			completed := time.Now()
			run.CompletedAt = &completed
			run.DurationMs = int32(time.Since(now).Milliseconds())
			_ = s.repo.UpdateRun(ctx, run)

			s.broadcast(runID, &MigrationStatusMessage{
				RunID: runID, State: RunStatusFailed,
				CompletedStatements: int32(i), TotalStatements: int32(total),
				ErrorMessage: run.ErrorMessage, StartedAt: now,
				ElapsedMs: time.Since(now).Milliseconds(),
				LastLog:   logEntry,
			})

			s.finalizeMigration(ctx, run.MigrationID, MigrationStatusFailed)
			success = false
			break
		}

		_ = s.repo.CreateLogEntry(ctx, logEntry)

		s.broadcast(runID, &MigrationStatusMessage{
			RunID: runID, State: RunStatusRunning,
			CompletedStatements: int32(i + 1), TotalStatements: int32(total),
			CurrentStatement: stmt, StartedAt: now,
			ElapsedMs: time.Since(now).Milliseconds(),
			LastLog:   logEntry,
		})
	}

	if success {
		if err := tx.Commit(ctx); err != nil {
			run.Status = RunStatusFailed
			run.ErrorMessage = fmt.Sprintf("commit transaction: %v", err)
			completed := time.Now()
			run.CompletedAt = &completed
			run.DurationMs = int32(time.Since(now).Milliseconds())
			_ = s.repo.UpdateRun(ctx, run)

			s.broadcast(runID, &MigrationStatusMessage{
				RunID: runID, State: RunStatusFailed,
				ErrorMessage: run.ErrorMessage, StartedAt: now,
				ElapsedMs: time.Since(now).Milliseconds(),
			})

			s.finalizeMigration(ctx, run.MigrationID, MigrationStatusFailed)
			return
		}

		run.Status = RunStatusCompleted
		completed := time.Now()
		run.CompletedAt = &completed
		run.DurationMs = int32(time.Since(now).Milliseconds())
		_ = s.repo.UpdateRun(ctx, run)

		s.broadcast(runID, &MigrationStatusMessage{
			RunID: runID, State: RunStatusCompleted,
			CompletedStatements: int32(total), TotalStatements: int32(total),
			StartedAt: now, ElapsedMs: time.Since(now).Milliseconds(),
		})

		s.finalizeMigration(ctx, run.MigrationID, MigrationStatusCompleted)
	}
}

func (s *MigrationService) finalizeMigration(ctx context.Context, migrationID string, status MigrationStatus) {
	m, err := s.repo.GetByID(ctx, migrationID)
	if err != nil {
		return
	}
	m.Status = status
	_ = s.repo.Update(ctx, m)
}

// ── Rollback ──

func (s *MigrationService) Rollback(ctx context.Context, runID, userID string) (*MigrationRun, error) {
	originalRun, err := s.repo.GetRunByID(ctx, runID)
	if err != nil {
		return nil, err
	}

	m, err := s.repo.GetByID(ctx, originalRun.MigrationID)
	if err != nil {
		return nil, err
	}

	if m.DownSQL == "" {
		return nil, errors.New("FAILED_PRECONDITION", "migration has no down_sql defined")
	}

	if originalRun.Status != RunStatusCompleted {
		return nil, errors.New("FAILED_PRECONDITION", "can only roll back a completed migration run")
	}

	activeRun, _ := s.repo.GetActiveRunForConnection(ctx, originalRun.ConnectionID)
	if activeRun != nil {
		return nil, errors.New("FAILED_PRECONDITION", "a migration is already running on this connection")
	}

	run := &MigrationRun{
		MigrationID:  originalRun.MigrationID,
		ConnectionID: originalRun.ConnectionID,
		Direction:    MigrationDirectionDown,
		Status:       RunStatusPending,
		ExecutedBy:   userID,
	}

	if err := s.repo.CreateRun(ctx, run); err != nil {
		return nil, fmt.Errorf("creating rollback run: %w", err)
	}

	go s.executeAsync(run.ID, m.DownSQL, originalRun.ConnectionID)

	return run, nil
}

// ── Watch ──

func (s *MigrationService) Subscribe(runID string) chan *MigrationStatusMessage {
	ch := make(chan *MigrationStatusMessage, 100)
	s.watchMu.Lock()
	s.watchers[runID] = append(s.watchers[runID], ch)
	s.watchMu.Unlock()
	return ch
}

func (s *MigrationService) Unsubscribe(runID string, ch chan *MigrationStatusMessage) {
	s.watchMu.Lock()
	defer s.watchMu.Unlock()
	watchers := s.watchers[runID]
	for i, w := range watchers {
		if w == ch {
			s.watchers[runID] = append(watchers[:i], watchers[i+1:]...)
			break
		}
	}
	close(ch)
}

func (s *MigrationService) broadcast(runID string, msg *MigrationStatusMessage) {
	s.watchMu.RLock()
	watchers := s.watchers[runID]
	s.watchMu.RUnlock()
	for _, ch := range watchers {
		select {
		case ch <- msg:
		default:
		}
	}
}

// ── Validate & Dry Run ──

func (s *MigrationService) Validate(ctx context.Context, upSQL, downSQL string) (bool, []string) {
	return s.validator.Validate(upSQL, downSQL)
}

func (s *MigrationService) DryRun(ctx context.Context, migrationID, connectionID string) (bool, []string, []string) {
	m, err := s.repo.GetByID(ctx, migrationID)
	if err != nil {
		return false, []string{err.Error()}, nil
	}

	valid, valErrors := s.validator.Validate(m.UpSQL, m.DownSQL)
	if !valid {
		return false, valErrors, nil
	}

	connStr, err := s.connString(ctx, connectionID)
	if err != nil {
		return false, []string{fmt.Sprintf("connection: %v", err)}, nil
	}

	pool, err := pgxpool.New(ctx, connStr)
	if err != nil {
		return false, []string{fmt.Sprintf("connecting: %v", err)}, nil
	}
	defer pool.Close()

	var warnings []string
	stmts := splitSQLStatements(m.UpSQL)
	for i, stmt := range stmts {
		if strings.TrimSpace(stmt) == "" {
			continue
		}
		_, err := pool.Exec(ctx, "SELECT 1 WHERE FALSE")
		if err != nil {
			warnings = append(warnings, fmt.Sprintf("statement %d: connection warning: %v", i+1, err))
		}
	}

	return true, nil, warnings
}

// ── Run Queries ──

func (s *MigrationService) GetRunByID(ctx context.Context, id string) (*MigrationRun, error) {
	return s.repo.GetRunByID(ctx, id)
}

func (s *MigrationService) ListRuns(ctx context.Context, migrationID, cursor string, pageSize int32) ([]*MigrationRun, string, int32, error) {
	if pageSize <= 0 || pageSize > 100 {
		pageSize = 20
	}
	return s.repo.ListRunsByMigrationID(ctx, migrationID, cursor, pageSize)
}

func (s *MigrationService) GetLogs(ctx context.Context, runID string) ([]*MigrationLogEntry, error) {
	return s.repo.ListLogsByRunID(ctx, runID)
}

func splitSQLStatements(sql string) []string {
	var stmts []string
	current := strings.Builder{}
	depth := 0
	inString := false
	stringChar := byte(0)

	for i := 0; i < len(sql); i++ {
		ch := sql[i]

		if inString {
			current.WriteByte(ch)
			if ch == stringChar && (i+1 >= len(sql) || sql[i+1] != stringChar) {
				inString = false
			}
			continue
		}

		switch ch {
		case '\'', '"':
			inString = true
			stringChar = ch
			current.WriteByte(ch)
		case '(':
			depth++
			current.WriteByte(ch)
		case ')':
			depth--
			current.WriteByte(ch)
		case ';':
			if depth == 0 {
				current.WriteByte(ch)
				stmts = append(stmts, current.String())
				current.Reset()
			} else {
				current.WriteByte(ch)
			}
		case '-':
			if i+1 < len(sql) && sql[i+1] == '-' {
				for i < len(sql) && sql[i] != '\n' {
					i++
				}
				continue
			}
			current.WriteByte(ch)
		default:
			current.WriteByte(ch)
		}
	}

	remaining := strings.TrimSpace(current.String())
	if remaining != "" {
		stmts = append(stmts, remaining)
	}

	return stmts
}

var _ = pgx.ErrNoRows
