package changeloggenerator

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestNewChangelogGenerator(t *testing.T) {
	cfg := DefaultConfig()
	plugin, err := NewChangelogGenerator(cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if plugin == nil {
		t.Fatal("expected non-nil plugin")
	}
	if plugin.config != cfg {
		t.Error("expected config to match")
	}
	if plugin.generator == nil {
		t.Error("expected generator to be created")
	}
}

func TestNewChangelogGenerator_NilConfig(t *testing.T) {
	plugin, err := NewChangelogGenerator(nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if plugin == nil {
		t.Fatal("expected non-nil plugin")
	}
	if plugin.config == nil {
		t.Error("expected default config to be used")
	}
	if plugin.config.Mode != "versioned" {
		t.Errorf("Mode = %q, want 'versioned' (default)", plugin.config.Mode)
	}
}

func TestPluginName(t *testing.T) {
	plugin, err := NewChangelogGenerator(DefaultConfig())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	name := plugin.Name()
	if name != "changelog-generator" {
		t.Errorf("Name() = %q, want 'changelog-generator'", name)
	}
}

func TestPluginDescription(t *testing.T) {
	plugin, err := NewChangelogGenerator(DefaultConfig())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	desc := plugin.Description()
	if desc == "" {
		t.Error("expected non-empty description")
	}
	if !strings.Contains(desc, "changelog") {
		t.Errorf("Description() = %q, expected to contain 'changelog'", desc)
	}
}

func TestPluginVersion(t *testing.T) {
	plugin, err := NewChangelogGenerator(DefaultConfig())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	version := plugin.Version()
	if version == "" {
		t.Error("expected non-empty version")
	}
	if !strings.HasPrefix(version, "v") {
		t.Errorf("Version() = %q, expected to start with 'v'", version)
	}
}

func TestPluginIsEnabled(t *testing.T) {
	tests := []struct {
		name     string
		enabled  bool
		expected bool
	}{
		{"Enabled", true, true},
		{"Disabled", false, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := DefaultConfig()
			cfg.Enabled = tt.enabled
			plugin, err := NewChangelogGenerator(cfg)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if got := plugin.IsEnabled(); got != tt.expected {
				t.Errorf("IsEnabled() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestPluginGetConfig(t *testing.T) {
	cfg := DefaultConfig()
	cfg.Mode = "unified"
	plugin, err := NewChangelogGenerator(cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	got := plugin.GetConfig()
	if got != cfg {
		t.Error("expected GetConfig() to return the same config")
	}
	if got.Mode != "unified" {
		t.Errorf("config.Mode = %q, want 'unified'", got.Mode)
	}
}

func TestGenerateForVersion_Disabled(t *testing.T) {
	cfg := DefaultConfig()
	cfg.Enabled = false
	plugin, err := NewChangelogGenerator(cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Should return nil without doing anything
	err = plugin.GenerateForVersion("v1.0.0", "v0.9.0", "patch")
	if err != nil {
		t.Errorf("expected nil error for disabled plugin, got %v", err)
	}
}

func TestGenerateForVersion_Enabled_VersionedMode(t *testing.T) {
	tmpDir := t.TempDir()

	cfg := DefaultConfig()
	cfg.Enabled = true
	cfg.Mode = "versioned"
	cfg.ChangesDir = filepath.Join(tmpDir, ".changes")
	cfg.Repository = &RepositoryConfig{
		Provider: "github",
		Host:     "github.com",
		Owner:    "testowner",
		Repo:     "testrepo",
	}
	plugin, err := NewChangelogGenerator(cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Mock GetCommitsWithMetaFn to return test commits
	originalFn := GetCommitsWithMetaFn
	GetCommitsWithMetaFn = func(since, until string) ([]CommitInfo, error) {
		return []CommitInfo{
			{Hash: "abc123", ShortHash: "abc123", Subject: "feat: test feature", Author: "Test", AuthorEmail: "test@example.com"},
		}, nil
	}
	defer func() { GetCommitsWithMetaFn = originalFn }()

	err = plugin.GenerateForVersion("v1.0.0", "v0.9.0", "minor")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Check file was created
	expectedPath := filepath.Join(cfg.ChangesDir, "v1.0.0.md")
	if _, err := os.Stat(expectedPath); os.IsNotExist(err) {
		t.Errorf("expected file at %s", expectedPath)
	}

	// Check content
	data, err := os.ReadFile(expectedPath)
	if err != nil {
		t.Fatalf("failed to read file: %v", err)
	}
	content := string(data)

	if !strings.Contains(content, "v1.0.0") {
		t.Error("expected version in content")
	}
	if !strings.Contains(content, "test feature") {
		t.Error("expected commit description in content")
	}
}

func TestGenerateForVersion_Enabled_UnifiedMode(t *testing.T) {
	tmpDir := t.TempDir()

	cfg := DefaultConfig()
	cfg.Enabled = true
	cfg.Mode = "unified"
	cfg.ChangelogPath = filepath.Join(tmpDir, "CHANGELOG.md")
	cfg.Repository = &RepositoryConfig{
		Provider: "github",
		Host:     "github.com",
		Owner:    "testowner",
		Repo:     "testrepo",
	}
	plugin, err := NewChangelogGenerator(cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Mock GetCommitsWithMetaFn
	originalFn := GetCommitsWithMetaFn
	GetCommitsWithMetaFn = func(since, until string) ([]CommitInfo, error) {
		return []CommitInfo{
			{Hash: "def456", ShortHash: "def456", Subject: "fix: test fix", Author: "Test", AuthorEmail: "test@example.com"},
		}, nil
	}
	defer func() { GetCommitsWithMetaFn = originalFn }()

	err = plugin.GenerateForVersion("v1.0.0", "v0.9.0", "patch")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Check file was created
	if _, err := os.Stat(cfg.ChangelogPath); os.IsNotExist(err) {
		t.Errorf("expected CHANGELOG.md at %s", cfg.ChangelogPath)
	}

	// Check content
	data, err := os.ReadFile(cfg.ChangelogPath)
	if err != nil {
		t.Fatalf("failed to read file: %v", err)
	}
	content := string(data)

	if !strings.Contains(content, "v1.0.0") {
		t.Error("expected version in content")
	}
}

func TestGenerateForVersion_Enabled_BothMode(t *testing.T) {
	tmpDir := t.TempDir()

	cfg := DefaultConfig()
	cfg.Enabled = true
	cfg.Mode = "both"
	cfg.ChangesDir = filepath.Join(tmpDir, ".changes")
	cfg.ChangelogPath = filepath.Join(tmpDir, "CHANGELOG.md")
	cfg.Repository = &RepositoryConfig{
		Provider: "github",
		Host:     "github.com",
		Owner:    "testowner",
		Repo:     "testrepo",
	}
	plugin, err := NewChangelogGenerator(cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Mock GetCommitsWithMetaFn
	originalFn := GetCommitsWithMetaFn
	GetCommitsWithMetaFn = func(since, until string) ([]CommitInfo, error) {
		return []CommitInfo{
			{Hash: "ghi789", ShortHash: "ghi789", Subject: "docs: update docs", Author: "Test", AuthorEmail: "test@example.com"},
		}, nil
	}
	defer func() { GetCommitsWithMetaFn = originalFn }()

	err = plugin.GenerateForVersion("v1.0.0", "v0.9.0", "patch")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Check both files were created
	versionedPath := filepath.Join(cfg.ChangesDir, "v1.0.0.md")
	if _, err := os.Stat(versionedPath); os.IsNotExist(err) {
		t.Errorf("expected versioned file at %s", versionedPath)
	}

	if _, err := os.Stat(cfg.ChangelogPath); os.IsNotExist(err) {
		t.Errorf("expected CHANGELOG.md at %s", cfg.ChangelogPath)
	}
}

func TestGenerateForVersion_NoCommits(t *testing.T) {
	cfg := DefaultConfig()
	cfg.Enabled = true
	plugin, err := NewChangelogGenerator(cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Mock GetCommitsWithMetaFn to return empty
	originalFn := GetCommitsWithMetaFn
	GetCommitsWithMetaFn = func(since, until string) ([]CommitInfo, error) {
		return []CommitInfo{}, nil
	}
	defer func() { GetCommitsWithMetaFn = originalFn }()

	err = plugin.GenerateForVersion("v1.0.0", "v0.9.0", "patch")
	if err != nil {
		t.Errorf("expected nil error for no commits, got %v", err)
	}
}

func TestGenerateForVersion_UnknownMode(t *testing.T) {
	tmpDir := t.TempDir()

	cfg := DefaultConfig()
	cfg.Enabled = true
	cfg.Mode = "invalid"
	cfg.ChangesDir = filepath.Join(tmpDir, ".changes")
	cfg.Repository = &RepositoryConfig{
		Provider: "github",
		Host:     "github.com",
		Owner:    "testowner",
		Repo:     "testrepo",
	}
	plugin, err := NewChangelogGenerator(cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Mock GetCommitsWithMetaFn
	originalFn := GetCommitsWithMetaFn
	GetCommitsWithMetaFn = func(since, until string) ([]CommitInfo, error) {
		return []CommitInfo{
			{Hash: "abc123", ShortHash: "abc123", Subject: "feat: test", Author: "Test", AuthorEmail: "test@example.com"},
		}, nil
	}
	defer func() { GetCommitsWithMetaFn = originalFn }()

	err = plugin.GenerateForVersion("v1.0.0", "v0.9.0", "patch")
	if err == nil {
		t.Error("expected error for unknown mode")
	}
	if !strings.Contains(err.Error(), "unknown mode") {
		t.Errorf("expected 'unknown mode' in error, got %v", err)
	}
}

func TestChangelogGeneratorInterface(t *testing.T) {
	// Verify that ChangelogGeneratorPlugin implements ChangelogGenerator
	var _ ChangelogGenerator = (*ChangelogGeneratorPlugin)(nil)
}

func TestWriteChangelog_Versioned(t *testing.T) {
	tmpDir := t.TempDir()

	cfg := DefaultConfig()
	cfg.Mode = "versioned"
	cfg.ChangesDir = filepath.Join(tmpDir, ".changes")
	plugin, err := NewChangelogGenerator(cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	content := "## v1.0.0\n\nTest content"
	err = plugin.writeChangelog("v1.0.0", content)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expectedPath := filepath.Join(cfg.ChangesDir, "v1.0.0.md")
	if _, err := os.Stat(expectedPath); os.IsNotExist(err) {
		t.Errorf("expected file at %s", expectedPath)
	}
}

func TestWriteChangelog_Unified(t *testing.T) {
	tmpDir := t.TempDir()

	cfg := DefaultConfig()
	cfg.Mode = "unified"
	cfg.ChangelogPath = filepath.Join(tmpDir, "CHANGELOG.md")
	plugin, err := NewChangelogGenerator(cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	content := "## v1.0.0\n\nTest content"
	err = plugin.writeChangelog("v1.0.0", content)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if _, err := os.Stat(cfg.ChangelogPath); os.IsNotExist(err) {
		t.Errorf("expected CHANGELOG.md at %s", cfg.ChangelogPath)
	}
}

func TestWriteChangelog_Both(t *testing.T) {
	tmpDir := t.TempDir()

	cfg := DefaultConfig()
	cfg.Mode = "both"
	cfg.ChangesDir = filepath.Join(tmpDir, ".changes")
	cfg.ChangelogPath = filepath.Join(tmpDir, "CHANGELOG.md")
	plugin, err := NewChangelogGenerator(cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	content := "## v1.0.0\n\nTest content"
	err = plugin.writeChangelog("v1.0.0", content)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Check both files
	versionedPath := filepath.Join(cfg.ChangesDir, "v1.0.0.md")
	if _, err := os.Stat(versionedPath); os.IsNotExist(err) {
		t.Errorf("expected versioned file at %s", versionedPath)
	}
	if _, err := os.Stat(cfg.ChangelogPath); os.IsNotExist(err) {
		t.Errorf("expected CHANGELOG.md at %s", cfg.ChangelogPath)
	}
}
