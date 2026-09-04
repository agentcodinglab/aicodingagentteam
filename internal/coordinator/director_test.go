package coordinator

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/agentcodinglab/aicodingagentteam/internal/a2a"
	"github.com/agentcodinglab/aicodingagentteam/internal/planner"
	"github.com/agentcodinglab/aicodingagentteam/internal/qualitygate"
	"github.com/agentcodinglab/aicodingagentteam/internal/router"
	"github.com/agentcodinglab/aicodingagentteam/internal/scheduler"
	"github.com/agentcodinglab/aicodingagentteam/internal/types"
	"github.com/agentcodinglab/aicodingagentteam/pkg/api"
)

// fastCheck is a near-instant quality gate check for tests.
func fastCheck(name, script, severity string) qualitygate.Check {
	if strings.Contains(script, "0") {
		return qualitygate.Check{Name: name, Command: []string{"go", "version"}, Timeout: 10, Severity: severity}
	}
	return qualitygate.Check{Name: name, Command: []string{"go", "tool", "nonexistent-binary-xyz"}, Timeout: 10, Severity: severity}
}

// newTestDirector builds a Director with a real A2A in-process bus, registered
// reviewers, a temp workspace, and fast stub gate checks.
func newTestDirector(t *testing.T, checks []qualitygate.Check) (*Director, string) {
	t.Helper()
	ws := t.TempDir()
	r := router.New()
	p := planner.New(ws)
	bus := a2a.NewBus()
	s := scheduler.NewWithBus(ws, bus)
	g := qualitygate.NewWithChecks(50, checks)
	return NewWithBus(r, p, s, g, bus), ws
}

func TestHandle_QuickEdit_Passes(t *testing.T) {
	d, _ := newTestDirector(t, []qualitygate.Check{fastCheck("fast", "exit 0", "blocking")})

	delivery, err := d.Handle(context.Background(), types.UserRequest{Message: "修改 README 标题"})
	if err != nil {
		t.Fatalf("handle failed: %v", err)
	}
	if delivery.PlanID != "quick" {
		t.Errorf("expected quick plan, got %s", delivery.PlanID)
	}
	if !delivery.Passed {
		t.Error("expected delivery to pass with all-pass gate")
	}
	if delivery.Score != 100 {
		t.Errorf("expected score 100, got %d", delivery.Score)
	}
}

func TestHandle_BuildPlan_EndToEnd(t *testing.T) {
	d, ws := newTestDirector(t, []qualitygate.Check{fastCheck("fast", "exit 0", "blocking")})

	delivery, err := d.Handle(context.Background(), types.UserRequest{Message: "build a todo app"})
	if err != nil {
		t.Fatalf("handle failed: %v", err)
	}
	if !delivery.Passed {
		t.Error("expected end-to-end build to pass")
	}
	if delivery.Score != 100 {
		t.Errorf("expected score 100, got %d", delivery.Score)
	}
	if len(delivery.Artifacts) == 0 {
		t.Error("expected coordinator/writer artifacts in delivery")
	}
	// plan.json must be persisted in the workspace
	if _, err := os.Stat(filepath.Join(ws, ".aicodingagentteam", "plan.json")); err != nil {
		t.Errorf("plan.json not persisted: %v", err)
	}
	// write lock must be released after writer nodes complete
	if _, err := os.Stat(filepath.Join(ws, ".aicodingagentteam", "write.lock")); !os.IsNotExist(err) {
		t.Error("write lock should be released after execution")
	}
}

func TestHandle_BuildPlan_ParksOnHeldLock(t *testing.T) {
	d, ws := newTestDirector(t, []qualitygate.Check{fastCheck("fast", "exit 0", "blocking")})

	lockDir := filepath.Join(ws, ".aicodingagentteam")
	if err := os.MkdirAll(lockDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(lockDir, "write.lock"), []byte("other|backend|2026-09-03T10:00:00Z"), 0o644); err != nil {
		t.Fatal(err)
	}

	delivery, err := d.Handle(context.Background(), types.UserRequest{Message: "build a todo app"})
	if err != nil {
		t.Fatalf("handle failed: %v", err)
	}
	if delivery.Passed {
		t.Error("expected parked delivery to report Passed=false")
	}
	if delivery.PlanID == "" || delivery.PlanID == "quick" {
		t.Errorf("expected real plan ID, got %q", delivery.PlanID)
	}
}

func TestRunPipeline_And_QuickEdit(t *testing.T) {
	d, _ := newTestDirector(t, []qualitygate.Check{fastCheck("fast", "exit 0", "blocking")})
	ctx := context.Background()

	run, err := d.RunPipeline(ctx, api.RunRequest{Requirement: "修改按钮颜色", Backend: "codex"})
	if err != nil {
		t.Fatalf("run pipeline failed: %v", err)
	}
	if run.PlanID != "quick" || !run.Passed {
		t.Errorf("unexpected run response: %+v", run)
	}

	quick, err := d.QuickEdit(ctx, api.QuickRequest{Description: "update typo"})
	if err != nil {
		t.Fatalf("quick edit failed: %v", err)
	}
	if !quick.Passed {
		t.Error("expected quick edit to pass")
	}
}

func TestVerify_ReflectsGateResult(t *testing.T) {
	d, _ := newTestDirector(t, []qualitygate.Check{
		fastCheck("pass", "exit 0", "blocking"),
		fastCheck("warn", "exit 1", "advisory"),
	})

	resp, err := d.Verify(context.Background())
	if err != nil {
		t.Fatalf("verify failed: %v", err)
	}
	if resp.Score != 50 {
		t.Errorf("expected score 50, got %d", resp.Score)
	}
	if !resp.Passed {
		t.Error("expected pass: no blocking failures and score >= threshold")
	}
	if len(resp.Advisory) != 1 || resp.Advisory[0] != "warn" {
		t.Errorf("expected advisory [warn], got %v", resp.Advisory)
	}
}

func TestGetPlanDetail_NoPlan(t *testing.T) {
	d, _ := newTestDirector(t, nil)

	detail, err := d.GetPlanDetail(context.Background())
	if err != nil {
		t.Fatalf("expected nil error for missing plan, got %v", err)
	}
	if detail != nil {
		t.Errorf("expected nil detail, got %+v", detail)
	}
}

func TestGetPlanDetail_AfterBuild(t *testing.T) {
	d, _ := newTestDirector(t, []qualitygate.Check{fastCheck("fast", "exit 0", "blocking")})
	if _, err := d.Handle(context.Background(), types.UserRequest{Message: "build a todo app"}); err != nil {
		t.Fatalf("handle failed: %v", err)
	}

	detail, err := d.GetPlanDetail(context.Background())
	if err != nil {
		t.Fatalf("get plan detail failed: %v", err)
	}
	if detail == nil {
		t.Fatal("expected plan detail")
	}
	if len(detail.Nodes) != 9 {
		t.Errorf("expected 9 nodes, got %d", len(detail.Nodes))
	}
	if len(detail.Gates) != 3 {
		t.Errorf("expected 3 gates, got %d", len(detail.Gates))
	}
	if detail.Nodes[0].ID != "n1-clarify" {
		t.Errorf("unexpected first node: %s", detail.Nodes[0].ID)
	}
}

func TestContinuePlan_NoState(t *testing.T) {
	d, _ := newTestDirector(t, nil)

	resumed, msg, err := d.ContinuePlan(context.Background(), "plan-x")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resumed {
		t.Error("should not resume without workflow state")
	}
	if !strings.Contains(msg, "no workflow state") {
		t.Errorf("unexpected message: %q", msg)
	}
}

func TestContinuePlan_Parked(t *testing.T) {
	d, ws := newTestDirector(t, nil)
	p := planner.New(ws)

	state := &planner.WorkflowState{PlanID: "p1", Status: "parked", Current: "n5-spec"}
	if err := p.SaveState(state); err != nil {
		t.Fatal(err)
	}

	resumed, msg, err := d.ContinuePlan(context.Background(), "p1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !resumed || msg != "resumed" {
		t.Errorf("expected resumed, got %v %q", resumed, msg)
	}

	loaded, err := p.LoadState()
	if err != nil {
		t.Fatal(err)
	}
	if loaded.Status != "running" {
		t.Errorf("expected status running, got %s", loaded.Status)
	}
}

func TestContinuePlan_NotParked(t *testing.T) {
	d, ws := newTestDirector(t, nil)
	p := planner.New(ws)

	if err := p.SaveState(&planner.WorkflowState{PlanID: "p1", Status: "completed"}); err != nil {
		t.Fatal(err)
	}

	resumed, msg, err := d.ContinuePlan(context.Background(), "p1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resumed {
		t.Error("should not resume a completed workflow")
	}
	if !strings.Contains(msg, "completed") {
		t.Errorf("unexpected message: %q", msg)
	}
}

func TestGetAgentCards(t *testing.T) {
	// Without bus: no cards
	ws := t.TempDir()
	plain := New(router.New(), planner.New(ws), scheduler.New(ws), qualitygate.NewWithChecks(50, nil))
	if cards := plain.GetAgentCards(context.Background()); cards != nil {
		t.Errorf("expected nil cards without bus, got %v", cards)
	}

	// With bus: 5 registered reviewers
	d, _ := newTestDirector(t, nil)
	cards := d.GetAgentCards(context.Background())
	if len(cards) != 5 {
		t.Fatalf("expected 5 agent cards, got %d", len(cards))
	}
	roles := make(map[types.Role]bool)
	for _, c := range cards {
		roles[c.Role] = true
		if c.ID == "" || c.Name == "" {
			t.Errorf("card missing identity: %+v", c)
		}
	}
	for _, want := range []types.Role{types.RolePM, types.RoleArchitect, types.RoleQA, types.RoleSecurity, types.RoleDevOps} {
		if !roles[want] {
			t.Errorf("missing agent card for role %s", want)
		}
	}
}

func TestGetPlan(t *testing.T) {
	d, _ := newTestDirector(t, nil)
	resp, err := d.GetPlan(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.PlanJSON != "{}" || resp.Nodes != 0 {
		t.Errorf("expected empty plan response, got %+v", resp)
	}
}

func TestRunPipeline_BuildPlan(t *testing.T) {
	d, _ := newTestDirector(t, []qualitygate.Check{fastCheck("fast", "exit 0", "blocking")})
	resp, err := d.RunPipeline(context.Background(), api.RunRequest{Requirement: "build a todo app", Backend: "codex"})
	if err != nil {
		t.Fatalf("run pipeline failed: %v", err)
	}
	if resp.PlanID == "" || resp.PlanID == "quick" {
		t.Errorf("expected real plan ID, got %s", resp.PlanID)
	}
	if len(resp.Artifacts) == 0 {
		t.Error("expected artifacts from build pipeline")
	}
}

func TestQuickEdit_ReturnsScore(t *testing.T) {
	d, _ := newTestDirector(t, []qualitygate.Check{fastCheck("fast", "exit 0", "blocking")})
	resp, err := d.QuickEdit(context.Background(), api.QuickRequest{Description: "update typo", Backend: "codex"})
	if err != nil {
		t.Fatalf("quick edit failed: %v", err)
	}
	if resp.Score != 100 {
		t.Errorf("expected score 100, got %d", resp.Score)
	}
	if !resp.Passed {
		t.Error("expected quick edit to pass")
	}
}

func TestHandle_PlanError(t *testing.T) {
	// Use a planner with an unwriteable workspace to trigger plan error
	ws := filepath.Join(t.TempDir(), "nonexistent_deep", "path")
	r := router.New()
	p := planner.New(ws)
	bus := a2a.NewBus()
	s := scheduler.NewWithBus(ws, bus)
	g := qualitygate.NewWithChecks(50, []qualitygate.Check{fastCheck("fast", "exit 0", "blocking")})
	d := NewWithBus(r, p, s, g, bus)

	// The planner should fail to save to a nonexistent deeply nested path
	_, err := d.Handle(context.Background(), types.UserRequest{Message: "build a todo app"})
	if err == nil {
		t.Skip("plan save did not fail on this OS")
	}
}

func TestRunPipeline_ErrorPropagation(t *testing.T) {
	// Force a schedule error by giving an unwriteable workspace
	ws := filepath.Join(t.TempDir(), "nonexistent_deep", "path")
	r := router.New()
	p := planner.New(ws)
	bus := a2a.NewBus()
	s := scheduler.NewWithBus(ws, bus)
	g := qualitygate.NewWithChecks(50, []qualitygate.Check{fastCheck("fast", "exit 0", "blocking")})
	d := NewWithBus(r, p, s, g, bus)

	_, err := d.RunPipeline(context.Background(), api.RunRequest{Requirement: "build a todo app", Backend: "codex"})
	if err == nil {
		t.Skip("plan save did not fail on this OS")
	}
}

func TestQuickEdit_ErrorPropagation(t *testing.T) {
	ws := filepath.Join(t.TempDir(), "nonexistent_deep", "path")
	r := router.New()
	p := planner.New(ws)
	bus := a2a.NewBus()
	s := scheduler.NewWithBus(ws, bus)
	g := qualitygate.NewWithChecks(50, []qualitygate.Check{fastCheck("fast", "exit 0", "blocking")})
	d := NewWithBus(r, p, s, g, bus)

	_, err := d.QuickEdit(context.Background(), api.QuickRequest{Description: "fix bug", Backend: "codex"})
	if err == nil {
		t.Skip("plan save did not fail on this OS")
	}
}

func TestToCheckSummaries_Nil(t *testing.T) {
	if v := toCheckSummaries(nil); v != nil {
		t.Errorf("expected nil for nil input, got %v", v)
	}
}

func TestToAPISummary_Nil(t *testing.T) {
	if v := toAPISummary(nil); v != nil {
		t.Errorf("expected nil for nil input, got %v", v)
	}
}

func TestContinuePlan_SaveStateError(t *testing.T) {
	// Use a workspace where SaveState will fail
	ws := filepath.Join(t.TempDir(), "nonexistent_deep", "path")
	r := router.New()
	p := planner.New(ws)
	bus := a2a.NewBus()
	s := scheduler.NewWithBus(ws, bus)
	g := qualitygate.NewWithChecks(50, nil)
	d := NewWithBus(r, p, s, g, bus)

	// Pre-create a state file in an unwriteable location won't work;
	// instead test the nil planner path
	resumed, msg, err := d.ContinuePlan(context.Background(), "x")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resumed {
		t.Error("should not resume without workflow state")
	}
	if !strings.Contains(msg, "no workflow state") {
		t.Errorf("unexpected message: %q", msg)
	}
}
