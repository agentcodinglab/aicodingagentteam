// Package agent implements reviewer role agents for A2A dispatch.
// Each agent processes tasks and returns structured Verdicts.
// MVP: stub agents that validate artifacts and return accept/blocking.
package agent

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/yourorg/aicodingagentteam/internal/a2a"
	"github.com/yourorg/aicodingagentteam/internal/types"
)

// ReviewerAgent is a generic reviewer role agent.
// It checks artifact existence and returns a Verdict.
type ReviewerAgent struct {
	card    a2a.AgentCard
	checkFn func(ctx context.Context, task a2a.Task) (types.Verdict, error)
}

// NewReviewer creates a reviewer agent with the given card and check function.
func NewReviewer(card a2a.AgentCard, fn func(ctx context.Context, task a2a.Task) (types.Verdict, error)) *ReviewerAgent {
	return &ReviewerAgent{card: card, checkFn: fn}
}

// Card returns the agent's capability declaration.
func (a *ReviewerAgent) Card() a2a.AgentCard { return a.card }

// Execute runs the agent's check function and returns a Result.
func (a *ReviewerAgent) Execute(ctx context.Context, task a2a.Task) (a2a.Result, error) {
	verdict, err := a.checkFn(ctx, task)
	if err != nil {
		return a2a.Result{TaskID: task.TaskID, Verdict: verdict}, err
	}
	if verdict.TaskID == "" {
		verdict.TaskID = task.TaskID
	}
	if verdict.Role == "" {
		verdict.Role = task.Role
	}
	return a2a.Result{TaskID: task.TaskID, Verdict: verdict}, nil
}

// Status returns the agent's current status.
func (a *ReviewerAgent) Status(ctx context.Context) string { return "idle" }

// DefaultReviewerCheck is the default check function for reviewer agents.
// It validates that required artifacts exist and returns accept or blocking.
func DefaultReviewerCheck(role types.Role) func(ctx context.Context, task a2a.Task) (types.Verdict, error) {
	return func(ctx context.Context, task a2a.Task) (types.Verdict, error) {
		verdict := types.Verdict{
			TaskID:    task.TaskID,
			Role:      role,
			Decision:  types.DecisionAccept,
			Artifacts: task.Payload.Artifacts,
		}

		// Check if required artifacts exist
		var missing []string
		for _, art := range task.Payload.Artifacts {
			if strings.HasSuffix(art, "/") {
				continue // directory artifact, skip existence check
			}
			if _, err := os.Stat(art); err != nil {
				missing = append(missing, art)
			}
		}

		if len(missing) > 0 && role != types.RolePM && role != types.RoleArchitect {
			// For QA/Security/DevOps, missing artifacts are blocking
			verdict.Decision = types.DecisionBlocking
			verdict.Severity = "critical"
			verdict.Findings = append(verdict.Findings, types.Finding{
				Check:  "artifact-existence",
				Status: "fail",
				Detail: fmt.Sprintf("missing artifacts: %s", strings.Join(missing, ", ")),
			})
		}

		return verdict, nil
	}
}

// NewPMAgent creates a Product Manager reviewer agent.
func NewPMAgent() *ReviewerAgent {
	return NewReviewer(a2a.AgentCard{
		ID: "agent-pm", Name: "Product Manager", Role: types.RolePM,
		Capabilities:  []string{"prd-generation", "acceptance-criteria"},
		MaxConcurrent: 1, TimeoutDefault: 300,
	}, DefaultReviewerCheck(types.RolePM))
}

// NewArchitectAgent creates an Architect reviewer agent.
func NewArchitectAgent() *ReviewerAgent {
	return NewReviewer(a2a.AgentCard{
		ID: "agent-architect", Name: "Architect", Role: types.RoleArchitect,
		Capabilities:  []string{"architecture-design", "api-contract"},
		MaxConcurrent: 1, TimeoutDefault: 300,
	}, DefaultReviewerCheck(types.RoleArchitect))
}

// NewQAAgent creates a QA reviewer agent.
func NewQAAgent() *ReviewerAgent {
	return NewReviewer(a2a.AgentCard{
		ID: "agent-qa", Name: "QA Engineer", Role: types.RoleQA,
		Capabilities:  []string{"test-generation", "runtime-probe", "coverage-analysis"},
		MaxConcurrent: 1, TimeoutDefault: 300,
	}, DefaultReviewerCheck(types.RoleQA))
}

// NewSecurityAgent creates a Security reviewer agent.
func NewSecurityAgent() *ReviewerAgent {
	return NewReviewer(a2a.AgentCard{
		ID: "agent-security", Name: "Security Engineer", Role: types.RoleSecurity,
		Capabilities:  []string{"threat-modeling", "sast-scan", "secret-detection"},
		MaxConcurrent: 1, TimeoutDefault: 300,
	}, DefaultReviewerCheck(types.RoleSecurity))
}

// NewDevOpsAgent creates a DevOps reviewer agent.
func NewDevOpsAgent() *ReviewerAgent {
	return NewReviewer(a2a.AgentCard{
		ID: "agent-devops", Name: "DevOps Engineer", Role: types.RoleDevOps,
		Capabilities:  []string{"dockerfile", "ci-config", "deploy-proof"},
		MaxConcurrent: 1, TimeoutDefault: 300,
	}, DefaultReviewerCheck(types.RoleDevOps))
}

// RegisterAllReviewers registers all 5 reviewer agents on the bus.
func RegisterAllReviewers(bus a2a.Bus) {
	bus.Register(NewPMAgent())
	bus.Register(NewArchitectAgent())
	bus.Register(NewQAAgent())
	bus.Register(NewSecurityAgent())
	bus.Register(NewDevOpsAgent())
}
