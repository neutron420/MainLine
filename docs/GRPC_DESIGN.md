# gRPC Design

> **Complete gRPC architecture for SchemaHub — why gRPC, service boundaries, streaming patterns, error handling, and scalability.**

---

## Table of Contents

- [Why gRPC](#why-grpc)
- [Why Protocol Buffers](#why-protocol-buffers)
- [Unary RPC](#unary-rpc)
- [Server Streaming](#server-streaming)
- [Client Streaming](#client-streaming)
- [Bidirectional Streaming](#bidirectional-streaming)
- [Service Boundaries](#service-boundaries)
- [Communication Flow](#communication-flow)
- [Error Handling](#error-handling)
- [Retry Strategy](#retry-strategy)
- [Connection Lifecycle](#connection-lifecycle)
- [Future Scalability](#future-scalability)

---

## Why gRPC

### Primary Reasons

| Reason | Explanation |
|---|---|
| **Strongly Typed Contracts** | Service interfaces are defined in `.proto` files, generating both server and client code. Schema metadata is complex; type safety at the API boundary prevents entire categories of bugs. |
| **Native Streaming** | gRPC provides first-class support for server streaming, client streaming, and bidirectional streaming. SchemaHub's real-time features (schema change notifications, migration progress) map naturally to these patterns. |
| **HTTP/2** | Multiplexed streams, header compression, and binary framing provide significant performance advantages over HTTP/1.1. |
| **Code Generation** | Protobuf definitions generate Go server code and TypeScript client code, ensuring frontend and backend stay synchronized. |
| **Performance** | Binary serialization is 3-10x faster than JSON serialization, with significantly smaller payload sizes. |

### Why Not REST

- REST over JSON requires manual schema validation on both client and server
- No built-in streaming — requires WebSocket or SSE as a separate mechanism
- No code generation — client and server types drift independently
- Larger payload sizes due to JSON verbosity

### gRPC-Web for Browser Compatibility

The browser cannot natively make gRPC calls (no direct HTTP/2 access from JavaScript). SchemaHub solves this with:

1. **Envoy Proxy** — Acts as a gRPC-Web proxy, converting browser gRPC-Web requests to gRPC and forwarding to backend services
2. **protoc-gen-grpc-web** — Generates TypeScript client code

---

## Why Protocol Buffers

### Advantages

| Aspect | Benefit |
|---|---|
| **Schema Evolution** | Forward and backward compatibility via field numbering |
| **Compact Binary Format** | 3-10x smaller than JSON |
| **Strong Typing** | Enums, oneof, map types |
| **Multi-Language** | Code generation for Go, TypeScript, Python, etc. |
| **Self-Documenting** | `.proto` files serve as authoritative API documentation |

### Versioning Strategy

See [Protobuf Contracts](PROTOBUF_CONTRACTS.md) for detailed versioning strategy.

---

## Unary RPC

### Typical Use Cases

| Operation | Service | Pattern |
|---|---|---|
| Create Project | Project Service | CreateProject → Project |
| List Schemas | Schema Service | ListSchemas → stream Schema |
| Get Migration | Migration Service | GetMigration → Migration |
| Login | Auth Service | Login → TokenResponse |

### Lifecycle

```
1. Client sends single request
2. Server validates (auth, input)
3. Server processes (business logic)
4. Server sends single response
5. Connection closed
```

### When to Use

- CRUD operations
- Authentication (login, register, refresh)
- Any operation that produces a deterministic, bounded result

---

## Server Streaming

### Typical Use Cases

| Operation | Service | Pattern |
|---|---|---|
| Subscribe to schema events | Event Service | Subscribe → stream Event |
| Watch migration progress | Migration Service | WatchMigration → stream MigrationStatus |
| Tail audit logs | Audit Service | TailAuditLog → stream AuditEntry |
| List all schema versions | Schema Service | ListVersions → stream SchemaVersion |

### Lifecycle

```
1. Client sends single request
2. Server validates request
3. Server opens stream
4. Server sends messages as events occur
5. Client or server closes stream
```

### When to Use

- Real-time event delivery
- Large result sets that benefit from paginated streaming
- Progress updates for long-running operations

### Implementation Considerations

- Server maintains a per-client send buffer (configurable, default 100 messages)
- Flow control via gRPC's built-in HTTP/2 flow control
- Keepalive pings every 30 seconds to detect dead connections
- Client reconnection with last-received event ID for continuity

---

## Client Streaming

### Typical Use Cases

| Operation | Service | Pattern |
|---|---|---|
| Bulk schema import | Schema Service | UploadSchemaDump → ImportResult |
| Batch migration submission | Migration Service | SubmitBatchMigrations → BatchResult |

### Lifecycle

```
1. Client opens stream
2. Client sends multiple messages
3. Server acknowledges each (optional)
4. Client closes stream
5. Server sends single response
```

### When to Use

- Bulk operations where data arrives over time
- Large payloads that benefit from chunked transmission

---

## Bidirectional Streaming

### Typical Use Cases

| Operation | Service | Pattern |
|---|---|---|
| Interactive migration debugging | Migration Service | DebugMigration → stream DebugEvent |
| Real-time collaboration (future) | Collaboration Service | Collaborate → stream CollaborationEvent |

### Lifecycle

```
1. Client opens stream
2. Both sides send and receive messages independently
3. Either side closes the stream
```

### When to Use

- Interactive debugging sessions
- Real-time collaboration features (v2+)
- Any protocol where both sides produce and consume events

---

## Service Boundaries

### Service Decomposition Principle

Each service owns its domain completely. A service's domain includes:

- Its data (stored in its own tables or schemas)
- Its business logic
- Its API contract
- Its event definitions

### Service Map

```
┌──────────────────────────────────────────────────────────────┐
│                     API GATEWAY (Envoy)                       │
│  Route by service prefix: /auth.v1.AuthService/               │
│  /project.v1.ProjectService/ /schema.v1.SchemaService/       │
│  /migration.v1.MigrationService/ /event.v1.EventService/     │
│  /audit.v1.AuditService/                                      │
└─────┬──────┬──────┬──────┬──────┬──────┬─────────────────────┘
      │      │      │      │      │      │
      ▼      ▼      ▼      ▼      ▼      ▼
  ┌──────┐ ┌──────┐ ┌──────┐ ┌──────┐ ┌──────┐ ┌──────┐
  │ Auth │ │Proj. │ │Schema│ │Migr. │ │Event │ │Audit │
  │ Serv │ │Serv  │ │Serv  │ │Serv  │ │Serv  │ │Serv  │
  └──────┘ └──────┘ └──────┘ └──────┘ └──────┘ └──────┘
```

### Inter-Service Communication

Services communicate with each other via gRPC. The event service is the central nervous system:

```
Schema Service  ──► Event Service ──► Redis Pub/Sub ──► Event Service ──► Clients
       │                                          │
       └──── PublishEvent(msg) ───────────────────┘
```

### Cross-Cutting Concerns

| Concern | Implementation |
|---|---|
| **Authentication** | Auth interceptor at gateway level |
| **Authorization** | Per-service RBAC checks in handler |
| **Rate Limiting** | Gateway interceptor + Redis counters |
| **Request ID** | Trace ID injected by gateway |
| **Logging** | Per-service structured logging |

---

## Communication Flow

### Internal Service Call (Auth Verification)

```
                  ┌──────────┐      ┌──────────┐
                  │  Service │      │   Auth   │
                  │    A     │      │  Service │
                  └────┬─────┘      └────┬─────┘
                       │                 │
                       │── VerifyToken ──►│
                       │   (gRPC call)   │
                       │                 │── Validate JWT
                       │                 │── Check cache
                       │                 │── Return user
                       │◄──── User ──────│
                       │                 │
```

### Event Publication Flow

```
┌──────────┐   ┌──────────┐   ┌──────────┐   ┌──────────┐
│ Service  │──►│  Event   │──►│  Redis   │──►│  Event   │──► Client
│ (mutates)│   │Publisher │   │ Pub/Sub  │   │ Service  │   (stream)
└──────────┘   └────┬─────┘   └──────────┘   └────┬─────┘
                    │                              │
                    ▼                              ▼
            ┌──────────────┐              ┌──────────────┐
            │  Audit Table │              │  Other       │
            │  (Postgres)  │              │  Services    │
            └──────────────┘              └──────────────┘
```

---

## Error Handling

### Standard gRPC Error Codes

| gRPC Code | HTTP Equivalent | SchemaHub Usage |
|---|---|---|
| OK (0) | 200 | Success |
| InvalidArgument (3) | 400 | Validation failure |
| NotFound (5) | 404 | Resource not found |
| AlreadyExists (6) | 409 | Duplicate resource |
| PermissionDenied (7) | 403 | Insufficient permissions |
| Unauthenticated (16) | 401 | Missing/invalid JWT |
| ResourceExhausted (8) | 429 | Rate limited |
| FailedPrecondition (9) | 400 | Invalid state for operation |
| Aborted (10) | 409 | Conflict (concurrent modification) |
| Internal (13) | 500 | Unexpected error |
| Unavailable (14) | 503 | Service temporarily down |

### Error Detail Enrichment

All errors include a structured detail message:

```protobuf
message ErrorDetail {
    string code = 1;                      // Machine-readable (e.g., "PROJECT_NOT_FOUND")
    string message = 2;                   // Human-readable
    map<string, string> field_errors = 3; // Field-level validation
    string request_id = 4;                // Correlation ID
    string docs_url = 5;                  // Link to documentation
}
```

### Client Error Handling

```typescript
// TypeScript client pattern
try {
    const response = await client.createProject(request);
} catch (error) {
    if (error.code === 'ALREADY_EXISTS') {
        // Handle duplicate project name
    } else if (error.code === 'PERMISSION_DENIED') {
        // Handle insufficient permissions
    } else {
        // Generic error handling
    }
}
```

---

## Retry Strategy

### Client-Side Retries

| Condition | Retry Behavior |
|---|---|
| `Unavailable` (503) | Retry up to 3 times with exponential backoff (100ms, 300ms, 900ms) |
| `ResourceExhausted` (429) | Retry after Retry-After header duration |
| `Internal` (500) | No retry — unexpected error, fail fast |
| `DeadlineExceeded` | No retry — request took too long |

### Idempotency

All mutation operations support idempotency keys:

- Client generates a UUID idempotency key for each mutation
- Server deduplicates requests with the same key within a 5-minute window
- If a client receives a timeout, it can safely retry with the same idempotency key

### Server-Side Retries (Database)

- Transient database errors (`SerializationFailure`, `DeadlockDetected`) are retried up to 3 times
- Connection errors trigger pool health check and retry
- Statement timeouts are not retried (the query needs optimization)

---

## Connection Lifecycle

### Client Connection

```
┌─────────┐                  ┌──────────┐
│  Client  │                 │  Server   │
└────┬────┘                  └────┬─────┘
     │                            │
     │── Open HTTP/2 Connection ──►│
     │                            │── TLS handshake
     │◄── Connection Established ─│
     │                            │
     │── Authenticate (JWT) ─────►│
     │                            │── Validate JWT
     │◄── Authenticated ──────────│
     │                            │
     │── RPC Calls ───────────────►│
     │◄── Responses ──────────────│
     │                            │
     │── Close ──────────────────►│
     │                            │── Cleanup resources
```

### Keepalive

- Client sends HTTP/2 PING every 30 seconds
- Server responds with PING ACK
- If no PING received for 60 seconds, server closes connection
- If no PING ACK received for 30 seconds, client closes connection

### Reconnection

On connection loss:

1. Client immediately attempts reconnect
2. If server-streaming subscription was active, client sends last received event ID
3. Server replays events from that ID (from in-memory buffer or audit log)
4. Client reconciles state

---

## Future Scalability

### Service Mesh

As SchemaHub grows, a service mesh (Istio, Linkerd) can provide:

- Mutual TLS between services
- Traffic splitting for canary deployments
- Circuit breaking
- Distributed tracing

### Event Sourcing with Kafka

If Redis Pub/Sub becomes insufficient:

- Replace Redis Pub/Sub with Kafka for event streaming
- Kafka provides message persistence, replay, and consumer groups
- Services consume events asynchronously rather than via synchronous gRPC

### gRPC Load Balancing

- Client-side load balancing with service discovery
- Look-aside load balancing via Envoy
- gRPC's native look-aside load balancing for HTTP/2
