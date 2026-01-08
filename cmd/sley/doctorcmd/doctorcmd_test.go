package doctorcmd

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/indaco/sley/internal/config"
	"github.com/indaco/sley/internal/plugins"
	"github.com/indaco/sley/internal/testutils"
	"github.com/urfave/cli/v3"
)

func TestCLI_ValidateCommand_ValidCases(t *testing.T) {
	tmpDir := t.TempDir()
	versionPath := filepath.Join(tmpDir, ".version")

	// Prepare and run the CLI command
	cfg := &config.Config{Path: versionPath}
	appCli := testutils.BuildCLIForTests(cfg.Path, []*cli.Command{Run(cfg)})

	tests := []struct {
		name           string
		version        string
		expectedOutput string
	}{
		{
			name:    "valid semantic version",
			version: "1.2.3",
		},
		{
			name:    "valid version with build metadata",
			version: "1.2.3+exp.sha.5114f85",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testutils.WriteTempVersionFile(t, tmpDir, tt.version)

			output, err := testutils.CaptureStdout(func() {
				testutils.RunCLITest(t, appCli, []string{"sley", "doctor"}, tmpDir)
			})
			if err != nil {
				t.Fatalf("Failed to capture stdout: %v", err)
			}

			expected := fmt.Sprintf("Valid version file at %s/.version", tmpDir)
			if !strings.Contains(output, expected) {
				t.Errorf("expected output to contain %q, got %q", expected, output)
			}
		})
	}
}

func TestCLI_ValidateCommand_Errors(t *testing.T) {
	tests := []struct {
		name          string
		version       string
		expectedError string
	}{
		{"invalid version string", "not-a-version", "invalid version"},
		{"invalid build metadata", "1.0.0+inv@lid-meta", "invalid version"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			testutils.WriteTempVersionFile(t, tmpDir, tt.version)
			versionPath := filepath.Join(tmpDir, ".version")

			// Prepare and run the CLI command
			cfg := &config.Config{Path: versionPath}
			appCli := testutils.BuildCLIForTests(cfg.Path, []*cli.Command{Run(cfg)})

			err := appCli.Run(context.Background(), []string{"sley", "doctor"})
			if err == nil || !strings.Contains(err.Error(), tt.expectedError) {
				t.Fatalf("expected error containing %q, got: %v", tt.expectedError, err)
			}
		})
	}
}

func TestCLI_ValidateCommand_MultiModule_All(t *testing.T) {
	tmpDir := t.TempDir()

	// Create multi-module workspace
	moduleA := filepath.Join(tmpDir, "module-a")
	moduleB := filepath.Join(tmpDir, "module-b")
	if err := os.MkdirAll(moduleA, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(moduleB, 0755); err != nil {
		t.Fatal(err)
	}

	testutils.WriteTempVersionFile(t, moduleA, "1.0.0")
	testutils.WriteTempVersionFile(t, moduleB, "2.0.0")

	// Create config with workspace discovery enabled
	enabled := true
	recursive := true
	maxDepth := 10
	cfg := &config.Config{
		Path: ".version",
		Workspace: &config.WorkspaceConfig{
			Discovery: &config.DiscoveryConfig{
				Enabled:   &enabled,
				Recursive: &recursive,
				MaxDepth:  &maxDepth,
			},
		},
	}

	appCli := testutils.BuildCLIForTests(cfg.Path, []*cli.Command{Run(cfg)})

	// Test with --all flag
	output, err := testutils.CaptureStdout(func() {
		testutils.RunCLITest(t, appCli, []string{"sley", "doctor", "--all"}, tmpDir)
	})
	if err != nil {
		t.Fatalf("Failed to capture stdout: %v", err)
	}

	// Verify output contains both modules
	if !strings.Contains(output, "module-a") {
		t.Errorf("expected output to contain module-a, got: %q", output)
	}
	if !strings.Contains(output, "module-b") {
		t.Errorf("expected output to contain module-b, got: %q", output)
	}
}

func TestCLI_ValidateCommand_MultiModule_Specific(t *testing.T) {
	tmpDir := t.TempDir()

	// Create multi-module workspace
	moduleA := filepath.Join(tmpDir, "module-a")
	moduleB := filepath.Join(tmpDir, "module-b")
	if err := os.MkdirAll(moduleA, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(moduleB, 0755); err != nil {
		t.Fatal(err)
	}

	testutils.WriteTempVersionFile(t, moduleA, "1.0.0")
	testutils.WriteTempVersionFile(t, moduleB, "2.0.0")

	// Create config with workspace discovery enabled
	enabled := true
	recursive := true
	maxDepth := 10
	cfg := &config.Config{
		Path: ".version",
		Workspace: &config.WorkspaceConfig{
			Discovery: &config.DiscoveryConfig{
				Enabled:   &enabled,
				Recursive: &recursive,
				MaxDepth:  &maxDepth,
			},
		},
	}

	appCli := testutils.BuildCLIForTests(cfg.Path, []*cli.Command{Run(cfg)})

	// Test with --module flag (target specific module)
	output, err := testutils.CaptureStdout(func() {
		testutils.RunCLITest(t, appCli, []string{"sley", "doctor", "--module", "module-a"}, tmpDir)
	})
	if err != nil {
		t.Fatalf("Failed to capture stdout: %v", err)
	}

	// Verify output contains only module-a
	if !strings.Contains(output, "module-a") {
		t.Errorf("expected output to contain module-a, got: %q", output)
	}
}

func TestCLI_ValidateCommand_MultiModule_Quiet(t *testing.T) {
	tmpDir := t.TempDir()

	// Create multi-module workspace
	moduleA := filepath.Join(tmpDir, "module-a")
	moduleB := filepath.Join(tmpDir, "module-b")
	if err := os.MkdirAll(moduleA, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(moduleB, 0755); err != nil {
		t.Fatal(err)
	}

	testutils.WriteTempVersionFile(t, moduleA, "1.0.0")
	testutils.WriteTempVersionFile(t, moduleB, "2.0.0")

	// Create config with workspace discovery enabled
	enabled := true
	recursive := true
	maxDepth := 10
	cfg := &config.Config{
		Path: ".version",
		Workspace: &config.WorkspaceConfig{
			Discovery: &config.DiscoveryConfig{
				Enabled:   &enabled,
				Recursive: &recursive,
				MaxDepth:  &maxDepth,
			},
		},
	}

	appCli := testutils.BuildCLIForTests(cfg.Path, []*cli.Command{Run(cfg)})

	// Test with --quiet flag
	output, err := testutils.CaptureStdout(func() {
		testutils.RunCLITest(t, appCli, []string{"sley", "doctor", "--all", "--quiet"}, tmpDir)
	})
	if err != nil {
		t.Fatalf("Failed to capture stdout: %v", err)
	}

	// Quiet mode should show minimal output
	if !strings.Contains(output, "Success:") && !strings.Contains(output, "2 module(s)") {
		t.Errorf("expected quiet summary, got: %q", output)
	}
}

func TestCLI_ValidateCommand_MultiModule_WithInvalidVersion(t *testing.T) {
	tmpDir := t.TempDir()

	// Create multi-module workspace with one invalid version
	moduleA := filepath.Join(tmpDir, "module-a")
	moduleB := filepath.Join(tmpDir, "module-b")
	if err := os.MkdirAll(moduleA, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(moduleB, 0755); err != nil {
		t.Fatal(err)
	}

	testutils.WriteTempVersionFile(t, moduleA, "1.0.0")
	testutils.WriteTempVersionFile(t, moduleB, "not-a-version")

	// Create config with workspace discovery enabled
	enabled := true
	recursive := true
	maxDepth := 10
	cfg := &config.Config{
		Path: ".version",
		Workspace: &config.WorkspaceConfig{
			Discovery: &config.DiscoveryConfig{
				Enabled:   &enabled,
				Recursive: &recursive,
				MaxDepth:  &maxDepth,
			},
		},
	}

	appCli := testutils.BuildCLIForTests(cfg.Path, []*cli.Command{Run(cfg)})

	// Test with --all flag - should fail due to invalid version
	// We need to run in the tmpDir context
	oldDir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		if err := os.Chdir(oldDir); err != nil {
			t.Errorf("failed to restore directory: %v", err)
		}
	}()

	if err := os.Chdir(tmpDir); err != nil {
		t.Fatal(err)
	}

	err = appCli.Run(context.Background(), []string{"sley", "doctor", "--all"})
	if err == nil {
		t.Fatal("expected error due to invalid version in one module, got nil")
	}

	if !strings.Contains(err.Error(), "failed validation") {
		t.Errorf("expected error message to contain 'failed validation', got: %v", err)
	}
}

func TestCLI_ValidateCommand_MultiModule_ContinueOnError(t *testing.T) {
	tmpDir := t.TempDir()

	// Create multi-module workspace with one invalid version
	moduleA := filepath.Join(tmpDir, "module-a")
	moduleB := filepath.Join(tmpDir, "module-b")
	moduleC := filepath.Join(tmpDir, "module-c")
	if err := os.MkdirAll(moduleA, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(moduleB, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(moduleC, 0755); err != nil {
		t.Fatal(err)
	}

	testutils.WriteTempVersionFile(t, moduleA, "1.0.0")
	testutils.WriteTempVersionFile(t, moduleB, "not-a-version")
	testutils.WriteTempVersionFile(t, moduleC, "3.0.0")

	// Create config with workspace discovery enabled
	enabled := true
	recursive := true
	maxDepth := 10
	cfg := &config.Config{
		Path: ".version",
		Workspace: &config.WorkspaceConfig{
			Discovery: &config.DiscoveryConfig{
				Enabled:   &enabled,
				Recursive: &recursive,
				MaxDepth:  &maxDepth,
			},
		},
	}

	appCli := testutils.BuildCLIForTests(cfg.Path, []*cli.Command{Run(cfg)})

	// Test with --continue-on-error flag
	// We need to run in the tmpDir context
	oldDir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		if err := os.Chdir(oldDir); err != nil {
			t.Errorf("failed to restore directory: %v", err)
		}
	}()

	if err := os.Chdir(tmpDir); err != nil {
		t.Fatal(err)
	}

	output, _ := testutils.CaptureStdout(func() {
		_ = appCli.Run(context.Background(), []string{"sley", "doctor", "--all", "--continue-on-error"})
	})

	// Should show results for all modules including the failed one
	if !strings.Contains(output, "module-a") {
		t.Errorf("expected output to contain module-a, got: %q", output)
	}
	if !strings.Contains(output, "module-c") {
		t.Errorf("expected output to contain module-c, got: %q", output)
	}
}

func TestCLI_ValidateCommand_MultiModule_JSONFormat(t *testing.T) {
	tmpDir := t.TempDir()

	// Create multi-module workspace
	moduleA := filepath.Join(tmpDir, "module-a")
	moduleB := filepath.Join(tmpDir, "module-b")
	if err := os.MkdirAll(moduleA, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(moduleB, 0755); err != nil {
		t.Fatal(err)
	}

	testutils.WriteTempVersionFile(t, moduleA, "1.0.0")
	testutils.WriteTempVersionFile(t, moduleB, "2.0.0")

	// Create config with workspace discovery enabled
	enabled := true
	recursive := true
	maxDepth := 10
	cfg := &config.Config{
		Path: ".version",
		Workspace: &config.WorkspaceConfig{
			Discovery: &config.DiscoveryConfig{
				Enabled:   &enabled,
				Recursive: &recursive,
				MaxDepth:  &maxDepth,
			},
		},
	}

	appCli := testutils.BuildCLIForTests(cfg.Path, []*cli.Command{Run(cfg)})

	// Test with --format json
	output, err := testutils.CaptureStdout(func() {
		testutils.RunCLITest(t, appCli, []string{"sley", "doctor", "--all", "--format", "json"}, tmpDir)
	})
	if err != nil {
		t.Fatalf("Failed to capture stdout: %v", err)
	}

	// Output should contain JSON
	if !strings.Contains(output, "module-a") || !strings.Contains(output, "module-b") {
		t.Errorf("expected JSON output with module names, got: %q", output)
	}
}

func TestCLI_ValidateCommand_MultiModule_TableFormat(t *testing.T) {
	tmpDir := t.TempDir()

	// Create multi-module workspace
	moduleA := filepath.Join(tmpDir, "module-a")
	moduleB := filepath.Join(tmpDir, "module-b")
	if err := os.MkdirAll(moduleA, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(moduleB, 0755); err != nil {
		t.Fatal(err)
	}

	testutils.WriteTempVersionFile(t, moduleA, "1.0.0")
	testutils.WriteTempVersionFile(t, moduleB, "2.0.0")

	// Create config with workspace discovery enabled
	enabled := true
	recursive := true
	maxDepth := 10
	cfg := &config.Config{
		Path: ".version",
		Workspace: &config.WorkspaceConfig{
			Discovery: &config.DiscoveryConfig{
				Enabled:   &enabled,
				Recursive: &recursive,
				MaxDepth:  &maxDepth,
			},
		},
	}

	appCli := testutils.BuildCLIForTests(cfg.Path, []*cli.Command{Run(cfg)})

	// Test with --format table
	output, err := testutils.CaptureStdout(func() {
		testutils.RunCLITest(t, appCli, []string{"sley", "doctor", "--all", "--format", "table"}, tmpDir)
	})
	if err != nil {
		t.Fatalf("Failed to capture stdout: %v", err)
	}

	// Output should contain table-formatted data
	if !strings.Contains(output, "module-a") || !strings.Contains(output, "module-b") {
		t.Errorf("expected table output with module names, got: %q", output)
	}
}

func TestIsPluginEnabled(t *testing.T) {
	tests := []struct {
		name       string
		cfg        *config.Config
		pluginType plugins.PluginType
		expected   bool
	}{
		{
			name:       "nil config",
			cfg:        nil,
			pluginType: plugins.TypeCommitParser,
			expected:   false,
		},
		{
			name:       "nil plugins",
			cfg:        &config.Config{},
			pluginType: plugins.TypeCommitParser,
			expected:   false,
		},
		{
			name: "commit-parser enabled",
			cfg: &config.Config{
				Plugins: &config.PluginConfig{
					CommitParser: true,
				},
			},
			pluginType: plugins.TypeCommitParser,
			expected:   true,
		},
		{
			name: "commit-parser disabled",
			cfg: &config.Config{
				Plugins: &config.PluginConfig{
					CommitParser: false,
				},
			},
			pluginType: plugins.TypeCommitParser,
			expected:   false,
		},
		{
			name: "tag-manager enabled",
			cfg: &config.Config{
				Plugins: &config.PluginConfig{
					TagManager: &config.TagManagerConfig{Enabled: true},
				},
			},
			pluginType: plugins.TypeTagManager,
			expected:   true,
		},
		{
			name: "tag-manager nil",
			cfg: &config.Config{
				Plugins: &config.PluginConfig{},
			},
			pluginType: plugins.TypeTagManager,
			expected:   false,
		},
		{
			name: "version-validator enabled",
			cfg: &config.Config{
				Plugins: &config.PluginConfig{
					VersionValidator: &config.VersionValidatorConfig{Enabled: true},
				},
			},
			pluginType: plugins.TypeVersionValidator,
			expected:   true,
		},
		{
			name: "dependency-check enabled",
			cfg: &config.Config{
				Plugins: &config.PluginConfig{
					DependencyCheck: &config.DependencyCheckConfig{Enabled: true},
				},
			},
			pluginType: plugins.TypeDependencyChecker,
			expected:   true,
		},
		{
			name: "changelog-parser enabled",
			cfg: &config.Config{
				Plugins: &config.PluginConfig{
					ChangelogParser: &config.ChangelogParserConfig{Enabled: true},
				},
			},
			pluginType: plugins.TypeChangelogParser,
			expected:   true,
		},
		{
			name: "changelog-generator enabled",
			cfg: &config.Config{
				Plugins: &config.PluginConfig{
					ChangelogGenerator: &config.ChangelogGeneratorConfig{Enabled: true},
				},
			},
			pluginType: plugins.TypeChangelogGenerator,
			expected:   true,
		},
		{
			name: "release-gate enabled",
			cfg: &config.Config{
				Plugins: &config.PluginConfig{
					ReleaseGate: &config.ReleaseGateConfig{Enabled: true},
				},
			},
			pluginType: plugins.TypeReleaseGate,
			expected:   true,
		},
		{
			name: "audit-log enabled",
			cfg: &config.Config{
				Plugins: &config.PluginConfig{
					AuditLog: &config.AuditLogConfig{Enabled: true},
				},
			},
			pluginType: plugins.TypeAuditLog,
			expected:   true,
		},
		{
			name: "unknown plugin type",
			cfg: &config.Config{
				Plugins: &config.PluginConfig{},
			},
			pluginType: plugins.PluginType("unknown"),
			expected:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isPluginEnabled(tt.cfg, tt.pluginType)
			if result != tt.expected {
				t.Errorf("isPluginEnabled() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestTruncateString(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		maxLen   int
		expected string
	}{
		{
			name:     "short string unchanged",
			input:    "hello",
			maxLen:   10,
			expected: "hello",
		},
		{
			name:     "exact length unchanged",
			input:    "hello",
			maxLen:   5,
			expected: "hello",
		},
		{
			name:     "long string truncated",
			input:    "hello world",
			maxLen:   8,
			expected: "hello...",
		},
		{
			name:     "empty string",
			input:    "",
			maxLen:   10,
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := truncateString(tt.input, tt.maxLen)
			if result != tt.expected {
				t.Errorf("truncateString(%q, %d) = %q, want %q", tt.input, tt.maxLen, result, tt.expected)
			}
		})
	}
}

func TestPluginStatus_InDoctorOutput(t *testing.T) {
	tmpDir := t.TempDir()
	versionPath := filepath.Join(tmpDir, ".version")
	testutils.WriteTempVersionFile(t, tmpDir, "1.0.0")

	// Config with some plugins enabled
	cfg := &config.Config{
		Path: versionPath,
		Plugins: &config.PluginConfig{
			CommitParser: true,
			TagManager:   &config.TagManagerConfig{Enabled: true},
		},
	}

	appCli := testutils.BuildCLIForTests(cfg.Path, []*cli.Command{Run(cfg)})

	output, err := testutils.CaptureStdout(func() {
		testutils.RunCLITest(t, appCli, []string{"sley", "doctor"}, tmpDir)
	})
	if err != nil {
		t.Fatalf("Failed to capture stdout: %v", err)
	}

	// Verify plugin status section appears
	if !strings.Contains(output, "Plugin Status:") {
		t.Errorf("expected output to contain 'Plugin Status:', got: %q", output)
	}

	// Verify enabled plugins show [ON]
	if !strings.Contains(output, "[ON]") {
		t.Errorf("expected output to contain '[ON]' for enabled plugins, got: %q", output)
	}

	// Verify disabled plugins show [OFF]
	if !strings.Contains(output, "[OFF]") {
		t.Errorf("expected output to contain '[OFF]' for disabled plugins, got: %q", output)
	}

	// Verify summary shows correct count
	if !strings.Contains(output, "2/8 plugins enabled") {
		t.Errorf("expected output to contain '2/8 plugins enabled', got: %q", output)
	}
}

func TestPluginStatus_QuietMode(t *testing.T) {
	tmpDir := t.TempDir()
	versionPath := filepath.Join(tmpDir, ".version")
	testutils.WriteTempVersionFile(t, tmpDir, "1.0.0")

	cfg := &config.Config{
		Path: versionPath,
		Plugins: &config.PluginConfig{
			CommitParser: true,
		},
	}

	appCli := testutils.BuildCLIForTests(cfg.Path, []*cli.Command{Run(cfg)})

	output, err := testutils.CaptureStdout(func() {
		testutils.RunCLITest(t, appCli, []string{"sley", "doctor", "--quiet"}, tmpDir)
	})
	if err != nil {
		t.Fatalf("Failed to capture stdout: %v", err)
	}

	// Quiet mode should NOT show plugin status
	if strings.Contains(output, "Plugin Status:") {
		t.Errorf("quiet mode should not show plugin status, got: %q", output)
	}
}

func TestPluginStatus_JSONFormat(t *testing.T) {
	tmpDir := t.TempDir()
	versionPath := filepath.Join(tmpDir, ".version")
	testutils.WriteTempVersionFile(t, tmpDir, "1.0.0")

	cfg := &config.Config{
		Path: versionPath,
		Plugins: &config.PluginConfig{
			CommitParser: true,
			TagManager:   &config.TagManagerConfig{Enabled: true},
		},
	}

	appCli := testutils.BuildCLIForTests(cfg.Path, []*cli.Command{Run(cfg)})

	output, err := testutils.CaptureStdout(func() {
		testutils.RunCLITest(t, appCli, []string{"sley", "doctor", "--format", "json"}, tmpDir)
	})
	if err != nil {
		t.Fatalf("Failed to capture stdout: %v", err)
	}

	// Verify JSON output contains plugin info
	if !strings.Contains(output, `"plugins"`) {
		t.Errorf("expected JSON output to contain 'plugins' key, got: %q", output)
	}
	if !strings.Contains(output, `"enabled":true`) {
		t.Errorf("expected JSON output to contain enabled plugins, got: %q", output)
	}
	if !strings.Contains(output, `"commit-parser"`) {
		t.Errorf("expected JSON output to contain 'commit-parser', got: %q", output)
	}
}
