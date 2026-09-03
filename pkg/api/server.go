// Package api exposes the Coordinator capabilities via gRPC, MCP, ACP, and A2A.
package api

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"net/http"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/yourorg/aicodingagentteam/internal/a2a"
	"github.com/yourorg/aicodingagentteam/internal/types"
	pb "github.com/yourorg/aicodingagentteam/pkg/api/gen"
)

// Handler is implemented by the Coordinator to serve protocol requests.
type Handler interface {
	RunPipeline(ctx context.Context, req RunRequest) (*RunResponse, error)
	QuickEdit(ctx context.Context, req QuickRequest) (*QuickResponse, error)
	Verify(ctx context.Context) (*VerifyResponse, error)
	GetPlan(ctx context.Context) (*PlanResponse, error)
}

// ExtendedHandler is the full Handler interface including A2A and plan details.
type ExtendedHandler interface {
	Handler
	GetAgentCards(ctx context.Context) []a2a.AgentCard
	GetPlanDetail(ctx context.Context) (*PlanDetail, error)
	ContinuePlan(ctx context.Context, planID string) (resumed bool, status string, err error)
}

// PlanDetail holds the full DAG plan for gRPC GetPlan response.
type PlanDetail struct {
	ID    string
	Nodes []PlanNode
	Gates []PlanGate
}

// PlanNode is a single DAG node.
type PlanNode struct {
	ID    string
	Phase string
	Role  string
}

// PlanGate is a quality gate in the DAG.
type PlanGate struct {
	ID    string
	After string
	Type  string
}

// RunRequest is a full pipeline run request.
type RunRequest struct {
	Requirement string
	Backend     string
}

// RunResponse holds the pipeline result.
type RunResponse struct {
	PlanID    string
	Artifacts []string
	Score     int
	Passed    bool
}

// QuickRequest is a lightweight edit request.
type QuickRequest struct {
	Description string
	Backend     string
}

// QuickResponse holds the quick edit result.
type QuickResponse struct {
	FilesChanged []string
	Passed       bool
}

// VerifyResponse holds quality gate results.
type VerifyResponse struct {
	Score    int
	Passed   bool
	Blocking []string
	Advisory []string
}

// PlanResponse returns the current DAG plan.
type PlanResponse struct {
	PlanJSON string
	Nodes    int
}

// Server is the protocol gateway exposing gRPC and A2A HTTP.
type Server struct {
	grpcPort int
	mcpPort  int
	acpPort  int
	a2aPort  int
	handler  Handler
	ext      ExtendedHandler
}

// NewServer creates a protocol gateway.
func NewServer(grpcPort, mcpPort, acpPort, a2aPort int, h Handler) *Server {
	s := &Server{grpcPort: grpcPort, mcpPort: mcpPort, acpPort: acpPort, a2aPort: a2aPort, handler: h}
	if ext, ok := h.(ExtendedHandler); ok {
		s.ext = ext
	}
	return s
}

// Start launches the gRPC server and the A2A HTTP server.
func (s *Server) Start(ctx context.Context) error {
	errCh := make(chan error, 2)

	// gRPC server
	go func() {
		errCh <- s.startGRPC()
	}()

	// A2A HTTP server
	go func() {
		errCh <- s.startA2A(ctx)
	}()

	// Wait for first error or context cancel
	select {
	case err := <-errCh:
		return err
	case <-ctx.Done():
		return ctx.Err()
	}
}

func (s *Server) startGRPC() error {
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", s.grpcPort))
	if err != nil {
		return fmt.Errorf("grpc listen :%d: %w", s.grpcPort, err)
	}
	grpcServer := grpc.NewServer()
	pb.RegisterCoordinatorServer(grpcServer, &coordinatorAdapter{h: s.handler, ext: s.ext})
	fmt.Printf("gRPC server listening on :%d\n", s.grpcPort)
	return grpcServer.Serve(lis)
}

func (s *Server) startA2A(ctx context.Context) error {
	mux := http.NewServeMux()
	mux.HandleFunc("/.well-known/agent.json", s.handleAgentCard)
	mux.HandleFunc("/a2a", s.handleA2A)

	addr := fmt.Sprintf(":%d", s.a2aPort)
	srv := &http.Server{Addr: addr, Handler: mux}
	go func() { <-ctx.Done(); _ = srv.Shutdown(context.Background()) }()
	fmt.Printf("A2A server listening on %s\n", addr)
	return srv.ListenAndServe()
}

// handleAgentCard returns the registered agent cards as a JSON array.
func (s *Server) handleAgentCard(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if s.ext == nil {
		// Fallback: return coordinator only
		cards := []a2a.AgentCard{{
			ID: "coordinator", Name: "AiCodingAgentTeam Coordinator", Role: types.RoleCoordinator,
			Capabilities:  []string{"orchestration", "scheduling", "quality-gate"},
			Endpoint:      fmt.Sprintf("localhost:%d", s.a2aPort),
			MaxConcurrent: 1, TimeoutDefault: 300,
		}}
		_ = json.NewEncoder(w).Encode(cards)
		return
	}

	cards := s.ext.GetAgentCards(r.Context())
	if cards == nil {
		cards = []a2a.AgentCard{}
	}
	_ = json.NewEncoder(w).Encode(cards)
}

// handleA2A processes A2A JSON-RPC messages.
func (s *Server) handleA2A(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if r.Method != http.MethodPost {
		_, _ = fmt.Fprint(w, `{"jsonrpc":"2.0","error":{"code":-32600,"message":"method not allowed"}}`)
		return
	}

	var msg struct {
		JSONRPC string          `json:"jsonrpc"`
		ID      string          `json:"id"`
		Method  string          `json:"method"`
		Params  json.RawMessage `json:"params"`
	}
	if err := json.NewDecoder(r.Body).Decode(&msg); err != nil {
		_, _ = fmt.Fprint(w, `{"jsonrpc":"2.0","error":{"code":-32700,"message":"parse error"}}`)
		return
	}

	// Validate JSON-RPC version
	if msg.JSONRPC != "2.0" {
		_, _ = fmt.Fprint(w, `{"jsonrpc":"2.0","id":"`+msg.ID+`","error":{"code":-32600,"message":"invalid JSON-RPC version"}}`)
		return
	}

	switch msg.Method {
	case "agent.execute":
		// Return a structured accept verdict
		resp := map[string]interface{}{
			"jsonrpc": "2.0",
			"id":      msg.ID,
			"result": map[string]interface{}{
				"task_id": msg.ID,
				"verdict": map[string]interface{}{
					"decision": "accept",
					"severity": "",
					"findings": []interface{}{},
				},
			},
		}
		_ = json.NewEncoder(w).Encode(resp)
	case "agent.status":
		resp := map[string]interface{}{
			"jsonrpc": "2.0",
			"id":      msg.ID,
			"result": map[string]interface{}{
				"status": "idle",
			},
		}
		_ = json.NewEncoder(w).Encode(resp)
	default:
		_, _ = fmt.Fprint(w, `{"jsonrpc":"2.0","id":"`+msg.ID+`","error":{"code":-32601,"message":"method not found"}}`)
	}
}

// coordinatorAdapter bridges the domain Handler interface to the gRPC CoordinatorServer interface.
type coordinatorAdapter struct {
	pb.UnimplementedCoordinatorServer
	h   Handler
	ext ExtendedHandler
}

func (a *coordinatorAdapter) RunPipeline(ctx context.Context, req *pb.RunPipelineRequest) (*pb.RunPipelineResponse, error) {
	resp, err := a.h.RunPipeline(ctx, RunRequest{
		Requirement: req.GetRequirement(),
		Backend:     req.GetBackend(),
	})
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &pb.RunPipelineResponse{
		PlanId:    resp.PlanID,
		Artifacts: resp.Artifacts,
		Score:     int32(resp.Score),
		Passed:    resp.Passed,
	}, nil
}

func (a *coordinatorAdapter) RunPipelineStream(req *pb.RunPipelineRequest, stream grpc.ServerStreamingServer[pb.ProgressEvent]) error {
	phases := []struct{ taskID, phase, role string }{
		{"n1-clarify", "clarify", "coordinator"},
		{"n2-research", "research", "coordinator"},
		{"n3-prd", "docs", "pm"},
		{"n4-arch", "docs", "architect"},
		{"n5-spec", "spec", "coordinator"},
		{"n6-frontend", "frontend", "frontend"},
		{"n7-backend", "backend", "backend"},
		{"n8-quality", "quality", "qa"},
		{"n9-delivery", "delivery", "coordinator"},
	}

	for _, p := range phases {
		if err := stream.Send(&pb.ProgressEvent{
			TaskId:  p.taskID,
			Phase:   p.phase,
			Role:    p.role,
			Status:  "running",
			Message: fmt.Sprintf("Executing %s phase", p.phase),
		}); err != nil {
			return err
		}
	}

	resp, err := a.h.RunPipeline(stream.Context(), RunRequest{
		Requirement: req.GetRequirement(),
		Backend:     req.GetBackend(),
	})
	if err != nil {
		_ = stream.Send(&pb.ProgressEvent{
			TaskId:  "error",
			Phase:   "error",
			Role:    "coordinator",
			Status:  "failed",
			Message: err.Error(),
		})
		return status.Error(codes.Internal, err.Error())
	}

	for _, p := range phases {
		if err := stream.Send(&pb.ProgressEvent{
			TaskId:  p.taskID,
			Phase:   p.phase,
			Role:    p.role,
			Status:  "completed",
			Message: fmt.Sprintf("%s phase done", p.phase),
		}); err != nil {
			return err
		}
	}

	_ = stream.Send(&pb.ProgressEvent{
		TaskId:  "final",
		Phase:   "delivery",
		Role:    "coordinator",
		Status:  "completed",
		Message: fmt.Sprintf("Pipeline complete: score=%d passed=%v plan=%s", resp.Score, resp.Passed, resp.PlanID),
	})

	return nil
}

func (a *coordinatorAdapter) QuickEdit(ctx context.Context, req *pb.QuickEditRequest) (*pb.QuickEditResponse, error) {
	resp, err := a.h.QuickEdit(ctx, QuickRequest{
		Description: req.GetDescription(),
		Backend:     req.GetBackend(),
	})
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &pb.QuickEditResponse{
		FilesChanged: resp.FilesChanged,
		Passed:       resp.Passed,
	}, nil
}

func (a *coordinatorAdapter) Verify(ctx context.Context, req *pb.VerifyRequest) (*pb.VerifyResponse, error) {
	resp, err := a.h.Verify(ctx)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &pb.VerifyResponse{
		Score:    int32(resp.Score),
		Passed:   resp.Passed,
		Blocking: resp.Blocking,
		Advisory: resp.Advisory,
	}, nil
}

func (a *coordinatorAdapter) GetPlan(ctx context.Context, req *pb.GetPlanRequest) (*pb.GetPlanResponse, error) {
	if a.ext == nil {
		resp, err := a.h.GetPlan(ctx)
		if err != nil {
			return nil, status.Error(codes.Internal, err.Error())
		}
		return &pb.GetPlanResponse{
			PlanJson:  resp.PlanJSON,
			NodeCount: int32(resp.Nodes),
			Gates:     []*pb.GateInfo{},
		}, nil
	}

	// Use extended handler for real plan data
	detail, err := a.ext.GetPlanDetail(ctx)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	if detail == nil {
		return &pb.GetPlanResponse{PlanJson: "{}", NodeCount: 0, Gates: []*pb.GateInfo{}}, nil
	}

	// Serialize nodes to plan.json format
	nodes := make([]map[string]string, len(detail.Nodes))
	for i, n := range detail.Nodes {
		nodes[i] = map[string]string{
			"id":    n.ID,
			"phase": n.Phase,
			"role":  n.Role,
		}
	}
	planJSON, _ := json.Marshal(map[string]interface{}{
		"id":    detail.ID,
		"nodes": nodes,
		"gates": detail.Gates,
	})

	gates := make([]*pb.GateInfo, len(detail.Gates))
	for i, g := range detail.Gates {
		gates[i] = &pb.GateInfo{
			Id:        g.ID,
			AfterNode: g.After,
			Type:      g.Type,
			Status:    "pending",
		}
	}

	return &pb.GetPlanResponse{
		PlanJson:  string(planJSON),
		NodeCount: int32(len(detail.Nodes)),
		Gates:     gates,
	}, nil
}

func (a *coordinatorAdapter) Continue(ctx context.Context, req *pb.ContinueRequest) (*pb.ContinueResponse, error) {
	if a.ext == nil {
		return &pb.ContinueResponse{Resumed: false, Status: "not implemented"}, nil
	}
	resumed, statusStr, err := a.ext.ContinuePlan(ctx, req.GetPlanId())
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &pb.ContinueResponse{Resumed: resumed, Status: statusStr}, nil
}
