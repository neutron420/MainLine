# OAuth Integration

> **Complete design for OAuth 2.0 social login in SchemaHub — Google, GitHub, and Slack authentication with account linking and security considerations.**

---

## Table of Contents

- [Overview](#overview)
- [Supported Providers](#supported-providers)
- [Authorization Code Flow](#authorization-code-flow)
- [Provider Configuration](#provider-configuration)
- [Account Linking](#account-linking)
- [Token Management](#token-management)
- [API Design](#api-design)
- [Frontend Flow](#frontend-flow)
- [Error Handling](#error-handling)
- [Security Considerations](#security-considerations)

---

## Overview

SchemaHub supports three OAuth 2.0 identity providers alongside email/password authentication:

| Provider | Use Case | Identity Scope |
|---|---|---|
| **Google** | Broad consumer & enterprise identity | `openid profile email` |
| **GitHub** | Developer-focused authentication | `user:email read:user` |
| **Slack** | Team-based workplace identity | `openid profile email` |

### Design Goals

- **One-click login** — Minimize friction for returning users
- **Account linking** — Connect OAuth identities to existing accounts
- **Verified email** — Delegate email verification to trusted providers
- **Graceful fallback** — Email/password remains available alongside OAuth

---

## Supported Providers

### Google OAuth 2.0

| Property | Value |
|---|---|
| **Authorization endpoint** | `https://accounts.google.com/o/oauth2/v2/auth` |
| **Token endpoint** | `https://oauth2.googleapis.com/token` |
| **Userinfo endpoint** | `https://openidconnect.googleapis.com/v1/userinfo` |
| **Grant type** | Authorization code |
| **Scopes** | `openid profile email` |
| **User identifier** | `sub` (Google Account ID) |
| **Email verification** | Google provides `email_verified` claim in ID token |

### GitHub OAuth 2.0

| Property | Value |
|---|---|
| **Authorization endpoint** | `https://github.com/login/oauth/authorize` |
| **Token endpoint** | `https://github.com/login/oauth/access_token` |
| **User endpoint** | `https://api.github.com/user` |
| **Emails endpoint** | `https://api.github.com/user/emails` |
| **Grant type** | Authorization code |
| **Scopes** | `read:user user:email` |
| **User identifier** | `id` (GitHub User ID) |
| **Email verification** | GitHub returns only verified primary email via `/user/emails` |

**Note:** GitHub does not expose `email_verified` on the primary `/user` endpoint. The emails endpoint (`/user/emails`) must be called separately to retrieve the verified primary email.

### Slack OAuth 2.0 (OpenID Connect)

| Property | Value |
|---|---|
| **Authorization endpoint** | `https://slack.com/openid/connect/authorize` |
| **Token endpoint** | `https://slack.com/api/openid.connect.token` |
| **Userinfo endpoint** | `https://slack.com/api/openid.connect.userInfo` |
| **Grant type** | Authorization code |
| **Scopes** | `openid profile email` |
| **User identifier** | `sub` (Slack Member ID) |
| **Email verification** | Slack provides verified email via OpenID Connect claims |

---

## Authorization Code Flow

All three providers use the standard **OAuth 2.0 Authorization Code Grant** flow.

```
┌──────────┐     ┌──────────┐     ┌──────────┐     ┌──────────────┐
│  Browser │     │ Frontend │     │ Backend │     │ OAuth        │
│          │     │ (Next.js)│     │ (Go)    │     │ Provider     │
└────┬─────┘     └────┬─────┘     └────┬─────┘     └──────┬───────┘
     │                │                │                   │
     │── "Sign in     │                │                   │
     │    with        │                │                   │
     │    Google" ───►│                │                   │
     │                │── GET          │                   │
     │                │   /api/auth/   │                   │
     │                │   google/url ─►│                   │
     │                │                │── Generate state  │
     │                │                │   + PKCE verifier │
     │                │◄── Auth URL ───│                   │
     │◄── Redirect ──│                │                   │
     │    to Google   │                │                   │
     │               ────────────────────────────────────►│
     │                │                │                   │
     │◄── Auth Code ──────────────────────────────────────│
     │    callback    │                │                   │
     │                │                │                   │
     │── POST         │                │                   │
     │   /api/auth/   │                │                   │
     │   google/      │                │                   │
     │   callback ───►│──── gRPC ─────►│                   │
     │                │                │── Exchange code   │
     │                │                │   for tokens ────►│
     │                │                │◄── Access + ID   │
     │                │                │    token ────────│
     │                │                │                   │
     │                │                │── Fetch userinfo  │
     │                │                │   (or decode ID   │
     │                │                │    token) ───────►│
     │                │                │◄── User info ────│
     │                │                │                   │
     │                │                │── Find or create  │
     │                │                │   user            │
     │                │                │── Issue JWT       │
     │                │◄── JWT +       │                   │
     │                │    user ───────│                   │
     │◄── Redirect ──│                │                   │
     │    to dashboard│                │                   │
```

### Step-by-Step

1. **Initiate** — User clicks "Sign in with Google/GitHub/Slack"
2. **Authorization URL** — Frontend calls backend to generate the provider's OAuth URL with `state` and PKCE `code_challenge`
3. **Redirect** — Browser is redirected to the provider's authorization page
4. **Consent** — User authorizes SchemaHub (may be skipped if already authorized)
5. **Callback** — Provider redirects to SchemaHub's callback URL with auth code and state
6. **Exchange** — Backend exchanges the auth code for tokens using the provider's token endpoint
7. **Userinfo** — Backend fetches user identity from the provider's userinfo endpoint (or decodes the ID token for OpenID Connect providers)
8. **Find or create** — Backend looks up existing user by provider + provider_user_id, or creates a new user
9. **Issue JWT** — Standard SchemaHub JWT access + refresh tokens are issued
10. **Redirect** — Frontend receives tokens, redirects to dashboard

---

## Provider Configuration

### Environment Variables

```
# Google
GOOGLE_CLIENT_ID=123456789-xxxx.apps.googleusercontent.com
GOOGLE_CLIENT_SECRET=GOCSPX-xxxxx
GOOGLE_CALLBACK_URL=https://api.schemahub.dev/auth/google/callback

# GitHub
GITHUB_CLIENT_ID=Ov23lixxxxx
GITHUB_CLIENT_SECRET=xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx
GITHUB_CALLBACK_URL=https://api.schemahub.dev/auth/github/callback

# Slack
SLACK_CLIENT_ID=1234567890.xxxxxxxx
SLACK_CLIENT_SECRET=xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx
SLACK_CALLBACK_URL=https://api.schemahub.dev/auth/slack/callback
```

### OAuth State

The `state` parameter is a signed JWT containing:

```json
{
    "sub": "state_01JABCDEFGHIJKLMNOPQRSTUV",
    "provider": "google",
    "redirect_to": "/projects/new",
    "linking": false,
    "code_challenge": "xxxxxxxxx",
    "exp": 1721664900
}
```

Signed with the OAuth state signing key to prevent tampering. Validated on callback.

### PKCE (Proof Key for Code Exchange)

All OAuth flows use PKCE to prevent authorization code interception:

```
code_verifier = random_bytes(32) → base64url
code_challenge = base64url(sha256(code_verifier))

Stored in state JWT for callback validation.
```

---

## Account Linking

### Strategy: Link on First Login + Manual Linking

Users can have multiple OAuth identities linked to a single SchemaHub account.

### Automatic Linking (New User)

```
First OAuth login with any provider:
  → No existing user with this provider identity
  → No existing user with this email
  → Create new user account
  → Link provider identity
  → Issue JWT
```

### Automatic Linking (Existing User, Same Email)

```
OAuth login with email matching existing email/password user:
  → No existing link for this provider identity
  → Existing user found with matching email (email/password account)
  → Prompt: "We found an existing account with this email.
     Sign in with your password to link {Provider} login."
  → User enters password → Authenticate → Link provider → Issue JWT
```

### Manual Linking (User Settings)

```
User → Settings → Connected Accounts
  → "Connect Google" / "Connect GitHub" / "Connect Slack"
  → OAuth flow with linking=true state
  → Provider identity linked to current user
```

### Unlinking

```
User → Settings → Connected Accounts → "Disconnect"
  → Remove provider identity link
  → User must have another authentication method (password or other provider)
```

### Database Schema

Extend the users domain with an `oauth_identities` table:

| Column | Type | Constraints | Purpose |
|---|---|---|---|
| id | UUID (v7) | PK | Unique identity record |
| user_id | UUID | FK → users.id, NOT NULL | Linked user |
| provider | VARCHAR(20) | NOT NULL | 'google', 'github', 'slack' |
| provider_user_id | VARCHAR(200) | NOT NULL | User ID from provider |
| provider_email | VARCHAR(320) | NOT NULL | Email from provider |
| access_token_encrypted | TEXT | Nullable | Provider access token (encrypted) |
| refresh_token_encrypted | TEXT | Nullable | Provider refresh token (encrypted) |
| expires_at | TIMESTAMPTZ | Nullable | Provider token expiry |
| created_at | TIMESTAMPTZ | NOT NULL | Link creation |
| last_used_at | TIMESTAMPTZ | Nullable | Last login via this provider |

**Indexes:**
- `idx_oauth_identities_provider_user` on `(provider, provider_user_id)` (unique)
- `idx_oauth_identities_user_id` on `user_id`

---

## Token Management

### SchemaHub Tokens (Unchanged)

OAuth login produces standard SchemaHub JWT tokens:
- Access token: 15 minutes (RS256)
- Refresh token: 7 days (opaque, hashed in DB)

### Provider Token Storage

Provider access and refresh tokens are stored encrypted (AES-256-GCM) for:
- Token refresh when provider tokens expire
- Future features: schema changes posted to Slack, GitHub commit status

### Provider Token Refresh

Each provider has different token lifetimes:

| Provider | Access Token | Refresh Token | Notes |
|---|---|---|---|
| Google | 1 hour | Available | Use `refresh_token` grant |
| GitHub | No expiry (classic PAT) or 8 hours (OAuth) | None* | GitHub tokens rarely expire |
| Slack | 1 hour | Available | Use `refresh_token` grant |

\* GitHub OAuth tokens do not expire unless revoked, but are limited to the initial scope.

---

## API Design

### Additional AuthService RPCs

```protobuf
service AuthService {
    // Existing methods...

    // OAuth
    rpc GetOAuthURL(GetOAuthURLRequest) returns (GetOAuthURLResponse);
    rpc HandleOAuthCallback(HandleOAuthCallbackRequest) returns (HandleOAuthCallbackResponse);
    rpc LinkOAuthIdentity(LinkOAuthIdentityRequest) returns (LinkOAuthIdentityResponse);
    rpc UnlinkOAuthIdentity(UnlinkOAuthIdentityRequest) returns (UnlinkOAuthIdentityResponse);
    rpc ListLinkedIdentities(ListLinkedIdentitiesRequest) returns (ListLinkedIdentitiesResponse);
}
```

### Key Messages

```protobuf
message GetOAuthURLRequest {
    OAuthProvider provider = 1;      // GOOGLE, GITHUB, SLACK
    string redirect_to = 2;          // Post-login redirect path
    bool linking = 3;                // Linking to existing account?
}

message GetOAuthURLResponse {
    string auth_url = 1;             // Full OAuth authorization URL
    string state_token = 2;          // State token for callback validation
}

message HandleOAuthCallbackRequest {
    OAuthProvider provider = 1;
    string code = 2;                 // Authorization code
    string state = 3;                // State parameter (JWT)
    string code_verifier = 4;        // PKCE verifier
}

message HandleOAuthCallbackResponse {
    string access_token = 1;
    string refresh_token = 2;
    int32 expires_in = 3;
    User user = 4;
    bool is_new_user = 5;            // First login?
    bool needs_linking = 6;          // Existing email found, needs password?
}

message LinkOAuthIdentityRequest {
    OAuthProvider provider = 1;
    string code = 2;                 // Authorization code
    string state = 3;                // State with linking=true
}

enum OAuthProvider {
    OAUTH_PROVIDER_UNSPECIFIED = 0;
    OAUTH_PROVIDER_GOOGLE = 1;
    OAUTH_PROVIDER_GITHUB = 2;
    OAUTH_PROVIDER_SLACK = 3;
}
```

---

## Frontend Flow

### Login Page

```
┌────────────────────────────────────┐
│         Sign in to SchemaHub       │
│                                    │
│   ┌──────────────────────────┐     │
│   │ Email                    │     │
│   └──────────────────────────┘     │
│   ┌──────────────────────────┐     │
│   │ Password                 │     │
│   └──────────────────────────┘     │
│   ┌──────────────────────────┐     │
│   │     Sign In              │     │
│   └──────────────────────────┘     │
│                                    │
│   ──── or continue with ────       │
│                                    │
│   ┌──────────┐ ┌──────────┐ ┌───┐ │
│   │  Google  │ │  GitHub  │ │Sla│ │
│   └──────────┘ └──────────┘ └───┘ │
└────────────────────────────────────┘
```

### OAuth Button Behavior

```tsx
function OAuthButton({ provider }: { provider: OAuthProvider }) {
    const { mutate: getUrl } = useGetOAuthUrl();

    const handleClick = () => {
        getUrl(
            { provider, redirect_to: window.location.pathname },
            {
                onSuccess: (data) => {
                    // Store state_token for callback validation
                    sessionStorage.setItem('oauth_state', data.state_token);
                    // Redirect to provider
                    window.location.href = data.auth_url;
                },
            }
        );
    };

    return <button onClick={handleClick}>Sign in with {provider}</button>;
}
```

### Callback Handler

```tsx
// Called on route: /auth/callback?provider=google&code=xxx&state=yyy
export async function oauthCallback(searchParams: URLSearchParams) {
    const provider = searchParams.get('provider');
    const code = searchParams.get('code');
    const state = searchParams.get('state');
    const savedState = sessionStorage.getItem('oauth_state');

    // Validate state matches
    if (state !== savedState) throw new Error('State mismatch — possible CSRF');

    // Exchange code for tokens
    const result = await authClient.handleOAuthCallback({
        provider,
        code,
        state,
        code_verifier: sessionStorage.getItem('code_verifier'),
    });

    // Handle linking case
    if (result.needs_linking) {
        // Show password dialog to link accounts
        return { view: 'LINKING', email: result.user.email };
    }

    // Store tokens and redirect
    return { view: 'SUCCESS', redirectTo: '/dashboard' };
}
```

---

## Error Handling

| Error | Condition | User Experience |
|---|---|---|
| **User cancelled** | User denies authorization on provider page | Redirect back to login, no error shown |
| **State mismatch** | State parameter doesn't match (CSRF detected) | "Login failed. Please try again." + log security event |
| **Code expired** | Authorization code expires (10 min window) | "Login timed out. Please try again." |
| **Email not verified** | Provider returns unverified email | "Please use a {Provider} account with a verified email." |
| **Email already linked** | Provider email already linked to different account | "This {Provider} account is already linked to another user." |
| **Provider unavailable** | Provider token endpoint unreachable | "{Provider} is temporarily unavailable. Try again later." |
| **Scope denied** | User does not grant required scopes | "SchemaHub needs access to your email address." |
| **Account linking required** | Existing email matches but not linked | Show password dialog for linking confirmation |

### Link-Account Dialog Flow

```
User uses Google login → Email matches existing password account
  → Backend returns needs_linking=true, email=user@example.com
  → Frontend shows dialog:

  ┌────────────────────────────────────┐
  │   Existing account found           │
  │                                    │
  │   An account with                  │
  │   user@example.com already exists. │
  │                                    │
  │   Enter your password to link      │
  │   Google login to this account.    │
  │                                    │
  │   ┌──────────────────────────┐     │
  │   │ Password                 │     │
  │   └──────────────────────────┘     │
  │                                    │
  │   ┌──────────┐  ┌──────────┐      │
  │   │ Cancel   │  │  Link    │      │
  │   └──────────┘  └──────────┘      │
  └────────────────────────────────────┘
```

---

## Security Considerations

### CSRF Protection

| Layer | Mechanism |
|---|---|
| **State parameter** | Signed JWT containing nonce, provider, redirect, expiry |
| **PKCE** | Code verifier required for token exchange |
| **Session binding** | State token stored in sessionStorage, validated on callback |

### Email Verification

- Google and Slack provide `email_verified` claims via OpenID Connect
- GitHub requires separate `/user/emails` API call, only verified primary email is used
- If a provider returns `email_verified: false`, the login is rejected
- SchemaHub does **not** create accounts with unverified emails from OAuth

### Provider Token Security

- Provider access/refresh tokens are encrypted with AES-256-GCM before storage
- Tokens are stored only when needed for future provider API calls
- Provider tokens are never exposed to the frontend

### Scope Minimization

| Provider | Scopes Requested | Why |
|---|---|---|
| Google | `openid profile email` | Identity only, no Google API access |
| GitHub | `read:user user:email` | Profile + verified email, no repo access |
| Slack | `openid profile email` | Identity only, no workspace access |

### Account Takeover Prevention

| Scenario | Protection |
|---|---|
| Attacker steals OAuth link before user completes | State parameter is single-use, expires in 10 minutes |
| Attacker links their OAuth to victim's account | Password confirmation required for linking |
| User loses access to OAuth provider | Email/password login still available if set |
| Provider account compromised | User can unlink provider in settings, admin can force-unlink |

### Rate Limiting

| Endpoint | Limit | Period |
|---|---|---|
| OAuth URL generation | 10 | 1 minute per user |
| OAuth callback | 20 | 1 minute per IP |
| OAuth linking | 5 | 1 minute per user |
| OAuth unlinking | 5 | 1 minute per user |

### Audit Events (OAuth)

| Event | Logged |
|---|---|
| OAuth login success | ✅ |
| OAuth login failure (invalid code) | ✅ |
| OAuth login failure (unverified email) | ✅ |
| OAuth account linked | ✅ |
| OAuth account unlinked | ✅ |
| OAuth state mismatch (CSRF attempt) | ✅ (security alert) |
