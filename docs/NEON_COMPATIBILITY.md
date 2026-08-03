# Neon Compatibility Audit

> Audit of the SchemaHub backend + migrations against Neon serverless
> PostgreSQL (branching, pooled mode, extension allow-list, no-superuser
> model). Conducted by grepping the codebase and migrations; no production
> behavior was changed as a result of this document.

Neon is the documented primary database for SchemaHub
(`docs/TECH_STACK.md` § PostgreSQL (Neon), `docs/PROJECT_CONTEXT.md`
Constraint #3). RDS (AWS) is the alternate control-plane deployment
(`infra/terraform/modules/rds`) — this audit targets Neon specifically.

## 1. Summary

| # | Finding | Verdict |
|---|---|---|
| 1 | `CREATE EXTENSION IF NOT EXISTS pgcrypto` in migration 001 | ✅ Compatible |
| 2 | `gen_random_uuid()` defaults and inserts | ✅ Compatible (PG13+ built-in) |
| 3 | Superuser usage (roles, `pg_read_server_files`, `ALTER SYSTEM`) | ✅ None found |
| 4 | Advisory locks (`pg_advisory_lock*`) | ✅ None found |
| 5 | Postgres `LISTEN`/`NOTIFY` | ✅ Not used (Redis pub/sub instead) |
| 6 | `dblink` / `lo` / FDW extensions | ✅ Not used |
| 7 | Named prepared statements vs. Neon pooler (pgx v5 default) | ⚠️ Needs change when using the pooled endpoint |
| 8 | `statement_timeout` / long-running queries | ⚠️ Needs change (tuning) |
| 9 | Connection pool sizing vs. Neon pooled mode | ⚠️ Needs change (tuning) |
| 10 | TCP keepalive | ✅ Compatible (pgx default, not disabled) |
| 11 | `CREATE EXTENSION` in migrations beyond allow-list | ℹ️ Info — keep to Neon allow-list |
| 12 | Partitioned `audit_logs` + partition worker | ✅ Compatible |
| 13 | `DO $$ ... $$` plpgsql block in migration 001 | ✅ Compatible |
| 14 | `INET`, `JSONB`, `TIMESTAMPTZ` types | ✅ Compatible |
| 15 | `sslmode` handling | ✅ Compatible (Neon requires `sslmode=require`) |
| 16 | Migration validator blocking server-level SQL | ✅ Proactively Neon-safe |

## 2. Detailed findings

### ✅ 1. `pgcrypto` extension — compatible

`backend/internal/pkg/database/migrations/001_init.sql:1`

```sql
CREATE EXTENSION IF NOT EXISTS pgcrypto;
```

Neon supports `pgcrypto` (in the extension allow-list). The `IF NOT EXISTS`
guard makes the migration idempotent. Note that nothing in the migrations
actually *needs* pgcrypto — `gen_random_uuid()` is built into PostgreSQL
13+ — so the extension is harmless but also redundant. Keeping it is fine.

### ✅ 2. `gen_random_uuid()` — compatible

Used as a column default throughout `001_init.sql` (e.g. `users.id`,
`refresh_tokens.id`, `audit_logs.id`) and directly in inserts:
`backend/internal/auth/repository/postgres/user_repo.go:24`,
`backend/internal/auth/repository/postgres/refresh_token_repo.go:30`,
`backend/internal/auth/repository/postgres/oauth_repo.go:25`.

Built into PostgreSQL since 13 — no extension or superuser needed. Neon runs
Postgres 15/16/17. Compatible.

### ✅ 3. Superuser usage — none found

Grep for `superuser`, `ALTER SYSTEM`, `pg_read_server_files`,
`pg_write_server_files`, `CREATE ROLE`, `CREATE USER`, `GRANT` in Go and SQL:
the only hits are a domain validator unit test
(`backend/internal/project/domain/project_test.go:122`, `ValidateRole` —
an RBAC role string, unrelated to Postgres) and the migration validator's
deny-list (finding 16). No code assumes superuser privileges.

### ✅ 4. Advisory locks — none found

Grep for `pg_advisory` across Go and SQL: zero hits. No migration or worker
relies on Postgres advisory locking. (If one is ever introduced, Neon
supports advisory locks per-connection — but never assume they survive
pooled connection rotation.)

### ✅ 5. `LISTEN`/`NOTIFY` — not used

Real-time events are delivered via **Redis pub/sub**
(`backend/internal/event/domain/service.go` uses the Redis client, wired in
`backend/cmd/server/main.go:180`), not Postgres `LISTEN/NOTIFY`. Neon
supports `LISTEN/NOTIFY` on direct connections only (it is not usable
through the pooled endpoint), so avoiding it sidesteps the pitfall entirely.

### ✅ 6. `dblink` / `lo` / FDW — not used

No `dblink`, `lo_import`, `lo_export`, `postgres_fdw`, or
`CREATE FOREIGN` anywhere in Go or SQL. Neon does not allow `dblink`/`lo`.
Compatible by absence.

### ⚠️ 7. Named prepared statements vs. Neon pooler — needs change (when pooled)

pgx v5 defaults to **cached named prepared statements**
(`QueryExecModeCacheStatement`). PgBouncer-style transaction pooling (the
Neon **pooled** endpoint) does not support named prepared statements — the
pool can rotate the server connection between statements and the cached
`PREPARE` is lost ("prepared statement does not exist" / "server is being
terminated" class errors).

`backend/internal/pkg/database/postgres.go:11-33` builds the pool with
`pgxpool.ParseConfig` and does not touch the query exec mode.

**Fix (when using the Neon pooled endpoint)** — either:
1. **Prefer the Neon *direct* endpoint** (non-pooled host) for the backend
   process pool — recommended for a server workload with a bounded pool, or
2. Force unnamed statements in `database.Connect`:
   ```go
   cfg.ConnConfig.DefaultQueryExecMode = pgx.QueryExecModeSimpleProtocol
   ```
   (or `QueryExecModeExec`/`QueryExecModeCacheStatement` off) — accepts the
   per-query parse cost.

Not applied: production behavior intentionally unchanged by this audit.

### ⚠️ 8. `statement_timeout` — needs change (tuning)

Nothing in the codebase or migrations sets `statement_timeout`. On Neon the
default is 0 (unlimited) for direct connections, but pooled connections and
long-running DDL/DDL-on-branch operations are subject to idle-timeout /
statement safeguards (Neon terminates long queries per plan limits).

**Recommendation** (deployment-time, not code): set
`statement_timeout` per session for introspection and migration execution,
or keep migrations small and run them against a branch first (see
`docs/DATABASE_DESIGN.md` § workspace isolation). This is a tuning item,
not a blocker.

### ⚠️ 9. Pool sizing vs. Neon pooled mode — needs change (tuning)

`backend/internal/pkg/database/postgres.go`:
- `MaxConns = maxConn` (default 20 via `DB_POOL_MAX`, config.go:77)
- `MaxConnLifetime = 30 * time.Minute`, `MaxConnIdleTime = 5 * time.Minute`,
  `HealthCheckPeriod = 30 * time.Second`

Against the Neon **pooled** endpoint, 20 client connections collapse onto a
small number of server connections and can saturate the compute's connection
limit. The lifetime/idle bounds are good hygiene for pooled mode (they force
connection recycling).

**Recommendation**: for the Neon pooled endpoint set `DB_POOL_MAX` to
10–15; for the direct endpoint 20 is fine. The RDS/ECS deployment
(`infra/terraform/modules/ecs`, `db_pool_max` variable) defaults to 20 —
reduce when Neon-backed.

### ✅ 10. TCP keepalive — compatible

pgx v5.10 does not tune keepalive itself — its default dialer
(`pgconn/config.go`, `makeDefaultDialer`) "relies on GOLANG KeepAlive
settings", and Go enables TCP keepalive by default (≈15 s). Neither
`backend/internal/pkg/database/postgres.go` nor any repo overrides this.
Neon requires keepalive-enabled clients to avoid stale connections after
compute idle-down. Compatible; keep the Go default.

### ℹ️ 11. Extension allow-list — info

Only `pgcrypto` is created by migrations (`001_init.sql:1`). Neon allows
only its documented extension list (pgcrypto is on it). Any future
migration introducing an extension must first confirm it is in Neon's
allow-list — `uuid-ossp`, `pg_trgm`, `hstore`, `pg_stat_statements` are
allowed; `dblink`, `lo`, `postgres_fdw`, `pg_cron` are not.

### ✅ 12. Partitioned `audit_logs` — compatible

`001_init.sql:181-205` declares `audit_logs` partitioned by RANGE on
`created_at` with concrete partitions. Neon supports declarative
partitioning on the main branch.

`backend/internal/pkg/worker/audit_partition.go:39-44` creates future
partitions with `CREATE TABLE ... PARTITION OF audit_logs`. This requires
ownership of `audit_logs` — the Neon database owner role has it. Compatible.

### ✅ 13. `DO $$ ... $$` plpgsql block — compatible

`001_init.sql:119-126` guards the `current_version_id` FK with a plpgsql
`DO` block. Neon runs standard plpgsql. Compatible.

### ✅ 14. Types — compatible

`INET` (`refresh_tokens.created_by_ip`, `audit_logs.ip_address`), `JSONB`
(`schema_versions.metadata`, `schema_objects.definition`,
`audit_logs.resource_changes`), `TIMESTAMPTZ`, `UUID` — all supported on
Neon.

### ✅ 15. `sslmode` — compatible

Neon requires TLS (`sslmode=require`); `docs/RUN_LOCALLY.md:79` already
documents this for the Neon `DATABASE_URL`. The `connections` table stores
per-connection `ssl_mode` with default `'require'`
(`001_init.sql:84`) — aligned. The compose stack uses `sslmode=disable`
locally, which is correct for the sandbox network and irrelevant to Neon.

### ✅ 16. Migration validator — proactively Neon-safe

`backend/internal/migration/domain/validator.go:12` rejects `ALTER SYSTEM`;
line 113 rejects `GRANT`, `REVOKE`, `SET`, `ANALYZE`, `VACUUM`, `REINDEX`
in user migrations. These are exactly the constructs Neon does not allow
(`ALTER SYSTEM` is unsupported; `SET`/`GRANT` are restricted per role).
This validator keeps *user-supplied* migrations Neon-safe by construction.

## 3. Required changes before a Neon production rollout

| Change | Where | Who |
|---|---|---|
| Use the Neon **direct** endpoint for the backend pool (or force simple-protocol query mode — finding 7) | `DATABASE_URL` / `backend/internal/pkg/database/postgres.go` | Platform |
| Keep `DB_POOL_MAX` ≤ 15 when Neon-backed (finding 9) | env / `infra/terraform` vars | Platform |
| Confirm the Neon plan's PITR retention ≥ 7 days (matches `docs/BACKUP_DR.md` targets) | Neon console | Platform |
| Verify branch-based migration workflow (run each migration on a branch first) | CI/CD | Platform |

Nothing in the application code requires an extension beyond pgcrypto or any
superuser privilege, so no functional code change is required for Neon.

## 4. Non-blocking recommendations

- Drop `pgcrypto` from migration 001 once no consumer needs it (it is
  currently unused — `gen_random_uuid()` is built-in) — reduces the
  allow-list surface.
- Set `statement_timeout` for introspection and migration execution
  (finding 8).
- Keep `MaxConnLifetime`/`MaxConnIdleTime` as-is — they are the correct
  behaviour behind a pooler.

## 5. References

- `docs/BACKUP_DR.md` — Neon PITR/branching as DR strategy
- `docs/TECH_STACK.md` § PostgreSQL (Neon) — rationale and trade-offs
- `docs/DEPLOYMENT.md` § Backup Strategy — RTO/RPO table
- `docs/RUN_LOCALLY.md` — local Neon `DATABASE_URL` setup
- `backend/internal/pkg/database/postgres.go` — pool configuration audited
  in findings 7–10
- `backend/internal/pkg/database/migrations/001_init.sql`,
  `002_project_template.sql` — migrations audited in findings 1, 2, 12–14
