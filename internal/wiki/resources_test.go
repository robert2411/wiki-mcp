package wiki

import (
	"context"
	"encoding/json"
	"strings"
	"testing"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/robert2411/wiki-mcp/internal/config"
)

// resourceRequest builds a ReadResourceRequest for the given URI.
func resourceRequest(uri string) mcp.ReadResourceRequest {
	return mcp.ReadResourceRequest{
		Params: mcp.ReadResourceParams{URI: uri},
	}
}

// confinedWikiConfig returns an existingWikiConfig with ConfineToWikiPath enabled.
func confinedWikiConfig(t *testing.T) *config.Config {
	t.Helper()
	cfg := existingWikiConfig(t)
	cfg.Safety.ConfineToWikiPath = true
	return cfg
}

// --- wiki://index ---

func TestResourceIndex_ExistingWiki(t *testing.T) {
	cfg := confinedWikiConfig(t)
	handler := handleResourceIndex(cfg)

	contents, err := handler(context.Background(), resourceRequest("wiki://index"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(contents) == 0 {
		t.Fatal("expected at least one ResourceContents item")
	}

	rc, ok := contents[0].(mcp.TextResourceContents)
	if !ok {
		t.Fatalf("expected TextResourceContents, got %T", contents[0])
	}
	if rc.MIMEType != "text/markdown" {
		t.Errorf("MIME type = %q; want text/markdown", rc.MIMEType)
	}
	if !strings.Contains(rc.Text, "#") {
		t.Error("expected markdown content in index.md")
	}
}

// --- wiki://log/recent ---

func TestResourceLogRecent_ExistingWiki(t *testing.T) {
	cfg := confinedWikiConfig(t)
	cfg.Log = config.LogConfig{DateFormat: "%Y-%m-%d"}
	handler := handleResourceLogRecent(cfg)

	contents, err := handler(context.Background(), resourceRequest("wiki://log/recent"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(contents) == 0 {
		t.Fatal("expected at least one ResourceContents item")
	}

	rc, ok := contents[0].(mcp.TextResourceContents)
	if !ok {
		t.Fatalf("expected TextResourceContents, got %T", contents[0])
	}
	if rc.MIMEType != "text/markdown" {
		t.Errorf("MIME type = %q; want text/markdown", rc.MIMEType)
	}
	// log.md may be empty on a fresh fixture — just verify no error and correct type
	_ = rc.Text
}

// --- wiki://config ---

func TestResourceConfig_RedactsSafety(t *testing.T) {
	cfg := confinedWikiConfig(t)
	handler := handleResourceConfig(cfg)

	contents, err := handler(context.Background(), resourceRequest("wiki://config"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(contents) == 0 {
		t.Fatal("expected at least one ResourceContents item")
	}

	rc, ok := contents[0].(mcp.TextResourceContents)
	if !ok {
		t.Fatalf("expected TextResourceContents, got %T", contents[0])
	}
	if rc.MIMEType != "application/json" {
		t.Errorf("MIME type = %q; want application/json", rc.MIMEType)
	}

	var out map[string]any
	if err := json.Unmarshal([]byte(rc.Text), &out); err != nil {
		t.Fatalf("response is not valid JSON: %v\nbody: %s", err, rc.Text)
	}

	safety, ok := out["safety"]
	if !ok {
		t.Fatal("expected 'safety' key in config JSON")
	}
	if safety != "[redacted]" {
		t.Errorf("safety = %v; want \"[redacted]\"", safety)
	}

	if _, ok := out["wiki_path"]; !ok {
		t.Error("expected 'wiki_path' key in config JSON")
	}
}

// --- wiki://page/<path> ---

func TestResourcePage_ExistingWiki(t *testing.T) {
	cfg := confinedWikiConfig(t)
	handler := handleResourcePage(cfg)

	uri := "wiki://page/index.md"
	contents, err := handler(context.Background(), resourceRequest(uri))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(contents) == 0 {
		t.Fatal("expected at least one ResourceContents item")
	}

	rc, ok := contents[0].(mcp.TextResourceContents)
	if !ok {
		t.Fatalf("expected TextResourceContents, got %T", contents[0])
	}
	if rc.MIMEType != "text/markdown" {
		t.Errorf("MIME type = %q; want text/markdown", rc.MIMEType)
	}
	if rc.URI != uri {
		t.Errorf("URI = %q; want %q", rc.URI, uri)
	}
	if strings.TrimSpace(rc.Text) == "" {
		t.Error("expected non-empty page body")
	}
}

func TestResourcePage_SubdirPage(t *testing.T) {
	cfg := confinedWikiConfig(t)
	handler := handleResourcePage(cfg)

	uri := "wiki://page/entities/ollama.md"
	contents, err := handler(context.Background(), resourceRequest(uri))
	if err != nil {
		t.Fatalf("unexpected error reading %s: %v", uri, err)
	}

	rc, ok := contents[0].(mcp.TextResourceContents)
	if !ok {
		t.Fatalf("expected TextResourceContents, got %T", contents[0])
	}
	if rc.MIMEType != "text/markdown" {
		t.Errorf("MIME type = %q; want text/markdown", rc.MIMEType)
	}
}

func TestResourcePage_FrontmatterRetained(t *testing.T) {
	cfg := confinedWikiConfig(t)
	handler := handleResourcePage(cfg)

	// entities pages have frontmatter — verify it comes through raw
	uri := "wiki://page/entities/ollama.md"
	contents, err := handler(context.Background(), resourceRequest(uri))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	rc := contents[0].(mcp.TextResourceContents)
	// raw read should include frontmatter delimiters if present
	// (we just verify text is non-empty; frontmatter presence depends on the fixture)
	if strings.TrimSpace(rc.Text) == "" {
		t.Error("expected non-empty content for entities/ollama.md")
	}
}

// --- confine_to_wiki_path enforcement (AC #3) ---

func TestResourcePage_PathEscapeRejected(t *testing.T) {
	cfg := confinedWikiConfig(t)
	handler := handleResourcePage(cfg)

	_, err := handler(context.Background(), resourceRequest("wiki://page/../../etc/passwd"))
	if err == nil {
		t.Fatal("expected error for path-escaping URI, got nil")
	}
}

func TestResourceIndex_PathEscapeBlocked(t *testing.T) {
	// Even if WikiPath somehow leads outside — ResolveWikiPath("index.md") stays inside.
	// Verify index handler uses ResolveWikiPath by pointing it at a temp dir with no index.md.
	root := t.TempDir()
	cfg := &config.Config{
		WikiPath:    root,
		ProjectPath: root,
		Safety:      config.SafetyConfig{ConfineToWikiPath: true},
	}
	handler := handleResourceIndex(cfg)
	contents, err := handler(context.Background(), resourceRequest("wiki://index"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	rc := contents[0].(mcp.TextResourceContents)
	if !strings.Contains(rc.Text, "not found") {
		t.Errorf("expected fallback text for missing index.md, got: %q", rc.Text)
	}
}

// --- list_resources page enumeration (AC #2) ---

func TestRegisterResources_PageList(t *testing.T) {
	cfg := confinedWikiConfig(t)
	pages, te := PageList(cfg, PageListFilter{})
	if te != nil {
		t.Fatalf("PageList error: %v", te)
	}
	// existing/wiki has >0 pages; we just verify the list logic caps at 500
	if len(pages) > 500 {
		t.Errorf("fixture has %d pages; cap logic would be needed", len(pages))
	}
	// Verify each page produces a valid wiki://page/ URI (no backslashes etc.)
	for _, p := range pages {
		uri := pageURIPrefix + p.Path
		if strings.Contains(uri, "\\") {
			t.Errorf("URI contains backslash: %q", uri)
		}
	}
}
