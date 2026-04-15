package wiki

import (
	"strings"
	"testing"
)

func TestWikiSearch_BasicMatch(t *testing.T) {
	root := t.TempDir()
	seedPage(t, root, "page1.md", "# Page One\nThis is about quantization.\n")
	seedPage(t, root, "page2.md", "# Page Two\nNothing related here.\n")
	cfg := testConfig(t, root)

	results, te := WikiSearch(cfg, "quantization", 20)
	if te != nil {
		t.Fatal(te)
	}

	if len(results) != 1 {
		t.Fatalf("got %d results, want 1", len(results))
	}
	if results[0].Path != "page1.md" {
		t.Errorf("path = %q, want page1.md", results[0].Path)
	}
	if results[0].Score < 1 {
		t.Errorf("score = %d, want ≥1", results[0].Score)
	}
	if len(results[0].Snippets) == 0 {
		t.Error("expected at least one snippet")
	}
}

func TestWikiSearch_CaseInsensitive(t *testing.T) {
	root := t.TempDir()
	seedPage(t, root, "page.md", "# Title\nSWE-Bench score is great.\n")
	cfg := testConfig(t, root)

	results, te := WikiSearch(cfg, "swe-bench", 20)
	if te != nil {
		t.Fatal(te)
	}

	if len(results) != 1 {
		t.Fatalf("case-insensitive search: got %d results, want 1", len(results))
	}
}

func TestWikiSearch_ScoredAndOrdered(t *testing.T) {
	root := t.TempDir()
	// page1 has 3 matches, page2 has 1
	seedPage(t, root, "page1.md", "ollama ollama ollama\n")
	seedPage(t, root, "page2.md", "only one ollama here\n")
	cfg := testConfig(t, root)

	results, te := WikiSearch(cfg, "ollama", 20)
	if te != nil {
		t.Fatal(te)
	}

	if len(results) < 2 {
		t.Fatalf("got %d results, want ≥2", len(results))
	}
	if results[0].Score < results[1].Score {
		t.Errorf("results not ordered by score: %d < %d", results[0].Score, results[1].Score)
	}
}

func TestWikiSearch_LimitRespected(t *testing.T) {
	root := t.TempDir()
	for i := 0; i < 5; i++ {
		seedPage(t, root, strings.Repeat("a", i+1)+".md", "keyword here\n")
	}
	cfg := testConfig(t, root)

	results, te := WikiSearch(cfg, "keyword", 3)
	if te != nil {
		t.Fatal(te)
	}

	if len(results) > 3 {
		t.Errorf("got %d results, want ≤3", len(results))
	}
}

func TestWikiSearch_FrontmatterExcluded(t *testing.T) {
	root := t.TempDir()
	// "secret" only in frontmatter — should NOT match
	seedPage(t, root, "page.md", "---\ntags: [secret]\n---\n# Normal body\n")
	cfg := testConfig(t, root)

	results, te := WikiSearch(cfg, "secret", 20)
	if te != nil {
		t.Fatal(te)
	}

	if len(results) != 0 {
		t.Errorf("frontmatter match should be excluded, got %v", results)
	}
}

func TestWikiSearch_ExistingWiki(t *testing.T) {
	cfg := existingWikiConfig(t)

	results, te := WikiSearch(cfg, "Qwen", 20)
	if te != nil {
		t.Fatal(te)
	}

	if len(results) == 0 {
		t.Error("expected results for 'Qwen' in existing wiki")
	}
	// All results should have score ≥ 1 and at least one snippet
	for _, r := range results {
		if r.Score < 1 {
			t.Errorf("result %q has score 0", r.Path)
		}
		if len(r.Snippets) == 0 {
			t.Errorf("result %q has no snippets", r.Path)
		}
	}
}

func TestWikiSearch_EmptyQuery(t *testing.T) {
	root := t.TempDir()
	cfg := testConfig(t, root)

	_, te := WikiSearch(cfg, "", 20)
	if te == nil || te.Code != ErrCodeBadRequest {
		t.Errorf("expected BadRequest for empty query, got %v", te)
	}
}
