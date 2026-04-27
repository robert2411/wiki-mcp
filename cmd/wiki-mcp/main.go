package main

import (
	"context"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"golang.org/x/sync/errgroup"

	"github.com/robert2411/wiki-mcp/internal/config"
	"github.com/robert2411/wiki-mcp/internal/server"
	"github.com/robert2411/wiki-mcp/internal/sources"
	"github.com/robert2411/wiki-mcp/internal/web"
	"github.com/robert2411/wiki-mcp/internal/wiki"
)

var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

const (
	transportStdio = "stdio"
	transportSSE   = "sse"
)

func main() {
	showVersion := flag.Bool("version", false, "print version and exit")
	configFile := flag.String("config", "", "path to TOML config file")
	wikiPath := flag.String("wiki-path", "", "path to the wiki directory")
	projectPath := flag.String("project", "", "scope all tools to this project subdirectory (must be within wiki-path)")
	subProjectPath := flag.String("sub-project", "", "scope all tools to this sub-project (must be within --project); enables parent-project write access")
	port := flag.Int("port", 0, "web UI HTTP listen port")
	bind := flag.String("bind", "", "bind address for MCP HTTP transport (and web UI)")
	mcpPort := flag.Int("mcp-port", 0, "MCP HTTP transport listen port (streamable-http, used with --transport sse)")
	authToken := flag.String("auth-token", "", "bearer token required for MCP HTTP transport requests")
	transport := flag.String("transport", transportStdio, "transport mode: stdio or sse")
	serve := flag.Bool("serve", false, "enable the web UI (sets Web.Enabled=true)")
	serveOnly := flag.Bool("serve-only", false, "run web UI only (implies --serve), no MCP stdio transport")

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "wiki-mcp %s — a personal wiki with MCP integration\n\n", version)
		fmt.Fprintf(os.Stderr, "Usage:\n  wiki-mcp [flags]\n\nFlags:\n")
		flag.PrintDefaults()
	}

	flag.Parse()

	if *showVersion {
		fmt.Printf("%s (commit=%s, built=%s)\n", version, commit, date)
		os.Exit(0)
	}

	if !*serveOnly && *transport != transportStdio && *transport != transportSSE {
		fmt.Fprintf(os.Stderr, "error: unknown transport %q (valid: stdio, sse)\n", *transport)
		os.Exit(1)
	}

	// Collect only explicitly-set flags into the Flags struct.
	var flags config.Flags
	flag.Visit(func(f *flag.Flag) {
		switch f.Name {
		case "config":
			flags.ConfigFile = configFile
		case "wiki-path":
			flags.WikiPath = wikiPath
		case "project":
			flags.ProjectPath = projectPath
		case "sub-project":
			flags.SubProjectPath = subProjectPath
		case "port":
			flags.Port = port
		case "bind":
			flags.Bind = bind
		case "mcp-port":
			flags.MCPPort = mcpPort
		case "auth-token":
			flags.AuthToken = authToken
		}
	})

	cfg, err := config.Load(flags)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}

	// --serve / --serve-only both force Web.Enabled on.
	if *serve || *serveOnly {
		cfg.Web.Enabled = true
	}

	logger := slog.New(slog.NewTextHandler(os.Stderr, nil))

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	mcpSrv := server.New(cfg, version, logger)
	wiki.RegisterTools(mcpSrv)
	wiki.RegisterPrompts(mcpSrv)
	wiki.RegisterResources(mcpSrv)
	sources.RegisterTools(mcpSrv)

	g, gctx := errgroup.WithContext(ctx)

	if !*serveOnly {
		if *transport == transportSSE {
			g.Go(func() error {
				return mcpSrv.RunStreamableHTTP(gctx)
			})
		} else {
			g.Go(func() error {
				return mcpSrv.RunStdio(gctx, os.Stdin, os.Stdout)
			})
		}
	}

	if cfg.Web.Enabled {
		webSrv, err := web.NewServer(cfg, logger)
		if err != nil {
			fmt.Fprintf(os.Stderr, "error: start web server: %v\n", err)
			os.Exit(1)
		}
		g.Go(func() error {
			return webSrv.Run(gctx, nil)
		})
	}

	if err := g.Wait(); err != nil && ctx.Err() == nil {
		logger.Error("server error", "err", err)
		os.Exit(1)
	}
}
