# Contributing to AiCodingAgentTeam

Thank you for your interest in contributing! This project follows the Vibe Coding methodology: **人对意图负责，AI 对实现负责**.

## Development Workflow

This project follows a three-phase closed loop for every feature:

```
规格(Spec) → 计划(Plan) → 实现(Implement) → 验证(Verify)
```

1. **Spec-first**: Write requirements to `docs/spec/{feature}.md` before coding.
2. **Plan**: Break down into atomic, verifiable steps in `docs/plan/{feature}.md`.
3. **Implement**: Test-first (TDD), one step at a time.
4. **Verify**: `make all` (lint + vet + test + build) must pass.

See [AGENTS.md](AGENTS.md) for the full specification.

## Prerequisites

- Go ≥ 1.25
- Node.js ≥ 20 (TUI client only)
- golangci-lint v2.13+
- Docker ≥ 24.0 (containerized deployment only)

## Getting Started

```bash
git clone https://github.com/agentcodinglab/aicodingagentteam.git
cd aicodingagentteam

# Build
make build

# Run all checks
make all

# Run coordinator
make run

# TUI client
cd tui && npm install && npm run build && node dist/cli.js --demo
```

## Local Development Quick Start

```bash
# 1. Clone and enter the repo
git clone https://github.com/agentcodinglab/aicodingagentteam.git
cd aicodingagentteam

# 2. Run all checks (lint + vet + test + build)
make all

# 3. Run Go tests with coverage
go test ./... -cover -count=1

# 4. Run the coordinator in demo mode (no API key needed)
go run ./cmd/aicodingagentteam knowledge demo

# 5. Website development (optional)
cd website && npm install && npm run dev

# 6. Generate OG images for social sharing
cd website && node scripts/gen-og-images.mjs
```

### Common test commands

| Command | What it does |
|---|---|
| `make all` | Full CI pipeline: lint + vet + test + build |
| `go test ./... -cover` | All Go tests with coverage report |
| `go test ./cmd/... -cover` | CLI smoke tests only |
| `go test ./internal/godocgen/... -cover` | Godocgen package only |
| `cd website && npm test` | Website perf-budget check |
| `cd website && npx tsc --noEmit` | TypeScript type check |


## Coding Standards

- **Go**: Follow `aicoding_docs/docs/standards/languages/go.md` — `internal/` + `pkg/` + `cmd/` layout, lowercase package names, explicit error handling.
- **TypeScript**: Follow `aicoding_docs/docs/standards/languages/typescript.md` — strict mode, no `any`, explicit types.
- **Tests**: Follow `aicoding_docs/docs/standards/testing/testing.md` — TDD, test pyramid, ≥80% Go coverage (core ≥90%).

## Quality Gates (do not silently lower)

- Go coverage ≥ 80% (coordinator/scheduler/router ≥ 90%)
- Lint: 0 new warnings (golangci-lint + eslint)
- Security: 0 high-severity vulnerabilities (govulncheck + npm audit)
- No hardcoded API keys in Coordinator or Agent code
- See [docs/CONSTRAINTS.md](docs/CONSTRAINTS.md) for full thresholds

## Git Workflow

- Branch from `main`, name as `feat/{scope}`, `fix/{scope}`, or `docs/{scope}`
- Commit message: `type(scope): description` (e.g., `feat(coordinator): wire RAG into Handle`)
- Squash-merge to `main` after CI passes
- Do not push directly to `main` without review

## Pull Request Checklist

- [ ] Spec written to `docs/spec/` (if new feature)
- [ ] Plan written to `docs/plan/` (if multi-step)
- [ ] Tests written and passing (`go test ./... -count=1`)
- [ ] Lint passes (`golangci-lint run ./...`)
- [ ] Build passes (`go build ./...`)
- [ ] No quality gate thresholds lowered
- [ ] Documentation updated (if behavior changed)
- [ ] No hardcoded secrets or API keys

## Reporting Issues

Use the issue templates in `.github/ISSUE_TEMPLATE/`. Provide:
- Expected vs actual behavior
- Steps to reproduce
- Environment (OS, Go version, backend CLI)
- Logs or error output

## Contributing a New Host Driver

AiCodingAgentTeam supports community-contributed host drivers via the plugin registry. To add a new AI coding CLI driver:

1. **Implement the `runtime.Runtime` interface** (`pkg/runtime/runtime.go`) — all 8 methods: StartSession, DestroySession, SendTask, Capabilities, ModelInfo, Pause, Resume, AuthStatus.

2. **Create a Go package** with an `init()` that self-registers:

   ```go
   package mydriver

   import (
       "github.com/agentcodinglab/aicodingagentteam/pkg/plugin"
       "github.com/agentcodinglab/aicodingagentteam/pkg/runtime"
   )

   func init() {
       plugin.Register(runtime.Backend("my-cli"), &Driver{})
   }
   ```

3. **Write tests** — target >=85% coverage, test Capabilities/AuthStatus/SendTask at minimum.

4. **Users activate your driver** with a blank import:

   ```go
   import _ "github.com/yourorg/mydriver"
   ```

5. **Reference**: see `pkg/plugin/examples/stub-driver/` for a complete working example.

### Rules

- **No API keys** in driver code (ADR-0005). Authentication is delegated to the underlying CLI.
- **Report capabilities honestly** — `HostCapabilities` fields must reflect what the CLI actually supports, never fake `true`.
- **Backend name** must be unique and lowercase-kebab (e.g., `my-cli`).
