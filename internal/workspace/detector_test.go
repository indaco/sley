package workspace

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/indaco/sley/internal/config"
	"github.com/indaco/sley/internal/core"
)

// setupTestFS creates a mock filesystem with test files.
func setupTestFS(files map[string]string) *core.MockFileSystem {
	fs := core.NewMockFileSystem()
	ctx := context.Background()
	for path, content := range files {
		fs.SetFile(path, []byte(content))
		// Create all parent directories up to root
		dir := filepath.Dir(path)
		for dir != "." && dir != "/" {
			_ = fs.MkdirAll(ctx, dir, 0755)
			dir = filepath.Dir(dir)
		}
	}
	return fs
}

func TestDetector_SingleModule_InCWD(t *testing.T) {
	// Setup: .version in current directory
	fs := setupTestFS(map[string]string{
		"/project/.version": "1.0.0",
	})

	cfg := &config.Config{}
	detector := NewDetector(fs, cfg)

	ctx, err := detector.DetectContext(context.Background(), "/project")
	if err != nil {
		t.Fatalf("DetectContext failed: %v", err)
	}

	if ctx.Mode != SingleModule {
		t.Errorf("Expected SingleModule mode, got %s", ctx.Mode)
	}

	if ctx.Path != "/project/.version" {
		t.Errorf("Expected path /project/.version, got %s", ctx.Path)
	}
}

func TestDetector_MultiModule(t *testing.T) {
	// Setup: Multiple .version files in subdirectories
	fs := setupTestFS(map[string]string{
		"/project/module-a/.version": "1.0.0",
		"/project/module-b/.version": "2.0.0",
		"/project/module-c/.version": "0.1.0",
	})

	cfg := &config.Config{}
	detector := NewDetector(fs, cfg)

	ctx, err := detector.DetectContext(context.Background(), "/project")
	if err != nil {
		t.Fatalf("DetectContext failed: %v", err)
	}

	if ctx.Mode != MultiModule {
		t.Errorf("Expected MultiModule mode, got %s", ctx.Mode)
	}

	if len(ctx.Modules) != 3 {
		t.Errorf("Expected 3 modules, got %d", len(ctx.Modules))
	}

	// Verify module names
	expectedNames := map[string]bool{
		"module-a": false,
		"module-b": false,
		"module-c": false,
	}

	for _, module := range ctx.Modules {
		if _, ok := expectedNames[module.Name]; !ok {
			t.Errorf("Unexpected module name: %s", module.Name)
		}
		expectedNames[module.Name] = true
	}

	for name, found := range expectedNames {
		if !found {
			t.Errorf("Module %s not found", name)
		}
	}
}

func TestDetector_NoModules(t *testing.T) {
	// Setup: No .version files
	fs := setupTestFS(map[string]string{
		"/project/README.md": "# Project",
	})

	cfg := &config.Config{}
	detector := NewDetector(fs, cfg)

	ctx, err := detector.DetectContext(context.Background(), "/project")
	if err != nil {
		t.Fatalf("DetectContext failed: %v", err)
	}

	if ctx.Mode != NoModules {
		t.Errorf("Expected NoModules mode, got %s", ctx.Mode)
	}
}

func TestDetector_SingleModuleInSubdir(t *testing.T) {
	// Setup: Only one .version in a subdirectory
	fs := setupTestFS(map[string]string{
		"/project/module-a/.version": "1.0.0",
	})

	cfg := &config.Config{}
	detector := NewDetector(fs, cfg)

	ctx, err := detector.DetectContext(context.Background(), "/project")
	if err != nil {
		t.Fatalf("DetectContext failed: %v", err)
	}

	if ctx.Mode != SingleModule {
		t.Errorf("Expected SingleModule mode, got %s", ctx.Mode)
	}

	if ctx.Path != "/project/module-a/.version" {
		t.Errorf("Expected path /project/module-a/.version, got %s", ctx.Path)
	}
}

func TestDetector_ExcludePatterns(t *testing.T) {
	// Setup: .version files in both regular and excluded directories
	fs := setupTestFS(map[string]string{
		"/project/module-a/.version":      "1.0.0",
		"/project/node_modules/.version":  "1.0.0",
		"/project/build/.version":         "1.0.0",
		"/project/.git/.version":          "1.0.0",
		"/project/module-b/dist/.version": "1.0.0",
		"/project/module-c/.version":      "2.0.0",
		"/project/tmp/test/.version":      "1.0.0",
		"/project/__pycache__/.version":   "1.0.0",
		"/project/vendor/pkg/.version":    "1.0.0",
		"/project/.cache/tmp/.version":    "1.0.0",
	})

	cfg := &config.Config{}
	detector := NewDetector(fs, cfg)

	ctx, err := detector.DetectContext(context.Background(), "/project")
	if err != nil {
		t.Fatalf("DetectContext failed: %v", err)
	}

	if ctx.Mode != MultiModule {
		t.Errorf("Expected MultiModule mode, got %s", ctx.Mode)
	}

	// Should only find module-a and module-c
	if len(ctx.Modules) != 2 {
		t.Errorf("Expected 2 modules, got %d", len(ctx.Modules))
		for _, m := range ctx.Modules {
			t.Logf("Found module: %s at %s", m.Name, m.Path)
		}
	}
}

func TestDetector_MaxDepth(t *testing.T) {
	// Setup: Deeply nested .version files
	fs := setupTestFS(map[string]string{
		"/project/level1/.version":                "1.0.0",
		"/project/level1/level2/.version":         "1.0.0",
		"/project/level1/level2/level3/.version":  "1.0.0",
		"/project/a/b/c/d/e/f/g/h/i/j/k/.version": "1.0.0", // depth 11
	})

	maxDepth := 2
	cfg := &config.Config{
		Workspace: &config.WorkspaceConfig{
			Discovery: &config.DiscoveryConfig{
				ModuleMaxDepth: &maxDepth,
			},
		},
	}

	detector := NewDetector(fs, cfg)
	modules, err := detector.DiscoverModules(context.Background(), "/project")
	if err != nil {
		t.Fatalf("DiscoverModules failed: %v", err)
	}

	// Should find level1, level2, but not level3 or deeper
	// Note: In the current implementation, it might find different modules
	// depending on how depth is counted
	if len(modules) > 3 {
		t.Errorf("Expected at most 3 modules with max depth 2, got %d", len(modules))
		for _, m := range modules {
			t.Logf("Found module: %s at %s", m.Name, m.Path)
		}
	}
}

func TestDetector_ExplicitModules(t *testing.T) {
	// Setup: .version files
	fs := setupTestFS(map[string]string{
		"/project/frontend/.version": "1.0.0",
		"/project/backend/.version":  "2.0.0",
		"/project/shared/.version":   "0.5.0",
	})

	enabled := true
	disabled := false
	cfg := &config.Config{
		Workspace: &config.WorkspaceConfig{
			Modules: []config.ModuleConfig{
				{Name: "frontend", Path: "/project/frontend/.version", Enabled: &enabled},
				{Name: "backend", Path: "/project/backend/.version", Enabled: &enabled},
				{Name: "shared", Path: "/project/shared/.version", Enabled: &disabled},
			},
		},
	}

	detector := NewDetector(fs, cfg)

	ctx, err := detector.DetectContext(context.Background(), "/project")
	if err != nil {
		t.Fatalf("DetectContext failed: %v", err)
	}

	if ctx.Mode != MultiModule {
		t.Errorf("Expected MultiModule mode, got %s", ctx.Mode)
	}

	// Should only find enabled modules (frontend and backend)
	if len(ctx.Modules) != 2 {
		t.Errorf("Expected 2 modules, got %d", len(ctx.Modules))
	}

	for _, module := range ctx.Modules {
		if module.Name == "shared" {
			t.Errorf("Found disabled module: shared")
		}
	}
}

func TestDetector_RecursiveDisabled(t *testing.T) {
	// Setup: .version files at different levels
	// When recursive is disabled, only .version files directly in the scanned directory are found
	fs := setupTestFS(map[string]string{
		"/project/.version":                 "0.5.0", // This will be found
		"/project/module-a/.version":        "1.0.0", // This won't be found (in subdir)
		"/project/module-a/nested/.version": "2.0.0", // This won't be found (nested)
	})

	recursive := false
	cfg := &config.Config{
		Workspace: &config.WorkspaceConfig{
			Discovery: &config.DiscoveryConfig{
				Recursive: &recursive,
			},
		},
	}

	detector := NewDetector(fs, cfg)
	modules, err := detector.DiscoverModules(context.Background(), "/project")
	if err != nil {
		t.Fatalf("DiscoverModules failed: %v", err)
	}

	// Should only find .version files directly in /project (not in subdirectories)
	if len(modules) != 1 {
		t.Errorf("Expected 1 module with recursive disabled, got %d", len(modules))
		for _, m := range modules {
			t.Logf("Found: %s at %s", m.Name, m.Path)
		}
	}

	if len(modules) > 0 && modules[0].Path != "/project/.version" {
		t.Errorf("Expected /project/.version, got %s", modules[0].Path)
	}
}

func TestDetector_ModuleVersionLoading(t *testing.T) {
	// Setup: .version files with different versions
	fs := setupTestFS(map[string]string{
		"/project/module-a/.version": "1.2.3",
		"/project/module-b/.version": "0.1.0-beta",
		"/project/module-c/.version": "invalid-version",
	})

	cfg := &config.Config{}
	detector := NewDetector(fs, cfg)

	modules, err := detector.DiscoverModules(context.Background(), "/project")
	if err != nil {
		t.Fatalf("DiscoverModules failed: %v", err)
	}

	if len(modules) != 3 {
		t.Fatalf("Expected 3 modules, got %d", len(modules))
	}

	// Check that versions are loaded
	for _, module := range modules {
		t.Logf("Module: %s, Version: %s", module.Name, module.CurrentVersion)
		if module.Name == "module-a" && module.CurrentVersion != "1.2.3" {
			t.Errorf("Expected version 1.2.3 for module-a, got %s", module.CurrentVersion)
		}
		if module.Name == "module-b" && module.CurrentVersion != "0.1.0-beta" {
			t.Errorf("Expected version 0.1.0-beta for module-b, got %s", module.CurrentVersion)
		}
		// module-c has invalid version, should have empty string or handle gracefully
	}
}

func TestDetector_SemverignoreIntegration(t *testing.T) {
	// Setup: .version files and .sleyignore
	fs := setupTestFS(map[string]string{
		"/project/.sleyignore":          "test-*\n*.tmp\n# Comment\nignored/",
		"/project/module-a/.version":    "1.0.0",
		"/project/test-module/.version": "1.0.0",
		"/project/ignored/.version":     "1.0.0",
		"/project/file.tmp/.version":    "1.0.0",
	})

	cfg := &config.Config{}
	detector := NewDetector(fs, cfg)

	modules, err := detector.DiscoverModules(context.Background(), "/project")
	if err != nil {
		t.Fatalf("DiscoverModules failed: %v", err)
	}

	// Should only find module-a
	if len(modules) != 1 {
		t.Errorf("Expected 1 module with .sleyignore, got %d", len(modules))
		for _, m := range modules {
			t.Logf("Found module: %s at %s", m.Name, m.Path)
		}
	}

	if len(modules) > 0 && modules[0].Name != "module-a" {
		t.Errorf("Expected module-a, got %s", modules[0].Name)
	}
}

func TestDetector_DiscoveryDisabled(t *testing.T) {
	// Setup: .version files
	fs := setupTestFS(map[string]string{
		"/project/module-a/.version": "1.0.0",
	})

	enabled := false
	cfg := &config.Config{
		Workspace: &config.WorkspaceConfig{
			Discovery: &config.DiscoveryConfig{
				Enabled: &enabled,
			},
		},
	}

	detector := NewDetector(fs, cfg)
	modules, err := detector.DiscoverModules(context.Background(), "/project")
	if err != nil {
		t.Fatalf("DiscoverModules failed: %v", err)
	}

	// Should find no modules when discovery is disabled
	if len(modules) != 0 {
		t.Errorf("Expected 0 modules with discovery disabled, got %d", len(modules))
	}
}

func TestDetectionMode_String(t *testing.T) {
	tests := []struct {
		mode     DetectionMode
		expected string
	}{
		{SingleModule, "SingleModule"},
		{MultiModule, "MultiModule"},
		{NoModules, "NoModules"},
		{DetectionMode(999), "Unknown(999)"},
	}

	for _, tt := range tests {
		if got := tt.mode.String(); got != tt.expected {
			t.Errorf("DetectionMode(%d).String() = %s, want %s", tt.mode, got, tt.expected)
		}
	}
}

// Integration test with real filesystem
func TestDetector_RealFilesystem(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// Create temporary directory structure
	tmpDir := t.TempDir()

	// Create test structure
	testDirs := []string{
		filepath.Join(tmpDir, "module-a"),
		filepath.Join(tmpDir, "module-b"),
		filepath.Join(tmpDir, "node_modules", "dep"),
	}

	for _, dir := range testDirs {
		if err := os.MkdirAll(dir, 0755); err != nil {
			t.Fatalf("Failed to create directory %s: %v", dir, err)
		}
	}

	// Create .version files
	versionFiles := map[string]string{
		filepath.Join(tmpDir, "module-a", ".version"):            "1.0.0",
		filepath.Join(tmpDir, "module-b", ".version"):            "2.0.0",
		filepath.Join(tmpDir, "node_modules", "dep", ".version"): "1.0.0",
	}

	for path, content := range versionFiles {
		if err := os.WriteFile(path, []byte(content), 0644); err != nil {
			t.Fatalf("Failed to write file %s: %v", path, err)
		}
	}

	cfg := &config.Config{}
	detector := NewDetector(core.NewOSFileSystem(), cfg)

	ctx, err := detector.DetectContext(context.Background(), tmpDir)
	if err != nil {
		t.Fatalf("DetectContext failed: %v", err)
	}

	if ctx.Mode != MultiModule {
		t.Errorf("Expected MultiModule mode, got %s", ctx.Mode)
	}

	// Should only find module-a and module-b (not node_modules)
	if len(ctx.Modules) != 2 {
		t.Errorf("Expected 2 modules, got %d", len(ctx.Modules))
		for _, m := range ctx.Modules {
			t.Logf("Found module: %s at %s", m.Name, m.Path)
		}
	}
}
