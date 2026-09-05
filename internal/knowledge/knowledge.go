// Package knowledge provides BM25 + vector hybrid retrieval (RRF + HyDE).
// MVP: pure BM25 with file-system indexing. Vector/cloud disabled by default.
package knowledge

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

// Chunk is a retrieved knowledge chunk.
type Chunk struct {
	Path    string
	Content string
	Score   float64
}

// Engine is the hybrid retrieval engine. BM25 is always available; vector is optional.
type Engine struct {
	mu         sync.RWMutex
	cloudEmbed bool
	bm25       *BM25Index
}

// New creates a knowledge Engine. cloudEmbed controls optional cloud vector embedding.
func New(cloudEmbed bool) *Engine {
	return &Engine{
		cloudEmbed: cloudEmbed,
		bm25:       NewBM25Index(),
	}
}

// isCloudEmbedAllowed checks if cloud embedding is explicitly enabled via double env vars.
func isCloudEmbedAllowed() bool {
	allow := os.Getenv("AICODINGAGENTTEAM_ALLOW_CLOUD_EMBED")
	key := os.Getenv("OPENAI_EMBED_KEY")
	return allow == "1" && key != ""
}

// IndexDirectory indexes all code and doc files in the given directory tree.
// Returns the number of files indexed. Respects ctx for cancellation.
// Equivalent to IndexDirectoryWithLimit(ctx, root, 0) (no cap).
func (e *Engine) IndexDirectory(ctx context.Context, root string) (int, error) {
	return e.IndexDirectoryWithLimit(ctx, root, 0)
}

// IndexDirectoryWithLimit is IndexDirectory with a maximum file count.
// maxFiles <= 0 means unlimited. Returns (indexed, error).
func (e *Engine) IndexDirectoryWithLimit(ctx context.Context, root string, maxFiles int) (int, error) {
	e.mu.Lock()
	defer e.mu.Unlock()

	if e.cloudEmbed && !isCloudEmbedAllowed() {
		e.cloudEmbed = false
	}

	indexed := 0
	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		if ctx.Err() != nil {
			return ctx.Err()
		}
		if info.IsDir() {
			return nil
		}
		if maxFiles > 0 && indexed >= maxFiles {
			return filepath.SkipDir
		}
		ext := strings.ToLower(filepath.Ext(path))
		if ext != ".go" && ext != ".ts" && ext != ".tsx" && ext != ".js" &&
			ext != ".md" && ext != ".py" && ext != ".java" {
			return nil
		}
		content, err := os.ReadFile(path)
		if err != nil {
			return nil
		}
		e.bm25.Add(path, string(content))
		indexed++
		return nil
	})
	return indexed, err
}

// IndexFile adds a single file to the index.
func (e *Engine) IndexFile(path, content string) {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.bm25.Add(path, content)
}

// Retrieve returns top-k relevant chunks for a query.
// BM25 is the guaranteed fallback; vector degrades gracefully if unavailable.
func (e *Engine) Retrieve(ctx context.Context, query string, topK int) []Chunk {
	e.mu.RLock()
	defer e.mu.RUnlock()

	results := e.bm25.Search(query, topK)
	if results == nil {
		results = []Chunk{}
	}
	return results
}

// DocCount returns the number of indexed documents.
func (e *Engine) DocCount() int {
	e.mu.RLock()
	defer e.mu.RUnlock()
	return e.bm25.DocCount()
}

// IsCloudEmbed reports whether cloud embedding is active.
func (e *Engine) IsCloudEmbed() bool {
	e.mu.RLock()
	defer e.mu.RUnlock()
	return e.cloudEmbed && isCloudEmbedAllowed()
}
