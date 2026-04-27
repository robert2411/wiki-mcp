package wiki

import (
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"github.com/robert2411/wiki-mcp/internal/config"
)

// WikiSearchResult is one item returned by wiki_search.
type WikiSearchResult struct {
	Path     string   `json:"path"`
	Title    string   `json:"title"`
	Snippets []string `json:"snippets"`
	Score    int      `json:"score"`
}

const (
	maxSnippets   = 3
	snippetRadius = 80 // characters of context around each match
)

// WikiSearch performs a case-insensitive substring (or regex) scan across page
// bodies (frontmatter is excluded). Returns up to limit results ordered by
// descending score (match count).
func WikiSearch(cfg *config.Config, query string, limit int) ([]WikiSearchResult, *ToolError) {
	if query == "" {
		return nil, NewToolError(ErrCodeBadRequest, "query must not be empty")
	}
	if limit <= 0 {
		limit = 20
	}

	re, err := buildSearchRegex(query)
	if err != nil {
		return nil, NewToolError(ErrCodeBadRequest, "invalid query: "+err.Error())
	}

	root := cfg.Root()
	var results []WikiSearchResult

	_ = filepath.WalkDir(root, func(p string, d os.DirEntry, walkErr error) error {
		if walkErr != nil || d.IsDir() || !strings.HasSuffix(p, ".md") {
			return nil
		}

		data, err := os.ReadFile(p)
		if err != nil {
			return nil
		}

		_, body := ParseFrontmatter(data)
		matches := re.FindAllStringIndex(body, -1)
		if len(matches) == 0 {
			return nil
		}

		rel, _ := filepath.Rel(root, p)
		snippets := extractSnippets(body, matches, maxSnippets)

		results = append(results, WikiSearchResult{
			Path:     rel,
			Title:    TitleFromPath(rel),
			Snippets: snippets,
			Score:    len(matches),
		})
		return nil
	})

	sort.Slice(results, func(i, j int) bool {
		return results[i].Score > results[j].Score
	})

	if len(results) > limit {
		results = results[:limit]
	}
	return results, nil
}

// buildSearchRegex compiles a case-insensitive regex for the query.
// If the query contains no regex metacharacters it is treated as a literal
// substring; otherwise it is compiled as a regex.
func buildSearchRegex(query string) (*regexp.Regexp, error) {
	pattern := query
	if !containsRegexMeta(query) {
		pattern = regexp.QuoteMeta(query)
	}
	return regexp.Compile("(?i)" + pattern)
}

func containsRegexMeta(s string) bool {
	return strings.ContainsAny(s, `\.+*?()|[]{}^$`)
}

// extractSnippets returns up to max context excerpts around match positions.
func extractSnippets(body string, matches [][]int, max int) []string {
	var snippets []string
	lastEnd := 0

	for _, m := range matches {
		if len(snippets) >= max {
			break
		}
		start := m[0] - snippetRadius
		if start < 0 {
			start = 0
		}
		end := m[1] + snippetRadius
		if end > len(body) {
			end = len(body)
		}

		// Snap to word boundaries (avoid cutting mid-word)
		for start > 0 && body[start] != ' ' && body[start] != '\n' {
			start--
		}
		for end < len(body) && body[end] != ' ' && body[end] != '\n' {
			end++
		}

		if start < lastEnd {
			continue // overlaps previous snippet
		}
		lastEnd = end

		snippet := strings.TrimSpace(body[start:end])
		snippet = strings.ReplaceAll(snippet, "\n", " ")
		snippets = append(snippets, snippet)
	}
	return snippets
}
