package wiki

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/robertstevens/wiki-mcp/internal/config"
)

func logTestConfig(t *testing.T, wikiPath string) *config.Config {
	t.Helper()
	cfg := config.Defaults()
	cfg.WikiPath = wikiPath
	cfg.Safety.ConfineToWikiPath = true
	cfg.Safety.MaxPageBytes = 1048576
	return &cfg
}

func TestLogTail_ExistingLog(t *testing.T) {
	src := filepath.Join("..", "..", "existing", "wiki", "log.md")
	data, err := os.ReadFile(src)
	if err != nil {
		t.Fatalf("cannot read existing log.md: %v", err)
	}

	cfg := logTestConfig(t, filepath.Join("..", "..", "existing", "wiki"))
	_ = data

	entries, te := LogTail(cfg, 10)
	if te != nil {
		t.Fatalf("LogTail failed: %v", te)
	}

	// existing/wiki/log.md has 13 entries
	if len(entries) != 10 {
		t.Errorf("expected 10 entries (tail of 13), got %d", len(entries))
	}

	// Last entry is the lint pass
	last := entries[len(entries)-1]
	if last.Operation != "lint" {
		t.Errorf("last entry operation: want lint, got %q", last.Operation)
	}
	if last.Date != "2026-04-11" {
		t.Errorf("last entry date: want 2026-04-11, got %q", last.Date)
	}
	if !strings.Contains(last.Title, "pass 1") {
		t.Errorf("last entry title: want 'pass 1', got %q", last.Title)
	}
}

func TestLogTail_AllEntries(t *testing.T) {
	cfg := logTestConfig(t, filepath.Join("..", "..", "existing", "wiki"))

	entries, te := LogTail(cfg, 100)
	if te != nil {
		t.Fatalf("LogTail failed: %v", te)
	}

	if len(entries) != 13 {
		t.Errorf("expected 13 total entries, got %d", len(entries))
	}

	// First entry
	first := entries[0]
	if first.Operation != "query" {
		t.Errorf("first entry operation: want query, got %q", first.Operation)
	}
	if first.Date != "2026-04-06" {
		t.Errorf("first entry date: want 2026-04-06, got %q", first.Date)
	}
}

func TestLogTail_MissingFile(t *testing.T) {
	dir := t.TempDir()
	cfg := logTestConfig(t, dir)

	entries, te := LogTail(cfg, 10)
	if te != nil {
		t.Fatalf("LogTail on missing file should not error: %v", te)
	}
	if len(entries) != 0 {
		t.Errorf("expected 0 entries from missing file, got %d", len(entries))
	}
}

func TestLogAppend_Basic(t *testing.T) {
	dir := t.TempDir()
	src := filepath.Join("..", "..", "existing", "wiki", "log.md")
	data, err := os.ReadFile(src)
	if err != nil {
		t.Fatalf("cannot read existing log.md: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dir, "log.md"), data, 0o644); err != nil {
		t.Fatal(err)
	}

	cfg := logTestConfig(t, dir)

	te := LogAppend(cfg, "ingest", "Test Source", "Some body text here.")
	if te != nil {
		t.Fatalf("LogAppend failed: %v", te)
	}

	entries, te := LogTail(cfg, 100)
	if te != nil {
		t.Fatalf("LogTail failed: %v", te)
	}

	if len(entries) != 14 {
		t.Errorf("expected 14 entries after append, got %d", len(entries))
	}

	last := entries[len(entries)-1]
	if last.Operation != "ingest" {
		t.Errorf("last entry operation: want ingest, got %q", last.Operation)
	}
	if last.Title != "Test Source" {
		t.Errorf("last entry title: want %q, got %q", "Test Source", last.Title)
	}
	if last.Body != "Some body text here." {
		t.Errorf("last entry body: want %q, got %q", "Some body text here.", last.Body)
	}
}

func TestLogAppend_InvalidOperation(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "log.md"), []byte(logTemplate), 0o644)
	cfg := logTestConfig(t, dir)

	te := LogAppend(cfg, "delete", "Bad op", "")
	if te == nil {
		t.Fatal("expected error for invalid operation")
	}
	if te.Code != ErrCodeBadRequest {
		t.Errorf("want code %q, got %q", ErrCodeBadRequest, te.Code)
	}
}

func TestLogAppend_CreatesFileFromTemplate(t *testing.T) {
	dir := t.TempDir()
	cfg := logTestConfig(t, dir)

	// No log.md exists
	te := LogAppend(cfg, "query", "First Entry", "First body.")
	if te != nil {
		t.Fatalf("LogAppend failed: %v", te)
	}

	data, err := os.ReadFile(filepath.Join(dir, "log.md"))
	if err != nil {
		t.Fatalf("log.md not created: %v", err)
	}

	content := string(data)
	if !strings.Contains(content, "# Wiki Log") {
		t.Error("template header not present")
	}
	if !strings.Contains(content, "## [") {
		t.Error("entry header not present")
	}
	if !strings.Contains(content, "First body.") {
		t.Error("entry body not present")
	}

	entries, te := LogTail(cfg, 10)
	if te != nil {
		t.Fatalf("LogTail failed: %v", te)
	}
	if len(entries) != 1 {
		t.Errorf("expected 1 entry, got %d", len(entries))
	}
}

func TestLogAppend_ReadOnly(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "log.md"), []byte(logTemplate), 0o644)
	cfg := logTestConfig(t, dir)
	cfg.Safety.ReadOnly = true

	te := LogAppend(cfg, "ingest", "test", "body")
	if te == nil {
		t.Fatal("expected read-only error")
	}
	if te.Code != ErrCodeReadOnly {
		t.Errorf("want code %q, got %q", ErrCodeReadOnly, te.Code)
	}
}

func TestLogAppend_NoBodyEntry(t *testing.T) {
	dir := t.TempDir()
	cfg := logTestConfig(t, dir)

	te := LogAppend(cfg, "lint", "Quick lint", "")
	if te != nil {
		t.Fatalf("LogAppend failed: %v", te)
	}

	entries, te := LogTail(cfg, 10)
	if te != nil {
		t.Fatalf("LogTail failed: %v", te)
	}
	if len(entries) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(entries))
	}
	if entries[0].Body != "" {
		t.Errorf("expected empty body, got %q", entries[0].Body)
	}
}
