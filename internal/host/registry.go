// Package host provides the host driver registry and 4 CLI driver implementations.
package host

import (
	"context"
	"fmt"

	"github.com/yourorg/aicodingagentteam/internal/host/claude"
	"github.com/yourorg/aicodingagentteam/internal/host/codex"
	"github.com/yourorg/aicodingagentteam/internal/host/dsh"
	"github.com/yourorg/aicodingagentteam/internal/host/opencode"
	"github.com/yourorg/aicodingagentteam/pkg/runtime"
)

// Registry holds all registered host drivers keyed by Backend.
type Registry struct {
	drivers map[runtime.Backend]runtime.Runtime
}

// NewRegistry creates a host Registry with all drivers registered.
func NewRegistry() *Registry {
	r := &Registry{drivers: make(map[runtime.Backend]runtime.Runtime)}
	r.Register(runtime.BackendCodex, codex.New())
	r.Register(runtime.BackendClaudeCode, claude.New())
	r.Register(runtime.BackendOpenCode, opencode.New())
	r.Register(runtime.BackendDSH, dsh.New())
	return r
}

// Register adds or replaces a host driver.
func (r *Registry) Register(b runtime.Backend, rt runtime.Runtime) {
	r.drivers[b] = rt
}

// Get returns the driver for a backend.
func (r *Registry) Get(b runtime.Backend) (runtime.Runtime, error) {
	rt, ok := r.drivers[b]
	if !ok {
		return nil, fmt.Errorf("no driver for backend %s", b)
	}
	return rt, nil
}

// List returns all registered backends.
func (r *Registry) List() []runtime.Backend {
	out := make([]runtime.Backend, 0, len(r.drivers))
	for b := range r.drivers {
		out = append(out, b)
	}
	return out
}

// AuthCheck verifies that a backend is ready before scheduling.
func (r *Registry) AuthCheck(ctx context.Context, b runtime.Backend) error {
	rt, err := r.Get(b)
	if err != nil {
		return err
	}
	status, err := rt.AuthStatus(ctx, "")
	if err != nil {
		return err
	}
	if !status.Ready {
		return fmt.Errorf("backend %s not authenticated: %s", b, status.Detail)
	}
	return nil
}
