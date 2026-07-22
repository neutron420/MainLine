# Authentication

> **Complete authentication and authorization architecture for SchemaHub — JWT, refresh tokens, RBAC, permission model, and security considerations.**

---

## Table of Contents

- [Authentication Overview](#authentication-overview)
- [Authentication Methods](#authentication-methods)
- [JWT Design](#jwt-design)
- [Refresh Token Design](#refresh-token-design)
- [Token Lifecycle](#token-lifecycle)
- [Authorization Model](#authorization-model)
- [Role-Based Access Control](#role-based-access-control)
- [Permission Matrix](#permission-matrix)
- [API Security](#api-security)
- [Security Considerations](#security-considerations)
- [OAuth / Social Login](#oauth--social-login)

---

## Authentication Overview

SchemaHub uses a **JWT-based authentication system** with short-lived access tokens and long-lived refresh tokens, supporting multiple authentication methods including email/password and OAuth social login. The design prioritizes:

1. **Stateless verification** — Access tokens can be verified without database lookups
2. **Security** — Short token lifetimes limit exposure from token theft
3. **User experience** — Automatic token refresh eliminates frequent re-login, one-click OAuth login
4. **gRPC compatibility** — Tokens are passed via gRPC metadata headers

```
┌─────────────────────────────────────────────────────────────┐
│                    AUTHENTICATION FLOW                       │
│                                                             │
│  ┌────────┐    ┌──────────┐    ┌───────────┐               │
│  │ Client │    │ Frontend │    │  Backend  │               │
│  │        │    │ (Next.js)│    │ (Go/gRPC) │               │
│  └────┬───┘    └────┬─────┘    └─────┬─────┘               │
│       │             │                │                      │
│       │ Login Form  │                │                      │
│       ├────────────►│                │                      │
│       │             │ Login gRPC     │                      │
│       │             ├───────────────►│                      │
│       │             │                │── Verify credentials │
│       │             │                │── Generate JWT + RT  │
│       │             │◄───────────────│                      │
│       │             │                │                      │
│       │  Store RT   │                │                      │
│       │◄────────────┤                │                      │
│       │             │                │                      │
│       │ API Call +  │                │                      │
│       │ JWT in Meta │                │                      │
│       ├────────────►│──── gRPC ─────►│── Verify JWT         │
│       │             │                │── Process request    │
│       │◄────────────│◄───────────────│                      │
│       │             │                │                      │
│       │  401 →      │                │                      │
│       │  Refresh    │                │                      │
│       ├────────────►│── Refresh ────►│── Verify RT          │
│       │             │                │── Rotate tokens      │
│       │◄────────────│◄───────────────│                      │
│       │  Retry API  │                │                      │
│       └─────────────┘                └──────────────────────┘
```

---

## JWT Design

### Token Structure

```json
{
    "header": {
        "alg": "RS256",
        "typ": "JWT",
        "kid": "2026-07-v1"
    },
    "payload": {
        "sub": "usr_01JABCDEFGHIJKLMNOPQRSTUV",
        "email": "user@example.com",
        "role": "user",
        "iat": 1721664000,
        "exp": 1721664900,
        "jti": "jti_01JABCDEFGHIJKLMNOPQRSTUV"
    }
}
```

### Claims

| Claim | Description | Example |
|---|---|---|
| `sub` | Subject — User ID | `usr_01JABCDEFGHIJKLMNOPQRSTUV` |
| `email` | User email | `user@example.com` |
| `role` | Global user role | `user` |
| `iat` | Issued at (Unix timestamp) | `1721664000` |
| `exp` | Expiration (Unix timestamp) | `1721664900` |
| `jti` | JWT ID — unique identifier | `jti_01JABCDEFGHIJKLMNOPQRSTUV` |

### Signing Algorithm

**RS256 (RSA Signature with SHA-256)** — asymmetric signing

- **Private key** — Used by Auth Service to sign tokens. Stored securely, never exposed.
- **Public key** — Used by all services to verify tokens. Distributed via configuration or JWKS endpoint.

**Why RS256 over HS256:**
- HS256 uses a shared secret — any service with the secret can forge tokens
- RS256 allows the public key to be safely distributed without compromising signing capability
- Key rotation is cleaner (add new key ID, phase out old)

### Token Lifetime

| Token | Lifetime | Purpose |
|---|---|---|
| Access Token | 15 minutes | Authorization for API calls |
| Refresh Token | 7 days | Obtain new access tokens |

---

## Refresh Token Design

### Token Structure

Refresh tokens are opaque random strings (not JWTs):

- Format: `rt_` + 48 bytes of cryptographically secure random data, base64url encoded
- Example: `rt_abc123...` (88 characters total)

### Storage

Refresh tokens are stored in PostgreSQL with a SHA-256 hash:

```sql
INSERT INTO refresh_tokens (id, user_id, token_hash, expires_at, created_by_ip, family)
VALUES ($1, $2, SHA256($3), now() + '7 days', $4, $5)
```

**Why hash storage:**
- If the database is compromised, refresh tokens cannot be extracted
- Token rotation: each use generates a new token in the same family
- Old tokens are revoked on use (detects token reuse)

### Token Family

Each refresh token belongs to a "family". When a token is rotated:

1. The current token is revoked
2. A new token is created in the same family
3. If a revoked token is used (reuse attack), **all tokens in the family are revoked**

This prevents token theft: if an attacker steals a refresh token, the legitimate user's next request will detect the reuse and invalidate the attacker's token.

### Rotation Pattern

```
                    Family: "fam_abc"
                         │
                    ┌────┴────┐
                    │         │
               Token v1    Token v2
               (revoked)   (active)

If token v1 is used AFTER being revoked:
  → All tokens in "fam_abc" are revoked
  → User must re-authenticate
```

---

## Token Lifecycle

### Issuance

```
POST / Login
  → Validate credentials
  → Generate access token (RS256, 15min)
  → Generate refresh token (random, 7 days)
  → Store refresh token hash in DB
  → Return both tokens
```

### Verification (Access Token)

```
Every gRPC call
  → Extract token from gRPC metadata (Authorization: Bearer <token>)
  → Verify signature with public key (RS256)
  → Check expiration (exp)
  → Extract user claims from payload
  → Attach user context to gRPC context
```

### Refresh (Access Token)

```
POST / RefreshToken
  → Hash received refresh token
  → Look up hash in refresh_tokens table
  → Verify token is not revoked or expired
  → Revoke current token (detect reuse)
  → Generate new access token
  → Generate new refresh token (same family)
  → Store new refresh token hash
  → Return both tokens
```

### Revocation

```
Logout
  → Hash refresh token
  → Set revoked_at = now()
  → Token can no longer be used

Force Logout (Admin)
  → Revoke ALL refresh tokens for a user
  → User must re-authenticate
```

---

## Authorization Model

### Two-Level Authorization

SchemaHub uses a two-level authorization model:

1. **Global Level** — User role (admin, user)
   - Determines global capabilities (system settings, user management)
   - Set at account creation, rarely changes

2. **Project Level** — Membership role (owner, admin, member, viewer)
   - Determines capabilities within a project
   - Set when user is added to project

### Authorization Flow

```
Request → Auth Interceptor → Parse JWT → Extract User ID + Global Role
         → Route to Service Handler
         → Check Project Membership (if project-scoped)
         → Check Project Role
         → Execute or Deny
```

### Authorization Check Pattern (Go)

```go
func (i *AuthInterceptor) Authorize(ctx context.Context, requiredRole ProjectRole) error {
    user := GetUserFromContext(ctx)     // From JWT
    project := GetProjectFromContext(ctx) // From request

    membership, err := i.repo.GetMembership(ctx, user.ID, project.ID)
    if err != nil {
        return fmt.Errorf("checking membership: %w", err)
    }
    if membership == nil {
        return ErrPermissionDenied
    }
    if membership.Role.Permissions() < requiredRole.Permissions() {
        return ErrPermissionDenied
    }
    return nil
}
```

---

## Role-Based Access Control

### Global Roles

| Role | Permissions |
|---|---|
| **admin** | Full system access, user management, all projects |
| **user** | Standard access, own projects, invited projects |

### Project Roles

| Role | Level | Permissions |
|---|---|---|
| **owner** | 100 | Full project control, delete project, manage roles |
| **admin** | 80 | Manage connections, migrations, members (except owner) |
| **member** | 50 | Create migrations, run introspection, view schemas |
| **viewer** | 10 | View schemas, versions, migration history (read-only) |

---

## Permission Matrix

| Operation | Owner | Admin | Member | Viewer |
|---|---|---|---|---|
| View project | ✅ | ✅ | ✅ | ✅ |
| Edit project settings | ✅ | ✅ | ❌ | ❌ |
| Delete project | ✅ | ❌ | ❌ | ❌ |
| Manage members | ✅ | ✅ (except owner) | ❌ | ❌ |
| Change roles | ✅ | ❌ | ❌ | ❌ |
| Add connection | ✅ | ✅ | ❌ | ❌ |
| Edit connection | ✅ | ✅ | ❌ | ❌ |
| Delete connection | ✅ | ✅ | ❌ | ❌ |
| View connections | ✅ | ✅ | ✅ | ✅ |
| Introspect schema | ✅ | ✅ | ✅ | ✅ |
| View schemas | ✅ | ✅ | ✅ | ✅ |
| View versions | ✅ | ✅ | ✅ | ✅ |
| View diffs | ✅ | ✅ | ✅ | ✅ |
| Create migration | ✅ | ✅ | ✅ | ❌ |
| Execute migration | ✅ | ✅ | ✅ | ❌ |
| Rollback migration | ✅ | ✅ | ✅ | ❌ |
| View audit logs | ✅ | ✅ | ✅ | ✅ |
| View drift events | ✅ | ✅ | ✅ | ✅ |

---

## API Security

### gRPC Metadata Headers

```
Authorization: Bearer <access_token>
X-Idempotency-Key: <uuid>
X-Trace-ID: <uuid>
```

### Unauthenticated Endpoints

The following endpoints do not require authentication:

| Service | Method | Reason |
|---|---|---|
| AuthService | Register | Account creation |
| AuthService | Login | Authentication |
| AuthService | RefreshToken | Token refresh |
| AuthService | SendVerificationEmail | Pre-auth |
| AuthService | VerifyEmail | Pre-auth |

All other endpoints require a valid JWT access token.

### CORS

The gRPC gateway (Envoy) enforces CORS policies:

- Allowed origins: configurable (development: `*`, production: specific domains)
- Allowed methods: POST, GET, OPTIONS
- Allowed headers: Authorization, Content-Type, X-Idempotency-Key, X-Trace-ID
- Credentials: included (cookies)

---

## Security Considerations

### Token Storage (Client)

- **Access token** — Stored in memory only (not persisted to localStorage/sessionStorage)
- **Refresh token** — Stored in HTTP-only, Secure, SameSite=Strict cookie
- **No tokens in URL** — Tokens are never passed as query parameters

### Token Theft Mitigation

| Threat | Mitigation |
|---|---|
| Access token theft | Short TTL (15 min), limited damage window |
| Refresh token theft | HTTP-only cookie prevents XSS access |
| Refresh token reuse | Token family revocation |
| XSS attack | No token in localStorage, HTTP-only cookies |
| CSRF attack | SameSite=Strict, CORS validation |
| Man-in-the-middle | TLS for all connections |

### Rate Limiting

| Endpoint | Rate Limit | Period |
|---|---|---|
| Login | 5 attempts | 1 minute per IP |
| Register | 3 attempts | 1 hour per IP |
| RefreshToken | 10 attempts | 1 minute per user |
| All other endpoints | 100 requests | 1 minute per user |
| Schema introspection | 10 requests | 1 minute per connection |

### Password Policy

- Minimum length: 8 characters
- Requires: uppercase, lowercase, digit
- Maximum length: 128 characters
- Hashing algorithm: bcrypt with cost factor 12
- No plaintext password storage
- Password change requires current password verification

### Session Management

- Maximum active sessions per user: 10
- Admin can force-logout all sessions
- Password change revokes all existing sessions
- Account deactivation prevents all authentication

### Audit Events (Authentication)

| Event | Logged |
|---|---|
| Login success | ✅ |
| Login failure | ✅ (without password) |
| OAuth login success | ✅ |
| OAuth login failure | ✅ |
| Token refresh | ✅ |
| Token revocation | ✅ |
| Password change | ✅ |
| Email verification | ✅ |
| Account creation | ✅ |
| Account deletion | ✅ |
| Role change | ✅ |
| Force logout | ✅ |

---

## OAuth / Social Login

SchemaHub supports OAuth 2.0 social login via Google, GitHub, and Slack. See [OAuth Integration](OAUTH_INTEGRATION.md) for the complete design covering:

- OAuth 2.0 authorization code flow for all three providers
- Account linking (OAuth ↔ existing email/password account)
- Automatic account creation on first OAuth login
- Provider-specific configuration (client IDs, secrets, scopes, callbacks)
- State parameter with PKCE for CSRF protection
- Token refresh and session management for OAuth-originated sessions
- Security considerations per provider (verified email enforcement, scope minimization)
