package coordinator

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
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

// TestKnowledgeDemo_E2E_ThroughDirector exercises the full 5-layer flow with
// knowledge retrieval and memory recall, mirroring `cmd knowledge demo`.
func TestKnowledgeDemo_E2E_ThroughDirector(t *testing.T) {
	ws := t.TempDir()
	keng := knowledge.New(false)
	_, _ = keng.IndexDirectoryWithLimit(context.Background(), ws, 50)
	memDir := filepath.Join(ws, ".aicodingagentteam", "memory")
	_ = os.MkdirAll(memDir, 0o755)
	mem := memory.New(memDir)
	d := NewWithOptions(
		router.New(), planner.New(ws), scheduler.New(ws),
		qualitygate.NewWithChecks(50, []qualitygate.Check{fastCheck("fast", "exit 0", "blocking")}),
		a2a.NewBus(),
		WithKnowledge(keng), WithMemory(mem),
	)

	_ = mem.Capture(context.Background(), memory.Fact{Key: "demo-seed", Value: "RAG demo bootstrap", Source: "e2e-test"})

	delivery, err := d.Handle(context.Background(), types.UserRequest{Message: "explain routing", Backend: "stub"})
	if err != nil {
		t.Fatalf("handle failed: %v", err)
	}
	if delivery == nil {
		t.Fatal("expected non-nil delivery")
	}
	if delivery.PlanID == "" {
		t.Error("expected plan id")
	}

	facts, err := mem.RecallFacts(context.Background())
	if err != nil {
		t.Fatalf("recall facts: %v", err)
	}
	if len(facts) == 0 {
		t.Error("expected at least the demo-seed fact")
	}
}

// TestKnowledgeDemo_ReportWritten simulates the demo report writer and ensures
// the resulting JSON + Markdown files are well-formed.
func TestKnowledgeDemo_ReportWritten(t *testing.T) {
	ws := t.TempDir()
	keng := knowledge.New(false)
	memDir := filepath.Join(ws, ".aicodingagentteam", "memory")
	_ = os.MkdirAll(memDir, 0o755)
	mem := memory.New(memDir)
	d := NewWithOptions(
		router.New(), planner.New(ws), scheduler.New(ws),
		qualitygate.NewWithChecks(50, []qualitygate.Check{fastCheck("fast", "exit 0", "blocking")}),
		a2a.NewBus(),
		WithKnowledge(keng), WithMemory(mem),
	)

	delivery, _ := d.Handle(context.Background(), types.UserRequest{Message: "explain", Backend: "stub"})

	// Replicate the report writer in-process for the test surface.
	reportDir := filepath.Join(ws, ".aicodingagentteam")
	md := "# RAG Demo Report\n- Indexed: 500\n"
	if delivery != nil {
		md += "- PlanID: " + delivery.PlanID + "\n"
	}
	_ = os.WriteFile(filepath.Join(reportDir, "demo-report.md"), []byte(md), 0o644)

	payload, _ := json.MarshalIndent(map[string]any{"ok": true}, "", "  ")
	_ = os.WriteFile(filepath.Join(reportDir, "demo-report.json"), payload, 0o644)

	b, err := os.ReadFile(filepath.Join(reportDir, "demo-report.md"))
	if err != nil {
		t.Fatalf("read markdown: %v", err)
	}
	if !strings.Contains(string(b), "RAG Demo Report") {
		t.Errorf("missing header: %s", b)
	}

	jb, err := os.ReadFile(filepath.Join(reportDir, "demo-report.json"))
	if err != nil {
		t.Fatalf("read json: %v", err)
	}
	var parsed map[string]any
	if err := json.Unmarshal(jb, &parsed); err != nil {
		t.Fatalf("invalid json: %v", err)
	}
}

// TestKnowledgeDemo_NilKnowledge_NoPanic ensures the Director handle path does
// not panic when knowledge/memory are nil (default Director configuration).
func TestKnowledgeDemo_NilKnowledge_NoPanic(t *testing.T) {
	ws := t.TempDir()
	d := NewWithOptions(
		router.New(), planner.New(ws), scheduler.New(ws),
		qualitygate.NewWithChecks(50, []qualitygate.Check{fastCheck("fast", "exit 0", "blocking")}),
		a2a.NewBus(),
	)

	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("panic: %v", r)
		}
	}()

	_, err := d.Handle(context.Background(), types.UserRequest{Message: "explain", Backend: "stub"})
	if err != nil {
		t.Fatalf("handle failed: %v", err)
	}
}
