# Changelog

All notable changes to AiCodingAgentTeam are documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added
- End-to-end RAG + memory demo through Director.Handle (cmd `knowledge demo`)
- IndexDirectoryWithLimit API for capped repo indexing
- RegisterAllReviewers nil-bus guard for standalone demo use
- E2E tests: TestKnowledgeDemo_E2E_ThroughDirector, TestKnowledgeDemo_ReportWritten, TestKnowledgeDemo_NilKnowledge_NoPanic
- Governance CI: rag-demo job (soft-fail, uploads demo-report.md + demo-report.json artifacts)
- ADR-0017: Direction D end-to-end RAG demo decision record
- Implementation plan: docs/plan/direction-d-rag-demo.md

### Changed
- Demo report output written to .aicodingagentteam/ (gitignored runtime artifacts)

## [0.4.0] - 2026-09-05

### Added
- Direction D: end-to-end RAG + memory demo through Director.Handle
- IndexDirectoryWithLimit API for capped repo indexing
- RegisterAllReviewers nil-bus guard for standalone demo use
- E2E tests covering demo flow, report writer, and nil-knowledge safety
- Governance CI: rag-demo job (soft-fail, uploads demo-report artifacts)
- ADR-0017: Direction D end-to-end RAG demo decision record
- Implementation plan: docs/plan/direction-d-rag-demo.md

### Changed
- Demo report output written to .aicodingagentteam/ (gitignored runtime artifacts)


### Added
- Go E2E integration tests (quick edit + build + park/continue + API surface)
- proof-pack zip generation (plan.json + verify.jsonl + scorecard.md + delivery-summary.md)
- Fail-open governance fault injection tests (closed engine + excluded path + disabled rule)
- Concurrent write-lock mutex enforcement tests
- A2A contract JSON schema validation (AgentCard / Task / Result / ProgressEvent round-trip)
- CI: gitleaks secret scanning step
- CI: Semgrep SAST static analysis step
- CI: Trivy container image scanning job
- CLI smoke tests (version, init, memory show, knowledge demo, verify, knowledge index)
- TUI component render tests (StatusBar + ResultPanel + HelpPanel via ink-testing-library)
- RAG spec 13 acceptance checkboxes closed
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

## [0.2.1] - 2026-09-05

### Fixed
- Website: collapse README language picker into a single line (pipe-separated)

## [0.2.0] - 2026-09-05

### Added
- Marketing website at `website/` (Next.js 14 App Router + Tailwind + next-intl)
- 9-language static documentation site (en, zh, ja, ko, fr, de, ru, es, it)
- GitHub Pages static deployment via `.github/workflows/docs-site.yml`
- Localized README set under `website/README.{lang}.md` with cross-navigation
- Stage 1 redesign: best-practice landing sections (Hero / LogoCloud / Features / Stats / Code / Architecture / Quickstart / FinalCTA)
- Stage 2 dark duotone redesign (cyan + magenta neon) with three-font system (Manrope / Space Grotesk / JetBrains Mono)
- Hero 3-slide auto-typing terminal demo with cursor blink + RESTART button
- IntersectionObserver-driven Reveal scroll-in animations
- 3D mouse-follow Tilt wrapper on FeatureGrid and ArchitectureDiagram
- Markdown-rendered documentation pages (react-markdown + remark-gfm) with locale-prefixed navigation
- Smart root `/` redirect with browser-language detection + localStorage persistence
- OG image (`og.png`, 1200x630) and gradient `favicon.svg`

### Changed
- Documentation organization: site-internal docs moved under `docs/`, public docs under `website/content/docs/`

## [0.1.0] - 2026-09-02

### Added
- Go E2E integration tests (quick edit + build + park/continue + API surface)
- proof-pack zip generation (plan.json + verify.jsonl + scorecard.md + delivery-summary.md)
- Fail-open governance fault injection tests (closed engine + excluded path + disabled rule)
- Concurrent write-lock mutex enforcement tests
- A2A contract JSON schema validation (AgentCard / Task / Result / ProgressEvent round-trip)
- CI: gitleaks secret scanning step
- CI: Semgrep SAST static analysis step
- CI: Trivy container image scanning job
- CLI smoke tests (version, init, memory show, knowledge demo, verify, knowledge index)
- TUI component render tests (StatusBar + ResultPanel + HelpPanel via ink-testing-library)
- RAG spec 13 acceptance checkboxes closed
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