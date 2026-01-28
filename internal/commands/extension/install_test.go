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
/* EXTENSION INSTALL COMMAND                                                 */
/* ------------------------------------------------------------------------- */

func TestExtensionIstallCmd_Success(t *testing.T) {
	// Set up a temporary directory for the version file and config
	tmpDir := t.TempDir()
	versionPath := filepath.Join(tmpDir, ".version")
	configPath := filepath.Join(tmpDir, ".sley.yaml")

	// Create .sley.yaml with the required path field
	configContent := `path: .version`
	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatalf("failed to create .sley.yaml: %v", err)
	}

	// Create a subdirectory for the extension to hold the extension.yaml file
	extensionDir := filepath.Join(tmpDir, "mock-extension")
	if err := os.Mkdir(extensionDir, 0755); err != nil {
		t.Fatalf("failed to create extension directory: %v", err)
	}

	// Create a valid extension.yaml file inside the extension directory
	extensionPath := filepath.Join(extensionDir, "extension.yaml")
	extensionContent := `name: mock-extension
version: 1.0.0
description: Mock Extension
author: Test Author
repository: https://github.com/test/repo
entry: mock-entry`

	if err := os.WriteFile(extensionPath, []byte(extensionContent), 0644); err != nil {
		t.Fatalf("failed to create extension.yaml: %v", err)
	}

	// Prepare and run the CLI command
	cfg := &config.Config{Path: versionPath}
	appCli := testutils.BuildCLIForTests(cfg.Path, []*cli.Command{Run()})

	// Ensure the extension directory is passed correctly
	if _, err := os.Stat(extensionDir); os.IsNotExist(err) {
		t.Fatalf("extension directory does not exist at %s", extensionDir)
	}

	// Run the command, ensuring we pass the correct extension directory
	output, _ := testutils.CaptureStdout(func() {
		testutils.RunCLITest(t, appCli, []string{
			"sley", "extension", "install", "--path", extensionDir,
			"--extension-dir", tmpDir}, tmpDir)
	})

	// Check the output for success
	if !strings.Contains(output, "Extension \"mock-extension\" registered successfully.") {
		t.Fatalf("expected success message, got: %s", output)
	}
}

func TestExtensionRegisterCmd_MissingPathArgument(t *testing.T) {
	if os.Getenv("TEST_SLEY_EXTENSION_MISSING_PATH") == "1" {
		tmp := t.TempDir()
		versionPath := filepath.Join(tmp, ".version")

		// Prepare and run the CLI command
		cfg := &config.Config{Path: versionPath}
		appCli := testutils.BuildCLIForTests(cfg.Path, []*cli.Command{Run()})

		// Run the CLI command with missing --path argument
		err := appCli.Run(context.Background(), []string{"sley", "extension", "install"})
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1) // expected non-zero exit
		}
		os.Exit(0) // should not happen
	}

	// Run the test with the custom environment variable to trigger the error condition
	cmd := exec.Command(os.Args[0], "-test.run=TestExtensionRegisterCmd_MissingPathArgument")
	cmd.Env = append(os.Environ(), "TEST_SLEY_EXTENSION_MISSING_PATH=1")
	output, err := cmd.CombinedOutput()

	// Ensure that the test exits with an error
	if err == nil {
		t.Fatal("expected non-zero exit status")
	}

	// Define the expected error message
	expected := "missing --path or --url for extension installation"

	// Check if the expected error message is in the captured output
	if !strings.Contains(string(output), expected) {
		t.Errorf("expected output to contain %q, got %q", expected, string(output))
	}
}

/* ------------------------------------------------------------------------- */
/* EXTENSION INSTALL COMMAND - ADDITIONAL TESTS                              */
/* ------------------------------------------------------------------------- */

func TestExtensionInstallCmd_RegisterLocalExtensionError(t *testing.T) {
	if os.Getenv("TEST_EXTENSION_REGISTER_ERROR") == "1" {
		tmp := t.TempDir()
		versionPath := filepath.Join(tmp, ".version")

		cfg := &config.Config{Path: versionPath}
		appCli := testutils.BuildCLIForTests(cfg.Path, []*cli.Command{Run()})

		// Run with a non-existent path to trigger the error
		err := appCli.Run(context.Background(), []string{"sley", "extension", "install", "--path", "/non/existent/path"})
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		os.Exit(0)
	}

	cmd := exec.Command(os.Args[0], "-test.run=TestExtensionInstallCmd_RegisterLocalExtensionError")
	cmd.Env = append(os.Environ(), "TEST_EXTENSION_REGISTER_ERROR=1")
	output, err := cmd.CombinedOutput()

	if err == nil {
		t.Fatal("expected non-zero exit status")
	}

	expected := "Failed to install extension"
	if !strings.Contains(string(output), expected) {
		t.Errorf("expected output to contain %q, got %q", expected, string(output))
	}
}

func TestExtensionInstallCmd_URLInPathFlag(t *testing.T) {
	if os.Getenv("TEST_EXTENSION_URL_IN_PATH") == "1" {
		tmp := t.TempDir()
		versionPath := filepath.Join(tmp, ".version")

		cfg := &config.Config{Path: versionPath}
		appCli := testutils.BuildCLIForTests(cfg.Path, []*cli.Command{Run()})

		// Try to pass a URL to --path flag (should be rejected)
		err := appCli.Run(context.Background(), []string{"sley", "extension", "install", "--path", "github.com/user/repo"})
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		os.Exit(0)
	}

	cmd := exec.Command(os.Args[0], "-test.run=TestExtensionInstallCmd_URLInPathFlag")
	cmd.Env = append(os.Environ(), "TEST_EXTENSION_URL_IN_PATH=1")
	output, err := cmd.CombinedOutput()

	if err == nil {
		t.Fatal("expected non-zero exit status")
	}

	expected := "URL detected in --path flag. Please use --url flag for remote installations."
	if !strings.Contains(string(output), expected) {
		t.Errorf("expected output to contain %q, got %q", expected, string(output))
	}
}
