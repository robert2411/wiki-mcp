package wiki

import (
	"context"
	"strings"
	"testing"

	"github.com/mark3labs/mcp-go/mcp"
)

func TestIngestPrompt_FilledText(t *testing.T) {
	handler := handleIngestPrompt()

	req := mcp.GetPromptRequest{
		Params: mcp.GetPromptParams{
			Name: "ingest",
			Arguments: map[string]string{
				"source": "https://example.com/article",
				"hint":   "focus on the benchmarks",
			},
		},
	}

	result, err := handler(context.Background(), req)
	if err != nil {
		t.Fatalf("handler error: %v", err)
	}

	if len(result.Messages) != 1 {
		t.Fatalf("expected 1 message, got %d", len(result.Messages))
	}

	tc, ok := result.Messages[0].Content.(mcp.TextContent)
	if !ok {
		t.Fatalf("expected TextContent, got %T", result.Messages[0].Content)
	}

	text := tc.Text

	// Source must appear
	if !strings.Contains(text, "https://example.com/article") {
		t.Error("source URL not in prompt text")
	}

	// Hint must appear
	if !strings.Contains(text, "focus on the benchmarks") {
		t.Error("hint not in prompt text")
	}

	// Must name all required tool calls
	for _, toolName := range []string{
		"source_fetch_url",
		"page_read",
		"page_write",
		"index_upsert_entry",
		"index_refresh_stats",
		"log_append",
	} {
		if !strings.Contains(text, toolName) {
			t.Errorf("prompt missing tool call: %q", toolName)
		}
	}

	// Must state one-source-at-a-time discipline
	if !strings.Contains(text, "one source") && !strings.Contains(text, "one at a time") {
		t.Error("prompt missing one-source-at-a-time discipline")
	}
}

func TestIngestPrompt_NoHint(t *testing.T) {
	handler := handleIngestPrompt()

	req := mcp.GetPromptRequest{
		Params: mcp.GetPromptParams{
			Name:      "ingest",
			Arguments: map[string]string{"source": "https://example.com"},
		},
	}

	result, err := handler(context.Background(), req)
	if err != nil {
		t.Fatalf("handler error: %v", err)
	}

	tc, ok := result.Messages[0].Content.(mcp.TextContent)
	if !ok {
		t.Fatalf("expected TextContent, got %T", result.Messages[0].Content)
	}

	// Hint block should not produce a stray empty line labelled "Caller note:"
	if strings.Contains(tc.Text, "Caller note:") {
		t.Error("caller note section present when no hint provided")
	}
}

func TestQueryPrompt_FilledText(t *testing.T) {
	handler := handleQueryPrompt()

	req := mcp.GetPromptRequest{
		Params: mcp.GetPromptParams{
			Name:      "query",
			Arguments: map[string]string{"question": "Which model fits on a 20GB GPU?"},
		},
	}

	result, err := handler(context.Background(), req)
	if err != nil {
		t.Fatalf("handler error: %v", err)
	}

	if len(result.Messages) != 1 {
		t.Fatalf("expected 1 message, got %d", len(result.Messages))
	}

	tc, ok := result.Messages[0].Content.(mcp.TextContent)
	if !ok {
		t.Fatalf("expected TextContent, got %T", result.Messages[0].Content)
	}
	text := tc.Text

	if !strings.Contains(text, "Which model fits on a 20GB GPU?") {
		t.Error("question not in prompt text")
	}

	// Must instruct index-first reading
	if !strings.Contains(text, "index_read") {
		t.Error("prompt missing index_read call")
	}
	if !strings.Contains(text, "page_read") {
		t.Error("prompt missing page_read call")
	}

	// file_answer=false → offer step, not direct-write instruction
	if !strings.Contains(text, "offer") {
		t.Error("file_answer=false branch should mention offering to file answer")
	}
	// Must not instruct to skip the offer
	if strings.Contains(text, "skip the") {
		t.Error("file_answer=false branch should not instruct to skip the offer step")
	}
}

func TestQueryPrompt_FileAnswer(t *testing.T) {
	handler := handleQueryPrompt()

	req := mcp.GetPromptRequest{
		Params: mcp.GetPromptParams{
			Name: "query",
			Arguments: map[string]string{
				"question":    "Best model for Java code review?",
				"file_answer": "true",
			},
		},
	}

	result, err := handler(context.Background(), req)
	if err != nil {
		t.Fatalf("handler error: %v", err)
	}

	tc, ok := result.Messages[0].Content.(mcp.TextContent)
	if !ok {
		t.Fatalf("expected TextContent, got %T", result.Messages[0].Content)
	}
	text := tc.Text

	// Must name page_write, index_upsert_entry, log_append
	for _, toolName := range []string{"page_write", "index_upsert_entry", "log_append"} {
		if !strings.Contains(text, toolName) {
			t.Errorf("file_answer=true prompt missing tool: %q", toolName)
		}
	}
}

func TestQueryPromptDef_Args(t *testing.T) {
	p := queryPromptDef()

	if p.Name != "query" {
		t.Errorf("prompt name: want query, got %q", p.Name)
	}

	if len(p.Arguments) != 2 {
		t.Fatalf("expected 2 arguments, got %d", len(p.Arguments))
	}

	if !p.Arguments[0].Required {
		t.Error("question arg should be required")
	}
	if p.Arguments[1].Required {
		t.Error("file_answer arg should not be required")
	}
}

func TestIngestPromptDef_Args(t *testing.T) {
	p := ingestPromptDef()

	if p.Name != "ingest" {
		t.Errorf("prompt name: want ingest, got %q", p.Name)
	}

	if len(p.Arguments) != 2 {
		t.Fatalf("expected 2 arguments, got %d", len(p.Arguments))
	}

	sourceArg := p.Arguments[0]
	if sourceArg.Name != "source" {
		t.Errorf("arg 0 name: want source, got %q", sourceArg.Name)
	}
	if !sourceArg.Required {
		t.Error("source arg should be required")
	}

	hintArg := p.Arguments[1]
	if hintArg.Name != "hint" {
		t.Errorf("arg 1 name: want hint, got %q", hintArg.Name)
	}
	if hintArg.Required {
		t.Error("hint arg should not be required")
	}
}
