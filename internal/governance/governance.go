// Package governance implements the fail-open governance rule engine.
package governance

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/yourorg/aicodingagentteam/internal/audit"
)

// Rule is a single governance check clause.
type Rule struct {
	ID       string
	Severity string // "advisory" or "blocking"
	Enabled  bool
	Group    string // "ui" / "security" / "contract" / "engineering"
}

// Violation is a governance rule violation.
type Violation struct {
	RuleID   string `json:"rule_id"`
	Severity string `json:"severity"`
	Path     string `json:"path"`
	Detail   string `json:"detail"`
}

// Engine evaluates governance rules. Fail-open: internal errors never block.
type Engine struct {
	mu     sync.RWMutex
	rules  []Rule
	config RuleConfig
	audit  *audit.Logger
	closed bool
}

// RuleConfig holds governance configuration loaded from rules.toml.
type RuleConfig struct {
	Disabled   DisabledCfg   `json:"disabled"`
	Exclusions ExclusionsCfg `json:"exclusions"`
}

// DisabledCfg lists rule IDs that are turned off.
type DisabledCfg struct {
	Clauses []string `json:"clauses"`
}

// ExclusionsCfg lists path patterns to skip.
type ExclusionsCfg struct {
	Paths []string `json:"paths"`
}

// New creates a governance Engine with default rules.
func New() *Engine {
	return &Engine{
		rules:  defaultRules(),
		config: RuleConfig{},
	}
}

// NewWithConfig creates a governance Engine with config loaded from a rules.toml file.
func NewWithConfig(configPath string, al *audit.Logger) *Engine {
	e := &Engine{
		rules: defaultRules(),
		audit: al,
	}
	if configPath != "" {
		_ = e.LoadConfig(configPath)
	}
	return e
}

// LoadConfig reads rules.toml (JSON format for simplicity) and applies disabled clauses and exclusions.
func (e *Engine) LoadConfig(path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}
	var cfg RuleConfig
	if err := json.Unmarshal(data, &cfg); err != nil {
		return fmt.Errorf("parse config: %w", err)
	}
	e.mu.Lock()
	defer e.mu.Unlock()
	e.config = cfg
	// Apply disabled clauses
	disabledSet := make(map[string]bool)
	for _, c := range cfg.Disabled.Clauses {
		disabledSet[c] = true
	}
	for i, r := range e.rules {
		if disabledSet[r.ID] {
			e.rules[i].Enabled = false
		}
	}
	return nil
}

// IsExcluded reports whether the path matches any exclusion pattern.
func (e *Engine) IsExcluded(path string) bool {
	e.mu.RLock()
	defer e.mu.RUnlock()
	for _, pattern := range e.config.Exclusions.Paths {
		if matched, _ := filepath.Match(pattern, path); matched {
			return true
		}
	}
	return false
}

// Check evaluates all enabled rules against the given file content.
// Fail-open: any internal panic/error returns empty violations, never blocking development.
func (e *Engine) Check(ctx context.Context, path, content string) []Violation {
	if e.closed {
		return nil
	}
	if e.IsExcluded(path) {
		return nil
	}

	var out []Violation
	e.mu.RLock()
	rules := make([]Rule, len(e.rules))
	copy(rules, e.rules)
	e.mu.RUnlock()

	for _, r := range rules {
		if !r.Enabled {
			continue
		}
		violations := checkRule(r, path, content)
		out = append(out, violations...)
	}

	// Audit violations
	if e.audit != nil && len(out) > 0 {
		for _, v := range out {
			_ = e.audit.Log(audit.Entry{
				TS:     time.Now(),
				Type:   "governance",
				Agent:  "governance",
				Task:   v.RuleID,
				Result: v.Severity,
				Detail: fmt.Sprintf("%s: %s", v.Path, v.Detail),
			})
		}
	}

	return out
}

// Close marks the engine as closed (for CI fail-close mode).
func (e *Engine) Close() {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.closed = true
}

// HasBlocking reports whether any violation is blocking severity.
func HasBlocking(violations []Violation) bool {
	for _, v := range violations {
		if v.Severity == "blocking" {
			return true
		}
	}
	return false
}

// Rules returns the current rule list.
func (e *Engine) Rules() []Rule {
	e.mu.RLock()
	defer e.mu.RUnlock()
	out := make([]Rule, len(e.rules))
	copy(out, e.rules)
	return out
}

// checkRule dispatches to the appropriate rule checker.
func checkRule(r Rule, path, content string) []Violation {
	switch r.ID {
	case "ui-emoji-icon":
		return checkEmoji(path, content)
	case "ui-hardcoded-color":
		return checkHardcodedColor(path, content)
	case "sec-secret-leak":
		return checkSecretLeak(path, content)
	case "sec-sql-injection":
		return checkSQLInjection(path, content)
	case "api-contract-mismatch":
		return checkAPIContract(path, content)
	case "eng-todo-placeholder":
		return checkTodoPlaceholder(path, content)
	case "eng-fake-data":
		return checkFakeData(path, content)
	default:
		return nil
	}
}

func defaultRules() []Rule {
	return []Rule{
		{ID: "ui-emoji-icon", Severity: "advisory", Enabled: true, Group: "ui"},
		{ID: "ui-hardcoded-color", Severity: "advisory", Enabled: true, Group: "ui"},
		{ID: "sec-secret-leak", Severity: "blocking", Enabled: true, Group: "security"},
		{ID: "sec-sql-injection", Severity: "blocking", Enabled: true, Group: "security"},
		{ID: "api-contract-mismatch", Severity: "blocking", Enabled: true, Group: "contract"},
		{ID: "eng-todo-placeholder", Severity: "advisory", Enabled: true, Group: "engineering"},
		{ID: "eng-fake-data", Severity: "advisory", Enabled: true, Group: "engineering"},
	}
}
