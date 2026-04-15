package web

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"html/template"
	"io/fs"
	"log/slog"
	"mime"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	"github.com/robertstevens/wiki-mcp/internal/config"
	"github.com/robertstevens/wiki-mcp/internal/web/render"
	"github.com/robertstevens/wiki-mcp/internal/wiki"
	webtheme "github.com/robertstevens/wiki-mcp/web"
)

// pageEntry holds a rendered page alongside the file metadata used for
// ETag / Last-Modified headers.
type pageEntry struct {
	page    *render.RenderedPage
	etag    string
	modTime time.Time
}

// Server is the read-only HTTP wiki UI.
type Server struct {
	cfg      *config.Config
	logger   *slog.Logger
	tmplPage *template.Template
	tmplSrch *template.Template
	themeFS  fs.FS

	// mu guards renderer, indexCache, navCache, titleCache, and pageCache. All
	// are rebuilt lazily on first use and nulled by InvalidateCache() when the
	// file watcher detects changes.
	mu          sync.RWMutex
	renderer    *render.Renderer
	indexCache  []SearchIndexEntry
	navCache    []navSection
	titleCache  string // empty = not yet loaded
	pageCache   map[string]*pageEntry
}

// pageData is passed to page.html.
type pageData struct {
	Title       string
	WikiTitle   string
	Query       string
	ContentHTML template.HTML
	Nav         []navSection
}

// searchData is passed to search.html.
type searchData struct {
	WikiTitle string
	Query     string
	Results   []SearchIndexEntry
	Nav       []navSection
}

// navSection groups nav links under a heading from index.md.
type navSection struct {
	Title string
	Links []navEntry
}

type navEntry struct {
	Path  string
	Title string
}

// NewServer creates a Server. Templates are parsed eagerly so errors surface
// at startup. The renderer and search index are built lazily on first request.
func NewServer(cfg *config.Config, logger *slog.Logger) (*Server, error) {
	sub, err := fs.Sub(webtheme.FS, "theme/default")
	if err != nil {
		return nil, fmt.Errorf("web: sub theme fs: %w", err)
	}

	tmplPage, err := template.New("page.html").ParseFS(sub, "page.html")
	if err != nil {
		return nil, fmt.Errorf("web: parse page.html: %w", err)
	}
	tmplSrch, err := template.New("search.html").ParseFS(sub, "search.html")
	if err != nil {
		return nil, fmt.Errorf("web: parse search.html: %w", err)
	}

	return &Server{
		cfg:      cfg,
		logger:   logger,
		tmplPage: tmplPage,
		tmplSrch: tmplSrch,
		themeFS:  sub,
	}, nil
}

// Handler returns the HTTP handler. Used in tests and by Run.
func (s *Server) Handler() http.Handler {
	return s.buildRouter()
}

// Run starts the HTTP server and blocks until ctx is cancelled or an error
// occurs. It performs a graceful shutdown with a 10-second drain timeout.
func (s *Server) Run(ctx context.Context) error {
	if s.cfg.Web.AutoRebuild {
		startWatcher(ctx, s.cfg.WikiPath, s.logger, s.InvalidateCache, fsnotify.NewWatcher)
	}

	addr := fmt.Sprintf("%s:%d", s.cfg.Web.Bind, s.cfg.Web.Port)
	srv := &http.Server{
		Addr:    addr,
		Handler: s.Handler(),
	}

	ln, err := net.Listen("tcp", addr)
	if err != nil {
		return fmt.Errorf("web: listen %s: %w", addr, err)
	}
	s.logger.Info("web UI listening", "addr", "http://"+addr)

	serveErr := make(chan error, 1)
	go func() {
		serveErr <- srv.Serve(ln)
	}()

	select {
	case <-ctx.Done():
		shutCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if err := srv.Shutdown(shutCtx); err != nil {
			return fmt.Errorf("web: shutdown: %w", err)
		}
		return nil
	case err := <-serveErr:
		if errors.Is(err, http.ErrServerClosed) {
			return nil
		}
		return err
	}
}

func (s *Server) buildRouter() http.Handler {
	r := chi.NewRouter()
	r.Use(middleware.RealIP)
	r.Use(middleware.Recoverer)

	r.Get("/", s.handlePage("index.md"))
	r.Get("/_log", s.handlePage("log.md"))
	r.Get("/_search", s.handleSearch)
	r.Get("/_search_index.json", s.handleSearchIndex)
	r.Get("/_theme/*", s.handleTheme)
	r.Get("/_assets/*", s.handleAsset)
	r.Get("/*", s.handleWikiPage) // catch-all: maps URL path → <path>.md

	return r
}

// InvalidateCache nils the renderer, search index, nav, title, and page cache
// so they are rebuilt on the next request. Called by the file watcher on disk
// changes; also exported for testing.
func (s *Server) InvalidateCache() {
	s.mu.Lock()
	s.renderer = nil
	s.indexCache = nil
	s.navCache = nil
	s.titleCache = ""
	s.pageCache = nil
	s.mu.Unlock()
}

// cachedIndex returns the search index, building it lazily on first call.
// invalidateCache() resets it so the next call rebuilds from disk.
func (s *Server) cachedIndex() []SearchIndexEntry {
	s.mu.RLock()
	idx := s.indexCache
	s.mu.RUnlock()
	if idx != nil {
		return idx
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	if s.indexCache == nil {
		built, err := BuildSearchIndex(s.cfg.WikiPath)
		if err != nil {
			s.logger.Error("build search index", "err", err)
		}
		s.indexCache = built
	}
	return s.indexCache
}

func (s *Server) navSections() []navSection {
	s.mu.RLock()
	nav := s.navCache
	s.mu.RUnlock()
	if nav != nil {
		return nav
	}

	// Build the search index first (acquires + releases its own lock) so that
	// the nav write lock below does not re-enter the mutex.
	idx := s.cachedIndex()

	s.mu.Lock()
	defer s.mu.Unlock()
	if s.navCache == nil {
		s.navCache = navFromIndex(s.cfg, idx)
	}
	return s.navCache
}

// navFromIndex parses index.md via wiki.ParseIndex for section headings and
// links. Falls back to a single unnamed section built from the flat search
// index when index.md is absent, unparseable, or contains no sections.
func navFromIndex(cfg *config.Config, fallback []SearchIndexEntry) []navSection {
	data, err := os.ReadFile(filepath.Join(cfg.WikiPath, "index.md"))
	if err != nil {
		return flatNavSection(fallback)
	}
	doc, err := wiki.ParseIndex(data, cfg)
	if err != nil {
		return flatNavSection(fallback)
	}
	var out []navSection
	for _, sec := range doc.Sections {
		if len(sec.Entries) == 0 {
			continue
		}
		links := make([]navEntry, 0, len(sec.Entries))
		for _, e := range sec.Entries {
			links = append(links, navEntry{
				Path:  strings.TrimSuffix(e.Path, ".md"),
				Title: e.Title,
			})
		}
		out = append(out, navSection{Title: sec.Title, Links: links})
	}
	if len(out) == 0 {
		return flatNavSection(fallback)
	}
	return out
}

func flatNavSection(entries []SearchIndexEntry) []navSection {
	links := make([]navEntry, len(entries))
	for i, e := range entries {
		links[i] = navEntry{Path: e.Path, Title: e.Title}
	}
	return []navSection{{Links: links}}
}

// cachedWikiTitle returns the wiki title (H1 from index.md), building and
// caching it on first call. Cleared by InvalidateCache().
func (s *Server) cachedWikiTitle() string {
	s.mu.RLock()
	t := s.titleCache
	s.mu.RUnlock()
	if t != "" {
		return t
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	if s.titleCache == "" {
		s.titleCache = wikiTitle(s.cfg.WikiPath)
	}
	return s.titleCache
}

// wikiTitle reads the H1 from index.md as the wiki title. Falls back to "wiki".
func wikiTitle(wikiPath string) string {
	data, err := os.ReadFile(filepath.Join(wikiPath, "index.md"))
	if err != nil {
		return "wiki"
	}
	_, body := wiki.ParseFrontmatter(data)
	if t := h1Title(body); t != "" {
		return t
	}
	return "wiki"
}

// handlePage returns a handler that renders a fixed page (e.g. index.md).
func (s *Server) handlePage(relPath string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		s.servePage(w, r, relPath)
	}
}

// handleWikiPage maps /{*} → {*}.md (catch-all route).
func (s *Server) handleWikiPage(w http.ResponseWriter, r *http.Request) {
	urlPath := chi.URLParam(r, "*")
	if urlPath == "" {
		s.servePage(w, r, "index.md")
		return
	}
	relPath := filepath.Clean(urlPath) + ".md"
	s.servePage(w, r, relPath)
}

func (s *Server) servePage(w http.ResponseWriter, r *http.Request, relPath string) {
	abs, err := s.cfg.ResolveWikiPath(relPath)
	if err != nil {
		http.NotFound(w, r)
		return
	}

	cp, err := s.cachedPageEntry(abs, relPath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			http.NotFound(w, r)
			return
		}
		http.Error(w, "render error", http.StatusInternalServerError)
		s.logger.Error("render page", "path", relPath, "err", err)
		return
	}

	w.Header().Set("ETag", cp.etag)
	w.Header().Set("Last-Modified", cp.modTime.UTC().Format(http.TimeFormat))
	data := pageData{
		Title:       cp.page.Title,
		WikiTitle:   s.cachedWikiTitle(),
		ContentHTML: template.HTML(cp.page.HTML), // #nosec G203 — content is produced by our own markdown renderer
		Nav:         s.navSections(),
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := s.tmplPage.Execute(w, data); err != nil {
		s.logger.Error("template execute", "err", err)
	}
}

// cachedPageEntry returns a rendered page from the page cache, building it on
// first call. The cache is invalidated by InvalidateCache().
func (s *Server) cachedPageEntry(abs, relPath string) (*pageEntry, error) {
	s.mu.RLock()
	var pe *pageEntry
	if s.pageCache != nil {
		pe = s.pageCache[abs]
	}
	s.mu.RUnlock()
	if pe != nil {
		return pe, nil
	}

	// Cache miss: stat + render under write lock.
	s.mu.Lock()
	defer s.mu.Unlock()

	// Re-check under write lock in case another goroutine populated it.
	if s.pageCache != nil {
		if pe = s.pageCache[abs]; pe != nil {
			return pe, nil
		}
	}

	info, err := os.Stat(abs)
	if err != nil {
		return nil, err
	}

	if s.renderer == nil {
		rdr, err := render.NewRenderer(s.cfg.WikiPath)
		if err != nil {
			return nil, fmt.Errorf("build renderer: %w", err)
		}
		s.renderer = rdr
	}

	page, err := s.renderer.RenderFile(abs, relPath)
	if err != nil {
		return nil, err
	}

	pe = &pageEntry{
		page:    page,
		etag:    fmt.Sprintf(`"%x-%x"`, info.ModTime().UnixNano(), info.Size()),
		modTime: info.ModTime(),
	}
	if s.pageCache == nil {
		s.pageCache = make(map[string]*pageEntry)
	}
	s.pageCache[abs] = pe
	return pe, nil
}

func (s *Server) handleSearch(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query().Get("q")
	data := searchData{
		WikiTitle: s.cachedWikiTitle(),
		Query:     q,
		Results:   Search(s.cachedIndex(), q),
		Nav:       s.navSections(),
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := s.tmplSrch.Execute(w, data); err != nil {
		s.logger.Error("search template execute", "err", err)
	}
}

func (s *Server) handleSearchIndex(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	enc := json.NewEncoder(w)
	enc.SetIndent("", "")
	if err := enc.Encode(s.cachedIndex()); err != nil {
		s.logger.Error("encode search index", "err", err)
	}
}

// handleTheme serves embedded theme static files (CSS, JS, etc.).
func (s *Server) handleTheme(w http.ResponseWriter, r *http.Request) {
	name := chi.URLParam(r, "*")
	data, err := fs.ReadFile(s.themeFS, name)
	if err != nil {
		http.NotFound(w, r)
		return
	}
	ct := mime.TypeByExtension(filepath.Ext(name))
	if ct == "" {
		ct = "application/octet-stream"
	}
	w.Header().Set("Content-Type", ct)
	_, _ = w.Write(data)
}

// handleAsset serves static files (images, PDFs, etc.) from the wiki dir.
// Rejects .md files and any path that would escape the wiki root.
func (s *Server) handleAsset(w http.ResponseWriter, r *http.Request) {
	name := filepath.Clean(chi.URLParam(r, "*"))

	if strings.HasSuffix(strings.ToLower(name), ".md") {
		http.NotFound(w, r)
		return
	}

	abs, err := s.cfg.ResolveWikiPath(name)
	if err != nil {
		http.NotFound(w, r)
		return
	}

	// ResolveWikiPath enforces confinement only when ConfineToWikiPath is set.
	// Always enforce it here regardless of that flag, since the web server must
	// never serve files outside the wiki root.
	wikiRoot := filepath.Clean(s.cfg.WikiPath)
	if !strings.HasPrefix(abs, wikiRoot+string(os.PathSeparator)) && abs != wikiRoot {
		http.NotFound(w, r)
		return
	}

	http.ServeFile(w, r, abs)
}
