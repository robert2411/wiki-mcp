package sources

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// --- slugFromURL ---

func TestSlugFromURL(t *testing.T) {
	tests := []struct {
		rawURL string
		want   string
	}{
		{"https://example.com/blog/my-post?q=1", "example.com-blog-my-post"},
		{"https://example.com/", "example.com"},
		{"https://example.com", "example.com"},
		{"https://foo.bar/a/b/c", "foo.bar-a-b-c"},
	}
	for _, tc := range tests {
		got, err := slugFromURL(tc.rawURL)
		if err != nil {
			t.Errorf("slugFromURL(%q): unexpected error: %v", tc.rawURL, err)
			continue
		}
		if got != tc.want {
			t.Errorf("slugFromURL(%q) = %q; want %q", tc.rawURL, got, tc.want)
		}
	}
}

// --- resolvePath ---

func TestResolvePath_Escape(t *testing.T) {
	root := t.TempDir()
	_, err := resolvePath(root, "../../etc/passwd")
	if err == nil {
		t.Fatal("expected error for escaping path, got nil")
	}
}

func TestResolvePath_Valid(t *testing.T) {
	root := t.TempDir()
	abs, err := resolvePath(root, "sub/file.md")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.HasPrefix(abs, root) {
		t.Errorf("resolved path %q should be under root %q", abs, root)
	}
}

// --- FetchURL ---

func TestFetchURL_HTMLConvertedToMarkdown(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		_, _ = w.Write([]byte(`<html><body><h1>Hello</h1><p>World</p></body></html>`))
	}))
	defer srv.Close()

	dir := t.TempDir()
	savedPath, te := FetchURL(dir, srv.URL, "test-page")
	if te != nil {
		t.Fatalf("FetchURL error: %v", te)
	}

	// AC #1: file saved with slug-based name
	abs := filepath.Join(dir, "test-page.md")
	data, err := os.ReadFile(abs)
	if err != nil {
		t.Fatalf("saved file not found at %q: %v", abs, err)
	}
	content := string(data)
	if !strings.Contains(content, "Hello") {
		t.Errorf("expected markdown content to contain 'Hello', got: %q", content)
	}
	if savedPath != "test-page.md" {
		t.Errorf("savedPath = %q; want 'test-page.md'", savedPath)
	}
}

func TestFetchURL_DefaultSlugFromURL(t *testing.T) {
	// AC #1: default slug derived from URL
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain")
		_, _ = w.Write([]byte("plain content"))
	}))
	defer srv.Close()

	dir := t.TempDir()
	savedPath, te := FetchURL(dir, srv.URL+"/some/page", "")
	if te != nil {
		t.Fatalf("FetchURL error: %v", te)
	}
	if savedPath == "" {
		t.Fatal("expected non-empty savedPath")
	}
	// slug should be based on host+path
	abs := filepath.Join(dir, savedPath)
	if _, err := os.Stat(abs); os.IsNotExist(err) {
		t.Errorf("expected file at %q to exist", abs)
	}
}

func TestFetchURL_NonHTMLSavedRaw(t *testing.T) {
	// AC #2 (indirectly): plain-text response saved without conversion
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain")
		_, _ = w.Write([]byte("raw text content"))
	}))
	defer srv.Close()

	dir := t.TempDir()
	_, te := FetchURL(dir, srv.URL, "raw-file")
	if te != nil {
		t.Fatalf("FetchURL error: %v", te)
	}

	data, err := os.ReadFile(filepath.Join(dir, "raw-file.md"))
	if err != nil {
		t.Fatalf("file not found: %v", err)
	}
	if string(data) != "raw text content" {
		t.Errorf("expected raw content, got %q", string(data))
	}
}

func TestFetchURL_PathEscapeRejected(t *testing.T) {
	// AC #5: slug that would escape sources_path is rejected
	dir := t.TempDir()
	_, te := FetchURL(dir, "http://example.com", "../../etc/passwd")
	if te == nil {
		t.Fatal("expected error for escaping slug, got nil")
	}
	if te.Code != ErrCodePathEscape {
		t.Errorf("expected PathEscape code, got %q", te.Code)
	}
}

func TestFetchURL_CreatesDirectoryOnFirstWrite(t *testing.T) {
	// sources dir does not exist yet — should be created
	parent := t.TempDir()
	sourcesDir := filepath.Join(parent, "sources")

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain")
		_, _ = w.Write([]byte("content"))
	}))
	defer srv.Close()

	_, te := FetchURL(sourcesDir, srv.URL, "myfile")
	if te != nil {
		t.Fatalf("FetchURL error: %v", te)
	}
	if _, err := os.Stat(filepath.Join(sourcesDir, "myfile.md")); os.IsNotExist(err) {
		t.Error("expected file to be created including parent directory")
	}
}

// --- PDFText ---

func TestPDFText_ExtractFromFixture(t *testing.T) {
	// AC #3: pure-Go extraction from fixture PDF (no pdftotext required)
	_ = os.Unsetenv("WIKI_MCP_PREFER_PDFTOTEXT")

	sourcesDir := t.TempDir()
	dest := filepath.Join(sourcesDir, "sample.pdf")
	data, err := os.ReadFile("testdata/sample.pdf")
	if err != nil {
		t.Fatalf("read fixture PDF: %v", err)
	}
	if err := os.WriteFile(dest, data, 0o644); err != nil {
		t.Fatalf("write PDF to temp dir: %v", err)
	}

	result, te := PDFText(sourcesDir, "sample.pdf")
	if te != nil {
		t.Fatalf("PDFText error: %v", te)
	}
	if result.PageCount <= 0 {
		t.Errorf("expected page_count > 0, got %d", result.PageCount)
	}
	// fixture PDF contains some text
	if strings.TrimSpace(result.Text) == "" {
		t.Error("expected non-empty extracted text")
	}
}

func TestPDFText_PathEscapeRejected(t *testing.T) {
	// AC #5
	dir := t.TempDir()
	_, te := PDFText(dir, "../../etc/passwd")
	if te == nil {
		t.Fatal("expected error for escaping path, got nil")
	}
	if te.Code != ErrCodePathEscape {
		t.Errorf("expected PathEscape code, got %q", te.Code)
	}
}

func TestPDFText_NotFound(t *testing.T) {
	dir := t.TempDir()
	_, te := PDFText(dir, "nonexistent.pdf")
	if te == nil {
		t.Fatal("expected error for missing file, got nil")
	}
	if te.Code != ErrCodeNotFound {
		t.Errorf("expected NotFound code, got %q", te.Code)
	}
}

// --- List ---

func TestList_EmptyWhenDirAbsent(t *testing.T) {
	// AC #4: no error, empty slice when dir does not exist
	nonExistent := filepath.Join(t.TempDir(), "no-such-dir")
	entries, te := List(nonExistent)
	if te != nil {
		t.Fatalf("List error: %v", te)
	}
	if len(entries) != 0 {
		t.Errorf("expected empty list, got %d entries", len(entries))
	}
}

func TestList_ReportsFiles(t *testing.T) {
	// AC #4: reports size + mtime
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "a.md"), []byte("hello"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "b.txt"), []byte("world!"), 0o644); err != nil {
		t.Fatal(err)
	}

	entries, te := List(dir)
	if te != nil {
		t.Fatalf("List error: %v", te)
	}
	if len(entries) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(entries))
	}
	for _, e := range entries {
		if e.Size == 0 {
			t.Errorf("entry %q has zero size", e.Name)
		}
		if e.Mtime.IsZero() {
			t.Errorf("entry %q has zero mtime", e.Name)
		}
	}
}

func TestList_SkipsSubdirectories(t *testing.T) {
	dir := t.TempDir()
	_ = os.Mkdir(filepath.Join(dir, "subdir"), 0o755)
	_ = os.WriteFile(filepath.Join(dir, "file.md"), []byte("x"), 0o644)

	entries, te := List(dir)
	if te != nil {
		t.Fatalf("List error: %v", te)
	}
	for _, e := range entries {
		if e.Name == "subdir" {
			t.Error("List should not include subdirectories")
		}
	}
}
