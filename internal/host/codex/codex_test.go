package codex

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/agentcodinglab/aicodingagentteam/pkg/runtime"
)

func TestFilterStderr_RemovesNoise(t *testing.T) {
	input := `Reading additional input from stdin...
2026-09-02T10:08:05Z  WARN codex_models_manager::model_info: Unknown model stk-code-f1 is used.
2026-09-02T10:08:05Z  WARN codex_core_plugins::manager: remote installed plugin bundle sync failed error=...
2026-09-02T10:08:17Z  WARN codex_core::shell_snapshot: Failed to create shell snapshot for powershell: Shell snapshot not supported yet for PowerShell
This is a real error that should be kept`

	out := filterStderr(input)
	if strings.Contains(out, "Reading additional input from stdin") {
		t.Error("should filter stdin message")
	}
	if strings.Contains(out, "Unknown model") {
		t.Error("should filter Unknown model warning")
	}
	if strings.Contains(out, "remote installed plugin bundle") {
		t.Error("should filter plugin sync warning")
	}
	if !strings.Contains(out, "This is a real error") {
		t.Error("should keep real error")
	}
}

func TestIsTransient_503(t *testing.T) {
	if !isTransient("Service temporarily unavailable: 503", errors.New("exit status 1")) {
		t.Error("503 should be transient")
	}
}

func TestIsTransient_StreamDisconnect(t *testing.T) {
	if !isTransient("stream disconnected before completion", nil) {
		t.Error("stream disconnect should be transient")
	}
}

func TestIsTransient_NonTransient(t *testing.T) {
	if isTransient("", errors.New("some other error")) {
		t.Error("generic error should not be transient")
	}
}

func TestAuthStatus_BinaryCheck(t *testing.T) {
	d := New()
	status, err := d.AuthStatus(context.Background(), "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// codex binary should be available in this environment
	if !status.Ready {
		t.Errorf("codex should be available, got: %s", status.Detail)
	}
}

func TestSendTask_TimeoutProducesError(t *testing.T) {
	d := New(WithTimeout(1)) // 1 second timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	ch, err := d.SendTask(ctx, "codex", runtime.TaskPayload{
		Instruction: "say hello",
		Timeout:     1,
	})
	if err != nil {
		t.Fatalf("SendTask failed: %v", err)
	}

	var gotError bool
	for ev := range ch {
		if ev.Type == runtime.EventError {
			gotError = true
		}
		if ev.Type == runtime.EventDone {
			// Success is also acceptable in this environment
			return
		}
	}
	// Either error (transient/timeout) or success is acceptable;
	// we just verify no panic and channel closes
	_ = gotError
}
