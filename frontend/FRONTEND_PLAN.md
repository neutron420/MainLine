# SchemaHub Frontend Plan

## Current Status

### ✅ Auth Pages (Live)

| Route | Page Component | Status |
|---|---|---|
| `/login` | `LoginPage` | ✅ Built |
| `/register` | `RegisterPage` | ✅ Built |
| `/forgot-password` | `ForgotPasswordPage` | ✅ Built |
| `/forgot-password/otp` | `OtpPage` | ✅ Built |
| `/forgot-password/reset` | `ResetPasswordPage` | ✅ Built |
| `/auth/callback/[provider]` | `OAuthCallbackPage` | ✅ Built |

### 📦 Installed shadcn/ui Components

Button, Input, Textarea, Label, Select, Card, Dialog, DropdownMenu, Tabs, Skeleton, Table, Form, Sidebar, Sheet, Separator, Tooltip, Sonner, Avatar, Badge, Checkbox, Drawer, Toggle, Chart, Breadcrumb, Collapsible, ToggleGroup

---

### ⬜ Pages To Build (Future)

| Route | Page Component |
|---|---|
| `/dashboard` | `DashboardPage` |
| `/projects` | `ProjectsPage` |
| `/projects/new` | `CreateProjectPage` |
| `/projects/[id]` | `ProjectDetailPage` |
| `/projects/[id]/connections` | `ConnectionsPage` |
| `/projects/[id]/connections/new` | `CreateConnectionPage` |
| `/projects/[id]/schemas` | `SchemasPage` |
| `/projects/[id]/schemas/[schemaId]` | `SchemaDetailPage` |
| `/projects/[id]/schemas/[schemaId]/erd` | `ErdPage` |
| `/projects/[id]/schemas/[schemaId]/compare` | `SchemaComparePage` |
| `/projects/[id]/migrations` | `MigrationsPage` |
| `/projects/[id]/migrations/new` | `CreateMigrationPage` |
| `/projects/[id]/migrations/[migrationId]` | `MigrationDetailPage` |
| `/projects/[id]/migrations/[migrationId]/run` | `MigrationRunPage` |
| `/projects/[id]/drift` | `DriftPage` |
| `/projects/[id]/drift/[driftId]` | `DriftDetailPage` |
| `/projects/[id]/audit` | `AuditPage` |
| `/projects/[id]/events` | `EventsPage` |
| `/projects/[id]/settings` | `ProjectSettingsPage` |
| `/projects/[id]/settings/members` | `MembersPage` |
| `/settings` | `SettingsPage` |
| `/settings/connections` | `LinkedAccountsPage` |

---

## Summary

| Category | Count | Progress |
|---|---|---|
| Auth Pages | 6 | ✅ Done |
| Protected Pages | 22 | ⬜ Not started |
| shadcn/ui Components | 26 | 📦 Installed |
