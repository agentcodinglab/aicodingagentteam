// Package memory stores project-local facts, recipes, pitfalls, and lessons.
// Memory never auto-shares across projects.
package memory

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

// Store manages the project memory directory.
type Store struct {
	mu           sync.RWMutex
	dir          string
	captureOn    bool
	recallOn     bool
	factsFile    string
	recipesFile  string
	pitfallsFile string
	lessonsDir   string
}

// Fact is a project environment fact with provenance.
type Fact struct {
	Key        string    `json:"key"`
	Value      string    `json:"value"`
	Source     string    `json:"source"`
	Tombstoned bool      `json:"tombstoned,omitempty"`
	TS         time.Time `json:"ts,omitempty"`
}

// Recipe is a historical delivery solution.
type Recipe struct {
	ID       string    `json:"id"`
	Stack    []string  `json:"stack"`
	Solution string    `json:"solution"`
	TS       time.Time `json:"ts,omitempty"`
}

// Pitfall is a recorded failure event.
type Pitfall struct {
	ID       string    `json:"id"`
	Detail   string    `json:"detail"`
	Count    int       `json:"count"`
	Verified bool      `json:"verified"`
	TS       time.Time `json:"ts,omitempty"`
}

// Lesson is a validated rule from a pitfall.
type Lesson struct {
	ID       string `json:"id"`
	Rule     string `json:"rule"`
	Verified bool   `json:"verified"`
}

// New creates a memory Store rooted at the given directory.
func New(dir string) *Store {
	return &Store{
		dir:          dir,
		captureOn:    true,
		recallOn:     true,
		factsFile:    filepath.Join(dir, "facts.jsonl"),
		recipesFile:  filepath.Join(dir, "recipes.jsonl"),
		pitfallsFile: filepath.Join(dir, "dev-errors.jsonl"),
		lessonsDir:   filepath.Join(dir, "learned-skills"),
	}
}

// SetCaptureOn enables or disables fact capture.
func (s *Store) SetCaptureOn(on bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.captureOn = on
}

// SetRecallOn enables or disables recipe recall.
func (s *Store) SetRecallOn(on bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.recallOn = on
}

// CaptureOn reports whether fact capture is enabled.
func (s *Store) CaptureOn() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.captureOn
}

// RecallOn reports whether recipe recall is enabled.
func (s *Store) RecallOn() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.recallOn
}

// Capture stores a new fact.
func (s *Store) Capture(ctx context.Context, f Fact) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if !s.captureOn {
		return nil // silently skip when capture is off
	}
	f.TS = time.Now()
	return s.appendJSONL(s.factsFile, f)
}

// RecallFacts retrieves all non-tombstoned facts.
func (s *Store) RecallFacts(ctx context.Context) ([]Fact, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	var facts []Fact
	err := s.readJSONL(s.factsFile, func(data []byte) {
		var f Fact
		if err := json.Unmarshal(data, &f); err == nil && !f.Tombstoned {
			facts = append(facts, f)
		}
	})
	return facts, err
}

// TombstoneFact marks a fact as expired by key.
func (s *Store) TombstoneFact(ctx context.Context, key string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	facts, err := s.readAllFacts()
	if err != nil {
		return err
	}
	// Rewrite all facts with tombstone on matching key
	var out []Fact
	for _, f := range facts {
		if f.Key == key {
			f.Tombstoned = true
		}
		out = append(out, f)
	}
	return s.writeJSONL(s.factsFile, out)
}

// Recall retrieves matching recipes by stack.
func (s *Store) Recall(ctx context.Context, stack []string) ([]Recipe, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if !s.recallOn {
		return nil, nil
	}
	var all []Recipe
	err := s.readJSONL(s.recipesFile, func(data []byte) {
		var r Recipe
		if err := json.Unmarshal(data, &r); err == nil {
			all = append(all, r)
		}
	})
	if err != nil {
		return nil, err
	}

	// Filter by stack match
	stackSet := make(map[string]bool)
	for _, s := range stack {
		stackSet[strings.ToLower(s)] = true
	}
	var matched []Recipe
	for _, r := range all {
		if stackMatches(r.Stack, stackSet) {
			matched = append(matched, r)
		}
	}
	return matched, nil
}

// CaptureRecipe stores a new recipe.
func (s *Store) CaptureRecipe(ctx context.Context, r Recipe) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	r.TS = time.Now()
	return s.appendJSONL(s.recipesFile, r)
}

// CapturePitfall records a failure event.
func (s *Store) CapturePitfall(ctx context.Context, p Pitfall) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	p.TS = time.Now()
	return s.appendJSONL(s.pitfallsFile, p)
}

// RecallPitfalls retrieves all pitfall events.
func (s *Store) RecallPitfalls(ctx context.Context) ([]Pitfall, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	var pitfalls []Pitfall
	err := s.readJSONL(s.pitfallsFile, func(data []byte) {
		var p Pitfall
		if err := json.Unmarshal(data, &p); err == nil {
			pitfalls = append(pitfalls, p)
		}
	})
	return pitfalls, err
}

// CaptureLesson stores a lesson (only if verified).
func (s *Store) CaptureLesson(ctx context.Context, l Lesson) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if !l.Verified {
		return fmt.Errorf("lesson must be verified before capture")
	}
	return s.appendJSONL(filepath.Join(s.lessonsDir, "lessons.jsonl"), l)
}

// RecallLessons retrieves all verified lessons.
func (s *Store) RecallLessons(ctx context.Context) ([]Lesson, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	var lessons []Lesson
	err := s.readJSONL(filepath.Join(s.lessonsDir, "lessons.jsonl"), func(data []byte) {
		var l Lesson
		if err := json.Unmarshal(data, &l); err == nil && l.Verified {
			lessons = append(lessons, l)
		}
	})
	return lessons, err
}

// Dir returns the memory directory (for project isolation checks).
func (s *Store) Dir() string {
	return s.dir
}

// --- helpers ---

func (s *Store) appendJSONL(path string, v interface{}) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	f, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		return err
	}
	defer func() { _ = f.Close() }()
	data, err := json.Marshal(v)
	if err != nil {
		return err
	}
	_, err = f.Write(append(data, '\n'))
	return err
}

func (s *Store) readJSONL(path string, fn func(data []byte)) error {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	for _, line := range strings.Split(string(data), "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		fn([]byte(line))
	}
	return nil
}

func (s *Store) writeJSONL(path string, items []Fact) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	var buf strings.Builder
	for _, item := range items {
		data, err := json.Marshal(item)
		if err != nil {
			continue
		}
		buf.Write(data)
		buf.WriteByte('\n')
	}
	return os.WriteFile(path, []byte(buf.String()), 0o644)
}

func (s *Store) readAllFacts() ([]Fact, error) {
	var facts []Fact
	err := s.readJSONL(s.factsFile, func(data []byte) {
		var f Fact
		if err := json.Unmarshal(data, &f); err == nil {
			facts = append(facts, f)
		}
	})
	return facts, err
}

func stackMatches(recipeStack []string, querySet map[string]bool) bool {
	for _, s := range recipeStack {
		if querySet[strings.ToLower(s)] {
			return true
		}
	}
	return false
}
