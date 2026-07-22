# Folder Structure

> **Complete repository layout for SchemaHub with explanations of every directory and file.**

---

## Table of Contents

- [Root Level](#root-level)
- [Frontend](#frontend)
- [Backend](#backend)
- [Protobuf Definitions](#protobuf-definitions)
- [Documentation](#documentation)
- [Docker](#docker)
- [Scripts](#scripts)
- [Infrastructure](#infrastructure)

---

## Root Level

```
schemahub/
├── frontend/                   # Next.js frontend application
├── backend/                    # Go backend services
├── proto/                      # Source protobuf definitions
├── docs/                       # Project documentation
├── docker/                     # Docker and Compose files
├── scripts/                    # Development and CI scripts
├── infra/                      # Infrastructure as Code
├── .github/                    # GitHub Actions, templates
├── .gitignore
├── .editorconfig
├── CLAUDE.md                   # AI assistant context
├── README.md                   # Project overview
├── LICENSE                     # MIT license
└── Makefile                    # Top-level build orchestration
```

---

## Frontend

```
frontend/
├── .next/                       # Next.js build output (gitignored)
├── public/                      # Static assets
│   ├── favicon.ico
│   ├── logo.svg
│   └── diagrams/               # Static diagram assets
├── src/
│   ├── app/                     # Next.js App Router pages and layouts
│   │   ├── (auth)/             # Authentication route group
│   │   │   ├── login/
│   │   │   │   └── page.tsx
│   │   │   ├── register/
│   │   │   │   └── page.tsx
│   │   │   └── layout.tsx
│   │   ├── (dashboard)/        # Dashboard route group
│   │   │   ├── projects/
│   │   │   │   ├── [slug]/
│   │   │   │   │   ├── schemas/
│   │   │   │   │   ├── migrations/
│   │   │   │   │   ├── settings/
│   │   │   │   │   └── page.tsx
│   │   │   │   └── page.tsx
│   │   │   └── layout.tsx
│   │   ├── layout.tsx          # Root layout
│   │   ├── loading.tsx         # Global loading state
│   │   ├── error.tsx           # Global error boundary
│   │   └── not-found.tsx       # 404 page
│   ├── components/             # Reusable UI components
│   │   ├── ui/                 # shadcn/ui primitives
│   │   │   ├── button.tsx
│   │   │   ├── dialog.tsx
│   │   │   ├── input.tsx
│   │   │   ├── select.tsx
│   │   │   ├── table.tsx
│   │   │   ├── tabs.tsx
│   │   │   ├── toast.tsx
│   │   │   └── ...
│   │   ├── layout/             # Layout components
│   │   │   ├── sidebar.tsx
│   │   │   ├── navbar.tsx
│   │   │   └── project-nav.tsx
│   │   ├── schema/             # Schema-related components
│   │   │   ├── schema-tree.tsx
│   │   │   ├── schema-detail.tsx
│   │   │   ├── column-list.tsx
│   │   │   ├── schema-diagram.tsx    # React Flow wrapper
│   │   │   └── version-timeline.tsx
│   │   ├── migration/          # Migration-related components
│   │   │   ├── migration-form.tsx
│   │   │   ├── migration-list.tsx
│   │   │   ├── migration-runner.tsx
│   │   │   ├── migration-status.tsx
│   │   │   └── rollback-button.tsx
│   │   ├── diff/               # Diff viewer components
│   │   │   ├── diff-viewer.tsx
│   │   │   └── diff-line.tsx
│   │   ├── audit/              # Audit log components
│   │   │   ├── audit-log.tsx
│   │   │   └── audit-entry.tsx
│   │   └── shared/             # Shared components
│   │       ├── loading-spinner.tsx
│   │       ├── empty-state.tsx
│   │       ├── error-state.tsx
│   │       ├── confirm-dialog.tsx
│   │       ├── pagination.tsx
│   │       └── connection-status.tsx
│   ├── lib/                    # Utilities and client code
│   │   ├── api/                # gRPC API client setup
│   │   │   ├── client.ts       # gRPC-Web client initialization
│   │   │   ├── auth.ts         # Auth-related API calls
│   │   │   ├── project.ts      # Project API calls
│   │   │   ├── schema.ts       # Schema API calls
│   │   │   ├── migration.ts    # Migration API calls
│   │   │   └── audit.ts        # Audit API calls
│   │   ├── hooks/              # Custom React hooks
│   │   │   ├── use-auth.ts
│   │   │   ├── use-project.ts
│   │   │   ├── use-schema.ts
│   │   │   ├── use-migration.ts
│   │   │   ├── use-realtime.ts # WebSocket/streaming hook
│   │   │   └── use-debounce.ts
│   │   ├── providers/          # React context providers
│   │   │   ├── auth-provider.tsx
│   │   │   ├── query-provider.tsx  # TanStack Query provider
│   │   │   └── theme-provider.tsx
│   │   ├── utils/              # Utility functions
│   │   │   ├── cn.ts           # clsx + tailwind-merge helper
│   │   │   ├── format.ts       # Date, number formatting
│   │   │   └── validation.ts   # Form validation rules
│   │   └── types/              # TypeScript type definitions
│   │       ├── generated/      # Generated protobuf types
│   │       ├── schema.ts
│   │       ├── migration.ts
│   │       └── common.ts
│   └── styles/                 # Global styles
│       ├── globals.css         # Tailwind imports, global styles
│       └── tailwind.config.ts  # Tailwind configuration
├── package.json
├── tsconfig.json               # strict: true
├── next.config.ts
├── tailwind.config.ts
├── postcss.config.js
└── .env.local                  # Local environment variables (gitignored)
```

---

## Backend

```
backend/
├── cmd/                        # Binary entry points
│   ├── server/                 # Main server binary
│   │   └── main.go             # Dependency injection, server startup
│   ├── migrate/                # Database migration tool
│   │   └── main.go
│   └── seed/                   # Development data seeder
│       └── main.go
├── internal/                   # Private application packages
│   ├── auth/                   # Authentication domain
│   │   ├── domain/             # Business logic, entities
│   │   │   ├── user.go
│   │   │   ├── service.go
│   │   │   └── errors.go
│   │   ├── repository/         # Data access layer
│   │   │   ├── interface.go    # Repository interfaces
│   │   │   └── postgres/       # PostgreSQL implementation
│   │   │       ├── user_repo.go
│   │   │       └── token_repo.go
│   │   └── handler/            # gRPC handlers
│   │       └── grpc.go
│   ├── project/                # Project domain
│   │   ├── domain/
│   │   ├── repository/
│   │   │   ├── interface.go
│   │   │   └── postgres/
│   │   └── handler/
│   ├── schema/                 # Schema domain
│   │   ├── domain/
│   │   │   ├── schema.go
│   │   │   ├── introspection.go
│   │   │   ├── differ.go       # Schema diff engine
│   │   │   └── service.go
│   │   ├── repository/
│   │   │   ├── interface.go
│   │   │   └── postgres/
│   │   └── handler/
│   ├── migration/              # Migration domain
│   │   ├── domain/
│   │   │   ├── migration.go
│   │   │   ├── executor.go     # SQL execution engine
│   │   │   ├── validator.go    # SQL validation
│   │   │   └── service.go
│   │   ├── repository/
│   │   │   ├── interface.go
│   │   │   └── postgres/
│   │   └── handler/
│   ├── event/                  # Event streaming domain
│   │   ├── domain/
│   │   │   ├── event.go
│   │   │   ├── subscriber.go
│   │   │   └── service.go
│   │   └── handler/
│   ├── audit/                  # Audit domain
│   │   ├── domain/
│   │   │   ├── audit.go
│   │   │   └── service.go
│   │   ├── repository/
│   │   │   ├── interface.go
│   │   │   └── postgres/
│   │   └── handler/
│   ├── drift/                  # Drift detection domain
│   │   ├── domain/
│   │   │   ├── drift.go
│   │   │   └── service.go
│   │   └── repository/
│   │       ├── interface.go
│   │       └── postgres/
│   └── pkg/                    # Internal shared packages
│       ├── config/             # Configuration loading
│       │   └── config.go
│       ├── database/           # Database connection, migration runner
│       │   ├── postgres.go
│       │   └── migrations/     # DB migration SQL files
│       ├── redis/              # Redis client, pub/sub helpers
│       │   └── redis.go
│       ├── jwt/                # JWT creation and verification
│       │   ├── manager.go
│       │   └── keys.go
│       ├── interceptor/        # Shared gRPC interceptors
│       │   ├── auth.go
│       │   ├── logging.go
│       │   ├── recovery.go
│       │   ├── rate_limit.go
│       │   └── validation.go
│       ├── logger/             # Structured logging setup
│       │   └── logger.go
│       ├── middleware/          # HTTP middleware (for gateway)
│       │   ├── cors.go
│       │   └── tracing.go
│       ├── errors/             # Error types and gRPC mapping
│       │   └── errors.go
│       └── testutil/           # Test helpers
│           ├── db.go
│           └── mock.go
├── pkg/                        # Public shared packages
│   └── encryption/             # Password/credential encryption
│       └── crypto.go
├── proto/                      # Generated protobuf Go code
│   ├── auth/
│   │   └── v1/
│   ├── project/
│   │   └── v1/
│   ├── schema/
│   │   └── v1/
│   ├── migration/
│   │   └── v1/
│   ├── event/
│   │   └── v1/
│   ├── audit/
│   │   └── v1/
│   └── common/
│       └── v1/
├── go.mod
├── go.sum
├── Makefile                    # Go-specific build commands
├── .golangci.yml               # Linter configuration
└── tools.go                    # Tool dependency tracking
```

---

## Protobuf Definitions

```
proto/
├── auth/
│   └── v1/
│       ├── auth_service.proto
│       └── auth_messages.proto
├── project/
│   └── v1/
│       ├── project_service.proto
│       └── project_messages.proto
├── schema/
│   └── v1/
│       ├── schema_service.proto
│       └── schema_messages.proto
├── migration/
│   └── v1/
│       ├── migration_service.proto
│       └── migration_messages.proto
├── event/
│   └── v1/
│       ├── event_service.proto
│       └── event_messages.proto
├── audit/
│   └── v1/
│       ├── audit_service.proto
│       └── audit_messages.proto
├── common/
│   └── v1/
│       ├── common.proto
│       └── pagination.proto
├── buf.yaml                   # Buf configuration
├── buf.gen.yaml               # Buf code generation config
└── Makefile                   # Proto generation commands
```

---

## Documentation

```
docs/
├── PROJECT_CONTEXT.md         # Single source of truth
├── ARCHITECTURE.md            # System architecture
├── TECH_STACK.md              # Technology decisions
├── DATABASE_DESIGN.md         # Database schema design
├── ER_DIAGRAM.md              # Entity-relationship diagram
├── GRPC_DESIGN.md             # gRPC architecture
├── PROTOBUF_CONTRACTS.md      # Service contracts
├── API_FLOW.md                # Request flows
├── AUTHENTICATION.md          # Auth and authorization
├── REALTIME_ARCHITECTURE.md   # Real-time streaming
├── FOLDER_STRUCTURE.md        # This file
├── FEATURE_SPECIFICATIONS.md  # Feature details
├── ROADMAP.md                 # Development roadmap
├── SECURITY.md                # Security architecture
├── DEPLOYMENT.md              # Deployment guide
├── TESTING_STRATEGY.md        # Testing approach
├── CODING_GUIDELINES.md       # Coding conventions
├── FUTURE_IDEAS.md            # Future feature ideas
└── CONTRIBUTING.md            # Contribution guidelines
```

---

## Docker

```
docker/
├── docker-compose.yml          # Full stack (backend, frontend, redis)
├── docker-compose.dev.yml      # Override for development
├── docker-compose.test.yml     # Override for testing
├── backend/
│   ├── Dockerfile              # Multi-stage Go build
│   └── Dockerfile.dev          # Hot-reload development build
├── frontend/
│   ├── Dockerfile              # Next.js production build
│   └── Dockerfile.dev          # Next.js development with HMR
├── envoy/
│   └── envoy.yaml              # Envoy proxy configuration
└── redis/
    └── redis.conf              # Redis configuration
```

---

## Scripts

```
scripts/
├── setup.sh                   # Initial project setup
├── dev.sh                     # Start development environment
├── lint.sh                    # Run all linters
├── test.sh                    # Run all tests
├── proto-gen.sh               # Generate protobuf code
├── migrate.sh                 # Run database migrations
├── seed.sh                    # Seed development data
└── pre-commit.sh              # Pre-commit hook
```

---

## Infrastructure

```
infra/
├── terraform/                  # Terraform configurations
│   ├── environments/
│   │   ├── dev/
│   │   ├── staging/
│   │   └── production/
│   ├── modules/
│   │   ├── database/
│   │   ├── redis/
│   │   ├── backend-service/
│   │   └── networking/
│   └── main.tf
├── pulumi/                     # Alternative: Pulumi configurations
│   └── index.ts
└── monitoring/
    ├── prometheus.yml
    └── grafana-dashboards/
        └── schemahub.json
```
