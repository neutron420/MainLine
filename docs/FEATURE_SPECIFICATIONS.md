# Feature Specifications

> **Detailed specifications for every SchemaHub feature — problem, solution, workflow, backend behavior, frontend behavior, and future improvements.**

---

## Table of Contents

- [Feature List](#feature-list)
- [1. Authentication & OAuth Login](#1-authentication--oauth-login)
- [2. Project Management](#2-project-management)
- [3. Database Connection Management](#3-database-connection-management)
- [4. Schema Exploration](#4-schema-exploration)
- [5. Schema Versioning](#5-schema-versioning)
- [6. Schema Diff](#6-schema-diff)
- [7. Migration Creation](#7-migration-creation)
- [8. Migration Execution](#8-migration-execution)
- [9. Migration Rollback](#9-migration-rollback)
- [10. Migration History](#10-migration-history)
- [11. Real-Time Events](#11-real-time-events)
- [12. Visual Schema Diagram](#12-visual-schema-diagram)
- [13. Audit Logging](#13-audit-logging)
- [14. Drift Detection](#14-drift-detection)

---

## Feature List

| # | Feature | Phase | Priority |
|---|---|---|---|---|
| 1 | Authentication & OAuth Login | 2 | P0 |
| 2 | Project Management | 3 | P0 |
| 3 | Database Connection Management | 3 | P0 |
| 4 | Schema Exploration | 3 | P0 |
| 5 | Schema Versioning | 3 | P0 |
| 6 | Schema Diff | 5 | P1 |
| 7 | Migration Creation | 4 | P0 |
| 8 | Migration Execution | 4 | P0 |
| 9 | Migration Rollback | 4 | P0 |
| 10 | Migration History | 5 | P1 |
| 11 | Real-Time Events | 6 | P1 |
| 12 | Visual Schema Diagram | 7 | P1 |
| 13 | Audit Logging | 5 | P1 |
| 14 | Drift Detection | 9 | P2 |

---

## 1. Authentication & OAuth Login

### Problem

Users need to sign up and log in to SchemaHub. Requiring email/password creates friction and password management burden. Engineers prefer using existing identities (Google, GitHub, Slack) for one-click authentication.

### Solution

Dual authentication model: traditional email/password + OAuth 2.0 social login via Google, GitHub, and Slack. Users can authenticate with any linked provider and manage connections in account settings.

### Workflow

#### Email/Password
```
User → Login form → Enter email + password
   → Backend validates credentials → Issues JWT
   → User redirected to dashboard
```

#### OAuth (Google / GitHub / Slack)
```
User → Click "Sign in with {Provider}" → Redirected to provider → Authorize
   → Provider redirects back to SchemaHub → Backend exchanges code for identity
   → New user: account auto-created with verified email
   → Existing user (same email): link accounts after password confirmation
   → JWT issued → Redirected to dashboard
```

### Backend Behavior

- Email/password: bcrypt verification, rate-limited (5 attempts/min/IP)
- OAuth: Authorization code flow with PKCE + state parameter (signed JWT)
- Provider tokens (access + refresh) stored encrypted for session refresh
- `oauth_identities` table links provider + provider_user_id → internal user
- Account linking requires password confirmation for existing accounts
- Unlinking allowed if user has at least one authentication method remaining

### Frontend Behavior

- Login page: email/password form + OAuth provider buttons (Google, GitHub, Slack)
- Registration page: email/password + OAuth options
- Account settings: "Connected Accounts" section with link/unlink controls
- Callback handler route: `/auth/callback` processes OAuth redirect
- Linking dialog: password prompt when OAuth email matches existing account

### Future Improvements

- Apple ID, Microsoft, GitLab OAuth providers
- SAML/SSO for enterprise
- Passkeys / WebAuthn support
- MFA / TOTP two-factor authentication

---

## 2. Project Management

### Problem

Users need an organizational unit to group related database connections, schemas, and migrations. Without projects, every resource is flat and disconnected.

### Solution

Projects provide a top-level container with access control, settings, and team membership.

### Workflow

```
User → "Create Project" → Enter: name, description, visibility
   → Backend creates project → User becomes "owner"
   → User can invite others → Members get role-based access
   → Project dashboard shows connected databases and recent migrations
```

### Backend Behavior

- Project Service handles CRUD with soft delete
- Membership Service handles role assignments
- Slug auto-generated from name (unique enforcement)

### Frontend Behavior

- Project list page (user's projects)
- Project settings (name, description, visibility)
- Member management (add/remove/change role)

### Future Improvements

- Project templates (pre-configured schemas)
- Nested sub-projects
- Project-level CI/CD configuration
- Project analytics dashboard

---

## 3. Database Connection Management

### Problem

Users need to tell SchemaHub which databases to manage. Connection details must be stored securely.

### Solution

Connections store database credentials and metadata. Passwords are encrypted at rest.

### Workflow

```
User → "Add Connection" → Enter: name, host, port, database, username, password, SSL mode
   → Backend encrypts password → Stores connection → Tests connectivity async
   → User sees connection status (connected/failed)
```

### Backend Behavior

- Credentials encrypted with AES-256-GCM
- Async connection test after creation
- Periodic connection health checks
- Connection status tracked and reported

### Frontend Behavior

- Connection form with input validation
- Connection status indicator (green/red/yellow)
- Test connection button with result display
- Edit/delete connection controls

### Future Improvements

- Connection pooling configuration
- SSH tunnel support
- IAM-based authentication (AWS RDS IAM, GCP Cloud SQL)
- Connection tagging and search

---

## 4. Schema Exploration

### Problem

Engineers need to browse database schemas without connecting to the database directly. They need to see tables, columns, indexes, and constraints in a readable UI.

### Solution

Schema Explorer introspects PostgreSQL system catalogs and presents the schema structure in a tree view with detail panels.

### Workflow

```
User → Selects connection → Selects schema (e.g., "public")
   → Backend introspects database → Returns full schema metadata
   → Frontend renders tree: Tables → Columns → Properties
   → User clicks a table → Detail panel shows columns, indexes, constraints
```

### Backend Behavior

- Introspection queries `information_schema` and `pg_catalog`
- Metadata cached in Redis (TTL: 5 minutes)
- Returns full schema structure as JSONB
- Supports refresh (re-introspect)

### Frontend Behavior

- Tree view of schemas → tables → columns
- Detail panel with column properties (type, nullable, default, constraints)
- Index list with type and columns
- Constraint display (PK, FK, unique, check)
- Search/filter within schema
- Refresh button for re-introspection

### Future Improvements

- Column statistics (null %, distinct values, data distribution)
- Table row count estimates
- Sample data viewer (SELECT TOP 100)
- Query execution plan viewer

---

## 5. Schema Versioning

### Problem

Schemas change over time. Without versioning, it is impossible to know what the schema looked like at any point in history.

### Solution

Every introspection creates an immutable schema version. Versions are content-addressed (SHA-256 checksum) for deduplication.

### Workflow

```
User → Introspects schema → Backend creates SchemaVersion
   → Version stored with timestamp, checksum, full metadata
   → Previous version linked as parent
   → User can browse version history
```

### Backend Behavior

- SHA-256 checksum of normalized JSONB metadata
- Deduplication: same checksum = no new version
- Immutable: versions are never deleted or modified
- Parent link for efficient reverse diffing

### Frontend Behavior

- Version timeline (vertical timeline component)
- Version metadata display (date, objects, checksum)
- Quick navigation between versions

### Future Improvements

- Version tags (e.g., "v1.0.0", "before-migration-42")
- Version compare from timeline
- Auto-generated release notes between versions

---

## 6. Schema Diff

### Problem

Teams need to understand what changed between schema versions. Manual comparison is error-prone and time-consuming.

### Solution

Schema Diff engine compares two versions and produces a structured diff of added, removed, and modified objects.

### Workflow

```
User → Selects two versions (A and B)
   → Backend: diff JSONB metadata → Return structured diff
   → Frontend: side-by-side or unified view
   → Added objects shown in green, removed in red, modified in yellow
```

### Backend Behavior

- Object-level comparison (tables, columns, indexes, constraints)
- Column-level comparison (type, nullable, default, position)
- Change categorization: added, removed, modified, unchanged
- Output: structured diff with before/after for modified objects

### Frontend Behavior

- Side-by-side version comparison
- Unified diff view option
- Filter: show only changes (hide unchanged)
- Expand/collapse sections
- Export diff as text or JSON

### Future Improvements

- Inline comments on diff entries
- Approval workflow for changes
- Breaking change detection (column removal, type changes)
- Impact analysis (what depends on changed objects)

---

## 7. Migration Creation

### Problem

Engineers need to write SQL migrations. The platform should validate SQL before execution to catch errors early.

### Solution

Migration creation form with SQL validation. Up and down SQL are required.

### Workflow

```
User → "New Migration" → Enters: title, version, up SQL, down SQL
   → Frontend syntax highlights SQL
   → Backend: validates SQL syntax → Checks version uniqueness
   → Migration saved as "draft"
```

### Backend Behavior

- SQL syntax validation (parse with PostgreSQL parser)
- Version uniqueness check
- Disallowed statement detection (DROP DATABASE, etc.)
- Checksum computation

### Frontend Behavior

- SQL editor with syntax highlighting (Monaco or CodeMirror)
- SQL validation with inline errors
- Down SQL auto-generation (for simple changes — future)
- Version string validation (semantic version format)

### Future Improvements

- AI-assisted migration generation from schema diff
- Migration template library
- Up/down SQL side-by-side editor
- Migration dry-run before saving

---

## 8. Migration Execution

### Problem

Running migrations against production databases is risky. The platform must provide safe, observable migration execution.

### Solution

Migration executor runs SQL in a transaction, streams progress, and reports results.

### Workflow

```
User → Selects migration → Selects target connection → "Execute"
   → Backend: BEGIN → Execute each statement → Log each → COMMIT
   → Frontend: Real-time progress stream
   → Migration status: "completed" or "failed"
```

### Backend Behavior

- Transactional execution (all or nothing)
- Per-statement logging
- Streaming progress via gRPC server-streaming
- Automatic rollback on error
- Statement timeout configuration
- Concurrent execution prevention per connection

### Frontend Behavior

- Execute button (with confirmation dialog)
- Real-time progress bar (statements completed / total)
- Per-statement status display
- Success/failure notification
- Execution details view

### Future Improvements

- Scheduled migrations
- Parallel migration execution
- Multi-environment execution (dev → staging → prod)
- Pre-flight checks (row counts, data validation)

---

## 9. Migration Rollback

### Problem

Migrations sometimes fail or need to be reversed. Rollback must be safe and observable.

### Solution

Rollback executes the migration's down SQL in a transaction.

### Workflow

```
User → Selects completed migration → "Rollback"
   → Backend: Creates rollback run → Execute down SQL → Update status
   → Migration status: "rolled_back"
```

### Backend Behavior

- Same execution engine as forward migration
- Rollback runs are logged separately
- Failed rollback requires manual intervention

### Frontend Behavior

- Rollback button on completed migration
- Confirmation dialog with warning
- Same progress streaming as forward execution
- Rollback history view

### Future Improvements

- Automatic rollback on failure (configurable)
- Rollback simulation (show what will be affected)
- Conditional rollback (safety checks before executing)
- Batch rollback (multiple migrations)

---

## 10. Migration History

### Problem

Teams need to know which migrations were executed, when, by whom, and with what result.

### Solution

Migration history provides a searchable, filterable list of all migration runs.

### Workflow

```
User → Migration History tab → Filtered list of runs
   → Each entry shows: migration, version, direction, status, timing, executor
   → Click run → Per-statement log details
```

### Backend Behavior

- Query migration_runs table with filters
- Paginated response
- Log entry retrieval per run

### Frontend Behavior

- List view with search/filter/sort
- Status badges (completed, failed, rolled back)
- Duration display
- Log detail expandable per run

### Future Improvements

- Migration metrics dashboard (execution times, failure rates)
- Migration timeline visualization
- Export history as CSV/JSON
- Compare migration runs across environments

---

## 11. Real-Time Events

### Problem

Users should know immediately when schema changes happen, migrations complete, or drift is detected.

### Solution

Real-time event streaming via gRPC server-streaming delivers events to connected clients.

### Workflow

```
User → Opens project page → Frontend subscribes to events
   → Backend Event Service streams events
   → UI updates reactively (new version, migration status, etc.)
   → User sees toast/badge for important events
```

### Backend Behavior

- Event Service manages subscriptions
- Redis Pub/Sub for cross-service event distribution
- Per-client buffering and flow control
- Reconnection replay with last-event-id

### Frontend Behavior

- Subscription management (what events to receive)
- Toast notifications for important events
- Badge indicators on navigation items
- Listener pattern for UI updates

### Future Improvements

- Email notification integration
- Webhook delivery for CI/CD
- Event filtering and prioritization
- Event history browser

---

## 12. Visual Schema Diagram

### Problem

Text-based schema browsing does not show relationships between tables. Engineers need visual ERDs.

### Solution

Interactive entity-relationship diagrams using React Flow.

### Workflow

```
User → "Diagram" tab for a schema version
   → Backend returns nodes (tables) and edges (foreign keys)
   → Frontend renders interactive diagram
   → User can: pan, zoom, drag, click for details
```

### Backend Behavior

- Compute nodes from schema objects (tables, views, enums)
- Compute edges from foreign key constraints
- Automatic layout positions (Dagre algorithm)
- Support for large schemas (pagination)

### Frontend Behavior

- React Flow canvas with interactive controls
- Table nodes with column list
- Foreign key edges with labels
- Click-to-expand table details
- Search and filter within diagram
- Export as SVG/PNG

### Future Improvements

- Custom layout save/load
- Diagram annotations
- Side-by-side diagram comparison
- Real-time diagram update on schema change

---

## 13. Audit Logging

### Problem

All operations must be auditable for compliance and debugging.

### Solution

Immutable, append-only audit log captures every operation with before/after state.

### Workflow

```
All mutation operations → Audit entry created → Stored in partitioned table
   → Accessible via Audit Service → Filterable, searchable, exportable
```

### Backend Behavior

- Write-ahead audit: entry created before or atomically with the operation
- Structured changes: field-level before/after
- Queryable with multiple filters (actor, resource, time range, event type)
- Partitioned by month for performance
- Configurable retention period

### Frontend Behavior

- Audit log table with filters
- Entry detail view (changes, metadata)
- Export functionality
- Real-time tail (streaming)

### Future Improvements

- Compliance report generation
- Alert rules on specific audit events
- Audit log anomaly detection

---

## 14. Drift Detection

### Problem

Database schemas can change outside of SchemaHub (hotfixes, manual changes, tooling). These changes go undetected and cause issues.

### Solution

Drift detection compares the live database schema against the tracked schema version and reports differences.

### Workflow

```
User → "Check Drift" or scheduled job
   → Backend: Introspect live DB → Compare with latest version
   → If mismatch: Create drift_event → Notify user
   → User: Review drift → Acknowledge or resolve
```

### Backend Behavior

- Re-introspect the live database
- Compare with latest schema version (same diff engine)
- Categorize drift: missing, extra, or modified objects
- Severity classification (info, warning, critical)
- Track resolution status

### Frontend Behavior

- Drift dashboard (open drift events)
- Drift detail view (what changed)
- Acknowledge/resolve buttons
- Drift history
- Settings: drift check interval, notification preferences

### Future Improvements

- Auto-remediation (apply missing changes)
- Drift policy engine (auto-reject certain drift types)
- Cross-environment drift comparison
- Drift trend analysis
