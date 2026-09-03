// Package coordinator is the core orchestration engine: route -> plan -> schedule -> verify -> finalize.
package coordinator

import (
	"context"
	"fmt"
	"time"

	"github.com/agentcodinglab/aicodingagentteam/internal/a2a"
	"github.com/agentcodinglab/aicodingagentteam/internal/agent"
	"github.com/agentcodinglab/aicodingagentteam/internal/planner"
	"github.com/agentcodinglab/aicodingagentteam/internal/qualitygate"
	"github.com/agentcodinglab/aicodingagentteam/internal/router"
	"github.com/agentcodinglab/aicodingagentteam/internal/scheduler"
	"github.com/agentcodinglab/aicodingagentteam/internal/types"
	"github.com/agentcodinglab/aicodingagentteam/pkg/api"
)

// Director is the core scheduling loop implementing the 5-layer model.
type Director struct {
	router  *router.Router
	planner *planner.Planner
	sched   *scheduler.Scheduler
	gate    *qualitygate.Engine
	bus     a2a.Bus
}

// New creates a Director wiring all engine components.
func New(r *router.Router, p *planner.Planner, s *scheduler.Scheduler, g *qualitygate.Engine) *Director {
	return &Director{router: r, planner: p, sched: s, gate: g}
}

// NewWithBus creates a Director with an A2A Bus and registered reviewer agents.
func NewWithBus(r *router.Router, p *planner.Planner, s *scheduler.Scheduler, g *qualitygate.Engine, bus a2a.Bus) *Director {
	agent.RegisterAllReviewers(bus)
	return &Director{router: r, planner: p, sched: s, gate: g, bus: bus}
}

// Handle processes a user request through the full 5-layer flow.
func (d *Director) Handle(ctx context.Context, req types.UserRequest) (*types.Delivery, error) {
	// ① Route
	intent := d.router.Route(ctx, req)

	// ② Plan
	plan, err := d.planner.Build(ctx, intent)
	if err != nil {
		return nil, fmt.Errorf("plan: %w", err)
	}

	// ③ Schedule
	result, err := d.sched.Execute(ctx, plan)
	if err != nil {
		return nil, fmt.Errorf("schedule: %w", err)
	}
	if result.Parked {
		return &types.Delivery{PlanID: plan.ID, CreatedAt: time.Now(), Passed: false}, nil
	}

	// ④ Verify
	verdict := d.gate.Verify(ctx, result.Artifacts)

	// ⑤ Finalize
	delivery := &types.Delivery{
		PlanID:    plan.ID,
		Artifacts: result.Artifacts,
		Score:     verdict.Score,
		Passed:    verdict.Passed,
		CreatedAt: time.Now(),
	}
	return delivery, nil
}

// Ensure Director implements the api.Handler and api.ExtendedHandler interfaces.
var _ api.Handler = (*Director)(nil)
var _ api.ExtendedHandler = (*Director)(nil)

// RunPipeline implements api.Handler.
func (d *Director) RunPipeline(ctx context.Context, req api.RunRequest) (*api.RunResponse, error) {
	delivery, err := d.Handle(ctx, types.UserRequest{Message: req.Requirement, Backend: req.Backend})
	if err != nil {
		return nil, err
	}
	return &api.RunResponse{PlanID: delivery.PlanID, Artifacts: delivery.Artifacts, Score: delivery.Score, Passed: delivery.Passed}, nil
}

// QuickEdit implements api.Handler.
func (d *Director) QuickEdit(ctx context.Context, req api.QuickRequest) (*api.QuickResponse, error) {
	delivery, err := d.Handle(ctx, types.UserRequest{Message: req.Description, Backend: req.Backend})
	if err != nil {
		return nil, err
	}
	return &api.QuickResponse{FilesChanged: delivery.Artifacts, Passed: delivery.Passed}, nil
}

// Verify implements api.Handler.
func (d *Director) Verify(ctx context.Context) (*api.VerifyResponse, error) {
	v := d.gate.Verify(ctx, nil)
	return &api.VerifyResponse{Score: v.Score, Passed: v.Passed, Blocking: v.Blocking, Advisory: v.Advisory}, nil
}

// GetPlan implements api.Handler.
func (d *Director) GetPlan(ctx context.Context) (*api.PlanResponse, error) {
	return &api.PlanResponse{PlanJSON: "{}", Nodes: 0}, nil
}

// GetAgentCards implements api.ExtendedHandler.
// Returns all registered agent cards from the A2A Bus.
func (d *Director) GetAgentCards(ctx context.Context) []a2a.AgentCard {
	if d.bus == nil {
		return nil
	}
	return d.bus.Discover()
}

// GetPlanDetail implements api.ExtendedHandler.
// Reads the persisted plan from .aicodingagentteam/plan.json.
func (d *Director) GetPlanDetail(ctx context.Context) (*api.PlanDetail, error) {
	if d.planner == nil {
		return nil, nil
	}
	plan, err := d.planner.Load()
	if err != nil {
		return nil, nil // no plan file yet
	}

	detail := &api.PlanDetail{ID: plan.ID}
	for _, n := range plan.Nodes {
		detail.Nodes = append(detail.Nodes, api.PlanNode{
			ID:    n.ID,
			Phase: string(n.Phase),
			Role:  string(n.Role),
		})
	}
	for _, g := range plan.Gates {
		detail.Gates = append(detail.Gates, api.PlanGate{
			ID:    g.ID,
			After: g.After,
			Type:  string(g.Type),
		})
	}
	return detail, nil
}

// ContinuePlan implements api.ExtendedHandler.
// Resumes a parked workflow by loading the workflow state and continuing execution.
func (d *Director) ContinuePlan(ctx context.Context, planID string) (bool, string, error) {
	if d.planner == nil {
		return false, "planner not configured", nil
	}

	state, err := d.planner.LoadState()
	if err != nil {
		return false, "no workflow state found", nil
	}

	if state.Status != "parked" && state.Status != "paused" {
		return false, fmt.Sprintf("workflow status is %s, not parked/paused", state.Status), nil
	}

	// Mark as resumed
	state.Status = "running"
	state.Current = ""
	if err := d.planner.SaveState(state); err != nil {
		return false, "", fmt.Errorf("save state: %w", err)
	}

	return true, "resumed", nil
}
