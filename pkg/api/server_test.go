package api

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"google.golang.org/grpc/metadata"

	"github.com/agentcodinglab/aicodingagentteam/internal/a2a"
	"github.com/agentcodinglab/aicodingagentteam/internal/types"
	pb "github.com/agentcodinglab/aicodingagentteam/pkg/api/gen"
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

func TestToPBDetails_Nil(t *testing.T) {
	if v := toPBDetails(nil); v != nil {
		t.Errorf("expected nil, got %v", v)
	}
}

func TestToPBDetails_Converts(t *testing.T) {
	in := []CheckSummary{
		{Name: "build", Status: "pass"},
		{Name: "vet", Status: "fail", Output: "syntax error"},
	}
	out := toPBDetails(in)
	if len(out) != 2 {
		t.Fatalf("expected 2, got %d", len(out))
	}
	if out[0].Name != "build" || out[0].Status != "pass" {
		t.Errorf("unexpected first: %+v", out[0])
	}
	if out[1].Output != "syntax error" {
		t.Errorf("unexpected second output: %s", out[1].Output)
	}
}

// TestCoordinatorAdapter_RPCCalls verifies the adapter methods translate
// Handler responses to pb responses, without starting a gRPC server.
func TestCoordinatorAdapter_RPCCalls(t *testing.T) {
	a := &coordinatorAdapter{h: &mockHandler{}}

	ctx := context.Background()

	// RunPipeline
	resp, err := a.RunPipeline(ctx, &pb.RunPipelineRequest{Requirement: "test", Backend: "codex"})
	if err != nil {
		t.Fatalf("RunPipeline: %v", err)
	}
	if resp.PlanId != "test" || !resp.Passed || resp.Score != 100 {
		t.Errorf("unexpected RunPipeline response: %+v", resp)
	}

	// QuickEdit
	qresp, err := a.QuickEdit(ctx, &pb.QuickEditRequest{Description: "fix", Backend: "codex"})
	if err != nil {
		t.Fatalf("QuickEdit: %v", err)
	}
	if !qresp.Passed || len(qresp.FilesChanged) != 1 {
		t.Errorf("unexpected QuickEdit response: %+v", qresp)
	}

	// Verify
	vresp, err := a.Verify(ctx, &pb.VerifyRequest{})
	if err != nil {
		t.Fatalf("Verify: %v", err)
	}
	if vresp.Score != 90 || !vresp.Passed {
		t.Errorf("unexpected Verify response: %+v", vresp)
	}

	// GetPlan
	presp, err := a.GetPlan(ctx, &pb.GetPlanRequest{})
	if err != nil {
		t.Fatalf("GetPlan: %v", err)
	}
	if presp.PlanJson != "{}" {
		t.Errorf("unexpected PlanJson: %s", presp.PlanJson)
	}
}

func TestCoordinatorAdapter_ExtendedHandler(t *testing.T) {
	a := &coordinatorAdapter{h: &mockHandler{}, ext: &mockExtendedHandler{}}
	ctx := context.Background()

	// GetPlan with ExtendedHandler
	resp, err := a.GetPlan(ctx, &pb.GetPlanRequest{})
	if err != nil {
		t.Fatalf("GetPlan: %v", err)
	}
	if resp.NodeCount != 2 {
		t.Errorf("expected 2 nodes, got %d", resp.NodeCount)
	}
	if len(resp.Gates) != 1 {
		t.Errorf("expected 1 gate, got %d", len(resp.Gates))
	}

	// Continue
	cresp, err := a.Continue(ctx, &pb.ContinueRequest{PlanId: "p1"})
	if err != nil {
		t.Fatalf("Continue: %v", err)
	}
	if !cresp.Resumed || cresp.Status != "resumed" {
		t.Errorf("unexpected Continue response: %+v", cresp)
	}
}

func TestCoordinatorAdapter_NoExtendedHandler(t *testing.T) {
	a := &coordinatorAdapter{h: &mockHandler{}}
	ctx := context.Background()

	// GetPlan should fall back to basic Handler
	resp, err := a.GetPlan(ctx, &pb.GetPlanRequest{})
	if err != nil {
		t.Fatalf("GetPlan: %v", err)
	}
	if resp.PlanJson != "{}" {
		t.Errorf("expected fallback PlanJson, got %s", resp.PlanJson)
	}

	// Continue should return not implemented
	cresp, err := a.Continue(ctx, &pb.ContinueRequest{PlanId: "p1"})
	if err != nil {
		t.Fatalf("Continue: %v", err)
	}
	if cresp.Resumed {
		t.Error("should not resume without ExtendedHandler")
	}
}

// mockStream implements grpc.ServerStreamingServer for testing RunPipelineStream.
type mockStream struct {
	events []*pb.ProgressEvent
	ctx    context.Context
}

func (m *mockStream) Send(ev *pb.ProgressEvent) error {
	m.events = append(m.events, ev)
	return nil
}
func (m *mockStream) Context() context.Context        { return m.ctx }
func (m *mockStream) SetHeader(md metadata.MD) error  { return nil }
func (m *mockStream) SendHeader(md metadata.MD) error { return nil }
func (m *mockStream) SetTrailer(md metadata.MD)       {}
func (m *mockStream) SendMsg(msg interface{}) error   { return nil }
func (m *mockStream) RecvMsg(msg interface{}) error   { return nil }

func TestRunPipelineStream_EmitsProgress(t *testing.T) {
	a := &coordinatorAdapter{h: &mockHandler{}}
	stream := &mockStream{ctx: context.Background()}
	err := a.RunPipelineStream(&pb.RunPipelineRequest{Requirement: "build app", Backend: "codex"}, stream)
	if err != nil {
		t.Fatalf("RunPipelineStream: %v", err)
	}
	if len(stream.events) < 9 {
		t.Fatalf("expected at least 9 progress events, got %d", len(stream.events))
	}
	if stream.events[0].Phase != "clarify" {
		t.Errorf("expected first phase clarify, got %s", stream.events[0].Phase)
	}
	last := stream.events[len(stream.events)-1]
	if last.Phase != "delivery" {
		t.Errorf("expected last phase delivery, got %s", last.Phase)
	}
}

func TestStart_AllServicesGracefulShutdown(t *testing.T) {
	h := &mockExtendedHandler{}
	srv := NewServer(0, 0, 0, 0, h)
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan error, 1)
	go func() {
		done <- srv.Start(ctx)
	}()
	// Wait for server to start serving
	time.Sleep(100 * time.Millisecond)
	cancel()
	select {
	case err := <-done:
		if err != nil && err != context.Canceled {
			t.Fatalf("Start returned error: %v", err)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("Start did not return after cancel")
	}
}
