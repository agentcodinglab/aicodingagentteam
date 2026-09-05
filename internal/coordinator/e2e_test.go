package coordinator

import (
	"context"
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
	"github.com/agentcodinglab/aicodingagentteam/pkg/api"
)

// TestE2E_QuickEdit_FullPipeline verifies the full 5-layer flow:
// Route -> Plan -> Schedule -> Verify -> Finalize for a quick edit.
func TestE2E_QuickEdit_FullPipeline(t *testing.T) {
	ws := t.TempDir()
	bus := a2a.NewBus()
	r := router.New()
	p := planner.New(ws)
	s := scheduler.NewWithBus(ws, bus)
	g := qualitygate.NewWithChecks(50, []qualitygate.Check{fastCheck("fast", "exit 0", "blocking")})
	memDir := filepath.Join(ws, ".aicodingagentteam", "memory")
	_ = os.MkdirAll(memDir, 0o755)
	keng := knowledge.New(false)
	mem := memory.New(memDir)

	d := NewWithOptions(r, p, s, g, bus, WithKnowledge(keng), WithMemory(mem))

	delivery, err := d.Handle(context.Background(), types.UserRequest{Message: "修改 README", Backend: "codex"})
	if err != nil {
		t.Fatalf("E2E quick edit failed: %v", err)
	}
	if delivery.PlanID != "quick" {
		t.Errorf("expected quick plan, got %s", delivery.PlanID)
	}
	if !delivery.Passed {
		t.Error("expected delivery to pass")
	}
	if delivery.Score != 100 {
		t.Errorf("expected score 100, got %d", delivery.Score)
	}

	// Verify memory was captured
	facts, _ := mem.RecallFacts(context.Background())
	if len(facts) == 0 {
		t.Error("expected facts to be captured after E2E run")
	}

	// Verify via api.Handler interface
	verifyResp, err := d.Verify(context.Background())
	if err != nil {
		t.Fatalf("verify failed: %v", err)
	}
	if !verifyResp.Passed {
		t.Error("verify should pass after successful delivery")
	}
}

// TestE2E_BuildPlan_FullPipeline verifies a build intent produces 9-node DAG.
func TestE2E_BuildPlan_FullPipeline(t *testing.T) {
	ws := t.TempDir()
	bus := a2a.NewBus()
	r := router.New()
	p := planner.New(ws)
	s := scheduler.NewWithBus(ws, bus)
	g := qualitygate.NewWithChecks(50, []qualitygate.Check{fastCheck("fast", "exit 0", "blocking")})
	d := NewWithBus(r, p, s, g, bus)

	delivery, err := d.Handle(context.Background(), types.UserRequest{Message: "build a todo app", Backend: "codex"})
	if err != nil {
		t.Fatalf("E2E build failed: %v", err)
	}
	if delivery.PlanID == "" || delivery.PlanID == "quick" {
		t.Errorf("expected real plan ID, got %s", delivery.PlanID)
	}
	if !delivery.Passed {
		t.Error("expected delivery to pass")
	}
	if len(delivery.Artifacts) == 0 {
		t.Error("expected artifacts from build pipeline")
	}

	// Verify plan.json was persisted
	planPath := filepath.Join(ws, ".aicodingagentteam", "plan.json")
	if _, err := os.Stat(planPath); err != nil {
		t.Errorf("plan.json not persisted: %v", err)
	}

	// Verify plan detail is accessible
	detail, err := d.GetPlanDetail(context.Background())
	if err != nil {
		t.Fatalf("GetPlanDetail failed: %v", err)
	}
	if detail == nil {
		t.Fatal("expected plan detail, got nil")
	}
	if len(detail.Nodes) != 9 {
		t.Errorf("expected 9 nodes, got %d", len(detail.Nodes))
	}
}

// TestE2E_ParkedWorkflow_Continue verifies park + continue flow.
func TestE2E_ParkedWorkflow_Continue(t *testing.T) {
	ws := t.TempDir()
	bus := a2a.NewBus()
	r := router.New()
	p := planner.New(ws)
	s := scheduler.NewWithBus(ws, bus)
	g := qualitygate.NewWithChecks(50, []qualitygate.Check{fastCheck("fast", "exit 0", "blocking")})
	d := NewWithBus(r, p, s, g, bus)

	// Pre-create a write.lock to force park
	lockDir := filepath.Join(ws, ".aicodingagentteam")
	_ = os.MkdirAll(lockDir, 0o755)
	_ = os.WriteFile(filepath.Join(lockDir, "write.lock"), []byte("other|backend|2026-09-03T10:00:00Z"), 0o644)

	delivery, err := d.Handle(context.Background(), types.UserRequest{Message: "build a todo app", Backend: "codex"})
	if err != nil {
		t.Fatalf("handle failed: %v", err)
	}
	if delivery.Passed {
		t.Error("parked delivery should report Passed=false")
	}

	// Remove lock and continue
	_ = os.Remove(filepath.Join(lockDir, "write.lock"))

	// Save a parked state
	state := &planner.WorkflowState{
		PlanID:  delivery.PlanID,
		Status:  "parked",
		Updated: planner.WorkflowState{}.Updated,
	}
	if err := p.SaveState(state); err != nil {
		t.Fatalf("save state: %v", err)
	}

	resumed, msg, err := d.ContinuePlan(context.Background(), delivery.PlanID)
	if err != nil {
		t.Fatalf("continue failed: %v", err)
	}
	if !resumed {
		t.Errorf("expected resume to succeed, got: %s", msg)
	}
	if !strings.Contains(msg, "resumed") {
		t.Errorf("unexpected message: %s", msg)
	}
}

// TestE2E_RunPipeline_API verifies the gRPC API surface works end-to-end.
func TestE2E_RunPipeline_API(t *testing.T) {
	ws := t.TempDir()
	bus := a2a.NewBus()
	r := router.New()
	p := planner.New(ws)
	s := scheduler.NewWithBus(ws, bus)
	g := qualitygate.NewWithChecks(50, []qualitygate.Check{fastCheck("fast", "exit 0", "blocking")})
	d := NewWithBus(r, p, s, g, bus)

	// Quick edit via API
	quickResp, err := d.QuickEdit(context.Background(), api.QuickRequest{Description: "fix typo", Backend: "codex"})
	if err != nil {
		t.Fatalf("QuickEdit API failed: %v", err)
	}
	if !quickResp.Passed || quickResp.Score != 100 {
		t.Errorf("unexpected quick response: passed=%v score=%d", quickResp.Passed, quickResp.Score)
	}

	// Build via API
	runResp, err := d.RunPipeline(context.Background(), api.RunRequest{Requirement: "build a web app", Backend: "codex"})
	if err != nil {
		t.Fatalf("RunPipeline API failed: %v", err)
	}
	if runResp.PlanID == "" || runResp.PlanID == "quick" {
		t.Errorf("expected real plan ID from build, got %s", runResp.PlanID)
	}
	if !runResp.Passed {
		t.Error("expected build to pass")
	}

	// Get agent cards
	cards := d.GetAgentCards(context.Background())
	if len(cards) != 5 {
		t.Errorf("expected 5 agent cards, got %d", len(cards))
	}

	// Get plan
	planResp, err := d.GetPlan(context.Background())
	if err != nil {
		t.Fatalf("GetPlan failed: %v", err)
	}
	if planResp.PlanJSON != "{}" {
		t.Errorf("expected empty plan JSON, got %s", planResp.PlanJSON)
	}
}
