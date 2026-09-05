package godocgen

import (
	"go/ast"
	"go/doc"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestGenerate_HappyPath(t *testing.T) {
	t.Parallel()
	tmp := t.TempDir()
	pkgRoot := filepath.Join(tmp, "pkg")
	outRoot := filepath.Join(tmp, "out")
	mustMkdir(t, pkgRoot, "demo")
	mustWrite(t, filepath.Join(pkgRoot, "demo", "demo.go"), `// Package demo is a tiny package for testing godocgen.
package demo

// Greeter returns a friendly hello.
type Greeter struct {
	Name string
}

// Hello is exported and documented.
func (g Greeter) Hello() string { return "hi " + g.Name }

// MaxLen is exported.
const MaxLen = 42
`)
	opts := Options{PkgDir: pkgRoot, OutDir: outRoot, Locale: "en", Version: "test"}
	if err := Generate(opts); err != nil {
		t.Fatalf("Generate: %v", err)
	}
	idx, err := os.ReadFile(filepath.Join(outRoot, "en", "index.md"))
	if err != nil {
		t.Fatalf("index.md missing: %v", err)
	}
	if !strings.Contains(string(idx), "Go Public API") {
		t.Errorf("index missing main heading:\n%s", idx)
	}
	pkgMD, err := os.ReadFile(filepath.Join(outRoot, "en", "pkg-demo.md"))
	if err != nil {
		t.Fatalf("pkg-demo.md missing: %v", err)
	}
	for _, must := range []string{"Greeter", "Hello", "MaxLen", "Hello is exported"} {
		if !strings.Contains(string(pkgMD), must) {
			t.Errorf("pkg-demo.md missing %q:\n%s", must, pkgMD)
		}
	}
}

func TestGenerate_NonEnglishLocale(t *testing.T) {
	t.Parallel()
	tmp := t.TempDir()
	pkgRoot := filepath.Join(tmp, "pkg")
	outRoot := filepath.Join(tmp, "out")
	mustMkdir(t, pkgRoot, "alpha")
	mustWrite(t, filepath.Join(pkgRoot, "alpha", "alpha.go"), `// Package alpha has one type.
package alpha

// Tag is exported.
type Tag string
`)
	opts := Options{PkgDir: pkgRoot, OutDir: outRoot, Locale: "zh", Version: "test"}
	if err := Generate(opts); err != nil {
		t.Fatalf("Generate: %v", err)
	}
	idx, err := os.ReadFile(filepath.Join(outRoot, "zh", "index.md"))
	if err != nil {
		t.Fatalf("zh index missing: %v", err)
	}
	if !strings.Contains(string(idx), "English index") {
		t.Errorf("zh index should link to English version:\n%s", idx)
	}
}

func TestGenerate_EmptyPkg(t *testing.T) {
	t.Parallel()
	tmp := t.TempDir()
	pkgRoot := filepath.Join(tmp, "pkg")
	outRoot := filepath.Join(tmp, "out")
	opts := Options{PkgDir: pkgRoot, OutDir: outRoot, Locale: "en", Version: "test"}
	mustMkdir(t, pkgRoot)
	if err := Generate(opts); err != nil {
		t.Errorf("empty pkg should not error: %v", err)
	}
	if _, err := os.Stat(filepath.Join(outRoot, "en", "index.md")); err != nil {
		t.Errorf("index.md should still be created: %v", err)
	}
}

func TestPrintDecl_RoundTrip(t *testing.T) {
	t.Parallel()
	tmp := t.TempDir()
	pkgRoot := filepath.Join(tmp, "pkg")
	mustMkdir(t, pkgRoot, "rt")
	src := `package rt

// Add sums two ints.
func Add(a, b int) int { return a + b }
`
	mustWrite(t, filepath.Join(pkgRoot, "rt.go"), src)
	fset, astPkg := mustParse(t, pkgRoot)
	dpkg := mustDoc(t, astPkg)
	if len(dpkg.Funcs) == 0 {
		t.Fatal("expected Add func")
	}
	got := printDecl(fset, dpkg.Funcs[0].Decl)
	if !strings.Contains(got, "Add(a, b int)") {
		t.Errorf("printDecl lost content: %q", got)
	}
}

func mustMkdir(t *testing.T, parts ...string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Join(parts...), 0o755); err != nil {
		t.Fatal(err)
	}
}

func mustWrite(t *testing.T, p, body string) {
	t.Helper()
	if err := os.WriteFile(p, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
}

func mustParse(t *testing.T, pkgRoot string) (*token.FileSet, *ast.Package) {
	t.Helper()
	fset := token.NewFileSet()
	pkgs, err := parser.ParseDir(fset, pkgRoot, nil, parser.ParseComments)
	if err != nil {
		t.Fatal(err)
	}
	for _, p := range pkgs {
		return fset, p
	}
	t.Fatal("no package found")
	return nil, nil
}

func mustDoc(t *testing.T, pkg *ast.Package) *doc.Package {
	t.Helper()
	dpkg := doc.New(pkg, pkg.Name, doc.AllDecls)
	if dpkg == nil {
		t.Fatalf("doc.New nil for %s", pkg.Name)
	}
	return dpkg
}






