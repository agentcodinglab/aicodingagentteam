package agent

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/agentcodinglab/aicodingagentteam/internal/a2a"
	"github.com/agentcodinglab/aicodingagentteam/internal/types"
)

func TestReviewerAgent_AcceptsWhenArtifactsExist(t *testing.T) {
	dir := t.TempDir()
	artFile := filepath.Join(dir, "prd.md")
	_ = os.WriteFile(artFile, []byte("content"), 0o644)

	qa := NewQAAgent()
	task := a2a.Task{
		TaskID: "n8",
		Role:   types.RoleQA,
		Payload: a2a.TaskPayload{
			Artifacts:   []string{artFile},
			Instruction: "review",
		},
	}
	result, err := qa.Execute(context.Background(), task)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Verdict.Decision != types.DecisionAccept {
		t.Errorf("expected accept when artifacts exist, got %s", result.Verdict.Decision)
	}
}

func TestReviewerAgent_BlocksWhenArtifactsMissing(t *testing.T) {
	qa := NewQAAgent()
	task := a2a.Task{
		TaskID: "n8",
		Role:   types.RoleQA,
		Payload: a2a.TaskPayload{
			Artifacts:   []string{"/nonexistent/file.md"},
			Instruction: "review",
		},
	}
	result, _ := qa.Execute(context.Background(), task)
	if result.Verdict.Decision != types.DecisionBlocking {
		t.Errorf("expected blocking for missing artifacts, got %s", result.Verdict.Decision)
	}
	if result.Verdict.Severity != "critical" {
		t.Error("blocking verdict should have critical severity")
	}
}

func TestReviewerAgent_PMAcceptsEvenWithMissingArtifacts(t *testing.T) {
	pm := NewPMAgent()
	task := a2a.Task{
		TaskID: "n3",
		Role:   types.RolePM,
		Payload: a2a.TaskPayload{
			Artifacts:   []string{"/nonexistent/input.md"},
			Instruction: "generate PRD",
		},
	}
	result, _ := pm.Execute(context.Background(), task)
	if result.Verdict.Decision != types.DecisionAccept {
		t.Error("PM should accept even with missing input artifacts (it generates them)")
	}
}

func TestRegisterAllReviewers(t *testing.T) {
	bus := a2a.NewBus()
	RegisterAllReviewers(bus)
	if bus.Count() != 5 {
		t.Errorf("expected 5 registered agents, got %d", bus.Count())
	}
	for _, role := range []types.Role{types.RolePM, types.RoleArchitect, types.RoleQA, types.RoleSecurity, types.RoleDevOps} {
		if !bus.IsRegistered(role) {
			t.Errorf("role %s should be registered", role)
		}
	}
}

func TestAgentCardFields(t *testing.T) {
	qa := NewQAAgent()
	card := qa.Card()
	if card.ID != "agent-qa" {
		t.Errorf("expected agent-qa, got %s", card.ID)
	}
	if card.Role != types.RoleQA {
		t.Error("role mismatch")
	}
	if len(card.Capabilities) == 0 {
		t.Error("should have capabilities")
	}
}
