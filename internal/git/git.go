// Package git provides a periodic auto-committer for the wiki directory.
// It uses the system git binary (os/exec) and never crashes on push failure.
package git

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

// Committer periodically stages, commits, and optionally pushes all changes
// in WikiPath. It requires a system git binary and that WikiPath is already
// inside a git repository.
type Committer struct {
	wikiPath    string
	interval    time.Duration
	push        bool
	authorName  string
	authorEmail string
	logger      *slog.Logger
}

// New returns a Committer configured with the given options.
func New(wikiPath string, interval time.Duration, push bool, authorName, authorEmail string, logger *slog.Logger) *Committer {
	return &Committer{
		wikiPath:    wikiPath,
		interval:    interval,
		push:        push,
		authorName:  authorName,
		authorEmail: authorEmail,
		logger:      logger,
	}
}

// CheckPrerequisites verifies the git binary is on PATH and that wikiPath is
// inside a git repository. Returns an error describing what is missing.
func CheckPrerequisites(wikiPath string) error {
	if _, err := exec.LookPath("git"); err != nil {
		return fmt.Errorf("git binary not found on PATH: %w", err)
	}
	out, err := runGit(context.Background(), wikiPath, "rev-parse", "--git-dir")
	if err != nil || strings.TrimSpace(string(out)) == "" {
		return fmt.Errorf("%q is not inside a git repository", wikiPath)
	}
	return nil
}

// Run starts the periodic commit loop. It blocks until ctx is cancelled.
func (c *Committer) Run(ctx context.Context) {
	ticker := time.NewTicker(c.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			c.CommitAll(ctx)
		}
	}
}

// CommitAll stages all changes under WikiPath, commits them, and (if push is
// enabled and a remote exists) pushes to the tracking remote. Push failures
// are logged but never returned as errors.
func (c *Committer) CommitAll(ctx context.Context) {
	dirty, err := c.isDirty(ctx)
	if err != nil {
		c.logger.Warn("git: status check failed", "err", err)
		return
	}
	if !dirty {
		return
	}

	if err := c.stage(ctx); err != nil {
		c.logger.Warn("git: stage failed", "err", err)
		return
	}

	// Re-check after staging: index might be empty if only untracked files were ignored.
	staged, err := c.hasStagedChanges(ctx)
	if err != nil {
		c.logger.Warn("git: staged check failed", "err", err)
		return
	}
	if !staged {
		return
	}

	if err := c.commit(ctx); err != nil {
		c.logger.Warn("git: commit failed", "err", err)
		return
	}

	c.logger.Info("git: committed wiki changes")

	if c.push {
		c.pushChanges(ctx)
	}
}

func (c *Committer) isDirty(ctx context.Context) (bool, error) {
	out, err := runGit(ctx, c.wikiPath, "status", "--porcelain")
	if err != nil {
		return false, err
	}
	return strings.TrimSpace(string(out)) != "", nil
}

func (c *Committer) stage(ctx context.Context) error {
	_, err := runGit(ctx, c.wikiPath, "add", "--all")
	return err
}

func (c *Committer) hasStagedChanges(ctx context.Context) (bool, error) {
	out, err := runGit(ctx, c.wikiPath, "diff", "--cached", "--name-only")
	if err != nil {
		return false, err
	}
	return strings.TrimSpace(string(out)) != "", nil
}

func (c *Committer) commit(ctx context.Context) error {
	args := []string{"commit", "-m", "chore: auto-commit wiki changes"}
	if c.authorName != "" && c.authorEmail != "" {
		args = append(args, fmt.Sprintf("--author=%s <%s>", c.authorName, c.authorEmail))
	}
	_, err := runGit(ctx, c.wikiPath, args...)
	return err
}

func (c *Committer) pushChanges(ctx context.Context) {
	hasRemote, err := c.hasRemote(ctx)
	if err != nil {
		c.logger.Warn("git: remote check failed", "err", err)
		return
	}
	if !hasRemote {
		return
	}

	if _, err := runGit(ctx, c.wikiPath, "push"); err != nil {
		c.logger.Warn("git: push failed (non-fatal)", "err", err)
		return
	}
	c.logger.Info("git: pushed wiki changes")
}

func (c *Committer) hasRemote(ctx context.Context) (bool, error) {
	out, err := runGit(ctx, c.wikiPath, "remote")
	if err != nil {
		return false, err
	}
	return strings.TrimSpace(string(out)) != "", nil
}

// runGit executes git with args in dir and returns combined output.
func runGit(ctx context.Context, dir string, args ...string) ([]byte, error) {
	cmd := exec.CommandContext(ctx, "git", args...) // #nosec G204 — args are caller-controlled, not user input
	cmd.Dir = filepath.Clean(dir)
	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &out

	if err := cmd.Run(); err != nil {
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) {
			return nil, fmt.Errorf("git %s: exit %d: %s", strings.Join(args, " "), exitErr.ExitCode(), strings.TrimSpace(out.String()))
		}
		return nil, err
	}
	return out.Bytes(), nil
}
