#requires -Version 5.1
<#
.SYNOPSIS
  AiCodingAgentTeam end-to-end demo: init -> quick -> verify -> govern -> report.

.DESCRIPTION
  Dogfoods the full coordinator pipeline against this repository.
  No host CLI (Codex/OpenCode) required: reviewer agents run in-process
  and the quality gate executes the local Go toolchain.

.PARAMETER SkipSlow
  Skip the `verify` and `report` steps (they run real go build/test/lint, ~2 min).
#>
param(
    [switch]$SkipSlow
)

$ErrorActionPreference = 'Stop'
$root = Split-Path -Parent $PSScriptRoot
Set-Location $root

$bin = Join-Path $root 'aicodingagentteam.exe'
$step = 0

function Step([string]$msg) {
    $script:step++
    Write-Host "`n=== [$script:step] $msg ===" -ForegroundColor Cyan
}

# --- 0. Build ----------------------------------------------------------------
Step 'build coordinator binary'
go build -o $bin ./cmd/aicodingagentteam
if ($LASTEXITCODE -ne 0) { throw 'go build failed' }

# --- 1. Init workspace --------------------------------------------------------
Step 'init workspace (.aicodingagentteam/)'
& $bin init

# --- 2. Quick edit (fast path: plan=quick, no DAG) ----------------------------
Step 'quick edit pipeline'
& $bin quick 'update README quick-start section'
if ($LASTEXITCODE -ne 0) { throw 'quick failed' }

# --- 3. Full run (DAG plan; parks when QA finds missing artifacts) ------------
Step 'full pipeline run (build intent -> 9-node DAG)'
& $bin run 'build a hello-world web service' --backend codex
if ($LASTEXITCODE -ne 0) { throw 'run failed' }

if (-not $SkipSlow) {
    # --- 4. Quality gate verify ---------------------------------------------------
    Step 'quality gate verify (real go build/vet/test/lint)'
    & $bin verify
    if ($LASTEXITCODE -ne 0) { throw 'verify failed' }

    # --- 5. Governance scan -------------------------------------------------------
    Step 'governance scan'
    & $bin govern

    # --- 6. Quality report ----------------------------------------------------------
    Step 'quality report (writes output/quality-gate.md)'
    & $bin report
    if ($LASTEXITCODE -ne 0) { throw 'report failed' }
}

# --- 7. Show produced artifacts -----------------------------------------------
Step 'artifacts produced'
$paths = @(
    '.aicodingagentteam/plan.json',
    '.aicodingagentteam/audit/events.jsonl',
    'output/quality-gate.md'
)
foreach ($p in $paths) {
    $full = Join-Path $root $p
    if (Test-Path $full) {
        $size = (Get-Item $full).Length
        Write-Host ("  [OK] {0} ({1} bytes)" -f $p, $size) -ForegroundColor Green
    } else {
        Write-Host ("  [--] {0} (not created)" -f $p) -ForegroundColor DarkGray
    }
}

Write-Host "`nDemo complete." -ForegroundColor Green