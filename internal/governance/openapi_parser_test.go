package governance

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

// TestLoadOpenAPIPaths_JSONParsed verifies JSON parsing via encoding/json
func TestLoadOpenAPIPaths_JSONParsed(t *testing.T) {
	spec := `{
		"openapi": "3.0.0",
		"paths": {
			"/api/users": {"get": {}},
			"/api/users/{id}": {"get": {}, "put": {}},
			"/api/posts": {"post": {}}
		}
	}`
	paths := loadOpenAPIPaths([]byte(spec))

	if !paths["/api/users"] {
		t.Error("expected /api/users in paths")
	}
	if !paths["/api/users/{id}"] {
		t.Error("expected /api/users/{id} in paths")
	}
	if !paths["/api/posts"] {
		t.Error("expected /api/posts in paths")
	}
}

// TestLoadOpenAPIPaths_CompactJSON verifies compact single-line JSON works
func TestLoadOpenAPIPaths_CompactJSON(t *testing.T) {
	spec := `{"paths":{"/api/users":{"get":{}},"/api/posts":{"post":{}}}}`
	paths := loadOpenAPIPaths([]byte(spec))

	if !paths["/api/users"] {
		t.Error("expected /api/users in compact JSON")
	}
	if !paths["/api/posts"] {
		t.Error("expected /api/posts in compact JSON")
	}
}

// TestLoadOpenAPIPaths_EmptySpec verifies empty spec returns empty map
func TestLoadOpenAPIPaths_EmptySpec(t *testing.T) {
	paths := loadOpenAPIPaths([]byte(`{"paths":{}}`))
	if len(paths) != 0 {
		t.Errorf("expected 0 paths, got %d", len(paths))
	}
}

// TestLoadOpenAPIPaths_InvalidJSON verifies invalid JSON falls back to line scanning
func TestLoadOpenAPIPaths_InvalidJSON(t *testing.T) {
	spec := `not valid json at all
"/api/fallback": {}
	"/api/another": {}`
	paths := loadOpenAPIPaths([]byte(spec))
	if !paths["/api/fallback"] {
		t.Error("expected /api/fallback via fallback scanning")
	}
}

// TestLoadOpenAPIPaths_RealFile verifies reading a real openapi.json file
func TestLoadOpenAPIPaths_RealFile(t *testing.T) {
	dir := t.TempDir()
	spec := map[string]any{
		"openapi": "3.0.0",
		"paths": map[string]any{
			"/api/items": map[string]any{"get": map[string]any{}},
		},
	}
	data, _ := json.Marshal(spec)
	specPath := filepath.Join(dir, "openapi.json")
	_ = os.WriteFile(specPath, data, 0644)

	loaded, _ := os.ReadFile(specPath)
	paths := loadOpenAPIPaths(loaded)
	if !paths["/api/items"] {
		t.Error("expected /api/items from real file")
	}
}
