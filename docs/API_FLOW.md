# API Flow

> **End-to-end request flows for every SchemaHub feature — user actions, backend actions, database interactions, validation, authentication, and failure modes.**

---

## Table of Contents

- [Flow Documentation Format](#flow-documentation-format)
- [Authentication Flows](#authentication-flows)
- [Project Management Flows](#project-management-flows)
- [Connection Management Flows](#connection-management-flows)
- [Schema Exploration Flows](#schema-exploration-flows)
- [Migration Execution Flows](#migration-execution-flows)
- [Version Management Flows](#version-management-flows)
- [Drift Detection Flows](#drift-detection-flows)
- [Real-Time Event Flows](#real-time-event-flows)

---

## Flow Documentation Format

Each flow follows this structure:

```
## Flow Name

**Trigger:** What initiates this flow
**Auth Required:** Yes/No
**Roles Required:** [List of roles]

### Sequence

[User Action] → [Frontend] → [API Call] → [Backend] → [Database]

### Validation Rules
- What is validated at each step

### Success Response
- What the client receives

### Failure Cases
- What can go wrong and how it is handled
```

---

## Authentication Flows

### User Registration

**Trigger:** User submits registration form
**Auth Required:** No

**Sequence:**

```
1. User fills registration form (email, password, display_name)
2. Frontend validates:
   - Email format (regex)
   - Password strength (min 8 chars, uppercase, lowercase, number)
   - Display name (non-empty, max 100 chars)
3. Frontend calls AuthService.Register() via gRPC-Web
4. Backend validates:
   - Email not already registered
   - Password meets strength requirements
5. Backend hashes password with bcrypt (cost 12)
6. Backend creates user record:
   - INSERT INTO users (id, email, password_hash, display_name)
7. Backend generates email verification token
8. Backend creates email_verification record
9. Backend sends verification email (async, non-blocking)
10. Backend generates JWT access + refresh tokens
11. Backend creates refresh_token record
12. Response: { user, access_token, refresh_token, expires_in }

Failure Cases:
- Email already exists → AlreadyExists error
- Password too weak → InvalidArgument with field errors
- Database failure → Internal error
```

### User Login

**Trigger:** User submits login form
**Auth Required:** No

**Sequence:**

```
1. User enters email and password
2. Frontend validates: email format, password non-empty
3. Frontend calls AuthService.Login()
4. Backend queries user by email
5. If user not found → Unauthenticated error (generic "invalid credentials")
6. Backend compares password_hash with bcrypt
7. If mismatch → Unauthenticated error (generic "invalid credentials")
8. Backend updates last_login_at
9. Backend generates JWT access token (15 min TTL)
10. Backend generates refresh token (7 day TTL)
11. Backend creates refresh_token record
12. Backend logs audit event: "user.login"
13. Response: { user, access_token, refresh_token, expires_in }

Failure Cases:
- Invalid credentials → Unauthenticated (no indication of which field is wrong)
- Account deactivated → Unauthenticated
- Rate limited → ResourceExhausted
```

### Token Refresh

**Trigger:** Access token expired, client has refresh token
**Auth Required:** Yes (via refresh token)

**Sequence:**

```
1. Frontend detects 401 Unauthenticated
2. Frontend calls AuthService.RefreshToken() with refresh_token
3. Backend hashes refresh_token, queries refresh_tokens table
4. Backend validates:
   - Token exists
   - Token not revoked
   - Token not expired
5. Backend performs token rotation:
   - Revokes current refresh token (revoked_at = now())
   - Generates new access token
   - Generates new refresh token (same family)
   - Creates new refresh_token record
6. Response: { access_token, refresh_token, expires_in }

Failure Cases:
- Token expired → Unauthenticated (user must re-login)
- Token revoked (reuse detection) → Unauthenticated, ALL tokens in family revoked
- Token not found → Unauthenticated
```

---

## Project Management Flows

### Create Project

**Trigger:** User fills "New Project" form
**Auth Required:** Yes

**Sequence:**

```
1. User enters project name, description, visibility
2. Frontend validates: name required, max 200 chars
3. Frontend generates slug from name (lowercased, hyphenated)
4. Frontend calls ProjectService.CreateProject()
5. Backend validates auth token → extracts user ID
6. Backend validates:
   - Slug uniqueness (SELECT COUNT WHERE slug = ?)
   - Name not empty
7. Backend begins transaction:
   a. INSERT INTO projects (name, slug, description, visibility, created_by)
   b. INSERT INTO project_members (project_id, user_id, role='owner')
   c. INSERT INTO audit_logs (event_type='project.created', ...)
8. Backend commits transaction
9. Backend publishes event: EventType.PROJECT_CREATED
10. Response: Project object

Failure Cases:
- Slug conflict → AlreadyExists with suggestion
- Name too long → InvalidArgument
- Database constraint violation → Internal
```

### List Projects

**Trigger:** User visits dashboard
**Auth Required:** Yes

**Sequence:**

```
1. Frontend calls ProjectService.ListProjects({ cursor, page_size })
2. Backend extracts user ID from auth
3. Backend queries:
   SELECT p.*
   FROM projects p
   JOIN project_members pm ON p.id = pm.project_id
   WHERE pm.user_id = ?
     AND p.deleted_at IS NULL
   ORDER BY p.updated_at DESC
4. Backend applies cursor pagination
5. Backend returns paginated response

Failure Cases:
- Invalid cursor → InvalidArgument
- None (read operation)
```

---

## Connection Management Flows

### Create Connection

**Trigger:** User adds database connection
**Auth Required:** Yes

**Sequence:**

```
1. User enters connection details:
   - Name, host, port, database, username, password, SSL mode
2. Frontend validates: all required fields present, port is number
3. Frontend calls ProjectService.CreateConnection()
4. Backend validates:
   - User has write access to project (role >= member)
   - Connection name unique within project
5. Backend encrypts password using AES-256-GCM:
   - Generate random nonce (12 bytes)
   - Encrypt with master key (from env)
   - Store: base64(nonce + ciphertext)
6. Backend inserts connection record
7. Backend starts async connection test:
   a. Open TCP connection to host:port (timeout: 10s)
   b. Perform TLS handshake if SSL enabled
   c. Authenticate with PostgreSQL
   d. Run SELECT version()
   e. Update connection_status and last_connected_at
8. Response: Connection object (password not included)
9. If async test fails, status = 'failed' with error stored

Failure Cases:
- Invalid credentials → Connection test fails (status='failed')
- Network unreachable → Connection test fails
- Duplicate name → AlreadyExists
- Insufficient permissions → PermissionDenied
```

### Test Connection

**Trigger:** User clicks "Test Connection" button
**Auth Required:** Yes

**Sequence:**

```
1. Frontend calls ProjectService.TestConnection({ connection_id })
2. Backend validates user has read access to project
3. Backend retrieves and decrypts stored password
4. Backend attempts connection:
   a. Dial TCP (timeout: 5s)
   b. TLS handshake if SSL enabled
   c. Authenticate with PostgreSQL
   d. Run SELECT version()
   e. Run SELECT current_database()
5. Backend updates connection_status
6. Backend records latency_ms
7. Response: { success, latency_ms, server_version, database_name }

Failure Cases:
- All connection failures → success=false with error details
- Encrypted password cannot be decrypted → Internal
```

---

## Schema Exploration Flows

### Introspect Schema

**Trigger:** User triggers schema introspection
**Auth Required:** Yes

**Sequence:**

```
1. User selects a connection and schema name
2. Frontend calls SchemaService.IntrospectSchema({ connection_id, schema_names })
3. Backend validates user has read access to connection's project
4. Backend retrieves connection details (decrypts password)
5. Backend opens connection to target database (pgx pool)
6. Backend queries PostgreSQL system catalogs:
   a. information_schema.tables — list all tables
   b. information_schema.columns — column definitions for each table
   c. pg_indexes — index definitions
   d. information_schema.table_constraints — constraints
   e. pg_stat_user_tables — row count estimates
   f. pg_type — custom enum types
   g. pg_extension — installed extensions
7. Backend constructs schema JSONB metadata
8. Backend computes SHA-256 checksum of metadata
9. Backend checks if checksum exists (content-addressed dedup):
   - If same checksum exists: link to existing version, no new version created
   - If new: create new schema_version
10. Backend updates schema's current_version_id
11. Backend populates schema_objects table
12. Backend publishes event: EventType.SCHEMA_VERSION_CREATED
13. Response: SchemaVersion object

Failure Cases:
- Connection failed → FailedPrecondition
- Schema does not exist → NotFound
- Large schema → streaming response with progress updates
- Permission denied in target DB → PermissionDenied
```

### Compare Schema Versions

**Trigger:** User selects two versions to compare
**Auth Required:** Yes

**Sequence:**

```
1. User selects two schema versions (A and B)
2. Frontend calls SchemaService.CompareSchemaVersions({
     version_a_id, version_b_id
   })
3. Backend retrieves both versions' JSONB metadata
4. Backend computes diff:
   a. Compare tables by name → added, removed, modified
   b. For modified tables, compare columns → added, removed, modified columns
   c. Compare indexes, constraints, enums, extensions
5. Backend structures diff response:
   - Added objects: { type, name, definition }
   - Removed objects: { type, name, definition }
   - Modified objects: { type, name, changes: [{ field, before, after }] }
6. Response: SchemaDiff object

Failure Cases:
- Version not found → NotFound
- Versions from different schemas → InvalidArgument
```

---

## Migration Execution Flows

### Create Migration

**Trigger:** User writes a migration
**Auth Required:** Yes

**Sequence:**

```
1. User provides title, version, up SQL, down SQL (optional)
2. Frontend validates: title non-empty, SQL non-empty
3. Frontend calls MigrationService.CreateMigration()
4. Backend validates:
   - User has write access to project
   - Version string is unique within project
   - Up SQL is valid PostgreSQL (parses without error)
   - Down SQL is valid PostgreSQL (if provided)
5. Backend computes SHA-256 checksum of up_sql
6. Backend inserts migration record
7. Response: Migration object (status = 'draft')

Failure Cases:
- Invalid SQL syntax → InvalidArgument with line/position details
- Version already exists → AlreadyExists
- SQL contains disallowed statements → InvalidArgument
```

### Execute Migration

**Trigger:** User executes a draft migration
**Auth Required:** Yes

**Sequence:**

```
1. User selects migration and target connection
2. Frontend calls MigrationService.ExecuteMigration({
     migration_id, connection_id
   })
3. Backend validates:
   - User has write access to project
   - Migration is in 'draft' or 'pending' status
   - Connection is connected (tested)
   - No other migration is running on this connection
4. Backend creates MigrationRun record (status = 'pending')
5. Backend updates migration status to 'running'
6. Backend opens connection to target database
7. Backend begins transaction (BEGIN)
8. Backend executes each SQL statement in up_sql:
   a. Split SQL into individual statements
   b. Execute each with pgx
   c. Log each to migration_logs
   d. On error: ROLLBACK, mark run as 'failed', update migration status
9. If all statements succeed: COMMIT
10. Backend updates migration status = 'completed'
11. Backend updates run: duration_ms, completed_at
12. Backend publishes event: EventType.MIGRATION_COMPLETED
13. Backend triggers schema introspection (async)
14. Response: MigrationRun object

Streaming: If WatchMigration was called, backends sends status updates:
  { state: 'RUNNING', completed_statements: 1, total_statements: 3 }
  { state: 'RUNNING', completed_statements: 2, total_statements: 3 }
  { state: 'COMPLETED', completed_statements: 3, total_statements: 3 }

Failure Cases:
- SQL execution error → Transaction rolled back, error details returned
- Connection lost during execution → Migration marked as 'failed', manual review needed
- Timeout → Statement timeout exceeded, migration marked as 'failed'
- Concurrent migration → Aborted error
```

### Rollback Migration

**Trigger:** User rolls back a completed migration
**Auth Required:** Yes

**Sequence:**

```
1. User selects a completed migration
2. Frontend calls MigrationService.RollbackMigration({ migration_run_id })
3. Backend validates:
   - User has write access
   - Migration has down_sql
   - Migration is 'completed'
4. Backend creates MigrationRun (direction = 'down')
5. Backend sets migration status = 'rolling_back'
6. Same execution flow as ExecuteMigration but with down_sql
7. On success: migration status = 'rolled_back'
8. On failure: migration status = 'failed_rollback' (needs manual intervention)
9. Response: MigrationRun object

Failure Cases:
- No down_sql provided → FailedPrecondition
- Rollback SQL fails → Requires manual intervention (status = 'failed_rollback')
```

---

## Version Management Flows

### List Schema Versions

**Trigger:** User views version history
**Auth Required:** Yes

**Sequence:**

```
1. Frontend calls SchemaService.ListSchemaVersions({
     schema_id, cursor, page_size
   })
2. Backend queries schema_versions table
3. Returns paginated version list with metadata
4. Each version includes: version number, checksum, object_count, created_at, created_by

Failure Cases:
- Schema not found → NotFound
```

### Get Version Diagram

**Trigger:** User views schema diagram for a version
**Auth Required:** Yes

**Sequence:**

```
1. Frontend calls SchemaService.GetSchemaDiagram({
     schema_version_id, include_details
   })
2. Backend retrieves schema_objects for this version
3. Backend computes node positions:
   - Group tables by schema
   - Layout algorithm (Dagre or similar)
   - Position tables with columns, indexes, constraints
4. Backend detects foreign key relationships:
   - Parse column definitions and constraints
   - Create edges between related table nodes
5. Returns DiagramData (nodes + edges)

Failure Cases:
- Version not found → NotFound
- No objects → Empty diagram
```

---

## Drift Detection Flows

### Manual Drift Check

**Trigger:** User triggers drift detection
**Auth Required:** Yes

**Sequence:**

```
1. User requests drift check for a schema
2. Frontend calls SchemaService.IntrospectSchema() (re-introspect)
3. Backend compares current_introspection with current_version
4. If mismatches found:
   a. Creates drift_event records
   b. Publishes EventType.DRIFT_DETECTED
   c. Returns drift summary
5. If no mismatches:
   a. Returns empty drift report

Failure Cases:
- Connection failed → FailedPrecondition
```

### Automatic Drift Detection (Background)

**Trigger:** Scheduled job (configurable interval)
**Auth Required:** Internal service

**Sequence:**

```
1. Drift Service queries all active connections
2. For each connection:
   a. Introspect current schema
   b. Compare with latest schema_version
   c. If drift found: create drift_event, publish notification
3. Drift events are stored for dashboard display
```

---

## Real-Time Event Flows

### Subscribe to Events

**Trigger:** User views a project page
**Auth Required:** Yes

**Sequence:**

```
1. Frontend opens gRPC server-streaming call:
   EventService.Subscribe({
     project_ids: [proj_1, proj_2],
     event_types: [EVENT_TYPE_MIGRATION_COMPLETED, EVENT_TYPE_DRIFT_DETECTED],
     last_event_id: 'evt_xxx'  // null for fresh connection
   })
2. Backend validates user has access to all requested projects
3. Backend subscribes to Redis channels for each project
4. If last_event_id is provided:
   a. Query audit_log for events after this ID
   b. Replay missed events
5. Stream remains open, delivering events as they occur
6. On disconnect → client reconnects with last_event_id

Failure Cases:
- No access to project → stream closed with PermissionDenied
- Invalid event type → InvalidArgument for that type, others continue
```

### Heartbeat

**Trigger:** Periodic (every 30s)
**Auth Required:** Yes

**Sequence:**

```
1. Frontend calls EventService.Heartbeat({})
2. Backend updates presence TTL in Redis
3. Response: { server_time, subscriptions_active }

Failure Cases:
- None (read operation)
```
