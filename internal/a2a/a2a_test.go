package a2a

import (
	"context"
	"testing"
	"time"

	"github.com/agentcodinglab/aicodingagentteam/internal/types"
)

// mockAgent is a test agent that returns a configurable verdict.
type mockAgent struct {
	card    AgentCard
	verdict types.Verdict
	err     error
}

func (m *mockAgent) Card() AgentCard { return m.card }
func (m *mockAgent) Execute(ctx context.Context, task Task) (Result, error) {
	v := m.verdict
	v.TaskID = task.TaskID
	v.Role = task.Role
	return Result{TaskID: task.TaskID, Verdict: v}, m.err
}
func (m *mockAgent) Status(ctx context.Context) string { return "idle" }

func TestRegisterAndDiscover(t *testing.T) {
	bus := NewBus()
	bus.Register(&mockAgent{
		card: AgentCard{ID: "qa-1", Name: "QA", Role: types.RoleQA},
	})
	bus.Register(&mockAgent{
		card: AgentCard{ID: "pm-1", Name: "PM", Role: types.RolePM},
	})
	cards := bus.Discover()
	if len(cards) != 2 {
		t.Errorf("expected 2 agents, got %d", len(cards))
	}
	if !bus.IsRegistered(types.RoleQA) {
		t.Error("QA should be registered")
	}
	if bus.IsRegistered(types.RoleFrontend) {
		t.Error("Frontend should not be registered")
	}
}

func TestDelegateRoutesByRole(t *testing.T) {
	bus := NewBus()
	bus.Register(&mockAgent{
		card:    AgentCard{ID: "qa-1", Role: types.RoleQA},
		verdict: types.Verdict{Decision: types.DecisionAccept},
	})
	task := Task{
		TaskID: "t1", Role: types.RoleQA,
		Payload: TaskPayload{Instruction: "test"},
	}
	result, err := bus.Delegate(context.Background(), task)
	if err != nil {
		t.Fatalf("delegate failed: %v", err)
	}
	if result.Verdict.Decision != types.DecisionAccept {
		t.Errorf("expected accept, got %s", result.Verdict.Decision)
	}
	if result.TaskID != "t1" {
		t.Errorf("expected task t1, got %s", result.TaskID)
	}
}

func TestDelegateUnregisteredRoleReturnsError(t *testing.T) {
	bus := NewBus()
	task := Task{TaskID: "t1", Role: types.RoleSecurity}
	_, err := bus.Delegate(context.Background(), task)
	if err == nil {
		t.Error("should return error for unregistered role")
	}
	// IsRegistered for unregistered role should return false (verified by error above)
}

func TestDelegateAgentErrorReturnsBlocking(t *testing.T) {
	bus := NewBus()
	bus.Register(&mockAgent{
		card: AgentCard{ID: "qa-1", Role: types.RoleQA},
		err:  context.DeadlineExceeded,
	})
	task := Task{TaskID: "t1", Role: types.RoleQA}
	result, _ := bus.Delegate(context.Background(), task)
	if result.Verdict.Decision != types.DecisionBlocking {
		t.Error("agent error should produce blocking verdict, not fake success")
	}
}

func TestDelegateParallelCollectsAllResults(t *testing.T) {
	bus := NewBus()
	bus.Register(&mockAgent{
		card:    AgentCard{ID: "qa-1", Role: types.RoleQA},
		verdict: types.Verdict{Decision: types.DecisionAccept},
	})
	bus.Register(&mockAgent{
		card:    AgentCard{ID: "pm-1", Role: types.RolePM},
		verdict: types.Verdict{Decision: types.DecisionBlocking},
	})
	tasks := []Task{
		{TaskID: "t1", Role: types.RoleQA},
		{TaskID: "t2", Role: types.RolePM},
	}
	results := bus.DelegateParallel(context.Background(), tasks)
	if len(results) != 2 {
		t.Fatalf("expected 2 results, got %d", len(results))
	}
	if results[0].Verdict.Decision != types.DecisionAccept {
		t.Error("t1 should be accept")
	}
	if results[1].Verdict.Decision != types.DecisionBlocking {
		t.Error("t2 should be blocking")
	}
}

func TestProgressChanEmitsEvents(t *testing.T) {
	bus := NewBus()
	bus.Register(&mockAgent{
		card:    AgentCard{ID: "qa-1", Role: types.RoleQA},
		verdict: types.Verdict{Decision: types.DecisionAccept},
	})
	go func() {
		time.Sleep(10 * time.Millisecond)
		_, _ = bus.Delegate(context.Background(), Task{TaskID: "t1", Role: types.RoleQA})
	}()
	select {
	case ev := <-bus.ProgressChan():
		if ev.TaskID != "t1" {
			t.Errorf("expected task t1, got %s", ev.TaskID)
		}
	case <-time.After(2 * time.Second):
		t.Error("timeout waiting for progress event")
	}
}

func TestCount(t *testing.T) {
	bus := NewBus()
	if bus.Count() != 0 {
		t.Error("empty bus should have 0 agents")
	}
	bus.Register(&mockAgent{card: AgentCard{ID: "a", Role: types.RoleQA}})
	if bus.Count() != 1 {
		t.Error("should have 1 agent")
	}
}
