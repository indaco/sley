package extensionmgr

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/goccy/go-yaml"

	"github.com/indaco/sley/internal/config"
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

	registrar := NewDefaultExtensionRegistrarInstance()
	err := registrar.Register(extensionDir, cfgPath, tmpDir)
	if err != nil {
		t.Fatalf("expected success, got error: %v", err)
	}
}

func TestRegisterLocalExtension_InvalidPath(t *testing.T) {
	tmpDir := os.TempDir()
	registrar := NewDefaultExtensionRegistrarInstance()
	err := registrar.Register("/nonexistent/path", ".sley.yaml", tmpDir)
	if err == nil || !strings.Contains(err.Error(), "extension path") {
		t.Errorf("expected extension path error, got: %v", err)
	}
}

func TestRegisterLocalExtension_NotDirectory(t *testing.T) {
	tmpDir := t.TempDir()
	file := filepath.Join(tmpDir, "file.txt")
	_ = os.WriteFile(file, []byte("test"), 0644)

	registrar := NewDefaultExtensionRegistrarInstance()
	err := registrar.Register(file, ".sley.yaml", tmpDir)
	if err == nil || !strings.Contains(err.Error(), "must be a directory") {
		t.Errorf("expected directory error, got: %v", err)
	}
}

func TestRegisterLocalExtension_InvalidManifest(t *testing.T) {
	tmpDir := t.TempDir()
	extensionDir := filepath.Join(tmpDir, "invalidextension")
	_ = os.Mkdir(extensionDir, 0755)
	_ = os.WriteFile(filepath.Join(extensionDir, "extension.yaml"), []byte("invalid: yaml:::"), 0644)

	registrar := NewDefaultExtensionRegistrarInstance()
	err := registrar.Register(extensionDir, ".sley.yaml", tmpDir)
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

	// Create mock file copier that fails
	mockFileCopier := &MockFileCopier{
		CopyDirFunc: func(src, dst string) error {
			return fmt.Errorf("simulated copy failure")
		},
	}

	// Create registrar with mocked file copier
	registrar := NewDefaultExtensionRegistrar(
		&DefaultManifestLoader{},
		NewDefaultConfigUpdater(&DefaultYAMLMarshaler{}),
		mockFileCopier,
		&OSHomeDirectory{},
	)

	// Call Register which should now fail due to the simulated copy error
	err := registrar.Register(extensionDir, configPath, tmpDir)
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

	// Setup working directory and cleanup
	setupWorkingDirForTest(t, tmpDir)

	// Register the extension for the first time
	registrar := NewDefaultExtensionRegistrarInstance()
	err := registrar.Register(tmpextensionDir, tmpConfigPath, tmpDir)
	if err != nil {
		t.Fatalf("expected no error on first extension registration, got: %v", err)
	}

	// Attempt duplicate registration - should fail
	verifyDuplicateRegistrationError(t, tmpextensionDir, tmpConfigPath, tmpDir)

	// Verify config state
	verifyConfigHasOneExtension(t, tmpConfigPath, ".version")

	// Attempt registration with empty configPath - should also fail
	verifyDuplicateRegistrationError(t, tmpextensionDir, "", tmpDir)

	// Verify config still has one extension
	verifyConfigHasOneExtension(t, tmpConfigPath, ".version")
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
	registrar := NewDefaultExtensionRegistrarInstance()
	err = registrar.Register(tmpextensionDir, tmpConfigPath, "")
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
	registrar := NewDefaultExtensionRegistrarInstance()
	err = registrar.Register(tmpExtensionDir, tmpConfigPath, tmpExtensionFolder)
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

func TestRegisterLocalExtension_DotExtensionDir(t *testing.T) {
	content := "path: .version"
	tmpConfigPath := testutils.WriteTempConfig(t, content)
	tmpDir := filepath.Dir(tmpConfigPath)
	tmpExtensionDir := setupextensionDir(t, "mock-extension", "1.0.0")

	origDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get working directory: %v", err)
	}

	// Change to the directory of the temporary config file
	if err := os.Chdir(tmpDir); err != nil {
		t.Fatalf("failed to change directory to %s: %v", tmpDir, err)
	}
	t.Cleanup(func() {
		if err := os.Chdir(origDir); err != nil {
			t.Fatalf("failed to restore working directory: %v", err)
		}
	})

	// Register the extension with "." as extension directory
	// This should install to ./.sley-extensions/, NOT $HOME/.sley-extensions/
	registrar := NewDefaultExtensionRegistrarInstance()
	err = registrar.Register(tmpExtensionDir, tmpConfigPath, ".")
	if err != nil {
		t.Fatalf("expected no error on extension registration, got: %v", err)
	}

	// Verify extension was installed to current directory, not home
	expectedPath := filepath.Join(tmpDir, ".sley-extensions", "mock-extension")
	if _, err := os.Stat(expectedPath); os.IsNotExist(err) {
		t.Fatalf("extension should be installed at %s but was not found", expectedPath)
	}

	// Verify it was NOT installed to home directory
	homeDir, err := os.UserHomeDir()
	if err != nil {
		t.Fatalf("failed to get user home directory: %v", err)
	}
	homePath := filepath.Join(homeDir, ".sley-extensions", "mock-extension")
	if _, err := os.Stat(homePath); err == nil {
		// Clean up if it exists (shouldn't happen)
		_ = os.RemoveAll(homePath)
		t.Fatalf("extension should NOT be installed at home directory %s", homePath)
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
	tmpExtensionDir := setupextensionDir(t, "mock-extension", "1.0.0")
	tmpConfigPath := testutils.WriteTempConfig(t, "path: .version")

	// Create registrar with mock home directory that fails
	mockHomeDir := &MockHomeDirectory{
		GetFunc: func() (string, error) {
			return "", errors.New("mocked failure")
		},
	}

	registrar := NewDefaultExtensionRegistrar(
		&DefaultManifestLoader{},
		NewDefaultConfigUpdater(&DefaultYAMLMarshaler{}),
		NewOSFileCopier(),
		mockHomeDir,
	)

	err := registrar.Register(tmpExtensionDir, tmpConfigPath, "")
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
	registrar := NewDefaultExtensionRegistrarInstance()
	err := registrar.Register(extensionDir, configPath, tmpDir)
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
	registrar := NewDefaultExtensionRegistrarInstance()
	err := registrar.Register(extensionDir, nonExistentConfigPath, tmpDir)
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
	registrar := NewDefaultExtensionRegistrarInstance()
	err := registrar.Register(tmpextensionDir, invalidConfigPath, os.TempDir())
	if err == nil {
		t.Fatal("expected error due to invalid config path resolution, got nil")
	}

	// Check if the error message is about the config file not being found
	expectedErr := "config file not found"
	if !strings.Contains(err.Error(), expectedErr) {
		t.Errorf("expected error message to contain %q, got: %v", expectedErr, err)
	}
}

func TestRegisterLocalExtension_PathIsRelative(t *testing.T) {
	// Create temp directory for test
	tmpDir := t.TempDir()

	// Create config file
	configPath := filepath.Join(tmpDir, ".sley.yaml")
	if err := os.WriteFile(configPath, []byte("path: .version\n"), 0644); err != nil {
		t.Fatalf("failed to create config: %v", err)
	}

	// Create extension directory
	extensionDir := setupextensionDir(t, "test-extension", "1.0.0")

	// Register extension with project-local installation
	registrar := NewDefaultExtensionRegistrarInstance()
	err := registrar.Register(extensionDir, configPath, tmpDir)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	// Read and parse the config file
	data, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatalf("failed to read config: %v", err)
	}

	var cfg config.Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		t.Fatalf("failed to parse config: %v", err)
	}

	// Verify extension was added
	if len(cfg.Extensions) != 1 {
		t.Fatalf("expected 1 extension, got %d", len(cfg.Extensions))
	}

	ext := cfg.Extensions[0]

	// Verify path is relative (not absolute)
	if filepath.IsAbs(ext.Path) {
		t.Errorf("expected relative path, got absolute path: %s", ext.Path)
	}

	// Verify path format is consistent (.sley-extensions/extension-name)
	expectedPath := filepath.Join(".sley-extensions", "test-extension")
	if ext.Path != expectedPath {
		t.Errorf("expected path %q, got %q", expectedPath, ext.Path)
	}
}

func TestRegisterLocalExtension_InstallExtensionToConfigError(t *testing.T) {
	// Set up the initial config
	tmpConfigPath := testutils.WriteTempConfig(t, `path: .version`)
	tmpExtensionDir := setupextensionDir(t, "mock-extension", "1.0.0")
	tmpInstallDir := t.TempDir()
	cfgPath := tmpConfigPath // Path to the config file

	// Create mock config updater that fails
	mockUpdater := &MockConfigUpdater{
		AddExtensionFunc: func(path string, extension config.ExtensionConfig) error {
			return fmt.Errorf("failed to update config: some error")
		},
	}

	// Create registrar with mocked config updater
	registrar := NewDefaultExtensionRegistrar(
		&DefaultManifestLoader{},
		mockUpdater,
		NewOSFileCopier(),
		&OSHomeDirectory{},
	)

	// Attempt to register the extension
	err := registrar.Register(tmpExtensionDir, cfgPath, tmpInstallDir)

	// Check that we get the expected error
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "failed to update config") {
		t.Errorf("unexpected error: %v", err)
	}
}

// setupWorkingDirForTest changes to the given directory and registers cleanup
func setupWorkingDirForTest(t *testing.T, targetDir string) {
	t.Helper()

	origDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get working directory: %v", err)
	}

	if err := os.Chdir(targetDir); err != nil {
		t.Fatalf("failed to change directory to %s: %v", targetDir, err)
	}

	t.Cleanup(func() {
		if err := os.Chdir(origDir); err != nil {
			t.Fatalf("failed to restore working directory: %v", err)
		}
	})
}

// verifyDuplicateRegistrationError attempts registration and verifies it fails with "already registered" error
func verifyDuplicateRegistrationError(t *testing.T, extensionDir, configPath, installDir string) {
	t.Helper()

	registrar := NewDefaultExtensionRegistrarInstance()
	err := registrar.Register(extensionDir, configPath, installDir)
	if err == nil {
		t.Fatal("expected error on duplicate extension registration, got nil")
	}

	if !strings.Contains(err.Error(), "already registered") {
		t.Errorf("expected error to contain 'already registered', got: %v", err)
	}
}

// verifyConfigHasOneExtension loads config and verifies it has exactly one extension with expected path
func verifyConfigHasOneExtension(t *testing.T, configPath, expectedPath string) {
	t.Helper()

	// Check config file exists
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		t.Fatalf(".sley.yaml file does not exist at %s", configPath)
	}

	// Load config
	cfg, err := config.LoadConfigFn()
	if err != nil {
		t.Fatalf("expected no error loading config, got: %v", err)
	}

	if cfg == nil {
		t.Fatal("config is nil after loading")
	}

	// Verify extension count
	if len(cfg.Extensions) != 1 {
		t.Fatalf("expected 1 extension in config, got: %d", len(cfg.Extensions))
	}

	// Verify config path
	if cfg.Path != expectedPath {
		t.Errorf("expected config path to be %s, got: %s", expectedPath, cfg.Path)
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
