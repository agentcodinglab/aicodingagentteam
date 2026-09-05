# Changelog

All notable changes to AiCodingAgentTeam are documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added
- Root README.md with architecture overview and quick start guide
- MIT LICENSE file
- CONTRIBUTING.md with development workflow
- GitHub Issue and Pull Request templates
- goreleaser.yml for cross-platform binary builds
- GitHub Release workflow (tag-triggered)
- CI govulncheck security scanning step
- CI codecov coverage reporting
- RedisBus unit tests (miniredis)
- host/registry unit tests
- MCP Server real JSON-RPC stdio implementation
- TUI component unit tests (ink-testing-library + vitest)
- A2A DelegateParallel p95 benchmark
- ACP Server real stdio JSON-RPC 2.0 (initialize, session/start, session/stop, session/list)
- CLI subcommands: `continue`, `mcp`, `a2a serve`, `acp`, `memory show|capture|recall`, `ci`
- claude/dsh stub driver tests (coverage 0% to 100%)
- mcp coverage 77.9% to 89.6% (subdirectory + tool definition tests)
- version/commit/date metadata via goreleaser ldflags

### Changed
- golangci-lint-action v8 to v9 (Node 20 deprecation fix)
- GitHub Actions checkout@v4 to v5, setup-go@v5 to v6, setup-node@v4 to v5

## [0.1.0] - 2026-09-02

### Added
- Go orchestration engine: router to planner to scheduler to coordinator 5-layer flow
- Host drivers: Codex (real exec), OpenCode (real exec), Claude (stub), DSH (stub)
- A2A protocol: InProcBus + RedisBus Pub/Sub
- 9 team role agents with A2A reviewer registration
- Quality gate engine with deterministic checks (golangci-lint + go vet + go test)
- Governance engine with 113+ rules and fail-open design
- RAG knowledge engine (BM25, pure Go, zero dependencies)
- Project memory store (facts/pitfalls/lessons/recipes)
- gRPC API with proto definitions
- A2A HTTP server with Agent Card discovery
- MCP/ACP server stubs
- TypeScript TUI client (Ink + React + Zustand + gRPC)
- Docker Compose deployment (coordinator + redis + agents + hosts)
- ADR-0001 through ADR-0012 architecture decision records
- Full documentation suite (requirements, architecture, design, deployment)