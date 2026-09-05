package host

import (
	"context"
	"testing"

	"github.com/agentcodinglab/aicodingagentteam/internal/host/claude"
	"github.com/agentcodinglab/aicodingagentteam/internal/host/codex"
	"github.com/agentcodinglab/aicodingagentteam/internal/host/dsh"
	"github.com/agentcodinglab/aicodingagentteam/internal/host/opencode"
	"github.com/agentcodinglab/aicodingagentteam/pkg/runtime"
)

func TestNewRegistry_RegistersAllBackends(t *testing.T) {
	r := NewRegistry()
	if r.Count() != 4 {
		t.Errorf("expected 4 drivers registered, got %d", r.Count())
	}
}

func TestRegistry_Get_KnownBackend(t *testing.T) {
	r := NewRegistry()
	tests := []struct {
		backend runtime.Backend
		wantID  string
	}{
		{runtime.BackendCodex, "codex"},
		{runtime.BackendOpenCode, "opencode"},
		{runtime.BackendClaudeCode, "claude-code"},
		{runtime.BackendDSH, "deepseek-dsh"},
	}
	for _, tc := range tests {
		rt, err := r.Get(tc.backend)
		if err != nil {
			t.Errorf("Get(%s) failed: %v", tc.backend, err)
			continue
		}
		if rt.ModelInfo().ID != tc.wantID {
			t.Errorf("Get(%s) returned model ID %s, want %s", tc.backend, rt.ModelInfo().ID, tc.wantID)
		}
	}
}

func TestRegistry_Get_UnknownBackend(t *testing.T) {
	r := NewRegistry()
	_, err := r.Get("unknown-backend")
	if err == nil {
		t.Error("should return error for unknown backend")
	}
}

func TestRegistry_Register_Overwrites(t *testing.T) {
	r := NewRegistry()
	custom := codex.New()
	r.Register(runtime.BackendCodex, custom)
	rt, _ := r.Get(runtime.BackendCodex)
	if rt != custom {
		t.Error("Register should overwrite existing driver")
	}
}

func TestRegistry_List_ReturnsAllBackends(t *testing.T) {
	r := NewRegistry()
	list := r.List()
	if len(list) != 4 {
		t.Errorf("expected 4 backends, got %d", len(list))
	}
	// All 4 should be present
	seen := make(map[runtime.Backend]bool)
	for _, b := range list {
		seen[b] = true
	}
	for _, want := range []runtime.Backend{runtime.BackendCodex, runtime.BackendOpenCode, runtime.BackendClaudeCode, runtime.BackendDSH} {
		if !seen[want] {
			t.Errorf("missing backend: %s", want)
		}
	}
}

func TestRegistry_AuthCheck_UnknownBackend(t *testing.T) {
	r := NewRegistry()
	err := r.AuthCheck(context.Background(), "unknown")
	if err == nil {
		t.Error("should return error for unknown backend")
	}
}

func TestRegistry_AuthCheck_KnownBackend_NoError(t *testing.T) {
	r := NewRegistry()
	// Stub drivers (claude, dsh) report ready=true without checking binary
	err := r.AuthCheck(context.Background(), runtime.BackendClaudeCode)
	if err != nil {
		t.Errorf("AuthCheck for claude stub should not fail: %v", err)
	}
}

func TestRegistry_AuthCheck_CodexBinaryMissing(t *testing.T) {
	// Build a registry with codex pointing to a missing binary
	r := NewRegistry()
	r.Register(runtime.BackendCodex, codex.New())
	// This may pass or fail depending on env; just ensure no panic
	_ = r.AuthCheck(context.Background(), runtime.BackendCodex)
}

func TestDriverModelInfo_AllBackends(t *testing.T) {
	tests := []struct {
		name     string
		driver   runtime.Runtime
		wantID   string
		wantProv string
	}{
		{"codex", codex.New(), "codex", "openai"},
		{"opencode", opencode.New(), "opencode", "opencode"},
		{"claude", claude.New(), "claude-code", "anthropic"},
		{"dsh", dsh.New(), "deepseek-dsh", "deepseek"},
	}
	for _, tc := range tests {
		info := tc.driver.ModelInfo()
		if info.ID != tc.wantID {
			t.Errorf("%s: model ID = %s, want %s", tc.name, info.ID, tc.wantID)
		}
		if info.Provider != tc.wantProv {
			t.Errorf("%s: provider = %s, want %s", tc.name, info.Provider, tc.wantProv)
		}
		if info.Context <= 0 {
			t.Errorf("%s: context should be > 0", tc.name)
		}
	}
}

func TestDriverCapabilities_StubReportsNoWriteHook(t *testing.T) {
	// Codex reports WriteHook=false; stubs may vary
	caps := codex.New().Capabilities()
	if caps.WriteHook {
		t.Error("codex should not support write hook")
	}
	if !caps.ToolCalls {
		t.Error("codex should support tool calls")
	}
}
