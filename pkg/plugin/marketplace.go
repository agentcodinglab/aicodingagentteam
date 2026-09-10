// Package plugin provides a global host driver registry and local marketplace
// index for community-contributed Runtime implementations.
package plugin

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// PluginMeta describes a community plugin entry in the local marketplace index.
type PluginMeta struct {
	Name        string `json:"name"`
	Backend     string `json:"backend"`
	Description string `json:"description"`
	ImportPath  string `json:"importPath"`
	Version     string `json:"version"`
	Author      string `json:"author"`
}

// Marketplace is a local index of discoverable plugins.
// It reads from a JSON index file on disk (default: .aicodingagentteam/plugins.json).
type Marketplace struct {
	indexPath string
}

// NewMarketplace creates a Marketplace rooted at the given project directory.
func NewMarketplace(projectDir string) *Marketplace {
	return &Marketplace{
		indexPath: filepath.Join(projectDir, ".aicodingagentteam", "plugins.json"),
	}
}

// Load reads the plugin index from disk. Returns an empty list if the file
// does not exist (not an error).
func (m *Marketplace) Load() ([]PluginMeta, error) {
	data, err := os.ReadFile(m.indexPath)
	if os.IsNotExist(err) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("read plugin index: %w", err)
	}
	var plugins []PluginMeta
	if err := json.Unmarshal(data, &plugins); err != nil {
		return nil, fmt.Errorf("parse plugin index: %w", err)
	}
	return plugins, nil
}

// Add registers a new plugin in the index. If a plugin with the same name
// already exists, it is replaced.
func (m *Marketplace) Add(p PluginMeta) error {
	plugins, err := m.Load()
	if err != nil {
		return err
	}
	found := false
	for i, existing := range plugins {
		if existing.Name == p.Name {
			plugins[i] = p
			found = true
			break
		}
	}
	if !found {
		plugins = append(plugins, p)
	}
	return m.save(plugins)
}

// Search returns plugins whose name, backend, or description matches the query (case-insensitive).
func (m *Marketplace) Search(query string) ([]PluginMeta, error) {
	plugins, err := m.Load()
	if err != nil {
		return nil, err
	}
	q := strings.ToLower(query)
	var results []PluginMeta
	for _, p := range plugins {
		if strings.Contains(strings.ToLower(p.Name), q) ||
			strings.Contains(strings.ToLower(p.Backend), q) ||
			strings.Contains(strings.ToLower(p.Description), q) {
			results = append(results, p)
		}
	}
	sort.Slice(results, func(i, j int) bool { return results[i].Name < results[j].Name })
	return results, nil
}

// Remove deletes a plugin from the index by name.
func (m *Marketplace) Remove(name string) error {
	plugins, err := m.Load()
	if err != nil {
		return err
	}
	var filtered []PluginMeta
	for _, p := range plugins {
		if p.Name != name {
			filtered = append(filtered, p)
		}
	}
	return m.save(filtered)
}

func (m *Marketplace) save(plugins []PluginMeta) error {
	dir := filepath.Dir(m.indexPath)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("create plugin index dir: %w", err)
	}
	data, err := json.MarshalIndent(plugins, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal plugin index: %w", err)
	}
	return os.WriteFile(m.indexPath, data, 0o644)
}
