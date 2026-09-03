# AiCodingAgentTeam TUI Client

Terminal UI client for the AiCodingAgentTeam orchestration platform, built with [Ink](https://github.com/vadimdemedes/ink) (React for CLIs).

## Quick Start

```bash
cd tui
npm install
npm run build
```

### Demo Mode (no coordinator needed)

```bash
node dist/cli.js --demo
```

Runs the TUI with mock data — useful for testing UI without a running Go backend.

### Live Mode

Start the coordinator first:

```bash
cd ..
go run ./cmd/aicodingagentteam serve
```

Then launch TUI:

```bash
node dist/cli.js
# or specify host/port
AICODINGAGENTTEAM_HOST=127.0.0.1 AICODINGAGENTTEAM_PORT=8080 node dist/cli.js
```

## Commands

| Command | Description |
|---|---|
| `/run <requirement>` | Run full pipeline |
| `/quick <description>` | Quick edit |
| `/verify` | Run quality gate |
| `/plan` | Show DAG plan |
| `/continue` | Resume parked task |
| `/backend <name>` | Switch host CLI backend |
| `/report` | Show quality report |
| `/exit` | Exit TUI |

Press `Ctrl+C` to exit at any time.

## Architecture

```
tui/
├── package.json
├── tsconfig.json
├── .eslintrc.js
├── proto/
│   └── aicodingagentteam.proto      # gRPC service definition (copied from root)
├── src/
│   ├── cli.ts                       # Binary entry, --demo flag, TTY detection
│   ├── app.tsx                      # Root React component, input handling
│   ├── store.ts                     # Zustand global state
│   ├── grpc/
│   │   ├── client.ts                # Real gRPC client (ClientInterface impl)
│   │   └── mock.ts                  # Mock client for demo mode
│   ├── hooks/
│   │   └── useCommands.ts           # Slash command dispatcher
│   └── components/
│       ├── StatusBar.tsx            # Connection + quality score
│       ├── PlanView.tsx             # DAG node list with status icons
│       ├── AuditLog.tsx             # Recent audit entries
│       ├── ResultPanel.tsx          # Score / blocking / artifacts
│       └── HelpPanel.tsx            # Command reference
└── dist/                            # Compiled output
```

## Development

```bash
npm run dev      # tsc --watch
npm run lint     # eslint
npm run build    # tsc (production)
```

## Environment Variables

| Variable | Default | Description |
|---|---|---|
| `AICODINGAGENTTEAM_HOST` | `localhost` | Coordinator gRPC host |
| `AICODINGAGENTTEAM_PORT` | `8080` | Coordinator gRPC port |
| `AICODINGAGENTTEAM_DEMO` | — | Set to `1` to force demo mode |