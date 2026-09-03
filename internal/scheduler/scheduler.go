// Package scheduler dispatches DAG tasks to role agents via A2A.
// Writer roles are serialized (single-writer model, ADR-0004).
// Reviewer roles are dispatched in parallel via the A2A Bus.
package scheduler

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/agentcodinglab/aicodingagentteam/internal/a2a"
	"github.com/agentcodinglab/aicodingagentteam/internal/audit"
	"github.com/agentcodinglab/aicodingagentteam/internal/types"
)

// Scheduler executes a Plan by dispatching tasks to role agents.
type Scheduler struct {
	mu        sync.Mutex // single-writer lock: only one writer at a time
	workspace string
	audit     *audit.Logger
	bus       a2a.Bus // A2A message bus for reviewer dispatch
	verdicts  map[string][]types.Verdict
}

// New creates a Scheduler with the given workspace root.
func New(workspace string) *Scheduler {
	if workspace == "" {
		workspace, _ = os.Getwd()
	}
	return &Scheduler{
		workspace: workspace,
		verdicts:  make(map[string][]types.Verdict),
	}
}

// NewWithBus creates a Scheduler connected to an A2A Bus.
func NewWithBus(workspace string, bus a2a.Bus) *Scheduler {
	s := New(workspace)
	s.bus = bus
	return s
}

// NewWithAudit creates a Scheduler with an audit logger.
func NewWithAudit(workspace string, al *audit.Logger) *Scheduler {
	s := New(workspace)
	s.audit = al
	return s
}

// NewFull creates a Scheduler with both Bus and audit logger.
func NewFull(workspace string, bus a2a.Bus, al *audit.Logger) *Scheduler {
	s := New(workspace)
	s.bus = bus
	s.audit = al
	return s
}

// Result is the output of a scheduled plan execution.
type Result struct {
	PlanID    string
	Verdicts  []types.Verdict
	Artifacts []string
	Parked    bool
}

// Execute runs the plan: writer roles serialized, reviewer roles in parallel via A2A.
func (s *Scheduler) Execute(ctx context.Context, plan *types.Plan) (*Result, error) {
	if plan == nil {
		return nil, fmt.Errorf("nil plan")
	}
	res := &Result{PlanID: plan.ID}

	// Phase 1: Collect reviewer tasks (non-writer, non-coordinator nodes) for parallel dispatch
	var reviewerTasks []a2a.Task
	var reviewerNodes []types.TaskNode
	var writerNodes []types.TaskNode
	var coordinatorNodes []types.TaskNode

	for _, node := range plan.Nodes {
		switch {
		case node.Role == types.RoleCoordinator:
			coordinatorNodes = append(coordinatorNodes, node)
		case node.Role.IsWriter():
			writerNodes = append(writerNodes, node)
		default:
			reviewerNodes = append(reviewerNodes, node)
		}
	}

	// Phase 2: Dispatch reviewer tasks in parallel via A2A Bus
	if s.bus != nil && len(reviewerNodes) > 0 {
		for _, node := range reviewerNodes {
			reviewerTasks = append(reviewerTasks, a2a.Task{
				TaskID: node.ID,
				Role:   node.Role,
				Payload: a2a.TaskPayload{
					Artifacts:   node.ArtifactsIn,
					Instruction: fmt.Sprintf("review task %s in phase %s", node.ID, node.Phase),
				},
				Deadline: time.Now().Add(300 * time.Second),
			})
		}
		results := s.bus.DelegateParallel(ctx, reviewerTasks)
		for i, r := range results {
			verdict := r.Verdict
			if verdict.TaskID == "" {
				verdict.TaskID = reviewerNodes[i].ID
			}
			if verdict.Role == "" {
				verdict.Role = reviewerNodes[i].Role
			}
			if len(verdict.Artifacts) == 0 {
				verdict.Artifacts = reviewerNodes[i].ArtifactsOut
			}
			if verdict.Decision == types.DecisionBlocking {
				res.Parked = true
			}
			res.Verdicts = append(res.Verdicts, verdict)
			s.logAudit(reviewerNodes[i], "a2a-delegate", string(verdict.Decision))
		}
	} else {
		// No bus: stub verdicts for reviewer nodes
		for _, node := range reviewerNodes {
			verdict := types.Verdict{
				TaskID: node.ID, Role: node.Role,
				Decision: types.DecisionAccept, Artifacts: node.ArtifactsOut,
			}
			res.Verdicts = append(res.Verdicts, verdict)
		}
	}

	// Phase 3: Coordinator nodes (internal, no A2A dispatch)
	for _, node := range coordinatorNodes {
		verdict := types.Verdict{
			TaskID: node.ID, Role: node.Role,
			Decision: types.DecisionAccept, Artifacts: node.ArtifactsOut,
		}
		res.Verdicts = append(res.Verdicts, verdict)
		res.Artifacts = append(res.Artifacts, node.ArtifactsOut...)
		s.logAudit(node, "coordinator", "pass")
	}

	// Phase 4: Writer nodes serialized (single-writer model)
	for _, node := range writerNodes {
		verdict := types.Verdict{
			TaskID: node.ID, Role: node.Role,
			Decision: types.DecisionAccept, Artifacts: node.ArtifactsOut,
		}

		if err := s.acquireWriteLock(ctx, node); err != nil {
			verdict.Decision = types.DecisionBlocking
			verdict.Findings = append(verdict.Findings, types.Finding{
				Check: "write-lock", Status: "fail", Detail: err.Error(),
			})
			res.Parked = true
			res.Verdicts = append(res.Verdicts, verdict)
			continue
		}
		s.mu.Lock()
		res.Artifacts = append(res.Artifacts, node.ArtifactsOut...)
		s.mu.Unlock()
		s.releaseWriteLock(node)
		res.Verdicts = append(res.Verdicts, verdict)
		s.logAudit(node, "write-complete", "pass")
	}

	return res, nil
}

// acquireWriteLock obtains a file lock for the writer role.
func (s *Scheduler) acquireWriteLock(ctx context.Context, node types.TaskNode) error {
	lockPath := filepath.Join(s.workspace, ".aicodingagentteam", "write.lock")
	if _, err := os.Stat(lockPath); err == nil {
		data, _ := os.ReadFile(lockPath)
		return fmt.Errorf("write lock held by another writer: %s", string(data))
	}
	if err := os.MkdirAll(filepath.Dir(lockPath), 0o755); err != nil {
		return err
	}
	content := fmt.Sprintf("%s|%s|%s", node.ID, node.Role, time.Now().Format(time.RFC3339))
	if err := os.WriteFile(lockPath, []byte(content), 0o644); err != nil {
		return fmt.Errorf("acquire write lock: %w", err)
	}
	return nil
}

// releaseWriteLock removes the write lock file.
func (s *Scheduler) releaseWriteLock(node types.TaskNode) {
	lockPath := filepath.Join(s.workspace, ".aicodingagentteam", "write.lock")
	_ = os.Remove(lockPath)
}

// GetVerdicts returns collected verdicts for a task node.
func (s *Scheduler) GetVerdicts(nodeID string) []types.Verdict {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.verdicts[nodeID]
}

// logAudit records a scheduler event.
func (s *Scheduler) logAudit(node types.TaskNode, eventType, result string) {
	if s.audit == nil {
		return
	}
	_ = s.audit.Log(audit.Entry{
		Type:   eventType,
		Agent:  string(node.Role),
		Task:   node.ID,
		Result: result,
	})
}
