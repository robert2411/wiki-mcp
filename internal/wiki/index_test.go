package wiki

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/robertstevens/wiki-mcp/internal/config"
)

func indexTestConfig(t *testing.T, wikiPath string) *config.Config {
	t.Helper()
	cfg := config.Defaults()
	cfg.WikiPath = wikiPath
	cfg.Safety.ConfineToWikiPath = true
	cfg.Safety.MaxPageBytes = 1048576
	return &cfg
}

func TestParseIndex_RoundTrip(t *testing.T) {
	// Use the real existing wiki index.md
	src := filepath.Join("..", "..", "existing", "wiki", "index.md")
	data, err := os.ReadFile(src)
	if err != nil {
		t.Fatalf("cannot read existing index.md: %v", err)
	}

	cfg := indexTestConfig(t, filepath.Join("..", "..", "existing", "wiki"))
	doc, parseErr := ParseIndex(data, cfg)
	if parseErr != nil {
		t.Fatalf("ParseIndex failed: %v", parseErr)
	}

	rendered := RenderIndex(doc)

	if string(rendered) != string(data) {
		// Find first difference
		a, b := string(data), string(rendered)
		aLines := strings.Split(a, "\n")
		bLines := strings.Split(b, "\n")
		for i := 0; i < len(aLines) || i < len(bLines); i++ {
			var al, bl string
			if i < len(aLines) {
				al = aLines[i]
			}
			if i < len(bLines) {
				bl = bLines[i]
			}
			if al != bl {
				t.Errorf("first diff at line %d:\n  want: %q\n  got:  %q", i+1, al, bl)
				break
			}
		}
		t.Fatalf("round-trip failed: lengths want=%d got=%d", len(data), len(rendered))
	}
}

func TestParseIndex_Sections(t *testing.T) {
	src := filepath.Join("..", "..", "existing", "wiki", "index.md")
	data, err := os.ReadFile(src)
	if err != nil {
		t.Fatalf("cannot read existing index.md: %v", err)
	}

	cfg := indexTestConfig(t, filepath.Join("..", "..", "existing", "wiki"))
	doc, parseErr := ParseIndex(data, cfg)
	if parseErr != nil {
		t.Fatalf("ParseIndex failed: %v", parseErr)
	}

	if len(doc.Sections) != 4 {
		t.Fatalf("expected 4 sections, got %d", len(doc.Sections))
	}

	expectedKeys := []string{"research", "entities", "concepts", "infrastructure"}
	for i, key := range expectedKeys {
		if doc.Sections[i].Key != key {
			t.Errorf("section %d: want key %q, got %q", i, key, doc.Sections[i].Key)
		}
	}

	// Research should have 8 entries
	if len(doc.Sections[0].Entries) != 8 {
		t.Errorf("research entries: want 8, got %d", len(doc.Sections[0].Entries))
	}

	// Stats should be parsed
	if doc.Stats.WikiPages != "22" {
		t.Errorf("wiki pages: want 22, got %q", doc.Stats.WikiPages)
	}
}

func TestIndexUpsertEntry_UpdateExisting(t *testing.T) {
	dir := t.TempDir()
	src := filepath.Join("..", "..", "existing", "wiki", "index.md")
	data, err := os.ReadFile(src)
	if err != nil {
		t.Fatalf("cannot read existing index.md: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dir, "index.md"), data, 0o644); err != nil {
		t.Fatal(err)
	}

	cfg := indexTestConfig(t, dir)

	te := IndexUpsertEntry(cfg, "entities", "Ollama", "entities/ollama.md", "Updated summary for Ollama")
	if te != nil {
		t.Fatalf("upsert failed: %v", te)
	}

	// Re-read and check
	doc, te := IndexRead(cfg)
	if te != nil {
		t.Fatalf("read failed: %v", te)
	}

	for _, sec := range doc.Sections {
		if sec.Key == "entities" {
			for _, e := range sec.Entries {
				if e.Title == "Ollama" && e.Path == "entities/ollama.md" {
					if e.Summary != "Updated summary for Ollama" {
						t.Errorf("summary not updated: got %q", e.Summary)
					}
					return
				}
			}
			t.Error("Ollama entry not found after upsert")
			return
		}
	}
	t.Error("entities section not found")
}

func TestIndexUpsertEntry_AppendNew(t *testing.T) {
	dir := t.TempDir()
	src := filepath.Join("..", "..", "existing", "wiki", "index.md")
	data, err := os.ReadFile(src)
	if err != nil {
		t.Fatalf("cannot read existing index.md: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dir, "index.md"), data, 0o644); err != nil {
		t.Fatal(err)
	}

	cfg := indexTestConfig(t, dir)

	te := IndexUpsertEntry(cfg, "concepts", "New Concept", "concepts/new-concept.md", "A brand new concept")
	if te != nil {
		t.Fatalf("upsert failed: %v", te)
	}

	doc, te := IndexRead(cfg)
	if te != nil {
		t.Fatalf("read failed: %v", te)
	}

	for _, sec := range doc.Sections {
		if sec.Key == "concepts" {
			if len(sec.Entries) != 3 {
				t.Errorf("concepts entries: want 3, got %d", len(sec.Entries))
			}
			last := sec.Entries[len(sec.Entries)-1]
			if last.Title != "New Concept" {
				t.Errorf("last entry title: want %q, got %q", "New Concept", last.Title)
			}
			return
		}
	}
	t.Error("concepts section not found")
}

func TestIndexUpsertEntry_NewSection(t *testing.T) {
	// Create a minimal index with only research section
	dir := t.TempDir()
	minimal := `# Wiki Index

## Pages

### 🔬 Research
- [Test](test.md) — A test page

---

## Stats

- Wiki pages: 1
- Last updated: 2026-04-14
`
	if err := os.WriteFile(filepath.Join(dir, "index.md"), []byte(minimal), 0o644); err != nil {
		t.Fatal(err)
	}

	cfg := indexTestConfig(t, dir)

	// Upsert into concepts — should create section between research and infrastructure
	te := IndexUpsertEntry(cfg, "concepts", "Test Concept", "concepts/test.md", "A test concept")
	if te != nil {
		t.Fatalf("upsert failed: %v", te)
	}

	doc, te := IndexRead(cfg)
	if te != nil {
		t.Fatalf("read failed: %v", te)
	}

	if len(doc.Sections) != 2 {
		t.Fatalf("expected 2 sections, got %d", len(doc.Sections))
	}

	// Concepts should come after research (config order: research=0, entities=1, concepts=2, infra=3)
	if doc.Sections[0].Key != "research" {
		t.Errorf("section 0: want research, got %q", doc.Sections[0].Key)
	}
	if doc.Sections[1].Key != "concepts" {
		t.Errorf("section 1: want concepts, got %q", doc.Sections[1].Key)
	}
}

func TestIndexRefreshStats(t *testing.T) {
	dir := t.TempDir()
	src := filepath.Join("..", "..", "existing", "wiki", "index.md")
	data, err := os.ReadFile(src)
	if err != nil {
		t.Fatalf("cannot read existing index.md: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dir, "index.md"), data, 0o644); err != nil {
		t.Fatal(err)
	}

	// Create a few test pages
	os.MkdirAll(filepath.Join(dir, "entities"), 0o755)
	os.WriteFile(filepath.Join(dir, "entities", "test1.md"), []byte("# Test 1"), 0o644)
	os.WriteFile(filepath.Join(dir, "entities", "test2.md"), []byte("# Test 2"), 0o644)

	cfg := indexTestConfig(t, dir)

	// Capture pages list before
	docBefore, te := IndexRead(cfg)
	if te != nil {
		t.Fatalf("read failed: %v", te)
	}
	entriesBefore := make([][]IndexEntry, len(docBefore.Sections))
	for i, sec := range docBefore.Sections {
		entriesBefore[i] = append([]IndexEntry{}, sec.Entries...)
	}

	te = IndexRefreshStats(cfg)
	if te != nil {
		t.Fatalf("refresh stats failed: %v", te)
	}

	docAfter, te := IndexRead(cfg)
	if te != nil {
		t.Fatalf("read failed: %v", te)
	}

	// Pages list should be unchanged
	for i, sec := range docAfter.Sections {
		if len(sec.Entries) != len(entriesBefore[i]) {
			t.Errorf("section %q entries changed: before=%d after=%d", sec.Key, len(entriesBefore[i]), len(sec.Entries))
		}
	}

	// Stats should be updated
	if docAfter.Stats.WikiPages != "2" {
		t.Errorf("wiki pages: want 2, got %q", docAfter.Stats.WikiPages)
	}

	// Sources ingested should be preserved
	if docAfter.Stats.SourcesIngested != docBefore.Stats.SourcesIngested {
		t.Errorf("sources ingested changed: was %q, now %q", docBefore.Stats.SourcesIngested, docAfter.Stats.SourcesIngested)
	}
}

func TestIndexUpsertEntry_ReadOnly(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "index.md"), []byte("# Index\n## Pages\n## Stats\n"), 0o644)

	cfg := indexTestConfig(t, dir)
	cfg.Safety.ReadOnly = true

	te := IndexUpsertEntry(cfg, "research", "Test", "test.md", "summary")
	if te == nil {
		t.Fatal("expected read-only error")
	}
	if te.Code != ErrCodeReadOnly {
		t.Errorf("want code %q, got %q", ErrCodeReadOnly, te.Code)
	}
}

func TestIndexRefreshStats_ReadOnly(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "index.md"), []byte("# Index\n## Pages\n## Stats\n"), 0o644)

	cfg := indexTestConfig(t, dir)
	cfg.Safety.ReadOnly = true

	te := IndexRefreshStats(cfg)
	if te == nil {
		t.Fatal("expected read-only error")
	}
	if te.Code != ErrCodeReadOnly {
		t.Errorf("want code %q, got %q", ErrCodeReadOnly, te.Code)
	}
}
