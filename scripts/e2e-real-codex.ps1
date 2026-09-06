# e2e-real-codex.ps1
# Local-only: run the real codex binary through the host driver e2e path.
# Requires codex CLI installed and authenticated (ADR-0005: key stays in codex CLI).
# Usage: pwsh scripts/e2e-real-codex.ps1

$ErrorActionPreference = "Stop"

Write-Host "[e2e-real-codex] checking codex binary..."
$codex = (Get-Command codex -ErrorAction SilentlyContinue).Source
if (-not $codex) {
    Write-Error "codex CLI not found. Install and authenticate codex first."
    exit 1
}
Write-Host "[e2e-real-codex] using: $codex"

Write-Host "[e2e-real-codex] running scheduler host-e2e tests against real codex..."
go test ./internal/scheduler/... -run "TestScheduler_HostE2E_StubBinary" -v
if ($LASTEXITCODE -ne 0) {
    Write-Error "host e2e tests failed against real codex (may be 503/timeout)."
    exit $LASTEXITCODE
}
Write-Host "[e2e-real-codex] done. Real codex host path verified locally."
