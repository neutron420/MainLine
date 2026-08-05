# SchemaHub — Status & Verification

> Last verified: 05 Aug 2026 — against **live Neon database** (user's own Neon instance) — **ALL GREEN**

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
> Migrations 001–006 are applied on Neon (21 tables, `last_introspected_at` dropped).

---

## Test accounts (registered on Neon)

| Role | Email | Password |
|---|---|---|
| Owner | `e2e.owner@schemahub.dev` | `E2eTest-Pass-123` |
| Member | `e2e.member@schemahub.dev` | `E2eTest-Pass-123` |

**Test projects (owned by e2e.owner):**
- `Invite E2E Test` — invite/accept flow verified here
- `E2E Neon Test` — real Neon connection + schema introspection verified here
- Re-verify anytime: `cd backend && go run ./cmd/e2eprobe`

---

## What is verified (green) on Neon

### Auth
- [x] Register with password policy (8+ chars, upper + lower + digit)
- [x] Login / refresh token flow
- [x] gRPC auth interceptor (401 on bad token)
- [x] Login rate limit (3/min)

### Projects & team
- [x] Create / list projects
- [x] Invite member by email (unregistered → invitation link; registered → added directly)
- [x] Invite duplicate → AlreadyExists; invalid role `developer` → InvalidArgument
- [x] Accept invitation via `/invite/accept?token=...` (7-day expiry; reuse blocked)
- [x] Roles: `owner/admin/member/viewer`

### Connections + Schema Introspection (real Neon DB)
- [x] Create connection (password encrypted — AES-256-GCM, SHA-256 derived key)
- [x] `TestConnection` → **PostgreSQL 18.4 on Neon, db=`neondb`** (latency ~2 s)
- [x] `IntrospectSchema` on `public` — 21 tables + enums + extensions, versioned (checksum dedupe)
- [x] `ListSchemas` shows introspected schema with current version
- [x] Re-introspection is idempotent (same checksum → no duplicate version)
- [x] Fixed: schema `ProjectID` was empty on first introspect (uuid error) — now resolved from connection

### Schema & migrations (backend built, verified by unit tests)
- [x] Introspect → schema versions, schema objects
- [x] Migration run engine, migration logs, drift events
- [x] Audit log for every project action

### Frontend
- [x] Dashboard, projects, schemas, team, connections, settings pages (HTTP 200)
- [x] Sidebar driven by real API data (projects, schemas, members, audit feed, stats)
- [x] Invite dialog, invite-accept page, login redirect
- [x] `tsc --noEmit` clean, `next lint` — No ESLint warnings or errors
- [x] Hydration mismatch fixed: React `19.2.0 → 19.2.8` (data-has-listeners bug) + musl SWC binary for alpine
- [x] Dead field `last_introspected_at` removed end-to-end (proto → DB column → API → UI shows "Last Updated")

### Backend quality
- [x] `go build ./...` clean, `go vet ./...` clean
- [x] `go test ./...` — all 23 packages pass

---

## Known issues

1. **Introspection latency on Neon pooler** — a full introspect takes ~40–60 s (one query per table/column/index/constraint round-trip through the pooler). Fine for on-demand runs; future optimization: run per-table queries concurrently.
2. **Full compose stack must start together** — backend exits if redis is down (observed once after a host restart).

## Verified on localhost (pre-Neon, for reference)

- Local Postgres: full migration suite + all 23 backend packages pass `go test ./...`
- Invite email sent via SMTP (Resend): sender-address syntax bug fixed (envelope = bare address, header = `Name <addr>`)

