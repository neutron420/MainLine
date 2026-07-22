# Testing Strategy

> **Complete testing strategy for SchemaHub — unit tests, integration tests, gRPC tests, database tests, performance tests, and end-to-end tests.**

---

## Table of Contents

- [Testing Philosophy](#testing-philosophy)
- [Test Pyramid](#test-pyramid)
- [Unit Testing](#unit-testing)
- [Integration Testing](#integration-testing)
- [gRPC API Testing](#grpc-api-testing)
- [Database Testing](#database-testing)
- [Frontend Testing](#frontend-testing)
- [Performance Testing](#performance-testing)
- [Load Testing](#load-testing)
- [End-to-End Testing](#end-to-end-testing)
- [Test Infrastructure](#test-infrastructure)

---

## Testing Philosophy

1. **Test behavior, not implementation** — Tests should break when requirements change, not when code is refactored
2. **Write tests alongside code** — TDD preferred for domain logic
3. **Every test must be deterministic** — No flaky tests, no time-dependent assertions
4. **Fast feedback** — Unit tests in milliseconds, integration tests in seconds
5. **Realistic environments** — Test against real PostgreSQL (not mocks) for integration tests

---

## Test Pyramid

```
         ╱╲
        ╱  ╲
       ╱ E2E╲           Few end-to-end tests
      ╱──────╲
     ╱ Inte-  ╲        More integration tests
    ╱ gration  ╲
   ╱────────────╲
  ╱   Unit       ╲     Most unit tests
 ╱     Tests      ╲
╱──────────────────╲
```

| Layer | Count Goal | Speed | Responsibility |
|---|---|---|---|
| **Unit** | 70%+ | Milliseconds | Domain logic, validation, errors |
| **Integration** | 20% | Seconds | Repository, gRPC handlers |
| **E2E** | 10% | Minutes | Full user workflows |

---

## Unit Testing

### Scope

- Domain logic (services, validators, calculators)
- Pure functions (diff engine, checksum computation)
- Error paths and edge cases

### Go Unit Tests

```go
func TestCreateProject(t *testing.T) {
    t.Parallel()

    tests := []struct {
        name    string
        nameArg string
        wantErr error
    }{
        {name: "valid name", nameArg: "My Service", wantErr: nil},
        {name: "empty name", nameArg: "", wantErr: ErrProjectNameRequired},
        {name: "name too long", nameArg: strings.Repeat("a", 201), wantErr: ErrProjectNameTooLong},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            mockRepo := new(MockProjectRepository)
            svc := NewService(mockRepo)

            _, err := svc.CreateProject(context.Background(), tt.nameArg, "user_123")
            if !errors.Is(err, tt.wantErr) {
                t.Errorf("got %v, want %v", err, tt.wantErr)
            }
        })
    }
}
```

### TypeScript Unit Tests

```tsx
import { render, screen } from '@testing-library/react';
import { SchemaTree } from './schema-tree';

describe('SchemaTree', () => {
    it('renders table names', () => {
        render(<SchemaTree tables={mockTables} />);
        expect(screen.getByText('users')).toBeInTheDocument();
    });

    it('shows empty state when no tables', () => {
        render(<SchemaTree tables={[]} />);
        expect(screen.getByText('No tables found')).toBeInTheDocument();
    });
});
```

---

## Integration Testing

### Scope

- Repository implementations against real PostgreSQL
- gRPC handler integration with services
- Redis integration (pub/sub, caching)
- End-to-end service flows

### Test Database Strategy

- **Ephemeral databases** — Each test run creates a fresh database
- **Testcontainers** — PostgreSQL and Redis containers spun up per test suite
- **Migration at start** — Run migrations before each test suite
- **Clean state** — Truncate all tables between test cases

### Go Integration Test Structure

```go
type Suite struct {
    db     *pgxpool.Pool
    redis  *redis.Client
    repo   *postgres.Repository
}

func TestMain(m *testing.M) {
    // Start testcontainers
    postgresContainer, _ := setupPostgres()
    redisContainer, _ := setupRedis()

    // Run migrations
    runMigrations(postgresContainer)

    // Run tests
    code := m.Run()

    // Cleanup
    postgresContainer.Terminate()
    redisContainer.Terminate()
    os.Exit(code)
}

func TestSchemaRepository(t *testing.T) {
    suite := setupSuite(t)
    defer suite.teardown()

    t.Run("create and retrieve schema", func(t *testing.T) {
        schema := &domain.Schema{...}
        err := suite.repo.Create(ctx, schema)
        assert.NoError(t, err)

        got, err := suite.repo.GetByID(ctx, schema.ID)
        assert.NoError(t, err)
        assert.Equal(t, schema.Name, got.Name)
    })
}
```

---

## gRPC API Testing

### Approach

1. **Unit test handlers** with mocked services
2. **Integration test** full gRPC call chain (handler → service → repository → database)
3. **Contract testing** — Verify proto definitions match handler implementations

### Testing gRPC Handlers

```go
func TestCreateProjectHandler(t *testing.T) {
    // Setup
    svc := new(MockProjectService)
    handler := NewGRPCHandler(svc)

    svc.On("CreateProject", mock.Anything, mock.Anything).
        Return(&domain.Project{ID: "proj_123", Name: "Test"}, nil)

    // Execute
    req := &pb.CreateProjectRequest{Name: "Test Project"}
    resp, err := handler.CreateProject(ctx, req)

    // Assert
    assert.NoError(t, err)
    assert.Equal(t, "proj_123", resp.Project.Id)
    svc.AssertExpectations(t)
}
```

### gRPC Client Testing

```go
func TestProjectServiceE2E(t *testing.T) {
    // Start real gRPC server with test dependencies
    server := startTestServer(t)
    defer server.Stop()

    // Create gRPC client
    conn, err := grpc.Dial(server.Addr(), grpc.WithInsecure())
    require.NoError(t, err)
    defer conn.Close()

    client := pb.NewProjectServiceClient(conn)

    // Test
    resp, err := client.CreateProject(ctx, &pb.CreateProjectRequest{
        Name: "Integration Test Project",
    })
    assert.NoError(t, err)
    assert.NotEmpty(t, resp.Project.Id)
}
```

---

## Database Testing

### Repository Tests

- Test each repository method against a real PostgreSQL instance
- Test edge cases: duplicate entries, foreign key violations, null handling
- Test transaction behavior: rollback on error, commit on success

### Migration Tests

- Test migration execution against a test database
- Verify schema changes are applied correctly
- Test rollback produces expected state
- Test that migrations are idempotent (run twice, same result)

### SQL Injection Tests

```go
func TestSQLInjectionPrevention(t *testing.T) {
    // Attempt SQL injection via repository methods
    maliciousName := "'; DROP TABLE users; --"
    _, err := repo.CreateProject(ctx, maliciousName)
    assert.NoError(t, err) // Should NOT drop the table
}
```

---

## Frontend Testing

### Component Tests

- Test rendering of all component states (loading, empty, error, populated)
- Test user interactions (clicks, form submission)
- Test accessibility (keyboard navigation, screen reader labels)

### Hook Tests

```tsx
describe('useSchema', () => {
    it('returns schema data on success', async () => {
        const { result } = renderHook(() => useSchema('proj_1', 'public'));

        await waitFor(() => expect(result.current.isSuccess).toBe(true));
        expect(result.current.data?.schemaName).toBe('public');
    });

    it('returns error on failure', async () => {
        server.use(
            http.get('*/schema/*', (req, res, ctx) =>
                res(ctx.status(500))
            )
        );

        const { result } = renderHook(() => useSchema('proj_1', 'public'));
        await waitFor(() => expect(result.current.isError).toBe(true));
    });
});
```

### Mocking Strategy

| Tool | Use |
|---|---|
| **MSW (Mock Service Worker)** | Mock gRPC-web responses at the network level |
| **Vitest** | Test runner and mocking framework |
| **Testing Library** | Component rendering and interaction |
| **Storybook** | Visual component testing and documentation |

---

## Performance Testing

### Test Scenarios

| Scenario | Target | Tool |
|---|---|---|
| Schema introspection (100 tables) | < 5 seconds | Go benchmark |
| Schema diff (500 objects) | < 2 seconds | Go benchmark |
| Migration execution (20 statements) | < 30 seconds | Timer test |
| Audit log query (1M entries) | < 500ms P99 | Database benchmark |
| Event delivery | < 100ms P99 | Custom measurement |

### Go Benchmarks

```go
func BenchmarkIntrospectLargeSchema(b *testing.B) {
    db := setupLargeTestDB(b) // 100 tables, 1000 columns
    svc := NewIntrospectionService(db)

    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        _, err := svc.Introspect(ctx, "public")
        if err != nil {
            b.Fatal(err)
        }
    }
}
```

---

## Load Testing

### k6 Test Scripts

```javascript
import http from 'k6/http';
import { check, sleep } from 'k6';

export const options = {
    stages: [
        { duration: '2m', target: 50 },  // Ramp up
        { duration: '5m', target: 50 },  // Stay
        { duration: '2m', target: 0 },   // Ramp down
    ],
    thresholds: {
        http_req_duration: ['p(99)<500'],  // 99% under 500ms
        http_req_failed: ['rate<0.01'],     // <1% errors
    },
};

export default function () {
    const res = http.get('https://api.schemahub.dev/v1/projects');
    check(res, {
        'status is 200': (r) => r.status === 200,
        'response time < 200ms': (r) => r.timings.duration < 200,
    });
    sleep(1);
}
```

### Load Test Scenarios

| Scenario | Concurrent Users | Duration | Success Criteria |
|---|---|---|---|
| Schema browsing | 100 | 10 min | P99 < 500ms |
| Migration execution | 20 | 10 min | P99 < 30s |
| Real-time subscriptions | 500 | 10 min | P99 delivery < 200ms |
| Audit log queries | 50 | 10 min | P99 < 1s |

---

## End-to-End Testing

### Cypress / Playwright Tests

```typescript
describe('Migration Workflow', () => {
    beforeEach(() => {
        cy.login('test@example.com', 'password123');
    });

    it('creates and executes a migration', () => {
        // Create project
        cy.visit('/projects/new');
        cy.get('[data-testid="project-name"]').type('Test Project');
        cy.get('[data-testid="create-project"]').click();

        // Add connection
        cy.get('[data-testid="add-connection"]').click();
        cy.get('[data-testid="connection-name"]').type('Test DB');
        cy.get('[data-testid="connection-host"]').type('localhost');
        // ... fill connection form
        cy.get('[data-testid="save-connection"]').click();

        // Create migration
        cy.get('[data-testid="new-migration"]').click();
        cy.get('[data-testid="migration-sql"]').type('ALTER TABLE users ADD COLUMN phone VARCHAR(20);');
        cy.get('[data-testid="save-migration"]').click();

        // Execute migration
        cy.get('[data-testid="execute-migration"]').click();
        cy.contains('Migration completed successfully');
    });
});
```

### E2E Test Scenarios

| Scenario | User Flow |
|---|---|
| **Complete user journey** | Register → Login → Create Project → Add Connection → Introspect → View Schema |
| **Migration lifecycle** | Create Migration → Validate → Execute → Verify → Rollback |
| **Schema versioning** | Introspect → View Version → Make Change → Introspect → Compare Versions |
| **Real-time updates** | Open project → Make change in other tab → See notification |
| **Team collaboration** | Create project → Invite member → Member views project → Member creates migration |
| **Error handling** | Submit invalid SQL → See validation error → Fix → Execute |
| **Authentication edge cases** | Expired token → Auto-refresh → Continue work → Logout → Cannot access |

---

## Test Infrastructure

### CI Pipeline

```
PR Created
  → Lint (golangci-lint, eslint, prettier)
  → Type check (tsc)
  → Unit tests (go test, vitest)
  → Integration tests (go test with testcontainers)
  → Build (go build, npm build)
  → Security scan (dependencies)
  → E2E tests (if full pipeline)
  → Deploy preview
```

### Test Coverage Targets

| Module | Minimum Coverage |
|---|---|
| Domain logic (Go) | 90% |
| Repository layer (Go) | 80% |
| gRPC handlers (Go) | 70% |
| Frontend components (TS) | 70% |
| Frontend hooks (TS) | 80% |
| **Overall** | **80%** |

### Tools

| Language | Testing Tool | Assertion Library | Mocking |
|---|---|---|---|
| **Go** | `testing` (stdlib) | `testify/assert` | `testify/mock` |
| **Go (integration)** | `testcontainers-go` | `testify/require` | N/A |
| **TypeScript** | Vitest | `@testing-library/jest-dom` | `vi.fn()` |
| **E2E** | Playwright | Built-in | MSW |
| **Load** | k6 | Built-in | N/A |
| **Benchmark** | Go `testing.B` | N/A | N/A |
