$ErrorActionPreference = "Stop"
$root = (Resolve-Path (Join-Path $PSScriptRoot "..")).Path
Push-Location -LiteralPath $root
try {
  Write-Host "[release-readiness] go build"
  & go build ./...
  Write-Host "[release-readiness] go vet"
  & go vet ./...
  Write-Host "[release-readiness] go test"
  & go test ./... -count=1
  Write-Host "[release-readiness] godocgen"
  $version = if ($env:VERSION) { $env:VERSION } else { "dev" }
  & go run ./cmd/godocgen --locale=en,zh,ja,ko,fr,de,ru,es,it --version=$version
  if (Get-Command goreleaser -ErrorAction SilentlyContinue) {
    Write-Host "[release-readiness] goreleaser check"
    & goreleaser check
  } else {
    Write-Host "[release-readiness] goreleaser not installed; CI validates the release config"
  }
  Write-Host "[release-readiness] PASS"
} finally {
  Pop-Location
}