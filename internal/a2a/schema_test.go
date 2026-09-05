package a2a

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/agentcodinglab/aicodingagentteam/internal/types"
)

func TestAgentCard_JSONSchema(t *testing.T) {
	card := AgentCard{
		ID:             "qa-1",
		Name:           "QA Agent",
		Role:           types.RoleQA,
		Capabilities:   []string{"review", "quality"},
		Endpoint:       "localhost:8083",
		MaxConcurrent:  2,
		TimeoutDefault: 300,
	}
	data, err := json.Marshal(card)
	if err != nil {
		t.Fatalf("marshal failed: %v", err)
	}

	var m map[string]interface{}
	if err := json.Unmarshal(data, &m); err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}

	requiredFields := []string{"id", "name", "role", "capabilities", "endpoint", "max_concurrent", "timeout_default"}
	for _, field := range requiredFields {
		if _, ok := m[field]; !ok {
			t.Errorf("AgentCard JSON missing required field: %s", field)
		}
	}
}

func TestTask_JSONSchema(t *testing.T) {
	task := Task{
		TaskID:  "t1",
		Role:    types.RoleQA,
		Payload: TaskPayload{Artifacts: []string{"output/app.md"}, Instruction: "review PRD"},
	}
	data, err := json.Marshal(task)
	if err != nil {
		t.Fatalf("marshal failed: %v", err)
	}

	var m map[string]interface{}
	if err := json.Unmarshal(data, &m); err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}

	requiredFields := []string{"task_id", "role", "payload", "deadline"}
	for _, field := range requiredFields {
		if _, ok := m[field]; !ok {
			t.Errorf("Task JSON missing required field: %s", field)
		}
	}
}

func TestResult_JSONSchema(t *testing.T) {
	result := Result{
		TaskID: "t1",
		Verdict: types.Verdict{
			TaskID:   "t1",
			Role:     types.RoleQA,
			Decision: types.DecisionAccept,
			Findings: []types.Finding{{Check: "lint", Status: "pass", Detail: "no issues"}},
		},
	}
	data, err := json.Marshal(result)
	if err != nil {
		t.Fatalf("marshal failed: %v", err)
	}

	var m map[string]interface{}
	if err := json.Unmarshal(data, &m); err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}

	requiredFields := []string{"task_id", "verdict"}
	for _, field := range requiredFields {
		if _, ok := m[field]; !ok {
			t.Errorf("Result JSON missing required field: %s", field)
		}
	}
}

func TestProgressEvent_JSONSchema(t *testing.T) {
	ev := ProgressEvent{
		TaskID:  "t1",
		Phase:   "quality",
		Status:  "running",
		Message: "linting",
	}
	data, err := json.Marshal(ev)
	if err != nil {
		t.Fatalf("marshal failed: %v", err)
	}

	var m map[string]interface{}
	if err := json.Unmarshal(data, &m); err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}

	requiredFields := []string{"task_id", "phase", "status", "message"}
	for _, field := range requiredFields {
		if _, ok := m[field]; !ok {
			t.Errorf("ProgressEvent JSON missing required field: %s", field)
		}
	}
}

func TestAgentCard_RoundTrip(t *testing.T) {
	original := AgentCard{
		ID:             "pm-1",
		Name:           "PM Agent",
		Role:           types.RolePM,
		Capabilities:   []string{"docs", "planning"},
		Endpoint:       "agent-pm:8083",
		MaxConcurrent:  1,
		TimeoutDefault: 120,
	}
	data, _ := json.Marshal(original)
	var decoded AgentCard
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}
	if decoded.ID != original.ID {
		t.Errorf("ID mismatch: %s vs %s", decoded.ID, original.ID)
	}
	if decoded.Role != original.Role {
		t.Errorf("Role mismatch: %s vs %s", decoded.Role, original.Role)
	}
	if len(decoded.Capabilities) != len(original.Capabilities) {
		t.Errorf("Capabilities length mismatch: %d vs %d", len(decoded.Capabilities), len(original.Capabilities))
	}
}

// Ensure Agent interface is satisfied by mockAgent
func TestAgentInterface_Satisfied(t *testing.T) {
	var _ Agent = &mockAgent{
		card: AgentCard{ID: "test", Role: types.RoleQA},
	}
}

// Test bus nil-safety with unregistered role
func TestDelegateNilSafe(t *testing.T) {
	bus := NewBus()
	result, err := bus.Delegate(context.Background(), Task{TaskID: "t1", Role: types.RoleFrontend})
	if err == nil {
		t.Error("expected error for unregistered role")
	}
	if result.TaskID != "" {
		t.Error("expected empty result for unregistered role")
	}
}
