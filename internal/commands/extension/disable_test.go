package extension

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/indaco/sley/internal/config"
	"github.com/indaco/sley/internal/testutils"
	"github.com/urfave/cli/v3"
)

/* ------------------------------------------------------------------------- */
/* EXTENSION DISABLE COMMAND                                                 */
/* ------------------------------------------------------------------------- */

func TestExtensionDisableCmd_Success(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, ".sley.yaml")

	content := `extensions:
  - name: mock-extension
    path: /some/path
    enabled: true`
	if err := os.WriteFile(configPath, []byte(content), 0644); err != nil {
		t.Fatalf("failed to write config file: %v", err)
	}

	cfg := &config.Config{Path: configPath}
	appCli := testutils.BuildCLIForTests(cfg.Path, []*cli.Command{Run()})

	output, err := testutils.CaptureStdout(func() {
		testutils.RunCLITest(t, appCli, []string{
			"sley", "extension", "disable", "--name", "mock-extension",
		}, tmpDir)
	})
	if err != nil {
		t.Fatalf("CLI run failed: %v", err)
	}

	expected := `Extension "mock-extension" disabled.`
	if !strings.Contains(output, expected) {
		t.Errorf("expected output to contain %q, got:\n%s", expected, output)
	}

	// Verify the extension is now disabled in the config
	data, readErr := os.ReadFile(configPath)
	if readErr != nil {
		t.Fatalf("failed to read config: %v", readErr)
	}
	if !strings.Contains(string(data), "enabled: false") {
		t.Errorf("expected config to contain 'enabled: false', got:\n%s", string(data))
	}
}

func TestExtensionDisableCmd_MissingName(t *testing.T) {
	if os.Getenv("TEST_EXTENSION_DISABLE_MISSING_NAME") == "1" {
		tmp := t.TempDir()
		configPath := filepath.Join(tmp, ".sley.yaml")
		content := `extensions:
  - name: mock-extension
    path: /some/path
    enabled: true`
		if err := os.WriteFile(configPath, []byte(content), 0644); err != nil {
			fmt.Fprintln(os.Stderr, "failed to write config:", err)
			os.Exit(1)
		}

		cfg := &config.Config{Path: configPath}
		appCli := testutils.BuildCLIForTests(cfg.Path, []*cli.Command{Run()})

		err := appCli.Run(context.Background(), []string{
			"sley", "extension", "disable",
		})
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		os.Exit(0)
	}

	cmd := exec.Command(os.Args[0], "-test.run=TestExtensionDisableCmd_MissingName") //nolint:gosec // G702: standard test re-exec pattern
	cmd.Env = append(os.Environ(), "TEST_EXTENSION_DISABLE_MISSING_NAME=1")
	output, err := cmd.CombinedOutput()

	if err == nil {
		t.Fatal("expected non-zero exit status")
	}

	expected := "please provide an extension name to disable"
	if !strings.Contains(string(output), expected) {
		t.Errorf("expected output to contain %q, got:\n%s", expected, output)
	}
}
