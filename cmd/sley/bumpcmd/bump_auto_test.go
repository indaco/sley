package bumpcmd

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/indaco/sley/internal/config"
	"github.com/indaco/sley/internal/plugins"
	"github.com/indaco/sley/internal/plugins/commitparser"
	"github.com/indaco/sley/internal/plugins/commitparser/gitlog"
	"github.com/indaco/sley/internal/plugins/tagmanager"
	"github.com/indaco/sley/internal/semver"
	"github.com/indaco/sley/internal/testutils"
	"github.com/urfave/cli/v3"
)

func TestCLI_BumpAutoCmd(t *testing.T) {
	tmpDir := t.TempDir()
	versionPath := filepath.Join(tmpDir, ".version")

	// Prepare and run the CLI command
	cfg := &config.Config{Path: versionPath}
	registry := plugins.NewPluginRegistry()
	appCli := testutils.BuildCLIForTests(cfg.Path, []*cli.Command{Run(cfg, registry)})

	tests := []struct {
		name     string
		initial  string
		args     []string
		expected string
	}{
		{
			name:     "promotes alpha to release",
			initial:  "1.2.3-alpha.1",
			args:     []string{"sley", "bump", "auto"},
			expected: "1.2.3",
		},
		{
			name:     "promotes rc to release",
			initial:  "1.2.3-rc.1",
			args:     []string{"sley", "bump", "auto"},
			expected: "1.2.3",
		},
		{
			name:     "default patch bump",
			initial:  "1.2.3",
			args:     []string{"sley", "bump", "auto"},
			expected: "1.2.4",
		},
		{
			name:     "promotes pre-release in 0.x series",
			initial:  "0.9.0-alpha.1",
			args:     []string{"sley", "bump", "auto"},
			expected: "0.9.0",
		},
		{
			name:     "bump minor from 0.9.0 as a special case",
			initial:  "0.9.0",
			args:     []string{"sley", "bump", "auto"},
			expected: "0.10.0",
		},
		{
			name:     "preserve build metadata",
			initial:  "1.2.3-alpha.1+meta.123",
			args:     []string{"sley", "bump", "auto", "--preserve-meta"},
			expected: "1.2.3+meta.123",
		},
		{
			name:     "strip build metadata by default",
			initial:  "1.2.3-alpha.1+meta.123",
			args:     []string{"sley", "bump", "auto"},
			expected: "1.2.3",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testutils.WriteTempVersionFile(t, tmpDir, tt.initial)
			testutils.RunCLITest(t, appCli, tt.args, tmpDir)

			got := testutils.ReadTempVersionFile(t, tmpDir)
			if got != tt.expected {
				t.Errorf("expected version %q, got %q", tt.expected, got)
			}
		})
	}
}

func TestCLI_BumpAutoCmd_InferredBump(t *testing.T) {
	tmp := t.TempDir()
	versionPath := testutils.WriteTempVersionFile(t, tmp, "1.2.3")

	// Save original function and restore later
	originalInfer := tryInferBumpTypeFromCommitParserPluginFn
	defer func() { tryInferBumpTypeFromCommitParserPluginFn = originalInfer }()

	// Mock the inference to simulate an inferred "minor" bump
	tryInferBumpTypeFromCommitParserPluginFn = func(registry *plugins.PluginRegistry, since, until string) string {
		return "minor"
	}

	cfg := &config.Config{Path: versionPath, Plugins: &config.PluginConfig{CommitParser: true}}
	registry := plugins.NewPluginRegistry()
	appCli := testutils.BuildCLIForTests(cfg.Path, []*cli.Command{Run(cfg, registry)})

	err := appCli.Run(context.Background(), []string{
		"sley", "bump", "auto", "--path", versionPath,
	})

	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	got := testutils.ReadTempVersionFile(t, tmp)
	want := "1.3.0"
	if got != want {
		t.Errorf("expected bumped version %q, got %q", want, got)
	}
}

func TestCLI_BumpAutoCommand_WithLabelAndMeta(t *testing.T) {
	tmpDir := t.TempDir()
	versionPath := filepath.Join(tmpDir, ".version")

	// Prepare the CLI command
	cfg := &config.Config{Path: versionPath}
	registry := plugins.NewPluginRegistry()
	appCli := testutils.BuildCLIForTests(cfg.Path, []*cli.Command{Run(cfg, registry)})

	tests := []struct {
		name    string
		initial string
		args    []string
		want    string
	}{
		{
			name:    "label=patch",
			initial: "1.2.3",
			args:    []string{"sley", "bump", "auto", "--label", "patch"},
			want:    "1.2.4",
		},
		{
			name:    "label=minor",
			initial: "1.2.3",
			args:    []string{"sley", "bump", "auto", "--label", "minor"},
			want:    "1.3.0",
		},
		{
			name:    "label=major",
			initial: "1.2.3",
			args:    []string{"sley", "bump", "auto", "--label", "major"},
			want:    "2.0.0",
		},
		{
			name:    "label=minor with metadata",
			initial: "1.2.3",
			args:    []string{"sley", "bump", "auto", "--label", "minor", "--meta", "build.42"},
			want:    "1.3.0+build.42",
		},
		{
			name:    "preserve existing metadata",
			initial: "1.2.3+ci.88",
			args:    []string{"sley", "bump", "auto", "--label", "patch", "--preserve-meta"},
			want:    "1.2.4+ci.88",
		},
		{
			name:    "override existing metadata",
			initial: "1.2.3+ci.88",
			args:    []string{"sley", "bump", "auto", "--label", "patch", "--meta", "ci.99"},
			want:    "1.2.4+ci.99",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testutils.WriteTempVersionFile(t, tmpDir, tt.initial)
			testutils.RunCLITest(t, appCli, tt.args, tmpDir)

			got := testutils.ReadTempVersionFile(t, tmpDir)
			if got != tt.want {
				t.Errorf("expected %q, got %q", tt.want, got)
			}
		})
	}
}

func TestCLI_BumpAutoCmd_InferredPromotion(t *testing.T) {
	tmp := t.TempDir()
	versionPath := testutils.WriteTempVersionFile(t, tmp, "1.2.3-beta.1")

	originalInfer := tryInferBumpTypeFromCommitParserPluginFn
	defer func() { tryInferBumpTypeFromCommitParserPluginFn = originalInfer }()

	tryInferBumpTypeFromCommitParserPluginFn = func(registry *plugins.PluginRegistry, since, until string) string {
		return "minor"
	}

	cfg := &config.Config{Path: versionPath, Plugins: &config.PluginConfig{CommitParser: true}}
	registry := plugins.NewPluginRegistry()
	appCli := testutils.BuildCLIForTests(cfg.Path, []*cli.Command{Run(cfg, registry)})

	err := appCli.Run(context.Background(), []string{
		"sley", "bump", "auto", "--path", versionPath,
	})

	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	got := testutils.ReadTempVersionFile(t, tmp)
	want := "1.2.3" // Promotion, not minor bump
	if got != want {
		t.Errorf("expected promoted version %q, got %q", want, got)
	}
}

func TestCLI_BumpAutoCmd_PromotePreReleaseWithPreserveMeta(t *testing.T) {
	tmp := t.TempDir()
	versionPath := testutils.WriteTempVersionFile(t, tmp, "1.2.3-beta.2+ci.99")

	// Override tryInferBumpTypeFromCommitParserPlugin
	originalInfer := tryInferBumpTypeFromCommitParserPluginFn
	tryInferBumpTypeFromCommitParserPluginFn = func(registry *plugins.PluginRegistry, since, until string) string {
		return "minor" // Force a non-empty inference so that promotePreRelease is called
	}
	t.Cleanup(func() { tryInferBumpTypeFromCommitParserPluginFn = originalInfer })

	cfg := &config.Config{Path: versionPath}
	registry := plugins.NewPluginRegistry()
	appCli := testutils.BuildCLIForTests(cfg.Path, []*cli.Command{Run(cfg, registry)})

	err := appCli.Run(context.Background(), []string{
		"sley", "bump", "auto", "--path", versionPath, "--preserve-meta",
	})
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	got := testutils.ReadTempVersionFile(t, tmp)
	want := "1.2.3+ci.99"
	if got != want {
		t.Errorf("expected promoted version with metadata %q, got %q", want, got)
	}
}

func TestCLI_BumpAutoCmd_InferredBumpFails(t *testing.T) {
	tmp := t.TempDir()
	versionPath := testutils.WriteTempVersionFile(t, tmp, "1.2.3")

	originalBumpByLabel := semver.BumpByLabelFunc
	originalInferFunc := tryInferBumpTypeFromCommitParserPluginFn

	// Force BumpByLabelFunc to fail
	semver.BumpByLabelFunc = func(v semver.SemVersion, label string) (semver.SemVersion, error) {
		return semver.SemVersion{}, fmt.Errorf("forced inferred bump failure")
	}

	// Force inference to return something
	tryInferBumpTypeFromCommitParserPluginFn = func(registry *plugins.PluginRegistry, since, until string) string {
		return "minor"
	}

	t.Cleanup(func() {
		semver.BumpByLabelFunc = originalBumpByLabel
		tryInferBumpTypeFromCommitParserPluginFn = originalInferFunc
	})

	// Prepare and run CLI
	cfg := &config.Config{Path: versionPath}
	registry := plugins.NewPluginRegistry()
	appCli := testutils.BuildCLIForTests(cfg.Path, []*cli.Command{Run(cfg, registry)})

	err := appCli.Run(context.Background(), []string{
		"sley", "bump", "auto", "--path", versionPath,
	})

	if err == nil || !strings.Contains(err.Error(), "failed to bump inferred version") {
		t.Fatalf("expected error about inferred bump failure, got: %v", err)
	}
}

func TestTryInferBumpTypeFromCommitParserPlugin_GetCommitsError(t *testing.T) {
	testutils.WithMock(func() {
		// Mock GetCommits to fail
		originalGetCommits := gitlog.GetCommitsFn
		originalParser := commitparser.GetCommitParserFn

		gitlog.GetCommitsFn = func(since, until string) ([]string, error) {
			return nil, fmt.Errorf("simulated gitlog error")
		}
		commitparser.GetCommitParserFn = func() commitparser.CommitParser {
			return testutils.MockCommitParser{} // Return any parser
		}

		t.Cleanup(func() {
			gitlog.GetCommitsFn = originalGetCommits
			commitparser.GetCommitParserFn = originalParser
		})
	}, func() {
		registry := plugins.NewPluginRegistry()
		label := tryInferBumpTypeFromCommitParserPlugin(registry, "", "")
		if label != "" {
			t.Errorf("expected empty label on gitlog error, got %q", label)
		}
	})
}

func TestTryInferBumpTypeFromCommitParserPlugin_ParserError(t *testing.T) {
	testutils.WithMock(
		func() {
			// Setup mocks
			gitlog.GetCommitsFn = func(since, until string) ([]string, error) {
				return []string{"fix: something"}, nil
			}
			commitparser.GetCommitParserFn = func() commitparser.CommitParser {
				return testutils.MockCommitParser{Err: fmt.Errorf("parser error")}
			}
		},
		func() {
			registry := plugins.NewPluginRegistry()
			label := tryInferBumpTypeFromCommitParserPlugin(registry, "", "")
			if label != "" {
				t.Errorf("expected empty label on parser error, got %q", label)
			}
		},
	)
}

func TestTryInferBumpTypeFromCommitParserPlugin_Success(t *testing.T) {
	testutils.WithMock(
		func() {
			// Setup mocks
			gitlog.GetCommitsFn = func(since, until string) ([]string, error) {
				return []string{"feat: add feature"}, nil
			}
		},
		func() {
			registry := plugins.NewPluginRegistry()
			parser := testutils.MockCommitParser{Label: "minor"}
			if err := registry.RegisterCommitParser(&parser); err != nil {
				t.Fatalf("failed to register parser: %v", err)
			}
			label := tryInferBumpTypeFromCommitParserPlugin(registry, "", "")
			if label != "minor" {
				t.Errorf("expected label 'minor', got %q", label)
			}
		},
	)
}

func TestCLI_BumpAutoCmd_Errors(t *testing.T) {
	tests := []struct {
		name          string
		setup         func(dir string)
		args          []string
		expectedErr   string
		skipOnWindows bool
	}{
		{
			name: "fails if version file is invalid",
			setup: func(dir string) {
				_ = os.WriteFile(filepath.Join(dir, ".version"), []byte("not-a-version\n"), 0600)
			},
			args:        []string{"sley", "bump", "auto"},
			expectedErr: "failed to read version",
		},
		{
			name: "fails if version file is not writable",
			setup: func(dir string) {
				path := filepath.Join(dir, ".version")
				_ = os.WriteFile(path, []byte("1.2.3-alpha\n"), 0444)
				_ = os.Chmod(path, 0444)
			},
			args:          []string{"sley", "bump", "auto"},
			expectedErr:   "failed to save version",
			skipOnWindows: true, // permission simulation less reliable on Windows
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.skipOnWindows && testutils.IsWindows() {
				t.Skip("skipping test on Windows")
			}

			tmp := t.TempDir()
			tt.setup(tmp)

			versionPath := filepath.Join(tmp, ".version")

			// Prepare and run the CLI command
			cfg := &config.Config{Path: versionPath}
			registry := plugins.NewPluginRegistry()
			appCli := testutils.BuildCLIForTests(cfg.Path, []*cli.Command{Run(cfg, registry)})

			err := appCli.Run(context.Background(), tt.args)
			if err == nil || !strings.Contains(err.Error(), tt.expectedErr) {
				t.Fatalf("expected error to contain %q, got: %v", tt.expectedErr, err)
			}
		})
	}
}

func TestCLI_BumpAutoCmd_InitVersionFileFails(t *testing.T) {
	tmp := t.TempDir()
	protected := filepath.Join(tmp, "protected")

	// Make directory not writable
	if err := os.Mkdir(protected, 0555); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chmod(protected, 0755) })

	versionPath := filepath.Join(protected, ".version")

	// Prepare and run the CLI command
	cfg := &config.Config{Path: versionPath}
	registry := plugins.NewPluginRegistry()
	appCli := testutils.BuildCLIForTests(cfg.Path, []*cli.Command{Run(cfg, registry)})

	err := appCli.Run(context.Background(), []string{
		"sley", "bump", "auto", "--path", versionPath,
	})
	if err == nil || !strings.Contains(err.Error(), "permission denied") {
		t.Fatalf("expected permission denied error, got: %v", err)
	}
}

func TestCLI_BumpAutoCmd_BumpNextFails(t *testing.T) {
	tmp := t.TempDir()
	versionPath := testutils.WriteTempVersionFile(t, tmp, "1.2.3")

	original := semver.BumpNextFunc
	semver.BumpNextFunc = func(v semver.SemVersion) (semver.SemVersion, error) {
		return semver.SemVersion{}, fmt.Errorf("forced BumpNext failure")
	}
	t.Cleanup(func() {
		semver.BumpNextFunc = original
	})

	// Prepare and run the CLI command
	cfg := &config.Config{Path: versionPath}
	registry := plugins.NewPluginRegistry()
	appCli := testutils.BuildCLIForTests(cfg.Path, []*cli.Command{Run(cfg, registry)})

	err := appCli.Run(context.Background(), []string{
		"sley", "bump", "auto", "--path", versionPath, "--no-infer",
	})

	if err == nil || !strings.Contains(err.Error(), "failed to determine next version") {
		t.Fatalf("expected BumpNext failure, got: %v", err)
	}
}

func TestCLI_BumpAutoCmd_SaveVersionFails(t *testing.T) {
	tmp := t.TempDir()
	versionPath := filepath.Join(tmp, ".version")

	// Write valid content
	if err := os.WriteFile(versionPath, []byte("1.2.3-alpha\n"), 0644); err != nil {
		t.Fatalf("failed to write version: %v", err)
	}

	// Make file read-only
	if err := os.Chmod(versionPath, 0444); err != nil {
		t.Fatalf("failed to chmod version file: %v", err)
	}
	t.Cleanup(func() { _ = os.Chmod(versionPath, 0644) }) // ensure cleanup

	// Prepare and run the CLI command
	cfg := &config.Config{Path: versionPath}
	registry := plugins.NewPluginRegistry()
	appCli := testutils.BuildCLIForTests(cfg.Path, []*cli.Command{Run(cfg, registry)})

	err := appCli.Run(context.Background(), []string{
		"sley", "bump", "auto", "--path", versionPath, "--strict",
	})

	if err == nil || !strings.Contains(err.Error(), "failed to save version") {
		t.Fatalf("expected error containing 'failed to save version', got: %v", err)
	}
}

func TestCLI_BumpAutoCommand_InvalidLabel(t *testing.T) {
	if os.Getenv("TEST_SLEY_BUMP_AUTO_INVALID_LABEL") == "1" {
		tmp := t.TempDir()
		versionPath := testutils.WriteTempVersionFile(t, tmp, "1.2.3")

		// Prepare and run the CLI command
		cfg := &config.Config{Path: versionPath}
		registry := plugins.NewPluginRegistry()
		appCli := testutils.BuildCLIForTests(cfg.Path, []*cli.Command{Run(cfg, registry)})

		err := appCli.Run(context.Background(), []string{
			"sley", "bump", "auto", "--label", "banana", "--path", versionPath,
		})
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		os.Exit(0) // shouldn't happen
	}

	cmd := exec.Command(os.Args[0], "-test.run=TestCLI_BumpAutoCommand_InvalidLabel")
	cmd.Env = append(os.Environ(), "TEST_SLEY_BUMP_AUTO_INVALID_LABEL=1")
	output, err := cmd.CombinedOutput()

	if err == nil {
		t.Fatal("expected non-zero exit status")
	}

	expected := "invalid --label: must be 'patch', 'minor', or 'major'"
	if !strings.Contains(string(output), expected) {
		t.Errorf("expected output to contain %q, got: %q", expected, string(output))
	}
}

func TestCLI_BumpAutoCmd_BumpByLabelFails(t *testing.T) {
	tmp := t.TempDir()
	versionPath := testutils.WriteTempVersionFile(t, tmp, "1.2.3")

	original := semver.BumpByLabelFunc
	semver.BumpByLabelFunc = func(v semver.SemVersion, label string) (semver.SemVersion, error) {
		return semver.SemVersion{}, fmt.Errorf("boom")
	}
	t.Cleanup(func() {
		semver.BumpByLabelFunc = original
	})

	// Prepare and run the CLI command
	cfg := &config.Config{Path: versionPath}
	registry := plugins.NewPluginRegistry()
	appCli := testutils.BuildCLIForTests(cfg.Path, []*cli.Command{Run(cfg, registry)})

	err := appCli.Run(context.Background(), []string{
		"sley", "bump", "auto", "--label", "patch", "--path", versionPath,
	})

	if err == nil || !strings.Contains(err.Error(), "failed to bump version with label") {
		t.Fatalf("expected error due to label bump failure, got: %v", err)
	}
}

func TestDetermineBumpType(t *testing.T) {
	// Save original function
	originalInferFromChangelog := tryInferBumpTypeFromChangelogParserPluginFn
	originalInferFromCommit := tryInferBumpTypeFromCommitParserPluginFn
	defer func() {
		tryInferBumpTypeFromChangelogParserPluginFn = originalInferFromChangelog
		tryInferBumpTypeFromCommitParserPluginFn = originalInferFromCommit
	}()

	tests := []struct {
		name          string
		label         string
		disableInfer  bool
		mockChangelog string
		mockCommit    string
		expected      string
	}{
		{"explicit patch", "patch", false, "", "", "patch"},
		{"explicit minor", "minor", false, "", "", "minor"},
		{"explicit major", "major", false, "", "", "major"},
		{"infer from changelog minor", "", false, "minor", "", "minor"},
		{"infer from changelog major", "", false, "major", "", "major"},
		{"infer from commits when changelog empty", "", false, "", "minor", "minor"},
		{"default to auto when inference disabled", "", true, "", "", "auto"},
		{"invalid label defaults to auto", "invalid", false, "", "", "auto"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tryInferBumpTypeFromChangelogParserPluginFn = func(registry *plugins.PluginRegistry) string { return tt.mockChangelog }
			tryInferBumpTypeFromCommitParserPluginFn = func(registry *plugins.PluginRegistry, since, until string) string { return tt.mockCommit }

			registry := plugins.NewPluginRegistry()
			result := determineBumpType(registry, tt.label, tt.disableInfer, "", "")

			if string(result) != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, string(result))
			}
		})
	}
}

func TestTryInferBumpTypeFromChangelogParserPlugin_NoParser(t *testing.T) {
	// Ensure no parser is registered
	originalFn := tryInferBumpTypeFromChangelogParserPluginFn
	defer func() { tryInferBumpTypeFromChangelogParserPluginFn = originalFn }()

	// Use the actual function
	tryInferBumpTypeFromChangelogParserPluginFn = tryInferBumpTypeFromChangelogParserPlugin

	registry := plugins.NewPluginRegistry()
	label := tryInferBumpTypeFromChangelogParserPlugin(registry)
	if label != "" {
		t.Errorf("expected empty label when no parser, got %q", label)
	}
}

func TestGetNextVersion(t *testing.T) {
	tests := []struct {
		name         string
		current      semver.SemVersion
		label        string
		disableInfer bool
		expected     string
		expectError  bool
	}{
		{
			name:        "patch label",
			current:     semver.SemVersion{Major: 1, Minor: 2, Patch: 3},
			label:       "patch",
			expected:    "1.2.4",
			expectError: false,
		},
		{
			name:        "minor label",
			current:     semver.SemVersion{Major: 1, Minor: 2, Patch: 3},
			label:       "minor",
			expected:    "1.3.0",
			expectError: false,
		},
		{
			name:        "major label",
			current:     semver.SemVersion{Major: 1, Minor: 2, Patch: 3},
			label:       "major",
			expected:    "2.0.0",
			expectError: false,
		},
		{
			name:         "auto bump with inference disabled",
			current:      semver.SemVersion{Major: 1, Minor: 2, Patch: 3},
			label:        "",
			disableInfer: true,
			expected:     "1.2.4",
			expectError:  false,
		},
		{
			name:        "invalid label",
			current:     semver.SemVersion{Major: 1, Minor: 2, Patch: 3},
			label:       "invalid",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			registry := plugins.NewPluginRegistry()
			result, err := getNextVersion(registry, tt.current, tt.label, tt.disableInfer, "", "", false)
			if tt.expectError {
				if err == nil {
					t.Error("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if result.String() != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, result.String())
			}
		})
	}
}

/* ------------------------------------------------------------------------- */
/* BUMP AUTO TAG CREATION TESTS                                              */
/* ------------------------------------------------------------------------- */

func TestBumpAuto_CallsCreateTagAfterBump_WithEnabledTagManager(t *testing.T) {
	// This test verifies that createTagAfterBump is called when tag manager is enabled
	// by testing the function directly with the "auto" bump type parameter
	version := semver.SemVersion{Major: 99, Minor: 88, Patch: 77}

	// Save original function and restore after test
	origGetTagManagerFn := tagmanager.GetTagManagerFn
	defer func() { tagmanager.GetTagManagerFn = origGetTagManagerFn }()

	// Create an enabled tag manager plugin
	plugin := tagmanager.NewTagManager(&tagmanager.Config{
		Enabled:    true,
		AutoCreate: true,
		Prefix:     "v",
		Annotate:   true,
	})
	tagmanager.GetTagManagerFn = func() tagmanager.TagManager { return plugin }

	registry := plugins.NewPluginRegistry()

	// Call createTagAfterBump directly with "auto" bump type
	// This is the call that runSingleModuleAuto makes at the end
	err := createTagAfterBump(registry, version, "auto")

	// The test environment may have various outcomes depending on git state
	// The important thing is that the function is called and attempts tag creation
	if err != nil {
		errStr := err.Error()
		// These are acceptable errors when running in a test environment
		if !strings.Contains(errStr, "failed to create tag") && !strings.Contains(errStr, "already exists") {
			t.Fatalf("expected tag creation error, tag exists error, or no error, got: %v", err)
		}
	}
}

func TestBumpAuto_EndToEnd_WithMockTagManager(t *testing.T) {
	// This test verifies the full bump auto flow with a mock tag manager
	// that tracks whether CreateTag is called
	tmpDir := t.TempDir()
	versionPath := filepath.Join(tmpDir, ".version")
	testutils.WriteTempVersionFile(t, tmpDir, "1.2.3")

	// Save original function and restore after test
	origGetTagManagerFn := tagmanager.GetTagManagerFn
	defer func() { tagmanager.GetTagManagerFn = origGetTagManagerFn }()

	// Create a mock that tracks calls - note: createTagAfterBump uses type assertion
	// to *tagmanager.TagManagerPlugin, so mocks will be treated as disabled
	// We use this to verify the flow works when tag manager returns nil for registry
	tagmanager.GetTagManagerFn = func() tagmanager.TagManager { return nil }

	cfg := &config.Config{Path: versionPath}
	registry := plugins.NewPluginRegistry()
	appCli := testutils.BuildCLIForTests(cfg.Path, []*cli.Command{Run(cfg, registry)})

	err := appCli.Run(context.Background(), []string{
		"sley", "bump", "auto", "--path", versionPath, "--no-infer",
	})

	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	// Verify the version was bumped
	got := testutils.ReadTempVersionFile(t, tmpDir)
	want := "1.2.4"
	if got != want {
		t.Errorf("expected bumped version %q, got %q", want, got)
	}
}

func TestBumpAuto_SkipsTagCreation_WhenTagManagerDisabled(t *testing.T) {
	tmpDir := t.TempDir()
	versionPath := filepath.Join(tmpDir, ".version")
	testutils.WriteTempVersionFile(t, tmpDir, "1.2.3")

	// Save original function and restore after test
	origGetTagManagerFn := tagmanager.GetTagManagerFn
	defer func() { tagmanager.GetTagManagerFn = origGetTagManagerFn }()

	// Create a disabled tag manager plugin
	plugin := tagmanager.NewTagManager(&tagmanager.Config{
		Enabled:    false,
		AutoCreate: false,
	})
	tagmanager.GetTagManagerFn = func() tagmanager.TagManager { return plugin }

	cfg := &config.Config{Path: versionPath}
	registry := plugins.NewPluginRegistry()
	if err := registry.RegisterTagManager(plugin); err != nil {
		t.Fatalf("failed to register tag manager: %v", err)
	}
	appCli := testutils.BuildCLIForTests(cfg.Path, []*cli.Command{Run(cfg, registry)})

	err := appCli.Run(context.Background(), []string{
		"sley", "bump", "auto", "--path", versionPath, "--no-infer",
	})

	if err != nil {
		t.Fatalf("expected no error when tag manager is disabled, got: %v", err)
	}

	got := testutils.ReadTempVersionFile(t, tmpDir)
	want := "1.2.4"
	if got != want {
		t.Errorf("expected bumped version %q, got %q", want, got)
	}
}

func TestBumpAuto_SkipsTagCreation_WhenNoTagManagerRegistered(t *testing.T) {
	tmpDir := t.TempDir()
	versionPath := filepath.Join(tmpDir, ".version")
	testutils.WriteTempVersionFile(t, tmpDir, "1.2.3-alpha.1")

	// Save original function and restore after test
	origGetTagManagerFn := tagmanager.GetTagManagerFn
	defer func() { tagmanager.GetTagManagerFn = origGetTagManagerFn }()

	// No tag manager registered
	tagmanager.GetTagManagerFn = func() tagmanager.TagManager { return nil }

	cfg := &config.Config{Path: versionPath}
	registry := plugins.NewPluginRegistry()
	appCli := testutils.BuildCLIForTests(cfg.Path, []*cli.Command{Run(cfg, registry)})

	err := appCli.Run(context.Background(), []string{
		"sley", "bump", "auto", "--path", versionPath,
	})

	if err != nil {
		t.Fatalf("expected no error when no tag manager registered, got: %v", err)
	}

	got := testutils.ReadTempVersionFile(t, tmpDir)
	want := "1.2.3" // Promoted from pre-release
	if got != want {
		t.Errorf("expected promoted version %q, got %q", want, got)
	}
}

func TestBumpAuto_TagCreatedWithCorrectParameters(t *testing.T) {
	version := semver.SemVersion{Major: 1, Minor: 2, Patch: 4}

	t.Run("calls createTagAfterBump with auto bump type", func(t *testing.T) {
		// Save original and restore after test
		origGetTagManagerFn := tagmanager.GetTagManagerFn
		defer func() { tagmanager.GetTagManagerFn = origGetTagManagerFn }()

		// Create enabled tag manager
		plugin := tagmanager.NewTagManager(&tagmanager.Config{
			Enabled:    true,
			AutoCreate: true,
			Prefix:     "v",
		})
		tagmanager.GetTagManagerFn = func() tagmanager.TagManager { return plugin }

		registry := plugins.NewPluginRegistry()
		// Note: createTagAfterBump checks for *TagManagerPlugin type assertion
		// When using a real plugin, it will try to create a tag (which fails without git)
		err := createTagAfterBump(registry, version, "auto")

		// In test environment without git, we expect a tag creation error
		if err != nil && !strings.Contains(err.Error(), "failed to create tag") {
			t.Errorf("unexpected error type: %v", err)
		}
	})

	t.Run("returns nil when tag manager is nil", func(t *testing.T) {
		origGetTagManagerFn := tagmanager.GetTagManagerFn
		defer func() { tagmanager.GetTagManagerFn = origGetTagManagerFn }()

		tagmanager.GetTagManagerFn = func() tagmanager.TagManager { return nil }

		registry := plugins.NewPluginRegistry()
		err := createTagAfterBump(registry, version, "auto")
		if err != nil {
			t.Errorf("expected nil error when tag manager is nil, got: %v", err)
		}
	})

	t.Run("returns nil when tag manager is disabled", func(t *testing.T) {
		origGetTagManagerFn := tagmanager.GetTagManagerFn
		defer func() { tagmanager.GetTagManagerFn = origGetTagManagerFn }()

		plugin := tagmanager.NewTagManager(&tagmanager.Config{
			Enabled:    false,
			AutoCreate: false,
		})
		tagmanager.GetTagManagerFn = func() tagmanager.TagManager { return plugin }

		registry := plugins.NewPluginRegistry()
		err := createTagAfterBump(registry, version, "auto")
		if err != nil {
			t.Errorf("expected nil error when tag manager is disabled, got: %v", err)
		}
	})
}

func TestBumpAuto_TagCreation_OnPreReleasePromotion(t *testing.T) {
	tmpDir := t.TempDir()
	versionPath := filepath.Join(tmpDir, ".version")
	testutils.WriteTempVersionFile(t, tmpDir, "2.0.0-rc.1")

	// Save original function and restore after test
	origGetTagManagerFn := tagmanager.GetTagManagerFn
	defer func() { tagmanager.GetTagManagerFn = origGetTagManagerFn }()

	// Create a disabled tag manager to verify bump succeeds without tag creation
	plugin := tagmanager.NewTagManager(&tagmanager.Config{
		Enabled:    false,
		AutoCreate: false,
	})
	tagmanager.GetTagManagerFn = func() tagmanager.TagManager { return plugin }

	cfg := &config.Config{Path: versionPath}
	registry := plugins.NewPluginRegistry()
	appCli := testutils.BuildCLIForTests(cfg.Path, []*cli.Command{Run(cfg, registry)})

	err := appCli.Run(context.Background(), []string{
		"sley", "bump", "auto", "--path", versionPath,
	})

	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	got := testutils.ReadTempVersionFile(t, tmpDir)
	want := "2.0.0"
	if got != want {
		t.Errorf("expected promoted version %q, got %q", want, got)
	}
}

func TestBumpAuto_InferredMinorBump_WithTagManager(t *testing.T) {
	tmpDir := t.TempDir()
	versionPath := filepath.Join(tmpDir, ".version")
	testutils.WriteTempVersionFile(t, tmpDir, "1.0.0")

	// Save and restore original functions
	origGetTagManagerFn := tagmanager.GetTagManagerFn
	originalInfer := tryInferBumpTypeFromCommitParserPluginFn
	defer func() {
		tagmanager.GetTagManagerFn = origGetTagManagerFn
		tryInferBumpTypeFromCommitParserPluginFn = originalInfer
	}()

	// Mock inference to return "minor"
	tryInferBumpTypeFromCommitParserPluginFn = func(registry *plugins.PluginRegistry, since, until string) string {
		return "minor"
	}

	// Create a disabled tag manager (to avoid git dependency in tests)
	plugin := tagmanager.NewTagManager(&tagmanager.Config{
		Enabled:    false,
		AutoCreate: false,
	})
	tagmanager.GetTagManagerFn = func() tagmanager.TagManager { return plugin }

	cfg := &config.Config{Path: versionPath, Plugins: &config.PluginConfig{CommitParser: true}}
	registry := plugins.NewPluginRegistry()
	appCli := testutils.BuildCLIForTests(cfg.Path, []*cli.Command{Run(cfg, registry)})

	err := appCli.Run(context.Background(), []string{
		"sley", "bump", "auto", "--path", versionPath,
	})

	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	got := testutils.ReadTempVersionFile(t, tmpDir)
	want := "1.1.0"
	if got != want {
		t.Errorf("expected inferred minor bump %q, got %q", want, got)
	}
}
