package extension

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/indaco/sley/internal/config"
	"github.com/indaco/sley/internal/testutils"
	"github.com/urfave/cli/v3"
)

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
