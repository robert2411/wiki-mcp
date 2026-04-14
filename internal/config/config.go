// Package config loads wiki-mcp configuration from TOML files, environment
// variables, and CLI flags with layered precedence.
package config

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/BurntSushi/toml"
)

// ErrReadOnly is returned by MustMutate when the config has ReadOnly set.
var ErrReadOnly = errors.New("wiki is in read-only mode")

// IndexSection describes one section header in index.md.
type IndexSection struct {
	Key   string `toml:"key"`
	Title string `toml:"title"`
}

// WebConfig controls the built-in web server.
type WebConfig struct {
	Enabled      bool   `toml:"enabled"`
	Port         int    `toml:"port"`
	Bind         string `toml:"bind"`
	Theme        string `toml:"theme"`
	AutoRebuild  bool   `toml:"auto_rebuild"`
}

// IndexConfig controls index.md rendering.
type IndexConfig struct {
	Sections []IndexSection `toml:"sections"`
}

// LogConfig controls log.md rendering.
type LogConfig struct {
	DateFormat string `toml:"date_format"`
}

// LinksConfig controls how wiki links are written.
type LinksConfig struct {
	Style string `toml:"style"`
}

// SafetyConfig contains hard guardrails.
type SafetyConfig struct {
	ReadOnly          bool `toml:"read_only"`
	ConfineToWikiPath bool `toml:"confine_to_wiki_path"`
	MaxPageBytes      int  `toml:"max_page_bytes"`
}

// Config is the top-level configuration struct.
type Config struct {
	WikiPath    string      `toml:"wiki_path"`
	SourcesPath string      `toml:"sources_path"`
	Web         WebConfig   `toml:"web"`
	Index       IndexConfig `toml:"index"`
	Log         LogConfig   `toml:"log"`
	Links       LinksConfig `toml:"links"`
	Safety      SafetyConfig `toml:"safety"`
}

// Flags holds CLI flag values that were explicitly set by the user.
// nil pointer means "not set by user."
type Flags struct {
	ConfigFile *string
	WikiPath   *string
	Port       *int
	Bind       *string
}

// Defaults returns a Config populated with built-in defaults.
func Defaults() Config {
	return Config{
		Web: WebConfig{
			Port:        9000,
			Bind:        "127.0.0.1",
			Theme:       "default",
			AutoRebuild: true,
		},
		Index: IndexConfig{
			Sections: []IndexSection{
				{Key: "research", Title: "🔬 Research"},
				{Key: "entities", Title: "🏷️ Entities"},
				{Key: "concepts", Title: "💡 Concepts"},
				{Key: "infrastructure", Title: "🏗️ Infrastructure"},
			},
		},
		Log: LogConfig{
			DateFormat: "%Y-%m-%d",
		},
		Links: LinksConfig{
			Style: "preserve",
		},
		Safety: SafetyConfig{
			ConfineToWikiPath: true,
			MaxPageBytes:      1048576,
		},
	}
}

// Load builds a Config by layering sources in precedence order (later wins):
//  1. Built-in defaults
//  2. ./wiki-mcp.toml in CWD
//  3. $XDG_CONFIG_HOME/wiki-mcp/config.toml
//  4. WIKI_MCP_CONFIG env file + scalar env overrides
//  5. CLI flags (via Flags struct)
func Load(flags Flags) (*Config, error) {
	cfg := Defaults()

	// Layer 2: CWD file
	_ = loadTOMLFile("wiki-mcp.toml", &cfg)

	// Layer 3: XDG config
	xdgPath := xdgConfigPath()
	if xdgPath != "" {
		_ = loadTOMLFile(xdgPath, &cfg)
	}

	// Layer 4a: WIKI_MCP_CONFIG env file
	if envFile, ok := os.LookupEnv("WIKI_MCP_CONFIG"); ok && envFile != "" {
		if err := loadTOMLFile(envFile, &cfg); err != nil {
			return nil, fmt.Errorf("WIKI_MCP_CONFIG=%q: %w", envFile, err)
		}
	}

	// Layer 4b: scalar env overrides
	applyEnvOverrides(&cfg)

	// Layer 5a: --config flag file (loaded before other flags override)
	if flags.ConfigFile != nil {
		if err := loadTOMLFile(*flags.ConfigFile, &cfg); err != nil {
			return nil, fmt.Errorf("--config %q: %w", *flags.ConfigFile, err)
		}
	}

	// Layer 5b: remaining CLI flags
	if flags.WikiPath != nil {
		cfg.WikiPath = *flags.WikiPath
	}
	if flags.Port != nil {
		cfg.Web.Port = *flags.Port
	}
	if flags.Bind != nil {
		cfg.Web.Bind = *flags.Bind
	}

	if err := validate(&cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}

// ResolveWikiPath joins rel to the wiki root and returns a cleaned absolute
// path. If ConfineToWikiPath is true and the result escapes the wiki root,
// an error is returned.
func (c *Config) ResolveWikiPath(rel string) (string, error) {
	resolved := filepath.Join(c.WikiPath, rel)

	if c.Safety.ConfineToWikiPath {
		root := c.WikiPath
		if resolved != root && !strings.HasPrefix(resolved, root+string(os.PathSeparator)) {
			return "", fmt.Errorf("path %q escapes wiki root %q", rel, root)
		}
	}

	return resolved, nil
}

// MustMutate returns ErrReadOnly when the config has ReadOnly set.
// Mutating tools should call this before making changes.
func (c *Config) MustMutate() error {
	if c.Safety.ReadOnly {
		return ErrReadOnly
	}
	return nil
}

func loadTOMLFile(path string, cfg *Config) error {
	_, err := toml.DecodeFile(path, cfg)
	if errors.Is(err, os.ErrNotExist) {
		return nil
	}
	return err
}

func xdgConfigPath() string {
	dir := os.Getenv("XDG_CONFIG_HOME")
	if dir == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return ""
		}
		dir = filepath.Join(home, ".config")
	}
	return filepath.Join(dir, "wiki-mcp", "config.toml")
}

func applyEnvOverrides(cfg *Config) {
	envStr("WIKI_MCP_WIKI_PATH", &cfg.WikiPath)
	envStr("WIKI_MCP_SOURCES_PATH", &cfg.SourcesPath)
	envBool("WIKI_MCP_WEB_ENABLED", &cfg.Web.Enabled)
	envInt("WIKI_MCP_WEB_PORT", &cfg.Web.Port)
	envStr("WIKI_MCP_WEB_BIND", &cfg.Web.Bind)
	envStr("WIKI_MCP_WEB_THEME", &cfg.Web.Theme)
	envStr("WIKI_MCP_LINKS_STYLE", &cfg.Links.Style)
	envBool("WIKI_MCP_SAFETY_READ_ONLY", &cfg.Safety.ReadOnly)
	envBool("WIKI_MCP_SAFETY_CONFINE", &cfg.Safety.ConfineToWikiPath)
}

func envStr(key string, dest *string) {
	if v, ok := os.LookupEnv(key); ok {
		*dest = v
	}
}

func envBool(key string, dest *bool) {
	if v, ok := os.LookupEnv(key); ok {
		*dest = v == "true" || v == "1"
	}
}

func envInt(key string, dest *int) {
	if v, ok := os.LookupEnv(key); ok {
		if p, err := strconv.Atoi(v); err == nil {
			*dest = p
		}
	}
}

func validate(cfg *Config) error {
	if cfg.WikiPath == "" {
		return fmt.Errorf(
			"wiki_path is required but not set\n\n" +
				"Set it via any of:\n" +
				"  - TOML file:   wiki_path = \"/path/to/wiki\"\n" +
				"  - Env var:     WIKI_MCP_WIKI_PATH=/path/to/wiki\n" +
				"  - CLI flag:    --wiki-path /path/to/wiki",
		)
	}

	abs, err := filepath.Abs(cfg.WikiPath)
	if err != nil {
		return fmt.Errorf("cannot resolve wiki_path %q: %w", cfg.WikiPath, err)
	}
	cfg.WikiPath = filepath.Clean(abs)

	if cfg.SourcesPath == "" {
		cfg.SourcesPath = filepath.Join(filepath.Dir(cfg.WikiPath), "sources")
	}

	return nil
}
