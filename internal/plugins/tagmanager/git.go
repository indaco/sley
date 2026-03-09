package tagmanager

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"
	"strings"

	"github.com/indaco/sley/internal/core"
)

// OSGitTagOperations implements core.GitTagOperations using actual git commands.
type OSGitTagOperations struct {
	execCommandContext func(ctx context.Context, name string, arg ...string) *exec.Cmd
}

// NewOSGitTagOperations creates a new OSGitTagOperations with the default exec.CommandContext.
func NewOSGitTagOperations() *OSGitTagOperations {
	return &OSGitTagOperations{
		execCommandContext: exec.CommandContext,
	}
}

// Verify OSGitTagOperations implements core.GitTagOperations.
var _ core.GitTagOperations = (*OSGitTagOperations)(nil)

func (g *OSGitTagOperations) CreateAnnotatedTag(ctx context.Context, name, message string) error {
	cmd := g.execCommandContext(ctx, "git", "tag", "-a", name, "-m", message)
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

func (g *OSGitTagOperations) CreateLightweightTag(ctx context.Context, name string) error {
	cmd := g.execCommandContext(ctx, "git", "tag", name)
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

func (g *OSGitTagOperations) CreateSignedTag(ctx context.Context, name, message, keyID string) error {
	var args []string
	if keyID != "" {
		args = []string{"tag", "-s", "-u", keyID, name, "-m", message}
	} else {
		args = []string{"tag", "-s", name, "-m", message}
	}

	cmd := g.execCommandContext(ctx, "git", args...)
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

func (g *OSGitTagOperations) TagExists(ctx context.Context, name string) (bool, error) {
	cmd := g.execCommandContext(ctx, "git", "tag", "-l", name)
	var stdout bytes.Buffer
	cmd.Stdout = &stdout

	if err := cmd.Run(); err != nil {
		return false, fmt.Errorf("failed to list tags: %w", err)
	}

	// If the tag exists, git tag -l will output the tag name
	output := strings.TrimSpace(stdout.String())
	return output == name, nil
}

func (g *OSGitTagOperations) GetLatestTag(ctx context.Context) (string, error) {
	cmd := g.execCommandContext(ctx, "git", "describe", "--tags", "--abbrev=0")
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

func (g *OSGitTagOperations) PushTag(ctx context.Context, name string) error {
	cmd := g.execCommandContext(ctx, "git", "push", "origin", name)
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

func (g *OSGitTagOperations) ListTags(ctx context.Context, pattern string) ([]string, error) {
	args := []string{"tag", "-l"}
	if pattern != "" {
		args = append(args, pattern)
	}

	cmd := g.execCommandContext(ctx, "git", args...)
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

func (g *OSGitTagOperations) DeleteTag(ctx context.Context, name string) error {
	cmd := g.execCommandContext(ctx, "git", "tag", "-d", name)
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

func (g *OSGitTagOperations) DeleteRemoteTag(ctx context.Context, name string) error {
	cmd := g.execCommandContext(ctx, "git", "push", "origin", "--delete", name)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		stderrMsg := strings.TrimSpace(stderr.String())
		if stderrMsg != "" {
			return fmt.Errorf("%s: %w", stderrMsg, err)
		}
		return fmt.Errorf("git push --delete tag failed: %w", err)
	}
	return nil
}

// OSGitCommitOperations implements core.GitCommitOperations using actual git commands.
type OSGitCommitOperations struct {
	execCommandContext func(ctx context.Context, name string, arg ...string) *exec.Cmd
}

// NewOSGitCommitOperations creates a new OSGitCommitOperations with the default exec.CommandContext.
func NewOSGitCommitOperations() *OSGitCommitOperations {
	return &OSGitCommitOperations{
		execCommandContext: exec.CommandContext,
	}
}

// Verify OSGitCommitOperations implements core.GitCommitOperations.
var _ core.GitCommitOperations = (*OSGitCommitOperations)(nil)

func (g *OSGitCommitOperations) StageFiles(ctx context.Context, files ...string) error {
	if len(files) == 0 {
		return nil
	}
	args := append([]string{"add", "--"}, files...)
	cmd := g.execCommandContext(ctx, "git", args...)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		stderrMsg := strings.TrimSpace(stderr.String())
		if stderrMsg != "" {
			return fmt.Errorf("%s: %w", stderrMsg, err)
		}
		return fmt.Errorf("git add failed: %w", err)
	}
	return nil
}

func (g *OSGitCommitOperations) Commit(ctx context.Context, message string) error {
	cmd := g.execCommandContext(ctx, "git", "commit", "-m", message)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		stderrMsg := strings.TrimSpace(stderr.String())
		if stderrMsg != "" {
			return fmt.Errorf("%s: %w", stderrMsg, err)
		}
		return fmt.Errorf("git commit failed: %w", err)
	}
	return nil
}

func (g *OSGitCommitOperations) GetModifiedFiles(ctx context.Context) ([]string, error) {
	cmd := g.execCommandContext(ctx, "git", "status", "--porcelain")
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		stderrMsg := strings.TrimSpace(stderr.String())
		if stderrMsg != "" {
			return nil, fmt.Errorf("%s: %w", stderrMsg, err)
		}
		return nil, fmt.Errorf("git status failed: %w", err)
	}

	output := strings.TrimSpace(stdout.String())
	if output == "" {
		return []string{}, nil
	}

	var files []string
	for line := range strings.SplitSeq(output, "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		// git status --porcelain output format: "XY filename"
		// The filename starts at position 3
		if len(line) > 3 {
			files = append(files, strings.TrimSpace(line[3:]))
		} else if len(line) > 2 {
			files = append(files, strings.TrimSpace(line[2:]))
		}
	}
	return files, nil
}

// defaultGitTagOps is the default git tag operations instance used by package-level functions.
var defaultGitTagOps = NewOSGitTagOperations()

// ListTags returns all git tags matching a pattern (package-level convenience function).
func ListTags(pattern string) ([]string, error) {
	return defaultGitTagOps.ListTags(context.Background(), pattern)
}

// DeleteTag deletes a local git tag (package-level convenience function).
func DeleteTag(name string) error {
	return defaultGitTagOps.DeleteTag(context.Background(), name)
}
