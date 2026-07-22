# Contributing

> **Guidelines for contributing to SchemaHub — how to get started, development workflow, code review, and community standards.**

---

## Table of Contents

- [Welcome](#welcome)
- [Getting Started](#getting-started)
- [Development Setup](#development-setup)
- [Development Workflow](#development-workflow)
- [Code Review](#code-review)
- [Testing](#testing)
- [Documentation](#documentation)
- [Commit Guidelines](#commit-guidelines)
- [Issue Reporting](#issue-reporting)
- [Feature Requests](#feature-requests)
- [Community Guidelines](#community-guidelines)

---

## Welcome

Thank you for considering contributing to SchemaHub! We believe in the power of community-driven development and welcome contributions of all kinds:

- **Code** — Bug fixes, features, performance improvements
- **Documentation** — Guides, tutorials, API documentation
- **Testing** — Test cases, edge cases, performance benchmarks
- **Design** — UI/UX improvements, accessibility
- **Feedback** — Feature requests, bug reports, usability feedback

This document outlines the contribution process. By participating, you agree to abide by our [Code of Conduct](#community-guidelines).

---

## Getting Started

### Prerequisites

- **Go** 1.22+ (backend)
- **Node.js** 20+ (frontend)
- **Docker** and **Docker Compose** (local development)
- **Buf** CLI (protobuf code generation)
- **Make** (build orchestration)

### Fork and Clone

```bash
# Fork the repository on GitHub
# Then clone your fork:
git clone https://github.com/your-username/schemahub.git
cd schemahub

# Add upstream remote
git remote add upstream https://github.com/schemahub/schemahub.git
```

### Verify Setup

```bash
# Run the full development environment
docker compose -f docker/docker-compose.yml -f docker/docker-compose.dev.yml up

# Run tests
cd backend && go test ./...
cd frontend && npm run test

# Verify linting
cd backend && golangci-lint run ./...
cd frontend && npm run lint
```

---

## Development Setup

### Local Backend

```bash
cd backend

# Install dependencies
go mod download

# Install development tools
go install github.com/air-verse/air@latest
go install github.com/bufbuild/buf/cmd/buf@latest

# Run with hot reload
air

# Run tests
go test -race -count=1 ./...
```

### Local Frontend

```bash
cd frontend

# Install dependencies
npm install

# Start development server
npm run dev

# Run tests
npm run test

# Run type check
npm run typecheck
```

### Environment Variables

Copy the example environment file:

```bash
cp backend/.env.example backend/.env
cp frontend/.env.local.example frontend/.env.local
```

See [Deployment](DEPLOYMENT.md) for all environment variable definitions.

---

## Development Workflow

### Branch Naming

- `feature/<name>` — New features
- `fix/<name>` — Bug fixes
- `docs/<name>` — Documentation changes
- `refactor/<name>` — Code refactoring
- `chore/<name>` — Tooling, CI, dependencies
- `test/<name>` — Test additions or modifications

### Workflow Steps

1. **Pick an issue** — Comment on the issue to express interest
2. **Create a branch** — Branch from `main` using the naming convention
3. **Write code** — Follow [Coding Guidelines](CODING_GUIDELINES.md)
4. **Write tests** — All new code must have tests
5. **Run checks** — Lint, type check, test all pass
6. **Commit** — Use conventional commit messages
7. **Push and PR** — Push branch, create pull request
8. **Address feedback** — Respond to review comments
9. **Merge** — Maintainer approves and merges

---

## Code Review

### Review Process

1. All PRs require at least one approval from a maintainer
2. PRs should be focused on a single concern (small PRs preferred)
3. Reviewers should provide constructive, specific feedback
4. Authors should respond to all comments

### What Reviewers Look For

| Category | Questions |
|---|---|
| **Correctness** | Does the code do what it claims? Are edge cases handled? |
| **Security** | Are there any security concerns? Is input validated? |
| **Performance** | Are there any obvious performance issues? |
| **Testing** | Are there tests for the new code? Do existing tests pass? |
| **Documentation** | Are relevant docs updated? Are APIs documented? |
| **Style** | Does the code follow project conventions? Does linter pass? |

### PR Title Format

PR titles must follow conventional commits:

```
feat: add schema introspection endpoint
fix: handle null pointer in diff engine
docs: update authentication documentation
refactor: extract migration executor interface
```

---

## Testing

### Mandatory Testing

- All new code must have corresponding tests
- Bug fixes must include a test that reproduces the bug
- Changes to domain logic require unit tests
- Changes to gRPC handlers require integration tests

### Running Tests

```bash
# Backend
cd backend
go test -race -count=1 ./...          # All tests
go test -race -run TestSchema ./...   # Specific tests
go test -bench=. ./...                # Benchmarks

# Frontend
cd frontend
npm run test                          # All tests
npm run test -- --coverage            # With coverage
npm run test -- --watch               # Watch mode
```

### Test Coverage

- Domain logic: minimum 90%
- Repository layer: minimum 80%
- Overall project: minimum 80%

---

## Documentation

### When to Update Docs

- **Feature changes** — Update relevant feature documentation
- **API changes** — Update protobuf contracts and API flow docs
- **Architecture changes** — Update architecture diagrams and descriptions
- **Configuration changes** — Update deployment and environment docs

### Documentation Standards

- All documentation is in the `docs/` directory
- Use Mermaid diagrams for architecture and flows
- Use tables for structured information
- Cross-reference related documents
- Follow the formatting of existing documents

---

## Commit Guidelines

### Format

```
<type>(<optional scope>): <description>

[optional body]

[optional footer]
```

### Types

| Type | Usage |
|---|---|
| `feat` | New feature |
| `fix` | Bug fix |
| `docs` | Documentation |
| `refactor` | Code refactoring |
| `test` | Tests |
| `chore` | Build, CI, dependencies |
| `perf` | Performance improvement |
| `style` | Formatting, linting |

### Examples

```
feat(schema): add introspection by schema name filter

Allow users to introspect specific schemas instead of all schemas
in a database. This improves performance for databases with many schemas.

Closes #123
```

```
fix(migration): handle empty down SQL gracefully

Rollback button was disabled when down SQL was empty.
Now shows a clear message explaining that down SQL is required for rollback.

Fixes #456
```

---

## Issue Reporting

### Bug Reports

When reporting a bug, include:

1. **Summary** — Clear, concise description
2. **Steps to reproduce** — Exact steps to trigger the bug
3. **Expected behavior** — What should happen
4. **Actual behavior** — What actually happens
5. **Environment** — Browser, OS, SchemaHub version, database type
6. **Logs/screenshots** — Relevant error messages, stack traces, screenshots
7. **Severity** — Critical (production impact), Major (blocked workflow), Minor (annoyance)

### Issue Template

```markdown
**Describe the bug:**
A clear and concise description of the bug.

**To Reproduce:**
1. Go to '...'
2. Click on '...'
3. See error

**Expected behavior:**
What should happen.

**Actual behavior:**
What actually happens.

**Environment:**
- SchemaHub version: [e.g., 1.0.0]
- Browser: [e.g., Chrome 120]
- Database: [e.g., PostgreSQL 16 on Neon]

**Additional context:**
Logs, screenshots, anything else relevant.
```

---

## Feature Requests

### How to Request

1. **Search existing issues** — Check if the feature has already been requested
2. **Create a new issue** — Use the feature request template
3. **Describe the problem** — What problem does this feature solve?
4. **Describe the solution** — How should it work?
5. **Describe alternatives** — What alternatives have you considered?

### Feature Request Template

```markdown
**Problem:**
A clear description of what problem this feature would solve.

**Proposed Solution:**
How the feature should work.

**Alternative Solutions:**
What other approaches could solve this problem?

**Additional Context:**
Any relevant screenshots, mockups, or references.
```

---

## Community Guidelines

### Code of Conduct

We are committed to providing a welcoming and inclusive environment.

**Expected behavior:**
- Be respectful and considerate
- Use welcoming language
- Accept constructive criticism
- Focus on what is best for the community

**Unacceptable behavior:**
- Harassment, intimidation, or discrimination
- Trolling, insulting, or derogatory comments
- Publishing private information without consent
- Any conduct that would be inappropriate in a professional setting

### Communication Channels

| Channel | Purpose |
|---|---|
| **GitHub Issues** | Bug reports, feature requests |
| **GitHub Discussions** | Questions, ideas, general discussion |
| **PR Reviews** | Code review and feedback |

### Recognition

Contributors are recognized in:
- Release notes
- Contributors page (README.md)
- All Contributors bot for automated recognition

---

**Thank you for contributing to SchemaHub!**
