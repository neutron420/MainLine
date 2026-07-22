# SchemaHub

> **"GitHub for Database Schemas" — Centralized schema management for modern engineering teams.**

SchemaHub is a production-grade developer platform that enables backend engineers and database administrators to manage PostgreSQL database schemas in a centralized, version-controlled, and collaborative environment. It is not an ORM, not a database, and not a SQL editor — it is a **schema management platform**.

---

## Features

| Feature | Description |
|---|---|
| **Project Management** | Create and organize database schema projects |
| **Schema Exploration** | Browse tables, columns, indexes, constraints, and relationships |
| **Version Tracking** | Every schema change is versioned and immutable |
| **Migration Engine** | Execute, validate, and track database migrations |
| **Schema Comparison** | Diff any two schema versions side-by-side |
| **Rollback Support** | Roll back migrations with full history preservation |
| **Real-Time Updates** | Live schema change notifications via WebSocket streams |
| **Schema Drift Detection** | Detect when a database diverges from its tracked state |
| **Visual Diagrams** | Interactive entity-relationship diagrams via React Flow |
| **Audit Logging** | Immutable, queryable audit trail for every operation |
| **Collaboration** | Team-based access control with role-based permissions |
| **Migration Monitoring** | Track execution time, status, and failure details |

---

## Architecture Overview

```
┌─────────────────────────────────────────────────────────────┐
│                        Frontend (Next.js)                    │
│  React · TypeScript · Tailwind · shadcn/ui · React Flow    │
│  TanStack Query (REST) + WebSocket (Real-time)              │
└────────────────────┬────────────────────────────────────────┘
                     │ HTTP/gRPC-Web
                     ▼
┌─────────────────────────────────────────────────────────────┐
│                     Backend (Go + gRPC)                      │
│  Services: Project · Schema · Migration · Auth · Audit      │
│  Streaming: Schema Events · Notifications · Presence        │
└──────┬──────────────────────┬───────────────────────────────┘
       │                      │
       ▼                      ▼
┌──────────────┐    ┌──────────────────┐
│  PostgreSQL   │    │      Redis       │
│  (Neon)       │    │  Caching · Pub   │
│  Primary DB   │    │  Sub · Sessions  │
└──────────────┘    └──────────────────┘
```

---

## Tech Stack

| Layer | Technology |
|---|---|
| Frontend | Next.js, React, TypeScript, Tailwind CSS, shadcn/ui, React Flow, TanStack Query |
| Backend | Go, gRPC, Protocol Buffers |
| Database | PostgreSQL (Neon) |
| Cache | Redis |
| Auth | JWT (access + refresh tokens) |
| Containerization | Docker, Docker Compose |
| Frontend Deployment | Vercel |
| Backend Deployment | Railway / Fly.io / AWS |

---

## Why SchemaHub Exists

Engineering teams manage hundreds of database schema changes across multiple environments. Without a centralized platform:

- Schema changes are undocumented
- Migrations are run ad-hoc from local machines
- Rollbacks are manual and error-prone
- Audit trails do not exist
- Collaboration happens over Slack or PR comments
- Schema drift goes undetected until production breaks

SchemaHub solves these problems by bringing database schema management into the same workflow paradigm as code management on GitHub.

---

## Documentation

| Document | Purpose |
|---|---|
| [Project Context](docs/PROJECT_CONTEXT.md) | Vision, goals, terminology, design philosophy |
| [Architecture](docs/ARCHITECTURE.md) | System architecture, request lifecycle, event flow |
| [Tech Stack](docs/TECH_STACK.md) | Technology decisions with trade-off analysis |
| [Database Design](docs/DATABASE_DESIGN.md) | Complete schema design, indexing, relationships |
| [ER Diagram](docs/ER_DIAGRAM.md) | Entity-relationship visualization |
| [gRPC Design](docs/GRPC_DESIGN.md) | gRPC architecture, streaming patterns, error handling |
| [Protobuf Contracts](docs/PROTOBUF_CONTRACTS.md) | Service definitions, message contracts, versioning |
| [API Flow](docs/API_FLOW.md) | End-to-end request flows for every feature |
| [Authentication](docs/AUTHENTICATION.md) | JWT, refresh tokens, RBAC, security |
| [Real-Time Architecture](docs/REALTIME_ARCHITECTURE.md) | WebSocket streaming, presence, notifications |
| [Folder Structure](docs/FOLDER_STRUCTURE.md) | Repository layout and conventions |
| [Coding Guidelines](docs/CODING_GUIDELINES.md) | Go style, naming, error handling, commits |
| [Feature Specifications](docs/FEATURE_SPECIFICATIONS.md) | Detailed feature definitions and workflows |
| [Roadmap](docs/ROADMAP.md) | Phased development plan with deliverables |
| [Security](docs/SECURITY.md) | Authentication, authorization, hardening |
| [Deployment](docs/DEPLOYMENT.md) | Dev/staging/production, Docker, monitoring |
| [Testing Strategy](docs/TESTING_STRATEGY.md) | Unit, integration, e2e, performance testing |
| [Future Ideas](docs/FUTURE_IDEAS.md) | AI analysis, CLI, VS Code extension, plugins |
| [Contributing](docs/CONTRIBUTING.md) | Contribution guidelines for open-source |

---

## Project Structure

```
schemahub/
├── frontend/          # Next.js application
│   ├── src/
│   │   ├── app/       # App router pages
│   │   ├── components/# Shared UI components
│   │   ├── lib/       # Utilities, hooks, API clients
│   │   └── styles/    # Global styles
│   └── package.json
├── backend/           # Go gRPC services
│   ├── cmd/           # Entry points
│   ├── internal/      # Private application code
│   ├── pkg/           # Shared libraries
│   ├── proto/         # Protocol Buffers definitions
│   └── go.mod
├── proto/             # Shared protobuf definitions
├── docs/              # Documentation
├── docker/            # Docker Compose files
├── scripts/           # Dev tooling scripts
└── infra/             # Infrastructure as Code
```

---

## Development Roadmap

| Phase | Focus |
|---|---|
| **Phase 1** | Foundation — Project scaffold, Docker, build toolchain |
| **Phase 2** | Authentication — JWT, registration, login, RBAC |
| **Phase 3** | Schema Explorer — Connect DB, browse schemas, introspect |
| **Phase 4** | Migration Engine — Execute, validate, track migrations |
| **Phase 5** | Version History — Compare, diff, rollback |
| **Phase 6** | Real-Time Features — WebSocket streams, notifications |
| **Phase 7** | Visual Schema Explorer — React Flow diagrams |
| **Phase 8** | Deployment — CI/CD, monitoring, staging/production |
| **Phase 9** | Advanced Features — Drift detection, CLI, plugins |

---

## Future Vision

SchemaHub aims to become the standard platform for database schema management across the industry. Long-term goals include:

- **AI-powered migration analysis** — automatic rollback recommendations and impact analysis
- **VS Code Extension** — manage schemas without leaving the editor
- **CLI Tool** — scriptable schema management for CI/CD pipelines
- **Plugin Ecosystem** — extensible hooks for custom validators, notifiers, and integrations
- **Multi-Database Support** — MySQL, SQLite, SQL Server beyond PostgreSQL
- **Schema Registry** — discoverable catalog of reusable schema patterns
- **Schema Simulation** — dry-run migrations on synthetic data to predict impact

---

## License

MIT
