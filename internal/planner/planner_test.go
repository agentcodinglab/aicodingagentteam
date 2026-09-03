package planner

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/yourorg/aicodingagentteam/internal/types"
)

func TestBuild_QuickEditReturnsTrivialPlan(t *testing.T) {
	p := New(t.TempDir())
	plan, err := p.Build(context.Background(), types.Intent{Type: types.IntentQuickEdit})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(plan.Nodes) != 0 {
		t.Errorf("quick edit should have 0 nodes, got %d", len(plan.Nodes))
	}
}

func TestBuild_BuildIntentReturnsFullPipeline(t *testing.T) {
	p := New(t.TempDir())
	plan, err := p.Build(context.Background(), types.Intent{Type: types.IntentBuild})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(plan.Nodes) != 9 {
		t.Errorf("build should have 9 nodes, got %d", len(plan.Nodes))
	}
	if len(plan.Gates) != 3 {
		t.Errorf("build should have 3 gates, got %d", len(plan.Gates))
	}
}

func TestBuild_FrontendNodeIsWriter(t *testing.T) {
	p := New(t.TempDir())
	plan, _ := p.Build(context.Background(), types.Intent{Type: types.IntentBuild})
	for _, n := range plan.Nodes {
		if n.Role == types.RoleFrontend && !n.Writer {
			t.Error("frontend node should be writer")
		}
		if n.Role == types.RoleBackend && !n.Writer {
			t.Error("backend node should be writer")
		}
	}
}

func TestBuild_PersistsPlanJSON(t *testing.T) {
	dir := t.TempDir()
	p := New(dir)
	plan, err := p.Build(context.Background(), types.Intent{Type: types.IntentBuild})
	if err != nil {
		t.Fatalf("build failed: %v", err)
	}
	planFile := filepath.Join(dir, ".aicodingagentteam", "plan.json")
	if _, err := os.Stat(planFile); err != nil {
		t.Fatalf("plan.json should exist: %v", err)
	}
	loaded, err := p.Load()
	if err != nil {
		t.Fatalf("load failed: %v", err)
	}
	if loaded.ID != plan.ID {
		t.Errorf("loaded ID %q != original %q", loaded.ID, plan.ID)
	}
	if len(loaded.Nodes) != 9 {
		t.Errorf("loaded should have 9 nodes, got %d", len(loaded.Nodes))
	}
}

func TestSaveState_PersistsWorkflowState(t *testing.T) {
	dir := t.TempDir()
	p := New(dir)
	state := &WorkflowState{
		PlanID:    "test-plan",
		Status:    "running",
		Completed: []string{"n1"},
		Current:   "n2",
		Gates:     map[string]string{"g1": "pending"},
	}
	if err := p.SaveState(state); err != nil {
		t.Fatalf("save failed: %v", err)
	}
	loaded, err := p.LoadState()
	if err != nil {
		t.Fatalf("load failed: %v", err)
	}
	if loaded.PlanID != "test-plan" {
		t.Errorf("plan id mismatch: %s", loaded.PlanID)
	}
	if loaded.Status != "running" {
		t.Errorf("status mismatch: %s", loaded.Status)
	}
	if loaded.Current != "n2" {
		t.Errorf("current mismatch: %s", loaded.Current)
	}
}

func TestLoad_NonExistentReturnsError(t *testing.T) {
	p := New(t.TempDir())
	_, err := p.Load()
	if err == nil {
		t.Error("should return error for non-existent plan")
	}
}
