// Package server implements the MCP server (stdio transport, tool/resource/prompt registry).
package server

import (
	"context"
	"errors"
	"io"
	"log"
	"log/slog"

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

// ErrSSENotImplemented is returned when --transport sse is requested.
var ErrSSENotImplemented = errors.New("transport \"sse\" is not implemented: see TASK-22")
