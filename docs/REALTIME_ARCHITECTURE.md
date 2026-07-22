# Real-Time Architecture

> **Real-time event streaming, notifications, presence tracking, and collaboration infrastructure for SchemaHub.**

---

## Table of Contents

- [Architecture Overview](#architecture-overview)
- [Event-Driven Design](#event-driven-design)
- [Streaming Infrastructure](#streaming-infrastructure)
- [Connection Lifecycle](#connection-lifecycle)
- [Event Delivery Guarantees](#event-delivery-guarantees)
- [Notifications](#notifications)
- [Presence Tracking](#presence-tracking)
- [Future Collaboration Features](#future-collaboration-features)

---

## Architecture Overview

SchemaHub's real-time architecture uses **gRPC server streaming** for event delivery to clients, with **Redis Pub/Sub** for cross-service event distribution. The Event Service acts as the central hub, managing client subscriptions and event routing.

```
┌──────────────────────────────────────────────────────────────────────┐
│                        REAL-TIME ARCHITECTURE                        │
│                                                                      │
│  ┌──────────┐  ┌──────────┐  ┌──────────┐  ┌──────────┐            │
│  │  Schema  │  │Migration │  │  Drift   │  │  Auth    │            │
│  │ Service  │  │ Service  │  │ Service  │  │ Service  │            │
│  └────┬─────┘  └────┬─────┘  └────┬─────┘  └────┬─────┘            │
│       │              │              │              │                 │
│       └──────────────┼──────────────┼──────────────┘                 │
│                      │              │                                │
│                      ▼              ▼                                │
│              ┌────────────────────────────┐                          │
│              │       Redis Pub/Sub        │                          │
│              │                            │                          │
│              │  Channels:                │                          │
│              │  schema:events:{projId}   │                          │
│              │  schema:events:global     │                          │
│              │  notifications:{userId}   │                          │
│              │  presence:{projectId}     │                          │
│              └────────────┬───────────────┘                          │
│                           │                                          │
│                           ▼                                          │
│              ┌────────────────────────────┐                          │
│              │       Event Service        │                          │
│              │                            │                          │
│              │  ┌──────────────────────┐  │                          │
│              │  │  Subscription Manager │  │                          │
│              │  │  - Per-client streams │  │                          │
│              │  │  - Connection pool   │  │                          │
│              │  │  - Buffer management │  │                          │
│              │  │  - Replay strategy   │  │                          │
│              │  └──────────────────────┘  │                          │
│              │                            │                          │
│              │  ┌──────────────────────┐  │                          │
│              │  │  Presence Manager    │  │                          │
│              │  │  - Heartbeat handler │  │                          │
│              │  │  - Online status     │  │                          │
│              │  │  - TTL cleanup       │  │                          │
│              │  └──────────────────────┘  │                          │
│              └────────────┬───────────────┘                          │
│                           │                                          │
│              ┌────────────┴──────────┐                               │
│              │                       │                               │
│              ▼                       ▼                               │
│  ┌──────────────────┐    ┌──────────────────┐                        │
│  │   Web Browser    │    │  CLI (Future)    │                        │
│  │   gRPC-Web       │    │  gRPC            │                        │
│  └──────────────────┘    └──────────────────┘                        │
└──────────────────────────────────────────────────────────────────────┘
```

---

## Event-Driven Design

### Domain Events

Every state change in SchemaHub produces a domain event. Events are:

1. **Published** to Redis Pub/Sub by the originating service
2. **Persisted** in the audit_logs table for history
3. **Streamed** to connected clients by the Event Service
4. **Dispatched** to other services that need to react

### Event Envelope

```json
{
    "id": "evt_01JABCDEFGHIJKLMNOPQRSTUV",
    "type": "SchemaVersionCreated",
    "version": 1,
    "timestamp": "2026-07-22T14:30:00.000000Z",
    "project_id": "proj_01JABCDEFGHIJKLMNOPQRSTUV",
    "actor": {
        "id": "usr_01JABCDEFGHIJKLMNOPQRSTUV",
        "email": "user@example.com"
    },
    "resource": {
        "type": "schema_version",
        "id": "sv_01JABCDEFGHIJKLMNOPQRSTUV"
    },
    "data": {
        "schema_id": "sch_01JABCDEFGHIJKLMNOPQRSTUV",
        "version": 5,
        "object_count": 42,
        "checksum": "abc123..."
    },
    "metadata": {
        "trace_id": "trace_01JABCDEFGHIJKLMNOPQRSTUV",
        "client_ip": "203.0.113.42"
    }
}
```

### Event Categories

| Category | Events | Consumers |
|---|---|---|
| **Schema** | VersionCreated, SchemaRefreshed, SchemaDeleted | Frontend, Drift Service, Audit |
| **Migration** | Started, Completed, Failed, RolledBack, Progress | Frontend, Audit, Notification |
| **Drift** | Detected, Resolved, Escalated | Frontend, Notification |
| **Connection** | Created, StatusChanged, Deleted | Frontend, Audit |
| **Project** | Created, Updated, Deleted, MemberAdded, MemberRemoved | Frontend, Audit |
| **Auth** | Login, Logout, RoleChanged, PasswordChanged | Audit |

---

## Streaming Infrastructure

### Event Service

The Event Service is responsible for:

1. **Managing client subscriptions** — Each gRPC streaming connection is tracked
2. **Subscribing to Redis channels** — Per-project and global event channels
3. **Buffering events** — Per-client send buffer with overflow handling
4. **Replaying missed events** — On reconnection, events missed during disconnection are replayed
5. **Filtering events** — Clients specify event types they're interested in

### Subscription Manager

```go
// Conceptual structure of the subscription manager
type SubscriptionManager struct {
    mu          sync.RWMutex
    subscribers map[string]*Subscriber  // subscriber_id → subscriber
    channels    map[string]map[string]bool // channel_name → set of subscriber IDs
}

type Subscriber struct {
    ID          string
    UserID      string
    Stream      pb.EventService_SubscribeServer
    EventTypes  []EventType
    ProjectIDs  []string
    Buffer      chan *SchemaEvent
    LastEventID string
    ConnectedAt time.Time
    Done        chan struct{}
}
```

### Redis Channel Structure

| Channel Pattern | Purpose | Retention |
|---|---|---|
| `schema:events:project:{project_id}` | Project-specific events | Fire-and-forget |
| `schema:events:global` | System-wide events | Fire-and-forget |
| `notifications:{user_id}` | User-specific notifications | Fire-and-forget |
| `presence:project:{project_id}` | Online users in project | TTL (30s) |

---

## Connection Lifecycle

### Connection Establishment

```
Client                    Event Service              Redis
  │                           │                       │
  │── Subscribe ─────────────►│                       │
  │   {project_ids,           │                       │
  │    event_types,           │                       │
  │    last_event_id}         │                       │
  │                           │── Validate Auth ─────►│
  │                           │◄── OK ───────────────│
  │                           │                       │
  │                           │── SUBSCRIBE ─────────►│
  │                           │  schema:events:       │
  │                           │  project:{proj1}      │
  │                           │◄── Confirmed ────────│
  │                           │                       │
  │                           │── PUBLISH ───────────►│
  │                           │  presence:{proj1}     │
  │                           │  user joined          │
  │                           │                       │
  │◄── Stream Open ──────────│                       │
```

### Message Delivery

```
  Redis                    Event Service              Client
    │                           │                       │
    │── PUBLISH ───────────────►│                       │
    │   schema:events:          │                       │
    │   project:{proj1}         │                       │
    │                           │── Filter event type  │
    │                           │── Check client perms │
    │                           │── Serialize to proto │
    │                           │── Send to buffer     │
    │                           │── Send to stream ────►│
    │                           │                       │
    │                           │── Wait for send OK   │
    │                           │── Buffer full?       │
    │                           │   → Drop oldest or   │
    │                           │     close connection │
```

### Disconnection and Reconnection

```
Client                    Event Service              Redis
  │                           │                       │
  │── Connection Lost ───────►│                       │
  │                           │── Remove subscriber  │
  │                           │── UNSUBSCRIBE ──────►│
  │                           │── Cancel presence ──►│
  │                           │                       │
  │     (Wait)                │                       │
  │                           │                       │
  │── Subscribe ─────────────►│                       │
  │   {last_event_id:         │                       │
  │    "evt_xxx"}             │                       │
  │                           │── Query audit_logs    │
  │                           │── Replay missed events│
  │                           │── Re-subscribe ──────►│
  │                           │── Re-establish       │
  │◄── Stream Open ──────────│                       │
```

---

## Event Delivery Guarantees

### Delivery Semantics

| Guarantee | Level | Explanation |
|---|---|---|
| **At-most-once** | Redis Pub/Sub → Event Service | Redis Pub/Sub has no delivery confirmation |
| **At-least-once** | Event Service → Client | Events are buffered until client acknowledges, or replayed on reconnect |
| **In-order** | Per-project, per-event-type | Events within a single project and event type are delivered in order |
| **No duplication** | Best effort | Clients should deduplicate by event ID |

### Buffer Management

Each subscriber has a send buffer (default: 100 events):

- **Normal operation:** Events are placed in buffer and sent to stream
- **Buffer approaching full:** Oldest events are dropped (sliding window)
- **Buffer overflow:** The stream is terminated and the client must reconnect

### Reconnection Replay

On reconnection with `last_event_id`:

1. Event Service queries `audit_logs` for events after the given ID
2. Results are limited to the client's subscribed projects and event types
3. Events are replayed in order
4. Live streaming resumes after replay completes

Replay is limited to the last 1000 events or 1 hour, whichever is smaller.

---

## Notifications

### Notification Types

| Type | Trigger | Delivery |
|---|---|---|
| Migration completed | Migration status → completed | Real-time stream |
| Migration failed | Migration status → failed | Real-time stream + optional email |
| Drift detected | Drift check finds mismatches | Real-time stream + scheduled digest |
| Member added | User added to project | Real-time stream |
| Role changed | User's project role changes | Real-time stream |
| Connection lost | Connection test fails | Real-time stream |

### Notification Channels

1. **In-app (real-time stream)** — Delivered via Event Service subscription
2. **Email (future)** — Digest-based or immediate for critical events
3. **Webhook (future)** — Configurable webhook endpoints for CI/CD integration

---

## Presence Tracking

### Purpose

Presence tracking shows which users are currently viewing a project, enabling:

- "3 users viewing this project" indicators
- Awareness of concurrent migration operations
- Future collaboration features (who is working on what)

### Implementation

```
Heartbeat (every 30s from each client):
  → Redis SET presence:{project_id}:{user_id} {timestamp} EX 60

Query Presence:
  → Redis KEYS presence:{project_id}:*
  → For each key, check if timestamp is within 60 seconds

Cleanup:
  → Redis TTL automatically removes stale entries after 60 seconds
```

### Presence Events

| Event | Trigger |
|---|---|
| `user.joined` | First heartbeat for a project after absence |
| `user.present` | Subsequent heartbeats (not broadcasted) |
| `user.left` | Heartbeat stops (detected by TTL expiry) |

---

## Future Collaboration Features

### Phase 2 Real-Time Features

| Feature | Description | Architecture Impact |
|---|---|---|
| **Collaborative cursors** | Show where other users are viewing in schema explorer | Add cursor position to presence data |
| **Migration pair programming** | Two users can work on the same migration | Bidirectional streaming for operation sync |
| **Review comments** | Inline comments on schema objects | Add comments table, comment events |
| **Approval workflows** | Real-time approval notifications | Workflow state machine + events |

### Scaling Considerations

| Component | Scaling Strategy |
|---|---|
| **Event Service** | Horizontal scaling behind load balancer, each instance subscribes to Redis channels independently |
| **Redis Pub/Sub** | Redis Cluster for high throughput. Channel partitioning by project. |
| **Per-client state** | Stored in memory (lightweight, ~2KB per subscriber). Target: 10,000 concurrent connections per instance. |
| **Event buffer** | In-memory ring buffer per subscriber. Overflow → reconnect. |

### Performance Targets

| Metric | Target |
|---|---|
| Event delivery latency (P99) | < 100ms |
| Concurrent connections per instance | 10,000 |
| Events per second (total) | 10,000 |
| Reconnection replay (1000 events) | < 1 second |
| Buffer size per subscriber | 100 events |
