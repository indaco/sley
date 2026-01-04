package extensioncmd

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/indaco/sley/internal/config"
	"github.com/indaco/sley/internal/testutils"
	"github.com/urfave/cli/v3"
)

/* ------------------------------------------------------------------------- */
/* HELPERS                                                                   */
/* ------------------------------------------------------------------------- */

func createExtensionDir(t *testing.T, path string) {
	t.Helper()
	if err := os.MkdirAll(path, 0755); err != nil {
		t.Fatalf("failed to create extension directory: %v", err)
	}
}

func writeConfigFile(t *testing.T, path string) {
	t.Helper()
	content := `extensions:
  - name: mock-extension
    path: /some/path
    enabled: true`
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("failed to write config file: %v", err)
	}
}

func checkCLIOutput(t *testing.T, output, extensionName string, deleted bool) {
	t.Helper()
	var expected string
	if deleted {
		expected = fmt.Sprintf("Extension %q and its directory removed successfully.", extensionName)
	} else {
		expected = fmt.Sprintf("Extension %q removed, but its directory is preserved.", extensionName)
	}
	if !strings.Contains(output, expected) {
		t.Errorf("expected output to contain %q, got:\n%s", expected, output)
	}
}

func checkExtensionDirDeleted(t *testing.T, dir string, expectDeleted bool) {
	t.Helper()
	_, err := os.Stat(dir)
	if expectDeleted {
		if !os.IsNotExist(err) {
			t.Errorf("expected extension directory to be deleted, but it still exists")
		}
	} else {
		if err != nil {
			t.Errorf("expected extension directory to exist, got: %v", err)
		}
	}
}

func checkExtensionDisabledInConfig(t *testing.T, configPath, extensionName string) {
	t.Helper()
	cfg, err := config.LoadConfigFn()
	if err != nil {
		t.Fatalf("failed to reload config: %v", err)
	}

	found := false
	for _, ext := range cfg.Extensions {
		if ext.Name == extensionName {
			found = true
			if ext.Enabled {
				t.Errorf("expected extension %q to be disabled, but it's still enabled", extensionName)
			}
			break
		}
	}
	if !found {
		t.Errorf("extension %q not found in config", extensionName)
	}
}

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
/* EXTENSION LIST COMMAND                                                    */
/* ------------------------------------------------------------------------- */

func TestExtensionListCmd(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, ".sley.yaml")

	// Test with no plugins
	err := os.WriteFile(configPath, []byte("extensions: []\n"), 0644)
	if err != nil {
		t.Fatalf("failed to create .sley.yaml: %v", err)
	}

	// Prepare and run the CLI command
	cfg := &config.Config{Path: configPath}
	appCli := testutils.BuildCLIForTests(cfg.Path, []*cli.Command{Run()})

	output, err := testutils.CaptureStdout(func() {
		testutils.RunCLITest(t, appCli, []string{"sley", "extension", "list"}, tmpDir)
	})
	if err != nil {
		t.Fatalf("Failed to capture stdout: %v", err)
	}

	if !strings.Contains(output, "No extensions registered.") {
		t.Errorf("expected output to contain 'No extensions registered.', got:\n%s", output)
	}

	// Add plugin entries
	extensionsContent := `
extensions:
  - name: mock-extension-1
    path: /path/to/mock-extension-1
    enabled: true
  - name: mock-extension-2
    path: /path/to/mock-extension-2
    enabled: false
`
	err = os.WriteFile(configPath, []byte(extensionsContent), 0644)
	if err != nil {
		t.Fatalf("failed to write .sley.yaml: %v", err)
	}

	output, err = testutils.CaptureStdout(func() {
		testutils.RunCLITest(t, appCli, []string{"sley", "extension", "list"}, tmpDir)
	})
	if err != nil {
		t.Fatalf("Failed to capture stdout: %v", err)
	}

	expectedRows := []string{
		"mock-extension-1",
		"true",
		"mock-extension-2",
		"false",
		"(no manifest)",
	}

	for _, expected := range expectedRows {
		if !strings.Contains(output, expected) {
			t.Errorf("expected output to contain %q, got:\n%s", expected, output)
		}
	}
}

func TestExtensionListCmd_LoadConfigError(t *testing.T) {
	// Create a mock of the LoadConfig function that returns an error
	originalLoadConfig := config.LoadConfigFn
	defer func() {
		// Restore the original LoadConfig function after the test
		config.LoadConfigFn = originalLoadConfig
	}()

	// Mock the LoadConfig function to simulate an error
	config.LoadConfigFn = func() (*config.Config, error) {
		return nil, fmt.Errorf("failed to load configuration")
	}

	// Set up a temporary directory for the config file (not used here, since LoadConfig will fail)
	tmpDir := t.TempDir()

	// Prepare and run the CLI command
	cfg := &config.Config{Path: tmpDir}
	appCli := testutils.BuildCLIForTests(cfg.Path, []*cli.Command{Run()})

	// Capture the output of the plugin list command again
	output, err := testutils.CaptureStdout(func() {
		err := appCli.Run(context.Background(), []string{"sley", "extension", "list"})
		// Capture the actual error during execution
		if err != nil {
			t.Fatalf("Expected no error, got: %v", err)
		}
	})

	if err != nil {
		t.Fatalf("Failed to capture stdout: %v", err)
	}

	// Check if the error message was properly printed
	expectedErrorMessage := "failed to load configuration"
	if !strings.Contains(output, expectedErrorMessage) {
		t.Errorf("Expected error message to contain %q, but got: %q", expectedErrorMessage, output)
	}
}

func TestExtensionListCmd_WithManifest(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, ".sley.yaml")

	extensionName := "test-extension"
	extensionDir := filepath.Join(tmpDir, extensionName)

	// Create extension directory with manifest
	if err := os.MkdirAll(extensionDir, 0755); err != nil {
		t.Fatalf("failed to create extension directory: %v", err)
	}

	manifestContent := `name: test-extension
version: 9.9.9
description: Test extension with manifest
author: Test Author
repository: https://github.com/test/repo
entry: hook.sh
hooks:
  - post-bump
`
	if err := os.WriteFile(filepath.Join(extensionDir, "extension.yaml"), []byte(manifestContent), 0644); err != nil {
		t.Fatalf("failed to write extension manifest: %v", err)
	}

	// Write .sley.yaml with extension pointing to the directory
	content := fmt.Sprintf(`extensions:
  - name: %s
    path: %s
    enabled: true
`, extensionName, extensionDir)
	if err := os.WriteFile(configPath, []byte(content), 0644); err != nil {
		t.Fatalf("failed to write config file: %v", err)
	}

	// Prepare and run the CLI command
	cfg := &config.Config{Path: configPath}
	appCli := testutils.BuildCLIForTests(cfg.Path, []*cli.Command{Run()})

	output, err := testutils.CaptureStdout(func() {
		testutils.RunCLITest(t, appCli, []string{"sley", "extension", "list"}, tmpDir)
	})
	if err != nil {
		t.Fatalf("Failed to capture stdout: %v", err)
	}

	// Ensure metadata from manifest was printed
	expectedValues := []string{
		extensionName,
		"9.9.9",
		"true",
		"Test extension with manifest",
	}
	for _, expected := range expectedValues {
		if !strings.Contains(output, expected) {
			t.Errorf("expected output to contain %q, got:\n%s", expected, output)
		}
	}
}

/* ------------------------------------------------------------------------- */
/* EXTENSION REMOVE COMMAND                                                  */
/* ------------------------------------------------------------------------- */

func TestExtensionRemoveCmd_DeleteFolderVariants(t *testing.T) {
	extensionName := "mock-extension"

	tests := []struct {
		name          string
		deleteFolder  bool
		expectDeleted bool
	}{
		{
			name:          "delete-folder=false",
			deleteFolder:  false,
			expectDeleted: false,
		},
		{
			name:          "delete-folder=true",
			deleteFolder:  true,
			expectDeleted: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()

			extensionsRoot := filepath.Join(tmpDir, ".sley-extensions")
			extensionDir := filepath.Join(extensionsRoot, extensionName)
			configPath := filepath.Join(tmpDir, ".sley.yaml")

			createExtensionDir(t, extensionDir)
			writeConfigFile(t, configPath)

			args := []string{"sley", "extension", "remove", "--name", extensionName}
			if tt.deleteFolder {
				args = append(args, "--delete-folder")
			}

			cfg := &config.Config{Path: configPath}
			appCli := testutils.BuildCLIForTests(cfg.Path, []*cli.Command{Run()})

			output, err := testutils.CaptureStdout(func() {
				testutils.RunCLITest(t, appCli, args, tmpDir)
			})
			if err != nil {
				t.Fatalf("CLI run failed: %v", err)
			}

			checkCLIOutput(t, output, extensionName, tt.deleteFolder)
			checkExtensionDirDeleted(t, extensionDir, tt.expectDeleted)
			checkExtensionDisabledInConfig(t, configPath, extensionName)
		})
	}
}

func TestExtensionRemoveCmd_DeleteFolderFailure(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("permission-based RemoveAll failure is unreliable on Windows")
	}

	tmpDir := t.TempDir()
	extensionName := "mock-extension"
	extensionDir := filepath.Join(tmpDir, ".sley-extensions", extensionName)
	extensionConfigPath := filepath.Join(tmpDir, ".sley.yaml")

	if err := os.MkdirAll(extensionDir, 0755); err != nil {
		t.Fatalf("failed to create extension directory: %v", err)
	}

	protectedFile := filepath.Join(extensionDir, "protected.txt")
	if err := os.WriteFile(protectedFile, []byte("protected"), 0400); err != nil {
		t.Fatalf("failed to create protected file: %v", err)
	}
	if err := os.Chmod(extensionDir, 0500); err != nil {
		t.Fatalf("failed to chmod extension dir: %v", err)
	}

	defer func() {
		_ = os.Chmod(protectedFile, 0600)
		_ = os.Chmod(extensionDir, 0700)
	}()

	content := `extensions:
  - name: mock-extension
    path: /some/path
    enabled: true`
	if err := os.WriteFile(extensionConfigPath, []byte(content), 0644); err != nil {
		t.Fatalf("failed to write config: %v", err)
	}

	// Prepare and run the CLI command
	cfg := &config.Config{Path: extensionConfigPath}
	appCli := testutils.BuildCLIForTests(cfg.Path, []*cli.Command{Run()})

	var cliErr error
	_, captureErr := testutils.CaptureStdout(func() {
		cliErr = testutils.RunCLITestAllowError(t, appCli, []string{
			"sley", "extension", "remove",
			"--name", extensionName,
			"--delete-folder",
		}, tmpDir)
	})
	if captureErr != nil {
		t.Fatalf("failed to capture output: %v", captureErr)
	}

	_ = os.Chmod(protectedFile, 0600)
	_ = os.Chmod(extensionDir, 0700)

	if cliErr == nil {
		t.Fatal("expected error but got nil")
	}

	expectedMsg := "failed to remove extension directory"
	if !strings.Contains(cliErr.Error(), expectedMsg) {
		t.Errorf("expected error message to contain %q, got: %v", expectedMsg, cliErr)
	}

}

func TestCLI_ExtensionRemove_MissingName(t *testing.T) {
	if os.Getenv("TEST_EXTENSION_REMOVE_MISSING_NAME") == "1" {
		tmp := t.TempDir()

		// Write valid .sley.yaml with 1 extension (won't be used, but still required)
		configPath := filepath.Join(tmp, ".sley.yaml")
		content := `extensions:
  - name: mock-extension
    path: /some/path
    enabled: true`
		if err := os.WriteFile(configPath, []byte(content), 0644); err != nil {
			fmt.Fprintln(os.Stderr, "failed to write config:", err)
			os.Exit(1)
		}

		// Prepare and run the CLI command
		cfg := &config.Config{Path: configPath}
		appCli := testutils.BuildCLIForTests(cfg.Path, []*cli.Command{Run()})

		// Run command WITHOUT --name (should trigger the validation)
		err := appCli.Run(context.Background(), []string{
			"sley", "extension", "remove", "--path", configPath,
		})
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}

		// Shouldn't reach here
		os.Exit(0)
	}

	cmd := exec.Command(os.Args[0], "-test.run=TestCLI_ExtensionRemove_MissingName")
	cmd.Env = append(os.Environ(), "TEST_EXTENSION_REMOVE_MISSING_NAME=1")
	output, err := cmd.CombinedOutput()

	if err == nil {
		t.Fatal("expected non-zero exit status")
	}

	expected := "please provide an extension name to remove"
	if !strings.Contains(string(output), expected) {
		t.Errorf("expected output to contain %q, got:\n%s", expected, output)
	}
}

func TestExtensionRemoveCmd_LoadConfigError(t *testing.T) {
	// Mock the LoadConfig function to simulate an error
	originalLoadConfig := config.LoadConfigFn
	defer func() {
		config.LoadConfigFn = originalLoadConfig
	}()

	config.LoadConfigFn = func() (*config.Config, error) {
		return nil, fmt.Errorf("failed to load configuration")
	}

	tmpDir := t.TempDir()
	extensionConfigPath := filepath.Join(tmpDir, ".sley.yaml")

	// Prepare and run the CLI command
	cfg := &config.Config{Path: extensionConfigPath}
	appCli := testutils.BuildCLIForTests(cfg.Path, []*cli.Command{Run()})

	// Capture the output
	output, err := testutils.CaptureStdout(func() {
		testutils.RunCLITest(t, appCli, []string{"sley", "extension", "remove", "--name", "mock-plugin"}, tmpDir)
	})

	if err != nil {
		t.Fatalf("Failed to capture stdout: %v", err)
	}

	// Expected error message
	expected := "failed to load configuration"
	if !strings.Contains(output, expected) {
		t.Errorf("expected output to contain %q, got %q", expected, output)
	}
}

func TestExtensionRemoveCmd_PluginNotFound(t *testing.T) {
	tmpDir := t.TempDir()
	extensionConfigPath := filepath.Join(tmpDir, ".sley.yaml")

	// Create a dummy .sley.yaml configuration file with no extension
	content := `extensions: []`
	if err := os.WriteFile(extensionConfigPath, []byte(content), 0644); err != nil {
		t.Fatalf("failed to create .sley.yaml: %v", err)
	}

	// Prepare and run the CLI command
	cfg := &config.Config{Path: extensionConfigPath}
	appCli := testutils.BuildCLIForTests(cfg.Path, []*cli.Command{Run()})

	// Capture the output
	output, err := testutils.CaptureStdout(func() {
		testutils.RunCLITest(t, appCli, []string{"sley", "extension", "remove", "--name", "mock-extension"}, tmpDir)
	})

	if err != nil {
		t.Fatalf("Failed to capture stdout: %v", err)
	}

	// Expected error message
	expected := "extension \"mock-extension\" not found"
	if !strings.Contains(output, expected) {
		t.Errorf("expected output to contain %q, got %q", expected, output)
	}
}

func TestExtensionRemoveCmd_SaveConfigError(t *testing.T) {
	// Mock the SaveConfig function to simulate an error
	originalSaveConfig := config.SaveConfigFn
	defer func() {
		config.SaveConfigFn = originalSaveConfig
	}()

	config.SaveConfigFn = func(cfg *config.Config) error {
		return fmt.Errorf("failed to save updated configuration")
	}

	tmpDir := t.TempDir()
	extensionConfigPath := filepath.Join(tmpDir, ".sley.yaml")

	// Create a dummy .sley.yaml configuration file
	content := `extensions:
  - name: mock-extension
    path: /path/to/extension
    enabled: true`
	if err := os.WriteFile(extensionConfigPath, []byte(content), 0644); err != nil {
		t.Fatalf("failed to create .sley.yaml: %v", err)
	}

	// Prepare and run the CLI command
	cfg := &config.Config{Path: extensionConfigPath}
	appCli := testutils.BuildCLIForTests(cfg.Path, []*cli.Command{Run()})

	// Capture the output
	output, err := testutils.CaptureStdout(func() {
		testutils.RunCLITest(t, appCli, []string{"sley", "extension", "remove", "--name", "mock-extension"}, tmpDir)
	})

	if err != nil {
		t.Fatalf("Failed to capture stdout: %v", err)
	}

	// Expected error message
	expected := "failed to save updated configuration"
	if !strings.Contains(output, expected) {
		t.Errorf("expected output to contain %q, got %q", expected, output)
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
