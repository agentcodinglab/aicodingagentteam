# Plugin Development Guide

> This guide covers how to write, register, discover, and install community-contributed host drivers for AiCodingAgentTeam.

## Overview

AiCodingAgentTeam supports a plugin ecosystem via `pkg/plugin`. Each plugin is a Go package that implements the `runtime.Runtime` interface and self-registers at `init()` time via a blank import.

## Writing a Driver

### 1. Implement `runtime.Runtime`

Your driver must implement all methods of the `runtime.Runtime` interface defined in `pkg/runtime/runtime.go`:

```go
type Runtime interface {
    StartSession(ctx context.Context, opts SessionOpts) (SessionID, error)
    DestroySession(ctx context.Context, id SessionID) error
    SendTask(ctx context.Context, id SessionID, task TaskPayload) (<-chan Event, error)
    Capabilities() HostCapabilities
    ModelInfo() ModelInfo
    Pause(ctx context.Context, id SessionID) error
    Resume(ctx context.Context, id SessionID) error
    AuthStatus(ctx context.Context, id SessionID) (AuthStatus, error)
}
```

### 2. Register at `init()`

```go
package mydriver

import (
    "github.com/agentcodinglab/aicodingagentteam/pkg/plugin"
    "github.com/agentcodinglab/aicodingagentteam/pkg/runtime"
)

const Backend runtime.Backend = "my-backend"

func init() {
    plugin.Register(Backend, &MyDriver{})
}
```

### 3. Users activate via blank import

```go
import _ "github.com/yourorg/my-driver"
```

After this import, `host.NewRegistry()` will include your driver.

## Scaffolding

Use the CLI to scaffold a new driver:

```bash
$ aicodingagentteam plugin new my-driver my-backend
Created driver package: pkg/plugin/examples/my-driver
Backend: my-backend

To activate, add a blank import:

  import _ "github.com/agentcodinglab/aicodingagentteam/pkg/plugin/examples/my-driver"
```

## Discovery

### Local Index

Plugins can be indexed locally in `.aicodingagentteam/plugins.json`:

```bash
$ aicodingagentteam plugin search gemini
  gemini-cli (backend=gemini, v0.1.0) -- Gemini CLI driver

$ aicodingagentteam plugin list
  gemini-cli (backend=gemini, v0.1.0) -- Gemini CLI driver
```

### Remote Index (PoC)

`pkg/plugin/remote_index.go` provides a `RemoteIndex` that fetches a JSON array of `PluginMeta` from an HTTP endpoint. Remote errors are fail-open: if the remote is unavailable, local results are returned without error.

```go
remote := plugin.NewRemoteIndex("https://example.com/plugins.json")
results, _ := plugin.SearchMerged(localMarketplace, remote, "query")
```

## Example Drivers

| Driver | Backend | Model | Real exec |
|---|---|---|---|
| stub-driver | stub-example | stub-model | No |
| gemini-driver | gemini | gemini-2.0-flash | Checks `exec.LookPath` |
| qwen-driver | qwen | qwen-2.5-coder | Checks `exec.LookPath` |

## Security Constraints

- **No API keys**: Drivers must NOT hold any API key or credential (ADR-0005). Authentication is delegated to the underlying CLI.
- **No fake capabilities**: `Capabilities()` and `ModelInfo()` must report real values. Missing capabilities must be reported as `false`, never faked.
- **Local-first**: Code and data stay inside the container. Cloud features require explicit opt-in.

## Testing

- Register/unregister tests should use `plugin.Clear()` in `t.Cleanup` to avoid cross-test pollution.
- Marketplace tests use `t.TempDir()` for isolation.
- Remote index tests use `httptest.NewServer` for mock endpoints.