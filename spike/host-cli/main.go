// Spike A1: verify that Go exec.Command can invoke AI coding CLIs non-interactively.
// This is a throwaway spike, not production code.
package main

import (
	"bytes"
	"context"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"
)

type cliSpec struct {
	name    string
	binary  string
	args    []string
	timeout int
}

func main() {
	cli := flag.String("cli", "codex", "which CLI to test: codex|opencode")
	prompt := flag.String("prompt", "say hello in one word", "prompt to send")
	flag.Parse()

	specs := map[string]cliSpec{
		"codex": {
			name:    "codex",
			binary:  "codex",
			args:    []string{"exec", "--skip-git-repo-check", *prompt},
			timeout: 30,
		},
		"opencode": {
			name:    "opencode",
			binary:  "opencode",
			args:    []string{"run", *prompt},
			timeout: 30,
		},
	}

	spec, ok := specs[*cli]
	if !ok {
		fmt.Fprintf(os.Stderr, "unknown CLI: %s\n", *cli)
		os.Exit(1)
	}

	runSpike(spec)
}

func runSpike(spec cliSpec) {
	fmt.Printf("=== Spike: %s ===\n", spec.name)
	fmt.Printf("binary: %s\n", spec.binary)
	fmt.Printf("args:   %v\n", spec.args)

	// Check binary exists
	path, err := exec.LookPath(spec.binary)
	if err != nil {
		fmt.Printf("[FAIL] binary not found in PATH: %s\n", spec.binary)
		return
	}
	fmt.Printf("found:  %s\n", path)

	// Run with timeout
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(spec.timeout)*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, spec.binary, spec.args...)
	cmd.Dir = os.TempDir()

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	start := time.Now()
	err = cmd.Run()
	elapsed := time.Since(start)

	fmt.Printf("elapsed: %v\n", elapsed)
	fmt.Printf("exit-err: %v\n", err)

	out := stdout.String()
	if len(out) > 500 {
		out = out[:500] + "...(truncated)"
	}
	fmt.Printf("stdout: %s\n", out)

	errStr := stderr.String()
	if len(errStr) > 500 {
		errStr = errStr[:500] + "...(truncated)"
	}
	if errStr != "" {
		fmt.Printf("stderr: %s\n", errStr)
	}

	// Verdict
	if ctx.Err() == context.DeadlineExceeded {
		fmt.Printf("[VERDICT] TIMEOUT after %ds - CLI may not support clean non-interactive exit\n", spec.timeout)
	} else if err != nil && !strings.Contains(err.Error(), "exit status") {
		fmt.Printf("[VERDICT] FAIL - failed to start: %v\n", err)
	} else {
		fmt.Printf("[VERDICT] OK - CLI started, ran, and exited (non-interactive mode works)\n")
	}
}
