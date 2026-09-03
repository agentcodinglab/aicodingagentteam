// Package acp implements the Agent Client Protocol server.
package acp

import "context"

// Server handles ACP session lifecycle for standard agent clients.
type Server struct{}

// New creates an ACP Server.
func New() *Server { return &Server{} }

// Serve starts the ACP server (stub: real impl uses stdio JSON-RPC).
func (s *Server) Serve(ctx context.Context) error {
	<-ctx.Done()
	return ctx.Err()
}
