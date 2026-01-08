package auditlog

import (
	"context"
	"fmt"
	"os/exec"
	"strings"

	"github.com/indaco/sley/internal/core"
)

// DefaultGitOps implements GitOperations using git commands.
type DefaultGitOps struct{}

// GetAuthor returns the git user name and email.
func (g *DefaultGitOps) GetAuthor() (string, error) {
	name, err := g.runGitCommand("config", "user.name")
	if err != nil {
		return "", err
	}

	email, err := g.runGitCommand("config", "user.email")
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("%s <%s>", name, email), nil
}

// GetCommitSHA returns the current commit SHA.
func (g *DefaultGitOps) GetCommitSHA() (string, error) {
	return g.runGitCommand("rev-parse", "HEAD")
}

// GetBranch returns the current branch name.
func (g *DefaultGitOps) GetBranch() (string, error) {
	return g.runGitCommand("rev-parse", "--abbrev-ref", "HEAD")
}

// runGitCommand executes a git command and returns the trimmed output.
func (g *DefaultGitOps) runGitCommand(args ...string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), core.TimeoutShort)
	defer cancel()

	cmd := exec.CommandContext(ctx, "git", args...)
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("git command failed: %w", err)
	}
	return strings.TrimSpace(string(output)), nil
}
