package api

import (
	"context"
	"fmt"
	"net"
	"testing"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"

	pb "github.com/agentcodinglab/aicodingagentteam/pkg/api/gen"

	"github.com/agentcodinglab/aicodingagentteam/internal/a2a"
)

// fakeHandler implements Handler for gRPC integration testing.
type fakeHandler struct{}

func (f *fakeHandler) RunPipeline(ctx context.Context, req RunRequest) (*RunResponse, error) {
	return &RunResponse{PlanID: "test-plan", Artifacts: []string{"a.go"}, Score: 95, Passed: true}, nil
}
func (f *fakeHandler) QuickEdit(ctx context.Context, req QuickRequest) (*QuickResponse, error) {
	return &QuickResponse{FilesChanged: []string{"main.go"}, Passed: true, Score: 90}, nil
}
func (f *fakeHandler) Verify(ctx context.Context) (*VerifyResponse, error) {
	return &VerifyResponse{Score: 88, Passed: true}, nil
}
func (f *fakeHandler) GetPlan(ctx context.Context) (*PlanResponse, error) {
	return &PlanResponse{PlanJSON: "{}", Nodes: 3}, nil
}

// fakeExtendedHandler implements Handler + ExtendedHandler for Continue testing.
type fakeExtendedHandler struct{ fakeHandler }

func (f *fakeExtendedHandler) GetAgentCards(ctx context.Context) []a2a.AgentCard {
	return nil
}

func (f *fakeExtendedHandler) GetPlanDetail(ctx context.Context) (*PlanDetail, error) {
	return &PlanDetail{ID: "test"}, nil
}

func (f *fakeExtendedHandler) ContinuePlan(ctx context.Context, planID string) (bool, string, error) {
	return true, "resumed", nil
}

// fakeEmptyHandler returns an empty plan (0 nodes).
type fakeEmptyHandler struct{ fakeHandler }

func (f *fakeEmptyHandler) GetPlan(ctx context.Context) (*PlanResponse, error) {
	return &PlanResponse{PlanJSON: "{}", Nodes: 0}, nil
}

// fakeErrorHandler returns an error from GetPlan.
type fakeErrorHandler struct{ fakeHandler }

func (f *fakeErrorHandler) GetPlan(ctx context.Context) (*PlanResponse, error) {
	return nil, fmt.Errorf("plan not found")
}

// startTestGRPC starts a gRPC server on a free port and returns the address + cleanup.
func startTestGRPC(t *testing.T, h Handler) (string, func()) {
	t.Helper()
	lis, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen: %v", err)
	}
	grpcServer := grpc.NewServer()
	pb.RegisterCoordinatorServer(grpcServer, &coordinatorAdapter{h: h})
	go func() { _ = grpcServer.Serve(lis) }()
	addr := lis.Addr().String()
	// Poll until the server is ready
	for i := 0; i < 50; i++ {
		conn, err := net.DialTimeout("tcp", addr, 100*time.Millisecond)
		if err == nil {
			conn.Close()
			break
		}
		time.Sleep(50 * time.Millisecond)
	}
	return addr, grpcServer.Stop
}

func TestGRPC_RunPipeline_EndToEnd(t *testing.T) {
	addr, stop := startTestGRPC(t, &fakeHandler{})
	defer stop()

	conn, err := grpc.Dial(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		t.Fatalf("failed to connect to gRPC server: %v", err)
	}
	defer conn.Close()

	client := pb.NewCoordinatorClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	resp, err := client.RunPipeline(ctx, &pb.RunPipelineRequest{Requirement: "build app", Backend: "codex"})
	if err != nil {
		t.Fatalf("RunPipeline RPC failed: %v", err)
	}
	if resp.PlanId != "test-plan" {
		t.Errorf("expected test-plan, got %s", resp.PlanId)
	}
	if resp.Score != 95 {
		t.Errorf("expected score 95, got %d", resp.Score)
	}
	if !resp.Passed {
		t.Error("expected passed=true")
	}
	if len(resp.Artifacts) != 1 || resp.Artifacts[0] != "a.go" {
		t.Errorf("unexpected artifacts: %v", resp.Artifacts)
	}
}

func TestGRPC_GetPlan_EndToEnd(t *testing.T) {
	addr, stop := startTestGRPC(t, &fakeHandler{})
	defer stop()

	conn, err := grpc.Dial(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		t.Fatalf("failed to connect: %v", err)
	}
	defer conn.Close()

	client := pb.NewCoordinatorClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	resp, err := client.GetPlan(ctx, &pb.GetPlanRequest{})
	if err != nil {
		t.Fatalf("GetPlan RPC failed: %v", err)
	}
	if resp.NodeCount != 3 {
		t.Errorf("expected 3 nodes, got %d", resp.NodeCount)
	}
}

func TestGRPC_Verify_EndToEnd(t *testing.T) {
	addr, stop := startTestGRPC(t, &fakeHandler{})
	defer stop()

	conn, err := grpc.Dial(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		t.Fatalf("failed to connect: %v", err)
	}
	defer conn.Close()

	client := pb.NewCoordinatorClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	resp, err := client.Verify(ctx, &pb.VerifyRequest{Runtime: false})
	if err != nil {
		t.Fatalf("Verify RPC failed: %v", err)
	}
	if resp.Score != 88 {
		t.Errorf("expected score 88, got %d", resp.Score)
	}
	if !resp.Passed {
		t.Error("expected passed=true")
	}
}

func TestGRPC_QuickEdit_EndToEnd(t *testing.T) {
	addr, stop := startTestGRPC(t, &fakeHandler{})
	defer stop()

	conn, err := grpc.Dial(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		t.Fatalf("failed to connect: %v", err)
	}
	defer conn.Close()

	client := pb.NewCoordinatorClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	resp, err := client.QuickEdit(ctx, &pb.QuickEditRequest{Description: "fix typo", Backend: "codex"})
	if err != nil {
		t.Fatalf("QuickEdit RPC failed: %v", err)
	}
	if !resp.Passed {
		t.Error("expected passed=true")
	}
	if resp.Score != 90 {
		t.Errorf("expected score 90, got %d", resp.Score)
	}
	_ = fmt.Sprintf("") // silence unused import if fmt not needed
}

func TestGRPC_Continue_FallbackPlanExists(t *testing.T) {
	addr, stop := startTestGRPC(t, &fakeHandler{})
	defer stop()

	conn, err := grpc.Dial(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		t.Fatalf("failed to connect: %v", err)
	}
	defer conn.Close()

	client := pb.NewCoordinatorClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	resp, err := client.Continue(ctx, &pb.ContinueRequest{PlanId: "p1"})
	if err != nil {
		t.Fatalf("Continue RPC failed: %v", err)
	}
	if resp.Resumed {
		t.Error("expected resumed=false without ExtendedHandler")
	}
	if resp.Status != "plan exists but resume requires extended handler" {
		t.Errorf("unexpected status: %s", resp.Status)
	}
}

func TestGRPC_Continue_FallbackNoPlan(t *testing.T) {
	addr, stop := startTestGRPC(t, &fakeEmptyHandler{})
	defer stop()

	conn, err := grpc.Dial(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		t.Fatalf("failed to connect: %v", err)
	}
	defer conn.Close()

	client := pb.NewCoordinatorClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	resp, err := client.Continue(ctx, &pb.ContinueRequest{PlanId: "p1"})
	if err != nil {
		t.Fatalf("Continue RPC failed: %v", err)
	}
	if resp.Resumed {
		t.Error("expected resumed=false")
	}
	if resp.Status != "no plan available to continue" {
		t.Errorf("unexpected status: %s", resp.Status)
	}
}

func TestGRPC_Continue_FallbackGetPlanError(t *testing.T) {
	addr, stop := startTestGRPC(t, &fakeErrorHandler{})
	defer stop()

	conn, err := grpc.Dial(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		t.Fatalf("failed to connect: %v", err)
	}
	defer conn.Close()

	client := pb.NewCoordinatorClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err = client.Continue(ctx, &pb.ContinueRequest{PlanId: "p1"})
	if err == nil {
		t.Fatal("expected error from Continue RPC")
	}
	st, ok := status.FromError(err)
	if !ok || st.Code() != codes.Internal {
		t.Errorf("expected Internal error, got %v", err)
	}
}

func TestGRPC_Continue_WithExtendedHandler(t *testing.T) {
	h := &fakeExtendedHandler{}
	grpcServer := grpc.NewServer()
	adapter := &coordinatorAdapter{h: h, ext: h}
	pb.RegisterCoordinatorServer(grpcServer, adapter)

	lis, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen: %v", err)
	}
	go func() { _ = grpcServer.Serve(lis) }()
	defer grpcServer.Stop()

	conn, err := grpc.Dial(lis.Addr().String(), grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		t.Fatalf("failed to connect: %v", err)
	}
	defer conn.Close()

	client := pb.NewCoordinatorClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	resp, err := client.Continue(ctx, &pb.ContinueRequest{PlanId: "p1"})
	if err != nil {
		t.Fatalf("Continue RPC failed: %v", err)
	}
	if !resp.Resumed {
		t.Error("expected resumed=true with ExtendedHandler")
	}
	if resp.Status != "resumed" {
		t.Errorf("expected status 'resumed', got %s", resp.Status)
	}
}
