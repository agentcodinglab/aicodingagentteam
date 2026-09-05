// Package godocgen renders the Go public API surface (pkg/...) into
// Markdown files consumed by the docs site (website/content/docs/api/{locale}/).
//
// ADR-0016: docs/adr/ADR-0016-direction-a-governance.md
package godocgen

import (
	"fmt"
	"go/ast"
	"go/doc"
	"go/parser"
	"go/printer"
	"go/token"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"unicode"
)

// PackageInfo describes one rendered package.
type PackageInfo struct {
	Name        string
	ImportPath  string
	Doc         string
	Decls       []Decl
	NumExamples int
}

// Decl is one exported declaration rendered as Markdown.
type Decl struct {
	Kind string // "type" | "func" | "const" | "var"
	Name string
	Doc  string
	Sign string
}

// Options controls generation.
type Options struct {
	PkgDir  string // absolute path to ./pkg
	OutDir  string // absolute path to website/content/docs/api
	Locale  string // "en" or 8 others
	Version string
}

// Generate walks pkgDir and writes Markdown files into outDir.
func Generate(opts Options) error {
	entries, err := os.ReadDir(opts.PkgDir)
	if err != nil {
		return fmt.Errorf("read pkg dir: %w", err)
	}
	pkgs := make([]PackageInfo, 0, len(entries))
	fset := token.NewFileSet()
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		pkgPath := filepath.Join(opts.PkgDir, e.Name())
		info, err := collectPackage(fset, pkgPath, e.Name(), opts.PkgDir)
		if err != nil {
			fmt.Fprintf(os.Stderr, "godocgen: skip %s: %v\n", e.Name(), err)
			continue
		}
		pkgs = append(pkgs, info)
	}
	sort.Slice(pkgs, func(i, j int) bool { return pkgs[i].Name < pkgs[j].Name })

	localeDir := filepath.Join(opts.OutDir, opts.Locale)
	if err := os.MkdirAll(localeDir, 0o755); err != nil {
		return fmt.Errorf("mkdir locale dir: %w", err)
	}
	if err := writeIndex(localeDir, pkgs, opts); err != nil {
		return fmt.Errorf("write index: %w", err)
	}
	for _, p := range pkgs {
		if err := writePackage(localeDir, p); err != nil {
			return fmt.Errorf("write %s: %w", p.Name, err)
		}
	}
	return nil
}

func collectPackage(fset *token.FileSet, pkgPath, pkgName, pkgRoot string) (PackageInfo, error) {
	pkgs, err := parser.ParseDir(fset, pkgPath, nil, parser.ParseComments)
	if err != nil {
		return PackageInfo{}, err
	}
	astPkg, ok := pkgs[pkgName]
	if !ok {
		return PackageInfo{}, fmt.Errorf("package %s not found in %s", pkgName, pkgPath)
	}
	dpkg := doc.New(astPkg, pkgPath, doc.AllDecls)

	// Normalize ImportPath to be relative to pkgRoot.
	importPath := dpkg.ImportPath
	if rel, err := filepath.Rel(pkgRoot, importPath); err == nil && !strings.HasPrefix(rel, "..") {
		importPath = rel
	}

	info := PackageInfo{
		Name:        dpkg.Name,
		ImportPath:  importPath,
		Doc:         strings.TrimSpace(dpkg.Doc),
		NumExamples: len(dpkg.Examples),
	}
	for _, t := range dpkg.Types {
		if !ast.IsExported(t.Name) {
			continue
		}
		info.Decls = append(info.Decls, Decl{
			Kind: "type",
			Name: t.Name,
			Doc:  strings.TrimSpace(t.Doc),
			Sign: printDecl(fset, t.Decl),
		})
		// Methods hang off the type via t.Methods, not t.Funcs.
		for _, f := range t.Methods {
			if !ast.IsExported(f.Name) {
				continue
			}
			info.Decls = append(info.Decls, Decl{
				Kind: "func",
				Name: f.Name,
				Doc:  strings.TrimSpace(f.Doc),
				Sign: printDecl(fset, f.Decl),
			})
		}
		for _, c := range t.Consts {
			for _, n := range c.Names {
				if !ast.IsExported(n) {
					continue
				}
				info.Decls = append(info.Decls, Decl{
					Kind: "const",
					Name: n,
					Doc:  strings.TrimSpace(c.Doc),
					Sign: printDecl(fset, c.Decl),
				})
			}
		}
		for _, v := range t.Vars {
			for _, n := range v.Names {
				if !ast.IsExported(n) {
					continue
				}
				info.Decls = append(info.Decls, Decl{
					Kind: "var",
					Name: n,
					Doc:  strings.TrimSpace(v.Doc),
					Sign: printDecl(fset, v.Decl),
				})
			}
		}
	}
	for _, f := range dpkg.Funcs {
		if !ast.IsExported(f.Name) {
			continue
		}
		info.Decls = append(info.Decls, Decl{
			Kind: "func",
			Name: f.Name,
			Doc:  strings.TrimSpace(f.Doc),
			Sign: printDecl(fset, f.Decl),
		})
	}
	for _, c := range dpkg.Consts {
		for _, n := range c.Names {
			if !ast.IsExported(n) {
				continue
			}
			info.Decls = append(info.Decls, Decl{
				Kind: "const",
				Name: n,
				Doc:  strings.TrimSpace(c.Doc),
				Sign: printDecl(fset, c.Decl),
			})
		}
	}
	for _, v := range dpkg.Vars {
		for _, n := range v.Names {
			if !ast.IsExported(n) {
				continue
			}
			info.Decls = append(info.Decls, Decl{
				Kind: "var",
				Name: n,
				Doc:  strings.TrimSpace(v.Doc),
				Sign: printDecl(fset, v.Decl),
			})
		}
	}
	return info, nil
}

// printDecl renders an ast.Decl via go/printer (UseSpaces, no source-position
// noise). Single-line output for embedding in markdown code fences.
func printDecl(fset *token.FileSet, decl ast.Decl) string {
	var b strings.Builder
	cfg := printer.Config{Mode: printer.UseSpaces, Tabwidth: 4}
	if err := cfg.Fprint(&b, fset, decl); err != nil {
		return ""
	}
	return strings.TrimSpace(b.String())
}

func writeIndex(dir string, pkgs []PackageInfo, opts Options) error {
	var b strings.Builder
	fmt.Fprintf(&b, "---\ntitle: Go Public API (%s)\n---\n\n", strings.ToUpper(opts.Locale))
	if opts.Locale != "en" {
		fmt.Fprintf(&b, "> **Note**: Go API documentation is canonical in English only. Full content lives in the [English index](./index.md).\n\n")
	} else {
		fmt.Fprintf(&b, "> Auto-generated by `internal/godocgen` (ADR-0016). Do not edit by hand.\n\n")
	}
	fmt.Fprintf(&b, "# Go Public API\n\n")
	fmt.Fprintf(&b, "Generated at `%s` from `./pkg/...` (%d packages).\n\n", opts.Version, len(pkgs))
	for _, p := range pkgs {
		fmt.Fprintf(&b, "## [`%s`](pkg-%s.md)\n\n", p.ImportPath, p.Name)
		if p.Doc != "" {
			fmt.Fprintf(&b, "%s\n\n", p.Doc)
		}
		fmt.Fprintf(&b, "_%d exported declarations_\n\n", len(p.Decls))
	}
	return os.WriteFile(filepath.Join(dir, "index.md"), []byte(b.String()), 0o644)
}

func writePackage(dir string, p PackageInfo) error {
	var b strings.Builder
	fmt.Fprintf(&b, "---\ntitle: %s\n---\n\n", p.Name)
	fmt.Fprintf(&b, "> Auto-generated by `internal/godocgen`. Do not edit by hand.\n\n")
	fmt.Fprintf(&b, "# package %s\n\n", p.ImportPath)
	if p.Doc != "" {
		fmt.Fprintf(&b, "%s\n\n", p.Doc)
	}
	fmt.Fprintf(&b, "_%d exported declarations, %d examples._\n\n", len(p.Decls), p.NumExamples)
	for _, d := range p.Decls {
		fmt.Fprintf(&b, "## %s `%s`\n\n", titleCaser(d.Kind), d.Name)
		if d.Sign != "" {
			fmt.Fprintf(&b, "```go\n%s\n```\n\n", d.Sign)
		}
		if d.Doc != "" {
			fmt.Fprintf(&b, "%s\n\n", d.Doc)
		}
	}
	return os.WriteFile(filepath.Join(dir, "pkg-"+p.Name+".md"), []byte(b.String()), 0o644)
}

func titleCaser(s string) string {
	if s == "" {
		return s
	}
	runes := []rune(s)
	runes[0] = unicode.ToUpper(runes[0])
	return string(runes)
}
