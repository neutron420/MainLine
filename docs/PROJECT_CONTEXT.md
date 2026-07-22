# Project Context

> **Single source of truth for all SchemaHub design decisions, conventions, and constraints.**

---

## Table of Contents

- [Vision](#vision)
- [Goals](#goals)
- [Problem Statement](#problem-statement)
- [Business Value](#business-value)
- [Target Users](#target-users)
- [Terminology](#terminology)
- [Core Concepts](#core-concepts)
- [Major Modules](#major-modules)
- [Current Project Status](#current-project-status)
- [Future Plans](#future-plans)
- [Constraints](#constraints)
- [Coding Philosophy](#coding-philosophy)
- [Documentation Philosophy](#documentation-philosophy)
- [Design Principles](#design-principles)
- [Engineering Standards](#engineering-standards)

---

## Vision

SchemaHub aims to become the industry standard for database schema management — a centralized, version-controlled, and collaborative platform that brings the same rigor to database schemas that GitHub brought to source code.

We envision a world where every database schema change is tracked, reviewed, auditable, and reversible. Where schema management is as natural and structured as code management. Where teams never again ask "who changed that column?" or "what migration is running on production?"

---

## Goals

1. **Centralization** — Provide a single source of truth for all database schema definitions across an organization
2. **Version Control** — Track every schema change with immutable, timestamped versions that can be compared, diffed, and rolled back
3. **Collaboration** — Enable teams to work on schema changes together with proper access controls and audit trails
4. **Safety** — Ensure migrations are validated, reversible, and monitored to prevent production incidents
5. **Visibility** — Give engineers and DBAs real-time insight into schema state, drift, and migration execution
6. **Extensibility** — Design for plugins, CLI integration, and CI/CD pipeline compatibility from day one

---

## Problem Statement

Modern engineering teams manage hundreds of microservices, each with its own database. Schema changes are:

- **Undocumented** — There is no canonical source of truth for what the schema looks like
- **Ad-hoc** — Migrations are run from local machines, CI pipelines, or manual scripts
- **Unsafe** — Rollbacks are manual, poorly tested, and often fail
- **Invisible** — No one knows what schemas exist across the organization
- **Uncoordinated** — Multiple teams modify schemas without visibility into each other's work
- **Undetectable** — Schema drift goes unnoticed until it causes production outages

Current solutions are fragmented: ORMs provide limited schema awareness, dedicated migration tools lack collaboration features, and spreadsheets or README files are not scalable.

---

## Business Value

| Value Driver | Impact |
|---|---|
| **Reduced Incidents** | Validated, reversible migrations prevent schema-related outages |
| **Faster Onboarding** | New engineers can immediately see all database schemas |
| **Audit Compliance** | Immutable audit trail satisfies SOC2, HIPAA, and SOX requirements |
| **Team Productivity** | No more Slack messages asking "what columns does this table have?" |
| **Drift Prevention** | Automated drift detection catches unauthorized changes |
| **Standardization** | Consistent migration patterns across all services |

---

## Target Users

| Persona | Needs |
|---|---|
| **Backend Engineer** | Explore schemas, run migrations, view history, collaborate on changes |
| **Database Administrator** | Monitor all schemas, audit changes, detect drift, enforce standards |
| **Platform Engineer** | Integrate SchemaHub into CI/CD pipelines, manage access controls |
| **Engineering Manager** | Visibility into schema change velocity, compliance, team collaboration |
| **DevOps/SRE** | Monitor migration execution, rollback coordination, incident response |

---

## Terminology

| Term | Definition |
|---|---|
| **Project** | A top-level container for related schemas and migrations |
| **Schema** | The structure of a database — tables, columns, indexes, constraints |
| **Schema Version** | An immutable snapshot of a schema at a point in time |
| **Migration** | A set of SQL statements that transform one schema version to another |
| **Migration Run** | A single execution of a migration against a target database |
| **Drift** | When a live database schema differs from its tracked version |
| **Introspection** | The process of reading schema metadata from a live database |
| **Snapshot** | A point-in-time capture of full schema metadata |
| **Workspace** | An isolated environment for testing migrations before production |
| **Connection** | Stored credentials and metadata for a target PostgreSQL database |

---

## Core Concepts

### Schema Versioning

Every schema change produces an immutable version. Versions are content-addressed by a SHA-256 hash of the schema metadata, ensuring that identical schemas produce identical version identifiers.

### Migration as Transaction

Migrations are treated as atomic transactions. If any statement in a migration fails, the entire migration is rolled back, and the error is captured for analysis.

### Stream-First Architecture

Schema changes produce events that are streamed to connected clients via gRPC server-streaming and WebSocket bridges. Every mutation is an event, and every event is persisted in the audit log.

### Introspection-Driven

SchemaHub connects to live databases and introspects their structure. It never assumes — it always reads the actual schema from the database and compares it to known versions.

---

## Major Modules

| Module | Responsibility |
|---|---|
| **Auth Service** | User registration, authentication, JWT issuance, RBAC |
| **Project Service** | CRUD for projects, member management, settings |
| **Connection Service** | Database connection management, credential encryption, connectivity testing |
| **Schema Service** | Introspection, version creation, diff computation, diagram data |
| **Migration Service** | Migration execution, validation, rollback, history |
| **Event Service** | Real-time event streaming, notification dispatch |
| **Audit Service** | Immutable audit log ingestion, querying, retention |
| **Drift Service** | Drift detection, alerting, remediation workflows |

---

## Current Project Status

| Aspect | Status |
|---|---|
| Documentation | Complete — all design documents are finalized |
| Architecture Design | Complete — services, data flow, and infrastructure are defined |
| Protobuf Contracts | Designed — service boundaries and message schemas are documented |
| Database Design | Complete — all entities, relationships, and indexes are designed |
| Implementation | Not started — waiting for documentation approval |
| CI/CD | Not configured — to be built during Phase 1 |

---

## Future Plans

### v1.0 (Initial Release)

- Full project and connection management
- Schema introspection and versioning
- Migration execution engine with rollback
- Real-time event streaming
- Schema visualization
- RBAC and audit logging

### v2.0 (Advanced)

- AI-powered migration analysis and rollback recommendations
- CLI tool for CI/CD integration
- VS Code extension
- Multi-database support (MySQL, SQLite)
- Plugin ecosystem

### v3.0 (Enterprise)

- Schema registry with discoverable patterns
- Multi-region replication
- On-premises deployment option
- Advanced compliance reporting
- Custom workflow automation

---

## Constraints

### Immutable Decisions

The following decisions are locked and will not be reconsidered without a formal RFC process:

1. **Go for backend** — No replacement of Go with another language
2. **gRPC for API** — No REST-only API (gRPC-Web for browser compatibility)
3. **PostgreSQL (Neon) as primary database** — No multi-DB support in v1
4. **Next.js for frontend** — No replacement of the frontend framework
5. **Protocol Buffers for contracts** — No alternative serialization formats
6. **JWT for auth** — No session-based authentication

### Design Constraints

1. **No SQL migrations in documentation** — Document schema design conceptually, not as SQL scripts
2. **No actual `.proto` files in documentation** — Document contracts conceptually
3. **No application code in doc files** — Documentation only
4. **Neon-compatible** — Must support serverless PostgreSQL with branching features
5. **Documentation first** — No implementation code without corresponding documentation

---

## Coding Philosophy

### Production-Grade from Day One

- Every line of code is written as if it will run in production tomorrow
- No shortcuts, no TODOs, no technical debt accepted without explicit tracking
- Tests are mandatory for all code paths
- Zero warnings on build — treat warnings as errors

### Simplicity Over Cleverness

- Code should be readable by a junior engineer
- Prefer standard library over external dependencies
- If a pattern is hard to explain, it should be refactored
- Nested abstractions are avoided in favor of flat, obvious structures

### Consistency Over Preference

- All code follows enforced formatters (`go fmt`, `prettier`)
- Patterns established in one part of the codebase apply everywhere
- Style discussions are settled by the linter, not by PR comments
- Configuration over convention where reasonable

---

## Documentation Philosophy

- Documentation is a first-class deliverable, equal in importance to code
- Every architectural decision must include its trade-offs in writing
- Documentation must be detailed enough that another engineer can build the project without asking questions
- Cross-references between documents are mandatory
- Documentation is kept in the same repository as code (docs-as-code)

---

## Design Principles

| Principle | Description |
|---|---|
| **Defense in Depth** | Multiple layers of security at every boundary |
| **Fail Closed** | On error, deny access rather than grant it |
| **Principle of Least Privilege** | Every component gets only the permissions it needs |
| **Idempotency** | Operations should be safe to retry |
| **Observability by Default** | Every operation produces logs, metrics, and traces |
| **Backward Compatibility** | API changes must not break existing clients |
| **Explicit Over Implicit** | Configuration and behavior should be visible and intentional |
| **Separation of Concerns** | Each service owns its domain completely |

---

## Engineering Standards

| Standard | Requirement |
|---|---|
| **Code Review** | Every PR requires at least one approval |
| **Test Coverage** | Minimum 80% for domain logic, 60% overall |
| **Documentation** | Every PR must update relevant docs |
| **Performance Budget** | P99 API response < 500ms, P99 migration introspection < 5s |
| **Security Review** | Auth and data handling changes require security review |
| **Commit Hygiene** | Conventional Commits, squashed merges with linear history |
| **Dependency Management** | Dependencies are pinned and scanned for vulnerabilities |

---

**This document is the single source of truth. All other documents derive from it. If a conflict exists between documents, this document takes precedence.**
