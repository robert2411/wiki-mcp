package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/robertstevens/wiki-mcp/internal/config"
)

var version = "dev"

func main() {
	showVersion := flag.Bool("version", false, "print version and exit")
	configFile := flag.String("config", "", "path to TOML config file")
	wikiPath := flag.String("wiki-path", "", "path to the wiki directory")
	port := flag.Int("port", 0, "HTTP listen port")
	bind := flag.String("bind", "", "HTTP bind address")

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

	// Placeholders — will be wired up in later tasks.
	_ = cfg

	fmt.Printf("wiki-mcp %s\n", version)
	fmt.Printf("wiki: %s\n", cfg.WikiPath)
}
