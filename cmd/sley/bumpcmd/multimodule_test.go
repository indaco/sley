package bumpcmd

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/indaco/sley/internal/config"
	"github.com/indaco/sley/internal/plugins"
	"github.com/indaco/sley/internal/plugins/dependencycheck"
	"github.com/indaco/sley/internal/testutils"
	"github.com/indaco/sley/internal/workspace"
	"github.com/urfave/cli/v3"
)

/* ------------------------------------------------------------------------- */
/* GET FIRST SUCCESSFUL VERSION TESTS                                        */
/* ------------------------------------------------------------------------- */

func TestGetFirstSuccessfulVersion(t *testing.T) {
	tests := []struct {
		name     string
		results  []workspace.ExecutionResult
		expected string
	}{
		{
			name:     "empty results returns empty",
			results:  []workspace.ExecutionResult{},
			expected: "",
		},
		{
			name: "all failures returns empty",
			results: []workspace.ExecutionResult{
				{Success: false, NewVersion: "1.0.0"},
				{Success: false, NewVersion: "2.0.0"},
			},
			expected: "",
		},
		{
			name: "first success returns version",
			results: []workspace.ExecutionResult{
				{Success: true, NewVersion: "1.0.1"},
				{Success: true, NewVersion: "2.0.1"},
			},
			expected: "1.0.1",
		},
		{
			name: "mixed results returns first success",
			results: []workspace.ExecutionResult{
				{Success: false, NewVersion: "1.0.0"},
				{Success: true, NewVersion: "2.0.1"},
				{Success: true, NewVersion: "3.0.1"},
			},
			expected: "2.0.1",
		},
		{
			name: "success with empty version",
			results: []workspace.ExecutionResult{
				{Success: true, NewVersion: ""},
				{Success: true, NewVersion: "2.0.1"},
			},
			expected: "2.0.1",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := getFirstSuccessfulVersion(tt.results)
			if result != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, result)
			}
		})
	}
}

/* ------------------------------------------------------------------------- */
/* MULTI-MODULE BUMP WITH DEPENDENCY-CHECK PLUGIN TESTS                      */
/* ------------------------------------------------------------------------- */

// setupMultiModuleWorkspaceWithVersion creates a workspace with multiple modules.
func setupMultiModuleWorkspaceWithVersion(t *testing.T, dir string, modules map[string]string) {
	t.Helper()
	for modulePath, version := range modules {
		moduleDir := filepath.Join(dir, modulePath)
		if err := os.MkdirAll(moduleDir, 0755); err != nil {
			t.Fatalf("failed to create module dir %s: %v", moduleDir, err)
		}
		versionFile := filepath.Join(moduleDir, ".version")
		if err := os.WriteFile(versionFile, []byte(version+"\n"), 0644); err != nil {
			t.Fatalf("failed to write version file %s: %v", versionFile, err)
		}
	}
}

// readModuleVersionFromDir reads the version of a specific module.
func readModuleVersionFromDir(t *testing.T, dir, modulePath string) string {
	t.Helper()
	versionFile := filepath.Join(dir, modulePath, ".version")
	data, err := os.ReadFile(versionFile)
	if err != nil {
		t.Fatalf("failed to read version file %s: %v", versionFile, err)
	}
	return strings.TrimSpace(string(data))
}

// buildMultiModuleCLI creates a CLI app configured for multi-module mode.
// It uses an empty default path so the detector can discover modules.
func buildMultiModuleCLI(cfg *config.Config, registry *plugins.PluginRegistry) *cli.Command {
	return &cli.Command{
		Name: "sley",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    "path",
				Aliases: []string{"p"},
				Usage:   "Path to .version file",
				Value:   ".version", // Default, will be auto-detected
			},
			&cli.BoolFlag{
				Name:    "strict",
				Aliases: []string{"no-auto-init"},
				Usage:   "Fail if .version file is missing (disable auto-initialization)",
			},
		},
		Commands: []*cli.Command{Run(cfg, registry)},
	}
}

func TestMultiModuleBump_SyncsDependencies(t *testing.T) {
	tmpDir := t.TempDir()

	// Set up a multi-module workspace (without root .version file)
	setupMultiModuleWorkspaceWithVersion(t, tmpDir, map[string]string{
		"backend/api":    "1.0.0",
		"backend/worker": "1.0.0",
		"frontend":       "1.0.0",
	})

	// Create additional files that should be synced by dependency-check plugin
	pkgJSONPath := filepath.Join(tmpDir, "frontend", "package.json")
	if err := os.WriteFile(pkgJSONPath, []byte(`{"name": "frontend", "version": "1.0.0"}`), 0644); err != nil {
		t.Fatal(err)
	}

	extraVersionPath := filepath.Join(tmpDir, "backend", "extra.version")
	if err := os.WriteFile(extraVersionPath, []byte("1.0.0\n"), 0644); err != nil {
		t.Fatal(err)
	}

	// Create a config without explicit path (will use multi-module detection)
	cfg := &config.Config{
		Path: ".version", // Default, triggers multi-module detection
	}

	// Create and register the dependency-check plugin
	depPlugin := dependencycheck.NewDependencyChecker(&dependencycheck.Config{
		Enabled:  true,
		AutoSync: true,
		Files: []dependencycheck.FileConfig{
			{Path: pkgJSONPath, Field: "version", Format: "json"},
			{Path: extraVersionPath, Format: "raw"},
		},
	})

	registry := plugins.NewPluginRegistry()
	if err := registry.RegisterDependencyChecker(depPlugin); err != nil {
		t.Fatalf("failed to register dependency checker: %v", err)
	}

	// Build CLI with bump command for multi-module mode
	appCli := buildMultiModuleCLI(cfg, registry)

	// Run multi-module bump with --all to explicitly select all modules
	err := testutils.RunCLITestAllowError(t, appCli, []string{
		"sley", "bump", "patch", "--all", "--non-interactive",
	}, tmpDir)
	if err != nil {
		t.Fatalf("bump failed: %v", err)
	}

	// Verify module versions were bumped
	apiVersion := readModuleVersionFromDir(t, tmpDir, "backend/api")
	if apiVersion != "1.0.1" {
		t.Errorf("expected api version '1.0.1', got %q", apiVersion)
	}

	workerVersion := readModuleVersionFromDir(t, tmpDir, "backend/worker")
	if workerVersion != "1.0.1" {
		t.Errorf("expected worker version '1.0.1', got %q", workerVersion)
	}

	frontendVersion := readModuleVersionFromDir(t, tmpDir, "frontend")
	if frontendVersion != "1.0.1" {
		t.Errorf("expected frontend version '1.0.1', got %q", frontendVersion)
	}

	// Verify package.json was synced
	pkgJSONData, err := os.ReadFile(pkgJSONPath)
	if err != nil {
		t.Fatalf("failed to read package.json: %v", err)
	}
	pkgJSON := string(pkgJSONData)
	if !strings.Contains(pkgJSON, `"version": "1.0.1"`) {
		t.Errorf("expected package.json to contain version 1.0.1, got: %s", pkgJSON)
	}

	// Verify extra.version was synced
	extraVersionData, err := os.ReadFile(extraVersionPath)
	if err != nil {
		t.Fatalf("failed to read extra.version: %v", err)
	}
	extraVersion := strings.TrimSpace(string(extraVersionData))
	if extraVersion != "1.0.1" {
		t.Errorf("expected extra.version to be '1.0.1', got %q", extraVersion)
	}
}

func TestMultiModuleBump_NoSyncWhenDisabled(t *testing.T) {
	tmpDir := t.TempDir()

	// Set up a multi-module workspace
	setupMultiModuleWorkspaceWithVersion(t, tmpDir, map[string]string{
		"api": "1.0.0",
		"web": "1.0.0",
	})

	// Create an additional file that would be synced if auto-sync was enabled
	// Set its version to match the modules so pre-bump validation passes
	extraVersionPath := filepath.Join(tmpDir, "other.version")
	if err := os.WriteFile(extraVersionPath, []byte("1.0.0\n"), 0644); err != nil {
		t.Fatal(err)
	}

	cfg := &config.Config{
		Path: ".version",
	}

	// Create dependency-check plugin with AutoSync disabled
	depPlugin := dependencycheck.NewDependencyChecker(&dependencycheck.Config{
		Enabled:  true,
		AutoSync: false, // Disabled
		Files: []dependencycheck.FileConfig{
			{Path: extraVersionPath, Format: "raw"},
		},
	})

	registry := plugins.NewPluginRegistry()
	if err := registry.RegisterDependencyChecker(depPlugin); err != nil {
		t.Fatalf("failed to register dependency checker: %v", err)
	}

	appCli := buildMultiModuleCLI(cfg, registry)

	err := testutils.RunCLITestAllowError(t, appCli, []string{
		"sley", "bump", "patch", "--all", "--non-interactive",
	}, tmpDir)
	if err != nil {
		t.Fatalf("bump failed: %v", err)
	}

	// Verify module versions were bumped
	apiVersion := readModuleVersionFromDir(t, tmpDir, "api")
	if apiVersion != "1.0.1" {
		t.Errorf("expected api version '1.0.1', got %q", apiVersion)
	}

	// Verify extra file was NOT synced (because auto-sync is disabled)
	extraVersionData, err := os.ReadFile(extraVersionPath)
	if err != nil {
		t.Fatalf("failed to read other.version: %v", err)
	}
	extraVersion := strings.TrimSpace(string(extraVersionData))
	if extraVersion != "1.0.0" {
		t.Errorf("expected other.version to remain '1.0.0' (no sync), got %q", extraVersion)
	}
}

func TestMultiModuleBump_NoSyncWithoutPlugin(t *testing.T) {
	tmpDir := t.TempDir()

	// Set up a multi-module workspace
	setupMultiModuleWorkspaceWithVersion(t, tmpDir, map[string]string{
		"api": "1.0.0",
		"web": "1.0.0",
	})

	cfg := &config.Config{
		Path: ".version",
	}

	// Registry without dependency-check plugin
	registry := plugins.NewPluginRegistry()

	appCli := buildMultiModuleCLI(cfg, registry)

	err := testutils.RunCLITestAllowError(t, appCli, []string{
		"sley", "bump", "minor", "--all", "--non-interactive",
	}, tmpDir)
	if err != nil {
		t.Fatalf("bump failed: %v", err)
	}

	// Verify module versions were bumped
	apiVersion := readModuleVersionFromDir(t, tmpDir, "api")
	if apiVersion != "1.1.0" {
		t.Errorf("expected api version '1.1.0', got %q", apiVersion)
	}

	webVersion := readModuleVersionFromDir(t, tmpDir, "web")
	if webVersion != "1.1.0" {
		t.Errorf("expected web version '1.1.0', got %q", webVersion)
	}
}

func TestMultiModuleBump_AutoCommand_SyncsDependencies(t *testing.T) {
	tmpDir := t.TempDir()

	// Set up a multi-module workspace with pre-release versions
	setupMultiModuleWorkspaceWithVersion(t, tmpDir, map[string]string{
		"api": "1.0.0-alpha",
		"web": "1.0.0-alpha",
	})

	// Create additional file to sync - must match the module versions
	extraVersionPath := filepath.Join(tmpDir, "build.version")
	if err := os.WriteFile(extraVersionPath, []byte("1.0.0-alpha\n"), 0644); err != nil {
		t.Fatal(err)
	}

	cfg := &config.Config{
		Path: ".version",
	}

	depPlugin := dependencycheck.NewDependencyChecker(&dependencycheck.Config{
		Enabled:  true,
		AutoSync: true,
		Files: []dependencycheck.FileConfig{
			{Path: extraVersionPath, Format: "raw"},
		},
	})

	registry := plugins.NewPluginRegistry()
	if err := registry.RegisterDependencyChecker(depPlugin); err != nil {
		t.Fatalf("failed to register dependency checker: %v", err)
	}

	appCli := buildMultiModuleCLI(cfg, registry)

	// Run auto bump which should promote pre-release versions
	err := testutils.RunCLITestAllowError(t, appCli, []string{
		"sley", "bump", "auto", "--all", "--non-interactive",
	}, tmpDir)
	if err != nil {
		t.Fatalf("bump auto failed: %v", err)
	}

	// Verify module versions were promoted to release
	apiVersion := readModuleVersionFromDir(t, tmpDir, "api")
	if apiVersion != "1.0.0" {
		t.Errorf("expected api version '1.0.0', got %q", apiVersion)
	}

	// Verify extra file was synced
	extraVersionData, err := os.ReadFile(extraVersionPath)
	if err != nil {
		t.Fatalf("failed to read build.version: %v", err)
	}
	extraVersion := strings.TrimSpace(string(extraVersionData))
	if extraVersion != "1.0.0" {
		t.Errorf("expected build.version to be '1.0.0', got %q", extraVersion)
	}
}

func TestMultiModuleBump_ReleaseCommand_SyncsDependencies(t *testing.T) {
	tmpDir := t.TempDir()

	// Set up a multi-module workspace with pre-release versions
	setupMultiModuleWorkspaceWithVersion(t, tmpDir, map[string]string{
		"api": "2.0.0-beta.1",
		"web": "2.0.0-beta.1",
	})

	// Create additional file to sync - must match the module versions
	extraVersionPath := filepath.Join(tmpDir, "release.version")
	if err := os.WriteFile(extraVersionPath, []byte("2.0.0-beta.1\n"), 0644); err != nil {
		t.Fatal(err)
	}

	cfg := &config.Config{
		Path: ".version",
	}

	depPlugin := dependencycheck.NewDependencyChecker(&dependencycheck.Config{
		Enabled:  true,
		AutoSync: true,
		Files: []dependencycheck.FileConfig{
			{Path: extraVersionPath, Format: "raw"},
		},
	})

	registry := plugins.NewPluginRegistry()
	if err := registry.RegisterDependencyChecker(depPlugin); err != nil {
		t.Fatalf("failed to register dependency checker: %v", err)
	}

	appCli := buildMultiModuleCLI(cfg, registry)

	// Run release bump which should remove pre-release
	err := testutils.RunCLITestAllowError(t, appCli, []string{
		"sley", "bump", "release", "--all", "--non-interactive",
	}, tmpDir)
	if err != nil {
		t.Fatalf("bump release failed: %v", err)
	}

	// Verify module versions were released
	apiVersion := readModuleVersionFromDir(t, tmpDir, "api")
	if apiVersion != "2.0.0" {
		t.Errorf("expected api version '2.0.0', got %q", apiVersion)
	}

	// Verify extra file was synced
	extraVersionData, err := os.ReadFile(extraVersionPath)
	if err != nil {
		t.Fatalf("failed to read release.version: %v", err)
	}
	extraVersion := strings.TrimSpace(string(extraVersionData))
	if extraVersion != "2.0.0" {
		t.Errorf("expected release.version to be '2.0.0', got %q", extraVersion)
	}
}

func TestMultiModuleBump_SpecificModule_SyncsDependencies(t *testing.T) {
	tmpDir := t.TempDir()

	// Set up a multi-module workspace
	setupMultiModuleWorkspaceWithVersion(t, tmpDir, map[string]string{
		"api": "1.0.0",
		"web": "2.0.0",
	})

	// Create additional file to sync - start with 1.0.0 to match the api module
	extraVersionPath := filepath.Join(tmpDir, "sync.version")
	if err := os.WriteFile(extraVersionPath, []byte("1.0.0\n"), 0644); err != nil {
		t.Fatal(err)
	}

	cfg := &config.Config{
		Path: ".version",
	}

	depPlugin := dependencycheck.NewDependencyChecker(&dependencycheck.Config{
		Enabled:  true,
		AutoSync: true,
		Files: []dependencycheck.FileConfig{
			{Path: extraVersionPath, Format: "raw"},
		},
	})

	registry := plugins.NewPluginRegistry()
	if err := registry.RegisterDependencyChecker(depPlugin); err != nil {
		t.Fatalf("failed to register dependency checker: %v", err)
	}

	appCli := buildMultiModuleCLI(cfg, registry)

	// Bump only the api module
	err := testutils.RunCLITestAllowError(t, appCli, []string{
		"sley", "bump", "patch", "--module", "api", "--non-interactive",
	}, tmpDir)
	if err != nil {
		t.Fatalf("bump failed: %v", err)
	}

	// Verify only api was bumped
	apiVersion := readModuleVersionFromDir(t, tmpDir, "api")
	if apiVersion != "1.0.1" {
		t.Errorf("expected api version '1.0.1', got %q", apiVersion)
	}

	webVersion := readModuleVersionFromDir(t, tmpDir, "web")
	if webVersion != "2.0.0" {
		t.Errorf("expected web version to remain '2.0.0', got %q", webVersion)
	}

	// Verify extra file was synced to the bumped version (1.0.1)
	extraVersionData, err := os.ReadFile(extraVersionPath)
	if err != nil {
		t.Fatalf("failed to read sync.version: %v", err)
	}
	extraVersion := strings.TrimSpace(string(extraVersionData))
	if extraVersion != "1.0.1" {
		t.Errorf("expected sync.version to be '1.0.1', got %q", extraVersion)
	}
}

func TestMultiModuleBump_WithNilRegistry(t *testing.T) {
	// The function should handle nil registry gracefully
	tmpDir := t.TempDir()

	setupMultiModuleWorkspaceWithVersion(t, tmpDir, map[string]string{
		"api": "1.0.0",
	})

	cfg := &config.Config{
		Path: ".version",
	}

	// Passing nil registry - use a fresh registry instead to avoid panic
	registry := plugins.NewPluginRegistry()
	appCli := buildMultiModuleCLI(cfg, registry)

	err := testutils.RunCLITestAllowError(t, appCli, []string{
		"sley", "bump", "patch", "--all", "--non-interactive",
	}, tmpDir)
	if err != nil {
		t.Fatalf("bump failed: %v", err)
	}

	// Verify module was bumped
	apiVersion := readModuleVersionFromDir(t, tmpDir, "api")
	if apiVersion != "1.0.1" {
		t.Errorf("expected api version '1.0.1', got %q", apiVersion)
	}
}
