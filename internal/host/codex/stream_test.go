package codex

import (
	"context"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"

	rt "github.com/agentcodinglab/aicodingagentteam/pkg/runtime"
)

// streamStub returns the absolute path to the multi-line codex-stream stub binary.
func streamStub(t *testing.T) string {
	t.Helper()
	name := "codex-stream"
	if runtime.GOOS == "windows" {
		name = "codex-stream.cmd"
	}
	_, thisFile, _, _ := runtime.Caller(0)
	// internal/host/codex -> repo root
	root := filepath.Dir(filepath.Dir(filepath.Dir(filepath.Dir(thisFile))))
	p := filepath.Join(root, "testdata", "stubbin", name)
	return p
}

// TestCodex_SendTask_StreamingEvents verifies that stdout is streamed as multiple
// EventMessage events before EventDone (ADR-0019), using the multi-line stub binary.
func TestCodex_SendTask_StreamingEvents(t *testing.T) {
	d := New(WithBinary(streamStub(t)), WithTimeout(30))
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	ch, err := d.SendTask(ctx, "codex", rt.TaskPayload{
		Instruction: "emit multiple lines",
		Timeout:     30,
	})
	if err != nil {
		t.Fatalf("SendTask failed: %v", err)
	}

	var messages []string
	var doneContent string
	var gotDone bool
	for ev := range ch {
		switch ev.Type {
		case rt.EventMessage:
			messages = append(messages, ev.Content)
		case rt.EventDone:
			doneContent = ev.Content
			gotDone = true
		case rt.EventError:
			t.Fatalf("unexpected error event: %s", ev.Content)
		}
	}

	if !gotDone {
		t.Fatal("expected EventDone")
	}
	if len(messages) < 2 {
		t.Errorf("expected at least 2 streamed messages, got %d (no streaming?)", len(messages))
	}
	for _, m := range messages {
		if !strings.Contains(m, "line") {
			t.Errorf("unexpected message: %s", m)
		}
	}
	// Done content should aggregate the streamed stdout.
	if doneContent == "" {
		t.Error("expected non-empty Done content (aggregated stdout)")
	}
}
