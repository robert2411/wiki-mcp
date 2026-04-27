package wiki

import (
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/robert2411/wiki-mcp/internal/config"
	"github.com/robert2411/wiki-mcp/internal/wiki/linkgraph"
)

// protectedLinkSources are pages whose links are not counted as incoming links
// when computing backlinks and orphans.
var protectedLinkSources = map[string]bool{
	"index.md": true,
	"log.md":   true,
}

// LinksOutgoingResult holds the result of parsing a page's outgoing links.
type LinksOutgoingResult struct {
	Internal []string `json:"internal"`
	External []string `json:"external"`
}

// LinksOutgoing returns all outgoing links from a wiki page.
func LinksOutgoing(cfg *config.Config, relPath string) (*LinksOutgoingResult, *ToolError) {
	abs, err := cfg.ResolveWikiPath(relPath)
	if err != nil {
		return nil, NewToolError(ErrCodePathEscape, err.Error())
	}

	data, err := os.ReadFile(abs)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, NewToolError(ErrCodeNotFound, "page not found: "+relPath)
		}
		return nil, NewToolError(ErrCodeInternal, err.Error())
	}

	_, body := ParseFrontmatter(data)
	links := linkgraph.ParseOutgoing(relPath, body)
	return &LinksOutgoingResult{
		Internal: links.Internal,
		External: links.External,
	}, nil
}

// LinksIncoming returns all pages that link to relPath.
func LinksIncoming(cfg *config.Config, relPath string) ([]string, *ToolError) {
	abs, err := cfg.ResolveWikiPath(relPath)
	if err != nil {
		return nil, NewToolError(ErrCodePathEscape, err.Error())
	}
	if _, err := os.Stat(abs); os.IsNotExist(err) {
		return nil, NewToolError(ErrCodeNotFound, "page not found: "+relPath)
	}

	targetTitle := TitleFromPath(relPath)
	root := cfg.Root()
	var backlinks []string

	_ = filepath.WalkDir(root, func(p string, d os.DirEntry, walkErr error) error {
		if walkErr != nil || d.IsDir() || !strings.HasSuffix(p, ".md") {
			return nil
		}
		rel, _ := filepath.Rel(root, p)
		if rel == relPath {
			return nil // skip the page itself
		}

		data, err := os.ReadFile(p)
		if err != nil {
			return nil
		}

		_, body := ParseFrontmatter(data)
		links := linkgraph.ParseOutgoing(rel, body)

		for _, internal := range links.Internal {
			// Match resolved path or wikilink title
			if internal == relPath || internal == targetTitle {
				backlinks = append(backlinks, rel)
				return nil
			}
		}
		return nil
	})

	sort.Strings(backlinks)
	return backlinks, nil
}

// Orphans returns pages with zero incoming links, excluding index.md and log.md
// both as candidates and as link sources.
func Orphans(cfg *config.Config) ([]string, *ToolError) {
	root := cfg.Root()

	var candidates []string
	hasIncoming := make(map[string]bool)

	_ = filepath.WalkDir(root, func(p string, d os.DirEntry, walkErr error) error {
		if walkErr != nil || d.IsDir() || !strings.HasSuffix(p, ".md") {
			return nil
		}
		rel, _ := filepath.Rel(root, p)
		protected := protectedLinkSources[filepath.Base(rel)]

		if !protected {
			candidates = append(candidates, rel)
		}

		data, err := os.ReadFile(p)
		if err != nil || protected {
			return nil // skip protected sources when building link map
		}

		_, body := ParseFrontmatter(data)
		for _, internal := range linkgraph.ParseOutgoing(rel, body).Internal {
			hasIncoming[internal] = true
		}
		return nil
	})

	var orphans []string
	for _, candidate := range candidates {
		title := TitleFromPath(candidate)
		if !hasIncoming[candidate] && !hasIncoming[title] {
			orphans = append(orphans, candidate)
		}
	}

	sort.Strings(orphans)
	return orphans, nil
}
