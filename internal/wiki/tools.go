package wiki

import (
	"context"
	"fmt"
	"time"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/robertstevens/wiki-mcp/internal/config"
	"github.com/robertstevens/wiki-mcp/internal/server"
)

// RegisterTools registers all page CRUD tools on the server.
func RegisterTools(srv *server.Server) {
	cfg := srv.Config()

	srv.RegisterTool(pageReadTool(), handlePageRead(cfg))
	srv.RegisterTool(pageWriteTool(), handlePageWrite(cfg))
	srv.RegisterTool(pageDeleteTool(), handlePageDelete(cfg))
	srv.RegisterTool(pageListTool(), handlePageList(cfg))
	srv.RegisterTool(pageMoveTool(), handlePageMove(cfg))

	srv.RegisterTool(indexReadTool(), handleIndexRead(cfg))
	srv.RegisterTool(indexUpsertEntryTool(), handleIndexUpsertEntry(cfg))
	srv.RegisterTool(indexRefreshStatsTool(), handleIndexRefreshStats(cfg))

	srv.RegisterTool(logAppendTool(), handleLogAppend(cfg))
	srv.RegisterTool(logTailTool(), handleLogTail(cfg))

	srv.RegisterTool(wikiSearchTool(), handleWikiSearch(cfg))
	srv.RegisterTool(linksOutgoingTool(), handleLinksOutgoing(cfg))
	srv.RegisterTool(linksIncomingTool(), handleLinksIncoming(cfg))
	srv.RegisterTool(orphansTool(), handleOrphans(cfg))
}

// --- Tool definitions ---

func pageReadTool() mcp.Tool {
	return mcp.NewTool("page_read",
		mcp.WithDescription("Read a wiki page by relative path. Returns frontmatter and body separately."),
		mcp.WithString("path", mcp.Required(), mcp.Description("Relative path to the page (e.g. entities/ollama.md)")),
		mcp.WithReadOnlyHintAnnotation(true),
		mcp.WithDestructiveHintAnnotation(false),
	)
}

func pageWriteTool() mcp.Tool {
	return mcp.NewTool("page_write",
		mcp.WithDescription("Create or overwrite a wiki page. Creates parent directories as needed."),
		mcp.WithString("path", mcp.Required(), mcp.Description("Relative path for the page")),
		mcp.WithString("body", mcp.Required(), mcp.Description("Markdown body content")),
		mcp.WithObject("frontmatter", mcp.Description("Optional YAML frontmatter as key-value pairs")),
		mcp.WithDestructiveHintAnnotation(true),
		mcp.WithIdempotentHintAnnotation(true),
	)
}

func pageDeleteTool() mcp.Tool {
	return mcp.NewTool("page_delete",
		mcp.WithDescription("Delete a wiki page. Cannot delete index.md or log.md."),
		mcp.WithString("path", mcp.Required(), mcp.Description("Relative path to the page to delete")),
		mcp.WithDestructiveHintAnnotation(true),
	)
}

func pageListTool() mcp.Tool {
	return mcp.NewTool("page_list",
		mcp.WithDescription("List wiki pages with optional filters."),
		mcp.WithString("dir", mcp.Description("Filter by directory prefix (e.g. entities)")),
		mcp.WithString("glob", mcp.Description("Filter by glob pattern (e.g. *.md)")),
		mcp.WithString("tag", mcp.Description("Filter by frontmatter tag")),
		mcp.WithString("updated_since", mcp.Description("Filter by updated date (YYYY-MM-DD)")),
		mcp.WithReadOnlyHintAnnotation(true),
		mcp.WithDestructiveHintAnnotation(false),
	)
}

func pageMoveTool() mcp.Tool {
	return mcp.NewTool("page_move",
		mcp.WithDescription("Move/rename a wiki page. Rewrites incoming links across the wiki."),
		mcp.WithString("old_path", mcp.Required(), mcp.Description("Current relative path of the page")),
		mcp.WithString("new_path", mcp.Required(), mcp.Description("New relative path for the page")),
		mcp.WithDestructiveHintAnnotation(true),
	)
}

// --- Tool handlers ---

func toolErrorResult(te *ToolError) *mcp.CallToolResult {
	r := mcp.NewToolResultText(te.JSON())
	r.IsError = true
	return r
}

func handlePageRead(cfg *config.Config) func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		path, err := req.RequireString("path")
		if err != nil {
			return toolErrorResult(NewToolError(ErrCodeBadRequest, err.Error())), nil
		}

		result, te := PageRead(cfg, path)
		if te != nil {
			return toolErrorResult(te), nil
		}

		return mcp.NewToolResultJSON(result)
	}
}

func handlePageWrite(cfg *config.Config) func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		path, err := req.RequireString("path")
		if err != nil {
			return toolErrorResult(NewToolError(ErrCodeBadRequest, err.Error())), nil
		}
		body, err := req.RequireString("body")
		if err != nil {
			return toolErrorResult(NewToolError(ErrCodeBadRequest, err.Error())), nil
		}

		var fm map[string]any
		args := req.GetArguments()
		if fmRaw, ok := args["frontmatter"]; ok && fmRaw != nil {
			if fmMap, ok := fmRaw.(map[string]any); ok {
				fm = fmMap
			}
		}

		if te := PageWrite(cfg, path, fm, body); te != nil {
			return toolErrorResult(te), nil
		}

		return mcp.NewToolResultText(fmt.Sprintf("page %q written successfully", path)), nil
	}
}

func handlePageDelete(cfg *config.Config) func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		path, err := req.RequireString("path")
		if err != nil {
			return toolErrorResult(NewToolError(ErrCodeBadRequest, err.Error())), nil
		}

		if te := PageDelete(cfg, path); te != nil {
			return toolErrorResult(te), nil
		}

		return mcp.NewToolResultText(fmt.Sprintf("page %q deleted successfully", path)), nil
	}
}

func handlePageList(cfg *config.Config) func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		filter := PageListFilter{
			Dir:          req.GetString("dir", ""),
			Glob:         req.GetString("glob", ""),
			Tag:          req.GetString("tag", ""),
			UpdatedSince: req.GetString("updated_since", ""),
		}

		entries, te := PageList(cfg, filter)
		if te != nil {
			return toolErrorResult(te), nil
		}

		return mcp.NewToolResultJSON(entries)
	}
}

func handlePageMove(cfg *config.Config) func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		oldPath, err := req.RequireString("old_path")
		if err != nil {
			return toolErrorResult(NewToolError(ErrCodeBadRequest, err.Error())), nil
		}
		newPath, err := req.RequireString("new_path")
		if err != nil {
			return toolErrorResult(NewToolError(ErrCodeBadRequest, err.Error())), nil
		}

		if te := PageMove(cfg, oldPath, newPath); te != nil {
			return toolErrorResult(te), nil
		}

		return mcp.NewToolResultText(fmt.Sprintf("page moved from %q to %q", oldPath, newPath)), nil
	}
}

func indexReadTool() mcp.Tool {
	return mcp.NewTool("index_read",
		mcp.WithDescription("Read and parse index.md into structured form with sections and stats."),
		mcp.WithReadOnlyHintAnnotation(true),
		mcp.WithDestructiveHintAnnotation(false),
	)
}

func indexUpsertEntryTool() mcp.Tool {
	return mcp.NewTool("index_upsert_entry",
		mcp.WithDescription("Add or update a single entry in index.md. Updates summary if title+path match; otherwise appends within the section."),
		mcp.WithString("section_key", mcp.Required(), mcp.Description("Section key (e.g. research, entities, concepts, infrastructure)")),
		mcp.WithString("title", mcp.Required(), mcp.Description("Page title")),
		mcp.WithString("path", mcp.Required(), mcp.Description("Relative path to the page")),
		mcp.WithString("summary", mcp.Required(), mcp.Description("One-line summary for the entry")),
		mcp.WithDestructiveHintAnnotation(false),
		mcp.WithIdempotentHintAnnotation(true),
	)
}

func indexRefreshStatsTool() mcp.Tool {
	return mcp.NewTool("index_refresh_stats",
		mcp.WithDescription("Recompute page count and last-updated date from disk and rewrite the Stats block in index.md."),
		mcp.WithDestructiveHintAnnotation(false),
		mcp.WithIdempotentHintAnnotation(true),
	)
}

func handleIndexRead(cfg *config.Config) func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		doc, te := IndexRead(cfg)
		if te != nil {
			return toolErrorResult(te), nil
		}
		return mcp.NewToolResultJSON(doc)
	}
}

func handleIndexUpsertEntry(cfg *config.Config) func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		sectionKey, err := req.RequireString("section_key")
		if err != nil {
			return toolErrorResult(NewToolError(ErrCodeBadRequest, err.Error())), nil
		}
		title, err := req.RequireString("title")
		if err != nil {
			return toolErrorResult(NewToolError(ErrCodeBadRequest, err.Error())), nil
		}
		path, err := req.RequireString("path")
		if err != nil {
			return toolErrorResult(NewToolError(ErrCodeBadRequest, err.Error())), nil
		}
		summary, err := req.RequireString("summary")
		if err != nil {
			return toolErrorResult(NewToolError(ErrCodeBadRequest, err.Error())), nil
		}

		if te := IndexUpsertEntry(cfg, sectionKey, title, path, summary); te != nil {
			return toolErrorResult(te), nil
		}

		return mcp.NewToolResultText(fmt.Sprintf("entry %q upserted in section %q", title, sectionKey)), nil
	}
}

func handleIndexRefreshStats(cfg *config.Config) func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		if te := IndexRefreshStats(cfg); te != nil {
			return toolErrorResult(te), nil
		}
		return mcp.NewToolResultText("index stats refreshed"), nil
	}
}

func logAppendTool() mcp.Tool {
	return mcp.NewTool("log_append",
		mcp.WithDescription("Append an entry to log.md. Creates log.md from template if missing. Operation must be one of: ingest, query, lint."),
		mcp.WithString("operation", mcp.Required(), mcp.Description("Log operation type: ingest, query, or lint")),
		mcp.WithString("title", mcp.Required(), mcp.Description("Entry title")),
		mcp.WithString("body", mcp.Description("Entry body (markdown)")),
		mcp.WithDestructiveHintAnnotation(false),
	)
}

func logTailTool() mcp.Tool {
	return mcp.NewTool("log_tail",
		mcp.WithDescription("Return the last N log entries from log.md parsed as structured objects. Default N=10."),
		mcp.WithNumber("n", mcp.Description("Number of entries to return (default 10)")),
		mcp.WithReadOnlyHintAnnotation(true),
		mcp.WithDestructiveHintAnnotation(false),
	)
}

func handleLogAppend(cfg *config.Config) func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		operation, err := req.RequireString("operation")
		if err != nil {
			return toolErrorResult(NewToolError(ErrCodeBadRequest, err.Error())), nil
		}
		title, err := req.RequireString("title")
		if err != nil {
			return toolErrorResult(NewToolError(ErrCodeBadRequest, err.Error())), nil
		}
		body := req.GetString("body", "")

		if te := LogAppend(cfg, operation, title, body); te != nil {
			return toolErrorResult(te), nil
		}

		return mcp.NewToolResultText(fmt.Sprintf("log entry appended: [%s] %s | %s", time.Now().Format("2006-01-02"), operation, title)), nil
	}
}

func handleLogTail(cfg *config.Config) func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		n := int(req.GetFloat("n", 10))

		entries, te := LogTail(cfg, n)
		if te != nil {
			return toolErrorResult(te), nil
		}

		return mcp.NewToolResultJSON(entries)
	}
}

func wikiSearchTool() mcp.Tool {
	return mcp.NewTool("wiki_search",
		mcp.WithDescription("Full-text search across wiki page bodies. Returns scored results with snippets."),
		mcp.WithString("query", mcp.Required(), mcp.Description("Search query (substring or regex)")),
		mcp.WithNumber("limit", mcp.Description("Max results to return (default 20)")),
		mcp.WithReadOnlyHintAnnotation(true),
		mcp.WithDestructiveHintAnnotation(false),
	)
}

func linksOutgoingTool() mcp.Tool {
	return mcp.NewTool("links_outgoing",
		mcp.WithDescription("Return all outgoing links from a wiki page. Internal links include wikilinks and relative paths; external links include full URLs."),
		mcp.WithString("path", mcp.Required(), mcp.Description("Relative path to the page")),
		mcp.WithReadOnlyHintAnnotation(true),
		mcp.WithDestructiveHintAnnotation(false),
	)
}

func linksIncomingTool() mcp.Tool {
	return mcp.NewTool("links_incoming",
		mcp.WithDescription("Return all pages that link to a given wiki page (backlinks)."),
		mcp.WithString("path", mcp.Required(), mcp.Description("Relative path to the target page")),
		mcp.WithReadOnlyHintAnnotation(true),
		mcp.WithDestructiveHintAnnotation(false),
	)
}

func orphansTool() mcp.Tool {
	return mcp.NewTool("orphans",
		mcp.WithDescription("Return pages with zero incoming links (excluding index.md and log.md)."),
		mcp.WithReadOnlyHintAnnotation(true),
		mcp.WithDestructiveHintAnnotation(false),
	)
}

func handleWikiSearch(cfg *config.Config) func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		query, err := req.RequireString("query")
		if err != nil {
			return toolErrorResult(NewToolError(ErrCodeBadRequest, err.Error())), nil
		}
		limit := int(req.GetFloat("limit", 20))

		results, te := WikiSearch(cfg, query, limit)
		if te != nil {
			return toolErrorResult(te), nil
		}

		return mcp.NewToolResultJSON(results)
	}
}

func handleLinksOutgoing(cfg *config.Config) func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		path, err := req.RequireString("path")
		if err != nil {
			return toolErrorResult(NewToolError(ErrCodeBadRequest, err.Error())), nil
		}

		result, te := LinksOutgoing(cfg, path)
		if te != nil {
			return toolErrorResult(te), nil
		}

		return mcp.NewToolResultJSON(result)
	}
}

func handleLinksIncoming(cfg *config.Config) func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		path, err := req.RequireString("path")
		if err != nil {
			return toolErrorResult(NewToolError(ErrCodeBadRequest, err.Error())), nil
		}

		backlinks, te := LinksIncoming(cfg, path)
		if te != nil {
			return toolErrorResult(te), nil
		}

		return mcp.NewToolResultJSON(backlinks)
	}
}

func handleOrphans(cfg *config.Config) func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		orphans, te := Orphans(cfg)
		if te != nil {
			return toolErrorResult(te), nil
		}

		return mcp.NewToolResultJSON(orphans)
	}
}
