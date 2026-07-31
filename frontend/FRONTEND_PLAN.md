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
| `/reviews` | `ReviewsPage` | ✅ Built |
| `/reviews/[id]` | `ReviewDetailPage` | ✅ Built |
| `/team` | `TeamPage` | ✅ Built |
| `/settings` | `SettingsPage` | ✅ Built |
| `/schemas` | `SchemasPage` — database explorer (tree + table detail + SQL DDL) | ✅ Built |
| `/projects/[id]/schemas/[schemaId]/erd` | `ErdPage` — React Flow entity relationship diagram | ✅ Built |

> `/projects/[id]` tabs are all live: Overview, Schemas, Migrations, Drift, Audit, Events (only Settings tab is a placeholder).

### ⬜ Pages To Build (Future)

| Route | Page Component |
|---|---|
| `/projects/new` | `CreateProjectPage` (currently a dialog) |
| `/projects/[id]/connections` | `ConnectionsPage` |
| `/projects/[id]/connections/new` | `CreateConnectionPage` |
| `/projects/[id]/schemas` | `SchemasPage` (project-scoped) |
| `/projects/[id]/schemas/[schemaId]` | `SchemaDetailPage` |
| `/projects/[id]/schemas/[schemaId]/compare` | `SchemaComparePage` |
| `/projects/[id]/migrations` | `MigrationsPage` (project-scoped) |
| `/projects/[id]/migrations/new` | `CreateMigrationPage` |
| `/projects/[id]/migrations/[migrationId]` | `MigrationDetailPage` |
| `/projects/[id]/migrations/[migrationId]/run` | `MigrationRunPage` |
| `/projects/[id]/drift` | `DriftPage` (project-scoped) |
| `/projects/[id]/drift/[driftId]` | `DriftDetailPage` |
| `/projects/[id]/audit` | `AuditPage` (project-scoped) |
| `/projects/[id]/events` | `EventsPage` (project-scoped) |
| `/projects/[id]/settings` | `ProjectSettingsPage` |
| `/projects/[id]/settings/members` | `MembersPage` |
| `/settings/connections` | `LinkedAccountsPage` |

### 📦 Installed shadcn/ui Components

Button, Input, Textarea, Label, Select, Card, Dialog, DropdownMenu, Tabs, Skeleton, Table, Form, Sidebar, Sheet, Separator, Tooltip, Sonner, Avatar, Badge, Checkbox, Drawer, Toggle, Chart, Breadcrumb, Collapsible, ToggleGroup, Alert, Popover

### 🔧 Shared Data Modules

- `src/lib/reviews-data.ts` — review types + mock data for `/reviews` and `/reviews/[id]`
- `src/lib/schemas-data.ts` — schema explorer types + mock data for `/schemas`

---

## Summary

| Category | Count | Progress |
|---|---|---|
| Auth Pages | 6 | ✅ Done |
| Product Pages | 9 | ✅ Done |
| Protected Pages | 13 | ⬜ Remaining |
| shadcn/ui Components | 28 | 📦 Installed |

> Build passes with 21 routes. Next up: migration sub-pages or schema compare.

