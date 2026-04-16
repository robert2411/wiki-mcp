package wiki

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"

	"github.com/mark3labs/mcp-go/mcp"
	mcpserver "github.com/mark3labs/mcp-go/server"
	"github.com/robertstevens/wiki-mcp/internal/config"
	"github.com/robertstevens/wiki-mcp/internal/server"
)

const (
	pageURIPrefix    = "wiki://page/"
	mimeMarkdown     = "text/markdown"
	mimeJSON         = "application/json"
	maxPageResources = 500 // cap for list_resources page enumeration
)

// textResource wraps text content in the single-element slice MCP resource handlers return.
func textResource(uri, mime, text string) []mcp.ResourceContents {
	return []mcp.ResourceContents{
		mcp.TextResourceContents{URI: uri, MIMEType: mime, Text: text},
	}
}

// RegisterResources registers the four MCP wiki resources on the server.
// It also scans disk to register individual wiki://page/<path> resources
// (up to maxPageResources) so that list_resources enumerates them.
func RegisterResources(srv *server.Server) {
	cfg := srv.Config()

	srv.RegisterResource(
		mcp.NewResource("wiki://index", "Wiki Index",
			mcp.WithResourceDescription("Contents of index.md"),
			mcp.WithMIMEType(mimeMarkdown)),
		handleResourceIndex(cfg),
	)

	srv.RegisterResource(
		mcp.NewResource("wiki://log/recent", "Recent Wiki Log",
			mcp.WithResourceDescription("Last 20 log entries from log.md"),
			mcp.WithMIMEType(mimeMarkdown)),
		handleResourceLogRecent(cfg),
	)

	srv.RegisterResource(
		mcp.NewResource("wiki://config", "Wiki Config",
			mcp.WithResourceDescription("Active configuration (safety section redacted)"),
			mcp.WithMIMEType(mimeJSON)),
		handleResourceConfig(cfg),
	)

	pageHandler := handleResourcePage(cfg)

	// Template handles reads for any path, including pages beyond the cap.
	srv.RegisterResourceTemplate(
		mcp.NewResourceTemplate("wiki://page/{+path}", "Wiki Page",
			mcp.WithTemplateDescription("Read any wiki page by relative path"),
			mcp.WithTemplateMIMEType(mimeMarkdown)),
		mcpserver.ResourceTemplateHandlerFunc(pageHandler),
	)

	// Enumerate individual page URIs so list_resources includes them.
	pages, err := PageList(cfg, PageListFilter{})
	if err != nil {
		slog.Warn("RegisterResources: PageList failed; page URIs will not appear in list_resources", "err", err)
		return
	}
	limit := min(len(pages), maxPageResources)
	for _, p := range pages[:limit] {
		uri := pageURIPrefix + filepath.ToSlash(p.Path)
		srv.RegisterResource(
			mcp.NewResource(uri, TitleFromPath(p.Path),
				mcp.WithMIMEType(mimeMarkdown)),
			pageHandler,
		)
	}
}

func handleResourceIndex(cfg *config.Config) mcpserver.ResourceHandlerFunc {
	// Resolve path once at construction; it cannot change after startup.
	abs, resolveErr := cfg.ResolveWikiPath("index.md")
	return func(_ context.Context, req mcp.ReadResourceRequest) ([]mcp.ResourceContents, error) {
		if resolveErr != nil {
			return nil, resolveErr
		}
		data, err := os.ReadFile(abs)
		if err != nil {
			if os.IsNotExist(err) {
				data = []byte("# Index\n\n(index.md not found)\n")
			} else {
				return nil, err
			}
		}
		return textResource(req.Params.URI, mimeMarkdown, string(data)), nil
	}
}

func handleResourceLogRecent(cfg *config.Config) mcpserver.ResourceHandlerFunc {
	return func(_ context.Context, req mcp.ReadResourceRequest) ([]mcp.ResourceContents, error) {
		entries, te := LogTail(cfg, 20)
		if te != nil {
			return nil, te
		}
		var sb strings.Builder
		for _, e := range entries {
			fmt.Fprintf(&sb, "## [%s] %s | %s\n", e.Date, e.Operation, e.Title)
			if e.Body != "" {
				sb.WriteString("\n")
				sb.WriteString(e.Body)
				sb.WriteString("\n")
			}
			sb.WriteString("\n")
		}
		return textResource(req.Params.URI, mimeMarkdown, sb.String()), nil
	}
}

// safeConfig is Config with the Safety section replaced by a redaction marker,
// guarding against future sensitive additions to SafetyConfig.
type safeConfig struct {
	WikiPath    string             `json:"wiki_path"`
	SourcesPath string             `json:"sources_path"`
	Web         config.WebConfig   `json:"web"`
	Index       config.IndexConfig `json:"index"`
	Log         config.LogConfig   `json:"log"`
	Links       config.LinksConfig `json:"links"`
	Safety      string             `json:"safety"`
}

func handleResourceConfig(cfg *config.Config) mcpserver.ResourceHandlerFunc {
	// Marshal once at construction; cfg is immutable after startup.
	b, err := json.Marshal(safeConfig{
		WikiPath:    cfg.WikiPath,
		SourcesPath: cfg.SourcesPath,
		Web:         cfg.Web,
		Index:       cfg.Index,
		Log:         cfg.Log,
		Links:       cfg.Links,
		Safety:      "[redacted]",
	})
	return func(_ context.Context, req mcp.ReadResourceRequest) ([]mcp.ResourceContents, error) {
		if err != nil {
			return nil, err
		}
		return textResource(req.Params.URI, mimeJSON, string(b)), nil
	}
}

func handleResourcePage(cfg *config.Config) mcpserver.ResourceHandlerFunc {
	return func(_ context.Context, req mcp.ReadResourceRequest) ([]mcp.ResourceContents, error) {
		uri := req.Params.URI
		if !strings.HasPrefix(uri, pageURIPrefix) {
			return nil, fmt.Errorf("invalid wiki page URI: %q", uri)
		}
		relPath := uri[len(pageURIPrefix):]
		if relPath == "" {
			return nil, fmt.Errorf("empty path in URI: %q", uri)
		}

		abs, err := cfg.ResolveWikiPath(relPath)
		if err != nil {
			return nil, err
		}

		data, err := os.ReadFile(abs)
		if err != nil {
			if os.IsNotExist(err) {
				return nil, fmt.Errorf("page %q not found", relPath)
			}
			return nil, err
		}

		return textResource(uri, mimeMarkdown, string(data)), nil
	}
}
