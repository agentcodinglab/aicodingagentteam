// Package a2a implements Agent-to-Agent protocol communication.
// Supports in-process channels (dev) and Redis Pub/Sub (containerized).
package a2a

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/agentcodinglab/aicodingagentteam/internal/audit"
	"github.com/agentcodinglab/aicodingagentteam/internal/types"
)

// AgentCard declares an agent's capabilities for discovery.
type AgentCard struct {
	ID             string     `json:"id"`
	Name           string     `json:"name"`
	Role           types.Role `json:"role"`
	Capabilities   []string   `json:"capabilities"`
	Endpoint       string     `json:"endpoint"`
	MaxConcurrent  int        `json:"max_concurrent"`
	TimeoutDefault int        `json:"timeout_default"`
}

// Task is an A2A task delegation message.
type Task struct {
	TaskID   string      `json:"task_id"`
	Role     types.Role  `json:"role"`
	Payload  TaskPayload `json:"payload"`
	Deadline time.Time   `json:"deadline"`
}

// TaskPayload carries the task instruction and artifacts.
type TaskPayload struct {
	Artifacts   []string `json:"artifacts"`
	Instruction string   `json:"instruction"`
}

// Result is the A2A task result with a structured Verdict.
type Result struct {
	TaskID  string        `json:"task_id"`
	Verdict types.Verdict `json:"verdict"`
}

// ProgressEvent is a streaming progress update from an agent.
type ProgressEvent struct {
	TaskID  string `json:"task_id"`
	Phase   string `json:"phase"`
	Status  string `json:"status"`
	Message string `json:"message"`
}

// Agent is the interface every role agent implements.
type Agent interface {
	Card() AgentCard
	Execute(ctx context.Context, task Task) (Result, error)
	Status(ctx context.Context) string
}

// Bus is the interface for A2A message bus implementations.
// In-process (InProcBus) for development; Redis (RedisBus) for containerized.
type Bus interface {
	Register(a Agent)
	Delegate(ctx context.Context, task Task) (Result, error)
	DelegateParallel(ctx context.Context, tasks []Task) []Result
	Discover() []AgentCard
	ProgressChan() <-chan ProgressEvent
	IsRegistered(role types.Role) bool
	Count() int
}

// InProcBus is the in-process A2A message bus.
type InProcBus struct {
	mu       sync.RWMutex
	agents   map[types.Role]Agent
	audit    *audit.Logger
	progress chan ProgressEvent
}

// NewBus creates an in-process A2A message bus.
func NewBus() *InProcBus {
	return &InProcBus{
		agents:   make(map[types.Role]Agent),
		progress: make(chan ProgressEvent, 64),
	}
}

// NewBusWithAudit creates an in-process Bus with an audit logger.
func NewBusWithAudit(al *audit.Logger) *InProcBus {
	b := NewBus()
	b.audit = al
	return b
}

// Register registers a role agent on the bus.
func (b *InProcBus) Register(a Agent) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.agents[a.Card().Role] = a
}

// Delegate sends a task to the agent for the task's role.
func (b *InProcBus) Delegate(ctx context.Context, task Task) (Result, error) {
	b.mu.RLock()
	a, ok := b.agents[task.Role]
	b.mu.RUnlock()
	if !ok {
		err := fmt.Errorf("no agent registered for role %s", task.Role)
		b.logAudit(task.TaskID, string(task.Role), "delegate-failed", err.Error())
		return Result{}, err
	}

	b.emitProgress(task.TaskID, string(task.PhaseFromRole()), "running", "")
	b.logAudit(task.TaskID, string(task.Role), "delegate", "dispatched")

	result, err := a.Execute(ctx, task)
	if err != nil {
		result = Result{
			TaskID: task.TaskID,
			Verdict: types.Verdict{
				TaskID:   task.TaskID,
				Role:     task.Role,
				Decision: types.DecisionBlocking,
				Findings: []types.Finding{{
					Check:  "agent-execution",
					Status: "fail",
					Detail: err.Error(),
				}},
			},
		}
		b.emitProgress(task.TaskID, "", "failed", err.Error())
		b.logAudit(task.TaskID, string(task.Role), "agent-error", err.Error())
		return result, err
	}

	status := "completed"
	if result.Verdict.Decision == types.DecisionBlocking {
		status = "parked"
	}
	b.emitProgress(task.TaskID, "", status, string(result.Verdict.Decision))
	b.logAudit(task.TaskID, string(task.Role), "agent-result", string(result.Verdict.Decision))
	return result, nil
}

// DelegateParallel dispatches multiple tasks concurrently.
func (b *InProcBus) DelegateParallel(ctx context.Context, tasks []Task) []Result {
	var wg sync.WaitGroup
	results := make([]Result, len(tasks))

	for i, task := range tasks {
		wg.Add(1)
		go func(idx int, t Task) {
			defer wg.Done()
			r, _ := b.Delegate(ctx, t)
			results[idx] = r
		}(i, task)
	}
	wg.Wait()
	return results
}

// Discover returns all registered agent cards.
func (b *InProcBus) Discover() []AgentCard {
	b.mu.RLock()
	defer b.mu.RUnlock()
	cards := make([]AgentCard, 0, len(b.agents))
	for _, a := range b.agents {
		cards = append(cards, a.Card())
	}
	return cards
}

// ProgressChan returns a read-only channel for streaming progress events.
func (b *InProcBus) ProgressChan() <-chan ProgressEvent {
	return b.progress
}

func (b *InProcBus) emitProgress(taskID, phase, status, msg string) {
	select {
	case b.progress <- ProgressEvent{TaskID: taskID, Phase: phase, Status: status, Message: msg}:
	default:
	}
}

func (b *InProcBus) logAudit(taskID, agent, eventType, detail string) {
	if b.audit == nil {
		return
	}
	_ = b.audit.Log(audit.Entry{
		Type:   "a2a_message",
		Task:   taskID,
		Agent:  agent,
		Result: eventType,
		Detail: detail,
	})
}

// IsRegistered checks if an agent for the given role is registered.
func (b *InProcBus) IsRegistered(role types.Role) bool {
	b.mu.RLock()
	defer b.mu.RUnlock()
	_, ok := b.agents[role]
	return ok
}

// Count returns the number of registered agents.
func (b *InProcBus) Count() int {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return len(b.agents)
}

// phaseFromRole returns the default phase for a task based on its role.
func (t Task) PhaseFromRole() types.Phase {
	switch t.Role {
	case types.RolePM:
		return types.PhaseDocs
	case types.RoleArchitect:
		return types.PhaseDocs
	case types.RoleQA:
		return types.PhaseQuality
	case types.RoleSecurity:
		return types.PhaseQuality
	case types.RoleDevOps:
		return types.PhaseDelivery
	default:
		return ""
	}
}
