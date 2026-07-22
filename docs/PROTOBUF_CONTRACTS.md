# Protobuf Contracts

> **Complete documentation of all gRPC service definitions, message contracts, naming conventions, versioning strategy, and best practices for SchemaHub.**

---

## Table of Contents

- [Design Philosophy](#design-philosophy)
- [Naming Conventions](#naming-conventions)
- [Package Organization](#package-organization)
- [Common Messages](#common-messages)
- [Auth Service](#auth-service)
- [Project Service](#project-service)
- [Schema Service](#schema-service)
- [Migration Service](#migration-service)
- [Event Service](#event-service)
- [Audit Service](#audit-service)
- [Versioning Strategy](#versioning-strategy)
- [Best Practices](#best-practices)

---

## Design Philosophy

### Principles

1. **Contracts are the source of truth** — The `.proto` files define the API. Server and client implementations must conform to them.
2. **Forward and backward compatible** — Field numbers are never reused. New fields are always optional. Messages can evolve without breaking existing clients.
3. **Explicit over implicit** — Every field is named and typed. No generic `map<string, string>` for structured data.
4. **Domain-driven** — Messages are organized by domain. Each service owns its message types.
5. **Pagination everywhere** — Every list response uses a cursor-based pagination pattern.

---

## Naming Conventions

### Files

```
{domain}/{major_version}/{service_name}.proto
```

Examples:
- `auth/v1/auth_service.proto`
- `schema/v1/schema_service.proto`

### Packages

```
schemahub.{domain}.v1
```

Examples:
- `schemahub.auth.v1`
- `schemahub.schema.v1`

### Messages

- **Request messages:** `{Action}Request` (e.g., `CreateProjectRequest`)
- **Response messages:** `{Action}Response` (e.g., `CreateProjectResponse`)
- **Entity messages:** PascalCase noun (e.g., `Project`, `SchemaVersion`)
- **Enum types:** PascalCase with `TYPE` suffix (e.g., `MigrationStatus`, `ObjectType`)
- **Field names:** snake_case (protobuf convention)

### Services

- **Service name:** `{Domain}Service` (e.g., `ProjectService`, `SchemaService`)
- **RPC methods:** Verb + Noun (e.g., `CreateProject`, `ListSchemas`, `SubscribeEvents`)

---

## Package Organization

```
proto/
├── auth/
│   └── v1/
│       ├── auth_service.proto        # AuthService definitions
│       └── auth_messages.proto       # Auth-specific types
├── project/
│   └── v1/
│       ├── project_service.proto     # ProjectService definitions
│       └── project_messages.proto    # Project-specific types
├── schema/
│   └── v1/
│       ├── schema_service.proto      # SchemaService definitions
│       └── schema_messages.proto     # Schema-specific types
├── migration/
│   └── v1/
│       ├── migration_service.proto   # MigrationService definitions
│       └── migration_messages.proto  # Migration-specific types
├── event/
│   └── v1/
│       ├── event_service.proto       # EventService definitions
│       └── event_messages.proto      # Event-specific types
├── audit/
│   └── v1/
│       ├── audit_service.proto       # AuditService definitions
│       └── audit_messages.proto      # Audit-specific types
└── common/
    └── v1/
        ├── common.proto              # Shared base types
        └── pagination.proto          # Pagination messages
```

---

## Common Messages

### Timestamp

Uses `google.protobuf.Timestamp` for all timestamp fields.

### UUID

All resource IDs use a UUID string field:

```protobuf
message ResourceIdentifier {
    string id = 1;  // UUID v7 formatted as string
}
```

### Pagination

```protobuf
message CursorPaginationRequest {
    string cursor = 1;    // Opaque cursor from previous response
    int32 page_size = 2;  // Max items per page (default: 20, max: 100)
}

message CursorPaginationResponse {
    string next_cursor = 1;   // Cursor for next page (empty if no more)
    string previous_cursor = 2; // Cursor for previous page
    int32 total_count = 3;    // Total items matching query
}
```

### Sorting

```protobuf
message SortOptions {
    string field = 1;      // Field to sort by
    SortDirection direction = 2;

    enum SortDirection {
        SORT_DIRECTION_UNSPECIFIED = 0;
        SORT_DIRECTION_ASC = 1;
        SORT_DIRECTION_DESC = 2;
    }
}
```

### Error

```protobuf
message ErrorResponse {
    string code = 1;
    string message = 2;
    map<string, string> field_errors = 3;
    string request_id = 4;
}
```

### User

```protobuf
message User {
    string id = 1;
    string email = 2;
    string display_name = 3;
    string avatar_url = 4;
    UserRole role = 5;
    bool is_active = 6;
    google.protobuf.Timestamp created_at = 7;
}

enum UserRole {
    USER_ROLE_UNSPECIFIED = 0;
    USER_ROLE_USER = 1;
    USER_ROLE_ADMIN = 2;
}
```

---

## Auth Service

### Service Definition

```protobuf
service AuthService {
    // Authentication
    rpc Register(RegisterRequest) returns (RegisterResponse);
    rpc Login(LoginRequest) returns (LoginResponse);
    rpc RefreshToken(RefreshTokenRequest) returns (RefreshTokenResponse);
    rpc Logout(LogoutRequest) returns (LogoutResponse);

    // Email Verification
    rpc SendVerificationEmail(SendVerificationEmailRequest) returns (SendVerificationEmailResponse);
    rpc VerifyEmail(VerifyEmailRequest) returns (VerifyEmailResponse);

    // User Management
    rpc GetCurrentUser(GetCurrentUserRequest) returns (GetCurrentUserResponse);
    rpc UpdateUser(UpdateUserRequest) returns (UpdateUserResponse);
    rpc ChangePassword(ChangePasswordRequest) returns (ChangePasswordResponse);
    rpc DeleteAccount(DeleteAccountRequest) returns (DeleteAccountResponse);
}
```

### Key Messages

| Message | Fields | Description |
|---|---|---|
| `LoginRequest` | email, password | User credentials |
| `LoginResponse` | access_token, refresh_token, expires_in, user | Authentication result |
| `RegisterRequest` | email, password, display_name | New user registration |
| `RefreshTokenRequest` | refresh_token | Token rotation |
| `RefreshTokenResponse` | access_token, refresh_token, expires_in | New tokens |
| `LogoutRequest` | refresh_token | Revoke refresh token |

### Auth Metadata

Authentication context is passed via gRPC metadata headers:

```
Authorization: Bearer <access_token>
X-Idempotency-Key: <uuid>
```

---

## Project Service

### Service Definition

```protobuf
service ProjectService {
    // Projects
    rpc CreateProject(CreateProjectRequest) returns (CreateProjectResponse);
    rpc GetProject(GetProjectRequest) returns (GetProjectResponse);
    rpc ListProjects(ListProjectsRequest) returns (ListProjectsResponse);
    rpc UpdateProject(UpdateProjectRequest) returns (UpdateProjectResponse);
    rpc DeleteProject(DeleteProjectRequest) returns (DeleteProjectResponse);

    // Project Members
    rpc AddMember(AddMemberRequest) returns (AddMemberResponse);
    rpc RemoveMember(RemoveMemberRequest) returns (RemoveMemberResponse);
    rpc UpdateMemberRole(UpdateMemberRoleRequest) returns (UpdateMemberRoleResponse);
    rpc ListMembers(ListMembersRequest) returns (ListMembersResponse);

    // Connections
    rpc CreateConnection(CreateConnectionRequest) returns (CreateConnectionResponse);
    rpc GetConnection(GetConnectionRequest) returns (GetConnectionResponse);
    rpc ListConnections(ListConnectionsRequest) returns (ListConnectionsResponse);
    rpc UpdateConnection(UpdateConnectionRequest) returns (UpdateConnectionResponse);
    rpc DeleteConnection(DeleteConnectionRequest) returns (DeleteConnectionResponse);
    rpc TestConnection(TestConnectionRequest) returns (TestConnectionResponse);
}
```

### Key Messages

| Message | Fields | Description |
|---|---|---|
| `Project` | id, name, slug, description, visibility, member_count, created_at | Full project entity |
| `Connection` | id, project_id, name, host, port, database_name, ssl_mode, status | Connection (without password) |
| `CreateConnectionRequest` | project_id, name, host, port, database_name, username, password, ssl_mode | New connection |
| `TestConnectionRequest` | connection_id (or inline connection params) | Test connectivity |
| `TestConnectionResponse` | success, latency_ms, error_message, server_version | Connection test result |
| `ProjectMember` | user_id, email, display_name, role, joined_at | Member with role |

---

## Schema Service

### Service Definition

```protobuf
service SchemaService {
    // Schema Management
    rpc IntrospectSchema(IntrospectSchemaRequest) returns (IntrospectSchemaResponse);
    rpc GetSchema(GetSchemaRequest) returns (GetSchemaResponse);
    rpc ListSchemas(ListSchemasRequest) returns (ListSchemasResponse);
    rpc RefreshSchema(RefreshSchemaRequest) returns (RefreshSchemaResponse);

    // Schema Versions
    rpc ListSchemaVersions(ListSchemaVersionsRequest) returns (ListSchemaVersionsResponse);
    rpc GetSchemaVersion(GetSchemaVersionRequest) returns (GetSchemaVersionResponse);
    rpc CompareSchemaVersions(CompareSchemaVersionsRequest) returns (CompareSchemaVersionsResponse);

    // Schema Objects
    rpc ListSchemaObjects(ListSchemaObjectsRequest) returns (ListSchemaObjectsResponse);
    rpc GetSchemaObject(GetSchemaObjectRequest) returns (GetSchemaObjectResponse);

    // Diagram
    rpc GetSchemaDiagram(GetSchemaDiagramRequest) returns (GetSchemaDiagramResponse);
}
```

### Key Messages

| Message | Fields | Description |
|---|---|---|
| `Schema` | id, project_id, connection_id, schema_name, current_version_id, last_introspected_at | Schema entity |
| `SchemaVersion` | id, schema_id, version, checksum, object_count, parent_version_id, created_at | Immutable version |
| `SchemaObject` | id, object_type, object_name, object_schema, definition | Individual object |
| `SchemaDiff` | added_objects, removed_objects, modified_objects | Version comparison result |
| `IntrospectSchemaRequest` | connection_id, schema_name(s) | Trigger introspection |
| `GetSchemaDiagramRequest` | schema_version_id, include_details | Diagram data request |
| `GetSchemaDiagramResponse` | nodes, edges | React Flow compatible data |

### Diagram Response Format

```protobuf
// Designed to map directly to React Flow's expected format
message DiagramNode {
    string id = 1;
    string type = 2;              // 'table', 'view', 'enum'
    DiagramPosition position = 3;
    DiagramNodeData data = 4;
}

message DiagramEdge {
    string id = 1;
    string source = 2;
    string target = 3;
    string source_handle = 4;
    string target_handle = 5;
    string label = 6;  // Foreign key constraint name
}

message DiagramPosition {
    double x = 1;
    double y = 2;
}
```

---

## Migration Service

### Service Definition

```protobuf
service MigrationService {
    // Migration Management
    rpc CreateMigration(CreateMigrationRequest) returns (CreateMigrationResponse);
    rpc GetMigration(GetMigrationRequest) returns (GetMigrationResponse);
    rpc ListMigrations(ListMigrationsRequest) returns (ListMigrationsResponse);
    rpc UpdateMigration(UpdateMigrationRequest) returns (UpdateMigrationResponse);
    rpc DeleteMigration(DeleteMigrationRequest) returns (DeleteMigrationResponse);

    // Execution
    rpc ExecuteMigration(ExecuteMigrationRequest) returns (ExecuteMigrationResponse);
    rpc WatchMigration(WatchMigrationRequest) returns (stream MigrationStatus);
    rpc RollbackMigration(RollbackMigrationRequest) returns (RollbackMigrationResponse);
    rpc WatchRollback(WatchRollbackRequest) returns (stream MigrationStatus);

    // Validation
    rpc ValidateMigration(ValidateMigrationRequest) returns (ValidateMigrationResponse);
    rpc DryRunMigration(DryRunMigrationRequest) returns (DryRunMigrationResponse);

    // History
    rpc GetMigrationRun(GetMigrationRunRequest) returns (GetMigrationRunResponse);
    rpc ListMigrationRuns(ListMigrationRunsRequest) returns (ListMigrationRunsResponse);
    rpc GetMigrationLogs(GetMigrationLogsRequest) returns (stream MigrationLogEntry);
}
```

### Key Messages

| Message | Fields | Description |
|---|---|---|
| `Migration` | id, project_id, title, version, up_sql, down_sql, checksum, status | Full migration entity |
| `MigrationRun` | id, migration_id, connection_id, direction, status, started_at, completed_at, duration_ms | Execution record |
| `MigrationStatus` | run_id, status, progress_percentage, current_statement, error_message | Live execution status |
| `MigrationLogEntry` | sequence, sql, duration_ms, rows_affected, error_message | Per-statement log |

### WatchMigration Streaming Pattern

```protobuf
// Server streaming for migration execution progress
message WatchMigrationRequest {
    string migration_run_id = 1;
}

message MigrationStatus {
    string run_id = 1;
    MigrationState state = 2;
    int32 total_statements = 3;
    int32 completed_statements = 4;
    string current_statement = 5;
    google.protobuf.Timestamp started_at = 6;
    google.protobuf.Duration elapsed = 7;
    string error_message = 8;
    MigrationLogEntry last_log = 9;
}
```

---

## Event Service

### Service Definition

```protobuf
service EventService {
    // Subscription
    rpc Subscribe(SubscribeRequest) returns (stream SchemaEvent);

    // Event Acknowledgment
    rpc AcknowledgeEvent(AcknowledgeEventRequest) returns (AcknowledgeEventResponse);

    // Connection Status
    rpc Heartbeat(HeartbeatRequest) returns (HeartbeatResponse);
}
```

### Key Messages

| Message | Fields | Description |
|---|---|---|
| `SubscribeRequest` | project_ids, event_types, last_event_id | Subscription parameters |
| `SchemaEvent` | id, type, timestamp, actor, resource, payload, metadata | Event envelope |
| `EventPayload` | One of: SchemaVersionCreated, MigrationStarted, MigrationCompleted, etc. | Typed event data |

### Event Types Enum

```protobuf
enum EventType {
    EVENT_TYPE_UNSPECIFIED = 0;

    // Schema Events
    EVENT_TYPE_SCHEMA_VERSION_CREATED = 1;
    EVENT_TYPE_SCHEMA_REFRESHED = 2;

    // Migration Events
    EVENT_TYPE_MIGRATION_STARTED = 3;
    EVENT_TYPE_MIGRATION_COMPLETED = 4;
    EVENT_TYPE_MIGRATION_FAILED = 5;
    EVENT_TYPE_MIGRATION_ROLLED_BACK = 6;

    // Drift Events
    EVENT_TYPE_DRIFT_DETECTED = 7;
    EVENT_TYPE_DRIFT_RESOLVED = 8;

    // Connection Events
    EVENT_TYPE_CONNECTION_CREATED = 9;
    EVENT_TYPE_CONNECTION_STATUS_CHANGED = 10;

    // Team Events
    EVENT_TYPE_MEMBER_ADDED = 11;
    EVENT_TYPE_MEMBER_REMOVED = 12;
    EVENT_TYPE_ROLE_CHANGED = 13;
}
```

---

## Audit Service

### Service Definition

```protobuf
service AuditService {
    rpc ListAuditEntries(ListAuditEntriesRequest) returns (ListAuditEntriesResponse);
    rpc GetAuditEntry(GetAuditEntryRequest) returns (GetAuditEntryResponse);
    rpc TailAuditEntries(TailAuditEntriesRequest) returns (stream AuditEntry);
    rpc GetAuditStats(GetAuditStatsRequest) returns (GetAuditStatsResponse);
}
```

### Key Messages

| Message | Fields | Description |
|---|---|---|
| `AuditEntry` | id, event_type, actor, action, resource, changes, metadata, ip, timestamp | Full audit record |
| `ListAuditEntriesRequest` | filters (event_type, actor, resource, date_range), pagination | Query parameters |
| `TailAuditEntriesRequest` | filter, since_event_id | Real-time audit tail |
| `ResourceChange` | field, before, after | Individual field diff |

---

## Versioning Strategy

### Major Version (Package Level)

- Breaking changes increment the major version in the package name: `schemahub.auth.v1` → `schemahub.auth.v2`
- Breaking changes include: removing fields, changing field types, changing RPC signatures
- Major versions are deployed as separate services during migration periods

### Minor Version (Field Level)

- New fields are added at the end of existing messages (high field numbers)
- New RPCs are added to existing services
- New messages are added to existing packages
- Old fields are deprecated with `deprecated = true` before removal

### Backward Compatibility Rules

1. Never remove a field — mark it `reserved` after a deprecation period
2. Never change a field's type or number
3. Never change a service's RPC signature
4. New fields must be optional or have sensible defaults
5. New enum values must be added to the end of the enum

### Deprecation Process

1. Field is marked `deprecated = true` in proto
2. Server still processes the field for one major version cycle
3. Field is moved to `reserved` in the next major version
4. Client migration docs are published

---

## Best Practices

### Message Design

- Use `google.protobuf.Timestamp` instead of Unix timestamps
- Use `google.protobuf.Duration` for time intervals
- Use wrapper types (`google.protobuf.StringValue`) for nullable primitive fields
- Use `oneof` for mutually exclusive fields
- Use maps sparingly — prefer repeated key-value messages

### Enum Design

- Always include an `UNSPECIFIED = 0` value as the first enum entry
- Use positive integers for all enum values (avoid -1)
- Document what happens on unspecified values

### Service Design

- Keep services focused on a single domain
- Prefer many small services over few large ones
- Use server streaming for real-time and large result sets
- Use client streaming for bulk operations
- Use unary RPC for standard CRUD

### Performance

- Keep individual messages under 4MB (gRPC default limit)
- Use streaming for responses expected to exceed 100 items
- Use pagination for all list operations
- Avoid deeply nested messages (max 3 levels)
