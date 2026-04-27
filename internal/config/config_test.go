package config

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func chdirTemp(t *testing.T, dir string) {
	t.Helper()
	origDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}
	t.Cleanup(func() {
		if chdirErr := os.Chdir(origDir); chdirErr != nil {
			t.Errorf("restore cwd: %v", chdirErr)
		}
	})
	if err := os.Chdir(dir); err != nil {
		t.Fatalf("chdir to temp dir: %v", err)
	}
}

func TestDefaults(t *testing.T) {
	d := Defaults()
	if d.Web.Port != 9000 {
		t.Errorf("default port = %d, want 9000", d.Web.Port)
	}
	if d.Web.Bind != "127.0.0.1" {
		t.Errorf("default bind = %q, want 127.0.0.1", d.Web.Bind)
	}
	if d.Safety.ConfineToWikiPath != true {
		t.Error("default ConfineToWikiPath should be true")
	}
	if d.Safety.MaxPageBytes != 1048576 {
		t.Errorf("default MaxPageBytes = %d, want 1048576", d.Safety.MaxPageBytes)
	}
	if len(d.Index.Sections) != 4 {
		t.Errorf("default sections count = %d, want 4", len(d.Index.Sections))
	}
}

func TestLoadFromTOMLFile(t *testing.T) {
	dir := t.TempDir()
	wikiDir := filepath.Join(dir, "wiki")
	if err := os.MkdirAll(wikiDir, 0o755); err != nil {
		t.Fatal(err)
	}

	tomlContent := `
wiki_path = "` + filepath.ToSlash(wikiDir) + `"

[web]
port = 8080
bind = "0.0.0.0"
`
	tomlPath := filepath.Join(dir, "wiki-mcp.toml")
	if err := os.WriteFile(tomlPath, []byte(tomlContent), 0o644); err != nil {
		t.Fatal(err)
	}

	chdirTemp(t, dir)

	cfg, err := Load(Flags{})
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if cfg.Web.Port != 8080 {
		t.Errorf("port = %d, want 8080", cfg.Web.Port)
	}
	if cfg.Web.Bind != "0.0.0.0" {
		t.Errorf("bind = %q, want 0.0.0.0", cfg.Web.Bind)
	}
	if cfg.Web.Theme != "default" {
		t.Errorf("theme = %q, want default", cfg.Web.Theme)
	}
}

func TestMissingTOMLFileFallsThrough(t *testing.T) {
	dir := t.TempDir()
	chdirTemp(t, dir)

	wikiDir := filepath.Join(dir, "wiki")
	if err := os.MkdirAll(wikiDir, 0o755); err != nil {
		t.Fatal(err)
	}
	wp := wikiDir
	_, err := Load(Flags{WikiPath: &wp})
	if err != nil {
		t.Fatalf("Load should succeed with no TOML file: %v", err)
	}
}

func TestEnvOverridesFile(t *testing.T) {
	dir := t.TempDir()
	wikiDir := filepath.Join(dir, "wiki")
	if err := os.MkdirAll(wikiDir, 0o755); err != nil {
		t.Fatal(err)
	}

	tomlContent := `
wiki_path = "` + filepath.ToSlash(wikiDir) + `"

[web]
port = 8080
`
	tomlPath := filepath.Join(dir, "wiki-mcp.toml")
	if err := os.WriteFile(tomlPath, []byte(tomlContent), 0o644); err != nil {
		t.Fatal(err)
	}

	chdirTemp(t, dir)

	t.Setenv("WIKI_MCP_WEB_PORT", "3000")

	cfg, err := Load(Flags{})
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if cfg.Web.Port != 3000 {
		t.Errorf("port = %d, want 3000 (env override)", cfg.Web.Port)
	}
}

func TestFlagsOverrideEnv(t *testing.T) {
	dir := t.TempDir()
	wikiDir := filepath.Join(dir, "wiki")
	if err := os.MkdirAll(wikiDir, 0o755); err != nil {
		t.Fatal(err)
	}

	chdirTemp(t, dir)

	t.Setenv("WIKI_MCP_WIKI_PATH", "/should/be/overridden")

	wp := wikiDir
	port := 7777
	cfg, err := Load(Flags{WikiPath: &wp, Port: &port})
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	absWiki, _ := filepath.Abs(wikiDir)
	if cfg.WikiPath != filepath.Clean(absWiki) {
		t.Errorf("WikiPath = %q, want %q (flag override)", cfg.WikiPath, absWiki)
	}
	if cfg.Web.Port != 7777 {
		t.Errorf("port = %d, want 7777 (flag override)", cfg.Web.Port)
	}
}

func TestWikiPathRequired(t *testing.T) {
	dir := t.TempDir()
	chdirTemp(t, dir)

	_, err := Load(Flags{})
	if err == nil {
		t.Fatal("Load should fail when wiki_path not set")
	}
	// Error message should mention all three ways to set it
	msg := err.Error()
	for _, want := range []string{"WIKI_MCP_WIKI_PATH", "--wiki-path", "wiki_path"} {
		if !strings.Contains(msg, want) {
			t.Errorf("error message should mention %q, got: %s", want, msg)
		}
	}
}

func TestInvalidTOML(t *testing.T) {
	dir := t.TempDir()
	bad := filepath.Join(dir, "bad.toml")
	if err := os.WriteFile(bad, []byte("not valid [[ toml"), 0o644); err != nil {
		t.Fatal(err)
	}

	chdirTemp(t, dir)

	t.Setenv("WIKI_MCP_CONFIG", bad)
	_, err := Load(Flags{})
	if err == nil {
		t.Fatal("Load should fail on invalid TOML")
	}
}

func TestEnvConfigFileNotFound(t *testing.T) {
	dir := t.TempDir()
	chdirTemp(t, dir)

	// WIKI_MCP_CONFIG pointing at nonexistent file should not error
	// (loadTOMLFile returns nil for missing files)
	t.Setenv("WIKI_MCP_CONFIG", filepath.Join(dir, "nonexistent.toml"))
	wp := filepath.Join(dir, "wiki")
	if err := os.MkdirAll(wp, 0o755); err != nil {
		t.Fatal(err)
	}
	_, err := Load(Flags{WikiPath: &wp})
	if err != nil {
		t.Fatalf("Load should succeed when WIKI_MCP_CONFIG points to missing file: %v", err)
	}
}

func TestConfigFlagFileNotFound(t *testing.T) {
	dir := t.TempDir()
	chdirTemp(t, dir)

	// --config pointing at nonexistent file should not error
	missing := filepath.Join(dir, "missing.toml")
	_, err := Load(Flags{ConfigFile: &missing})
	// Actually this should still require wiki_path, but config file
	// itself being missing is silently skipped
	if err == nil {
		t.Fatal("should still fail for missing wiki_path")
	}
}

func TestResolveWikiPath(t *testing.T) {
	root := filepath.Join("home", "user", "wiki")
	tests := []struct {
		name    string
		rel     string
		confine bool
		wantErr bool
		want    string
	}{
		{
			name:    "simple subpath",
			rel:     "pages/intro.md",
			confine: true,
			want:    filepath.Join(root, "pages", "intro.md"),
		},
		{
			name:    "dot path",
			rel:     "./pages/../pages/intro.md",
			confine: true,
			want:    filepath.Join(root, "pages", "intro.md"),
		},
		{
			name:    "traversal blocked",
			rel:     filepath.Join("..", "etc", "passwd"),
			confine: true,
			wantErr: true,
		},
		{
			name:    "traversal with dot segments",
			rel:     filepath.Join("pages", "..", "..", "etc", "passwd"),
			confine: true,
			wantErr: true,
		},
		{
			name:    "traversal allowed when confine off",
			rel:     filepath.Join("..", "etc", "passwd"),
			confine: false,
			want:    filepath.Clean(filepath.Join(root, "..", "etc", "passwd")),
		},
		{
			name:    "root itself",
			rel:     ".",
			confine: true,
			want:    root,
		},
		{
			name:    "empty rel",
			rel:     "",
			confine: true,
			want:    root,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &Config{
				WikiPath:    root,
				ProjectPath: root, // pin tests to root so Root() == root
				Safety: SafetyConfig{
					ConfineToWikiPath: tt.confine,
				},
			}
			got, err := cfg.ResolveWikiPath(tt.rel)
			if tt.wantErr {
				if err == nil {
					t.Errorf("expected error for rel=%q, got path=%q", tt.rel, got)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tt.want {
				t.Errorf("ResolveWikiPath(%q) = %q, want %q", tt.rel, got, tt.want)
			}
		})
	}
}

func TestMustMutate(t *testing.T) {
	cfg := &Config{Safety: SafetyConfig{ReadOnly: true}}
	err := cfg.MustMutate()
	if !errors.Is(err, ErrReadOnly) {
		t.Errorf("MustMutate() = %v, want ErrReadOnly", err)
	}

	cfg.Safety.ReadOnly = false
	if err := cfg.MustMutate(); err != nil {
		t.Errorf("MustMutate() should return nil when not read-only: %v", err)
	}
}

func TestSourcesPathDefault(t *testing.T) {
	dir := t.TempDir()
	wikiDir := filepath.Join(dir, "wiki")
	if err := os.MkdirAll(wikiDir, 0o755); err != nil {
		t.Fatal(err)
	}

	chdirTemp(t, dir)

	wp := wikiDir
	cfg, err := Load(Flags{WikiPath: &wp})
	if err != nil {
		t.Fatalf("Load: %v", err)
	}

	want := filepath.Join(filepath.Dir(cfg.WikiPath), "sources")
	if cfg.SourcesPath != want {
		t.Errorf("SourcesPath = %q, want %q", cfg.SourcesPath, want)
	}
}

func TestXDGConfigPath(t *testing.T) {
	dir := t.TempDir()
	xdgDir := filepath.Join(dir, "xdg")
	wikiDir := filepath.Join(dir, "wiki")
	if err := os.MkdirAll(wikiDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(xdgDir, "wiki-mcp"), 0o755); err != nil {
		t.Fatal(err)
	}

	toml := `wiki_path = "` + filepath.ToSlash(wikiDir) + `"
[web]
port = 4444
`
	if err := os.WriteFile(filepath.Join(xdgDir, "wiki-mcp", "config.toml"), []byte(toml), 0o644); err != nil {
		t.Fatal(err)
	}

	chdirTemp(t, dir) // no CWD file

	t.Setenv("XDG_CONFIG_HOME", xdgDir)

	cfg, err := Load(Flags{})
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if cfg.Web.Port != 4444 {
		t.Errorf("port = %d, want 4444 from XDG config", cfg.Web.Port)
	}
}

func TestEnvScalarOverrides(t *testing.T) {
	dir := t.TempDir()
	wikiDir := filepath.Join(dir, "wiki")
	if err := os.MkdirAll(wikiDir, 0o755); err != nil {
		t.Fatal(err)
	}

	chdirTemp(t, dir)

	t.Setenv("WIKI_MCP_WIKI_PATH", wikiDir)
	t.Setenv("WIKI_MCP_WEB_BIND", "0.0.0.0")
	t.Setenv("WIKI_MCP_WEB_THEME", "minimal")
	t.Setenv("WIKI_MCP_LINKS_STYLE", "obsidian")
	t.Setenv("WIKI_MCP_SAFETY_READ_ONLY", "true")
	t.Setenv("WIKI_MCP_SAFETY_CONFINE", "false")

	cfg, err := Load(Flags{})
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if cfg.Web.Bind != "0.0.0.0" {
		t.Errorf("bind = %q, want 0.0.0.0", cfg.Web.Bind)
	}
	if cfg.Web.Theme != "minimal" {
		t.Errorf("theme = %q, want minimal", cfg.Web.Theme)
	}
	if cfg.Links.Style != "obsidian" {
		t.Errorf("links style = %q, want obsidian", cfg.Links.Style)
	}
	if !cfg.Safety.ReadOnly {
		t.Error("ReadOnly should be true")
	}
	if cfg.Safety.ConfineToWikiPath {
		t.Error("ConfineToWikiPath should be false")
	}
}

func TestRoot_NoProject(t *testing.T) {
	cfg := &Config{WikiPath: "/wiki"}
	want := filepath.Join("/wiki", "default")
	if cfg.Root() != want {
		t.Errorf("Root() = %q, want %q", cfg.Root(), want)
	}
}

func TestRoot_WithProject(t *testing.T) {
	cfg := &Config{WikiPath: "/wiki", ProjectPath: "/wiki/myproject"}
	if cfg.Root() != "/wiki/myproject" {
		t.Errorf("Root() = %q, want /wiki/myproject", cfg.Root())
	}
}

func TestProjectPath_MustBeWithinWikiPath(t *testing.T) {
	dir := t.TempDir()
	wikiDir := filepath.Join(dir, "wiki")
	outsideDir := filepath.Join(dir, "outside")
	for _, d := range []string{wikiDir, outsideDir} {
		if err := os.MkdirAll(d, 0o755); err != nil {
			t.Fatal(err)
		}
	}

	wp := wikiDir
	pp := outsideDir
	_, err := Load(Flags{WikiPath: &wp, ProjectPath: &pp})
	if err == nil {
		t.Fatal("expected error for project_path outside wiki_path")
	}
	if !strings.Contains(err.Error(), "must be within") {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestProjectPath_ValidSubdir(t *testing.T) {
	dir := t.TempDir()
	wikiDir := filepath.Join(dir, "wiki")
	projectDir := filepath.Join(wikiDir, "myproject")
	if err := os.MkdirAll(projectDir, 0o755); err != nil {
		t.Fatal(err)
	}

	wp := wikiDir
	pp := projectDir
	cfg, err := Load(Flags{WikiPath: &wp, ProjectPath: &pp})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Root() != projectDir {
		t.Errorf("Root() = %q, want %q", cfg.Root(), projectDir)
	}
}

func TestProjectPath_RelativeResolvesAgainstWikiPath(t *testing.T) {
	dir := t.TempDir()
	wikiDir := filepath.Join(dir, "wiki")
	projectDir := filepath.Join(wikiDir, "myproject")
	if err := os.MkdirAll(projectDir, 0o755); err != nil {
		t.Fatal(err)
	}

	wp := wikiDir
	pp := "myproject" // relative — should be resolved against wiki_path
	cfg, err := Load(Flags{WikiPath: &wp, ProjectPath: &pp})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Root() != projectDir {
		t.Errorf("Root() = %q, want %q", cfg.Root(), projectDir)
	}
}

func TestProjectPath_RelativeNestedResolvesAgainstWikiPath(t *testing.T) {
	dir := t.TempDir()
	wikiDir := filepath.Join(dir, "wiki")
	projectDir := filepath.Join(wikiDir, "research", "2026")
	if err := os.MkdirAll(projectDir, 0o755); err != nil {
		t.Fatal(err)
	}

	wp := wikiDir
	pp := "research/2026"
	cfg, err := Load(Flags{WikiPath: &wp, ProjectPath: &pp})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Root() != projectDir {
		t.Errorf("Root() = %q, want %q", cfg.Root(), projectDir)
	}
}

func TestProjectPath_RelativeEscapeRejected(t *testing.T) {
	dir := t.TempDir()
	wikiDir := filepath.Join(dir, "wiki")
	if err := os.MkdirAll(wikiDir, 0o755); err != nil {
		t.Fatal(err)
	}

	wp := wikiDir
	pp := "../outside"
	_, err := Load(Flags{WikiPath: &wp, ProjectPath: &pp})
	if err == nil {
		t.Fatal("expected error for relative path escaping wiki_path")
	}
	if !strings.Contains(err.Error(), "must be within") {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestSubProjectPath_Validation(t *testing.T) {
	dir := t.TempDir()
	wikiDir := filepath.Join(dir, "wiki")
	projectDir := filepath.Join(wikiDir, "msb-cb")
	subProjectDir := filepath.Join(projectDir, "mappers")

	wp := wikiDir
	pp := projectDir
	sp := subProjectDir

	t.Run("valid", func(t *testing.T) {
		cfg, err := Load(Flags{WikiPath: &wp, ProjectPath: &pp, SubProjectPath: &sp})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if cfg.SubProjectPath != subProjectDir {
			t.Errorf("SubProjectPath = %q, want %q", cfg.SubProjectPath, subProjectDir)
		}
		if cfg.Root() != subProjectDir {
			t.Errorf("Root() = %q, want %q", cfg.Root(), subProjectDir)
		}
	})

	t.Run("requires project_path", func(t *testing.T) {
		_, err := Load(Flags{WikiPath: &wp, SubProjectPath: &sp})
		if err == nil || !strings.Contains(err.Error(), "requires project_path") {
			t.Errorf("expected requires project_path error, got: %v", err)
		}
	})

	t.Run("must differ from project_path", func(t *testing.T) {
		_, err := Load(Flags{WikiPath: &wp, ProjectPath: &pp, SubProjectPath: &pp})
		if err == nil || !strings.Contains(err.Error(), "must differ") {
			t.Errorf("expected must differ error, got: %v", err)
		}
	})

	t.Run("must be within project_path", func(t *testing.T) {
		outside := filepath.Join(wikiDir, "other-project", "mappers")
		_, err := Load(Flags{WikiPath: &wp, ProjectPath: &pp, SubProjectPath: &outside})
		if err == nil || !strings.Contains(err.Error(), "must be within") {
			t.Errorf("expected must be within error, got: %v", err)
		}
	})
}

func TestMustAllowWrite_SubProject(t *testing.T) {
	dir := t.TempDir()
	wikiDir := filepath.Join(dir, "wiki")
	projectDir := filepath.Join(wikiDir, "msb-cb")
	subProjectDir := filepath.Join(projectDir, "mappers")
	siblingDir := filepath.Join(projectDir, "other")

	wp := wikiDir
	pp := projectDir
	sp := subProjectDir
	cfg, err := Load(Flags{WikiPath: &wp, ProjectPath: &pp, SubProjectPath: &sp})
	if err != nil {
		t.Fatalf("load: %v", err)
	}

	cases := []struct {
		name    string
		path    string
		wantErr bool
	}{
		{"own sub-project file", filepath.Join(subProjectDir, "foo.md"), false},
		{"own sub-project nested", filepath.Join(subProjectDir, "sub", "bar.md"), false},
		{"parent direct file", filepath.Join(projectDir, "readme.md"), false},
		{"sibling sub-project", filepath.Join(siblingDir, "foo.md"), true},
		{"wiki root file", filepath.Join(wikiDir, "foo.md"), true},
		{"parent subdir (sibling)", filepath.Join(projectDir, "other", "deep", "x.md"), true},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := cfg.MustAllowWrite(tc.path)
			if tc.wantErr && err == nil {
				t.Error("expected error, got nil")
			}
			if !tc.wantErr && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}

func TestMustAllowWrite_NoSubProject(t *testing.T) {
	dir := t.TempDir()
	wikiDir := filepath.Join(dir, "wiki")
	wp := wikiDir
	cfg, err := Load(Flags{WikiPath: &wp})
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	// Without sub-project all paths are allowed.
	if err := cfg.MustAllowWrite("/any/path/at/all"); err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestResolveWikiPath_SubProjectBoundary(t *testing.T) {
	dir := t.TempDir()
	wikiDir := filepath.Join(dir, "wiki")
	projectDir := filepath.Join(wikiDir, "msb-cb")
	subProjectDir := filepath.Join(projectDir, "mappers")

	wp := wikiDir
	pp := projectDir
	sp := subProjectDir
	cfg, err := Load(Flags{WikiPath: &wp, ProjectPath: &pp, SubProjectPath: &sp})
	if err != nil {
		t.Fatalf("load: %v", err)
	}

	// Relative path to parent file via ".." should resolve successfully.
	abs, err := cfg.ResolveWikiPath("../parent.md")
	if err != nil {
		t.Errorf("unexpected error resolving ../parent.md: %v", err)
	}
	if abs != filepath.Join(projectDir, "parent.md") {
		t.Errorf("resolved = %q, want %q", abs, filepath.Join(projectDir, "parent.md"))
	}

	// Path escaping above ProjectPath must be rejected.
	_, err = cfg.ResolveWikiPath("../../escape.md")
	if err == nil {
		t.Error("expected error for path escaping project boundary, got nil")
	}
}
