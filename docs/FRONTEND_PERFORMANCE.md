# Frontend Performance Optimization

Status: **Implemented** — verification steps at the end.

## 1. Problem Statement

The app shell was duplicated in every authenticated page (sidebar, header,
notifications), the sidebar fired one gRPC query per project to render
member/connection counts, realtime streams stayed open even when hidden, and
the gRPC transport had no timeout, no reconnect backoff, and stale query keys.

Symptoms:
- Slow initial navigation (22 duplicated shells, redundant requests).
- N+1 member/connection queries on the sidebar.
- Unbounded payloads from `ListAuditEntries` (no page size).
- React Query cache collisions (`useDriftEvent` key `["drift", id]` collided
  with the list key `["drift", connectionId]`; date/driftType filters were
  missing from keys, so filter changes served stale cache data).
- One-off @heroui dependency for a single Tooltip.
- Dead files (template SVGs, duplicate `use-mobile`, unused `globals.css`).

## 2. Changes

### 2.1 Shared route-group layout (`(app)/layout.tsx`)

- Moved `dashboard`, `projects`, `schemas`, `settings`, `team` into
  `src/app/(app)/`; single `layout.tsx` renders `SidebarProvider`,
  `AppSidebar`, and the header (sidebar trigger, section title derived from
  `usePathname`, `NotificationsPopover`) **once**.
- Stripped the duplicated shell (sidebar/header/title/popover code) from all
  22 pages; pages now return bare content. Removed `dashboard/layout.tsx`.
- Result: one shell, one client-side mount, ~22x less duplicate code, no
  per-page client JS for the shell.

### 2.2 AppSidebar N+1 removal

- Before: `useQueries` fired `GetProjectMembers` + `GetConnections` per
  project listed in the sidebar.
- After: a single `useConnections(firstProjectId)` + `useMembers(firstProjectId)`
  feed the count badges; active item derived from `usePathname()` (no
  `useState` mirroring); activity items memoized with `useMemo`.

### 2.3 Realtime streams

- `NotificationsPopover` lazy-mounts the event stream: `useEventStream` is
  `enabled` only while the popover is open; events clear on close.
- `useEventStream` (`use-realtime.ts`):
  - new `enabled` option;
  - exponential backoff with jitter on reconnect (`reconnectDelayMs * 2^attempt`,
    capped at 30s, ±25% jitter), attempt counter reset on each received event;
  - fixed the glued `EventStreamOptions` type declaration.

### 2.4 Query keys & bounded payloads

- `useAuditEntries`: `pageSize` (default 50) now sent to the API and included
  in the query key; `dateFrom`/`dateTo` added to the key.
- `useDriftEvents`: `driftType` added to the query key.
- `useDriftEvent` / `useResolveDriftEvent`: key namespaced to
  `["drift", "event", id]` so it can never collide with the list key.

### 2.5 Transport

- `transport.ts`: `defaultTimeoutMs: 60_000` on the gRPC-Web transport.

### 2.6 Dependency & asset cleanup

- Replaced the single `@heroui/react` Tooltip in `team/page.tsx` with
  `ui/tooltip` (shadcn/radix) and removed the dependency (44 packages) plus
  its `optimizePackageImports` entry in `next.config.ts`.
- `privacy/page.tsx` converted from client to server component (no hooks used).
  This surfaced a Next 15 + MDX issue: compiled MDX imports
  `useMDXComponents` from `@mdx-js/react`, which calls `createContext` at
  module scope and crashed as an RSC module (`createContext only works in
  Client Components`). Fixed with the documented `src/mdx-components.tsx`
  import-source override. Note: Turbopack dev needs a container restart to
  pick up import-source changes.
- Deleted: `src/app/globals.css` (unused duplicate of `src/styles/globals.css`),
  `src/lib/hooks/use-mobile.ts` (dead duplicate of `src/hooks/use-mobile.ts`),
  default template SVGs and unused images in `public/` (`next.svg`,
  `vercel.svg`, `window.svg`, `file.svg`, `globe.svg`, `footer.svg`,
  `slack.png`, `logo.png`), session scratch scripts under `frontend/scripts/`,
  stale `frontend/tsconfig.tsbuildinfo`, and scratch docs
  (`FRONTEND_PLAN.md`, `PERFORMANCE_OPTIMIZATION_PLAN.md`).
- Fixed `components.json` css path (`src/styles/globals.css`).

## 3. Not Done (future candidates)

- ERD canvas tuning: `onlyRenderVisibleElements`, node memoization, dynamic
  chunk loading of `@xyflow/react` — deferred; verify with the React Profiler
  first.
- Column-table virtualization for schemas with 50+ columns
  (`@tanstack/react-virtual`).
- Route-level `staleTime` tuning and hover prefetching.

## 4. Verification

```bash
cd frontend
npx tsc --noEmit          # clean
npm run lint              # clean
npm run build             # no warnings
# then load-check in the running stack (docker compose):
# /login, /signup, /dashboard, /projects, project detail,
# /projects/[id]/schemas/[schemaId]/erd, /settings, /team, /audit, /migrations
# Network tab: sidebar issues at most 2 list queries (no per-project N+1);
# Notifications popover: stream connects only after opening.
```

Verified live (dev stack): all routes return 200, including `/privacy` after
the RSC conversion. `/audit`, `/migrations`, `/projects/[id]/schemas` return
404 by design — those routes live under `/projects/[id]/...`.
