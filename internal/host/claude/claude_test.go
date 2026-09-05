package claude

import (
	"context"
	"testing"

	"github.com/agentcodinglab/aicodingagentteam/pkg/runtime"
)

func TestModelInfo(t *testing.T) {
	d := New()
	info := d.ModelInfo()
	if info.ID != "claude-code" {
		t.Errorf("expected model ID claude-code, got %s", info.ID)
	}
	if info.Provider != "anthropic" {
		t.Errorf("expected provider anthropic, got %s", info.Provider)
	}
	if info.Context != 200000 {
		t.Errorf("expected context 200000, got %d", info.Context)
	}
}

func TestCapabilities(t *testing.T) {
	d := New()
	caps := d.Capabilities()
	if !caps.SessionResume {
		t.Error("claude should support session resume")
	}
	if !caps.ToolCalls {
		t.Error("claude should support tool calls")
	}
	if !caps.WebSearch {
		t.Error("claude should support web search")
	}
	if !caps.WriteHook {
		t.Error("claude should support write hook")
	}
}

func TestStartAndDestroySession(t *testing.T) {
	d := New()
	ctx := context.Background()
	id, err := d.StartSession(ctx, runtime.SessionOpts{})
	if err != nil {
		t.Fatalf("StartSession error: %v", err)
	}
	if id != "claude-session" {
		t.Errorf("expected session id claude-session, got %s", id)
	}
	if err := d.DestroySession(ctx, id); err != nil {
		t.Fatalf("DestroySession error: %v", err)
	}
}

func TestSendTask(t *testing.T) {
	d := New()
	ctx := context.Background()
	ch, err := d.SendTask(ctx, "claude-session", runtime.TaskPayload{Instruction: "test"})
	if err != nil {
		t.Fatalf("SendTask failed: %v", err)
	}
	var gotDone bool
	for ev := range ch {
		if ev.Type == runtime.EventDone {
			gotDone = true
			if ev.Content == "" {
				t.Error("expected non-empty content in done event")
			}
		}
	}
	if !gotDone {
		t.Error("expected done event")
	}
}

func TestAuthStatus(t *testing.T) {
	d := New()
	status, err := d.AuthStatus(context.Background(), "")
	if err != nil {
		t.Fatalf("AuthStatus error: %v", err)
	}
	if !status.Ready {
		t.Error("stub should report ready=true")
	}
	if status.Detail != "logged in" {
		t.Errorf("expected detail 'logged in', got %s", status.Detail)
	}
}

func TestPauseAndResume(t *testing.T) {
	d := New()
	ctx := context.Background()
	if err := d.Pause(ctx, "claude-session"); err != nil {
		t.Errorf("Pause error: %v", err)
	}
	if err := d.Resume(ctx, "claude-session"); err != nil {
		t.Errorf("Resume error: %v", err)
	}
}
