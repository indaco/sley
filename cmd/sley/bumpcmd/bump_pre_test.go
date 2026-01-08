package bumpcmd

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"
	"testing"

	"github.com/indaco/sley/internal/clix"
	"github.com/indaco/sley/internal/config"
	"github.com/indaco/sley/internal/hooks"
	"github.com/indaco/sley/internal/plugins"
	"github.com/indaco/sley/internal/testutils"
	"github.com/urfave/cli/v3"
)

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
				hooks.RunPreReleaseHooksFn = func(ctx context.Context, skip bool) error {
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
