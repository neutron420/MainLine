# Deployment

> **Deployment architecture, environment configuration, Docker setup, CI/CD pipeline, monitoring, and operational procedures for SchemaHub.**

---

## Table of Contents

- [Deployment Overview](#deployment-overview)
- [Architecture](#architecture)
- [Development Environment](#development-environment)
- [Staging Environment](#staging-environment)
- [Production Environment](#production-environment)
- [Docker Configuration](#docker-configuration)
- [Environment Variables](#environment-variables)
- [CI/CD Pipeline](#cicd-pipeline)
- [Monitoring](#monitoring)
- [Logging](#logging)
- [Backup Strategy](#backup-strategy)
- [Disaster Recovery](#disaster-recovery)
- [Runbooks](#runbooks)

---

## Deployment Overview

SchemaHub uses a **three-environment deployment model** with automated CI/CD pipelines.

```
┌─────────────────┐     ┌─────────────────┐     ┌─────────────────┐
│   Development   │────►│    Staging      │────►│   Production    │
│                 │     │                 │     │                 │
│  • Local Docker │     │  • Preview env  │     │  • Multi-region │
│  • Hot reload   │     │  • Test data    │     │  • Auto-scaling │
│  • Dev database │     │  • CI pipeline  │     │  • Traffic mgmt │
└─────────────────┘     └─────────────────┘     └─────────────────┘
```

---

## Architecture

```
                         ┌──────────────┐
                         │   Vercel     │
                         │  (Frontend)  │
                         │  Next.js SSR │
                         └──────┬───────┘
                                │
                        gRPC-Web │
                                ▼
                    ┌──────────────────────┐
                    │   Envoy Proxy        │
                    │   (TLS, Routing)     │
                    └──────────┬───────────┘
                               │ gRPC
                               ▼
               ┌───────────────────────────────┐
               │    Backend Services (Go)       │
               │   ┌─────┐ ┌─────┐ ┌─────┐    │
               │   │Auth │ │Proj │ │Sch  │    │
               │   └─────┘ └─────┘ └─────┘    │
               │   ┌─────┐ ┌─────┐ ┌─────┐    │
               │   │Migr │ │Event│ │Audit│    │
               │   └─────┘ └─────┘ └─────┘    │
               └──────┬──────────┬─────────────┘
                      │          │
                      ▼          ▼
            ┌──────────────┐  ┌──────────┐
            │  PostgreSQL  │  │  Redis   │
            │   (Neon)     │  │  Cluster │
            └──────────────┘  └──────────┘
```

---

## Development Environment

### Local Development with Docker Compose

```yaml
services:
  backend:
    build:
      context: ./backend
      dockerfile: ../docker/backend/Dockerfile.dev
    ports:
      - "50051:50051"
    environment:
      - DATABASE_URL=postgres://postgres:postgres@postgres:5432/schemahub
      - REDIS_URL=redis://redis:6379/0
    volumes:
      - ./backend:/app  # Hot reload
    depends_on:
      - postgres
      - redis

  frontend:
    build:
      context: ./frontend
      dockerfile: ../docker/frontend/Dockerfile.dev
    ports:
      - "3000:3000"
    environment:
      - NEXT_PUBLIC_API_URL=http://localhost:50051
    volumes:
      - ./frontend:/app

  postgres:
    image: postgres:16-alpine
    environment:
      POSTGRES_DB: schemahub
      POSTGRES_PASSWORD: postgres
    ports:
      - "5432:5432"
    volumes:
      - pgdata:/var/lib/postgresql/data

  redis:
    image: redis:7-alpine
    ports:
      - "6379:6379"

  envoy:
    image: envoyproxy/envoy:v1.28-latest
    volumes:
      - ./docker/envoy/envoy.yaml:/etc/envoy/envoy.yaml
    ports:
      - "8080:8080"
```

### Quick Start

```bash
# Clone and start
git clone https://github.com/schemahub/schemahub.git
cd schemahub
docker compose -f docker/docker-compose.yml -f docker/docker-compose.dev.yml up

# Run migrations
docker compose exec backend go run ./cmd/migrate

# Development with hot reload
# Backend uses air (Go hot reload)
# Frontend uses Next.js HMR
```

---

## Staging Environment

### Purpose

- Integration testing before production
- Preview deployments for PRs
- Load testing and performance validation

### Configuration

| Component | Configuration |
|---|---|
| **Frontend** | Vercel preview deployment |
| **Backend** | Single instance, Railway/Fly.io |
| **Database** | Neon branch (fork from production schema) |
| **Redis** | Single instance, Railway/Fly.io |
| **Domain** | `staging.schemahub.dev` |

### Data

- Anonymized production data subset
- Synthetic test data for load testing
- Database reset on each deployment

---

## Production Environment

### Infrastructure

| Component | Provider | Configuration |
|---|---|---|
| **Frontend** | Vercel | 2+ instances, auto-scaling, CDN |
| **Backend** | Railway / Fly.io / AWS ECS | 2+ instances, auto-scaling, multi-AZ |
| **Database** | Neon | Serverless, auto-scaling, read replicas |
| **Redis** | Upstash / Redis Cloud | Cluster mode, replication |
| **DNS** | Vercel DNS / Cloudflare | CDN, DDoS protection |
| **Monitoring** | Grafana Cloud | Metrics, logs, alerts |

### Backend Scaling

```
Load Balancer (Envoy)
       │
       ├── Backend Instance 1 (us-east-1)
       ├── Backend Instance 2 (us-east-1)
       ├── Backend Instance 3 (eu-west-1)
       └── Backend Instance N (auto-scaling)
```

### Service-Level Objectives

| Metric | Target |
|---|---|
| API availability (monthly) | 99.9% |
| API response time (P99) | < 500ms |
| Event delivery (P99) | < 200ms |
| Schema introspection (< 100 tables) | < 5s |
| Migration execution | As reported (depends on SQL) |

---

## Docker Configuration

### Backend Dockerfile (Production)

```dockerfile
# Build stage
FROM golang:1.22-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /server ./cmd/server

# Runtime stage
FROM gcr.io/distroless/base-debian12
COPY --from=builder /server /server
EXPOSE 50051
ENTRYPOINT ["/server"]
```

### Frontend Dockerfile (Production)

```dockerfile
FROM node:20-alpine AS builder
WORKDIR /app
COPY package*.json ./
RUN npm ci
COPY . .
RUN npm run build

FROM node:20-alpine AS runner
WORKDIR /app
COPY --from=builder /app/.next ./.next
COPY --from=builder /app/public ./public
COPY --from=builder /app/package.json ./package.json
EXPOSE 3000
ENTRYPOINT ["npm", "start"]
```

---

## Environment Variables

### Backend

```bash
# Required
DATABASE_URL=postgres://user:pass@host:5432/schemahub?sslmode=require
REDIS_URL=redis://:password@host:6379/0
JWT_PRIVATE_KEY=-----BEGIN RSA PRIVATE KEY-----\n...
JWT_PUBLIC_KEY=-----BEGIN PUBLIC KEY-----\n...
ENCRYPTION_MASTER_KEY=base64-32byte-key

# Optional (with defaults)
PORT=50051
LOG_LEVEL=info
LOG_FORMAT=json
RATE_LIMIT_PER_MINUTE=100
STREAM_BUFFER_SIZE=100
DB_POOL_MIN=2
DB_POOL_MAX=20
```

### Frontend

```bash
# Build-time (public)
NEXT_PUBLIC_API_URL=https://api.schemahub.dev
NEXT_PUBLIC_WS_URL=wss://api.schemahub.dev
NEXT_PUBLIC_SENTRY_DSN=...

# Runtime (server-side only)
API_INTERNAL_URL=http://backend:50051
SCHEMAHUB_ENV=production
```

---

## CI/CD Pipeline

### GitHub Actions

```yaml
name: CI/CD Pipeline

on:
  push:
    branches: [main]
  pull_request:
    branches: [main]

jobs:
  lint:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - name: Lint Go
        run: cd backend && golangci-lint run
      - name: Lint Frontend
        run: cd frontend && npm run lint

  test:
    runs-on: ubuntu-latest
    services:
      postgres:
        image: postgres:16-alpine
        env:
          POSTGRES_PASSWORD: postgres
        options: >-
          --health-cmd pg_isready
          --health-interval 10s
          --health-timeout 5s
          --health-retries 5
      redis:
        image: redis:7-alpine
        options: >-
          --health-cmd "redis-cli ping"
          --health-interval 10s

    steps:
      - uses: actions/checkout@v4
      - name: Run Go Tests
        run: cd backend && go test -race -coverprofile=coverage.out ./...
      - name: Run Frontend Tests
        run: cd frontend && npm run test

  deploy-staging:
    needs: [lint, test]
    if: github.event_name == 'pull_request'
    steps:
      - name: Deploy Backend to Staging
        run: # ... deploy to Railway/Fly.io
      - name: Deploy Frontend to Vercel Preview
        run: # ... Vercel CLI

  deploy-production:
    needs: [lint, test]
    if: github.ref == 'refs/heads/main'
    steps:
      - name: Deploy Backend
        run: # ... deploy to production
      - name: Deploy Frontend
        run: # ... Vercel production deploy
```

### Pipeline Stages

```
PR Opened → Lint → Test → Build → Deploy Preview → Preview URL
Merge to main → Lint → Test → Build → Deploy Staging → Run E2E Tests → Deploy Production
```

---

## Monitoring

### Metrics (Prometheus)

| Metric | Type | Description |
|---|---|---|
| `grpc_requests_total` | Counter | Total gRPC requests by method and status |
| `grpc_request_duration_seconds` | Histogram | Request latency distribution |
| `migration_duration_seconds` | Histogram | Migration execution time |
| `introspection_duration_seconds` | Histogram | Schema introspection time |
| `active_subscriptions` | Gauge | Current real-time subscriptions |
| `db_connections` | Gauge | Active database connections |
| `event_delivery_latency` | Histogram | Event delivery time from publish to client |

### Alerts

| Alert | Condition | Severity |
|---|---|---|
| High error rate | Error rate > 1% for 5 minutes | Critical |
| High latency | P99 latency > 1s for 5 minutes | Warning |
| Migration failure | Migration execution fails | Critical |
| Schema drift detected | Drift event created | Warning |
| Database connection pool exhausted | Pool utilization > 90% | Critical |
| Certificate expiring | TLS cert expires in < 30 days | Warning |

### Dashboards (Grafana)

- **Overview** — Request rate, error rate, latency, active users
- **Schema** — Introspection duration, version count, drift events
- **Migrations** — Execution duration, success rate, rollback rate
- **Real-time** — Active subscriptions, event delivery latency
- **Database** — Connection pool, query performance, cache hit rate

---

## Logging

### Log Collection

- **Structured JSON logs** from all services
- **Log shipping** via vector/fluentbit to Grafana Loki or similar
- **Centralized search** in Grafana or dedicated log management

### Log Retention

| Environment | Retention | Storage |
|---|---|---|
| Development | 7 days | Local files |
| Staging | 30 days | Centralized |
| Production | 90 days (hot) + 1 year (cold archive) | S3/GCS |

### Alerting on Logs

- ERROR level logs trigger alerts
- Repeated errors within time window trigger escalation
- Known error patterns can be suppressed via configuration

---

## Backup Strategy

### Database Backups (Neon)

| Backup Type | Frequency | Retention | Recovery |
|---|---|---|---|
| **Point-in-time recovery** | Continuous | 7 days | Any point in last 7 days |
| **Daily snapshot** | Daily | 30 days | Specific day |
| **Weekly snapshot** | Weekly | 3 months | Specific week |

### Application Backups

- No application state to backup (stateless services)
- Configuration is in version control (Infrastructure as Code)

### Disaster Recovery

| Scenario | RTO | RPO | Recovery Method |
|---|---|---|---|
| Single instance failure | < 1 min | 0 | Auto-scaling group replaces |
| AZ failure | < 5 min | < 1 min | Cross-AZ failover |
| Region failure | < 30 min | < 5 min | Cross-region failover |
| Data corruption | < 1 hour | < 1 hour | PITR to before corruption |

---

## Runbooks

### Common Operational Tasks

- **Deploy a new version**: `git push main` → automated pipeline
- **Rollback a deployment**: Revert commit or use deployment rollback button
- **View logs**: `grafana logs explorer → filter by service and trace_id`
- **Scale up**: Adjust auto-scaling group min/max (or Neon auto-scaling handles it)
- **Database migration**: Apply via deployment pipeline (not manual)
- **SSL certificate renewal**: Automated via LetsEncrypt/Cert Manager
- **Incident response**: Follow incident response runbook in internal docs
