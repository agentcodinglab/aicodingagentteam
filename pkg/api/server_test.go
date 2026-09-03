package api

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/yourorg/aicodingagentteam/internal/a2a"
	"github.com/yourorg/aicodingagentteam/internal/types"
)

// mockHandler implements Handler for testing.
type mockHandler struct{}

func (m *mockHandler) RunPipeline(ctx context.Context, req RunRequest) (*RunResponse, error) {
	return &RunResponse{PlanID: "test", Score: 100, Passed: true}, nil
}
func (m *mockHandler) QuickEdit(ctx context.Context, req QuickRequest) (*QuickResponse, error) {
	return &QuickResponse{FilesChanged: []string{"a.go"}, Passed: true}, nil
}
func (m *mockHandler) Verify(ctx context.Context) (*VerifyResponse, error) {
	return &VerifyResponse{Score: 90, Passed: true}, nil
}
func (m *mockHandler) GetPlan(ctx context.Context) (*PlanResponse, error) {
	return &PlanResponse{PlanJSON: "{}", Nodes: 0}, nil
}

// mockExtendedHandler implements ExtendedHandler for testing.
type mockExtendedHandler struct {
	mockHandler
	cards []a2a.AgentCard
}

func (m *mockExtendedHandler) GetAgentCards(ctx context.Context) []a2a.AgentCard {
	return m.cards
}
func (m *mockExtendedHandler) ContinuePlan(ctx context.Context, planID string) (bool, string, error) {
	return true, "resumed", nil
}
func (m *mockExtendedHandler) GetPlanDetail(ctx context.Context) (*PlanDetail, error) {
	return &PlanDetail{
		ID: "test-plan",
		Nodes: []PlanNode{
			{ID: "n1", Phase: "clarify", Role: "coordinator"},
			{ID: "n2", Phase: "docs", Role: "pm"},
		},
		Gates: []PlanGate{
			{ID: "g1", After: "n2", Type: "human"},
		},
	}, nil
}

func TestAgentCardReturnsRegisteredAgents(t *testing.T) {
	cards := []a2a.AgentCard{
		{ID: "agent-pm", Name: "PM", Role: types.RolePM, Capabilities: []string{"prd"}, Endpoint: "localhost:8083", MaxConcurrent: 1, TimeoutDefault: 300},
		{ID: "agent-qa", Name: "QA", Role: types.RoleQA, Capabilities: []string{"test"}, Endpoint: "localhost:8083", MaxConcurrent: 1, TimeoutDefault: 300},
	}
	h := &mockExtendedHandler{cards: cards}
	srv := NewServer(0, 0, 0, 8083, h)

	req := httptest.NewRequest(http.MethodGet, "/.well-known/agent.json", nil)
	rec := httptest.NewRecorder()
	srv.handleAgentCard(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}

	var result []a2a.AgentCard
	if err := json.Unmarshal(rec.Body.Bytes(), &result); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}
	if len(result) != 2 {
		t.Fatalf("expected 2 cards, got %d", len(result))
	}
	if result[0].ID != "agent-pm" {
		t.Errorf("expected agent-pm, got %s", result[0].ID)
	}
}

func TestAgentCardFallbackWithoutExtendedHandler(t *testing.T) {
	srv := NewServer(0, 0, 0, 8083, &mockHandler{})

	req := httptest.NewRequest(http.MethodGet, "/.well-known/agent.json", nil)
	rec := httptest.NewRecorder()
	srv.handleAgentCard(rec, req)

	var result []a2a.AgentCard
	if err := json.Unmarshal(rec.Body.Bytes(), &result); err != nil {
		t.Fatalf("failed to parse: %v", err)
	}
	if len(result) != 1 {
		t.Fatalf("expected 1 fallback card, got %d", len(result))
	}
	if result[0].ID != "coordinator" {
		t.Errorf("expected coordinator, got %s", result[0].ID)
	}
}

func TestA2ARejectsGET(t *testing.T) {
	srv := NewServer(0, 0, 0, 8083, &mockHandler{})

	req := httptest.NewRequest(http.MethodGet, "/a2a", nil)
	rec := httptest.NewRecorder()
	srv.handleA2A(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	if !strings.Contains(rec.Body.String(), "method not allowed") {
		t.Error("GET should be rejected")
	}
}

func TestA2AExecuteReturnsAcceptVerdict(t *testing.T) {
	srv := NewServer(0, 0, 0, 8083, &mockHandler{})

	body := strings.NewReader(`{"jsonrpc":"2.0","id":"task-1","method":"agent.execute","params":{}}`)
	req := httptest.NewRequest(http.MethodPost, "/a2a", body)
	rec := httptest.NewRecorder()
	srv.handleA2A(rec, req)

	var result map[string]interface{}
	if err := json.Unmarshal(rec.Body.Bytes(), &result); err != nil {
		t.Fatalf("failed to parse: %v", err)
	}
	if result["jsonrpc"] != "2.0" {
		t.Error("expected jsonrpc 2.0")
	}
	resultMap, ok := result["result"].(map[string]interface{})
	if !ok {
		t.Fatal("expected result object")
	}
	verdict, ok := resultMap["verdict"].(map[string]interface{})
	if !ok {
		t.Fatal("expected verdict object")
	}
	if verdict["decision"] != "accept" {
		t.Errorf("expected accept, got %v", verdict["decision"])
	}
}

func TestA2ARejectsInvalidJSONRPCVersion(t *testing.T) {
	srv := NewServer(0, 0, 0, 8083, &mockHandler{})

	body := strings.NewReader(`{"jsonrpc":"1.0","id":"t1","method":"agent.execute","params":{}}`)
	req := httptest.NewRequest(http.MethodPost, "/a2a", body)
	rec := httptest.NewRecorder()
	srv.handleA2A(rec, req)

	if !strings.Contains(rec.Body.String(), "invalid JSON-RPC version") {
		t.Error("should reject invalid JSON-RPC version")
	}
}

func TestA2ARejectsUnknownMethod(t *testing.T) {
	srv := NewServer(0, 0, 0, 8083, &mockHandler{})

	body := strings.NewReader(`{"jsonrpc":"2.0","id":"t1","method":"unknown.method","params":{}}`)
	req := httptest.NewRequest(http.MethodPost, "/a2a", body)
	rec := httptest.NewRecorder()
	srv.handleA2A(rec, req)

	if !strings.Contains(rec.Body.String(), "method not found") {
		t.Error("should reject unknown method")
	}
}

func TestAgentCardHasRequiredFields(t *testing.T) {
	cards := []a2a.AgentCard{
		{ID: "test", Name: "Test Agent", Role: types.RoleQA, Capabilities: []string{"test"}, Endpoint: "localhost:8083", MaxConcurrent: 1, TimeoutDefault: 300},
	}
	h := &mockExtendedHandler{cards: cards}
	srv := NewServer(0, 0, 0, 8083, h)

	req := httptest.NewRequest(http.MethodGet, "/.well-known/agent.json", nil)
	rec := httptest.NewRecorder()
	srv.handleAgentCard(rec, req)

	var result []map[string]interface{}
	_ = json.Unmarshal(rec.Body.Bytes(), &result)
	if len(result) == 0 {
		t.Fatal("expected at least 1 card")
	}
	required := []string{"id", "name", "role", "capabilities", "endpoint"}
	for _, field := range required {
		if result[0][field] == nil {
			t.Errorf("missing required field: %s", field)
		}
	}
}
