package sources

import (
	"context"
	"fmt"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/robert2411/wiki-mcp/internal/server"
)

// RegisterTools registers source helper tools on the MCP server.
func RegisterTools(srv *server.Server) {
	sourcesPath := srv.Config().SourcesPath
	srv.RegisterTool(sourceFetchURLTool(), handleSourceFetchURL(sourcesPath))
	srv.RegisterTool(sourcePDFTextTool(), handleSourcePDFText(sourcesPath))
	srv.RegisterTool(sourceListTool(), handleSourceList(sourcesPath))
}

func toolErrorResult(te *ToolError) *mcp.CallToolResult {
	r := mcp.NewToolResultText(te.JSON())
	r.IsError = true
	return r
}

func sourceFetchURLTool() mcp.Tool {
	return mcp.NewTool("source_fetch_url",
		mcp.WithDescription("Fetch a URL and save it as a markdown file in the sources directory. Uses pure Go net/http — no external curl required. HTML is converted to markdown; other content types are saved as-is."),
		mcp.WithString("url", mcp.Required(), mcp.Description("URL to fetch")),
		mcp.WithString("slug", mcp.Description("Optional filename slug (without .md). Defaults to sanitized URL host+path.")),
		mcp.WithDestructiveHintAnnotation(false),
	)
}

func sourcePDFTextTool() mcp.Tool {
	return mcp.NewTool("source_pdf_text",
		mcp.WithDescription("Extract plain text from a PDF stored in the sources directory. Returns extracted text and page count."),
		mcp.WithString("path", mcp.Required(), mcp.Description("Relative path to the PDF within the sources directory")),
		mcp.WithReadOnlyHintAnnotation(true),
		mcp.WithDestructiveHintAnnotation(false),
	)
}

func sourceListTool() mcp.Tool {
	return mcp.NewTool("source_list",
		mcp.WithDescription("List files in the sources directory with name, size, and modification time. Returns an empty list when the directory does not exist yet."),
		mcp.WithReadOnlyHintAnnotation(true),
		mcp.WithDestructiveHintAnnotation(false),
	)
}

func handleSourceFetchURL(sourcesPath string) func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		rawURL, err := req.RequireString("url")
		if err != nil {
			return toolErrorResult(newToolError(ErrCodeBadRequest, err.Error())), nil
		}
		slug := req.GetString("slug", "")

		savedPath, te := FetchURL(sourcesPath, rawURL, slug)
		if te != nil {
			return toolErrorResult(te), nil
		}
		return mcp.NewToolResultText(fmt.Sprintf("saved %q to %q", rawURL, savedPath)), nil
	}
}

func handleSourcePDFText(sourcesPath string) func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		path, err := req.RequireString("path")
		if err != nil {
			return toolErrorResult(newToolError(ErrCodeBadRequest, err.Error())), nil
		}

		result, te := PDFText(sourcesPath, path)
		if te != nil {
			return toolErrorResult(te), nil
		}
		return mcp.NewToolResultJSON(result)
	}
}

func handleSourceList(sourcesPath string) func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		entries, te := List(sourcesPath)
		if te != nil {
			return toolErrorResult(te), nil
		}
		return mcp.NewToolResultJSON(entries)
	}
}
