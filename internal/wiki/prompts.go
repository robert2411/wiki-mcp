package wiki

import (
	"bytes"
	"context"
	_ "embed"
	"strings"
	"text/template"

	"github.com/mark3labs/mcp-go/mcp"
	mcpserver "github.com/mark3labs/mcp-go/server"
	"github.com/robertstevens/wiki-mcp/internal/server"
)

//go:embed prompts/ingest.md
var ingestPromptTemplate string

//go:embed prompts/query.md
var queryPromptTemplate string

//go:embed prompts/lint.md
var lintPromptText string

// RegisterPrompts registers all MCP prompts on the server.
func RegisterPrompts(srv *server.Server) {
	srv.RegisterPrompt(ingestPromptDef(), handleIngestPrompt())
	srv.RegisterPrompt(queryPromptDef(), handleQueryPrompt())
	srv.RegisterPrompt(lintPromptDef(), handleLintPrompt())
}

func ingestPromptDef() mcp.Prompt {
	return mcp.NewPrompt("ingest",
		mcp.WithPromptDescription("Ingest a source (URL, PDF, or text) into the wiki. Creates pages, updates index.md, and appends to log.md."),
		mcp.WithArgument("source",
			mcp.RequiredArgument(),
			mcp.ArgumentDescription("URL, local path to a PDF/image, or raw text to ingest"),
		),
		mcp.WithArgument("hint",
			mcp.ArgumentDescription("Optional note about what matters most in this source"),
		),
	)
}

type ingestPromptData struct {
	Source string
	Hint   string
}

var ingestTmpl = template.Must(template.New("ingest").Parse(ingestPromptTemplate))

func handleIngestPrompt() mcpserver.PromptHandlerFunc {
	return func(ctx context.Context, req mcp.GetPromptRequest) (*mcp.GetPromptResult, error) {
		args := req.Params.Arguments
		source := strings.TrimSpace(args["source"])
		hint := strings.TrimSpace(args["hint"])

		var buf bytes.Buffer
		if err := ingestTmpl.Execute(&buf, ingestPromptData{Source: source, Hint: hint}); err != nil {
			return nil, err
		}

		return mcp.NewGetPromptResult(
			"Ingest a source into the wiki",
			[]mcp.PromptMessage{
				{
					Role:    mcp.RoleUser,
					Content: mcp.NewTextContent(buf.String()),
				},
			},
		), nil
	}
}

func queryPromptDef() mcp.Prompt {
	return mcp.NewPrompt("query",
		mcp.WithPromptDescription("Answer a question using the wiki. Reads index then relevant pages. Optionally files the answer as a new page."),
		mcp.WithArgument("question",
			mcp.RequiredArgument(),
			mcp.ArgumentDescription("The question to answer from wiki content"),
		),
		mcp.WithArgument("file_answer",
			mcp.ArgumentDescription("If 'true', file the answer as a new wiki page without asking"),
		),
	)
}

type queryPromptData struct {
	Question   string
	FileAnswer bool
}

var queryTmpl = template.Must(template.New("query").Parse(queryPromptTemplate))

func lintPromptDef() mcp.Prompt {
	return mcp.NewPrompt("lint",
		mcp.WithPromptDescription("Lint the wiki: find orphans, contradictions, staleness, gaps, and missing cross-refs. Reports findings and offers fixes. Appends a lint pass entry to log.md on completion."),
	)
}

func handleLintPrompt() mcpserver.PromptHandlerFunc {
	return func(ctx context.Context, req mcp.GetPromptRequest) (*mcp.GetPromptResult, error) {
		return mcp.NewGetPromptResult(
			"Lint the wiki",
			[]mcp.PromptMessage{
				{
					Role:    mcp.RoleUser,
					Content: mcp.NewTextContent(lintPromptText),
				},
			},
		), nil
	}
}

func handleQueryPrompt() mcpserver.PromptHandlerFunc {
	return func(ctx context.Context, req mcp.GetPromptRequest) (*mcp.GetPromptResult, error) {
		args := req.Params.Arguments
		question := strings.TrimSpace(args["question"])
		fileAnswer := strings.TrimSpace(strings.ToLower(args["file_answer"])) == "true"

		var buf bytes.Buffer
		if err := queryTmpl.Execute(&buf, queryPromptData{Question: question, FileAnswer: fileAnswer}); err != nil {
			return nil, err
		}

		return mcp.NewGetPromptResult(
			"Answer a question from the wiki",
			[]mcp.PromptMessage{
				{
					Role:    mcp.RoleUser,
					Content: mcp.NewTextContent(buf.String()),
				},
			},
		), nil
	}
}
