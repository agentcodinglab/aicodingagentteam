package memory

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

func TestCaptureAndRecallFacts(t *testing.T) {
	tmpDir := t.TempDir()
	s := New(tmpDir)

	err := s.Capture(context.Background(), Fact{
		Key:    "language",
		Value:  "go",
		Source: "go.mod",
	})
	if err != nil {
		t.Fatalf("Capture error: %v", err)
	}

	facts, err := s.RecallFacts(context.Background())
	if err != nil {
		t.Fatalf("RecallFacts error: %v", err)
	}
	if len(facts) != 1 {
		t.Fatalf("expected 1 fact, got %d", len(facts))
	}
	if facts[0].Key != "language" || facts[0].Value != "go" {
		t.Errorf("unexpected fact: %+v", facts[0])
	}
}

func TestTombstoneFact(t *testing.T) {
	tmpDir := t.TempDir()
	s := New(tmpDir)

	_ = s.Capture(context.Background(), Fact{Key: "db", Value: "postgres", Source: "config"})
	_ = s.Capture(context.Background(), Fact{Key: "cache", Value: "redis", Source: "config"})

	err := s.TombstoneFact(context.Background(), "db")
	if err != nil {
		t.Fatalf("TombstoneFact error: %v", err)
	}

	facts, _ := s.RecallFacts(context.Background())
	if len(facts) != 1 {
		t.Fatalf("expected 1 non-tombstoned fact, got %d", len(facts))
	}
	if facts[0].Key != "cache" {
		t.Errorf("expected cache fact, got %s", facts[0].Key)
	}
}

func TestRecallRecipesByStack(t *testing.T) {
	tmpDir := t.TempDir()
	s := New(tmpDir)

	_ = s.CaptureRecipe(context.Background(), Recipe{
		ID:       "r1",
		Stack:    []string{"go", "gin"},
		Solution: "REST API with gin framework",
	})
	_ = s.CaptureRecipe(context.Background(), Recipe{
		ID:       "r2",
		Stack:    []string{"python", "flask"},
		Solution: "REST API with flask",
	})
	_ = s.CaptureRecipe(context.Background(), Recipe{
		ID:       "r3",
		Stack:    []string{"go", "echo"},
		Solution: "REST API with echo framework",
	})

	recipes, err := s.Recall(context.Background(), []string{"go"})
	if err != nil {
		t.Fatalf("Recall error: %v", err)
	}
	if len(recipes) != 2 {
		t.Errorf("expected 2 recipes matching 'go', got %d", len(recipes))
	}

	recipes, _ = s.Recall(context.Background(), []string{"python"})
	if len(recipes) != 1 {
		t.Errorf("expected 1 recipe matching 'python', got %d", len(recipes))
	}
}

func TestProjectIsolation(t *testing.T) {
	dir1 := t.TempDir()
	dir2 := t.TempDir()
	s1 := New(dir1)
	s2 := New(dir2)

	_ = s1.Capture(context.Background(), Fact{Key: "lang", Value: "go", Source: "test"})

	facts1, _ := s1.RecallFacts(context.Background())
	facts2, _ := s2.RecallFacts(context.Background())

	if len(facts1) != 1 {
		t.Errorf("project1 should have 1 fact, got %d", len(facts1))
	}
	if len(facts2) != 0 {
		t.Errorf("project2 should have 0 facts, got %d", len(facts2))
	}
}

func TestCaptureOff(t *testing.T) {
	tmpDir := t.TempDir()
	s := New(tmpDir)
	s.SetCaptureOn(false)

	_ = s.Capture(context.Background(), Fact{Key: "lang", Value: "go", Source: "test"})
	facts, _ := s.RecallFacts(context.Background())
	if len(facts) != 0 {
		t.Error("capture should be off, no facts stored")
	}
}

func TestRecallOff(t *testing.T) {
	tmpDir := t.TempDir()
	s := New(tmpDir)
	_ = s.CaptureRecipe(context.Background(), Recipe{ID: "r1", Stack: []string{"go"}, Solution: "test"})

	s.SetRecallOn(false)
	recipes, _ := s.Recall(context.Background(), []string{"go"})
	if recipes != nil {
		t.Error("recall should be off, expected nil")
	}
}

func TestCaptureOnRecallOnDefaults(t *testing.T) {
	s := New(t.TempDir())
	if !s.CaptureOn() {
		t.Error("capture should default to on")
	}
	if !s.RecallOn() {
		t.Error("recall should default to on")
	}
}

func TestCapturePitfall(t *testing.T) {
	tmpDir := t.TempDir()
	s := New(tmpDir)

	_ = s.CapturePitfall(context.Background(), Pitfall{
		ID:     "p1",
		Detail: "build failed: undefined symbol",
		Count:  1,
	})

	pitfalls, err := s.RecallPitfalls(context.Background())
	if err != nil {
		t.Fatalf("RecallPitfalls error: %v", err)
	}
	if len(pitfalls) != 1 {
		t.Errorf("expected 1 pitfall, got %d", len(pitfalls))
	}
}

func TestLessonMustBeVerified(t *testing.T) {
	s := New(t.TempDir())

	err := s.CaptureLesson(context.Background(), Lesson{ID: "l1", Rule: "always test", Verified: false})
	if err == nil {
		t.Error("unverified lesson should be rejected")
	}

	err = s.CaptureLesson(context.Background(), Lesson{ID: "l1", Rule: "always test", Verified: true})
	if err != nil {
		t.Errorf("verified lesson should be accepted: %v", err)
	}
}

func TestFactsFileExists(t *testing.T) {
	tmpDir := t.TempDir()
	s := New(tmpDir)
	_ = s.Capture(context.Background(), Fact{Key: "test", Value: "val", Source: "test"})

	if _, err := os.Stat(filepath.Join(tmpDir, "facts.jsonl")); err != nil {
		t.Error("facts.jsonl should be created after Capture")
	}
}

func TestRecipesFileExists(t *testing.T) {
	tmpDir := t.TempDir()
	s := New(tmpDir)
	_ = s.CaptureRecipe(context.Background(), Recipe{ID: "r1", Stack: []string{"go"}, Solution: "test"})

	if _, err := os.Stat(filepath.Join(tmpDir, "recipes.jsonl")); err != nil {
		t.Error("recipes.jsonl should be created after CaptureRecipe")
	}
}

func TestPitfallsFileExists(t *testing.T) {
	tmpDir := t.TempDir()
	s := New(tmpDir)
	_ = s.CapturePitfall(context.Background(), Pitfall{ID: "p1", Detail: "error", Count: 1})

	if _, err := os.Stat(filepath.Join(tmpDir, "dev-errors.jsonl")); err != nil {
		t.Error("dev-errors.jsonl should be created after CapturePitfall")
	}
}

func TestMemoryDir(t *testing.T) {
	dir := t.TempDir()
	s := New(dir)
	if s.Dir() != dir {
		t.Errorf("expected dir %s, got %s", dir, s.Dir())
	}
}
