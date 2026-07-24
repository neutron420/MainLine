<#
.SYNOPSIS
    Regenerate protobuf code from .proto definitions.

.DESCRIPTION
    Runs buf generate on the proto/ directory to regenerate Go code from
    Protocol Buffer definitions. Optionally checks git diff to verify that
    generated code is up-to-date, and supports a file watcher mode.

.PARAMETER Watch
    Enable file watcher mode — re-runs buf generate whenever a .proto file changes.

.EXAMPLE
    .\scripts\gen-proto.ps1
    .\scripts\gen-proto.ps1 -Watch
#>

param(
    [switch]$Watch
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

# ── Verify tooling ──────────────────────────────────────────────────────────
if (-not (Get-Command 'buf' -ErrorAction SilentlyContinue)) {
    Write-Err "buf CLI is not installed. Install it via 'go install github.com/bufbuild/buf/cmd/buf@latest'."
}

if (-not (Test-Path "$RepoRoot\proto\buf.yaml")) {
    Write-Err "proto/buf.yaml not found — are you in the correct directory?"
}

if (-not (Test-Path "$RepoRoot\proto\buf.gen.yaml")) {
    Write-Err "proto/buf.gen.yaml not found — cannot generate code."
}

# ── Generate ────────────────────────────────────────────────────────────────
function Generate-Proto {
    Write-Step "Generating protobuf code"

    Push-Location $RepoRoot
    try {
        Write-Host "  → Running buf generate ... " @Yellow -NoNewline
        $output = & buf generate proto 2>&1
        $exitCode = $LASTEXITCODE
        if ($exitCode -ne 0) {
            Write-Host "FAILED" @Red
            Write-Host $output @Red
            return $false
        }
        Write-Host "OK" @Green
        return $true
    } finally { Pop-Location }
}

# ── Check git diff ──────────────────────────────────────────────────────────
function Check-GitDiff {
    Write-Step "Verifying generated code is up-to-date"

    Push-Location $RepoRoot
    try {
        $diff = & git diff --stat 2>&1
        if ($LASTEXITCODE -ne 0) {
            Write-Warn "Not a git repository or git unavailable — skipping diff check"
            return
        }

        if ($diff) {
            Write-Host "  ⚠ Uncommitted changes detected:" @Yellow
            & git diff --stat 2>&1 | ForEach-Object { Write-Host "    $_" }
            Write-Warn "Generated code differs from the committed version."
            Write-Warn "Review the changes and commit them if expected."
        } else {
            Write-OK "No differences — generated code is in sync"
        }
    } finally { Pop-Location }
}

# ═══════════════════════════════════════════════════════════════════════════════
# Main
# ═══════════════════════════════════════════════════════════════════════════════

if (-not $Watch) {
    $ok = Generate-Proto
    if ($ok) { Check-GitDiff }
    exit $(if ($ok) { 0 } else { 1 })
}

# ── Watch mode ──────────────────────────────────────────────────────────────
Write-Step "Watch mode"
Write-Host "  Watching .proto files in proto/ for changes ..." @Cyan
Write-Host "  Press Ctrl+C to stop." @Yellow

$watcher = New-Object System.IO.FileSystemWatcher
$watcher.Path = "$RepoRoot\proto"
$watcher.Filter = '*.proto'
$watcher.IncludeSubdirectories = $true
$watcher.EnableRaisingEvents = $true

$timer = $null

Register-ObjectEvent $watcher 'Changed' -Action {
    # Debounce: coalesce rapid changes into a single run
    if ($timer) { $timer.Dispose() }
    $timer = New-Object System.Timers.Timer(500)
    $timer.AutoReset = $false
    Register-ObjectEvent $timer 'Elapsed' -Action {
        Write-Host "`n  Change detected — regenerating ..." @Yellow
        Push-Location "$($Event.MessageData.RepoRoot)"
        try {
            $output = & buf generate proto 2>&1
            if ($LASTEXITCODE -eq 0) {
                Write-Host "  ✓ Regenerated successfully at $(Get-Date -Format 'HH:mm:ss')" @Green
            } else {
                Write-Host "  ✗ Regeneration failed:" @Red
                Write-Host $output @Red
            }
        } finally { Pop-Location }
        $timer.Dispose()
    } -MessageData @{ RepoRoot = $RepoRoot } | Out-Null
    $timer.Start()
}.GetNewClosure()

try {
    # Keep the script alive
    while ($true) { Start-Sleep -Seconds 1 }
} finally {
    $watcher.EnableRaisingEvents = $false
    $watcher.Dispose()
    Write-Host "`n" @Yellow
    Write-Warn "Watcher stopped"
}

exit 0
