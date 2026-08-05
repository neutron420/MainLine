# Frontend Performance Optimization & Housekeeping Plan

## Executive Summary

This plan addresses performance bottlenecks in the SchemaHub Next.js frontend (e.g., rendering delays, canvas framing drops in ERD/diff pages, unoptimized gRPC transport, and unnecessary re-renders) while cleaning up obsolete and unwanted files across the frontend repository.

---

## 1. Codebase Housekeeping & Unwanted File Cleanup

### Identified Unwanted / Redundant Files for Removal
1. `frontend/next-dev.err.log` — Development server error log file.
2. `frontend/next-dev.log` — Development server stdout log file.
3. `frontend/tsconfig.tsbuildinfo` — Stale TypeScript build cache committed in root.
4. `frontend/bun.lock` — Duplicate package lockfile (`npm` with `package-lock.json` is the standard lockfile).
5. `frontend/CLAUDE.md` — Redundant 11-byte stub file pointing to `AGENTS.md`.

### Git & Ignored Pattern Safeguards
- Update `frontend/.gitignore` to include `*.log`, `*.tsbuildinfo`, and `bun.lock`.

---

## 2. Architecture & Performance Optimization Pillars

```
+-----------------------------------------------------------------------------------+
|                           FRONTEND PERFORMANCE PILLARS                           |
+------------------------------------+----------------------------------------------+
| 1. API & ConnectRPC Layer          | 2. Rendering & React Flow Tuning             |
| - Request deduplication interceptor| - React Flow node memoization & windowing    |
| - Pre-fetching on route hover      | - Schema DDL & Column Table virtualization   |
| - Optimistic cache updates         | - useTransition for filtering/search state   |
+------------------------------------+----------------------------------------------+
| 3. Asset & Bundle Optimization     | 4. State Management Isolation                |
| - Dynamic imports for heavy canvas | - Sliced Zustand/Context subscribers         |
| - Package import optimization      | - Removal of blocking external scripts       |
| - Font loading with swap & preload | - UI-only state decoupled from server state  |
+------------------------------------+----------------------------------------------+
```

### Pillar 1: API & Network Communication (ConnectRPC / gRPC-Web)
- **Request Deduplication & Caching Interceptor**:
  - Implement a custom ConnectRPC transport interceptor to automatically deduplicate in-flight identical RPC queries.
- **TanStack React Query Cache Tuning**:
  - Configure route-level `staleTime` (60s for schemas/projects, 15s for drift/migrations).
  - Implement speculative prefetching on hover for primary navigation items (e.g., Project Detail, Schema Detail, ERD Canvas).
- **Optimistic UI Updates**:
  - Apply immediate optimistic updates to React Query cache on user mutations (e.g., connection status toggle, drift resolution, member role updates).

### Pillar 2: Rendering & Component Optimization
- **ERD Canvas Optimization (`@xyflow/react`)**:
  - Memoize custom node & edge components (`React.memo` with custom comparison).
  - Enable viewport rendering bounds (`onlyRenderVisibleElements={true}`).
  - Throttle canvas zoom and pan event callbacks.
- **Schema DDL & Large Table Virtualization**:
  - Use windowing/virtualization (`@tanstack/react-virtual`) for schemas with 50+ columns or extensive migration logs.
- **Non-Blocking UI State Updates**:
  - Wrap search filters, schema tree expansion, and drift table queries in `React.useTransition` to maintain 60 FPS responsiveness during user typing.

### Pillar 3: Bundle Size & Resource Loading
- **Dynamic Imports for Heavy Components**:
  - Lazy-load React Flow (`@xyflow/react`), SQL diff highlighters, and MDX renderers with `next/dynamic` and Skeleton fallback states.
- **Next.js Package Import Optimization**:
  - Configure `experimental.optimizePackageImports` in `next.config.ts` for `@heroui/react`, `lucide-react`, and `@radix-ui/*`.
- **Remove Blocking Third-Party Scripts**:
  - Remove external render-blocking script tag (`https://tweakcn.com/live-preview.min.js`) from root `layout.tsx`.
- **Font Preloading**:
  - Ensure `display: "swap"` and proper preloading for local DM Sans font variants.

### Pillar 4: State Management & Memory Hygiene
- **Granular Context Selectors**:
  - De-couple global Auth & Theme context updates from static UI layout wrappers.
- **Unsubscribe & Event Stream Cleanup**:
  - Ensure notification SSE/gRPC streams and window event listeners cleanly tear down on component unmount to prevent memory leaks.

---

## 3. Step-by-Step Execution Plan

| Phase | Milestone | Actions | Target Impact |
|---|---|---|---|
| **Phase 1** | **Housekeeping & Cleanup** | Delete unwanted files (`next-dev*.log`, `tsconfig.tsbuildinfo`, `bun.lock`, `CLAUDE.md`). Update `.gitignore`. | Clean repository state & reproducible builds |
| **Phase 2** | **Network & Transport** | Implement ConnectRPC interceptors, tune React Query `staleTime`, add prefetching hooks & optimistic updates. | 40-60% reduction in redundant backend requests |
| **Phase 3** | **Render & Component Optimization** | Optimize ERD canvas, memoize custom nodes, add `useTransition` to filter controls, lazy-load heavy components. | Eliminate input delay, 60 FPS canvas panning |
| **Phase 4** | **Bundle & Layout Polish** | Remove blocking external scripts, configure `optimizePackageImports`, refine local font loading strategy. | Faster FCP/LCP, lower bundle payload |

---

## 4. Verification & Metrics Checklist

- [ ] `npm run build` completes cleanly without TypeScript or ESLint warnings.
- [ ] No unhandled re-renders logged on input interactions (verified via React DevTools Profiler).
- [ ] ERD page loads with dynamic chunk splitting for `@xyflow/react`.
- [ ] In-flight gRPC queries deduplicated; hover pre-fetching reduces page navigation latency.
- [ ] Unwanted log/build files permanently removed from filesystem and `.gitignore` updated.
