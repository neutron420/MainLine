<#
.SYNOPSIS
    Initial SchemaHub project setup — checks prerequisites, installs tooling, and prepares the workspace.

.DESCRIPTION
    Verifies Go 1.22+, Node 20+, and Docker are installed; installs air, golangci-lint, buf CLI;
    sets up pnpm, frontend dependencies, Go modules, and generates protobuf code.

.PARAMETER NoDocker
    Skip Docker prerequisite checks.

.EXAMPLE
    .\scripts\setup.ps1
    .\scripts\setup.ps1 -NoDocker
#>

param(
    [switch]$NoDocker
)

$ErrorActionPreference = 'Stop'
$ScriptDir = $PSScriptRoot
$RepoRoot = Resolve-Path "$ScriptDir\.."

# ── Colours ─────────────────────────────────────────────────────────────────
$Green   = @{ ForegroundColor = 'Green'  }
$Red     = @{ ForegroundColor = 'Red'    }
$Yellow  = @{ ForegroundColor = 'Yellow' }
$Magenta = @{ ForegroundColor = 'Magenta'}

function Write-Step($Message) { Write-Host "`n━━━ $Message ━━━" @Magenta }
function Write-OK  ($Message) { Write-Host "  ✓ $Message" @Green }
function Write-Warn($Message) { Write-Host "  ⚠ $Message" @Yellow }
function Write-Err ($Message) { Write-Host "  ✗ $Message" @Red; exit 1 }

# ── Helper: check a command exists and optionally verify its version ─────────
function Test-Command($Name, $MinVersion, $VersionArg) {
    $cmd = Get-Command $Name -ErrorAction SilentlyContinue
    if (-not $cmd) {
        Write-Err "$Name not found. Install it and ensure it is on your PATH."
    }

    if ($MinVersion) {
        try {
            $verLine = & $Name $VersionArg 2>&1 | Out-String
            $verMatch = [regex]::Match($verLine, '(\d+\.\d+\.?\d*)')
            if (-not $verMatch.Success) { throw "could not parse version" }
            $installed = [Version]$verMatch.Groups[1].Value
            $required  = [Version]$MinVersion
            if ($installed -lt $required) {
                Write-Err "$Name version $MinVersion+ required (found $installed)"
            }
            Write-OK "$Name $installed (minimum $MinVersion)"
        } catch {
            Write-OK "$Name (version check skipped: $($_.Exception.Message))"
        }
    } else {
        Write-OK "$Name found"
    }
}

# ── Helper: install a Go tool via go install ─────────────────────────────────
function Install-GoTool($Name, $Module) {
    $toolPath = "$env:GOPATH\bin\$Name"
    if (-not (Get-Command $Name -ErrorAction SilentlyContinue)) {
        Write-Host "  → Installing $Name ... " @Yellow -NoNewline
        try {
            & go install $Module @('>', '$null', '2>&1')
            if ($LASTEXITCODE -ne 0) { throw "go install failed" }
            Write-Host "OK" @Green
        } catch {
            Write-Host "FAILED" @Red
            Write-Warn "Could not install $Name. You may need to install it manually."
        }
    } else {
        Write-OK "$Name already installed"
    }
}

function Install-NpmTool($Name) {
    if (-not (Get-Command $Name -ErrorAction SilentlyContinue)) {
        Write-Host "  → Installing $Name ... " @Yellow -NoNewline
        try {
            & npm install -g $Name @('>', '$null', '2>&1')
            if ($LASTEXITCODE -ne 0) { throw "npm install failed" }
            Write-Host "OK" @Green
        } catch {
            Write-Host "FAILED" @Red
            Write-Warn "Could not install $Name. You may need to install it manually."
        }
    } else {
        Write-OK "$Name already installed"
    }
}

# ═══════════════════════════════════════════════════════════════════════════════
Write-Step "Checking prerequisites"

Test-Command -Name 'go'       -MinVersion '1.22' -VersionArg 'version'
Test-Command -Name 'node'     -MinVersion '20.0' -VersionArg '--version'
Test-Command -Name 'npm'                       -VersionArg '--version'

if (-not $NoDocker) {
    Test-Command -Name 'docker' -MinVersion '24.0' -VersionArg '--version'
    Test-Command -Name 'docker-compose' -MinVersion '2.0' -VersionArg 'version'
}

# ── Go tooling ───────────────────────────────────────────────────────────────
Write-Step "Installing / checking Go tooling"

Install-GoTool -Name 'air'          -Module 'github.com/air-verse/air@latest'
Install-GoTool -Name 'golangci-lint' -Module 'github.com/golangci/golangci-lint/cmd/golangci-lint@latest'
Install-GoTool -Name 'buf'          -Module 'github.com/bufbuild/buf/cmd/buf@latest'

# ── pnpm ─────────────────────────────────────────────────────────────────────
Write-Step "Setting up pnpm"

Install-NpmTool -Name 'pnpm'

# ── Frontend dependencies ───────────────────────────────────────────────────
Write-Step "Installing frontend dependencies"

$frontendDir = "$RepoRoot\frontend"
if (Test-Path "$frontendDir\package.json") {
    Push-Location $frontendDir
    try {
        & pnpm install
        if ($LASTEXITCODE -ne 0) { Write-Err "pnpm install failed" }
        Write-OK "Frontend dependencies installed"
    } finally { Pop-Location }
} else {
    Write-Warn "frontend/package.json not found — skipping pnpm install"
}

# ── Go modules ───────────────────────────────────────────────────────────────
Write-Step "Downloading Go modules"

$backendDir = "$RepoRoot\backend"
if (Test-Path "$backendDir\go.mod") {
    Push-Location $backendDir
    try {
        & go mod download
        if ($LASTEXITCODE -ne 0) { Write-Err "go mod download failed" }
        Write-OK "Go modules downloaded"
    } finally { Pop-Location }
} else {
    Write-Err "backend/go.mod not found — is this the correct repository?"
}

# ── Protobuf code generation ────────────────────────────────────────────────
Write-Step "Generating protobuf code"

if (Get-Command 'buf' -ErrorAction SilentlyContinue) {
    $protoDir = "$RepoRoot\proto"
    if (Test-Path "$protoDir\buf.yaml") {
        Push-Location $RepoRoot
        try {
            & buf generate proto
            if ($LASTEXITCODE -ne 0) { Write-Err "buf generate failed" }
            Write-OK "Protobuf code generated"
        } finally { Pop-Location }
    } else {
        Write-Warn "proto/buf.yaml not found — skipping buf generate"
    }
} else {
    Write-Warn "buf CLI not available — skipping protobuf generation"
}

# ── Docker Compose instructions ─────────────────────────────────────────────
Write-Step "Next steps"

if (-not $NoDocker) {
    Write-Host @"

  Your environment is ready. To start the full stack:

    docker compose -f docker/docker-compose.yml -f docker/docker-compose.dev.yml up -d

  To start the frontend only (outside Docker):

    cd frontend && pnpm dev

  To start the backend only (outside Docker):

    cd backend && air

"@
} else {
    Write-Host @"

  Setup complete (Docker checks skipped). Use the commands above for Docker-based
  development, or run services individually with pnpm dev / go run / air.

"@
}

Write-Host "`nSetup finished successfully!`n" @Green
exit 0
