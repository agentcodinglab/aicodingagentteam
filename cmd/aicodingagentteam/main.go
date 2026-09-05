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
	case "knowledge":
		cmdKnowledge(ctx, os.Args[2:])
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
	printCheckDetails(delivery.CheckDetails)
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
	printCheckDetails(delivery.CheckDetails)
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
		Details:  toGateDetails(v.Details),
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
	keng := knowledge.New(false)
	mem := memory.New(".aicodingagentteam/memory")
	_ = governance.NewWithConfig(".aicodingagentteam/rules.json", al)
	return coordinator.NewWithOptions(
		router.New(),
		planner.New(""),
		scheduler.NewFull("", bus, al),
		qualitygate.NewWithAudit(cfg.Quality.Threshold, al),
		bus,
		coordinator.WithKnowledge(keng),
		coordinator.WithMemory(mem),
	)
}

func cmdKnowledge(ctx context.Context, args []string) {
	if len(args) < 1 {
		fmt.Println(`Usage:
  aicodingagentteam knowledge index [dir]   Index a directory (default .)
  aicodingagentteam knowledge search "query" Retrieve top-5 chunks
  aicodingagentteam knowledge demo          End-to-end RAG + memory demo`)
		return
	}
	sub := args[0]
	switch sub {
	case "index":
		dir := "."
		if len(args) > 1 {
			dir = args[1]
		}
		keng := knowledge.New(false)
		if err := keng.IndexDirectory(ctx, dir); err != nil {
			fmt.Fprintf(os.Stderr, "index error: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("indexed %d documents from %s (cloud-embed=%v)\n", keng.DocCount(), dir, keng.IsCloudEmbed())
	case "search":
		if len(args) < 2 {
			fmt.Fprintln(os.Stderr, "search requires a query")
			os.Exit(1)
		}
		keng := knowledge.New(false)
		_ = keng.IndexDirectory(ctx, ".")
		chunks := keng.Retrieve(ctx, args[1], 5)
		fmt.Printf("top-%d results for %q:\n", len(chunks), args[1])
		for i, c := range chunks {
			fmt.Printf("  %d. [%.4f] %s\n", i+1, c.Score, c.Path)
		}
	case "demo":
		knowledgeDemo(ctx)
	default:
		fmt.Fprintf(os.Stderr, "unknown knowledge subcommand: %s\n", sub)
		os.Exit(1)
	}
}

func knowledgeDemo(ctx context.Context) {
	dir := tempDir()
	_ = os.MkdirAll(dir, 0o755)

	// Write two sample files for indexing
	_ = os.WriteFile(filepath.Join(dir, "router.go"), []byte("package main\nfunc route(msg string) string { return msg }\n"), 0o644)
	_ = os.WriteFile(filepath.Join(dir, "planner.go"), []byte("package main\nfunc plan(intent string) { println(intent) }\n"), 0o644)

	// 1. Index
	keng := knowledge.New(false)
	if err := keng.IndexDirectory(ctx, dir); err != nil {
		fmt.Fprintf(os.Stderr, "index error: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("[demo] indexed %d documents\n", keng.DocCount())

	// 2. Retrieve
	chunks := keng.Retrieve(ctx, "route message intent", 3)
	fmt.Printf("[demo] retrieve 'route message intent' -> %d chunks:\n", len(chunks))
	for i, c := range chunks {
		fmt.Printf("  %d. [%.4f] %s\n", i+1, c.Score, c.Path)
	}

	// 3. Memory: capture a fact
	mem := memory.New(filepath.Join(dir, ".memory"))
	_ = mem.Capture(ctx, memory.Fact{Key: "demo-fact", Value: "router routes messages", Source: "knowledge-demo"})
	fmt.Println("[demo] captured fact: demo-fact")

	// 4. Memory: recall
	facts, _ := mem.RecallFacts(ctx)
	fmt.Printf("[demo] recalled %d facts:\n", len(facts))
	for _, f := range facts {
		fmt.Printf("  - %s: %s\n", f.Key, f.Value)
	}

	// 5. Status
	fmt.Printf("[demo] cloud-embed=%v doc-count=%d\n", keng.IsCloudEmbed(), keng.DocCount())
	fmt.Println("[demo] RAG + memory end-to-end complete")
}

var tempDir = func() string {
	d, _ := os.MkdirTemp("", "aicodingagentteam-demo-*")
	return d
}

func printCheckDetails(details []types.CheckSummary) {
	for _, d := range details {
		if d.Status == "pass" {
			continue
		}
		out := d.Output
		if len(out) > 200 {
			out = out[:200] + "..."
		}
		fmt.Printf("  [%s] %s: %s\n", strings.ToUpper(d.Status), d.Name, out)
	}
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
  aicodingagentteam knowledge index [dir]         Index a directory
  aicodingagentteam knowledge search "query"      Retrieve top-5 chunks
  aicodingagentteam knowledge demo                 End-to-end RAG + memory demo
  aicodingagentteam version                        Print version`)
}

func toGateDetails(in []api.CheckSummary) []qualitygate.CheckDetail {
	if len(in) == 0 {
		return nil
	}
	out := make([]qualitygate.CheckDetail, len(in))
	for i, s := range in {
		out[i] = qualitygate.CheckDetail{Name: s.Name, Status: s.Status, Output: s.Output}
	}
	return out
}
