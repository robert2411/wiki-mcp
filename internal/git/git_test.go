package git_test

import (
	"context"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"

	internalgit "github.com/robert2411/wiki-mcp/internal/git"
)

func skipIfNoGit(t *testing.T) {
	t.Helper()
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not on PATH")
	}
}

func initRepo(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	mustGit(t, dir, "init")
	mustGit(t, dir, "config", "user.email", "test@example.com")
	mustGit(t, dir, "config", "user.name", "Test")
	return dir
}

func mustGit(t *testing.T, dir string, args ...string) {
	t.Helper()
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("git %v: %v\n%s", args, err, out)
	}
}

func TestCheckPrerequisites_NotARepo(t *testing.T) {
	skipIfNoGit(t)
	dir := t.TempDir()
	if err := internalgit.CheckPrerequisites(dir); err == nil {
		t.Error("expected error for non-git directory")
	}
}

func TestCheckPrerequisites_ValidRepo(t *testing.T) {
	skipIfNoGit(t)
	dir := initRepo(t)
	if err := internalgit.CheckPrerequisites(dir); err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestCommitAll_NoChanges(t *testing.T) {
	skipIfNoGit(t)
	dir := initRepo(t)
	// Initial commit so HEAD exists.
	mustGit(t, dir, "commit", "--allow-empty", "-m", "init")

	logger := slog.New(slog.NewTextHandler(os.Stderr, nil))
	c := internalgit.New(dir, time.Hour, false, "", "", logger)
	// Should not panic or error — nothing to commit.
	c.CommitAll(context.Background())
}

func TestCommitAll_WithChanges(t *testing.T) {
	skipIfNoGit(t)
	dir := initRepo(t)
	mustGit(t, dir, "commit", "--allow-empty", "-m", "init")

	// Write a file.
	if err := os.WriteFile(filepath.Join(dir, "page.md"), []byte("# Hello"), 0o644); err != nil {
		t.Fatal(err)
	}

	logger := slog.New(slog.NewTextHandler(os.Stderr, nil))
	c := internalgit.New(dir, time.Hour, false, "Bot", "bot@example.com", logger)
	c.CommitAll(context.Background())

	// Confirm the file is committed.
	cmd := exec.Command("git", "show", "--name-only", "--format=", "HEAD")
	cmd.Dir = dir
	out, err := cmd.Output()
	if err != nil {
		t.Fatalf("git show: %v", err)
	}
	if string(out) == "" || filepath.Base(string(out)) == "" {
		t.Error("expected committed file in last commit")
	}
}

func TestCommitAll_PushNoRemote(t *testing.T) {
	skipIfNoGit(t)
	dir := initRepo(t)
	mustGit(t, dir, "commit", "--allow-empty", "-m", "init")

	if err := os.WriteFile(filepath.Join(dir, "a.md"), []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}

	logger := slog.New(slog.NewTextHandler(os.Stderr, nil))
	c := internalgit.New(dir, time.Hour, true, "", "", logger)
	// push=true but no remote configured — should not panic.
	c.CommitAll(context.Background())
}
