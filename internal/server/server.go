// Package server implements the MCP server (stdio transport, tool/resource/prompt registry).
package server

import (
	"context"
	"crypto/subtle"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
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

// RegisterTool adds a tool. Every handler is wrapped with audit logging.
func (s *Server) RegisterTool(tool mcp.Tool, handler mcpserver.ToolHandlerFunc) {
	s.mcp.AddTool(tool, s.auditWrap(tool.Name, handler))
}

// auditWrap returns a handler that calls the original and then appends an
// entry to audit.md in the wiki root. Audit failures are silently dropped so
// they never affect tool call results.
func (s *Server) auditWrap(toolName string, handler mcpserver.ToolHandlerFunc) mcpserver.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		result, err := handler(ctx, req)
		go s.appendAuditEntry(toolName, req)
		return result, err
	}
}

func (s *Server) appendAuditEntry(toolName string, req mcp.CallToolRequest) {
	auditPath := filepath.Join(s.cfg.WikiPath, "audit.md")

	now := time.Now()
	date := now.Format("2006-01-02")
	timeStr := now.Format("15:04:05")
	project := s.cfg.ProjectPath
	if project == "" {
		project = "-"
	}

	args := req.GetArguments()
	var paramsSummary string
	if b, err := json.Marshal(args); err == nil {
		paramsSummary = string(b)
		if len(paramsSummary) > 200 {
			paramsSummary = paramsSummary[:197] + "..."
		}
	}

	entry := fmt.Sprintf("| %s | %s | %s | %s | %s |\n",
		date, timeStr, project, toolName, paramsSummary)

	f, err := os.OpenFile(auditPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		return
	}
	defer func() {
		_ = f.Close()
	}()

	// Write header if file is empty.
	if fi, err := f.Stat(); err == nil && fi.Size() == 0 {
		_, _ = f.WriteString("# Audit Log\n\n| Date | Time | Project | Tool | Params |\n|------|------|---------|------|--------|\n")
	}
	_, _ = f.WriteString(entry)
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
