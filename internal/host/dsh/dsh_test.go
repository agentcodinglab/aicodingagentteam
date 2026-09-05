package dsh

import (
	"context"
	"testing"

	"github.com/agentcodinglab/aicodingagentteam/pkg/runtime"
)

func TestModelInfo(t *testing.T) {
	d := New()
	info := d.ModelInfo()
	if info.ID != "deepseek-dsh" {
		t.Errorf("expected model ID deepseek-dsh, got %s", info.ID)
	}
	if info.Provider != "deepseek" {
		t.Errorf("expected provider deepseek, got %s", info.Provider)
	}
	if info.Context != 128000 {
		t.Errorf("expected context 128000, got %d", info.Context)
	}
}

func TestCapabilities(t *testing.T) {
	d := New()
	caps := d.Capabilities()
	if !caps.SessionResume {
		t.Error("dsh should support session resume")
	}
	if !caps.ToolCalls {
		t.Error("dsh should support tool calls")
	}
	if !caps.WebSearch {
		t.Error("dsh should support web search")
	}
	if caps.WriteHook {
		t.Error("dsh should not support write hook")
	}
}

func TestStartAndDestroySession(t *testing.T) {
	d := New()
	ctx := context.Background()
	id, err := d.StartSession(ctx, runtime.SessionOpts{})
	if err != nil {
		t.Fatalf("StartSession error: %v", err)
	}
	if id != "dsh-session" {
		t.Errorf("expected session id dsh-session, got %s", id)
	}
	if err := d.DestroySession(ctx, id); err != nil {
		t.Fatalf("DestroySession error: %v", err)
	}
}

func TestSendTask(t *testing.T) {
	d := New()
	ctx := context.Background()
	ch, err := d.SendTask(ctx, "dsh-session", runtime.TaskPayload{Instruction: "test"})
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
	if err := d.Pause(ctx, "dsh-session"); err != nil {
		t.Errorf("Pause error: %v", err)
	}
	if err := d.Resume(ctx, "dsh-session"); err != nil {
		t.Errorf("Resume error: %v", err)
	}
}
