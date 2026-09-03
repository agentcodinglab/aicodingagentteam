// Package repomap scans Go source files for symbol declarations (functions, types).
package knowledge

import (
	"bufio"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"sync"
)

// Symbol is a code symbol (function, type, method).
type Symbol struct {
	Name     string
	Type     string // "func" / "type" / "method"
	Path     string
	Line     int
	Receiver string // for methods
}

// Index is a symbol index over a directory tree.
type Index struct {
	mu      sync.RWMutex
	symbols []Symbol
}

// NewIndex creates an empty symbol index.
func NewIndex() *Index {
	return &Index{}
}

// ScanDirectory walks a directory tree and indexes all .go files.
func (idx *Index) ScanDirectory(root string) error {
	idx.mu.Lock()
	defer idx.mu.Unlock()
	idx.symbols = nil

	return filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() {
			return nil
		}
		if !strings.HasSuffix(path, ".go") {
			return nil
		}
		// Skip test files
		if strings.HasSuffix(path, "_test.go") {
			return nil
		}
		return idx.scanFile(path)
	})
}

var funcRe = regexp.MustCompile(`^func\s+(?:\([^)]*\)\s+)?([A-Za-z_][A-Za-z0-9_]*)\s*\(`)
var typeRe = regexp.MustCompile(`^type\s+([A-Za-z_][A-Za-z0-9_]*)\s+`)
var methodRe = regexp.MustCompile(`^func\s+\(([^)]*)\)\s+([A-Za-z_][A-Za-z0-9_]*)\s*\(`)

func (idx *Index) scanFile(path string) error {
	f, err := os.Open(path)
	if err != nil {
		return nil
	}
	defer func() { _ = f.Close() }()

	scanner := bufio.NewScanner(f)
	scanner.Buffer(make([]byte, 0, 1024*1024), 1024*1024)
	lineNum := 0
	for scanner.Scan() {
		lineNum++
		line := scanner.Text()
		trimmed := strings.TrimSpace(line)

		// Method: func (receiver) Name(
		if matches := methodRe.FindStringSubmatch(trimmed); len(matches) > 0 {
			idx.symbols = append(idx.symbols, Symbol{
				Name:     matches[2],
				Type:     "method",
				Path:     path,
				Line:     lineNum,
				Receiver: matches[1],
			})
			continue
		}

		// Function: func Name(
		if matches := funcRe.FindStringSubmatch(trimmed); len(matches) > 0 {
			idx.symbols = append(idx.symbols, Symbol{
				Name: matches[1],
				Type: "func",
				Path: path,
				Line: lineNum,
			})
			continue
		}

		// Type: type Name
		if matches := typeRe.FindStringSubmatch(trimmed); len(matches) > 0 {
			idx.symbols = append(idx.symbols, Symbol{
				Name: matches[1],
				Type: "type",
				Path: path,
				Line: lineNum,
			})
		}
	}
	return scanner.Err()
}

// Symbols returns all indexed symbols.
func (idx *Index) Symbols() []Symbol {
	idx.mu.RLock()
	defer idx.mu.RUnlock()
	out := make([]Symbol, len(idx.symbols))
	copy(out, idx.symbols)
	return out
}

// SymbolsByPath returns symbols for a specific file.
func (idx *Index) SymbolsByPath(path string) []Symbol {
	idx.mu.RLock()
	defer idx.mu.RUnlock()
	var out []Symbol
	for _, s := range idx.symbols {
		if s.Path == path {
			out = append(out, s)
		}
	}
	return out
}

// Count returns the total number of indexed symbols.
func (idx *Index) Count() int {
	idx.mu.RLock()
	defer idx.mu.RUnlock()
	return len(idx.symbols)
}

// Summary returns a formatted summary string of all symbols, grouped by file.
func (idx *Index) Summary() string {
	idx.mu.RLock()
	defer idx.mu.RUnlock()

	// Group by path
	byPath := make(map[string][]Symbol)
	for _, s := range idx.symbols {
		byPath[s.Path] = append(byPath[s.Path], s)
	}

	var paths []string
	for p := range byPath {
		paths = append(paths, p)
	}
	sort.Strings(paths)

	var buf strings.Builder
	for _, p := range paths {
		buf.WriteString(p)
		buf.WriteString(":\n")
		for _, s := range byPath[p] {
			buf.WriteString("  ")
			buf.WriteString(s.Type)
			buf.WriteString(" ")
			if s.Receiver != "" {
				buf.WriteString("(")
				buf.WriteString(s.Receiver)
				buf.WriteString(") ")
			}
			buf.WriteString(s.Name)
			buf.WriteString("\n")
		}
	}
	return buf.String()
}
