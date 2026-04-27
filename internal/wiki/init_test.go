package wiki

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/robert2411/wiki-mcp/internal/config"
)

func twoSectionCfg(t *testing.T, root string) *config.Config {
	t.Helper()
	cfg := testConfig(t, root)
	cfg.Index = config.IndexConfig{
		Sections: []config.IndexSection{
			{Key: "research", Title: "🔬 Research"},
			{Key: "entities", Title: "🏷️ Entities"},
		},
	}
	return cfg
}

func TestWikiInit_FreshDirectory(t *testing.T) {
	root := filepath.Join(t.TempDir(), "wiki")
	cfg := twoSectionCfg(t, root)

	result, te := WikiInit(cfg)
	if te != nil {
		t.Fatalf("unexpected error: %v", te)
	}

	if _, err := os.Stat(root); err != nil {
		t.Error("wiki directory should exist")
	}
	for _, key := range []string{"research", "entities"} {
		if _, err := os.Stat(filepath.Join(root, key)); err != nil {
			t.Errorf("section dir %q should exist", key)
		}
	}
	if _, err := os.Stat(filepath.Join(root, "index.md")); err != nil {
		t.Error("index.md should exist")
	}
	if _, err := os.Stat(filepath.Join(root, "log.md")); err != nil {
		t.Error("log.md should exist")
	}
	if len(result.Created) == 0 {
		t.Error("expected created items")
	}
	if result.WikiPath != root {
		t.Errorf("wiki_path: got %q, want %q", result.WikiPath, root)
	}
}

func TestWikiInit_IndexIsParseable(t *testing.T) {
	root := filepath.Join(t.TempDir(), "wiki")
	cfg := twoSectionCfg(t, root)

	if _, te := WikiInit(cfg); te != nil {
		t.Fatal(te)
	}

	data, err := os.ReadFile(filepath.Join(root, "index.md"))
	if err != nil {
		t.Fatal(err)
	}
	doc, err := ParseIndex(data, cfg)
	if err != nil {
		t.Fatalf("index.md not parseable: %v", err)
	}
	if len(doc.Sections) != 2 {
		t.Errorf("expected 2 sections, got %d", len(doc.Sections))
	}
	if doc.Sections[0].Key != "research" {
		t.Errorf("first section key: got %q, want %q", doc.Sections[0].Key, "research")
	}
	if doc.Stats.WikiPages != "0" {
		t.Errorf("WikiPages: got %q, want %q", doc.Stats.WikiPages, "0")
	}
}

func TestWikiInit_NoOverwrite(t *testing.T) {
	root := t.TempDir()
	cfg := twoSectionCfg(t, root)
	seedPage(t, root, "index.md", "# My Custom Index\n")
	seedPage(t, root, "log.md", "# My Custom Log\n")

	result, te := WikiInit(cfg)
	if te != nil {
		t.Fatal(te)
	}

	data, _ := os.ReadFile(filepath.Join(root, "index.md"))
	if !strings.Contains(string(data), "My Custom Index") {
		t.Error("existing index.md should not be overwritten")
	}
	logData, _ := os.ReadFile(filepath.Join(root, "log.md"))
	if !strings.Contains(string(logData), "My Custom Log") {
		t.Error("existing log.md should not be overwritten")
	}

	skipped := strings.Join(result.Skipped, ",")
	if !strings.Contains(skipped, "index.md") {
		t.Errorf("index.md should be in skipped, got: %v", result.Skipped)
	}
	if !strings.Contains(skipped, "log.md") {
		t.Errorf("log.md should be in skipped, got: %v", result.Skipped)
	}
}

func TestWikiInit_Idempotent(t *testing.T) {
	root := filepath.Join(t.TempDir(), "wiki")
	cfg := twoSectionCfg(t, root)

	if _, te := WikiInit(cfg); te != nil {
		t.Fatal(te)
	}
	result2, te := WikiInit(cfg)
	if te != nil {
		t.Fatalf("second init failed: %v", te)
	}
	if len(result2.Created) != 0 {
		t.Errorf("second init should create nothing, got: %v", result2.Created)
	}
	if len(result2.Skipped) == 0 {
		t.Error("second init should report skipped items")
	}
}

func TestWikiInit_ReadOnly(t *testing.T) {
	root := t.TempDir()
	cfg := twoSectionCfg(t, root)
	cfg.Safety.ReadOnly = true

	_, te := WikiInit(cfg)
	if te == nil {
		t.Fatal("expected error")
	}
	if te.Code != ErrCodeReadOnly {
		t.Errorf("expected ReadOnly, got %q", te.Code)
	}
}

func TestProjectList_Empty(t *testing.T) {
	root := t.TempDir()
	cfg := testConfig(t, root)

	projects, te := ProjectList(cfg)
	if te != nil {
		t.Fatalf("unexpected error: %v", te)
	}
	if len(projects) != 0 {
		t.Errorf("expected 0 projects, got %d", len(projects))
	}
}

func TestProjectList_SubdirsWithIndex(t *testing.T) {
	root := t.TempDir()
	cfg := testConfig(t, root)

	// Create two project dirs with index.md and one without
	for _, name := range []string{"alpha", "beta"} {
		dir := filepath.Join(root, name)
		if err := os.MkdirAll(dir, 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(dir, "index.md"), []byte("# "+name), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	if err := os.MkdirAll(filepath.Join(root, "noindex"), 0o755); err != nil {
		t.Fatal(err)
	}

	projects, te := ProjectList(cfg)
	if te != nil {
		t.Fatalf("unexpected error: %v", te)
	}
	if len(projects) != 2 {
		t.Errorf("expected 2 projects, got %d: %v", len(projects), projects)
	}
	names := make(map[string]bool)
	for _, p := range projects {
		names[p.Name] = true
	}
	for _, want := range []string{"alpha", "beta"} {
		if !names[want] {
			t.Errorf("expected project %q in list", want)
		}
	}
	if names["noindex"] {
		t.Error("noindex should not appear (no index.md)")
	}
}

func TestWikiInit_ScopedToProject(t *testing.T) {
	root := filepath.Join(t.TempDir(), "wiki")
	if err := os.MkdirAll(root, 0o755); err != nil {
		t.Fatal(err)
	}
	cfg := twoSectionCfg(t, root)
	projectDir := filepath.Join(root, "myproject")
	cfg.ProjectPath = projectDir

	result, te := WikiInit(cfg)
	if te != nil {
		t.Fatalf("unexpected error: %v", te)
	}
	if result.WikiPath != projectDir {
		t.Errorf("WikiPath: got %q, want %q", result.WikiPath, projectDir)
	}
	if _, err := os.Stat(filepath.Join(projectDir, "index.md")); err != nil {
		t.Error("index.md should exist inside project dir")
	}
}
