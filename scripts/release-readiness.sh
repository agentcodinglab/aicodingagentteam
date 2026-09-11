#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$ROOT_DIR"

echo "[release-readiness] go build"
go build ./...
echo "[release-readiness] go vet"
go vet ./...
echo "[release-readiness] go test"
go test ./... -count=1
echo "[release-readiness] godocgen"
go run ./cmd/godocgen --locale=en,zh,ja,ko,fr,de,ru,es,it --version="${VERSION:-dev}"

if command -v goreleaser >/dev/null 2>&1; then
  echo "[release-readiness] goreleaser check"
  goreleaser check
else
  echo "[release-readiness] goreleaser not installed; CI validates the release config"
fi

echo "[release-readiness] PASS"