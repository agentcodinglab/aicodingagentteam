// Package a2a implements the Redis-backed A2A bus for cross-container communication.
// This file provides a Redis Pub/Sub implementation of the Bus interface.
// When AICODINGAGENTTEAM_A2A_BUS env var is set (e.g. redis://host:6379),
// the Redis Bus is used instead of the in-process channel Bus.
package a2a

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/redis/go-redis/v9"

	"github.com/agentcodinglab/aicodingagentteam/internal/audit"
	"github.com/agentcodinglab/aicodingagentteam/internal/types"
)

// RedisBus implements the A2A message bus using Redis Pub/Sub.
// It is used when running in containerized mode across multiple containers.
type RedisBus struct {
	mu       sync.RWMutex
	agents   map[types.Role]Agent
	client   *redis.Client
	audit    *audit.Logger
	progress chan ProgressEvent
}

// NewRedisBus creates a Redis-backed A2A bus from a connection URL.
// URL format: redis://host:port or redis://user:pass@host:port
func NewRedisBus(redisURL string, al *audit.Logger) (*RedisBus, error) {
	opts, err := redis.ParseURL(redisURL)
	if err != nil {
		return nil, fmt.Errorf("parse redis URL: %w", err)
	}
	client := redis.NewClient(opts)

	// Test connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("redis ping: %w", err)
	}

	bus := &RedisBus{
		agents:   make(map[types.Role]Agent),
		client:   client,
		audit:    al,
		progress: make(chan ProgressEvent, 64),
	}
	return bus, nil
}

// Register registers a role agent on the bus and subscribes to its channel.
func (b *RedisBus) Register(a Agent) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.agents[a.Card().Role] = a

	// Subscribe to Redis channel for this role
	role := string(a.Card().Role)
	go b.subscribe(context.Background(), role)
}

// subscribe listens for A2A messages on the Redis channel for a given role.
func (b *RedisBus) subscribe(ctx context.Context, role string) {
	pubsub := b.client.Subscribe(ctx, "a2a:"+role)
	defer func() { _ = pubsub.Close() }()

	ch := pubsub.Channel()
	for {
		select {
		case <-ctx.Done():
			return
		case msg, ok := <-ch:
			if !ok {
				return
			}
			b.handleRedisMessage(ctx, msg.Payload)
		}
	}
}

// handleRedisMessage processes an A2A message received from Redis.
func (b *RedisBus) handleRedisMessage(ctx context.Context, payload string) {
	var task Task
	if err := json.Unmarshal([]byte(payload), &task); err != nil {
		return
	}

	b.emitProgress(task.TaskID, string(task.PhaseFromRole()), "running", "")
	b.logAudit(task.TaskID, string(task.Role), "redis-delegate", "dispatched")

	b.mu.RLock()
	a, ok := b.agents[task.Role]
	b.mu.RUnlock()
	if !ok {
		return
	}

	result, err := a.Execute(ctx, task)
	if err != nil {
		result = Result{
			TaskID: task.TaskID,
			Verdict: types.Verdict{
				TaskID:   task.TaskID,
				Role:     task.Role,
				Decision: types.DecisionBlocking,
				Findings: []types.Finding{{
					Check:  "agent-execution",
					Status: "fail",
					Detail: err.Error(),
				}},
			},
		}
		b.emitProgress(task.TaskID, "", "failed", err.Error())
		b.logAudit(task.TaskID, string(task.Role), "redis-agent-error", err.Error())
	} else {
		status := "completed"
		if result.Verdict.Decision == types.DecisionBlocking {
			status = "parked"
		}
		b.emitProgress(task.TaskID, "", status, string(result.Verdict.Decision))
		b.logAudit(task.TaskID, string(task.Role), "redis-agent-result", string(result.Verdict.Decision))
	}

	// Publish result back to coordinator channel
	resultData, _ := json.Marshal(result)
	_ = b.client.Publish(ctx, "a2a:coordinator", string(resultData))
}

// Delegate sends a task to the agent for the task's role via Redis Pub/Sub.
func (b *RedisBus) Delegate(ctx context.Context, task Task) (Result, error) {
	b.mu.RLock()
	a, ok := b.agents[task.Role]
	b.mu.RUnlock()
	if !ok {
		err := fmt.Errorf("no agent registered for role %s", task.Role)
		b.logAudit(task.TaskID, string(task.Role), "redis-delegate-failed", err.Error())
		return Result{}, err
	}

	// If agent is in-process (same container), execute directly
	result, err := a.Execute(ctx, task)
	if err != nil {
		result = Result{
			TaskID: task.TaskID,
			Verdict: types.Verdict{
				TaskID:   task.TaskID,
				Role:     task.Role,
				Decision: types.DecisionBlocking,
				Findings: []types.Finding{{
					Check:  "agent-execution",
					Status: "fail",
					Detail: err.Error(),
				}},
			},
		}
		b.emitProgress(task.TaskID, "", "failed", err.Error())
		b.logAudit(task.TaskID, string(task.Role), "redis-agent-error", err.Error())
		return result, err
	}

	status := "completed"
	if result.Verdict.Decision == types.DecisionBlocking {
		status = "parked"
	}
	b.emitProgress(task.TaskID, "", status, string(result.Verdict.Decision))
	b.logAudit(task.TaskID, string(task.Role), "redis-agent-result", string(result.Verdict.Decision))
	return result, nil
}

// DelegateParallel dispatches multiple tasks concurrently.
func (b *RedisBus) DelegateParallel(ctx context.Context, tasks []Task) []Result {
	var wg sync.WaitGroup
	results := make([]Result, len(tasks))

	for i, task := range tasks {
		wg.Add(1)
		go func(idx int, t Task) {
			defer wg.Done()
			r, _ := b.Delegate(ctx, t)
			results[idx] = r
		}(i, task)
	}
	wg.Wait()
	return results
}

// Discover returns all registered agent cards.
func (b *RedisBus) Discover() []AgentCard {
	b.mu.RLock()
	defer b.mu.RUnlock()
	cards := make([]AgentCard, 0, len(b.agents))
	for _, a := range b.agents {
		cards = append(cards, a.Card())
	}
	return cards
}

// ProgressChan returns a read-only channel for streaming progress events.
func (b *RedisBus) ProgressChan() <-chan ProgressEvent {
	return b.progress
}

func (b *RedisBus) emitProgress(taskID, phase, status, msg string) {
	select {
	case b.progress <- ProgressEvent{TaskID: taskID, Phase: phase, Status: status, Message: msg}:
	default:
	}
}

func (b *RedisBus) logAudit(taskID, agent, eventType, detail string) {
	if b.audit == nil {
		return
	}
	_ = b.audit.Log(audit.Entry{
		Type:   "a2a_message",
		Task:   taskID,
		Agent:  agent,
		Result: eventType,
		Detail: detail,
	})
}

// IsRegistered checks if an agent for the given role is registered.
func (b *RedisBus) IsRegistered(role types.Role) bool {
	b.mu.RLock()
	defer b.mu.RUnlock()
	_, ok := b.agents[role]
	return ok
}

// Count returns the number of registered agents.
func (b *RedisBus) Count() int {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return len(b.agents)
}

// Close closes the Redis connection.
func (b *RedisBus) Close() error {
	return b.client.Close()
}

// NewBusFromEnv creates the appropriate Bus based on environment configuration.
// If AICODINGAGENTTEAM_A2A_BUS is set to a redis:// URL, uses RedisBus.
// Otherwise, uses the in-process Bus (for development).
func NewBusFromEnv(al *audit.Logger) Bus {
	busURL := os.Getenv("AICODINGAGENTTEAM_A2A_BUS")
	if busURL != "" && strings.HasPrefix(busURL, "redis://") {
		bus, err := NewRedisBus(busURL, al)
		if err != nil {
			// Fail-open: fall back to in-process bus if Redis is unavailable
			fmt.Fprintf(os.Stderr, "WARNING: redis bus unavailable, falling back to in-process: %v\n", err)
			return NewBusWithAudit(al)
		}
		fmt.Println("A2A bus: Redis Pub/Sub mode")
		return bus
	}
	return NewBus()
}
