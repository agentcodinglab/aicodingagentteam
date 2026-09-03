// Package mcp exposes governance and orchestration capabilities via MCP.
package mcp

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/agentcodinglab/aicodingagentteam/internal/governance"
)

// GovernanceResult holds the output of a govern_file scan.
type GovernanceResult struct {
	Path       string
	Violations []governance.Violation
	Blocking   bool
}

// Server exposes MCP tools to external MCP clients.
type Server struct {
	engine *governance.Engine
}

// New creates an MCP Server with the given governance engine.
func New(engine *governance.Engine) *Server {
	return &Server{engine: engine}
}

// GovernFile runs governance on a single file and returns violations.
func (s *Server) GovernFile(ctx context.Context, path string) (*GovernanceResult, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read file %s: %w", path, err)
	}
	violations := s.engine.Check(ctx, path, string(content))
	return &GovernanceResult{
		Path:       path,
		Violations: violations,
		Blocking:   governance.HasBlocking(violations),
	}, nil
}

// GovernDirectory runs governance on all code files in a directory tree.
func (s *Server) GovernDirectory(ctx context.Context, root string) ([]GovernanceResult, error) {
	var results []GovernanceResult
	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}
		ext := strings.ToLower(filepath.Ext(path))
		if ext != ".go" && ext != ".ts" && ext != ".tsx" && ext != ".js" &&
			ext != ".jsx" && ext != ".py" && ext != ".java" && ext != ".rb" {
			return nil
		}
		content, err := os.ReadFile(path)
		if err != nil {
			return nil
		}
		violations := s.engine.Check(ctx, path, string(content))
		if len(violations) > 0 {
			results = append(results, GovernanceResult{
				Path:       path,
				Violations: violations,
				Blocking:   governance.HasBlocking(violations),
			})
		}
		return nil
	})
	return results, err
}

// Serve starts the MCP server (stub: real impl uses stdio JSON-RPC).
func (s *Server) Serve(ctx context.Context) error {
	<-ctx.Done()
	return ctx.Err()
}
