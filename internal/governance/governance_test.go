package governance

import (
	"context"
	"testing"
)

func TestCheckNoViolations(t *testing.T) {
	e := New()
	violations := e.Check(context.Background(), "main.go", "package main\n\nfunc main() {}\n")
	if len(violations) != 0 {
		t.Errorf("expected 0 violations, got %d: %v", len(violations), violations)
	}
}

func TestSecretLeakAWSKey(t *testing.T) {
	e := New()
	content := `const key = "AKIAIOSFODNN7EXAMPLE"`
	violations := e.Check(context.Background(), "config.go", content)
	if len(violations) == 0 {
		t.Fatal("expected secret leak violation for AWS key")
	}
	found := false
	for _, v := range violations {
		if v.RuleID == "sec-secret-leak" && v.Severity == "blocking" {
			found = true
		}
	}
	if !found {
		t.Error("expected sec-secret-leak blocking violation")
	}
}

func TestSecretLeakGitHubToken(t *testing.T) {
	e := New()
	content := `token := "ghp_ABCDEFGHIJKLMNOPQRSTUVWXYZabcdef12"`
	violations := e.Check(context.Background(), "auth.go", content)
	found := false
	for _, v := range violations {
		if v.RuleID == "sec-secret-leak" {
			found = true
		}
	}
	if !found {
		t.Error("expected sec-secret-leak for GitHub token")
	}
}

func TestSecretLeakPrivateKey(t *testing.T) {
	e := New()
	content := `-----BEGIN RSA PRIVATE KEY-----\nMIIE...`
	violations := e.Check(context.Background(), "key.pem", content)
	found := false
	for _, v := range violations {
		if v.RuleID == "sec-secret-leak" {
			found = true
		}
	}
	if !found {
		t.Error("expected sec-secret-leak for private key")
	}
}

func TestSecretLeakSkipsTestFiles(t *testing.T) {
	e := New()
	content := `const key = "AKIAIOSFODNN7EXAMPLE"`
	violations := e.Check(context.Background(), "config_test.go", content)
	for _, v := range violations {
		if v.RuleID == "sec-secret-leak" {
			t.Error("should skip secret check in test files")
		}
	}
}

func TestEmojiInTSX(t *testing.T) {
	e := New()
	content := `export function Icon() { return <span>\u{1F600}</span> }`
	violations := e.Check(context.Background(), "component.tsx", content)
	found := false
	for _, v := range violations {
		if v.RuleID == "ui-emoji-icon" {
			found = true
		}
	}
	if !found {
		t.Error("expected emoji violation in .tsx file")
	}
}

func TestEmojiNotInGoFile(t *testing.T) {
	e := New()
	content := "// this has emoji \xf0\x9f\x98\x80 in comment"
	violations := e.Check(context.Background(), "main.go", content)
	for _, v := range violations {
		if v.RuleID == "ui-emoji-icon" {
			t.Error("emoji rule should not fire for .go files")
		}
	}
}

func TestHardcodedColor(t *testing.T) {
	e := New()
	content := `const style = { color: "#ff0000" }`
	violations := e.Check(context.Background(), "styles.tsx", content)
	found := false
	for _, v := range violations {
		if v.RuleID == "ui-hardcoded-color" {
			found = true
		}
	}
	if !found {
		t.Error("expected hardcoded color violation")
	}
}

func TestHardcodedColorSkipsComments(t *testing.T) {
	e := New()
	content := `// #ff0000 is red`
	violations := e.Check(context.Background(), "styles.css", content)
	for _, v := range violations {
		if v.RuleID == "ui-hardcoded-color" {
			t.Error("should skip comment lines")
		}
	}
}

func TestSQLInjection(t *testing.T) {
	e := New()
	content := `query := "SELECT * FROM users WHERE id = " + userId`
	violations := e.Check(context.Background(), "db.go", content)
	found := false
	for _, v := range violations {
		if v.RuleID == "sec-sql-injection" && v.Severity == "blocking" {
			found = true
		}
	}
	if !found {
		t.Error("expected SQL injection violation")
	}
}

func TestTodoPlaceholder(t *testing.T) {
	e := New()
	content := "// TODO: implement this later"
	violations := e.Check(context.Background(), "main.go", content)
	found := false
	for _, v := range violations {
		if v.RuleID == "eng-todo-placeholder" {
			found = true
		}
	}
	if !found {
		t.Error("expected TODO placeholder violation")
	}
}

func TestFakeData(t *testing.T) {
	e := New()
	content := `email := "test@example.com"`
	violations := e.Check(context.Background(), "user.go", content)
	found := false
	for _, v := range violations {
		if v.RuleID == "eng-fake-data" {
			found = true
		}
	}
	if !found {
		t.Error("expected fake data violation")
	}
}

func TestFakeDataSkipsTestFiles(t *testing.T) {
	e := New()
	content := `email := "test@example.com"`
	violations := e.Check(context.Background(), "user_test.go", content)
	for _, v := range violations {
		if v.RuleID == "eng-fake-data" {
			t.Error("should skip fake data check in test files")
		}
	}
}

func TestDisabledRuleProducesNoViolation(t *testing.T) {
	e := New()
	e.rules[0].Enabled = false // disable ui-emoji-icon
	content := `export function Icon() { return <span>\u{1F600}</span> }`
	violations := e.Check(context.Background(), "component.tsx", content)
	for _, v := range violations {
		if v.RuleID == "ui-emoji-icon" {
			t.Error("disabled rule should not produce violations")
		}
	}
}

func TestExcludedPathProducesNoViolation(t *testing.T) {
	e := New()
	e.config.Exclusions.Paths = []string{"src/legacy/**"}
	content := `const key = "AKIAIOSFODNN7EXAMPLE"`
	violations := e.Check(context.Background(), "src/legacy/config.go", content)
	if len(violations) != 0 {
		t.Errorf("excluded path should produce 0 violations, got %d", len(violations))
	}
}

func TestHasBlocking(t *testing.T) {
	blocking := []Violation{{Severity: "blocking"}}
	if !HasBlocking(blocking) {
		t.Error("expected HasBlocking=true for blocking violations")
	}
	advisory := []Violation{{Severity: "advisory"}}
	if HasBlocking(advisory) {
		t.Error("expected HasBlocking=false for advisory-only violations")
	}
	if HasBlocking(nil) {
		t.Error("expected HasBlocking=false for empty violations")
	}
}

func TestClosedEngineReturnsNil(t *testing.T) {
	e := New()
	e.Close()
	violations := e.Check(context.Background(), "main.go", "const key = \"AKIAIOSFODNN7EXAMPLE\"")
	if violations != nil {
		t.Error("closed engine should return nil")
	}
}

func TestRulesList(t *testing.T) {
	e := New()
	rules := e.Rules()
	if len(rules) != 7 {
		t.Errorf("expected 7 rules, got %d", len(rules))
	}
	ids := make(map[string]bool)
	for _, r := range rules {
		ids[r.ID] = true
	}
	expected := []string{"ui-emoji-icon", "ui-hardcoded-color", "sec-secret-leak", "sec-sql-injection", "api-contract-mismatch", "eng-todo-placeholder", "eng-fake-data"}
	for _, id := range expected {
		if !ids[id] {
			t.Errorf("missing rule: %s", id)
		}
	}
}
