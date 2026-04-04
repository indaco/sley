package gitlog

import (
	"bytes"
	"fmt"
	"os/exec"
	"regexp"
	"strings"

	"github.com/indaco/sley/internal/git"
)

// validGitRef matches safe git reference names: alphanumeric, dots, hyphens, slashes, tildes, carets.
// Rejects shell metacharacters, spaces, and ".." sequences within ref names.
var validGitRef = regexp.MustCompile(`^[a-zA-Z0-9._/~^@{}\-]+$`)

// validateGitRef checks that a git reference is safe to use in a revision range.
func validateGitRef(ref string) error {
	if ref == "" {
		return fmt.Errorf("git reference cannot be empty")
	}
	if strings.Contains(ref, "..") {
		return fmt.Errorf("git reference %q contains invalid '..' sequence", ref)
	}
	if !validGitRef.MatchString(ref) {
		return fmt.Errorf("git reference %q contains invalid characters", ref)
	}
	return nil
}

// ExecCommandFunc is the function signature matching exec.Command.
type ExecCommandFunc func(name string, arg ...string) *exec.Cmd

// GitLog retrieves commit messages from a git repository.
type GitLog struct {
	ExecCommandFn ExecCommandFunc
}

// NewGitLog creates a GitLog with the real exec.Command implementation.
func NewGitLog() *GitLog {
	return &GitLog{ExecCommandFn: exec.Command}
}

// GetCommitsFn is the function type for GetCommits (used for dependency injection).
type GetCommitsFn func(since, until string) ([]string, error)

// DefaultGetCommitsFn returns a GetCommitsFn backed by a real GitLog instance.
func DefaultGetCommitsFn() GetCommitsFn {
	return NewGitLog().GetCommits
}

// GetCommits returns commit messages between since and until refs.
func (g *GitLog) GetCommits(since string, until string) ([]string, error) {
	if until == "" {
		until = "HEAD"
	}

	if since == "" {
		lastTag, err := g.getLastTag()
		if err != nil {
			// No tags found - fall back to a safe recent range.
			// Use HEAD~10 if enough commits exist, otherwise use the repo root.
			since = git.SafeFallbackSince(g.ExecCommandFn, 10)
		} else {
			since = lastTag
		}
	}

	// Validate git references to prevent unexpected behavior
	if err := validateGitRef(since); err != nil {
		return nil, fmt.Errorf("invalid 'since' reference: %w", err)
	}
	if err := validateGitRef(until); err != nil {
		return nil, fmt.Errorf("invalid 'until' reference: %w", err)
	}

	revRange := since + ".." + until
	cmd := g.ExecCommandFn("git", "log", "--pretty=format:%s", revRange)

	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	output, err := cmd.Output()
	if err != nil {
		stderrMsg := strings.TrimSpace(stderr.String())
		if stderrMsg != "" {
			return nil, fmt.Errorf("git log failed: %s: %w", stderrMsg, err)
		}
		return nil, fmt.Errorf("git log failed: %w", err)
	}

	lines := strings.Split(strings.TrimSpace(string(output)), "\n")
	if len(lines) == 1 && lines[0] == "" {
		return []string{}, nil
	}
	return lines, nil
}

func (g *GitLog) getLastTag() (string, error) {
	cmd := g.ExecCommandFn("git", "describe", "--tags", "--abbrev=0")
	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	out, err := cmd.Output()
	if err != nil {
		stderrMsg := strings.TrimSpace(stderr.String())
		if stderrMsg != "" {
			return "", fmt.Errorf("git describe failed: %s: %w", stderrMsg, err)
		}
		return "", fmt.Errorf("git describe failed: %w", err)
	}

	tag := strings.TrimSpace(string(out))
	if tag == "" {
		return "", fmt.Errorf("no tags found")
	}

	return tag, nil
}
