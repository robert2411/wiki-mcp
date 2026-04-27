package wiki

import (
	"testing"

	"github.com/robert2411/wiki-mcp/internal/config"
)

func existingWikiConfig(t *testing.T) *config.Config {
	t.Helper()
	p := existingWikiPath(t)
	return &config.Config{WikiPath: p, ProjectPath: p}
}

// --- LinksOutgoing ---

func TestLinksOutgoing_WikiLinks(t *testing.T) {
	root := t.TempDir()
	seedPage(t, root, "page.md", "See [[Target Title]] and [[Other|alias]].\n")
	cfg := testConfig(t, root)

	result, te := LinksOutgoing(cfg, "page.md")
	if te != nil {
		t.Fatal(te)
	}

	if len(result.Internal) != 2 {
		t.Fatalf("Internal len = %d, want 2: %v", len(result.Internal), result.Internal)
	}
	if result.Internal[0] != "Target Title" {
		t.Errorf("Internal[0] = %q, want %q", result.Internal[0], "Target Title")
	}
	if result.Internal[1] != "Other" {
		t.Errorf("Internal[1] = %q, want %q", result.Internal[1], "Other")
	}
}

func TestLinksOutgoing_MDLinks(t *testing.T) {
	root := t.TempDir()
	seedPage(t, root, "concepts/page.md", "[Coder](../entities/qwen2.5-coder.md) and [ext](https://example.com).\n")
	cfg := testConfig(t, root)

	result, te := LinksOutgoing(cfg, "concepts/page.md")
	if te != nil {
		t.Fatal(te)
	}

	if len(result.Internal) != 1 || result.Internal[0] != "entities/qwen2.5-coder.md" {
		t.Errorf("Internal = %v, want [entities/qwen2.5-coder.md]", result.Internal)
	}
	if len(result.External) != 1 {
		t.Errorf("External len = %d, want 1", len(result.External))
	}
}

func TestLinksOutgoing_NotFound(t *testing.T) {
	root := t.TempDir()
	cfg := testConfig(t, root)

	_, te := LinksOutgoing(cfg, "ghost.md")
	if te == nil || te.Code != ErrCodeNotFound {
		t.Errorf("expected NotFound, got %v", te)
	}
}

// --- LinksIncoming ---

func TestLinksIncoming_KnownPage(t *testing.T) {
	root := t.TempDir()
	seedPage(t, root, "target.md", "# Target\n")
	seedPage(t, root, "a.md", "[Link](target.md)\n")
	seedPage(t, root, "b.md", "[[Target]]\n")
	seedPage(t, root, "c.md", "no link here\n")
	cfg := testConfig(t, root)

	backlinks, te := LinksIncoming(cfg, "target.md")
	if te != nil {
		t.Fatal(te)
	}

	if len(backlinks) != 2 {
		t.Fatalf("backlinks = %v, want [a.md b.md]", backlinks)
	}
}

func TestLinksIncoming_ExistingWiki(t *testing.T) {
	cfg := existingWikiConfig(t)

	// qwen2.5-coder.md is heavily linked
	backlinks, te := LinksIncoming(cfg, "entities/qwen2.5-coder.md")
	if te != nil {
		t.Fatal(te)
	}

	if len(backlinks) < 5 {
		t.Errorf("expected ≥5 backlinks for qwen2.5-coder.md, got %d: %v", len(backlinks), backlinks)
	}
}

// --- Orphans ---

func TestOrphans_SimpleFixture(t *testing.T) {
	root := t.TempDir()
	seedPage(t, root, "index.md", "[A](a.md)\n") // index links to a.md
	seedPage(t, root, "a.md", "# A\n")           // linked only from index → orphan
	seedPage(t, root, "b.md", "[A](a.md)\n")     // b links to a
	seedPage(t, root, "c.md", "# Standalone\n")  // no incoming links → orphan
	cfg := testConfig(t, root)

	orphans, te := Orphans(cfg)
	if te != nil {
		t.Fatal(te)
	}

	// a.md is linked only from index.md (excluded source) → orphan
	// b.md is linked by nobody → orphan
	// c.md is linked by nobody → orphan
	// a.md is linked by b.md → NOT orphan
	// Let's recalculate: b.md links to a.md, so a.md has an incoming link from b.md
	// So orphans should be: b.md and c.md
	wantOrphans := map[string]bool{"b.md": true, "c.md": true}
	for _, o := range orphans {
		if !wantOrphans[o] {
			t.Errorf("unexpected orphan: %q", o)
		}
	}
	if len(orphans) != len(wantOrphans) {
		t.Errorf("orphans = %v, want %v", orphans, wantOrphans)
	}
}

func TestOrphans_ExistingWiki(t *testing.T) {
	cfg := existingWikiConfig(t)

	orphans, te := Orphans(cfg)
	if te != nil {
		t.Fatal(te)
	}

	// Must contain the 2 known orphans from lint pass 1
	want := map[string]bool{
		"concepts/spring-boot-maven-docker.md": false,
		"infrastructure/wiki-ui.md":            false,
	}
	for _, o := range orphans {
		if _, known := want[o]; known {
			want[o] = true
		}
	}
	for path, found := range want {
		if !found {
			t.Errorf("expected orphan %q not found in: %v", path, orphans)
		}
	}
}
