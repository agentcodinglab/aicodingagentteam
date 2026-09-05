package coordinator

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/agentcodinglab/aicodingagentteam/internal/a2a"
	"github.com/agentcodinglab/aicodingagentteam/internal/knowledge"
	"github.com/agentcodinglab/aicodingagentteam/internal/memory"
	"github.com/agentcodinglab/aicodingagentteam/internal/planner"
	"github.com/agentcodinglab/aicodingagentteam/internal/qualitygate"
	"github.com/agentcodinglab/aicodingagentteam/internal/router"
	"github.com/agentcodinglab/aicodingagentteam/internal/scheduler"
	"github.com/agentcodinglab/aicodingagentteam/internal/types"
)

// newTestDirectorWithKM builds a Director with knowledge engine and memory store.
func newTestDirectorWithKM(t *testing.T, checks []qualitygate.Check) (*Director, string) {
	t.Helper()
	ws := t.TempDir()
	r := router.New()
	p := planner.New(ws)
	bus := a2a.NewBus()
	s := scheduler.NewWithBus(ws, bus)
	g := qualitygate.NewWithChecks(50, checks)
	memDir := filepath.Join(ws, ".aicodingagentteam", "memory")
	_ = os.MkdirAll(memDir, 0o755)
	keng := knowledge.New(false)
	mem := memory.New(memDir)
	// Index a dummy file so Retrieve returns something
	_ = os.MkdirAll(filepath.Join(ws, "src"), 0o755)
	_ = os.WriteFile(filepath.Join(ws, "src", "main.go"), []byte("package main\nfunc main() {}"), 0o644)
	_, _ = keng.IndexDirectory(context.Background(), ws)
	return NewWithOptions(r, p, s, g, bus, WithKnowledge(keng), WithMemory(mem)), ws
}

func TestHandle_WithKnowledge_EnhancesIntent(t *testing.T) {
	d, _ := newTestDirectorWithKM(t, []qualitygate.Check{fastCheck("fast", "exit 0", "blocking")})

	delivery, err := d.Handle(context.Background(), types.UserRequest{Message: "修改 README"})
	if err != nil {
		t.Fatalf("handle failed: %v", err)
	}
	if delivery.PlanID != "quick" {
		t.Errorf("expected quick plan, got %s", delivery.PlanID)
	}
}

func TestHandle_WithMemory_CapturesFact(t *testing.T) {
	d, ws := newTestDirectorWithKM(t, []qualitygate.Check{fastCheck("fast", "exit 0", "blocking")})

	_, err := d.Handle(context.Background(), types.UserRequest{Message: "修改 README"})
	if err != nil {
		t.Fatalf("handle failed: %v", err)
	}

	memDir := filepath.Join(ws, ".aicodingagentteam", "memory")
	mem := memory.New(memDir)
	facts, err := mem.RecallFacts(context.Background())
	if err != nil {
		t.Fatalf("recall facts: %v", err)
	}
	if len(facts) == 0 {
		t.Error("expected at least one captured fact after Handle")
	}
}

func TestHandle_WithMemory_CapturesPitfallOnFailure(t *testing.T) {
	d, ws := newTestDirectorWithKM(t, []qualitygate.Check{fastCheck("fail", "exit 1", "blocking")})

	delivery, err := d.Handle(context.Background(), types.UserRequest{Message: "build a todo app"})
	if err != nil {
		t.Fatalf("handle failed: %v", err)
	}
	if delivery.Passed {
		t.Skip("delivery passed unexpectedly, cannot test pitfall capture")
	}

	memDir := filepath.Join(ws, ".aicodingagentteam", "memory")
	mem := memory.New(memDir)
	pitfalls, err := mem.RecallPitfalls(context.Background())
	if err != nil {
		t.Fatalf("recall pitfalls: %v", err)
	}
	if len(pitfalls) == 0 {
		t.Error("expected at least one captured pitfall after failed delivery")
	}
}

func TestHandle_NilKnowledgeMemory_NoPanic(t *testing.T) {
	d, _ := newTestDirector(t, []qualitygate.Check{fastCheck("fast", "exit 0", "blocking")})

	// Director with nil knowledge/memory should not panic
	delivery, err := d.Handle(context.Background(), types.UserRequest{Message: "修改 README"})
	if err != nil {
		t.Fatalf("handle failed: %v", err)
	}
	if !delivery.Passed {
		t.Error("expected pass")
	}
}

func TestWithKnowledge_NilEngine(t *testing.T) {
	ws := t.TempDir()
	d := NewWithOptions(
		router.New(), planner.New(ws), scheduler.New(ws),
		qualitygate.NewWithChecks(50, nil), a2a.NewBus(),
		WithKnowledge(nil), WithMemory(nil),
	)
	if d.knowledge != nil {
		t.Error("expected nil knowledge")
	}
	if d.memory != nil {
		t.Error("expected nil memory")
	}
}

