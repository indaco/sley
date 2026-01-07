package extensionmgr

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/indaco/sley/internal/config"
	"github.com/indaco/sley/internal/extensions"
	"github.com/indaco/sley/internal/testutils"
)

func TestRegisterLocalExtension_Success(t *testing.T) {
	tmpDir := t.TempDir()
	extensionDir := filepath.Join(tmpDir, "myextension")
	if err := os.Mkdir(extensionDir, 0755); err != nil {
		t.Fatal(err)
	}

	manifestContent := `
name: test-extension
version: 1.0.0
description: A test extension
author: John Doe
repository: https://github.com/test/extension
entry: extension.go
`
	if err := os.WriteFile(filepath.Join(extensionDir, "extension.yaml"), []byte(manifestContent), 0644); err != nil {
		t.Fatal(err)
	}

	cfgPath := filepath.Join(tmpDir, ".sley.yaml")
	if err := os.WriteFile(cfgPath, []byte("path: .version\n"), 0644); err != nil {
		t.Fatal(err)
	}

	// Override .sley-extensions dir for test
	originalCopyDir := copyDirFn
	defer func() { copyDirFn = originalCopyDir }()

	copyDirFn = func(src, dst string) error {
		if !strings.Contains(src, "myextension") || !strings.Contains(dst, "test-extension") {
			t.Errorf("unexpected copy src=%q dst=%q", src, dst)
		}
		return nil
	}

	err := RegisterLocalExtensionFn(extensionDir, cfgPath, tmpDir)
	if err != nil {
		t.Fatalf("expected success, got error: %v", err)
	}
}

func TestRegisterLocalExtension_InvalidPath(t *testing.T) {
	tmpDir := os.TempDir()
	err := RegisterLocalExtensionFn("/nonexistent/path", ".sley.yaml", tmpDir)
	if err == nil || !strings.Contains(err.Error(), "extension path") {
		t.Errorf("expected extension path error, got: %v", err)
	}
}

func TestRegisterLocalExtension_NotDirectory(t *testing.T) {
	tmpDir := t.TempDir()
	file := filepath.Join(tmpDir, "file.txt")
	_ = os.WriteFile(file, []byte("test"), 0644)

	err := RegisterLocalExtensionFn(file, ".sley.yaml", tmpDir)
	if err == nil || !strings.Contains(err.Error(), "must be a directory") {
		t.Errorf("expected directory error, got: %v", err)
	}
}

func TestRegisterLocalExtension_InvalidManifest(t *testing.T) {
	tmpDir := t.TempDir()
	extensionDir := filepath.Join(tmpDir, "invalidextension")
	_ = os.Mkdir(extensionDir, 0755)
	_ = os.WriteFile(filepath.Join(extensionDir, "extension.yaml"), []byte("invalid: yaml:::"), 0644)

	err := RegisterLocalExtensionFn(extensionDir, ".sley.yaml", tmpDir)
	if err == nil || !strings.Contains(err.Error(), "failed to load extension manifest") {
		t.Errorf("expected manifest load error, got: %v", err)
	}
}

func TestRegisterLocalExtension_CopyDirFails(t *testing.T) {
	tmpDir := os.TempDir()
	// Setup mock extension directory
	extensionDir := setupextensionDir(t, "mock-extension", "1.0.0")

	// Create the config file
	configPath := testutils.WriteTempConfig(t, "extensions: []\n")

	// Temporarily override CopyDirFn to simulate failure
	originalCopyDirFn := copyDirFn
	copyDirFn = func(src, dst string) error {
		return fmt.Errorf("simulated copy failure")
	}
	defer func() {
		// Restore original CopyDir function
		copyDirFn = originalCopyDirFn
	}()

	// Call RegisterLocalExtensionFn which should now fail due to the simulated copy error
	err := RegisterLocalExtensionFn(extensionDir, configPath, tmpDir)
	if err == nil {
		t.Fatal("expected error when copying, got nil")
	}

	if !strings.Contains(err.Error(), "simulated copy failure") {
		t.Fatalf("expected simulated copy error, got: %v", err)
	}
}

func TestRegisterLocalExtension_DefaultConfigPath(t *testing.T) {
	content := "path: .version"
	tmpConfigPath := testutils.WriteTempConfig(t, content)
	tmpDir := filepath.Dir(tmpConfigPath)
	tmpextensionDir := setupextensionDir(t, "mock-extension", "1.0.0")

	origDir, err := os.Getwd() // Get the original working directory to restore later
	if err != nil {
		t.Fatalf("failed to get working directory: %v", err)
	}

	if err := os.Chdir(tmpDir); err != nil { // Change to the directory of the temporary config file
		t.Fatalf("failed to change directory to %s: %v", tmpDir, err)
	}
	t.Cleanup(func() { // Ensure we restore the original working directory after the test
		if err := os.Chdir(origDir); err != nil {
			t.Fatalf("failed to restore working directory: %v", err)
		}
	})

	// Register the extension for the first time
	err = RegisterLocalExtensionFn(tmpextensionDir, tmpConfigPath, tmpDir)
	if err != nil {
		t.Fatalf("expected no error on first extension registration, got: %v", err)
	}

	// Register the extension again
	err = RegisterLocalExtensionFn(tmpextensionDir, tmpConfigPath, tmpDir)
	if err != nil {
		t.Fatalf("expected no error on second extension registration, got: %v", err)
	}

	// Check if the .sley.yaml file exists before loading it
	if _, err := os.Stat(tmpConfigPath); os.IsNotExist(err) {
		t.Fatalf(".sley.yaml file does not exist at %s", tmpConfigPath)
	}

	// Ensure the config file has the extension registered
	cfg, err := config.LoadConfigFn()
	if err != nil {
		t.Fatalf("expected no error loading config, got: %v", err)
	}

	// Guard check for nil config
	if cfg == nil {
		t.Fatal("config is nil after loading")
	}

	// Check that there's exactly one extensions
	if len(cfg.Extensions) != 1 {
		t.Fatalf("expected 1 extension in config, got: %d", len(cfg.Extensions))
	}

	// Ensure that the default config path has been used if configPath is empty
	if cfg.Path != ".version" {
		t.Errorf("expected config path to be .version, got: %s", cfg.Path)
	}

	// Test that the default config path is used when no configPath is passed
	err = RegisterLocalExtensionFn(tmpextensionDir, "", tmpDir)
	if err != nil {
		t.Fatalf("expected no error on second extension registration with empty configPath, got: %v", err)
	}

	// Verify the path has been set to the default value when configPath is empty
	cfg, err = config.LoadConfigFn()
	if err != nil {
		t.Fatalf("expected no error loading config, got: %v", err)
	}

	// Ensure the path is still the default
	if cfg.Path != ".version" {
		t.Errorf("expected config path to be .version, got: %s", cfg.Path)
	}
}

func TestRegisterLocalExtension_DefaultConfigPathUsed_CurrentWorkingDir(t *testing.T) {
	content := "path: .version"
	tmpConfigPath := testutils.WriteTempConfig(t, content)
	tmpDir := filepath.Dir(tmpConfigPath)
	tmpextensionDir := setupextensionDir(t, "mock-extension", "1.0.0")

	// Resolve expected path in $HOME/.sley-extensions
	homeDir, err := os.UserHomeDir()
	if err != nil {
		t.Fatalf("failed to get user home directory: %v", err)
	}
	extensionsPath := filepath.Join(homeDir, ".sley-extensions", "mock-extension")

	// Cleanup: remove it before and after the test
	_ = os.RemoveAll(extensionsPath)
	t.Cleanup(func() {
		_ = os.RemoveAll(extensionsPath)
	})

	origDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get working directory: %v", err)
	}

	// Change to the directory of the temporary config file
	if err := os.Chdir(tmpDir); err != nil {
		t.Fatalf("failed to change directory to %s: %v", tmpDir, err)
	}
	t.Cleanup(func() {
		// Restore original working directory
		if err := os.Chdir(origDir); err != nil {
			t.Fatalf("failed to restore working directory: %v", err)
		}
	})

	// Register the extension with default extension path
	err = RegisterLocalExtensionFn(tmpextensionDir, tmpConfigPath, "")
	if err != nil {
		t.Fatalf("expected no error on extension registration, got: %v", err)
	}

	// Assert extension was copied into $HOME/.sley-extensions
	if _, err := os.Stat(extensionsPath); os.IsNotExist(err) {
		t.Fatalf("extension folder does not exist at %s", extensionsPath)
	}

	// Ensure the config file has the extension registered
	cfg, err := config.LoadConfigFn()
	if err != nil {
		t.Fatalf("expected no error loading config, got: %v", err)
	}

	if len(cfg.Extensions) != 1 {
		t.Fatalf("expected 1 extension in config, got: %d", len(cfg.Extensions))
	}
}

func TestRegisterLocalExtension_DefaultConfigPathUsed_OtherDir(t *testing.T) {
	content := "path: .version"
	tmpConfigPath := testutils.WriteTempConfig(t, content)
	tmpDir := filepath.Dir(tmpConfigPath)
	tmpExtensionDir := setupextensionDir(t, "mock-extension", "1.0.0")

	// Set up a temporary directory for the extension
	tmpExtensionFolder := t.TempDir()

	origDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get working directory: %v", err)
	}

	// Change to the directory of the temporary config file
	if err := os.Chdir(tmpDir); err != nil {
		t.Fatalf("failed to change directory to %s: %v", tmpDir, err)
	}
	t.Cleanup(func() {
		// Ensure we restore the original working directory after the test
		if err := os.Chdir(origDir); err != nil {
			t.Fatalf("failed to restore working directory: %v", err)
		}
	})

	// Register the extension with the temporary extension folder
	err = RegisterLocalExtensionFn(tmpExtensionDir, tmpConfigPath, tmpExtensionFolder)
	if err != nil {
		t.Fatalf("expected no error on extension registration, got: %v", err)
	}

	// Ensure the extension was copied into the temporary extension folder
	extensionPath := filepath.Join(tmpExtensionFolder, ".sley-extensions", "mock-extension")
	if _, err := os.Stat(extensionPath); os.IsNotExist(err) {
		t.Fatalf("extension folder does not exist at %s", extensionPath)
	}

	// Ensure the config file has the extension registered
	cfg, err := config.LoadConfigFn()
	if err != nil {
		t.Fatalf("expected no error loading config, got: %v", err)
	}

	if len(cfg.Extensions) != 1 {
		t.Fatalf("expected 1 extension in config, got: %d", len(cfg.Extensions))
	}
}

func TestRegisterLocalExtension_UserHomeDirError(t *testing.T) {
	// Backup and restore the original function
	originalFn := userHomeDirFn
	defer func() { userHomeDirFn = originalFn }()

	userHomeDirFn = func() (string, error) {
		return "", errors.New("mocked failure")
	}

	tmpExtensionDir := setupextensionDir(t, "mock-extension", "1.0.0")
	tmpConfigPath := testutils.WriteTempConfig(t, "path: .version")

	err := RegisterLocalExtensionFn(tmpExtensionDir, tmpConfigPath, "")
	if err == nil || !strings.Contains(err.Error(), "failed to get user home directory") {
		t.Fatalf("expected user home dir error, got: %v", err)
	}
}

func TestRegisterLocalExtension_ValidConfigPath(t *testing.T) {
	// Set up temporary directories
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, ".sley.yaml")

	// Create a mock config file at the given path
	if err := os.WriteFile(configPath, []byte("path: .version"), 0644); err != nil {
		t.Fatalf("failed to create mock .sley.yaml: %v", err)
	}

	// Create a mock extension directory
	extensionDir := t.TempDir()
	extensionPath := filepath.Join(extensionDir, "extension.yaml")

	// Create a mock extension manifest file
	extensionManifest := []byte(`
name: mock-extension
version: "1.0.0"
description: Mock extension
author: Test Author
repository: https://github.com/test/repo
entry: mock-extension.go
`)

	if err := os.WriteFile(extensionPath, extensionManifest, 0644); err != nil {
		t.Fatalf("failed to create mock extension.yaml: %v", err)
	}

	// Call the RegisterLocalExtension function
	err := RegisterLocalExtensionFn(extensionDir, configPath, tmpDir)
	if err != nil {
		t.Fatalf("expected no error during extension registration, got: %v", err)
	}
}

func TestRegisterLocalExtension_InvalidConfigPath(t *testing.T) {
	// Set up temporary directories
	tmpDir := t.TempDir()

	// Use a non-existent config path for testing
	nonExistentConfigPath := filepath.Join(tmpDir, "nonexistent-config.yaml")

	// Create a mock extension directory
	extensionDir := t.TempDir()
	extensionPath := filepath.Join(extensionDir, "extension.yaml")

	// Create a mock extension manifest file
	extensionManifest := []byte(`
name: mock-extension
version: "1.0.0"
description: Mock extension
author: Test Author
repository: https://github.com/test/repo
entry: mock-extension.go
`)

	if err := os.WriteFile(extensionPath, extensionManifest, 0644); err != nil {
		t.Fatalf("failed to create mock extension.yaml: %v", err)
	}

	// Call the RegisterLocalExtension function with an invalid config path
	err := RegisterLocalExtensionFn(extensionDir, nonExistentConfigPath, tmpDir)
	if err == nil {
		t.Fatal("expected error due to non-existent config file, got nil")
	}

	// Check that the error message contains "config file not found"
	expectedErr := fmt.Sprintf("config file not found at %s", nonExistentConfigPath)
	if !strings.Contains(err.Error(), expectedErr) {
		t.Errorf("expected error to contain %q, got: %v", expectedErr, err)
	}
}

func TestRegisterLocalExtension_InvalidConfigPathResolution(t *testing.T) {
	// Create a temporary extension directory
	tmpextensionDir := setupextensionDir(t, "mock-extension", "1.0.0")

	// Simulate an invalid config path
	invalidConfigPath := "/invalid/path/to/.sley.yaml"

	// Try registering the extension with the invalid config path
	err := RegisterLocalExtensionFn(tmpextensionDir, invalidConfigPath, os.TempDir())
	if err == nil {
		t.Fatal("expected error due to invalid config path resolution, got nil")
	}

	// Check if the error message is about the config file not being found
	expectedErr := "config file not found"
	if !strings.Contains(err.Error(), expectedErr) {
		t.Errorf("expected error message to contain %q, got: %v", expectedErr, err)
	}
}

func TestRegisterLocalExtension_InstallExtensionToConfigError(t *testing.T) {
	// Set up the initial config
	tmpConfigPath := testutils.WriteTempConfig(t, `path: .version`)
	tmpextensionDir := setupextensionDir(t, "mock-extension", "1.0.0")

	// Simulate the error returned by InstallExtensionToConfig
	localPath := t.TempDir()
	cfgPath := tmpConfigPath // Path to the config file

	// Mock the LoadExtensionManifest function to return a mock manifest
	mockManifest := &extensions.ExtensionManifest{
		Name:        "mock-extension",
		Version:     "1.0.0",
		Description: "Mock Extension",
		Author:      "Test Author",
		Repository:  "https://github.com/test/repo",
		Entry:       "mock-entry",
	}

	// Mock LoadExtensionManifest to return the mock manifest
	extensions.LoadExtensionManifestFn = func(path string) (*extensions.ExtensionManifest, error) {
		return mockManifest, nil
	}

	// Simulate AddExtensionToConfig error by overriding the function
	originalAddExtensionToConfig := AddExtensionToConfigFn
	defer func() {
		AddExtensionToConfigFn = originalAddExtensionToConfig // Restore original after test
	}()
	AddExtensionToConfigFn = func(path string, extension config.ExtensionConfig) error {
		return fmt.Errorf("failed to update config: some error")
	}

	// Attempt to register the extension
	err := RegisterLocalExtensionFn(localPath, cfgPath, tmpextensionDir)

	// Check that we get the expected error
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "failed to update config") {
		t.Errorf("unexpected error: %v", err)
	}
}

func setupextensionDir(t *testing.T, name, version string) string {
	t.Helper()

	dir := t.TempDir()
	manifestContent := fmt.Sprintf(`name: %s
version: %s
description: test extension
author: test
repository: https://example.com/%s.git
entry: extension.go
`, name, version, name)

	if err := os.WriteFile(filepath.Join(dir, "extension.yaml"), []byte(manifestContent), 0644); err != nil {
		t.Fatalf("failed to write extension.yaml: %v", err)
	}

	return dir
}
