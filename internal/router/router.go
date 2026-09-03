// Package router classifies user requests into intent types for scheduling.
package router

import (
	"context"
	"strings"

	"github.com/yourorg/aicodingagentteam/internal/types"
)

// Router routes user requests to intent classifications.
type Router struct{}

// New creates a Router.
func New() *Router { return &Router{} }

// Route evaluates the request and returns an Intent.
// This is a rule-based stub; a real implementation would use knowledge retrieval + clarification.
func (r *Router) Route(ctx context.Context, req types.UserRequest) types.Intent {
	msg := strings.ToLower(req.Message)
	intent := types.Intent{Scope: req.Message, WriteAccess: false}

	switch {
	case strings.Contains(msg, "build") || strings.Contains(msg, "搭建") || strings.Contains(msg, "创建项目"):
		intent.Type = types.IntentBuild
		intent.Depth = types.DepthBuild
		intent.WriteAccess = true
	case strings.Contains(msg, "fix") || strings.Contains(msg, "修复") || strings.Contains(msg, "bug"):
		intent.Type = types.IntentDebug
		intent.Depth = types.DepthFeature
		intent.WriteAccess = true
	case strings.Contains(msg, "修改") || strings.Contains(msg, "change") || strings.Contains(msg, "update"):
		intent.Type = types.IntentQuickEdit
		intent.Depth = types.DepthTrivial
		intent.WriteAccess = true
	case strings.Contains(msg, "explain") || strings.Contains(msg, "解释"):
		intent.Type = types.IntentExplain
		intent.Depth = types.DepthTrivial
	default:
		intent.Type = types.IntentChat
		intent.Depth = types.DepthTrivial
	}
	return intent
}
