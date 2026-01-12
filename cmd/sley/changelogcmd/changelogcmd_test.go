package changelogcmd

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/indaco/sley/internal/config"
	"github.com/indaco/sley/internal/plugins/changeloggenerator"
	"github.com/indaco/sley/internal/testutils"
	"github.com/urfave/cli/v3"
)

/* ------------------------------------------------------------------------- */
/* HELPERS                                                                   */
/* ------------------------------------------------------------------------- */

// createVersionedChangelogFiles creates test versioned changelog files.
func createVersionedChangelogFiles(t *testing.T, changesDir string) {
	t.Helper()

	if err := os.MkdirAll(changesDir, 0755); err != nil {
		t.Fatalf("failed to create changes directory: %v", err)
	}

	versions := []struct {
		filename string
		content  string
	}{
		{
			filename: "v1.0.0.md",
			content: `## v1.0.0 - 2025-01-01

### Features

- Initial release
`,
		},
		{
			filename: "v1.1.0.md",
			content: `## v1.1.0 - 2025-02-01

### Features

- Add new feature
`,
		},
		{
			filename: "v1.2.0.md",
			content: `## v1.2.0 - 2025-03-01

### Fixes

- Fix critical bug
`,
		},
	}

	for _, v := range versions {
		path := filepath.Join(changesDir, v.filename)
		if err := os.WriteFile(path, []byte(v.content), 0644); err != nil {
			t.Fatalf("failed to write %s: %v", v.filename, err)
		}
	}
}

// createCustomHeader creates a custom header template file.
func createCustomHeader(t *testing.T, path string) {
	t.Helper()

	content := `# My Project Changelog

All notable changes to My Project are documented here.

See [Semantic Versioning](https://semver.org/) for versioning guidelines.`

	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("failed to write custom header: %v", err)
	}
}

/* ------------------------------------------------------------------------- */
/* CHANGELOG MERGE COMMAND                                                   */
/* ------------------------------------------------------------------------- */

func TestChangelogMergeCmd_Success(t *testing.T) {
	tmpDir := t.TempDir()
	changesDir := filepath.Join(tmpDir, ".changes")
	outputPath := filepath.Join(tmpDir, "CHANGELOG.md")

	// Create versioned changelog files
	createVersionedChangelogFiles(t, changesDir)

	// Prepare and run the CLI command
	cfg := &config.Config{Path: tmpDir}
	appCli := testutils.BuildCLIForTests(cfg.Path, []*cli.Command{Run(cfg)})

	output, err := testutils.CaptureStdout(func() {
		testutils.RunCLITest(t, appCli, []string{
			"sley", "changelog", "merge",
			"--changes-dir", changesDir,
			"--output", outputPath,
		}, tmpDir)
	})
	if err != nil {
		t.Fatalf("failed to capture stdout: %v", err)
	}

	// Check success message
	expectedMsg := "Successfully merged changelog files"
	if !strings.Contains(output, expectedMsg) {
		t.Errorf("expected output to contain %q, got:\n%s", expectedMsg, output)
	}

	// Verify output file exists
	if _, err := os.Stat(outputPath); os.IsNotExist(err) {
		t.Fatalf("expected output file to exist at %s", outputPath)
	}

	// Read and verify content
	content, err := os.ReadFile(outputPath)
	if err != nil {
		t.Fatalf("failed to read output file: %v", err)
	}

	contentStr := string(content)

	// Check that all versions are present
	versions := []string{"v1.2.0", "v1.1.0", "v1.0.0"}
	for _, version := range versions {
		if !strings.Contains(contentStr, version) {
			t.Errorf("expected content to contain version %s", version)
		}
	}

	// Check that default header is present
	if !strings.Contains(contentStr, "# Changelog") {
		t.Errorf("expected content to contain default header")
	}

	// Verify order: newest version should appear before older versions
	v12Idx := strings.Index(contentStr, "v1.2.0")
	v11Idx := strings.Index(contentStr, "v1.1.0")
	v10Idx := strings.Index(contentStr, "v1.0.0")

	if v12Idx > v11Idx || v11Idx > v10Idx {
		t.Errorf("expected versions to be ordered newest first")
	}
}

func TestChangelogMergeCmd_CustomHeader(t *testing.T) {
	tmpDir := t.TempDir()
	changesDir := filepath.Join(tmpDir, ".changes")
	outputPath := filepath.Join(tmpDir, "CHANGELOG.md")
	headerPath := filepath.Join(tmpDir, "header.md")

	// Create versioned changelog files
	createVersionedChangelogFiles(t, changesDir)

	// Create custom header
	createCustomHeader(t, headerPath)

	// Prepare and run the CLI command
	cfg := &config.Config{Path: tmpDir}
	appCli := testutils.BuildCLIForTests(cfg.Path, []*cli.Command{Run(cfg)})

	_, err := testutils.CaptureStdout(func() {
		testutils.RunCLITest(t, appCli, []string{
			"sley", "changelog", "merge",
			"--changes-dir", changesDir,
			"--output", outputPath,
			"--header-template", headerPath,
		}, tmpDir)
	})
	if err != nil {
		t.Fatalf("failed to capture stdout: %v", err)
	}

	// Read and verify content
	content, err := os.ReadFile(outputPath)
	if err != nil {
		t.Fatalf("failed to read output file: %v", err)
	}

	contentStr := string(content)

	// Check that custom header is present
	if !strings.Contains(contentStr, "# My Project Changelog") {
		t.Errorf("expected content to contain custom header")
	}

	// Verify versions are still present
	if !strings.Contains(contentStr, "v1.2.0") {
		t.Errorf("expected content to contain version v1.2.0")
	}
}

func TestChangelogMergeCmd_NoVersionedFiles(t *testing.T) {
	tmpDir := t.TempDir()
	changesDir := filepath.Join(tmpDir, ".changes")
	outputPath := filepath.Join(tmpDir, "CHANGELOG.md")

	// Create empty changes directory
	if err := os.MkdirAll(changesDir, 0755); err != nil {
		t.Fatalf("failed to create changes directory: %v", err)
	}

	// Prepare and run the CLI command
	cfg := &config.Config{Path: tmpDir}
	appCli := testutils.BuildCLIForTests(cfg.Path, []*cli.Command{Run(cfg)})

	output, err := testutils.CaptureStdout(func() {
		testutils.RunCLITest(t, appCli, []string{
			"sley", "changelog", "merge",
			"--changes-dir", changesDir,
			"--output", outputPath,
		}, tmpDir)
	})
	if err != nil {
		t.Fatalf("failed to capture stdout: %v", err)
	}

	// Should still report success (nothing to merge)
	expectedMsg := "Successfully merged changelog files"
	if !strings.Contains(output, expectedMsg) {
		t.Errorf("expected output to contain %q, got:\n%s", expectedMsg, output)
	}
}

func TestChangelogMergeCmd_WithConfig(t *testing.T) {
	tmpDir := t.TempDir()
	changesDir := filepath.Join(tmpDir, "custom-changes")
	outputPath := filepath.Join(tmpDir, "CUSTOM.md")

	// Create versioned changelog files
	createVersionedChangelogFiles(t, changesDir)

	// Create config with plugin settings
	cfg := &config.Config{
		Path: tmpDir,
		Plugins: &config.PluginConfig{
			ChangelogGenerator: &config.ChangelogGeneratorConfig{
				Enabled:       true,
				ChangesDir:    "custom-changes",
				ChangelogPath: "CUSTOM.md",
			},
		},
	}

	// Prepare and run the CLI command (without flags, should use config)
	appCli := testutils.BuildCLIForTests(cfg.Path, []*cli.Command{Run(cfg)})

	_, err := testutils.CaptureStdout(func() {
		testutils.RunCLITest(t, appCli, []string{
			"sley", "changelog", "merge",
		}, tmpDir)
	})
	if err != nil {
		t.Fatalf("failed to capture stdout: %v", err)
	}

	// Verify output file exists at configured path
	if _, err := os.Stat(outputPath); os.IsNotExist(err) {
		t.Fatalf("expected output file to exist at %s", outputPath)
	}

	// Read and verify content
	content, err := os.ReadFile(outputPath)
	if err != nil {
		t.Fatalf("failed to read output file: %v", err)
	}

	// Check that versions are present
	if !strings.Contains(string(content), "v1.2.0") {
		t.Errorf("expected content to contain version v1.2.0")
	}
}

func TestChangelogMergeCmd_FlagOverridesConfig(t *testing.T) {
	tmpDir := t.TempDir()
	flagChangesDir := filepath.Join(tmpDir, "flag-changes")
	configOutputPath := filepath.Join(tmpDir, "CONFIG.md")
	flagOutputPath := filepath.Join(tmpDir, "FLAG.md")

	// Create versioned changelog files in flag-specified directory
	createVersionedChangelogFiles(t, flagChangesDir)

	// Create config with different settings
	cfg := &config.Config{
		Path: tmpDir,
		Plugins: &config.PluginConfig{
			ChangelogGenerator: &config.ChangelogGeneratorConfig{
				Enabled:       true,
				ChangesDir:    "config-changes",
				ChangelogPath: "CONFIG.md",
			},
		},
	}

	// Prepare and run the CLI command with flags (should override config)
	appCli := testutils.BuildCLIForTests(cfg.Path, []*cli.Command{Run(cfg)})

	_, err := testutils.CaptureStdout(func() {
		testutils.RunCLITest(t, appCli, []string{
			"sley", "changelog", "merge",
			"--changes-dir", flagChangesDir,
			"--output", flagOutputPath,
		}, tmpDir)
	})
	if err != nil {
		t.Fatalf("failed to capture stdout: %v", err)
	}

	// Verify output file exists at flag-specified path
	if _, err := os.Stat(flagOutputPath); os.IsNotExist(err) {
		t.Fatalf("expected output file to exist at %s", flagOutputPath)
	}

	// Verify config path was NOT used
	if _, err := os.Stat(configOutputPath); !os.IsNotExist(err) {
		t.Errorf("expected config output path to NOT be used")
	}

	// Read and verify content
	content, err := os.ReadFile(flagOutputPath)
	if err != nil {
		t.Fatalf("failed to read output file: %v", err)
	}

	// Check that versions are present
	if !strings.Contains(string(content), "v1.2.0") {
		t.Errorf("expected content to contain version v1.2.0")
	}
}

func TestChangelogMergeCmd_MissingChangesDir(t *testing.T) {
	tmpDir := t.TempDir()
	nonExistentDir := filepath.Join(tmpDir, "nonexistent")
	outputPath := filepath.Join(tmpDir, "CHANGELOG.md")

	// Prepare and run the CLI command
	cfg := &config.Config{Path: tmpDir}
	appCli := testutils.BuildCLIForTests(cfg.Path, []*cli.Command{Run(cfg)})

	var cliErr error
	_, captureErr := testutils.CaptureStdout(func() {
		cliErr = testutils.RunCLITestAllowError(t, appCli, []string{
			"sley", "changelog", "merge",
			"--changes-dir", nonExistentDir,
			"--output", outputPath,
		}, tmpDir)
	})
	if captureErr != nil {
		t.Fatalf("failed to capture output: %v", captureErr)
	}

	// Should return error for missing directory
	if cliErr == nil {
		t.Fatal("expected error for missing changes directory, got nil")
	}

	expectedMsg := "failed to merge changelog files"
	if !strings.Contains(cliErr.Error(), expectedMsg) {
		t.Errorf("expected error to contain %q, got: %v", expectedMsg, cliErr)
	}
}

/* ------------------------------------------------------------------------- */
/* BUILD CONFIG HELPER                                                       */
/* ------------------------------------------------------------------------- */

func TestBuildGeneratorConfig_DefaultValues(t *testing.T) {
	genCfg := changeloggenerator.DefaultConfig()

	// Should use defaults
	if genCfg.ChangesDir != ".changes" {
		t.Errorf("expected ChangesDir to be '.changes', got %s", genCfg.ChangesDir)
	}
	if genCfg.ChangelogPath != "CHANGELOG.md" {
		t.Errorf("expected ChangelogPath to be 'CHANGELOG.md', got %s", genCfg.ChangelogPath)
	}
}

/* ------------------------------------------------------------------------- */
/* CHANGELOG MERGE COMMAND - PLUGIN WARNINGS                                 */
/* ------------------------------------------------------------------------- */

func TestChangelogMergeCmd_WarningWhenPluginDisabled(t *testing.T) {
	tmpDir := t.TempDir()
	changesDir := filepath.Join(tmpDir, ".changes")
	outputPath := filepath.Join(tmpDir, "CHANGELOG.md")

	// Create versioned changelog files
	createVersionedChangelogFiles(t, changesDir)

	// Config without changelog-generator enabled
	cfg := &config.Config{Path: tmpDir}
	appCli := testutils.BuildCLIForTests(cfg.Path, []*cli.Command{Run(cfg)})

	output, err := testutils.CaptureStdout(func() {
		testutils.RunCLITest(t, appCli, []string{
			"sley", "changelog", "merge",
			"--changes-dir", changesDir,
			"--output", outputPath,
		}, tmpDir)
	})
	if err != nil {
		t.Fatalf("failed to capture stdout: %v", err)
	}

	// Should show warning about plugin not being enabled
	expectedWarning := "changelog-generator plugin is not enabled"
	if !strings.Contains(output, expectedWarning) {
		t.Errorf("expected output to contain warning %q, got:\n%s", expectedWarning, output)
	}

	// Should still succeed with merge
	expectedMsg := "Successfully merged changelog files"
	if !strings.Contains(output, expectedMsg) {
		t.Errorf("expected output to contain %q, got:\n%s", expectedMsg, output)
	}
}

func TestChangelogMergeCmd_WarningWhenMergeAfterImmediate(t *testing.T) {
	tmpDir := t.TempDir()
	changesDir := filepath.Join(tmpDir, ".changes")
	outputPath := filepath.Join(tmpDir, "CHANGELOG.md")

	// Create versioned changelog files
	createVersionedChangelogFiles(t, changesDir)

	// Config with merge-after set to immediate
	cfg := &config.Config{
		Path: tmpDir,
		Plugins: &config.PluginConfig{
			ChangelogGenerator: &config.ChangelogGeneratorConfig{
				Enabled:       true,
				ChangesDir:    changesDir,
				ChangelogPath: outputPath,
				MergeAfter:    "immediate",
			},
		},
	}
	appCli := testutils.BuildCLIForTests(cfg.Path, []*cli.Command{Run(cfg)})

	output, err := testutils.CaptureStdout(func() {
		testutils.RunCLITest(t, appCli, []string{
			"sley", "changelog", "merge",
		}, tmpDir)
	})
	if err != nil {
		t.Fatalf("failed to capture stdout: %v", err)
	}

	// Should show warning about merge-after being set
	expectedWarning := "'merge-after' is set to 'immediate'"
	if !strings.Contains(output, expectedWarning) {
		t.Errorf("expected output to contain warning %q, got:\n%s", expectedWarning, output)
	}

	// Should still succeed with merge
	expectedMsg := "Successfully merged changelog files"
	if !strings.Contains(output, expectedMsg) {
		t.Errorf("expected output to contain %q, got:\n%s", expectedMsg, output)
	}
}

func TestChangelogMergeCmd_NoWarningWhenMergeAfterManual(t *testing.T) {
	tmpDir := t.TempDir()
	changesDir := filepath.Join(tmpDir, ".changes")
	outputPath := filepath.Join(tmpDir, "CHANGELOG.md")

	// Create versioned changelog files
	createVersionedChangelogFiles(t, changesDir)

	// Config with merge-after set to manual (default)
	cfg := &config.Config{
		Path: tmpDir,
		Plugins: &config.PluginConfig{
			ChangelogGenerator: &config.ChangelogGeneratorConfig{
				Enabled:       true,
				ChangesDir:    changesDir,
				ChangelogPath: outputPath,
				MergeAfter:    "manual",
			},
		},
	}
	appCli := testutils.BuildCLIForTests(cfg.Path, []*cli.Command{Run(cfg)})

	output, err := testutils.CaptureStdout(func() {
		testutils.RunCLITest(t, appCli, []string{
			"sley", "changelog", "merge",
		}, tmpDir)
	})
	if err != nil {
		t.Fatalf("failed to capture stdout: %v", err)
	}

	// Should NOT show merge-after warning
	unexpectedWarning := "'merge-after' is set to"
	if strings.Contains(output, unexpectedWarning) {
		t.Errorf("expected output to NOT contain warning %q, got:\n%s", unexpectedWarning, output)
	}

	// Should NOT show plugin disabled warning
	unexpectedWarning2 := "plugin is not enabled"
	if strings.Contains(output, unexpectedWarning2) {
		t.Errorf("expected output to NOT contain warning %q, got:\n%s", unexpectedWarning2, output)
	}

	// Should still succeed with merge
	expectedMsg := "Successfully merged changelog files"
	if !strings.Contains(output, expectedMsg) {
		t.Errorf("expected output to contain %q, got:\n%s", expectedMsg, output)
	}
}

func TestIsChangelogGeneratorEnabled(t *testing.T) {
	tests := []struct {
		name     string
		cfg      *config.Config
		expected bool
	}{
		{
			name:     "nil config",
			cfg:      nil,
			expected: false,
		},
		{
			name:     "nil plugins",
			cfg:      &config.Config{},
			expected: false,
		},
		{
			name: "nil changelog generator",
			cfg: &config.Config{
				Plugins: &config.PluginConfig{},
			},
			expected: false,
		},
		{
			name: "disabled changelog generator",
			cfg: &config.Config{
				Plugins: &config.PluginConfig{
					ChangelogGenerator: &config.ChangelogGeneratorConfig{
						Enabled: false,
					},
				},
			},
			expected: false,
		},
		{
			name: "enabled changelog generator",
			cfg: &config.Config{
				Plugins: &config.PluginConfig{
					ChangelogGenerator: &config.ChangelogGeneratorConfig{
						Enabled: true,
					},
				},
			},
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isChangelogGeneratorEnabled(tt.cfg)
			if result != tt.expected {
				t.Errorf("isChangelogGeneratorEnabled() = %v, want %v", result, tt.expected)
			}
		})
	}
}
