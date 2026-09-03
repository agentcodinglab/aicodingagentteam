// Package planner builds DAG task plans from routing intents and persists them.
package planner

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/yourorg/aicodingagentteam/internal/types"
)

// Planner constructs DAG plans from intents and persists them to plan.json.
type Planner struct {
	workspace string // workspace root for .aicodingagentteam/ dir
}

// New creates a Planner with the given workspace root.
// Pass "" to use the current directory.
func New(workspace string) *Planner {
	if workspace == "" {
		workspace, _ = os.Getwd()
	}
	return &Planner{workspace: workspace}
}

// Build generates a plan.json-equivalent DAG for the given intent.
func (p *Planner) Build(ctx context.Context, intent types.Intent) (*types.Plan, error) {
	if intent.Type == types.IntentQuickEdit || intent.Type == types.IntentChat {
		plan := &types.Plan{ID: "quick"}
		return plan, nil
	}

	planID := fmt.Sprintf("plan-%d", time.Now().UnixNano())
	plan := &types.Plan{ID: planID}
	plan.Nodes = []types.TaskNode{
		{ID: "n1-clarify", Phase: types.PhaseClarify, Role: types.RoleCoordinator, ArtifactsOut: []string{"output/app-clarify.md"}},
		{ID: "n2-research", Phase: types.PhaseResearch, Role: types.RoleCoordinator, DependsOn: []string{"n1-clarify"}, ArtifactsIn: []string{"output/app-clarify.md"}, ArtifactsOut: []string{"output/app-research.md"}},
		{ID: "n3-prd", Phase: types.PhaseDocs, Role: types.RolePM, DependsOn: []string{"n2-research"}, ArtifactsIn: []string{"output/app-research.md"}, ArtifactsOut: []string{"output/app-prd.md"}},
		{ID: "n4-arch", Phase: types.PhaseDocs, Role: types.RoleArchitect, DependsOn: []string{"n2-research"}, ArtifactsIn: []string{"output/app-research.md"}, ArtifactsOut: []string{"output/app-architecture.md", ".aicodingagentteam/contracts/openapi.json"}},
		{ID: "n5-spec", Phase: types.PhaseSpec, Role: types.RoleCoordinator, DependsOn: []string{"n3-prd", "n4-arch"}, ArtifactsIn: []string{"output/app-prd.md", "output/app-architecture.md"}, ArtifactsOut: []string{"output/app-execution-plan.md"}},
		{ID: "n6-frontend", Phase: types.PhaseFrontend, Role: types.RoleFrontend, DependsOn: []string{"n5-spec"}, Writer: true, ArtifactsIn: []string{"output/app-architecture.md"}, ArtifactsOut: []string{"src/frontend/"}},
		{ID: "n7-backend", Phase: types.PhaseBackend, Role: types.RoleBackend, DependsOn: []string{"n6-frontend"}, Writer: true, ArtifactsIn: []string{".aicodingagentteam/contracts/openapi.json"}, ArtifactsOut: []string{"src/backend/"}},
		{ID: "n8-quality", Phase: types.PhaseQuality, Role: types.RoleQA, DependsOn: []string{"n7-backend"}, ArtifactsOut: []string{"output/app-quality-gate.md"}},
		{ID: "n9-delivery", Phase: types.PhaseDelivery, Role: types.RoleCoordinator, DependsOn: []string{"n8-quality"}, ArtifactsOut: []string{".aicodingagentteam/proof/proof-pack.zip"}},
	}
	plan.Gates = []types.Gate{
		{ID: "g1-docs", After: "n4-arch", Type: types.GateHuman},
		{ID: "g2-preview", After: "n6-frontend", Type: types.GateHuman},
		{ID: "g3-quality", After: "n8-quality", Type: types.GateAuto},
	}

	if err := p.Save(plan); err != nil {
		return plan, fmt.Errorf("persist plan: %w", err)
	}
	return plan, nil
}

// planFile is the persisted plan.json structure.
type planFile struct {
	ID    string           `json:"id"`
	Nodes []types.TaskNode `json:"nodes"`
	Gates []types.Gate     `json:"gates"`
}

// Save writes the plan to .aicodingagentteam/plan.json.
func (p *Planner) Save(plan *types.Plan) error {
	dir := filepath.Join(p.workspace, ".aicodingagentteam")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	data, err := json.MarshalIndent(planFile{
		ID:    plan.ID,
		Nodes: plan.Nodes,
		Gates: plan.Gates,
	}, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(dir, "plan.json"), data, 0o644)
}

// Load reads the plan from .aicodingagentteam/plan.json.
func (p *Planner) Load() (*types.Plan, error) {
	data, err := os.ReadFile(filepath.Join(p.workspace, ".aicodingagentteam", "plan.json"))
	if err != nil {
		return nil, err
	}
	var pf planFile
	if err := json.Unmarshal(data, &pf); err != nil {
		return nil, err
	}
	return &types.Plan{ID: pf.ID, Nodes: pf.Nodes, Gates: pf.Gates}, nil
}

// WorkflowState tracks runtime execution state, persisted to workflow-state.json.
type WorkflowState struct {
	PlanID    string            `json:"plan_id"`
	Status    string            `json:"status"`    // running / paused / completed / parked
	Completed []string          `json:"completed"` // completed node IDs
	Current   string            `json:"current"`   // current running node ID
	Gates     map[string]string `json:"gates"`     // gate ID -> pending / approved / rejected
	Updated   time.Time         `json:"updated"`
}

// SaveState persists the workflow state.
func (p *Planner) SaveState(state *WorkflowState) error {
	dir := filepath.Join(p.workspace, ".aicodingagentteam")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	data, err := json.MarshalIndent(state, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(dir, "workflow-state.json"), data, 0o644)
}

// LoadState reads the workflow state.
func (p *Planner) LoadState() (*WorkflowState, error) {
	data, err := os.ReadFile(filepath.Join(p.workspace, ".aicodingagentteam", "workflow-state.json"))
	if err != nil {
		return nil, err
	}
	var state WorkflowState
	if err := json.Unmarshal(data, &state); err != nil {
		return nil, err
	}
	return &state, nil
}
