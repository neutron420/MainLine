<#
.SYNOPSIS
    Build all SchemaHub artifacts — Go backend, TypeScript frontend, and Docker images.

.DESCRIPTION
    Compiles the Go backend to the bin/ directory, builds the Next.js frontend,
    and optionally builds Docker images. Supports cross-compilation via -Platform.

.PARAMETER Platform
    Target OS for Go build. Valid values: windows, linux, darwin. Defaults to the
    current OS.

.EXAMPLE
    .\scripts\build.ps1
    .\scripts\build.ps1 -Platform linux
    .\scripts\build.ps1 -Platform windows
#>

param(
    [ValidateSet('windows', 'linux', 'darwin')]
    [string]$Platform
)

$ErrorActionPreference = 'Stop'
$ScriptDir = $PSScriptRoot
$RepoRoot = Resolve-Path "$ScriptDir\.."

$Green   = @{ ForegroundColor = 'Green'  }
$Red     = @{ ForegroundColor = 'Red'    }
$Yellow  = @{ ForegroundColor = 'Yellow' }
$Magenta = @{ ForegroundColor = 'Magenta'}
$Cyan    = @{ ForegroundColor = 'Cyan'   }

$BuildErrors = $false

function Write-Step($Message) { Write-Host "`n━━━ $Message ━━━" @Magenta }
function Write-OK  ($Message) { Write-Host "  ✓ $Message" @Green }
function Write-Err ($Message) { Write-Host "  ✗ $Message" @Red }

# ── Determine target platform ───────────────────────────────────────────────
if (-not $Platform) {
    $goos = if ($env:GOOS) { $env:GOOS } else { (go env GOOS 2>$null) }
    if (-not $goos) { $goos = 'windows' }
} else {
    $goos = $Platform
}

$goarch = if ($env:GOARCH) { $env:GOARCH } else { (go env GOARCH 2>$null) }
if (-not $goarch) { $goarch = 'amd64' }

$binaryExt = if ($goos -eq 'windows') { '.exe' } else { '' }

Write-Host "  Target:     $goos/$goarch" @Cyan
Write-Host "  Binary ext: $binaryExt" @Cyan

# ═══════════════════════════════════════════════════════════════════════════════
# Go build
# ═══════════════════════════════════════════════════════════════════════════════
Write-Step "Building Go backend"

$backendDir = "$RepoRoot\backend"
if (-not (Test-Path "$backendDir\go.mod")) {
    Write-Err "backend/go.mod not found — skipping Go build"
} else {
    $binDir = "$RepoRoot\bin"
    New-Item -ItemType Directory -Path $binDir -Force | Out-Null

    $outputName = "server$binaryExt"
    $outputPath = "$binDir\$outputName"

    Write-Host "  → Compiling cmd/server -> $outputPath ... " @Yellow -NoNewline

    Push-Location $backendDir
    try {
        $env:GOOS   = $goos
        $env:GOARCH = $goarch
        $env:CGO_ENABLED = '0'

        & go build -ldflags="-s -w" -o $outputPath ./cmd/server 2>&1
        if ($LASTEXITCODE -ne 0) {
            Write-Host "FAILED" @Red
            $BuildErrors = $true
        } else {
            Write-Host "OK" @Green
            Write-OK "Binary: $outputPath"
        }
    } finally {
        Remove-Item -Path 'env:GOOS', 'env:GOARCH', 'env:CGO_ENABLED' -ErrorAction SilentlyContinue
        Pop-Location
    }
}

# ═══════════════════════════════════════════════════════════════════════════════
# Frontend build
# ═══════════════════════════════════════════════════════════════════════════════
Write-Step "Building frontend"

$frontendDir = "$RepoRoot\frontend"
if (-not (Test-Path "$frontendDir\package.json")) {
    Write-Warn "frontend/package.json not found — skipping frontend build"
} else {
    Push-Location $frontendDir
    try {
        Write-Host "  → npm run build ... " @Yellow -NoNewline
        & npm run build 2>&1
        if ($LASTEXITCODE -ne 0) {
            Write-Host "FAILED" @Red
            $BuildErrors = $true
        } else {
            Write-Host "OK" @Green
            Write-OK "Frontend built"
        }
    } finally { Pop-Location }
}

# ═══════════════════════════════════════════════════════════════════════════════
# Docker build
# ═══════════════════════════════════════════════════════════════════════════════
Write-Step "Building Docker images"

$composeFile = "$RepoRoot\docker\docker-compose.yml"
if (-not (Test-Path $composeFile)) {
    Write-Warn "docker/docker-compose.yml not found — skipping Docker build"
} else {
    Push-Location $RepoRoot
    try {
        Write-Host "  → docker compose build ... " @Yellow -NoNewline
        & docker compose -f $composeFile build 2>&1
        if ($LASTEXITCODE -ne 0) {
            Write-Host "FAILED" @Red
            $BuildErrors = $true
        } else {
            Write-Host "OK" @Green
            Write-OK "Docker images built"
        }
    } finally { Pop-Location }
}

# ═══════════════════════════════════════════════════════════════════════════════
# Summary
# ═══════════════════════════════════════════════════════════════════════════════
Write-Host "`n━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━" @Magenta
if ($BuildErrors) {
    Write-Err "Build completed with errors"
    exit 1
} else {
    Write-OK "All artifacts built successfully"
    exit 0
}
