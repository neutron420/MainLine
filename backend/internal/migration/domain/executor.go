package domain

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Executor struct {
	connString func(ctx context.Context, connID string) (string, error)
	repo       MigrationRepository
	broadcast  func(runID string, msg *MigrationStatusMessage)
}

func NewExecutor(repo MigrationRepository, connString func(ctx context.Context, connID string) (string, error), broadcastFn func(runID string, msg *MigrationStatusMessage)) *Executor {
	return &Executor{
		connString: connString,
		repo:       repo,
		broadcast:  broadcastFn,
	}
}

func (e *Executor) ExecuteAsync(ctx context.Context, runID, sql, connID string) {
	run, err := e.repo.GetRunByID(ctx, runID)
	if err != nil {
		return
	}

	run.Status = RunStatusRunning
	now := time.Now()
	run.StartedAt = &now
	_ = e.repo.UpdateRun(ctx, run)

	e.broadcast(runID, &MigrationStatusMessage{
		RunID: runID, State: RunStatusRunning, StartedAt: now,
	})

	connStr, err := e.connString(ctx, connID)
	if err != nil {
		e.failRun(ctx, run, fmt.Sprintf("connection string: %v", err), now)
		return
	}

	pool, err := pgxpool.New(ctx, connStr)
	if err != nil {
		e.failRun(ctx, run, fmt.Sprintf("connecting: %v", err), now)
		return
	}
	defer pool.Close()

	stmts := splitSQLStatements(sql)
	total := len(stmts)

	e.broadcast(runID, &MigrationStatusMessage{
		RunID: runID, State: RunStatusRunning, TotalStatements: int32(total),
		CompletedStatements: 0, StartedAt: now,
	})

	tx, err := pool.Begin(ctx)
	if err != nil {
		e.failRun(ctx, run, fmt.Sprintf("begin transaction: %v", err), now)
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
			_ = e.repo.CreateLogEntry(ctx, logEntry)
			_ = tx.Rollback(ctx)

			run.Status = RunStatusFailed
			run.ErrorMessage = fmt.Sprintf("statement %d: %v", i+1, err)
			completed := time.Now()
			run.CompletedAt = &completed
			run.DurationMs = int32(time.Since(now).Milliseconds())
			_ = e.repo.UpdateRun(ctx, run)

			e.broadcast(runID, &MigrationStatusMessage{
				RunID: runID, State: RunStatusFailed,
				CompletedStatements: int32(i), TotalStatements: int32(total),
				ErrorMessage: run.ErrorMessage, StartedAt: now,
				ElapsedMs: time.Since(now).Milliseconds(),
				LastLog:   logEntry,
			})

			success = false
			break
		}

		_ = e.repo.CreateLogEntry(ctx, logEntry)
		e.broadcast(runID, &MigrationStatusMessage{
			RunID: runID, State: RunStatusRunning,
			CompletedStatements: int32(i + 1), TotalStatements: int32(total),
			CurrentStatement: stmt, StartedAt: now,
			ElapsedMs: time.Since(now).Milliseconds(),
			LastLog:   logEntry,
		})
	}

	if success {
		if err := tx.Commit(ctx); err != nil {
			e.failRun(ctx, run, fmt.Sprintf("commit transaction: %v", err), now)
			return
		}

		run.Status = RunStatusCompleted
		completed := time.Now()
		run.CompletedAt = &completed
		run.DurationMs = int32(time.Since(now).Milliseconds())
		_ = e.repo.UpdateRun(ctx, run)

		e.broadcast(runID, &MigrationStatusMessage{
			RunID: runID, State: RunStatusCompleted,
			CompletedStatements: int32(total), TotalStatements: int32(total),
			StartedAt: now, ElapsedMs: time.Since(now).Milliseconds(),
		})
	}
}

func (e *Executor) failRun(ctx context.Context, run *MigrationRun, msg string, startedAt time.Time) {
	run.Status = RunStatusFailed
	run.ErrorMessage = msg
	completed := time.Now()
	run.CompletedAt = &completed
	run.DurationMs = int32(time.Since(startedAt).Milliseconds())
	_ = e.repo.UpdateRun(ctx, run)
	e.broadcast(run.ID, &MigrationStatusMessage{
		RunID: run.ID, State: RunStatusFailed, ErrorMessage: msg,
	})
}
