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

	"github.com/robertstevens/wiki-mcp/internal/config"
	"github.com/robertstevens/wiki-mcp/internal/server"
	"github.com/robertstevens/wiki-mcp/internal/web"
	"github.com/robertstevens/wiki-mcp/internal/wiki"
)

var version = "dev"

func main() {
	showVersion := flag.Bool("version", false, "print version and exit")
	configFile := flag.String("config", "", "path to TOML config file")
	wikiPath := flag.String("wiki-path", "", "path to the wiki directory")
	port := flag.Int("port", 0, "HTTP listen port")
	bind := flag.String("bind", "", "HTTP bind address")
	transport := flag.String("transport", "stdio", "transport mode: stdio or sse")
	serve := flag.Bool("serve", false, "enable the web UI (sets Web.Enabled=true)")

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "wiki-mcp %s — a personal wiki with MCP integration\n\n", version)
		fmt.Fprintf(os.Stderr, "Usage:\n  wiki-mcp [flags]\n\nFlags:\n")
		flag.PrintDefaults()
	}

	flag.Parse()

	if *showVersion {
		fmt.Println(version)
		os.Exit(0)
	}

	if *transport == "sse" {
		fmt.Fprintln(os.Stderr, server.ErrSSENotImplemented)
		os.Exit(1)
	}

	if *transport != "stdio" {
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
		case "port":
			flags.Port = port
		case "bind":
			flags.Bind = bind
		}
	})

	cfg, err := config.Load(flags)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}

	// --serve flag forces Web.Enabled on.
	if *serve {
		cfg.Web.Enabled = true
	}

	logger := slog.New(slog.NewTextHandler(os.Stderr, nil))

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	mcpSrv := server.New(cfg, version, logger)
	wiki.RegisterTools(mcpSrv)
	wiki.RegisterPrompts(mcpSrv)

	g, gctx := errgroup.WithContext(ctx)

	g.Go(func() error {
		return mcpSrv.RunStdio(gctx, os.Stdin, os.Stdout)
	})

	if cfg.Web.Enabled {
		webSrv, err := web.NewServer(cfg, logger)
		if err != nil {
			fmt.Fprintf(os.Stderr, "error: start web server: %v\n", err)
			os.Exit(1)
		}
		g.Go(func() error {
			return webSrv.Run(gctx)
		})
	}

	if err := g.Wait(); err != nil && ctx.Err() == nil {
		logger.Error("server error", "err", err)
		os.Exit(1)
	}
}
