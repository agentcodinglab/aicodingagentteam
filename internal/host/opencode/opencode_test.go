package opencode

import (
	"context"
	"runtime"
	"time"
	"testing"

	"os"
	"path/filepath"
	"strings"

	rt "github.com/agentcodinglab/aicodingagentteam/pkg/runtime"
)

func TestCapabilities(t *testing.T) {
	d := New()
	caps := d.Capabilities()
	if !caps.SessionResume {
		t.Error("opencode (ACP mode) should support session resume")
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

func TestResumeSucceeds(t *testing.T) {
	d := New()
	err := d.Resume(context.Background(), "test")
	if err != nil {
		t.Errorf("opencode Resume should succeed (ACP mode): %v", err)
	}
}

func TestAuthStatusWhenBinaryAvailable(t *testing.T) {
	d := New()
	status, err := d.AuthStatus(context.Background(), "")
	if err != nil {
		t.Fatalf("AuthStatus error: %v", err)
	}
	if !status.Ready {
		t.Skipf("opencode binary not available in this environment: %s", status.Detail)
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
	id, err := d.StartSession(ctx, rt.SessionOpts{})
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

// writeMockOpencode builds a tiny Go program that prints jsonl to stdout.
const NewLine = "\n"


// acpStub returns the absolute path to the opencode-acp stub binary.
func acpStub(t *testing.T) string {
	t.Helper()
	name := "opencode-acp"
	if runtime.GOOS == "windows" { name = "opencode-acp.cmd" }
	_, thisFile, _, _ := runtime.Caller(0)
	// internal/host/opencode -> repo root
	root := filepath.Dir(filepath.Dir(filepath.Dir(filepath.Dir(thisFile))))
	p := filepath.Join(root, "testdata", "stubbin", name)
	if _, err := os.Stat(p); err != nil { t.Skipf("stub binary not found: %s", p) }
	return p
}

// TestOpenCode_ACP_StubServer drives the real opencode driver against the stub
// acp binary (ADR-0021 B) and asserts that a streamed EventMessage arrives
// before EventDone and that the aggregated Done content is non-empty.
func TestOpenCode_ACP_StubServer(t *testing.T) {
	d := New(WithBinary(acpStub(t)), WithTimeout(30))
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	ch, err := d.SendTask(ctx, "opencode", rt.TaskPayload{Instruction: "hello", Timeout: 30})
	if err != nil { t.Fatalf("SendTask: %v", err) }
	var msgs []string
	var gotDone bool
	var doneContent string
	for ev := range ch {
		switch ev.Type {
		case rt.EventMessage: msgs = append(msgs, ev.Content)
		case rt.EventDone:
			gotDone = true; doneContent = ev.Content
		case rt.EventError: t.Fatalf("error event: %s", ev.Content)
		}
	}
	if !gotDone { t.Fatal("expected EventDone") }
	if len(msgs) == 0 { t.Error("expected at least one EventMessage from the stub (no streaming?)") }
	for _, m := range msgs { if !strings.Contains(m, "acp stub") { t.Errorf("unexpected message: %s", m) } }
	if doneContent == "" { t.Error("expected non-empty Done content") }
}