# SchemaHub — Status & Verification

> Last verified: 06 Aug 2026 — against **live Neon database** (user's own Neon instance) — **ALL GREEN**
> Session notes: fresh DB wipe + full E2E on real Neon; audit trail now writes end-to-end; probe tools removed after use.

---

## How to run the full stack

```bash
docker compose -f docker/docker-compose.dev.yml up -d   # from repo root
```

| Service  | URL |
|---|---|
| Frontend (Next.js) | http://localhost:3000 |
| Backend (gRPC) | localhost:50051 |
| Envoy (gRPC-Web proxy) | localhost:8080 |
| Redis | localhost:6379 |
| Postgres (local, fallback) | localhost:5432 |

> The app database is **Neon** (see `.env` → `DATABASE_URL`), not local Postgres.
> Migrations 001–006 are applied on Neon (`last_introspected_at` dropped; audit `resource_id`/`trace_id` nullable).
> **Backend code changes require `docker compose -f docker/docker-compose.dev.yml restart backend`** — air's auto-reload on Docker Desktop does not reliably fire (observed: binary stayed at old mtime until explicit restart).

---

## Test accounts (registered on Neon — current DB state)

| Role | Email | Password |
|---|---|---|
| Owner | `ritesh.singh@schemahub.dev` | `Ritesh-Singh-123` |
| Member | `aarav.sharma@schemahub.dev` | `Member-Dev-123` |
| Member | `priya.verma@schemahub.dev` | `Member-Dev-123` |
| Member | `vikram.patel@schemahub.dev` | `Member-Dev-123` |
| Member | `sneha.gupta@schemahub.dev` | `Member-Dev-123` |
| Member (admin) | `arjun.mehta@schemahub.dev` | `Member-Dev-123` |
| Pending invite (unregistered) | `future.dev@schemahub.dev` | — (invitation link) |

**Test project (owned by Ritesh Singh):**
- `Production DB` — id `d02994c6-a756-42dd-810e-84c959085e6e`
- Connection: `Neon Production` — id `1c94488d-5405-4e68-9a3a-14a0ae5eb998` (real Neon, `neondb`)
- Introspected schema version `50ba365a-49a2-4aaf-8ebd-742397eb0516` — **23 objects**
- ERD diagram: **21 nodes / 154 edges**
- Migration: `Add profile_bio to users` v1.0.1 — executed then rolled back on the live connection (both runs completed)

> Note: older `e2e.owner` / `e2e.member` accounts were wiped during the fresh E2E run and no longer exist.
> The temporary `cmd/e2eprobe` and `cmd/auditprobe` tools were deleted after verification (per user request). Re-verify via UI or a fresh throwaway probe.

---

## What is verified (green) on Neon

### Auth
- [x] Register with password policy (8+ chars, upper + lower + digit) — rate limit 3/min per email
- [x] Login / refresh token flow
- [x] gRPC auth interceptor (401 on bad token)
- [x] Login rate limit (3/min)

### Projects & team
- [x] Create / list projects
- [x] Invite member by email (unregistered → invitation link; registered → added directly)
- [x] Invite duplicate → AlreadyExists; invalid role → InvalidArgument
- [x] Accept invitation via `/invite/accept?token=...` (7-day expiry; reuse blocked)
- [x] Roles: `owner/admin/member/viewer`; role update verified (Arjun → admin)
- [x] `ListMembers` returns proper array (frontend crash root cause was stale JS, not the API)

### Connections + Schema Introspection (real Neon DB)
- [x] Create connection (password encrypted — AES-256-GCM, SHA-256 derived from 40-char master key)
- [x] `TestConnection` → **PostgreSQL 18.4 on Neon, db=`neondb`** (latency ~2 s)
- [x] `IntrospectSchema` on `public` — 23 objects, versioned (checksum dedupe), `project_id` now populated
- [x] `ListSchemas` shows introspected schema with current version
- [x] `GetSchemaDiagram` — 21 nodes / 154 edges
- [x] Pooler reliability: pgxpool lifetime/health settings + one retry on transient `unexpected EOF` errors

### Migrations (executed on real Neon connection)
- [x] Validate / Create / DryRun / Execute / Logs / Rollback — full cycle PASS on live connection
- [x] Migration run status transitions (running → completed), rollback completed

### Audit trail
- [x] `AuditInterceptor` records every successful mutating RPC (skips AuthService + Get/List/Tail/Watch/Validate/DryRun)
- [x] Audit write verified end-to-end: `CreateProject` → row in `audit_logs`, `ListAuditEntries` + `GetAuditStats` return it
- [x] Fixed: `resource_id`/`trace_id` NOT NULL on partitioned table (dropped via ALTER + `001_init.sql`), NULL-safe `scanEntry` (resource_id/trace_id/resource-changes as pointers)

### Frontend
- [x] Dashboard, projects, schemas, team, connections, settings pages (HTTP 200)
- [x] Hydration mismatch fixed: React `19.2.0 → 19.2.8` (data-has-listeners bug) + musl SWC binary for alpine
- [x] Dead field `last_introspected_at` removed end-to-end (proto → DB column → API → UI shows "Last Updated")
- [x] Defensive fix for `members.map is not a function` — pages now `Array.isArray`-guard member lists (project detail + settings/members)
- [x] `tsc --noEmit` clean, `npm run lint` — No ESLint warnings or errors
- [x] **Performance pass (see `docs/FRONTEND_PERFORMANCE.md`)**: routes moved to `(app)/` route group with single shared shell (was 22 duplicated page shells); sidebar N+1 removed (was one members+connections query per project); notifications stream lazy-mounts on popover open; `useEventStream` exponential backoff + jitter; audit `pageSize` 50 + date/driftType query-key fixes; transport `defaultTimeoutMs 60s`; `@heroui/react` removed (44 pkgs); `privacy` page is now a server component (needed `src/mdx-components.tsx` for Next MDX RSC context fix); dead files removed (`app/globals.css`, `lib/hooks/use-mobile.ts`, template SVGs, scratch scripts/docs). All key routes HTTP 200 re-verified.

### Backend quality
- [x] `go build ./...` clean, `go vet ./...` clean
- [x] `go test ./...` — all 23 packages pass

---

## Known issues

1. **Introspection latency on Neon pooler** — a full introspect takes ~40–60 s (one query per table/column/index/constraint round-trip through the pooler). Fine for on-demand runs; future optimization: run per-table queries concurrently.
2. **Full compose stack must start together** — backend exits if redis is down (observed once after a host restart).
3. **air auto-reload unreliable on Docker Desktop** — after editing Go code, always `docker compose -f docker/docker-compose.dev.yml restart backend`.
4. **Audit entries currently log `CreateProject` with empty `resource_id`** — the interceptor prefers project id from the request, but CreateProject has no project id yet; entry still recorded correctly with actor metadata.
5. **Next/Turbopack dev needs a container restart after import-source changes** — e.g. adding `src/mdx-components.tsx` was not hot-picked-up; `docker compose -f docker/docker-compose.dev.yml restart frontend` fixed it.

## Verified on localhost (pre-Neon, for reference)

- Local Postgres: full migration suite + all 23 backend packages pass `go test ./...`
- Invite email sent via SMTP (Resend): sender-address syntax bug fixed (envelope = bare address, header = `Name <addr>`)
