package opencode

import (
	"context"
	"testing"

	"github.com/yourorg/aicodingagentteam/pkg/runtime"
)

func TestCapabilities(t *testing.T) {
	d := New()
	caps := d.Capabilities()
	if caps.SessionResume {
		t.Error("opencode should not support session resume")
	}
	if !caps.ToolCalls {
		t.Error("opencode should support tool calls")
	}
}

func TestModelInfo(t *testing.T) {
	d := New()
	info := d.ModelInfo()
	if info.ID != "opencode" {
		t.Errorf("expected model ID opencode, got %s", info.ID)
	}
	if info.Provider != "opencode" {
		t.Errorf("expected provider opencode, got %s", info.Provider)
	}
}

func TestResumeReturnsError(t *testing.T) {
	d := New()
	err := d.Resume(context.Background(), "test")
	if err == nil {
		t.Error("opencode Resume should return error")
	}
}

func TestAuthStatusWhenBinaryAvailable(t *testing.T) {
	d := New()
	status, err := d.AuthStatus(context.Background(), "")
	if err != nil {
		t.Fatalf("AuthStatus error: %v", err)
	}
	if !status.Ready {
		t.Errorf("opencode should be ready, got: %s", status.Detail)
	}
}

func TestAuthStatusWhenBinaryMissing(t *testing.T) {
	d := New(WithBinary("nonexistent-binary-12345"))
	status, err := d.AuthStatus(context.Background(), "")
	if err != nil {
		t.Fatalf("AuthStatus error: %v", err)
	}
	if status.Ready {
		t.Error("should report not ready for missing binary")
	}
}

func TestFilterStderr(t *testing.T) {
	input := "downloading model\nplugin loaded\nwarning: deprecated\nreal error: connection refused"
	out := filterStderr(input)
	if contains(out, "downloading") {
		t.Error("should filter 'downloading' lines")
	}
	if contains(out, "plugin") {
		t.Error("should filter 'plugin' lines")
	}
	if contains(out, "warning:") {
		t.Error("should filter 'warning:' lines")
	}
	if !contains(out, "real error") {
		t.Error("should keep real error lines")
	}
}

func TestIsTransient(t *testing.T) {
	if !isTransient("503 Service unavailable", nil) {
		t.Error("503 should be transient")
	}
	if !isTransient("connection refused", nil) {
		t.Error("connection refused should be transient")
	}
	if isTransient("syntax error", nil) {
		t.Error("syntax error should not be transient")
	}
}

func TestWithOptions(t *testing.T) {
	d := New(WithBinary("/custom/opencode"), WithTimeout(120), WithMaxRetries(5))
	if d.binary != "/custom/opencode" {
		t.Errorf("expected /custom/opencode, got %s", d.binary)
	}
	if d.timeout != 120 {
		t.Errorf("expected 120, got %d", d.timeout)
	}
	if d.maxRetries != 5 {
		t.Errorf("expected 5, got %d", d.maxRetries)
	}
}

func TestStartAndDestroySession(t *testing.T) {
	d := New()
	ctx := context.Background()
	id, err := d.StartSession(ctx, runtime.SessionOpts{})
	if err != nil {
		t.Fatalf("StartSession error: %v", err)
	}
	if id != "opencode" {
		t.Errorf("expected session id opencode, got %s", id)
	}
	if err := d.DestroySession(ctx, id); err != nil {
		t.Fatalf("DestroySession error: %v", err)
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || (len(s) > len(substr) && indexString(s, substr) >= 0))
}

func indexString(s, substr string) int {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return i
		}
	}
	return -1
}
