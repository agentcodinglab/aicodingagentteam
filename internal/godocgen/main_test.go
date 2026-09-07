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







// --- Coverage gap tests (P7.3) ---

func TestGenerate_NonDirEntries(t *testing.T) {
	t.Parallel()
	tmp := t.TempDir()
	pkgRoot := filepath.Join(tmp, "pkg")
	outRoot := filepath.Join(tmp, "out")
	mustMkdir(t, pkgRoot, "realpkg")
	mustWrite(t, filepath.Join(pkgRoot, "realpkg", "real.go"), `package realpkg

// Real is exported.
type Real struct{}
`)
	// drop a stray file in pkgDir root — should be skipped by !e.IsDir()
	mustWrite(t, filepath.Join(pkgRoot, "README.md"), "not a package")
	opts := Options{PkgDir: pkgRoot, OutDir: outRoot, Locale: "en", Version: "test"}
	if err := Generate(opts); err != nil {
		t.Fatalf("Generate with stray file: %v", err)
	}
}

func TestGenerate_BadPackageDir(t *testing.T) {
	t.Parallel()
	opts := Options{PkgDir: "/nonexistent/path/xyz", OutDir: t.TempDir(), Locale: "en", Version: "test"}
	if err := Generate(opts); err == nil {
		t.Error("expected error for nonexistent pkg dir")
	}
}

func TestGenerate_MkdirAllFail(t *testing.T) {
	t.Parallel()
	tmp := t.TempDir()
	pkgRoot := filepath.Join(tmp, "pkg")
	mustMkdir(t, pkgRoot, "ok")
	mustWrite(t, filepath.Join(pkgRoot, "ok", "ok.go"), `package ok
type T struct{}
`)
	// point OutDir at a file to force MkdirAll failure
	mustWrite(t, filepath.Join(tmp, "blocker"), "x")
	opts := Options{PkgDir: pkgRoot, OutDir: filepath.Join(tmp, "blocker"), Locale: "en", Version: "test"}
	if err := Generate(opts); err == nil {
		t.Error("expected mkdir error")
	}
}

func TestCollectPackage_NameMismatch(t *testing.T) {
	t.Parallel()
	tmp := t.TempDir()
	pkgRoot := filepath.Join(tmp, "pkg")
	dirName := "mydir"
	mustMkdir(t, pkgRoot, dirName)
	// source declares package "different" but dir is "mydir" -> name mismatch
	mustWrite(t, filepath.Join(pkgRoot, dirName, "f.go"), `package different
type X struct{}
`)
	fset := token.NewFileSet()
	_, err := collectPackage(fset, filepath.Join(pkgRoot, dirName), dirName, pkgRoot)
	if err == nil {
		t.Error("expected error for package name mismatch")
	}
}

func TestCollectPackage_ParseError(t *testing.T) {
	t.Parallel()
	tmp := t.TempDir()
	pkgRoot := filepath.Join(tmp, "pkg")
	dirName := "badpkg"
	mustMkdir(t, pkgRoot, dirName)
	mustWrite(t, filepath.Join(pkgRoot, dirName, "broken.go"), `package badpkg
this is not valid go syntax !!!
`)
	fset := token.NewFileSet()
	_, err := collectPackage(fset, filepath.Join(pkgRoot, dirName), dirName, pkgRoot)
	if err == nil {
		t.Error("expected parse error for broken go source")
	}
}

func TestCollectPackage_VarsAndFuncs(t *testing.T) {
	t.Parallel()
	tmp := t.TempDir()
	pkgRoot := filepath.Join(tmp, "pkg")
	dirName := "mixedpkg"
	mustMkdir(t, pkgRoot, dirName)
	mustWrite(t, filepath.Join(pkgRoot, dirName, "mixed.go"), `package mixedpkg

// Holder has a var and methods.
type Holder struct {
	Val int
}

// Get returns Val.
func (h Holder) Get() int { return h.Val }

// Default is an exported var on the type group.
var Default = Holder{Val: 1}

// Pi is a package-level exported var.
var Pi = 3.14

// MaxQ is a package-level exported const.
const MaxQ = 100

// unexportedVar should not appear.
var unexportedVar = 1
`)
	fset := token.NewFileSet()
	info, err := collectPackage(fset, filepath.Join(pkgRoot, dirName), dirName, pkgRoot)
	if err != nil {
		t.Fatalf("collectPackage: %v", err)
	}
	found := make(map[string]string)
	for _, d := range info.Decls {
		found[d.Name] = d.Kind
	}
	if _, ok := found["Pi"]; !ok {
		t.Error("expected package-level var Pi in decls")
	}
	if _, ok := found["MaxQ"]; !ok {
		t.Error("expected package-level const MaxQ in decls")
	}
	if _, ok := found["Get"]; !ok {
		t.Error("expected method Get in decls")
	}
	if _, ok := found["unexportedVar"]; ok {
		t.Error("unexportedVar should not appear")
	}
}

func TestPrintDecl_NilSafe(t *testing.T) {
	t.Parallel()
	// passing nil decl should return empty string, not panic
	got := printDecl(token.NewFileSet(), nil)
	if got != "" {
		t.Errorf("expected empty for nil decl, got %q", got)
	}
}

func TestTitleCaser_Empty(t *testing.T) {
	t.Parallel()
	if got := titleCaser(""); got != "" {
		t.Errorf("expected empty, got %q", got)
	}
}

func TestTitleCaser_Lower(t *testing.T) {
	t.Parallel()
	if got := titleCaser("func"); got != "Func" {
		t.Errorf("expected Func, got %q", got)
	}
}

// --- Additional coverage tests (P7.3 continued) ---

func TestCollectPackage_PackageLevelFuncsAndTypeConsts(t *testing.T) {
	t.Parallel()
	tmp := t.TempDir()
	pkgRoot := filepath.Join(tmp, "pkg")
	dirName := "fullpkg"
	mustMkdir(t, pkgRoot, dirName)
	mustWrite(t, filepath.Join(pkgRoot, dirName, "full.go"), `package fullpkg

// Color is a type with associated consts.
type Color int

const (
	// Red is a color.
	Red Color = iota
	// Blue is a color.
	Blue
)

// Standalone is a package-level exported func.
func Standalone() string { return "standalone" }

// unexported should not appear.
func unexported() {}
`)
	fset := token.NewFileSet()
	info, err := collectPackage(fset, filepath.Join(pkgRoot, dirName), dirName, pkgRoot)
	if err != nil {
		t.Fatalf("collectPackage: %v", err)
	}
	found := make(map[string]string)
	for _, d := range info.Decls {
		found[d.Name] = d.Kind
	}
	if _, ok := found["Standalone"]; !ok {
		t.Error("expected package-level func Standalone in decls")
	}
	if _, ok := found["Red"]; !ok {
		t.Error("expected type-level const Red in decls")
	}
	if _, ok := found["Blue"]; !ok {
		t.Error("expected type-level const Blue in decls")
	}
	if _, ok := found["unexported"]; ok {
		t.Error("unexported func should not appear")
	}
}

func TestGenerate_SkipsBadPackageButContinues(t *testing.T) {
	t.Parallel()
	tmp := t.TempDir()
	pkgRoot := filepath.Join(tmp, "pkg")
	outRoot := filepath.Join(tmp, "out")
	mustMkdir(t, pkgRoot, "good")
	mustWrite(t, filepath.Join(pkgRoot, "good", "good.go"), `package good
type T struct{}
`)
	// "bad" dir has a package name mismatch — Generate should skip it and continue
	mustMkdir(t, pkgRoot, "bad")
	mustWrite(t, filepath.Join(pkgRoot, "bad", "f.go"), `package mismatched
type X struct{}
`)
	opts := Options{PkgDir: pkgRoot, OutDir: outRoot, Locale: "en", Version: "test"}
	if err := Generate(opts); err != nil {
		t.Fatalf("Generate should skip bad pkg and succeed: %v", err)
	}
	// good package should still be rendered
	if _, err := os.Stat(filepath.Join(outRoot, "en", "pkg-good.md")); err != nil {
		t.Errorf("pkg-good.md should exist: %v", err)
	}
}

func TestGenerate_MkdirLocaleFail(t *testing.T) {
	t.Parallel()
	tmp := t.TempDir()
	pkgRoot := filepath.Join(tmp, "pkg")
	mustMkdir(t, pkgRoot, "ok")
	mustWrite(t, filepath.Join(pkgRoot, "ok", "ok.go"), `package ok
type T struct{}
`)
	// Create a file, then use it as OutDir/locale path to force MkdirAll failure
	blocker := filepath.Join(tmp, "blocker")
	mustWrite(t, blocker, "x")
	opts := Options{PkgDir: pkgRoot, OutDir: blocker, Locale: "en", Version: "test"}
	if err := Generate(opts); err == nil {
		t.Error("expected mkdir error when locale dir cannot be created")
	}
}
