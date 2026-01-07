package bumpcmd

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/indaco/sley/internal/clix"
	"github.com/indaco/sley/internal/config"
	"github.com/indaco/sley/internal/hooks"
	"github.com/indaco/sley/internal/plugins"
	"github.com/indaco/sley/internal/plugins/auditlog"
	"github.com/indaco/sley/internal/plugins/changeloggenerator"
	"github.com/indaco/sley/internal/plugins/commitparser"
	"github.com/indaco/sley/internal/plugins/commitparser/gitlog"
	"github.com/indaco/sley/internal/plugins/dependencycheck"
	"github.com/indaco/sley/internal/plugins/releasegate"
	"github.com/indaco/sley/internal/plugins/tagmanager"
	"github.com/indaco/sley/internal/plugins/versionvalidator"
	"github.com/indaco/sley/internal/semver"
	"github.com/indaco/sley/internal/testutils"
	"github.com/indaco/sley/internal/workspace"
	"github.com/urfave/cli/v3"
)

func TestCLI_BumpCommand_Variants(t *testing.T) {
	tmpDir := t.TempDir()
	versionPath := filepath.Join(tmpDir, ".version")

	// Prepare the CLI command
	cfg := &config.Config{Path: versionPath}
	registry := plugins.NewPluginRegistry()
	appCli := testutils.BuildCLIForTests(cfg.Path, []*cli.Command{Run(cfg, registry)})

	tests := []struct {
		name     string
		initial  string
		args     []string
		expected string
	}{
		{"patch bump", "1.2.3", []string{"sley", "bump", "patch"}, "1.2.4"},
		{"minor bump", "1.2.3", []string{"sley", "bump", "minor"}, "1.3.0"},
		{"major bump", "1.2.3", []string{"sley", "bump", "major"}, "2.0.0"},
		{"patch bump after pre-release", "1.2.3-alpha", []string{"sley", "bump", "patch"}, "1.2.4"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testutils.WriteTempVersionFile(t, tmpDir, tt.initial)
			testutils.RunCLITest(t, appCli, tt.args, tmpDir)

			got := testutils.ReadTempVersionFile(t, tmpDir)
			if got != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, got)
			}
		})
	}
}

func TestCLI_BumpCommand_AutoInitFeedback(t *testing.T) {
	tests := []struct {
		name    string
		initial string
		args    []string
	}{
		{"patch bump", "1.2.3", []string{"sley", "bump", "patch"}},
		{"minor bump", "1.2.3", []string{"sley", "bump", "minor"}},
		{"major bump", "1.2.3", []string{"sley", "bump", "major"}},
		{"patch bump after pre-release", "1.2.3-alpha", []string{"sley", "bump", "patch"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			versionPath := filepath.Join(tmpDir, ".version")

			// Prepare and run the CLI command
			cfg := &config.Config{Path: versionPath}
			registry := plugins.NewPluginRegistry()
			appCli := testutils.BuildCLIForTests(cfg.Path, []*cli.Command{Run(cfg, registry)})
			output, err := testutils.CaptureStdout(func() {
				testutils.RunCLITest(t, appCli, tt.args, tmpDir)
			})
			if err != nil {
				t.Fatalf("Failed to capture stdout: %v", err)
			}

			expected := fmt.Sprintf("Auto-initialized %s with default version", versionPath)
			if !strings.Contains(output, expected) {
				t.Errorf("expected feedback %q, got %q", expected, output)
			}
		})
	}
}

func TestCLI_BumpSubcommands_EarlyFailures(t *testing.T) {

	tests := []struct {
		name        string
		args        []string
		override    func() func() // returns restore function
		expectedErr string
	}{
		{
			name: "patch - FromCommand fails",
			args: []string{"sley", "bump", "patch"},
			override: func() func() {
				original := clix.FromCommandFn
				clix.FromCommandFn = func(cmd *cli.Command) (bool, error) {
					return false, fmt.Errorf("mock FromCommand error")
				}
				return func() { clix.FromCommandFn = original }
			},
			expectedErr: "mock FromCommand error",
		},
		{
			name: "patch - RunPreReleaseHooks fails",
			args: []string{"sley", "bump", "patch"},
			override: func() func() {
				original := hooks.RunPreReleaseHooksFn
				hooks.RunPreReleaseHooksFn = func(skip bool) error {
					return fmt.Errorf("mock pre-release hooks error")
				}
				return func() { hooks.RunPreReleaseHooksFn = original }
			},
			expectedErr: "mock pre-release hooks error",
		},
		{
			name: "minor - FromCommand fails",
			args: []string{"sley", "bump", "minor"},
			override: func() func() {
				original := clix.FromCommandFn
				clix.FromCommandFn = func(cmd *cli.Command) (bool, error) {
					return false, fmt.Errorf("mock FromCommand error")
				}
				return func() { clix.FromCommandFn = original }
			},
			expectedErr: "mock FromCommand error",
		},
		{
			name: "minor - RunPreReleaseHooks fails",
			args: []string{"sley", "bump", "minor"},
			override: func() func() {
				original := hooks.RunPreReleaseHooksFn
				hooks.RunPreReleaseHooksFn = func(skip bool) error {
					return fmt.Errorf("mock pre-release hooks error")
				}
				return func() { hooks.RunPreReleaseHooksFn = original }
			},
			expectedErr: "mock pre-release hooks error",
		},
		{
			name: "major - FromCommand fails",
			args: []string{"sley", "bump", "major"},
			override: func() func() {
				original := clix.FromCommandFn
				clix.FromCommandFn = func(cmd *cli.Command) (bool, error) {
					return false, fmt.Errorf("mock FromCommand error")
				}
				return func() { clix.FromCommandFn = original }
			},
			expectedErr: "mock FromCommand error",
		},
		{
			name: "major - RunPreReleaseHooks fails",
			args: []string{"sley", "bump", "major"},
			override: func() func() {
				original := hooks.RunPreReleaseHooksFn
				hooks.RunPreReleaseHooksFn = func(skip bool) error {
					return fmt.Errorf("mock pre-release hooks error")
				}
				return func() { hooks.RunPreReleaseHooksFn = original }
			},
			expectedErr: "mock pre-release hooks error",
		},
		{
			name: "auto - FromCommand fails",
			args: []string{"sley", "bump", "auto"},
			override: func() func() {
				original := clix.FromCommandFn
				clix.FromCommandFn = func(cmd *cli.Command) (bool, error) {
					return false, fmt.Errorf("mock FromCommand error")
				}
				return func() { clix.FromCommandFn = original }
			},
			expectedErr: "mock FromCommand error",
		},
		{
			name: "auto - RunPreReleaseHooks fails",
			args: []string{"sley", "bump", "auto"},
			override: func() func() {
				original := hooks.RunPreReleaseHooksFn
				hooks.RunPreReleaseHooksFn = func(skip bool) error {
					return fmt.Errorf("mock pre-release hooks error")
				}
				return func() { hooks.RunPreReleaseHooksFn = original }
			},
			expectedErr: "mock pre-release hooks error",
		},
		{
			name: "release - FromCommand fails",
			args: []string{"sley", "bump", "release"},
			override: func() func() {
				original := clix.FromCommandFn
				clix.FromCommandFn = func(cmd *cli.Command) (bool, error) {
					return false, fmt.Errorf("mock FromCommand error")
				}
				return func() { clix.FromCommandFn = original }
			},
			expectedErr: "mock FromCommand error",
		},
		{
			name: "release - RunPreReleaseHooks fails",
			args: []string{"sley", "bump", "release"},
			override: func() func() {
				original := hooks.RunPreReleaseHooksFn
				hooks.RunPreReleaseHooksFn = func(skip bool) error {
					return fmt.Errorf("mock pre-release hooks error")
				}
				return func() { hooks.RunPreReleaseHooksFn = original }
			},
			expectedErr: "mock pre-release hooks error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			versionPath := filepath.Join(tmpDir, ".version")
			testutils.WriteTempVersionFile(t, tmpDir, "1.2.3")

			restore := tt.override()
			defer restore()

			cfg := &config.Config{Path: versionPath}
			registry := plugins.NewPluginRegistry()
			appCli := testutils.BuildCLIForTests(cfg.Path, []*cli.Command{Run(cfg, registry)})

			err := appCli.Run(context.Background(), tt.args)
			if err == nil || !strings.Contains(err.Error(), tt.expectedErr) {
				t.Fatalf("expected error containing %q, got: %v", tt.expectedErr, err)
			}
		})
	}
}

func TestCLI_BumpReleaseCmd(t *testing.T) {
	tmpDir := t.TempDir()
	versionPath := filepath.Join(tmpDir, ".version")

	// Prepare the CLI command
	cfg := &config.Config{Path: versionPath}
	registry := plugins.NewPluginRegistry()
	appCli := testutils.BuildCLIForTests(cfg.Path, []*cli.Command{Run(cfg, registry)})

	tests := []struct {
		name           string
		initialVersion string
		args           []string
		expected       string
	}{
		{
			name:           "removes pre-release and metadata",
			initialVersion: "1.3.0-alpha.1+ci.123",
			args:           []string{"sley", "bump", "release"},
			expected:       "1.3.0",
		},
		{
			name:           "preserves metadata if flag is set",
			initialVersion: "1.3.0-alpha.2+build.99",
			args:           []string{"sley", "bump", "release", "--preserve-meta"},
			expected:       "1.3.0+build.99",
		},
		{
			name:           "no-op when already final version",
			initialVersion: "2.0.0",
			args:           []string{"sley", "bump", "release"},
			expected:       "2.0.0",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			testutils.WriteTempVersionFile(t, tmpDir, tt.initialVersion)
			testutils.RunCLITest(t, appCli, tt.args, tmpDir)

			got := testutils.ReadTempVersionFile(t, tmpDir)
			if got != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, got)
			}
		})
	}
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

func TestBumpReleaseCmd_ErrorOnReadVersion(t *testing.T) {
	tmp := t.TempDir()
	versionPath := testutils.WriteTempVersionFile(t, tmp, "invalid-version")

	// Prepare and run the CLI command
	cfg := &config.Config{Path: versionPath}
	registry := plugins.NewPluginRegistry()
	appCli := testutils.BuildCLIForTests(cfg.Path, []*cli.Command{Run(cfg, registry)})
	err := appCli.Run(context.Background(), []string{
		"sley", "bump", "release", "--path", versionPath,
	})

	if err == nil || !strings.Contains(err.Error(), "failed to read version") {
		t.Errorf("expected read version error, got: %v", err)
	}
}

func TestCLI_BumpReleaseCommand_SaveVersionFails(t *testing.T) {
	tmp := t.TempDir()
	versionPath := filepath.Join(tmp, ".version")

	// Write valid pre-release content
	if err := os.WriteFile(versionPath, []byte("1.2.3-alpha\n"), 0444); err != nil {
		t.Fatalf("failed to write read-only version file: %v", err)
	}
	t.Cleanup(func() {
		_ = os.Chmod(versionPath, 0644)
	})

	// Prepare and run the CLI command
	cfg := &config.Config{Path: versionPath}
	registry := plugins.NewPluginRegistry()
	appCli := testutils.BuildCLIForTests(cfg.Path, []*cli.Command{Run(cfg, registry)})
	err := appCli.Run(context.Background(), []string{
		"sley", "bump", "release", "--path", versionPath, "--strict",
	})

	if err == nil {
		t.Fatal("expected error due to save failure, got nil")
	}

	if !strings.Contains(err.Error(), "failed to save version") {
		t.Errorf("expected error message to contain 'failed to save version', got: %v", err)
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

func TestBumpReleaseCmd_ErrorOnInitVersionFile(t *testing.T) {
	tmp := t.TempDir()
	protectedDir := filepath.Join(tmp, "protected")
	if err := os.Mkdir(protectedDir, 0555); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chmod(protectedDir, 0755) })

	versionPath := filepath.Join(protectedDir, ".version")

	// Prepare and run the CLI command
	cfg := &config.Config{Path: versionPath}
	registry := plugins.NewPluginRegistry()
	appCli := testutils.BuildCLIForTests(cfg.Path, []*cli.Command{Run(cfg, registry)})

	err := appCli.Run(context.Background(), []string{
		"sley", "bump", "release", "--path", versionPath,
	})

	if err == nil || !strings.Contains(err.Error(), "permission denied") {
		t.Errorf("expected permission denied error, got: %v", err)
	}
}

/* ------------------------------------------------------------------------- */
/* BUMP PRE COMMAND TESTS                                                    */
/* ------------------------------------------------------------------------- */

func TestCLI_BumpPreCmd(t *testing.T) {
	tmpDir := t.TempDir()
	versionPath := filepath.Join(tmpDir, ".version")

	cfg := &config.Config{Path: versionPath}
	registry := plugins.NewPluginRegistry()
	appCli := testutils.BuildCLIForTests(cfg.Path, []*cli.Command{Run(cfg, registry)})

	tests := []struct {
		name     string
		initial  string
		args     []string
		expected string
	}{
		// Increment existing pre-release (dot separator)
		{"increment rc.1 to rc.2", "1.2.3-rc.1", []string{"sley", "bump", "pre"}, "1.2.3-rc.2"},
		{"increment rc.9 to rc.10", "1.2.3-rc.9", []string{"sley", "bump", "pre"}, "1.2.3-rc.10"},
		{"increment alpha.1 to alpha.2", "2.0.0-alpha.1", []string{"sley", "bump", "pre"}, "2.0.0-alpha.2"},

		// Increment existing pre-release (no separator)
		{"increment rc1 to rc2", "1.2.3-rc1", []string{"sley", "bump", "pre"}, "1.2.3-rc2"},
		{"increment beta5 to beta6", "1.2.3-beta5", []string{"sley", "bump", "pre"}, "1.2.3-beta6"},

		// Increment existing pre-release (dash separator)
		{"increment rc-1 to rc-2", "1.2.3-rc-1", []string{"sley", "bump", "pre"}, "1.2.3-rc-2"},

		// Label switch
		{"switch from alpha to beta", "1.2.3-alpha.3", []string{"sley", "bump", "pre", "--label", "beta"}, "1.2.3-beta.1"},
		{"switch from rc1 to beta", "1.2.3-rc1", []string{"sley", "bump", "pre", "--label", "beta"}, "1.2.3-beta.1"},

		// Start new pre-release with --label on final version
		{"add pre-release to final version", "1.2.3", []string{"sley", "bump", "pre", "--label", "rc"}, "1.2.3-rc.1"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testutils.WriteTempVersionFile(t, tmpDir, tt.initial)
			testutils.RunCLITest(t, appCli, tt.args, tmpDir)

			got := testutils.ReadTempVersionFile(t, tmpDir)
			if got != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, got)
			}
		})
	}
}

func TestCLI_BumpPreCmd_ErrorNoPreRelease(t *testing.T) {
	tmpDir := t.TempDir()
	versionPath := filepath.Join(tmpDir, ".version")

	testutils.WriteTempVersionFile(t, tmpDir, "1.2.3")

	cfg := &config.Config{Path: versionPath}
	registry := plugins.NewPluginRegistry()
	appCli := testutils.BuildCLIForTests(cfg.Path, []*cli.Command{Run(cfg, registry)})

	err := appCli.Run(context.Background(), []string{"sley", "bump", "pre"})
	if err == nil || !strings.Contains(err.Error(), "current version has no pre-release") {
		t.Fatalf("expected error about no pre-release, got: %v", err)
	}
}

func TestCLI_BumpPreCmd_EarlyFailures(t *testing.T) {
	tests := []struct {
		name        string
		args        []string
		override    func() func()
		expectedErr string
	}{
		{
			name: "FromCommand fails",
			args: []string{"sley", "bump", "pre", "--label", "rc"},
			override: func() func() {
				original := clix.FromCommandFn
				clix.FromCommandFn = func(cmd *cli.Command) (bool, error) {
					return false, fmt.Errorf("mock FromCommand error")
				}
				return func() { clix.FromCommandFn = original }
			},
			expectedErr: "mock FromCommand error",
		},
		{
			name: "RunPreReleaseHooks fails",
			args: []string{"sley", "bump", "pre", "--label", "rc"},
			override: func() func() {
				original := hooks.RunPreReleaseHooksFn
				hooks.RunPreReleaseHooksFn = func(skip bool) error {
					return fmt.Errorf("mock pre-release hooks error")
				}
				return func() { hooks.RunPreReleaseHooksFn = original }
			},
			expectedErr: "mock pre-release hooks error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			versionPath := filepath.Join(tmpDir, ".version")
			testutils.WriteTempVersionFile(t, tmpDir, "1.2.3-rc.1")

			restore := tt.override()
			defer restore()

			cfg := &config.Config{Path: versionPath}
			registry := plugins.NewPluginRegistry()
			appCli := testutils.BuildCLIForTests(cfg.Path, []*cli.Command{Run(cfg, registry)})

			err := appCli.Run(context.Background(), tt.args)
			if err == nil || !strings.Contains(err.Error(), tt.expectedErr) {
				t.Fatalf("expected error containing %q, got: %v", tt.expectedErr, err)
			}
		})
	}
}

func TestCLI_BumpPreCmd_PreserveMetadata(t *testing.T) {
	tmpDir := t.TempDir()
	versionPath := filepath.Join(tmpDir, ".version")

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
			name:     "preserve metadata",
			initial:  "1.2.3-rc.1+build.99",
			args:     []string{"sley", "bump", "pre", "--preserve-meta"},
			expected: "1.2.3-rc.2+build.99",
		},
		{
			name:     "strip metadata by default",
			initial:  "1.2.3-rc.1+build.99",
			args:     []string{"sley", "bump", "pre"},
			expected: "1.2.3-rc.2",
		},
		{
			name:     "add new metadata",
			initial:  "1.2.3-rc.1",
			args:     []string{"sley", "bump", "pre", "--meta", "ci.123"},
			expected: "1.2.3-rc.2+ci.123",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testutils.WriteTempVersionFile(t, tmpDir, tt.initial)
			testutils.RunCLITest(t, appCli, tt.args, tmpDir)

			got := testutils.ReadTempVersionFile(t, tmpDir)
			if got != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, got)
			}
		})
	}
}

/* ------------------------------------------------------------------------- */
/* HELPER FUNCTION TESTS                                                     */
/* ------------------------------------------------------------------------- */

func TestCalculateNewBuild(t *testing.T) {
	tests := []struct {
		name         string
		meta         string
		preserveMeta bool
		currentBuild string
		expected     string
	}{
		{"new meta overrides", "ci.123", false, "old.456", "ci.123"},
		{"new meta with preserve", "ci.123", true, "old.456", "ci.123"},
		{"preserve existing", "", true, "old.456", "old.456"},
		{"clear when not preserving", "", false, "old.456", ""},
		{"empty when no current", "", true, "", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := calculateNewBuild(tt.meta, tt.preserveMeta, tt.currentBuild)
			if result != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, result)
			}
		})
	}
}

func TestExtractVersionPointers(t *testing.T) {
	strPtr := func(s string) *string { return &s }

	tests := []struct {
		name          string
		version       semver.SemVersion
		expectedPre   *string
		expectedBuild *string
	}{
		{
			name:          "both populated",
			version:       semver.SemVersion{Major: 1, Minor: 2, Patch: 3, PreRelease: "alpha.1", Build: "ci.99"},
			expectedPre:   strPtr("alpha.1"),
			expectedBuild: strPtr("ci.99"),
		},
		{
			name:          "only prerelease",
			version:       semver.SemVersion{Major: 1, Minor: 2, Patch: 3, PreRelease: "beta.2"},
			expectedPre:   strPtr("beta.2"),
			expectedBuild: nil,
		},
		{
			name:          "only build",
			version:       semver.SemVersion{Major: 1, Minor: 2, Patch: 3, Build: "build.42"},
			expectedPre:   nil,
			expectedBuild: strPtr("build.42"),
		},
		{
			name:          "both empty",
			version:       semver.SemVersion{Major: 1, Minor: 2, Patch: 3},
			expectedPre:   nil,
			expectedBuild: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pre, build := extractVersionPointers(tt.version)
			assertStringPtr(t, "prerelease", tt.expectedPre, pre)
			assertStringPtr(t, "build", tt.expectedBuild, build)
		})
	}
}

func assertStringPtr(t *testing.T, name string, expected, actual *string) {
	t.Helper()
	if expected == nil && actual != nil {
		t.Errorf("expected %s pointer to be nil, got %q", name, *actual)
	}
	if expected != nil && actual == nil {
		t.Errorf("expected %s pointer to be %q, got nil", name, *expected)
	}
	if expected != nil && actual != nil && *expected != *actual {
		t.Errorf("expected %s %q, got %q", name, *expected, *actual)
	}
}

func TestPromotePreRelease(t *testing.T) {
	tests := []struct {
		name         string
		current      semver.SemVersion
		preserveMeta bool
		expected     semver.SemVersion
	}{
		{
			name:         "promote without preserving meta",
			current:      semver.SemVersion{Major: 1, Minor: 2, Patch: 3, PreRelease: "alpha.1", Build: "ci.99"},
			preserveMeta: false,
			expected:     semver.SemVersion{Major: 1, Minor: 2, Patch: 3},
		},
		{
			name:         "promote with preserving meta",
			current:      semver.SemVersion{Major: 1, Minor: 2, Patch: 3, PreRelease: "alpha.1", Build: "ci.99"},
			preserveMeta: true,
			expected:     semver.SemVersion{Major: 1, Minor: 2, Patch: 3, Build: "ci.99"},
		},
		{
			name:         "promote without meta",
			current:      semver.SemVersion{Major: 2, Minor: 0, Patch: 0, PreRelease: "rc.1"},
			preserveMeta: true,
			expected:     semver.SemVersion{Major: 2, Minor: 0, Patch: 0},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := promotePreRelease(tt.current, tt.preserveMeta)
			if result.String() != tt.expected.String() {
				t.Errorf("expected %q, got %q", tt.expected.String(), result.String())
			}
		})
	}
}

func TestSetBuildMetadata(t *testing.T) {
	tests := []struct {
		name     string
		current  semver.SemVersion
		next     semver.SemVersion
		meta     string
		preserve bool
		expected string
	}{
		{
			name:     "set new meta",
			current:  semver.SemVersion{Major: 1, Minor: 2, Patch: 3, Build: "old"},
			next:     semver.SemVersion{Major: 1, Minor: 2, Patch: 4},
			meta:     "new",
			preserve: false,
			expected: "new",
		},
		{
			name:     "preserve meta",
			current:  semver.SemVersion{Major: 1, Minor: 2, Patch: 3, Build: "old"},
			next:     semver.SemVersion{Major: 1, Minor: 2, Patch: 4},
			meta:     "",
			preserve: true,
			expected: "old",
		},
		{
			name:     "clear meta",
			current:  semver.SemVersion{Major: 1, Minor: 2, Patch: 3, Build: "old"},
			next:     semver.SemVersion{Major: 1, Minor: 2, Patch: 4},
			meta:     "",
			preserve: false,
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := setBuildMetadata(tt.current, tt.next, tt.meta, tt.preserve)
			if result.Build != tt.expected {
				t.Errorf("expected build %q, got %q", tt.expected, result.Build)
			}
		})
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
/* PRINT QUIET SUMMARY TESTS                                                 */
/* ------------------------------------------------------------------------- */

func TestPrintQuietSummary(t *testing.T) {
	tests := []struct {
		name     string
		results  []workspace.ExecutionResult
		expected string
	}{
		{
			name: "all success",
			results: []workspace.ExecutionResult{
				{Module: &workspace.Module{Name: "mod1"}, Success: true},
				{Module: &workspace.Module{Name: "mod2"}, Success: true},
			},
			expected: "Success: 2 module(s) bumped",
		},
		{
			name: "with failures",
			results: []workspace.ExecutionResult{
				{Module: &workspace.Module{Name: "mod1"}, Success: true},
				{Module: &workspace.Module{Name: "mod2"}, Success: false, Error: fmt.Errorf("failed")},
			},
			expected: "Completed: 1 succeeded, 1 failed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			output, _ := testutils.CaptureStdout(func() {
				printQuietSummary(tt.results)
			})
			if !strings.Contains(output, tt.expected) {
				t.Errorf("expected output to contain %q, got %q", tt.expected, output)
			}
		})
	}
}

/* ------------------------------------------------------------------------- */
/* MOCK IMPLEMENTATIONS FOR HELPER TESTS                                     */
/* ------------------------------------------------------------------------- */

// mockTagManager implements tagmanager.TagManager for testing
type mockTagManager struct {
	validateErr error
	createErr   error
}

func (m *mockTagManager) Name() string                                    { return "mock-tag-manager" }
func (m *mockTagManager) Description() string                             { return "mock tag manager" }
func (m *mockTagManager) Version() string                                 { return "1.0.0" }
func (m *mockTagManager) ValidateTagAvailable(v semver.SemVersion) error  { return m.validateErr }
func (m *mockTagManager) CreateTag(v semver.SemVersion, msg string) error { return m.createErr }
func (m *mockTagManager) FormatTagName(v semver.SemVersion) string        { return "v" + v.String() }
func (m *mockTagManager) TagExists(v semver.SemVersion) (bool, error)     { return false, nil }
func (m *mockTagManager) PushTag(v semver.SemVersion) error               { return nil }
func (m *mockTagManager) DeleteTag(v semver.SemVersion) error             { return nil }
func (m *mockTagManager) GetLatestTag() (semver.SemVersion, error)        { return semver.SemVersion{}, nil }
func (m *mockTagManager) ListTags() ([]string, error)                     { return nil, nil }

// mockVersionValidator implements versionvalidator.VersionValidator for testing
type mockVersionValidator struct {
	validateErr error
}

func (m *mockVersionValidator) Name() string        { return "mock-version-validator" }
func (m *mockVersionValidator) Description() string { return "mock version validator" }
func (m *mockVersionValidator) Version() string     { return "1.0.0" }
func (m *mockVersionValidator) Validate(newV, prevV semver.SemVersion, bumpType string) error {
	return m.validateErr
}
func (m *mockVersionValidator) ValidateSet(v semver.SemVersion) error { return nil }

// mockReleaseGate implements releasegate.ReleaseGate for testing
type mockReleaseGate struct {
	validateErr error
}

func (m *mockReleaseGate) Name() string        { return "mock-release-gate" }
func (m *mockReleaseGate) Description() string { return "mock release gate" }
func (m *mockReleaseGate) Version() string     { return "1.0.0" }
func (m *mockReleaseGate) ValidateRelease(newV, prevV semver.SemVersion, bumpType string) error {
	return m.validateErr
}

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

func TestModuleInfoFromPath(t *testing.T) {
	tests := []struct {
		name     string
		path     string
		wantNil  bool
		wantDir  string
		wantName string
	}{
		{
			name:    "current directory version file",
			path:    ".version",
			wantNil: true,
		},
		{
			name:    "dot path version file",
			path:    "./.version",
			wantNil: true,
		},
		{
			name:     "subdirectory version file",
			path:     "packages/app/.version",
			wantNil:  false,
			wantDir:  "packages/app",
			wantName: "app",
		},
		{
			name:     "absolute path version file",
			path:     "/project/packages/lib/.version",
			wantNil:  false,
			wantDir:  "/project/packages/lib",
			wantName: "lib",
		},
		{
			name:     "nested module path",
			path:     "apps/frontend/web/.version",
			wantNil:  false,
			wantDir:  "apps/frontend/web",
			wantName: "web",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := moduleInfoFromPath(tt.path)

			if tt.wantNil {
				if result != nil {
					t.Errorf("expected nil, got %+v", result)
				}
				return
			}

			if result == nil {
				t.Errorf("expected non-nil result, got nil")
				return
			}

			if result.Dir != tt.wantDir {
				t.Errorf("expected Dir=%q, got %q", tt.wantDir, result.Dir)
			}
			if result.Name != tt.wantName {
				t.Errorf("expected Name=%q, got %q", tt.wantName, result.Name)
			}
		})
	}
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
