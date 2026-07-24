<#
.SYNOPSIS
    Start the SchemaHub development environment via Docker Compose.

.DESCRIPTION
    Starts Docker Compose in detached mode and tails container logs.
    Handles Ctrl+C gracefully by stopping containers.

.PARAMETER BackendOnly
    Start only backend and its dependencies (postgres, redis).

.PARAMETER FrontendOnly
    Start only the frontend service.

.PARAMETER All
    Start all services (default). Explicit override of other flags.

.EXAMPLE
    .\scripts\dev.ps1
    .\scripts\dev.ps1 -BackendOnly
    .\scripts\dev.ps1 -FrontendOnly
#>

param(
    [switch]$BackendOnly,
    [switch]$FrontendOnly,
    [switch]$All
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
function Write-Warn($Message) { Write-Host "  ⚠ $Message" @Yellow }
function Write-Err ($Message) { Write-Host "  ✗ $Message" @Red; exit 1 }

# ── Preflight ────────────────────────────────────────────────────────────────
if (-not (Get-Command 'docker' -ErrorAction SilentlyContinue)) {
    Write-Err "docker is not installed or not on PATH."
}
if (-not (Get-Command 'docker-compose' -ErrorAction SilentlyContinue)) {
    Write-Warn "docker-compose plugin not found — trying docker compose"
}

$composeFiles = @(
    "$RepoRoot\docker\docker-compose.yml"
    "$RepoRoot\docker\docker-compose.dev.yml"
)

foreach ($f in $composeFiles) {
    if (-not (Test-Path $f)) {
        Write-Err "Required compose file not found: $f"
    }
}

# ── Determine which services to start ───────────────────────────────────────
$services = @()

if ($FrontendOnly) {
    Write-Step "Starting frontend only"
    $services = @('frontend')
} elseif ($BackendOnly) {
    Write-Step "Starting backend and infrastructure"
    $services = @('postgres', 'redis', 'backend', 'envoy')
} else {
    Write-Step "Starting all services"
}

$composeArgs = @()
foreach ($f in $composeFiles) {
    $composeArgs += '-f', $f
}
$composeArgs += 'up', '-d'

if ($services.Count -gt 0) {
    $composeArgs += $services
}

Write-Host "Running: docker compose $($composeArgs -join ' ')" @Cyan

# ── Start containers ────────────────────────────────────────────────────────
Push-Location $RepoRoot
try {
    $result = & docker compose $composeArgs 2>&1
    $exitCode = $LASTEXITCODE
    if ($exitCode -ne 0) {
        Write-Err "Docker Compose failed:`n$($result -join "`n")"
    }
    Write-OK "Containers started"
} finally { Pop-Location }

# ── Tail logs ───────────────────────────────────────────────────────────────
Write-Step "Tailing logs (press Ctrl+C to stop)"

$logArgs = @()
foreach ($f in $composeFiles) {
    $logArgs += '-f', $f
}
$logArgs += 'logs', '-f'

if ($services.Count -gt 0) {
    $logArgs += $services
}

try {
    Push-Location $RepoRoot
    & docker compose $logArgs 2>&1
} catch {
    # Ctrl+C or interrupt — fall through to clean up
} finally {
    Pop-Location
}

# ── Graceful shutdown on Ctrl+C ─────────────────────────────────────────────
Write-Host "`n" @Yellow
Write-Warn "Shutting down containers ..."

$downArgs = @()
foreach ($f in $composeFiles) {
    $downArgs += '-f', $f
}
$downArgs += 'down'

Push-Location $RepoRoot
try {
    & docker compose $downArgs 2>&1
    Write-OK "Containers stopped and cleaned up"
} catch {
    Write-Warn "Failed to stop containers gracefully — you may need to run 'docker compose down' manually"
} finally {
    Pop-Location
}

exit 0
