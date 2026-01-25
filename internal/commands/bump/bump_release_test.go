package bump

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/indaco/sley/internal/config"
	"github.com/indaco/sley/internal/plugins"
	"github.com/indaco/sley/internal/testutils"
	"github.com/urfave/cli/v3"
)

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
