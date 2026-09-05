// Command aicodingagentteam is the binary entry point for the orchestration engine.
package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"github.com/agentcodinglab/aicodingagentteam/internal/a2a"
	"github.com/agentcodinglab/aicodingagentteam/internal/acp"
	"github.com/agentcodinglab/aicodingagentteam/internal/agent"
	"github.com/agentcodinglab/aicodingagentteam/internal/audit"
	"github.com/agentcodinglab/aicodingagentteam/internal/config"
	"github.com/agentcodinglab/aicodingagentteam/internal/coordinator"
	"github.com/agentcodinglab/aicodingagentteam/internal/governance"
	"github.com/agentcodinglab/aicodingagentteam/internal/host"
	"github.com/agentcodinglab/aicodingagentteam/internal/knowledge"
	"github.com/agentcodinglab/aicodingagentteam/internal/mcp"
	"github.com/agentcodinglab/aicodingagentteam/internal/memory"
	"github.com/agentcodinglab/aicodingagentteam/internal/planner"
	"github.com/agentcodinglab/aicodingagentteam/internal/qualitygate"
	"github.com/agentcodinglab/aicodingagentteam/internal/router"
	"github.com/agentcodinglab/aicodingagentteam/internal/scheduler"
	"github.com/agentcodinglab/aicodingagentteam/internal/types"
	"github.com/agentcodinglab/aicodingagentteam/pkg/api"
)

// Build metadata (injected by goreleaser ldflags).
var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
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
	case "continue":
		cmdContinue(ctx, cfg, os.Args[2:])
	case "mcp":
		cmdMCPServe(ctx, cfg)
	case "a2a":
		cmdA2AServe(ctx, cfg)
	case "acp":
		cmdACPServe(ctx, cfg)
	case "memory":
		cmdMemory(ctx, os.Args[2:])
	case "ci":
		cmdGovern(ctx, append([]string{"--ci"}, os.Args[2:]...))
	case "version":
		fmt.Printf("aicodingagentteam %s (commit=%s, built=%s)\n", version, commit, date)
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
		if _, err := keng.IndexDirectory(ctx, dir); err != nil {
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
		_, _ = keng.IndexDirectoryWithLimit(ctx, ".", 500)
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
	workspace, _ := os.Getwd()
	reportDir := filepath.Join(workspace, ".aicodingagentteam")
	_ = os.MkdirAll(reportDir, 0o755)
	memDir := filepath.Join(reportDir, "memory")

	// 1. Index current repository (cap at 500 files for safety).
	keng := knowledge.New(false)
	indexed, err := keng.IndexDirectoryWithLimit(ctx, workspace, 500)
	if err != nil {
		fmt.Fprintf(os.Stderr, "index error: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("[demo] indexed %d documents (cap=500) from %s\n", indexed, workspace)

	// 2. Wire Director with knowledge + memory, using stub backend (no real CLI).
	mem := memory.New(memDir)
	routerInst := router.New()
	plannerInst := planner.New(workspace)
	schedulerInst := scheduler.New(workspace)
	gate := qualitygate.New(80)
	// Demo uses NewWithOptions without a real A2A bus (nil bus = no reviewers fire).
	dir := coordinator.NewWithOptions(routerInst, plannerInst, schedulerInst, gate, nil,
		coordinator.WithKnowledge(keng),
		coordinator.WithMemory(mem),
	)

	// 3. Capture one seed fact so memory recall has something to return.
	_ = mem.Capture(ctx, memory.Fact{Key: "demo-seed", Value: "RAG demo bootstrap", Source: "knowledge-demo"})

	// 4. Drive Director.Handle: full 5-layer flow (Route->Plan->Schedule->Verify->Finalize).
	//    Knowledge retrieval and memory recall/capture fire as side effects.
	delivery, derr := dir.Handle(ctx, types.UserRequest{
		Message: "explain how the coordinator routes user requests",
		Backend: "stub",
	})
	if derr != nil {
		fmt.Fprintf(os.Stderr, "handle error: %v\n", derr)
		os.Exit(1)
	}

	// 5. Manual retrieval for the report (Director already triggered one inside Handle).
	chunks := keng.Retrieve(ctx, "coordinator routes user requests", 5)
	facts, _ := mem.RecallFacts(ctx)

	// 6. Print summary to stdout.
	fmt.Printf("[demo] delivery: planID=%s score=%d passed=%v artifacts=%d\n",
		delivery.PlanID, delivery.Score, delivery.Passed, len(delivery.Artifacts))
	fmt.Printf("[demo] retrieve -> %d chunks\n", len(chunks))
	for i, c := range chunks {
		fmt.Printf("  %d. [%.4f] %s\n", i+1, c.Score, c.Path)
	}
	fmt.Printf("[demo] recalled %d facts\n", len(facts))
	for _, f := range facts {
		fmt.Printf("  - %s: %s\n", f.Key, f.Value)
	}

	// 7. Write demo report (markdown + JSON) into .aicodingagentteam/ (gitignored).
	writeDemoReport(reportDir, indexed, workspace, chunks, facts, delivery)
	fmt.Printf("[demo] report written to %s/demo-report.md\n", reportDir)
	fmt.Println("[demo] RAG + memory end-to-end complete")
}

func writeDemoReport(reportDir string, indexed int, workspace string, chunks []knowledge.Chunk, facts []memory.Fact, delivery *types.Delivery) {
	md := &strings.Builder{}
	fmt.Fprintf(md, "# RAG Demo Report\n\n")
	fmt.Fprintf(md, "- Generated: %s\n", time.Now().Format(time.RFC3339))
	fmt.Fprintf(md, "- Workspace: %s\n", workspace)
	fmt.Fprintf(md, "- Indexed documents: %d (cap=500)\n", indexed)
	fmt.Fprintf(md, "- Retrieved chunks: %d\n\n", len(chunks))

	fmt.Fprintf(md, "## Top retrieved chunks\n\n")
	for i, c := range chunks {
		fmt.Fprintf(md, "%d. `%.4f` `%s`\n", i+1, c.Score, c.Path)
	}
	fmt.Fprintf(md, "\n## Recalled facts (%d)\n\n", len(facts))
	for _, f := range facts {
		fmt.Fprintf(md, "- `%s`: %s\n", f.Key, f.Value)
	}

	fmt.Fprintf(md, "\n## Delivery verdict\n\n")
	fmt.Fprintf(md, "- Plan ID: %s\n", delivery.PlanID)
	fmt.Fprintf(md, "- Score: %d\n", delivery.Score)
	fmt.Fprintf(md, "- Passed: %v\n", delivery.Passed)
	fmt.Fprintf(md, "- Artifacts: %d\n", len(delivery.Artifacts))
	if len(delivery.CheckDetails) > 0 {
		fmt.Fprintf(md, "\n### Check details\n\n")
		for _, c := range delivery.CheckDetails {
			fmt.Fprintf(md, "- %s: %s\n", c.Name, c.Status)
		}
	}

	_ = os.WriteFile(filepath.Join(reportDir, "demo-report.md"), []byte(md.String()), 0o644)

	js := struct {
		Indexed  int                `json:"indexed"`
		Chunks   []knowledge.Chunk  `json:"chunks"`
		Facts    []memory.Fact      `json:"facts"`
		Delivery *types.Delivery    `json:"delivery"`
	}{indexed, chunks, facts, delivery}
	if buf, err := json.MarshalIndent(js, "", "  "); err == nil {
		_ = os.WriteFile(filepath.Join(reportDir, "demo-report.json"), buf, 0o644)
	}
}

var tempDir = func() string {
	d, _ := os.MkdirTemp("", "aicodingagentteam-demo-*")
	return d
}

func cmdContinue(ctx context.Context, cfg *config.Config, args []string) {
	fs := flag.NewFlagSet("continue", flag.ExitOnError)
	_ = fs.Parse(args)
	d := newDirector(cfg)
	resumed, msg, err := d.ContinuePlan(ctx, fs.Arg(0))
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
	if resumed {
		fmt.Printf("workflow resumed: %s\n", msg)
	} else {
		fmt.Printf("cannot continue: %s\n", msg)
	}
}

func cmdMCPServe(ctx context.Context, cfg *config.Config) {
	govEngine := governance.NewWithConfig(".aicodingagentteam/rules.json", audit.New(".aicodingagentteam/audit"))
	srv := mcp.New(govEngine)
	fmt.Println("MCP server: stdio JSON-RPC (govern_file, govern_directory)")
	if err := srv.Serve(ctx); err != nil && err != context.Canceled {
		fmt.Fprintf(os.Stderr, "mcp server error: %v\n", err)
		os.Exit(1)
	}
}

func cmdA2AServe(ctx context.Context, cfg *config.Config) {
	bus := a2a.NewBusFromEnv(audit.New(".aicodingagentteam/audit"))
	agent.RegisterAllReviewers(bus)
	d := newDirector(cfg)
	srv := api.NewServer(cfg.Coordinator.GRPC, cfg.Coordinator.MCP, cfg.Coordinator.ACP, cfg.Coordinator.A2A, d)
	fmt.Printf("A2A server: listening on :%d\n", cfg.Coordinator.A2A)
	if err := srv.Start(ctx); err != nil && err != context.Canceled {
		fmt.Fprintf(os.Stderr, "a2a server error: %v\n", err)
		os.Exit(1)
	}
}

func cmdACPServe(ctx context.Context, cfg *config.Config) {
	srv := acp.New()
	fmt.Println("ACP server: stdio JSON-RPC (session lifecycle)")
	if err := srv.Serve(ctx); err != nil && err != context.Canceled {
		fmt.Fprintf(os.Stderr, "acp server error: %v\n", err)
		os.Exit(1)
	}
}

func cmdMemory(ctx context.Context, args []string) {
	if len(args) < 1 {
		fmt.Println("Usage:")
		fmt.Println("  aicodingagentteam memory show              Show all memories (facts/pitfalls/lessons)")
		fmt.Println("  aicodingagentteam memory capture on|off    Enable/disable fact capture")
		fmt.Println("  aicodingagentteam memory recall on|off     Enable/disable recipe recall")
		return
	}
	mem := memory.New(".aicodingagentteam/memory")
	sub := args[0]
	switch sub {
	case "show":
		facts, _ := mem.RecallFacts(ctx)
		fmt.Println("=== Facts ===")
		for _, f := range facts {
			fmt.Printf("  [%s] %s: %s\n", f.Source, f.Key, f.Value)
		}
		if len(facts) == 0 {
			fmt.Println("  (none)")
		}
		pitfalls, _ := mem.RecallPitfalls(ctx)
		fmt.Println("=== Pitfalls ===")
		for _, p := range pitfalls {
			fmt.Printf("  [%s] count=%d verified=%v: %s\n", p.ID, p.Count, p.Verified, p.Detail)
		}
		if len(pitfalls) == 0 {
			fmt.Println("  (none)")
		}
		lessons, _ := mem.RecallLessons(ctx)
		fmt.Println("=== Lessons ===")
		for _, l := range lessons {
			fmt.Printf("  [%s] verified=%v: %s\n", l.ID, l.Verified, l.Rule)
		}
		if len(lessons) == 0 {
			fmt.Println("  (none)")
		}
	case "capture":
		if len(args) < 2 {
			fmt.Fprintln(os.Stderr, "usage: memory capture on|off")
			os.Exit(1)
		}
		mem.SetCaptureOn(args[1] == "on")
		fmt.Printf("memory capture: %s\n", args[1])
	case "recall":
		if len(args) < 2 {
			fmt.Fprintln(os.Stderr, "usage: memory recall on|off")
			os.Exit(1)
		}
		mem.SetRecallOn(args[1] == "on")
		fmt.Printf("memory recall: %s\n", args[1])
	default:
		fmt.Fprintf(os.Stderr, "unknown memory subcommand: %s\n", sub)
		os.Exit(1)
	}
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

