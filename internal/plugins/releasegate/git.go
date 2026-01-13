package releasegate

import (
	"bytes"
	"fmt"
	"os/exec"
	"strings"
)

// GitOperations defines the interface for git operations used by release gate.
type GitOperations interface {
	// IsWorktreeClean checks if the git working tree has uncommitted changes.
	IsWorktreeClean() (bool, error)

	// GetCurrentBranch retrieves the current git branch name.
	GetCurrentBranch() (string, error)

	// GetRecentCommits retrieves the last N commit messages.
	GetRecentCommits(count int) ([]string, error)
}

// OSGitOperations implements GitOperations using actual git commands.
type OSGitOperations struct {
	execCommand func(name string, arg ...string) *exec.Cmd
}

// NewOSGitOperations creates a new OSGitOperations with the default exec.Command.
func NewOSGitOperations() *OSGitOperations {
	return &OSGitOperations{
		execCommand: exec.Command,
	}
}

// Verify OSGitOperations implements GitOperations.
var _ GitOperations = (*OSGitOperations)(nil)

// IsWorktreeClean checks if the git working tree has uncommitted changes.
// Returns true if the working tree is clean (no uncommitted changes).
func (g *OSGitOperations) IsWorktreeClean() (bool, error) {
	cmd := g.execCommand("git", "status", "--porcelain")
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		stderrMsg := strings.TrimSpace(stderr.String())
		if stderrMsg != "" {
			return false, fmt.Errorf("failed to check git status: %s: %w", stderrMsg, err)
		}
		return false, fmt.Errorf("failed to check git status: %w", err)
	}

	// Empty output means clean working tree
	output := strings.TrimSpace(stdout.String())
	return output == "", nil
}

// GetCurrentBranch retrieves the current git branch name.
func (g *OSGitOperations) GetCurrentBranch() (string, error) {
	cmd := g.execCommand("git", "rev-parse", "--abbrev-ref", "HEAD")
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		stderrMsg := strings.TrimSpace(stderr.String())
		if stderrMsg != "" {
			return "", fmt.Errorf("failed to get current branch: %s: %w", stderrMsg, err)
		}
		return "", fmt.Errorf("failed to get current branch: %w", err)
	}

	branch := strings.TrimSpace(stdout.String())
	if branch == "" {
		return "", fmt.Errorf("failed to determine current branch")
	}

	return branch, nil
}

// GetRecentCommits retrieves the last N commit messages.
func (g *OSGitOperations) GetRecentCommits(count int) ([]string, error) {
	if count <= 0 {
		count = 10
	}

	cmd := g.execCommand("git", "log", fmt.Sprintf("-n%d", count), "--oneline", "--no-decorate")
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		stderrMsg := strings.TrimSpace(stderr.String())
		if stderrMsg != "" {
			return nil, fmt.Errorf("failed to get commit history: %s: %w", stderrMsg, err)
		}
		return nil, fmt.Errorf("failed to get commit history: %w", err)
	}

	output := strings.TrimSpace(stdout.String())
	if output == "" {
		return []string{}, nil
	}

	commits := strings.Split(output, "\n")
	return commits, nil
}
