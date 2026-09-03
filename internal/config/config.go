// Package config loads global and project configuration.
package config

import (
	"encoding/json"
	"os"
	"path/filepath"
)

// Config is the merged global + project configuration.
type Config struct {
	Default     DefaultCfg `json:"default"`
	Coordinator Ports      `json:"coordinator"`
	Host        HostCfg    `json:"host"`
	Quality     QualityCfg `json:"quality"`
}

// DefaultCfg holds default settings.
type DefaultCfg struct {
	Backend          string `json:"backend"`
	AutoApproveGates bool   `json:"auto_approve_gates"`
}

// Ports holds server port assignments.
type Ports struct {
	GRPC int `json:"grpc"`
	MCP  int `json:"mcp"`
	ACP  int `json:"acp"`
	A2A  int `json:"a2a"`
}

// HostCfg holds host CLI paths and timeouts.
type HostCfg struct {
	ClaudeBin     string `json:"claude_bin"`
	CodexBin      string `json:"codex_bin"`
	OpenCodeBin   string `json:"opencode_bin"`
	DSHBin        string `json:"dsh_bin"`
	WorkerTimeout int    `json:"worker_timeout"`
}

// QualityCfg holds quality gate settings.
type QualityCfg struct {
	Threshold int `json:"threshold"`
}

// Default returns the default configuration.
func Default() *Config {
	return &Config{
		Default:     DefaultCfg{Backend: "codex"},
		Coordinator: Ports{GRPC: 8080, MCP: 8081, ACP: 8082, A2A: 8083},
		Host:        HostCfg{WorkerTimeout: 300},
		Quality:     QualityCfg{Threshold: 90},
	}
}

// Load reads config from ~/.aicodingagentteam/config.json if present.
func Load() (*Config, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return Default(), nil
	}
	p := filepath.Join(home, ".aicodingagentteam", "config.json")
	data, err := os.ReadFile(p)
	if err != nil {
		return Default(), nil
	}
	var c Config
	if err := json.Unmarshal(data, &c); err != nil {
		return Default(), nil
	}
	return &c, nil
}
