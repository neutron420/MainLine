# Run Locally

How to run the full SchemaHub stack on your machine.

## Prerequisites

| Tool | Why |
|---|---|
| Go 1.26+ | Backend |
| Node.js 24 + npm | Frontend |
| Docker Desktop | Envoy gateway (optional — see note below) |
| Neon Postgres URL | Database (serverless) |
| Upstash Redis URL | Cache / pub-sub |
| RSA key pair | JWT signing (see `scripts/setup.ps1`) |

Create `.env` in the repo root by copying `.env.example` and filling in real values:

```powershell
Copy-Item .env.example .env
```

The backend reads `../.env` on startup and **auto-runs database migrations** the first time it connects — no manual schema setup needed.

## 1. Backend

```powershell
cd backend
go run ./cmd/server
```

Listens on `:50051` (configurable via `PORT`). Auto-runs migrations and starts all workers (connection health checks, drift alerts, hard-deletes, audit partitions).

## 2. API Gateway (Envoy — required)

The frontend speaks gRPC-Web, the backend speaks plain gRPC. Envoy is the
required bridge in between:

```powershell
docker compose -f docker/docker-compose.yml up envoy
```

Envoy exposes `http://localhost:8080` (gRPC-Web) → `backend:50051`.

**Without Docker:** run the whole stack via `docker compose -f docker/docker-compose.yml up` (includes Envoy), or add a gRPC-Web wrapper to the backend in a future iteration.

## 3. Frontend

```powershell
cd frontend
npm install
npm run dev
```

Open http://localhost:3000. The default API base is `http://localhost:8080`
(Envoy); override with `NEXT_PUBLIC_API_URL`.

## 4. Verify it works

1. Register an account at http://localhost:3000/auth/register
2. Create a project → Dashboard
3. Add a database connection → introspect → schema explorer
4. Settings → Members → invite a teammate by email (must be a registered user)

## Makefile shortcuts (macOS / Linux / WSL)

```bash
make dev          # run backend
make frontend     # run frontend
make test         # backend + frontend tests
make lint         # go fmt/vet + golangci-lint + eslint + buf lint
make proto-gen    # regenerate Go + TS bindings
```

## Troubleshooting

| Symptom | Fix |
|---|---|
| `failed to load config` | `.env` missing required vars — copy `.env.example` and fill it |
| `failed to connect to database` | Check `DATABASE_URL` (Neon: `sslmode=require`) |
| `failed to connect to redis` | Check `REDIS_URL` (Upstash: `rediss://...`) |
| Frontend can't reach API | Envoy not running — `docker compose -f docker/docker-compose.yml up envoy`, or set `NEXT_PUBLIC_API_URL` |
| `protoc plugin es` errors | `cd frontend && npm install` (provides `protoc-gen-es` in `node_modules/.bin`) |
| OAuth callback fails | `*_CALLBACK_URL` must match exactly what you registered in the OAuth provider app |
