# SchemaHub Frontend Plan

## After Login

User gets redirected to → **`/dashboard`**

---

## Status Legend

| Icon | Meaning |
|------|---------|
| ✅ | Built (page or component exists) |
| ⬜ | Not yet built |
| 📦 | shadcn/ui component installed |

---

## All Pages & Components

---

### 1. Auth Pages (✅ Page logic done, pages exist)

| Route | Page Component | Child Components (File Name) |
|---|---|---|
| `/login` | `LoginPage.tsx` | `LoginForm.tsx`, `OAuthButtons.tsx`, `Divider.tsx` |
| `/register` | `RegisterPage.tsx` | `RegisterForm.tsx`, `OAuthButtons.tsx` |
| `/forgot-password` | `ForgotPasswordPage.tsx` | `EmailForm.tsx` |
| `/forgot-password/otp` | `OtpPage.tsx` | `OtpInput.tsx`, `Timer.tsx` |
| `/forgot-password/reset` | `ResetPasswordPage.tsx` | `ResetPasswordForm.tsx` |
| `/auth/callback/[provider]` | `OAuthCallbackPage.tsx` | `LoadingSpinner.tsx` |

---

### 2. Dashboard

| Route: `/dashboard` |
|---|
| **Page**: `DashboardPage.tsx` |
| **Components**: `ProjectCard.tsx`, `RecentActivity.tsx`, `QuickStats.tsx`, `WelcomeBanner.tsx`, `EmptyState.tsx` |

---

### 3. Projects

| Route: `/projects` |
|---|
| **Page**: `ProjectsPage.tsx` |
| **Components**: `ProjectCard.tsx`, `CreateProjectDialog.tsx`, `ProjectFilters.tsx`, `ProjectGrid.tsx`, `EmptyState.tsx` |

| Route: `/projects/new` |
|---|
| **Page**: `CreateProjectPage.tsx` |
| **Components**: `ProjectForm.tsx` |

---

### 4. Project Detail

| Route: `/projects/[id]` |
|---|
| **Page**: `ProjectDetailPage.tsx` |
| **Components**: `ProjectHeader.tsx`, `ConnectionList.tsx`, `SchemaSummary.tsx`, `RecentMigrations.tsx`, `DriftAlerts.tsx` |

---

### 5. Connections

| Route: `/projects/[id]/connections` |
|---|
| **Page**: `ConnectionsPage.tsx` |
| **Components**: `ConnectionCard.tsx`, `ConnectionFormDialog.tsx`, `ConnectionTestButton.tsx`, `ConnectionStringInput.tsx`, `EmptyState.tsx` |

| Route: `/projects/[id]/connections/new` |
|---|
| **Page**: `CreateConnectionPage.tsx` |
| **Components**: `ConnectionForm.tsx`, `ConnectionTestButton.tsx` |

---

### 6. Schemas

| Route: `/projects/[id]/schemas` |
|---|
| **Page**: `SchemasPage.tsx` |
| **Components**: `SchemaCard.tsx`, `SchemaVersionBadge.tsx`, `IntrospectButton.tsx`, `EmptyState.tsx` |

| Route: `/projects/[id]/schemas/[schemaId]` |
|---|
| **Page**: `SchemaDetailPage.tsx` |
| **Components**: `SchemaHeader.tsx`, `VersionHistory.tsx`, `ObjectTree.tsx`, `SchemaDiffView.tsx` |

---

### 7. Schema ERD

| Route: `/projects/[id]/schemas/[schemaId]/erd` |
|---|
| **Page**: `ErdPage.tsx` |
| **Components**: `ErdCanvas.tsx` (React Flow), `TableNode.tsx`, `RelationEdge.tsx`, `ObjectPalette.tsx`, `ErdControls.tsx` |

---

### 8. Schema Compare

| Route: `/projects/[id]/schemas/[schemaId]/compare` |
|---|
| **Page**: `SchemaComparePage.tsx` |
| **Components**: `VersionSelector.tsx`, `DiffTable.tsx`, `ObjectDiffCard.tsx`, `SqlPreview.tsx` |

---

### 9. Migrations

| Route: `/projects/[id]/migrations` |
|---|
| **Page**: `MigrationsPage.tsx` |
| **Components**: `MigrationCard.tsx`, `MigrationStatusBadge.tsx`, `CreateMigrationDialog.tsx`, `EmptyState.tsx` |

| Route: `/projects/[id]/migrations/new` |
|---|
| **Page**: `CreateMigrationPage.tsx` |
| **Components**: `MigrationForm.tsx`, `SqlEditor.tsx`, `DestructiveOpWarning.tsx`, `MigrationPreview.tsx` |

| Route: `/projects/[id]/migrations/[migrationId]` |
|---|
| **Page**: `MigrationDetailPage.tsx` |
| **Components**: `MigrationTimeline.tsx`, `SqlViewer.tsx`, `RunLog.tsx`, `RollbackButton.tsx` |

| Route: `/projects/[id]/migrations/[migrationId]/run` |
|---|
| **Page**: `MigrationRunPage.tsx` |
| **Components**: `RunProgress.tsx` (streaming), `StatementLog.tsx`, `ResultSummary.tsx` |

---

### 10. Drift Detection

| Route: `/projects/[id]/drift` |
|---|
| **Page**: `DriftPage.tsx` |
| **Components**: `DriftCard.tsx`, `SeverityBadge.tsx`, `DriftFilter.tsx`, `DriftDetailModal.tsx`, `ResolveButton.tsx`, `EmptyState.tsx` |

| Route: `/projects/[id]/drift/[driftId]` |
|---|
| **Page**: `DriftDetailPage.tsx` |
| **Components**: `DiffViewer.tsx`, `ObjectComparison.tsx`, `ResolveForm.tsx` |

---

### 11. Audit Logs

| Route: `/projects/[id]/audit` |
|---|
| **Page**: `AuditPage.tsx` |
| **Components**: `AuditTable.tsx`, `AuditFilter.tsx`, `EventDetailModal.tsx`, `EmptyState.tsx` |

---

### 12. Real-time Events

| Route: `/projects/[id]/events` |
|---|
| **Page**: `EventsPage.tsx` |
| **Components**: `EventStream.tsx`, `EventCard.tsx`, `EmptyState.tsx` |

---

### 13. User Settings

| Route: `/settings` |
|---|
| **Page**: `SettingsPage.tsx` |
| **Components**: `ProfileForm.tsx`, `AvatarUpload.tsx`, `ChangePasswordForm.tsx` |

| Route: `/settings/connections` |
|---|
| **Page**: `LinkedAccountsPage.tsx` |
| **Components**: `OAuthProviderCard.tsx` (Google / GitHub / Slack) |

---

### 14. Project Settings

| Route: `/projects/[id]/settings` |
|---|
| **Page**: `ProjectSettingsPage.tsx` |
| **Components**: `ProjectForm.tsx`, `DangerZone.tsx`, `MemberList.tsx`, `InviteMemberDialog.tsx` |

| Route: `/projects/[id]/settings/members` |
|---|
| **Page**: `MembersPage.tsx` |
| **Components**: `MemberRow.tsx`, `RoleSelector.tsx`, `RemoveMemberDialog.tsx` |

---

## Shared / Global Components

| File Name | Status | Description | Used In |
|---|---|---|---|
| **shadcn/ui Components** (installed from registry) |
| `Button` | 📦 | Reusable button (default, destructive, outline, secondary, ghost, link) | All pages |
| `Input` | 📦 | Text input with validation states | All forms |
| `Textarea` | 📦 | Multi-line text input | Forms |
| `Label` | 📦 | Form label with peer-disabled | All forms |
| `Select` | 📦 | Dropdown select with search | All forms |
| `Card` | 📦 | Card with header, content, footer | Dashboard, Projects |
| `Dialog` | 📦 | Modal dialog | Create/Delete/Invite |
| `DropdownMenu` | 📦 | Context menu / user menu | TopNav, UserMenu |
| `Tabs` | 📦 | Tabbed navigation | Project Detail, Settings |
| `Skeleton` | 📦 | Loading placeholder | All pages |
| `Table` | 📦 | Sortable data table | Audit, Events, Members |
| `Form` | 📦 | Form with validation (react-hook-form + zod) | All forms |
| `Sidebar` | 📦 | Collapsible sidebar (shadcn) | Protected layout |
| `Sheet` | 📦 | Slide-over panel | Sidebar mobile, dialogs |
| `Separator` | 📦 | Visual divider | Layout, menus |
| `Tooltip` | 📦 | Hover tooltip | Actions, icons |
| `Sonner` | 📦 | Toast notifications | Global |
| **Custom Components (⬜ not yet built)** |
| `TopNav` | ⬜ | Top navigation bar | All protected pages |
| `ProjectSwitcher` | ⬜ | Dropdown to switch projects | `TopNav` |
| `UserMenu` | ⬜ | Avatar + dropdown (profile/settings/logout) | `TopNav` |
| `Avatar` | ⬜ | User avatar with fallback | TopNav, UserMenu, Members |
| `Badge` | ⬜ | Status badge (success, warning, error) | Multiple pages |
| `EmptyState` | ⬜ | Empty state with icon + message + action | All list pages |
| `DataTable` | ⬜ | Sortable/filterable table wrapper | Audit, Events, Members |
| `SqlEditor` | ⬜ | SQL code editor (CodeMirror/Monaco) | Migrations, Schema Diff |
| `LoadingSpinner` | ⬜ | Loading indicator | All pages |

---

## Pages Status

| Route | Page Component | Status |
|---|---|---|
| **Auth** |
| `/login` | `LoginPage` | ✅ Done |
| `/register` | `RegisterPage` | ✅ Done |
| `/forgot-password/*` | Forgot flow | ✅ Done |
| `/auth/callback/[provider]` | `OAuthCallbackPage` | ✅ Done |
| **Protected** |
| `/dashboard` | `DashboardPage` | ⬜ |
| `/projects` | `ProjectsPage` | ⬜ |
| `/projects/new` | `CreateProjectPage` | ⬜ |
| `/projects/[id]` | `ProjectDetailPage` | ⬜ |
| `/projects/[id]/connections` | `ConnectionsPage` | ⬜ |
| `/projects/[id]/connections/new` | `CreateConnectionPage` | ⬜ |
| `/projects/[id]/schemas` | `SchemasPage` | ⬜ |
| `/projects/[id]/schemas/[schemaId]` | `SchemaDetailPage` | ⬜ |
| `/projects/[id]/schemas/[schemaId]/erd` | `ErdPage` | ⬜ |
| `/projects/[id]/schemas/[schemaId]/compare` | `SchemaComparePage` | ⬜ |
| `/projects/[id]/migrations` | `MigrationsPage` | ⬜ |
| `/projects/[id]/migrations/new` | `CreateMigrationPage` | ⬜ |
| `/projects/[id]/migrations/[migrationId]` | `MigrationDetailPage` | ⬜ |
| `/projects/[id]/migrations/[migrationId]/run` | `MigrationRunPage` | ⬜ |
| `/projects/[id]/drift` | `DriftPage` | ⬜ |
| `/projects/[id]/drift/[driftId]` | `DriftDetailPage` | ⬜ |
| `/projects/[id]/audit` | `AuditPage` | ⬜ |
| `/projects/[id]/events` | `EventsPage` | ⬜ |
| `/projects/[id]/settings` | `ProjectSettingsPage` | ⬜ |
| `/projects/[id]/settings/members` | `MembersPage` | ⬜ |
| `/settings` | `SettingsPage` | ⬜ |
| `/settings/connections` | `LinkedAccountsPage` | ⬜ |

---

## Summary

| Category | Count | Status |
|---|---|---|
| Total Routes | ~24 | 6 ✅ / 18 ⬜ |
| Auth Pages | 6 | 6 ✅ **Done** |
| Protected Pages | ~18 | 0 ✅ / 18 ⬜ |
| shadcn/ui Components | 17 | 17 📦 **Installed** |
| Custom Components | 9 | 0 ✅ / 9 ⬜ |
| Child Components (domain) | ~65 | 0 ✅ / 65 ⬜ |
