package search

import (
	"encoding/json"
	"os"
	"path/filepath"
)

// Document represents an indexed page for search.
type Document struct {
	Title   string `json:"title"`
	URL     string `json:"url"`
	Section string `json:"section"`
	Content string `json:"content"`
}

// Index holds the list of indexed documents.
type Index struct {
	Docs []Document `json:"docs"`
}

// NewIndex creates an empty search index.
func NewIndex() *Index {
	return &Index{Docs: make([]Document, 0)}
}

// Add appends a document to the search index.
func (idx *Index) Add(title, url, section, content string) {
	idx.Docs = append(idx.Docs, Document{
		Title:   title,
		URL:     url,
		Section: section,
		Content: content,
	})
}

// Save writes the search index to a JSON file.
func (idx *Index) Save(outPath string) error {
	if err := os.MkdirAll(filepath.Dir(outPath), 0755); err != nil {
		return err
	}

	data, err := json.Marshal(idx.Docs)
	if err != nil {
		return err
	}

	return os.WriteFile(outPath, data, 0644)
}
