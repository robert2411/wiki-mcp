package wiki

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/robertstevens/wiki-mcp/internal/config"
)

func testConfig(t *testing.T, wikiPath string) *config.Config {
	t.Helper()
	return &config.Config{
		WikiPath: wikiPath,
		Safety: config.SafetyConfig{
			ConfineToWikiPath: true,
			MaxPageBytes:      1048576,
		},
	}
}

func seedPage(t *testing.T, root, rel, content string) {
	t.Helper()
	abs := filepath.Join(root, rel)
	if err := os.MkdirAll(filepath.Dir(abs), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(abs, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

// --- Frontmatter ---

func TestParseFrontmatter(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		wantFM   bool
		wantTags int
		wantBody string
	}{
		{
			name:     "with frontmatter",
			input:    "---\ntags: [entity, model]\nupdated: 2026-04-06\n---\n# Title\nBody here.\n",
			wantFM:   true,
			wantTags: 2,
			wantBody: "# Title\nBody here.\n",
		},
		{
			name:     "no frontmatter",
			input:    "# Title\nBody here.\n",
			wantFM:   false,
			wantBody: "# Title\nBody here.\n",
		},
		{
			name:     "empty frontmatter",
			input:    "---\n---\n# Title\n",
			wantFM:   false, // empty yaml parses to nil map
			wantBody: "# Title\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fm, body := ParseFrontmatter([]byte(tt.input))
			if (fm != nil) != tt.wantFM {
				t.Errorf("frontmatter presence: got %v, want %v", fm != nil, tt.wantFM)
			}
			if tt.wantFM && tt.wantTags > 0 {
				tags, _ := fm["tags"].([]any)
				if len(tags) != tt.wantTags {
					t.Errorf("tags count: got %d, want %d", len(tags), tt.wantTags)
				}
			}
			if body != tt.wantBody {
				t.Errorf("body:\ngot:  %q\nwant: %q", body, tt.wantBody)
			}
		})
	}
}

func TestRenderFrontmatter_Empty(t *testing.T) {
	out := RenderFrontmatter(nil, "# Hello\n")
	if string(out) != "# Hello\n" {
		t.Errorf("expected no frontmatter block, got: %q", out)
	}

	out = RenderFrontmatter(map[string]any{}, "# Hello\n")
	if string(out) != "# Hello\n" {
		t.Errorf("expected no frontmatter block for empty map, got: %q", out)
	}
}

func TestRenderFrontmatter_WithData(t *testing.T) {
	fm := map[string]any{"tags": []string{"a", "b"}}
	out := RenderFrontmatter(fm, "# Hello\n")
	s := string(out)
	if !strings.HasPrefix(s, "---\n") {
		t.Error("expected frontmatter to start with ---")
	}
	if !strings.Contains(s, "tags:") {
		t.Error("expected tags in frontmatter")
	}
	if !strings.Contains(s, "# Hello") {
		t.Error("expected body after frontmatter")
	}
}

// --- PageRead ---

func TestPageRead_Success(t *testing.T) {
	root := t.TempDir()
	seedPage(t, root, "test.md", "---\ntags: [x]\n---\n# Test\nBody\n")
	cfg := testConfig(t, root)

	result, te := PageRead(cfg, "test.md")
	if te != nil {
		t.Fatalf("unexpected error: %v", te)
	}
	if result.Path != "test.md" {
		t.Errorf("path: got %q", result.Path)
	}
	if result.Frontmatter == nil {
		t.Error("expected frontmatter")
	}
	if !strings.Contains(result.Body, "# Test") {
		t.Error("expected body")
	}
}

func TestPageRead_NotFound(t *testing.T) {
	root := t.TempDir()
	cfg := testConfig(t, root)

	_, te := PageRead(cfg, "nonexistent.md")
	if te == nil {
		t.Fatal("expected error")
	}
	if te.Code != ErrCodeNotFound {
		t.Errorf("expected NotFound, got %q", te.Code)
	}
}

func TestPageRead_PathEscape(t *testing.T) {
	root := t.TempDir()
	cfg := testConfig(t, root)

	_, te := PageRead(cfg, "../../../etc/passwd")
	if te == nil {
		t.Fatal("expected error")
	}
	if te.Code != ErrCodePathEscape {
		t.Errorf("expected PathEscape, got %q", te.Code)
	}
}

// --- PageWrite ---

func TestPageWrite_CreatesParentDirs(t *testing.T) {
	root := t.TempDir()
	cfg := testConfig(t, root)

	te := PageWrite(cfg, "new/subdir/page.md", nil, "# New Page\n")
	if te != nil {
		t.Fatalf("unexpected error: %v", te)
	}

	data, err := os.ReadFile(filepath.Join(root, "new/subdir/page.md"))
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != "# New Page\n" {
		t.Errorf("content mismatch: %q", data)
	}
}

func TestPageWrite_NoEmptyFrontmatter(t *testing.T) {
	root := t.TempDir()
	cfg := testConfig(t, root)

	te := PageWrite(cfg, "page.md", nil, "# Hello\n")
	if te != nil {
		t.Fatal(te)
	}

	data, _ := os.ReadFile(filepath.Join(root, "page.md"))
	if strings.HasPrefix(string(data), "---") {
		t.Error("should not emit frontmatter block when no metadata provided")
	}
}

func TestPageWrite_WithFrontmatter(t *testing.T) {
	root := t.TempDir()
	cfg := testConfig(t, root)

	fm := map[string]any{"tags": []string{"test"}}
	te := PageWrite(cfg, "page.md", fm, "# Hello\n")
	if te != nil {
		t.Fatal(te)
	}

	data, _ := os.ReadFile(filepath.Join(root, "page.md"))
	if !strings.HasPrefix(string(data), "---\n") {
		t.Error("expected frontmatter block")
	}
}

func TestPageWrite_ReadOnly(t *testing.T) {
	root := t.TempDir()
	cfg := testConfig(t, root)
	cfg.Safety.ReadOnly = true

	te := PageWrite(cfg, "page.md", nil, "# Hello\n")
	if te == nil {
		t.Fatal("expected error")
	}
	if te.Code != ErrCodeReadOnly {
		t.Errorf("expected ReadOnly, got %q", te.Code)
	}
}

func TestPageWrite_PathEscape(t *testing.T) {
	root := t.TempDir()
	cfg := testConfig(t, root)

	te := PageWrite(cfg, "../../escape.md", nil, "bad")
	if te == nil {
		t.Fatal("expected error")
	}
	if te.Code != ErrCodePathEscape {
		t.Errorf("expected PathEscape, got %q", te.Code)
	}
}

// --- PageDelete ---

func TestPageDelete_Success(t *testing.T) {
	root := t.TempDir()
	seedPage(t, root, "deleteme.md", "# Delete me\n")
	cfg := testConfig(t, root)

	te := PageDelete(cfg, "deleteme.md")
	if te != nil {
		t.Fatal(te)
	}

	if _, err := os.Stat(filepath.Join(root, "deleteme.md")); !os.IsNotExist(err) {
		t.Error("file should have been deleted")
	}
}

func TestPageDelete_ProtectedFiles(t *testing.T) {
	root := t.TempDir()
	seedPage(t, root, "index.md", "# Index\n")
	seedPage(t, root, "log.md", "# Log\n")
	cfg := testConfig(t, root)

	for _, name := range []string{"index.md", "log.md"} {
		te := PageDelete(cfg, name)
		if te == nil {
			t.Errorf("expected error deleting %s", name)
			continue
		}
		if te.Code != ErrCodeForbidden {
			t.Errorf("expected Forbidden for %s, got %q", name, te.Code)
		}
	}
}

func TestPageDelete_ProtectedSubdir(t *testing.T) {
	root := t.TempDir()
	seedPage(t, root, "subdir/index.md", "# Sub Index\n")
	cfg := testConfig(t, root)

	te := PageDelete(cfg, "subdir/index.md")
	if te == nil {
		t.Fatal("expected error deleting subdir/index.md")
	}
	if te.Code != ErrCodeForbidden {
		t.Errorf("expected Forbidden, got %q", te.Code)
	}
}

func TestPageDelete_NotFound(t *testing.T) {
	root := t.TempDir()
	cfg := testConfig(t, root)

	te := PageDelete(cfg, "ghost.md")
	if te == nil {
		t.Fatal("expected error")
	}
	if te.Code != ErrCodeNotFound {
		t.Errorf("expected NotFound, got %q", te.Code)
	}
}

// --- PageList ---

func existingWikiPath(t *testing.T) string {
	t.Helper()
	// Find the existing wiki fixture relative to this test file
	// Walk up from internal/wiki to project root, then into existing/wiki
	wd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	root := filepath.Join(wd, "..", "..", "existing", "wiki")
	if _, err := os.Stat(root); err != nil {
		t.Skipf("existing wiki fixture not found at %s", root)
	}
	return root
}

func TestPageList_ExistingWiki(t *testing.T) {
	wikiPath := existingWikiPath(t)

	tests := []struct {
		name      string
		filter    PageListFilter
		wantMin   int
		wantCheck func(entries []PageListEntry) error
	}{
		{
			name:    "all pages",
			filter:  PageListFilter{},
			wantMin: 20,
		},
		{
			name:   "dir filter entities",
			filter: PageListFilter{Dir: "entities"},
			wantCheck: func(entries []PageListEntry) error {
				for _, e := range entries {
					if !strings.HasPrefix(e.Path, "entities/") {
						return &ToolError{Code: "test", Message: "unexpected path: " + e.Path}
					}
				}
				if len(entries) < 5 {
					return &ToolError{Code: "test", Message: "expected at least 5 entity pages"}
				}
				return nil
			},
		},
		{
			name:   "tag filter entity",
			filter: PageListFilter{Tag: "entity"},
			wantCheck: func(entries []PageListEntry) error {
				if len(entries) == 0 {
					return &ToolError{Code: "test", Message: "expected at least 1 page with tag 'entity'"}
				}
				return nil
			},
		},
		{
			name:   "updated_since filter",
			filter: PageListFilter{Tag: "entity", UpdatedSince: "2026-04-01"},
			wantCheck: func(entries []PageListEntry) error {
				if len(entries) == 0 {
					return &ToolError{Code: "test", Message: "expected at least 1 recently updated entity"}
				}
				return nil
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &config.Config{WikiPath: wikiPath}
			entries, te := PageList(cfg, tt.filter)
			if te != nil {
				t.Fatalf("unexpected error: %v", te)
			}
			if tt.wantMin > 0 && len(entries) < tt.wantMin {
				t.Errorf("got %d entries, want at least %d", len(entries), tt.wantMin)
			}
			if tt.wantCheck != nil {
				if err := tt.wantCheck(entries); err != nil {
					t.Error(err)
				}
			}
		})
	}
}

// --- PageMove ---

func TestPageMove_RewritesLinks(t *testing.T) {
	root := t.TempDir()

	// Create the page to move
	seedPage(t, root, "entities/old-name.md", "# Old Name\n\nSee [Other](../concepts/foo.md)\n")

	// Create a page that links to it
	seedPage(t, root, "concepts/foo.md",
		"# Foo\n\nSee [[Old Name]] and [old link](../entities/old-name.md)\n")

	// Create index that links to it
	seedPage(t, root, "index.md",
		"- [Old Name](entities/old-name.md) — description\n")

	cfg := testConfig(t, root)

	te := PageMove(cfg, "entities/old-name.md", "entities/new-name.md")
	if te != nil {
		t.Fatalf("unexpected error: %v", te)
	}

	// Verify old file gone, new file exists
	if _, err := os.Stat(filepath.Join(root, "entities/old-name.md")); !os.IsNotExist(err) {
		t.Error("old file should not exist")
	}
	newData, err := os.ReadFile(filepath.Join(root, "entities/new-name.md"))
	if err != nil {
		t.Fatal("new file should exist")
	}

	// Verify outgoing links in moved page still work
	if !strings.Contains(string(newData), "../concepts/foo.md") {
		t.Error("outgoing link should still point to concepts/foo.md")
	}

	// Verify incoming [[Title]] link rewritten
	fooData, _ := os.ReadFile(filepath.Join(root, "concepts/foo.md"))
	if !strings.Contains(string(fooData), "[[New Name]]") {
		t.Errorf("wiki link not rewritten in foo.md: %s", fooData)
	}

	// Verify incoming [text](path) link rewritten
	if !strings.Contains(string(fooData), "../entities/new-name.md") {
		t.Errorf("markdown link not rewritten in foo.md: %s", fooData)
	}

	// Verify index link rewritten
	idxData, _ := os.ReadFile(filepath.Join(root, "index.md"))
	if !strings.Contains(string(idxData), "entities/new-name.md") {
		t.Errorf("index link not rewritten: %s", idxData)
	}
}

func TestPageMove_CrossDirectory(t *testing.T) {
	root := t.TempDir()

	seedPage(t, root, "entities/page.md", "# Page\n\nSee [Concept](../concepts/idea.md)\n")
	seedPage(t, root, "concepts/idea.md", "# Idea\n\nSee [Page](../entities/page.md)\n")

	cfg := testConfig(t, root)

	te := PageMove(cfg, "entities/page.md", "archive/old/page.md")
	if te != nil {
		t.Fatal(te)
	}

	// Moved page's outgoing links should be adjusted
	movedData, _ := os.ReadFile(filepath.Join(root, "archive/old/page.md"))
	if !strings.Contains(string(movedData), "../../concepts/idea.md") {
		t.Errorf("outgoing link not adjusted: %s", movedData)
	}

	// Incoming link should point to new location
	ideaData, _ := os.ReadFile(filepath.Join(root, "concepts/idea.md"))
	if !strings.Contains(string(ideaData), "../archive/old/page.md") {
		t.Errorf("incoming link not rewritten: %s", ideaData)
	}
}

func TestPageMove_NotFound(t *testing.T) {
	root := t.TempDir()
	cfg := testConfig(t, root)

	te := PageMove(cfg, "ghost.md", "new.md")
	if te == nil {
		t.Fatal("expected error")
	}
	if te.Code != ErrCodeNotFound {
		t.Errorf("expected NotFound, got %q", te.Code)
	}
}

// --- Error codes ---

func TestToolError_JSON(t *testing.T) {
	te := NewToolError(ErrCodeNotFound, "page not found")
	j := te.JSON()
	if !strings.Contains(j, `"code":"wiki.NotFound"`) {
		t.Errorf("unexpected JSON: %s", j)
	}
	if !strings.Contains(j, `"message":"page not found"`) {
		t.Errorf("unexpected JSON: %s", j)
	}
}

func TestTitleFromPath(t *testing.T) {
	tests := []struct {
		path string
		want string
	}{
		{"entities/qwen2.5-coder.md", "Qwen2.5 Coder"},
		{"concepts/llm-quantization.md", "Llm Quantization"},
		{"summary.md", "Summary"},
	}
	for _, tt := range tests {
		got := TitleFromPath(tt.path)
		if got != tt.want {
			t.Errorf("TitleFromPath(%q) = %q, want %q", tt.path, got, tt.want)
		}
	}
}
