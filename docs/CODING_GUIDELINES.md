# Coding Guidelines

> **Standardized coding conventions, naming patterns, package organization, and engineering practices for SchemaHub.**

---

## Table of Contents

- [General Principles](#general-principles)
- [Folder Conventions](#folder-conventions)
- [Go Style Guide](#go-style-guide)
- [TypeScript / React Style Guide](#typescript--react-style-guide)
- [Package Organization](#package-organization)
- [Error Handling](#error-handling)
- [Logging Standards](#logging-standards)
- [Commenting Standards](#commenting-standards)
- [Documentation Standards](#documentation-standards)
- [Commit Message Conventions](#commit-message-conventions)
- [Git Branching Strategy](#git-branching-strategy)

---

## General Principles

1. **Production-grade from day one** — Every line of code is written as if it will run in production tomorrow
2. **Make it correct, then make it fast** — Correctness over optimization
3. **Explicit over implicit** — Code should be obvious, not clever
4. **Consistency over preference** — Follow established patterns, not personal style
5. **Document the why** — Comments explain reasoning, not what the code does
6. **Test the behavior, not the implementation** — Tests should break when behavior changes, not when code is refactored
7. **Zero warnings** — All linters and type checkers pass with zero warnings at all times

---

## Folder Conventions

### Go

```
pkg/                            # Public packages (imported by external projects)
internal/                       # Private packages (not importable externally)
  └── {domain}/                 # One folder per bounded context
      ├── domain/               # Business logic, entities, services
      ├── repository/           # Data access interfaces and implementations
      │   └── postgres/         # PostgreSQL implementations
      └── handler/              # gRPC or transport handlers
```

### TypeScript

```
src/
├── app/                        # Next.js App Router pages and layouts
├── components/                 # Reusable React components
│   ├── ui/                     # shadcn/ui primitives
│   ├── {domain}/               # Domain-specific components
│   └── shared/                 # Cross-domain shared components
├── lib/                        # Utilities, hooks, client code
│   ├── api/                    # API client functions
│   ├── hooks/                  # Custom React hooks
│   ├── providers/              # Context providers
│   ├── utils/                  # Utility functions
│   └── types/                  # TypeScript type definitions
└── styles/                     # Global styles
```

---

## Go Style Guide

### Code Formatting

- All Go code must pass `go fmt` — this is non-negotiable
- Run `go vet` on all code — no exceptions
- Use `golangci-lint` with the project's `.golangci.yml` configuration

### Naming Conventions

| Element | Convention | Example |
|---|---|---|
| Variables | camelCase | `userID`, `migrationRun` |
| Exported functions | PascalCase | `CreateProject`, `ValidateMigration` |
| Unexported functions | camelCase | `validateSQL`, `computeChecksum` |
| Interfaces | PascalCase with `er` suffix | `UserRepository`, `SchemaService` |
| Structs | PascalCase | `Project`, `MigrationRun` |
| Constants | PascalCase | `MaxRetryCount`, `DefaultPageSize` |
| Files | snake_case | `user_repository.go`, `migration_handler.go` |
| Packages | lowercase, single word | `auth`, `schema`, `migration` |

### Package Organization

- **One package per directory** — No `package foo` in multiple directories
- **Small interfaces** — 1-3 methods per interface. Prefer many small interfaces over few large ones.
- **Accept interfaces, return structs** — Functions accept interfaces and return concrete types
- **No package-level variables** — Use dependency injection instead
- **Naming** — Package names should be short (1-3 lowercase words)

### Imports

```go
import (
    // Standard library
    "context"
    "fmt"

    // Third-party
    "github.com/jackc/pgx/v5"
    "google.golang.org/grpc"

    // Internal
    "github.com/schemahub/backend/internal/auth/domain"
    "github.com/schemahub/backend/internal/pkg/config"
)
```

### Function Style

```go
// Good: Named return values for documentation, defer for cleanup
func (s *Service) CreateProject(ctx context.Context, name string, ownerID string) (project *Project, err error) {
    // Validate
    if name == "" {
        return nil, fmt.Errorf("project name is required")
    }

    // Business logic
    project = &Project{
        ID:        uuid.NewString(),
        Name:      name,
        OwnerID:   ownerID,
        CreatedAt: time.Now(),
    }

    // Persist
    if err := s.repo.Create(ctx, project); err != nil {
        return nil, fmt.Errorf("creating project: %w", err)
    }

    return project, nil
}
```

### Context

- `context.Context` is always the first parameter in functions that make database calls or service calls
- Never store context in a struct — pass it explicitly
- Use `context.WithTimeout` and `context.WithCancel` for operations with deadlines

---

## TypeScript / React Style Guide

### Configuration

- `strict: true` in `tsconfig.json`
- `noUncheckedIndexedAccess: true`
- `noImplicitReturns: true`

### Naming Conventions

| Element | Convention | Example |
|---|---|---|
| Components | PascalCase | `SchemaExplorer`, `MigrationRunner` |
| Functions | camelCase | `useAuth`, `formatDate` |
| Variables | camelCase | `migrationStatus`, `projectList` |
| Types/Interfaces | PascalCase with `Type`/`I` prefix | `ProjectType`, `IMigration` (TS conventions vary; prefer `Type` suffix) |
| Enums | PascalCase | `MigrationStatus`, `EventType` |
| Files (components) | kebab-case | `schema-explorer.tsx` |
| Files (utilities) | camelCase | `formatDate.ts` |
| Constants | UPPER_SNAKE_CASE | `MAX_RETRY_COUNT` |

### Component Structure

```tsx
// Good: Explicit types, no `any`, props interface defined
interface SchemaExplorerProps {
    projectId: string;
    schemaName: string;
}

export function SchemaExplorer({ projectId, schemaName }: SchemaExplorerProps) {
    // Hooks at the top
    const { data: schema, isLoading } = useSchema(projectId, schemaName);

    // Early returns for loading/error states
    if (isLoading) return <LoadingSpinner />;
    if (!schema) return <ErrorState message="Schema not found" />;

    // Render
    return (
        <div>
            {/* ... */}
        </div>
    );
}
```

### Server vs Client Components

- **Default to Server Components** — Only use `'use client'` when interactivity is needed
- **Client Components** — Event handlers, state, effects, browser APIs
- **Server Components** — Data fetching, static rendering, SEO

### State Management

- **Server state** — TanStack Query (all data from API)
- **URL state** — Next.js search params, path parameters
- **Form state** — React Hook Form
- **UI state** — `useState` / `useReducer` (local to component)
- **No global state store** — TanStack Query replaces Redux/Zustand for SchemaHub

### TanStack Query Patterns

```tsx
// Query
export function useProjects() {
    return useQuery({
        queryKey: ['projects'],
        queryFn: () => projectClient.listProjects({}),
    });
}

// Mutation
export function useCreateProject() {
    const queryClient = useQueryClient();
    return useMutation({
        mutationFn: (data: CreateProjectData) => projectClient.createProject(data),
        onSuccess: () => {
            queryClient.invalidateQueries({ queryKey: ['projects'] });
        },
    });
}
```

---

## Error Handling

### Go Error Handling

| Rule | Example |
|---|---|
| Always check errors | `if err != nil { return ... }` |
| Wrap errors with context | `fmt.Errorf("fetching user %s: %w", id, err)` |
| Never use `_` to ignore errors | Always handle or explicitly ignore with comment |
| Define sentinel errors for domain | `var ErrNotFound = errors.New("resource not found")` |
| Use typed errors for recoverable | Custom error types with metadata |
| Map domain errors to gRPC codes | `domain.ErrNotFound → gRPC NotFound` |

### TypeScript Error Handling

```tsx
try {
    const result = await api.call();
} catch (error) {
    if (error instanceof ApiError) {
        switch (error.code) {
            case 'NOT_FOUND':
                // Handle not found
                break;
            case 'UNAUTHENTICATED':
                // Handle auth failure (trigger refresh)
                break;
        }
    }
}
```

---

## Logging Standards

### Go (slog)

```go
// INFO level for normal operations
slog.InfoContext(ctx, "migration completed",
    "migration_id", migrationID,
    "duration_ms", duration,
    "status", "completed",
)

// ERROR level for failures
slog.ErrorContext(ctx, "migration failed",
    "migration_id", migrationID,
    "error", err,
    "duration_ms", duration,
)
```

### Log Levels

| Level | Usage |
|---|---|
| `DEBUG` | Detailed debugging, enabled only in development |
| `INFO` | Normal operations, state changes, request completion |
| `WARN` | Unexpected but handled, slow queries, approaching limits |
| `ERROR` | Operation failures, database errors, unexpected conditions |

### Structured Fields

Every log entry should include:
- `trace_id` — Correlation ID from request context
- `service` — Service name
- `user_id` — Authenticated user (if applicable)
- Operation-specific fields (duration, status, resource ID)

---

## Commenting Standards

### Go Comments

```go
// Package domain contains the core business logic for project management.
// It is independent of any transport, database, or external concern.
package domain

// Project represents a SchemaHub project.
// Projects are the top-level organizational unit that contain
// database connections, schemas, and migrations.
type Project struct {
    ID   string
    Name string
    // ...
}

// CreateProject creates a new project with the given name.
// Returns ErrProjectAlreadyExists if a project with the same slug exists.
func (s *Service) CreateProject(ctx context.Context, name string, ownerID string) (*Project, error) {
    // ...
}
```

### What to Comment

- **Packages** — What the package does, who uses it
- **Exported types** — What they represent
- **Exported functions** — What they do, parameters, return values, errors
- **Complex logic** — Why a particular approach was taken (not what the code does)
- **Workarounds** — Explanation of why a non-obvious approach was necessary

### What NOT to Comment

- Obvious code (comments that just repeat the code)
- TODO comments with no associated issue number

---

## Documentation Standards

### When to Document

- Every new feature must have corresponding documentation before code
- Every API change must update the relevant documentation
- Every architectural decision must include trade-off analysis

### Documentation Structure

- All documentation is in the `docs/` directory
- Cross-reference related documents
- Use Mermaid diagrams for architecture and flows
- Use tables for structured information
- Include table of contents for documents longer than 3 sections

---

## Commit Message Conventions

### Format

```
<type>: <short summary> (max 72 characters)

<body> (optional, wrap at 72 characters)

<footer> (optional)
```

### Types

| Type | Usage |
|---|---|
| `feat` | New feature |
| `fix` | Bug fix |
| `docs` | Documentation changes |
| `refactor` | Code refactoring (no behavior change) |
| `test` | Adding or modifying tests |
| `chore` | Build, CI, dependencies, tooling |
| `perf` | Performance improvement |
| `style` | Formatting, linting (no behavior change) |

### Examples

```
feat: add schema introspection endpoint

Implement IntrospectSchema RPC that connects to a PostgreSQL
database and reads full schema metadata from system catalogs.

Closes #42
```

```
fix: handle empty schema name in introspection

Empty schema names caused a panic in the introspection parser.
Added validation to reject empty schema names with a clear error.

Fixes #87
```

---

## Git Branching Strategy

### Branch Naming

- `main` — Production-ready code, protected
- `feature/<name>` — New features (e.g., `feature/schema-introspection`)
- `fix/<name>` — Bug fixes (e.g., `fix/empty-schema-panic`)
- `docs/<name>` — Documentation changes
- `refactor/<name>` — Code refactoring
- `chore/<name>` — Tooling, CI, dependencies

### Workflow

```
main ──────┬───────────────────────────────────
            \                                  /
feature/     └── feature/schema-explorer ──────
                                              \
fix/                                          └── fix/empty-name ────
```

### Rules

1. Branch from `main`
2. Use squash merge with linear history
3. PR title must match commit convention (e.g., `feat: add schema introspection`)
4. All PRs require:
   - At least one approval
   - All CI checks passing
   - Zero linter warnings
   - Tests passing
5. No direct commits to `main` (protected branch)
