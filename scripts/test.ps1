<#
.SYNOPSIS
    Run all SchemaHub tests — Go unit/integration tests, TypeScript tests, and protobuf lint checks.

.DESCRIPTION
    Executes go test for Go packages, npm run test for TypeScript, and buf lint for protobuf.
    Supports coverage, short mode, and integration test flags.

.PARAMETER Coverage
    Generate Go coverage profile (coverage.out) and display per-package coverage.

.PARAMETER Short
    Run tests in short mode (-short flag for Go, avoiding slow integration tests).

.PARAMETER Integration
    Include integration tests (clears the -short flag, runs full test suite).

.EXAMPLE
    .\scripts\test.ps1
    .\scripts\test.ps1 -Coverage
    .\scripts\test.ps1 -Short
    .\scripts\test.ps1 -Integration
    .\scripts\test.ps1 -Coverage -Short
#>

param(
    [switch]$Coverage,
    [switch]$Short,
    [switch]$Integration
)

$ErrorActionPreference = 'Stop'
$ScriptDir = $PSScriptRoot
$RepoRoot = Resolve-Path "$ScriptDir\.."

$Green   = @{ ForegroundColor = 'Green'  }
$Red     = @{ ForegroundColor = 'Red'    }
$Yellow  = @{ ForegroundColor = 'Yellow' }
$Magenta = @{ ForegroundColor = 'Magenta'}
$Cyan    = @{ ForegroundColor = 'Cyan'   }

$global:HasErrors = $false
$global:Results   = @()

function Write-Step($Message) { Write-Host "`n━━━ $Message ━━━" @Magenta }
function Write-OK  ($Message) { Write-Host "  ✓ $Message" @Green }
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
    }
    $global:Results += @{ Name = $Name; Passed = $LASTEXITCODE -eq 0 }
}

# ═══════════════════════════════════════════════════════════════════════════════
# Go tests
# ═══════════════════════════════════════════════════════════════════════════════
Write-Step "Go tests"

$backendDir = "$RepoRoot\backend"
if (-not (Test-Path "$backendDir\go.mod")) {
    Write-Warn "backend/go.mod not found — skipping Go tests"
} else {
    $goTestArgs = @('test', './...')

    if ($Integration) {
        $goTestArgs += '-tags=integration'
        Write-Host "  (integration tests enabled)" @Cyan
    } elseif ($Short) {
        $goTestArgs += '-short'
        Write-Host "  (short mode)" @Cyan
    }

    if ($Coverage) {
        $goTestArgs += '-coverprofile=coverage.out', '-covermode=atomic'
    }

    if (-not $Integration) {
        $goTestArgs += '-race'
    }

    $goTestArgs += '-count=1'

    Invoke-Check 'go test' {
        Push-Location $backendDir
        try {
            & go $goTestArgs 2>&1
        } finally { Pop-Location }
    }

    if ($Coverage -and (Test-Path "$backendDir\coverage.out")) {
        Write-Host "`n  Coverage by package:" @Cyan
        Push-Location $backendDir
        try {
            & go tool cover -func=coverage.out 2>&1 | ForEach-Object {
                Write-Host "    $_"
            }
            $totalLine = (& go tool cover -func=coverage.out 2>&1 | Select-String 'total:')
            if ($totalLine) {
                $match = [regex]::Match($totalLine, '(\d+\.\d+)%')
                if ($match.Success) {
                    Write-Host "`n  Total coverage: $($match.Groups[1].Value)%" @Cyan
                }
            }
        } finally { Pop-Location }
    }
}

# ═══════════════════════════════════════════════════════════════════════════════
# TypeScript tests
# ═══════════════════════════════════════════════════════════════════════════════
Write-Step "TypeScript tests"

$frontendDir = "$RepoRoot\frontend"
if (-not (Test-Path "$frontendDir\package.json")) {
    Write-Warn "frontend/package.json not found — skipping TypeScript tests"
} else {
    $npmArgs = @('run', 'test')
    if ($Integration) { $npmArgs += '--', '--integration' }

    Invoke-Check 'npm run test' {
        Push-Location $frontendDir
        try {
            & npm $npmArgs 2>&1
        } finally { Pop-Location }
    }
}

# ═══════════════════════════════════════════════════════════════════════════════
# Proto lint
# ═══════════════════════════════════════════════════════════════════════════════
Write-Step "Protocol Buffers lint"

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
Write-Host "  Test Results Summary" @Cyan
Write-Host "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━" @Magenta

$passed = 0
$failed = 0
foreach ($r in $Results) {
    if ($r.Passed) { $passed++ } else { $failed++ }
    $icon = if ($r.Passed) { '✓' } else { '✗' }
    $color = if ($r.Passed) { $Green } else { $Red }
    Write-Host "  $icon $($r.Name)" @color
}

Write-Host "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━" @Magenta
Write-Host "  $passed passed, $failed failed" @(if ($failed -gt 0) { $Red } else { $Green })

if ($HasErrors) {
    exit 1
}
exit 0
