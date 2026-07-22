# Architecture

> **Complete system architecture of SchemaHub including component responsibilities, request lifecycle, communication patterns, streaming, error handling, and scalability.**

---

## Table of Contents

- [System Overview](#system-overview)
- [Component Architecture](#component-architecture)
- [Service Architecture](#service-architecture)
- [Request Lifecycle](#request-lifecycle)
- [Communication Patterns](#communication-patterns)
- [gRPC Architecture](#grpc-architecture)
- [Internal Package Structure](#internal-package-structure)
- [Database Interaction](#database-interaction)
- [Streaming Architecture](#streaming-architecture)
- [Event Flow](#event-flow)
- [Error Handling](#error-handling)
- [Logging](#logging)
- [Scalability Considerations](#scalability-considerations)

---

## System Overview

SchemaHub follows a **layered microservice architecture** with a Go backend exposing gRPC endpoints, a Next.js frontend consuming those endpoints via gRPC-Web, and PostgreSQL (Neon) as the primary data store with Redis for caching, pub/sub, and session management.

```
┌──────────────────────────────────────────────────────────────────┐
│                         CLIENT LAYER                             │
│                                                                  │
│  ┌─────────────┐  ┌──────────────┐  ┌───────────┐  ┌─────────┐  │
│  │  Web Browser │  │  CLI (v2)    │  │ CI/CD     │  │ VS Code │  │
│  │  (Next.js)   │  │              │  │ Pipeline  │  │ Ext     │  │
│  └──────┬───────┘  └──────┬───────┘  └─────┬─────┘  └────┬────┘  │
│         │                 │                │              │       │
└─────────┼─────────────────┼────────────────┼──────────────┼───────┘
          │                 │                │              │
          │    gRPC-Web     │    gRPC        │   gRPC       │  gRPC
          ▼                 ▼                ▼              ▼
┌──────────────────────────────────────────────────────────────────┐
│                        API GATEWAY LAYER                         │
│                                                                  │
│  ┌──────────────────────────────────────────────────────────┐    │
│  │              gRPC Gateway / Envoy Proxy                   │    │
│  │  - TLS termination   - Rate limiting    - Auth forwarding │    │
│  │  - gRPC-Web conversion - Request routing - CORS          │    │
│  └──────────────────────────────────────────────────────────┘    │
└──────────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌──────────────────────────────────────────────────────────────────┐
│                       SERVICE LAYER                              │
│                                                                  │
│  ┌──────────┐  ┌──────────┐  ┌──────────┐  ┌──────────┐       │
│  │  Auth    │  │  Project │  │  Schema  │  │ Migration│       │
│  │  Service │  │  Service │  │  Service │  │ Service  │       │
│  └────┬─────┘  └────┬─────┘  └────┬─────┘  └────┬─────┘       │
│       │              │              │              │            │
│  ┌──────────┐  ┌──────────┐  ┌──────────┐  ┌──────────┐       │
│  │  Event   │  │  Audit   │  │  Drift   │  │  ...     │       │
│  │  Service │  │  Service │  │  Service │  │          │       │
│  └────┬─────┘  └────┬─────┘  └────┬─────┘  └──────────┘       │
└───────┼──────────────┼──────────────┼───────────────────────────┘
        │              │              │
        ▼              ▼              ▼
┌──────────────────────────────────────────────────────────────────┐
│                       DATA LAYER                                 │
│                                                                  │
│  ┌────────────────────────┐  ┌────────────────────────┐         │
│  │    PostgreSQL (Neon)   │  │        Redis            │         │
│  │  - Primary data store  │  │  - Cache (TTL-based)    │         │
│  │  - Audit log storage   │  │  - Pub/Sub (events)     │         │
│  │  - Migration tracking  │  │  - Rate limiter         │         │
│  │  - Schema metadata     │  │  - Session store        │         │
│  └────────────────────────┘  └────────────────────────┘         │
└──────────────────────────────────────────────────────────────────┘
```

---

## Component Architecture

### Frontend Components

| Component | Responsibility |
|---|---|
| **App Shell** | Layout, navigation, authentication gate |
| **Project List** | Displays and manages user projects |
| **Schema Explorer** | Tree view of database schemas with detail panels |
| **Schema Diagram** | React Flow-based interactive ERD |
| **Migration Runner** | UI for executing and monitoring migrations |
| **Version Timeline** | Visual timeline of schema versions |
| **Diff Viewer** | Side-by-side schema comparison |
| **Audit Log** | Filterable, searchable audit trail |
| **Settings** | Project settings, connection management, team management |

### Backend Services

| Service | Language | Responsibilities |
|---|---|---|
| **Auth Service** | Go | Authentication, JWT issuance, user management, RBAC |
| **Project Service** | Go | Project CRUD, workspace management, team membership |
| **Schema Service** | Go | Schema introspection, versioning, diffing, diagram data |
| **Migration Service** | Go | Migration execution, validation, rollback, history |
| **Event Service** | Go | Real-time event streaming, notification dispatch |
| **Audit Service** | Go | Audit log ingestion, storage, querying, retention |
| **Drift Service** | Go | Drift detection, comparison, alerting |

### Shared Infrastructure

| Component | Purpose |
|---|---|
| **PostgreSQL (Neon)** | Primary database — projects, users, schemas, migrations, audit logs |
| **Redis** | Cache for schema metadata, pub/sub for real-time events, rate limiting counters |
| **gRPC Gateway** | HTTP-to-gRPC conversion for browser clients |
| **Envoy Proxy** | Traffic routing, TLS termination, rate limiting, observability |

---

## Service Architecture

Each backend service follows a consistent internal structure:

```
internal/
├── domain/           # Business logic, entities, value objects
│   ├── models.go     # Domain models
│   ├── service.go    # Business logic implementation
│   └── errors.go     # Domain-specific errors
├── repository/       # Database access layer (interfaces + implementations)
│   ├── interface.go  # Repository interfaces
│   └── postgres/     # PostgreSQL implementations
├── handler/          # gRPC handler implementations
│   └── grpc.go       # gRPC service handlers
├── middleware/       # Service-specific middleware
│   ├── auth.go       # Auth interceptor
│   └── validation.go # Request validation
└── config.go         # Service configuration
```

### Dependency Flow

```
gRPC Handler → Service (domain) → Repository Interface → PostgreSQL
                    ↓
             Event Publisher → Redis Pub/Sub
                    ↓
             Audit Logger → PostgreSQL (audit_logs)
```

Handlers never access the database directly. Repositories never contain business logic. Services never know about gRPC or HTTP.

---

## Request Lifecycle

### Unary Request (e.g., List Projects)

```
Client                  gRPC Gateway            Auth Service              PostgreSQL
  │                         │                       │                        │
  │── gRPC-Web Request ────►│                       │                        │
  │                         │── Verify JWT ────────►│                        │
  │                         │◄── OK/User ──────────│                        │
  │                         │                       │                        │
  │                         │── Forward Request ────►                        │
  │                         │                       │── Query Projects ─────►│
  │                         │                       │◄── Result ────────────│
  │                         │◄── Response ──────────│                        │
  │◄── gRPC-Web Response ──│                       │                        │
```

### Streaming Request (e.g., Schema Change Notifications)

```
Client                  Event Service          Redis Pub/Sub        Schema Service
  │                         │                       │                      │
  │── Subscribe ───────────►│                       │                      │
  │                         │── Subscribe Channel ─►│                      │
  │                         │◄── Confirmed ────────│                      │
  │                         │                       │                      │
  │                         │                       │                      │
  │                   (Schema Change Occurs)        │                      │
  │                         │                       │◄── Publish Event ────│
  │                         │◄── Event ────────────│                      │
  │◄── Stream Update ──────│                       │                      │
  │                         │                       │                      │
  │                   (Another Change)              │                      │
  │                         │                       │◄── Publish Event ────│
  │                         │◄── Event ────────────│                      │
  │◄── Stream Update ──────│                       │                      │
  │                         │                       │                      │
  │── Unsubscribe ─────────►│                       │                      │
  │                         │── Unsubscribe ───────►│                      │
```

---

## Communication Patterns

| Pattern | Use Case | Protocol |
|---|---|---|
| **Unary RPC** | CRUD operations, authentication | gRPC (HTTP/2) |
| **Server Streaming** | Schema change notifications, audit log tailing | gRPC (HTTP/2) |
| **Client Streaming** | Bulk schema import, batch migration submission | gRPC (HTTP/2) |
| **Bidirectional Streaming** | Interactive migration debugging, real-time collaboration | gRPC (HTTP/2) |
| **Redis Pub/Sub** | Cross-service event distribution | Internal |
| **Database Polling** | Drift detection (periodic) | SQL |

---

## gRPC Architecture

### Why gRPC

- **Strongly typed contracts** — Service interfaces are defined in protobuf, generating both server and client code
- **Performance** — HTTP/2 multiplexing, binary encoding, streaming built-in
- **Code generation** — TypeScript types generated for frontend, Go types for backend
- **Streaming** — Native support for real-time event streaming without WebSocket server management

### Service Definitions

```
proto/
├── auth/
│   └── v1/
│       └── auth.proto          # Auth service
├── project/
│   └── v1/
│       └── project.proto       # Project service
├── schema/
│   └── v1/
│       └── schema.proto        # Schema service
├── migration/
│   └── v1/
│       └── migration.proto     # Migration service
├── event/
│   └── v1/
│       └── event.proto         # Event service
├── audit/
│   └── v1/
│       └── audit.proto         # Audit service
├── common/
│   └── v1/
│       ├── common.proto        # Shared types
│       └── pagination.proto    # Pagination messages
└── google/
    └── api/
        └── annotations.proto   # Google API annotations
```

### Interceptor Chain

Each gRPC request passes through the following interceptors (in order):

1. **Recovery Interceptor** — Catches panics, converts to gRPC errors
2. **Logging Interceptor** — Logs request method, duration, status
3. **Auth Interceptor** — Validates JWT (except for public endpoints)
4. **Rate Limit Interceptor** — Applies rate limiting per user/IP
5. **Validation Interceptor** — Validates request message fields

---

## Internal Package Structure

```
backend/
├── cmd/
│   └── server/
│       └── main.go                 # Service entry point with DI wiring
├── internal/
│   ├── auth/
│   │   ├── domain/
│   │   ├── repository/
│   │   ├── handler/
│   │   └── middleware/
│   ├── project/
│   │   ├── domain/
│   │   ├── repository/
│   │   └── handler/
│   ├── schema/
│   │   ├── domain/
│   │   ├── repository/
│   │   └── handler/
│   ├── migration/
│   │   ├── domain/
│   │   ├── repository/
│   │   └── handler/
│   ├── event/
│   │   ├── domain/
│   │   └── handler/
│   ├── audit/
│   │   ├── domain/
│   │   ├── repository/
│   │   └── handler/
│   ├── drift/
│   │   ├── domain/
│   │   └── repository/
│   └── pkg/
│       ├── config/                 # Configuration loading
│       ├── database/               # DB connection pool, migrations
│       ├── redis/                  # Redis client, pub/sub helpers
│       ├── jwt/                    # JWT creation and validation
│       ├── interceptor/            # Shared gRPC interceptors
│       ├── logger/                 # Structured logging
│       ├── middleware/             # Shared HTTP/gRPC middleware
│       └── errors/                 # Error types and mapping
├── pkg/                            # Public shared packages
├── proto/                          # Generated protobuf Go code
└── go.mod
```

---

## Database Interaction

### Connection Pooling

- **pgx pool** — Connection pooling via `pgx/v5/pgxpool`
- Pool size: min 2, max 20 per service instance
- Connection health checks every 30 seconds
- Statement timeout: 30s for queries, 120s for migrations
- All queries use prepared statements to prevent SQL injection

### Transaction Management

- **Read operations** — Use read replicas when available (Neon read replicas)
- **Write operations** — Use primary with explicit transaction control
- **Migration execution** — Each migration runs in its own transaction
- **Audit logging** — Audit entries are written in the same transaction as the operation, or independently if that is not possible

### Repository Pattern

```go
// Each repository is defined as an interface
type SchemaRepository interface {
    Create(ctx context.Context, schema *domain.Schema) error
    GetByID(ctx context.Context, id string) (*domain.Schema, error)
    GetByProjectID(ctx context.Context, projectID string) ([]*domain.Schema, error)
    CreateVersion(ctx context.Context, version *domain.SchemaVersion) error
    GetVersions(ctx context.Context, schemaID string) ([]*domain.SchemaVersion, error)
}
```

---

## Streaming Architecture

```
                    ┌─────────────────────────────┐
                    │      Schema Service          │
                    │                               │
                    │  ┌───────────────────────┐   │
                    │  │   Migration Executor  │   │
                    │  └──────────┬────────────┘   │
                    │             │                 │
                    │             ▼                 │
                    │  ┌───────────────────────┐   │
                    │  │   Event Publisher     │   │
                    │  └──────────┬────────────┘   │
                    └─────────────┼────────────────┘
                                  │ Publish
                                  ▼
                    ┌─────────────────────────────┐
                    │         Redis Pub/Sub        │
                    │  ┌───────────────────────┐   │
                    │  │ schema:events:{projID} │   │
                    │  │ audit:events          │   │
                    │  │ notifications:{userID}│   │
                    │  └───────────────────────┘   │
                    └─────────────────────────────┘
                                  │ Subscribe
                    ┌─────────────┼────────────────┐
                    │             ▼                 │
                    │  ┌───────────────────────┐   │
                    │  │   Event Service        │   │
                    │  │  ┌─────────────────┐   │   │
                    │  │  │ Stream Manager  │   │   │
                    │  │  │ - Per-client    │   │   │
                    │  │  │   subscriptions │   │   │
                    │  │  │ - Connection    │   │   │
                    │  │  │   lifecycle     │   │   │
                    │  │  └─────────────────┘   │   │
                    │  └───────────────────────┘   │
                    └─────────────────────────────┘
                                  │ gRPC Server Stream
                                  ▼
                    ┌─────────────────────────────┐
                    │         Clients              │
                    │  (Browser, CLI, API)         │
                    └─────────────────────────────┘
```

---

## Event Flow

### Event Types

| Event | Producer | Consumers |
|---|---|---|
| `SchemaVersionCreated` | Schema Service | Event Service, Audit Service, Drift Service |
| `MigrationStarted` | Migration Service | Event Service, Audit Service |
| `MigrationCompleted` | Migration Service | Event Service, Audit Service |
| `MigrationFailed` | Migration Service | Event Service, Audit Service |
| `DriftDetected` | Drift Service | Event Service, Audit Service |
| `ConnectionCreated` | Project Service | Audit Service |
| `UserRoleChanged` | Auth Service | Audit Service |

### Event Schema

```json
{
    "id": "evt_01JABCDEFGHIJKLMNOPQRSTUV",
    "type": "MigrationCompleted",
    "version": "1.0",
    "timestamp": "2026-07-22T14:30:00.000Z",
    "actor": {
        "id": "usr_01JABCDEFGHIJKLMNOPQRSTUV",
        "email": "user@example.com"
    },
    "resource": {
        "type": "migration",
        "id": "mig_01JABCDEFGHIJKLMNOPQRSTUV"
    },
    "payload": {
        "project_id": "proj_01JABCDEFGHIJKLMNOPQRSTUV",
        "migration_id": "mig_01JABCDEFGHIJKLMNOPQRSTUV",
        "status": "completed",
        "duration_ms": 1234
    },
    "metadata": {
        "trace_id": "trace_01JABCDEFGHIJKLMNOPQRSTUV"
    }
}
```

### Event Flow Diagram

```
┌──────────┐   ┌──────────┐   ┌──────────┐   ┌──────────┐   ┌──────────┐
│ Service  │──►│ Event    │──►│ Redis    │──►│ Event    │──►│ Client   │
│ (mutates)│   │Publisher │   │ Pub/Sub  │   │ Service  │   │ (stream) │
└──────────┘   └────┬─────┘   └──────────┘   └────┬─────┘   └──────────┘
                    │                              │
                    │                              │
                    ▼                              ▼
            ┌──────────────┐              ┌──────────────┐
            │  Audit Table │              │  Drift       │
            │  (Postgres)  │              │  Service     │
            └──────────────┘              └──────────────┘
```

---

## Error Handling

### Error Classification

| Category | Examples | HTTP/gRPC Code |
|---|---|---|
| **Validation** | Invalid input, missing required fields | `InvalidArgument` (400) |
| **Authentication** | Missing/expired JWT, invalid token | `Unauthenticated` (401) |
| **Authorization** | Insufficient permissions | `PermissionDenied` (403) |
| **Not Found** | Resource does not exist | `NotFound` (404) |
| **Conflict** | Duplicate resource, version conflict | `AlreadyExists` (409) |
| **Rate Limited** | Too many requests | `ResourceExhausted` (429) |
| **Internal** | Database failure, unexpected error | `Internal` (500) |
| **Unavailable** | Service temporarily unavailable | `Unavailable` (503) |

### Error Response Format

```protobuf
message ErrorResponse {
    string code = 1;                    // Machine-readable error code
    string message = 2;                 // Human-readable message
    map<string, string> details = 3;    // Field-level validation errors
    string request_id = 4;              // For correlation
}
```

### Error Wrapping Pattern (Go)

```go
// Domain errors are defined as sentinel values
var ErrProjectNotFound = errors.New("project not found")
var ErrMigrationInProgress = errors.New("migration already in progress")

// Services wrap errors with context
func (s *Service) ExecuteMigration(ctx context.Context, req *pb.ExecuteMigrationRequest) error {
    project, err := s.repo.GetProject(ctx, req.ProjectId)
    if err != nil {
        return fmt.Errorf("fetching project: %w", err)
    }
    if project.Status == domain.ProjectStatusMigrating {
        return fmt.Errorf("%w for project %s", ErrMigrationInProgress, project.ID)
    }
    // ...
}
```

---

## Logging

### Structured Logging

All logs use **log/slog** (Go 1.21+) with JSON output format:

```json
{
    "time": "2026-07-22T14:30:00.000Z",
    "level": "INFO",
    "msg": "migration completed",
    "service": "migration",
    "trace_id": "trace_01JABCDEFGHIJKLMNOPQRSTUV",
    "user_id": "usr_01JABCDEFGHIJKLMNOPQRSTUV",
    "project_id": "proj_01JABCDEFGHIJKLMNOPQRSTUV",
    "migration_id": "mig_01JABCDEFGHIJKLMNOPQRSTUV",
    "duration_ms": 1234,
    "status": "completed"
}
```

### Log Levels

| Level | Usage |
|---|---|
| `DEBUG` | Detailed debugging — not enabled in production |
| `INFO` | Normal operations — request starts/completes, state changes |
| `WARN` | Unexpected but handled — rate limit approaching, slow query |
| `ERROR` | Operation failures — migration failure, DB connection loss |
| `FATAL` | Unrecoverable — service cannot start, config error |

### Correlation

Every request gets a `trace_id` (UUID v7) that propagates through gRPC metadata and is included in all log entries and database operations.

---

## Scalability Considerations

### Horizontal Scaling

- **Backend services** are stateless and scale horizontally behind the gRPC gateway
- **PostgreSQL (Neon)** provides serverless scaling with automatic read replicas
- **Redis** can be clustered for high availability
- **Event subscribers** are load-balanced across service instances

### Performance Budgets

| Operation | Target Latency (P99) |
|---|---|
| API CRUD operations | < 200ms |
| Schema introspection (small DB) | < 3s |
| Schema introspection (large DB, 500+ tables) | < 10s |
| Migration execution (per statement) | < 1s |
| Schema diff computation | < 2s |
| Event delivery (end-to-end) | < 100ms |
| Audit log query (with pagination) | < 500ms |

### Bottleneck Mitigation

| Bottleneck | Mitigation |
|---|---|
| Schema introspection on large databases | Streaming results, parallel table introspection, paginated responses |
| Migration execution blocking | Non-blocking execution with status polling |
| Event delivery to many clients | Redis pub/sub fan-out, per-client buffering |
| Audit log volume | Partitioned tables, time-based retention, archival |
| Concurrent introspection requests | Connection pooling limits, request queuing |

### Caching Strategy

| Cache | Key Pattern | TTL | Invalidation |
|---|---|---|---|
| Schema metadata | `schema:{id}` | 5 minutes | On schema version creation |
| Project list | `user:{id}:projects` | 1 minute | On project CRUD |
| Connection status | `conn:{id}:status` | 30 seconds | On test connection |
| User permissions | `user:{id}:perms:{projID}` | 5 minutes | On role change |
| Introspection results | `introspect:{connID}:schema` | 1 minute | On manual refresh |
