package wiki

import (
	"context"
	"fmt"

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
