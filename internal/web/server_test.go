package web_test

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"log/slog"

	"github.com/robertstevens/wiki-mcp/internal/config"
	"github.com/robertstevens/wiki-mcp/internal/web"
)

func testServer(t *testing.T, wikiPath string) *httptest.Server {
	t.Helper()
	return testServerWithCfg(t, &config.Config{
		WikiPath: wikiPath,
		Web:      config.WebConfig{Port: 0, Bind: "127.0.0.1", Enabled: true},
		Safety:   config.SafetyConfig{ConfineToWikiPath: true},
	})
}

func testServerWithCfg(t *testing.T, cfg *config.Config) *httptest.Server {
	t.Helper()
	logger := slog.New(slog.NewTextHandler(os.Stderr, nil))
	srv, err := web.NewServer(cfg, logger)
	if err != nil {
		t.Fatalf("NewServer: %v", err)
	}
	return httptest.NewServer(srv.Handler())
}

func TestIndexPageRenders(t *testing.T) {
	wikiPath := t.TempDir()
	writeFile(t, wikiPath, "index.md", "# Test Wiki\n\nHello world.")
	ts := testServer(t, wikiPath)
	defer ts.Close()

	resp, err := http.Get(ts.URL + "/")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Errorf("GET / = %d, want 200", resp.StatusCode)
	}
}

func TestPageRenders(t *testing.T) {
	wikiPath := t.TempDir()
	writeFile(t, wikiPath, "index.md", "# Index\n")
	writeFile(t, wikiPath, "entities/ollama.md", "# Ollama\n\nAn LLM runner.")
	ts := testServer(t, wikiPath)
	defer ts.Close()

	resp, err := http.Get(ts.URL + "/entities/ollama")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Errorf("GET /entities/ollama = %d, want 200", resp.StatusCode)
	}
}

func TestMissingPageReturns404(t *testing.T) {
	wikiPath := t.TempDir()
	writeFile(t, wikiPath, "index.md", "# Index\n")
	ts := testServer(t, wikiPath)
	defer ts.Close()

	resp, err := http.Get(ts.URL + "/nonexistent")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusNotFound {
		t.Errorf("GET /nonexistent = %d, want 404", resp.StatusCode)
	}
}

func TestPathTraversalReturns404(t *testing.T) {
	wikiPath := t.TempDir()
	writeFile(t, wikiPath, "index.md", "# Index\n")
	ts := testServer(t, wikiPath)
	defer ts.Close()

	cases := []string{
		"/_assets/../../etc/passwd",
		"/_assets/../index.md",
		"/_assets/%2e%2e%2fetc%2fpasswd",
	}
	for _, path := range cases {
		resp, err := http.Get(ts.URL + path)
		if err != nil {
			t.Fatalf("GET %s: %v", path, err)
		}
		resp.Body.Close()
		if resp.StatusCode == http.StatusOK {
			t.Errorf("GET %s = 200, want non-200 (path traversal not blocked)", path)
		}
	}
}

func TestSearchReturnsMatches(t *testing.T) {
	wikiPath := t.TempDir()
	writeFile(t, wikiPath, "index.md", "# Index\n")
	writeFile(t, wikiPath, "entities/qwen.md", "# Qwen\n\nA large language model from Alibaba.")
	ts := testServer(t, wikiPath)
	defer ts.Close()

	resp, err := http.Get(ts.URL + "/_search?q=qwen")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Errorf("GET /_search?q=qwen = %d, want 200", resp.StatusCode)
	}
}

func TestSearchIndexJSON(t *testing.T) {
	wikiPath := t.TempDir()
	writeFile(t, wikiPath, "index.md", "# Index\n\nWelcome.")
	writeFile(t, wikiPath, "page.md", "# Page\n\nContent here.")
	ts := testServer(t, wikiPath)
	defer ts.Close()

	resp, err := http.Get(ts.URL + "/_search_index.json")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Errorf("GET /_search_index.json = %d, want 200", resp.StatusCode)
	}
	if ct := resp.Header.Get("Content-Type"); ct != "application/json; charset=utf-8" {
		t.Errorf("Content-Type = %q, want application/json", ct)
	}

	var entries []web.SearchIndexEntry
	if err := json.NewDecoder(resp.Body).Decode(&entries); err != nil {
		t.Fatalf("decode JSON: %v", err)
	}
	if len(entries) < 2 {
		t.Errorf("got %d entries, want >= 2", len(entries))
	}
}

func TestStyleCSS(t *testing.T) {
	wikiPath := t.TempDir()
	writeFile(t, wikiPath, "index.md", "# Index\n")
	ts := testServer(t, wikiPath)
	defer ts.Close()

	resp, err := http.Get(ts.URL + "/_theme/style.css")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Errorf("GET /_theme/style.css = %d, want 200", resp.StatusCode)
	}
}

// TestDiskChangeReflected verifies AC#1: editing a page on disk is visible
// on the next request once the cache is invalidated.
func TestDiskChangeReflected(t *testing.T) {
	wikiPath := t.TempDir()
	writeFile(t, wikiPath, "index.md", "# Original Title\n\nOriginal content.")
	writeFile(t, wikiPath, "page.md", "# Page\n\nHello.")

	cfg := &config.Config{
		WikiPath: wikiPath,
		Web:      config.WebConfig{Port: 0, Bind: "127.0.0.1", Enabled: true, AutoRebuild: true},
		Safety:   config.SafetyConfig{ConfineToWikiPath: true},
	}
	logger := slog.New(slog.NewTextHandler(os.Stderr, nil))
	srv, err := web.NewServer(cfg, logger)
	if err != nil {
		t.Fatalf("NewServer: %v", err)
	}
	ts := httptest.NewServer(srv.Handler())
	defer ts.Close()

	// First request — should see original content.
	body1 := httpGetBody(t, ts.URL+"/")
	if !strings.Contains(body1, "Original") {
		t.Errorf("first request: expected 'Original' in body, got:\n%s", body1)
	}

	// Edit the file on disk.
	writeFile(t, wikiPath, "index.md", "# Updated Title\n\nUpdated content.")

	// Invalidate cache (simulates what the file watcher does).
	srv.InvalidateCache()

	// Next request should reflect the update.
	body2 := httpGetBody(t, ts.URL+"/")
	if !strings.Contains(body2, "Updated") {
		t.Errorf("post-invalidation request: expected 'Updated' in body, got:\n%s", body2)
	}
}

// TestETagLastModified verifies AC#2: ETag and Last-Modified headers are present.
func TestETagLastModified(t *testing.T) {
	wikiPath := t.TempDir()
	writeFile(t, wikiPath, "index.md", "# Index\n\nHello.")
	ts := testServer(t, wikiPath)
	defer ts.Close()

	resp, err := http.Get(ts.URL + "/")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("GET / = %d, want 200", resp.StatusCode)
	}
	if etag := resp.Header.Get("ETag"); etag == "" {
		t.Error("ETag header missing from rendered-page response")
	}
	if lm := resp.Header.Get("Last-Modified"); lm == "" {
		t.Error("Last-Modified header missing from rendered-page response")
	}
}

// TestAutoRebuildFalse verifies AC#3: when AutoRebuild=false, disk edits are
// not reflected until server restart (cache is never invalidated).
func TestAutoRebuildFalse(t *testing.T) {
	wikiPath := t.TempDir()
	writeFile(t, wikiPath, "index.md", "# Stable Title\n\nStable content.")

	cfg := &config.Config{
		WikiPath: wikiPath,
		Web:      config.WebConfig{Port: 0, Bind: "127.0.0.1", Enabled: true, AutoRebuild: false},
		Safety:   config.SafetyConfig{ConfineToWikiPath: true},
	}
	ts := testServerWithCfg(t, cfg)
	defer ts.Close()

	// Warm the cache.
	body1 := httpGetBody(t, ts.URL+"/")
	if !strings.Contains(body1, "Stable") {
		t.Fatalf("first request: expected 'Stable' in body")
	}

	// Edit the file — watcher is off, cache should NOT be invalidated.
	writeFile(t, wikiPath, "index.md", "# Changed Title\n\nChanged content.")

	body2 := httpGetBody(t, ts.URL+"/")
	if strings.Contains(body2, "Changed") {
		t.Error("AutoRebuild=false: disk edit was visible without cache invalidation")
	}
	if !strings.Contains(body2, "Stable") {
		t.Errorf("AutoRebuild=false: expected cached 'Stable' content, got:\n%s", body2)
	}
}

func httpGetBody(t *testing.T, url string) string {
	t.Helper()
	resp, err := http.Get(url) //nolint:noctx
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	var b strings.Builder
	if _, err := io.Copy(&b, resp.Body); err != nil {
		t.Fatal(err)
	}
	return b.String()
}

func writeFile(t *testing.T, dir, relPath, content string) {
	t.Helper()
	abs := filepath.Join(dir, relPath)
	if err := os.MkdirAll(filepath.Dir(abs), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(abs, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}
