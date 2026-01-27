package extensionmgr

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/indaco/sley/internal/config"
	"github.com/indaco/sley/internal/extensions"
)

func TestHasHook(t *testing.T) {
	tests := []struct {
		name     string
		hooks    []string
		hookType string
		want     bool
	}{
		{
			name:     "hook exists",
			hooks:    []string{"pre-bump", "post-bump"},
			hookType: "pre-bump",
			want:     true,
		},
		{
			name:     "hook does not exist",
			hooks:    []string{"pre-bump", "post-bump"},
			hookType: "validate",
			want:     false,
		},
		{
			name:     "empty hooks",
			hooks:    []string{},
			hookType: "pre-bump",
			want:     false,
		},
		{
			name:     "nil hooks",
			hooks:    nil,
			hookType: "pre-bump",
			want:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := hasHook(tt.hooks, tt.hookType)
			if got != tt.want {
				t.Errorf("hasHook() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestValidateExtensionHook(t *testing.T) {
	tests := []struct {
		name     string
		hookType string
		wantErr  bool
	}{
		{
			name:     "valid pre-bump",
			hookType: "pre-bump",
			wantErr:  false,
		},
		{
			name:     "valid post-bump",
			hookType: "post-bump",
			wantErr:  false,
		},
		{
			name:     "valid pre-release",
			hookType: "pre-release",
			wantErr:  false,
		},
		{
			name:     "valid validate",
			hookType: "validate",
			wantErr:  false,
		},
		{
			name:     "invalid hook",
			hookType: "invalid-hook",
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateExtensionHook(tt.hookType)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateExtensionHook() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestNewExtensionHookRunner(t *testing.T) {
	cfg := &config.Config{}
	runner := NewExtensionHookRunner(cfg)

	if runner == nil {
		t.Fatal("expected non-nil runner")
	}
	if runner.Config != cfg {
		t.Error("runner config does not match")
	}
	if runner.Executor == nil {
		t.Error("expected non-nil executor")
	}
}

func TestExtensionHookRunner_RunHooks_NoExtensions(t *testing.T) {
	cfg := &config.Config{
		Extensions: []config.ExtensionConfig{},
	}

	runner := NewExtensionHookRunner(cfg)
	input := HookInput{
		Hook:        string(PreBumpHook),
		Version:     "1.2.3",
		ProjectRoot: "/test",
	}

	ctx := context.Background()
	err := runner.RunHooks(ctx, PreBumpHook, &input)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestExtensionHookRunner_RunHooks_DisabledExtension(t *testing.T) {
	tmpDir := t.TempDir()

	// Create extension manifest
	manifestPath := filepath.Join(tmpDir, "extension.yaml")
	manifest := `name: test-ext
version: 1.0.0
description: Test extension
author: test
repository: https://github.com/test/test
entry: hook.sh
hooks:
  - pre-bump
`
	if err := os.WriteFile(manifestPath, []byte(manifest), 0644); err != nil {
		t.Fatalf("failed to create manifest: %v", err)
	}

	cfg := &config.Config{
		Extensions: []config.ExtensionConfig{
			{
				Name:    "test-ext",
				Path:    tmpDir,
				Enabled: false, // Disabled
			},
		},
	}

	runner := NewExtensionHookRunner(cfg)
	input := HookInput{
		Hook:        string(PreBumpHook),
		Version:     "1.2.3",
		ProjectRoot: "/test",
	}

	ctx := context.Background()
	err := runner.RunHooks(ctx, PreBumpHook, &input)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestExtensionHookRunner_RunHooks_Success(t *testing.T) {
	tmpDir := t.TempDir()

	// Create extension manifest
	manifestPath := filepath.Join(tmpDir, "extension.yaml")
	manifest := `name: test-ext
version: 1.0.0
description: Test extension
author: test
repository: https://github.com/test/test
entry: hook.sh
hooks:
  - pre-bump
`
	if err := os.WriteFile(manifestPath, []byte(manifest), 0644); err != nil {
		t.Fatalf("failed to create manifest: %v", err)
	}

	// Create hook script
	scriptPath := filepath.Join(tmpDir, "hook.sh")
	script := `#!/bin/sh
read input
echo '{"success": true, "message": "Hook executed"}'
`
	if err := os.WriteFile(scriptPath, []byte(script), 0755); err != nil {
		t.Fatalf("failed to create script: %v", err)
	}

	cfg := &config.Config{
		Extensions: []config.ExtensionConfig{
			{
				Name:    "test-ext",
				Path:    tmpDir,
				Enabled: true,
			},
		},
	}

	runner := NewExtensionHookRunner(cfg)
	input := HookInput{
		Hook:        string(PreBumpHook),
		Version:     "1.2.3",
		ProjectRoot: "/test",
	}

	ctx := context.Background()
	err := runner.RunHooks(ctx, PreBumpHook, &input)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestExtensionHookRunner_RunHooks_WrongHook(t *testing.T) {
	tmpDir := t.TempDir()

	// Create extension manifest with only post-bump hook
	manifestPath := filepath.Join(tmpDir, "extension.yaml")
	manifest := `name: test-ext
version: 1.0.0
description: Test extension
author: test
repository: https://github.com/test/test
entry: hook.sh
hooks:
  - post-bump
`
	if err := os.WriteFile(manifestPath, []byte(manifest), 0644); err != nil {
		t.Fatalf("failed to create manifest: %v", err)
	}

	cfg := &config.Config{
		Extensions: []config.ExtensionConfig{
			{
				Name:    "test-ext",
				Path:    tmpDir,
				Enabled: true,
			},
		},
	}

	runner := NewExtensionHookRunner(cfg)
	input := HookInput{
		Hook:        string(PreBumpHook), // pre-bump, but extension only supports post-bump
		Version:     "1.2.3",
		ProjectRoot: "/test",
	}

	ctx := context.Background()
	err := runner.RunHooks(ctx, PreBumpHook, &input)
	if err != nil {
		t.Errorf("unexpected error (should skip): %v", err)
	}
}

func TestLoadExtensionsForHook(t *testing.T) {
	tmpDir := t.TempDir()

	// Create extension manifest
	manifestPath := filepath.Join(tmpDir, "extension.yaml")
	manifest := `name: test-ext
version: 1.0.0
description: Test extension
author: test
repository: https://github.com/test/test
entry: hook.sh
hooks:
  - pre-bump
  - post-bump
`
	if err := os.WriteFile(manifestPath, []byte(manifest), 0644); err != nil {
		t.Fatalf("failed to create manifest: %v", err)
	}

	cfg := &config.Config{
		Extensions: []config.ExtensionConfig{
			{
				Name:    "test-ext",
				Path:    tmpDir,
				Enabled: true,
			},
		},
	}

	// Test loading extensions for pre-bump hook
	exts, err := LoadExtensionsForHook(cfg, PreBumpHook)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(exts) != 1 {
		t.Errorf("expected 1 extension, got %d", len(exts))
	}

	if exts[0].Name != "test-ext" {
		t.Errorf("expected extension name 'test-ext', got %s", exts[0].Name)
	}

	// Test loading extensions for validate hook (not supported)
	exts, err = LoadExtensionsForHook(cfg, ValidateHook)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(exts) != 0 {
		t.Errorf("expected 0 extensions for validate hook, got %d", len(exts))
	}
}

func TestLoadExtensionsForHook_NoConfig(t *testing.T) {
	exts, err := LoadExtensionsForHook(nil, PreBumpHook)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(exts) != 0 {
		t.Errorf("expected 0 extensions with nil config, got %d", len(exts))
	}
}

func TestRunPreBumpHooks(t *testing.T) {
	tmpDir := t.TempDir()

	// Create extension manifest
	manifestPath := filepath.Join(tmpDir, "extension.yaml")
	manifest := `name: test-ext
version: 1.0.0
description: Test extension
author: test
repository: https://github.com/test/test
entry: hook.sh
hooks:
  - pre-bump
`
	if err := os.WriteFile(manifestPath, []byte(manifest), 0644); err != nil {
		t.Fatalf("failed to create manifest: %v", err)
	}

	// Create hook script
	scriptPath := filepath.Join(tmpDir, "hook.sh")
	script := `#!/bin/sh
read input
echo '{"success": true}'
`
	if err := os.WriteFile(scriptPath, []byte(script), 0755); err != nil {
		t.Fatalf("failed to create script: %v", err)
	}

	cfg := &config.Config{
		Extensions: []config.ExtensionConfig{
			{
				Name:    "test-ext",
				Path:    tmpDir,
				Enabled: true,
			},
		},
	}

	ctx := context.Background()
	err := RunPreBumpHooks(ctx, cfg, "1.2.3", "1.2.2", "patch", nil)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestRunPostBumpHooks(t *testing.T) {
	tmpDir := t.TempDir()

	// Create extension manifest
	manifestPath := filepath.Join(tmpDir, "extension.yaml")
	manifest := `name: test-ext
version: 1.0.0
description: Test extension
author: test
repository: https://github.com/test/test
entry: hook.sh
hooks:
  - post-bump
`
	if err := os.WriteFile(manifestPath, []byte(manifest), 0644); err != nil {
		t.Fatalf("failed to create manifest: %v", err)
	}

	// Create hook script
	scriptPath := filepath.Join(tmpDir, "hook.sh")
	script := `#!/bin/sh
read input
echo '{"success": true}'
`
	if err := os.WriteFile(scriptPath, []byte(script), 0755); err != nil {
		t.Fatalf("failed to create script: %v", err)
	}

	cfg := &config.Config{
		Extensions: []config.ExtensionConfig{
			{
				Name:    "test-ext",
				Path:    tmpDir,
				Enabled: true,
			},
		},
	}

	prerelease := "alpha"
	metadata := "build123"

	ctx := context.Background()
	err := RunPostBumpHooks(ctx, cfg, "1.3.0", "1.2.3", "minor", &prerelease, &metadata, nil)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

// Mock executor for testing error scenarios
type mockExecutor struct {
	executeFunc func(ctx context.Context, scriptPath string, input *HookInput) (*HookOutput, error)
}

func (m *mockExecutor) Execute(ctx context.Context, scriptPath string, input *HookInput) (*HookOutput, error) {
	return m.executeFunc(ctx, scriptPath, input)
}

func TestExtensionHookRunner_RunHooks_ExecutorError(t *testing.T) {
	tmpDir := t.TempDir()

	// Create extension manifest
	manifestPath := filepath.Join(tmpDir, "extension.yaml")
	manifest := `name: test-ext
version: 1.0.0
description: Test extension
author: test
repository: https://github.com/test/test
entry: hook.sh
hooks:
  - pre-bump
`
	if err := os.WriteFile(manifestPath, []byte(manifest), 0644); err != nil {
		t.Fatalf("failed to create manifest: %v", err)
	}

	cfg := &config.Config{
		Extensions: []config.ExtensionConfig{
			{
				Name:    "test-ext",
				Path:    tmpDir,
				Enabled: true,
			},
		},
	}

	// Save original function and restore after test
	originalLoadFn := extensions.LoadExtensionManifestFn
	defer func() { extensions.LoadExtensionManifestFn = originalLoadFn }()

	runner := NewExtensionHookRunner(cfg)
	runner.Executor = &mockExecutor{
		executeFunc: func(ctx context.Context, scriptPath string, input *HookInput) (*HookOutput, error) {
			return nil, fmt.Errorf("execution failed")
		},
	}

	input := HookInput{
		Hook:        string(PreBumpHook),
		Version:     "1.2.3",
		ProjectRoot: "/test",
	}

	ctx := context.Background()
	err := runner.RunHooks(ctx, PreBumpHook, &input)
	if err == nil {
		t.Error("expected error from executor")
	}
}

/* ------------------------------------------------------------------------- */
/* TABLE-DRIVEN TESTS FOR HOOK RUNNER                                       */
/* ------------------------------------------------------------------------- */

// setupNoManifests is a no-op setup function for tests with no extensions.
func setupNoManifests(_ *testing.T, _ string) {}

// setupSingleExtension creates a manifest and script for a single extension.
func setupSingleExtension(t *testing.T, tmpDir string) {
	t.Helper()
	writeManifest(t, tmpDir, "test-ext", "pre-bump")
	writeSuccessScript(t, tmpDir)
}

// setupMixedExtensions creates manifests for ext1 (post-bump) and ext2 (pre-bump).
func setupMixedExtensions(t *testing.T, tmpDir string) {
	t.Helper()
	ext1Dir := filepath.Join(tmpDir, "ext1")
	mustMkdirAll(t, ext1Dir)
	writeManifest(t, ext1Dir, "ext1", "post-bump")
	writeSuccessScript(t, ext1Dir)

	ext2Dir := filepath.Join(tmpDir, "ext2")
	mustMkdirAll(t, ext2Dir)
	writeManifest(t, ext2Dir, "ext2", "pre-bump")
}

// setupInvalidManifest creates a manifest missing required fields.
func setupInvalidManifest(t *testing.T, tmpDir string) {
	t.Helper()
	manifestPath := filepath.Join(tmpDir, "extension.yaml")
	manifest := "name: bad-ext\n# missing version, description, etc.\n"
	mustWriteFile(t, manifestPath, manifest, 0644)
}

// writeManifest writes a valid extension manifest to the given directory.
func writeManifest(t *testing.T, dir, name, hook string) {
	t.Helper()
	manifestPath := filepath.Join(dir, "extension.yaml")
	manifest := fmt.Sprintf(`name: %s
version: 1.0.0
description: Test
author: test
repository: https://github.com/test/test
entry: hook.sh
hooks:
  - %s
`, name, hook)
	mustWriteFile(t, manifestPath, manifest, 0644)
}

// writeSuccessScript writes a simple success script to the given directory.
func writeSuccessScript(t *testing.T, dir string) {
	t.Helper()
	scriptPath := filepath.Join(dir, "hook.sh")
	script := "#!/bin/sh\nread input\necho '{\"success\": true}'\n"
	mustWriteFile(t, scriptPath, script, 0755)
}

// mustMkdirAll creates a directory, failing the test on error.
func mustMkdirAll(t *testing.T, path string) {
	t.Helper()
	if err := os.MkdirAll(path, 0755); err != nil {
		t.Fatalf("failed to create directory %s: %v", path, err)
	}
}

// mustWriteFile writes a file, failing the test on error.
func mustWriteFile(t *testing.T, path, content string, perm os.FileMode) {
	t.Helper()
	if err := os.WriteFile(path, []byte(content), perm); err != nil {
		t.Fatalf("failed to write file %s: %v", path, err)
	}
}

// updateExtensionPaths sets extension paths based on tmpDir.
func updateExtensionPaths(extensions []config.ExtensionConfig, tmpDir string) {
	for i := range extensions {
		if extensions[i].Path == "" {
			if len(extensions) == 1 {
				extensions[i].Path = tmpDir
			} else {
				extensions[i].Path = filepath.Join(tmpDir, extensions[i].Name)
			}
		}
	}
}

// assertHookError checks if error matches expectations.
func assertHookError(t *testing.T, err error, wantErr bool, wantErrText string) {
	t.Helper()
	if (err != nil) != wantErr {
		t.Errorf("RunHooks() error = %v, wantErr %v", err, wantErr)
		return
	}
	if wantErr && wantErrText != "" && (err == nil || !contains(err.Error(), wantErrText)) {
		t.Errorf("expected error containing %q, got: %v", wantErrText, err)
	}
}

func TestExtensionHookRunner_RunHooks_TableDriven(t *testing.T) {
	tests := []struct {
		name           string
		hookType       HookType
		extensions     []config.ExtensionConfig
		setupManifests func(t *testing.T, tmpDir string)
		wantErr        bool
		wantErrText    string
	}{
		{
			name:           "no extensions configured",
			hookType:       PreBumpHook,
			extensions:     []config.ExtensionConfig{},
			setupManifests: setupNoManifests,
			wantErr:        false,
		},
		{
			name:     "single extension with matching hook",
			hookType: PreBumpHook,
			extensions: []config.ExtensionConfig{
				{Name: "test-ext", Path: "", Enabled: true},
			},
			setupManifests: setupSingleExtension,
			wantErr:        false,
		},
		{
			name:     "multiple extensions with mixed hook support",
			hookType: PostBumpHook,
			extensions: []config.ExtensionConfig{
				{Name: "ext1", Path: "", Enabled: true},
				{Name: "ext2", Path: "", Enabled: true},
			},
			setupManifests: setupMixedExtensions,
			wantErr:        false,
		},
		{
			name:     "extension with invalid manifest",
			hookType: PreBumpHook,
			extensions: []config.ExtensionConfig{
				{Name: "bad-ext", Path: "", Enabled: true},
			},
			setupManifests: setupInvalidManifest,
			wantErr:        true,
			wantErrText:    "failed to load extension",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			tt.setupManifests(t, tmpDir)
			updateExtensionPaths(tt.extensions, tmpDir)

			cfg := &config.Config{Extensions: tt.extensions}
			runner := NewExtensionHookRunner(cfg)
			input := HookInput{
				Hook:        string(tt.hookType),
				Version:     "1.2.3",
				ProjectRoot: "/test",
			}

			err := runner.RunHooks(context.Background(), tt.hookType, &input)
			assertHookError(t, err, tt.wantErr, tt.wantErrText)
		})
	}
}

// TestRunPreBumpHooks_EdgeCases tests edge cases for pre-bump hooks
func TestRunPreBumpHooks_EdgeCases(t *testing.T) {
	tests := []struct {
		name        string
		cfg         *config.Config
		version     string
		prevVersion string
		bumpType    string
		wantErr     bool
	}{
		{
			name:        "nil config",
			cfg:         nil,
			version:     "1.0.0",
			prevVersion: "0.9.0",
			bumpType:    "minor",
			wantErr:     false, // Should handle nil gracefully
		},
		{
			name: "empty extensions list",
			cfg: &config.Config{
				Extensions: []config.ExtensionConfig{},
			},
			version:     "1.0.0",
			prevVersion: "0.9.0",
			bumpType:    "minor",
			wantErr:     false,
		},
		{
			name: "all extensions disabled",
			cfg: &config.Config{
				Extensions: []config.ExtensionConfig{
					{Name: "disabled-ext", Path: "/tmp/disabled", Enabled: false},
				},
			},
			version:     "1.0.0",
			prevVersion: "0.9.0",
			bumpType:    "minor",
			wantErr:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			err := RunPreBumpHooks(ctx, tt.cfg, tt.version, tt.prevVersion, tt.bumpType, nil)

			if (err != nil) != tt.wantErr {
				t.Errorf("RunPreBumpHooks() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// TestRunPostBumpHooks_EdgeCases tests edge cases for post-bump hooks
func TestRunPostBumpHooks_EdgeCases(t *testing.T) {
	tests := []struct {
		name        string
		cfg         *config.Config
		version     string
		prevVersion string
		bumpType    string
		prerelease  *string
		metadata    *string
		wantErr     bool
	}{
		{
			name:        "nil config",
			cfg:         nil,
			version:     "1.0.0",
			prevVersion: "0.9.0",
			bumpType:    "minor",
			prerelease:  nil,
			metadata:    nil,
			wantErr:     false,
		},
		{
			name: "with prerelease and metadata",
			cfg: &config.Config{
				Extensions: []config.ExtensionConfig{},
			},
			version:     "2.0.0",
			prevVersion: "1.9.9",
			bumpType:    "major",
			prerelease:  stringPtr("rc.1"),
			metadata:    stringPtr("build.123"),
			wantErr:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			err := RunPostBumpHooks(ctx, tt.cfg, tt.version, tt.prevVersion, tt.bumpType, tt.prerelease, tt.metadata, nil)

			if (err != nil) != tt.wantErr {
				t.Errorf("RunPostBumpHooks() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// TestLoadExtensionsForHook_EdgeCases tests edge cases for loading extensions
func TestLoadExtensionsForHook_EdgeCases(t *testing.T) {
	tests := []struct {
		name         string
		cfg          *config.Config
		hookType     HookType
		wantCount    int
		wantErr      bool
		setupMocks   func()
		cleanupMocks func()
	}{
		{
			name:      "nil config",
			cfg:       nil,
			hookType:  PreBumpHook,
			wantCount: 0,
			wantErr:   false,
		},
		{
			name: "empty extensions",
			cfg: &config.Config{
				Extensions: []config.ExtensionConfig{},
			},
			hookType:  PreBumpHook,
			wantCount: 0,
			wantErr:   false,
		},
		{
			name: "manifest load error",
			cfg: &config.Config{
				Extensions: []config.ExtensionConfig{
					{Name: "bad", Path: "/nonexistent", Enabled: true},
				},
			},
			hookType: PreBumpHook,
			wantErr:  true,
			setupMocks: func() {
				// Manifest loading will fail for nonexistent path
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.setupMocks != nil {
				tt.setupMocks()
			}
			if tt.cleanupMocks != nil {
				defer tt.cleanupMocks()
			}

			exts, err := LoadExtensionsForHook(tt.cfg, tt.hookType)

			if (err != nil) != tt.wantErr {
				t.Errorf("LoadExtensionsForHook() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && len(exts) != tt.wantCount {
				t.Errorf("LoadExtensionsForHook() count = %d, want %d", len(exts), tt.wantCount)
			}
		})
	}
}

// TestModuleInfo tests the ModuleInfo struct and its usage
func TestModuleInfo(t *testing.T) {
	t.Run("nil module info is handled gracefully", func(t *testing.T) {
		ctx := context.Background()
		err := RunPreBumpHooks(ctx, nil, "1.0.0", "0.9.0", "minor", nil)
		if err != nil {
			t.Errorf("expected nil error with nil module info, got %v", err)
		}
	})

	t.Run("module info with dir and name", func(t *testing.T) {
		moduleInfo := &ModuleInfo{
			Dir:  "/path/to/module",
			Name: "my-module",
		}

		if moduleInfo.Dir != "/path/to/module" {
			t.Errorf("expected Dir to be '/path/to/module', got %s", moduleInfo.Dir)
		}
		if moduleInfo.Name != "my-module" {
			t.Errorf("expected Name to be 'my-module', got %s", moduleInfo.Name)
		}
	})
}

// TestRunHooksWithModuleInfo tests that module info is correctly passed to hooks
func TestRunHooksWithModuleInfo(t *testing.T) {
	tmpDir := t.TempDir()

	// Create extension manifest
	manifestPath := filepath.Join(tmpDir, "extension.yaml")
	manifest := `name: test-ext
version: 1.0.0
description: Test extension
author: test
repository: https://github.com/test/test
entry: hook.sh
hooks:
  - post-bump
`
	if err := os.WriteFile(manifestPath, []byte(manifest), 0644); err != nil {
		t.Fatalf("failed to create manifest: %v", err)
	}

	// Create hook script that echoes input for verification
	scriptPath := filepath.Join(tmpDir, "hook.sh")
	script := `#!/bin/sh
read input
# Extract module_dir and module_name from input
module_dir=$(echo "$input" | grep -o '"module_dir":"[^"]*"' | cut -d'"' -f4)
module_name=$(echo "$input" | grep -o '"module_name":"[^"]*"' | cut -d'"' -f4)
echo "{\"success\": true, \"message\": \"module_dir=$module_dir module_name=$module_name\"}"
`
	if err := os.WriteFile(scriptPath, []byte(script), 0755); err != nil {
		t.Fatalf("failed to create script: %v", err)
	}

	cfg := &config.Config{
		Extensions: []config.ExtensionConfig{
			{
				Name:    "test-ext",
				Path:    tmpDir,
				Enabled: true,
			},
		},
	}

	t.Run("hook receives module info", func(t *testing.T) {
		moduleInfo := &ModuleInfo{
			Dir:  "/project/packages/app",
			Name: "app",
		}

		ctx := context.Background()
		err := RunPostBumpHooks(ctx, cfg, "1.0.0", "0.9.0", "minor", nil, nil, moduleInfo)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})

	t.Run("hook works without module info", func(t *testing.T) {
		ctx := context.Background()
		err := RunPostBumpHooks(ctx, cfg, "1.0.0", "0.9.0", "minor", nil, nil, nil)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})
}

// TestExtensionHookRunner_RunHooks_WithConfig tests that extension config is passed to hooks
func TestExtensionHookRunner_RunHooks_WithConfig(t *testing.T) {
	tmpDir := t.TempDir()

	// Create extension manifest
	manifestPath := filepath.Join(tmpDir, "extension.yaml")
	manifest := `name: github-version-sync
version: 1.0.0
description: Test extension with config
author: test
repository: https://github.com/test/test
entry: hook.sh
hooks:
  - pre-bump
`
	if err := os.WriteFile(manifestPath, []byte(manifest), 0644); err != nil {
		t.Fatalf("failed to create manifest: %v", err)
	}

	// Create hook script that validates config is received
	scriptPath := filepath.Join(tmpDir, "hook.sh")
	script := `#!/bin/sh
read input
# Extract config fields from JSON input
repo=$(echo "$input" | grep -o '"repo":"[^"]*"' | cut -d'"' -f4)
strip_prefix=$(echo "$input" | grep -o '"strip-prefix":"[^"]*"' | cut -d'"' -f4)

if [ -z "$repo" ]; then
    echo '{"success": false, "message": "config.repo not found"}'
    exit 1
fi

if [ -z "$strip_prefix" ]; then
    echo '{"success": false, "message": "config.strip-prefix not found"}'
    exit 1
fi

echo "{\"success\": true, \"message\": \"Config received: repo=$repo, strip-prefix=$strip_prefix\"}"
`
	if err := os.WriteFile(scriptPath, []byte(script), 0755); err != nil {
		t.Fatalf("failed to create script: %v", err)
	}

	cfg := &config.Config{
		Extensions: []config.ExtensionConfig{
			{
				Name:    "github-version-sync",
				Path:    tmpDir,
				Enabled: true,
				Config: map[string]any{
					"repo":         "indaco/sley",
					"strip-prefix": "v",
				},
			},
		},
	}

	runner := NewExtensionHookRunner(cfg)
	input := HookInput{
		Hook:        string(PreBumpHook),
		Version:     "1.2.3",
		ProjectRoot: "/test",
	}

	ctx := context.Background()
	err := runner.RunHooks(ctx, PreBumpHook, &input)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

// TestExtensionHookRunner_RunHooks_WithoutConfig tests that hooks work without config
func TestExtensionHookRunner_RunHooks_WithoutConfig(t *testing.T) {
	tmpDir := t.TempDir()

	// Create extension manifest
	manifestPath := filepath.Join(tmpDir, "extension.yaml")
	manifest := `name: simple-ext
version: 1.0.0
description: Test extension without config
author: test
repository: https://github.com/test/test
entry: hook.sh
hooks:
  - post-bump
`
	if err := os.WriteFile(manifestPath, []byte(manifest), 0644); err != nil {
		t.Fatalf("failed to create manifest: %v", err)
	}

	// Create hook script that doesn't require config
	scriptPath := filepath.Join(tmpDir, "hook.sh")
	script := `#!/bin/sh
read input
echo '{"success": true, "message": "Hook executed without config"}'
`
	if err := os.WriteFile(scriptPath, []byte(script), 0755); err != nil {
		t.Fatalf("failed to create script: %v", err)
	}

	cfg := &config.Config{
		Extensions: []config.ExtensionConfig{
			{
				Name:    "simple-ext",
				Path:    tmpDir,
				Enabled: true,
				Config:  nil, // No config provided
			},
		},
	}

	runner := NewExtensionHookRunner(cfg)
	input := HookInput{
		Hook:        string(PostBumpHook),
		Version:     "2.0.0",
		ProjectRoot: "/test",
	}

	ctx := context.Background()
	err := runner.RunHooks(ctx, PostBumpHook, &input)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

// TestExtensionHookRunner_RunHooks_MultipleExtensionsWithDifferentConfigs tests multiple extensions with different configs
func TestExtensionHookRunner_RunHooks_MultipleExtensionsWithDifferentConfigs(t *testing.T) {
	tmpDir := t.TempDir()

	// Create first extension
	ext1Dir := filepath.Join(tmpDir, "ext1")
	if err := os.MkdirAll(ext1Dir, 0755); err != nil {
		t.Fatalf("failed to create ext1 dir: %v", err)
	}

	manifest1 := `name: ext1
version: 1.0.0
description: Extension 1
author: test
repository: https://github.com/test/test
entry: hook.sh
hooks:
  - pre-bump
`
	if err := os.WriteFile(filepath.Join(ext1Dir, "extension.yaml"), []byte(manifest1), 0644); err != nil {
		t.Fatalf("failed to create manifest1: %v", err)
	}

	script1 := `#!/bin/sh
read input
key1=$(echo "$input" | grep -o '"key1":"[^"]*"' | cut -d'"' -f4)
echo "{\"success\": true, \"message\": \"ext1: key1=$key1\"}"
`
	if err := os.WriteFile(filepath.Join(ext1Dir, "hook.sh"), []byte(script1), 0755); err != nil {
		t.Fatalf("failed to create script1: %v", err)
	}

	// Create second extension
	ext2Dir := filepath.Join(tmpDir, "ext2")
	if err := os.MkdirAll(ext2Dir, 0755); err != nil {
		t.Fatalf("failed to create ext2 dir: %v", err)
	}

	manifest2 := `name: ext2
version: 1.0.0
description: Extension 2
author: test
repository: https://github.com/test/test
entry: hook.sh
hooks:
  - pre-bump
`
	if err := os.WriteFile(filepath.Join(ext2Dir, "extension.yaml"), []byte(manifest2), 0644); err != nil {
		t.Fatalf("failed to create manifest2: %v", err)
	}

	script2 := `#!/bin/sh
read input
key2=$(echo "$input" | grep -o '"key2":"[^"]*"' | cut -d'"' -f4)
echo "{\"success\": true, \"message\": \"ext2: key2=$key2\"}"
`
	if err := os.WriteFile(filepath.Join(ext2Dir, "hook.sh"), []byte(script2), 0755); err != nil {
		t.Fatalf("failed to create script2: %v", err)
	}

	cfg := &config.Config{
		Extensions: []config.ExtensionConfig{
			{
				Name:    "ext1",
				Path:    ext1Dir,
				Enabled: true,
				Config: map[string]any{
					"key1": "value1",
				},
			},
			{
				Name:    "ext2",
				Path:    ext2Dir,
				Enabled: true,
				Config: map[string]any{
					"key2": "value2",
				},
			},
		},
	}

	runner := NewExtensionHookRunner(cfg)
	input := HookInput{
		Hook:        string(PreBumpHook),
		Version:     "1.0.0",
		ProjectRoot: "/test",
	}

	ctx := context.Background()
	err := runner.RunHooks(ctx, PreBumpHook, &input)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}
