package a2a

import (
	"context"
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"

	"github.com/agentcodinglab/aicodingagentteam/internal/types"
)

// redisMockAgent is a test agent for RedisBus.
type redisMockAgent struct {
	card    AgentCard
	verdict types.Verdict
	err     error
}

func (m *redisMockAgent) Card() AgentCard { return m.card }
func (m *redisMockAgent) Execute(ctx context.Context, task Task) (Result, error) {
	v := m.verdict
	v.TaskID = task.TaskID
	v.Role = task.Role
	return Result{TaskID: task.TaskID, Verdict: v}, m.err
}
func (m *redisMockAgent) Status(ctx context.Context) string { return "idle" }

func newTestRedisBus(t *testing.T) (*RedisBus, *miniredis.Miniredis) {
	t.Helper()
	mr := miniredis.RunT(t)
	bus, err := NewRedisBus("redis://"+mr.Addr(), nil)
	if err != nil {
		t.Fatalf("NewRedisBus failed: %v", err)
	}
	t.Cleanup(func() { _ = bus.Close() })
	return bus, mr
}

func TestRedisBus_RegisterAndDiscover(t *testing.T) {
	bus, _ := newTestRedisBus(t)
	bus.Register(&redisMockAgent{
		card: AgentCard{ID: "qa-1", Name: "QA", Role: types.RoleQA},
	})
	// Allow subscribe goroutine to start
	time.Sleep(50 * time.Millisecond)
	if !bus.IsRegistered(types.RoleQA) {
		t.Error("QA should be registered")
	}
	if bus.Count() != 1 {
		t.Errorf("expected count 1, got %d", bus.Count())
	}
	cards := bus.Discover()
	if len(cards) != 1 || cards[0].Role != types.RoleQA {
		t.Errorf("unexpected cards: %+v", cards)
	}
}

func TestRedisBus_DelegateUnregisteredRole(t *testing.T) {
	bus, _ := newTestRedisBus(t)
	task := Task{TaskID: "t1", Role: types.RoleSecurity}
	_, err := bus.Delegate(context.Background(), task)
	if err == nil {
		t.Error("should return error for unregistered role")
	}
	if !strings.Contains(err.Error(), "no agent registered") {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestRedisBus_DelegateRoutesByRole(t *testing.T) {
	bus, _ := newTestRedisBus(t)
	bus.Register(&redisMockAgent{
		card:    AgentCard{ID: "qa-1", Role: types.RoleQA},
		verdict: types.Verdict{Decision: types.DecisionAccept},
	})
	result, err := bus.Delegate(context.Background(), Task{
		TaskID:  "t1",
		Role:    types.RoleQA,
		Payload: TaskPayload{Instruction: "test"},
	})
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

func TestRedisBus_DelegateAgentErrorReturnsBlocking(t *testing.T) {
	bus, _ := newTestRedisBus(t)
	bus.Register(&redisMockAgent{
		card: AgentCard{ID: "qa-1", Role: types.RoleQA},
		err:  context.DeadlineExceeded,
	})
	result, _ := bus.Delegate(context.Background(), Task{TaskID: "t1", Role: types.RoleQA})
	if result.Verdict.Decision != types.DecisionBlocking {
		t.Error("agent error should produce blocking verdict")
	}
	if result.TaskID != "t1" {
		t.Errorf("expected task t1, got %s", result.TaskID)
	}
}

func TestRedisBus_DelegateBlockingVerdict(t *testing.T) {
	bus, _ := newTestRedisBus(t)
	bus.Register(&redisMockAgent{
		card:    AgentCard{ID: "qa-1", Role: types.RoleQA},
		verdict: types.Verdict{Decision: types.DecisionBlocking},
	})
	result, err := bus.Delegate(context.Background(), Task{TaskID: "t1", Role: types.RoleQA})
	if err != nil {
		t.Fatalf("delegate failed: %v", err)
	}
	if result.Verdict.Decision != types.DecisionBlocking {
		t.Error("expected blocking verdict")
	}
}

func TestRedisBus_DelegateParallel(t *testing.T) {
	bus, _ := newTestRedisBus(t)
	bus.Register(&redisMockAgent{
		card:    AgentCard{ID: "qa-1", Role: types.RoleQA},
		verdict: types.Verdict{Decision: types.DecisionAccept},
	})
	bus.Register(&redisMockAgent{
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

func TestRedisBus_ProgressChan(t *testing.T) {
	bus, _ := newTestRedisBus(t)
	bus.Register(&redisMockAgent{
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

func TestRedisBus_HandleRedisMessage_InvalidJSON(t *testing.T) {
	bus, _ := newTestRedisBus(t)
	// Should not panic on invalid JSON
	bus.handleRedisMessage(context.Background(), "not-json{")
}

func TestRedisBus_HandleRedisMessage_UnregisteredRole(t *testing.T) {
	bus, mr := newTestRedisBus(t)
	bus.handleRedisMessage(context.Background(), `{"task_id":"t1","role":"frontend","payload":{"instruction":"test"},"deadline":"2026-01-01T00:00:00Z"}`)
	_ = mr
}

func TestRedisBus_HandleRedisMessage_AgentError(t *testing.T) {
	bus, _ := newTestRedisBus(t)
	bus.Register(&redisMockAgent{
		card: AgentCard{ID: "qa-1", Role: types.RoleQA},
		err:  context.DeadlineExceeded,
	})
	taskJSON, _ := json.Marshal(Task{TaskID: "t1", Role: types.RoleQA})
	bus.handleRedisMessage(context.Background(), string(taskJSON))
}

func TestRedisBus_NewRedisBus_InvalidURL(t *testing.T) {
	_, err := NewRedisBus("not-a-valid-url", nil)
	if err == nil {
		t.Error("should fail on invalid URL")
	}
}

func TestRedisBus_NewRedisBus_UnreachableHost(t *testing.T) {
	_, err := NewRedisBus("redis://127.0.0.1:1", nil)
	if err == nil {
		t.Error("should fail on unreachable host")
	}
}

func TestRedisBus_LogAudit_NilAudit(t *testing.T) {
	bus, _ := newTestRedisBus(t)
	// Should not panic with nil audit
	bus.logAudit("t1", "qa", "test", "detail")
	bus.emitProgress("t1", "p", "running", "")
}

func TestRedisBus_IsRegistered_False(t *testing.T) {
	bus, _ := newTestRedisBus(t)
	if bus.IsRegistered(types.RoleArchitect) {
		t.Error("unregistered role should return false")
	}
}

func TestRedisBus_Register_MultipleAgents(t *testing.T) {
	bus, _ := newTestRedisBus(t)
	bus.Register(&redisMockAgent{card: AgentCard{ID: "qa", Role: types.RoleQA}})
	bus.Register(&redisMockAgent{card: AgentCard{ID: "pm", Role: types.RolePM}})
	bus.Register(&redisMockAgent{card: AgentCard{ID: "arch", Role: types.RoleArchitect}})
	if bus.Count() != 3 {
		t.Errorf("expected 3 agents, got %d", bus.Count())
	}
	if !bus.IsRegistered(types.RoleQA) || !bus.IsRegistered(types.RolePM) || !bus.IsRegistered(types.RoleArchitect) {
		t.Error("all three roles should be registered")
	}
}
