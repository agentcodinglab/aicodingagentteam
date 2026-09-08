// Package plugin provides a global host driver registry that enables
// community-contributed Runtime implementations to self-register at init time.
//
// Usage for driver authors:
//
//	import "github.com/agentcodinglab/aicodingagentteam/pkg/plugin"
//	import "github.com/agentcodinglab/aicodingagentteam/pkg/runtime"
//
//	func init() {
//	    plugin.Register("my-backend", NewMyDriver())
//	}
//
// Users then import the driver package (blank import) to activate it:
//
//	import _ "github.com/community/my-driver"
package plugin

import (
	"fmt"
	"sort"
	"sync"

	"github.com/agentcodinglab/aicodingagentteam/pkg/runtime"
)

var (
	mu       sync.RWMutex
	registry = make(map[runtime.Backend]runtime.Runtime)
)

// Register adds or replaces a host driver in the global plugin registry.
// Called from init() functions in community driver packages.
func Register(backend runtime.Backend, drv runtime.Runtime) {
	mu.Lock()
	defer mu.Unlock()
	registry[backend] = drv
}

// Get returns the driver for a backend, or nil if not registered.
func Get(backend runtime.Backend) (runtime.Runtime, bool) {
	mu.RLock()
	defer mu.RUnlock()
	drv, ok := registry[backend]
	return drv, ok
}

// All returns all registered plugin drivers keyed by backend.
// The returned map is a copy; mutations do not affect the global registry.
func All() map[runtime.Backend]runtime.Runtime {
	mu.RLock()
	defer mu.RUnlock()
	out := make(map[runtime.Backend]runtime.Runtime, len(registry))
	for k, v := range registry {
		out[k] = v
	}
	return out
}

// Backends returns all registered backend names, sorted alphabetically.
func Backends() []runtime.Backend {
	mu.RLock()
	defer mu.RUnlock()
	out := make([]runtime.Backend, 0, len(registry))
	for b := range registry {
		out = append(out, b)
	}
	sort.Slice(out, func(i, j int) bool { return string(out[i]) < string(out[j]) })
	return out
}

// Count returns the number of registered plugin drivers.
func Count() int {
	mu.RLock()
	defer mu.RUnlock()
	return len(registry)
}

// Clear removes all registered plugin drivers.
// Intended for testing only.
func Clear() {
	mu.Lock()
	defer mu.Unlock()
	registry = make(map[runtime.Backend]runtime.Runtime)
}

// String returns a human-readable summary for debugging.
func String() string {
	mu.RLock()
	defer mu.RUnlock()
	return fmt.Sprintf("plugin registry: %d drivers registered", len(registry))
}
