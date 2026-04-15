// Package linkgraph provides link parsing for wiki markdown pages.
package linkgraph

import (
	"path/filepath"
	"regexp"
	"strings"
)

var (
	wikiLinkRe = regexp.MustCompile(`\[\[([^\]]+)\]\]`)
	mdLinkRe   = regexp.MustCompile(`\[([^\]]*)\]\(([^)]+)\)`)
)

// Links holds the outgoing links parsed from a page.
type Links struct {
	// Internal contains wiki-root-relative paths (for md links) or wikilink
	// titles (for [[Title]] / [[Title|alias]] forms).
	Internal []string `json:"internal"`
	// External contains full URLs (links containing "://").
	External []string `json:"external"`
}

// ParseOutgoing extracts all outgoing links from markdown content.
//
// relPath is the page's path relative to the wiki root. It is used to resolve
// relative md-style links into wiki-root-relative paths.
//
// Wikilinks [[Title]] and [[Title|alias]] are added to Internal as the target
// title string (no path resolution — titles are not mapped to paths here).
// Relative md links are resolved against relPath and added as wiki-root-relative
// paths. External URLs (containing "://") are added to External.
func ParseOutgoing(relPath, content string) Links {
	var internal, external []string
	seen := make(map[string]bool)

	add := func(list *[]string, val string) {
		if !seen[val] {
			seen[val] = true
			*list = append(*list, val)
		}
	}

	pageDir := filepath.Dir(relPath)

	// [[Title]] and [[Title|alias]]
	for _, m := range wikiLinkRe.FindAllStringSubmatch(content, -1) {
		raw := m[1]
		title := raw
		if idx := strings.Index(raw, "|"); idx >= 0 {
			title = raw[:idx] // take the target title, not the alias
		}
		add(&internal, title)
	}

	// [text](path) — md links
	for _, m := range mdLinkRe.FindAllStringSubmatch(content, -1) {
		target := m[2]
		if strings.Contains(target, "://") {
			add(&external, target)
			continue
		}
		// Resolve relative path to wiki root
		resolved := filepath.ToSlash(filepath.Clean(filepath.Join(pageDir, target)))
		add(&internal, resolved)
	}

	return Links{Internal: internal, External: external}
}
