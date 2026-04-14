package render

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/robertstevens/wiki-mcp/internal/config"
)

const existingWiki = "../../../existing/wiki"

// resolveWiki returns the absolute path to the test wiki fixture.
func resolveWiki(t *testing.T) string {
	t.Helper()
	abs, err := filepath.Abs(existingWiki)
	if err != nil {
		t.Fatalf("resolve wiki: %v", err)
	}
	if _, err := os.Stat(abs); err != nil {
		t.Skipf("existing/wiki not present: %v", err)
	}
	return abs
}

// AC1: RenderPage returns HTML + title + metadata for all pages without errors.
func TestRenderPage_AllExistingPages(t *testing.T) {
	wikiPath := resolveWiki(t)
	c := &config.Config{
		WikiPath: wikiPath,
		Safety:   config.SafetyConfig{ConfineToWikiPath: true},
	}

	entries, err := filepath.Glob(filepath.Join(wikiPath, "**/*.md"))
	if err != nil {
		t.Fatalf("glob: %v", err)
	}
	// Also pick up top-level .md files.
	top, _ := filepath.Glob(filepath.Join(wikiPath, "*.md"))
	entries = append(entries, top...)

	if len(entries) == 0 {
		// Walk if glob returns nothing (OS glob may not expand **).
		err = filepath.WalkDir(wikiPath, func(path string, d os.DirEntry, e error) error {
			if e != nil || d.IsDir() || !strings.HasSuffix(path, ".md") {
				return e
			}
			entries = append(entries, path)
			return nil
		})
		if err != nil {
			t.Fatalf("walk: %v", err)
		}
	}

	if len(entries) == 0 {
		t.Fatal("no .md pages found")
	}

	t.Logf("rendering %d pages", len(entries))
	for _, abs := range entries {
		rel, _ := filepath.Rel(wikiPath, abs)
		page, err := RenderPage(c, rel)
		if err != nil {
			t.Errorf("RenderPage(%s): %v", rel, err)
			continue
		}
		if page.HTML == "" {
			t.Errorf("RenderPage(%s): empty HTML", rel)
		}
		if page.Title == "" {
			t.Errorf("RenderPage(%s): empty title", rel)
		}
	}
}

// AC2: [[Qwen3.5]] resolves to href "entities/qwen3.5".
func TestWikilink_Resolved(t *testing.T) {
	wikiPath := resolveWiki(t)
	idx, err := BuildTitleIndex(wikiPath)
	if err != nil {
		t.Fatalf("BuildTitleIndex: %v", err)
	}

	rdr := newRendererWithIndex(idx)
	md := []byte("See [[Qwen3.5]] for details.")
	page, err := rdr.renderBytes(md, "test.md")
	if err != nil {
		t.Fatalf("renderBytes: %v", err)
	}

	want := `href="/entities/qwen3.5"`
	if !strings.Contains(page.HTML, want) {
		t.Errorf("expected %q in HTML\ngot: %s", want, page.HTML)
	}
}

// AC3: [[Missing Title]] renders with class "broken-link".
func TestWikilink_Broken(t *testing.T) {
	rdr := newRendererWithIndex(map[string]string{})
	md := []byte("See [[Missing Title]] for details.")
	page, err := rdr.renderBytes(md, "test.md")
	if err != nil {
		t.Fatalf("renderBytes: %v", err)
	}

	if !strings.Contains(page.HTML, `class="broken-link"`) {
		t.Errorf("expected broken-link class in HTML\ngot: %s", page.HTML)
	}
	// Must contain the text.
	if !strings.Contains(page.HTML, "Missing Title") {
		t.Errorf("expected display text in HTML\ngot: %s", page.HTML)
	}
}

// AC4: Frontmatter renders as metadata block, not raw YAML.
func TestFrontmatter_RenderedAsBlock(t *testing.T) {
	rdr := newRendererWithIndex(map[string]string{})
	md := []byte("---\ntags: [go, wiki]\nupdated: 2026-04-15\n---\n\n# My Page\n\nContent here.\n")
	page, err := rdr.renderBytes(md, "test.md")
	if err != nil {
		t.Fatalf("renderBytes: %v", err)
	}

	if strings.Contains(page.HTML, "---") {
		t.Errorf("raw YAML delimiters found in HTML output\ngot: %s", page.HTML)
	}
	if !strings.Contains(page.HTML, `class="frontmatter"`) {
		t.Errorf("expected frontmatter div in HTML\ngot: %s", page.HTML)
	}
	if !strings.Contains(page.HTML, "go") {
		t.Errorf("expected tag value in HTML\ngot: %s", page.HTML)
	}
	// Metadata map populated.
	if page.Metadata["updated"] == nil {
		t.Errorf("expected 'updated' in metadata, got: %v", page.Metadata)
	}
}

// AC5: Tables, fenced code, task lists render on summary.md.
func TestGFM_SummaryPage(t *testing.T) {
	wikiPath := resolveWiki(t)
	c := &config.Config{
		WikiPath: wikiPath,
		Safety:   config.SafetyConfig{ConfineToWikiPath: true},
	}

	page, err := RenderPage(c, "ollama-java-code-review/summary.md")
	if err != nil {
		t.Fatalf("RenderPage summary.md: %v", err)
	}

	checks := []struct {
		want string
		desc string
	}{
		{"<table>", "table element"},
		{"<code>", "code element"},
	}
	for _, ch := range checks {
		if !strings.Contains(page.HTML, ch.want) {
			t.Errorf("expected %s (%q) in HTML", ch.desc, ch.want)
		}
	}
}

// TestMdLinkStripping ensures .md is stripped from standard markdown links.
func TestMdLinkStripping(t *testing.T) {
	rdr := newRendererWithIndex(map[string]string{})
	md := []byte("[See page](entities/qwen3.5.md)")
	page, err := rdr.renderBytes(md, "test.md")
	if err != nil {
		t.Fatalf("renderBytes: %v", err)
	}

	if strings.Contains(page.HTML, ".md") {
		t.Errorf(".md not stripped from link href\ngot: %s", page.HTML)
	}
	if !strings.Contains(page.HTML, `href="entities/qwen3.5"`) {
		t.Errorf("expected stripped href\ngot: %s", page.HTML)
	}
}

// TestWikilinkAlias ensures [[Title|alias]] uses alias as display text.
func TestWikilinkAlias(t *testing.T) {
	wikiPath := resolveWiki(t)
	idx, err := BuildTitleIndex(wikiPath)
	if err != nil {
		t.Fatalf("BuildTitleIndex: %v", err)
	}

	rdr := newRendererWithIndex(idx)
	md := []byte("See [[Qwen3.5|the Qwen model]] for details.")
	page, err := rdr.renderBytes(md, "test.md")
	if err != nil {
		t.Fatalf("renderBytes: %v", err)
	}

	if !strings.Contains(page.HTML, "the Qwen model") {
		t.Errorf("expected alias text in HTML\ngot: %s", page.HTML)
	}
	if !strings.Contains(page.HTML, `href="/entities/qwen3.5"`) {
		t.Errorf("expected resolved href\ngot: %s", page.HTML)
	}
}
