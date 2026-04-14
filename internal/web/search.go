package web

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/robertstevens/wiki-mcp/internal/wiki"
)

// SearchIndexEntry is one page entry in the pre-built search index.
type SearchIndexEntry struct {
	Path    string `json:"path"`
	Title   string `json:"title"`
	Snippet string `json:"snippet"`
}

// BuildSearchIndex walks wikiPath and returns an index entry per .md file.
// Each entry contains the page path, title (H1 or filename), and first 500
// characters of body text.
func BuildSearchIndex(wikiPath string) ([]SearchIndexEntry, error) {
	var entries []SearchIndexEntry
	err := filepath.WalkDir(wikiPath, func(abs string, d os.DirEntry, err error) error {
		if err != nil || d.IsDir() || !strings.HasSuffix(abs, ".md") {
			return nil
		}
		data, readErr := os.ReadFile(abs)
		if readErr != nil {
			return nil
		}
		rel, _ := filepath.Rel(wikiPath, abs)
		urlPath := strings.TrimSuffix(rel, ".md")

		fm, body := wiki.ParseFrontmatter(data)
		entries = append(entries, SearchIndexEntry{
			Path:    urlPath,
			Title:   extractTitle(fm, body, rel),
			Snippet: bodySnippet(body, 500),
		})
		return nil
	})
	return entries, err
}

// Search filters a pre-built index for pages matching query (case-insensitive substring).
func Search(entries []SearchIndexEntry, query string) []SearchIndexEntry {
	if query == "" {
		return nil
	}
	lower := strings.ToLower(query)
	var results []SearchIndexEntry
	for _, e := range entries {
		if strings.Contains(strings.ToLower(e.Title), lower) ||
			strings.Contains(strings.ToLower(e.Snippet), lower) {
			results = append(results, e)
		}
	}
	return results
}

func extractTitle(fm map[string]any, body, rel string) string {
	if s, _ := fm["title"].(string); s != "" {
		return s
	}
	if t := h1Title(body); t != "" {
		return t
	}
	return wiki.TitleFromPath(rel)
}

func h1Title(body string) string {
	for _, line := range strings.SplitN(body, "\n", 20) {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "# ") {
			return strings.TrimSpace(line[2:])
		}
	}
	return ""
}

func bodySnippet(body string, maxLen int) string {
	body = strings.TrimSpace(body)
	if len(body) <= maxLen {
		return body
	}
	return body[:maxLen] + "…"
}

// matchSnippet returns text around the first occurrence of query in body.
func matchSnippet(body, lowerQuery string, window int) string {
	lowerBody := strings.ToLower(body)
	idx := strings.Index(lowerBody, lowerQuery)
	if idx < 0 {
		return bodySnippet(body, window)
	}
	start := idx - window/2
	if start < 0 {
		start = 0
	}
	end := start + window
	if end > len(body) {
		end = len(body)
	}
	snip := body[start:end]
	if start > 0 {
		snip = "…" + snip
	}
	if end < len(body) {
		snip += "…"
	}
	return snip
}
