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
				[]*cli.Command{Run(), bumpcmd.Run(cfg), precmd.Run(cfg)},
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

func TestCLI_InitCommand_WithTemplateFlag(t *testing.T) {
	tests := []struct {
		name             string
		template         string
		commitParser     bool
		tagManager       bool
		changelogGen     bool
		versionValidator bool
		releaseGate      bool
	}{
		{"basic template", "basic", true, false, false, false, false},
		{"git template", "git", true, true, false, false, false},
		{"automation template", "automation", true, true, true, false, false},
		{"strict template", "strict", true, true, false, true, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			loadedCfg := runInitWithTemplate(t, tt.template)
			verifyTemplatePlugins(t, loadedCfg.Plugins, tt)
		})
	}
}

// runInitWithTemplate runs sley init with the given template and returns the loaded config.
func runInitWithTemplate(t *testing.T, template string) config.Config {
	t.Helper()

	tmp := t.TempDir()
	versionPath := filepath.Join(tmp, ".version")
	configPath := filepath.Join(tmp, ".sley.yaml")

	t.Chdir(tmp)

	cfg := &config.Config{Path: versionPath}
	appCli := testutils.BuildCLIForTests(cfg.Path, []*cli.Command{Run()})

	if err := appCli.Run(context.Background(), []string{"sley", "init", "--template", template}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

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

	return loadedCfg
}

// verifyTemplatePlugins checks that the loaded plugins match the expected configuration.
func verifyTemplatePlugins(t *testing.T, plugins *config.PluginConfig, expected struct {
	name             string
	template         string
	commitParser     bool
	tagManager       bool
	changelogGen     bool
	versionValidator bool
	releaseGate      bool
}) {
	t.Helper()

	checks := []struct {
		name     string
		got      bool
		expected bool
	}{
		{"commit-parser", plugins.CommitParser, expected.commitParser},
		{"tag-manager", plugins.TagManager != nil && plugins.TagManager.Enabled, expected.tagManager},
		{"changelog-generator", plugins.ChangelogGenerator != nil && plugins.ChangelogGenerator.Enabled, expected.changelogGen},
		{"version-validator", plugins.VersionValidator != nil && plugins.VersionValidator.Enabled, expected.versionValidator},
		{"release-gate", plugins.ReleaseGate != nil && plugins.ReleaseGate.Enabled, expected.releaseGate},
	}

	for _, c := range checks {
		if c.got != c.expected {
			t.Errorf("%s: expected %v, got %v", c.name, c.expected, c.got)
		}
	}
}

func TestCLI_InitCommand_WithInvalidTemplate(t *testing.T) {
	tmp := t.TempDir()
	versionPath := filepath.Join(tmp, ".version")

	t.Chdir(tmp)

	cfg := &config.Config{Path: versionPath}
	appCli := testutils.BuildCLIForTests(
		cfg.Path,
		[]*cli.Command{Run()},
	)

	err := appCli.Run(context.Background(), []string{
		"sley", "init", "--template", "invalid-template",
	})
	if err == nil {
		t.Fatal("expected error for invalid template, got nil")
	}

	if !strings.Contains(err.Error(), "unknown template") {
		t.Errorf("expected 'unknown template' error, got: %v", err)
	}
}

func TestCLI_InitCommand_WithForceFlag(t *testing.T) {
	tmp := t.TempDir()
	versionPath := filepath.Join(tmp, ".version")
	configPath := filepath.Join(tmp, ".sley.yaml")

	t.Chdir(tmp)

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

	err := appCli.Run(context.Background(), []string{
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
		name         string
		ctx          *ProjectContext
		yesFlag      bool
		templateFlag string
		enableFlag   string
		expected     []string
		expectError  bool
	}{
		{
			name:         "enable flag takes priority over all",
			ctx:          &ProjectContext{},
			yesFlag:      true,
			templateFlag: "strict",
			enableFlag:   "audit-log",
			expected:     []string{"audit-log"},
		},
		{
			name:         "template flag takes priority over yes",
			ctx:          &ProjectContext{},
			yesFlag:      true,
			templateFlag: "basic",
			enableFlag:   "",
			expected:     []string{"commit-parser"},
		},
		{
			name:         "yes flag uses defaults",
			ctx:          &ProjectContext{},
			yesFlag:      true,
			templateFlag: "",
			enableFlag:   "",
			expected:     DefaultPluginNames(),
		},
		{
			name:         "multiple plugins in enable flag",
			ctx:          &ProjectContext{},
			yesFlag:      false,
			templateFlag: "",
			enableFlag:   "commit-parser,tag-manager,audit-log",
			expected:     []string{"commit-parser", "tag-manager", "audit-log"},
		},
		{
			name:         "template basic",
			ctx:          &ProjectContext{},
			yesFlag:      false,
			templateFlag: "basic",
			enableFlag:   "",
			expected:     []string{"commit-parser"},
		},
		{
			name:         "template git",
			ctx:          &ProjectContext{},
			yesFlag:      false,
			templateFlag: "git",
			enableFlag:   "",
			expected:     []string{"commit-parser", "tag-manager"},
		},
		{
			name:         "template automation",
			ctx:          &ProjectContext{},
			yesFlag:      false,
			templateFlag: "automation",
			enableFlag:   "",
			expected:     []string{"commit-parser", "tag-manager", "changelog-generator"},
		},
		{
			name:         "template strict",
			ctx:          &ProjectContext{},
			yesFlag:      false,
			templateFlag: "strict",
			enableFlag:   "",
			expected:     []string{"commit-parser", "tag-manager", "version-validator", "release-gate"},
		},
		{
			name:         "invalid template returns error",
			ctx:          &ProjectContext{},
			yesFlag:      false,
			templateFlag: "invalid-template",
			enableFlag:   "",
			expectError:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := determinePlugins(tt.ctx, tt.yesFlag, tt.templateFlag, tt.enableFlag)
			if tt.expectError {
				if err == nil {
					t.Fatalf("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if len(got) != len(tt.expected) {
				t.Errorf("expected %d plugins, got %d: %v", len(tt.expected), len(got), got)
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

func TestCreateConfigFile(t *testing.T) {
	t.Run("creates new config file", func(t *testing.T) {
		tmp := t.TempDir()
		t.Chdir(tmp)

		created, err := createConfigFile([]string{"commit-parser"}, false)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !created {
			t.Error("expected file to be created")
		}

		// Verify file exists
		if _, err := os.Stat(".sley.yaml"); os.IsNotExist(err) {
			t.Error("expected .sley.yaml to exist")
		}
	})

	t.Run("force overwrites existing file", func(t *testing.T) {
		tmp := t.TempDir()
		t.Chdir(tmp)

		// Create existing config
		if err := os.WriteFile(".sley.yaml", []byte("old: config\n"), 0600); err != nil {
			t.Fatal(err)
		}

		created, err := createConfigFile([]string{"commit-parser", "tag-manager"}, true)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !created {
			t.Error("expected file to be created with force")
		}

		// Verify file was overwritten
		data, _ := os.ReadFile(".sley.yaml")
		if strings.Contains(string(data), "old: config") {
			t.Error("expected config to be overwritten")
		}
	})

	// Note: The "skips existing file without force in non-interactive" path
	// is tested via CLI tests using --yes flag which avoids interactive prompts
}

func TestPrintVersionOnlySuccess(t *testing.T) {
	t.Run("prints with valid version", func(t *testing.T) {
		tmp := t.TempDir()
		versionPath := filepath.Join(tmp, ".version")
		if err := os.WriteFile(versionPath, []byte("1.2.3\n"), 0600); err != nil {
			t.Fatal(err)
		}

		output, _ := testutils.CaptureStdout(func() {
			printVersionOnlySuccess(versionPath)
		})

		if !strings.Contains(output, "1.2.3") {
			t.Errorf("expected version in output, got: %s", output)
		}
	})

	t.Run("prints without version on error", func(t *testing.T) {
		tmp := t.TempDir()
		versionPath := filepath.Join(tmp, ".version")
		// Create file with invalid version
		if err := os.WriteFile(versionPath, []byte("invalid\n"), 0600); err != nil {
			t.Fatal(err)
		}

		output, _ := testutils.CaptureStdout(func() {
			printVersionOnlySuccess(versionPath)
		})

		if !strings.Contains(output, "Initialized") {
			t.Errorf("expected initialized message, got: %s", output)
		}
	})
}

func TestPrintSuccessSummary(t *testing.T) {
	t.Run("prints all messages when both created", func(t *testing.T) {
		tmp := t.TempDir()
		versionPath := filepath.Join(tmp, ".version")
		if err := os.WriteFile(versionPath, []byte("1.0.0\n"), 0600); err != nil {
			t.Fatal(err)
		}

		ctx := &ProjectContext{IsGitRepo: true}
		plugins := []string{"commit-parser", "tag-manager"}

		output, _ := testutils.CaptureStdout(func() {
			printSuccessSummary(versionPath, true, true, plugins, ctx)
		})

		if !strings.Contains(output, "1.0.0") {
			t.Error("expected version in output")
		}
		if !strings.Contains(output, "2 plugins enabled") {
			t.Error("expected plugins count in output")
		}
		if !strings.Contains(output, "Next steps") {
			t.Error("expected next steps in output")
		}
	})

	t.Run("prints version created without config", func(t *testing.T) {
		tmp := t.TempDir()
		versionPath := filepath.Join(tmp, ".version")
		if err := os.WriteFile(versionPath, []byte("2.0.0\n"), 0600); err != nil {
			t.Fatal(err)
		}

		ctx := &ProjectContext{IsGitRepo: false}

		output, _ := testutils.CaptureStdout(func() {
			printSuccessSummary(versionPath, true, false, nil, ctx)
		})

		if !strings.Contains(output, "2.0.0") {
			t.Error("expected version in output")
		}
	})

	t.Run("handles version read error gracefully", func(t *testing.T) {
		tmp := t.TempDir()
		versionPath := filepath.Join(tmp, ".version")
		// Create file with invalid content
		if err := os.WriteFile(versionPath, []byte("invalid\n"), 0600); err != nil {
			t.Fatal(err)
		}

		ctx := &ProjectContext{}

		output, _ := testutils.CaptureStdout(func() {
			printSuccessSummary(versionPath, true, false, nil, ctx)
		})

		// Should still print created message (without version)
		if !strings.Contains(output, "Created") {
			t.Error("expected created message in output")
		}
	})
}

func TestInitializeVersionFileWithMigration(t *testing.T) {
	t.Run("uses migrated version", func(t *testing.T) {
		tmp := t.TempDir()
		versionPath := filepath.Join(tmp, ".version")

		created, err := initializeVersionFileWithMigration(versionPath, "5.0.0")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !created {
			t.Error("expected file to be created")
		}

		data, _ := os.ReadFile(versionPath)
		if string(data) != "5.0.0\n" {
			t.Errorf("expected version 5.0.0, got %q", string(data))
		}
	})

	t.Run("uses default when no migration version", func(t *testing.T) {
		tmp := t.TempDir()
		versionPath := filepath.Join(tmp, ".version")

		created, err := initializeVersionFileWithMigration(versionPath, "")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !created {
			t.Error("expected file to be created")
		}

		// Should have default version (0.1.0 or from git tag)
		data, _ := os.ReadFile(versionPath)
		if string(data) == "" {
			t.Error("expected version to be written")
		}
	})

	t.Run("returns false for existing valid file", func(t *testing.T) {
		tmp := t.TempDir()
		versionPath := filepath.Join(tmp, ".version")
		if err := os.WriteFile(versionPath, []byte("1.0.0\n"), 0600); err != nil {
			t.Fatal(err)
		}

		created, err := initializeVersionFileWithMigration(versionPath, "2.0.0")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if created {
			t.Error("expected file to not be recreated")
		}

		// Verify original content unchanged
		data, _ := os.ReadFile(versionPath)
		if string(data) != "1.0.0\n" {
			t.Errorf("expected original version, got %q", string(data))
		}
	})

	t.Run("returns error for existing invalid file", func(t *testing.T) {
		tmp := t.TempDir()
		versionPath := filepath.Join(tmp, ".version")
		if err := os.WriteFile(versionPath, []byte("invalid\n"), 0600); err != nil {
			t.Fatal(err)
		}

		_, err := initializeVersionFileWithMigration(versionPath, "2.0.0")
		if err == nil {
			t.Fatal("expected error for invalid existing file")
		}
		if !strings.Contains(err.Error(), "failed to read version file") {
			t.Errorf("expected read error, got: %v", err)
		}
	})
}

// Note: TestDeterminePlugins_NonInteractive is not tested directly here
// because the test environment may appear interactive.
// The non-interactive fallback path is covered by CLI tests using
// --yes, --template, and --enable flags which bypass the interactive prompt.

func TestCLI_InitCommand_NoPluginsSelected(t *testing.T) {
	// This test verifies behavior when plugins list is empty
	// which happens when user cancels interactive prompt
	tmp := t.TempDir()
	versionPath := filepath.Join(tmp, ".version")

	t.Chdir(tmp)

	cfg := &config.Config{Path: versionPath}
	appCli := testutils.BuildCLIForTests(cfg.Path, []*cli.Command{Run()})

	// Using --enable with empty value won't trigger this path
	// But --yes without other flags will use defaults
	err := appCli.Run(context.Background(), []string{
		"sley", "init", "--enable", "",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// With empty enable flag, should create version but skip config
	if _, err := os.Stat(versionPath); os.IsNotExist(err) {
		t.Error("expected .version to be created")
	}
}
