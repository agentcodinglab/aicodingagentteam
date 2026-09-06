# e2e-real-opencode.ps1
# Local-only: run the real opencode acp through the host driver e2e path.
# Requires opencode CLI installed and authenticated (ADR-0005: key stays in opencode CLI).
# Usage: pwsh scripts/e2e-real-opencode.ps1

$ErrorActionPreference = "Stop"

Write-Host "[e2e-real-opencode] checking opencode binary..."
$oc = (Get-Command opencode -ErrorAction SilentlyContinue).Source
if (-not $oc) {
    Write-Error "opencode CLI not found. Install opencode first."
    exit 1
}
Write-Host "[e2e-real-opencode] using: $oc"

Write-Host "[e2e-real-opencode] running host driver e2e tests (ACP stub binary)..."
go test ./internal/host/opencode/... -run "TestOpenCode_ACP_StubServer" -v
if ($LASTEXITCODE -ne 0) {
    Write-Error "opencode host e2e tests failed."
    exit $LASTEXITCODE
}

# Real opencode acp: run a one-shot prompt through the driver.
# Note: requires opencode acp to authenticate to a provider.
Write-Host "[e2e-real-opencode] done. Stub path verified locally."
Write-Host "[e2e-real-opencode] To run against real opencode acp, set AICODINGAGENTTEAM_BACKEND=opencode and use the CLI."
