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

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	"github.com/robertstevens/wiki-mcp/internal/config"
	"github.com/robertstevens/wiki-mcp/internal/web/render"
	webtheme "github.com/robertstevens/wiki-mcp/web"
)

// Server is the read-only HTTP wiki UI.
type Server struct {
	cfg      *config.Config
	logger   *slog.Logger
	renderer *render.Renderer
	tmplPage *template.Template
	tmplSrch *template.Template
	themeFS  fs.FS

	indexOnce  sync.Once
	indexCache []SearchIndexEntry
}

// pageData is passed to page.html.
type pageData struct {
	Title       string
	Query       string
	ContentHTML template.HTML
	Nav         []navEntry
}

// searchData is passed to search.html.
type searchData struct {
	Query   string
	Results []SearchIndexEntry
	Nav     []navEntry
}

type navEntry struct {
	Path  string
	Title string
}

// NewServer creates a Server. Renderer is initialised eagerly so template
// parse errors surface at startup rather than on first request.
func NewServer(cfg *config.Config, logger *slog.Logger) (*Server, error) {
	rdr, err := render.NewRenderer(cfg.WikiPath)
	if err != nil {
		return nil, fmt.Errorf("web: build renderer: %w", err)
	}

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
		renderer: rdr,
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

// cachedIndex returns the search index, building it once on first call.
// task-17 (file watcher) will invalidate this cache on disk changes.
func (s *Server) cachedIndex() []SearchIndexEntry {
	s.indexOnce.Do(func() {
		idx, err := BuildSearchIndex(s.cfg.WikiPath)
		if err != nil {
			s.logger.Error("build search index", "err", err)
		}
		s.indexCache = idx
	})
	return s.indexCache
}

func (s *Server) navLinks() []navEntry {
	index := s.cachedIndex()
	nav := make([]navEntry, 0, len(index))
	for _, e := range index {
		nav = append(nav, navEntry{Path: e.Path, Title: e.Title})
	}
	return nav
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
	page, err := s.renderer.RenderFile(abs, relPath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			http.NotFound(w, r)
			return
		}
		http.Error(w, "render error", http.StatusInternalServerError)
		s.logger.Error("render page", "path", relPath, "err", err)
		return
	}

	data := pageData{
		Title:       page.Title,
		ContentHTML: template.HTML(page.HTML), // #nosec G203 — content is produced by our own markdown renderer
		Nav:         s.navLinks(),
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := s.tmplPage.Execute(w, data); err != nil {
		s.logger.Error("template execute", "err", err)
	}
}

func (s *Server) handleSearch(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query().Get("q")
	data := searchData{
		Query:   q,
		Results: Search(s.cachedIndex(), q),
		Nav:     s.navLinks(),
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
