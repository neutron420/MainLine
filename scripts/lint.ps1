<#
.SYNOPSIS
    Run all SchemaHub linters — Go, TypeScript, and Protobuf.

.DESCRIPTION
    Executes go fmt, go vet, golangci-lint for Go code; npm run lint and
    npm run typecheck for TypeScript; buf lint for protobuf. Exits on first failure.

.PARAMETER Fix
    Auto-fix Go formatting with go fmt (runs before other checks).

.EXAMPLE
    .\scripts\lint.ps1
    .\scripts\lint.ps1 -Fix
#>

param(
    [switch]$Fix
)

$ErrorActionPreference = 'Stop'
$ScriptDir = $PSScriptRoot
$RepoRoot = Resolve-Path "$ScriptDir\.."

$Green   = @{ ForegroundColor = 'Green'  }
$Red     = @{ ForegroundColor = 'Red'    }
$Yellow  = @{ ForegroundColor = 'Yellow' }
$Magenta = @{ ForegroundColor = 'Magenta'}

$global:HasErrors = $false
$global:Failures  = @()

function Write-Step($Message) { Write-Host "`n━━━ $Message ━━━" @Magenta }
function Write-OK  ($Message) { Write-Host "  ✓ $Message" @Green }
function Write-Warn($Message) { Write-Host "  ⚠ $Message" @Yellow }
function Write-Err ($Message) { Write-Host "  ✗ $Message" @Red }

function Invoke-Check($Name, $ScriptBlock) {
    Write-Host "  → $Name ... " @Yellow -NoNewline
    try {
        & $ScriptBlock
        if ($LASTEXITCODE -ne 0) { throw "exit code $LASTEXITCODE" }
        Write-Host "PASS" @Green
    } catch {
        Write-Host "FAIL" @Red
        $global:HasErrors = $true
        $global:Failures += $Name
        throw $_
    }
}

# ═══════════════════════════════════════════════════════════════════════════════
# Go
# ═══════════════════════════════════════════════════════════════════════════════
Write-Step "Go"

$backendDir = "$RepoRoot\backend"
if (-not (Test-Path "$backendDir\go.mod")) {
    Write-Warn "backend/go.mod not found — skipping Go checks"
} else {
    Push-Location $backendDir
    try {
        if ($Fix) {
            Write-Host "  → go fmt (fix) ... " @Yellow -NoNewline
            & go fmt ./... 2>&1
            if ($LASTEXITCODE -eq 0) { Write-Host "DONE" @Green } else { Write-Host "DONE (with changes)" @Green }
        }

        Invoke-Check 'go fmt'       { $result = & go fmt ./... 2>&1; if ($result) { throw "unformatted files:`n$result" } }
        Invoke-Check 'go vet'       { & go vet ./... 2>&1 }
    } finally { Pop-Location }

    if (Get-Command 'golangci-lint' -ErrorAction SilentlyContinue) {
        Invoke-Check 'golangci-lint run' {
            Push-Location $backendDir
            try { & golangci-lint run --out-format=line-number 2>&1 } finally { Pop-Location }
        }
    } else {
        Write-Warn "golangci-lint not found — skipping"
    }
}

# ═══════════════════════════════════════════════════════════════════════════════
# TypeScript
# ═══════════════════════════════════════════════════════════════════════════════
Write-Step "TypeScript"

$frontendDir = "$RepoRoot\frontend"
if (-not (Test-Path "$frontendDir\package.json")) {
    Write-Warn "frontend/package.json not found — skipping TypeScript checks"
} else {
    Push-Location $frontendDir
    try {
        Invoke-Check 'npm run lint'       { & npm run lint 2>&1 }
        Invoke-Check 'npm run typecheck'  { & npm run typecheck 2>&1 }
    } finally { Pop-Location }
}

# ═══════════════════════════════════════════════════════════════════════════════
# Proto
# ═══════════════════════════════════════════════════════════════════════════════
Write-Step "Protocol Buffers"

if (Get-Command 'buf' -ErrorAction SilentlyContinue) {
    $protoDir = "$RepoRoot\proto"
    if (Test-Path "$protoDir\buf.yaml") {
        Invoke-Check 'buf lint' {
            Push-Location $RepoRoot
            try { & buf lint proto 2>&1 } finally { Pop-Location }
        }
    } else {
        Write-Warn "proto/buf.yaml not found — skipping buf lint"
    }
} else {
    Write-Warn "buf CLI not found — skipping buf lint"
}

# ═══════════════════════════════════════════════════════════════════════════════
# Summary
# ═══════════════════════════════════════════════════════════════════════════════
Write-Host "`n━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━" @Magenta
if ($HasErrors) {
    Write-Err "FAILED — $($Failures.Count) check(s) failed:"
    foreach ($f in $Failures) { Write-Host "    • $f" @Red }
    exit 1
} else {
    Write-OK "All checks passed"
    exit 0
}
