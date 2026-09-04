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
// applyEnvOverrides lets the operator override ports/backend via env vars
// (AICODINGAGENTTEAM_GRPC_PORT etc.), matching the env contract documented
// for the TUI client. File config takes precedence over env; env over defaults.
func applyEnvOverrides(c *Config) {
	if v := os.Getenv("AICODINGAGENTTEAM_PORT"); v != "" {
		if p := atoiSafe(v); p > 0 {
			c.Coordinator.GRPC = p
		}
	}
	if v := os.Getenv("AICODINGAGENTTEAM_BACKEND"); v != "" {
		c.Default.Backend = v
	}
	if v := os.Getenv("AICODINGAGENTTEAM_QUALITY_THRESHOLD"); v != "" {
		if p := atoiSafe(v); p > 0 {
			c.Quality.Threshold = p
		}
	}
}

func atoiSafe(s string) int {
	n := 0
	for _, r := range s {
		if r < '0' || r > '9' {
			return 0
		}
		n = n*10 + int(r-'0')
	}
	return n
}
func Load() (*Config, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		c := Default()
		applyEnvOverrides(c)
		return c, nil
	}
	p := filepath.Join(home, ".aicodingagentteam", "config.json")
	data, err := os.ReadFile(p)
	if err != nil {
		c := Default()
		applyEnvOverrides(c)
		return c, nil
	}
	c := Default()
	if len(data) > 0 {
		if err := json.Unmarshal(data, c); err != nil {
			applyEnvOverrides(c)
			return c, nil
		}
	}
	applyEnvOverrides(c)
	return c, nil
}
