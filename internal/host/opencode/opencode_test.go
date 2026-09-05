package opencode

import (
	"context"
	"testing"

	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/agentcodinglab/aicodingagentteam/pkg/runtime"
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

// writeMockOpencode builds a tiny Go program that prints jsonl to stdout.
const NewLine = "\n"

func writeMockOpencode(t *testing.T, dir, jsonl string) string {
	t.Helper()
	// Use a sentinel to avoid quote escaping issues when jsonl contains ".
	const sentinel = "SENTINEL_TOKEN_HERE"
	src := "package main" + NewLine + "import \"fmt\"" + NewLine + "func main() { fmt.Println(`" + sentinel + "`) }" + NewLine
	src = strings.ReplaceAll(src, sentinel, jsonl)
	mockSrc := filepath.Join(dir, "mockopencode.go")
	if err := os.WriteFile(mockSrc, []byte(src), 0o644); err != nil {
		t.Fatal(err)
	}
	binPath := filepath.Join(dir, "mockopencode")
	if os.PathSeparator == '\\' {
		binPath += ".exe"
	}
	cmd := exec.Command("go", "build", "-o", binPath, mockSrc)
	cmd.Dir = dir
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("build mock: %v\n%s", err, out)
	}
	return binPath
}

func TestSendTask_HappyPath(t *testing.T) {
	dir := t.TempDir()
	// Emit a valid JSONL event that the driver recognizes
	mockBin := writeMockOpencode(t, dir, `{"type":"text","part":{"text":"hello from mock"}}`)
	d := New(WithBinary(mockBin))
	ch, err := d.SendTask(context.Background(), "s1", runtime.TaskPayload{Instruction: "test", Timeout: 10})
	if err != nil {
		t.Fatalf("SendTask: %v", err)
	}
	var got []runtime.Event
	for ev := range ch {
		got = append(got, ev)
	}
	if len(got) == 0 {
		t.Fatal("expected events, got none")
	}
	if got[0].Type != runtime.EventStart {
		t.Errorf("expected first event Start, got %s", got[0].Type)
	}
	if got[len(got)-1].Type != runtime.EventDone {
		t.Errorf("expected last event Done, got %s", got[len(got)-1].Type)
	}
}

func TestIsTransient_AllBranches(t *testing.T) {
	tests := []struct {
		name   string
		stderr string
		err    error
		want   bool
	}{
		{"503", "503 Service Unavailable", nil, true},
		{"deadline exceeded", "context deadline exceeded", nil, true},
		{"connection refused", "connection refused", nil, true},
		{"service temporarily unavailable", "Service temporarily unavailable", nil, true},
		{"ECONNREFUSED", "ECONNREFUSED", nil, true},
		{"permanent error", "syntax error: bad prompt", nil, false},
		{"empty", "", nil, false},
	}
	for _, tc := range tests {
		if got := isTransient(tc.stderr, tc.err); got != tc.want {
			t.Errorf("isTransient(%q) = %v, want %v", tc.name, got, tc.want)
		}
	}
}

func TestRunOnce_Direct(t *testing.T) {
	dir := t.TempDir()
	mockBin := writeMockOpencode(t, dir, `{"type":"text","part":{"text":"hi"}}`)
	d := New(WithBinary(mockBin))
	out, stderr, err := d.runOnce(context.Background(), "test", 10)
	if err != nil {
		t.Fatalf("runOnce failed: %v (stderr: %s)", err, stderr)
	}
	if !strings.Contains(out, "text") {
		t.Errorf("expected JSONL output containing text, got %s", out)
	}
}
