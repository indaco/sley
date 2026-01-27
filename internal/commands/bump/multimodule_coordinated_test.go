package bump

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/indaco/sley/internal/config"
	"github.com/indaco/sley/internal/plugins"
	"github.com/indaco/sley/internal/plugins/dependencycheck"
	"github.com/indaco/sley/internal/testutils"
	"github.com/urfave/cli/v3"
)

/* ------------------------------------------------------------------------- */
/* COORDINATED VERSIONING TESTS                                              */
/* ------------------------------------------------------------------------- */

func TestCoordinatedVersioning_SingleRootBumpSyncsSubmodules(t *testing.T) {
	tmpDir := t.TempDir()

	// Create root .version file
	rootVersionPath := filepath.Join(tmpDir, ".version")
	if err := os.WriteFile(rootVersionPath, []byte("1.0.0\n"), 0644); err != nil {
		t.Fatal(err)
	}

	// Create submodule .version files (simulating coordinated versioning setup)
	submodules := []string{"services/api", "services/web"}
	for _, submod := range submodules {
		submodDir := filepath.Join(tmpDir, submod)
		if err := os.MkdirAll(submodDir, 0755); err != nil {
			t.Fatalf("failed to create submodule dir %s: %v", submodDir, err)
		}
		subVersionPath := filepath.Join(submodDir, ".version")
		if err := os.WriteFile(subVersionPath, []byte("1.0.0\n"), 0644); err != nil {
			t.Fatal(err)
		}
	}

	// Create config with explicit root path (single-module mode) and dependency-check
	// configured to sync the submodule .version files (coordinated versioning)
	cfg := &config.Config{
		Path: rootVersionPath, // Explicit path triggers single-module mode
	}

	// Configure dependency-check plugin with submodule .version files as sync targets
	depPlugin := dependencycheck.NewDependencyChecker(&dependencycheck.Config{
		Enabled:  true,
		AutoSync: true,
		Files: []dependencycheck.FileConfig{
			{Path: filepath.Join(tmpDir, "services/api/.version"), Format: "raw"},
			{Path: filepath.Join(tmpDir, "services/web/.version"), Format: "raw"},
		},
	})

	registry := plugins.NewPluginRegistry()
	if err := registry.RegisterDependencyChecker(depPlugin); err != nil {
		t.Fatalf("failed to register dependency checker: %v", err)
	}

	// Build CLI with explicit path for single-module mode
	appCli := testutils.BuildCLIForTests(rootVersionPath, []*cli.Command{Run(cfg, registry)})

	// Run single-module bump (no --all flag, no --module flag)
	err := testutils.RunCLITestAllowError(t, appCli, []string{
		"sley", "bump", "patch",
	}, tmpDir)
	if err != nil {
		t.Fatalf("bump patch failed: %v", err)
	}

	// Verify root version was bumped
	rootData, err := os.ReadFile(rootVersionPath)
	if err != nil {
		t.Fatalf("failed to read root .version: %v", err)
	}
	rootVersion := strings.TrimSpace(string(rootData))
	if rootVersion != "1.0.1" {
		t.Errorf("expected root version '1.0.1', got %q", rootVersion)
	}

	// Verify all submodule .version files were synced to 1.0.1
	for _, submod := range submodules {
		subVersionPath := filepath.Join(tmpDir, submod, ".version")
		subData, err := os.ReadFile(subVersionPath)
		if err != nil {
			t.Fatalf("failed to read %s/.version: %v", submod, err)
		}
		subVersion := strings.TrimSpace(string(subData))
		if subVersion != "1.0.1" {
			t.Errorf("expected %s version '1.0.1', got %q", submod, subVersion)
		}
	}
}

func TestCoordinatedVersioning_MinorBumpSyncsSubmodules(t *testing.T) {
	tmpDir := t.TempDir()

	// Create root .version file
	rootVersionPath := filepath.Join(tmpDir, ".version")
	if err := os.WriteFile(rootVersionPath, []byte("1.2.3\n"), 0644); err != nil {
		t.Fatal(err)
	}

	// Create submodule .version files
	submodDir := filepath.Join(tmpDir, "packages/core")
	if err := os.MkdirAll(submodDir, 0755); err != nil {
		t.Fatal(err)
	}
	subVersionPath := filepath.Join(submodDir, ".version")
	if err := os.WriteFile(subVersionPath, []byte("1.2.3\n"), 0644); err != nil {
		t.Fatal(err)
	}

	cfg := &config.Config{
		Path: rootVersionPath,
	}

	depPlugin := dependencycheck.NewDependencyChecker(&dependencycheck.Config{
		Enabled:  true,
		AutoSync: true,
		Files: []dependencycheck.FileConfig{
			{Path: subVersionPath, Format: "raw"},
		},
	})

	registry := plugins.NewPluginRegistry()
	if err := registry.RegisterDependencyChecker(depPlugin); err != nil {
		t.Fatalf("failed to register dependency checker: %v", err)
	}

	appCli := testutils.BuildCLIForTests(rootVersionPath, []*cli.Command{Run(cfg, registry)})

	// Run minor bump
	err := testutils.RunCLITestAllowError(t, appCli, []string{
		"sley", "bump", "minor",
	}, tmpDir)
	if err != nil {
		t.Fatalf("bump minor failed: %v", err)
	}

	// Verify root version was bumped to 1.3.0
	rootData, err := os.ReadFile(rootVersionPath)
	if err != nil {
		t.Fatalf("failed to read root .version: %v", err)
	}
	rootVersion := strings.TrimSpace(string(rootData))
	if rootVersion != "1.3.0" {
		t.Errorf("expected root version '1.3.0', got %q", rootVersion)
	}

	// Verify submodule was synced
	subData, err := os.ReadFile(subVersionPath)
	if err != nil {
		t.Fatalf("failed to read packages/core/.version: %v", err)
	}
	subVersion := strings.TrimSpace(string(subData))
	if subVersion != "1.3.0" {
		t.Errorf("expected packages/core version '1.3.0', got %q", subVersion)
	}
}

func TestCoordinatedVersioning_SyncsManifestsAndVersionFiles(t *testing.T) {
	tmpDir := t.TempDir()

	// Create root .version file
	rootVersionPath := filepath.Join(tmpDir, ".version")
	if err := os.WriteFile(rootVersionPath, []byte("2.0.0\n"), 0644); err != nil {
		t.Fatal(err)
	}

	// Create submodule .version file
	submodDir := filepath.Join(tmpDir, "app")
	if err := os.MkdirAll(submodDir, 0755); err != nil {
		t.Fatal(err)
	}
	subVersionPath := filepath.Join(submodDir, ".version")
	if err := os.WriteFile(subVersionPath, []byte("2.0.0\n"), 0644); err != nil {
		t.Fatal(err)
	}

	// Create package.json in the app directory
	pkgJSONPath := filepath.Join(submodDir, "package.json")
	if err := os.WriteFile(pkgJSONPath, []byte(`{"name": "app", "version": "2.0.0"}`), 0644); err != nil {
		t.Fatal(err)
	}

	cfg := &config.Config{
		Path: rootVersionPath,
	}

	// Configure dependency-check to sync both .version files and manifests
	depPlugin := dependencycheck.NewDependencyChecker(&dependencycheck.Config{
		Enabled:  true,
		AutoSync: true,
		Files: []dependencycheck.FileConfig{
			{Path: subVersionPath, Format: "raw"},
			{Path: pkgJSONPath, Field: "version", Format: "json"},
		},
	})

	registry := plugins.NewPluginRegistry()
	if err := registry.RegisterDependencyChecker(depPlugin); err != nil {
		t.Fatalf("failed to register dependency checker: %v", err)
	}

	appCli := testutils.BuildCLIForTests(rootVersionPath, []*cli.Command{Run(cfg, registry)})

	// Run major bump
	err := testutils.RunCLITestAllowError(t, appCli, []string{
		"sley", "bump", "major",
	}, tmpDir)
	if err != nil {
		t.Fatalf("bump major failed: %v", err)
	}

	// Verify root version was bumped to 3.0.0
	rootData, err := os.ReadFile(rootVersionPath)
	if err != nil {
		t.Fatalf("failed to read root .version: %v", err)
	}
	rootVersion := strings.TrimSpace(string(rootData))
	if rootVersion != "3.0.0" {
		t.Errorf("expected root version '3.0.0', got %q", rootVersion)
	}

	// Verify submodule .version was synced
	subData, err := os.ReadFile(subVersionPath)
	if err != nil {
		t.Fatalf("failed to read app/.version: %v", err)
	}
	subVersion := strings.TrimSpace(string(subData))
	if subVersion != "3.0.0" {
		t.Errorf("expected app version '3.0.0', got %q", subVersion)
	}

	// Verify package.json was synced
	pkgData, err := os.ReadFile(pkgJSONPath)
	if err != nil {
		t.Fatalf("failed to read package.json: %v", err)
	}
	if !strings.Contains(string(pkgData), `"version": "3.0.0"`) {
		t.Errorf("expected package.json to contain version 3.0.0, got: %s", string(pkgData))
	}
}
