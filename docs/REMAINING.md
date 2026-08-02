# Remaining Work

> **Complete audit of what is done vs what needs to be built.**

---

## Status Overview

| Area | Completion | Lines of Code |
|---|---|---|
| **Backend** | 100% | 87 Go files, build+vet pass, gofmt clean |
| **Proto** | 100% | 16 canonical .proto files + buf config |
| **Docker** | 100% | compose + Dockerfiles + envoy + redis |
| **CI/CD** | 100% | 4 GitHub Actions workflows + dependabot + coverage gate |
| **Scripts** | 100% | 7 PowerShell scripts |
| **Documentation** | 95% | 21 .md files (incl. RUN_LOCALLY.md) |
| **Infrastructure** | Skeleton | infra/ dir, Terraform modules, prometheus + grafana placeholders |
| **Frontend** | 100% | 34 routes, 21/21 product pages (see FRONTEND_PLAN.md) |
| **Testing** | 75% | Unit (domain + handlers) complete; integration vs real Postgres pending |
| **Tooling** | 100% | Makefiles x3, .golangci.yml, tools.go |

---

## Backend — DONE

### All 7 Services Fully Implemented

| Service | Domain | Repository (Postgres) | Handler | Notes |
|---|---|---|---|---|
| Auth | user.go, service.go, oauth.go, repositories.go | user_repo, oauth_repo, refresh_token_repo, verification_token_repo | grpc.go, errors.go | Register/Login/Logout/RefreshToken/UpdateUser/ChangePassword/DeleteAccount/ForgotPassword/ResetPassword/SendVerificationEmail/VerifyEmail + OAuth (Google/GitHub/Slack) + PKCE + account linking. All recovery RPCs exposed via gRPC and wired in frontend (`/forgot-password`, `/forgot-password/reset`) |
| Project | project.go, connection.go, service.go, connection_service.go | project_repo.go, connection_repo.go | grpc.go | Full CRUD + members + roles + slug generation + connection encryption (AES-256-GCM) + async connection test |
| Schema | schema.go, introspection.go, differ.go, cache.go, service.go | schema_repo.go | grpc.go | Introspection (information_schema + pg_catalog), versioning (SHA-256 content-addressed), diff engine, diagram data (nodes/edges), Redis cache |
| Migration | migration.go, service.go, executor.go, validator.go | migration_repo.go | grpc.go | Full CRUD + async execution + transaction management + per-statement logging + streaming progress + rollback + dry-run + SQL validation |
| Event | models.go, service.go | (uses AuditRepository) | grpc.go | Redis pub/sub, subscription manager, per-client buffering, reconnection replay, presence tracking (heartbeat) |
| Audit | models.go, interface.go, service.go | postgres.go | grpc.go | Insert/List/GetByID/ListAfterID/GetStats, streaming tail with polling, partitioned table support |
| Drift | models.go, interface.go, service.go | postgres.go | grpc.go | CheckDrift (re-introspect + compare), List/GetByID/Resolve/GetStats, severity classification |

### All Shared Packages

| Package | Files | Purpose |
|---|---|---|
| config | config.go | Env var loading (DB, Redis, JWT, OAuth, encryption, rate limiting) |
| database | postgres.go, migrate.go | pgxpool connection + SQL migration runner |
| errors | errors.go | Domain error types + gRPC status code mapping |
| interceptor | 11 files | auth, recovery, logging, rate-limit, rbac, idempotency, metrics, db-retry, validation, context helpers |
| jwt | manager.go | RS256 access token + refresh token generation/verification |
| logger | logger.go | Structured slog logger (JSON/text, level config) |
| middleware | cors.go, tracing.go | CORS header injection, trace ID propagation |
| rbac | rbac.go | Role-based project permission checker |
| redis | redis.go | Redis client connection |
| worker | 6 files | Connection health check, audit partition creation, hard delete, OAuth token refresh, drift alert, generic runner |
| encryption (pkg/) | crypto.go | AES-256-GCM encrypt/decrypt |

### Generated Proto Code

| Directory | Generated Files |
|---|---|
| proto/auth/v1/ | auth_service.pb.go, auth_service_grpc.pb.go |
| proto/project/v1/ | project_messages.pb.go, project_service_grpc.pb.go |
| proto/schema/v1/ | schema_messages.pb.go, schema_service_grpc.pb.go |
| proto/migration/v1/ | migration_messages.pb.go, migration_service_grpc.pb.go |
| proto/event/v1/ | event_messages.pb.go, event_service_grpc.pb.go |
| proto/audit/v1/ | audit_messages.pb.go, audit_service_grpc.pb.go |
| proto/drift/v1/ | drift_messages.pb.go, drift_service_grpc.pb.go |
| proto/common/v1/ | common.pb.go |

---

## What's Left To Do

### 1. Backend Polish — DONE

`.golangci.yml`, `backend/tools.go`, and 3 Makefiles (`./Makefile`,
`backend/Makefile`, `proto/Makefile`) all created.

### 2. Testing — In Progress (unit layer incl. handlers done, ~75%)

`go test ./...` passes for all packages; gofmt clean; go vet clean.
CI enforces a 25% coverage gate on `internal/ + pkg/`. Coverage per package:

| Package | Coverage |
|---|---|
| internal/pkg/config | 100% |
| internal/pkg/errors | 100% |
| internal/pkg/rbac | 100% |
| internal/audit/domain | 95.7% |
| internal/audit/handler | 40.0% |
| internal/pkg/interceptor | 62.5% (redis-backed paths need integration) |
| internal/pkg/jwt | 89.6% |
| pkg/encryption | 85.7% |
| internal/drift/domain | 83.3% |
| internal/migration/domain | 57.9% (executor paths need integration) |
| internal/project/domain | ~45% (incl. new AddMember email-invite tests) |
| internal/schema/domain | 27.2% (introspection/service need integration) |

Handler tests added for all 7 services (31 test functions: auth, project,
schema, migration, audit, drift, event).

**Remaining test work:** repository layer (integration tests against real
Postgres via docker-compose), executor + introspection integration tests,
raise the CI coverage gate from 25%.

### 3. Documentation — DONE

`LICENSE` (MIT) and `.editorconfig` added.

### 4. Infrastructure — Skeleton (was 0%)

| Item | Description |
|---|---|
| `infra/` directory | Created: `terraform/` (environments + modules) + `monitoring/` |
| Terraform IaC | `main.tf` + `variables.tf` + 4 module stubs + 3 env tfvars examples |
| Prometheus config | `prometheus.yml` + `alerts.yml` ready to use |
| Grafana dashboards | `schemahub.json` placeholder (Overview, gRPC, DB, Real-time) |
| Monitoring/alerting | Alert rules defined; wiring depends on backend metrics export |
| Backup/DR procedures | Not implemented |

**Infra modules are stubs** — real resources (VPC, RDS, ElastiCache, ECS/Fly)
are TODO inside each module.

### 4. Frontend — DONE (21/21 product pages)

34 routes built, lint+build clean, every page wired to gRPC-Web real data — zero mock data files remain (`src/lib/*-data.ts` deleted). Notifications popover is real-time (Redis pub/sub event stream). Reviews feature removed (no RPCs in the backend contract). Details in `frontend/FRONTEND_PLAN.md`.

---

## .gitignore — Already Exists

A comprehensive `.gitignore` already exists at root. Here's what it covers:

| Pattern | What It Ignores |
|---|---|
| `.env`, `.env.local`, `.env.production`, `.env.*.local` | Environment files with secrets |
| `node_modules/`, `vendor/` | Dependencies |
| `.next/`, `dist/`, `out/` | Build outputs |
| `*.exe`, `*.dll`, `*.so`, `*.dylib` | Compiled binaries |
| `backend/vendor/`, `backend/server`, `backend/tmp/` | Go artifacts + air temp |
| `go.work`, `go.work.sum` | Go workspace files |
| `.idea/`, `.vscode/`, `*.swp`, `*.swo` | IDE files |
| `*.log`, `logs/` | Log files |
| `coverage/`, `coverage.out`, `*.test`, `*.test.exe` | Test artifacts |
| `__pycache__/`, `*.pyc` | Python cache |
| `infra/terraform/.terraform/`, `*tfstate*`, `*.tfplan` | Terraform state |
| `*.pem`, `*.key` | Certificates/keys |
| `tmp/` | Temp directory |

**.gitignore is good — no changes needed.**

---

## Summary

| Area | Status | Priority |
|---|---|---|
| Backend | **100% DONE** | - |
| Proto files | **100% DONE** | - |
| Docker | **100% DONE** | - |
| CI/CD (incl. coverage gate) | **100% DONE** | - |
| Scripts | **100% DONE** | - |
| .gitignore | **EXISTS** | - |
| .env.example | **EXISTS** | - |
| Backend polish (Makefile, linter, tools.go) | **DONE** | - |
| Documentation (LICENSE, .editorconfig, RUN_LOCALLY) | **DONE** | - |
| Infrastructure (infra/, monitoring) | **SKELETON (modules need resources)** | Low (post-MVP) |
| **Frontend** | **100% DONE (34 routes)** | - |
| **Tests** | **~75% (unit incl. handlers; integration pending)** | **High** |
