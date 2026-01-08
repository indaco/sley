package versionvalidator

import (
	"context"
	"os/exec"
	"strings"

	"github.com/indaco/sley/internal/core"
)

// OSGitBranchReader implements core.GitBranchReader using actual git commands.
type OSGitBranchReader struct{}

// NewOSGitBranchReader creates a new OSGitBranchReader.
func NewOSGitBranchReader() *OSGitBranchReader {
	return &OSGitBranchReader{}
}

// Verify OSGitBranchReader implements core.GitBranchReader.
var _ core.GitBranchReader = (*OSGitBranchReader)(nil)

// GetCurrentBranch returns the current git branch name.
func (g *OSGitBranchReader) GetCurrentBranch() (string, error) {
	cmd := exec.CommandContext(context.Background(), "git", "rev-parse", "--abbrev-ref", "HEAD")
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}

	return strings.TrimSpace(string(output)), nil
}

// defaultBranchReader is the default branch reader for backward compatibility.
var defaultBranchReader = NewOSGitBranchReader()
