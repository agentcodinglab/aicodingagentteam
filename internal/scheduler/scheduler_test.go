package scheduler

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/agentcodinglab/aicodingagentteam/internal/a2a"
	"github.com/agentcodinglab/aicodingagentteam/internal/agent"
	"github.com/agentcodinglab/aicodingagentteam/internal/audit"
	"github.com/agentcodinglab/aicodingagentteam/internal/types"
)

func TestExecute_WriterAcquiresLock(t *testing.T) {
	dir := t.TempDir()
	s := New(dir)
	plan := &types.Plan{
		ID: "test",
		Nodes: []types.TaskNode{
			{ID: "n6", Phase: types.PhaseFrontend, Role: types.RoleFrontend, Writer: true, ArtifactsOut: []string{"src/"}},
		},
	}
	res, err := s.Execute(context.Background(), plan)
	if err != nil {
		t.Fatalf("execute failed: %v", err)
	}
	if res.Parked {
		t.Error("should not park on successful write")
	}
	// lock should be released after execution
	lockPath := filepath.Join(dir, ".aicodingagentteam", "write.lock")
	if _, err := os.Stat(lockPath); err == nil {
		t.Error("write lock should be released after execution")
	}
}

func TestExecute_SecondWriterBlockedWhenLockHeld(t *testing.T) {
	dir := t.TempDir()
	s := New(dir)

	// Pre-create a lock file simulating another writer
	lockDir := filepath.Join(dir, ".aicodingagentteam")
	_ = os.MkdirAll(lockDir, 0o755)
	_ = os.WriteFile(filepath.Join(lockDir, "write.lock"), []byte("other|backend|2026-09-02T10:00:00Z"), 0o644)

	plan := &types.Plan{
		ID: "test",
		Nodes: []types.TaskNode{
			{ID: "n6", Phase: types.PhaseFrontend, Role: types.RoleFrontend, Writer: true},
		},
	}
	res, _ := s.Execute(context.Background(), plan)
	if !res.Parked {
		t.Error("should park when write lock is held by another writer")
	}
}

func TestExecute_NonWriterDoesNotAcquireLock(t *testing.T) {
	dir := t.TempDir()
	s := New(dir)
	plan := &types.Plan{
		ID: "test",
		Nodes: []types.TaskNode{
			{ID: "n3", Phase: types.PhaseDocs, Role: types.RolePM, ArtifactsOut: []string{"output/prd.md"}},
		},
	}
	res, err := s.Execute(context.Background(), plan)
	if err != nil {
		t.Fatalf("execute failed: %v", err)
	}
	if res.Parked {
		t.Error("reviewer role should not park")
	}
	// no lock file should be created for non-writers
	lockPath := filepath.Join(dir, ".aicodingagentteam", "write.lock")
	if _, err := os.Stat(lockPath); err == nil {
		t.Error("non-writer should not create write lock")
	}
}

func TestExecute_NilPlanReturnsError(t *testing.T) {
	s := New(t.TempDir())
	_, err := s.Execute(context.Background(), nil)
	if err == nil {
		t.Error("should return error for nil plan")
	}
}

func TestExecute_MultipleWritersSerialized(t *testing.T) {
	dir := t.TempDir()
	s := New(dir)
	plan := &types.Plan{
		ID: "test",
		Nodes: []types.TaskNode{
			{ID: "n6", Phase: types.PhaseFrontend, Role: types.RoleFrontend, Writer: true, ArtifactsOut: []string{"src/frontend/"}},
			{ID: "n7", Phase: types.PhaseBackend, Role: types.RoleBackend, Writer: true, ArtifactsOut: []string{"src/backend/"}},
		},
	}
	res, err := s.Execute(context.Background(), plan)
	if err != nil {
		t.Fatalf("execute failed: %v", err)
	}
	if res.Parked {
		t.Error("serialized writers should not park")
	}
	if len(res.Artifacts) != 2 {
		t.Errorf("expected 2 artifacts, got %d", len(res.Artifacts))
	}
	// Both locks should be released
	lockPath := filepath.Join(dir, ".aicodingagentteam", "write.lock")
	if _, err := os.Stat(lockPath); err == nil {
		t.Error("all locks should be released")
	}
}

func TestNewWithBus_ConnectsBus(t *testing.T) {
	bus := a2a.NewBus()
	s := NewWithBus(t.TempDir(), bus)
	if s.bus == nil {
		t.Error("expected bus to be set")
	}
}

func TestNewWithAudit_SetsAudit(t *testing.T) {
	al := audit.New(t.TempDir())
	s := NewWithAudit(t.TempDir(), al)
	if s.audit == nil {
		t.Error("expected audit logger to be set")
	}
}

func TestNewFull_SetsBusAndAudit(t *testing.T) {
	bus := a2a.NewBus()
	al := audit.New(t.TempDir())
	s := NewFull(t.TempDir(), bus, al)
	if s.bus == nil {
		t.Error("expected bus to be set")
	}
	if s.audit == nil {
		t.Error("expected audit logger to be set")
	}
}

func TestExecute_WithBus_DelegatesReviewerTasks(t *testing.T) {
	dir := t.TempDir()
	bus := a2a.NewBus()
	// Register a reviewer that always accepts
	bus.Register(agent.NewReviewer(
		a2a.AgentCard{ID: "pm", Name: "PM", Role: types.RolePM, MaxConcurrent: 1, TimeoutDefault: 30},
		func(ctx context.Context, task a2a.Task) (types.Verdict, error) {
			return types.Verdict{
				TaskID:    task.TaskID,
				Role:      types.RolePM,
				Decision:  types.DecisionAccept,
				Artifacts: task.Payload.Artifacts,
			}, nil
		},
	))
	s := NewWithBus(dir, bus)
	plan := &types.Plan{
		ID: "p1",
		Nodes: []types.TaskNode{
			{ID: "n3", Phase: types.PhaseDocs, Role: types.RolePM, ArtifactsOut: []string{"output/prd.md"}},
		},
	}
	res, err := s.Execute(context.Background(), plan)
	if err != nil {
		t.Fatalf("execute failed: %v", err)
	}
	if len(res.Verdicts) != 1 {
		t.Fatalf("expected 1 verdict, got %d", len(res.Verdicts))
	}
	if res.Verdicts[0].Decision != types.DecisionAccept {
		t.Errorf("expected accept verdict, got %s", res.Verdicts[0].Decision)
	}
}

func TestExecute_WithBus_ReviewerBlockingParks(t *testing.T) {
	dir := t.TempDir()
	bus := a2a.NewBus()
	bus.Register(agent.NewReviewer(
		a2a.AgentCard{ID: "qa", Name: "QA", Role: types.RoleQA, MaxConcurrent: 1, TimeoutDefault: 30},
		func(ctx context.Context, task a2a.Task) (types.Verdict, error) {
			return types.Verdict{
				TaskID:   task.TaskID,
				Role:     types.RoleQA,
				Decision: types.DecisionBlocking,
			}, nil
		},
	))
	s := NewWithBus(dir, bus)
	plan := &types.Plan{
		ID: "p1",
		Nodes: []types.TaskNode{
			{ID: "n8", Phase: types.PhaseQuality, Role: types.RoleQA},
		},
	}
	res, _ := s.Execute(context.Background(), plan)
	if !res.Parked {
		t.Error("expected parked when reviewer returns blocking")
	}
}

func TestExecute_CoordinatorNodeProducesArtifacts(t *testing.T) {
	dir := t.TempDir()
	s := New(dir)
	plan := &types.Plan{
		ID: "p1",
		Nodes: []types.TaskNode{
			{ID: "n1", Phase: types.PhaseClarify, Role: types.RoleCoordinator, ArtifactsOut: []string{"output/clarify.md"}},
		},
	}
	res, err := s.Execute(context.Background(), plan)
	if err != nil {
		t.Fatalf("execute failed: %v", err)
	}
	if len(res.Artifacts) != 1 || res.Artifacts[0] != "output/clarify.md" {
		t.Errorf("expected coordinator artifact, got %v", res.Artifacts)
	}
}

func TestExecute_AuditLogsWritten(t *testing.T) {
	dir := t.TempDir()
	al := audit.New(filepath.Join(dir, "audit"))
	s := NewWithAudit(dir, al)
	plan := &types.Plan{
		ID: "p1",
		Nodes: []types.TaskNode{
			{ID: "n1", Phase: types.PhaseClarify, Role: types.RoleCoordinator},
		},
	}
	_, err := s.Execute(context.Background(), plan)
	if err != nil {
		t.Fatalf("execute failed: %v", err)
	}
	// audit log should have been created
	if _, err := os.Stat(filepath.Join(dir, "audit")); os.IsNotExist(err) {
		t.Error("expected audit directory to be created")
	}
}

func TestGetVerdicts_EmptyByDefault(t *testing.T) {
	s := New(t.TempDir())
	if v := s.GetVerdicts("nonexistent"); v != nil {
		t.Errorf("expected nil for unknown node, got %v", v)
	}
}

func TestNew_WorkspaceDefaultsToCwd(t *testing.T) {
	s := New("")
	if s.workspace == "" {
		t.Error("expected workspace to default to cwd")
	}
}
