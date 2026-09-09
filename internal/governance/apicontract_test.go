package governance

import (
	"os"
	"path/filepath"
	"testing"
)

func TestCheckAPIContract_NoOpenAPISpec(t *testing.T) {
	dir := t.TempDir()
	tsFile := filepath.Join(dir, "api.ts")
	_ = os.WriteFile(tsFile, []byte("fetch('/api/users')"), 0644)

	violations := checkAPIContract(tsFile, "fetch('/api/users')")
	// Without an OpenAPI spec co-located, no violations are reported
	if len(violations) != 0 {
		t.Errorf("expected 0 violations without spec, got %d", len(violations))
	}
}

func TestCheckAPIContract_MismatchDetected(t *testing.T) {
	dir := t.TempDir()
	tsFile := filepath.Join(dir, "api.ts")
	specFile := filepath.Join(dir, "openapi.json")

	_ = os.WriteFile(tsFile, []byte("fetch('/api/unknown')"), 0644)
	_ = os.WriteFile(specFile, []byte(`{
	"paths": {
		"/api/users": {}
	}
}`), 0644)

	violations := checkAPIContract(tsFile, "fetch('/api/unknown')")
	if len(violations) == 0 {
		t.Error("expected violation for mismatched path")
	}
	if violations[0].RuleID != "api-contract-mismatch" {
		t.Errorf("expected rule api-contract-mismatch, got %s", violations[0].RuleID)
	}
}

func TestCheckAPIContract_PathMatchesSpec(t *testing.T) {
	dir := t.TempDir()
	tsFile := filepath.Join(dir, "api.ts")
	specFile := filepath.Join(dir, "openapi.json")

	_ = os.WriteFile(tsFile, []byte("fetch('/api/users')"), 0644)
	_ = os.WriteFile(specFile, []byte(`{
	"paths": {
		"/api/users": {}
	}
}`), 0644)

	violations := checkAPIContract(tsFile, "fetch('/api/users')")
	if len(violations) != 0 {
		t.Errorf("expected 0 violations for matching path, got %d", len(violations))
	}
}

func TestCheckAPIContract_PathParamMatch(t *testing.T) {
	dir := t.TempDir()
	tsFile := filepath.Join(dir, "api.ts")
	specFile := filepath.Join(dir, "openapi.json")

	_ = os.WriteFile(tsFile, []byte("fetch('/api/users/123')"), 0644)
	_ = os.WriteFile(specFile, []byte(`{
	"paths": {
		"/api/users/{id}": {}
	}
}`), 0644)

	violations := checkAPIContract(tsFile, "fetch('/api/users/123')")
	if len(violations) != 0 {
		t.Errorf("expected 0 violations for path param match, got %d", len(violations))
	}
}

func TestCheckAPIContract_NonJSFile(t *testing.T) {
	violations := checkAPIContract("main.go", "fetch('/api/users')")
	if len(violations) != 0 {
		t.Errorf("expected 0 violations for non-JS file, got %d", len(violations))
	}
}

func TestCheckAPIContract_NoFetchOrAxios(t *testing.T) {
	violations := checkAPIContract("api.ts", "const x = 1")
	if len(violations) != 0 {
		t.Errorf("expected 0 violations without fetch/axios, got %d", len(violations))
	}
}

func TestCheckAPIContract_AxiosCall(t *testing.T) {
	dir := t.TempDir()
	tsFile := filepath.Join(dir, "api.ts")
	specFile := filepath.Join(dir, "openapi.json")

	_ = os.WriteFile(tsFile, []byte("axios.get('/api/unknown')"), 0644)
	_ = os.WriteFile(specFile, []byte(`{
	"paths": {
		"/api/users": {}
	}
}`), 0644)

	violations := checkAPIContract(tsFile, "axios.get('/api/unknown')")
	if len(violations) == 0 {
		t.Error("expected violation for axios mismatch")
	}
}

func TestCheckAPIContract_QueryStringStripped(t *testing.T) {
	dir := t.TempDir()
	tsFile := filepath.Join(dir, "api.ts")
	specFile := filepath.Join(dir, "openapi.json")

	_ = os.WriteFile(tsFile, []byte("fetch('/api/users?page=1')"), 0644)
	_ = os.WriteFile(specFile, []byte(`{
	"paths": {
		"/api/users": {}
	}
}`), 0644)

	violations := checkAPIContract(tsFile, "fetch('/api/users?page=1')")
	if len(violations) != 0 {
		t.Errorf("expected 0 violations after stripping query, got %d", len(violations))
	}
}

func TestExtractPathFromURL_StripQuery(t *testing.T) {
	got := extractPathFromURL("/api/users?page=1&size=10")
	if got != "/api/users" {
		t.Errorf("expected /api/users, got %s", got)
	}
}

func TestExtractPathFromURL_StripHost(t *testing.T) {
	got := extractPathFromURL("https://example.com/api/users")
	if got != "/api/users" {
		t.Errorf("expected /api/users, got %s", got)
	}
}

func TestExtractPathFromURL_PlainPath(t *testing.T) {
	got := extractPathFromURL("/api/users")
	if got != "/api/users" {
		t.Errorf("expected /api/users, got %s", got)
	}
}

func TestExtractPathFromURL_NoPath(t *testing.T) {
	got := extractPathFromURL("https://example.com")
	if got != "" {
		t.Errorf("expected empty string, got %s", got)
	}
}

func TestLoadOpenAPIPaths_ExtractsPaths(t *testing.T) {
	data := []byte(`{
		"openapi": "3.0",
		"paths": {
			"/api/users": {},
			"/api/posts": {}
		}
	}`)
	paths := loadOpenAPIPaths(data)
	if !paths["/api/users"] {
		t.Error("expected /api/users in paths")
	}
	if !paths["/api/posts"] {
		t.Error("expected /api/posts in paths")
	}
}

func TestPathPrefixMatch_ExactMatch(t *testing.T) {
	if !pathPrefixMatch("/api/users", "/api/users") {
		t.Error("expected exact match")
	}
}

func TestPathPrefixMatch_ParamMatch(t *testing.T) {
	if !pathPrefixMatch("/api/users/123", "/api/users/{id}") {
		t.Error("expected param match")
	}
}

func TestPathPrefixMatch_DifferentLength(t *testing.T) {
	if pathPrefixMatch("/api/users", "/api/users/{id}") {
		t.Error("expected no match for different length")
	}
}

func TestPathPrefixMatch_SegmentMismatch(t *testing.T) {
	if pathPrefixMatch("/api/posts/123", "/api/users/{id}") {
		t.Error("expected no match for different segment")
	}
}
