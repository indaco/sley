package semver

import (
	"context"
	"fmt"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/indaco/sley/internal/core"
)

// VersionManager handles version file operations with injected dependencies.
// This enables proper testing without global state mutation.
type VersionManager struct {
	fs  core.FileSystem
	git GitTagReader
}

// GitTagReader abstracts git tag reading for testability.
type GitTagReader interface {
	// DescribeTags returns the most recent tag, or an error if none exists.
	DescribeTags(ctx context.Context) (string, error)
}

// NewVersionManager creates a VersionManager with the given dependencies.
func NewVersionManager(fs core.FileSystem, git GitTagReader) *VersionManager {
	return &VersionManager{fs: fs, git: git}
}

// DefaultVersionManager returns a VersionManager using real OS and git.
func DefaultVersionManager() *VersionManager {
	return NewVersionManager(core.NewOSFileSystem(), &realGitClient{})
}

// Read reads a version from the given path.
func (m *VersionManager) Read(ctx context.Context, path string) (SemVersion, error) {
	data, err := m.fs.ReadFile(ctx, path)
	if err != nil {
		return SemVersion{}, err
	}
	return ParseVersion(string(data))
}

// Save writes a version to the given path.
func (m *VersionManager) Save(ctx context.Context, path string, version SemVersion) error {
	// Ensure parent directory exists
	if err := m.fs.MkdirAll(ctx, filepath.Dir(path), 0755); err != nil {
		return err
	}
	return m.fs.WriteFile(ctx, path, []byte(version.String()+"\n"), VersionFilePerm)
}

// Initialize creates a version file if it doesn't exist.
// It tries to use the latest git tag, or falls back to 0.1.0.
func (m *VersionManager) Initialize(ctx context.Context, path string) error {
	if _, err := m.fs.Stat(ctx, path); err == nil {
		return nil // Already exists
	}

	version := SemVersion{Major: 0, Minor: 1, Patch: 0} // Default

	if m.git != nil {
		tag, err := m.git.DescribeTags(ctx)
		if err == nil {
			tag = strings.TrimSpace(tag)
			tag = strings.TrimPrefix(tag, "v")
			if parsed, parseErr := ParseVersion(tag); parseErr == nil {
				version = parsed
			}
		}
	}

	return m.Save(ctx, path, version)
}

// InitializeWithFeedback initializes the version file and returns whether it was created.
func (m *VersionManager) InitializeWithFeedback(ctx context.Context, path string) (created bool, err error) {
	if _, err := m.fs.Stat(ctx, path); err == nil {
		return false, nil
	}

	if err := m.Initialize(ctx, path); err != nil {
		return false, err
	}

	return true, nil
}

// Update reads, bumps, and saves the version.
func (m *VersionManager) Update(ctx context.Context, path string, bumpType string, pre string, meta string, preserve bool) error {
	version, err := m.Read(ctx, path)
	if err != nil {
		return err
	}

	switch bumpType {
	case "patch":
		version.Patch++
	case "minor":
		version.Minor++
		version.Patch = 0
	case "major":
		version.Major++
		version.Minor = 0
		version.Patch = 0
	default:
		return fmt.Errorf("invalid bump type: %s", bumpType)
	}

	version.PreRelease = pre

	if meta != "" {
		version.Build = meta
	} else if !preserve {
		version.Build = ""
	}

	return m.Save(ctx, path, version)
}

// UpdatePreRelease updates only the pre-release portion of the version.
// If label is provided, it switches to that label (starting at .1).
// If label is empty, it increments the existing pre-release number.
// Returns an error if no label is provided and the version has no pre-release.
func (m *VersionManager) UpdatePreRelease(ctx context.Context, path string, label string, meta string, preserve bool) error {
	version, err := m.Read(ctx, path)
	if err != nil {
		return err
	}

	if label != "" {
		// Switch to new label or increment if same base label
		version.PreRelease = IncrementPreRelease(version.PreRelease, label)
	} else {
		// Increment existing pre-release
		if version.PreRelease == "" {
			return fmt.Errorf("current version has no pre-release; use --label to specify one")
		}
		// Extract base label and increment
		base := extractPreReleaseBase(version.PreRelease)
		version.PreRelease = IncrementPreRelease(version.PreRelease, base)
	}

	if meta != "" {
		version.Build = meta
	} else if !preserve {
		version.Build = ""
	}

	return m.Save(ctx, path, version)
}

// extractPreReleaseBase extracts the base label from a pre-release string.
// e.g., "rc.1" -> "rc", "beta.2" -> "beta", "alpha" -> "alpha", "rc1" -> "rc"
func extractPreReleaseBase(pre string) string {
	// First, check for dot followed by a number
	for i := len(pre) - 1; i >= 0; i-- {
		if pre[i] == '.' {
			// Check if everything after the dot is numeric
			suffix := pre[i+1:]
			isNumeric := true
			for _, c := range suffix {
				if c < '0' || c > '9' {
					isNumeric = false
					break
				}
			}
			if isNumeric && len(suffix) > 0 {
				return pre[:i]
			}
		}
	}

	// Check for trailing digits without dot (e.g., "rc1" -> "rc")
	lastNonDigit := -1
	for i := len(pre) - 1; i >= 0; i-- {
		if pre[i] < '0' || pre[i] > '9' {
			lastNonDigit = i
			break
		}
	}
	if lastNonDigit >= 0 && lastNonDigit < len(pre)-1 {
		return pre[:lastNonDigit+1]
	}

	return pre
}

// realGitClient implements GitTagReader using actual git commands.
type realGitClient struct{}

func (g *realGitClient) DescribeTags(ctx context.Context) (string, error) {
	cmd := exec.CommandContext(ctx, "git", "describe", "--tags", "--abbrev=0")
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return string(output), nil
}

// MockGitTagReader is a test helper for mocking git tag reading.
type MockGitTagReader struct {
	Tag string
	Err error
}

func (m *MockGitTagReader) DescribeTags(ctx context.Context) (string, error) {
	return m.Tag, m.Err
}

// Ensure interfaces are satisfied.
var (
	_ core.FileSystem = (*core.OSFileSystem)(nil)
	_ GitTagReader    = (*realGitClient)(nil)
	_ GitTagReader    = (*MockGitTagReader)(nil)
)

// defaultManager is the singleton used by legacy functions.
var defaultManager = DefaultVersionManager()

// SetDefaultManager allows tests to inject a custom manager.
// Returns a function to restore the original manager.
func SetDefaultManager(m *VersionManager) func() {
	old := defaultManager
	defaultManager = m
	return func() { defaultManager = old }
}
