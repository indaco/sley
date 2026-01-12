package tagmanager

import (
	"bytes"
	"fmt"
	"os/exec"
	"strings"

	"github.com/indaco/sley/internal/core"
)

// OSGitTagOperations implements core.GitTagOperations using actual git commands.
type OSGitTagOperations struct {
	execCommand func(name string, arg ...string) *exec.Cmd
}

// NewOSGitTagOperations creates a new OSGitTagOperations with the default exec.Command.
func NewOSGitTagOperations() *OSGitTagOperations {
	return &OSGitTagOperations{
		execCommand: exec.Command,
	}
}

// Verify OSGitTagOperations implements core.GitTagOperations.
var _ core.GitTagOperations = (*OSGitTagOperations)(nil)

func (g *OSGitTagOperations) CreateAnnotatedTag(name, message string) error {
	cmd := g.execCommand("git", "tag", "-a", name, "-m", message)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		stderrMsg := strings.TrimSpace(stderr.String())
		if stderrMsg != "" {
			return fmt.Errorf("%s: %w", stderrMsg, err)
		}
		return fmt.Errorf("git tag (annotated) failed: %w", err)
	}
	return nil
}

func (g *OSGitTagOperations) CreateLightweightTag(name string) error {
	cmd := g.execCommand("git", "tag", name)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		stderrMsg := strings.TrimSpace(stderr.String())
		if stderrMsg != "" {
			return fmt.Errorf("%s: %w", stderrMsg, err)
		}
		return fmt.Errorf("git tag (lightweight) failed: %w", err)
	}
	return nil
}

func (g *OSGitTagOperations) CreateSignedTag(name, message, keyID string) error {
	var args []string
	if keyID != "" {
		args = []string{"tag", "-s", "-u", keyID, name, "-m", message}
	} else {
		args = []string{"tag", "-s", name, "-m", message}
	}

	cmd := g.execCommand("git", args...)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		stderrMsg := strings.TrimSpace(stderr.String())
		if stderrMsg != "" {
			return fmt.Errorf("%s: %w", stderrMsg, err)
		}
		return fmt.Errorf("git tag (signed) failed: %w", err)
	}
	return nil
}

func (g *OSGitTagOperations) TagExists(name string) (bool, error) {
	cmd := g.execCommand("git", "tag", "-l", name)
	var stdout bytes.Buffer
	cmd.Stdout = &stdout

	if err := cmd.Run(); err != nil {
		return false, fmt.Errorf("failed to list tags: %w", err)
	}

	// If the tag exists, git tag -l will output the tag name
	output := strings.TrimSpace(stdout.String())
	return output == name, nil
}

func (g *OSGitTagOperations) GetLatestTag() (string, error) {
	cmd := g.execCommand("git", "describe", "--tags", "--abbrev=0")
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		stderrMsg := strings.TrimSpace(stderr.String())
		if stderrMsg != "" {
			return "", fmt.Errorf("%s: %w", stderrMsg, err)
		}
		return "", fmt.Errorf("no tags found: %w", err)
	}

	tag := strings.TrimSpace(stdout.String())
	if tag == "" {
		return "", fmt.Errorf("no tags found")
	}

	return tag, nil
}

func (g *OSGitTagOperations) PushTag(name string) error {
	cmd := g.execCommand("git", "push", "origin", name)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		stderrMsg := strings.TrimSpace(stderr.String())
		if stderrMsg != "" {
			return fmt.Errorf("%s: %w", stderrMsg, err)
		}
		return fmt.Errorf("git push tag failed: %w", err)
	}
	return nil
}

// defaultGitTagOps is the default git tag operations for backward compatibility.
var defaultGitTagOps = NewOSGitTagOperations()

// Function variables for backward compatibility during migration.
// These delegate to the interface-based implementation.
var (
	createAnnotatedTagFn   = func(name, message string) error { return defaultGitTagOps.CreateAnnotatedTag(name, message) }
	createLightweightTagFn = func(name string) error { return defaultGitTagOps.CreateLightweightTag(name) }
	createSignedTagFn      = func(name, message, keyID string) error { return defaultGitTagOps.CreateSignedTag(name, message, keyID) }
	tagExistsFn            = func(name string) (bool, error) { return defaultGitTagOps.TagExists(name) }
	getLatestTagFn         = func() (string, error) { return defaultGitTagOps.GetLatestTag() }
	pushTagFn              = func(name string) error { return defaultGitTagOps.PushTag(name) }
	execCommand            = exec.Command
)

// ListTags returns all git tags matching a pattern.
func ListTags(pattern string) ([]string, error) {
	args := []string{"tag", "-l"}
	if pattern != "" {
		args = append(args, pattern)
	}

	cmd := execCommand("git", args...)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		stderrMsg := strings.TrimSpace(stderr.String())
		if stderrMsg != "" {
			return nil, fmt.Errorf("%s: %w", stderrMsg, err)
		}
		return nil, fmt.Errorf("git tag list failed: %w", err)
	}

	output := strings.TrimSpace(stdout.String())
	if output == "" {
		return []string{}, nil
	}

	return strings.Split(output, "\n"), nil
}

// DeleteTag deletes a local git tag.
func DeleteTag(name string) error {
	cmd := execCommand("git", "tag", "-d", name)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		stderrMsg := strings.TrimSpace(stderr.String())
		if stderrMsg != "" {
			return fmt.Errorf("%s: %w", stderrMsg, err)
		}
		return fmt.Errorf("git tag delete failed: %w", err)
	}
	return nil
}
