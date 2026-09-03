package router

import (
	"context"
	"testing"

	"github.com/agentcodinglab/aicodingagentteam/internal/types"
)

func TestRoute_BuildIntent(t *testing.T) {
	r := New()
	cases := []struct {
		msg  string
		want types.IntentType
	}{
		{"搭建一个记账应用", types.IntentBuild},
		{"build a web app", types.IntentBuild},
		{"修复登录bug", types.IntentDebug},
		{"修改按钮颜色", types.IntentQuickEdit},
		{"explain this function", types.IntentExplain},
		{"你好", types.IntentChat},
	}
	for _, c := range cases {
		got := r.Route(context.Background(), types.UserRequest{Message: c.msg}).Type
		if got != c.want {
			t.Errorf("Route(%q) = %v, want %v", c.msg, got, c.want)
		}
	}
}

func TestRoute_WriteAccess(t *testing.T) {
	r := New()
	intent := r.Route(context.Background(), types.UserRequest{Message: "搭建应用"})
	if !intent.WriteAccess {
		t.Error("Build intent should have write access")
	}
	intent = r.Route(context.Background(), types.UserRequest{Message: "你好"})
	if intent.WriteAccess {
		t.Error("Chat intent should not have write access")
	}
}
