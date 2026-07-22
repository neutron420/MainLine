# Security

> **Security architecture for SchemaHub — authentication, authorization, input validation, secrets management, rate limiting, transport security, and best practices.**

---

## Table of Contents

- [Security Philosophy](#security-philosophy)
- [Authentication](#authentication)
- [Authorization](#authorization)
- [Input Validation](#input-validation)
- [SQL Injection Prevention](#sql-injection-prevention)
- [Secrets Management](#secrets-management)
- [Rate Limiting](#rate-limiting)
- [Transport Security](#transport-security)
- [Infrastructure Security](#infrastructure-security)
- [Incident Response](#incident-response)

---

## Security Philosophy

SchemaHub follows a **defense-in-depth** approach with multiple security layers:

1. **Network layer** — TLS, network policies, firewall rules
2. **Transport layer** — mTLS between services, HTTP/2 encryption
3. **API layer** — Authentication, authorization, input validation, rate limiting
4. **Application layer** — Parameterized queries, encryption at rest, secure configuration
5. **Data layer** — Encrypted secrets, audit logging, backup encryption

---

## Authentication

See [Authentication](AUTHENTICATION.md) for the complete authentication design.

### Key Security Properties

| Property | Implementation |
|---|---|
| **Password storage** | bcrypt with cost factor 12 |
| **Token signing** | RS256 (asymmetric) — private key only on Auth Service |
| **Access token lifetime** | 15 minutes |
| **Refresh token lifetime** | 7 days |
| **Token storage (client)** | Memory (access) + HTTP-only cookie (refresh) |
| **Token revocation** | Refresh token family revocation on detected theft |
| **Brute force protection** | Rate limiting per IP and per user |

### Authentication Endpoints

All authentication endpoints are rate-limited and monitored for anomalies. Failed login attempts do not indicate whether the email or password was incorrect.

---

## Authorization

### Two-Level Authorization

| Level | Granularity | Enforcement |
|---|---|---|
| **Global** | System-wide (admin/user) | Auth interceptor |
| **Project** | Per-project (owner/admin/member/viewer) | Service handler |

### Authorization Check Pattern

```go
func (s *Service) executeMigration(ctx context.Context, req *pb.ExecuteMigrationRequest) error {
    // 1. Authenticate — extract user from JWT (auth interceptor)
    user := auth.GetUserFromContext(ctx)

    // 2. Authorize — check project-level permissions
    membership, err := s.membershipRepo.Get(ctx, user.ID, req.ProjectId)
    if err != nil {
        return ErrPermissionDenied
    }
    if membership.Role.PermissionLevel < ProjectRoleMember {
        return ErrPermissionDenied
    }

    // 3. Execute
    // ...
}
```

### Authorization Rules

- Every project-scoped operation checks the user's project role
- Owner and admin roles have access to sensitive operations (connection management, member management)
- Viewer role is read-only
- Admin global role bypasses project-level checks

---

## Input Validation

### Server-Side Validation

Every gRPC endpoint validates its input before processing:

| Validation | Implementation | Enforcement |
|---|---|---|
| **Required fields** | Proto field presence check | Validation interceptor |
| **String length** | Configurable max/min | Validation interceptor |
| **Enum values** | Proto enum validation | gRPC code generation |
| **Email format** | Regex validation | Handler-specific |
| **UUID format** | Parse as UUID | Handler-specific |
| **SQL syntax** | PostgreSQL parser | Migration service |

### Client-Side Validation

Frontend validates input before sending to reduce unnecessary API calls:

- Form field validation with error messages
- Real-time validation feedback
- All client validation is duplicated on the server

### Validation Rules

```
email:    required, max 320 chars, regex: /^[^\s@]+@[^\s@]+\.[^\s@]+$/
password: required, min 8 chars, max 128 chars, must contain uppercase, lowercase, digit
name:     required, max 200 chars
slug:     required, max 200 chars, regex: /^[a-z0-9-]+$/
SQL:      valid PostgreSQL syntax, no disallowed statements
host:     required, valid hostname or IP
port:     1-65535
```

---

## SQL Injection Prevention

### Primary Defense: Parameterized Queries

All database queries use pgx's parameterized query interface. User input is never concatenated into SQL strings.

```go
// GOOD: Parameterized query
rows, err := pool.Query(ctx, "SELECT * FROM users WHERE email = $1", email)

// BAD: String concatenation (never used)
rows, err := pool.Query(ctx, "SELECT * FROM users WHERE email = '"+email+"'")
```

### Secondary Defense: Migration SQL Validation

When users provide SQL for migrations, the system:

1. Parses the SQL with a PostgreSQL-compatible parser
2. Validates that no dangerous statements are included:
   - `DROP DATABASE`
   - `DROP TABLESPACE`
   - `ALTER SYSTEM`
   - `CREATE EXTENSION` (restricted)
3. Rejects SQL that cannot be parsed

### Defense Layers

```
User Input → gRPC Layer (binary protobuf) → Validation Layer → Parameterized Query
                                                                    ↓
                                                            PostgreSQL (query parsing)
```

---

## Secrets Management

### Types of Secrets

| Secret | Storage | Encryption |
|---|---|---|
| Database passwords (connections) | PostgreSQL (connections table) | AES-256-GCM at application layer |
| JWT signing private key | Environment variable or secret manager | At rest by infrastructure |
| Password encryption master key | Environment variable or secret manager | At rest by infrastructure |
| Database connection string (SchemaHub's own DB) | Environment variable | At rest by platform |
| Redis password | Environment variable | At rest by platform |

### Encryption Flow (Connection Passwords)

```
User submits password
  → Frontend sends via gRPC-Web (TLS encrypted)
  → Backend receives in handler
  → Encryption service:
    1. Generate random 12-byte nonce
    2. Encrypt with AES-256-GCM using master key
    3. Encode as base64(nonce + ciphertext)
  → Stored in database
```

### Decryption Flow

```
Backend needs to connect to user's database
  → Encryption service:
    1. Decode base64 to get nonce + ciphertext
    2. Decrypt with AES-256-GCM using master key
  → Password available in memory for connection
  → Password reference is zeroed after use (not guaranteed, but attempted)
```

### Environment Variables

```bash
# Required
DATABASE_URL=postgres://user:pass@host:5432/schemahub
REDIS_URL=redis://:password@host:6379/0
JWT_PRIVATE_KEY=-----BEGIN RSA PRIVATE KEY-----\n...
JWT_PUBLIC_KEY=-----BEGIN PUBLIC KEY-----\n...
ENCRYPTION_MASTER_KEY=base64-encoded-32-byte-key

# Optional (with defaults)
PORT=50051
LOG_LEVEL=info
RATE_LIMIT_PER_MINUTE=100
```

---

## Rate Limiting

### Strategy

SchemaHub uses a **token bucket** algorithm implemented with Redis:

```
Key: ratelimit:{identifier}:{endpoint_group}
Expiry: 60 seconds
Value: remaining tokens (decremented on each request)

On request:
  1. GET key → remaining tokens
  2. If ≤ 0 → return ResourceExhausted
  3. DECR key → continue
  4. If key doesn't exist → SET key with limit-1, EX 60
```

### Rate Limits

| Endpoint Group | Limit | Period | Scope |
|---|---|---|---|
| Authentication (login) | 5 | 1 minute | Per IP |
| Authentication (register) | 3 | 1 hour | Per IP |
| Authentication (refresh) | 10 | 1 minute | Per user |
| Project CRUD | 30 | 1 minute | Per user |
| Schema introspection | 10 | 1 minute | Per connection |
| Migration execution | 5 | 1 minute | Per connection |
| General API | 100 | 1 minute | Per user |
| Event subscription | 10 | 1 minute | Per user |

### Headers

Rate limit information is returned via gRPC metadata (or HTTP headers through gateway):

```
X-RateLimit-Limit: 100
X-RateLimit-Remaining: 87
X-RateLimit-Reset: 1621584000
```

---

## Transport Security

### TLS Configuration

| Connection | TLS | Notes |
|---|---|---|
| Browser → Envoy | TLS 1.3 | Public certificates (LetsEncrypt) |
| Envoy → Backend Services | mTLS (optional) | Internal CA |
| Backend → PostgreSQL (Neon) | TLS 1.2+ | Required by Neon |
| Backend → Redis | TLS 1.2+ | Configurable |

### gRPC Security

- All gRPC calls go through TLS-terminated Envoy proxy
- gRPC metadata carries auth tokens
- Inter-service calls use mTLS in production (optional in development)

### CORS

```
AllowedOrigins: production domain(s)
AllowedMethods: POST, GET, OPTIONS
AllowedHeaders: Authorization, Content-Type, X-Idempotency-Key, X-Trace-ID
AllowCredentials: true
```

---

## Infrastructure Security

| Layer | Controls |
|---|---|
| **Network** | VPC, private subnets for databases, security group rules |
| **Compute** | Minimal base images, no root access, read-only filesystem |
| **Database** | Private network, TLS connections, automated backups, point-in-time recovery |
| **Redis** | Password authentication, private network |
| **CI/CD** | Scoped access tokens, no secrets in build logs |
| **Monitoring** | Anomaly detection on auth patterns, DDoS monitoring |

### Container Security

- Images scanned for vulnerabilities in CI
- Run as non-root user
- Read-only root filesystem
- No shell in production containers (distroless base images)
- Resource limits enforced

---

## Incident Response

### Security Event Detection

| Event | Detection | Response |
|---|---|---|
| Multiple failed logins | Rate limiter triggers | Temporary IP block, notify admin |
| Token reuse | Refresh token family revocation | Revoke all family tokens, notify user |
| Possible SQL injection | Query pattern analysis | Log, alert, block IP if confirmed |
| Unauthorized access attempt | Auth interceptor denial | Log with details, alert if repeated |
| Rate limit threshold exceeded | Rate limiter | Log, temporary block |

### Response Procedures

1. **Automated** — Rate limiting, token revocation, IP blocking
2. **Manual** — Admin review of audit logs, account suspension, database rollback
3. **Post-mortem** — Security incident analysis, remediation plan

### Compliance Considerations

SchemaHub is designed to support:

- **SOC 2** — Audit logging, access controls, change management
- **HIPAA** — Encryption at rest and in transit, access controls, audit trails
- **GDPR** — Data deletion, export, and privacy controls
- **SOX** — Immutable audit trails, change tracking, access controls
