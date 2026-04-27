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
	Enabled     bool   `toml:"enabled"`
	Port        int    `toml:"port"`
	Bind        string `toml:"bind"`
	Theme       string `toml:"theme"`
	AutoRebuild bool   `toml:"auto_rebuild"`
}

// MCPConfig controls the HTTP-based MCP transport (streamable-HTTP, MCP 2025-03 spec).
type MCPConfig struct {
	Port      int    `toml:"port"`
	Bind      string `toml:"bind"`
	AuthToken string `toml:"auth_token"`
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
	WikiPath       string       `toml:"wiki_path"`
	ProjectPath    string       `toml:"project_path"`
	SubProjectPath string       `toml:"sub_project_path"`
	SourcesPath    string       `toml:"sources_path"`
	Web            WebConfig    `toml:"web"`
	MCP            MCPConfig    `toml:"mcp"`
	Index          IndexConfig  `toml:"index"`
	Log            LogConfig    `toml:"log"`
	Links          LinksConfig  `toml:"links"`
	Safety         SafetyConfig `toml:"safety"`
}

// Flags holds CLI flag values that were explicitly set by the user.
// nil pointer means "not set by user."
type Flags struct {
	ConfigFile     *string
	WikiPath       *string
	ProjectPath    *string
	SubProjectPath *string
	Port           *int
	Bind           *string
	MCPPort        *int
	AuthToken      *string
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
		MCP: MCPConfig{
			Port: 8765,
			Bind: "127.0.0.1",
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
	if flags.ProjectPath != nil {
		cfg.ProjectPath = *flags.ProjectPath
	}
	if flags.SubProjectPath != nil {
		cfg.SubProjectPath = *flags.SubProjectPath
	}
	if flags.Port != nil {
		cfg.Web.Port = *flags.Port
	}
	if flags.Bind != nil {
		cfg.Web.Bind = *flags.Bind
		cfg.MCP.Bind = *flags.Bind
	}
	if flags.MCPPort != nil {
		cfg.MCP.Port = *flags.MCPPort
	}
	if flags.AuthToken != nil {
		cfg.MCP.AuthToken = *flags.AuthToken
	}

	if err := validate(&cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}

// Root returns the effective working root for wiki tools. When a sub-project
// path is configured it returns that; when a project path is configured it
// returns that; otherwise it returns the "default" subdirectory of the wiki
// root so that the wiki root itself remains a meta-level container.
func (c *Config) Root() string {
	if c.SubProjectPath != "" {
		return c.SubProjectPath
	}
	if c.ProjectPath != "" {
		return c.ProjectPath
	}
	return filepath.Join(c.WikiPath, "default")
}

// ResolveWikiPath joins rel to the effective root and returns a cleaned
// absolute path. If ConfineToWikiPath is true and the result escapes the
// confinement boundary, an error is returned.
//
// When a sub-project is active the confinement boundary is ProjectPath (the
// parent project), not SubProjectPath. This lets sub-projects address parent
// files via "../file.md" while still being blocked from escaping above the
// parent. MustAllowWrite enforces sibling exclusion on write operations.
func (c *Config) ResolveWikiPath(rel string) (string, error) {
	root := c.Root()
	resolved := filepath.Join(root, rel)

	if c.Safety.ConfineToWikiPath {
		boundary := root
		if c.SubProjectPath != "" {
			boundary = c.ProjectPath
		}
		if resolved != boundary && !strings.HasPrefix(resolved, boundary+string(os.PathSeparator)) {
			return "", fmt.Errorf("path %q escapes wiki root %q", rel, boundary)
		}
	}

	return resolved, nil
}

// MustAllowWrite returns ErrCodeForbidden when a sub-project is active and
// absPath falls outside the allowed write scope:
//   - within SubProjectPath (own scope, any depth), or
//   - a direct child of ProjectPath (parent project top-level files only).
//
// When no sub-project is configured every path is allowed.
func (c *Config) MustAllowWrite(absPath string) error {
	if c.SubProjectPath == "" {
		return nil
	}
	// Own sub-project scope.
	if absPath == c.SubProjectPath || strings.HasPrefix(absPath, c.SubProjectPath+string(os.PathSeparator)) {
		return nil
	}
	// Direct child of parent project (top-level file, not inside a sibling sub-project).
	if filepath.Dir(absPath) == c.ProjectPath {
		return nil
	}
	return fmt.Errorf("sub-project %q cannot write to %q: outside own scope and parent project root", c.SubProjectPath, absPath)
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
	envStr("WIKI_MCP_PROJECT_PATH", &cfg.ProjectPath)
	envStr("WIKI_MCP_SUB_PROJECT_PATH", &cfg.SubProjectPath)
	envStr("WIKI_MCP_SOURCES_PATH", &cfg.SourcesPath)
	envBool("WIKI_MCP_WEB_ENABLED", &cfg.Web.Enabled)
	envInt("WIKI_MCP_WEB_PORT", &cfg.Web.Port)
	envStr("WIKI_MCP_WEB_BIND", &cfg.Web.Bind)
	envStr("WIKI_MCP_WEB_THEME", &cfg.Web.Theme)
	envInt("WIKI_MCP_MCP_PORT", &cfg.MCP.Port)
	envStr("WIKI_MCP_MCP_BIND", &cfg.MCP.Bind)
	envStr("WIKI_MCP_MCP_AUTH_TOKEN", &cfg.MCP.AuthToken)
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

	if cfg.ProjectPath != "" {
		// Relative paths are resolved against wiki_path so callers can pass
		// just the subpath (e.g. "my-project") without repeating the wiki root.
		raw := cfg.ProjectPath
		if !filepath.IsAbs(raw) {
			raw = filepath.Join(cfg.WikiPath, raw)
		}
		absProject, err := filepath.Abs(raw)
		if err != nil {
			return fmt.Errorf("cannot resolve project_path %q: %w", cfg.ProjectPath, err)
		}
		absProject = filepath.Clean(absProject)
		if absProject != cfg.WikiPath && !strings.HasPrefix(absProject, cfg.WikiPath+string(os.PathSeparator)) {
			return fmt.Errorf("project_path %q must be within wiki_path %q", absProject, cfg.WikiPath)
		}
		cfg.ProjectPath = absProject
	}

	if cfg.SubProjectPath != "" {
		if cfg.ProjectPath == "" {
			return fmt.Errorf("sub_project_path requires project_path to be set")
		}
		raw := cfg.SubProjectPath
		if !filepath.IsAbs(raw) {
			raw = filepath.Join(cfg.ProjectPath, raw)
		}
		absSubProject, err := filepath.Abs(raw)
		if err != nil {
			return fmt.Errorf("cannot resolve sub_project_path %q: %w", cfg.SubProjectPath, err)
		}
		absSubProject = filepath.Clean(absSubProject)
		if absSubProject == cfg.ProjectPath {
			return fmt.Errorf("sub_project_path %q must differ from project_path %q", absSubProject, cfg.ProjectPath)
		}
		if !strings.HasPrefix(absSubProject, cfg.ProjectPath+string(os.PathSeparator)) {
			return fmt.Errorf("sub_project_path %q must be within project_path %q", absSubProject, cfg.ProjectPath)
		}
		cfg.SubProjectPath = absSubProject
	}

	if cfg.SourcesPath == "" {
		cfg.SourcesPath = filepath.Join(filepath.Dir(cfg.WikiPath), "sources")
	}

	return nil
}
