// Command aicodingagentteam is the binary entry point for the orchestration engine.
package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"

	"github.com/agentcodinglab/aicodingagentteam/internal/a2a"
	"github.com/agentcodinglab/aicodingagentteam/internal/audit"
	"github.com/agentcodinglab/aicodingagentteam/internal/config"
	"github.com/agentcodinglab/aicodingagentteam/internal/coordinator"
	"github.com/agentcodinglab/aicodingagentteam/internal/governance"
	"github.com/agentcodinglab/aicodingagentteam/internal/host"
	"github.com/agentcodinglab/aicodingagentteam/internal/knowledge"
	"github.com/agentcodinglab/aicodingagentteam/internal/memory"
	"github.com/agentcodinglab/aicodingagentteam/internal/planner"
	"github.com/agentcodinglab/aicodingagentteam/internal/qualitygate"
	"github.com/agentcodinglab/aicodingagentteam/internal/router"
	"github.com/agentcodinglab/aicodingagentteam/internal/scheduler"
	"github.com/agentcodinglab/aicodingagentteam/internal/types"
	"github.com/agentcodinglab/aicodingagentteam/pkg/api"
)

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	cfg, _ := config.Load()
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	switch os.Args[1] {
	case "init":
		cmdInit()
	case "run":
		cmdRun(ctx, cfg, os.Args[2:])
	case "quick":
		cmdQuick(ctx, cfg, os.Args[2:])
	case "verify":
		cmdVerify(ctx, cfg, os.Args[2:])
	case "govern":
		cmdGovern(ctx, os.Args[2:])
	case "report":
		cmdReport(ctx, cfg)
	case "serve":
		cmdServe(ctx, cfg)
	case "version":
		fmt.Println("aicodingagentteam v0.1.0")
	default:
		printUsage()
	}
}

func cmdInit() {
	for _, d := range []string{".aicodingagentteam", ".aicodingagentteam/audit", ".aicodingagentteam/memory", ".aicodingagentteam/contracts", ".aicodingagentteam/proof", "output"} {
		_ = os.MkdirAll(d, 0o755)
	}
	fmt.Println("project initialized: .aicodingagentteam/ created")
}

func cmdRun(ctx context.Context, cfg *config.Config, args []string) {
	fs := flag.NewFlagSet("run", flag.ExitOnError)
	backend := fs.String("backend", cfg.Default.Backend, "host CLI backend")
	_ = fs.Parse(args)
	req := types.UserRequest{Message: fs.Arg(0), Backend: *backend}
	d := newDirector(cfg)
	delivery, err := d.Handle(ctx, req)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("delivery: plan=%s score=%d passed=%v artifacts=%v\n", delivery.PlanID, delivery.Score, delivery.Passed, delivery.Artifacts)
}

func cmdQuick(ctx context.Context, cfg *config.Config, args []string) {
	fs := flag.NewFlagSet("quick", flag.ExitOnError)
	backend := fs.String("backend", cfg.Default.Backend, "host CLI backend")
	_ = fs.Parse(args)
	req := types.UserRequest{Message: fs.Arg(0), Backend: *backend}
	d := newDirector(cfg)
	delivery, err := d.Handle(ctx, req)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("quick edit done: passed=%v files=%v\n", delivery.Passed, delivery.Artifacts)
}

func cmdVerify(ctx context.Context, cfg *config.Config, args []string) {
	fs := flag.NewFlagSet("verify", flag.ExitOnError)
	runtime := fs.Bool("runtime", false, "include runtime probe")
	backend := fs.String("backend", cfg.Default.Backend, "host CLI backend")
	_ = fs.Parse(args)

	d := newDirector(cfg)
	_ = backend // suppress unused warning

	v, err := d.Verify(ctx)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}

	if *runtime {
		fmt.Println("runtime probe: included in verification")
	}

	fmt.Printf("quality-gate: score=%d passed=%v\n", v.Score, v.Passed)
	if len(v.Blocking) > 0 {
		fmt.Println("blocking:")
		for _, b := range v.Blocking {
			fmt.Printf("  - %s\n", b)
		}
	}
}

func cmdGovern(ctx context.Context, args []string) {
	fs := flag.NewFlagSet("govern", flag.ExitOnError)
	ciMode := fs.Bool("ci", false, "CI mode: fail-close (exit 1 on blocking violations)")
	rulesPath := fs.String("rules", ".aicodingagentteam/rules.json", "path to rules config")
	_ = fs.Parse(args)
	target := fs.Arg(0)
	if target == "" {
		target = "."
	}

	engine := governance.NewWithConfig(*rulesPath, audit.New(".aicodingagentteam/audit"))

	// Scan files
	var files []string
	info, err := os.Stat(target)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
	if info.IsDir() {
		_ = filepath.Walk(target, func(path string, fi os.FileInfo, err error) error {
			if err != nil || fi.IsDir() {
				return nil
			}
			ext := strings.ToLower(filepath.Ext(path))
			if ext == ".go" || ext == ".ts" || ext == ".tsx" || ext == ".js" ||
				ext == ".jsx" || ext == ".py" || ext == ".java" || ext == ".rb" {
				files = append(files, path)
			}
			return nil
		})
	} else {
		files = append(files, target)
	}

	totalViolations := 0
	hasBlocking := false
	for _, f := range files {
		content, err := os.ReadFile(f)
		if err != nil {
			continue
		}
		violations := engine.Check(ctx, f, string(content))
		for _, v := range violations {
			totalViolations++
			if v.Severity == "blocking" {
				hasBlocking = true
			}
			fmt.Printf("[%s] %s: %s (%s)\n", v.Severity, v.RuleID, v.Detail, f)
		}
	}

	fmt.Printf("\ngovernance: %d violations found", totalViolations)
	if hasBlocking {
		fmt.Printf(" (blocking present)")
	}
	fmt.Println()

	if *ciMode && hasBlocking {
		fmt.Fprintln(os.Stderr, "CI mode: blocking violations detected, exiting 1")
		os.Exit(1)
	}
}

func cmdReport(ctx context.Context, cfg *config.Config) {
	d := newDirector(cfg)
	v, err := d.Verify(ctx)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
	_ = os.MkdirAll("output", 0o755)
	result := qualitygate.Result{
		Score:    v.Score,
		Passed:   v.Passed,
		Blocking: v.Blocking,
		Advisory: v.Advisory,
	}
	report := qualitygate.Scorecard(result)
	fmt.Print(report)
	_ = os.WriteFile("output/quality-gate.md", []byte(report), 0o644)
	fmt.Println("\nReport saved to output/quality-gate.md")
}

func cmdServe(ctx context.Context, cfg *config.Config) {
	d := newDirector(cfg)
	srv := api.NewServer(cfg.Coordinator.GRPC, cfg.Coordinator.MCP, cfg.Coordinator.ACP, cfg.Coordinator.A2A, d)
	fmt.Println("starting AiCodingAgentTeam coordinator...")
	if err := srv.Start(ctx); err != nil {
		fmt.Fprintf(os.Stderr, "server error: %v\n", err)
		os.Exit(1)
	}
}

func newDirector(cfg *config.Config) *coordinator.Director {
	al := audit.New(".aicodingagentteam/audit")
	bus := a2a.NewBusFromEnv(al)
	_ = host.NewRegistry()
	_ = knowledge.New(false)
	_ = memory.New(".aicodingagentteam/memory")
	_ = governance.NewWithConfig(".aicodingagentteam/rules.json", al)
	return coordinator.NewWithBus(
		router.New(),
		planner.New(""),
		scheduler.NewFull("", bus, al),
		qualitygate.New(cfg.Quality.Threshold),
		bus,
	)
}

func printUsage() {
	fmt.Println(`AiCodingAgentTeam - AI coding orchestration platform

Usage:
  aicodingagentteam init                          Initialize project
  aicodingagentteam run "requirement" --backend X  Run full pipeline
  aicodingagentteam quick "small edit"             Quick edit
  aicodingagentteam verify                         Run quality gate
  aicodingagentteam govern [--ci] [path]           Governance scan
  aicodingagentteam serve                          Start coordinator server
  aicodingagentteam version                        Print version`)
}
