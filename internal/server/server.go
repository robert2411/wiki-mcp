// Package server implements the MCP server (stdio transport, tool/resource/prompt registry).
package server

import (
	"context"
	"crypto/subtle"
	"errors"
	"fmt"
	"io"
	"log"
	"log/slog"
	"net/http"
	"time"

	"github.com/mark3labs/mcp-go/mcp"
	mcpserver "github.com/mark3labs/mcp-go/server"
	"github.com/robertstevens/wiki-mcp/internal/config"
)

// Server wraps an MCP server with wiki-mcp configuration and registration.
type Server struct {
	mcp    *mcpserver.MCPServer
	cfg    *config.Config
	logger *slog.Logger
}

// New creates a Server from the loaded config. It does not start listening.
func New(cfg *config.Config, version string, logger *slog.Logger) *Server {
	opts := []mcpserver.ServerOption{
		mcpserver.WithToolCapabilities(true),
		mcpserver.WithResourceCapabilities(false, true),
		mcpserver.WithPromptCapabilities(true),
		mcpserver.WithLogging(),
	}

	m := mcpserver.NewMCPServer("wiki-mcp", version, opts...)

	return &Server{
		mcp:    m,
		cfg:    cfg,
		logger: logger,
	}
}

// RegisterTool adds a tool. M2+ tasks call this from their own packages.
func (s *Server) RegisterTool(tool mcp.Tool, handler mcpserver.ToolHandlerFunc) {
	s.mcp.AddTool(tool, handler)
}

func (s *Server) RegisterResource(resource mcp.Resource, handler mcpserver.ResourceHandlerFunc) {
	s.mcp.AddResource(resource, handler)
}

func (s *Server) RegisterResourceTemplate(tmpl mcp.ResourceTemplate, handler mcpserver.ResourceTemplateHandlerFunc) {
	s.mcp.AddResourceTemplate(tmpl, handler)
}

func (s *Server) RegisterPrompt(prompt mcp.Prompt, handler mcpserver.PromptHandlerFunc) {
	s.mcp.AddPrompt(prompt, handler)
}

func (s *Server) Config() *config.Config {
	return s.cfg
}

// RunStdio starts the stdio transport loop. It blocks until ctx is cancelled
// or an error occurs. Uses the provided reader/writer (typically os.Stdin/os.Stdout).
func (s *Server) RunStdio(ctx context.Context, stdin io.Reader, stdout io.Writer) error {
	s.logger.Info("starting MCP server",
		"transport", "stdio",
		"wiki_path", s.cfg.WikiPath,
		"read_only", s.cfg.Safety.ReadOnly,
	)

	stdio := mcpserver.NewStdioServer(s.mcp)
	stdio.SetErrorLogger(log.New(&slogLogWriter{logger: s.logger}, "", 0))
	return stdio.Listen(ctx, stdin, stdout)
}

type slogLogWriter struct {
	logger *slog.Logger
}

func (w *slogLogWriter) Write(p []byte) (n int, err error) {
	w.logger.Log(context.Background(), slog.LevelError, string(p))
	return len(p), nil
}

// RunStreamableHTTP starts the streamable-HTTP (MCP 2025-03 spec) transport.
// Blocks until ctx is cancelled or an error occurs.
func (s *Server) RunStreamableHTTP(ctx context.Context) error {
	// "" also binds all interfaces (OS default), so warn on both.
	if s.cfg.MCP.Bind == "0.0.0.0" || s.cfg.MCP.Bind == "" {
		s.logger.Warn("MCP transport bound to all interfaces; " +
			"ensure firewall rules are in place. Use a reverse proxy + TLS for anything beyond a trusted LAN.")
	}

	addr := fmt.Sprintf("%s:%d", s.cfg.MCP.Bind, s.cfg.MCP.Port)
	s.logger.Info("starting MCP server",
		"transport", "streamable-http",
		"addr", addr,
		"wiki_path", s.cfg.WikiPath,
		"auth", s.cfg.MCP.AuthToken != "",
	)

	httpSrv := mcpserver.NewStreamableHTTPServer(s.mcp)

	var handler http.Handler = httpSrv
	if s.cfg.MCP.AuthToken != "" {
		handler = BearerAuthMiddleware(s.cfg.MCP.AuthToken, httpSrv)
	}

	srv := &http.Server{
		Addr:              addr,
		Handler:           handler,
		ReadHeaderTimeout: 10 * time.Second,
	}

	errCh := make(chan error, 1)
	go func() {
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			errCh <- err
		}
		close(errCh)
	}()

	select {
	case <-ctx.Done():
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		return srv.Shutdown(shutdownCtx)
	case err := <-errCh:
		return err
	}
}

// BearerAuthMiddleware rejects requests whose Authorization header does not
// match "Bearer <token>". Uses constant-time comparison to prevent timing attacks.
func BearerAuthMiddleware(token string, next http.Handler) http.Handler {
	want := []byte("Bearer " + token)
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		got := []byte(r.Header.Get("Authorization"))
		if subtle.ConstantTimeCompare(got, want) != 1 {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}
		next.ServeHTTP(w, r)
	})
}
