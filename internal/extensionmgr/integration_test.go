package extensionmgr

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/indaco/sley/internal/config"
)

// setupTestExtension creates a test extension in the given directory
func setupTestExtension(t *testing.T, extensionSrcDir string) {
	t.Helper()

	if err := os.MkdirAll(extensionSrcDir, 0755); err != nil {
		t.Fatalf("failed to create extension source directory: %v", err)
	}

	manifestContent := `name: test-lifecycle-ext
version: 1.0.0
description: Test extension for lifecycle validation
author: test
repository: https://github.com/test/test-lifecycle-ext
entry: hook.sh
hooks:
  - pre-bump
  - post-bump
`
	manifestPath := filepath.Join(extensionSrcDir, "extension.yaml")
	if err := os.WriteFile(manifestPath, []byte(manifestContent), 0644); err != nil {
		t.Fatalf("failed to create extension manifest: %v", err)
	}

	hookScript := `#!/bin/sh
read input
echo '{"success": true, "message": "Lifecycle test hook executed"}'
`
	hookPath := filepath.Join(extensionSrcDir, "hook.sh")
	if err := os.WriteFile(hookPath, []byte(hookScript), 0755); err != nil {
		t.Fatalf("failed to create hook script: %v", err)
	}
}

// setupTestProject creates a test project directory with .sley.yaml
func setupTestProject(t *testing.T, projectDir string) string {
	t.Helper()

	if err := os.MkdirAll(projectDir, 0755); err != nil {
		t.Fatalf("failed to create project directory: %v", err)
	}

	configPath := filepath.Join(projectDir, ".sley.yaml")
	initialConfig := `path: .version
extensions: []
`
	if err := os.WriteFile(configPath, []byte(initialConfig), 0644); err != nil {
		t.Fatalf("failed to create config file: %v", err)
	}

	return configPath
}

// verifyExtensionInstalled checks that the extension was installed correctly
func verifyExtensionInstalled(t *testing.T, tmpDir, extensionName string) {
	t.Helper()

	extensionDir := filepath.Join(tmpDir, ".sley-extensions")
	installedExtPath := filepath.Join(extensionDir, extensionName)

	if _, err := os.Stat(installedExtPath); os.IsNotExist(err) {
		t.Error("extension was not installed to the expected directory")
	}

	installedManifest := filepath.Join(installedExtPath, "extension.yaml")
	if _, err := os.Stat(installedManifest); os.IsNotExist(err) {
		t.Error("extension manifest was not copied to installed location")
	}
}

// verifyConfigUpdated checks that the config file was updated correctly
func verifyConfigUpdated(t *testing.T, configPath string) *config.Config {
	t.Helper()

	cfgData, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatalf("failed to read config file: %v", err)
	}

	if len(cfgData) == 0 {
		t.Error("config file is empty")
	}

	cfg, err := config.LoadConfigFn()
	if err != nil {
		t.Fatalf("failed to load config: %v", err)
	}

	if len(cfg.Extensions) != 1 {
		t.Errorf("expected 1 extension in config, got %d", len(cfg.Extensions))
	}

	if cfg.Extensions[0].Name != "test-lifecycle-ext" {
		t.Errorf("expected extension name 'test-lifecycle-ext', got %s", cfg.Extensions[0].Name)
	}

	if !cfg.Extensions[0].Enabled {
		t.Error("extension should be enabled by default")
	}

	return cfg
}

// TestExtensionLifecycle tests the full lifecycle of an extension from installation to execution
func TestExtensionLifecycle(t *testing.T) {
	tmpDir := t.TempDir()
	projectDir := filepath.Join(tmpDir, "project")
	extensionSrcDir := filepath.Join(tmpDir, "extension-src")

	// Setup test extension and project
	setupTestExtension(t, extensionSrcDir)
	configPath := setupTestProject(t, projectDir)

	// Change to project directory
	originalDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get current directory: %v", err)
	}
	defer func() {
		if err := os.Chdir(originalDir); err != nil {
			t.Logf("failed to restore directory: %v", err)
		}
	}()

	if err := os.Chdir(projectDir); err != nil {
		t.Fatalf("failed to change to project directory: %v", err)
	}

	// Install the extension
	if err := registerLocalExtension(extensionSrcDir, configPath, tmpDir); err != nil {
		t.Fatalf("failed to install extension: %v", err)
	}

	// Verify installation
	verifyExtensionInstalled(t, tmpDir, "test-lifecycle-ext")
	cfg := verifyConfigUpdated(t, configPath)

	// Execute hooks
	ctx := context.Background()
	runner := NewExtensionHookRunner(cfg)

	preBumpInput := HookInput{
		Hook:            "pre-bump",
		Version:         "1.2.3",
		PreviousVersion: "1.2.2",
		BumpType:        "patch",
		ProjectRoot:     projectDir,
	}

	if err := runner.RunHooks(ctx, PreBumpHook, &preBumpInput); err != nil {
		t.Errorf("failed to run pre-bump hooks: %v", err)
	}

	postBumpInput := HookInput{
		Hook:            "post-bump",
		Version:         "1.2.3",
		PreviousVersion: "1.2.2",
		BumpType:        "patch",
		ProjectRoot:     projectDir,
	}

	if err := runner.RunHooks(ctx, PostBumpHook, &postBumpInput); err != nil {
		t.Errorf("failed to run post-bump hooks: %v", err)
	}
}

// TestExtensionExecutionWithMultipleHooks tests an extension that supports multiple hooks
func TestExtensionExecutionWithMultipleHooks(t *testing.T) {
	tmpDir := t.TempDir()

	// Create extension with multiple hook support
	manifestContent := `name: multi-hook-ext
version: 1.0.0
description: Extension supporting multiple hooks
author: test
repository: https://github.com/test/multi-hook
entry: hook.sh
hooks:
  - pre-bump
  - post-bump
  - validate
`
	manifestPath := filepath.Join(tmpDir, "extension.yaml")
	if err := os.WriteFile(manifestPath, []byte(manifestContent), 0644); err != nil {
		t.Fatalf("failed to create manifest: %v", err)
	}

	// Create hook script that echoes back which hook was called
	hookScript := `#!/bin/sh
read input
hook=$(echo "$input" | grep -o '"hook":"[^"]*"' | cut -d'"' -f4)
echo "{\"success\": true, \"message\": \"Executed $hook hook\"}"
`
	hookPath := filepath.Join(tmpDir, "hook.sh")
	if err := os.WriteFile(hookPath, []byte(hookScript), 0755); err != nil {
		t.Fatalf("failed to create hook script: %v", err)
	}

	// Create config
	cfg := &config.Config{
		Extensions: []config.ExtensionConfig{
			{
				Name:    "multi-hook-ext",
				Path:    tmpDir,
				Enabled: true,
			},
		},
	}

	runner := NewExtensionHookRunner(cfg)
	ctx := context.Background()

	// Test each hook type
	hookTypes := []HookType{PreBumpHook, PostBumpHook, ValidateHook}
	for _, hookType := range hookTypes {
		input := HookInput{
			Hook:        string(hookType),
			Version:     "1.0.0",
			ProjectRoot: tmpDir,
		}

		if err := runner.RunHooks(ctx, hookType, &input); err != nil {
			t.Errorf("failed to run %s hook: %v", hookType, err)
		}
	}
}

// TestExtensionErrorHandling tests that extension errors are properly propagated
func TestExtensionErrorHandling(t *testing.T) {
	tmpDir := t.TempDir()

	// Create extension manifest
	manifestContent := `name: error-ext
version: 1.0.0
description: Extension that fails
author: test
repository: https://github.com/test/error-ext
entry: fail.sh
hooks:
  - pre-bump
`
	manifestPath := filepath.Join(tmpDir, "extension.yaml")
	if err := os.WriteFile(manifestPath, []byte(manifestContent), 0644); err != nil {
		t.Fatalf("failed to create manifest: %v", err)
	}

	// Create hook script that reports failure
	hookScript := `#!/bin/sh
read input
echo '{"success": false, "message": "Extension reported an error"}'
`
	hookPath := filepath.Join(tmpDir, "fail.sh")
	if err := os.WriteFile(hookPath, []byte(hookScript), 0755); err != nil {
		t.Fatalf("failed to create hook script: %v", err)
	}

	// Create config
	cfg := &config.Config{
		Extensions: []config.ExtensionConfig{
			{
				Name:    "error-ext",
				Path:    tmpDir,
				Enabled: true,
			},
		},
	}

	runner := NewExtensionHookRunner(cfg)
	ctx := context.Background()

	input := HookInput{
		Hook:        "pre-bump",
		Version:     "1.0.0",
		ProjectRoot: tmpDir,
	}

	err := runner.RunHooks(ctx, PreBumpHook, &input)
	if err == nil {
		t.Error("expected error from failing extension, got nil")
	}
}

// TestMultipleExtensionsExecution tests that multiple extensions are executed in order
func TestMultipleExtensionsExecution(t *testing.T) {
	tmpDir := t.TempDir()

	// Create first extension
	ext1Dir := filepath.Join(tmpDir, "ext1")
	if err := os.MkdirAll(ext1Dir, 0755); err != nil {
		t.Fatalf("failed to create ext1 directory: %v", err)
	}

	manifest1 := `name: ext1
version: 1.0.0
description: First extension
author: test
repository: https://github.com/test/ext1
entry: hook.sh
hooks:
  - post-bump
`
	if err := os.WriteFile(filepath.Join(ext1Dir, "extension.yaml"), []byte(manifest1), 0644); err != nil {
		t.Fatalf("failed to create manifest1: %v", err)
	}

	script1 := `#!/bin/sh
read input
echo '{"success": true, "message": "Extension 1 executed"}'
`
	if err := os.WriteFile(filepath.Join(ext1Dir, "hook.sh"), []byte(script1), 0755); err != nil {
		t.Fatalf("failed to create script1: %v", err)
	}

	// Create second extension
	ext2Dir := filepath.Join(tmpDir, "ext2")
	if err := os.MkdirAll(ext2Dir, 0755); err != nil {
		t.Fatalf("failed to create ext2 directory: %v", err)
	}

	manifest2 := `name: ext2
version: 1.0.0
description: Second extension
author: test
repository: https://github.com/test/ext2
entry: hook.sh
hooks:
  - post-bump
`
	if err := os.WriteFile(filepath.Join(ext2Dir, "extension.yaml"), []byte(manifest2), 0644); err != nil {
		t.Fatalf("failed to create manifest2: %v", err)
	}

	script2 := `#!/bin/sh
read input
echo '{"success": true, "message": "Extension 2 executed"}'
`
	if err := os.WriteFile(filepath.Join(ext2Dir, "hook.sh"), []byte(script2), 0755); err != nil {
		t.Fatalf("failed to create script2: %v", err)
	}

	// Create config with both extensions
	cfg := &config.Config{
		Extensions: []config.ExtensionConfig{
			{
				Name:    "ext1",
				Path:    ext1Dir,
				Enabled: true,
			},
			{
				Name:    "ext2",
				Path:    ext2Dir,
				Enabled: true,
			},
		},
	}

	runner := NewExtensionHookRunner(cfg)
	ctx := context.Background()

	input := HookInput{
		Hook:        "post-bump",
		Version:     "1.0.0",
		ProjectRoot: tmpDir,
	}

	// Both extensions should execute successfully
	if err := runner.RunHooks(ctx, PostBumpHook, &input); err != nil {
		t.Errorf("failed to run hooks for multiple extensions: %v", err)
	}
}
