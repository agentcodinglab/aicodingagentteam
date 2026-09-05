# Quality Constraints

The project's red lines. These are not negotiable. Lowering a threshold requires an ADR and explicit user confirmation.

## Coverage

| Package | Minimum |
|---|---|
| `internal/coordinator` | 90% |
| `internal/scheduler` | 90% |
| `internal/router` | 90% |
| `internal/qualitygate` | 90% |
| `internal/host/codex` | 90% |
| `internal/host/opencode` | 90% |
| All other packages | 80% |

## Lint

- `golangci-lint run ./...` exits 0
- Zero new warnings introduced per PR
- `nolint:` directives require justification in the PR description

## Tests

- All unit tests pass with `-race`
- All integration tests pass with `-count=1`
- End-to-end tests cover the full pipeline

## Build

- `go build ./...` exits 0
- `go test ./...` exits 0
- CI total runtime ≤ 5 minutes

## Security

- `govulncheck ./...` reports 0 high-severity findings
- `npm audit` (for TUI) reports 0 high-severity findings
- `gitleaks detect` reports 0 real secrets (test fixtures are allow-listed)

## Anti-patterns

The following actions are forbidden:

- Silencing a failing check to make CI green
- Skipping tests to make a deadline
- Removing assertions to make code "pass"
- Lowering a threshold without an ADR

See [Domain Model](/docs/domain-model) for the project's ubiquitous language.