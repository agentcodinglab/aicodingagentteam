package acp

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/agentcodinglab/aicodingagentteam/internal/types"
)

type fakeDirector struct {
	mu     sync.Mutex
	calls  int
	failOn int
}

func (f *fakeDirector) Handle(ctx context.Context, req types.UserRequest) (*types.Delivery, error) {
	f.mu.Lock()
	f.calls++
	shouldFail := f.failOn > 0 && f.calls == f.failOn
	f.mu.Unlock()
	if shouldFail {
		return nil, fmt.Errorf("injected failure on call %d", f.calls)
	}
	d := types.Delivery{PlanID: "plan-x", Score: 100, Passed: true, Artifacts: []string{"src/a.go", "src/b.go"}}
	return &d, nil
}

func TestACP_SessionNewTask_StreamsEvents(t *testing.T) {
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
	go func() {
		_ = s.ServeReader(ctx, inR, outW)
		close(serverDone)
	}()

	clientR := bufio.NewReader(outR)

	writeReq := func(t *testing.T, req map[string]interface{}) {
		t.Helper()
		b, err := json.Marshal(req)
		if err != nil {
			t.Fatalf("marshal: %v", err)
		}
		b = append(b, byte(10)) // newline
		if _, err := inW.Write(b); err != nil {
			t.Fatalf("write: %v", err)
		}
	}

	readMessage := func(t *testing.T) map[string]interface{} {
		t.Helper()
		for {
			line, err := clientR.ReadString(byte(10))
			if err != nil {
				t.Fatalf("read: %v", err)
			}
			if line == "" || line == "\n" {
				continue
			}
			var m map[string]interface{}
			if err := json.Unmarshal([]byte(line), &m); err != nil {
				t.Fatalf("unmarshal: %v raw=%q", err, line)
			}
			return m
		}
	}

	// 1) session/start
	writeReq(t, map[string]interface{}{
		"jsonrpc": "2.0",
		"id":      1,
		"method":  "session/start",
		"params":  map[string]string{"agentId": "codex"},
	})
	startResp := readMessage(t)
	result, _ := startResp["result"].(map[string]interface{})
	sessionID, _ := result["sessionId"].(string)
	if sessionID == "" {
		t.Fatalf("no sessionId in %v", startResp)
	}

	// 2) session/newTask
	writeReq(t, map[string]interface{}{
		"jsonrpc": "2.0",
		"id":      2,
		"method":  "session/newTask",
		"params": map[string]interface{}{
			"sessionId": sessionID,
			"agentId":   "codex",
			"prompt":    "build hello world",
		},
	})
	newTaskResp := readMessage(t)
	taskResult, _ := newTaskResp["result"].(map[string]interface{})
	taskID, _ := taskResult["taskId"].(string)
	if taskID == "" {
		t.Fatalf("no taskId in %v", newTaskResp)
	}

	// 3) Wait for dispatch goroutine to push 4 notifications.
	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		notifMu.Lock()
		n := len(notifs)
		notifMu.Unlock()
		if n >= 4 {
			break
		}
		time.Sleep(20 * time.Millisecond)
	}

	notifMu.Lock()
	got := append([]string(nil), notifs...)
	notifMu.Unlock()

	if len(got) < 4 {
		t.Fatalf("expected at least 4 notifications, got %d: %v", len(got), got)
	}
	for i, m := range got[:4] {
		if m != "notifications/session/update" {
			t.Errorf("notif[%d] = %s, want notifications/session/update", i, m)
		}
	}

	cancel()
	inW.Close()
	outW.Close()
	<-serverDone
}

// silences unused-import when test never uses strings
var _ = strings.NewReader
