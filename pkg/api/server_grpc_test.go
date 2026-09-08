package api

import (
	"context"
	"fmt"
	"net"
	"testing"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	pb "github.com/agentcodinglab/aicodingagentteam/pkg/api/gen"
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
