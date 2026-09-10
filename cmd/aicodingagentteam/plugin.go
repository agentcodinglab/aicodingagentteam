package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/agentcodinglab/aicodingagentteam/pkg/plugin"
)

// cmdPlugin handles the `plugin` subcommand: scaffold new drivers and manage the local marketplace index.
func cmdPlugin(args []string) {
	if len(args) == 0 {
		printPluginUsage()
		return
	}
	switch args[0] {
	case "new":
		cmdPluginNew(args[1:])
	case "search":
		cmdPluginSearch(args[1:])
	case "list":
		cmdPluginList(args[1:])
	default:
		printPluginUsage()
	}
}

// cmdPluginNew scaffolds a new community driver package in the specified output directory.
func cmdPluginNew(args []string) {
	if len(args) < 1 {
		fmt.Fprintln(os.Stderr, "Usage: aicodingagentteam plugin new <name> [backend]")
		os.Exit(1)
	}
	name := args[0]
	backend := name
	if len(args) >= 2 {
		backend = args[1]
	}
	dir := filepath.Join("pkg", "plugin", "examples", name)
	if err := scaffoldDriver(dir, name, backend); err != nil {
		fmt.Fprintf(os.Stderr, "scaffold: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("Created driver package: %s\n", dir)
	fmt.Printf("Backend: %s\n", backend)
	fmt.Printf("\nTo activate, add a blank import:\n\n")
	fmt.Printf("  import _ \"github.com/agentcodinglab/aicodingagentteam/pkg/plugin/examples/%s\"\n", name)
}

// scaffoldDriver writes a minimal driver.go to the given directory.
func scaffoldDriver(dir, name, backend string) error {
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("mkdir: %w", err)
	}
	pkgName := strings.ReplaceAll(name, "-", "")
	content := fmt.Sprintf(`package %s

import (
	"context"
	"fmt"

	"github.com/agentcodinglab/aicodingagentteam/pkg/plugin"
	"github.com/agentcodinglab/aicodingagentteam/pkg/runtime"
)

const Backend runtime.Backend = "%s"

type driver struct{}

func init() {
	plugin.Register(Backend, &driver{})
}

func (d *driver) StartSession(ctx context.Context, opts runtime.SessionOpts) (runtime.SessionID, error) {
	return runtime.SessionID("%s-session"), nil
}
func (d *driver) DestroySession(ctx context.Context, id runtime.SessionID) error { return nil }
func (d *driver) SendTask(ctx context.Context, id runtime.SessionID, task runtime.TaskPayload) (<-chan runtime.Event, error) {
	ch := make(chan runtime.Event, 1)
	go func() {
		defer close(ch)
		ch <- runtime.Event{Type: runtime.EventDone, Content: fmt.Sprintf("%s processed: %%s", task.Instruction)}
	}()
	return ch, nil
}
func (d *driver) Capabilities() runtime.HostCapabilities {
	return runtime.HostCapabilities{SessionResume: false, ToolCalls: false, WebSearch: false, WriteHook: false}
}
func (d *driver) ModelInfo() runtime.ModelInfo {
	return runtime.ModelInfo{ID: "%s-model", Provider: "community", Context: 8192}
}
func (d *driver) Pause(ctx context.Context, id runtime.SessionID) error  { return nil }
func (d *driver) Resume(ctx context.Context, id runtime.SessionID) error { return nil }
func (d *driver) AuthStatus(ctx context.Context, id runtime.SessionID) (runtime.AuthStatus, error) {
	return runtime.AuthStatus{Ready: true, Detail: "%s driver needs no auth"}, nil
}
`, pkgName, backend, backend, backend, backend, backend)
	return os.WriteFile(filepath.Join(dir, "driver.go"), []byte(content), 0o644)
}

func printPluginUsage() {
	fmt.Println(`aicodingagentteam plugin <subcommand>

Subcommands:
  new <name> [backend]    Scaffold a new community driver package
  search <query>          Search the local plugin marketplace index
  list                    List all plugins in the local index`)
}

// cmdPluginSearch searches the local marketplace index.
func cmdPluginSearch(args []string) {
	if len(args) < 1 {
		fmt.Fprintln(os.Stderr, "Usage: aicodingagentteam plugin search <query>")
		os.Exit(1)
	}
	cwd, _ := os.Getwd()
	m := plugin.NewMarketplace(cwd)
	results, err := m.Search(args[0])
	if err != nil {
		fmt.Fprintf(os.Stderr, "search: %v\n", err)
		os.Exit(1)
	}
	if len(results) == 0 {
		fmt.Println("No plugins found.")
		return
	}
	for _, p := range results {
		fmt.Printf("  %s (backend=%s, v%s) -- %s\n", p.Name, p.Backend, p.Version, p.Description)
	}
}

// cmdPluginList lists all plugins in the local index.
func cmdPluginList(args []string) {
	cwd, _ := os.Getwd()
	m := plugin.NewMarketplace(cwd)
	plugins, err := m.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "list: %v\n", err)
		os.Exit(1)
	}
	if len(plugins) == 0 {
		fmt.Println("No plugins registered. Use 'plugin new' to scaffold a driver.")
		return
	}
	for _, p := range plugins {
		fmt.Printf("  %s (backend=%s, v%s) -- %s\n", p.Name, p.Backend, p.Version, p.Description)
	}
}
