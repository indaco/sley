package git

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/indaco/sley/internal/cmdrunner"
	"github.com/indaco/sley/internal/core"
)

// CloneOrUpdate clones a repository if it doesn't exist, or updates it if it does.
func CloneOrUpdate(ctx context.Context, repoURL, repoPath string) error {
	if IsValidGitRepo(repoPath) {
		return UpdateRepo(ctx, repoPath)
	}
	return CloneRepo(ctx, repoURL, repoPath)
}

// UpdateRepo pulls the latest changes in an existing repository.
func UpdateRepo(ctx context.Context, repoPath string) error {
	// Apply default timeout if context has no deadline
	if _, hasDeadline := ctx.Deadline(); !hasDeadline {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, core.TimeoutGit)
		defer cancel()
	}
	return cmdrunner.RunCommandContext(ctx, repoPath, "git", "pull")
}

func CloneRepo(ctx context.Context, repoURL, repoPath string) error {
	// Apply default timeout if context has no deadline
	if _, hasDeadline := ctx.Deadline(); !hasDeadline {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, core.TimeoutGit)
		defer cancel()
	}
	return cmdrunner.RunCommandContext(ctx, ".", "git", "clone", repoURL, repoPath)
}

func ForceReclone(ctx context.Context, repoURL, repoPath string) error {
	if err := os.RemoveAll(repoPath); err != nil {
		return fmt.Errorf("failed to remove existing repository: %w", err)
	}
	return CloneRepo(ctx, repoURL, repoPath)
}

func IsValidGitRepo(repoPath string) bool {
	_, err := os.Stat(filepath.Join(repoPath, ".git"))
	return err == nil
}
