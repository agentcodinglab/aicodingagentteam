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
	"os"
	"path/filepath"
	"strings"

	"github.com/agentcodinglab/aicodingagentteam/internal/godocgen"
)

func main() {
	var (
		pkgDir = flag.String("pkg", "./pkg", "path to pkg/ directory")
		outDir = flag.String("out", "./website/content/docs/api", "path to docs/api/ output directory")
		locale = flag.String("locale", "en", "comma-separated locale codes")
		ver    = flag.String("version", "dev", "version label written into index.md")
	)
	flag.Parse()

	pkgAbs, err := filepath.Abs(*pkgDir)
	if err != nil {
		fail("abs pkg: %v", err)
	}
	outAbs, err := filepath.Abs(*outDir)
	if err != nil {
		fail("abs out: %v", err)
	}
	for _, code := range strings.Split(*locale, ",") {
		code = strings.TrimSpace(code)
		if code == "" {
			continue
		}
		fmt.Printf("[godocgen] locale=%s out=%s\n", code, outAbs)
		opts := godocgen.Options{
			PkgDir:  pkgAbs,
			OutDir:  outAbs,
			Locale:  code,
			Version: *ver,
		}
		if err := godocgen.Generate(opts); err != nil {
			fail("generate %s: %v", code, err)
		}
	}
	fmt.Println("[godocgen] OK")
}

func fail(format string, args ...any) {
	fmt.Fprintf(os.Stderr, "[godocgen] "+format+"\n", args...)
	os.Exit(1)
}

