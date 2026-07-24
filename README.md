# SchemaHub

> **"GitHub for Database Schemas"** — Centralized, version-controlled, collaborative PostgreSQL schema management.

SchemaHub is a production-grade developer platform for backend engineers and DBAs to manage database schemas with the same rigor as source code on GitHub. Track every change, review every migration, audit every operation.

---

## Tech Stack

![Go](https://img.shields.io/badge/Go-00ADD8?style=for-the-badge&logo=go&logoColor=white)
![gRPC](https://img.shields.io/badge/gRPC-4285F4?style=for-the-badge&logo=grpc&logoColor=white)
![Protocol Buffers](https://img.shields.io/badge/Protobuf-34A853?style=for-the-badge&logo=protocolbuffers&logoColor=white)
![PostgreSQL](https://img.shields.io/badge/PostgreSQL-316192?style=for-the-badge&logo=postgresql&logoColor=white)
![Redis](https://img.shields.io/badge/Redis-DC382D?style=for-the-badge&logo=redis&logoColor=white)
![Next.js](https://img.shields.io/badge/Next.js-000000?style=for-the-badge&logo=nextdotjs&logoColor=white)
![React](https://img.shields.io/badge/React-20232A?style=for-the-badge&logo=react&logoColor=61DAFB)
![TypeScript](https://img.shields.io/badge/TypeScript-007ACC?style=for-the-badge&logo=typescript&logoColor=white)
![Tailwind CSS](https://img.shields.io/badge/Tailwind-06B6D4?style=for-the-badge&logo=tailwindcss&logoColor=white)
![Docker](https://img.shields.io/badge/Docker-2496ED?style=for-the-badge&logo=docker&logoColor=white)
![JWT](https://img.shields.io/badge/JWT-000000?style=for-the-badge&logo=jsonwebtokens&logoColor=white)
![Vercel](https://img.shields.io/badge/Vercel-000000?style=for-the-badge&logo=vercel&logoColor=white)
![Neon](https://img.shields.io/badge/Neon-00E59B?style=for-the-badge&logo=neon&logoColor=white)

---

## Architecture

```mermaid
flowchart TB
    subgraph Frontend["Frontend (Next.js)"]
        A[Next.js App Router]
        B[React + TypeScript]
        C[Tailwind + shadcn/ui]
        D[React Flow]
        E[TanStack Query]
        F[gRPC-Web Client]
    end

    subgraph Envoy["API Gateway (Envoy)"]
        G[TLS Termination]
        H[gRPC-Web Proxy]
        I[Rate Limiting]
        J[CORS]
    end

    subgraph Backend["Backend (Go + gRPC)"]
        K[Auth Service]
        L[Project Service]
        M[Schema Service]
        N[Migration Service]
        O[Event Service]
        P[Audit Service]
        Q[Drift Service]
    end

    subgraph Data["Data Layer"]
        R[PostgreSQL<br/>Neon]
        S[Redis<br/>Cache + Pub/Sub]
    end

    A -->|gRPC-Web| G
    G -->|gRPC| Backend
    K --> R
    L --> R
    M --> R
    N --> R
    O --> S
    P --> R
    Q --> R
    M -.-> S
    N -.-> S
    O -.-> R
```

---

## Features

| Feature | Description |
|---|---|
| **Project Management** | Create projects, manage members, set roles (owner/admin/member/viewer) |
| **Connection Management** | Store encrypted DB credentials, async connectivity testing |
| **Schema Exploration** | Browse tables, columns, indexes, constraints via PostgreSQL introspection |
| **Schema Versioning** | Content-addressed immutable versions with SHA-256 checksum dedup |
| **Schema Diff** | Side-by-side comparison of any two schema versions |
| **Schema Diagrams** | Interactive ERD diagrams via React Flow with Dagre layout |
| **Migration Engine** | Create, validate, execute, and roll back SQL migrations |
| **Progress Streaming** | Real-time migration status via gRPC server-streaming |
| **Validation Engine** | Pre-execution SQL validation with disallowed statement detection |
| **Dry-Run** | Test migrations against live DB without committing changes |
| **Rollback Support** | Execute down SQL with full history and failure handling |
| **Version Timeline** | Visual history of schema changes |
| **Real-Time Events** | Live notifications via Redis pub/sub + gRPC streams |
| **Presence Tracking** | Online user indicators with heartbeat management |
| **Event Replay** | Reconnection recovery with last-event-id |
| **Audit Logging** | Immutable, partitioned, filterable audit trail for every operation |
| **Drift Detection** | Automated comparison of live DB vs tracked schema version |
| **Drift Alerting** | Periodic drift checks with severity classification |
| **RBAC** | Two-level authorization: global roles + project roles |
| **OAuth Login** | Google, GitHub, Slack social login with PKCE + account linking |

---

## User Flows

### Authentication Flow

```mermaid
sequenceDiagram
    actor U as User
    participant F as Frontend
    participant B as Backend
    participant P as PostgreSQL
    participant R as Redis

    U->>F: Enter email + password
    F->>B: LoginRequest
    B->>P: Query user by email
    P-->>B: User record
    B->>B: Verify bcrypt hash
    B->>B: Generate JWT (15min) + Refresh Token (7d)
    B->>P: Store refresh token hash
    B->>R: Rate limit check
    B-->>F: LoginResponse{access_token, refresh_token}
    F-->>U: Redirect to dashboard
```

### Migration Execution Flow

```mermaid
sequenceDiagram
    actor U as User
    participant F as Frontend
    participant B as Backend
    participant P as PostgreSQL
    participant R as Redis

    U->>F: Select migration + connection → Execute
    F->>B: ExecuteMigrationRequest
    B->>B: Validate permissions + migration status
    B->>B: Check no concurrent run on connection
    B->>P: Create MigrationRun (pending)
    B-->>F: ExecuteMigrationResponse{run_id}

    par Async Execution
        B->>P: BEGIN transaction
        B->>P: Execute statement 1
        B->>P: Log to migration_logs
        B-->>F: Stream: MigrationStatus{1/3, RUNNING}
        B->>P: Execute statement 2
        B->>P: Log to migration_logs
        B-->>F: Stream: MigrationStatus{2/3, RUNNING}
        B->>P: Execute statement 3
        B->>P: Log to migration_logs
        B->>P: COMMIT
        B-->>F: Stream: MigrationStatus{3/3, COMPLETED}
        B->>R: Publish MigrationCompleted event
    end
```

### Schema Introspection Flow

```mermaid
sequenceDiagram
    actor U as User
    participant F as Frontend
    participant B as Backend
    participant P as PostgreSQL (Target DB)
    participant S as PostgreSQL (SchemaHub)
    participant R as Redis

    U->>F: Select connection → "Introspect"
    F->>B: IntrospectSchemaRequest
    B->>B: Decrypt stored credentials
    B->>P: Query information_schema.tables
    B->>P: Query information_schema.columns
    B->>P: Query pg_indexes
    B->>P: Query table_constraints
    B->>P: Query pg_enum
    B->>P: Query pg_extension
    B->>B: Build JSONB metadata
    B->>B: Compute SHA-256 checksum
    B->>S: Check if checksum exists
    alt New Version
        B->>S: Insert schema_version
        B->>S: Insert schema_objects
        B->>S: Update schema.current_version_id
        B->>R: Cache metadata (TTL: 5min)
        B->>R: Publish SchemaVersionCreated event
    else Duplicate
        B->>S: Link to existing version
    end
    B-->>F: SchemaVersion response
    F-->>U: Render schema tree
```

---

## Project Structure

```
schemahub/
├── frontend/          # Next.js application (App Router)
│   └── src/
│       ├── app/       # Route groups, pages, layouts
│       ├── components/# Reusable UI components
│       ├── lib/       # gRPC clients, hooks, utilities
│       └── styles/    # Tailwind config, global CSS
├── backend/           # Go mono-repo (7 services)
│   ├── cmd/server/    # Entry point with DI wiring
│   ├── internal/      # Domain logic, repos, handlers
│   │   ├── auth/      # Authentication + OAuth (Google/GitHub/Slack)
│   │   ├── project/   # Projects, members, connections
│   │   ├── schema/    # Introspection, versioning, diff, diagrams
│   │   ├── migration/ # Execution, validation, rollback, streaming
│   │   ├── event/     # Redis pub/sub, subscriptions, presence
│   │   ├── audit/     # Immutable partitioned audit logs
│   │   ├── drift/     # Drift detection, alerting, resolution
│   │   └── pkg/       # Shared: config, DB, JWT, interceptor, RBAC, workers
│   ├── proto/         # Generated protobuf Go code
│   └── pkg/encryption/# AES-256-GCM credential encryption
├── proto/             # Source .proto files (service contracts)
├── docker/            # Compose + Dockerfiles + Envoy + Redis
├── scripts/           # PowerShell dev tooling (7 scripts)
├── .github/           # CI/CD workflows + Dependabot
└── docs/              # 20 documentation files
```

---

## Quick Start

```bash
# Prerequisites: Go 1.22+, Node 20+, Docker

# Start full stack
docker compose -f docker/docker-compose.yml -f docker/docker-compose.dev.yml up

# Or use scripts
.\scripts\setup.ps1        # Check prerequisites, install tools
.\scripts\dev.ps1           # Start dev environment
.\scripts\lint.ps1          # Run all linters
.\scripts\test.ps1          # Run all tests
.\scripts\gen-proto.ps1     # Regenerate protobuf code
```

---

## Services

| Service | Role | RPCs |
|---|---|---|
| **Auth** | Register, login, JWT, OAuth | 15 |
| **Project** | Project CRUD, members, connections | 15 |
| **Schema** | Introspection, versions, diff, diagrams | 10 |
| **Migration** | CRUD, execute, rollback, validate, dry-run, streaming | 13 |
| **Event** | Subscribe, heartbeat, acknowledge | 3 |
| **Audit** | List, get, tail stream, stats | 4 |
| **Drift** | Check, list, resolve, stats | 5 |
| **Total** | | **65 gRPC RPCs** |

---

## Documentation

| Document | Description |
|---|---|
| [Project Context](docs/PROJECT_CONTEXT.md) | Vision, goals, terminology, design decisions |
| [Architecture](docs/ARCHITECTURE.md) | System architecture, request lifecycle, event flow |
| [Tech Stack](docs/TECH_STACK.md) | Technology choices with trade-off analysis |
| [Database Design](docs/DATABASE_DESIGN.md) | Complete schema design, indexes, relationships |
| [ER Diagram](docs/ER_DIAGRAM.md) | Entity-relationship visualization (Mermaid) |
| [gRPC Design](docs/GRPC_DESIGN.md) | gRPC architecture, streaming, error handling |
| [Protobuf Contracts](docs/PROTOBUF_CONTRACTS.md) | Service definitions, messages, versioning |
| [API Flow](docs/API_FLOW.md) | End-to-end request flows for every feature |
| [Authentication](docs/AUTHENTICATION.md) | JWT, refresh tokens, RBAC, token security |
| [OAuth Integration](docs/OAUTH_INTEGRATION.md) | Google/GitHub/Slack, PKCE, account linking |
| [Real-Time Architecture](docs/REALTIME_ARCHITECTURE.md) | Streaming, presence, notifications |
| [Folder Structure](docs/FOLDER_STRUCTURE.md) | Repository layout with explanations |
| [Coding Guidelines](docs/CODING_GUIDELINES.md) | Go/TS style, naming, error handling |
| [Feature Specifications](docs/FEATURE_SPECIFICATIONS.md) | 14 features with workflows |
| [Roadmap](docs/ROADMAP.md) | 9-phase development plan |
| [Security](docs/SECURITY.md) | Auth, secrets, rate limiting, transport security |
| [Deployment](docs/DEPLOYMENT.md) | Docker, env vars, CI/CD, monitoring |
| [Testing Strategy](docs/TESTING_STRATEGY.md) | Unit, integration, e2e, performance |
| [Remaining Work](docs/REMAINING.md) | Current completion audit |

---

## License

MIT
