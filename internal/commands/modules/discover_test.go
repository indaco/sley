package modules

import (
	"context"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/indaco/sley/internal/config"
	"github.com/indaco/sley/internal/testutils"
	"github.com/urfave/cli/v3"
)

// assertOutputContains checks that output contains all expected strings.
func assertOutputContains(t *testing.T, output string, expectedStrings []string) {
	t.Helper()
	for _, expected := range expectedStrings {
		if !strings.Contains(output, expected) {
			t.Errorf("expected %q in output, got: %q", expected, output)
		}
	}
}

func TestDiscoverCmd(t *testing.T) {
	cmd := discoverCmd()

	if cmd == nil {
		t.Fatal("discoverCmd() returned nil")
	}

	if cmd.Name != "discover" {
		t.Errorf("command name = %q, want %q", cmd.Name, "discover")
	}

	expectedUsage := "Test module discovery without running operations"
	if cmd.Usage != expectedUsage {
		t.Errorf("command usage = %q, want %q", cmd.Usage, expectedUsage)
	}

	if cmd.Action == nil {
		t.Error("command has no action")
	}
}

func TestDiscoverCmd_Flags(t *testing.T) {
	cmd := discoverCmd()

	expectedFlags := map[string]bool{
		"dry-run": true,
	}

	foundFlags := make(map[string]bool)
	for _, flag := range cmd.Flags {
		for _, name := range flag.Names() {
			foundFlags[name] = true
		}
	}

	for expectedFlag := range expectedFlags {
		if !foundFlags[expectedFlag] {
			t.Errorf("missing expected flag %q", expectedFlag)
		}
	}
}

func TestDiscoverCmd_DryRunDefault(t *testing.T) {
	cmd := discoverCmd()

	// Verify the dry-run flag exists
	var foundDryRunFlag bool
	for _, flag := range cmd.Flags {
		if slices.Contains(flag.Names(), "dry-run") {
			foundDryRunFlag = true
		}
	}

	if !foundDryRunFlag {
		t.Fatal("dry-run flag not found")
	}
}

func TestRunDiscover_ConfigLoadError(t *testing.T) {
	// Save original and restore after test
	originalLoadConfig := config.LoadConfigFn
	defer func() {
		config.LoadConfigFn = originalLoadConfig
	}()

	// Mock LoadConfigFn to return an error
	config.LoadConfigFn = func() (*config.Config, error) {
		return nil, os.ErrNotExist
	}

	ctx := context.Background()

	// Create a mock CLI command context
	mockCmd := &cli.Command{
		Name: "test",
		Flags: []cli.Flag{
			&cli.BoolFlag{Name: "dry-run", Value: true},
		},
	}

	err := runDiscover(ctx, mockCmd)
	if err == nil {
		t.Fatal("expected error from runDiscover, got nil")
	}

	expectedErrMsg := "failed to load config"
	if !strings.Contains(err.Error(), expectedErrMsg) {
		t.Errorf("error message = %q, want to contain %q", err.Error(), expectedErrMsg)
	}
}

func TestRunDiscover_GetWdError(t *testing.T) {
	// Save original and restore after test
	originalLoadConfig := config.LoadConfigFn
	defer func() {
		config.LoadConfigFn = originalLoadConfig
	}()

	// Mock LoadConfigFn to return valid config
	config.LoadConfigFn = func() (*config.Config, error) {
		return &config.Config{}, nil
	}

	// Change to a directory that will be deleted
	tmpDir := t.TempDir()
	subDir := filepath.Join(tmpDir, "subdir")
	if err := os.Mkdir(subDir, 0755); err != nil {
		t.Fatalf("failed to create subdirectory: %v", err)
	}

	origDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get working directory: %v", err)
	}
	defer func() { _ = os.Chdir(origDir) }()

	if err := os.Chdir(subDir); err != nil {
		t.Fatalf("failed to change to subdir: %v", err)
	}

	// Remove the directory we're in
	if err := os.RemoveAll(tmpDir); err != nil {
		t.Fatalf("failed to remove tmpDir: %v", err)
	}

	ctx := context.Background()

	mockCmd := &cli.Command{
		Name: "test",
		Flags: []cli.Flag{
			&cli.BoolFlag{Name: "dry-run", Value: true},
		},
	}

	err = runDiscover(ctx, mockCmd)
	// On some systems this might succeed, on others it fails
	// Just verify it doesn't panic
	_ = err
}

func TestRunDiscover_NoModules(t *testing.T) {
	tmpDir := t.TempDir()

	// Save original and restore after test
	originalLoadConfig := config.LoadConfigFn
	defer func() {
		config.LoadConfigFn = originalLoadConfig
	}()

	// Mock LoadConfigFn
	config.LoadConfigFn = func() (*config.Config, error) {
		return &config.Config{}, nil
	}

	origDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get working directory: %v", err)
	}
	defer func() { _ = os.Chdir(origDir) }()

	if err := os.Chdir(tmpDir); err != nil {
		t.Fatalf("failed to change directory: %v", err)
	}

	ctx := context.Background()

	mockCmd := &cli.Command{
		Name: "test",
		Flags: []cli.Flag{
			&cli.BoolFlag{Name: "dry-run", Value: true},
		},
	}

	output, err := testutils.CaptureStdout(func() {
		if err := runDiscover(ctx, mockCmd); err != nil {
			t.Errorf("runDiscover failed: %v", err)
		}
	})

	if err != nil {
		t.Fatalf("failed to capture stdout: %v", err)
	}

	// Should show discovery settings
	if !strings.Contains(output, "Discovery settings:") {
		t.Errorf("expected 'Discovery settings:' in output, got: %q", output)
	}

	// Should show detection mode
	if !strings.Contains(output, "Detection mode:") {
		t.Errorf("expected 'Detection mode:' in output, got: %q", output)
	}

	// Should show no modules
	if !strings.Contains(output, "No modules found in workspace") {
		t.Errorf("expected 'No modules found in workspace' in output, got: %q", output)
	}
}

func TestRunDiscover_SingleModule(t *testing.T) {
	tmpDir := t.TempDir()

	// Create a version file in the root
	versionFile := filepath.Join(tmpDir, ".version")
	if err := os.WriteFile(versionFile, []byte("1.0.0"), 0644); err != nil {
		t.Fatalf("failed to create version file: %v", err)
	}

	// Save original and restore after test
	originalLoadConfig := config.LoadConfigFn
	defer func() {
		config.LoadConfigFn = originalLoadConfig
	}()

	// Mock LoadConfigFn
	config.LoadConfigFn = func() (*config.Config, error) {
		enabled := true
		recursive := true
		maxDepth := 10
		return &config.Config{
			Workspace: &config.WorkspaceConfig{
				Discovery: &config.DiscoveryConfig{
					Enabled:   &enabled,
					Recursive: &recursive,
					MaxDepth:  &maxDepth,
				},
			},
		}, nil
	}

	origDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get working directory: %v", err)
	}
	defer func() { _ = os.Chdir(origDir) }()

	if err := os.Chdir(tmpDir); err != nil {
		t.Fatalf("failed to change directory: %v", err)
	}

	ctx := context.Background()

	mockCmd := &cli.Command{
		Name: "test",
		Flags: []cli.Flag{
			&cli.BoolFlag{Name: "dry-run", Value: true},
		},
	}

	output, err := testutils.CaptureStdout(func() {
		if err := runDiscover(ctx, mockCmd); err != nil {
			t.Errorf("runDiscover failed: %v", err)
		}
	})

	if err != nil {
		t.Fatalf("failed to capture stdout: %v", err)
	}

	// Should show discovery settings
	if !strings.Contains(output, "Discovery settings:") {
		t.Errorf("expected 'Discovery settings:' in output, got: %q", output)
	}

	// Should show detection mode
	if !strings.Contains(output, "Detection mode:") {
		t.Errorf("expected 'Detection mode:' in output, got: %q", output)
	}

	// Should show single module
	if !strings.Contains(output, "Single module found") {
		t.Errorf("expected 'Single module found' in output, got: %q", output)
	}
}

func TestRunDiscover_MultipleModules(t *testing.T) {
	tmpDir := t.TempDir()

	// Create multiple module directories
	moduleA := filepath.Join(tmpDir, "module-a")
	if err := os.MkdirAll(moduleA, 0755); err != nil {
		t.Fatalf("failed to create module-a: %v", err)
	}
	if err := os.WriteFile(filepath.Join(moduleA, ".version"), []byte("1.0.0"), 0644); err != nil {
		t.Fatalf("failed to create .version for module-a: %v", err)
	}

	moduleB := filepath.Join(tmpDir, "module-b")
	if err := os.MkdirAll(moduleB, 0755); err != nil {
		t.Fatalf("failed to create module-b: %v", err)
	}
	if err := os.WriteFile(filepath.Join(moduleB, ".version"), []byte("2.0.0"), 0644); err != nil {
		t.Fatalf("failed to create .version for module-b: %v", err)
	}

	// Save original and restore after test
	originalLoadConfig := config.LoadConfigFn
	defer func() {
		config.LoadConfigFn = originalLoadConfig
	}()

	// Mock LoadConfigFn
	config.LoadConfigFn = func() (*config.Config, error) {
		enabled := true
		recursive := true
		maxDepth := 10
		return &config.Config{
			Workspace: &config.WorkspaceConfig{
				Discovery: &config.DiscoveryConfig{
					Enabled:   &enabled,
					Recursive: &recursive,
					MaxDepth:  &maxDepth,
				},
			},
		}, nil
	}

	origDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get working directory: %v", err)
	}
	defer func() { _ = os.Chdir(origDir) }()

	if err := os.Chdir(tmpDir); err != nil {
		t.Fatalf("failed to change directory: %v", err)
	}

	ctx := context.Background()

	mockCmd := &cli.Command{
		Name: "test",
		Flags: []cli.Flag{
			&cli.BoolFlag{Name: "dry-run", Value: true},
		},
	}

	output, err := testutils.CaptureStdout(func() {
		if err := runDiscover(ctx, mockCmd); err != nil {
			t.Errorf("runDiscover failed: %v", err)
		}
	})

	if err != nil {
		t.Fatalf("failed to capture stdout: %v", err)
	}

	assertOutputContains(t, output, []string{
		"Discovery settings:",
		"Enabled:",
		"Recursive:",
		"Max depth:",
		"Exclude patterns:",
		"Detection mode:",
		"Multiple modules found",
	})
}

func TestRunDiscover_ShowsDiscoveryConfig(t *testing.T) {
	tmpDir := t.TempDir()

	// Save original and restore after test
	originalLoadConfig := config.LoadConfigFn
	defer func() {
		config.LoadConfigFn = originalLoadConfig
	}()

	// Mock LoadConfigFn with specific discovery settings
	enabled := true
	recursive := false
	maxDepth := 5
	config.LoadConfigFn = func() (*config.Config, error) {
		return &config.Config{
			Workspace: &config.WorkspaceConfig{
				Discovery: &config.DiscoveryConfig{
					Enabled:   &enabled,
					Recursive: &recursive,
					MaxDepth:  &maxDepth,
					Exclude:   []string{"test", "tmp"},
				},
			},
		}, nil
	}

	origDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get working directory: %v", err)
	}
	defer func() { _ = os.Chdir(origDir) }()

	if err := os.Chdir(tmpDir); err != nil {
		t.Fatalf("failed to change directory: %v", err)
	}

	ctx := context.Background()

	mockCmd := &cli.Command{
		Name: "test",
		Flags: []cli.Flag{
			&cli.BoolFlag{Name: "dry-run", Value: true},
		},
	}

	output, err := testutils.CaptureStdout(func() {
		if err := runDiscover(ctx, mockCmd); err != nil {
			t.Errorf("runDiscover failed: %v", err)
		}
	})

	if err != nil {
		t.Fatalf("failed to capture stdout: %v", err)
	}

	// Verify settings are displayed
	if !strings.Contains(output, "Enabled: true") {
		t.Errorf("expected 'Enabled: true' in output, got: %q", output)
	}

	if !strings.Contains(output, "Recursive: false") {
		t.Errorf("expected 'Recursive: false' in output, got: %q", output)
	}

	if !strings.Contains(output, "Max depth: 5") {
		t.Errorf("expected 'Max depth: 5' in output, got: %q", output)
	}
}

func TestRunDiscover_NilConfig(t *testing.T) {
	tmpDir := t.TempDir()

	// Save original and restore after test
	originalLoadConfig := config.LoadConfigFn
	defer func() {
		config.LoadConfigFn = originalLoadConfig
	}()

	// Mock LoadConfigFn to return nil config
	config.LoadConfigFn = func() (*config.Config, error) {
		return nil, nil
	}

	origDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get working directory: %v", err)
	}
	defer func() { _ = os.Chdir(origDir) }()

	if err := os.Chdir(tmpDir); err != nil {
		t.Fatalf("failed to change directory: %v", err)
	}

	ctx := context.Background()

	mockCmd := &cli.Command{
		Name: "test",
		Flags: []cli.Flag{
			&cli.BoolFlag{Name: "dry-run", Value: true},
		},
	}

	output, err := testutils.CaptureStdout(func() {
		if err := runDiscover(ctx, mockCmd); err != nil {
			t.Errorf("runDiscover failed: %v", err)
		}
	})

	if err != nil {
		t.Fatalf("failed to capture stdout: %v", err)
	}

	// Should still work with default config
	if !strings.Contains(output, "Discovery settings:") {
		t.Errorf("expected 'Discovery settings:' in output, got: %q", output)
	}
}

func TestRunDiscover_ModuleWithoutVersion(t *testing.T) {
	tmpDir := t.TempDir()

	// Create a module with empty version
	moduleA := filepath.Join(tmpDir, "module-a")
	if err := os.MkdirAll(moduleA, 0755); err != nil {
		t.Fatalf("failed to create module-a: %v", err)
	}
	if err := os.WriteFile(filepath.Join(moduleA, ".version"), []byte(""), 0644); err != nil {
		t.Fatalf("failed to create .version for module-a: %v", err)
	}

	moduleB := filepath.Join(tmpDir, "module-b")
	if err := os.MkdirAll(moduleB, 0755); err != nil {
		t.Fatalf("failed to create module-b: %v", err)
	}
	if err := os.WriteFile(filepath.Join(moduleB, ".version"), []byte("2.0.0"), 0644); err != nil {
		t.Fatalf("failed to create .version for module-b: %v", err)
	}

	// Save original and restore after test
	originalLoadConfig := config.LoadConfigFn
	defer func() {
		config.LoadConfigFn = originalLoadConfig
	}()

	// Mock LoadConfigFn
	config.LoadConfigFn = func() (*config.Config, error) {
		enabled := true
		recursive := true
		maxDepth := 10
		return &config.Config{
			Workspace: &config.WorkspaceConfig{
				Discovery: &config.DiscoveryConfig{
					Enabled:   &enabled,
					Recursive: &recursive,
					MaxDepth:  &maxDepth,
				},
			},
		}, nil
	}

	origDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get working directory: %v", err)
	}
	defer func() { _ = os.Chdir(origDir) }()

	if err := os.Chdir(tmpDir); err != nil {
		t.Fatalf("failed to change directory: %v", err)
	}

	ctx := context.Background()

	mockCmd := &cli.Command{
		Name: "test",
		Flags: []cli.Flag{
			&cli.BoolFlag{Name: "dry-run", Value: true},
		},
	}

	output, err := testutils.CaptureStdout(func() {
		if err := runDiscover(ctx, mockCmd); err != nil {
			t.Errorf("runDiscover failed: %v", err)
		}
	})

	if err != nil {
		t.Fatalf("failed to capture stdout: %v", err)
	}

	// Should show "unknown" for module without version
	if !strings.Contains(output, "(unknown)") {
		t.Errorf("expected '(unknown)' for module without version, got: %q", output)
	}
}

func TestRunDiscover_DetectContextError(t *testing.T) {
	tmpDir := t.TempDir()

	// Create an inaccessible subdirectory to trigger detection error
	restrictedDir := filepath.Join(tmpDir, "restricted")
	if err := os.MkdirAll(restrictedDir, 0755); err != nil {
		t.Fatalf("failed to create restricted directory: %v", err)
	}

	// Create a .version file in the restricted directory
	versionFile := filepath.Join(restrictedDir, ".version")
	if err := os.WriteFile(versionFile, []byte("1.0.0"), 0644); err != nil {
		t.Fatalf("failed to create version file: %v", err)
	}

	// Make directory inaccessible (this might not work on all systems)
	if err := os.Chmod(restrictedDir, 0000); err != nil {
		t.Fatalf("failed to change permissions: %v", err)
	}
	defer func() { _ = os.Chmod(restrictedDir, 0755) }()

	// Save original and restore after test
	originalLoadConfig := config.LoadConfigFn
	defer func() {
		config.LoadConfigFn = originalLoadConfig
	}()

	// Mock LoadConfigFn
	config.LoadConfigFn = func() (*config.Config, error) {
		enabled := true
		recursive := true
		maxDepth := 10
		return &config.Config{
			Workspace: &config.WorkspaceConfig{
				Discovery: &config.DiscoveryConfig{
					Enabled:   &enabled,
					Recursive: &recursive,
					MaxDepth:  &maxDepth,
				},
			},
		}, nil
	}

	origDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get working directory: %v", err)
	}
	defer func() { _ = os.Chdir(origDir) }()

	if err := os.Chdir(tmpDir); err != nil {
		t.Fatalf("failed to change directory: %v", err)
	}

	ctx := context.Background()

	mockCmd := &cli.Command{
		Name: "test",
		Flags: []cli.Flag{
			&cli.BoolFlag{Name: "dry-run", Value: true},
		},
	}

	// This test might not fail on all systems due to permission handling differences
	// We just ensure it doesn't panic
	_, _ = testutils.CaptureStdout(func() {
		_ = runDiscover(ctx, mockCmd)
	})
}
