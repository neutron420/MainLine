# SchemaHub Frontend Plan

## After Login

User gets redirected to → **`/dashboard`**

---

## All Pages & Components

---

### 1. Auth Pages (Done ✓)

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

| File Name | Description | Used In |
|---|---|---|
| `TopNav.tsx` | Top navigation bar | All protected pages |
| `Sidebar.tsx` | Side navigation (collapsible) | All protected pages |
| `ProjectSwitcher.tsx` | Dropdown to switch projects | `TopNav.tsx` |
| `UserMenu.tsx` | Avatar + dropdown (profile/settings/logout) | `TopNav.tsx` |
| `Button.tsx` | Reusable button (variants: primary, secondary, danger, ghost) | All pages |
| `Input.tsx` | Text input with label + error | All forms |
| `Badge.tsx` | Status badge (success, warning, error, info) | Multiple pages |
| `Card.tsx` | Generic card container | Dashboard, Projects, etc. |
| `Dialog.tsx` | Modal dialog wrapper | Create, Delete, Invite actions |
| `LoadingSpinner.tsx` | Loading indicator | All pages |
| `EmptyState.tsx` | Empty state with icon + message + action | All list pages |
| `ConfirmDialog.tsx` | Confirmation modal | Delete/rollback actions |
| `Toast.tsx` | Toast notification | Global |
| `DataTable.tsx` | Sortable/filterable table | Audit, Events, Members |
| `SqlEditor.tsx` | SQL code editor (CodeMirror/Monaco) | Migrations, Schema Diff |
| `IconButton.tsx` | Icon-only button | Toolbar actions |

---

## Summary

| Category | Count |
|---|---|
| Total Routes | ~24 |
| Auth Pages (done) | 6 |
| Protected Pages | ~18 |
| Page Components | ~24 |
| Unique Child Components | ~65 |
| Shared/Global Components | ~16 |
