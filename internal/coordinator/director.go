// Package coordinator is the core orchestration engine: route -> plan -> schedule -> verify -> finalize.
package coordinator

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/agentcodinglab/aicodingagentteam/internal/a2a"
	"github.com/agentcodinglab/aicodingagentteam/internal/agent"
	"github.com/agentcodinglab/aicodingagentteam/internal/knowledge"
	"github.com/agentcodinglab/aicodingagentteam/internal/memory"
	"github.com/agentcodinglab/aicodingagentteam/internal/planner"
	"github.com/agentcodinglab/aicodingagentteam/internal/qualitygate"
	"github.com/agentcodinglab/aicodingagentteam/internal/router"
	"github.com/agentcodinglab/aicodingagentteam/internal/scheduler"
	"github.com/agentcodinglab/aicodingagentteam/pkg/runtime"
	"github.com/agentcodinglab/aicodingagentteam/internal/types"
	"github.com/agentcodinglab/aicodingagentteam/pkg/api"
)

// Director is the core scheduling loop implementing the 5-layer model.
type Director struct {
	router    *router.Router
	planner   *planner.Planner
	sched     *scheduler.Scheduler
	gate      *qualitygate.Engine
	bus       a2a.Bus
	knowledge *knowledge.Engine // optional; nil skips RAG retrieval
	memory    *memory.Store     // optional; nil skips memory recall/capture
}

// DirectorOption configures optional Director components (knowledge, memory).
type DirectorOption func(*Director)

// WithKnowledge attaches a knowledge engine for RAG retrieval in the Handle flow.
func WithKnowledge(e *knowledge.Engine) DirectorOption {
	return func(d *Director) { d.knowledge = e }
}

// WithMemory attaches a memory store for fact/pitfall recall and capture.
func WithMemory(s *memory.Store) DirectorOption {
	return func(d *Director) { d.memory = s }
}

// WithDriver attaches a host driver so writer nodes dispatch to a real backend.
// When omitted, writer nodes use the legacy stub (just record planned artifacts).
func WithDriver(drv runtime.Runtime) DirectorOption {
	return func(d *Director) { d.sched.SetDriver(drv) }
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

// NewWithOptions creates a Director with bus, reviewers, and optional knowledge/memory.
func NewWithOptions(r *router.Router, p *planner.Planner, s *scheduler.Scheduler, g *qualitygate.Engine, bus a2a.Bus, opts ...DirectorOption) *Director {
	d := NewWithBus(r, p, s, g, bus)
	for _, opt := range opts {
		opt(d)
	}
	return d
}

// Handle processes a user request through the full 5-layer flow.
func (d *Director) Handle(ctx context.Context, req types.UserRequest) (*types.Delivery, error) {
	// ① Route
	intent := d.router.Route(ctx, req)

	// RAG retrieval: inject relevant knowledge chunks into intent context (fail-safe).
	d.enhanceWithKnowledge(ctx, req.Message, &intent)

	// Memory recall: surface known facts and recipes before planning (fail-safe).
	d.recallMemory(ctx, &intent)

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
		PlanID:       plan.ID,
		Artifacts:    result.Artifacts,
		Score:        verdict.Score,
		Passed:       verdict.Passed,
		CreatedAt:    time.Now(),
		CheckDetails: toCheckSummaries(verdict.Details),
	}

	// Memory capture: persist facts/pitfalls from this delivery (fail-safe).
	d.captureMemory(ctx, delivery)

	return delivery, nil
}

// enhanceWithKnowledge retrieves relevant code/doc chunks via the knowledge engine
// and appends their paths to the intent scope for downstream context.
// Fail-safe: any error or nil engine is silently skipped.
func (d *Director) enhanceWithKnowledge(ctx context.Context, query string, intent *types.Intent) {
	if d.knowledge == nil {
		return
	}
	chunks := d.knowledge.Retrieve(ctx, query, 5)
	if len(chunks) == 0 {
		return
	}
	var paths []string
	for _, c := range chunks {
		paths = append(paths, c.Path)
	}
	if intent.Scope == "" {
		intent.Scope = strings.Join(paths, ", ")
	} else {
		intent.Scope += " | knowledge: " + strings.Join(paths, ", ")
	}
	log.Printf("[director] knowledge retrieve: %d chunks for query %q", len(chunks), query)
}

// recallMemory surfaces known project facts and matching recipes before planning.
// Fail-safe: any error or nil store is silently skipped.
func (d *Director) recallMemory(ctx context.Context, intent *types.Intent) {
	if d.memory == nil {
		return
	}
	if facts, err := d.memory.RecallFacts(ctx); err == nil && len(facts) > 0 {
		log.Printf("[director] memory recall: %d facts", len(facts))
	}
}

// captureMemory persists the delivery outcome as a fact (success) or pitfall (failure).
// Fail-safe: any error or nil store is silently skipped.
func (d *Director) captureMemory(ctx context.Context, delivery *types.Delivery) {
	if d.memory == nil {
		return
	}
	status := "passed"
	if !delivery.Passed {
		status = "failed"
		// Capture pitfall for failed deliveries.
		var findings []string
		for _, c := range delivery.CheckDetails {
			if c.Status != "pass" {
				findings = append(findings, c.Name+":"+c.Status)
			}
		}
		detail := fmt.Sprintf("plan=%s score=%d failures=%s", delivery.PlanID, delivery.Score, strings.Join(findings, ";"))
		if err := d.memory.CapturePitfall(ctx, memory.Pitfall{ID: delivery.PlanID, Detail: detail, Verified: false}); err != nil {
			log.Printf("[director] memory capture pitfall error: %v", err)
		}
	}
	// Capture fact for every delivery (backend, score, status).
	fact := memory.Fact{
		Key:    "delivery:" + delivery.PlanID,
		Value:  fmt.Sprintf("score=%d status=%s backend=auto", delivery.Score, status),
		Source: "coordinator.Handle",
	}
	if err := d.memory.Capture(ctx, fact); err != nil {
		log.Printf("[director] memory capture fact error: %v", err)
	}
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
	return &api.QuickResponse{FilesChanged: delivery.Artifacts, Passed: delivery.Passed, Score: delivery.Score, Details: toAPISummary(delivery.CheckDetails)}, nil
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

// toCheckSummaries converts qualitygate check details into portable CheckSummary values.
func toCheckSummaries(details []qualitygate.CheckDetail) []types.CheckSummary {
	if len(details) == 0 {
		return nil
	}
	out := make([]types.CheckSummary, len(details))
	for i, d := range details {
		out[i] = types.CheckSummary{Name: d.Name, Status: d.Status, Output: d.Output}
	}
	return out
}

// toAPISummary maps types.CheckSummary to api.CheckSummary.
func toAPISummary(in []types.CheckSummary) []api.CheckSummary {
	if len(in) == 0 {
		return nil
	}
	out := make([]api.CheckSummary, len(in))
	for i, s := range in {
		out[i] = api.CheckSummary{Name: s.Name, Status: s.Status, Output: s.Output}
	}
	return out
}
