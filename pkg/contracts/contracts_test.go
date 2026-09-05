package contracts

import "testing"

func TestCrossCheck_NoMismatches(t *testing.T) {
	c := &Contract{
		Paths: map[string]PathItem{
			"/api/users": {Get: &Operation{OperationID: "listUsers", Method: "GET", Path: "/api/users"}},
		},
	}
	calls := []CallSite{
		{File: "frontend.ts", Method: "GET", Path: "/api/users", Line: 10},
	}
	mismatches := CrossCheck(c, calls)
	if len(mismatches) != 0 {
		t.Errorf("expected 0 mismatches, got %d: %v", len(mismatches), mismatches)
	}
}

func TestCrossCheck_PathNotInContract(t *testing.T) {
	c := &Contract{
		Paths: map[string]PathItem{
			"/api/users": {Get: &Operation{}},
		},
	}
	calls := []CallSite{
		{File: "app.ts", Method: "GET", Path: "/api/unknown", Line: 5},
	}
	mismatches := CrossCheck(c, calls)
	if len(mismatches) != 1 {
		t.Fatalf("expected 1 mismatch, got %d", len(mismatches))
	}
	if mismatches[0].Severity != "blocking" {
		t.Errorf("expected blocking severity, got %s", mismatches[0].Severity)
	}
	if mismatches[0].Call.Path != "/api/unknown" {
		t.Errorf("unexpected path: %s", mismatches[0].Call.Path)
	}
}

func TestCrossCheck_MethodNotAllowed(t *testing.T) {
	c := &Contract{
		Paths: map[string]PathItem{
			"/api/users": {Get: &Operation{OperationID: "listUsers", Method: "GET"}},
		},
	}
	calls := []CallSite{
		{File: "app.ts", Method: "DELETE", Path: "/api/users", Line: 12},
	}
	mismatches := CrossCheck(c, calls)
	if len(mismatches) != 1 {
		t.Fatalf("expected 1 mismatch, got %d", len(mismatches))
	}
	if mismatches[0].Severity != "blocking" {
		t.Errorf("expected blocking, got %s", mismatches[0].Severity)
	}
}

func TestCrossCheck_AllMethods(t *testing.T) {
	c := &Contract{
		Paths: map[string]PathItem{
			"/api/r": {
				Get:    &Operation{Method: "GET"},
				Post:   &Operation{Method: "POST"},
				Put:    &Operation{Method: "PUT"},
				Delete: &Operation{Method: "DELETE"},
				Patch:  &Operation{Method: "PATCH"},
			},
		},
	}
	for _, m := range []string{"GET", "POST", "PUT", "DELETE", "PATCH"} {
		calls := []CallSite{{Method: m, Path: "/api/r"}}
		if mm := CrossCheck(c, calls); len(mm) != 0 {
			t.Errorf("method %s should be allowed, got mismatches: %v", m, mm)
		}
	}
}

func TestMethodExists_UnknownMethod(t *testing.T) {
	item := PathItem{Get: &Operation{}}
	if methodExists(item, "HEAD") {
		t.Error("HEAD should not be recognized")
	}
}

func TestCrossCheck_EmptyContract(t *testing.T) {
	c := &Contract{Paths: map[string]PathItem{}}
	calls := []CallSite{{File: "f.ts", Method: "GET", Path: "/anything"}}
	mismatches := CrossCheck(c, calls)
	if len(mismatches) != 1 {
		t.Fatalf("expected 1 mismatch for empty contract, got %d", len(mismatches))
	}
}
