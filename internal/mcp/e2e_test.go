package mcp

import (
	"bufio"
	"context"
	"encoding/json"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/agentcodinglab/aicodingagentteam/internal/governance"
)

// TestMCP_E2E_FullFlow exercises the complete MCP server lifecycle:
// initialize -> tools/list -> tools/call (govern_file) -> tools/call (govern_directory).
func TestMCP_E2E_FullFlow(t *testing.T) {
	dir := t.TempDir()
	_ = os.WriteFile(filepath.Join(dir, "bad.go"), []byte("package main\nvar token = \"ghp_1234567890abcdefghijklmnopqrstuvwxyz\"\n"), 0644)

	inR, inW := io.Pipe()
	outR, outW := io.Pipe()

	s := New(governance.New())
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	serverDone := make(chan struct{})
	go func() { _ = s.ServeReader(ctx, inR, outW); close(serverDone) }()

	clientR := bufio.NewReader(outR)
	writeReq := func(req map[string]interface{}) {
		b, _ := json.Marshal(req)
		b = append(b, '\n')
		_, _ = inW.Write(b)
	}
	readMsg := func() map[string]interface{} {
		for {
			line, _ := clientR.ReadString('\n')
			if line == "" || line == "\n" {
				continue
			}
			var m map[string]interface{}
			_ = json.Unmarshal([]byte(line), &m)
			return m
		}
	}

	// 1) initialize
	writeReq(map[string]interface{}{
		"jsonrpc": "2.0", "id": 1, "method": "initialize",
		"params": map[string]string{"clientInfo": "test"},
	})
	initResp := readMsg()
	if initResp["result"] == nil {
		t.Fatalf("initialize failed: %v", initResp)
	}

	// 2) tools/list
	writeReq(map[string]interface{}{
		"jsonrpc": "2.0", "id": 2, "method": "tools/list",
	})
	listResp := readMsg()
	result, _ := listResp["result"].(map[string]interface{})
	tools, _ := result["tools"].([]interface{})
	if len(tools) < 2 {
		t.Fatalf("expected at least 2 tools, got %d", len(tools))
	}

	// 3) tools/call govern_file
	writeReq(map[string]interface{}{
		"jsonrpc": "2.0", "id": 3, "method": "tools/call",
		"params": map[string]interface{}{
			"name": "govern_file",
			"arguments": map[string]interface{}{
				"path": filepath.Join(dir, "bad.go"),
			},
		},
	})
	fileResp := readMsg()
	if fileResp["error"] != nil {
		t.Errorf("govern_file error: %v", fileResp["error"])
	}

	// 4) tools/call govern_directory
	writeReq(map[string]interface{}{
		"jsonrpc": "2.0", "id": 4, "method": "tools/call",
		"params": map[string]interface{}{
			"name": "govern_directory",
			"arguments": map[string]interface{}{
				"root": dir,
			},
		},
	})
	dirResp := readMsg()
	if dirResp["error"] != nil {
		t.Errorf("govern_directory error: %v", dirResp["error"])
	}

	cancel()
	inW.Close()
	outW.Close()
	<-serverDone
}

// TestMCP_E2E_InvalidToolCall verifies error handling for unknown tools.
func TestMCP_E2E_InvalidToolCall(t *testing.T) {
	inR, inW := io.Pipe()
	outR, outW := io.Pipe()

	s := New(governance.New())
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	serverDone := make(chan struct{})
	go func() { _ = s.ServeReader(ctx, inR, outW); close(serverDone) }()

	clientR := bufio.NewReader(outR)
	writeReq := func(req map[string]interface{}) {
		b, _ := json.Marshal(req)
		b = append(b, '\n')
		_, _ = inW.Write(b)
	}
	readMsg := func() map[string]interface{} {
		for {
			line, _ := clientR.ReadString('\n')
			if line == "" || line == "\n" {
				continue
			}
			var m map[string]interface{}
			_ = json.Unmarshal([]byte(line), &m)
			return m
		}
	}

	// initialize first
	writeReq(map[string]interface{}{
		"jsonrpc": "2.0", "id": 1, "method": "initialize",
	})
	readMsg()

	// call unknown tool
	writeReq(map[string]interface{}{
		"jsonrpc": "2.0", "id": 2, "method": "tools/call",
		"params": map[string]interface{}{
			"name": "nonexistent_tool",
		},
	})
	resp := readMsg()
	errObj, _ := resp["error"].(map[string]interface{})
	if errObj == nil {
		t.Fatal("expected error for unknown tool")
	}
	if !strings.Contains(errObj["message"].(string), "unknown tool") {
		t.Errorf("expected 'unknown tool' in error, got %v", errObj["message"])
	}

	cancel()
	inW.Close()
	outW.Close()
	<-serverDone

}
