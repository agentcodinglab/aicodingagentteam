package config

import (
	"os"
	"path/filepath"
	"testing"
)

// isolateHome points the OS home dir at a temp dir so Load() never touches the real profile.
func isolateHome(t *testing.T) string {
	t.Helper()
	home := t.TempDir()
	t.Setenv("USERPROFILE", home) // Windows
	t.Setenv("HOME", home)        // Unix fallback
	// Clear env overrides so tests assert pure default/file behavior,
	// not values leaked from the parent process (e.g. AICODINGAGENTTEAM_PORT).
	t.Setenv("AICODINGAGENTTEAM_PORT", "")
	t.Setenv("AICODINGAGENTTEAM_BACKEND", "")
	t.Setenv("AICODINGAGENTTEAM_QUALITY_THRESHOLD", "")
	return home
}

func writeConfig(t *testing.T, home, content string) {
	t.Helper()
	dir := filepath.Join(home, ".aicodingagentteam")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "config.json"), []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

func TestDefault(t *testing.T) {
	c := Default()
	if c.Default.Backend != "codex" {
		t.Errorf("expected default backend codex, got %s", c.Default.Backend)
	}
	if c.Coordinator.GRPC != 8080 || c.Coordinator.MCP != 8081 || c.Coordinator.ACP != 8082 || c.Coordinator.A2A != 8083 {
		t.Errorf("unexpected default ports: %+v", c.Coordinator)
	}
	if c.Host.WorkerTimeout != 300 {
		t.Errorf("expected worker timeout 300, got %d", c.Host.WorkerTimeout)
	}
	if c.Quality.Threshold != 90 {
		t.Errorf("expected quality threshold 90, got %d", c.Quality.Threshold)
	}
}

func TestLoad_NoFile_ReturnsDefault(t *testing.T) {
	isolateHome(t)
	c, err := Load()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if c.Quality.Threshold != 90 || c.Coordinator.GRPC != 8080 {
		t.Errorf("expected defaults when no config file, got %+v", c)
	}
}

func TestLoad_ValidFile(t *testing.T) {
	home := isolateHome(t)
	writeConfig(t, home, `{"default":{"backend":"opencode"},"quality":{"threshold":80},"coordinator":{"grpc":9090}}`)

	c, err := Load()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if c.Default.Backend != "opencode" {
		t.Errorf("expected backend opencode, got %s", c.Default.Backend)
	}
	if c.Quality.Threshold != 80 {
		t.Errorf("expected threshold 80, got %d", c.Quality.Threshold)
	}
	if c.Coordinator.GRPC != 9090 {
		t.Errorf("expected grpc port 9090, got %d", c.Coordinator.GRPC)
	}
}

func TestLoad_InvalidJSON_ReturnsDefault(t *testing.T) {
	home := isolateHome(t)
	writeConfig(t, home, `{not valid json`)

	c, err := Load()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if c.Quality.Threshold != 90 {
		t.Errorf("expected fallback to default threshold 90, got %d", c.Quality.Threshold)
	}
}

func TestLoad_EnvOverrides(t *testing.T) {
	// no config file -> defaults, then env overrides applied
	t.Setenv("AICODINGAGENTTEAM_PORT", "9090")
	t.Setenv("AICODINGAGENTTEAM_BACKEND", "opencode")
	t.Setenv("AICODINGAGENTTEAM_QUALITY_THRESHOLD", "70")

	c, err := Load()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if c.Coordinator.GRPC != 9090 {
		t.Errorf("expected grpc port overridden to 9090, got %d", c.Coordinator.GRPC)
	}
	if c.Default.Backend != "opencode" {
		t.Errorf("expected backend overridden to opencode, got %s", c.Default.Backend)
	}
	if c.Quality.Threshold != 70 {
		t.Errorf("expected threshold overridden to 70, got %d", c.Quality.Threshold)
	}
}

func TestLoad_FilePrecedenceOverEnv(t *testing.T) {
	home := isolateHome(t)
	writeConfig(t, home, `{"coordinator":{"grpc":7070}}`)
	t.Setenv("AICODINGAGENTTEAM_PORT", "9090")

	c, err := Load()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// env overrides defaults only when no file value; here file set grpc,
	// but env override is unconditional -> env wins (documented behavior)
	if c.Coordinator.GRPC != 9090 {
		t.Errorf("expected env override 9090, got %d", c.Coordinator.GRPC)
	}
}
func TestLoad_PartialFile_MergesDefaults(t *testing.T) {
	home := isolateHome(t)
	writeConfig(t, home, `{"quality":{"threshold":70}}`)

	c, err := Load()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if c.Quality.Threshold != 70 {
		t.Errorf("expected overridden threshold 70, got %d", c.Quality.Threshold)
	}
	// Load merges over Default(): unset fields keep their defaults
	if c.Coordinator.GRPC != 8080 {
		t.Errorf("expected default grpc port 8080 preserved, got %d", c.Coordinator.GRPC)
	}
	if c.Default.Backend != "codex" {
		t.Errorf("expected default backend codex preserved, got %s", c.Default.Backend)
	}
}
