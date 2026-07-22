# CLAUDE.md — Persistent AI Context for SchemaHub

## ⚠️ First Instruction

**Before making any changes to this project, read `docs/PROJECT_CONTEXT.md` thoroughly.** It is the single source of truth for all design decisions, conventions, and constraints.

---

## Project Identity

- **Name:** SchemaHub
- **Tagline:** "GitHub for Database Schemas"
- **Type:** Developer platform for PostgreSQL schema management
- **Not:** An ORM, a database, or a SQL editor
- **Maturity:** Greenfield — all documentation exists; no implementation code exists yet

---

## Core Philosophy

- Design before code
- Production-grade from day one
- Every decision must have a documented trade-off
- Documentation is a first-class deliverable
- No implementation code should be written without corresponding documentation

---

## Tech Stack (Immutable Decisions)

| Component | Choice | Rationale |
|---|---|---|
| Frontend | Next.js + React + TypeScript | Industry standard; server components; Vercel ecosystem |
| Styling | Tailwind CSS + shadcn/ui | Consistent design system; minimal CSS overhead |
| Diagrams | React Flow | Interactive ERD and migration flow visualization |
| State/Server | TanStack Query | Declarative caching; WebSocket integration |
| Backend | Go | Performance; goroutine-based streaming; excellent gRPC support |
| API Protocol | gRPC + Protocol Buffers | Strong typing; streaming; code generation; performance |
| Database | PostgreSQL (Neon) | Mature; JSONB; pgx driver; serverless-ready |
| Cache | Redis | Pub/sub for real-time; session storage; rate limiting |
| Auth | JWT (access + refresh tokens) | Stateless; gRPC-compatible; widely supported |
| Containerization | Docker + Docker Compose | Reproducible environments; CI/CD consistency |
| Frontend Deploy | Vercel | Zero-config Next.js hosting |
| Backend Deploy | Railway / Fly.io / AWS | Flexible; Go-native; gRPC-friendly |

---

## Folder Conventions

```
schemahub/
├── frontend/          # Next.js application (app router)
│   └── src/
│       ├── app/       # Route groups, pages, layouts
│       ├── components/# Reusable UI components
│       ├── lib/       # API clients, hooks, utilities
│       └── styles/    # Global CSS, Tailwind config
├── backend/           # Go mono-repo
│   ├── cmd/           # Binary entry points (one per service)
│   ├── internal/      # Private packages (domain logic, repositories)
│   ├── pkg/           # Shared public packages
│   ├── proto/         # Generated protobuf Go code
│   └── go.mod
├── proto/             # Source .proto files (shared interface contract)
├── docs/              # All project documentation
├── docker/            # Dockerfiles and Compose files
├── scripts/           # Automation scripts (make, husky, pre-commit)
└── infra/             # Terraform / Pulumi IaC
```

---

## Coding Rules

### General

- No implementation code until documentation is approved
- Write tests alongside implementation (TDD preferred)
- Zero warnings on build
- No secrets in code — use environment variables or secret manager
- All public APIs must have godoc comments (Go) or JSDoc (TS)

### Go Specific

- Follow `go fmt` and `go vet` — they are enforced in CI
- Use `internal/` packages to enforce encapsulation boundaries
- Error wrapping with `fmt.Errorf("context: %w", err)` — never swallow errors
- Use `context.Context` as first parameter for all gRPC handlers and database calls
- Domain logic must be in `internal/domain/` — not in handlers
- Repository layer abstracts database access behind interfaces
- gRPC interceptors for auth, logging, recovery, and rate limiting

### TypeScript / React Specific

- Use strict TypeScript mode (`strict: true`)
- Server components by default; client components only when interactivity is needed
- Never use `any` — prefer `unknown` with type guards
- All data fetching through TanStack Query (no raw `useEffect` + `fetch`)
- WebSocket connections managed through a dedicated hook or service

### Git & Commits

- Conventional Commits: `feat:`, `fix:`, `docs:`, `refactor:`, `test:`, `chore:`
- Branch from `main`, squash merge with linear history
- Branch naming: `feature/<name>`, `fix/<name>`, `docs/<name>`
- PR titles match commit conventions

---

## Important Constraints

1. **No SQL migrations in documentation** — document schema design, not SQL generation
2. **No actual `.proto` files in documentation** — document contracts conceptually
3. **No application code in any doc file** — documentation only
4. **Go gRPC backend** — not REST, not GraphQL
5. **PostgreSQL only** — no multi-DB support in v1
6. **Neon-compatible** — must support serverless PostgreSQL with branching
7. **Documentation first** — no code without docs

---

## Key Documents Index

| Document | When to Reference |
|---|---|
| `docs/PROJECT_CONTEXT.md` | Always — the single source of truth |
| `docs/ARCHITECTURE.md` | When designing new components or changing flows |
| `docs/DATABASE_DESIGN.md` | When adding/modifying database entities |
| `docs/GRPC_DESIGN.md` | When designing APIs or streaming logic |
| `docs/PROTOBUF_CONTRACTS.md` | When defining or updating service contracts |
| `docs/API_FLOW.md` | When implementing user-facing features |
| `docs/CODING_GUIDELINES.md` | Before writing any code |
| `docs/TESTING_STRATEGY.md` | Before writing tests |
| `docs/SECURITY.md` | When handling auth or sensitive data |
| `docs/OAUTH_INTEGRATION.md` | When implementing OAuth social login (Google, GitHub, Slack) |
| `docs/DEPLOYMENT.md` | When setting up infrastructure |

---

## Commands Reference (When Implemented)

```bash
# Development
cd backend && go run ./cmd/server          # Start backend
cd frontend && npm run dev                  # Start frontend
docker compose up                           # Full stack locally

# Code Quality
cd backend && go fmt ./...                  # Format Go code
cd backend && go vet ./...                  # Vet Go code
cd backend && golangci-lint run             # Full lint
cd frontend && npm run lint                 # Lint frontend
cd frontend && npm run typecheck            # TypeScript check

# Testing
cd backend && go test ./...                 # Run all Go tests
cd backend && go test -race ./...           # Race detection
cd frontend && npm run test                 # Frontend tests
```

---

## AI Assistant Behavior

When working as an AI assistant on SchemaHub:

1. Always start by reading `docs/PROJECT_CONTEXT.md` to understand full context
2. Reference relevant documentation before proposing changes
3. Never write implementation code unless explicitly asked
4. If asked to write code, follow all conventions in `docs/CODING_GUIDELINES.md`
5. Cross-reference at least 3 documentation files before making architectural decisions
6. Prefer patterns already established in the project over introducing new ones
