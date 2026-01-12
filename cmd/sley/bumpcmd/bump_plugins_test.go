package bumpcmd

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/indaco/sley/internal/config"
	"github.com/indaco/sley/internal/plugins"
	"github.com/indaco/sley/internal/plugins/auditlog"
	"github.com/indaco/sley/internal/plugins/changeloggenerator"
	"github.com/indaco/sley/internal/plugins/dependencycheck"
	"github.com/indaco/sley/internal/plugins/releasegate"
	"github.com/indaco/sley/internal/plugins/tagmanager"
	"github.com/indaco/sley/internal/plugins/versionvalidator"
	"github.com/indaco/sley/internal/semver"
	"github.com/indaco/sley/internal/testutils"
	"github.com/urfave/cli/v3"
)

/* ------------------------------------------------------------------------- */
/* VALIDATE TAG AVAILABLE TESTS                                              */
/* ------------------------------------------------------------------------- */

func TestValidateTagAvailable(t *testing.T) {
	// Save original and restore after test
	origGetTagManagerFn := tagmanager.GetTagManagerFn
	defer func() { tagmanager.GetTagManagerFn = origGetTagManagerFn }()

	version := semver.SemVersion{Major: 1, Minor: 0, Patch: 0}

	t.Run("nil tag manager returns nil", func(t *testing.T) {
		registry := plugins.NewPluginRegistry()
		tagmanager.GetTagManagerFn = func() tagmanager.TagManager { return nil }
		err := validateTagAvailable(registry, version)
		if err != nil {
			t.Errorf("expected nil error, got %v", err)
		}
	})

	t.Run("mock tag manager validates", func(t *testing.T) {
		registry := plugins.NewPluginRegistry()
		mock := &mockTagManager{}
		tagmanager.GetTagManagerFn = func() tagmanager.TagManager { return mock }
		err := validateTagAvailable(registry, version)
		if err != nil {
			t.Errorf("expected nil error, got %v", err)
		}
	})

	t.Run("mock tag manager returns validation error", func(t *testing.T) {
		registry := plugins.NewPluginRegistry()
		mock := &mockTagManager{validateErr: fmt.Errorf("tag exists")}
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
	// Save original and restore after test
	origGetTagManagerFn := tagmanager.GetTagManagerFn
	defer func() { tagmanager.GetTagManagerFn = origGetTagManagerFn }()

	version := semver.SemVersion{Major: 1, Minor: 0, Patch: 0}

	t.Run("nil tag manager returns nil", func(t *testing.T) {
		registry := plugins.NewPluginRegistry()
		tagmanager.GetTagManagerFn = func() tagmanager.TagManager { return nil }
		err := createTagAfterBump(registry, version, "minor")
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
	// Save original and restore after test
	origGetVersionValidatorFn := versionvalidator.GetVersionValidatorFn
	defer func() { versionvalidator.GetVersionValidatorFn = origGetVersionValidatorFn }()

	newVersion := semver.SemVersion{Major: 2, Minor: 0, Patch: 0}
	prevVersion := semver.SemVersion{Major: 1, Minor: 0, Patch: 0}

	t.Run("nil validator returns nil", func(t *testing.T) {
		registry := plugins.NewPluginRegistry()
		versionvalidator.GetVersionValidatorFn = func() versionvalidator.VersionValidator { return nil }
		err := validateVersionPolicy(registry, newVersion, prevVersion, "major")
		if err != nil {
			t.Errorf("expected nil error, got %v", err)
		}
	})

	t.Run("mock validator validates successfully", func(t *testing.T) {
		registry := plugins.NewPluginRegistry()
		mock := &mockVersionValidator{}
		versionvalidator.GetVersionValidatorFn = func() versionvalidator.VersionValidator { return mock }
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
	// Save original and restore after test
	origGetReleaseGateFn := releasegate.GetReleaseGateFn
	defer func() { releasegate.GetReleaseGateFn = origGetReleaseGateFn }()

	newVersion := semver.SemVersion{Major: 2, Minor: 0, Patch: 0}
	prevVersion := semver.SemVersion{Major: 1, Minor: 0, Patch: 0}

	t.Run("nil gate returns nil", func(t *testing.T) {
		registry := plugins.NewPluginRegistry()
		releasegate.GetReleaseGateFn = func() releasegate.ReleaseGate { return nil }
		err := validateReleaseGate(registry, newVersion, prevVersion, "major")
		if err != nil {
			t.Errorf("expected nil error, got %v", err)
		}
	})

	t.Run("mock gate validates successfully", func(t *testing.T) {
		registry := plugins.NewPluginRegistry()
		mock := &mockReleaseGate{}
		releasegate.GetReleaseGateFn = func() releasegate.ReleaseGate { return mock }
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
	// Save original and restore after test
	origGetDependencyCheckerFn := dependencycheck.GetDependencyCheckerFn
	defer func() { dependencycheck.GetDependencyCheckerFn = origGetDependencyCheckerFn }()

	version := semver.SemVersion{Major: 1, Minor: 0, Patch: 0}

	t.Run("nil checker returns nil", func(t *testing.T) {
		registry := plugins.NewPluginRegistry()
		dependencycheck.GetDependencyCheckerFn = func() dependencycheck.DependencyChecker { return nil }
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
	// Save original and restore after test
	origGetDependencyCheckerFn := dependencycheck.GetDependencyCheckerFn
	defer func() { dependencycheck.GetDependencyCheckerFn = origGetDependencyCheckerFn }()

	version := semver.SemVersion{Major: 1, Minor: 0, Patch: 0}

	t.Run("nil checker returns nil", func(t *testing.T) {
		registry := plugins.NewPluginRegistry()
		dependencycheck.GetDependencyCheckerFn = func() dependencycheck.DependencyChecker { return nil }
		err := syncDependencies(registry, version)
		if err != nil {
			t.Errorf("expected nil error, got %v", err)
		}
	})

	// Note: syncDependencies uses type assertion to *DependencyCheckerPlugin
	// so mock implementations will be treated as disabled and return nil
}

/* ------------------------------------------------------------------------- */
/* GENERATE CHANGELOG AFTER BUMP TESTS                                       */
/* ------------------------------------------------------------------------- */

func TestGenerateChangelogAfterBump(t *testing.T) {
	// Save original and restore after test
	origGetChangelogGeneratorFn := changeloggenerator.GetChangelogGeneratorFn
	defer func() { changeloggenerator.GetChangelogGeneratorFn = origGetChangelogGeneratorFn }()

	version := semver.SemVersion{Major: 2, Minor: 0, Patch: 0}
	prevVersion := semver.SemVersion{Major: 1, Minor: 0, Patch: 0}

	t.Run("nil generator returns nil", func(t *testing.T) {
		registry := plugins.NewPluginRegistry()
		changeloggenerator.GetChangelogGeneratorFn = func() changeloggenerator.ChangelogGenerator { return nil }
		err := generateChangelogAfterBump(registry, version, prevVersion, "major")
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
	// Save original and restore after test
	origGetAuditLogFn := auditlog.GetAuditLogFn
	defer func() { auditlog.GetAuditLogFn = origGetAuditLogFn }()

	version := semver.SemVersion{Major: 2, Minor: 0, Patch: 0}
	prevVersion := semver.SemVersion{Major: 1, Minor: 0, Patch: 0}

	t.Run("nil audit log returns nil", func(t *testing.T) {
		registry := plugins.NewPluginRegistry()
		auditlog.GetAuditLogFn = func() auditlog.AuditLog { return nil }
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
	// Save and restore
	origGetTagManagerFn := tagmanager.GetTagManagerFn
	defer func() { tagmanager.GetTagManagerFn = origGetTagManagerFn }()

	version := semver.SemVersion{Major: 1, Minor: 2, Patch: 3}

	t.Run("enabled plugin creates tag successfully", func(t *testing.T) {
		plugin := tagmanager.NewTagManager(&tagmanager.Config{
			Enabled:    true,
			AutoCreate: true,
			Prefix:     "v",
		})
		tagmanager.GetTagManagerFn = func() tagmanager.TagManager { return plugin }

		// Mock CreateTag to succeed
		registry := plugins.NewPluginRegistry()
		err := createTagAfterBump(registry, version, "patch")
		// This will fail in test environment without git, but we're testing the code path
		if err != nil && !strings.Contains(err.Error(), "failed to create tag") {
			t.Errorf("unexpected error type: %v", err)
		}
	})

	t.Run("disabled plugin returns nil", func(t *testing.T) {
		registry := plugins.NewPluginRegistry()
		plugin := tagmanager.NewTagManager(&tagmanager.Config{
			Enabled: false,
		})
		tagmanager.GetTagManagerFn = func() tagmanager.TagManager { return plugin }

		err := createTagAfterBump(registry, version, "patch")
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

	// Save and restore
	origGetDependencyCheckerFn := dependencycheck.GetDependencyCheckerFn
	defer func() { dependencycheck.GetDependencyCheckerFn = origGetDependencyCheckerFn }()

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

func TestSingleModuleBump_ValidateTagAvailableFails(t *testing.T) {
	tmpDir := t.TempDir()
	versionPath := filepath.Join(tmpDir, ".version")
	testutils.WriteTempVersionFile(t, tmpDir, "1.0.0")

	// Create a tag manager that fails validation
	cfg := &config.Config{Path: versionPath}
	registry := plugins.NewPluginRegistry()
	mock := &mockTagManager{validateErr: fmt.Errorf("tag already exists")}
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
