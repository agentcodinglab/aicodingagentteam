package plugin

import (
	"os"
	"path/filepath"
	"testing"
)

func TestMarketplace_Load_EmptyWhenNoFile(t *testing.T) {
	dir := t.TempDir()
	m := NewMarketplace(dir)
	plugins, err := m.Load()
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if len(plugins) != 0 {
		t.Errorf("expected 0 plugins, got %d", len(plugins))
	}
}

func TestMarketplace_Add_AndLoad(t *testing.T) {
	dir := t.TempDir()
	m := NewMarketplace(dir)
	p := PluginMeta{Name: "gemini-cli", Backend: "gemini", Description: "Gemini CLI driver", ImportPath: "github.com/x/gemini-driver", Version: "0.1.0", Author: "x"}
	if err := m.Add(p); err != nil {
		t.Fatalf("Add: %v", err)
	}
	plugins, err := m.Load()
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if len(plugins) != 1 || plugins[0].Name != "gemini-cli" {
		t.Fatalf("expected 1 plugin gemini-cli, got %v", plugins)
	}
}

func TestMarketplace_Add_ReplacesExisting(t *testing.T) {
	dir := t.TempDir()
	m := NewMarketplace(dir)
	_ = m.Add(PluginMeta{Name: "gemini-cli", Backend: "gemini", Version: "0.1.0"})
	_ = m.Add(PluginMeta{Name: "gemini-cli", Backend: "gemini", Version: "0.2.0"})
	plugins, _ := m.Load()
	if len(plugins) != 1 {
		t.Fatalf("expected 1 plugin, got %d", len(plugins))
	}
	if plugins[0].Version != "0.2.0" {
		t.Errorf("expected version 0.2.0, got %s", plugins[0].Version)
	}
}

func TestMarketplace_Search_ByBackend(t *testing.T) {
	dir := t.TempDir()
	m := NewMarketplace(dir)
	_ = m.Add(PluginMeta{Name: "gemini-cli", Backend: "gemini"})
	_ = m.Add(PluginMeta{Name: "qwen-cli", Backend: "qwen"})
	results, err := m.Search("gem")
	if err != nil {
		t.Fatalf("Search: %v", err)
	}
	if len(results) != 1 || results[0].Backend != "gemini" {
		t.Fatalf("expected 1 gemini result, got %v", results)
	}
}

func TestMarketplace_Search_CaseInsensitive(t *testing.T) {
	dir := t.TempDir()
	m := NewMarketplace(dir)
	_ = m.Add(PluginMeta{Name: "Gemini-CLI", Backend: "gemini", Description: "Google AI driver"})
	results, _ := m.Search("GEMINI")
	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
}

func TestMarketplace_Remove(t *testing.T) {
	dir := t.TempDir()
	m := NewMarketplace(dir)
	_ = m.Add(PluginMeta{Name: "gemini-cli", Backend: "gemini"})
	_ = m.Add(PluginMeta{Name: "qwen-cli", Backend: "qwen"})
	if err := m.Remove("gemini-cli"); err != nil {
		t.Fatalf("Remove: %v", err)
	}
	plugins, _ := m.Load()
	if len(plugins) != 1 || plugins[0].Name != "qwen-cli" {
		t.Fatalf("expected 1 qwen-cli, got %v", plugins)
	}
}

func TestMarketplace_Search_EmptyQuery_ReturnsAll(t *testing.T) {
	dir := t.TempDir()
	m := NewMarketplace(dir)
	_ = m.Add(PluginMeta{Name: "a", Backend: "a"})
	_ = m.Add(PluginMeta{Name: "b", Backend: "b"})
	results, _ := m.Search("")
	if len(results) != 2 {
		t.Fatalf("expected 2 results for empty query, got %d", len(results))
	}
}

func TestMarketplace_IndexFileCreated(t *testing.T) {
	dir := t.TempDir()
	m := NewMarketplace(dir)
	_ = m.Add(PluginMeta{Name: "test", Backend: "test"})
	path := filepath.Join(dir, ".aicodingagentteam", "plugins.json")
	if _, err := os.Stat(path); err != nil {
		t.Fatalf("index file not created: %v", err)
	}
}
