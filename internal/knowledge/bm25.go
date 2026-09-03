// Package knowledge implements a pure-Go BM25 search engine with CJK bigram support.
// Zero external dependencies. Used as the guaranteed fallback for knowledge retrieval.
package knowledge

import (
	"strings"
	"unicode"
)

// Document is an indexed document with a path and content.
type Document struct {
	Path    string
	Content string
	tokens  []string
}

// Tokenizer splits text into tokens.
type Tokenizer struct{}

// Tokenize splits text into lowercase tokens.
func (t Tokenizer) Tokenize(text string) []string {
	var tokens []string
	var current strings.Builder

	for _, r := range text {
		if unicode.Is(unicode.Han, r) || unicode.Is(unicode.Hangul, r) || unicode.Is(unicode.Hiragana, r) || unicode.Is(unicode.Katakana, r) {
			if current.Len() > 0 {
				tokens = append(tokens, strings.ToLower(current.String()))
				current.Reset()
			}
			tokens = append(tokens, string(r))
			continue
		}

		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			current.WriteRune(unicode.ToLower(r))
		} else if current.Len() > 0 {
			tokens = append(tokens, strings.ToLower(current.String()))
			current.Reset()
		}
	}

	if current.Len() > 0 {
		tokens = append(tokens, strings.ToLower(current.String()))
	}

	return tokens
}

// BM25Index is the inverted index for BM25 scoring.
type BM25Index struct {
	docs    []Document
	docFreq map[string]int
	posting map[string][]int
	avgLen  float64
	k1      float64
	b       float64
}

// NewBM25Index creates an empty index.
func NewBM25Index() *BM25Index {
	return &BM25Index{
		docFreq: make(map[string]int),
		posting: make(map[string][]int),
		k1:      1.5,
		b:       0.75,
	}
}

// Add indexes a document.
func (idx *BM25Index) Add(path, content string) {
	tok := Tokenizer{}
	tokens := tok.Tokenize(content)
	docIdx := len(idx.docs)
	idx.docs = append(idx.docs, Document{Path: path, Content: content, tokens: tokens})

	seen := make(map[string]bool)
	for _, t := range tokens {
		if !seen[t] {
			seen[t] = true
			idx.docFreq[t]++
		}
		idx.posting[t] = append(idx.posting[t], docIdx)
	}

	total := 0
	for _, d := range idx.docs {
		total += len(d.tokens)
	}
	idx.avgLen = float64(total) / float64(len(idx.docs))
}

// Search returns documents matching the query, scored by BM25.
func (idx *BM25Index) Search(query string, topK int) []Chunk {
	if len(idx.docs) == 0 {
		return nil
	}

	tok := Tokenizer{}
	queryTokens := tok.Tokenize(query)
	if len(queryTokens) == 0 {
		return nil
	}

	N := float64(len(idx.docs))
	scores := make(map[int]float64)

	for _, qt := range queryTokens {
		docs, ok := idx.posting[qt]
		if !ok {
			continue
		}
		df := float64(idx.docFreq[qt])
		idf := max(0, ln((N-df+0.5)/(df+0.5)+1))

		for _, docIdx := range docs {
			d := &idx.docs[docIdx]
			tf := 0.0
			for _, t := range d.tokens {
				if t == qt {
					tf++
				}
			}
			if tf == 0 {
				continue
			}
			docLen := float64(len(d.tokens))
			numerator := tf * (idx.k1 + 1)
			denominator := tf + idx.k1*(1-idx.b+idx.b*docLen/idx.avgLen)
			scores[docIdx] += idf * numerator / denominator
		}
	}

	var sorted []scoredDoc
	for i, s := range scores {
		sorted = append(sorted, scoredDoc{i, s})
	}
	sortScoredDocs(sorted)

	if topK <= 0 || topK > len(sorted) {
		topK = len(sorted)
	}

	results := make([]Chunk, 0, topK)
	for _, sd := range sorted[:topK] {
		results = append(results, Chunk{
			Path:    idx.docs[sd.idx].Path,
			Content: idx.docs[sd.idx].Content,
			Score:   sd.score,
		})
	}
	return results
}

// DocCount returns the number of indexed documents.
func (idx *BM25Index) DocCount() int {
	return len(idx.docs)
}

type scoredDoc struct {
	idx   int
	score float64
}

func sortScoredDocs(docs []scoredDoc) {
	for i := 1; i < len(docs); i++ {
		for j := i; j > 0 && docs[j].score > docs[j-1].score; j-- {
			docs[j], docs[j-1] = docs[j-1], docs[j]
		}
	}
}

func ln(x float64) float64 {
	if x <= 0 {
		return 0
	}
	return logApprox(x)
}

func logApprox(x float64) float64 {
	if x == 1.0 {
		return 0
	}
	y := x - 1.0
	if y > -0.5 && y < 0.5 {
		term := y
		sum := y
		for i := 2; i < 20; i++ {
			term *= -y
			sum += term / float64(i)
		}
		return sum
	}
	if x > 0 {
		sqrt := x
		for i := 0; i < 20; i++ {
			sqrt = (sqrt + x/sqrt) / 2
		}
		return 2 * logApprox(sqrt)
	}
	return 0
}

func max(a, b float64) float64 {
	if a > b {
		return a
	}
	return b
}
