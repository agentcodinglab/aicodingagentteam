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