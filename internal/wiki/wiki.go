// Package wiki implements page CRUD, index, log, and link-graph operations.
package wiki

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"gopkg.in/yaml.v3"

	"github.com/robertstevens/wiki-mcp/internal/config"
)

// Error codes returned in structured tool errors.
const (
	ErrCodeNotFound   = "wiki.NotFound"
	ErrCodeReadOnly   = "wiki.ReadOnly"
	ErrCodePathEscape = "wiki.PathEscape"
	ErrCodeForbidden  = "wiki.Forbidden"
	ErrCodeBadRequest = "wiki.BadRequest"
	ErrCodeInternal   = "wiki.Internal"
)

// ToolError is a structured error returned by wiki tools.
type ToolError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

func (e *ToolError) Error() string { return fmt.Sprintf("%s: %s", e.Code, e.Message) }

func (e *ToolError) JSON() string {
	b, _ := json.Marshal(e)
	return string(b)
}

// NewToolError creates a ToolError with the given code and message.
func NewToolError(code, msg string) *ToolError {
	return &ToolError{Code: code, Message: msg}
}

// PageReadResult is what page_read returns.
type PageReadResult struct {
	Path        string         `json:"path"`
	Frontmatter map[string]any `json:"frontmatter,omitempty"`
	Body        string         `json:"body"`
}

// PageListEntry is one item in page_list results.
type PageListEntry struct {
	Path    string         `json:"path"`
	Tags    []string       `json:"tags,omitempty"`
	Updated string         `json:"updated,omitempty"`
	Title   string         `json:"title,omitempty"`
}

// ParseFrontmatter splits a markdown file into YAML frontmatter and body.
func ParseFrontmatter(data []byte) (map[string]any, string) {
	s := string(data)
	if !strings.HasPrefix(s, "---\n") && !strings.HasPrefix(s, "---\r\n") {
		return nil, s
	}

	// Find closing ---
	rest := s[4:] // skip opening "---\n"

	// Handle empty frontmatter: ---\n---\n
	if strings.HasPrefix(rest, "---\n") {
		return nil, rest[4:]
	}
	if strings.HasPrefix(rest, "---\r\n") {
		return nil, rest[5:]
	}

	idx := strings.Index(rest, "\n---\n")
	if idx < 0 {
		idx = strings.Index(rest, "\r\n---\r\n")
		if idx < 0 {
			if strings.HasSuffix(rest, "\n---") {
				idx = len(rest) - 4
			} else {
				return nil, s
			}
		}
	}

	yamlBlock := rest[:idx]
	body := rest[idx:]
	// strip the closing --- line
	if i := strings.Index(body, "---"); i >= 0 {
		body = body[i+3:]
		if len(body) > 0 && body[0] == '\n' {
			body = body[1:]
		} else if len(body) > 1 && body[0] == '\r' && body[1] == '\n' {
			body = body[2:]
		}
	}

	var fm map[string]any
	if err := yaml.Unmarshal([]byte(yamlBlock), &fm); err != nil {
		return nil, s
	}
	return fm, body
}

// RenderFrontmatter serialises frontmatter + body. If fm is nil/empty, only body is written.
func RenderFrontmatter(fm map[string]any, body string) []byte {
	if len(fm) == 0 {
		return []byte(body)
	}
	var buf bytes.Buffer
	buf.WriteString("---\n")
	enc := yaml.NewEncoder(&buf)
	enc.SetIndent(2)
	_ = enc.Encode(fm)
	enc.Close()
	buf.WriteString("---\n")
	buf.WriteString(body)
	return buf.Bytes()
}

// PageRead reads a page by relative path.
func PageRead(cfg *config.Config, relPath string) (*PageReadResult, *ToolError) {
	abs, err := cfg.ResolveWikiPath(relPath)
	if err != nil {
		return nil, NewToolError(ErrCodePathEscape, err.Error())
	}

	data, err := os.ReadFile(abs)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, NewToolError(ErrCodeNotFound, fmt.Sprintf("page %q not found", relPath))
		}
		return nil, NewToolError(ErrCodeInternal, err.Error())
	}

	fm, body := ParseFrontmatter(data)
	return &PageReadResult{
		Path:        relPath,
		Frontmatter: fm,
		Body:        body,
	}, nil
}

// PageWrite creates or overwrites a page.
func PageWrite(cfg *config.Config, relPath string, fm map[string]any, body string) *ToolError {
	if err := cfg.MustMutate(); err != nil {
		return NewToolError(ErrCodeReadOnly, err.Error())
	}

	abs, err := cfg.ResolveWikiPath(relPath)
	if err != nil {
		return NewToolError(ErrCodePathEscape, err.Error())
	}

	content := RenderFrontmatter(fm, body)

	if cfg.Safety.MaxPageBytes > 0 && len(content) > cfg.Safety.MaxPageBytes {
		return NewToolError(ErrCodeBadRequest,
			fmt.Sprintf("page exceeds max size (%d > %d bytes)", len(content), cfg.Safety.MaxPageBytes))
	}

	if err := os.MkdirAll(filepath.Dir(abs), 0o755); err != nil {
		return NewToolError(ErrCodeInternal, fmt.Sprintf("cannot create directory: %v", err))
	}

	if err := os.WriteFile(abs, content, 0o644); err != nil {
		return NewToolError(ErrCodeInternal, err.Error())
	}
	return nil
}

// protectedBasenames are filenames that cannot be deleted.
var protectedBasenames = map[string]bool{
	"index.md": true,
	"log.md":   true,
}

// PageDelete removes a page. Rejects index.md and log.md.
func PageDelete(cfg *config.Config, relPath string) *ToolError {
	if err := cfg.MustMutate(); err != nil {
		return NewToolError(ErrCodeReadOnly, err.Error())
	}

	abs, err := cfg.ResolveWikiPath(relPath)
	if err != nil {
		return NewToolError(ErrCodePathEscape, err.Error())
	}

	base := filepath.Base(abs)
	if protectedBasenames[base] {
		return NewToolError(ErrCodeForbidden,
			fmt.Sprintf("cannot delete %q: it is a protected wiki file", base))
	}

	if err := os.Remove(abs); err != nil {
		if os.IsNotExist(err) {
			return NewToolError(ErrCodeNotFound, fmt.Sprintf("page %q not found", relPath))
		}
		return NewToolError(ErrCodeInternal, err.Error())
	}
	return nil
}

// PageListFilter holds optional filters for page_list.
type PageListFilter struct {
	Dir          string `json:"dir,omitempty"`
	Glob         string `json:"glob,omitempty"`
	Tag          string `json:"tag,omitempty"`
	UpdatedSince string `json:"updated_since,omitempty"`
}

// PageList lists wiki pages with optional filters.
func PageList(cfg *config.Config, filter PageListFilter) ([]PageListEntry, *ToolError) {
	var entries []PageListEntry
	var updatedSince time.Time
	if filter.UpdatedSince != "" {
		t, err := time.Parse("2006-01-02", filter.UpdatedSince)
		if err != nil {
			return nil, NewToolError(ErrCodeBadRequest,
				fmt.Sprintf("invalid updated_since date %q: use YYYY-MM-DD format", filter.UpdatedSince))
		}
		updatedSince = t
	}

	root := cfg.WikiPath
	err := filepath.WalkDir(root, func(abs string, d os.DirEntry, err error) error {
		if err != nil {
			return nil // skip unreadable
		}
		if d.IsDir() {
			return nil
		}
		if !strings.HasSuffix(abs, ".md") {
			return nil
		}

		rel, _ := filepath.Rel(root, abs)

		// Dir filter
		if filter.Dir != "" {
			prefix := filepath.Clean(filter.Dir)
			if !strings.HasPrefix(rel, prefix+string(os.PathSeparator)) && rel != prefix {
				return nil
			}
		}

		// Glob filter
		if filter.Glob != "" {
			matched, _ := filepath.Match(filter.Glob, rel)
			if !matched {
				// Also try matching just the filename
				matched, _ = filepath.Match(filter.Glob, filepath.Base(rel))
			}
			if !matched {
				return nil
			}
		}

		entry := PageListEntry{Path: rel}

		// Read frontmatter for tag/updated filters and metadata
		if filter.Tag != "" || !updatedSince.IsZero() {
			data, readErr := os.ReadFile(abs)
			if readErr != nil {
				return nil
			}
			fm, _ := ParseFrontmatter(data)
			if fm != nil {
				// Extract tags
				if tags, ok := fm["tags"]; ok {
					switch v := tags.(type) {
					case []any:
						for _, t := range v {
							if s, ok := t.(string); ok {
								entry.Tags = append(entry.Tags, s)
							}
						}
					case []string:
						entry.Tags = v
					}
				}
				// Extract updated
				if u, ok := fm["updated"]; ok {
					switch v := u.(type) {
					case string:
						entry.Updated = v
					case time.Time:
						entry.Updated = v.Format("2006-01-02")
					}
				}
			}

			// Tag filter
			if filter.Tag != "" {
				found := false
				for _, t := range entry.Tags {
					if t == filter.Tag {
						found = true
						break
					}
				}
				if !found {
					return nil
				}
			}

			// Updated since filter
			if !updatedSince.IsZero() && entry.Updated != "" {
				pageDate, parseErr := time.Parse("2006-01-02", entry.Updated)
				if parseErr != nil || pageDate.Before(updatedSince) {
					return nil
				}
			} else if !updatedSince.IsZero() && entry.Updated == "" {
				return nil // no date means we can't confirm it's recent enough
			}
		}

		entries = append(entries, entry)
		return nil
	})

	if err != nil {
		return nil, NewToolError(ErrCodeInternal, err.Error())
	}
	return entries, nil
}

var (
	// [[Title]] pattern
	wikiLinkRe = regexp.MustCompile(`\[\[([^\]]+)\]\]`)
	// [text](path) pattern — captures text and path separately
	mdLinkRe = regexp.MustCompile(`\[([^\]]*)\]\(([^)]+)\)`)
)

// titleFromPath derives a page title from a relative path.
// e.g. "entities/qwen2.5-coder.md" → "Qwen2.5 Coder"
func titleFromPath(rel string) string {
	base := strings.TrimSuffix(filepath.Base(rel), ".md")
	base = strings.ReplaceAll(base, "-", " ")
	// Title-case first letter of each word
	words := strings.Fields(base)
	for i, w := range words {
		if len(w) > 0 {
			words[i] = strings.ToUpper(w[:1]) + w[1:]
		}
	}
	return strings.Join(words, " ")
}

// PageMove renames/moves a page and rewrites links across the wiki.
func PageMove(cfg *config.Config, oldRel, newRel string) *ToolError {
	if err := cfg.MustMutate(); err != nil {
		return NewToolError(ErrCodeReadOnly, err.Error())
	}

	oldAbs, err := cfg.ResolveWikiPath(oldRel)
	if err != nil {
		return NewToolError(ErrCodePathEscape, err.Error())
	}
	newAbs, err := cfg.ResolveWikiPath(newRel)
	if err != nil {
		return NewToolError(ErrCodePathEscape, err.Error())
	}

	if _, err := os.Stat(oldAbs); os.IsNotExist(err) {
		return NewToolError(ErrCodeNotFound, fmt.Sprintf("page %q not found", oldRel))
	}

	// Create destination directory
	if err := os.MkdirAll(filepath.Dir(newAbs), 0o755); err != nil {
		return NewToolError(ErrCodeInternal, fmt.Sprintf("cannot create directory: %v", err))
	}

	// Move the file
	if err := os.Rename(oldAbs, newAbs); err != nil {
		return NewToolError(ErrCodeInternal, err.Error())
	}

	// Rewrite links in all other pages
	oldTitle := titleFromPath(oldRel)
	newTitle := titleFromPath(newRel)

	root := cfg.WikiPath
	_ = filepath.WalkDir(root, func(abs string, d os.DirEntry, walkErr error) error {
		if walkErr != nil || d.IsDir() || !strings.HasSuffix(abs, ".md") {
			return nil
		}

		data, readErr := os.ReadFile(abs)
		if readErr != nil {
			return nil
		}

		pageRel, _ := filepath.Rel(root, abs)
		original := string(data)
		result := original

		// Rewrite [[Old Title]] → [[New Title]]
		result = strings.ReplaceAll(result, "[["+oldTitle+"]]", "[["+newTitle+"]]")

		// Rewrite [text](old/relative/path) → [text](new/relative/path)
		pageDir := filepath.Dir(pageRel)
		oldRelFromPage, _ := filepath.Rel(pageDir, oldRel)
		newRelFromPage, _ := filepath.Rel(pageDir, newRel)

		result = rewriteMarkdownLinks(result, oldRelFromPage, newRelFromPage)

		if result != original {
			_ = os.WriteFile(abs, []byte(result), 0o644)
		}
		return nil
	})

	// Fix outgoing links in the moved page itself
	movedData, readErr := os.ReadFile(newAbs)
	if readErr == nil {
		movedContent := string(movedData)
		fixed := fixOutgoingLinks(movedContent, oldRel, newRel)
		if fixed != movedContent {
			_ = os.WriteFile(newAbs, []byte(fixed), 0o644)
		}
	}

	return nil
}

// rewriteMarkdownLinks replaces [text](oldPath) with [text](newPath) in content.
func rewriteMarkdownLinks(content, oldPath, newPath string) string {
	return mdLinkRe.ReplaceAllStringFunc(content, func(match string) string {
		subs := mdLinkRe.FindStringSubmatch(match)
		if len(subs) != 3 {
			return match
		}
		linkPath := subs[2]
		if filepath.Clean(linkPath) == filepath.Clean(oldPath) {
			return "[" + subs[1] + "](" + newPath + ")"
		}
		return match
	})
}

// fixOutgoingLinks adjusts relative links in a moved page so they still point
// to the same targets.
func fixOutgoingLinks(content, oldRel, newRel string) string {
	oldDir := filepath.Dir(oldRel)
	newDir := filepath.Dir(newRel)
	if oldDir == newDir {
		return content
	}

	return mdLinkRe.ReplaceAllStringFunc(content, func(match string) string {
		subs := mdLinkRe.FindStringSubmatch(match)
		if len(subs) != 3 {
			return match
		}
		linkPath := subs[2]
		// Skip external links
		if strings.Contains(linkPath, "://") {
			return match
		}
		// Resolve target relative to old location, then compute new relative path
		target := filepath.Join(oldDir, linkPath)
		target = filepath.Clean(target)
		newLink, _ := filepath.Rel(newDir, target)
		return "[" + subs[1] + "](" + newLink + ")"
	})
}
