# Changelog

All notable changes to AiCodingAgentTeam are documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Changed
- docs: revised host-driver spec + 02/03 architecture docs to reflect ADR-0021 (opencode serve is Web UI, not JSON API; driver uses opencode acp)
- docs: ADR-0007 cross-referenced as superseded by ADR-0021 for the OpenCode serve recommendation
- docs: ADR README table extended to ADR-0021


### Added
- P5: opencode driver now drives `opencode acp` (stdio JSON-RPC) instead of `opencode run --format json` (ADR-0021 B)
- testdata/stubbin/opencode-acp stub binary (Unix + Windows) simulating ACP server
- TestOpenCode_ACP_StubServer drives the real opencode driver against the stub acp binary
- scripts/e2e-real-opencode.ps1 for local real opencode verification
- spike/opencode-serve/notes.md: OpenCode v1.18.18 serve is Web UI, not JSON API; opencode acp is the programming interface

### Changed
- opencode driver rewritten: runOnceStreaming -> runACPSession (initialize/session/new/session/prompt + notifications/session/update streaming)
- opencode Capabilities.SessionResume now true (ACP supports session reuse)
- TestCapabilities + TestResumeReturnsError updated for ACP mode

## [0.8.0] - 2026-09-06

### Added
- P5: opencode driver drives `opencode acp` (stdio JSON-RPC, ADR-0021 B)
- testdata/stubbin/opencode-acp stub + TestOpenCode_ACP_StubServer
- scripts/e2e-real-opencode.ps1; spike/opencode-serve/notes.md

### Changed
- opencode driver rewritten (runOnceStreaming -> runACPSession)
- opencode SessionResume true; TestCapabilities/TestResume updated


### Added
- P6: ACP v1 session/newTask method dispatching to Director in a goroutine
- ACP Server NewWithDirector constructor + DirectorLike interface
- notifications/session/update JSON-RPC push (no ID) for streamed TaskEvents (start/message/tool_call/done/error)
- Task + TaskEvent types for in-memory event history
- TestACP_SessionNewTask_StreamsEvents e2e over in-process stdio pipe

### Changed
- ACP Session struct now tracks active tasks per session

## [0.7.0] - 2026-09-06

### Added
- P6: ACP session/newTask + notifications/session/update (ADR-0020)
- DirectorLike abstraction; NewWithDirector constructor
- Task + TaskEvent types; TestACP_SessionNewTask_StreamsEvents

### Changed
- ACP Session struct now tracks active tasks per session


### Added
- P4: codex + opencode driver stdout streaming (StdoutPipe + bufio.Scanner, ADR-0019)
- testdata/stubbin/codex-stream multi-line stub for streaming tests
- TestCodex_SendTask_StreamingEvents verifying incremental EventMessage before EventDone

### Changed
- codex runOnce -> runOnceStreaming (aggregated stdout preserved for EventDone)
- opencode runOnce -> runOnceStreaming (JSON Lines parsed inline as they arrive)

## [0.6.0] - 2026-09-06

### Added
- P4: codex + opencode driver stdout streaming
- testdata/stubbin/codex-stream multi-line stub
- TestCodex_SendTask_StreamingEvents

### Changed
- codex + opencode runOnce -> runOnceStreaming


### Added
- Direction C: real host driver end-to-end verification via codex stub binary (Scheduler.WithDriver + Director.WithDriver)
- Scheduler writer nodes now dispatch to host driver.SendTask, persist stdout artifact to .aicodingagentteam/host/<nodeID>.txt
- Director.WithDriver option propagates driver to scheduler
- testdata/stubbin/codex (Unix) + codex.cmd (Windows) stub binaries for CI e2e (no API key)
- TestScheduler_HostE2E_StubBinary + TestScheduler_HostE2E_NilDriver_NoPanic
- governance CI: host-e2e job (soft-fail, codex stub binary)
- scripts/e2e-real-codex.ps1 for local real codex verification
- ADR-0018: Direction C real host e2e decision record
- Implementation plan: docs/plan/direction-c-real-host-e2e.md

### Changed
- Scheduler driver=nil degrades to legacy stub path (backward compatible)

## [0.5.0] - 2026-09-06

### Added
- Direction C: real host driver end-to-end verification via codex stub binary
- Scheduler writer nodes dispatch to host driver.SendTask, stdout artifact persisted
- Director.WithDriver option; testdata/stubbin stub binaries for CI
- TestScheduler_HostE2E_StubBinary + NilDriver_NoPanic
- governance CI host-e2e job (soft-fail)
- scripts/e2e-real-codex.ps1 for local real codex
- ADR-0018 + docs/plan/direction-c-real-host-e2e.md

### Changed
- Scheduler driver=nil degrades to legacy stub path


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