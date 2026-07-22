# Tech Stack

> **Detailed analysis of every technology choice in the SchemaHub stack, including rationale, advantages, disadvantages, alternatives, and trade-offs.**

---

## Table of Contents

- [Overview](#overview)
- [Frontend](#frontend)
- [Backend](#backend)
- [API Protocol](#api-protocol)
- [Database](#database)
- [Cache](#cache)
- [Authentication](#authentication)
- [Containerization](#containerization)
- [Deployment](#deployment)
- [Technology Matrix](#technology-matrix)

---

## Overview

SchemaHub's tech stack was chosen to satisfy four core requirements:

1. **Real-time capabilities** — Schema changes must be streamed to connected clients
2. **Strong typing at every boundary** — Database schemas are complex; type safety prevents errors
3. **Performance** — Schema introspection and migration execution must be fast
4. **Production readiness** — Every component must be battle-tested at scale

---

## Frontend

### Next.js (App Router)

| Aspect | Detail |
|---|---|
| **Version** | 14+ |
| **Router** | App Router |
| **Rendering** | Server Components by default, Client Components for interactivity |

**Why:** Next.js provides the best developer experience for React applications with built-in routing, server-side rendering, and strong TypeScript support. The App Router enables nested layouts, loading states, and error boundaries without additional libraries. The Vercel ecosystem provides zero-configuration deployment.

**Advantages:**
- Server Components reduce client-side JavaScript
- Automatic code splitting and lazy loading
- Built-in image optimization and font loading
- Middleware for auth checks and request rewriting

**Disadvantages:**
- Tight coupling to Vercel for optimal deployment
- Server Components add complexity around client state management
- Bundle size can grow quickly without careful monitoring

**Alternatives:** Remix, SvelteKit, Astro

**Trade-off:** Next.js provides the most mature React ecosystem but comes with Vercel lock-in for the best deployment experience.

### React

| Aspect | Detail |
|---|---|
| **Version** | 18+ |
| **Component Model** | Functional with hooks |
| **Server Components** | Yes (Next.js) |

**Why:** React is the industry standard for building user interfaces. The component model maps well to SchemaHub's modular UI. Hooks enable clean separation of state and effects.

### TypeScript (strict mode)

| Aspect | Detail |
|---|---|
| **Config** | `strict: true`, `noUncheckedIndexedAccess` |
| **Generated Types** | From protobuf via `protoc-gen-ts` |

**Why:** TypeScript catches entire categories of bugs at compile time. For a platform dealing with complex database metadata, type safety is critical.

### Tailwind CSS

**Why:** Utility-first CSS enables rapid UI development without leaving the HTML. Consistent design tokens can be configured centrally. Tree-shaking produces minimal production CSS.

**Alternatives:** CSS Modules, Styled Components, Vanilla Extract

**Trade-off:** Utility classes can make JSX verbose, but shadcn/ui patterns mitigate this with composed components.

### shadcn/ui

**Why:** Provides accessible, unstyled React components that are copied into the project (not a dependency). Every component can be customized directly. Built on top of Radix UI for accessibility.

### React Flow

**Why:** The leading library for interactive node-based diagrams in React. Used for schema visualization (ERDs), migration flow diagrams, and dependency graphs.

**Alternatives:** D3.js, Vis.js, Mermaid (rendered)

**Trade-off:** React Flow is opinionated about layout; custom node types require more effort. However, it provides interactivity (pan, zoom, drag) out of the box.

### TanStack Query

| Aspect | Detail |
|---|---|
| **Role** | Server state management |
| **Caching** | Automatic, configurable TTL |
| **Mutations** | Optimistic updates, rollback on error |

**Why:** TanStack Query eliminates the need for global state stores (Redux, Zustand) for server data. It provides declarative caching, automatic refetching, pagination, and WebSocket integration.

**Alternatives:** Redux Toolkit Query, SWR, Apollo Client

**Trade-off:** Adds ~12KB to the bundle but eliminates hundreds of lines of state management code.

---

## Backend

### Go

| Aspect | Detail |
|---|---|
| **Version** | 1.22+ |
| **Build** | Single binary |
| **Standard Library** | Extensive — HTTP/2, crypto, context |

**Why:** Go is the gold standard for backend services that need performance, concurrency, and simplicity. For SchemaHub:

- **goroutines** enable efficient per-connection streaming
- **Fast compilation** means rapid development cycles
- **Single binary deployment** simplifies containerization
- **Excellent gRPC support** via the official Go gRPC library

**Advantages:**
- Low memory footprint — critical for many concurrent streaming connections
- Goroutines are lightweight (2KB stack) vs threads (1MB+)
- Strong standard library reduces third-party dependencies
- Static typing catches errors at compile time

**Disadvantages:**
- No generics in early versions (1.18+ addresses this)
- Verbose error handling
- Smaller ecosystem than Java or Node.js for certain domains

**Alternatives:** Rust (more performant but slower development), Node.js (weaker concurrency model), Java (heavier, slower startup)

**Trade-off:** Go sacrifices expressiveness for simplicity and performance. This is the right trade for a streaming-heavy Platform Engineering tool.

### gRPC + Protocol Buffers

| Aspect | Detail |
|---|---|
| **Protocol** | HTTP/2 |
| **Serialization** | Protocol Buffers (binary) |
| **Code Generation** | Both server (Go) and client (TS) |

**Why:** gRPC is the best choice for service-to-service communication in a streaming-heavy platform. See [gRPC Design](GRPC_DESIGN.md) for detailed reasoning.

### pgx (PostgreSQL Driver)

| Aspect | Detail |
|---|---|
| **Role** | PostgreSQL driver for Go |
| **Pool** | pgxpool |
| **Prepared Statements** | Automatic |

**Why:** pgx is the fastest PostgreSQL driver for Go, with native support for PostgreSQL-specific features (COPY, LISTEN/NOTIFY, array types, JSONB). It is actively maintained and battle-tested at scale.

**Alternatives:** database/sql + lib/pq, GORM (ORM — not suitable)

**Trade-off:** pgx is lower-level than an ORM, but SchemaHub needs direct SQL control for schema introspection and migration execution.

---

## API Protocol

### gRPC vs REST vs GraphQL

| Criteria | gRPC | REST | GraphQL |
|---|---|---|---|
| **Strong typing** | ✅ Built-in (protobuf) | ❌ Manual validation | ✅ Schema |
| **Streaming** | ✅ Native | ❌ SSE or WebSocket | ❌ Subscriptions |
| **Performance** | ✅ Binary encoding | ❌ JSON | ❌ JSON |
| **Code generation** | ✅ Both sides | ❌ Manual | ✅ Schema → TS |
| **Browser support** | ❌ Needs gateway | ✅ Native | ✅ Native |
| **Debugging** | ❌ Binary hard to read | ✅ Easy with curl | ✅ Easy |
| **Tooling** | ⭐ Good | ⭐ Excellent | ⭐ Good |

**Decision:** SchemaHub uses **gRPC** for all service-to-service and backend-to-frontend communication, with **gRPC-Web** (via Envoy or gRPC Gateway) for browser clients. The streaming requirements make gRPC the clear winner despite the gateway overhead.

---

## Database

### PostgreSQL (Neon)

| Aspect | Detail |
|---|---|
| **Version** | 16 (Neon-compatible) |
| **Hosting** | Neon (serverless PostgreSQL) |
| **Driver** | pgx (Go) |

**Why PostgreSQL:**
- Most advanced open-source relational database
- JSONB for flexible schema metadata storage
- Extensive indexing (B-tree, GiST, GIN, BRIN)
- LISTEN/NOTIFY for simple pub/sub
- Mature tooling (pg_stat_statements, pg_stat_activity)
- Strong consistency guarantees

**Why Neon:**
- Serverless PostgreSQL with automatic scaling
- Database branching for isolated schema workspaces
- Read replicas for scaling introspection queries
- Cold start mitigation (pools connections)

**Alternatives for PostgreSQL:** AWS RDS, Supabase, TimescaleDB

**Alternatives for Database:** MySQL, SQLite, CockroachDB

**Trade-off:** PostgreSQL is slower than MySQL for simple reads but significantly more capable for complex queries. Neon's serverless model adds latency for cold starts but provides automatic scaling.

### Data Storage Strategy

| Data Type | Storage | Rationale |
|---|---|---|
| **Entities (users, projects)** | PostgreSQL tables | Strong consistency, relational queries |
| **Schema metadata** | PostgreSQL tables + JSONB | Structured metadata with flexible schema diffs |
| **Migration history** | PostgreSQL tables | Immutable, append-only, transactional |
| **Audit logs** | PostgreSQL tables (partitioned) | Append-only, time-range queries |
| **Events** | Redis Pub/Sub + PostgreSQL | Low-latency delivery + persistent history |
| **Sessions** | Redis | TTL-based, fast lookup |
| **Rate limiting** | Redis | Atomic counters, TTL-based |

---

## Cache

### Redis

| Aspect | Detail |
|---|---|
| **Version** | 7+ |
| **Mode** | Standalone (development), Sentinel/Cluster (production) |
| **Go Client** | go-redis/redis |

**Why:** Redis provides the pub/sub infrastructure needed for real-time event distribution, plus caching and rate limiting in a single well-understood system.

**Advantages:**
- Blazing fast for key-value operations (< 1ms)
- Pub/Sub with channel pattern matching
- Atomic operations for rate limiting
- TTL-based key expiration for automatic cache invalidation

**Disadvantages:**
- Data must fit in memory
- Pub/Sub messages are fire-and-forget (no persistence)
- No built-in message replay

**Trade-off:** Redis Pub/Sub is not a message queue. If a subscriber disconnects, messages are lost. For SchemaHub, this is acceptable because:
- Events are also written to PostgreSQL (audit log)
- Clients re-subscribe and receive the latest state on reconnection
- Critical notifications can be implemented with Redis Streams (which support persistence)

---

## Authentication

### JWT (Access + Refresh Tokens)

| Aspect | Detail |
|---|---|
| **Token Type** | Access (short-lived) + Refresh (long-lived) |
| **Signing** | RS256 (asymmetric) |
| **Storage** | HTTP-only cookies (access) + secure cookie (refresh) |

**Why:** JWT provides stateless authentication that works natively with gRPC. The access/refresh pattern balances security (short-lived access tokens) with user experience (automatic refresh).

See [Authentication](AUTHENTICATION.md) for detailed design.

---

## Containerization

### Docker + Docker Compose

| Aspect | Detail |
|---|---|
| **Base Image (Go)** | `golang:1.22-alpine` (build), `gcr.io/distroless/base` (runtime) |
| **Base Image (Node)** | `node:20-alpine` |
| **Compose Services** | backend, frontend, redis, envoy |

**Why:** Docker provides reproducible environments across development, CI, and production. Docker Compose simplifies local development with a single command.

---

## Deployment

### Frontend: Vercel

**Why:** Zero-configuration Next.js deployment, automatic preview deployments, global CDN, environment variable management.

### Backend: Railway / Fly.io / AWS

| Option | Use Case |
|---|---|
| **Railway** | Fastest setup, integrated PostgreSQL |
| **Fly.io** | Global edge deployment, WireGuard networking |
| **AWS ECS/EKS** | Enterprise compliance, existing infrastructure |

---

## Technology Matrix

```
                    ┌─────────────────────────────────────────────────────┐
                    │                   SCHEMAHUB                         │
                    │                                                     │
                    │   ┌─────────────────────────────────────────────┐   │
                    │   │              FRONTEND (Next.js)              │   │
                    │   │  React + TypeScript + Tailwind + shadcn/ui  │   │
                    │   │  React Flow (diagrams)                      │   │
                    │   │  TanStack Query (data)                      │   │
                    │   │  gRPC-Web + protobuf-ts (API client)        │   │
                    │   └─────────────────────────────────────────────┘   │
                    │                        │                            │
                    │                        │ gRPC-Web                   │
                    │                        ▼                            │
                    │   ┌─────────────────────────────────────────────┐   │
                    │   │            ENVOY / GRPC GATEWAY              │   │
                    │   │  TLS · Rate Limit · Auth · gRPC-Web Proxy  │   │
                    │   └─────────────────────────────────────────────┘   │
                    │                        │                            │
                    │                        │ gRPC                       │
                    │                        ▼                            │
                    │   ┌─────────────────────────────────────────────┐   │
                    │   │              BACKEND (Go)                    │   │
                    │   │  Auth · Project · Schema · Migration        │   │
                    │   │  Event · Audit · Drift                      │   │
                    │   │  pgx (PostgreSQL) · go-redis (Redis)        │   │
                    │   └─────────────────────────────────────────────┘   │
                    │                        │                            │
                    │            ┌───────────┴───────────┐                │
                    │            ▼                       ▼                │
                    │   ┌──────────────┐       ┌──────────────┐          │
                    │   │  PostgreSQL  │       │    Redis      │          │
                    │   │   (Neon)     │       │  Cache/PubSub │          │
                    │   └──────────────┘       └──────────────┘          │
                    └─────────────────────────────────────────────────────┘
```
