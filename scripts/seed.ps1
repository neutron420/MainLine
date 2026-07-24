<#
.SYNOPSIS
    Seed the SchemaHub database with sample data for development.

.DESCRIPTION
    Connects to PostgreSQL (via docker exec) and inserts sample projects, users,
    schemas, and migrations. Reads connection details from environment variables
    or .env file.

.PARAMETER Clean
    Drop and recreate the schema before seeding (destructive).

.EXAMPLE
    .\scripts\seed.ps1
    .\scripts\seed.ps1 -Clean
#>

param(
    [switch]$Clean
)

$ErrorActionPreference = 'Stop'
$ScriptDir = $PSScriptRoot
$RepoRoot = Resolve-Path "$ScriptDir\.."

$Green   = @{ ForegroundColor = 'Green'  }
$Red     = @{ ForegroundColor = 'Red'    }
$Yellow  = @{ ForegroundColor = 'Yellow' }
$Magenta = @{ ForegroundColor = 'Magenta'}
$Cyan    = @{ ForegroundColor = 'Cyan'   }

function Write-Step($Message) { Write-Host "`n━━━ $Message ━━━" @Magenta }
function Write-OK  ($Message) { Write-Host "  ✓ $Message" @Green }
function Write-Err ($Message) { Write-Host "  ✗ $Message" @Red; exit 1 }

# ── Load env vars from .env if present ──────────────────────────────────────
$envFile = "$RepoRoot\.env"
if (Test-Path $envFile) {
    Write-Host "  → Loading .env file ... " @Yellow -NoNewline
    Get-Content $envFile | ForEach-Object {
        if ($_ -match '^\s*([^#=]+)=(.*)\s*$') {
            $name = $matches[1].Trim()
            $value = $matches[2].Trim().Trim('"', "'")
            Set-Item -Path "env:$name" -Value $value -ErrorAction SilentlyContinue
        }
    }
    Write-Host "DONE" @Green
}

# ── Connection config ───────────────────────────────────────────────────────
$pgUser     = if ($env:POSTGRES_USER)     { $env:POSTGRES_USER }     else { 'schemahub' }
$pgPassword = if ($env:POSTGRES_PASSWORD) { $env:POSTGRES_PASSWORD } else { 'schemahub' }
$pgDB       = if ($env:POSTGRES_DB)       { $env:POSTGRES_DB }       else { 'schemahub' }
$pgHost     = 'localhost'
$pgPort     = '5432'

$containerName = 'schemahub-postgres-1'

Write-Step "Database connection"
Write-Host "  Host:      $pgHost`:$pgPort" @Cyan
Write-Host "  Database:  $pgDB" @Cyan
Write-Host "  User:      $pgUser" @Cyan
Write-Host "  Container: $containerName" @Cyan

# ── Verify container is running ─────────────────────────────────────────────
$containerRunning = docker ps --filter "name=$containerName" --filter "status=running" --format "{{.Names}}" 2>$null
if (-not $containerRunning) {
    Write-Err "PostgreSQL container '$containerName' is not running. Start it with: docker compose -f docker/docker-compose.yml up -d postgres"
}
Write-OK "Container is running"

# ── Helper: execute SQL via docker exec ─────────────────────────────────────
function Invoke-PSQL($Sql, $Description) {
    Write-Host "  → $Description ... " @Yellow -NoNewline
    $result = docker exec -i $containerName psql -U $pgUser -d $pgDB -c $Sql 2>&1
    if ($LASTEXITCODE -ne 0) {
        Write-Host "FAILED" @Red
        Write-Err "SQL error: $result"
    }
    Write-Host "OK" @Green
}

# ── Clean if requested ──────────────────────────────────────────────────────
if ($Clean) {
    Write-Step "Cleaning database"
    Write-Warn "This will DROP ALL TABLES in the '$pgDB' database!"
    Write-Host "  Press Ctrl+C within 5 seconds to cancel ... " @Yellow
    Start-Sleep -Seconds 5

    Invoke-PSQL @"
DROP SCHEMA IF EXISTS public CASCADE;
CREATE SCHEMA public;
"@ "Dropping and recreating public schema"

    # Re-run the application's schema init if a SQL file exists
    $schemaFile = "$RepoRoot\backend\migrations\init.sql"
    if (Test-Path $schemaFile) {
        Write-Host "  → Running schema init ... " @Yellow -NoNewline
        docker exec -i $containerName psql -U $pgUser -d $pgDB < $schemaFile 2>&1 | Out-Null
        if ($LASTEXITCODE -ne 0) { Write-Host "FAILED" @Red; Write-Err "Schema init failed" }
        Write-Host "OK" @Green
    }
}

# ═══════════════════════════════════════════════════════════════════════════════
# Seed data
# ═══════════════════════════════════════════════════════════════════════════════
Write-Step "Seeding sample data"

# Users
Invoke-PSQL @"
INSERT INTO users (id, email, password_hash, display_name, role, is_active, email_verified_at, created_at, updated_at)
VALUES
  ('a0000000-0000-0000-0000-000000000001', 'alice@example.com', '\$2a\$12\$LJ3m4ys3Lk0TSwHnbfOMiOXPm1Qlq3UzO0q.7pLq7pLq7pLq7pLq', 'Alice Admin', 'admin', true, NOW(), NOW(), NOW()),
  ('a0000000-0000-0000-0000-000000000002', 'bob@example.com',   '\$2a\$12\$LJ3m4ys3Lk0TSwHnbfOMiOXPm1Qlq3UzO0q.7pLq7pLq7pLq7pLq', 'Bob Builder',  'user',  true, NOW(), NOW(), NOW()),
  ('a0000000-0000-0000-0000-000000000003', 'carol@example.com', '\$2a\$12\$LJ3m4ys3Lk0TSwHnbfOMiOXPm1Qlq3UzO0q.7pLq7pLq7pLq7pLq', 'Carol Dev',    'user',  true, NOW(), NOW(), NOW())
ON CONFLICT (email) DO NOTHING;
"@ "Inserting sample users"

# Projects
Invoke-PSQL @"
INSERT INTO projects (id, name, slug, description, visibility, created_by, created_at, updated_at)
VALUES
  ('b0000000-0000-0000-0000-000000000001', 'User Service',       'user-service',       'Microservice for user management and authentication', 'private', 'a0000000-0000-0000-0000-000000000001', NOW(), NOW()),
  ('b0000000-0000-0000-0000-000000000002', 'Payment Service',    'payment-service',    'Handles billing, invoices, and payment processing',    'team',    'a0000000-0000-0000-0000-000000000001', NOW(), NOW()),
  ('b0000000-0000-0000-0000-000000000003', 'Analytics Pipeline', 'analytics-pipeline', 'Data warehouse for product analytics',                  'public',  'a0000000-0000-0000-0000-000000000002', NOW(), NOW())
ON CONFLICT (slug) DO NOTHING;
"@ "Inserting sample projects"

# Project members
Invoke-PSQL @"
INSERT INTO project_members (id, project_id, user_id, role, joined_at, created_at)
VALUES
  ('c0000000-0000-0000-0000-000000000001', 'b0000000-0000-0000-0000-000000000001', 'a0000000-0000-0000-0000-000000000001', 'owner',  NOW(), NOW()),
  ('c0000000-0000-0000-0000-000000000002', 'b0000000-0000-0000-0000-000000000001', 'a0000000-0000-0000-0000-000000000002', 'member', NOW(), NOW()),
  ('c0000000-0000-0000-0000-000000000003', 'b0000000-0000-0000-0000-000000000002', 'a0000000-0000-0000-0000-000000000001', 'owner',  NOW(), NOW()),
  ('c0000000-0000-0000-0000-000000000004', 'b0000000-0000-0000-0000-000000000002', 'a0000000-0000-0000-0000-000000000003', 'member', NOW(), NOW()),
  ('c0000000-0000-0000-0000-000000000005', 'b0000000-0000-0000-0000-000000000003', 'a0000000-0000-0000-0000-000000000002', 'owner',  NOW(), NOW())
ON CONFLICT (project_id, user_id) DO NOTHING;
"@ "Inserting project members"

# Connections
Invoke-PSQL @"
INSERT INTO connections (id, project_id, name, host, port, database_name, username, password_encrypted, ssl_mode, connection_status, created_by, created_at, updated_at)
VALUES
  ('d0000000-0000-0000-0000-000000000001', 'b0000000-0000-0000-0000-000000000001', 'User DB - Production', 'prod-db.example.com', 5432, 'users_db', 'schemahub_ro', 'encrypted_placeholder', 'require', 'connected', 'a0000000-0000-0000-0000-000000000001', NOW(), NOW()),
  ('d0000000-0000-0000-0000-000000000002', 'b0000000-0000-0000-0000-000000000001', 'User DB - Staging',   'staging-db.example.com', 5432, 'users_db', 'schemahub_ro', 'encrypted_placeholder', 'require', 'connected', 'a0000000-0000-0000-0000-000000000001', NOW(), NOW()),
  ('d0000000-0000-0000-0000-000000000003', 'b0000000-0000-0000-0000-000000000002', 'Payment DB',          'pay-db.example.com', 5432, 'payments', 'schemahub_ro', 'encrypted_placeholder', 'require', 'unknown',  'a0000000-0000-0000-0000-000000000001', NOW(), NOW())
ON CONFLICT DO NOTHING;
"@ "Inserting database connections"

# Schemas
Invoke-PSQL @"
INSERT INTO schemas (id, project_id, connection_id, schema_name, created_at, updated_at)
VALUES
  ('e0000000-0000-0000-0000-000000000001', 'b0000000-0000-0000-0000-000000000001', 'd0000000-0000-0000-0000-000000000001', 'public', NOW(), NOW()),
  ('e0000000-0000-0000-0000-000000000002', 'b0000000-0000-0000-0000-000000000002', 'd0000000-0000-0000-0000-000000000003', 'public', NOW(), NOW())
ON CONFLICT DO NOTHING;
"@ "Inserting schemas"

# Schema versions (metadata as simplified JSONB)
Invoke-PSQL @"
INSERT INTO schema_versions (id, schema_id, version, checksum, metadata, object_count, created_by, created_at)
VALUES
  ('f0000000-0000-0000-0000-000000000001', 'e0000000-0000-0000-0000-000000000001', 1, 'sha256_abc123', '{"tables": [{"name": "users", "columns": [{"name": "id", "type": "uuid"}, {"name": "email", "type": "varchar"}]}]}', 1, 'a0000000-0000-0000-0000-000000000001', NOW()),
  ('f0000000-0000-0000-0000-000000000002', 'e0000000-0000-0000-0000-000000000002', 1, 'sha256_def456', '{"tables": [{"name": "invoices", "columns": [{"name": "id", "type": "uuid"}, {"name": "amount", "type": "decimal"}]}]}', 1, 'a0000000-0000-0000-0000-000000000001', NOW())
ON CONFLICT DO NOTHING;
"@ "Inserting schema versions"

# Migrations
Invoke-PSQL @"
INSERT INTO migrations (id, project_id, title, description, version, up_sql, down_sql, checksum, status, created_by, created_at, updated_at)
VALUES
  ('g0000000-0000-0000-0000-000000000001', 'b0000000-0000-0000-0000-000000000001', 'Create users table', 'Initial migration to create the users table', 'v1.0.0', 'CREATE TABLE users (id UUID PRIMARY KEY, email VARCHAR(320) NOT NULL);', 'DROP TABLE IF EXISTS users;', 'sha256_mig001', 'completed', 'a0000000-0000-0000-0000-000000000001', NOW(), NOW()),
  ('g0000000-0000-0000-0000-000000000002', 'b0000000-0000-0000-0000-000000000001', 'Add role column',    'Add role column to users table',               'v1.1.0', 'ALTER TABLE users ADD COLUMN role VARCHAR(20);',  'ALTER TABLE users DROP COLUMN role;',  'sha256_mig002', 'draft',     'a0000000-0000-0000-0000-000000000002', NOW(), NOW()),
  ('g0000000-0000-0000-0000-000000000003', 'b0000000-0000-0000-0000-000000000002', 'Create invoices',   'Initial payment schema',                         'v1.0.0', 'CREATE TABLE invoices (id UUID PRIMARY KEY, amount DECIMAL(10,2) NOT NULL);', 'DROP TABLE IF EXISTS invoices;', 'sha256_mig003', 'completed', 'a0000000-0000-0000-0000-000000000001', NOW(), NOW())
ON CONFLICT DO NOTHING;
"@ "Inserting migrations"

# Migration runs
Invoke-PSQL @"
INSERT INTO migration_runs (id, migration_id, connection_id, direction, status, started_at, completed_at, duration_ms, executed_by, created_at)
VALUES
  ('h0000000-0000-0000-0000-000000000001', 'g0000000-0000-0000-0000-000000000001', 'd0000000-0000-0000-0000-000000000001', 'up', 'completed', NOW() - INTERVAL '1 day', NOW() - INTERVAL '1 day', 1234, 'a0000000-0000-0000-0000-000000000001', NOW())
ON CONFLICT DO NOTHING;
"@ "Inserting migration runs"

Write-OK "Database seeded successfully"
exit 0
