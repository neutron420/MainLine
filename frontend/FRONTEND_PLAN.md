# SchemaHub Frontend Plan

## Current Status

### ✅ Auth Pages (Live)

| Route | Page Component | Status |
|---|---|---|
| `/login` | `LoginPage` | ✅ Built |
| `/register` | `RegisterPage` (route: `/signup`) | ✅ Built |
| `/forgot-password` | `ForgotPasswordPage` | ✅ Built |
| `/forgot-password/otp` | `OtpPage` | ✅ Built |
| `/forgot-password/reset` | `ResetPasswordPage` | ✅ Built |
| `/auth/callback/[provider]` | `OAuthCallbackPage` | ✅ Built |

### ✅ Product Pages (Built)

| Route | Page Component | Status |
|---|---|---|
| `/dashboard` | `DashboardPage` | ✅ Built |
| `/projects` | `ProjectsPage` | ✅ Built |
| `/projects/[id]` | `ProjectDetailPage` | ✅ Built |
| `/team` | `TeamPage` — members aggregated across all projects (real data) | ✅ Built |
| `/settings` | `SettingsPage` | ✅ Built |
| `/schemas` | `SchemasPage` — database explorer (tree + table detail + SQL DDL) | ✅ Built |
| `/projects/[id]/schemas/[schemaId]/erd` | `ErdPage` — React Flow entity relationship diagram | ✅ Built |
| `/projects/[id]/schemas/[schemaId]` | `SchemaDetailPage` — table list + stats + ERD/Compare actions | ✅ Built |
| `/projects/[id]/schemas/[schemaId]/compare` | `SchemaComparePage` — two-pane Columns/Indexes/Relations diff | ✅ Built |
| `/projects/[id]/migrations/new` | `CreateMigrationPage` | ✅ Built |
| `/projects/[id]/migrations/[migrationId]` | `MigrationDetailPage` | ✅ Built |
| `/projects/[id]/migrations/[migrationId]/run` | `MigrationRunPage` | ✅ Built |
| `/projects/[id]/connections` | `ConnectionsPage` — linked DB cards (status/latency/SSL) | ✅ Built |
| `/projects/[id]/connections/new` | `CreateConnectionPage` — form + test connection + success state | ✅ Built |
| `/projects/[id]/drift` | `DriftPage` — project-scoped drift reports with status filter | ✅ Built |
| `/projects/[id]/drift/[driftId]` | `DriftDetailPage` — severity banner + schema diff + resolve actions | ✅ Built |
| `/projects/[id]/audit` | `AuditPage` — project-scoped audit log | ✅ Built |
| `/projects/[id]/events` | `EventsPage` — project-scoped timeline with type filter | ✅ Built |
| `/projects/[id]/settings` | `ProjectSettingsPage` — general, migration policy, danger zone | ✅ Built |
| `/projects/[id]/settings/members` | `ProjectMembersPage` — roles + invite dialog | ✅ Built |
| `/projects/new` | `CreateProjectPage` — templates + Neon source + create flow | ✅ Built |
| `/settings/connections` | `LinkedAccountsPage` — OAuth accounts + connected databases | ✅ Built |

> `/projects/[id]` tabs are all live including Settings (links to settings/members pages). Drift tab rows link into drift detail; Schemas tab links into schema detail; ERD/Compare linked.

### ⬜ Pages To Build (Future)

None — all planned pages are built.

### 📦 Installed shadcn/ui Components

Button, Input, Textarea, Label, Select, Card, Dialog, DropdownMenu, Tabs, Skeleton, Table, Form, Sidebar, Sheet, Separator, Tooltip, Sonner, Avatar, Badge, Checkbox, Drawer, Toggle, Chart, Breadcrumb, Collapsible, ToggleGroup, Alert, Popover

### 🔧 Shared Data Modules

### 🔧 Data Layer

All pages are wired to real gRPC-Web clients (`src/lib/api/` + `src/lib/gen/`). No mock data modules remain — the former `src/lib/*-data.ts` files were deleted. Reviews pages were removed because no Reviews service exists in the backend contract. Notifications popover streams live events (Redis pub/sub) and seeds from the audit log.

---

## Summary

| Category | Count | Progress |
|---|---|---|
| Auth Pages | 6 | ✅ Done |
| Product Pages | 21 | ✅ Done |
| Protected Pages | 0 | ✅ Done |
| shadcn/ui Components | 28 | 📦 Installed |

> Build passes with 34 routes. All frontend pages complete and wired to the real API.

