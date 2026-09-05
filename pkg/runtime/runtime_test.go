package runtime

import (
	"context"
	"testing"
)

func TestBackendConstants(t *testing.T) {
	backends := map[Backend]string{
		BackendClaudeCode: "claude-code",
		BackendCodex:      "codex",
		BackendOpenCode:   "opencode",
		BackendDSH:        "deepseek-dsh",
	}
	for b, want := range backends {
		if string(b) != want {
			t.Errorf("backend %q expected %q, got %q", b, want, string(b))
		}
	}
}

func TestEventTypes(t *testing.T) {
	cases := []struct {
		got  EventType
		want string
	}{
		{EventStart, "start"},
		{EventMessage, "message"},
		{EventToolCall, "tool_call"},
		{EventDone, "done"},
		{EventError, "error"},
	}
	for _, c := range cases {
		if string(c.got) != c.want {
			t.Errorf("event type %q expected %q", c.got, c.want)
		}
	}
}

// TestRuntimeInterfaceCompiles ensures the Runtime interface is satisfiable
// and its method signatures remain stable (compile-time check).
type stubRuntime struct{}

func (stubRuntime) StartSession(ctx context.Context, opts SessionOpts) (SessionID, error) {
	return "", nil
}
func (stubRuntime) DestroySession(ctx context.Context, id SessionID) error { return nil }
func (stubRuntime) SendTask(ctx context.Context, id SessionID, task TaskPayload) (<-chan Event, error) {
	return nil, nil
}
func (stubRuntime) Capabilities() HostCapabilities                 { return HostCapabilities{} }
func (stubRuntime) ModelInfo() ModelInfo                           { return ModelInfo{} }
func (stubRuntime) Pause(ctx context.Context, id SessionID) error  { return nil }
func (stubRuntime) Resume(ctx context.Context, id SessionID) error { return nil }
func (stubRuntime) AuthStatus(ctx context.Context, id SessionID) (AuthStatus, error) {
	return AuthStatus{}, nil
}

var _ Runtime = stubRuntime{}

func TestRuntimeInterfaceImplemented(t *testing.T) {
	var r Runtime = stubRuntime{}
	if _, err := r.AuthStatus(context.TODO(), ""); err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}
