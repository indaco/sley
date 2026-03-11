package gitlog

import (
	"bytes"
	"fmt"
	"os/exec"
	"regexp"
	"strings"
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

var (
	GetCommitsFn = getCommits
	execCommand  = exec.Command
)

func getCommits(since string, until string) ([]string, error) {
	if until == "" {
		until = "HEAD"
	}

	if since == "" {
		lastTag, err := getLastTag()
		if err != nil {
			since = "HEAD~10"
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
	cmd := execCommand("git", "log", "--pretty=format:%s", revRange)

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

func getLastTag() (string, error) {
	cmd := execCommand("git", "describe", "--tags", "--abbrev=0")
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
