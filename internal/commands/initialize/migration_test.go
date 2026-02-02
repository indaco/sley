package initialize

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/indaco/sley/internal/config"
	"github.com/indaco/sley/internal/testutils"
	"github.com/urfave/cli/v3"
)

func TestDetectPackageJSONVersion(t *testing.T) {
	tmp := t.TempDir()
	pkgPath := filepath.Join(tmp, "package.json")

	tests := []struct {
		name     string
		content  string
		expected string
		hasError bool
	}{
		{
			name:     "valid version",
			content:  `{"name": "test", "version": "1.2.3"}`,
			expected: "1.2.3",
		},
		{
			name:     "version with pre-release",
			content:  `{"name": "test", "version": "1.0.0-beta.1"}`,
			expected: "1.0.0-beta.1",
		},
		{
			name:     "no version field",
			content:  `{"name": "test"}`,
			expected: "",
		},
		{
			name:     "invalid json",
			content:  `{invalid}`,
			hasError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := os.WriteFile(pkgPath, []byte(tt.content), 0600); err != nil {
				t.Fatal(err)
			}

			version, err := detectPackageJSONVersion(pkgPath)
			if tt.hasError {
				if err == nil {
					t.Error("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if version != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, version)
			}
		})
	}
}

func TestDetectCargoVersion(t *testing.T) {
	tmp := t.TempDir()
	cargoPath := filepath.Join(tmp, "Cargo.toml")

	tests := []struct {
		name     string
		content  string
		expected string
	}{
		{
			name: "valid version",
			content: `[package]
name = "myapp"
version = "0.1.0"
edition = "2021"`,
			expected: "0.1.0",
		},
		{
			name: "version with pre-release",
			content: `[package]
name = "myapp"
version = "1.0.0-alpha.1"`,
			expected: "1.0.0-alpha.1",
		},
		{
			name: "no version",
			content: `[package]
name = "myapp"`,
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := os.WriteFile(cargoPath, []byte(tt.content), 0600); err != nil {
				t.Fatal(err)
			}

			version, err := detectCargoVersion(cargoPath)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if version != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, version)
			}
		})
	}
}

func TestDetectPyprojectVersion(t *testing.T) {
	tmp := t.TempDir()
	pyprojectPath := filepath.Join(tmp, "pyproject.toml")

	tests := []struct {
		name     string
		content  string
		expected string
	}{
		{
			name: "PEP 621 format",
			content: `[project]
name = "myproject"
version = "2.0.0"`,
			expected: "2.0.0",
		},
		{
			name: "poetry format",
			content: `[tool.poetry]
name = "myproject"
version = "1.5.0"`,
			expected: "1.5.0",
		},
		{
			name: "single quotes",
			content: `[project]
name = 'myproject'
version = '3.0.0'`,
			expected: "3.0.0",
		},
		{
			name: "no version",
			content: `[project]
name = "myproject"`,
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := os.WriteFile(pyprojectPath, []byte(tt.content), 0600); err != nil {
				t.Fatal(err)
			}

			version, err := detectPyprojectVersion(pyprojectPath)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if version != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, version)
			}
		})
	}
}

func TestDetectChartVersion(t *testing.T) {
	tmp := t.TempDir()
	chartPath := filepath.Join(tmp, "Chart.yaml")

	tests := []struct {
		name     string
		content  string
		expected string
	}{
		{
			name: "valid chart",
			content: `apiVersion: v2
name: mychart
version: 1.0.0
appVersion: "1.16.0"`,
			expected: "1.0.0",
		},
		{
			name: "no version",
			content: `apiVersion: v2
name: mychart`,
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := os.WriteFile(chartPath, []byte(tt.content), 0600); err != nil {
				t.Fatal(err)
			}

			version, err := detectChartVersion(chartPath)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if version != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, version)
			}
		})
	}
}

func TestDetectPlainTextVersion(t *testing.T) {
	tmp := t.TempDir()
	versionPath := filepath.Join(tmp, "VERSION")

	tests := []struct {
		name     string
		content  string
		expected string
	}{
		{"simple version", "1.2.3\n", "1.2.3"},
		{"with v prefix", "v2.0.0\n", "2.0.0"},
		{"with whitespace", "  3.0.0  \n", "3.0.0"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := os.WriteFile(versionPath, []byte(tt.content), 0600); err != nil {
				t.Fatal(err)
			}

			version, err := detectPlainTextVersion(versionPath)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if version != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, version)
			}
		})
	}
}

func TestIsValidSemver(t *testing.T) {
	tests := []struct {
		version string
		valid   bool
	}{
		{"1.0.0", true},
		{"0.1.0", true},
		{"1.2.3", true},
		{"1.0.0-alpha", true},
		{"1.0.0-beta.1", true},
		{"1.0.0-rc.1+build.123", true},
		{"v1.0.0", true},
		{"1.0", false},
		{"1", false},
		{"invalid", false},
		{"1.0.0.0", false},
	}

	for _, tt := range tests {
		t.Run(tt.version, func(t *testing.T) {
			got := isValidSemver(tt.version)
			if got != tt.valid {
				t.Errorf("isValidSemver(%q): expected %v, got %v", tt.version, tt.valid, got)
			}
		})
	}
}

func TestGetBestVersionSource(t *testing.T) {
	tests := []struct {
		name     string
		sources  []VersionSource
		expected string
	}{
		{
			name:     "empty sources",
			sources:  []VersionSource{},
			expected: "",
		},
		{
			name: "single source",
			sources: []VersionSource{
				{File: "Cargo.toml", Version: "1.0.0"},
			},
			expected: "Cargo.toml",
		},
		{
			name: "package.json takes priority",
			sources: []VersionSource{
				{File: "Cargo.toml", Version: "1.0.0"},
				{File: "package.json", Version: "2.0.0"},
			},
			expected: "package.json",
		},
		{
			name: "priority order",
			sources: []VersionSource{
				{File: "VERSION", Version: "3.0.0"},
				{File: "pyproject.toml", Version: "2.0.0"},
				{File: "Cargo.toml", Version: "1.0.0"},
			},
			expected: "Cargo.toml",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			best := GetBestVersionSource(tt.sources)
			if tt.expected == "" {
				if best != nil {
					t.Errorf("expected nil, got %v", best)
				}
				return
			}
			if best == nil {
				t.Fatal("expected non-nil result")
			}
			if best.File != tt.expected {
				t.Errorf("expected file %q, got %q", tt.expected, best.File)
			}
		})
	}
}

func TestDetectExistingVersions(t *testing.T) {
	tmp := t.TempDir()

	// Create a package.json
	pkgContent := `{"name": "test", "version": "1.2.3"}`
	if err := os.WriteFile(filepath.Join(tmp, "package.json"), []byte(pkgContent), 0600); err != nil {
		t.Fatal(err)
	}

	// Create a Cargo.toml
	cargoContent := `[package]
name = "test"
version = "2.0.0"`
	if err := os.WriteFile(filepath.Join(tmp, "Cargo.toml"), []byte(cargoContent), 0600); err != nil {
		t.Fatal(err)
	}

	t.Chdir(tmp)

	sources := DetectExistingVersions()

	if len(sources) != 2 {
		t.Errorf("expected 2 sources, got %d", len(sources))
	}

	// Verify best source is package.json
	best := GetBestVersionSource(sources)
	if best == nil {
		t.Fatal("expected best source")
	}
	if best.File != "package.json" {
		t.Errorf("expected package.json, got %s", best.File)
	}
	if best.Version != "1.2.3" {
		t.Errorf("expected version 1.2.3, got %s", best.Version)
	}
}

func TestCLI_InitCommand_WithMigrateFlag(t *testing.T) {
	tmp := t.TempDir()
	versionPath := filepath.Join(tmp, ".version")

	// Create a package.json with version
	pkgContent := `{"name": "test-app", "version": "3.5.0"}`
	if err := os.WriteFile(filepath.Join(tmp, "package.json"), []byte(pkgContent), 0600); err != nil {
		t.Fatal(err)
	}

	t.Chdir(tmp)

	cfg := &config.Config{Path: versionPath}
	appCli := testutils.BuildCLIForTests(cfg.Path, []*cli.Command{Run()})

	err := appCli.Run(context.Background(), []string{
		"sley", "init", "--migrate", "--yes",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify .version was created with migrated version
	data, err := os.ReadFile(versionPath)
	if err != nil {
		t.Fatalf("failed to read .version: %v", err)
	}

	version := string(data)
	if version != "3.5.0\n" {
		t.Errorf("expected version '3.5.0\\n', got %q", version)
	}
}

func TestCLI_InitCommand_MigrateNoSources(t *testing.T) {
	tmp := t.TempDir()
	versionPath := filepath.Join(tmp, ".version")

	t.Chdir(tmp)

	cfg := &config.Config{Path: versionPath}
	appCli := testutils.BuildCLIForTests(cfg.Path, []*cli.Command{Run()})

	err := appCli.Run(context.Background(), []string{
		"sley", "init", "--migrate", "--yes",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify .version was created with default version (no sources to migrate)
	data, err := os.ReadFile(versionPath)
	if err != nil {
		t.Fatalf("failed to read .version: %v", err)
	}

	version := string(data)
	// Should fall back to default (0.0.0)
	if version != "0.0.0\n" {
		t.Errorf("expected version '0.0.0\\n', got %q", version)
	}
}

func TestFormatVersionSources(t *testing.T) {
	sources := []VersionSource{
		{File: "package.json", Version: "1.0.0", Format: "Node.js (package.json)"},
		{File: "Cargo.toml", Version: "2.0.0", Format: "Rust (Cargo.toml)"},
	}

	output := FormatVersionSources(sources)

	if output == "" {
		t.Error("expected non-empty output")
	}

	// Check that both sources are in the output (uses contains from detection_test.go)
	if !contains(output, "1.0.0") {
		t.Error("expected version 1.0.0 in output")
	}
	if !contains(output, "package.json") {
		t.Error("expected package.json in output")
	}
	if !contains(output, "2.0.0") {
		t.Error("expected version 2.0.0 in output")
	}
	if !contains(output, "Cargo.toml") {
		t.Error("expected Cargo.toml in output")
	}
}
