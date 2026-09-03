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
	// Must have BOTH AICODINGAGENTTEAM_ALLOW_CLOUD_EMBED=1 AND an embedding key
	allow := os.Getenv("AICODINGAGENTTEAM_ALLOW_CLOUD_EMBED")
	key := os.Getenv("OPENAI_EMBED_KEY")
	return allow == "1" && key != ""
}

// IndexDirectory indexes all code and doc files in the given directory tree.
func (e *Engine) IndexDirectory(ctx context.Context, root string) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	// If cloud embed is on but env vars not set, degrade to BM25 silently
	if e.cloudEmbed && !isCloudEmbedAllowed() {
		e.cloudEmbed = false // graceful degradation
	}

	return filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil // skip unreadable
		}
		if info.IsDir() {
			return nil
		}
		// Index code files and markdown docs
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
		return nil
	})
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

	// BM25 is always available; cloud vector is optional and degrades silently
	results := e.bm25.Search(query, topK)
	if results == nil {
		results = []Chunk{} // non-nil empty slice per spec
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
