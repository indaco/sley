package initcmd

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/goccy/go-yaml"
	"github.com/indaco/sley/cmd/sley/bumpcmd"
	"github.com/indaco/sley/cmd/sley/precmd"
	"github.com/indaco/sley/internal/config"
	"github.com/indaco/sley/internal/testutils"
	"github.com/urfave/cli/v3"
)

func TestCLI_InitCommand_CreatesFile(t *testing.T) {
	tmp := t.TempDir()
	versionPath := filepath.Join(tmp, ".version")

	// Use t.Chdir which properly handles directory restoration
	t.Chdir(tmp)

	// Prepare and run the CLI command with --yes to skip interactive prompts
	cfg := &config.Config{Path: versionPath}
	appCli := testutils.BuildCLIForTests(
		cfg.Path,
		[]*cli.Command{Run(), bumpcmd.Run(cfg)},
	)

	output, err := testutils.CaptureStdout(func() {
		testutils.RunCLITest(t, appCli, []string{"sley", "init", "--yes"}, tmp)
	})
	if err != nil {
		t.Fatalf("Failed to capture stdout: %v", err)
	}

	got := testutils.ReadTempVersionFile(t, tmp)
	if got != "0.1.0" {
		t.Errorf("expected version '0.1.0', got %q", got)
	}

	// With --yes flag, output includes config creation message
	if !strings.Contains(output, fmt.Sprintf("Created %s with version 0.1.0", versionPath)) {
		t.Errorf("expected output to contain version creation message, got: %q", output)
	}
}

func TestCLI_InitCommand_InitializationError(t *testing.T) {
	tmp := t.TempDir()
	noWrite := filepath.Join(tmp, "nowrite")
	if err := os.Mkdir(noWrite, 0555); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		_ = os.Chmod(noWrite, 0755)
	})

	versionPath := filepath.Join(noWrite, ".version")

	// Prepare and run the CLI command
	cfg := &config.Config{Path: versionPath}
	appCli := testutils.BuildCLIForTests(
		cfg.Path,
		[]*cli.Command{Run(), bumpcmd.Run(cfg)},
	)

	err := appCli.Run(context.Background(), []string{"sley", "init"})
	if err == nil {
		t.Fatal("expected initialization error, got nil")
	}
	if !strings.Contains(err.Error(), "permission denied") {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestCLI_InitCommand_FileAlreadyExists(t *testing.T) {
	tmpDir := t.TempDir()
	testutils.WriteTempVersionFile(t, tmpDir, "1.2.3")
	versionPath := filepath.Join(tmpDir, ".version")
	configPath := filepath.Join(tmpDir, ".sley.yaml")

	t.Chdir(tmpDir)

	// Prepare and run the CLI command with --yes flag to avoid prompts
	cfg := &config.Config{Path: versionPath}
	appCli := testutils.BuildCLIForTests(
		cfg.Path,
		[]*cli.Command{Run(), bumpcmd.Run(cfg)},
	)

	err := appCli.Run(context.Background(), []string{"sley", "init", "--yes"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify .version file still has original content
	content := testutils.ReadTempVersionFile(t, tmpDir)
	if content != "1.2.3" {
		t.Errorf("expected version file to remain '1.2.3', got %q", content)
	}

	// Verify .sley.yaml was created
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		t.Error("expected .sley.yaml to be created")
	}
}

func TestCLI_InitCommand_ExistingInvalidContent(t *testing.T) {
	tmp := t.TempDir()
	path := filepath.Join(tmp, ".version")

	// Create a file with invalid version content (simulating manual corruption)
	if err := os.WriteFile(path, []byte("not-a-version\n"), 0600); err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}

	// Prepare and run the CLI command
	cfg := &config.Config{Path: path}
	appCli := testutils.BuildCLIForTests(
		cfg.Path,
		[]*cli.Command{Run(), bumpcmd.Run(cfg)},
	)

	err := appCli.Run(context.Background(), []string{
		"sley", "init", "--path", path,
	})
	if err == nil {
		t.Fatal("expected error due to invalid version content, got nil")
	}
	if !strings.Contains(err.Error(), "failed to read version file") {
		t.Errorf("unexpected error: %v", err)
	}
	if !strings.Contains(err.Error(), "invalid version format") {
		t.Errorf("expected invalid version format message, got %v", err)
	}
}

func TestCLI_Command_InitializeVersionFilePermissionErrors(t *testing.T) {
	tests := []struct {
		name    string
		command []string
	}{
		{"bump minor", []string{"sley", "bump", "minor"}},
		{"bump major", []string{"sley", "bump", "major"}},
		{"pre label", []string{"sley", "pre", "--label", "alpha"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmp := t.TempDir()
			noWrite := filepath.Join(tmp, "protected")
			if err := os.Mkdir(noWrite, 0555); err != nil {
				t.Fatal(err)
			}
			t.Cleanup(func() {
				_ = os.Chmod(noWrite, 0755)
			})
			protectedPath := filepath.Join(noWrite, ".version")

			// Prepare and run the CLI command
			cfg := &config.Config{Path: protectedPath}
			appCli := testutils.BuildCLIForTests(
				cfg.Path,
				[]*cli.Command{Run(), bumpcmd.Run(cfg), precmd.Run()},
			)

			err := appCli.Run(context.Background(), append(tt.command, "--path", protectedPath))
			if err == nil || !strings.Contains(err.Error(), "permission denied") {
				t.Fatalf("expected permission denied error, got: %v", err)
			}
		})
	}
}

func TestCLI_InitCommand_WithYesFlag(t *testing.T) {
	tmp := t.TempDir()
	versionPath := filepath.Join(tmp, ".version")
	configPath := filepath.Join(tmp, ".sley.yaml")

	t.Chdir(tmp)

	cfg := &config.Config{Path: versionPath}
	appCli := testutils.BuildCLIForTests(
		cfg.Path,
		[]*cli.Command{Run()},
	)

	err := appCli.Run(context.Background(), []string{"sley", "init", "--yes"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify .version was created
	if _, err := os.Stat(versionPath); os.IsNotExist(err) {
		t.Error("expected .version file to be created")
	}

	// Verify .sley.yaml was created
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		t.Error("expected .sley.yaml file to be created")
	}

	// Verify config contains default plugins
	configData, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatalf("failed to read config: %v", err)
	}

	var loadedCfg config.Config
	if err := yaml.Unmarshal(configData, &loadedCfg); err != nil {
		t.Fatalf("failed to parse config: %v", err)
	}

	if loadedCfg.Plugins == nil {
		t.Fatal("expected plugins config")
	}

	// Verify default plugins are enabled
	if !loadedCfg.Plugins.CommitParser {
		t.Error("expected commit-parser to be enabled by default")
	}
	if loadedCfg.Plugins.TagManager == nil || !loadedCfg.Plugins.TagManager.Enabled {
		t.Error("expected tag-manager to be enabled by default")
	}
}

func TestCLI_InitCommand_WithEnableFlag(t *testing.T) {
	tmp := t.TempDir()
	versionPath := filepath.Join(tmp, ".version")
	configPath := filepath.Join(tmp, ".sley.yaml")

	t.Chdir(tmp)

	cfg := &config.Config{Path: versionPath}
	appCli := testutils.BuildCLIForTests(
		cfg.Path,
		[]*cli.Command{Run()},
	)

	err := appCli.Run(context.Background(), []string{
		"sley", "init",
		"--enable", "commit-parser,changelog-generator,audit-log",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify config was created
	configData, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatalf("failed to read config: %v", err)
	}

	var loadedCfg config.Config
	if err := yaml.Unmarshal(configData, &loadedCfg); err != nil {
		t.Fatalf("failed to parse config: %v", err)
	}

	if loadedCfg.Plugins == nil {
		t.Fatal("expected plugins config")
	}

	// Verify specified plugins are enabled
	if !loadedCfg.Plugins.CommitParser {
		t.Error("expected commit-parser to be enabled")
	}
	if loadedCfg.Plugins.ChangelogGenerator == nil || !loadedCfg.Plugins.ChangelogGenerator.Enabled {
		t.Error("expected changelog-generator to be enabled")
	}
	if loadedCfg.Plugins.AuditLog == nil || !loadedCfg.Plugins.AuditLog.Enabled {
		t.Error("expected audit-log to be enabled")
	}

	// Verify tag-manager is NOT enabled (not in --enable list)
	if loadedCfg.Plugins.TagManager != nil && loadedCfg.Plugins.TagManager.Enabled {
		t.Error("expected tag-manager to NOT be enabled")
	}
}

func TestCLI_InitCommand_WithForceFlag(t *testing.T) {
	tmp := t.TempDir()
	versionPath := filepath.Join(tmp, ".version")
	configPath := filepath.Join(tmp, ".sley.yaml")

	originalDir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		_ = os.Chdir(originalDir)
	}()

	if err := os.Chdir(tmp); err != nil {
		t.Fatal(err)
	}

	// Create existing config with different content
	existingConfig := []byte("path: .version\nplugins:\n  commit-parser: false\n")
	if err := os.WriteFile(configPath, existingConfig, 0600); err != nil {
		t.Fatal(err)
	}

	cfg := &config.Config{Path: versionPath}
	appCli := testutils.BuildCLIForTests(
		cfg.Path,
		[]*cli.Command{Run()},
	)

	err = appCli.Run(context.Background(), []string{
		"sley", "init", "--yes", "--force",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify config was overwritten
	configData, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatalf("failed to read config: %v", err)
	}

	var loadedCfg config.Config
	if err := yaml.Unmarshal(configData, &loadedCfg); err != nil {
		t.Fatalf("failed to parse config: %v", err)
	}

	// New config should have commit-parser enabled (default)
	if !loadedCfg.Plugins.CommitParser {
		t.Error("expected commit-parser to be enabled after force overwrite")
	}
}

func TestParseEnableFlag(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []string
	}{
		{
			name:     "single plugin",
			input:    "commit-parser",
			expected: []string{"commit-parser"},
		},
		{
			name:     "multiple plugins",
			input:    "commit-parser,tag-manager",
			expected: []string{"commit-parser", "tag-manager"},
		},
		{
			name:     "with spaces",
			input:    "commit-parser, tag-manager, audit-log",
			expected: []string{"commit-parser", "tag-manager", "audit-log"},
		},
		{
			name:     "empty string",
			input:    "",
			expected: []string{},
		},
		{
			name:     "trailing comma",
			input:    "commit-parser,tag-manager,",
			expected: []string{"commit-parser", "tag-manager"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := parseEnableFlag(tt.input)

			if len(got) != len(tt.expected) {
				t.Errorf("expected %d plugins, got %d", len(tt.expected), len(got))
				return
			}

			for i, exp := range tt.expected {
				if got[i] != exp {
					t.Errorf("plugin[%d]: expected %q, got %q", i, exp, got[i])
				}
			}
		})
	}
}

func TestInitializeVersionFile(t *testing.T) {
	tmp := t.TempDir()
	versionPath := filepath.Join(tmp, ".version")

	// First initialization should create file
	created, err := initializeVersionFile(versionPath)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !created {
		t.Error("expected file to be created")
	}

	// Second initialization should not create file
	created, err = initializeVersionFile(versionPath)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if created {
		t.Error("expected file to already exist")
	}
}

func TestDeterminePlugins(t *testing.T) {
	tests := []struct {
		name       string
		ctx        *ProjectContext
		yesFlag    bool
		enableFlag string
		expected   []string
	}{
		{
			name:       "enable flag takes priority",
			ctx:        &ProjectContext{},
			yesFlag:    true,
			enableFlag: "audit-log",
			expected:   []string{"audit-log"},
		},
		{
			name:       "yes flag uses defaults",
			ctx:        &ProjectContext{},
			yesFlag:    true,
			enableFlag: "",
			expected:   DefaultPluginNames(),
		},
		{
			name:       "multiple plugins in enable flag",
			ctx:        &ProjectContext{},
			yesFlag:    false,
			enableFlag: "commit-parser,tag-manager,audit-log",
			expected:   []string{"commit-parser", "tag-manager", "audit-log"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := determinePlugins(tt.ctx, tt.yesFlag, tt.enableFlag)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if len(got) != len(tt.expected) {
				t.Errorf("expected %d plugins, got %d", len(tt.expected), len(got))
				return
			}

			for i, exp := range tt.expected {
				if got[i] != exp {
					t.Errorf("plugin[%d]: expected %q, got %q", i, exp, got[i])
				}
			}
		})
	}
}

func TestPluralize(t *testing.T) {
	tests := []struct {
		count    int
		expected string
	}{
		{0, "s"},
		{1, ""},
		{2, "s"},
		{100, "s"},
	}

	for _, tt := range tests {
		t.Run(fmt.Sprintf("count=%d", tt.count), func(t *testing.T) {
			got := pluralize(tt.count)
			if got != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, got)
			}
		})
	}
}
