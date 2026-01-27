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
)

/* ------------------------------------------------------------------------- */
/* MULTI-MODULE PRE-RELEASE BUMP TESTS                                       */
/* ------------------------------------------------------------------------- */

func TestMultiModuleBump_PreCommand_WithAll(t *testing.T) {
	tmpDir := t.TempDir()

	// Set up a multi-module workspace with pre-release versions
	setupMultiModuleWorkspaceWithVersion(t, tmpDir, map[string]string{
		"api": "1.0.0-rc.1",
		"web": "1.0.0-rc.1",
	})

	cfg := &config.Config{
		Path: ".version",
	}

	registry := plugins.NewPluginRegistry()
	appCli := buildMultiModuleCLI(cfg, registry)

	// Run pre bump with --all to increment pre-release on all modules
	err := testutils.RunCLITestAllowError(t, appCli, []string{
		"sley", "bump", "pre", "--all", "--non-interactive",
	}, tmpDir)
	if err != nil {
		t.Fatalf("bump pre --all failed: %v", err)
	}

	// Verify all module versions were bumped to rc.2
	apiVersion := readModuleVersionFromDir(t, tmpDir, "api")
	if apiVersion != "1.0.0-rc.2" {
		t.Errorf("expected api version '1.0.0-rc.2', got %q", apiVersion)
	}

	webVersion := readModuleVersionFromDir(t, tmpDir, "web")
	if webVersion != "1.0.0-rc.2" {
		t.Errorf("expected web version '1.0.0-rc.2', got %q", webVersion)
	}
}

func TestMultiModuleBump_PreCommand_WithModule(t *testing.T) {
	tmpDir := t.TempDir()

	// Set up a multi-module workspace with pre-release versions
	setupMultiModuleWorkspaceWithVersion(t, tmpDir, map[string]string{
		"api": "1.0.0-beta.1",
		"web": "2.0.0-alpha.1",
	})

	cfg := &config.Config{
		Path: ".version",
	}

	registry := plugins.NewPluginRegistry()
	appCli := buildMultiModuleCLI(cfg, registry)

	// Run pre bump on just the api module
	err := testutils.RunCLITestAllowError(t, appCli, []string{
		"sley", "bump", "pre", "--module", "api", "--non-interactive",
	}, tmpDir)
	if err != nil {
		t.Fatalf("bump pre --module api failed: %v", err)
	}

	// Verify only api was bumped
	apiVersion := readModuleVersionFromDir(t, tmpDir, "api")
	if apiVersion != "1.0.0-beta.2" {
		t.Errorf("expected api version '1.0.0-beta.2', got %q", apiVersion)
	}

	// Verify web was NOT bumped
	webVersion := readModuleVersionFromDir(t, tmpDir, "web")
	if webVersion != "2.0.0-alpha.1" {
		t.Errorf("expected web version to remain '2.0.0-alpha.1', got %q", webVersion)
	}
}

func TestMultiModuleBump_PreCommand_WithLabel(t *testing.T) {
	tmpDir := t.TempDir()

	// Set up a multi-module workspace with stable versions
	setupMultiModuleWorkspaceWithVersion(t, tmpDir, map[string]string{
		"api": "1.0.0",
		"web": "1.0.0",
	})

	cfg := &config.Config{
		Path: ".version",
	}

	registry := plugins.NewPluginRegistry()
	appCli := buildMultiModuleCLI(cfg, registry)

	// Run pre bump with --label to add pre-release
	err := testutils.RunCLITestAllowError(t, appCli, []string{
		"sley", "bump", "pre", "--all", "--label", "rc", "--non-interactive",
	}, tmpDir)
	if err != nil {
		t.Fatalf("bump pre --all --label rc failed: %v", err)
	}

	// Verify all modules got the rc.1 pre-release
	apiVersion := readModuleVersionFromDir(t, tmpDir, "api")
	if apiVersion != "1.0.0-rc.1" {
		t.Errorf("expected api version '1.0.0-rc.1', got %q", apiVersion)
	}

	webVersion := readModuleVersionFromDir(t, tmpDir, "web")
	if webVersion != "1.0.0-rc.1" {
		t.Errorf("expected web version '1.0.0-rc.1', got %q", webVersion)
	}
}

func TestMultiModuleBump_PreCommand_SyncsDependencies(t *testing.T) {
	tmpDir := t.TempDir()

	// Set up a multi-module workspace with pre-release versions
	setupMultiModuleWorkspaceWithVersion(t, tmpDir, map[string]string{
		"api": "1.0.0-rc.1",
		"web": "1.0.0-rc.1",
	})

	// Create additional file to sync
	extraVersionPath := filepath.Join(tmpDir, "pre.version")
	if err := os.WriteFile(extraVersionPath, []byte("1.0.0-rc.1\n"), 0644); err != nil {
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

	// Run pre bump with --all
	err := testutils.RunCLITestAllowError(t, appCli, []string{
		"sley", "bump", "pre", "--all", "--non-interactive",
	}, tmpDir)
	if err != nil {
		t.Fatalf("bump pre --all failed: %v", err)
	}

	// Verify modules were bumped
	apiVersion := readModuleVersionFromDir(t, tmpDir, "api")
	if apiVersion != "1.0.0-rc.2" {
		t.Errorf("expected api version '1.0.0-rc.2', got %q", apiVersion)
	}

	// Verify extra file was synced
	extraVersionData, err := os.ReadFile(extraVersionPath)
	if err != nil {
		t.Fatalf("failed to read pre.version: %v", err)
	}
	extraVersion := strings.TrimSpace(string(extraVersionData))
	if extraVersion != "1.0.0-rc.2" {
		t.Errorf("expected pre.version to be '1.0.0-rc.2', got %q", extraVersion)
	}
}
