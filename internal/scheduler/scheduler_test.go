package scheduler

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/yourorg/aicodingagentteam/internal/types"
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
