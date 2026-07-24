package migrationv1

// ── Enums ──

type MigrationStatus string

const (
	MigrationStatusDraft      MigrationStatus = "draft"
	MigrationStatusPending    MigrationStatus = "pending"
	MigrationStatusRunning    MigrationStatus = "running"
	MigrationStatusCompleted  MigrationStatus = "completed"
	MigrationStatusFailed     MigrationStatus = "failed"
	MigrationStatusRolledBack MigrationStatus = "rolled_back"
)

type RunStatus string

const (
	RunStatusPending    RunStatus = "pending"
	RunStatusRunning    RunStatus = "running"
	RunStatusCompleted  RunStatus = "completed"
	RunStatusFailed     RunStatus = "failed"
	RunStatusRolledBack RunStatus = "rolled_back"
)

type MigrationDirection string

const (
	MigrationDirectionUp   MigrationDirection = "up"
	MigrationDirectionDown MigrationDirection = "down"
)

// ── Entities ──

type Migration struct {
	ID          string `json:"id"`
	ProjectID   string `json:"project_id"`
	Title       string `json:"title"`
	Description string `json:"description"`
	Version     string `json:"version"`
	UpSQL       string `json:"up_sql"`
	DownSQL     string `json:"down_sql"`
	Checksum    string `json:"checksum"`
	Status      string `json:"status"`
	CreatedBy   string `json:"created_by"`
	CreatedAt   string `json:"created_at"`
	UpdatedAt   string `json:"updated_at"`
}

type MigrationRun struct {
	ID           string `json:"id"`
	MigrationID  string `json:"migration_id"`
	ConnectionID string `json:"connection_id"`
	Direction    string `json:"direction"`
	Status       string `json:"status"`
	StartedAt    string `json:"started_at"`
	CompletedAt  string `json:"completed_at"`
	DurationMs   int32  `json:"duration_ms"`
	ErrorMessage string `json:"error_message"`
	ExecutedBy   string `json:"executed_by"`
	CreatedAt    string `json:"created_at"`
}

type MigrationLogEntry struct {
	Sequence     int32  `json:"sequence"`
	SQL          string `json:"sql"`
	DurationMs   int32  `json:"duration_ms"`
	RowsAffected int32  `json:"rows_affected"`
	ErrorMessage string `json:"error_message"`
	CreatedAt    string `json:"created_at"`
}

type MigrationStatusMessage struct {
	RunID               string             `json:"run_id"`
	State               string             `json:"state"`
	TotalStatements     int32              `json:"total_statements"`
	CompletedStatements int32              `json:"completed_statements"`
	CurrentStatement    string             `json:"current_statement"`
	StartedAt           string             `json:"started_at"`
	ElapsedMs           int64              `json:"elapsed_ms"`
	ErrorMessage        string             `json:"error_message"`
	LastLog             *MigrationLogEntry `json:"last_log,omitempty"`
}

// ── Requests / Responses ──

type CreateMigrationRequest struct {
	ProjectID   string `json:"project_id"`
	Title       string `json:"title"`
	Version     string `json:"version"`
	UpSQL       string `json:"up_sql"`
	DownSQL     string `json:"down_sql"`
	Description string `json:"description"`
}

type CreateMigrationResponse struct {
	Migration *Migration `json:"migration"`
}

type GetMigrationRequest struct {
	ID string `json:"id"`
}

type GetMigrationResponse struct {
	Migration *Migration `json:"migration"`
}

type ListMigrationsRequest struct {
	ProjectID string `json:"project_id"`
	Cursor    string `json:"cursor"`
	PageSize  int32  `json:"page_size"`
}

type ListMigrationsResponse struct {
	Migrations []*Migration `json:"migrations"`
	NextCursor string       `json:"next_cursor"`
	TotalCount int32        `json:"total_count"`
}

type UpdateMigrationRequest struct {
	ID          string  `json:"id"`
	Title       *string `json:"title,omitempty"`
	Description *string `json:"description,omitempty"`
	UpSQL       *string `json:"up_sql,omitempty"`
	DownSQL     *string `json:"down_sql,omitempty"`
	Status      *string `json:"status,omitempty"`
}

type UpdateMigrationResponse struct {
	Migration *Migration `json:"migration"`
}

type DeleteMigrationRequest struct {
	ID string `json:"id"`
}

type DeleteMigrationResponse struct{}

type ExecuteMigrationRequest struct {
	MigrationID  string `json:"migration_id"`
	ConnectionID string `json:"connection_id"`
}

type ExecuteMigrationResponse struct {
	Run *MigrationRun `json:"run"`
}

type WatchMigrationRequest struct {
	RunID string `json:"run_id"`
}

type RollbackMigrationRequest struct {
	RunID string `json:"run_id"`
}

type RollbackMigrationResponse struct {
	Run *MigrationRun `json:"run"`
}

type WatchRollbackRequest struct {
	RunID string `json:"run_id"`
}

type ValidateMigrationRequest struct {
	UpSQL   string `json:"up_sql"`
	DownSQL string `json:"down_sql"`
}

type ValidateMigrationResponse struct {
	Valid  bool     `json:"valid"`
	Errors []string `json:"errors,omitempty"`
}

type DryRunMigrationRequest struct {
	MigrationID  string `json:"migration_id"`
	ConnectionID string `json:"connection_id"`
}

type DryRunMigrationResponse struct {
	Valid    bool     `json:"valid"`
	Errors   []string `json:"errors,omitempty"`
	Warnings []string `json:"warnings,omitempty"`
}

type GetMigrationRunRequest struct {
	ID string `json:"id"`
}

type GetMigrationRunResponse struct {
	Run *MigrationRun `json:"run"`
}

type ListMigrationRunsRequest struct {
	MigrationID string `json:"migration_id"`
	Cursor      string `json:"cursor"`
	PageSize    int32  `json:"page_size"`
}

type ListMigrationRunsResponse struct {
	Runs      []*MigrationRun `json:"runs"`
	NextCursor string         `json:"next_cursor"`
	TotalCount int32          `json:"total_count"`
}

type GetMigrationLogsRequest struct {
	RunID string `json:"run_id"`
}
