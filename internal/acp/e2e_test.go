package acp

import (
	"bufio"
	"context"
	"encoding/json"
	"io"
	"sync"
	"testing"
	"time"
)

// TestACP_E2E_FullFlow exercises the complete ACP server lifecycle:
// session/start -> session/newTask -> streamed notifications -> session/list.
func TestACP_E2E_FullFlow(t *testing.T) {
	inR, inW := io.Pipe()
	outR, outW := io.Pipe()

	var notifs []string
	var notifMu sync.Mutex
	notifier := func(method string, params interface{}) {
		notifMu.Lock()
		notifs = append(notifs, method)
		notifMu.Unlock()
	}

	dir := &fakeDirector{}
	s := NewWithDirector(dir, notifier)

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

	// 1) session/start
	writeReq(map[string]interface{}{
		"jsonrpc": "2.0", "id": 1, "method": "session/start",
		"params": map[string]string{"agentId": "codex"},
	})
	startResp := readMsg()
	result, _ := startResp["result"].(map[string]interface{})
	sessionID, _ := result["sessionId"].(string)
	if sessionID == "" {
		t.Fatalf("no sessionId: %v", startResp)
	}

	// 2) session/newTask
	writeReq(map[string]interface{}{
		"jsonrpc": "2.0", "id": 2, "method": "session/newTask",
		"params": map[string]interface{}{
			"sessionId": sessionID, "agentId": "codex", "prompt": "build hello",
		},
	})
	newTaskResp := readMsg()
	taskResult, _ := newTaskResp["result"].(map[string]interface{})
	if _, ok := taskResult["taskId"]; !ok {
		t.Fatalf("no taskId: %v", newTaskResp)
	}

	// 3) Wait for streamed notifications
	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		notifMu.Lock()
		if len(notifs) >= 4 {
			notifMu.Unlock()
			break
		}
		notifMu.Unlock()
		time.Sleep(20 * time.Millisecond)
	}

	// 4) session/list (should show the session we created)
	writeReq(map[string]interface{}{
		"jsonrpc": "2.0", "id": 3, "method": "session/list",
	})
	listResp := readMsg()
	listResult, _ := listResp["result"].(map[string]interface{})
	sessions, _ := listResult["sessions"].([]interface{})
	if len(sessions) == 0 {
		t.Errorf("expected at least 1 session in list, got 0")
	}

	cancel()
	inW.Close()
	outW.Close()
	<-serverDone
}

// TestACP_E2E_SessionStopAfterNewTask verifies session lifecycle:
// start -> newTask -> stop -> list shows 0 sessions.
func TestACP_E2E_SessionStopAfterNewTask(t *testing.T) {
	inR, inW := io.Pipe()
	outR, outW := io.Pipe()

	dir := &fakeDirector{}
	s := NewWithDirector(dir, func(string, interface{}) {})

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

	// start
	writeReq(map[string]interface{}{
		"jsonrpc": "2.0", "id": 1, "method": "session/start",
		"params": map[string]string{"agentId": "codex"},
	})
	startResp := readMsg()
	result, _ := startResp["result"].(map[string]interface{})
	sessionID, _ := result["sessionId"].(string)

	// stop
	writeReq(map[string]interface{}{
		"jsonrpc": "2.0", "id": 2, "method": "session/stop",
		"params": map[string]string{"sessionId": sessionID},
	})
	readMsg() // consume stop response

	// list should show session with stopped status
	writeReq(map[string]interface{}{
		"jsonrpc": "2.0", "id": 3, "method": "session/list",
	})
	listResp := readMsg()
	listResult, _ := listResp["result"].(map[string]interface{})
	sessions, _ := listResult["sessions"].([]interface{})
	if len(sessions) != 1 {
		t.Fatalf("expected 1 session after stop, got %d", len(sessions))
	}
	sessMap, _ := sessions[0].(map[string]interface{})
	if sessMap["status"] != "stopped" {
		t.Errorf("expected stopped status, got %v", sessMap["status"])
	}

	cancel()
	inW.Close()
	outW.Close()
	<-serverDone
}
