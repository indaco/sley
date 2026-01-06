package initcmd

import (
	"context"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/goccy/go-yaml"
	"github.com/indaco/sley/internal/config"
	"github.com/indaco/sley/internal/testutils"
	"github.com/urfave/cli/v3"
)

func TestDiscoverVersionFiles(t *testing.T) {
	tmp := t.TempDir()

	// Create a monorepo structure
	dirs := []struct {
		path    string
		version string
	}{
		{"services/api", "1.0.0"},
		{"services/auth", "2.0.0"},
		{"packages/shared", "0.5.0"},
		{"apps/web", "3.1.0"},
	}

	for _, d := range dirs {
		dirPath := filepath.Join(tmp, d.path)
		if err := os.MkdirAll(dirPath, 0755); err != nil {
			t.Fatalf("failed to create dir %s: %v", dirPath, err)
		}
		versionFile := filepath.Join(dirPath, ".version")
		if err := os.WriteFile(versionFile, []byte(d.version+"\n"), 0600); err != nil {
			t.Fatalf("failed to write version file: %v", err)
		}
	}

	// Create directories that should be excluded
	excludedDirs := []string{"node_modules/pkg", "vendor/lib", ".git/objects"}
	for _, d := range excludedDirs {
		dirPath := filepath.Join(tmp, d)
		if err := os.MkdirAll(dirPath, 0755); err != nil {
			t.Fatalf("failed to create excluded dir: %v", err)
		}
		versionFile := filepath.Join(dirPath, ".version")
		if err := os.WriteFile(versionFile, []byte("0.0.1\n"), 0600); err != nil {
			t.Fatalf("failed to write excluded version file: %v", err)
		}
	}

	// Change to temp directory for discovery
	t.Chdir(tmp)

	modules, err := discoverVersionFiles(".")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Should find exactly 4 modules (not the excluded ones)
	if len(modules) != 4 {
		t.Errorf("expected 4 modules, got %d", len(modules))
		for _, m := range modules {
			t.Logf("  found: %s at %s", m.Name, m.RelPath)
		}
	}

	// Verify module names
	expectedNames := map[string]string{
		"api":    "1.0.0",
		"auth":   "2.0.0",
		"shared": "0.5.0",
		"web":    "3.1.0",
	}

	for _, mod := range modules {
		expectedVersion, ok := expectedNames[mod.Name]
		if !ok {
			t.Errorf("unexpected module: %s", mod.Name)
			continue
		}
		if mod.Version != expectedVersion {
			t.Errorf("module %s: expected version %q, got %q", mod.Name, expectedVersion, mod.Version)
		}
	}
}

func TestDiscoverVersionFiles_Empty(t *testing.T) {
	tmp := t.TempDir()

	t.Chdir(tmp)

	modules, err := discoverVersionFiles(".")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(modules) != 0 {
		t.Errorf("expected 0 modules, got %d", len(modules))
	}
}

func TestDiscoverVersionFiles_SkipsRootVersion(t *testing.T) {
	tmp := t.TempDir()

	// Create root .version file (should be skipped)
	if err := os.WriteFile(filepath.Join(tmp, ".version"), []byte("1.0.0\n"), 0600); err != nil {
		t.Fatal(err)
	}

	// Create subdirectory with .version
	subDir := filepath.Join(tmp, "submodule")
	if err := os.MkdirAll(subDir, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(subDir, ".version"), []byte("2.0.0\n"), 0600); err != nil {
		t.Fatal(err)
	}

	t.Chdir(tmp)

	modules, err := discoverVersionFiles(".")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Should only find the submodule, not the root
	if len(modules) != 1 {
		t.Errorf("expected 1 module, got %d", len(modules))
	}

	if len(modules) > 0 && modules[0].Name != "submodule" {
		t.Errorf("expected module name 'submodule', got %q", modules[0].Name)
	}
}

func TestGenerateWorkspaceConfigWithComments(t *testing.T) {
	modules := []DiscoveredModule{
		{Name: "api", RelPath: "services/api/.version", Version: "1.0.0"},
		{Name: "web", RelPath: "apps/web/.version", Version: "2.0.0"},
	}
	plugins := []string{"commit-parser", "tag-manager"}

	data, err := GenerateWorkspaceConfigWithComments(plugins, modules)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	dataStr := string(data)

	// Verify header
	if !strings.Contains(dataStr, "sley configuration file") {
		t.Error("expected header comment")
	}

	// Verify workspace section
	if !strings.Contains(dataStr, "workspace:") {
		t.Error("expected workspace section")
	}

	// Verify discovery settings
	if !strings.Contains(dataStr, "discovery:") {
		t.Error("expected discovery section")
	}
	if !strings.Contains(dataStr, "enabled: true") {
		t.Error("expected discovery enabled")
	}

	// Verify discovered modules are commented
	if !strings.Contains(dataStr, "# Discovered modules") {
		t.Error("expected discovered modules comment")
	}
	if !strings.Contains(dataStr, "#   - name: api") {
		t.Error("expected api module in comments")
	}
	if !strings.Contains(dataStr, "#   - name: web") {
		t.Error("expected web module in comments")
	}

	// Verify it's valid YAML (excluding comments)
	var cfg config.Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		t.Fatalf("generated config is not valid YAML: %v", err)
	}

	if cfg.Workspace == nil {
		t.Error("expected workspace config")
	}
	if cfg.Workspace.Discovery == nil {
		t.Error("expected discovery config")
	}
}

func TestGenerateWorkspaceConfigWithComments_NoModules(t *testing.T) {
	modules := []DiscoveredModule{}
	plugins := []string{"commit-parser"}

	data, err := GenerateWorkspaceConfigWithComments(plugins, modules)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	dataStr := string(data)

	// Should still have workspace section
	if !strings.Contains(dataStr, "workspace:") {
		t.Error("expected workspace section")
	}

	// Should not have discovered modules section
	if strings.Contains(dataStr, "# Discovered modules") {
		t.Error("should not have discovered modules comment when no modules")
	}
}

func TestCLI_InitCommand_WithWorkspaceFlag(t *testing.T) {
	tmp := t.TempDir()
	versionPath := filepath.Join(tmp, ".version")
	configPath := filepath.Join(tmp, ".sley.yaml")

	// Create some module directories with .version files
	modules := []struct {
		dir     string
		version string
	}{
		{"services/api", "1.0.0"},
		{"services/auth", "2.0.0"},
	}

	for _, m := range modules {
		dir := filepath.Join(tmp, m.dir)
		if err := os.MkdirAll(dir, 0755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(dir, ".version"), []byte(m.version+"\n"), 0600); err != nil {
			t.Fatal(err)
		}
	}

	t.Chdir(tmp)

	cfg := &config.Config{Path: versionPath}
	appCli := testutils.BuildCLIForTests(cfg.Path, []*cli.Command{Run()})

	err := appCli.Run(context.Background(), []string{
		"sley", "init", "--workspace", "--yes",
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

	// Verify workspace section exists
	if loadedCfg.Workspace == nil {
		t.Fatal("expected workspace config")
	}

	if loadedCfg.Workspace.Discovery == nil {
		t.Fatal("expected discovery config")
	}

	// Verify discovery is enabled
	if loadedCfg.Workspace.Discovery.Enabled != nil && !*loadedCfg.Workspace.Discovery.Enabled {
		t.Error("expected discovery to be enabled")
	}
}

func TestCLI_InitCommand_WorkspaceWithTemplate(t *testing.T) {
	tmp := t.TempDir()
	versionPath := filepath.Join(tmp, ".version")
	configPath := filepath.Join(tmp, ".sley.yaml")

	t.Chdir(tmp)

	cfg := &config.Config{Path: versionPath}
	appCli := testutils.BuildCLIForTests(cfg.Path, []*cli.Command{Run()})

	err := appCli.Run(context.Background(), []string{
		"sley", "init", "--workspace", "--template", "automation",
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

	// Verify plugins from automation template
	if !loadedCfg.Plugins.CommitParser {
		t.Error("expected commit-parser enabled")
	}
	if loadedCfg.Plugins.TagManager == nil || !loadedCfg.Plugins.TagManager.Enabled {
		t.Error("expected tag-manager enabled")
	}
	if loadedCfg.Plugins.ChangelogGenerator == nil || !loadedCfg.Plugins.ChangelogGenerator.Enabled {
		t.Error("expected changelog-generator enabled")
	}

	// Verify workspace section
	if loadedCfg.Workspace == nil {
		t.Error("expected workspace config")
	}
}

func TestCreateWorkspaceConfigFile(t *testing.T) {
	t.Run("creates new workspace config", func(t *testing.T) {
		tmp := t.TempDir()
		t.Chdir(tmp)

		modules := []DiscoveredModule{
			{Name: "api", RelPath: "services/api/.version", Version: "1.0.0"},
		}

		created, err := createWorkspaceConfigFile([]string{"commit-parser"}, modules, false)
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

	t.Run("force overwrites existing config", func(t *testing.T) {
		tmp := t.TempDir()
		t.Chdir(tmp)

		// Create existing config
		if err := os.WriteFile(".sley.yaml", []byte("old: config\n"), 0600); err != nil {
			t.Fatal(err)
		}

		modules := []DiscoveredModule{}
		created, err := createWorkspaceConfigFile([]string{"commit-parser"}, modules, true)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !created {
			t.Error("expected file to be created with force")
		}

		// Verify file was overwritten with workspace config
		data, _ := os.ReadFile(".sley.yaml")
		if strings.Contains(string(data), "old: config") {
			t.Error("expected config to be overwritten")
		}
		if !strings.Contains(string(data), "workspace:") {
			t.Error("expected workspace section in config")
		}
	})

	// Note: The "skips existing config without force" path is tested
	// via CLI tests using --yes flag which avoids interactive prompts
}

func TestPrintWorkspaceSuccessSummary(t *testing.T) {
	t.Run("prints with discovered modules", func(t *testing.T) {
		modules := []DiscoveredModule{
			{Name: "api", RelPath: "services/api/.version", Version: "1.0.0"},
			{Name: "web", RelPath: "apps/web/.version", Version: "2.0.0"},
		}
		plugins := []string{"commit-parser", "tag-manager"}
		ctx := &ProjectContext{IsGitRepo: true}

		output, _ := captureStdout(func() {
			printWorkspaceSuccessSummary(true, plugins, modules, ctx)
		})

		if !strings.Contains(output, "2 plugins") {
			t.Error("expected plugins count in output")
		}
		if !strings.Contains(output, "Discovered 2 module") {
			t.Error("expected modules count in output")
		}
		if !strings.Contains(output, "api") {
			t.Error("expected api module in output")
		}
		if !strings.Contains(output, "web") {
			t.Error("expected web module in output")
		}
	})

	t.Run("prints without modules", func(t *testing.T) {
		modules := []DiscoveredModule{}
		plugins := []string{"commit-parser"}
		ctx := &ProjectContext{}

		output, _ := captureStdout(func() {
			printWorkspaceSuccessSummary(true, plugins, modules, ctx)
		})

		if !strings.Contains(output, "No existing .version files found") {
			t.Error("expected no modules message")
		}
		if !strings.Contains(output, "Create .version files") {
			t.Error("expected instruction to create version files")
		}
	})

	t.Run("prints module with unknown version", func(t *testing.T) {
		modules := []DiscoveredModule{
			{Name: "api", RelPath: "services/api/.version", Version: ""},
		}
		plugins := []string{"commit-parser"}
		ctx := &ProjectContext{}

		output, _ := captureStdout(func() {
			printWorkspaceSuccessSummary(true, plugins, modules, ctx)
		})

		if !strings.Contains(output, "unknown") {
			t.Error("expected 'unknown' version in output")
		}
	})

	t.Run("skips config message when not created", func(t *testing.T) {
		modules := []DiscoveredModule{}
		plugins := []string{"commit-parser"}
		ctx := &ProjectContext{}

		output, _ := captureStdout(func() {
			printWorkspaceSuccessSummary(false, plugins, modules, ctx)
		})

		if strings.Contains(output, "Created .sley.yaml") {
			t.Error("should not print created message when configCreated is false")
		}
	})
}

func TestWritePluginConfig(t *testing.T) {
	tests := []struct {
		name     string
		plugin   string
		expected string
	}{
		{"commit-parser uses simple format", "commit-parser", "commit-parser: true"},
		{"tag-manager uses enabled format", "tag-manager", "tag-manager:\n    enabled: true"},
		{"unknown plugin uses enabled format", "custom-plugin", "custom-plugin:\n    enabled: true"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var sb strings.Builder
			writePluginConfig(&sb, tt.plugin)
			result := sb.String()

			if !strings.Contains(result, tt.expected) {
				t.Errorf("expected %q in output, got: %s", tt.expected, result)
			}
		})
	}
}

func TestCLI_InitCommand_WorkspaceWithEnableFlag(t *testing.T) {
	tmp := t.TempDir()
	versionPath := filepath.Join(tmp, ".version")
	configPath := filepath.Join(tmp, ".sley.yaml")

	t.Chdir(tmp)

	cfg := &config.Config{Path: versionPath}
	appCli := testutils.BuildCLIForTests(cfg.Path, []*cli.Command{Run()})

	err := appCli.Run(context.Background(), []string{
		"sley", "init", "--workspace", "--enable", "commit-parser,audit-log",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify config was created with specified plugins
	configData, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatalf("failed to read config: %v", err)
	}

	var loadedCfg config.Config
	if err := yaml.Unmarshal(configData, &loadedCfg); err != nil {
		t.Fatalf("failed to parse config: %v", err)
	}

	if !loadedCfg.Plugins.CommitParser {
		t.Error("expected commit-parser enabled")
	}
	if loadedCfg.Plugins.AuditLog == nil || !loadedCfg.Plugins.AuditLog.Enabled {
		t.Error("expected audit-log enabled")
	}
	// tag-manager should NOT be enabled
	if loadedCfg.Plugins.TagManager != nil && loadedCfg.Plugins.TagManager.Enabled {
		t.Error("expected tag-manager NOT enabled")
	}
}

func TestCLI_InitCommand_WorkspaceWithForce(t *testing.T) {
	tmp := t.TempDir()
	versionPath := filepath.Join(tmp, ".version")
	configPath := filepath.Join(tmp, ".sley.yaml")

	// Create existing config
	if err := os.WriteFile(configPath, []byte("old: config\n"), 0600); err != nil {
		t.Fatal(err)
	}

	t.Chdir(tmp)

	cfg := &config.Config{Path: versionPath}
	appCli := testutils.BuildCLIForTests(cfg.Path, []*cli.Command{Run()})

	err := appCli.Run(context.Background(), []string{
		"sley", "init", "--workspace", "--yes", "--force",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify config was overwritten
	configData, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatalf("failed to read config: %v", err)
	}

	if strings.Contains(string(configData), "old: config") {
		t.Error("expected config to be overwritten")
	}
	if !strings.Contains(string(configData), "workspace:") {
		t.Error("expected workspace section in config")
	}
}

// captureStdout captures stdout during function execution
func captureStdout(f func()) (string, error) {
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	f()

	w.Close()
	os.Stdout = old

	out, err := io.ReadAll(r)
	if err != nil {
		return "", err
	}
	return string(out), nil
}
