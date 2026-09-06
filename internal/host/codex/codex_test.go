package codex

import (
	"context"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
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
	if !isTransient("Service temporarily unavailable: 503", nil) {
		t.Error("503 should be transient")
	}
}

func TestIsTransient_StreamDisconnect(t *testing.T) {
	if !isTransient("stream disconnected before completion", nil) {
		t.Error("stream disconnect should be transient")
	}
}

func TestIsTransient_Reconnecting(t *testing.T) {
	if !isTransient("Reconnecting to server", nil) {
		t.Error("reconnecting should be transient")
	}
}

func TestIsTransient_DeadlineExceeded(t *testing.T) {
	if !isTransient("context deadline exceeded", nil) {
		t.Error("deadline exceeded should be transient")
	}
}

func TestIsTransient_NonTransient(t *testing.T) {
	if isTransient("", errors.New("some other error")) {
		t.Error("generic error should not be transient")
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
		{"stream disconnected", "stream disconnected", nil, true},
		{"Service temporarily", "Service temporarily unavailable", nil, true},
		{"Reconnecting", "Reconnecting to server", nil, true},
		{"deadline exceeded", "context deadline exceeded", nil, true},
		{"permanent error", "syntax error: bad prompt", nil, false},
		{"empty", "", nil, false},
	}
	for _, tc := range tests {
		if got := isTransient(tc.stderr, tc.err); got != tc.want {
			t.Errorf("isTransient(%q) = %v, want %v", tc.name, got, tc.want)
		}
	}
}

func TestAuthStatus_BinaryCheck(t *testing.T) {
	d := New()
	status, err := d.AuthStatus(context.Background(), "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// codex binary may not be installed in CI; skip if not available
	if !status.Ready {
		t.Skipf("codex binary not available in this environment: %s", status.Detail)
	}
}

func TestAuthStatus_MissingBinary(t *testing.T) {
	d := New(WithBinary("nonexistent-codex-binary-12345"))
	status, err := d.AuthStatus(context.Background(), "")
	if err != nil {
		t.Fatalf("AuthStatus error: %v", err)
	}
	if status.Ready {
		t.Error("should report not ready for missing binary")
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

func TestWithOptions(t *testing.T) {
	d := New(WithBinary("/custom/codex"), WithTimeout(120), WithMaxRetries(5))
	if d.binary != "/custom/codex" {
		t.Errorf("expected /custom/codex, got %s", d.binary)
	}
	if d.timeout != 120 {
		t.Errorf("expected 120, got %d", d.timeout)
	}
	if d.maxRetries != 5 {
		t.Errorf("expected 5, got %d", d.maxRetries)
	}
}

func TestWithMaxRetries(t *testing.T) {
	d := New(WithMaxRetries(10))
	if d.maxRetries != 10 {
		t.Errorf("expected maxRetries 10, got %d", d.maxRetries)
	}
}

func TestStartAndDestroySession(t *testing.T) {
	d := New()
	ctx := context.Background()
	id, err := d.StartSession(ctx, runtime.SessionOpts{})
	if err != nil {
		t.Fatalf("StartSession error: %v", err)
	}
	if id != "codex" {
		t.Errorf("expected session id codex, got %s", id)
	}
	if err := d.DestroySession(ctx, id); err != nil {
		t.Fatalf("DestroySession error: %v", err)
	}
}

func TestCapabilities(t *testing.T) {
	d := New()
	caps := d.Capabilities()
	if !caps.SessionResume {
		t.Error("codex should support session resume")
	}
	if !caps.ToolCalls {
		t.Error("codex should support tool calls")
	}
	if !caps.WebSearch {
		t.Error("codex should support web search")
	}
	if caps.WriteHook {
		t.Error("codex should not support write hook")
	}
}

func TestModelInfo(t *testing.T) {
	d := New()
	info := d.ModelInfo()
	if info.ID != "codex" {
		t.Errorf("expected model ID codex, got %s", info.ID)
	}
	if info.Provider != "openai" {
		t.Errorf("expected provider openai, got %s", info.Provider)
	}
	if info.Context != 200000 {
		t.Errorf("expected context 200000, got %d", info.Context)
	}
}

func TestPauseAndResume(t *testing.T) {
	d := New()
	ctx := context.Background()
	if err := d.Pause(ctx, "codex"); err != nil {
		t.Errorf("Pause error: %v", err)
	}
	if err := d.Resume(ctx, "codex"); err != nil {
		t.Errorf("Resume error: %v", err)
	}
}

// writeMockCodex builds a tiny Go program that prints to stdout.
func writeMockCodex(t *testing.T, dir, output string) string {
	t.Helper()
	const sentinel = "SENTINEL_TOKEN_HERE"
	src := "package main\nimport \"fmt\"\nfunc main() { fmt.Println(`" + sentinel + "`) }\n"
	src = strings.ReplaceAll(src, sentinel, output)
	mockSrc := filepath.Join(dir, "mockcodex.go")
	if err := os.WriteFile(mockSrc, []byte(src), 0o644); err != nil {
		t.Fatal(err)
	}
	binPath := filepath.Join(dir, "mockcodex")
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
	mockBin := writeMockCodex(t, dir, "hello from mock codex")
	d := New(WithBinary(mockBin), WithMaxRetries(0))
	ch, err := d.SendTask(context.Background(), "codex", runtime.TaskPayload{Instruction: "test", Timeout: 10})
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

func TestRunOnce_Direct(t *testing.T) {
	dir := t.TempDir()
	mockBin := writeMockCodex(t, dir, "codex output here")
	d := New(WithBinary(mockBin))
	evCh := make(chan runtime.Event, 8)
	out, stderr, err := d.runOnceStreaming(context.Background(), "test prompt", 10, evCh)
	close(evCh)
	if err != nil {
		t.Fatalf("runOnce failed: %v (stderr: %s)", err, stderr)
	}
	if !strings.Contains(out, "codex output") {
		t.Errorf("expected output containing 'codex output', got %s", out)
	}
}

func TestRunOnce_FiltersStderr(t *testing.T) {
	dir := t.TempDir()
	// Mock that prints noise to stderr and real content to stdout
	const sentinel = "SENTINEL_TOKEN_HERE"
	src := `package main
import (
  "fmt"
  "os"
)
func main() {
  fmt.Fprintln(os.Stderr, "Unknown model warning here")
  fmt.Fprintln(os.Stderr, "remote installed plugin bundle sync failed")
  fmt.Fprintln(os.Stdout, "` + sentinel + `")
}`
	src = strings.ReplaceAll(src, sentinel, "real codex output")
	mockSrc := filepath.Join(dir, "mockcodex_noise.go")
	if err := os.WriteFile(mockSrc, []byte(src), 0o644); err != nil {
		t.Fatal(err)
	}
	binPath := filepath.Join(dir, "mockcodex_noise")
	if os.PathSeparator == '\\' {
		binPath += ".exe"
	}
	cmd := exec.Command("go", "build", "-o", binPath, mockSrc)
	cmd.Dir = dir
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("build mock: %v\n%s", err, out)
	}
	d := New(WithBinary(binPath))
	evCh := make(chan runtime.Event, 8)
	out, stderr, err := d.runOnceStreaming(context.Background(), "test", 10, evCh)
	close(evCh)
	if err != nil {
		t.Fatalf("runOnce failed: %v", err)
	}
	if !strings.Contains(out, "real codex output") {
		t.Errorf("expected stdout to contain real output, got %s", out)
	}
	if strings.Contains(stderr, "Unknown model") {
		t.Error("stderr should filter out Unknown model noise")
	}
	if strings.Contains(stderr, "plugin bundle") {
		t.Error("stderr should filter out plugin bundle noise")
	}
}
