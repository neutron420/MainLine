# Database Design

> **Complete database schema design for SchemaHub — every table, column, relationship, index, and constraint with reasoning.**

---

## Table of Contents

- [Design Philosophy](#design-philosophy)
- [Entity Overview](#entity-overview)
- [Table Specifications](#table-specifications)
- [Relationships](#relationships)
- [Indexes](#indexes)
- [Constraints](#constraints)
- [Versioning Strategy](#versioning-strategy)
- [Audit Strategy](#audit-strategy)
- [Soft Delete Strategy](#soft-delete-strategy)
- [Normalization](#normalization)
- [Future Extensibility](#future-extensibility)

---

## Design Philosophy

### Principles

1. **Schemas as content-addressed objects** — Schema versions are identified by a SHA-256 hash of their content, not by sequential IDs. This ensures that identical schemas produce identical version identifiers regardless of when or where they are created.

2. **Immutable event sourcing** — All state changes are recorded as events in the audit log. The current state is derived from applying events in order, but the events themselves are never deleted or modified.

3. **Soft deletes with hard retention** — User-facing data uses soft deletes (deleted_at timestamp) for recoverability. Audit logs and migration history use hard deletes only through retention policies.

4. **JSONB for flexible metadata** — Schema metadata (column definitions, constraints, indexes) is stored as JSONB to accommodate PostgreSQL's rich schema system without an explosion of relational tables.

5. **Time-based partitioning** — High-volume tables (audit_logs, migration_runs) are partitioned by time for query performance and data management.

---

## Entity Overview

```
┌──────────┐     ┌──────────────┐     ┌──────────────┐
│  users   │─────│  projects    │─────│ connections  │
└──────────┘     └──────────────┘     └──────────────┘
                       │                      │
                       │                      │
                       ▼                      ▼
                ┌──────────────┐     ┌──────────────┐
                │project_members│    │  schemas     │
                └──────────────┘     └──────┬───────┘
                                            │
                                            ▼
                                    ┌──────────────┐
                                    │schema_versions│
                                    └──────┬───────┘
                                           │
                                           ▼
                                    ┌──────────────┐
                                    │schema_objects │
                                    └──────────────┘

┌──────────┐     ┌──────────────┐     ┌──────────────┐
│migrations│─────│ migration_runs│    │  drift_events │
└──────────┘     └──────────────┘     └──────────────┘
                       │                      │
                       ▼                      ▼
                ┌──────────────┐     ┌──────────────┐
                │migration_logs│     │    audit_logs │
                └──────────────┘     └──────────────┘
```

---

## Table Specifications

### users

| Column | Type | Constraints | Purpose |
|---|---|---|---|
| id | UUID (v7) | PK | Unique user identifier |
| email | VARCHAR(320) | UNIQUE, NOT NULL | Login identifier |
| password_hash | VARCHAR(255) | NOT NULL | bcrypt hash |
| display_name | VARCHAR(100) | NOT NULL | User-facing name |
| avatar_url | VARCHAR(512) | Nullable | Profile image |
| role | ENUM('admin','user') | NOT NULL, DEFAULT 'user' | Global role |
| is_active | BOOLEAN | NOT NULL, DEFAULT true | Account status |
| email_verified_at | TIMESTAMPTZ | Nullable | Email verification |
| last_login_at | TIMESTAMPTZ | Nullable | Last sign-in timestamp |
| created_at | TIMESTAMPTZ | NOT NULL, DEFAULT now() | Row creation |
| updated_at | TIMESTAMPTZ | NOT NULL, DEFAULT now() | Row update |
| deleted_at | TIMESTAMPTZ | Nullable | Soft delete |

**Indexes:**
- `idx_users_email` on `email` (unique)

**Reasoning:** UUID v7 is time-ordered, reducing B-tree index fragmentation. Email is the natural identifier but scoped as the unique login credential.

---

### projects

| Column | Type | Constraints | Purpose |
|---|---|---|---|
| id | UUID (v7) | PK | Unique project identifier |
| name | VARCHAR(200) | NOT NULL | Project name |
| slug | VARCHAR(200) | UNIQUE, NOT NULL | URL-friendly identifier |
| description | TEXT | Nullable | Project description |
| visibility | ENUM('private','team','public') | NOT NULL, DEFAULT 'private' | Access scope |
| created_by | UUID | FK → users.id, NOT NULL | Creator reference |
| created_at | TIMESTAMPTZ | NOT NULL, DEFAULT now() | Row creation |
| updated_at | TIMESTAMPTZ | NOT NULL, DEFAULT now() | Row update |
| deleted_at | TIMESTAMPTZ | Nullable | Soft delete |

**Indexes:**
- `idx_projects_slug` on `slug` (unique)
- `idx_projects_created_by` on `created_by`

**Reasoning:** Slug enables human-readable URLs like `/projects/my-service-db`. Visibility supports future team collaboration features.

---

### project_members

| Column | Type | Constraints | Purpose |
|---|---|---|---|
| id | UUID (v7) | PK | Unique membership identifier |
| project_id | UUID | FK → projects.id, NOT NULL | Project reference |
| user_id | UUID | FK → users.id, NOT NULL | User reference |
| role | ENUM('owner','admin','member','viewer') | NOT NULL, DEFAULT 'member' | Project-level role |
| invited_by | UUID | FK → users.id, Nullable | Who invited this user |
| joined_at | TIMESTAMPTZ | Nullable | When user accepted |
| created_at | TIMESTAMPTZ | NOT NULL, DEFAULT now() | Row creation |

**Indexes:**
- `idx_project_members_project_user` on `(project_id, user_id)` (unique)
- `idx_project_members_user_id` on `user_id`

**Constraints:**
- UNIQUE(project_id, user_id) — one membership per user per project

---

### connections

| Column | Type | Constraints | Purpose |
|---|---|---|---|
| id | UUID (v7) | PK | Unique connection identifier |
| project_id | UUID | FK → projects.id, NOT NULL | Project reference |
| name | VARCHAR(200) | NOT NULL | Display name |
| host | VARCHAR(500) | NOT NULL | Database hostname |
| port | INTEGER | NOT NULL, DEFAULT 5432 | Connection port |
| database_name | VARCHAR(200) | NOT NULL | Database name |
| username | VARCHAR(200) | NOT NULL | DB username |
| password_encrypted | TEXT | NOT NULL | AES-256-GCM encrypted |
| ssl_mode | ENUM('disable','allow','prefer','require','verify-ca','verify-full') | NOT NULL, DEFAULT 'require' | TLS mode |
| connection_status | ENUM('unknown','connected','failed') | NOT NULL, DEFAULT 'unknown' | Last check result |
| last_connected_at | TIMESTAMPTZ | Nullable | Last successful connection |
| created_by | UUID | FK → users.id, NOT NULL | Creator |
| created_at | TIMESTAMPTZ | NOT NULL, DEFAULT now() | Row creation |
| updated_at | TIMESTAMPTZ | NOT NULL, DEFAULT now() | Row update |
| deleted_at | TIMESTAMPTZ | Nullable | Soft delete |

**Indexes:**
- `idx_connections_project_id` on `project_id`

**Security:** Passwords are encrypted at the application layer using AES-256-GCM with a key derived from a master secret stored in environment variables. The encryption key is never stored in the database.

---

### schemas

| Column | Type | Constraints | Purpose |
|---|---|---|---|
| id | UUID (v7) | PK | Unique schema identifier |
| project_id | UUID | FK → projects.id, NOT NULL | Project reference |
| connection_id | UUID | FK → connections.id, NOT NULL | Source connection |
| schema_name | VARCHAR(200) | NOT NULL | PostgreSQL schema name (e.g., public) |
| current_version_id | UUID | FK → schema_versions.id, Nullable | Current active version |
| last_introspected_at | TIMESTAMPTZ | Nullable | Last introspection time |
| created_at | TIMESTAMPTZ | NOT NULL, DEFAULT now() | Row creation |
| updated_at | TIMESTAMPTZ | NOT NULL, DEFAULT now() | Row update |
| deleted_at | TIMESTAMPTZ | Nullable | Soft delete |

**Indexes:**
- `idx_schemas_project_connection` on `(project_id, connection_id, schema_name)` (unique)

---

### schema_versions

| Column | Type | Constraints | Purpose |
|---|---|---|---|
| id | UUID (v7) | PK | Unique version identifier |
| schema_id | UUID | FK → schemas.id, NOT NULL | Schema reference |
| version | INTEGER | NOT NULL | Monotonic version number |
| checksum | VARCHAR(64) | NOT NULL | SHA-256 of schema content |
| metadata | JSONB | NOT NULL | Full schema structure |
| object_count | INTEGER | NOT NULL, DEFAULT 0 | Number of objects in schema |
| parent_version_id | UUID | FK → schema_versions.id, Nullable | Previous version for diff |
| created_by | UUID | FK → users.id, NOT NULL | Who triggered version |
| created_at | TIMESTAMPTZ | NOT NULL, DEFAULT now() | Version creation |

**Indexes:**
- `idx_schema_versions_schema_id` on `schema_id`
- `idx_schema_versions_checksum` on `checksum`
- `idx_schema_versions_schema_version` on `(schema_id, version)` (unique)

**Constraints:**
- UNIQUE(schema_id, version) — each schema has monotonically increasing versions

**JSONB Structure (metadata):**

```json
{
    "tables": [
        {
            "name": "users",
            "schema": "public",
            "columns": [
                {
                    "name": "id",
                    "data_type": "uuid",
                    "is_nullable": false,
                    "default": "gen_random_uuid()",
                    "character_maximum_length": null
                }
            ],
            "indexes": [
                {
                    "name": "idx_users_email",
                    "columns": ["email"],
                    "unique": true,
                    "index_type": "btree"
                }
            ],
            "constraints": {
                "primary_key": ["id"],
                "foreign_keys": [],
                "uniques": [],
                "checks": []
            },
            "row_count_estimate": 10000
        }
    ],
    "enums": [
        {
            "name": "user_role",
            "values": ["admin", "user"]
        }
    ],
    "extensions": ["pgcrypto", "uuid-ossp"]
}
```

**Reasoning:** The entire schema structure is stored as JSONB to enable flexible diffing, search, and visualization without complex relational queries. The checksum enables content-addressed deduplication.

---

### schema_objects

| Column | Type | Constraints | Purpose |
|---|---|---|---|
| id | UUID (v7) | PK | Unique object identifier |
| schema_version_id | UUID | FK → schema_versions.id, NOT NULL | Version reference |
| object_type | ENUM('table','view','index','sequence','function','trigger','enum','extension') | NOT NULL | Object category |
| object_name | VARCHAR(200) | NOT NULL | Object name |
| object_schema | VARCHAR(200) | NOT NULL, DEFAULT 'public' | Schema name |
| definition | JSONB | NOT NULL | Object metadata |
| parent_object_id | UUID | FK → schema_objects.id, Nullable | For nested objects |

**Indexes:**
- `idx_schema_objects_version_id` on `schema_version_id`
- `idx_schema_objects_type_name` on `(object_type, object_name)`
- `idx_schema_objects_version_type` on `(schema_version_id, object_type)`

**Reasoning:** Schema objects are normalized from the JSONB metadata for efficient querying (e.g., "find all tables named 'users' across all versions"). This enables powerful search and filtering.

---

### migrations

| Column | Type | Constraints | Purpose |
|---|---|---|---|
| id | UUID (v7) | PK | Unique migration identifier |
| project_id | UUID | FK → projects.id, NOT NULL | Project reference |
| title | VARCHAR(300) | NOT NULL | Human-readable title |
| description | TEXT | Nullable | Detailed description |
| version | VARCHAR(50) | NOT NULL | Semantic version or timestamp |
| up_sql | TEXT | NOT NULL | Forward migration SQL |
| down_sql | TEXT | Nullable | Rollback migration SQL |
| checksum | VARCHAR(64) | NOT NULL | SHA-256 of up_sql |
| status | ENUM('draft','pending','running','completed','failed','rolled_back') | NOT NULL, DEFAULT 'draft' | Current status |
| created_by | UUID | FK → users.id, NOT NULL | Author |
| created_at | TIMESTAMPTZ | NOT NULL, DEFAULT now() | Row creation |
| updated_at | TIMESTAMPTZ | NOT NULL, DEFAULT now() | Row update |
| deleted_at | TIMESTAMPTZ | Nullable | Soft delete |

**Indexes:**
- `idx_migrations_project_id` on `project_id`
- `idx_migrations_project_version` on `(project_id, version)` (unique)
- `idx_migrations_checksum` on `checksum`
- `idx_migrations_status` on `status`

**Constraints:**
- UNIQUE(project_id, version) — no duplicate migration versions per project

---

### migration_runs

| Column | Type | Constraints | Purpose |
|---|---|---|---|
| id | UUID (v7) | PK | Unique run identifier |
| migration_id | UUID | FK → migrations.id, NOT NULL | Migration reference |
| connection_id | UUID | FK → connections.id, NOT NULL | Target connection |
| direction | ENUM('up','down') | NOT NULL | Migration direction |
| status | ENUM('pending','running','completed','failed','rolled_back') | NOT NULL | Execution status |
| started_at | TIMESTAMPTZ | Nullable | Execution start |
| completed_at | TIMESTAMPTZ | Nullable | Execution end |
| duration_ms | INTEGER | Nullable | Execution duration |
| error_message | TEXT | Nullable | Error details on failure |
| executed_by | UUID | FK → users.id, NOT NULL | Who initiated |
| created_at | TIMESTAMPTZ | NOT NULL, DEFAULT now() | Row creation |

**Indexes:**
- `idx_migration_runs_migration_id` on `migration_id`
- `idx_migration_runs_connection_id` on `connection_id`
- `idx_migration_runs_status` on `status`
- `idx_migration_runs_created_at` on `created_at`

---

### migration_logs

| Column | Type | Constraints | Purpose |
|---|---|---|---|
| id | UUID (v7) | PK | Unique log entry |
| migration_run_id | UUID | FK → migration_runs.id, NOT NULL | Run reference |
| sequence | INTEGER | NOT NULL | Statement order |
| sql | TEXT | NOT NULL | Executed SQL statement |
| duration_ms | INTEGER | Nullable | Statement execution time |
| rows_affected | INTEGER | Nullable | Number of rows |
| error_message | TEXT | Nullable | Statement error |
| created_at | TIMESTAMPTZ | NOT NULL, DEFAULT now() | Log entry time |

**Indexes:**
- `idx_migration_logs_run_id` on `migration_run_id`
- `idx_migration_logs_run_sequence` on `(migration_run_id, sequence)` (unique)

---

### audit_logs (Partitioned)

| Column | Type | Constraints | Purpose |
|---|---|---|---|
| id | UUID (v7) | PK | Unique audit entry |
| event_type | VARCHAR(100) | NOT NULL | Event category |
| actor_id | UUID | FK → users.id, Nullable | Who performed the action |
| actor_email | VARCHAR(320) | Nullable | Denormalized for query speed |
| action | VARCHAR(100) | NOT NULL | Performed action |
| resource_type | VARCHAR(50) | NOT NULL | Resource category |
| resource_id | UUID | NOT NULL | Affected resource |
| resource_changes | JSONB | Nullable | Before/after diff |
| metadata | JSONB | Nullable | Additional context |
| ip_address | INET | Nullable | Request origin |
| user_agent | TEXT | Nullable | Client information |
| trace_id | UUID | NOT NULL | Correlation ID |
| created_at | TIMESTAMPTZ | NOT NULL, DEFAULT now() | Event time |

**Partitioning:** By `created_at` using monthly range partitions.

**Indexes:**
- `idx_audit_logs_created_at` on `created_at` (local to each partition)
- `idx_audit_logs_actor_id` on `actor_id` (local)
- `idx_audit_logs_resource_type_id` on `(resource_type, resource_id)` (local)
- `idx_audit_logs_event_type` on `event_type` (local)

---

### drift_events

| Column | Type | Constraints | Purpose |
|---|---|---|---|
| id | UUID (v7) | PK | Unique drift event |
| connection_id | UUID | FK → connections.id, NOT NULL | Affected connection |
| schema_id | UUID | FK → schemas.id | Affected schema |
| expected_version_id | UUID | FK → schema_versions.id | Expected version |
| drift_type | ENUM('missing_object','extra_object','modified_object','type_change') | NOT NULL | Drift category |
| object_type | VARCHAR(50) | NOT NULL | Affected object type |
| object_name | VARCHAR(200) | NOT NULL | Affected object name |
| expected_definition | JSONB | Nullable | Expected metadata |
| actual_definition | JSONB | Nullable | Actual metadata |
| diff_summary | TEXT | Nullable | Human-readable diff |
| severity | ENUM('info','warning','critical') | NOT NULL | Impact level |
| status | ENUM('open','acknowledged','resolved','false_positive') | NOT NULL, DEFAULT 'open' | Resolution status |
| detected_at | TIMESTAMPTZ | NOT NULL, DEFAULT now() | Detection time |
| resolved_at | TIMESTAMPTZ | Nullable | Resolution time |
| resolved_by | UUID | FK → users.id, Nullable | Who resolved |

**Indexes:**
- `idx_drift_events_connection_id` on `connection_id`
- `idx_drift_events_status` on `status`
- `idx_drift_events_detected_at` on `detected_at`
- `idx_drift_events_schema_object` on `(schema_id, object_name)`

---

### refresh_tokens

| Column | Type | Constraints | Purpose |
|---|---|---|---|
| id | UUID (v7) | PK | Unique token identifier |
| user_id | UUID | FK → users.id, NOT NULL | Token owner |
| token_hash | VARCHAR(64) | NOT NULL | SHA-256 of refresh token |
| expires_at | TIMESTAMPTZ | NOT NULL | Token expiry |
| revoked_at | TIMESTAMPTZ | Nullable | Token revocation |
| created_at | TIMESTAMPTZ | NOT NULL, DEFAULT now() | Token creation |
| created_by_ip | INET | NOT NULL | IP at creation |
| family | VARCHAR(50) | NOT NULL | Token family for rotation |

**Indexes:**
- `idx_refresh_tokens_user_id` on `user_id`
- `idx_refresh_tokens_token_hash` on `token_hash` (unique)

---

### email_verifications

| Column | Type | Constraints | Purpose |
|---|---|---|---|
| id | UUID (v7) | PK | Unique verification record |
| user_id | UUID | FK → users.id, NOT NULL | User reference |
| token_hash | VARCHAR(64) | NOT NULL | Verification token |
| expires_at | TIMESTAMPTZ | NOT NULL | Expiry time |
| verified_at | TIMESTAMPTZ | Nullable | Verification time |
| created_at | TIMESTAMPTZ | NOT NULL, DEFAULT now() | Record creation |

---

## Relationships

```
users 1──N project_members N──1 projects
users 1──N connections
users 1──N schema_versions
users 1──N migrations
users 1──N migration_runs
users 1──N refresh_tokens

projects 1──N connections
projects 1──N schemas
projects 1──N migrations
projects 1──N project_members

connections 1──N schemas
connections 1──N migration_runs

schemas 1──N schema_versions
schemas 1──N drift_events

schema_versions 1──N schema_objects
schema_versions 1──1 schemas (current_version)

migrations 1──N migration_runs

migration_runs 1──N migration_logs
```

---

## Indexes Summary

| Table | Index | Type | Purpose |
|---|---|---|---|
| users | idx_users_email | UNIQUE BTREE | Login lookup |
| projects | idx_projects_slug | UNIQUE BTREE | URL routing |
| project_members | idx_pm_project_user | UNIQUE BTREE | Membership constraint |
| schemas | idx_schemas_proj_conn_schema | UNIQUE BTREE | Schema uniqueness |
| schema_versions | idx_sv_schema_version | UNIQUE BTREE | Version ordering |
| schema_versions | idx_sv_checksum | BTREE | Content dedup |
| schema_objects | idx_so_version_type | BTREE | Type-based queries |
| migrations | idx_mig_project_version | UNIQUE BTREE | Version constraint |
| migration_runs | idx_mr_status | BTREE | Status filtering |
| audit_logs | idx_al_created_at | BTREE (local) | Time-range queries |
| drift_events | idx_de_status | BTREE | Open drift queries |

---

## Constraints Summary

| Type | Example |
|---|---|
| **Primary Keys** | UUID v7 on all tables |
| **Foreign Keys** | All cross-table references enforced |
| **Unique Constraints** | Email, slug, project+version, schema+connection |
| **Check Constraints** | Enum validation at DB level |
| **NOT NULL** | Required fields enforced |

---

## Versioning Strategy

### Schema Versioning

- **Content-addressed** — Each schema version has a SHA-256 checksum of its JSONB metadata
- **Monotonic** — Version numbers increase by 1 within each schema
- **Immutable** — Once created, schema versions are never modified
- **Linked** — Each version references its parent for efficient reverse diffing

### Migration Versioning

- **Semantic versioning** — Migrations use SemVer or timestamp-based versions
- **Uniqueness** — No duplicate versions within a project
- **Ordering** — Migrations are executed in version order
- **Idempotency** — Each migration is tracked by checksum to prevent duplicate execution

---

## Audit Strategy

### What Is Audited

- All mutation operations (create, update, delete) on projects, connections, schemas, migrations
- Authentication events (login, logout, token refresh, failure)
- Permission changes (role assignments, membership changes)
- Migration execution (start, completion, failure, rollback)

### Audit Design

- **Append-only** — Audit entries are never modified
- **Before/after diffs** — Where possible, store the resource state before and after the change
- **Correlation** — Every entry carries a trace_id linking it to the originating request
- **Partitioned** — Monthly partitions prevent any single table from growing unbounded
- **Retention** — Configurable retention period (default: 1 year), with archival to cold storage

---

## Soft Delete Strategy

### SchemaHub Standard

- **Applies to:** users, projects, connections, schemas, migrations
- **Does not apply to:** schema_versions, migration_runs, migration_logs, audit_logs (immutable)
- **Mechanism:** `deleted_at TIMESTAMPTZ` column (null = active, non-null = deleted)
- **Filtering:** All queries include `WHERE deleted_at IS NULL` (implemented via repository layer, not application code)
- **Restoration:** Setting `deleted_at = NULL` restores the resource
- **Cascade:** Soft-deleting a project cascades to connections and schemas (soft delete)
- **Hard delete:** A separate background job permanently removes soft-deleted records after 90 days

---

## Normalization

### Normalization Level

The schema is in **BCNF** (Boyce-Codd Normal Form) with deliberate denormalization in specific areas:

| Denormalization | Location | Rationale |
|---|---|---|
| `actor_email` on `audit_logs` | audit_logs | Avoids JOIN to users table for audit queries (user email may change) |
| `object_count` on `schema_versions` | schema_versions | Avoids COUNT query for dashboard display |
| `resource_changes` as JSONB | audit_logs | Flexible before/after storage without schema explosion |

### Why BCNF

- All functional dependencies are on candidate keys
- No redundant data (except the documented denormalizations)
- Update anomalies are impossible

---

## Future Extensibility

### Ready for Multi-DB

Adding support for additional database types (MySQL, SQLite, SQL Server):

- Add a `db_type` column to `connections`
- Add a `db_type` column to `schemas`
- Make `metadata` in `schema_versions` type-aware (different DBs have different metadata)
- The JSONB structure already accommodates flexible metadata

### Ready for Workspaces

When implementing workspace isolation (Neon branching):

- Add `workspace_id` to `connections` and `migrations`
- Add a `workspaces` table with branch metadata
- Schema versions will naturally map to workspace branches

### Ready for Plugin System

When implementing plugins:

- Add plugin metadata to the `projects` table (JSONB)
- Create a `plugin_runs` table for tracking plugin executions
- The audit log already captures plugin actions via `event_type`

### Extensibility Points

| Future Feature | Database Impact |
|---|---|
| **Multi-DB support** | Add `db_type` to connections, schemas |
| **Workspaces** | Add `workspaces` table, FK on connections |
| **CI/CD integration** | Add `ci_pipeline_id` to migrations |
| **Tags/Labels** | Add `tags` JSONB to projects, schemas, migrations |
| **Comments/Annotations** | Add `comments` table with polymorphic FK |
| **Approval workflows** | Add `approvals` table referencing migrations |
| **Plugin executions** | Add `plugin_runs` table |
| **Custom dashboards** | Use audit_logs as data source |
