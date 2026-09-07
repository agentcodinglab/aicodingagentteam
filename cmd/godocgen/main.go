// Command godocgen renders ./pkg/... into Markdown for the docs site.
//
// Usage:
//   godocgen [--pkg=./pkg] [--out=./website/content/docs/api]
//            [--locale=en,zh,ja,ko,fr,de,ru,es,it] [--version=vX.Y.Z]
//
// ADR-0016: docs/adr/ADR-0016-direction-a-governance.md
package main

import (
	"flag"
	"fmt"
	"github.com/agentcodinglab/aicodingagentteam/internal/godocgen"
	"os"
	"path/filepath"
	"strings"
)

func main() {
	var (
		pkgDir = flag.String("pkg", "./pkg", "path to pkg/ directory")
		outDir = flag.String("out", "./website/content/docs/api", "path to docs/api/ output directory")
		locale = flag.String("locale", "en", "comma-separated locale codes")
		ver    = flag.String("version", "dev", "version label written into index.md")
	)
	flag.Parse()
	if err := runGodocGen(*pkgDir, *outDir, *locale, *ver); err != nil {
		fail("%v", err)
	}
}

func runGodocGen(pkgDir, outDir, locale, ver string) error {
	pkgAbs, err := filepath.Abs(pkgDir)
	if err != nil {
		return fmt.Errorf("abs pkg: %w", err)
	}
	outAbs, err := filepath.Abs(outDir)
	if err != nil {
		return fmt.Errorf("abs out: %w", err)
	}
	for _, code := range strings.Split(locale, ",") {
		code = strings.TrimSpace(code)
		if code == "" {
			continue
		}
		fmt.Printf("[godocgen] locale=%s out=%s\n", code, outAbs)
		opts := godocgen.Options{
			PkgDir:  pkgAbs,
			OutDir:  outAbs,
			Locale:  code,
			Version: ver,
		}
		if err := godocgen.Generate(opts); err != nil {
			return fmt.Errorf("generate %s: %w", code, err)
		}
	}
	fmt.Println("[godocgen] OK")
	return nil
}

func fail(format string, args ...any) {
	fmt.Fprintf(os.Stderr, "[godocgen] "+format+"\n", args...)
	os.Exit(1)
}
