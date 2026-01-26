package initialize

import (
	"strings"
	"testing"

	"github.com/goccy/go-yaml"
	"github.com/indaco/sley/internal/config"
	"github.com/indaco/sley/internal/discovery"
	"github.com/indaco/sley/internal/parser"
)

func TestGenerateConfigWithComments_Empty(t *testing.T) {
	data, err := GenerateConfigWithComments(".version", []string{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(data) == 0 {
		t.Error("expected non-empty config data")
	}

	// Should contain header comments
	dataStr := string(data)
	if !strings.Contains(dataStr, "sley configuration file") {
		t.Error("expected header comment")
	}
}

func TestGenerateConfigWithComments_CommitParser(t *testing.T) {
	data, err := GenerateConfigWithComments(".version", []string{"commit-parser"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	dataStr := string(data)

	// Should contain plugin in enabled list
	if !strings.Contains(dataStr, "commit-parser") {
		t.Error("expected commit-parser in config")
	}

	// Parse and verify structure
	var cfg config.Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		t.Fatalf("failed to parse generated config: %v", err)
	}

	if cfg.Plugins == nil {
		t.Fatal("expected plugins config")
	}

	if !cfg.Plugins.CommitParser {
		t.Error("expected commit-parser to be enabled")
	}
}

func TestGenerateConfigWithComments_TagManager(t *testing.T) {
	data, err := GenerateConfigWithComments(".version", []string{"tag-manager"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var cfg config.Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		t.Fatalf("failed to parse generated config: %v", err)
	}

	if cfg.Plugins == nil || cfg.Plugins.TagManager == nil {
		t.Fatal("expected tag-manager config")
	}

	if !cfg.Plugins.TagManager.Enabled {
		t.Error("expected tag-manager to be enabled")
	}
}

func TestGenerateConfigWithComments_MultiplePlugins(t *testing.T) {
	plugins := []string{
		"commit-parser",
		"tag-manager",
		"version-validator",
		"dependency-check",
	}

	data, err := GenerateConfigWithComments(".version", plugins)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var cfg config.Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		t.Fatalf("failed to parse generated config: %v", err)
	}

	if cfg.Plugins == nil {
		t.Fatal("expected plugins config")
	}

	// Verify all plugins are enabled
	if !cfg.Plugins.CommitParser {
		t.Error("expected commit-parser to be enabled")
	}
	if cfg.Plugins.TagManager == nil || !cfg.Plugins.TagManager.Enabled {
		t.Error("expected tag-manager to be enabled")
	}
	if cfg.Plugins.VersionValidator == nil || !cfg.Plugins.VersionValidator.Enabled {
		t.Error("expected version-validator to be enabled")
	}
	if cfg.Plugins.DependencyCheck == nil || !cfg.Plugins.DependencyCheck.Enabled {
		t.Error("expected dependency-check to be enabled")
	}
}

func TestGenerateConfigWithComments_AllPlugins(t *testing.T) {
	plugins := []string{
		"commit-parser",
		"tag-manager",
		"version-validator",
		"dependency-check",
		"changelog-parser",
		"changelog-generator",
		"release-gate",
		"audit-log",
	}

	data, err := GenerateConfigWithComments(".version", plugins)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var cfg config.Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		t.Fatalf("failed to parse generated config: %v", err)
	}

	if cfg.Plugins == nil {
		t.Fatal("expected plugins config")
	}

	// Verify all plugins are enabled using helper
	verifyAllPluginsEnabled(t, cfg.Plugins)
}

// verifyAllPluginsEnabled checks that all plugins are properly enabled.
func verifyAllPluginsEnabled(t *testing.T, plugins *config.PluginConfig) {
	t.Helper()

	checks := []struct {
		name    string
		enabled bool
	}{
		{"commit-parser", plugins.CommitParser},
		{"tag-manager", plugins.TagManager != nil && plugins.TagManager.Enabled},
		{"version-validator", plugins.VersionValidator != nil && plugins.VersionValidator.Enabled},
		{"dependency-check", plugins.DependencyCheck != nil && plugins.DependencyCheck.Enabled},
		{"changelog-parser", plugins.ChangelogParser != nil && plugins.ChangelogParser.Enabled},
		{"changelog-generator", plugins.ChangelogGenerator != nil && plugins.ChangelogGenerator.Enabled},
		{"release-gate", plugins.ReleaseGate != nil && plugins.ReleaseGate.Enabled},
		{"audit-log", plugins.AuditLog != nil && plugins.AuditLog.Enabled},
	}

	for _, check := range checks {
		if !check.enabled {
			t.Errorf("expected %s to be enabled", check.name)
		}
	}
}

func TestGenerateConfigWithComments_PathField(t *testing.T) {
	tests := []struct {
		name     string
		path     string
		wantPath string
	}{
		{"default path", ".version", ".version"},
		{"custom path", "internal/version/.version", "internal/version/.version"},
		{"custom directory", "custom/path/.version", "custom/path/.version"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data, err := GenerateConfigWithComments(tt.path, []string{"commit-parser"})
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			var cfg config.Config
			if err := yaml.Unmarshal(data, &cfg); err != nil {
				t.Fatalf("failed to parse generated config: %v", err)
			}

			if cfg.Path != tt.wantPath {
				t.Errorf("expected path to be %q, got %q", tt.wantPath, cfg.Path)
			}
		})
	}
}

func TestGenerateConfigWithComments_HeaderComments(t *testing.T) {
	data, err := GenerateConfigWithComments(".version", []string{"commit-parser"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	dataStr := string(data)

	// Verify header comments are present
	expectedComments := []string{
		"sley configuration file",
		"Documentation:",
		"Enabled plugins:",
	}

	for _, comment := range expectedComments {
		if !strings.Contains(dataStr, comment) {
			t.Errorf("expected comment %q in output", comment)
		}
	}
}

func TestGenerateConfigWithComments_InlineComments(t *testing.T) {
	plugins := []string{"commit-parser", "tag-manager"}
	data, err := GenerateConfigWithComments(".version", plugins)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	dataStr := string(data)

	// Verify inline comments are added for plugins
	// The exact format may vary, but we should see plugin descriptions as comments
	if !strings.Contains(dataStr, "commit-parser") {
		t.Error("expected commit-parser in output")
	}
	if !strings.Contains(dataStr, "tag-manager") {
		t.Error("expected tag-manager in output")
	}
}

func TestGenerateConfigWithDiscovery(t *testing.T) {
	plugins := []string{"dependency-check", "tag-manager"}
	syncCandidates := []discovery.SyncCandidate{
		{
			Path:        "package.json",
			Format:      parser.FormatJSON,
			Field:       "version",
			Description: "Node.js",
		},
		{
			Path:        "Cargo.toml",
			Format:      parser.FormatTOML,
			Field:       "package.version",
			Description: "Rust",
		},
	}

	data, err := GenerateConfigWithDiscovery(".version", plugins, syncCandidates)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var cfg config.Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		t.Fatalf("failed to parse generated config: %v", err)
	}

	// Verify dependency-check is configured
	if cfg.Plugins == nil || cfg.Plugins.DependencyCheck == nil {
		t.Fatal("expected dependency-check config")
	}

	depCheck := cfg.Plugins.DependencyCheck
	if !depCheck.Enabled {
		t.Error("expected dependency-check to be enabled")
	}
	if !depCheck.AutoSync {
		t.Error("expected auto-sync to be enabled")
	}
	if len(depCheck.Files) != 2 {
		t.Errorf("expected 2 files, got %d", len(depCheck.Files))
	}

	// Verify file configurations
	if depCheck.Files[0].Path != "package.json" {
		t.Errorf("expected first file to be package.json, got %s", depCheck.Files[0].Path)
	}
	if depCheck.Files[0].Format != "json" {
		t.Errorf("expected first file format to be json, got %s", depCheck.Files[0].Format)
	}
	if depCheck.Files[1].Path != "Cargo.toml" {
		t.Errorf("expected second file to be Cargo.toml, got %s", depCheck.Files[1].Path)
	}
}

func TestGenerateConfigWithDiscovery_EmptyCandidates(t *testing.T) {
	plugins := []string{"dependency-check"}
	syncCandidates := []discovery.SyncCandidate{}

	data, err := GenerateConfigWithDiscovery(".version", plugins, syncCandidates)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var cfg config.Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		t.Fatalf("failed to parse generated config: %v", err)
	}

	// Should still have dependency-check enabled, but no files
	if cfg.Plugins == nil || cfg.Plugins.DependencyCheck == nil {
		t.Fatal("expected dependency-check config")
	}

	if len(cfg.Plugins.DependencyCheck.Files) != 0 {
		t.Errorf("expected 0 files for empty candidates, got %d", len(cfg.Plugins.DependencyCheck.Files))
	}
}

func TestGenerateConfigWithDiscovery_WithRegexPattern(t *testing.T) {
	plugins := []string{"dependency-check"}
	syncCandidates := []discovery.SyncCandidate{
		{
			Path:        "version.go",
			Format:      parser.FormatRegex,
			Pattern:     `Version = "(.*?)"`,
			Description: "Go version constant",
		},
	}

	data, err := GenerateConfigWithDiscovery(".version", plugins, syncCandidates)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var cfg config.Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		t.Fatalf("failed to parse generated config: %v", err)
	}

	if len(cfg.Plugins.DependencyCheck.Files) != 1 {
		t.Fatalf("expected 1 file, got %d", len(cfg.Plugins.DependencyCheck.Files))
	}

	file := cfg.Plugins.DependencyCheck.Files[0]
	if file.Format != "regex" {
		t.Errorf("expected format to be regex, got %s", file.Format)
	}
	if file.Pattern != `Version = "(.*?)"` {
		t.Errorf("expected pattern to be preserved, got %s", file.Pattern)
	}
}
