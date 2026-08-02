# Remaining Work

> **Complete, verified audit of what is done vs what still needs to be built.**
> Last audited: 2026-08-02 (codebase-level review of `backend/`, `frontend/`, `infra/`).

---

## Status Overview

| Area | Status | Notes |
|---|---|---|
| Backend (7 gRPC services) | **100%** | P0 fully closed; P1 items 6-13 all resolved (template field, dead code removed) |
| Frontend (34 routes) | **99%** | All pages real gRPC-backed; search/ack/rollback-streaming/templates wired |
| Proto / codegen | **100%** | 16 canonical .proto files, buf config, CI/CD proto-check green |
| Docker | **100%** | Compose + Dockerfiles + Envoy + Redis |
| CI/CD | **100%** | 4 workflows, coverage gate (25%), proto-check, deploys, dependabot |
| Testing | **80%** | Unit layer done (28.5% coverage, gate 25%); repository/introspection integration tests pending |
| Documentation | **95%** | 22 .md files |
| Infrastructure | **SKELETON** | Terraform module stubs, monitoring placeholders |
| DB migrations | **2 of N** | `001_init.sql` + `002_project_template.sql` |

---

## What Is DONE (verified)

- **All 7 gRPC services implemented end-to-end**: Auth, Project (incl. connections + email invites), Schema (introspection/diff/diagram), Migration (transactional executor + streaming), Event (Redis pub/sub + replay + presence), Audit (tail/partitioning), Drift (compare + severity).
- **Real connection testing**: `internal/project/domain/connection_service.go` — AES-256-GCM decrypt → pgxpool → `SELECT version()` + latency + status persist.
- **Real migration executor**: `migration/domain/service.go` `executeAsync` — SQL splitter, transaction, per-statement logs, rollback, dry-run; bounded queue (32) + 4-worker pool, no unbounded goroutines.
- **Real OAuth**: Google/GitHub/Slack with PKCE, signed-JWT state, token refresh worker with client secrets wired.
- **Auth**: register/login, JWT rotation (bcrypt + RSA256), email verification + password reset with real SMTP delivery.
- **Interceptors**: recovery, auth (+ stream auth), RBAC (unary + stream), rate-limit (Redis per-user + token bucket), idempotency (Redis), metrics, DB retry, tracing + CORS — all real and wired.
- **Workers**: connection health, audit partition, hard delete, OAuth refresh, drift alert, drift check — 6 real workers started in `cmd/server/main.go`.
- **RBAC**: server-side enforcement for ~37 gRPC methods across projects/connections/schemas/migrations/drifts/audits, unary + streaming.
- **Email**: SMTP mailer (dev-mode logs instead of sending), verification + password-reset emails, `/verify-email` page.
- **Config**: fail-fast validation of required env vars.
- **Frontend**: 34 routes, all 7 service clients wired (Connect + gRPC-Web binary), React Flow ERD, migration streaming UI, realtime event timeline, OAuth pages, team/members, settings, CSV export.
- **CI/CD**: proto-check, coverage gate `>= 25%`, Docker builds, deploys — all green.

---

## What Is LEFT

Prioritized. P0 = do before anything else (security/correctness), P1 = functional gaps, P2 = cleanup, P3 = polish, P4 = testing, P5 = infra/ops.

### P0 — Security & Correctness (HIGH PRIORITY)

| # | Item | Location | Detail |
|---|---|---|---|
| 1 | ~~RBAC is disabled~~ **DONE** | `backend/cmd/server/rbac.go` | `rbacEnforcer` wired in `main.go` with ~37 method scopes, project/connection/schema/migration/drift/audit resolvers, stream + unary RBAC interceptors. Tests in `internal/pkg/interceptor/`. |
| 2 | ~~No auth on streaming RPCs~~ **DONE** | `backend/internal/pkg/interceptor/auth.go` | `StreamAuthInterceptor` authenticates streams and swaps the context; `StreamRBACInterceptor` checks the first `RecvMsg` request. Wired for WatchMigration, Subscribe, TailAuditEntries, etc. |
| 3 | ~~Email delivery does not exist~~ **DONE** | `backend/internal/pkg/mailer/` | SMTP mailer with fake-server tests; `EmailSender` wired into auth service for verification + password-reset emails; new `/verify-email` frontend page. |
| 4 | ~~Redis per-user rate limit unwired~~ **DONE** | `backend/internal/pkg/interceptor/rate_limit.go` | Wired with `cfg.RateLimit`; also fixed a bug where public endpoints (login/register) were skipped entirely. |
| 5 | ~~OAuth refresh goes out without client_secret~~ **DONE** | `backend/cmd/server/main.go` | `SetClientSecret` called for google/github/slack at startup; `GOOGLE_CLIENT_SECRET` added to config + OAuth exchange. |

### P1 — Functional Gaps (MEDIUM PRIORITY)

| # | Item | Location | Detail |
|---|---|---|---|
| 6 | ~~`AcknowledgeEvent` is a no-op~~ **DONE** | `backend/internal/event/` | Redis set `schema:events:acked:{userID}` (TTL 24h); `Acknowledge`/`IsAcknowledged` in event domain service; real handler + miniredis tests. |
| 7 | ~~No migration run queue / worker~~ **DONE** | `backend/internal/migration/domain/service.go` | Bounded queue (32) + 4-worker pool; `RESOURCE_EXHAUSTED` when full; failed-enqueue marks run failed. DB-backed persistence + restart recovery still future work (documented). |
| 8 | ~~No drift scheduler~~ **DONE** | `backend/internal/pkg/worker/drift_check.go` | `DriftCheckWorker` runs every 10 min per active connection (10-min interval; alerts already covered by DriftAlertWorker). **Also fixed 2 production bugs in `schemaDriftComparator`**: introspection ran with empty connID, and `CompareVersions` received a connectionID instead of the baseline version ID — manual drift checks always failed before. |
| 9 | ~~Event repository missing~~ **DONE** | `backend/internal/event/` | Deliberate reuse: events are persisted via the `AuditLogger` interface (`event/domain/service.go` `Publish` + `replayEvents`). No dedicated event repo — documented decision, not an omission. |
| 10 | ~~Config has no validation~~ **DONE** | `backend/internal/pkg/config/config.go` | `Validate()` fails fast on missing `DATABASE_URL`/`REDIS_URL`/JWT keys and short `ENCRYPTION_MASTER_KEY`; tests added. |
| 11 | ~~Introspection swallows errors~~ **DONE** | `backend/internal/schema/domain/introspection.go` | Enum/extension query errors now propagate with context instead of `err == nil` guards. |
| 12 | ~~Drift check creates a schema version as a side effect~~ **DONE** | `backend/cmd/server/main.go` | Comparator passes the real connID and compares against the tracked baseline `CurrentVersionID`; introspection records a new version on change as drift history — documented baseline-pinning as future work. |
| 13 | ~~Firebase config fields declared, no verification code~~ **DONE** | `backend/internal/pkg/config/config.go` | Dead fields (`FirebaseProjectID`/`FirebasePrivateKey`/`FirebaseClientEmail`) removed from struct, `Load()`, and `.env.example`. Google OAuth uses JWT exchange instead of ID-token verification. |

### P2 — Dead Code & Cleanup (LOW PRIORITY)

| # | Item | Location | Detail |
|---|---|---|---|
| 14 | ~~`migration/domain/executor.go` is dead~~ **DONE** | file deleted | `NewExecutor` was never instantiated and duplicated `executeAsync` logic; deleted (service inline logic is the single source of truth, tests green). |
| 15 | ~~`jwt.GenerateRefreshToken` placeholder~~ **DONE** | `backend/internal/pkg/jwt/manager.go` | Now uses `crypto/rand` (48 bytes). |
| 16 | ~~CORS + Tracing middleware not wired~~ **DONE** | `backend/cmd/server/main.go` | `TracingInterceptor` + `CORSInterceptor` wired in unary chain; `CORSStreamInterceptor` added + wired in stream chain (allowed origin = `FRONTEND_URL`). |
| 17 | ~~`Unimplemented<Service>Server` embeds~~ **DONE** | every handler | Audited 2026-08-03: script compared generated `*_service_grpc.pb.go` method lists against all 7 handler structs — every proto method has a real implementation (auth 12, project 19, schema 6, migration 11, event 3, audit 5, drift 5). Embeds are standard forward-compat only. |
| 18 | ~~Empty frontend component dirs~~ **DONE** | `frontend/src/components/{erd,migrations,audit,drift,events,connections,schemas,projects,settings,dashboard,shared}/` | All 11 empty directories deleted; `ui/` + `blocks/` + root components remain. |

### P3 — Frontend Polish (LOW PRIORITY)

| # | Item | Location | Detail |
|---|---|---|---|
| 19 | ~~Search inputs are decorative~~ **DONE** | header of every `src/app/projects/[id]/**` page | Client-side filtering wired: projects list (header + body), schemas (project overview), schema objects, drift reports, audit log (export respects filter), connections, events (type + text), members, migration runs, execution logs. Removed from detail/form pages with no list (erd, compare, drift detail, settings, new-project, new-connection, new-migration). |
| 20 | ~~Dashboard quick actions decorative~~ **DONE** | `src/app/dashboard/page.tsx` | All 4 actions now link to real pages: New Project → `/projects/new`, New Schema → first project's schemas, Run Migration → first project's migrations, Invite Team → `/settings`. |
| 21 | ~~Project templates + "Neon — Connected" badge static~~ **DONE** | `src/app/projects/new/page.tsx`, proto | `template` field added to `CreateProjectRequest`/`Project` (proto #4/#10, regenerated Go+TS), migration `002_project_template.sql`, domain validation (`ValidTemplate`), repo + handler wired; page passes selection to `createProject`. |
| 22 | ~~Notification preferences saved only locally~~ **DONE** | `src/app/settings/page.tsx` | Preferences persisted to `localStorage` (`schemahub.notification_prefs`), loaded on mount, saved on toggle; card copy labels them as browser-local. |
| 23 | ~~OAuth "Remember me" not persisted~~ **DONE** | `src/app/(marketing)/login/page.tsx`, `src/lib/api/auth-store.ts` | Checkbox wired to `authStore.setTokens(access, refresh, remember)` — localStorage when checked, sessionStorage when not; getters fall back across both; logout clears both. |
| 24 | ~~Contact form posts nowhere~~ **DONE** | `src/components/blocks/contact-form.tsx` | Rewired as a real mailto link (prefilled subject/body) to `hello@schemahub.dev`; dead `src/actions/server-action.ts` + `safe-action.ts` deleted, `next-safe-action` dep removed. |
| 25 | ~~`WatchRollback` streaming UI unused~~ **DONE** | `src/lib/api/hooks/use-migrations.ts`, run page | New `useWatchRollback` hook (same message shape as WatchMigration); run page shows live rollback progress (state, statements %, current statement) and terminal rollback outcome. |
| 26 | ~~`AcknowledgeEvent`/`Heartbeat` unused in UI~~ **DONE** | `src/app/projects/[id]/events/page.tsx` | Per-event "Ack" button (Redis-backed, `acked` state), presence heartbeat (`useHeartbeat`, 30s interval) with Active/Idle badge in the events page header. |
| 27 | ~~Unused API hooks~~ **DONE** | `useDriftStats`, `useValidateMigration`, `useDryRunMigration`, `useUpdateMigration`, `useAuditStats` | All 5 deleted (verified zero consumers). `useListLinkedIdentities` kept — consumed by `settings/connections/page.tsx`. |
| 28 | ~~Dashboard stat deltas hardcoded~~ **DONE** | `src/app/dashboard/page.tsx` | Deltas now computed from real data: projects created this month, audit entries in last 24h; stat card detail lines updated; header search filters the projects table. |

### P4 — Testing (HIGH PRIORITY — next focus)

| # | Item | Detail |
|---|---|---|
| 29 | **Repository integration tests** | Real Postgres via `docker-compose.dev.yml` — all `repository/postgres/*` packages have 0 tests today. |
| 30 | **Migration executor integration tests** | `executeAsync` path (transaction, rollback, logs) needs a real target DB. |
| 31 | **Introspection integration tests** | `information_schema` queries against a seeded Postgres. |
| 32 | **auth/domain and event/domain are at 0% coverage** | No test files at all. Add: OAuth flow (PKCE, exchange, link/unlink), email-token flows, Redis pub/sub service. |
| 33 | **Raise the coverage gate** | From 25% to ~50% after the above (CI `.github/workflows/ci.yml` awk threshold). |
| 34 | Low-coverage handlers | migration/handler 17.3%, event/handler 20%, schema/domain 27.2%, project/handler 30.2%, auth/handler 36.5%, audit/handler 40%, drift/handler 46.2%. |
| 35 | Interceptor stream-auth tests | Once P0-2 is implemented, test unauthorized stream access. |

### P5 — Infrastructure & Ops (POST-MVP)

| # | Item | Location | Detail |
|---|---|---|---|
| 36 | **Terraform modules are stubs** | `infra/terraform/modules/*` | "TODO" placeholders. Implement real resources: VPC, RDS (Postgres), ElastiCache (Redis), ECS/Fly app, networking. |
| 37 | Backup/DR procedures | not documented anywhere | DB snapshots, restore runbooks, recovery-time targets. |
| 38 | Grafana dashboard is a placeholder | `infra/monitoring/grafana-dashboards/schemahub.json` | Single overview panel; build real panels per service + latency/error-rate + DB pool + Redis. |
| 39 | Monitoring wiring | `infra/monitoring/` | Prometheus scrape config assumes backend metrics export — verify `/metrics` endpoint and wiring. |
| 40 | **OAuth provider credentials** | env config | Google/GitHub/Slack OAuth apps + secrets must be created and injected per environment. |
| 41 | **SMTP provider credentials** | env config | Required for P0-3. |
| 42 | Neon-compatible deployment | deployment | Verify backend runs on Neon serverless Postgres (pooled mode, no local superuser assumptions). |
| 43 | Envoy gRPC-Web bridge | `docker/envoy/envoy.yaml` | Required for every frontend deployment (backend has no gRPC-Web wrapper); confirm prod topology. |

---

## Verification Checklist (run before declaring MVP)

```bash
# Backend
cd backend
go build ./... && go vet ./... && gofmt -l .           # zero findings
go test ./internal/... ./pkg/... -count=1 -coverprofile=coverage.out
go tool cover -func coverage.out | tail -1             # >= 25% (raise later)

# Frontend
cd frontend
npm run lint && npx tsc --noEmit && npm run build

# Full stack
docker compose -f docker/docker-compose.dev.yml up
# → register → verify email link arrives → create project → add connection →
#   test connection (real ping) → introspect schema → run migration → see drift
```

Manual smoke test (end-to-end proof of P0-3):
1. Register → **verification email actually arrives** in inbox.
2. Forgot password → **reset email arrives**.
3. Two users on one project — member with `viewer` role **cannot delete** (RBAC enforced, not just UI).
4. Open `WatchMigration` stream in a second tab while unauthenticated → **rejected**.
5. OAuth refresh happens with client_secret → no 401s on token refresh.
