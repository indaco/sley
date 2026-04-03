package bump

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/indaco/sley/internal/config"
	"github.com/indaco/sley/internal/operations"
	"github.com/indaco/sley/internal/plugins"
	"github.com/indaco/sley/internal/plugins/changeloggenerator"
	"github.com/indaco/sley/internal/plugins/dependencycheck"
	"github.com/indaco/sley/internal/plugins/tagmanager"
	"github.com/indaco/sley/internal/semver"
	"github.com/indaco/sley/internal/testutils"
	"github.com/urfave/cli/v3"
)

/* ------------------------------------------------------------------------- */
/* VALIDATE TAG AVAILABLE TESTS                                              */
/* ------------------------------------------------------------------------- */

func TestValidateTagAvailable(t *testing.T) {

	version := semver.SemVersion{Major: 1, Minor: 0, Patch: 0}

	t.Run("nil tag manager returns nil", func(t *testing.T) {

		registry := plugins.NewPluginRegistry()
		err := validateTagAvailable(registry, version)
		if err != nil {
			t.Errorf("expected nil error, got %v", err)
		}
	})

	t.Run("mock tag manager returns validation error", func(t *testing.T) {

		registry := plugins.NewPluginRegistry()
		mock := &mockTagManager{validateErr: fmt.Errorf("tag exists"), autoCreateEnabled: true}
		if err := registry.RegisterTagManager(mock); err != nil {
			t.Fatalf("failed to register tag manager: %v", err)
		}
		err := validateTagAvailable(registry, version)
		if err == nil {
			t.Error("expected error, got nil")
		}
	})
}

/* ------------------------------------------------------------------------- */
/* CREATE TAG AFTER BUMP TESTS                                               */
/* ------------------------------------------------------------------------- */

func TestCreateTagAfterBump(t *testing.T) {

	version := semver.SemVersion{Major: 1, Minor: 0, Patch: 0}

	t.Run("nil tag manager returns nil", func(t *testing.T) {

		registry := plugins.NewPluginRegistry()
		err := createTagAfterBump(registry, version, "minor", nil)
		if err != nil {
			t.Errorf("expected nil error, got %v", err)
		}
	})

	// Note: createTagAfterBump uses type assertion to *TagManagerPlugin
	// so mock implementations will be treated as disabled and return nil
}

/* ------------------------------------------------------------------------- */
/* VALIDATE VERSION POLICY TESTS                                             */
/* ------------------------------------------------------------------------- */

func TestValidateVersionPolicy(t *testing.T) {

	newVersion := semver.SemVersion{Major: 2, Minor: 0, Patch: 0}
	prevVersion := semver.SemVersion{Major: 1, Minor: 0, Patch: 0}

	t.Run("nil validator returns nil", func(t *testing.T) {

		registry := plugins.NewPluginRegistry()
		err := validateVersionPolicy(registry, newVersion, prevVersion, "major")
		if err != nil {
			t.Errorf("expected nil error, got %v", err)
		}
	})

	t.Run("mock validator returns error", func(t *testing.T) {

		registry := plugins.NewPluginRegistry()
		mock := &mockVersionValidator{validateErr: fmt.Errorf("policy violation")}
		if err := registry.RegisterVersionValidator(mock); err != nil {
			t.Fatalf("failed to register version validator: %v", err)
		}
		err := validateVersionPolicy(registry, newVersion, prevVersion, "major")
		if err == nil {
			t.Error("expected error, got nil")
		}
	})
}

/* ------------------------------------------------------------------------- */
/* VALIDATE RELEASE GATE TESTS                                               */
/* ------------------------------------------------------------------------- */

func TestValidateReleaseGate(t *testing.T) {

	newVersion := semver.SemVersion{Major: 2, Minor: 0, Patch: 0}
	prevVersion := semver.SemVersion{Major: 1, Minor: 0, Patch: 0}

	t.Run("nil gate returns nil", func(t *testing.T) {

		registry := plugins.NewPluginRegistry()
		err := validateReleaseGate(registry, newVersion, prevVersion, "major")
		if err != nil {
			t.Errorf("expected nil error, got %v", err)
		}
	})

	t.Run("mock gate returns error", func(t *testing.T) {

		registry := plugins.NewPluginRegistry()
		mock := &mockReleaseGate{validateErr: fmt.Errorf("gate failed")}
		if err := registry.RegisterReleaseGate(mock); err != nil {
			t.Fatalf("failed to register release gate: %v", err)
		}
		err := validateReleaseGate(registry, newVersion, prevVersion, "major")
		if err == nil {
			t.Error("expected error, got nil")
		}
	})
}

/* ------------------------------------------------------------------------- */
/* VALIDATE DEPENDENCY CONSISTENCY TESTS                                     */
/* ------------------------------------------------------------------------- */

func TestValidateDependencyConsistency(t *testing.T) {

	version := semver.SemVersion{Major: 1, Minor: 0, Patch: 0}

	t.Run("nil checker returns nil", func(t *testing.T) {

		registry := plugins.NewPluginRegistry()
		err := validateDependencyConsistency(registry, version)
		if err != nil {
			t.Errorf("expected nil error, got %v", err)
		}
	})

	// Note: validateDependencyConsistency uses type assertion to *DependencyCheckerPlugin
	// so mock implementations will be treated as disabled and return nil
}

/* ------------------------------------------------------------------------- */
/* SYNC DEPENDENCIES TESTS                                                   */
/* ------------------------------------------------------------------------- */

func TestSyncDependencies(t *testing.T) {

	version := semver.SemVersion{Major: 1, Minor: 0, Patch: 0}

	t.Run("nil checker returns nil", func(t *testing.T) {

		registry := plugins.NewPluginRegistry()
		err := operations.SyncDependencies(registry, version)
		if err != nil {
			t.Errorf("expected nil error, got %v", err)
		}
	})

	// Note: shared.SyncDependencies uses type assertion to *DependencyCheckerPlugin
	// so mock implementations will be treated as disabled and return nil
}

/* ------------------------------------------------------------------------- */
/* GENERATE CHANGELOG AFTER BUMP TESTS                                       */
/* ------------------------------------------------------------------------- */

func TestGenerateChangelogAfterBump(t *testing.T) {

	version := semver.SemVersion{Major: 2, Minor: 0, Patch: 0}
	prevVersion := semver.SemVersion{Major: 1, Minor: 0, Patch: 0}

	t.Run("nil generator returns nil", func(t *testing.T) {

		registry := plugins.NewPluginRegistry()
		err := generateChangelogAfterBump(registry, version, prevVersion, "major", "", "")
		if err != nil {
			t.Errorf("expected nil error, got %v", err)
		}
	})

	// Note: generateChangelogAfterBump uses type assertion to *ChangelogGeneratorPlugin
	// so mock implementations will be treated as disabled and return nil
}

/* ------------------------------------------------------------------------- */
/* RECORD AUDIT LOG ENTRY TESTS                                              */
/* ------------------------------------------------------------------------- */

func TestRecordAuditLogEntry(t *testing.T) {

	version := semver.SemVersion{Major: 2, Minor: 0, Patch: 0}
	prevVersion := semver.SemVersion{Major: 1, Minor: 0, Patch: 0}

	t.Run("nil audit log returns nil", func(t *testing.T) {

		registry := plugins.NewPluginRegistry()
		err := recordAuditLogEntry(registry, version, prevVersion, "major")
		if err != nil {
			t.Errorf("expected nil error, got %v", err)
		}
	})

	// Note: recordAuditLogEntry uses type assertion to *AuditLogPlugin
	// so mock implementations will be treated as disabled and return nil
}

/* ------------------------------------------------------------------------- */
/* RUN PRE/POST BUMP EXTENSION HOOKS TESTS                                   */
/* ------------------------------------------------------------------------- */

func TestRunPreBumpExtensionHooks(t *testing.T) {

	ctx := context.Background()
	cfg := &config.Config{}

	t.Run("skip hooks returns nil", func(t *testing.T) {

		err := runPreBumpExtensionHooks(ctx, cfg, ".version", "1.0.0", "0.9.0", "minor", true)
		if err != nil {
			t.Errorf("expected nil error when skipping hooks, got %v", err)
		}
	})

	t.Run("nil config with skip returns nil", func(t *testing.T) {

		err := runPreBumpExtensionHooks(ctx, nil, ".version", "1.0.0", "0.9.0", "minor", true)
		if err != nil {
			t.Errorf("expected nil error when skipping hooks with nil config, got %v", err)
		}
	})
}

func TestRunPostBumpExtensionHooks(t *testing.T) {

	ctx := context.Background()
	tmpDir := t.TempDir()
	versionPath := filepath.Join(tmpDir, ".version")
	cfg := &config.Config{Path: versionPath}

	// Create a version file
	if err := os.WriteFile(versionPath, []byte("1.0.0\n"), 0644); err != nil {
		t.Fatal(err)
	}

	t.Run("skip hooks returns nil", func(t *testing.T) {

		err := runPostBumpExtensionHooks(ctx, cfg, versionPath, "0.9.0", "minor", true)
		if err != nil {
			t.Errorf("expected nil error when skipping hooks, got %v", err)
		}
	})
}

/* ------------------------------------------------------------------------- */
/* PLUGIN HELPER FUNCTION TESTS - ENABLED PLUGINS                          */
/* ------------------------------------------------------------------------- */

func TestCreateTagAfterBump_Enabled(t *testing.T) {

	version := semver.SemVersion{Major: 1, Minor: 2, Patch: 3}

	t.Run("disabled plugin returns nil", func(t *testing.T) {

		registry := plugins.NewPluginRegistry()
		plugin := tagmanager.NewTagManagerWithOps(&tagmanager.Config{
			Enabled: false,
		}, &tagmanager.MockGitTagOperations{}, &tagmanager.MockGitCommitOperations{})
		if err := registry.RegisterTagManager(plugin); err != nil {
			t.Fatalf("failed to register tag manager: %v", err)
		}

		err := createTagAfterBump(registry, version, "patch", nil)
		if err != nil {
			t.Errorf("expected nil error for disabled plugin, got %v", err)
		}
	})
}

func TestRunPostBumpExtensionHooks_WithError(t *testing.T) {

	ctx := context.Background()
	tmpDir := t.TempDir()
	versionPath := filepath.Join(tmpDir, ".version")

	// Write invalid version
	if err := os.WriteFile(versionPath, []byte("invalid\n"), 0644); err != nil {
		t.Fatal(err)
	}

	cfg := &config.Config{Path: versionPath}

	err := runPostBumpExtensionHooks(ctx, cfg, versionPath, "1.0.0", "patch", false)
	if err == nil {
		t.Error("expected error when reading invalid version")
	}
}

/* ------------------------------------------------------------------------- */
/* BUMP AUTO EXTENSION HOOKS TESTS                                           */
/* ------------------------------------------------------------------------- */

func TestBumpAuto_SkipHooksFlag(t *testing.T) {

	tmpDir := t.TempDir()
	versionPath := filepath.Join(tmpDir, ".version")
	testutils.WriteTempVersionFile(t, tmpDir, "1.0.0-alpha")

	cfg := &config.Config{Path: versionPath}
	registry := plugins.NewPluginRegistry()
	appCli := testutils.BuildCLIForTests(cfg.Path, []*cli.Command{Run(cfg, registry)})

	// Run with --skip-hooks flag
	err := appCli.Run(context.Background(), []string{
		"sley", "bump", "auto", "--skip-hooks",
	})
	if err != nil {
		t.Fatalf("expected no error with --skip-hooks, got: %v", err)
	}

	got := testutils.ReadTempVersionFile(t, tmpDir)
	if got != "1.0.0" {
		t.Errorf("expected version 1.0.0, got %q", got)
	}
}

func TestBumpAuto_ExtensionHooksCalledWithLabel(t *testing.T) {

	tmpDir := t.TempDir()
	versionPath := filepath.Join(tmpDir, ".version")
	testutils.WriteTempVersionFile(t, tmpDir, "1.0.0")

	cfg := &config.Config{Path: versionPath}
	registry := plugins.NewPluginRegistry()
	appCli := testutils.BuildCLIForTests(cfg.Path, []*cli.Command{Run(cfg, registry)})

	// Run with --label to ensure extension hooks path is exercised
	err := appCli.Run(context.Background(), []string{
		"sley", "bump", "auto", "--label", "patch",
	})
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	got := testutils.ReadTempVersionFile(t, tmpDir)
	if got != "1.0.1" {
		t.Errorf("expected version 1.0.1, got %q", got)
	}
}

/* ------------------------------------------------------------------------- */
/* BUMP RELEASE EXTENSION HOOKS TESTS                                        */
/* ------------------------------------------------------------------------- */

func TestBumpRelease_SkipHooksFlag(t *testing.T) {

	tmpDir := t.TempDir()
	versionPath := filepath.Join(tmpDir, ".version")
	testutils.WriteTempVersionFile(t, tmpDir, "1.0.0-beta.1")

	cfg := &config.Config{Path: versionPath}
	registry := plugins.NewPluginRegistry()
	appCli := testutils.BuildCLIForTests(cfg.Path, []*cli.Command{Run(cfg, registry)})

	// Run with --skip-hooks flag
	err := appCli.Run(context.Background(), []string{
		"sley", "bump", "release", "--skip-hooks",
	})
	if err != nil {
		t.Fatalf("expected no error with --skip-hooks, got: %v", err)
	}

	got := testutils.ReadTempVersionFile(t, tmpDir)
	if got != "1.0.0" {
		t.Errorf("expected version 1.0.0, got %q", got)
	}
}

func TestBumpRelease_ExtensionHooksCalledOnPromotion(t *testing.T) {

	tmpDir := t.TempDir()
	versionPath := filepath.Join(tmpDir, ".version")
	testutils.WriteTempVersionFile(t, tmpDir, "2.0.0-rc.1")

	cfg := &config.Config{Path: versionPath}
	registry := plugins.NewPluginRegistry()
	appCli := testutils.BuildCLIForTests(cfg.Path, []*cli.Command{Run(cfg, registry)})

	// Run release to promote pre-release
	err := appCli.Run(context.Background(), []string{
		"sley", "bump", "release",
	})
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	got := testutils.ReadTempVersionFile(t, tmpDir)
	if got != "2.0.0" {
		t.Errorf("expected version 2.0.0, got %q", got)
	}
}

/* ------------------------------------------------------------------------- */
/* SINGLE MODULE BUMP PLUGIN ERROR PATHS                                    */
/* ------------------------------------------------------------------------- */

func TestSingleModuleBump_ValidateReleaseGateFails(t *testing.T) {

	tmpDir := t.TempDir()
	versionPath := filepath.Join(tmpDir, ".version")
	testutils.WriteTempVersionFile(t, tmpDir, "1.0.0")

	// Create a release gate that fails validation
	cfg := &config.Config{Path: versionPath}
	registry := plugins.NewPluginRegistry()
	mock := &mockReleaseGate{validateErr: fmt.Errorf("release gate failed")}
	if err := registry.RegisterReleaseGate(mock); err != nil {
		t.Fatalf("failed to register release gate: %v", err)
	}
	appCli := testutils.BuildCLIForTests(cfg.Path, []*cli.Command{Run(cfg, registry)})

	err := appCli.Run(context.Background(), []string{
		"sley", "bump", "patch",
	})
	if err == nil || !strings.Contains(err.Error(), "release gate failed") {
		t.Errorf("expected release gate error, got: %v", err)
	}
}

func TestSingleModuleBump_ValidateVersionPolicyFails(t *testing.T) {

	tmpDir := t.TempDir()
	versionPath := filepath.Join(tmpDir, ".version")
	testutils.WriteTempVersionFile(t, tmpDir, "1.0.0")

	// Create a validator that fails
	cfg := &config.Config{Path: versionPath}
	registry := plugins.NewPluginRegistry()
	mock := &mockVersionValidator{validateErr: fmt.Errorf("policy violation")}
	if err := registry.RegisterVersionValidator(mock); err != nil {
		t.Fatalf("failed to register version validator: %v", err)
	}
	appCli := testutils.BuildCLIForTests(cfg.Path, []*cli.Command{Run(cfg, registry)})

	err := appCli.Run(context.Background(), []string{
		"sley", "bump", "minor",
	})
	if err == nil || !strings.Contains(err.Error(), "policy violation") {
		t.Errorf("expected policy violation error, got: %v", err)
	}
}

func TestSingleModuleBump_ValidateDependencyConsistencyFails(t *testing.T) {

	tmpDir := t.TempDir()
	versionPath := filepath.Join(tmpDir, ".version")
	testutils.WriteTempVersionFile(t, tmpDir, "1.0.0")

	// Create package.json with different version
	pkgPath := filepath.Join(tmpDir, "package.json")
	if err := os.WriteFile(pkgPath, []byte(`{"version": "0.9.0"}`), 0644); err != nil {
		t.Fatal(err)
	}

	// Create dependency checker that finds inconsistencies
	plugin := dependencycheck.NewDependencyChecker(&dependencycheck.Config{
		Enabled: true,
		Files: []dependencycheck.FileConfig{
			{Path: pkgPath, Field: "version", Format: "json"},
		},
	})

	cfg := &config.Config{Path: versionPath}
	registry := plugins.NewPluginRegistry()
	if err := registry.RegisterDependencyChecker(plugin); err != nil {
		t.Fatalf("failed to register dependency checker: %v", err)
	}
	appCli := testutils.BuildCLIForTests(cfg.Path, []*cli.Command{Run(cfg, registry)})

	err := appCli.Run(context.Background(), []string{
		"sley", "bump", "major",
	})
	if err == nil || !strings.Contains(err.Error(), "version inconsistencies detected") {
		t.Errorf("expected dependency inconsistency error, got: %v", err)
	}
}

func TestSingleModuleBump_ValidateDependencyConsistencyWithAutoSync(t *testing.T) {

	tmpDir := t.TempDir()
	versionPath := filepath.Join(tmpDir, ".version")
	testutils.WriteTempVersionFile(t, tmpDir, "1.0.0")

	// Create package.json with different version
	pkgPath := filepath.Join(tmpDir, "package.json")
	if err := os.WriteFile(pkgPath, []byte(`{"version": "0.9.0"}`), 0644); err != nil {
		t.Fatal(err)
	}

	// Create dependency checker with auto-sync enabled - should NOT fail on inconsistencies
	plugin := dependencycheck.NewDependencyChecker(&dependencycheck.Config{
		Enabled:  true,
		AutoSync: true, // This should cause the validation to pass
		Files: []dependencycheck.FileConfig{
			{Path: pkgPath, Field: "version", Format: "json"},
		},
	})

	cfg := &config.Config{Path: versionPath}
	registry := plugins.NewPluginRegistry()
	if err := registry.RegisterDependencyChecker(plugin); err != nil {
		t.Fatalf("failed to register dependency checker: %v", err)
	}
	appCli := testutils.BuildCLIForTests(cfg.Path, []*cli.Command{Run(cfg, registry)})

	// With auto-sync enabled, bump should succeed even with inconsistent versions
	err := appCli.Run(context.Background(), []string{
		"sley", "bump", "major",
	})
	if err != nil {
		t.Errorf("expected no error with auto-sync enabled, got: %v", err)
	}

	// Verify the version was bumped
	got := testutils.ReadTempVersionFile(t, tmpDir)
	if got != "2.0.0" {
		t.Errorf("expected version 2.0.0, got %q", got)
	}
}

func TestValidateDependencyConsistency_AutoSyncSkipsError(t *testing.T) {

	tmpDir := t.TempDir()

	// Create package.json with different version
	pkgPath := filepath.Join(tmpDir, "package.json")
	if err := os.WriteFile(pkgPath, []byte(`{"version": "0.9.0"}`), 0644); err != nil {
		t.Fatal(err)
	}

	version := semver.SemVersion{Major: 1, Minor: 0, Patch: 0}

	tests := []struct {
		name      string
		autoSync  bool
		expectErr bool
	}{
		{
			name:      "auto-sync disabled returns error on inconsistencies",
			autoSync:  false,
			expectErr: true,
		},
		{
			name:      "auto-sync enabled skips error on inconsistencies",
			autoSync:  true,
			expectErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			plugin := dependencycheck.NewDependencyChecker(&dependencycheck.Config{
				Enabled:  true,
				AutoSync: tt.autoSync,
				Files: []dependencycheck.FileConfig{
					{Path: pkgPath, Field: "version", Format: "json"},
				},
			})

			registry := plugins.NewPluginRegistry()
			if err := registry.RegisterDependencyChecker(plugin); err != nil {
				t.Fatalf("failed to register dependency checker: %v", err)
			}

			err := validateDependencyConsistency(registry, version)
			if tt.expectErr && err == nil {
				t.Error("expected error, got nil")
			}
			if !tt.expectErr && err != nil {
				t.Errorf("expected no error, got: %v", err)
			}
		})
	}
}

func TestSingleModuleBump_ValidateTagAvailableFails(t *testing.T) {

	tmpDir := t.TempDir()
	versionPath := filepath.Join(tmpDir, ".version")
	testutils.WriteTempVersionFile(t, tmpDir, "1.0.0")

	// Create a tag manager that fails validation
	cfg := &config.Config{Path: versionPath}
	registry := plugins.NewPluginRegistry()
	mock := &mockTagManager{validateErr: fmt.Errorf("tag already exists"), autoCreateEnabled: true}
	if err := registry.RegisterTagManager(mock); err != nil {
		t.Fatalf("failed to register tag manager: %v", err)
	}
	appCli := testutils.BuildCLIForTests(cfg.Path, []*cli.Command{Run(cfg, registry)})

	err := appCli.Run(context.Background(), []string{
		"sley", "bump", "patch",
	})
	if err == nil || !strings.Contains(err.Error(), "tag already exists") {
		t.Errorf("expected tag validation error, got: %v", err)
	}
}

func TestSingleModuleBump_UpdateVersionFails(t *testing.T) {

	tmpDir := t.TempDir()
	versionPath := filepath.Join(tmpDir, ".version")

	// Create read-only version file
	if err := os.WriteFile(versionPath, []byte("1.0.0\n"), 0444); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		_ = os.Chmod(versionPath, 0644)
	})

	cfg := &config.Config{Path: versionPath}
	registry := plugins.NewPluginRegistry()
	appCli := testutils.BuildCLIForTests(cfg.Path, []*cli.Command{Run(cfg, registry)})

	err := appCli.Run(context.Background(), []string{
		"sley", "bump", "minor", "--strict",
	})
	if err == nil {
		t.Error("expected error when updating read-only version file")
	}
}

/* ------------------------------------------------------------------------- */
/* APPLY MODULE CHANGELOG DIR TESTS                                          */
/* ------------------------------------------------------------------------- */

func TestApplyModuleChangelogDir(t *testing.T) {

	tests := []struct {
		name           string
		modulePath     string
		wantChangesDir string
	}{
		{
			name:           "empty modulePath returns noop with config unchanged",
			modulePath:     "",
			wantChangesDir: ".changes",
		},
		{
			name:           "simple module path scopes changes directory",
			modulePath:     "cobra",
			wantChangesDir: filepath.Join(".changes", "cobra"),
		},
		{
			name:           "nested module path scopes changes directory",
			modulePath:     filepath.Join("packages", "core"),
			wantChangesDir: filepath.Join(".changes", "packages", "core"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			cfg := &changeloggenerator.Config{
				Enabled:       true,
				Mode:          "versioned",
				Format:        "grouped",
				ChangesDir:    ".changes",
				ChangelogPath: "CHANGELOG.md",
			}
			plugin, err := changeloggenerator.NewChangelogGenerator(cfg)
			if err != nil {
				t.Fatalf("failed to create changelog generator: %v", err)
			}

			cleanup := applyModuleChangelog(plugin, tt.modulePath, tt.modulePath, "")

			gotCfg := plugin.GetConfig()
			if gotCfg.ChangesDir != tt.wantChangesDir {
				t.Errorf("ChangesDir: expected %q, got %q", tt.wantChangesDir, gotCfg.ChangesDir)
			}
			// ChangelogPath should remain unchanged (unified mode uses module name header instead)
			if gotCfg.ChangelogPath != "CHANGELOG.md" {
				t.Errorf("ChangelogPath: expected %q, got %q", "CHANGELOG.md", gotCfg.ChangelogPath)
			}

			// Call cleanup and verify originals are restored
			cleanup()

			restoredCfg := plugin.GetConfig()
			if restoredCfg.ChangesDir != ".changes" {
				t.Errorf("after cleanup ChangesDir: expected %q, got %q", ".changes", restoredCfg.ChangesDir)
			}
			if restoredCfg.ChangelogPath != "CHANGELOG.md" {
				t.Errorf("after cleanup ChangelogPath: expected %q, got %q", "CHANGELOG.md", restoredCfg.ChangelogPath)
			}
		})
	}
}

func TestApplyModuleChangelogDir_NonPluginType(t *testing.T) {

	mock := &mockChangelogGenerator{
		config: &changeloggenerator.Config{
			Enabled:       true,
			Mode:          "versioned",
			ChangesDir:    ".changes",
			ChangelogPath: "CHANGELOG.md",
		},
	}

	// Should return noop and not panic when cg is not *ChangelogGeneratorPlugin
	cleanup := applyModuleChangelog(mock, "cobra", "cobra", "")

	// Config should be unchanged because the type assertion fails
	gotCfg := mock.GetConfig()
	if gotCfg.ChangesDir != ".changes" {
		t.Errorf("ChangesDir should be unchanged, got %q", gotCfg.ChangesDir)
	}
	if gotCfg.ChangelogPath != "CHANGELOG.md" {
		t.Errorf("ChangelogPath should be unchanged, got %q", gotCfg.ChangelogPath)
	}

	// Cleanup should be safe to call (noop)
	cleanup()
}

func TestGenerateChangelogAfterBump_NilGeneratorWithModulePath(t *testing.T) {

	registry := plugins.NewPluginRegistry()
	version := semver.SemVersion{Major: 2, Minor: 0, Patch: 0}
	prevVersion := semver.SemVersion{Major: 1, Minor: 0, Patch: 0}

	// modulePath="cobra" with nil generator should return nil without panic
	err := generateChangelogAfterBump(registry, version, prevVersion, "major", "cobra", "cobra")
	if err != nil {
		t.Errorf("expected nil error, got %v", err)
	}
}

/* ------------------------------------------------------------------------- */
/* MOCK CHANGELOG GENERATOR FOR NON-PLUGIN TYPE TESTS                        */
/* ------------------------------------------------------------------------- */

// mockChangelogGenerator implements changeloggenerator.ChangelogGenerator
// but is not a *ChangelogGeneratorPlugin, used to test type assertion paths.
type mockChangelogGenerator struct {
	config *changeloggenerator.Config
}

func (m *mockChangelogGenerator) Name() string                          { return "mock-changelog-generator" }
func (m *mockChangelogGenerator) Description() string                   { return "mock changelog generator" }
func (m *mockChangelogGenerator) Version() string                       { return "1.0.0" }
func (m *mockChangelogGenerator) IsEnabled() bool                       { return m.config.Enabled }
func (m *mockChangelogGenerator) GetConfig() *changeloggenerator.Config { return m.config }
func (m *mockChangelogGenerator) GenerateForVersion(_, _, _ string) error {
	return nil
}
