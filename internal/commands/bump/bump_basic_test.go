package bump

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"
	"testing"

	"github.com/indaco/sley/internal/config"
	"github.com/indaco/sley/internal/hooks"
	"github.com/indaco/sley/internal/plugins"
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

// errorHook is a PreReleaseHook that always returns an error.
type errorHook struct{}

func (errorHook) HookName() string { return "error-hook" }
func (errorHook) Run(_ context.Context) error {
	return fmt.Errorf("mock pre-release hooks error")
}

func TestCLI_BumpSubcommands_EarlyFailures(t *testing.T) {

	tests := []struct {
		name        string
		args        []string
		setup       func(t *testing.T, tmpDir string)
		expectedErr string
	}{
		{
			name: "patch - FromCommand fails (strict + missing file)",
			args: []string{"sley", "bump", "patch", "--strict"},
			setup: func(t *testing.T, tmpDir string) {
				// Do NOT create .version file so strict mode fails
			},
			expectedErr: "version file not found",
		},
		{
			name: "patch - RunPreReleaseHooks fails",
			args: []string{"sley", "bump", "patch"},
			setup: func(t *testing.T, tmpDir string) {
				testutils.WriteTempVersionFile(t, tmpDir, "1.2.3")
				hooks.ResetPreReleaseHooks()
				hooks.RegisterPreReleaseHook(errorHook{})
				t.Cleanup(func() { hooks.ResetPreReleaseHooks() })
			},
			expectedErr: "mock pre-release hooks error",
		},
		{
			name:        "minor - FromCommand fails (strict + missing file)",
			args:        []string{"sley", "bump", "minor", "--strict"},
			setup:       func(t *testing.T, tmpDir string) {},
			expectedErr: "version file not found",
		},
		{
			name: "minor - RunPreReleaseHooks fails",
			args: []string{"sley", "bump", "minor"},
			setup: func(t *testing.T, tmpDir string) {
				testutils.WriteTempVersionFile(t, tmpDir, "1.2.3")
				hooks.ResetPreReleaseHooks()
				hooks.RegisterPreReleaseHook(errorHook{})
				t.Cleanup(func() { hooks.ResetPreReleaseHooks() })
			},
			expectedErr: "mock pre-release hooks error",
		},
		{
			name:        "major - FromCommand fails (strict + missing file)",
			args:        []string{"sley", "bump", "major", "--strict"},
			setup:       func(t *testing.T, tmpDir string) {},
			expectedErr: "version file not found",
		},
		{
			name: "major - RunPreReleaseHooks fails",
			args: []string{"sley", "bump", "major"},
			setup: func(t *testing.T, tmpDir string) {
				testutils.WriteTempVersionFile(t, tmpDir, "1.2.3")
				hooks.ResetPreReleaseHooks()
				hooks.RegisterPreReleaseHook(errorHook{})
				t.Cleanup(func() { hooks.ResetPreReleaseHooks() })
			},
			expectedErr: "mock pre-release hooks error",
		},
		{
			name:        "auto - FromCommand fails (strict + missing file)",
			args:        []string{"sley", "bump", "auto", "--strict"},
			setup:       func(t *testing.T, tmpDir string) {},
			expectedErr: "version file not found",
		},
		{
			name: "auto - RunPreReleaseHooks fails",
			args: []string{"sley", "bump", "auto"},
			setup: func(t *testing.T, tmpDir string) {
				testutils.WriteTempVersionFile(t, tmpDir, "1.2.3")
				hooks.ResetPreReleaseHooks()
				hooks.RegisterPreReleaseHook(errorHook{})
				t.Cleanup(func() { hooks.ResetPreReleaseHooks() })
			},
			expectedErr: "mock pre-release hooks error",
		},
		{
			name:        "release - FromCommand fails (strict + missing file)",
			args:        []string{"sley", "bump", "release", "--strict"},
			setup:       func(t *testing.T, tmpDir string) {},
			expectedErr: "version file not found",
		},
		{
			name: "release - RunPreReleaseHooks fails",
			args: []string{"sley", "bump", "release"},
			setup: func(t *testing.T, tmpDir string) {
				testutils.WriteTempVersionFile(t, tmpDir, "1.2.3")
				hooks.ResetPreReleaseHooks()
				hooks.RegisterPreReleaseHook(errorHook{})
				t.Cleanup(func() { hooks.ResetPreReleaseHooks() })
			},
			expectedErr: "mock pre-release hooks error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			versionPath := filepath.Join(tmpDir, ".version")

			tt.setup(t, tmpDir)

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
