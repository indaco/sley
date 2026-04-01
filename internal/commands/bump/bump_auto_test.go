package bump

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
	"github.com/indaco/sley/internal/plugins/tagmanager"
	"github.com/indaco/sley/internal/semver"
	"github.com/indaco/sley/internal/testutils"
	"github.com/urfave/cli/v3"
)

// testContext returns a context with custom bumpDeps for testing.
func testContext(deps *bumpDeps) context.Context {
	return context.WithValue(context.Background(), bumpDepsKey{}, deps)
}

// defaultTestDeps returns a bumpDeps with real defaults (no mocking).
func defaultTestDeps() *bumpDeps {
	return newBumpDeps()
}

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

	deps := defaultTestDeps()
	deps.inferFromCommits = func(registry *plugins.PluginRegistry, since, until string) string {
		return "minor"
	}
	ctx := testContext(deps)

	cfg := &config.Config{Path: versionPath, Plugins: &config.PluginConfig{CommitParser: true}}
	registry := plugins.NewPluginRegistry()
	appCli := testutils.BuildCLIForTests(cfg.Path, []*cli.Command{Run(cfg, registry)})

	err := appCli.Run(ctx, []string{
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

	deps := defaultTestDeps()
	deps.inferFromCommits = func(registry *plugins.PluginRegistry, since, until string) string {
		return "minor"
	}
	ctx := testContext(deps)

	cfg := &config.Config{Path: versionPath, Plugins: &config.PluginConfig{CommitParser: true}}
	registry := plugins.NewPluginRegistry()
	appCli := testutils.BuildCLIForTests(cfg.Path, []*cli.Command{Run(cfg, registry)})

	err := appCli.Run(ctx, []string{
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

	deps := defaultTestDeps()
	deps.inferFromCommits = func(registry *plugins.PluginRegistry, since, until string) string {
		return "minor" // Force a non-empty inference so that promotePreRelease is called
	}
	ctx := testContext(deps)

	cfg := &config.Config{Path: versionPath}
	registry := plugins.NewPluginRegistry()
	appCli := testutils.BuildCLIForTests(cfg.Path, []*cli.Command{Run(cfg, registry)})

	err := appCli.Run(ctx, []string{
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

// mockBumper is a VersionBumper for testing that returns configurable results.
type mockBumper struct {
	bumpNextErr    error
	bumpByLabelErr error
}

func (m mockBumper) BumpNext(v semver.SemVersion) (semver.SemVersion, error) {
	if m.bumpNextErr != nil {
		return semver.SemVersion{}, m.bumpNextErr
	}
	return semver.BumpNext(v)
}

func (m mockBumper) BumpByLabel(v semver.SemVersion, label string) (semver.SemVersion, error) {
	if m.bumpByLabelErr != nil {
		return semver.SemVersion{}, m.bumpByLabelErr
	}
	return semver.BumpByLabel(v, label)
}

func TestCLI_BumpAutoCmd_InferredBumpFails(t *testing.T) {
	tmp := t.TempDir()
	versionPath := testutils.WriteTempVersionFile(t, tmp, "1.2.3")

	deps := defaultTestDeps()
	deps.newBumper = func() semver.VersionBumper {
		return mockBumper{bumpByLabelErr: fmt.Errorf("forced inferred bump failure")}
	}
	deps.inferFromCommits = func(registry *plugins.PluginRegistry, since, until string) string {
		return "minor"
	}
	ctx := testContext(deps)

	// Prepare and run CLI
	cfg := &config.Config{Path: versionPath}
	registry := plugins.NewPluginRegistry()
	appCli := testutils.BuildCLIForTests(cfg.Path, []*cli.Command{Run(cfg, registry)})

	err := appCli.Run(ctx, []string{
		"sley", "bump", "auto", "--path", versionPath,
	})

	if err == nil || !strings.Contains(err.Error(), "failed to bump inferred version") {
		t.Fatalf("expected error about inferred bump failure, got: %v", err)
	}
}

func TestTryInferBumpTypeFromCommitParserPlugin_GetCommitsError(t *testing.T) {
	// tryInferBumpTypeFromCommitParserPlugin now uses gitlog.NewGitLog() internally,
	// so we test it by registering a mock commit parser and verifying the function
	// returns empty string when git operations fail (no real git repo in test dir).
	registry := plugins.NewPluginRegistry()
	parser := testutils.MockCommitParser{}
	if err := registry.RegisterCommitParser(&parser); err != nil {
		t.Fatalf("failed to register parser: %v", err)
	}
	// Without a real git repo, getCommits will fail → should return ""
	label := tryInferBumpTypeFromCommitParserPlugin(registry, "", "")
	if label != "" {
		t.Errorf("expected empty label on gitlog error, got %q", label)
	}
}

func TestTryInferBumpTypeFromCommitParserPlugin_ParserError(t *testing.T) {
	// Without a real git repo, this will fail at the git level, returning ""
	registry := plugins.NewPluginRegistry()
	parser := testutils.MockCommitParser{Err: fmt.Errorf("parser error")}
	if err := registry.RegisterCommitParser(&parser); err != nil {
		t.Fatalf("failed to register parser: %v", err)
	}
	label := tryInferBumpTypeFromCommitParserPlugin(registry, "", "")
	if label != "" {
		t.Errorf("expected empty label on error, got %q", label)
	}
}

func TestTryInferBumpTypeFromCommitParserPlugin_NoParser(t *testing.T) {
	registry := plugins.NewPluginRegistry()
	label := tryInferBumpTypeFromCommitParserPlugin(registry, "", "")
	if label != "" {
		t.Errorf("expected empty label when no parser, got %q", label)
	}
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

	deps := defaultTestDeps()
	deps.newBumper = func() semver.VersionBumper {
		return mockBumper{bumpNextErr: fmt.Errorf("forced BumpNext failure")}
	}
	ctx := testContext(deps)

	// Prepare and run the CLI command
	cfg := &config.Config{Path: versionPath}
	registry := plugins.NewPluginRegistry()
	appCli := testutils.BuildCLIForTests(cfg.Path, []*cli.Command{Run(cfg, registry)})

	err := appCli.Run(ctx, []string{
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

	cmd := exec.Command(os.Args[0], "-test.run=TestCLI_BumpAutoCommand_InvalidLabel") //nolint:gosec // G702: standard test re-exec pattern
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

	deps := defaultTestDeps()
	deps.newBumper = func() semver.VersionBumper {
		return mockBumper{bumpByLabelErr: fmt.Errorf("boom")}
	}
	ctx := testContext(deps)

	// Prepare and run the CLI command
	cfg := &config.Config{Path: versionPath}
	registry := plugins.NewPluginRegistry()
	appCli := testutils.BuildCLIForTests(cfg.Path, []*cli.Command{Run(cfg, registry)})

	err := appCli.Run(ctx, []string{
		"sley", "bump", "auto", "--label", "patch", "--path", versionPath,
	})

	if err == nil || !strings.Contains(err.Error(), "failed to bump version with label") {
		t.Fatalf("expected error due to label bump failure, got: %v", err)
	}
}

func TestDetermineBumpType(t *testing.T) {
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
			deps := &bumpDeps{
				inferFromChangelog: func(registry *plugins.PluginRegistry) string { return tt.mockChangelog },
				inferFromCommits:   func(registry *plugins.PluginRegistry, since, until string) string { return tt.mockCommit },
			}

			registry := plugins.NewPluginRegistry()
			result := determineBumpType(deps, registry, tt.label, tt.disableInfer, "", "")

			if string(result) != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, string(result))
			}
		})
	}
}

func TestTryInferBumpTypeFromChangelogParserPlugin_NoParser(t *testing.T) {
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
			deps := defaultTestDeps()
			registry := plugins.NewPluginRegistry()
			bumper := semver.NewDefaultBumper()
			result, err := getNextVersion(deps, bumper, registry, tt.current, tt.label, tt.disableInfer, "", "", false)
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
	version := semver.SemVersion{Major: 99, Minor: 88, Patch: 77}

	// Create an enabled tag manager plugin with mock git operations
	// to avoid creating real tags in the working repository.
	mockGitOps := &tagmanager.MockGitTagOperations{}
	mockCommitOps := &tagmanager.MockGitCommitOperations{}
	plugin := tagmanager.NewTagManagerWithOps(&tagmanager.Config{
		Enabled:    true,
		AutoCreate: true,
		Prefix:     "v",
		Annotate:   true,
	}, mockGitOps, mockCommitOps)

	registry := plugins.NewPluginRegistry()
	if err := registry.RegisterTagManager(plugin); err != nil {
		t.Fatalf("failed to register tag manager: %v", err)
	}

	err := createTagAfterBump(registry, version, "auto", nil)

	if err != nil {
		errStr := err.Error()
		if !strings.Contains(errStr, "failed to create tag") && !strings.Contains(errStr, "already exists") && !strings.Contains(errStr, "failed to commit") {
			t.Fatalf("expected tag creation error, tag exists error, commit error, or no error, got: %v", err)
		}
	}
}

func TestBumpAuto_EndToEnd_WithMockTagManager(t *testing.T) {
	tmpDir := t.TempDir()
	versionPath := filepath.Join(tmpDir, ".version")
	testutils.WriteTempVersionFile(t, tmpDir, "1.2.3")

	cfg := &config.Config{Path: versionPath}
	registry := plugins.NewPluginRegistry()
	appCli := testutils.BuildCLIForTests(cfg.Path, []*cli.Command{Run(cfg, registry)})

	err := appCli.Run(context.Background(), []string{
		"sley", "bump", "auto", "--path", versionPath, "--no-infer",
	})

	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

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

	plugin := tagmanager.NewTagManagerWithOps(&tagmanager.Config{
		Enabled:    false,
		AutoCreate: false,
	}, &tagmanager.MockGitTagOperations{}, &tagmanager.MockGitCommitOperations{})

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
		// Use mock git operations to avoid creating real tags.
		mockGitOps := &tagmanager.MockGitTagOperations{}
		mockCommitOps := &tagmanager.MockGitCommitOperations{}
		plugin := tagmanager.NewTagManagerWithOps(&tagmanager.Config{
			Enabled:    true,
			AutoCreate: true,
			Prefix:     "v",
		}, mockGitOps, mockCommitOps)

		registry := plugins.NewPluginRegistry()
		if err := registry.RegisterTagManager(plugin); err != nil {
			t.Fatalf("failed to register tag manager: %v", err)
		}
		err := createTagAfterBump(registry, version, "auto", nil)

		if err != nil && !strings.Contains(err.Error(), "failed to create tag") && !strings.Contains(err.Error(), "failed to commit") {
			t.Errorf("unexpected error type: %v", err)
		}
	})

	t.Run("returns nil when tag manager is nil", func(t *testing.T) {
		registry := plugins.NewPluginRegistry()
		err := createTagAfterBump(registry, version, "auto", nil)
		if err != nil {
			t.Errorf("expected nil error when tag manager is nil, got: %v", err)
		}
	})

	t.Run("returns nil when tag manager is disabled", func(t *testing.T) {
		plugin := tagmanager.NewTagManagerWithOps(&tagmanager.Config{
			Enabled:    false,
			AutoCreate: false,
		}, &tagmanager.MockGitTagOperations{}, &tagmanager.MockGitCommitOperations{})

		registry := plugins.NewPluginRegistry()
		if err := registry.RegisterTagManager(plugin); err != nil {
			t.Fatalf("failed to register tag manager: %v", err)
		}
		err := createTagAfterBump(registry, version, "auto", nil)
		if err != nil {
			t.Errorf("expected nil error when tag manager is disabled, got: %v", err)
		}
	})
}

func TestBumpAuto_TagCreation_OnPreReleasePromotion(t *testing.T) {
	tmpDir := t.TempDir()
	versionPath := filepath.Join(tmpDir, ".version")
	testutils.WriteTempVersionFile(t, tmpDir, "2.0.0-rc.1")

	plugin := tagmanager.NewTagManagerWithOps(&tagmanager.Config{
		Enabled:    false,
		AutoCreate: false,
	}, &tagmanager.MockGitTagOperations{}, &tagmanager.MockGitCommitOperations{})

	cfg := &config.Config{Path: versionPath}
	registry := plugins.NewPluginRegistry()
	if err := registry.RegisterTagManager(plugin); err != nil {
		t.Fatalf("failed to register tag manager: %v", err)
	}
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

	deps := defaultTestDeps()
	deps.inferFromCommits = func(registry *plugins.PluginRegistry, since, until string) string {
		return "minor"
	}
	ctx := testContext(deps)

	plugin := tagmanager.NewTagManagerWithOps(&tagmanager.Config{
		Enabled:    false,
		AutoCreate: false,
	}, &tagmanager.MockGitTagOperations{}, &tagmanager.MockGitCommitOperations{})

	cfg := &config.Config{Path: versionPath, Plugins: &config.PluginConfig{CommitParser: true}}
	registry := plugins.NewPluginRegistry()
	if err := registry.RegisterTagManager(plugin); err != nil {
		t.Fatalf("failed to register tag manager: %v", err)
	}
	appCli := testutils.BuildCLIForTests(cfg.Path, []*cli.Command{Run(cfg, registry)})

	err := appCli.Run(ctx, []string{
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
